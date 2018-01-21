package jobs

import (
	"crashtested-backend/common/http/helpers"
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ImportHelmetDataJob struct {
	ProductRepository *repositories.ProductRepository
}

// Get SHARP data
// Get SNELL data

// Must be run first:
// For each helmet in SHARP, try to find helmets by manufacturer+model combo
// does it already exist and are the SHARP fields different? If so, replace SHARP subdocument; else, create document.

// The below 2 steps can be run in parallel:

// For each helmet in SNELL, try to find helmets by manufacturer+model combo
// does it already exist? If so, set document.certifications.SNELL to true if it isn't already true; else, create document and log a warning that we couldn't find a matching SHARP helmet.

// For each helmet in the database, query CJ Affiliate's product data using Helmet manufacturer + model. Order by price descending, take top result, get product description.
// If no results, log a warning; if results:
// does description contain "DOT"? Set DOT to true.
// set price to the price
// if request limit reached, wait for 1.5 minutes and keep going
func (*ImportHelmetDataJob) Run() error {
	products := make([]*entities.ProductDocument, 0)
	form := url.Values{}
	form.Add("action", "more_helmet_ajax")
	form.Add("postsperpage", "500000") // "500000") // TODO: use 500000
	form.Add("manufacturer", "All")
	form.Add("model", "All")
	form.Add("pageNumber", "1")
	form.Add("type", "1")

	resp, err := helpers.MakeFormPOSTRequest("https://sharp.dft.gov.uk/wp-admin/admin-ajax.php", form)
	if err != nil {
		return err
	}
	resp = "<html><table>" + resp + "</table></html>" // SHARP's undocumented API returns invalid HTML with no root node, so we have to add the root nodes ourselves
	responseReader := strings.NewReader(resp)

	doc, err := goquery.NewDocumentFromReader(responseReader)
	if err != nil {
		return err
	}
	rows := doc.Find("a[href*='sharp.dft.gov.uk/helmets/']")
	helmetUrlsMap := make(map[string]bool)
	for linkIndex, _ := range rows.Nodes {
		linkSelection := rows.Eq(linkIndex)
		url, linkExists := linkSelection.Attr("href")
		if !linkExists {
			logrus.WithField("linkIndex", linkIndex).Warn("Encountered an empty link while parsing SHARP data")
			continue
		}

		if _, exists := helmetUrlsMap[url]; !exists {
			helmetUrlsMap[url] = true
		}
	}

	starsRegexp := regexp.MustCompile(`rating-star-(\d)`)
	topImpactZoneRegexp := regexp.MustCompile(`front-(\d)-(\d)\.jpg`) // SHARP calls this front-front and front-rear which isn't correct, it's actually top-front and top-rear
	leftImpactZoneRegexp := regexp.MustCompile(`left-(\d)\.jpg`)
	rightImpactZoneRegexp := regexp.MustCompile(`right-(\d)\.jpg`)
	rearImpactZoneRegexp := regexp.MustCompile(`rear-(\d)\.jpg`)
	weightRegexp := regexp.MustCompile(`(\d\.\d\d)`)
	startTime := time.Now()
	productsChannel := make(chan *entities.ProductDocument, len(helmetUrlsMap))
	for helmetUrl := range helmetUrlsMap {
		product, err := parseProductBySHARPHelmetUrl(helmetUrl, productsChannel, weightRegexp, starsRegexp, topImpactZoneRegexp, leftImpactZoneRegexp, rightImpactZoneRegexp, rearImpactZoneRegexp)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"helmetUrl": helmetUrl,
				"error":     err,
			}).Error("Failed to parse a helmet, aborting process")
			return err
		}
		products = append(products, product)
	}
	timeTaken := time.Now().Unix() - startTime.Unix() // 404 seconds synchronously
	logrus.Info(len(products), timeTaken)
	return nil
}

func parseProductBySHARPHelmetUrl(helmetUrl string, productsChannel chan *entities.ProductDocument, weightRegexp *regexp.Regexp, starsRegexp *regexp.Regexp, topImpactZoneRegexp *regexp.Regexp, leftImpactZoneRegexp *regexp.Regexp, rightImpactZoneRegexp *regexp.Regexp, rearImpactZoneRegexp *regexp.Regexp) (*entities.ProductDocument, error) {
	helmetLogger := logrus.WithField("helmetUrl", helmetUrl)
	resp, err := http.Get(helmetUrl)
	if err != nil {
		return nil, err
	}

	helmetDetailsDoc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	productImageUrl, found := helmetDetailsDoc.Find(".wp-post-image").First().Attr("src")
	if !found {
		helmetLogger.Warn("Product image not found")
	}

	impactZoneRatings := &entities.SHARPImpactZoneRatingsDocument{}
	impactZoneImages := helmetDetailsDoc.Find("img[src*='impact-zones/dots']")
	for index, _ := range impactZoneImages.Nodes {
		impactZoneImageSelection := impactZoneImages.Eq(index)
		impactZoneRatings, err = getImpactZoneRatings(helmetLogger, impactZoneImageSelection, leftImpactZoneRegexp, rightImpactZoneRegexp, topImpactZoneRegexp, rearImpactZoneRegexp)
		if err != nil {
			return nil, err
		}
	}

	model := findDetailsTextByHeader(helmetDetailsDoc, "model")

	starsSelection := findDetailsSelectionByHeader(helmetDetailsDoc, "helmet rating")
	starsImageUrl, _ := starsSelection.ChildrenFiltered("img").First().Attr("src")
	subMatchArray := starsRegexp.FindStringSubmatch(starsImageUrl)
	if len(subMatchArray) < 2 {
		return nil, errors.New("Encountered an unexpected star rating array")
	}
	starsValue, err := strconv.Atoi(subMatchArray[1])
	if err != nil {
		return nil, err
	}

	manufacturer := findDetailsTextByHeader(helmetDetailsDoc, "manufacturer")
	rawWeightText := findDetailsTextByHeader(helmetDetailsDoc, "helmet weight")
	weightInLbs := float64(-1)

	weightMatches := weightRegexp.FindStringSubmatch(strings.Replace(rawWeightText, ",", ".", -1))
	if len(weightMatches) > 1 {
		weightText := weightMatches[1]
		weightInKg, err := strconv.ParseFloat(weightText, 64)
		if err != nil {
			return nil, err
		}
		weightInLbs = float64(2.20462) * weightInKg
		weightInLbs = float64(int64(weightInLbs/0.01+0.5)) * 0.01
	} else {
		helmetLogger.Warning("An unexpected weight was encountered, setting weight to -1 lbs")
	}

	sizesText := findDetailsTextByHeader(helmetDetailsDoc, "helmet sizes")
	sizes := strings.Split(sizesText, " ")

	subtypeText := findDetailsTextByHeader(helmetDetailsDoc, "helmet type")
	subtype := ""
	if strings.EqualFold(subtypeText, "system") {
		subtype = "modular"
	} else if strings.EqualFold(subtypeText, "full face") {
		subtype = "full"
	}

	latchPercentageArray := (strings.Split(helmetDetailsDoc.Find(".percentage-overlay").Text(), "%"))
	latchPercentage := -1
	if len(latchPercentageArray) >= 2 {
		var err error
		latchPercentage, err = strconv.Atoi(latchPercentageArray[0])
		if err != nil {
			return nil, err
		}
	} else if subtype == "modular" {
		helmetLogger.Warn("Encountered a modular helmet with a latch percentage array that did not contain at least 2 elements, assuming empty latch percentage")
	}

	retentionSystem := findDetailsTextByHeader(helmetDetailsDoc, "retention system")
	materials := findDetailsTextByHeader(helmetDetailsDoc, "materials")
	otherStandardsText := findDetailsTextByHeader(helmetDetailsDoc, "other standards")
	isECERated := strings.Contains(otherStandardsText, "ECE")

	product := &entities.ProductDocument{
		UUID:            uuid.New(),
		Type:            "helmet",
		Subtype:         subtype,
		Model:           model,
		Manufacturer:    manufacturer,
		ImageURL:        productImageUrl,
		LatchPercentage: latchPercentage,
		WeightInLbs:     weightInLbs,
		Sizes:           sizes,
		RetentionSystem: retentionSystem,
		Materials:       materials,
	}
	product.Certifications.ECE = isECERated
	product.Certifications.SHARP = &entities.SHARPCertificationDocument{Stars: starsValue, ImpactZoneRatings: impactZoneRatings}
	return product, nil
}

func getImpactZoneRatings(helmetLogger *logrus.Entry, impactZoneImageSelection *goquery.Selection, leftImpactZoneRegexp *regexp.Regexp, rightImpactZoneRegexp *regexp.Regexp, topImpactZoneRegexp *regexp.Regexp, rearImpactZoneRegexp *regexp.Regexp) (*entities.SHARPImpactZoneRatingsDocument, error) {
	impactZoneImageUrl, impactZoneImageUrlExists := impactZoneImageSelection.Attr("src")
	if !impactZoneImageUrlExists {
		errString := "Impact zone image url not found"
		helmetLogger.Error(errString)
		return nil, errors.New(errString)
	}

	var err error
	impactZoneRatings := &entities.SHARPImpactZoneRatingsDocument{}
	if strings.Index(impactZoneImageUrl, "left") >= 0 {
		impactZoneRatings.Left, err = getImpactZoneRating(impactZoneImageUrl, leftImpactZoneRegexp)
		if err != nil {
			return nil, err
		}
	} else if strings.Index(impactZoneImageUrl, "right") >= 0 {
		impactZoneRatings.Right, err = getImpactZoneRating(impactZoneImageUrl, rightImpactZoneRegexp)
		if err != nil {
			return nil, err
		}
	} else if strings.Index(impactZoneImageUrl, "front") >= 0 {
		impactZoneRatings.Top.Front, impactZoneRatings.Top.Rear, err = getTopImpactZoneRatings(impactZoneImageUrl, topImpactZoneRegexp)
		if err != nil {
			return nil, err
		}
	} else if strings.Index(impactZoneImageUrl, "rear") >= 0 {
		impactZoneRatings.Rear, err = getImpactZoneRating(impactZoneImageUrl, rearImpactZoneRegexp)
		if err != nil {
			return nil, err
		}
	} else {
		unknownImpactZoneError := "Encountered an unknown impact zone rating"
		helmetLogger.Info(unknownImpactZoneError)
		return nil, errors.New(unknownImpactZoneError)
	}

	return impactZoneRatings, nil
}

func getImpactZoneRating(url string, regexp *regexp.Regexp) (int, error) {
	if regexp.MatchString(url) {
		matches := regexp.FindStringSubmatch(url)
		if len(matches) < 2 {
			return -1, errors.New(fmt.Sprintf("Encountered less than two matches for the %s impact zone regex", url))
		}
		rating, err := strconv.Atoi(matches[1])
		if err != nil {
			return -1, err
		}
		return rating, nil
	}
	return -1, errors.New("The url doesn't match the expected regex")
}

func getTopImpactZoneRatings(url string, regexp *regexp.Regexp) (int, int, error) {
	if regexp.MatchString(url) {
		matches := regexp.FindStringSubmatch(url)
		if len(matches) < 3 {
			return -1, -1, errors.New(fmt.Sprintf("Encountered less than three matches for the %s impact zone regex", url))
		}
		topFrontRating, err := strconv.Atoi(matches[1])
		if err != nil {
			return -1, -1, err
		}

		topRearRating, err := strconv.Atoi(matches[2])
		if err != nil {
			return -1, -1, err
		}
		return topFrontRating, topRearRating, nil
	}
	return -1, -1, errors.New("The url doesn't match the expected regex")
}

func findDetailsSelectionByHeader(doc *goquery.Document, headerKey string) *goquery.Selection {
	return doc.Find("tr>td").FilterFunction(func(i int, selection *goquery.Selection) bool {
		currHeader := strings.TrimSpace(selection.Siblings().Filter("th").First().Text())
		return strings.EqualFold(currHeader, headerKey)
	}).First()
}

func findDetailsTextByHeader(doc *goquery.Document, headerKey string) string {
	originalText := findDetailsSelectionByHeader(doc, headerKey).Text()
	tabsRegexp := regexp.MustCompile(`[\t]{2,}`)
	cleanedText := tabsRegexp.ReplaceAllString(originalText, " ")
	return strings.TrimSpace(cleanedText)
}

package parsers

import (
	"crashtested-backend/common/http/helpers"
	"crashtested-backend/persistence/entities"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/sirupsen/logrus"
)

// SHARPHelmetParser contains functions used to scrape helmet data from SHARP's website
type SHARPHelmetParser struct {
	Limit int
}

// GetAll scrapes and returns all helmet data from SHARP's website, or an error if there was a problem fetching/scraping the HTML
func (r *SHARPHelmetParser) GetAll() ([]*entities.SHARPHelmet, error) {
	logrus.Info("Started getting all SHARP helmets")
	helmets := make([]*entities.SHARPHelmet, 0)
	starsRegexp := regexp.MustCompile(`rating-star-(\d)`)
	topImpactZoneRegexp := regexp.MustCompile(`front-(\d)-(\d)\.jpg`) // SHARP calls this front-front and front-rear which isn't correct, it's actually top-front and top-rear
	leftImpactZoneRegexp := regexp.MustCompile(`left-(\d)\.jpg`)
	rightImpactZoneRegexp := regexp.MustCompile(`right-(\d)\.jpg`)
	rearImpactZoneRegexp := regexp.MustCompile(`rear-(\d)\.jpg`)
	weightRegexp := regexp.MustCompile(`(\d\.\d\d)`)
	latchPercentageRegexp := regexp.MustCompile("[0-9]+")

	startTime := time.Now()
	helmetResultsChannel := make(chan *parseHelmetResult)
	httpRequestSemaphore := make(chan struct{}, 4) // maximum of 4 concurrent http requests
	helmetUrlsMap, err := r.GetHelmetUrls()
	if err != nil {
		return nil, err
	}

	numHelmetUrls := len(helmetUrlsMap)
	if numHelmetUrls < 400 {
		return nil, errors.New("Too few helmets were found; check to see if the SHARP website changed its layout")
	}

	pooledHTTPClient := cleanhttp.DefaultPooledClient() // use a pooled http client so that the SSL session is reused between connections
	for helmetURL := range helmetUrlsMap {
		go parseSHARPHelmetByURL(pooledHTTPClient, httpRequestSemaphore, helmetURL, helmetResultsChannel, weightRegexp, starsRegexp, topImpactZoneRegexp, leftImpactZoneRegexp, rightImpactZoneRegexp, rearImpactZoneRegexp, latchPercentageRegexp)
	}

	for index := 0; index < numHelmetUrls; index++ {
		productResult := <-helmetResultsChannel
		if productResult.err != nil {
			logrus.WithFields(logrus.Fields{
				"helmetUrl": productResult.helmetURL,
				"error":     productResult.err,
			}).Error("Encountered an error while processing a SHARP helmet")
			continue
		}

		helmets = append(helmets, productResult.helmet)
	}

	if len(helmets) != numHelmetUrls {
		return nil, errors.New("Did not successfully process all SHARP helmets")
	}

	millisecondsElapsed := time.Now().Sub(startTime).Seconds() * 1000
	logrus.WithField("millisecondsElapsed", millisecondsElapsed).Info("Finished getting all SHARP helmets")
	return helmets, nil
}

// GetHelmetUrls calls an undocumented SHARP endpoint to retrieve a hash set of all the helmet urls on the SHARP website
func (r *SHARPHelmetParser) GetHelmetUrls() (map[string]bool, error) {
	limitToUse := strconv.Itoa(r.Limit)
	if r.Limit < 0 {
		limitToUse = "500000"
	}

	form := url.Values{}
	form.Add("action", "more_helmet_ajax")
	form.Add("postsperpage", limitToUse)
	form.Add("manufacturer", "All")
	form.Add("model", "All")
	form.Add("pageNumber", "1")
	form.Add("type", "1")

	resp, err := helpers.MakeFormPOSTRequest("https://sharp.dft.gov.uk/wp-admin/admin-ajax.php", form)
	if err != nil {
		return nil, err
	}
	resp = "<html><table>" + resp + "</table></html>" // SHARP's undocumented API returns invalid HTML with no root node, so we have to add the root nodes ourselves
	responseReader := strings.NewReader(resp)

	doc, err := goquery.NewDocumentFromReader(responseReader)
	if err != nil {
		return nil, err
	}
	rows := doc.Find("a[href*='sharp.dft.gov.uk/helmets/']")
	helmetUrlsMap := make(map[string]bool)

	numLinks := len(rows.Nodes)
	for linkIndex := 0; linkIndex < numLinks; linkIndex++ {
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
	return helmetUrlsMap, nil
}

type parseHelmetResult struct {
	helmet    *entities.SHARPHelmet
	helmetURL string
	err       error
}

func parseSHARPHelmetByURL(pooledHTTPClient *http.Client, httpRequestsSemaphore chan struct{}, helmetURL string, helmetResultsChannel chan *parseHelmetResult, weightRegexp *regexp.Regexp, starsRegexp *regexp.Regexp, topImpactZoneRegexp *regexp.Regexp, leftImpactZoneRegexp *regexp.Regexp, rightImpactZoneRegexp *regexp.Regexp, rearImpactZoneRegexp *regexp.Regexp, latchPercentageRegexp *regexp.Regexp) {
	helmetLogger := logrus.WithField("helmetUrl", helmetURL)
	helmetLogger.Info("Starting to parse helmet data")

	// increment while we're waiting for the request to finish
	var emptyItem struct{}
	httpRequestsSemaphore <- emptyItem
	resp, err := pooledHTTPClient.Get(helmetURL)
	result := &parseHelmetResult{helmetURL: helmetURL}
	if err != nil {
		result.err = err
		helmetResultsChannel <- result
		return
	}
	helmetDetailsDoc, err := goquery.NewDocumentFromResponse(resp)
	<-httpRequestsSemaphore
	// ^ decrement after the request is done

	if err != nil {
		result.err = err
		helmetResultsChannel <- result
		return
	}

	productImageURL, found := helmetDetailsDoc.Find(".wp-post-image").First().Attr("src")
	if !found {
		helmetLogger.Warn("Product image not found")
	}

	impactZoneRatings := &entities.SHARPImpactZoneRatingsDocument{}
	impactZoneImages := helmetDetailsDoc.Find("img[src*='impact-zones/dots']")
	impactZoneRatings, err = getImpactZoneRatings(helmetLogger, impactZoneImages, leftImpactZoneRegexp, rightImpactZoneRegexp, topImpactZoneRegexp, rearImpactZoneRegexp)
	if err != nil {
		result.err = err
		helmetResultsChannel <- result
		return
	}

	model := findDetailsTextByHeader(helmetDetailsDoc, "model")
	if model == "" {
		helmetLogger.Warn("The model of the helmet was missing")
	}

	priceFrom := findDetailsTextByHeader(helmetDetailsDoc, "price from")
	priceFromItems := strings.Split(priceFrom, "£")
	approximateMSRPCents := 0
	if len(priceFromItems) > 1 {
		approximateMSRPDollars, _ := strconv.ParseFloat(priceFromItems[1], 64)
		approximateMSRPCents = int(math.Trunc(approximateMSRPDollars * 100))
	} else {
		helmetLogger.Warning("The price of the helmet was missing")
	}

	starsSelection := findDetailsSelectionByHeader(helmetDetailsDoc, "helmet rating")
	starsImageURL, _ := starsSelection.ChildrenFiltered("img").First().Attr("src")
	subMatchArray := starsRegexp.FindStringSubmatch(starsImageURL)
	if len(subMatchArray) < 2 {
		result.err = errors.New("Encountered an unexpected star rating array")
		helmetResultsChannel <- result
		return
	}
	starsValue, err := strconv.Atoi(subMatchArray[1])
	if err != nil {
		result.err = err
		helmetResultsChannel <- result
		return
	}

	manufacturer := findDetailsTextByHeader(helmetDetailsDoc, "manufacturer")
	rawWeightText := findDetailsTextByHeader(helmetDetailsDoc, "helmet weight")
	var weightInLbs float64 = -1

	weightMatches := weightRegexp.FindStringSubmatch(strings.Replace(rawWeightText, ",", ".", -1))
	if len(weightMatches) > 1 {
		weightText := weightMatches[1]
		weightInKg, err := strconv.ParseFloat(weightText, 64)
		if err == nil {
			weightInLbs = float64(2.20462) * weightInKg
		} else {
			helmetLogger.WithError(err).Warning("The weight was in an unexpected format")
		}
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
	latchPercentageString := helmetDetailsDoc.Find(".percentage-overlay").Text()
	latchPercentageNumberMatches := latchPercentageRegexp.FindAllString(latchPercentageString, -1)
	latchPercentage := -1
	if len(latchPercentageNumberMatches) == 1 {
		var err error
		latchPercentage, err = strconv.Atoi(strings.TrimSpace(latchPercentageNumberMatches[0]))
		if err != nil {
			helmetLogger.WithError(err).Warning("The latch percentage was in an unexpected format")
			latchPercentage = -1
		}
	} else if subtype == "modular" {
		helmetLogger.Warn("Encountered a modular helmet with a latch percentage array that did not contain at least 2 elements, assuming empty latch percentage")
	}

	retentionSystem := findDetailsTextByHeader(helmetDetailsDoc, "retention system")
	materials := findDetailsTextByHeader(helmetDetailsDoc, "materials")
	otherStandardsText := findDetailsTextByHeader(helmetDetailsDoc, "other standards")
	isECERated := strings.Contains(otherStandardsText, "ECE")

	helmet := &entities.SHARPHelmet{
		Subtype:              subtype,
		Model:                model,
		Manufacturer:         manufacturer,
		ImageURL:             productImageURL,
		LatchPercentage:      latchPercentage,
		WeightInLbs:          weightInLbs,
		Sizes:                sizes,
		RetentionSystem:      retentionSystem,
		Materials:            materials,
		IsECECertified:       isECERated,
		Certifications:       &entities.SHARPCertificationDocument{Stars: starsValue, ImpactZoneRatings: impactZoneRatings},
		ApproximateMSRPCents: approximateMSRPCents,
	}

	result.helmet = helmet
	helmetResultsChannel <- result
	helmetLogger.Info("Finished parsing helmet data")
}

func getImpactZoneRatings(helmetLogger *logrus.Entry, impactZoneImagesSelection *goquery.Selection, leftImpactZoneRegexp *regexp.Regexp, rightImpactZoneRegexp *regexp.Regexp, topImpactZoneRegexp *regexp.Regexp, rearImpactZoneRegexp *regexp.Regexp) (*entities.SHARPImpactZoneRatingsDocument, error) {
	impactZoneRatings := &entities.SHARPImpactZoneRatingsDocument{}
	var err error
	impactZoneImagesSelection.Each(func(index int, selection *goquery.Selection) {
		if err != nil {
			return
		}

		impactZoneImageURL, impactZoneImageURLExists := selection.Attr("src")
		if !impactZoneImageURLExists {
			errString := "Impact zone image url not found"
			helmetLogger.Error(errString)
			err = errors.New(errString)
			return
		}

		if strings.Index(impactZoneImageURL, "left") >= 0 {
			impactZoneRatings.Left, err = getImpactZoneRating(impactZoneImageURL, leftImpactZoneRegexp)
			if err != nil {
				return
			}
		} else if strings.Index(impactZoneImageURL, "right") >= 0 {
			impactZoneRatings.Right, err = getImpactZoneRating(impactZoneImageURL, rightImpactZoneRegexp)
			if err != nil {
				return
			}
		} else if strings.Index(impactZoneImageURL, "front") >= 0 {
			impactZoneRatings.Top.Front, impactZoneRatings.Top.Rear, err = getTopImpactZoneRatings(impactZoneImageURL, topImpactZoneRegexp)
			if err != nil {
				return
			}
		} else if strings.Index(impactZoneImageURL, "rear") >= 0 {
			impactZoneRatings.Rear, err = getImpactZoneRating(impactZoneImageURL, rearImpactZoneRegexp)
			if err != nil {
				return
			}
		} else {
			unknownImpactZoneError := "Encountered an unknown impact zone rating"
			helmetLogger.Info(unknownImpactZoneError)
			err = errors.New(unknownImpactZoneError)
			return
		}
	})

	if err != nil {
		return nil, err
	}
	return impactZoneRatings, nil
}

func getImpactZoneRating(url string, regexp *regexp.Regexp) (int, error) {
	if regexp.MatchString(url) {
		matches := regexp.FindStringSubmatch(url)
		if len(matches) < 2 {
			return -1, fmt.Errorf("Encountered less than two matches for the %s impact zone regex", url)
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
			return -1, -1, fmt.Errorf("Encountered less than three matches for the %s impact zone regex", url)
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

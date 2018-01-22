package repositories

import (
	"crashtested-backend/common/http/helpers"
	"crashtested-backend/persistence/entities"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SHARPHelmetRepository struct {
	Limit int
}

func (self *SHARPHelmetRepository) GetAllHelmets() ([]*entities.SHARPHelmet, error) {
	logrus.Info("Started getting all SHARP helmets")
	helmets := make([]*entities.SHARPHelmet, 0)
	starsRegexp := regexp.MustCompile(`rating-star-(\d)`)
	topImpactZoneRegexp := regexp.MustCompile(`front-(\d)-(\d)\.jpg`) // SHARP calls this front-front and front-rear which isn't correct, it's actually top-front and top-rear
	leftImpactZoneRegexp := regexp.MustCompile(`left-(\d)\.jpg`)
	rightImpactZoneRegexp := regexp.MustCompile(`right-(\d)\.jpg`)
	rearImpactZoneRegexp := regexp.MustCompile(`rear-(\d)\.jpg`)
	weightRegexp := regexp.MustCompile(`(\d\.\d\d)`)
	startTime := time.Now()
	helmetResultsChannel := make(chan *parseHelmetResult)
	httpRequestSemaphore := make(chan struct{}, 4) // maximum of 4 concurrent http requests
	helmetUrlsMap, err := self.GetHelmetUrls()
	if err != nil {
		return nil, err
	}

	numHelmetUrls := len(helmetUrlsMap)
	if numHelmetUrls < 400 {
		return nil, errors.New("Too few helmets were found; check to see if the SHARP website changed its layout")
	}

	pooledHttpClient := cleanhttp.DefaultPooledClient() // use a pooled http client so that the SSL session is reused between connections
	for helmetUrl := range helmetUrlsMap {
		go parseSHARPHelmetByUrl(pooledHttpClient, httpRequestSemaphore, helmetUrl, helmetResultsChannel, weightRegexp, starsRegexp, topImpactZoneRegexp, leftImpactZoneRegexp, rightImpactZoneRegexp, rearImpactZoneRegexp)
	}

	for index := 0; index < numHelmetUrls; index++ {
		productResult := <-helmetResultsChannel
		if productResult.err != nil {
			logrus.WithFields(logrus.Fields{
				"helmetUrl": productResult.helmetUrl,
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

func (self *SHARPHelmetRepository) GetHelmetUrls() (map[string]bool, error) {
	limitToUse := strconv.Itoa(self.Limit)
	if self.Limit < 0 {
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
	return helmetUrlsMap, nil
}

type parseHelmetResult struct {
	helmet    *entities.SHARPHelmet
	helmetUrl string
	err       error
}

func parseSHARPHelmetByUrl(pooledHttpClient *http.Client, httpRequestsSemaphore chan struct{}, helmetUrl string, helmetResultsChannel chan *parseHelmetResult, weightRegexp *regexp.Regexp, starsRegexp *regexp.Regexp, topImpactZoneRegexp *regexp.Regexp, leftImpactZoneRegexp *regexp.Regexp, rightImpactZoneRegexp *regexp.Regexp, rearImpactZoneRegexp *regexp.Regexp) {
	helmetLogger := logrus.WithField("helmetUrl", helmetUrl)
	helmetLogger.Info("Starting to parse helmet data")
	var emptyItem struct{}

	// increment while we're waiting for the request to finish
	httpRequestsSemaphore <- emptyItem
	resp, err := pooledHttpClient.Get(helmetUrl)
	result := &parseHelmetResult{helmetUrl: helmetUrl}
	if err != nil {
		result.err = err
		helmetResultsChannel <- result
		return
	}
	<-httpRequestsSemaphore
	// ^ decrement after the request is done

	helmetDetailsDoc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		result.err = err
		helmetResultsChannel <- result
		return
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
			result.err = err
			helmetResultsChannel <- result
			return
		}
	}

	model := findDetailsTextByHeader(helmetDetailsDoc, "model")

	starsSelection := findDetailsSelectionByHeader(helmetDetailsDoc, "helmet rating")
	starsImageUrl, _ := starsSelection.ChildrenFiltered("img").First().Attr("src")
	subMatchArray := starsRegexp.FindStringSubmatch(starsImageUrl)
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
	weightInLbs := float64(-1)

	weightMatches := weightRegexp.FindStringSubmatch(strings.Replace(rawWeightText, ",", ".", -1))
	if len(weightMatches) > 1 {
		weightText := weightMatches[1]
		weightInKg, err := strconv.ParseFloat(weightText, 64)
		if err != nil {
			result.err = err
			helmetResultsChannel <- result
			return
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
			result.err = err
			helmetResultsChannel <- result
			return
		}
	} else if subtype == "modular" {
		helmetLogger.Warn("Encountered a modular helmet with a latch percentage array that did not contain at least 2 elements, assuming empty latch percentage")
	}

	retentionSystem := findDetailsTextByHeader(helmetDetailsDoc, "retention system")
	materials := findDetailsTextByHeader(helmetDetailsDoc, "materials")
	otherStandardsText := findDetailsTextByHeader(helmetDetailsDoc, "other standards")
	isECERated := strings.Contains(otherStandardsText, "ECE")

	helmet := &entities.SHARPHelmet{
		Subtype:         subtype,
		Model:           model,
		Manufacturer:    manufacturer,
		ImageURL:        productImageUrl,
		LatchPercentage: latchPercentage,
		WeightInLbs:     weightInLbs,
		Sizes:           sizes,
		RetentionSystem: retentionSystem,
		Materials:       materials,
		IsECERated:      isECERated,
		Certifications:  &entities.SHARPCertificationDocument{Stars: starsValue, ImpactZoneRatings: impactZoneRatings},
	}

	result.helmet = helmet
	helmetResultsChannel <- result
	helmetLogger.Info("Finished parsing helmet data")
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

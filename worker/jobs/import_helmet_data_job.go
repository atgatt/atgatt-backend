package jobs

import (
	"crashtested-backend/common/http/helpers"
	"crashtested-backend/persistence/repositories"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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
	form := url.Values{}
	form.Add("action", "more_helmet_ajax")
	form.Add("postsperpage", "2") // "500000") // TODO: use 500000
	form.Add("manufacturer", "All")
	form.Add("model", "All")
	form.Add("pageNumber", "1")
	form.Add("type", "1'")

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
	for linkIndex := range rows.Nodes {
		linkSelection := rows.Eq(linkIndex)
		url, exists := linkSelection.Attr("href")
		if !exists {
			logrus.WithField("linkIndex", linkIndex).Warn("Encountered an empty link while parsing SHARP data")
			continue
		}

		if _, exists := helmetUrlsMap[url]; !exists {
			helmetUrlsMap[url] = true
			logrus.Info(url)
		}
	}

	starsRegexp := regexp.MustCompile(`rating-star-(\d)`)
	for helmetUrl := range helmetUrlsMap {
		resp, err := http.Get(helmetUrl)
		if err != nil {
			return err
		}

		helmetDetailsDoc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			return err
		}

		model := findDetailsTextByHeader(helmetDetailsDoc, "model")

		starsSelection := findDetailsSelectionByHeader(helmetDetailsDoc, "helmet rating")
		starsImageUrl, _ := starsSelection.ChildrenFiltered("img").First().Attr("src")
		subMatchArray := starsRegexp.FindStringSubmatch(starsImageUrl)
		if len(subMatchArray) < 2 {
			return errors.New("Encountered an unexpected star rating array for " + model + ", aborting")
		}
		starsValue, err := strconv.Atoi(subMatchArray[1])
		if err != nil {
			return err
		}

		manufacturer := findDetailsTextByHeader(helmetDetailsDoc, "manufacturer")
		weight := findDetailsTextByHeader(helmetDetailsDoc, "helmet weight")
		// Skipping price from as it's inaccurate
		sizesText := findDetailsTextByHeader(helmetDetailsDoc, "helmet sizes")
		sizes := strings.Split(sizesText, " ")

		subtypeText := findDetailsTextByHeader(helmetDetailsDoc, "helmet type")
		subtype := ""
		if strings.EqualFold(subtypeText, "system") {
			subtype = "modular"
		} else if strings.EqualFold(subtypeText, "full face") {
			subtype = "full"
		}

		retentionSystem := findDetailsTextByHeader(helmetDetailsDoc, "retention system")
		materials := findDetailsTextByHeader(helmetDetailsDoc, "materials")
		otherStandardsText := findDetailsTextByHeader(helmetDetailsDoc, "other standards")
		isECERated := strings.Contains(otherStandardsText, "UN ECE REG 22.05")
		// Skipping manufacturer website as they only list the UK websites
		logrus.Info(starsValue)
	}
	return nil
}

func findDetailsSelectionByHeader(doc *goquery.Document, headerKey string) *goquery.Selection {
	return doc.Find("tr>td").FilterFunction(func(i int, selection *goquery.Selection) bool {
		currHeader := strings.TrimSpace(selection.Siblings().Filter("th").First().Text())
		return strings.EqualFold(currHeader, headerKey)
	}).First()
}

func findDetailsTextByHeader(doc *goquery.Document, headerKey string) string {
	return strings.TrimSpace(findDetailsSelectionByHeader(doc, headerKey).Text())
}

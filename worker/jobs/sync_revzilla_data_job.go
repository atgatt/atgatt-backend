package jobs

import (
	httpHelpers "crashtested-backend/common/http/helpers"
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/worker/jobs/helpers"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	golinq "github.com/ahmetb/go-linq"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/sirupsen/logrus"
	"github.com/xrash/smetrics"
)

// SyncRevzillaDataJob syncs revzilla price and buy urls by calling the CJ Affiliate API and pointing it at RevZilla's advertiser ID
type SyncRevzillaDataJob struct {
	ProductRepository *repositories.ProductRepository
	CJAPIKey          string
}

// Run executes the job
func (j *SyncRevzillaDataJob) Run() error {
	pooledClient := cleanhttp.DefaultPooledClient()

	return helpers.ForEachProduct(j.ProductRepository, func(product *entities.ProductDocument, productLogger *logrus.Entry) error {
		req, err := http.NewRequest("GET", fmt.Sprintf("https://product-search.api.cj.com/v2/product-search?website-id=8505854&advertiser-ids=3318586&keywords=%%2B\"%s\"+%%2B\"%s\"+%%2Bhelmet&page-number=1&records-per-page=100&low-price=200",
			url.QueryEscape(product.Manufacturer), url.QueryEscape(product.Model)), nil)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", j.CJAPIKey)

		resp, err := pooledClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("CJ API returned a status code of %d but expected 200", resp.StatusCode)
		}

		responseBodyBytes, _ := ioutil.ReadAll(resp.Body)
		cjResp := &entities.CJProductsResponseWrapper{}
		if err = xml.Unmarshal(responseBodyBytes, cjResp); err != nil {
			return err
		}

		var matchingRevzillaProductsSlice []entities.CJProduct
		confidenceMap := make(map[string]float64)
		expectedLowerProductName := strings.ToLower(fmt.Sprintf("%s %s", product.Manufacturer, product.Model))

		golinq.From(cjResp.Products.Data).WhereT(func(product entities.CJProduct) bool {
			return product.IsHelmet()
		}).OrderByDescendingT(func(product entities.CJProduct) interface{} {
			lowerProductName := strings.ToLower(product.Name)
			matchConfidence := smetrics.JaroWinkler(lowerProductName, expectedLowerProductName, boostThreshold, prefixSize)
			if _, exists := confidenceMap[lowerProductName]; !exists {
				confidenceMap[lowerProductName] = matchConfidence
			}

			return matchConfidence
		}).ToSlice(&matchingRevzillaProductsSlice)

		if len(matchingRevzillaProductsSlice) > 0 {
			bestMatchRevzillaProduct := &matchingRevzillaProductsSlice[0]
			bestMatchConfidence := confidenceMap[strings.ToLower(bestMatchRevzillaProduct.Name)]
			buyURLContents, err := httpHelpers.GetContentsAtURL(bestMatchRevzillaProduct.BuyURL)
			if err != nil {
				return err
			}

			isDiscontinued := strings.Contains(buyURLContents, "the item you were searching for has been discontinued")
			confidenceLogFields := logrus.Fields{
				"matchConfidence":     bestMatchConfidence,
				"revzillaProductName": bestMatchRevzillaProduct.Name,
				"isDiscontinued":      isDiscontinued,
			}

			if isDiscontinued {
				productLogger.WithFields(confidenceLogFields).Warning("This product is discontinued, updating the discontinued flag and continuing to the next product")
				product.IsDiscontinued = true
				j.ProductRepository.UpdateProduct(product)
			} else if bestMatchConfidence >= 0.8 {
				product.RevzillaBuyURL = bestMatchRevzillaProduct.BuyURL
				product.RevzillaPriceCents = int(bestMatchRevzillaProduct.Price * 100)
				product.UpdateCertificationsByDescription(bestMatchRevzillaProduct.Description)
				product.UpdateSearchPrice()
				productLogger.WithFields(confidenceLogFields).Info("Set new price and buy URL from RevZilla")
				j.ProductRepository.UpdateProduct(product)
			} else {
				productLogger.WithFields(confidenceLogFields).Warning("Could not find a price or buy URL from RevZilla because the best match had a low confidence score")
			}
		} else {
			productLogger.Info("Could not find a price or buy URL from RevZilla because no results were returned")
		}

		time.Sleep(3 * time.Second)
		return nil
	})
}

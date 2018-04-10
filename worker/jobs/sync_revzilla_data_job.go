package jobs

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/worker/jobs/helpers"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/xrash/smetrics"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/sirupsen/logrus"
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

		responseBodyBytes, _ := ioutil.ReadAll(resp.Body)
		cjResp := &entities.CJProductsResponseWrapper{}
		if err = xml.Unmarshal(responseBodyBytes, cjResp); err != nil {
			return err
		}

		matchingRevzillaProducts := cjResp.Products.Data
		if len(matchingRevzillaProducts) > 0 {
			expectedProductName := strings.ToLower(fmt.Sprintf("%s %s helmet", product.Manufacturer, product.Model))
			confidenceMap := make(map[string]float64)

			sort.Slice(matchingRevzillaProducts, func(i, j int) bool {
				firstProduct := matchingRevzillaProducts[i]
				secondProduct := matchingRevzillaProducts[j]
				firstProductName := strings.ToLower(firstProduct.Name)
				secondProductName := strings.ToLower(secondProduct.Name)

				firstProductMatchConfidence := smetrics.JaroWinkler(firstProductName, expectedProductName, boostThreshold, prefixSize)
				secondProductMatchConfidence := smetrics.JaroWinkler(secondProductName, expectedProductName, boostThreshold, prefixSize)

				if !firstProduct.IsHelmet() {
					firstProductMatchConfidence = 0
				}

				if !secondProduct.IsHelmet() {
					secondProductMatchConfidence = 0
				}

				if _, exists := confidenceMap[firstProductName]; !exists {
					confidenceMap[firstProductName] = firstProductMatchConfidence
				}

				if _, exists := confidenceMap[secondProductName]; !exists {
					confidenceMap[secondProductName] = secondProductMatchConfidence
				}

				return firstProductMatchConfidence > secondProductMatchConfidence
			})

			matchingRevzillaProduct := &matchingRevzillaProducts[0]
			matchConfidence := confidenceMap[strings.ToLower(matchingRevzillaProduct.Name)]
			if matchingRevzillaProduct.IsHelmet() && matchConfidence >= 0.8 {
				product.RevzillaBuyURL = matchingRevzillaProduct.BuyURL
				product.RevzillaPriceInUSDMultiple = int(matchingRevzillaProduct.Price * 100)
				product.UpdateCertificationsByDescription(matchingRevzillaProduct.Description)
				product.UpdateMinPrice()
				productLogger.WithFields(logrus.Fields{
					"matchConfidence":     confidenceMap[strings.ToLower(matchingRevzillaProduct.Name)],
					"revzillaProductName": matchingRevzillaProduct.Name,
				}).Info("Set new price and buy URL from RevZilla")
				j.ProductRepository.UpdateProduct(product)
			} else {
				productLogger.Info("Could not find a price or buy URL from RevZilla because the best match was not a helmet or had a low confidence score")
			}
		} else {
			productLogger.Info("Could not find a price or buy URL from RevZilla because no results were returned")
		}

		time.Sleep(3 * time.Second)
		return nil
	})
}

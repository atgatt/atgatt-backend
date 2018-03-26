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
	"time"

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
		req, err := http.NewRequest("GET", fmt.Sprintf("https://product-search.api.cj.com/v2/product-search?website-id=8505854&advertiser-ids=3318586&keywords=%%2B\"%s\"+%%2B\"%s\"+%%2Bhelmet&page-number=1&records-per-page=25&sort-by=price&sort-order=asc&low-price=100",
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
			var matchingRevzillaProduct *entities.CJProduct
			for _, revzillaProduct := range matchingRevzillaProducts {
				if revzillaProduct.IsHelmet() {
					matchingRevzillaProduct = &revzillaProduct
					break
				}
			}

			if matchingRevzillaProduct != nil {
				product.RevzillaBuyURL = matchingRevzillaProduct.BuyURL
				product.RevzillaPriceInUSDMultiple = int(matchingRevzillaProduct.Price * 100)
				product.UpdateCertificationsByDescription(matchingRevzillaProduct.Description)

				productLogger.Info("Set new price and buy URL from RevZilla")
				j.ProductRepository.UpdateProduct(product)
			} else {
				productLogger.Info("Could not find a price or buy URL from RevZilla because none of the results were helmets")
			}
		} else {
			productLogger.Info("Could not find a price or buy URL from RevZilla because no results were returned")
		}

		time.Sleep(3 * time.Second)
		return nil
	})
}

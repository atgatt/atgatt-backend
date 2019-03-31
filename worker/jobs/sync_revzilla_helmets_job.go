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
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/sirupsen/logrus"
	"github.com/xrash/smetrics"
)

// SyncRevzillaHelmetsJob syncs revzilla price and buy urls by calling the CJ Affiliate API and pointing it at RevZilla's advertiser ID
type SyncRevzillaHelmetsJob struct {
	ProductRepository *repositories.ProductRepository
	CJAPIKey          string
}

const bestMatchConfidenceThreshold float64 = 0.8

// Run executes the job
func (j *SyncRevzillaHelmetsJob) Run() error {
	pooledClient := cleanhttp.DefaultPooledClient()
	return helpers.ForEachProduct(j.ProductRepository, func(product *entities.Product, productLogger *logrus.Entry) error {
		modelsToTry := []string{product.Model}
		modelAliasStrings := []string{}
		golinq.From(product.ModelAliases).SelectT(func(modelAlias *entities.ProductModelAlias) string {
			return modelAlias.ModelAlias
		}).ToSlice(&modelAliasStrings)
		modelsToTry = append(modelsToTry, modelAliasStrings...)
		for _, modelToTry := range modelsToTry {
			// Adding a delay because CJ has an absurdly low threshold for requests per minute
			time.Sleep(3 * time.Second)
			synced, err := j.syncDataForProduct(pooledClient, product, modelToTry, productLogger)
			if err != nil {
				return err
			}

			if synced {
				productLogger.Info("Successfully synced product")
				return nil
			}
		}

		productLogger.Info("Could not find a matching product on revzilla, continuing to the next product")
		return nil
	})
}

func (j *SyncRevzillaHelmetsJob) syncDataForProduct(pooledClient *http.Client, product *entities.Product, modelToTry string, productLogger *logrus.Entry) (bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://product-search.api.cj.com/v2/product-search?website-id=8505854&advertiser-ids=3318586&keywords=%%2B\"%s\"+%%2B\"%s\"+%%2Bhelmet&page-number=1&records-per-page=100&low-price=200",
		url.QueryEscape(product.Manufacturer), url.QueryEscape(modelToTry)), nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Bearer "+j.CJAPIKey)

	resp, err := pooledClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("CJ API returned a status code of %d but expected 200", resp.StatusCode)
	}

	responseBodyBytes, _ := ioutil.ReadAll(resp.Body)
	cjResp := &entities.CJProductsResponseWrapper{}
	if err = xml.Unmarshal(responseBodyBytes, cjResp); err != nil {
		return false, err
	}

	var matchingRevzillaProductsSlice []entities.CJProduct
	confidenceMap := make(map[string]float64)
	expectedLowerProductName := strings.ToLower(fmt.Sprintf("%s %s", product.Manufacturer, modelToTry))

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

	syncedProduct := false
	if len(matchingRevzillaProductsSlice) > 0 {
		bestMatchRevzillaProduct := &matchingRevzillaProductsSlice[0]
		bestMatchConfidence := confidenceMap[strings.ToLower(bestMatchRevzillaProduct.Name)]
		buyURLContents, err := httpHelpers.GetContentsAtURL(bestMatchRevzillaProduct.BuyURL)
		if err != nil {
			return false, err
		}

		// If we don't have a product summary, it means we couldn't find the product
		isDiscontinued := !strings.Contains(strings.ToLower(buyURLContents), "product-show-summary")
		confidenceLogFields := logrus.Fields{
			"matchConfidence":             bestMatchConfidence,
			"matchingRevzillaProductName": bestMatchRevzillaProduct.Name,
			"isDiscontinued":              isDiscontinued,
			"modelToTry":                  modelToTry,
		}

		if isDiscontinued && bestMatchConfidence >= bestMatchConfidenceThreshold {
			productLogger.WithFields(confidenceLogFields).Warning("This product is discontinued, updating the discontinued flag and continuing to the next product")
			product.IsDiscontinued = true
			err := j.ProductRepository.UpdateProduct(product)
			if err != nil {
				return false, err
			}
			syncedProduct = true
		} else if bestMatchConfidence >= bestMatchConfidenceThreshold {
			product.RevzillaBuyURL = bestMatchRevzillaProduct.BuyURL
			product.RevzillaPriceCents = int(bestMatchRevzillaProduct.Price * 100)
			product.IsDiscontinued = false
			product.UpdateHelmetCertificationsByDescription(bestMatchRevzillaProduct.Description)
			product.UpdateSearchPrice()
			product.UpdateSafetyPercentage()
			productLogger.WithFields(confidenceLogFields).Info("Set new price and buy URL from RevZilla")
			err := j.ProductRepository.UpdateProduct(product)
			if err != nil {
				return false, err
			}
			syncedProduct = true
		} else {
			productLogger.WithFields(confidenceLogFields).Warning("Could not find a price or buy URL from RevZilla because the best match had a low confidence score")
		}
	} else {
		productLogger.Info("Could not find a price or buy URL from RevZilla because no results were returned")
	}

	return syncedProduct, nil
}

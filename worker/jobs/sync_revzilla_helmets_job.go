package jobs

import (
	httpHelpers "atgatt-backend/common/http"
	"atgatt-backend/persistence/entities"
	"atgatt-backend/persistence/repositories"
	"atgatt-backend/worker/jobs/helpers"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
		var highestConfidenceProductMatch *productMatch = nil
		for _, modelToTry := range modelsToTry {
			currProductMatch, err := j.getBestMatchForProduct(pooledClient, product, modelToTry, productLogger)
			if err != nil {
				return err
			}

			if currProductMatch == nil {
				continue
			}

			if highestConfidenceProductMatch == nil || (highestConfidenceProductMatch.ConfidenceScore < currProductMatch.ConfidenceScore) {
				highestConfidenceProductMatch = currProductMatch
			}
		}

		if highestConfidenceProductMatch == nil {
			productLogger.Info("Could not find a matching product on revzilla, continuing to the next product")
			return nil
		}

		err := j.updateProduct(product, highestConfidenceProductMatch, productLogger)
		if err != nil {
			return err
		}

		productLogger.Info("Successfully synced product")
		return nil
	})
}

func (j *SyncRevzillaHelmetsJob) updateProduct(product *entities.Product, productMatch *productMatch, productLogger *logrus.Entry) error {
	confidenceLogFields := logrus.Fields{
		"matchConfidence":             productMatch.ConfidenceScore,
		"matchingRevzillaProductName": productMatch.CJProduct.Name,
		"isDiscontinued":              productMatch.IsDiscontinued,
		"manufacturer":                product.Manufacturer,
		"modelToTry":                  product.Model,
	}

	if productMatch.IsDiscontinued {
		productLogger.WithFields(confidenceLogFields).Warning("This product is discontinued, updating the discontinued flag and continuing to the next product")
		product.IsDiscontinued = true
		err := j.ProductRepository.UpdateProduct(product)
		if err != nil {
			return err
		}
	} else {
		product.RevzillaBuyURL = productMatch.CJProduct.LinkCode.ClickURL
		product.RevzillaPriceCents = int(productMatch.CJProduct.GetPrice() * 100)
		product.IsDiscontinued = false
		product.UpdateHelmetCertificationsByDescription(productMatch.CJProduct.Description)
		product.UpdateSearchPrice()
		product.UpdateSafetyPercentage()
		product.Description = productMatch.CJProduct.Description
		productLogger.WithFields(confidenceLogFields).Info("Set new price and buy URL from RevZilla")
		err := j.ProductRepository.UpdateProduct(product)
		if err != nil {
			return err
		}
	}
	return nil
}

func (j *SyncRevzillaHelmetsJob) searchCJProducts(pooledClient *http.Client, manufacturer string, model string) (*entities.CJProductsResponseWrapper, error) {
	// Adding a delay to each request here because CJ has an absurdly low threshold for requests per minute
	time.Sleep(3 * time.Second)
	req, err := http.NewRequest(http.MethodPost, "https://ads.api.cj.com/query", strings.NewReader(fmt.Sprintf(`{
		shoppingProducts(companyId: "5023498", partnerIds: ["3318586"], keywords: ["%s", "%s", "helmet"], offset: 1, limit: 100, lowPrice: 200) {
		  resultList {
			linkCode(pid: "8505854") {
			  clickUrl
			},
			title,
			price {
			  amount
			},
			imageLink,
			productType,
			description
		  }
		}
	  }`, manufacturer, model)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", j.CJAPIKey))

	resp, err := pooledClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CJ API returned a status code of %d but expected 200", resp.StatusCode)
	}

	responseBodyBytes, _ := ioutil.ReadAll(resp.Body)
	cjResp := &entities.CJProductsResponseWrapper{}
	if err = json.Unmarshal(responseBodyBytes, cjResp); err != nil {
		return nil, err
	}
	return cjResp, nil
}

type productMatch struct {
	CJProduct       *entities.CJProduct
	ConfidenceScore float64
	IsDiscontinued  bool
}

func (j *SyncRevzillaHelmetsJob) getBestMatchForProduct(pooledClient *http.Client, product *entities.Product, modelToTry string, productLogger *logrus.Entry) (*productMatch, error) {
	cjResp, err := j.searchCJProducts(pooledClient, product.Manufacturer, modelToTry)
	if err != nil {
		return nil, err
	}

	var matchingRevzillaProductsSlice []entities.CJProduct
	confidenceMap := make(map[string]float64)
	expectedLowerProductName := strings.ToLower(fmt.Sprintf("%s %s", product.Manufacturer, modelToTry))

	golinq.From(cjResp.Data.ShoppingProducts.ResultList).WhereT(func(product entities.CJProduct) bool {
		return product.IsHelmet()
	}).OrderByDescendingT(func(product entities.CJProduct) interface{} {
		lowerProductName := strings.ToLower(product.Name)
		matchConfidence := smetrics.JaroWinkler(lowerProductName, expectedLowerProductName, boostThreshold, prefixSize)
		if _, exists := confidenceMap[lowerProductName]; !exists {
			confidenceMap[lowerProductName] = matchConfidence
		}

		return matchConfidence
	}).ToSlice(&matchingRevzillaProductsSlice)

	if len(matchingRevzillaProductsSlice) <= 0 {
		productLogger.Info("Could not find a price or buy URL from RevZilla because no results were returned")
		return nil, nil
	}

	bestMatchRevzillaProduct := &matchingRevzillaProductsSlice[0]
	bestMatchConfidence := confidenceMap[strings.ToLower(bestMatchRevzillaProduct.Name)]
	buyURLContents, err := httpHelpers.GetContentsAtURL(bestMatchRevzillaProduct.LinkCode.ClickURL)
	if err != nil {
		return nil, err
	}

	// If we don't have a product summary, it means we couldn't find the product
	isDiscontinued := !strings.Contains(strings.ToLower(buyURLContents), "product-show-summary")
	if bestMatchConfidence < bestMatchConfidenceThreshold {
		productLogger.WithFields(logrus.Fields{
			"matchConfidence":             bestMatchConfidence,
			"matchingRevzillaProductName": bestMatchRevzillaProduct.Name,
			"isDiscontinued":              isDiscontinued,
			"manufacturer":                product.Manufacturer,
			"modelToTry":                  modelToTry,
		}).Warning("Could not find a price or buy URL from RevZilla because the best match had a low confidence score")
		return nil, nil
	}

	return &productMatch{CJProduct: bestMatchRevzillaProduct, ConfidenceScore: bestMatchConfidence, IsDiscontinued: isDiscontinued}, nil
}

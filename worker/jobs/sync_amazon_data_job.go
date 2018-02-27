package jobs

import (
	"crashtested-backend/persistence/repositories"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/bakatz/go-amazon-product-advertising-api/amazon"
	"github.com/sirupsen/logrus"
)

// SyncAmazonDataJob sync Amazon price data for each product in the database.
type SyncAmazonDataJob struct {
	ProductRepository *repositories.ProductRepository
	AmazonClient      *amazon.Client
}

// Run invokes the job and returns an error if there were any errors encountered while processing the price data
func (j *SyncAmazonDataJob) Run() error {
	start := 0
	limit := 25
	currProducts, err := j.ProductRepository.GetAllPaged(start, limit)
	if err != nil {
		return err
	}

	for len(currProducts) > 0 {
		for _, product := range currProducts {
			productLogger := logrus.WithFields(
				logrus.Fields{
					"productUUID":  product.UUID,
					"manufacturer": product.Manufacturer,
					"model":        product.Model,
				})
			itemSearchRequest := j.AmazonClient.ItemSearch(amazon.ItemSearchParameters{
				SearchIndex:  amazon.SearchIndexAll,
				Keywords:     fmt.Sprintf("%s %s helmet -shield", product.Manufacturer, product.Model),
				MinimumPrice: 20000, // This is the same as $200.00
			})

			var searchResp *amazon.ItemSearchResponse
			var doSearchErr error
			url := ""
			priceInUsdMultiple := 0
			hasNewDOTCertification := false
			hasNewECECertification := false

			time.Sleep(1 * time.Second)
			if searchResp, doSearchErr = itemSearchRequest.Do(); doSearchErr != nil {
				productLogger.WithField("error", doSearchErr).Warn("Encountered an error while searching for the product, continuing to the next product")
				continue
			}

			if searchRespErr := searchResp.Error(); searchRespErr != nil {
				productLogger.WithField("error", searchRespErr).Warn("Encountered an error while searching for the product, continuing to the next product")
				continue
			}

			if searchResp.Items.TotalResults > 0 {
				bestResult := searchResp.Items.Item[0]
				itemLookupRequest := j.AmazonClient.ItemLookup(amazon.ItemLookupParameters{
					IDType:         amazon.IDTypeASIN,
					ResponseGroups: []amazon.ItemLookupResponseGroup{amazon.ItemLookupResponseGroupOffers, amazon.ItemLookupResponseGroupEditorialReview},
					ItemIDs:        []string{bestResult.ASIN},
				})

				time.Sleep(1 * time.Second)
				lookupResp, lookupErr := itemLookupRequest.Do()
				if lookupErr != nil {
					productLogger.WithField("error", lookupErr).Error("Encountered an error while getting item details, continuing to the next product")
					continue
				}

				if lookupRespErr := lookupResp.Error(); lookupRespErr != nil {
					productLogger.WithField("error", lookupRespErr).Error("Encountered an error while getting item details, continuing to the next product")
					continue
				}

				if len(lookupResp.Items.Item) > 0 {
					productDescription := ""
					if len(lookupResp.Items.Item[0].Offers.Offer) > 0 {
						firstItem := lookupResp.Items.Item[0]
						lowestNewPriceAmount, _ := strconv.Atoi(firstItem.OfferSummary.LowestNewPrice.Amount)
						lowestUsedPriceAmount, _ := strconv.Atoi(firstItem.OfferSummary.LowestUsedPrice.Amount)

						if lowestNewPriceAmount > 0 && lowestUsedPriceAmount > 0 {
							priceInUsdMultiple = int(math.Min(float64(lowestUsedPriceAmount), float64(lowestNewPriceAmount)))
						} else {
							priceInUsdMultiple = int(math.Max(float64(lowestUsedPriceAmount), float64(lowestNewPriceAmount)))
						}
						url = bestResult.DetailPageURL
					}

					if resp, err := http.Get(bestResult.DetailPageURL); err == nil {
						if doc, err := goquery.NewDocumentFromResponse(resp); err == nil {
							detailsText := doc.Find("#prodDetails").Text()
							productDescription += detailsText
						}
					}

					reviewContent := lookupResp.Items.Item[0].EditorialReviews.EditorialReview.Content
					productDescription += reviewContent

					lowerDescription := strings.ToLower(productDescription)
					containsDOT := strings.Contains(productDescription, "DOT") || strings.Contains(productDescription, "D.O.T")
					containsECE := strings.Contains(productDescription, "ECE") || strings.Contains(productDescription, "22/05") || strings.Contains(productDescription, "22.05")
					containsSNELL := strings.Contains(lowerDescription, "snell") || strings.Contains(lowerDescription, "m2010") || strings.Contains(lowerDescription, "m2015")

					if !product.Certifications.DOT && (containsDOT || containsSNELL) {
						hasNewDOTCertification = true
					}

					if !product.Certifications.ECE && containsECE {
						hasNewECECertification = true
					}
				}
			}

			if url == "" && priceInUsdMultiple <= 0 && !hasNewDOTCertification && !hasNewECECertification {
				productLogger.Warning("Got Amazon data for the product, but it was empty. Skipping.")
			} else {
				commitLogger := productLogger.WithFields(logrus.Fields{
					"buyUrl":             url,
					"priceInUsdMultiple": priceInUsdMultiple,
					"hasNewDOTRating":    hasNewDOTCertification,
				})

				if priceInUsdMultiple > 0 {
					commitLogger.Info("Saving new price")
					(&product).PriceInUSDMultiple = priceInUsdMultiple
				}

				if url != "" {
					commitLogger.Info("Saving new url")
					(&product).BuyURL = url
				}

				if hasNewDOTCertification {
					commitLogger.Info("Saving new DOT certification")
					(&product).Certifications.DOT = true
				}

				if hasNewECECertification {
					commitLogger.Info("Saving new ECE certification")
					(&product).Certifications.ECE = true
				}

				err = j.ProductRepository.UpdateProduct(&product)
				if err != nil {
					return err
				}
			}
		}

		start += limit
		currProducts, err = j.ProductRepository.GetAllPaged(start, limit)
		if err != nil {
			return err
		}
	}
	return nil
}

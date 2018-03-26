package jobs

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/worker/jobs/helpers"
	"fmt"
	"math"
	"net/http"
	"strconv"
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
	return helpers.ForEachProduct(j.ProductRepository, func(product *entities.ProductDocument, productLogger *logrus.Entry) error {
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
			return nil
		}

		if searchRespErr := searchResp.Error(); searchRespErr != nil {
			productLogger.WithField("error", searchRespErr).Warn("Encountered an error while searching for the product, continuing to the next product")
			return nil
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
				return nil
			}

			if lookupRespErr := lookupResp.Error(); lookupRespErr != nil {
				productLogger.WithField("error", lookupRespErr).Error("Encountered an error while getting item details, continuing to the next product")
				return nil
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

						featuresText := doc.Find("#feature-bullets").Text()
						productDescription += featuresText
					}
				}

				reviewContent := lookupResp.Items.Item[0].EditorialReviews.EditorialReview.Content
				productDescription += reviewContent
				hasNewDOTCertification, hasNewECECertification = product.UpdateCertificationsByDescription(productDescription)
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
				product.AmazonPriceInUSDMultiple = priceInUsdMultiple
			}

			if url != "" {
				commitLogger.Info("Saving new url")
				product.AmazonBuyURL = url
			}

			if hasNewDOTCertification {
				commitLogger.Info("Saving new DOT certification")
				product.Certifications.DOT = true
			}

			if hasNewECECertification {
				commitLogger.Info("Saving new ECE certification")
				product.Certifications.ECE = true
			}

			err := j.ProductRepository.UpdateProduct(product)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

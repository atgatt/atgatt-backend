package jobs

import (
	"crashtested-backend/persistence/repositories"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/ngs/go-amazon-product-advertising-api/amazon"
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
	currProducts, err := j.ProductRepository.GetAllWithoutPricePaged(start, limit)
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
				SearchIndex:  amazon.SearchIndexAutomotive,
				Keywords:     fmt.Sprintf("%s -shield", product.Model),
				Manufacturer: product.Manufacturer,
				MinimumPrice: 10000, // This is the same as $100.00
			})

			var searchResp *amazon.ItemSearchResponse
			var doSearchErr error
			url := ""
			priceInUsdMultiple := 0

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
					ResponseGroups: []amazon.ItemLookupResponseGroup{amazon.ItemLookupResponseGroupOffers},
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

				if len(lookupResp.Items.Item) > 0 && len(lookupResp.Items.Item[0].Offers.Offer) > 0 {
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
			}

			if url == "" || priceInUsdMultiple == 0 {
				productLogger.Error("Got Amazon data for the product, but it was empty. Skipping.")
			} else {
				productLogger.Infof("Successfully got Amazon data - URL: %s, Price: %d", url, priceInUsdMultiple)

				(&product).PriceInUSDMultiple = priceInUsdMultiple
				(&product).BuyURL = url

				err := j.ProductRepository.UpdateProduct(&product)
				if err != nil {
					return err
				}
			}
		}

		start += limit
		currProducts, err = j.ProductRepository.GetAllWithoutPricePaged(start, limit)
		if err != nil {
			return err
		}
	}
	return nil
}

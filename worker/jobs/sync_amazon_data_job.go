package jobs

import (
	"crashtested-backend/persistence/repositories"
	"fmt"

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
	currProducts, err := j.ProductRepository.GetAllPaged(start, limit)
	if err != nil {
		return err
	}

	for len(currProducts) > 0 {
		for _, product := range currProducts {
			itemSearchRequest := j.AmazonClient.ItemSearch(amazon.ItemSearchParameters{
				SearchIndex:  amazon.SearchIndexAll,
				Keywords:     fmt.Sprintf("%s %s", product.Manufacturer, product.Model),
				MinimumPrice: 50,
			})

			var resp *amazon.ItemSearchResponse
			var doErr error
			if resp, doErr = itemSearchRequest.Do(); doErr != nil {
				return doErr
			}

			if respErr := resp.Error(); respErr != nil {
				return respErr
			}

			if resp.Items.TotalResults > 0 {
				bestResult := resp.Items.Item[0]
				itemLookupRequest := j.AmazonClient.ItemLookup(amazon.ItemLookupParameters{
					IDType:         amazon.IDTypeASIN,
					ResponseGroups: []amazon.ItemLookupResponseGroup{amazon.ItemLookupResponseGroupOffers},
					ItemIDs:        []string{bestResult.ASIN},
				})

				resp, err := itemLookupRequest.Do()
				if err != nil {
					return err
				}

				logrus.Info(resp)
			}
			logrus.Info(resp)
		}

		start += limit
		currProducts, err = j.ProductRepository.GetAllPaged(start, limit)
		if err != nil {
			return err
		}
	}
	return nil
}

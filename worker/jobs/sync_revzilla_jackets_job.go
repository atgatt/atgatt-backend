package jobs

import (
	appEntities "crashtested-backend/application/entities"
	s3Helpers "crashtested-backend/common/s3"
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/remeh/sizedwaitgroup"
	"github.com/sirupsen/logrus"
)

const revzillaBaseURL string = "https://www.revzilla.com"
const minRevzillaProducts int = 1000

// SyncRevzillaJacketsJob scrapes all of RevZilla's helmet data
type SyncRevzillaJacketsJob struct {
	ProductRepository *repositories.ProductRepository
	S3Uploader        s3manageriface.UploaderAPI
	S3Bucket          string
}

// Run executes the job
func (j *SyncRevzillaJacketsJob) Run() error {
	pooledClient := cleanhttp.DefaultPooledClient()
	resp, err := pooledClient.Get("https://www.revzilla.com/motorcycle-jackets-vests?page=1&sort=featured&limit=10000&rating=-1&price=&price_min=3&price_max=1700&is_new=false&is_sale=false&is_made_in_usa=false&has_video=false&is_holiday=false&is_blemished=false&view_all=true")
	if err != nil {
		return err
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return err
	}

	revzillaProductsToScrape := getRevzillaProductsToScrape(doc)
	if len(revzillaProductsToScrape) < minRevzillaProducts {
		return errors.New("Not enough URLs found, check RevZilla's HTML for changes")
	}

	sizedWg := sizedwaitgroup.New(4)
	for _, revzillaProduct := range revzillaProductsToScrape {
		sizedWg.Add()
		go func(p *appEntities.RevzillaProduct) {
			defer sizedWg.Done()
			productLogger := logrus.WithFields(logrus.Fields{
				"externalID": p.ID,
				"name":       p.Name,
			})
			productLogger.Info("Starting to get a description for a product")
			p.DescriptionParts, err = getDescriptionPartsForProduct(pooledClient, p)
			if err != nil {
				productLogger.WithError(err).Error("Failed to get a description for a product")
			}

			if len(p.DescriptionParts) == 0 {
				productLogger.Warning("Could not find a description for a product, continuing to the next one")
			} else {
				productLogger.Info("Finished getting a description for a product")
			}

			existingProduct, err := j.ProductRepository.GetByExternalID(p.ID)
			if err != nil {
				productLogger.WithError(err).Error("Failed to get a product from the database by external ID")
			}

			if existingProduct != nil {
				existingProduct.RevzillaPriceCents = p.GetPriceCents()
				existingProduct.RevzillaBuyURL = p.URL
				existingProduct.IsDiscontinued = len(p.DescriptionParts) <= 0
				existingProduct.UpdateJacketCertificationsByDescriptionParts(p.DescriptionParts)
				existingProduct.UpdateSearchPrice()
				existingProduct.UpdateSafetyPercentage()

				err = j.ProductRepository.UpdateProduct(existingProduct)
			} else {
				productToPersist := &entities.Product{
					OriginalImageURL:   p.ImageURL,
					Manufacturer:       p.Brand,
					Model:              p.GetModel(),
					RevzillaPriceCents: p.GetPriceCents(),
					Type:               "jacket",
					UUID:               uuid.New(),
					RevzillaBuyURL:     fmt.Sprintf("http://www.anrdoezrs.net/links/8505854/type/dlg/%s", p.URL),
					ExternalID:         p.ID,
				}

				productToPersist.UpdateJacketCertificationsByDescriptionParts(p.DescriptionParts)

				if productToPersist.OriginalImageURL != "" {
					key, err := s3Helpers.CopyImageToS3FromURL(productLogger, j.S3Uploader, p.ImageURL, j.S3Bucket)
					if err != nil {
						productLogger.WithError(err).Warning("Failed to upload an image to S3, continuing")
					}
					productToPersist.ImageKey = key
				} else {
					productLogger.Warning("Skipping uploading image to S3 because the URL is empty, continuing")
				}

				err = j.ProductRepository.CreateProduct(productToPersist)
			}

			if err != nil {
				productLogger.WithError(err).Error("Failed to upsert a product into the database")
			}

		}(revzillaProduct)
	}

	sizedWg.Wait()
	return nil
}

func getDescriptionPartsForProduct(pooledClient *http.Client, revzillaProduct *appEntities.RevzillaProduct) ([]string, error) {
	resp, err := pooledClient.Get(revzillaProduct.URL)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	parts := []string{}
	detailsNode := doc.Find(".product-details__details")

	aggregateText := func(index int, item *goquery.Selection) {
		parts = append(parts, item.Text())
	}

	detailsNode.Find("p").Each(aggregateText)
	detailsNode.Find("li").Each(aggregateText)
	return parts, nil
}

func getMetaValue(item *goquery.Selection, key string) string {
	val, _ := item.Find(fmt.Sprintf("meta[itemprop='%s']", key)).Attr("content")
	return val
}

func getRevzillaProductsToScrape(doc *goquery.Document) []*appEntities.RevzillaProduct {
	productNodesCollection := doc.Find("*[data-product-id]")

	revzillaProducts := []*appEntities.RevzillaProduct{}
	productNodesCollection.Map(func(i int, item *goquery.Selection) string {
		urlSuffix, exists := item.Attr("href")
		externalID, _ := item.Attr("data-product-id")
		if !exists {
			logrus.WithField("externalID", externalID).Error("Could not find href for product")
			return ""
		}

		price := getMetaValue(item, "price")
		priceCurrency := getMetaValue(item, "priceCurrency")
		url := fmt.Sprintf("%s%s", revzillaBaseURL, urlSuffix)
		imageURL := getMetaValue(item, "image")
		name := getMetaValue(item, "name")
		brand := getMetaValue(item, "brand")

		revzillaProducts = append(revzillaProducts, &appEntities.RevzillaProduct{
			ID:            externalID,
			Brand:         brand,
			ImageURL:      imageURL,
			Price:         price,
			PriceCurrency: priceCurrency,
			Name:          name,
			URL:           url,
		})

		return ""
	})

	return revzillaProducts
}

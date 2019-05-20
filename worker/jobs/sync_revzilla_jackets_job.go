package jobs

import (
	"strings"
	"crashtested-backend/common/text"
	"crashtested-backend/application/clients"
	appEntities "crashtested-backend/application/entities"
	s3Helpers "crashtested-backend/common/s3"
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/remeh/sizedwaitgroup"
	"github.com/sirupsen/logrus"
)

const revzillaBaseURL string = "https://www.revzilla.com"
const minRevzillaProducts int = 1000

// SyncRevzillaJacketsJob scrapes all of RevZilla's helmet data
type SyncRevzillaJacketsJob struct {
	ProductRepository      *repositories.ProductRepository
	RevzillaClient         clients.RevzillaClient
	S3Uploader             s3manageriface.UploaderAPI
	S3Bucket               string
	EnableMinProductsCheck bool
}

func getRevzillaURL(url string) string {
	return fmt.Sprintf("http://www.anrdoezrs.net/links/8505854/type/dlg/%s", url)
}

// Run executes the job
func (j *SyncRevzillaJacketsJob) Run() error {
	doc, err := j.RevzillaClient.GetAllJacketOverviewsHTML()
	if err != nil {
		return err
	}

	revzillaProductsToScrape := getRevzillaProductsToScrape(doc)
	if j.EnableMinProductsCheck && len(revzillaProductsToScrape) < minRevzillaProducts {
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
			p.DescriptionParts, err = j.getDescriptionPartsForProduct(p, productLogger)
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
				existingProduct.RevzillaBuyURL = getRevzillaURL(p.URL)
				existingProduct.IsDiscontinued = len(p.DescriptionParts) <= 0
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
					RevzillaBuyURL:     getRevzillaURL(p.URL),
					ExternalID:         p.ID,
				}

				productToPersist.UpdateJacketCertificationsByDescriptionParts(p.DescriptionParts)
				productToPersist.UpdateJacketSubtypeByDescriptionParts(p.DescriptionParts)

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

func (j *SyncRevzillaJacketsJob) getDescriptionPartsForProduct(revzillaProduct *appEntities.RevzillaProduct, productLogger *logrus.Entry) ([]string, error) {
	doc, err := j.RevzillaClient.GetDescriptionPartsHTMLByURL(revzillaProduct.URL)
	if err != nil {
		return nil, err
	}

	parts := []string{}
	detailsNode := doc.Find(".product-details__details")

	detailsNode.Find("p").Each(func(index int, item *goquery.Selection) {
		rawSummary := item.Text()

		// the ML model isn't trained for motorcycle terms, so sometimes it screws up like on Category II (thinks its another sentence)
		sentences, err := text.GetSentencesFromString(strings.Replace(rawSummary, "Cat. II", "Cat II", -1))
		if err != nil {
			productLogger.WithError(err).Error("Got an error while splitting the summary, using the raw text as is")
			parts = append(parts, rawSummary)
			return
		}

		if len(sentences) == 0 {
			productLogger.Warn("Got 0 sentences back while splitting the summary, using the raw text as is")
			parts = append(parts, rawSummary)
			return
		}
		
		parts = append(parts, sentences...)
	})

	detailsNode.Find("li").Each(func(index int, item *goquery.Selection) {
		parts = append(parts, item.Text())
	})
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

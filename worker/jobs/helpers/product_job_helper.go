package helpers

import (
	"crashtested-backend/application/clients"
	appEntities "crashtested-backend/application/entities"
	s3Helpers "crashtested-backend/common/s3"
	"crashtested-backend/common/text"
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/google/uuid"
	"github.com/remeh/sizedwaitgroup"

	"github.com/PuerkitoBio/goquery"

	"github.com/sirupsen/logrus"
)

// RevzillaBaseURL refers to the root URL of revzilla.com
const RevzillaBaseURL string = "https://www.revzilla.com"

// MinRevzillaProducts represents the minimum number of products that are expected to be returned before an error is thrown (prevents against importing bad HTML when revzilla changes their webpage)
const MinRevzillaProducts int = 500

// ForEachProduct iterates over all the products in the database and runs the current action on the given product
func ForEachProduct(productRepository *repositories.ProductRepository, action func(product *entities.Product, productLogger *logrus.Entry) error) error {
	start := 0
	limit := 25
	currProducts, err := productRepository.GetAllPaged(start, limit)
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
			err := action(&product, productLogger)
			if err != nil {
				return err
			}
		}

		start += limit
		currProducts, err = productRepository.GetAllPaged(start, limit)
		if err != nil {
			return err
		}
	}
	return nil
}

func getMetaValue(item *goquery.Selection, key string) string {
	val, _ := item.Find(fmt.Sprintf("meta[itemprop='%s']", key)).Attr("content")
	return val
}

// GetRevzillaProductsToScrape fetches all revzilla products on a given search result page along with their URLs for further inspection
func GetRevzillaProductsToScrape(doc *goquery.Document) []*appEntities.RevzillaProduct {
	if doc == nil {
		return []*appEntities.RevzillaProduct{}
	}

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
		url := fmt.Sprintf("%s%s", RevzillaBaseURL, urlSuffix)
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

// GetRevzillaAffiliateURL returns an affiliate marketing link to monetize a url on revzilla.com
func GetRevzillaAffiliateURL(url string) string {
	return fmt.Sprintf("http://www.anrdoezrs.net/links/8505854/type/dlg/%s", url)
}

// GetDescriptionPartsForRevzillaProduct returns all of the description text as an array for a given Revzilla product
func GetDescriptionPartsForRevzillaProduct(revzillaProduct *appEntities.RevzillaProduct, productLogger *logrus.Entry, revzillaClient clients.RevzillaClient) ([]string, error) {
	if revzillaProduct == nil {
		return nil, errors.New("revzillaProduct must be defined")
	}

	if productLogger == nil {
		return nil, errors.New("productLogger must be defined")
	}

	if revzillaClient == nil {
		return nil, errors.New("revzillaClient must be defined")
	}

	doc, err := revzillaClient.GetDescriptionPartsHTMLByURL(revzillaProduct.URL)
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

// RunRevzillaImport is a generic function that imports (creates/updates) products found on revzilla.com of the given product type given a doc (goquery HTML doc representing the markup for all of the products in a given category)
func RunRevzillaImport(
	productURLPrefix string,
	productType string,
	revzillaClient clients.RevzillaClient,
	productRepository *repositories.ProductRepository,
	s3Uploader s3manageriface.UploaderAPI,
	s3Bucket string,
	enableMinProductsCheck bool,
	updateCertificationsFunc func(productToPersist *entities.Product, revzillaProduct *appEntities.RevzillaProduct),
) error {
	if productURLPrefix == "" {
		return errors.New("productURLPrefix cannot be empty")
	}

	if productType == "" {
		return errors.New("productType cannot be empty")
	}

	if revzillaClient == nil {
		return errors.New("revzillaClient cannot be nil")
	}

	if productRepository == nil {
		return errors.New("productRepository cannot be nil")
	}

	if s3Uploader == nil {
		return errors.New("s3Uploader cannot be nil")
	}

	if s3Bucket == "" {
		return errors.New("s3Bucket cannot be empty")
	}

	doc, err := revzillaClient.GetAllProductOverviewsHTML(productURLPrefix)
	if err != nil {
		return err
	}

	revzillaProductsToScrape := GetRevzillaProductsToScrape(doc)
	if enableMinProductsCheck && len(revzillaProductsToScrape) < MinRevzillaProducts {
		return errors.New("Not enough URLs found, check RevZilla's HTML for changes")
	}

	sizedWg := sizedwaitgroup.New(4)
	for _, revzillaProduct := range revzillaProductsToScrape {
		sizedWg.Add()
		go func(revzillaProduct *appEntities.RevzillaProduct) {
			defer sizedWg.Done()
			productLogger := logrus.WithFields(logrus.Fields{
				"externalID": revzillaProduct.ID,
				"name":       revzillaProduct.Name,
			})
			productLogger.Info("Starting to get a description for a product")
			revzillaProduct.DescriptionParts, err = GetDescriptionPartsForRevzillaProduct(revzillaProduct, productLogger, revzillaClient)
			if err != nil {
				productLogger.WithError(err).Error("Failed to get a description for a product")
			}

			if len(revzillaProduct.DescriptionParts) == 0 {
				productLogger.Warning("Could not find a description for a product, continuing to the next one")
			} else {
				productLogger.Info("Finished getting a description for a product")
			}

			existingProduct, err := productRepository.GetByExternalID(revzillaProduct.ID)
			if err != nil {
				productLogger.WithError(err).Error(fmt.Sprintf("Could not find a product with externalID: %v", revzillaProduct.ID))
			}

			if existingProduct != nil {
				existingProduct.RevzillaPriceCents = revzillaProduct.GetPriceCents()
				existingProduct.RevzillaBuyURL = GetRevzillaAffiliateURL(revzillaProduct.URL)
				existingProduct.IsDiscontinued = len(revzillaProduct.DescriptionParts) <= 0
				existingProduct.UpdateSearchPrice()
				existingProduct.UpdateSafetyPercentage()

				err = productRepository.UpdateProduct(existingProduct)
			} else {
				productToPersist := &entities.Product{
					OriginalImageURL:   revzillaProduct.ImageURL,
					Description:        strings.Join(revzillaProduct.DescriptionParts, "<br \\>"),
					Manufacturer:       revzillaProduct.Brand,
					Model:              revzillaProduct.GetModel(),
					RevzillaPriceCents: revzillaProduct.GetPriceCents(),
					Type:               productType,
					UUID:               uuid.New(),
					RevzillaBuyURL:     GetRevzillaAffiliateURL(revzillaProduct.URL),
					ExternalID:         revzillaProduct.ID,
				}

				updateCertificationsFunc(productToPersist, revzillaProduct)

				if productToPersist.OriginalImageURL != "" {
					key, err := s3Helpers.CopyImageToS3FromURL(productLogger, s3Uploader, revzillaProduct.ImageURL, s3Bucket)
					if err != nil {
						productLogger.WithError(err).Warning("Failed to upload an image to S3, continuing")
					}
					productToPersist.ImageKey = key
				} else {
					productLogger.Warning("Skipping uploading image to S3 because the URL is empty, continuing")
				}

				productToPersist.UpdateSafetyPercentage()
				err = productRepository.CreateProduct(productToPersist)
			}

			if err != nil {
				productLogger.WithError(err).Error("Failed to upsert a product into the database")
			}

		}(revzillaProduct)
	}

	sizedWg.Wait()
	return nil
}

package jobs

import (
	"atgatt-backend/application/parsers"
	s3Helpers "atgatt-backend/common/s3"
	"atgatt-backend/persistence/entities"
	"atgatt-backend/persistence/repositories"
	"fmt"
	"math"
	"strings"

	golinq "github.com/ahmetb/go-linq"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/xrash/smetrics"
)

// ImportHelmetsJob imports all helmet data from SHARP and SNELL into the database. It tries to normalize helmet models and manufacturers while doing this in order to have a clean data set. TODO: Refactor to not upsert if the product already exists, write tests
type ImportHelmetsJob struct {
	ProductRepository      *repositories.ProductRepository
	SNELLHelmetParser      *parsers.SNELLHelmetParser
	SHARPHelmetParser      *parsers.SHARPHelmetParser
	ManufacturerRepository *repositories.ManufacturerRepository
	S3Uploader             s3manageriface.UploaderAPI
	S3Bucket               string
}

const helmetType string = "helmet"

// Run invokes the job and returns an error if any errors occurred while processing the helmet data.
func (j *ImportHelmetsJob) Run() error {
	sharpProducts := []*entities.Product{}
	snellOnlyProducts := []*entities.Product{}

	manufacturers, err := j.ManufacturerRepository.GetAll()
	if err != nil {
		return err
	}

	manufacturerAliases, err := j.ProductRepository.GetAllManufacturerAliases()
	if err != nil {
		return err
	}

	manufacturerAliasesMap := make(map[string]string)
	for _, manufacturerAlias := range manufacturerAliases {
		manufacturerAliasesMap[manufacturerAlias.Manufacturer] = manufacturerAlias.ManufacturerAlias
	}

	allModelAliases, err := j.ProductRepository.GetAllModelAliases()
	if err != nil {
		return err
	}

	// NOTE: This call blocks for about a minute on average as we need to fetch 400+ HTML files and scrape them for data.
	sharpHelmets, err := j.SHARPHelmetParser.GetAll()
	if err != nil {
		return err
	}

	matchedAllProducts := true
	for _, sharpHelmet := range sharpHelmets {
		cleanedManufacturer := findCleanedManufacturer(sharpHelmet.Manufacturer, manufacturers, manufacturerAliasesMap)
		matchingModelAliases := findAliasesForModel(allModelAliases, cleanedManufacturer, sharpHelmet.Model)
		product := &entities.Product{
			OriginalImageURL: sharpHelmet.ImageURL,
			LatchPercentage:  sharpHelmet.LatchPercentage,
			Manufacturer:     cleanedManufacturer,
			Materials:        sharpHelmet.Materials,
			Model:            sharpHelmet.Model,
			ModelAliases:     matchingModelAliases,
			MSRPCents:        sharpHelmet.ApproximateMSRPCents,
			RetentionSystem:  sharpHelmet.RetentionSystem,
			Sizes:            sharpHelmet.Sizes,
			Subtype:          sharpHelmet.Subtype,
			Type:             helmetType,
			UUID:             uuid.New(),
			WeightInLbs:      sharpHelmet.WeightInLbs,
		}

		if len(matchingModelAliases) > 0 {
			logrus.WithFields(logrus.Fields{
				"model":        product.Model,
				"modelAliases": matchingModelAliases,
			}).Info("Found some aliases for the given model")
		}

		product.HelmetCertifications.SHARP = sharpHelmet.Certifications
		product.HelmetCertifications.ECE = sharpHelmet.IsECECertified
		sharpProducts = append(sharpProducts, product)
	}

	snellHelmets, err := j.SNELLHelmetParser.GetAllByCertification("M2015")
	if err != nil {
		return err
	}

	for _, snellHelmet := range snellHelmets {
		cleanedManufacturer := findCleanedManufacturer(snellHelmet.Manufacturer, manufacturers, manufacturerAliasesMap)
		matchingSHARPProduct := findMatchingSHARPProduct(cleanedManufacturer, snellHelmet.Model, sharpProducts)
		if matchedAllProducts && matchingSHARPProduct == nil {
			matchedAllProducts = false
		}

		if matchingSHARPProduct != nil {
			logrus.WithFields(logrus.Fields{
				"manufacturer": matchingSHARPProduct.Manufacturer,
				"model":        matchingSHARPProduct.Model,
			}).Info("Updated a SHARP helmet to have SNELL and DOT ratings")
			matchingSHARPProduct.HelmetCertifications.SNELL = true
			matchingSHARPProduct.HelmetCertifications.DOT = true
		} else {
			logrus.WithFields(logrus.Fields{
				"manufacturer": cleanedManufacturer,
				"model":        snellHelmet.Model,
			}).Info("Could not find a matching SHARP helmet, so initializing a helmet with only SNELL and DOT ratings")

			sizes := strings.Split(snellHelmet.Size, ",")
			snellOnlyProduct := &entities.Product{
				Manufacturer: cleanedManufacturer,
				Model:        snellHelmet.Model,
				UUID:         uuid.New(),
				Type:         helmetType,
				Subtype:      snellHelmet.FaceConfig,
				Sizes:        sizes,
			}
			snellOnlyProduct.HelmetCertifications.SNELL = true
			snellOnlyProduct.HelmetCertifications.DOT = true
			snellOnlyProducts = append(snellOnlyProducts, snellOnlyProduct)
		}
	}

	combinedProductsList := append(sharpProducts, snellOnlyProducts...)
	for _, product := range combinedProductsList {
		productLogger := logrus.WithFields(logrus.Fields{
			"manufacturer": product.Manufacturer,
			"model":        product.Model,
		})
		productLogger.Info("Starting to upsert the product into the database")
		validator := &entities.ProductValidator{Product: product}
		validationErr := validator.Validate()
		if validationErr != nil {
			productLogger.WithField("validationError", validationErr).Warning("Validation failed, continuing to the next helmet")
			continue
		}

		existingProduct, err := j.ProductRepository.GetByModel(product.Manufacturer, product.Model, "helmet")
		if err == repositories.ErrEntityNotFound {
			productLogger.WithError(err).Error(fmt.Sprintf("Could not find a product with manufacturer: %s, model: %s, type: %s",
				product.Manufacturer,
				product.Model,
				"helmet"),
			)
		} else if err != nil {
			return err
		}

		if product.OriginalImageURL != "" && existingProduct == nil {
			key, err := s3Helpers.CopyImageToS3FromURL(productLogger, j.S3Uploader, product.OriginalImageURL, j.S3Bucket)
			if err != nil {
				productLogger.Warning("Could not upload image to S3, saving the product to the DB anyway")
			} else {
				product.ImageKey = key
			}
		} else {
			productLogger.Info("Not uploading anything to S3, saving the product to the DB anyway")
		}

		product.UpdateSafetyPercentage()

		if existingProduct == nil {
			err := j.ProductRepository.CreateProduct(product)
			if err != nil {
				return err
			}
		} else {
			productLogger.WithField("existingUUID", existingProduct.UUID).Warning("Product already exists, skipping it")
		}

		productLogger.Info("Successfully finished upserting the product")
	}

	return nil
}

func findAliasesForModel(allAliases []*entities.ProductModelAlias, manufacturer string, model string) []*entities.ProductModelAlias {
	matchingAliases := []*entities.ProductModelAlias{}
	for _, alias := range allAliases {
		if strings.EqualFold(alias.Manufacturer, manufacturer) && strings.EqualFold(alias.Model, model) {
			matchingAliases = append(matchingAliases, alias)
		}
	}
	return matchingAliases
}

const boostThreshold float64 = 0.7
const prefixSize int = 4

func findMatchingSHARPProduct(cleanedSNELLManufacturer string, rawSNELLModel string, sharpProducts []*entities.Product) *entities.Product {
	possibleSHARPHelmets := []*entities.Product{}
	for _, sharpHelmet := range sharpProducts {
		if sharpHelmet.Manufacturer == cleanedSNELLManufacturer {
			possibleSHARPHelmets = append(possibleSHARPHelmets, sharpHelmet)
		}
	}

	if len(possibleSHARPHelmets) <= 0 {
		logrus.WithFields(logrus.Fields{
			"manufacturer": cleanedSNELLManufacturer,
			"model":        rawSNELLModel,
		}).Warn("No helmets found for the given manufacturer")
		return nil
	}

	confidenceMap := make(map[string]float64)
	orderedSHARPHelmets := []*entities.Product{}
	lowerSNELLModel := strings.ToLower(rawSNELLModel)
	golinq.From(possibleSHARPHelmets).OrderByDescendingT(func(helmet *entities.Product) interface{} {
		var maxConfidence float64
		for _, alias := range helmet.ModelAliases {
			lowerAlias := strings.ToLower(alias.ModelAlias)
			aliasConfidence := smetrics.JaroWinkler(lowerAlias, lowerSNELLModel, boostThreshold, prefixSize)
			maxConfidence = math.Max(maxConfidence, aliasConfidence)
		}

		modelConfidence := smetrics.JaroWinkler(strings.ToLower(helmet.Model), lowerSNELLModel, boostThreshold, prefixSize)
		maxConfidence = math.Max(maxConfidence, modelConfidence)
		confidenceMap[helmet.Model] = maxConfidence
		return maxConfidence
	}).ToSlice(&orderedSHARPHelmets)

	mostLikelySHARPHelmet := orderedSHARPHelmets[0]
	confidence := confidenceMap[mostLikelySHARPHelmet.Model]
	logEntry := logrus.WithFields(logrus.Fields{
		"rawSNELLModel":               rawSNELLModel,
		"mostLikelySHARPModel":        mostLikelySHARPHelmet.Model,
		"mostLikelySHARPModelAliases": mostLikelySHARPHelmet.ModelAliases,
		"confidence":                  confidence,
	})

	// if we're 90% confident that the model matches, use the value
	if confidence >= 0.9 {
		logEntry.Info("High confidence: found matching SHARP model using Jaro-Winkler algorithm")
		return mostLikelySHARPHelmet
	}

	logEntry.Warn("Low confidence: SHARP match found, but confidence too low. Ignoring.")
	return nil
}

func findCleanedManufacturer(rawManufacturer string, cleanedManufacturers []string, manufacturerAliasesMap map[string]string) string {
	mostLikelyManufacturers := make([]string, len(cleanedManufacturers))
	confidenceMap := make(map[string]float64)
	lowercaseRawManufacturer := strings.ToLower(rawManufacturer)

	golinq.From(cleanedManufacturers).OrderByDescendingT(func(cleanedManufacturer string) interface{} {
		lowercaseCleanedManufacturer := strings.ToLower(cleanedManufacturer)
		matchConfidence := smetrics.JaroWinkler(lowercaseRawManufacturer, lowercaseCleanedManufacturer, boostThreshold, prefixSize)
		if _, exists := confidenceMap[cleanedManufacturer]; !exists {
			confidenceMap[cleanedManufacturer] = matchConfidence
		}
		return matchConfidence
	}).ToSlice(&mostLikelyManufacturers)

	mostLikelyManufacturer := mostLikelyManufacturers[0]
	confidence := confidenceMap[mostLikelyManufacturer]
	logEntry := logrus.WithFields(logrus.Fields{
		"rawManufacturer":     rawManufacturer,
		"cleanedManufacturer": mostLikelyManufacturer,
		"confidence":          confidence,
	})

	manufacturerToReturn := ""

	// if we're 70% confident that the manufacturer matches, use the cleaned value
	if confidence >= 0.7 {
		logEntry.Info("High confidence: replaced raw manufacturer with cleaned manufacturer using Jaro-Winkler algorithm")
		manufacturerToReturn = mostLikelyManufacturer
	} else {
		foundCleanedManufacturer := false
		// Otherwise, do a stupider contains search to try to clean up the manufacturer
		for _, cleanedManufacturer := range cleanedManufacturers {
			lowercaseCleanedManufacturer := strings.ToLower(cleanedManufacturer)
			lowercaseRawManufacturer := strings.ToLower(rawManufacturer)

			if strings.HasPrefix(lowercaseRawManufacturer, lowercaseCleanedManufacturer) || strings.Contains(lowercaseRawManufacturer, fmt.Sprintf(" %s", lowercaseCleanedManufacturer)) {
				logrus.WithFields(logrus.Fields{
					"rawManufacturer":     rawManufacturer,
					"cleanedManufacturer": cleanedManufacturer,
				}).Warn("Low confidence: Replaced raw manufacturer with cleaned manufacturer by contains search")
				manufacturerToReturn = cleanedManufacturer
				foundCleanedManufacturer = true
				break
			}
		}

		if !foundCleanedManufacturer {
			// Worst case, use the raw value
			logrus.WithFields(logrus.Fields{"rawManufacturer": rawManufacturer}).Error("Could not find an appropriate match for the given raw manufacturer, using the value as-is")
			manufacturerToReturn = rawManufacturer
		}
	}

	if alias, exists := manufacturerAliasesMap[manufacturerToReturn]; exists {
		logrus.WithFields(logrus.Fields{"manufacturerToReturn": manufacturerToReturn, "manufacturerAlias": alias}).Info("Returning an alias for the given manufacturer")
		return alias
	}

	return manufacturerToReturn
}

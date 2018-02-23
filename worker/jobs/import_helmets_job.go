package jobs

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/xrash/smetrics"
)

// ImportHelmetsJob imports all helmet data from SHARP and SNELL into the database. It tries to normalize helmet models and manufacturers while doing this in order to have a clean data set.
type ImportHelmetsJob struct {
	ProductRepository      *repositories.ProductRepository
	SNELLHelmetRepository  *repositories.SNELLHelmetRepository
	SHARPHelmetRepository  *repositories.SHARPHelmetRepository
	ManufacturerRepository *repositories.ManufacturerRepository
}

const helmetType string = "helmet"

// Run invokes the job and returns an error if any errors occurred while processing the helmet data.
func (j *ImportHelmetsJob) Run() error {
	sharpProducts := make([]*entities.ProductDocument, 0)
	snellProducts := make([]*entities.ProductDocument, 0)

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

	modelAliases, err := j.ProductRepository.GetAllModelAliases()
	if err != nil {
		return err
	}

	// NOTE: This call blocks for about a minute on average as we need to fetch 400+ HTML files and scrape them for data.
	sharpHelmets, err := j.SHARPHelmetRepository.GetAll()
	if err != nil {
		return err
	}

	var jobCompletedWithWarnings bool
	for _, sharpHelmet := range sharpHelmets {
		cleanedManufacturer := findCleanedManufacturer(sharpHelmet.Manufacturer, manufacturers, manufacturerAliasesMap)
		modelAlias := findAliasForModel(modelAliases, cleanedManufacturer, sharpHelmet.Model)
		product := &entities.ProductDocument{
			ImageURL:            sharpHelmet.ImageURL,
			LatchPercentage:     sharpHelmet.LatchPercentage,
			Manufacturer:        cleanedManufacturer,
			Materials:           sharpHelmet.Materials,
			Model:               sharpHelmet.Model,
			ModelAlias:          "",
			PriceInUSDMultiple:  sharpHelmet.ApproximatePriceInUsdMultiple,
			RetentionSystem:     sharpHelmet.RetentionSystem,
			Sizes:               sharpHelmet.Sizes,
			Subtype:             sharpHelmet.Subtype,
			Type:                helmetType,
			UUID:                uuid.New(),
			WeightInLbsMultiple: sharpHelmet.WeightInLbsMultiple,
		}

		if modelAlias != "" {
			logrus.WithFields(logrus.Fields{
				"model":      product.Model,
				"modelAlias": modelAlias,
			}).Info("Replacing model with an alias")
			product.ModelAlias = product.Model
			product.Model = modelAlias
		}

		product.Certifications.SHARP = sharpHelmet.Certifications
		product.Certifications.ECE = sharpHelmet.IsECECertified
		sharpProducts = append(sharpProducts, product)
	}

	snellHelmets, err := j.SNELLHelmetRepository.GetAllByCertification("M2015")
	if err != nil {
		return err
	}

	for _, snellHelmet := range snellHelmets {
		cleanedManufacturer := findCleanedManufacturer(snellHelmet.Manufacturer, manufacturers, manufacturerAliasesMap)
		matchingSHARPProduct, success := findMatchingSHARPProduct(cleanedManufacturer, snellHelmet.Model, sharpProducts)
		if !jobCompletedWithWarnings && !success {
			jobCompletedWithWarnings = true
		}

		if matchingSHARPProduct != nil {
			logrus.WithFields(logrus.Fields{
				"manufacturer": matchingSHARPProduct.Manufacturer,
				"model":        matchingSHARPProduct.Model,
			}).Info("Updated a SHARP helmet to have SNELL and DOT ratings")
			matchingSHARPProduct.Certifications.SNELL = true
			matchingSHARPProduct.Certifications.DOT = true
		} else {
			logrus.WithFields(logrus.Fields{
				"manufacturer": cleanedManufacturer,
				"model":        snellHelmet.Model,
			}).Info("Could not find a matching SHARP helmet, so initializing a helmet with only SNELL and DOT ratings")

			sizes := strings.Split(snellHelmet.Size, ",")
			snellProduct := &entities.ProductDocument{Manufacturer: cleanedManufacturer, Model: snellHelmet.Model, UUID: uuid.New(), Type: helmetType, Subtype: snellHelmet.FaceConfig, Sizes: sizes}
			snellProduct.Certifications.SNELL = true
			snellProduct.Certifications.DOT = true
			snellProducts = append(snellProducts, snellProduct)
		}
	}

	combinedProductsList := append(sharpProducts, snellProducts...)
	for _, product := range combinedProductsList {
		validator := &entities.ProductDocumentValidator{Product: product}
		validationErr := validator.Validate()
		if validationErr != nil {
			logrus.WithFields(logrus.Fields{
				"manufacturer":    product.Manufacturer,
				"model":           product.Model,
				"validationError": validationErr,
			}).Warning("Validation failed, continuing to the next helmet")
			continue
		}

		existingProduct, err := j.ProductRepository.GetByModel(product.Manufacturer, product.Model)
		if err != nil {
			return err
		}

		if existingProduct == nil {
			err := j.ProductRepository.CreateProduct(product)
			if err != nil {
				return err
			}
		} else {
			product.UUID = existingProduct.UUID
			err := j.ProductRepository.UpdateProduct(product)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func findAliasForModel(aliases []entities.ProductModelAlias, manufacturer string, model string) string {
	for _, alias := range aliases {
		if strings.EqualFold(alias.Manufacturer, manufacturer) && strings.EqualFold(alias.Model, model) {
			return alias.ModelAlias
		}
	}
	return ""
}

const boostThreshold float64 = 0.7
const prefixSize int = 4

func findMatchingSHARPProduct(cleanedSNELLManufacturer string, rawSNELLModel string, sharpProducts []*entities.ProductDocument) (*entities.ProductDocument, bool) {
	possibleSHARPHelmets := make([]*entities.ProductDocument, 0)
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
		return nil, false
	}

	confidenceMap := make(map[string]float64)
	sort.Slice(possibleSHARPHelmets, func(i int, j int) bool {
		firstSHARPHelmet := possibleSHARPHelmets[i]
		secondSHARPHelmet := possibleSHARPHelmets[j]

		lowercaseRawSNELLModel := strings.ToLower(rawSNELLModel)

		lowercaseFirstSHARPModel := strings.ToLower(firstSHARPHelmet.Model)
		lowercaseSecondSHARPModel := strings.ToLower(secondSHARPHelmet.Model)

		lowercaseFirstSHARPModelAlias := strings.ToLower(firstSHARPHelmet.ModelAlias)
		lowercaseSecondSHARPModelAlias := strings.ToLower(secondSHARPHelmet.ModelAlias)

		firstModelMatchConfidence := smetrics.JaroWinkler(lowercaseRawSNELLModel, lowercaseFirstSHARPModel, boostThreshold, prefixSize)
		secondModelMatchConfidence := smetrics.JaroWinkler(lowercaseRawSNELLModel, lowercaseSecondSHARPModel, boostThreshold, prefixSize)

		firstModelAliasMatchConfidence := smetrics.JaroWinkler(lowercaseRawSNELLModel, lowercaseFirstSHARPModelAlias, boostThreshold, prefixSize)
		secondModelAliasMatchConfidence := smetrics.JaroWinkler(lowercaseRawSNELLModel, lowercaseSecondSHARPModelAlias, boostThreshold, prefixSize)

		if _, exists := confidenceMap[firstSHARPHelmet.Model]; !exists {
			confidenceMap[firstSHARPHelmet.Model] = math.Max(firstModelMatchConfidence, firstModelAliasMatchConfidence)
		}

		if _, exists := confidenceMap[secondSHARPHelmet.Model]; !exists {
			confidenceMap[secondSHARPHelmet.Model] = math.Max(secondModelMatchConfidence, secondModelAliasMatchConfidence)
		}

		return (firstModelMatchConfidence + firstModelAliasMatchConfidence) > (secondModelMatchConfidence + secondModelAliasMatchConfidence)
	})

	mostLikelySHARPHelmet := possibleSHARPHelmets[0]
	confidence := confidenceMap[mostLikelySHARPHelmet.Model]
	logEntry := logrus.WithFields(logrus.Fields{
		"rawSNELLModel":             rawSNELLModel,
		"mostLikelySHARPModel":      mostLikelySHARPHelmet.Model,
		"mostLikelySHARPModelAlias": mostLikelySHARPHelmet.ModelAlias,
		"confidence":                confidence,
	})

	// if we're 90% confident that the model matches, use the value
	if confidence >= 0.9 {
		logEntry.Info("High confidence: found matching SHARP model using Jaro-Winkler algorithm")
		return mostLikelySHARPHelmet, true
	}

	logEntry.Warn("Low confidence: SHARP match found, but confidence too low. Ignoring.")
	return nil, false
}

func findCleanedManufacturer(rawManufacturer string, cleanedManufacturers []string, manufacturerAliasesMap map[string]string) string {
	mostLikelyManufacturers := make([]string, len(cleanedManufacturers))
	copy(mostLikelyManufacturers, cleanedManufacturers)

	confidenceMap := make(map[string]float64)

	sort.Slice(mostLikelyManufacturers, func(i int, j int) bool {
		firstManufacturer := mostLikelyManufacturers[i]
		secondManufacturer := mostLikelyManufacturers[j]

		lowercaseRawManufacturer := strings.ToLower(rawManufacturer)

		lowercaseFirstManufacturer := strings.ToLower(mostLikelyManufacturers[i])
		lowercaseSecondManufacturer := strings.ToLower(mostLikelyManufacturers[j])

		firstManufacturerMatchConfidence := smetrics.JaroWinkler(lowercaseRawManufacturer, lowercaseFirstManufacturer, boostThreshold, prefixSize)
		secondManufacturerMatchConfidence := smetrics.JaroWinkler(lowercaseRawManufacturer, lowercaseSecondManufacturer, boostThreshold, prefixSize)

		if _, exists := confidenceMap[firstManufacturer]; !exists {
			confidenceMap[firstManufacturer] = firstManufacturerMatchConfidence
		}

		if _, exists := confidenceMap[secondManufacturer]; !exists {
			confidenceMap[secondManufacturer] = secondManufacturerMatchConfidence
		}

		return firstManufacturerMatchConfidence > secondManufacturerMatchConfidence
	})

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

package jobs

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	// "errors"
	"fmt"
	"strings"
	// "fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/xrash/smetrics"
	"sort"
	// "strings"
)

type ImportHelmetsJob struct {
	ProductRepository      *repositories.ProductRepository
	SNELLHelmetRepository  *repositories.SNELLHelmetRepository
	SHARPHelmetRepository  *repositories.SHARPHelmetRepository
	ManufacturerRepository *repositories.ManufacturerRepository
}

const helmetType string = "helmet"

// Get SHARP data
// Get SNELL data

// Must be run first:
// For each helmet in SHARP, try to find helmets by manufacturer+model combo
// does it already exist and are the SHARP fields different? If so, replace SHARP subdocument; else, create document.

// The below 2 steps can be run in parallel:

// For each helmet in SNELL, try to find helmets by manufacturer+model combo
// does it already exist? If so, set document.certifications.SNELL to true if it isn't already true; else, create document and log a warning that we couldn't find a matching SHARP helmet.

// For each helmet in the database, query CJ Affiliate's product data using Helmet manufacturer + model. Order by price descending, take top result, get product description.
// set price to the price
// if request limit reached, wait for 1.5 minutes and keep going
func (self *ImportHelmetsJob) Run() error {
	sharpProducts := make([]*entities.ProductDocument, 0)
	snellProducts := make([]*entities.ProductDocument, 0)

	manufacturers, err := self.ManufacturerRepository.GetAll()
	if err != nil {
		return err
	}

	// NOTE: This call blocks for about a minute on average as we need to fetch 400+ HTML files and scrape them for data.
	sharpHelmets, err := self.SHARPHelmetRepository.GetAll()
	if err != nil {
		return err
	}

	var jobCompletedWithWarnings bool
	for _, sharpHelmet := range sharpHelmets {
		cleanedManufacturer, success := findCleanedManufacturer(sharpHelmet.Manufacturer, manufacturers)
		if !jobCompletedWithWarnings && !success {
			jobCompletedWithWarnings = true
		}
		sharpHelmet.Manufacturer = cleanedManufacturer

		product := &entities.ProductDocument{
			ImageURL:            sharpHelmet.ImageURL,
			LatchPercentage:     sharpHelmet.LatchPercentage,
			Manufacturer:        sharpHelmet.Manufacturer,
			Materials:           sharpHelmet.Materials,
			Model:               sharpHelmet.Model,
			ModelAlias:          "",
			PriceInUsdMultiple:  0,
			RetentionSystem:     sharpHelmet.RetentionSystem,
			Sizes:               sharpHelmet.Sizes,
			Subtype:             sharpHelmet.Subtype,
			Type:                helmetType,
			UUID:                uuid.New(),
			WeightInLbsMultiple: sharpHelmet.WeightInLbsMultiple,
		}

		product.Certifications.SHARP = sharpHelmet.Certifications
		product.Certifications.ECE = sharpHelmet.IsECECertified
		sharpProducts = append(sharpProducts, product)
	}

	snellHelmets, err := self.SNELLHelmetRepository.GetAllByCertification("M2015")
	if err != nil {
		return err
	}

	for _, snellHelmet := range snellHelmets {
		cleanedManufacturer, success := findCleanedManufacturer(snellHelmet.Manufacturer, manufacturers)
		if !jobCompletedWithWarnings && !success {
			jobCompletedWithWarnings = true
		}
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
		existingProduct, err := self.ProductRepository.GetByModel(product.Manufacturer, product.Model)
		if err != nil {
			return err
		}

		if existingProduct == nil {
			err := self.ProductRepository.CreateProduct(product)
			if err != nil {
				return err
			}
		} else {
			product.UUID = existingProduct.UUID
			err := self.ProductRepository.UpdateProduct(product)
			if err != nil {
				return err
			}
		}
	}

	return nil
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

		firstModelMatchConfidence := smetrics.JaroWinkler(lowercaseRawSNELLModel, lowercaseFirstSHARPModel, boostThreshold, prefixSize)
		secondModelMatchConfidence := smetrics.JaroWinkler(lowercaseRawSNELLModel, lowercaseSecondSHARPModel, boostThreshold, prefixSize)

		if _, exists := confidenceMap[firstSHARPHelmet.Model]; !exists {
			confidenceMap[firstSHARPHelmet.Model] = firstModelMatchConfidence
		}

		if _, exists := confidenceMap[secondSHARPHelmet.Model]; !exists {
			confidenceMap[secondSHARPHelmet.Model] = secondModelMatchConfidence
		}

		return firstModelMatchConfidence > secondModelMatchConfidence
	})

	mostLikelySHARPHelmet := possibleSHARPHelmets[0]
	confidence := confidenceMap[mostLikelySHARPHelmet.Model]
	logEntry := logrus.WithFields(logrus.Fields{
		"rawSNELLModel":        rawSNELLModel,
		"mostLikelySHARPModel": mostLikelySHARPHelmet.Model,
		"confidence":           confidence,
	})

	// if we're 90% confident that the model matches, use the value
	if confidence >= 0.9 {
		logEntry.Info("High confidence: found matching SHARP model using Jaro-Winkler algorithm")
		return mostLikelySHARPHelmet, true
	} else {
		logEntry.Warn("Low confidence: SHARP match found, but confidence too low. Ignoring.")
		return nil, false
	}
}

func findCleanedManufacturer(rawManufacturer string, cleanedManufacturers []string) (string, bool) {
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

	// if we're 70% confident that the manufacturer matches, use the value
	if confidence >= 0.7 {
		logEntry.Info("High confidence: replaced raw manufacturer with cleaned manufacturer using Jaro-Winkler algorithm")
		return mostLikelyManufacturer, true
	} else { // otherwise, try a stupider contains search to see if anything matches
		for _, cleanedManufacturer := range cleanedManufacturers {
			lowercaseCleanedManufacturer := strings.ToLower(cleanedManufacturer)
			lowercaseRawManufacturer := strings.ToLower(rawManufacturer)

			if strings.HasPrefix(lowercaseRawManufacturer, lowercaseCleanedManufacturer) || strings.Contains(lowercaseRawManufacturer, fmt.Sprintf(" %s", lowercaseCleanedManufacturer)) {
				logrus.WithFields(logrus.Fields{
					"rawManufacturer":     rawManufacturer,
					"cleanedManufacturer": cleanedManufacturer,
				}).Warn("Low confidence: Replaced raw manufacturer with cleaned manufacturer by contains search")
				return cleanedManufacturer, true
			}
		}
		// worst case, return the raw value
		logrus.WithFields(logrus.Fields{"rawManufacturer": rawManufacturer}).Error("Could not find an appropriate match for the given raw manufacturer, using the value as-is", rawManufacturer)
		return rawManufacturer, false
	}
}

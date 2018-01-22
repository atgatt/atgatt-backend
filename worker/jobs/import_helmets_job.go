package jobs

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ImportHelmetsJob struct {
	ProductRepository     *repositories.ProductRepository
	SNELLHelmetRepository *repositories.SNELLHelmetRepository
	SHARPHelmetRepository *repositories.SHARPHelmetRepository
}

// Get SHARP data
// Get SNELL data

// Must be run first:
// For each helmet in SHARP, try to find helmets by manufacturer+model combo
// does it already exist and are the SHARP fields different? If so, replace SHARP subdocument; else, create document.

// The below 2 steps can be run in parallel:

// For each helmet in SNELL, try to find helmets by manufacturer+model combo
// does it already exist? If so, set document.certifications.SNELL to true if it isn't already true; else, create document and log a warning that we couldn't find a matching SHARP helmet.

// For each helmet in the database, query CJ Affiliate's product data using Helmet manufacturer + model. Order by price descending, take top result, get product description.
// If no results, log a warning; if results:
// does description contain "DOT"? Set DOT to true.
// set price to the price
// if request limit reached, wait for 1.5 minutes and keep going
func (self *ImportHelmetsJob) Run() error {
	products := make([]*entities.ProductDocument, 0)
	// NOTE: This call blocks for about a minute on average as we need to fetch 400+ HTML files and scrape them for data.
	sharpHelmets, err := self.SHARPHelmetRepository.GetAllHelmets()
	if err != nil {
		return err
	}

	for _, sharpHelmet := range sharpHelmets {
		product := &entities.ProductDocument{
			ImageURL:        sharpHelmet.ImageURL,
			LatchPercentage: sharpHelmet.LatchPercentage,
			Manufacturer:    sharpHelmet.Manufacturer,
			Materials:       sharpHelmet.Materials,
			Model:           sharpHelmet.Model,
			ModelAlias:      "",
			PriceInUsd:      "",
			RetentionSystem: sharpHelmet.RetentionSystem,
			Sizes:           sharpHelmet.Sizes,
			Subtype:         sharpHelmet.Subtype,
			Type:            "helmet",
			UUID:            uuid.New(),
			WeightInLbs:     sharpHelmet.WeightInLbs,
		}
		product.Certifications.SHARP = sharpHelmet.Certifications
		products = append(products, product)
	}
	logrus.Info(sharpHelmets)

	snellHelmets, err := self.SNELLHelmetRepository.GetAllHelmets()
	if err != nil {
		return err
	}
	logrus.Info(snellHelmets)

	return nil
}

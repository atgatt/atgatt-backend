package jobs

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/worker/jobs/helpers"

	"github.com/sirupsen/logrus"
)

// RecalculateSafetyPercentagesJob recalculates all safety percentages for all products in the database
type RecalculateSafetyPercentagesJob struct {
	ProductRepository *repositories.ProductRepository
}

// Run executes the job
func (j *RecalculateSafetyPercentagesJob) Run() error {
	return helpers.ForEachProduct(j.ProductRepository, func(product *entities.ProductDocument, productLogger *logrus.Entry) error {
		oldSafetyPercentage := product.SafetyPercentage
		newSafetyPercentage := product.CalculateSafetyPercentage()
		productLogger.WithFields(
			logrus.Fields{
				"manufacturer":        product.Manufacturer,
				"model":               product.Model,
				"oldSafetyPercentage": oldSafetyPercentage,
				"newSafetyPercentage": newSafetyPercentage,
			}).Info("Updating safety percentage for product")

		product.SafetyPercentage = newSafetyPercentage
		err := j.ProductRepository.UpdateProduct(product)
		if err != nil {
			return err
		}

		return nil
	})
}

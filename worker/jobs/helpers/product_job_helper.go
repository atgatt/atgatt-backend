package helpers

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"

	"github.com/sirupsen/logrus"
)

// ForEachProduct iterates over all the products in the database and runs the current action on the given product
func ForEachProduct(productRepository *repositories.ProductRepository, action func(product *entities.ProductDocument, productLogger *logrus.Entry) error) error {
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

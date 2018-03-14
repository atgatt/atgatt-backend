package helpers

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
)

// ForEachProduct iterates over all the products in the database and runs the current action on the given product
func ForEachProduct(productRepository *repositories.ProductRepository, action func(product *entities.ProductDocument) error) error {
	start := 0
	limit := 25
	currProducts, err := productRepository.GetAllPaged(start, limit)
	if err != nil {
		return err
	}

	for len(currProducts) > 0 {
		for _, product := range currProducts {
			err := action(&product)
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

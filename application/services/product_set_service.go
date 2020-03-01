package services

import (
	"atgatt-backend/persistence/entities"
	"atgatt-backend/persistence/repositories"

	"github.com/google/uuid"
)

// ProductSetService contains service methods to deal with productset data
type ProductSetService struct {
	ProductSetRepository *repositories.ProductSetRepository
	ProductRepository    *repositories.ProductRepository
}

// UpsertProductSet either creates a new product set or gets an existing one if an exact match is found in the DB
func (s *ProductSetService) UpsertProductSet(sourceProductSetID *uuid.UUID, productID uuid.UUID) (uuid.UUID, error) {
	var productSet *entities.ProductSet

	product, err := s.ProductRepository.GetByUUID(productID.String())
	if err != nil {
		return uuid.Nil, err
	}

	if sourceProductSetID != nil {
		productSet, err = s.ProductSetRepository.GetByUUID(*sourceProductSetID)
		if err != nil {
			return uuid.Nil, err
		}
	}

	if productSet == nil {
		productSet = &entities.ProductSet{UUID: uuid.New()}
	}

	productSet.AddOrReplaceProduct(product)

	matchingUUID, err := s.ProductSetRepository.GetMatchingProductSetUUID(productSet)
	if err != nil {
		return uuid.Nil, err
	}

	if matchingUUID != uuid.Nil {
		return matchingUUID, nil
	}

	uuidCreated, err := s.ProductSetRepository.Create(productSet)
	return uuidCreated, err
}

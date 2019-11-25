package services

import (
	"crashtested-backend/persistence/repositories"

	"github.com/google/uuid"
)

// ProductSetService contains service methods to deal with productset data
type ProductSetService struct {
	ProductSetRepository *repositories.ProductSetRepository
}

// UpsertProductSet either creates a new product set or gets an existing one if an exact match is found in the DB
func (s *ProductSetService) UpsertProductSet(existingProductSetID *uuid.UUID, productID *uuid.UUID, productType string) {

}

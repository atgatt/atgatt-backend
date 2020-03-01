package dtos

import (
	"atgatt-backend/persistence/entities"

	"github.com/google/uuid"
)

// ProductSetProductsDTO represents a collection of products associated with a product set
type ProductSetProductsDTO struct {
	UUID          uuid.UUID
	HelmetProduct *entities.Product
	JacketProduct *entities.Product
	PantsProduct  *entities.Product
	BootsProduct  *entities.Product
	GlovesProduct *entities.Product
}

package responses

import (
	"crashtested-backend/persistence/entities"

	"github.com/google/uuid"
)

// GetProductSetDetailsResponse returns the details of the product set
type GetProductSetDetailsResponse struct {
	ID uuid.UUID `json:"id"`

	// Products
	HelmetProduct *entities.Product `json:"helmetProduct"`
	JacketProduct *entities.Product `json:"jacketProduct"`
	PantsProduct  *entities.Product `json:"pantsProduct"`
	BootsProduct  *entities.Product `json:"bootsProduct"`
	GlovesProduct *entities.Product `json:"glovesProduct"`
}

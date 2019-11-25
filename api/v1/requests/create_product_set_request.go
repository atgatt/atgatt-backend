package requests

import (
	"github.com/google/uuid"
)

// CreateProductSetRequest represents a request to create a new product set
type CreateProductSetRequest struct {
	SourceProductSetID *uuid.UUID `json:"source_product_set_id"`
	ProductID          *uuid.UUID `json:"product_id"`
}

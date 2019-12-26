package requests

import (
	"github.com/google/uuid"
)

// CreateProductSetRequest represents a request to create a new product set
type CreateProductSetRequest struct {
	SourceProductSetID *uuid.UUID `json:"sourceProductSetID"`
	ProductID          uuid.UUID  `json:"productID"`
}

package requests

import "github.com/google/uuid"

// GetProductSetDetails represents a request to get all of the products identified by this product set ID
type GetProductSetDetails struct {
	ProductSetID uuid.UUID `json:"productSetID"`
}

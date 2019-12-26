package responses

import (
	"github.com/google/uuid"
)

// CreateProductSetResponse returns the UUID of the created/found product set
type CreateProductSetResponse struct {
	ID uuid.UUID `json:"id"`
}

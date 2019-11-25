package entities

import "github.com/google/uuid"

// ProductSet represents a user-defined group of gear to be bought together
type ProductSet struct {
	ID   int
	UUID *uuid.UUID

	Name            string
	Description     string
	HelmetProductID *int
	JacketProductID *int
	PantsProductID  *int
	BootsProductID  *int
	GlovesProductID *int
}

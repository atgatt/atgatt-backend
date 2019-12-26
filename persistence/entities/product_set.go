package entities

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ProductSet represents a user-defined group of gear to be bought together
type ProductSet struct {
	// General fields
	ID          int
	UUID        uuid.UUID
	Name        string
	Description string

	// Products
	HelmetProductID *int
	HelmetProduct   *Product

	JacketProductID *int
	JacketProduct   *Product

	PantsProductID *int
	PantsProduct   *Product

	BootsProductID *int
	BootsProduct   *Product

	GlovesProductID *int
	GlovesProduct   *Product
}

// AddOrReplaceProduct adds or overwrites a given product on this product set, using the type to determine which product to create/update.
func (p *ProductSet) AddOrReplaceProduct(product *Product) error {
	if product == nil {
		return errors.New("product cannot be nil")
	}

	switch product.Type {
	case ProductTypeHelmet:
		p.HelmetProduct = product
		p.HelmetProductID = &product.ID
	case ProductTypeJacket:
		p.JacketProduct = product
		p.JacketProductID = &product.ID
	case ProductTypePants:
		p.PantsProduct = product
		p.PantsProductID = &product.ID
	case ProductTypeBoots:
		p.BootsProduct = product
		p.BootsProductID = &product.ID
	case ProductTypeGloves:
		p.GlovesProduct = product
		p.GlovesProductID = &product.ID
	default:
		return fmt.Errorf("Unexpected product type %s", product.Type)
	}

	return nil
}

package entities

import validation "github.com/go-ozzo/ozzo-validation"

// ProductValidator validates that a product has its basic fields set. Just used during the import process for now.
type ProductValidator struct {
	Product *Product
}

// Validate checks that the manufacturer and model are defined for the given product
func (v *ProductValidator) Validate() error {
	err := validation.ValidateStruct(v.Product,
		validation.Field(&v.Product.Manufacturer,
			validation.Required.Error("The manufacturer must be specified"),
		),
		validation.Field(&v.Product.Model,
			validation.Required.Error("The model must be specified"),
		),
	)

	if err != nil {
		return err
	}

	return nil
}

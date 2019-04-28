package queries

import (
	"errors"

	"github.com/go-ozzo/ozzo-validation"
)

// FilterProductsQueryValidator is responsible for validating (or returning an error) for a single FilterProductsQuery
type FilterProductsQueryValidator struct {
	Query              *FilterProductsQuery
	AllowedOrderFields map[string]bool
}

// Validate returns an error if validation failed, or nil if it was successful
func (v *FilterProductsQueryValidator) Validate() error {
	err := validation.ValidateStruct(v.Query,
		validation.Field(&v.Query.Start,
			validation.Min(0),
		),
		validation.Field(&v.Query.Limit,
			validation.Required.Error("The limit must be specified"),
			validation.Min(1),
			validation.Max(25),
		),
		validation.Field(&v.Query.UsdPriceRange,
			validation.Required,
			validation.Length(2, 2).Error("The price range array must contain exactly two elements"),
			validation.By(PriceRange),
		),
		validation.Field(&v.Query.HelmetCertifications,
			validation.By(v.HelmetCertifications),
		),
	)
	if err != nil {
		return err
	}

	err = validation.Validate(v.Query.Order.Field, validation.Required, validation.By(v.OrderByField))
	if err != nil {
		validationErrors := validation.Errors{}
		validationErrors["order.field"] = errors.New("Ordering is not allowed by this field")
		return validationErrors
	}

	return nil
}

// OrderByField ensures that the user only orders by one of the allowed values and not a random DB column
func (v *FilterProductsQueryValidator) OrderByField(value interface{}) error {
	orderByField := value.(string)

	if _, exists := v.AllowedOrderFields[orderByField]; !exists {
		return errors.New("The order field that was specified is not allowed to be used")
	}
	return nil
}

// HelmetCertifications ensures we either have helmet certifications or jacket certifications but not both
func (v *FilterProductsQueryValidator) HelmetCertifications(value interface{}) error {
	if v.Query.HelmetCertifications != nil && v.Query.JacketCertifications != nil {
		return errors.New("Helmet and Jacket certifications cannot be supplied together")
	}

	return nil
}

// PriceRange ensures that the priceRange is valid
func PriceRange(value interface{}) error {
	priceRange := value.([]int)
	if priceRange[0] > priceRange[1] {
		return errors.New("The minimum price cannot be greater than the maximum price")
	}
	if priceRange[0] < 0 {
		return errors.New("The minimum price must be greater than or equal to $0")
	}
	if priceRange[1] <= 0 {
		return errors.New("The maximum price must be positive")
	}

	return nil
}

package queries

import (
	"errors"
	"github.com/go-ozzo/ozzo-validation"
)

type FilterProductsQueryValidator struct {
	Query *FilterProductsQuery
}

func (self *FilterProductsQueryValidator) Validate() error {
	err := validation.ValidateStruct(self.Query,
		validation.Field(&self.Query.Start,
			validation.Min(0),
		),
		validation.Field(&self.Query.Limit,
			validation.Required.Error("The limit must be specified"),
			validation.Min(1),
			validation.Max(25),
		),
		validation.Field(&self.Query.UsdPriceRange,
			validation.Required,
			validation.Length(2, 2).Error("The price range array must contain exactly two elements"),
			validation.By(PriceRange),
		),
	)
	if err != nil {
		return err
	}

	err = validation.Validate(self.Query.Order.Field, validation.Required, validation.By(OrderByField))
	if err != nil {
		validationErrors := validation.Errors{}
		validationErrors["order.field"] = errors.New("Ordering is not allowed by this field")
		return validationErrors
	}

	return nil
}

func OrderByField(value interface{}) error {
	orderByField := value.(string)
	allowedOrderFields := make(map[string]bool)
	allowedOrderFields["document->>'priceInUsdMultiple'"] = true
	allowedOrderFields["document->>'manufacturer'"] = true
	allowedOrderFields["document->>'model'"] = true
	allowedOrderFields["created_at_utc"] = true
	allowedOrderFields["updated_at_utc"] = true
	allowedOrderFields["id"] = true

	if _, exists := allowedOrderFields[orderByField]; !exists {
		return errors.New("The order field that was specified is not allowed to be used")
	}
	return nil
}

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

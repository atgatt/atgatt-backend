package handlers

import (
	"crashtested-backend/persistence/queries"
	"crashtested-backend/persistence/repositories"
	"net/http"

	"github.com/labstack/echo"
)

// ProductHandler contains functions related to filtering and updating Products
type ProductHandler struct {
	Repository         *repositories.ProductRepository
	AllowedOrderFields map[string]bool
}

// FilterProducts returns a subset of products from the database based off a user-supplied query, where all parameters are AND'd together
func (p *ProductHandler) FilterProducts(context echo.Context) (err error) {
	query := new(queries.FilterProductsQuery)
	if err := context.Bind(query); err != nil {
		return err
	}

	err = (&queries.FilterProductsQueryValidator{Query: query, AllowedOrderFields: p.AllowedOrderFields}).Validate()
	if err != nil {
		return context.JSON(http.StatusBadRequest, err)
	}

	products, err := p.Repository.FilterProducts(query)
	if err != nil {
		return err
	}

	return context.JSON(http.StatusOK, products)
}

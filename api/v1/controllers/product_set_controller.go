package controllers

import (
	"crashtested-backend/api/v1/requests"
	"crashtested-backend/persistence/repositories"
	"net/http"

	"github.com/labstack/echo/v4"
)

// ProductSetController contains functions related to filtering and updating ProductSets
type ProductSetController struct {
	Repository *repositories.ProductRepository
}

// CreateProductSet creates a product set based off an existing product set, or creates a new one in the database
func (p *ProductSetController) CreateProductSet(context echo.Context) (err error) {
	query := new(requests.CreateProductSetRequest)
	if err := context.Bind(query); err != nil {
		return err
	}

	return context.NoContent(http.StatusOK)
}

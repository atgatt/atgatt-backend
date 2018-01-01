package handlers

import (
	"crashtested-backend/persistence/queries"
	"crashtested-backend/persistence/repositories"
	"net/http"

	"github.com/labstack/echo"
)

type ProductHandler struct {
	Repository *repositories.ProductRepository
}

func (self *ProductHandler) FilterProducts(context echo.Context) (err error) {
	query := new(queries.FilterProductsQuery)
	if err := context.Bind(query); err != nil {
		return err
	}

	err = (&queries.FilterProductsQueryValidator{Query: query}).Validate()
	if err != nil {
		return context.JSON(http.StatusBadRequest, err)
	}

	products, err := self.Repository.FilterProducts(query)
	if err != nil {
		return err
	}

	return context.JSON(http.StatusOK, products)
}

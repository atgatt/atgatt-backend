package main

import (
	"crashtested-backend/api/requests"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func filterProducts(context echo.Context) error {
	request := new(requests.FilterProductsRequest)
	if err := context.Bind(request); err != nil {
		return err
	}
	return context.JSON(http.StatusOK, request)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	e.POST("/api/v1/products/filter", filterProducts)

	e.Logger.Fatal(e.Start(":5000"))
}

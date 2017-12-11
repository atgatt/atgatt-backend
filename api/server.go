package api

import (
	"crashtested-backend/api/handlers"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	healthCheckHandler := &handlers.HealthCheckHandler{BuildNumber: os.Getenv("BUILD_NUMBER")}
	productsHandler := &handlers.ProductsHandler{}

	e.GET("/", healthCheckHandler.Healthcheck)
	e.POST("/api/v1/products/filter", productsHandler.FilterProducts)

	err := e.Start(":5000")
	e.Logger.Fatal(err)
}

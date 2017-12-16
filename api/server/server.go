package server

import (
	"crashtested-backend/api/handlers"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Server struct {
	Port string
}

func (self *Server) Build() *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())

	healthCheckHandler := &handlers.HealthCheckHandler{BuildNumber: os.Getenv("BUILD_NUMBER"), Name: "crashtested-api", Version: "1.0.1"}
	productsHandler := &handlers.ProductsHandler{}

	e.GET("/", healthCheckHandler.Healthcheck)
	e.POST("/v1/products/filter", productsHandler.FilterProducts)

	return e
}

func (self *Server) StartAndBlock() {
	e := self.Build()
	err := e.Start(self.Port)
	e.Logger.Fatal(err)
}

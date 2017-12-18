package handlers

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Server struct {
	Port        string
	Name        string
	Version     string
	BuildNumber string

	Echo *echo.Echo
}

func (self *Server) Build() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: []string{"https://staging.crashtested.co", "https://www.staging.crashtested.co", "https://crashtested.co", "https://www.crashtested.co"}}))

	healthCheckHandler := &HealthCheckHandler{Name: self.Name, Version: self.Version, BuildNumber: self.BuildNumber}
	productsHandler := &ProductsHandler{}

	e.GET("/", healthCheckHandler.Healthcheck)
	e.POST("/v1/products/filter", productsHandler.FilterProducts)

	self.Echo = e
}

func (self *Server) StartAndBlock() {
	self.Build()
	err := self.Echo.Start(self.Port)
	self.Echo.Logger.Fatal(err)
}

func (self *Server) Stop() {
	self.Echo.Close()
}

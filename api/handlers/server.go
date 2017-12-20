package handlers

import (
	"crashtested-backend/api/configuration"
	"fmt"
	"os"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/color"
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
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: []string{"https://staging.crashtested.co", "https://www.staging.crashtested.co", "https://crashtested.co", "https://www.crashtested.co"}}))

	config := &configuration.Configuration{DatabaseConnectionString: os.Getenv("DATABASE_CONNECTION_STRING")}
	err := validation.ValidateStruct(config,
		validation.Field(&config.DatabaseConnectionString, validation.Required),
	)
	if err != nil {
		fmt.Printf("Failed to start the API because the app configuration could not be validated: %s", err.Error())
		os.Exit(-1)
	}

	healthCheckHandler := &HealthCheckHandler{Name: self.Name, Version: self.Version, BuildNumber: self.BuildNumber}
	productsHandler := &ProductHandler{}

	e.GET("/", healthCheckHandler.Healthcheck)
	e.POST("/v1/products/filter", productsHandler.FilterProducts)

	self.Echo = e
}

func (self *Server) StartAndBlock() {
	self.Build()
	coloredConsole := color.New()
	coloredConsole.Printf("⇨ http server started on %s\n", color.Green(fmt.Sprintf("%s%s", "http://localhost", self.Port)))
	err := self.Echo.Start(self.Port)
	self.Echo.Logger.Fatal(err)
}

func (self *Server) Stop() {
	self.Echo.Close()
}

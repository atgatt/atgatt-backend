package handlers

import (
	"crashtested-backend/api/configuration"
	"crashtested-backend/persistence/helpers"
	"crashtested-backend/persistence/repositories"
	"fmt"
	"os"

	"github.com/bakatz/echo-logrusmiddleware"
	"github.com/bshuster-repo/logruzio"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
)

type Server struct {
	Port          string
	Name          string
	Version       string
	BuildNumber   string
	Configuration *configuration.Configuration
	echoInstance  *echo.Echo
}

func (self *Server) Build() {
	e := echo.New()
	e.HideBanner = true
	e.Logger = logrusmiddleware.Logger{Logger: logrus.StandardLogger()}
	e.Use(logrusmiddleware.Hook())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: []string{"https://staging.crashtested.co", "https://www.staging.crashtested.co", "https://crashtested.co", "https://www.crashtested.co"}}))

	if self.Configuration == nil {
		logrus.Fatal("Failed to start the API because the app configuration was not specified")
		os.Exit(-1)
	}

	err := validation.ValidateStruct(self.Configuration,
		validation.Field(&self.Configuration.DatabaseConnectionString, validation.Required),
		validation.Field(&self.Configuration.AppEnvironment, validation.Required),
	)
	if err != nil {
		logrus.Fatalf("Failed to start the API because the app configuration could not be validated: %s", err.Error())
		os.Exit(-1)
	}

	if self.Configuration.LogzioToken != "" {
		logContext := logrus.Fields{
			"BuildNumber": self.BuildNumber,
			"Version":     self.Version,
		}
		logzioHook, err := logruzio.New(self.Configuration.LogzioToken, fmt.Sprintf("%s-%s", self.Name, self.Configuration.AppEnvironment), logContext)
		if err != nil {
			logrus.Fatalf("Failed to start the API because the logger could not be initialized: %s", err.Error())
		}
		logrus.AddHook(logzioHook)
	} else {
		logrus.Warn("LOGZIO_TOKEN was not set, so all application logs are going to stdout")
	}

	err = helpers.RunMigrations(self.Configuration.DatabaseConnectionString, "persistence/migrations")
	if err != nil {
		logrus.Errorf("Failed to run migrations, but starting the app anyway: %s", err.Error())
	}

	healthCheckHandler := &HealthCheckHandler{Name: self.Name, Version: self.Version, BuildNumber: self.BuildNumber, MigrationsRepository: &repositories.MigrationsRepository{ConnectionString: self.Configuration.DatabaseConnectionString}}
	productsHandler := &ProductHandler{Repository: &repositories.ProductRepository{ConnectionString: self.Configuration.DatabaseConnectionString}}

	e.GET("/", healthCheckHandler.Healthcheck)
	e.HEAD("/", healthCheckHandler.Healthcheck)
	e.POST("/v1/products/filter", productsHandler.FilterProducts)

	self.echoInstance = e
}

func (self *Server) StartAndBlock() {
	self.Build()
	logrus.Infof("-> http server started on %s%s", "http://localhost", self.Port)
	err := self.echoInstance.Start(self.Port)
	logrus.Fatalf("Failed to start the server: %s", err.Error())
}

func (self *Server) Stop() {
	self.echoInstance.Close()
}

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

// Server contains the bootstrapping code for the API
type Server struct {
	Port          string
	Name          string
	Version       string
	BuildNumber   string
	CommitHash    string
	Configuration *configuration.Configuration
	echoInstance  *echo.Echo
}

// Build initializes all dependencies required by the API and exits with a nonzero status code if there's a problem
func (s *Server) Build() {
	e := echo.New()
	e.HideBanner = true
	e.Logger = logrusmiddleware.Logger{Logger: logrus.StandardLogger()}
	e.Use(middleware.RequestID())
	e.Use(logrusmiddleware.Hook())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: []string{"https://staging.crashtested.co", "https://www.staging.crashtested.co", "https://crashtested.co", "https://www.crashtested.co"}}))

	if s.Configuration == nil {
		logrus.Fatal("Failed to start the API because the app configuration was not specified")
		os.Exit(-1)
	}

	err := validation.ValidateStruct(s.Configuration,
		validation.Field(&s.Configuration.DatabaseConnectionString, validation.Required),
		validation.Field(&s.Configuration.AppEnvironment, validation.Required),
	)
	if err != nil {
		logrus.Fatalf("Failed to start the API because the app configuration could not be validated: %s", err.Error())
		os.Exit(-1)
	}

	if s.Configuration.LogzioToken != "" {
		logContext := logrus.Fields{
			"BuildNumber":    s.BuildNumber,
			"Version":        s.Version,
			"CommitHash":     s.CommitHash,
			"AppEnvironment": s.Configuration.AppEnvironment,
		}
		logzioHook, err := logruzio.New(s.Configuration.LogzioToken, fmt.Sprintf("%s-%s", s.Name, s.Configuration.AppEnvironment), logContext)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to start the API because the logger could not be initialized")
			os.Exit(-1)
		}
		logrus.AddHook(logzioHook)
	} else {
		logrus.Warn("LOGZIO_TOKEN was not set, so all application logs are going to stdout")
	}

	err = helpers.RunMigrations(s.Configuration.DatabaseConnectionString, "persistence/migrations")
	if err != nil {
		logrus.WithError(err).Error("Failed to run migrations, but starting the app anyway: %s")
	}

	healthCheckHandler := &HealthCheckHandler{Name: s.Name, Version: s.Version, BuildNumber: s.BuildNumber, CommitHash: s.CommitHash, MigrationsRepository: &repositories.MigrationsRepository{ConnectionString: s.Configuration.DatabaseConnectionString}}

	allowedOrderFields := make(map[string]bool)
	allowedOrderFields["document->>'priceInUsdMultiple'"] = true
	allowedOrderFields["document->>'manufacturer'"] = true
	allowedOrderFields["document->>'model'"] = true
	allowedOrderFields["document->>'safetyPercentage'"] = true
	allowedOrderFields["created_at_utc"] = true
	allowedOrderFields["updated_at_utc"] = true
	allowedOrderFields["id"] = true
	productsHandler := &ProductHandler{Repository: &repositories.ProductRepository{ConnectionString: s.Configuration.DatabaseConnectionString}, AllowedOrderFields: allowedOrderFields}

	e.GET("/", healthCheckHandler.Healthcheck)
	e.HEAD("/", healthCheckHandler.Healthcheck)
	e.POST("/v1/products/filter", productsHandler.FilterProducts)

	s.echoInstance = e
}

// StartAndBlock first initializes the server, then starts it up and blocks
func (s *Server) StartAndBlock() {
	s.Build()
	logrus.Infof("-> http server started on %s%s", "http://localhost", s.Port)
	err := s.echoInstance.Start(s.Port)
	logrus.Fatalf("Failed to start the server: %s", err.Error())
}

// Stop just ensures that the echoInstance is closed
func (s *Server) Stop() {
	s.echoInstance.Close()
}

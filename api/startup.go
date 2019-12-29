package api

import (
	"crashtested-backend/api/settings"
	"crashtested-backend/api/v1/controllers"
	"crashtested-backend/application/services"
	helpers "crashtested-backend/common/auth"
	persistenceHelpers "crashtested-backend/persistence/helpers"
	"crashtested-backend/persistence/repositories"
	"fmt"
	"net/http"
	"os"

	logrusmiddleware "github.com/bakatz/echo-logrusmiddleware"
	"github.com/bshuster-repo/logruzio"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	// Importing the PostgreSQL driver with side effects because we need to call sql.Open() to run queries
	_ "github.com/jackc/pgx/v4/stdlib"
)

// Server contains the bootstrapping code for the API
type Server struct {
	Port         string
	Name         string
	Version      string
	BuildNumber  string
	CommitHash   string
	Settings     *settings.Settings
	echoInstance *echo.Echo
}

// Bootstrap first initializes the server, then starts it up and blocks
func (s *Server) Bootstrap() {
	e := echo.New()
	e.HideBanner = true
	e.Logger = logrusmiddleware.Logger{Logger: logrus.StandardLogger()}
	e.Use(middleware.RequestID())

	if s.Settings == nil {
		logrus.Fatal("Failed to start the API because the app configuration was not specified")
		os.Exit(-1)
	}

	config := &logrusmiddleware.Config{
		IncludeRequestBodies:  s.Settings.LogAPIRequests,
		IncludeResponseBodies: s.Settings.LogAPIRequests,
	}

	signingKey, err := helpers.GetAuth0PublicKey(s.Settings.Auth0Domain)
	if err != nil {
		logrus.Fatalf("Failed to start the API because the JWT signing key could not be retrieved from auth0: %s", err.Error())
		os.Exit(-1)
	}

	e.Use(logrusmiddleware.HookWithConfig(*config))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://master.crashtested.co", "https://www.master.crashtested.co", "https://crashtested.co", "https://www.crashtested.co"},
	}))

	err = validation.ValidateStruct(s.Settings,
		validation.Field(&s.Settings.DatabaseConnectionString, validation.Required),
		validation.Field(&s.Settings.AppEnvironment, validation.Required),
	)
	if err != nil {
		logrus.Fatalf("Failed to start the API because the app configuration could not be validated: %s", err.Error())
		os.Exit(-1)
	}

	if s.Settings.LogzioToken != "" {
		logContext := logrus.Fields{
			"BuildNumber":    s.BuildNumber,
			"Version":        s.Version,
			"CommitHash":     s.CommitHash,
			"AppEnvironment": s.Settings.AppEnvironment,
		}
		logzioHook, err := logruzio.New(s.Settings.LogzioToken, fmt.Sprintf("%s-%s", s.Name, s.Settings.AppEnvironment), logContext)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to start the API because the logger could not be initialized")
			os.Exit(-1)
		}
		logrus.AddHook(logzioHook)
	} else {
		logrus.Warn("LOGZIO_TOKEN was not set, so all application logs are going to stdout")
	}

	err = persistenceHelpers.RunMigrations(s.Settings.DatabaseConnectionString, "persistence/migrations")
	if err != nil {
		logrus.WithError(err).Error("Failed to run migrations, but starting the app anyway")
	}

	db, err := sqlx.Open("pgx", s.Settings.DatabaseConnectionString)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to start the API because the database connection could not be established")
		os.Exit(-1)
	}

	healthCheckController := &controllers.HealthCheckController{Name: s.Name, Version: s.Version, BuildNumber: s.BuildNumber, CommitHash: s.CommitHash, MigrationsRepository: &repositories.MigrationsRepository{DB: db}}

	allowedOrderFields := make(map[string]bool)
	allowedOrderFields["document->>'searchPriceCents'"] = true
	allowedOrderFields["document->>'manufacturer'"] = true
	allowedOrderFields["document->>'model'"] = true
	allowedOrderFields["document->>'safetyPercentage'"] = true
	allowedOrderFields["created_at_utc"] = true
	allowedOrderFields["updated_at_utc"] = true
	allowedOrderFields["id"] = true

	productRepository := &repositories.ProductRepository{DB: db}
	productSetRepository := &repositories.ProductSetRepository{DB: db}

	productsController := &controllers.ProductController{Repository: productRepository, AllowedOrderFields: allowedOrderFields}
	productSetController := &controllers.ProductSetController{Service: &services.ProductSetService{ProductRepository: productRepository, ProductSetRepository: productSetRepository}, Repository: productSetRepository}
	marketingController := &controllers.MarketingController{Repository: &repositories.MarketingRepository{DB: db}}

	jwtMiddleware := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningMethod: "RS256",
		SigningKey:    signingKey,
		Skipper: func(c echo.Context) bool {
			return c.Request().Method == http.MethodOptions
		},
	})

	// Translate all persistence errors to the correct json response, leave the default behavior for all other internal errors
	e.HTTPErrorHandler = func(err error, context echo.Context) {
		if err == repositories.ErrEntityNotFound {
			err = context.NoContent(http.StatusNotFound)
			if err != nil {
				logrus.WithError(err).Error("Failed to send response")
			}
		} else {
			e.DefaultHTTPErrorHandler(err, context)
		}
	}

	// Public endpoints
	e.GET("/", healthCheckController.Healthcheck)
	e.HEAD("/", healthCheckController.Healthcheck)
	e.POST("/v1/products/filter", productsController.FilterProducts)
	e.GET("/v1/products/:uuid", productsController.GetProductDetails)
	e.POST("/v1/product-sets", productSetController.CreateProductSet)
	e.GET("/v1/product-sets/:uuid", productSetController.GetProductSetDetails)
	e.POST("/v1/marketing/email", marketingController.CreateMarketingEmail)

	// Endpoints requiring JWT authentication
	e.POST("/v1/products/:id/reviews", productsController.CreateReview, jwtMiddleware)

	err = e.Start(s.Port)
	if err != nil {
		logrus.WithError(err).Error("Failed to start the server")
	}
}

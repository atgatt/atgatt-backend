package worker

import (
	"crashtested-backend/common/logging/helpers"
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/worker/jobs"
	"crashtested-backend/worker/settings"
	"net/http"
	"os"
	"runtime"

	"github.com/borderstech/artifex"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	logrusmiddleware "github.com/bakatz/echo-logrusmiddleware"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"

	// Importing the PostgreSQL driver with side effects because we need to call sql.Open() to run queries
	_ "github.com/lib/pq"
)

// Server contains the bootstrapping code for the worker
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

	config := s.Settings

	logrusMiddlewareConf := &logrusmiddleware.Config{
		IncludeRequestBodies:  true,
		IncludeResponseBodies: true,
	}
	e.Use(logrusmiddleware.HookWithConfig(*logrusMiddlewareConf))

	if s.Settings == nil {
		logrus.Fatal("Failed to start the API because the app configuration was not specified")
		os.Exit(-1)
	}

	err := validation.ValidateStruct(s.Settings,
		validation.Field(&s.Settings.DatabaseConnectionString, validation.Required),
		validation.Field(&s.Settings.AppEnvironment, validation.Required),
	)
	if err != nil {
		logrus.Fatalf("Failed to start the API because the app configuration could not be validated: %s", err.Error())
		os.Exit(-1)
	}

	logContext := logrus.Fields{
		"BuildNumber":    s.BuildNumber,
		"Version":        s.Version,
		"CommitHash":     s.CommitHash,
		"AppEnvironment": s.Settings.AppEnvironment,
	}
	helpers.InitializeLogzio(s.Settings.LogzioToken, s.Name, s.Settings.AppEnvironment, logContext)

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewEnvCredentials(),
	}))

	s3Uploader := s3manager.NewUploader(sess)

	db, err := sqlx.Open("postgres", config.DatabaseConnectionString)
	if err != nil {
		logrus.WithError(err).Error("Encountered an error while opening a database connection")
		os.Exit(-1)
	}
	defer db.Close()

	productRepository := &repositories.ProductRepository{DB: db}

	importHelmetsJob := &jobs.ImportHelmetsJob{
		ProductRepository:      productRepository,
		SHARPHelmetRepository:  &repositories.SHARPHelmetRepository{Limit: -1},
		SNELLHelmetRepository:  &repositories.SNELLHelmetRepository{},
		ManufacturerRepository: &repositories.ManufacturerRepository{DB: db},
		S3Uploader:             s3Uploader,
		S3Bucket:               config.AWS.S3Bucket,
	}

	syncRevzillaDataJob := &jobs.SyncRevzillaDataJob{ProductRepository: productRepository, CJAPIKey: config.CJAPIKey}

	// 10 workers, 100 max in job queue
	numWorkers := runtime.NumCPU()
	logrus.WithField("numWorkers", numWorkers).Info("Starting job queue")
	jobQueue := artifex.NewDispatcher(numWorkers, 100)
	jobQueue.Start()
	logrus.Info("Job queue started")

	// NOTE: for now, just running both jobs at the same time. Should refactor this to be separate jobs once there are more than two that need to run (it makes sense to group these two together for now)
	e.POST("/jobs", func(context echo.Context) (err error) {
		logrus.Info("Got a message, dispatching work to job queue")
		jobQueue.Dispatch(func() {
			logrus.Info("Starting Import Helmets Job")
			err = importHelmetsJob.Run()
			if err != nil {
				logrus.WithError(err).Error("Import Helmets Job completed with errors")
				return
			}
			logrus.Info("Import Helmets Job completed successfully")

			logrus.Info("Starting Revzilla Sync Job")
			err = syncRevzillaDataJob.Run()
			if err != nil {
				logrus.WithError(err).Error("Sync RevZilla job completed with errors")
				return
			}
			logrus.Info("Sync RevZilla job completed successfully")
		})

		logrus.Info("Finished dispatching work to job queue, returning OK")
		var emptyResponse struct{}
		return context.JSON(http.StatusOK, emptyResponse)
	})

	err = e.Start(s.Port)
	if err != nil {
		logrus.WithError(err).Error("Failed to start the server")
	}
}

package worker

import (
	"crashtested-backend/application/parsers"
	loggingHelpers "crashtested-backend/common/logging"
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
	loggingHelpers.InitializeLogzio(s.Settings.LogzioToken, s.Name, s.Settings.AppEnvironment, logContext)

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewEnvCredentials(),
		Endpoint:    &s.Settings.AWS.MinioEndpoint, // If MINIO_ENDPOINT is defined, we use the simulated S3 service for integration tests; otherwise use the real AWS S3 service
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
		SHARPHelmetParser:      &parsers.SHARPHelmetParser{Limit: -1},
		SNELLHelmetParser:      &parsers.SNELLHelmetParser{},
		ManufacturerRepository: &repositories.ManufacturerRepository{DB: db},
		S3Uploader:             s3Uploader,
		S3Bucket:               config.AWS.S3Bucket,
	}

	syncRevzillaHelmetsJob := &jobs.SyncRevzillaHelmetsJob{ProductRepository: productRepository, CJAPIKey: config.CJAPIKey}
	syncRevzillaJacketsJob := &jobs.SyncRevzillaJacketsJob{ProductRepository: productRepository, S3Uploader: s3Uploader, S3Bucket: config.AWS.S3Bucket}

	numWorkers := runtime.NumCPU()
	logrus.WithField("numWorkers", numWorkers).Info("Starting job queue")
	jobQueue := artifex.NewDispatcher(numWorkers, 100)
	jobQueue.Start()
	logrus.Info("Job queue started")

	// Jobs
	s.registerJob(e, jobQueue, "import_helmets", importHelmetsJob)
	s.registerJob(e, jobQueue, "sync_revzilla_helmets", syncRevzillaHelmetsJob)
	s.registerJob(e, jobQueue, "sync_revzilla_jackets", syncRevzillaJacketsJob)

	// Healthcheck endpoint
	e.GET("/", func(context echo.Context) error {
		var emptyResponse struct{}
		return context.JSON(http.StatusOK, emptyResponse)
	})

	err = e.Start(s.Port)
	if err != nil {
		logrus.WithError(err).Error("Failed to start the server")
	}
}

func (s *Server) registerJob(e *echo.Echo, jobQueue *artifex.Dispatcher, name string, job jobs.Job) {
	e.POST("/jobs/"+name, func(context echo.Context) error {
		jobLogger := logrus.WithField("jobName", name)
		jobLogger.Info("Triggered, dispatching work to job queue")

		runJobFunc := func() error {
			jobLogger.Info("Starting Job")
			err := job.Run()
			if err != nil {
				jobLogger.WithError(err).Error("Job completed with errors")
				return err
			}
			jobLogger.Info("Job completed successfully")
			return nil
		}

		runJobFuncWrapper := func() {
			_ = runJobFunc()
		}

		if s.Settings.UseSynchronousJobRunner {
			err := runJobFunc()
			if err != nil {
				return err
			}
		} else {
			err := jobQueue.Dispatch(runJobFuncWrapper)
			if err != nil {
				jobLogger.WithError(err).Error("Could not start job due to an error")
				return err
			}
		}

		jobLogger.Info("Finished dispatching work to job queue, returning OK")
		var emptyResponse struct{}
		return context.JSON(http.StatusOK, emptyResponse)
	})
}

package main

import (
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/worker/jobs"
	"crashtested-backend/worker/settings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	// Importing the PostgreSQL driver with side effects because we need to call sql.Open() to run queries
	_ "github.com/lib/pq"
)

func main() {
	config := settings.GetSettingsFromEnvironment()

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewEnvCredentials(),
	}))

	s3Uploader := s3manager.NewUploader(sess)

	db, err := sqlx.Open("postgres", config.DatabaseConnectionString)
	if err != nil {
		logrus.WithError(err).Error("Encountered an error while opening a database connection")
		return
	}

	productRepository := &repositories.ProductRepository{DB: db}

	importHelmetsJob := &jobs.ImportHelmetsJob{
		ProductRepository:      productRepository,
		SHARPHelmetRepository:  &repositories.SHARPHelmetRepository{Limit: -1},
		SNELLHelmetRepository:  &repositories.SNELLHelmetRepository{},
		ManufacturerRepository: &repositories.ManufacturerRepository{DB: db},
		S3Uploader:             s3Uploader,
		S3Bucket:               config.AWS.S3Bucket,
	}

	recaclulateSafetyPercentagesJob := &jobs.RecalculateSafetyPercentagesJob{ProductRepository: productRepository}
	syncRevzillaDataJob := &jobs.SyncRevzillaDataJob{ProductRepository: productRepository, CJAPIKey: config.CJAPIKey}

	err = importHelmetsJob.Run()
	if err != nil {
		logrus.WithError(err).Error("Import Helmets Job completed with errors")
	} else {
		logrus.Info("Import Helmets Job completed successfully")
	}

	err = syncRevzillaDataJob.Run()
	if err != nil {
		logrus.WithError(err).Error("Sync RevZilla job completed with errors")
	} else {
		logrus.Info("Sync RevZilla job completed successfully")
	}

	err = recaclulateSafetyPercentagesJob.Run()
	if err != nil {
		logrus.WithError(err).Error("Recalculate Safety Job completed with errors")
	} else {
		logrus.Info("Recalculate Safety Job completed successfully")
	}
}

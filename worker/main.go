package main

import (
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/worker/configuration"
	"crashtested-backend/worker/jobs"

	"github.com/ngs/go-amazon-product-advertising-api/amazon"
	"github.com/sirupsen/logrus"
)

func main() {
	config := configuration.GetDefaultConfiguration()

	amazonClient, err := amazon.New(config.AmazonAssociates.AccessKey, config.AmazonAssociates.SecretKey, config.AmazonAssociates.AssociateID, amazon.RegionUS)
	if err != nil {
		logrus.WithError(err).Error("Encountered an error while creating an Amazon Client")
		return
	}

	productRepository := &repositories.ProductRepository{ConnectionString: config.DatabaseConnectionString}

	syncAmazonDataJob := &jobs.SyncAmazonDataJob{AmazonClient: amazonClient, ProductRepository: productRepository}
	importHelmetsJob := &jobs.ImportHelmetsJob{
		ProductRepository:      productRepository,
		SHARPHelmetRepository:  &repositories.SHARPHelmetRepository{Limit: -1},
		SNELLHelmetRepository:  &repositories.SNELLHelmetRepository{},
		ManufacturerRepository: &repositories.ManufacturerRepository{ConnectionString: config.DatabaseConnectionString},
	}

	err = importHelmetsJob.Run()
	if err != nil {
		logrus.WithError(err).Error("Import Helmets Job completed with errors")
	} else {
		logrus.Info("Import Helmets Job completed successfully")
	}

	err = syncAmazonDataJob.Run()
	if err != nil {
		logrus.WithError(err).Error("Amazon Sync Job completed with errors")
	} else {
		logrus.Info("Amazon Sync Job completed successfully")
	}
}

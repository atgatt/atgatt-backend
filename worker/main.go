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
		logrus.Errorf("Encountered an error while creating an Amazon Client: %s", err.Error())
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

	err = syncAmazonDataJob.Run()
	if err != nil {
		logrus.Errorf("Amazon Sync Job completed with errors: %s", err.Error())
	} else {
		logrus.Info("Amazon Sync Job completed successfully")
	}
	err = importHelmetsJob.Run()
	if err != nil {
		logrus.Errorf("Import Helmets Job completed with errors: %s", err.Error())
	} else {
		logrus.Info("Import Helmets Job completed successfully")
	}
}

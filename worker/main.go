package main

import (
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/worker/configuration"
	"crashtested-backend/worker/jobs"
	"github.com/sirupsen/logrus"
)

func main() {
	config := configuration.GetDefaultConfiguration()
	job := &jobs.ImportHelmetsJob{
		ProductRepository:      &repositories.ProductRepository{ConnectionString: config.DatabaseConnectionString},
		SHARPHelmetRepository:  &repositories.SHARPHelmetRepository{Limit: -1},
		SNELLHelmetRepository:  &repositories.SNELLHelmetRepository{},
		ManufacturerRepository: &repositories.ManufacturerRepository{ConnectionString: config.DatabaseConnectionString},
	}
	err := job.Run()
	if err != nil {
		logrus.Errorf("Job completed with errors: %s", err.Error())
	} else {
		logrus.Info("Job completed successfully")
	}
}

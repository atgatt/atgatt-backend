package main

import (
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/worker/configuration"
	"crashtested-backend/worker/jobs"
)

func main() {
	config := configuration.GetDefaultConfiguration()
	job := &jobs.ImportHelmetsJob{
		ProductRepository:     &repositories.ProductRepository{ConnectionString: config.DatabaseConnectionString},
		SHARPHelmetRepository: &repositories.SHARPHelmetRepository{Limit: -1},
		SNELLHelmetRepository: &repositories.SNELLHelmetRepository{},
	}
	job.Run()
}

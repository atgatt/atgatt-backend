package main

import (
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/worker/configuration"
	"crashtested-backend/worker/jobs"
)

func main() {
	config := configuration.GetDefaultConfiguration()
	job := &jobs.ImportHelmetDataJob{ProductRepository: &repositories.ProductRepository{ConnectionString: config.DatabaseConnectionString}}
	job.Run()
}

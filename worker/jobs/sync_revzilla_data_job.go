package jobs

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/worker/jobs/helpers"
)

// SyncRevzillaDataJob syncs revzilla price and buy urls by calling the CJ Affiliate API and pointing it at RevZilla's advertiser ID
type SyncRevzillaDataJob struct {
	ProductRepository *repositories.ProductRepository
}

// Run executes the job
func (j *SyncRevzillaDataJob) Run() error {
	return helpers.ForEachProduct(j.ProductRepository, func(product *entities.ProductDocument) error {
		return nil
	})
}

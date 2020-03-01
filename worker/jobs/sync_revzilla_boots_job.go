package jobs

import (
	"atgatt-backend/application/clients"
	appEntities "atgatt-backend/application/entities"
	"atgatt-backend/persistence/entities"
	"atgatt-backend/persistence/repositories"
	"atgatt-backend/worker/jobs/helpers"

	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

// SyncRevzillaBootsJob scrapes all of RevZilla's boots data
type SyncRevzillaBootsJob struct {
	ProductRepository      *repositories.ProductRepository
	RevzillaClient         clients.RevzillaClient
	S3Uploader             s3manageriface.UploaderAPI
	S3Bucket               string
	EnableMinProductsCheck bool
}

// Run executes the job
func (j *SyncRevzillaBootsJob) Run() error {
	updateCertsFunc := func(productToPersist *entities.Product, revzillaProduct *appEntities.RevzillaProduct) {
		updated, newZone := productToPersist.UpdateSingleZoneCertificationsByDescriptionParts(productToPersist.BootsCertifications.Overall, revzillaProduct.DescriptionParts)
		if updated {
			productToPersist.BootsCertifications.Overall = newZone
		}
		productToPersist.UpdateGenericSubtypeByDescriptionParts(revzillaProduct.DescriptionParts)
	}

	return helpers.RunRevzillaImport("motorcycle-boots", "boots", j.RevzillaClient, j.ProductRepository, j.S3Uploader, j.S3Bucket, j.EnableMinProductsCheck, updateCertsFunc)
}

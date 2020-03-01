package jobs

import (
	"atgatt-backend/application/clients"
	appEntities "atgatt-backend/application/entities"
	"atgatt-backend/persistence/entities"
	"atgatt-backend/persistence/repositories"
	"atgatt-backend/worker/jobs/helpers"

	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

// SyncRevzillaGlovesJob scrapes all of RevZilla's gloves data
type SyncRevzillaGlovesJob struct {
	ProductRepository      *repositories.ProductRepository
	RevzillaClient         clients.RevzillaClient
	S3Uploader             s3manageriface.UploaderAPI
	S3Bucket               string
	EnableMinProductsCheck bool
}

// Run executes the job
func (j *SyncRevzillaGlovesJob) Run() error {
	updateCertsFunc := func(productToPersist *entities.Product, revzillaProduct *appEntities.RevzillaProduct) {
		updated, newZone := productToPersist.UpdateSingleZoneCertificationsByDescriptionParts(productToPersist.GlovesCertifications.Overall, revzillaProduct.DescriptionParts)
		if updated {
			productToPersist.GlovesCertifications.Overall = newZone
		}
		productToPersist.UpdateGenericSubtypeByDescriptionParts(revzillaProduct.DescriptionParts)
	}

	return helpers.RunRevzillaImport("motorcycle-gloves", "gloves", j.RevzillaClient, j.ProductRepository, j.S3Uploader, j.S3Bucket, j.EnableMinProductsCheck, updateCertsFunc)
}

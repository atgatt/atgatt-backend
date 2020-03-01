package jobs

import (
	"atgatt-backend/application/clients"
	appEntities "atgatt-backend/application/entities"
	"atgatt-backend/persistence/entities"
	"atgatt-backend/persistence/repositories"
	"atgatt-backend/worker/jobs/helpers"

	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

// SyncRevzillaPantsJob scrapes all of RevZilla's pants data
type SyncRevzillaPantsJob struct {
	ProductRepository      *repositories.ProductRepository
	RevzillaClient         clients.RevzillaClient
	S3Uploader             s3manageriface.UploaderAPI
	S3Bucket               string
	EnableMinProductsCheck bool
}

// Run executes the job
func (j *SyncRevzillaPantsJob) Run() error {
	updateCertsFunc := func(productToPersist *entities.Product, revzillaProduct *appEntities.RevzillaProduct) {
		productToPersist.UpdatePantsCertificationsByDescriptionParts(revzillaProduct.DescriptionParts)
		productToPersist.UpdatePantsSubtypeByDescriptionParts(revzillaProduct.DescriptionParts)
	}

	return helpers.RunRevzillaImport("motorcycle-pants", "pants", j.RevzillaClient, j.ProductRepository, j.S3Uploader, j.S3Bucket, j.EnableMinProductsCheck, updateCertsFunc)
}

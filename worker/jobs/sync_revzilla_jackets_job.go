package jobs

import (
	"atgatt-backend/application/clients"
	appEntities "atgatt-backend/application/entities"
	"atgatt-backend/persistence/entities"
	"atgatt-backend/persistence/repositories"
	"atgatt-backend/worker/jobs/helpers"

	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

// SyncRevzillaJacketsJob scrapes all of RevZilla's jacket data
type SyncRevzillaJacketsJob struct {
	ProductRepository      *repositories.ProductRepository
	RevzillaClient         clients.RevzillaClient
	S3Uploader             s3manageriface.UploaderAPI
	S3Bucket               string
	EnableMinProductsCheck bool
}

// Run executes the job
func (j *SyncRevzillaJacketsJob) Run() error {
	updateCertsFunc := func(productToPersist *entities.Product, revzillaProduct *appEntities.RevzillaProduct) {
		productToPersist.UpdateJacketCertificationsByDescriptionParts(revzillaProduct.DescriptionParts)
		productToPersist.UpdateGenericSubtypeByDescriptionParts(revzillaProduct.DescriptionParts)
	}

	return helpers.RunRevzillaImport("motorcycle-jackets-vests", "jacket", j.RevzillaClient, j.ProductRepository, j.S3Uploader, j.S3Bucket, j.EnableMinProductsCheck, updateCertsFunc)
}

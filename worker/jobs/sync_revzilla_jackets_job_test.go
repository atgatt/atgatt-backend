package jobs_test

import (
	"crashtested-backend/application/clients"
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/worker/jobs"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/jmoiron/sqlx"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	_ "github.com/lib/pq"
	. "github.com/onsi/gomega"
)

var httpRevzillaClient = clients.NewHTTPRevzillaClient()

type mockRevzillaClient struct {
	overviewsHTML string
}

func (r *mockRevzillaClient) GetAllJacketOverviewsHTML() (*goquery.Document, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r.overviewsHTML))
	return doc, err
}

func (r *mockRevzillaClient) GetDescriptionPartsHTMLByURL(url string) (*goquery.Document, error) {
	return httpRevzillaClient.GetDescriptionPartsHTMLByURL(url)
}

func (r *mockRevzillaClient) SetOverviewsHTML(htmlFilePath string) {
	bytes, _ := ioutil.ReadFile(htmlFilePath)
	r.overviewsHTML = string(bytes)
}

func Test_Run_should_create_all_new_jackets_if_none_exist(t *testing.T) {
	RegisterTestingT(t)
	productRepository := &repositories.ProductRepository{DB: sqlx.MustConnect("postgres", TestDatabaseConnectionString)}
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewEnvCredentials(),
	}))

	expectedProductIds := []string{"391137", "391141"}
	mockRevzillaClient := &mockRevzillaClient{}
	mockRevzillaClient.SetOverviewsHTML("../../seeds/mock-jackets-response.html")

	s3Uploader := s3manager.NewUploader(sess)
	job := &jobs.SyncRevzillaJacketsJob{ProductRepository: productRepository, S3Uploader: s3Uploader, S3Bucket: "junk", RevzillaClient: mockRevzillaClient}
	job.Run()

	product, _ := productRepository.GetByExternalID(expectedProductIds[0])
	Expect(product).ToNot(BeNil())
	Expect(product.JacketCertifications.Back).To(Equal(&entities.CEImpactZone{
		IsApproved: false,
		IsEmpty:    true,
		IsLevel2:   false,
	}))
	Expect(product.JacketCertifications.Chest).To(Equal(&entities.CEImpactZone{
		IsApproved: false,
		IsEmpty:    true,
		IsLevel2:   false,
	}))

	Expect(product.JacketCertifications.Elbow).To(Equal(&entities.CEImpactZone{
		IsApproved: false,
		IsEmpty:    false,
		IsLevel2:   true,
	}))

	Expect(product.JacketCertifications.Shoulder).To(Equal(&entities.CEImpactZone{
		IsApproved: false,
		IsEmpty:    false,
		IsLevel2:   true,
	}))

	Expect(product.JacketCertifications.FitsAirbag).To(BeFalse())
	Expect(product.Manufacturer).To(Equal("REAX"))
	Expect(product.Model).To(Equal("Folsom Leather Jacket"))
	Expect(product.Subtype).To(Equal("leather"))

	product, _ = productRepository.GetByExternalID(expectedProductIds[1])
	Expect(product).ToNot(BeNil())
	Expect(product.Manufacturer).To(Equal("REAX"))
	Expect(product.Model).To(Equal("Alta Mesh Jacket"))
	Expect(product.Subtype).To(Equal("textile"))
}

func Test_sync_revzilla_jackets_should_complete_successfully_with_a_full_set_of_data(t *testing.T) {
	RegisterTestingT(t)

	resp, err := http.Post(APIBaseURL+"/jobs/sync_revzilla_jackets", "application/json", strings.NewReader("{}"))
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

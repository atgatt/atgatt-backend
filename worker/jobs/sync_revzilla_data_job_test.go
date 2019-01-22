package jobs_test

import (
	"crashtested-backend/persistence/repositories"
	"net/http"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	. "github.com/onsi/gomega"
)

func Test_sync_revzilla_data_should_sync_revzilla_data_for_discontinued_and_active_products(t *testing.T) {
	RegisterTestingT(t)

	resp, err := http.Post(APIBaseURL+"/jobs/sync_revzilla_data", "application/json", strings.NewReader("{}"))
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	productRepository := &repositories.ProductRepository{DB: sqlx.MustOpen("postgres", TestDatabaseConnectionString)}
	activeAliasProduct, err := productRepository.GetByModel("Shoei", "X Spirit lll")
	Expect(err).To(BeNil())
	Expect(activeAliasProduct).ToNot(BeNil())
	Expect(activeAliasProduct.SearchPriceCents).To(BeNumerically(">", 0))
	Expect(activeAliasProduct.RevzillaBuyURL).ToNot(BeEmpty())
	Expect(activeAliasProduct.IsDiscontinued).To(BeFalse())

	activeNormalProduct, err := productRepository.GetByModel("Bell", "Star")
	Expect(err).To(BeNil())
	Expect(activeNormalProduct).ToNot(BeNil())
	Expect(activeNormalProduct.SearchPriceCents).To(BeNumerically(">", 0))
	Expect(activeNormalProduct.RevzillaBuyURL).ToNot(BeEmpty())
	Expect(activeNormalProduct.IsDiscontinued).To(BeFalse())

	discontinuedProduct, err := productRepository.GetByModel("Shoei", "X-12")
	Expect(err).To(BeNil())
	Expect(discontinuedProduct).ToNot(BeNil())
	Expect(discontinuedProduct.SearchPriceCents).To(Equal(0)) // make sure we didn't change the price for a discontinued product
	Expect(discontinuedProduct.RevzillaBuyURL).To(BeEmpty())
	Expect(discontinuedProduct.IsDiscontinued).To(BeTrue()) // make sure we set discontinued to true

	notFoundProduct, err := productRepository.GetByModel("IAMNOTREAL", "IDONOTEXIST")
	Expect(err).To(BeNil())
	Expect(notFoundProduct).ToNot(BeNil())
	Expect(notFoundProduct.SearchPriceCents).To(Equal(0)) // make sure we didn't change the price for a nonexistent product
	Expect(notFoundProduct.RevzillaBuyURL).To(BeEmpty())
	Expect(notFoundProduct.IsDiscontinued).To(BeFalse()) // make sure we didn't mark a nonexistent product as discontinued
}

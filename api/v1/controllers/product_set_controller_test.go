package controllers_test

import (
	"crashtested-backend/api/v1/requests"
	"crashtested-backend/api/v1/responses"
	httpHelpers "crashtested-backend/common/http"
	"crashtested-backend/seeds"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"

	. "github.com/onsi/gomega"
)

func Test_CreateProductSet_should_create_a_ProductSet_with_all_products_defined_without_an_existing_product_set(t *testing.T) {
	RegisterTestingT(t)

	mockProduct := seeds.GetProductSeeds()[0]
	request := &requests.CreateProductSetRequest{ProductID: mockProduct.UUID}

	responseBody := &responses.CreateProductSetResponse{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/product-sets", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	Expect(responseBody.ID).To(Not(Equal(uuid.Nil)))

	getResponseBody := &responses.GetProductSetDetailsResponse{}
	resp, err = httpHelpers.MakeJSONGETRequest(fmt.Sprintf("%s/v1/product-sets/%s", APIBaseURL, responseBody.ID.String()), getResponseBody)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(getResponseBody.HelmetProduct).To(Not(BeNil()))
}

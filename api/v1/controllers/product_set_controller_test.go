package controllers_test

import (
	"atgatt-backend/api/v1/requests"
	"atgatt-backend/api/v1/responses"
	httpHelpers "atgatt-backend/common/http"
	"atgatt-backend/persistence/entities"
	"atgatt-backend/seeds"
	"fmt"
	"net/http"
	"testing"

	golinq "github.com/ahmetb/go-linq"
	"github.com/google/uuid"

	. "github.com/onsi/gomega"
)

func applyProductToProductSet(expectedProduct *entities.Product, sourceProductSetID *uuid.UUID) (uuid.UUID, *responses.GetProductSetDetailsResponse) {
	request := &requests.CreateProductSetRequest{ProductID: expectedProduct.UUID, SourceProductSetID: sourceProductSetID}

	responseBody := &responses.CreateProductSetResponse{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/product-sets", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	Expect(responseBody.ID).To(Not(Equal(uuid.Nil)))

	initialUUID := responseBody.ID

	getResponseBody := &responses.GetProductSetDetailsResponse{}
	resp, err = httpHelpers.MakeJSONGETRequest(fmt.Sprintf("%s/v1/product-sets/%s", APIBaseURL, initialUUID.String()), getResponseBody)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	Expect(getResponseBody.ID).To(Equal(initialUUID))

	return initialUUID, getResponseBody
}

func Test_GetProductSetDetails_should_return_a_StatusBadRequest_status_code_when_the_product_set_uuid_is_malformed(t *testing.T) {
	RegisterTestingT(t)

	resp, err := httpHelpers.MakeJSONGETRequest(fmt.Sprintf("%s/v1/product-sets/some-weird-string", APIBaseURL), &responses.GetProductSetDetailsResponse{})

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_GetProductSetDetails_should_return_a_StatusNotFound_status_code_when_the_product_set_doesnt_exist(t *testing.T) {
	RegisterTestingT(t)

	resp, err := httpHelpers.MakeJSONGETRequest(fmt.Sprintf("%s/v1/product-sets/%s", APIBaseURL, uuid.Nil.String()), &responses.GetProductSetDetailsResponse{})

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
}

func Test_CreateProductSet_should_return_a_StatusNotFound_status_code_when_the_product_doesnt_exist(t *testing.T) {
	RegisterTestingT(t)

	request := &requests.CreateProductSetRequest{ProductID: uuid.Nil, SourceProductSetID: nil}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/product-sets", APIBaseURL), request, &responses.CreateProductSetResponse{})

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
}

func Test_CreateProductSet_should_create_a_ProductSet_with_one_product_without_an_existing_product_set(t *testing.T) {
	RegisterTestingT(t)

	helmetSeeds := []*entities.Product{}
	golinq.From(seeds.GetProductSeeds()).WhereT(func(product *entities.Product) bool {
		return product.Type == entities.ProductTypeHelmet
	}).ToSlice(&helmetSeeds)

	expectedHelmet := helmetSeeds[0]
	productSetID, getResponseBody := applyProductToProductSet(expectedHelmet, nil)
	Expect(productSetID).To(Not(Equal(uuid.Nil)))
	Expect(getResponseBody.HelmetProduct).To(Equal(expectedHelmet))
	Expect(getResponseBody.JacketProduct).To(BeNil())
	Expect(getResponseBody.PantsProduct).To(BeNil())
	Expect(getResponseBody.BootsProduct).To(BeNil())
	Expect(getResponseBody.GlovesProduct).To(BeNil())
}

func Test_CreateProductSet_should_return_an_existing_ProductSet_when_the_input_product_matches(t *testing.T) {
	RegisterTestingT(t)

	helmetSeeds := []*entities.Product{}
	golinq.From(seeds.GetProductSeeds()).WhereT(func(product *entities.Product) bool {
		return product.Type == entities.ProductTypeHelmet
	}).ToSlice(&helmetSeeds)

	expectedHelmet := helmetSeeds[2]
	initialUUID, getResponseBody := applyProductToProductSet(expectedHelmet, nil)
	Expect(getResponseBody.HelmetProduct).To(Equal(expectedHelmet))
	Expect(getResponseBody.JacketProduct).To(BeNil())
	Expect(getResponseBody.PantsProduct).To(BeNil())
	Expect(getResponseBody.BootsProduct).To(BeNil())
	Expect(getResponseBody.GlovesProduct).To(BeNil())

	// Now try to create the exact same product set that already exists, make sure we find the initial one again and don't insert a new row
	nextUUID, getResponseBody := applyProductToProductSet(expectedHelmet, nil)
	Expect(nextUUID).To(Equal(initialUUID))
	Expect(getResponseBody.HelmetProduct).To(Equal(expectedHelmet))
	Expect(getResponseBody.JacketProduct).To(BeNil())
	Expect(getResponseBody.PantsProduct).To(BeNil())
	Expect(getResponseBody.BootsProduct).To(BeNil())
	Expect(getResponseBody.GlovesProduct).To(BeNil())
}

func Test_CreateProductSet_should_create_a_ProductSet_with_all_products_without_an_existing_product_set(t *testing.T) {
	RegisterTestingT(t)

	seeds := seeds.GetProductSeeds()

	helmetSeeds := []*entities.Product{}
	golinq.From(seeds).WhereT(func(product *entities.Product) bool {
		return product.Type == entities.ProductTypeHelmet
	}).ToSlice(&helmetSeeds)

	jacketSeeds := []*entities.Product{}
	golinq.From(seeds).WhereT(func(product *entities.Product) bool {
		return product.Type == entities.ProductTypeJacket
	}).ToSlice(&jacketSeeds)

	pantsSeeds := []*entities.Product{}
	golinq.From(seeds).WhereT(func(product *entities.Product) bool {
		return product.Type == entities.ProductTypePants
	}).ToSlice(&pantsSeeds)

	bootsSeeds := []*entities.Product{}
	golinq.From(seeds).WhereT(func(product *entities.Product) bool {
		return product.Type == entities.ProductTypeBoots
	}).ToSlice(&bootsSeeds)

	glovesSeeds := []*entities.Product{}
	golinq.From(seeds).WhereT(func(product *entities.Product) bool {
		return product.Type == entities.ProductTypeGloves
	}).ToSlice(&glovesSeeds)

	expectedHelmet := helmetSeeds[1]
	initialUUID, getResponseBody := applyProductToProductSet(expectedHelmet, nil)
	firstUUID := initialUUID // save the first generated UUID, we should never use this one again

	Expect(getResponseBody.HelmetProduct).To(Equal(expectedHelmet))
	Expect(getResponseBody.JacketProduct).To(BeNil())
	Expect(getResponseBody.PantsProduct).To(BeNil())
	Expect(getResponseBody.BootsProduct).To(BeNil())
	Expect(getResponseBody.GlovesProduct).To(BeNil())

	expectedJacket := jacketSeeds[0]
	nextUUID, getResponseBody := applyProductToProductSet(expectedJacket, &initialUUID)
	Expect(nextUUID).To(Not(Equal(initialUUID)))
	Expect(nextUUID).To(Not(Equal(firstUUID)))
	initialUUID = nextUUID

	expectedPants := pantsSeeds[0]
	nextUUID, getResponseBody = applyProductToProductSet(expectedPants, &initialUUID)
	Expect(nextUUID).To(Not(Equal(initialUUID)))
	Expect(nextUUID).To(Not(Equal(firstUUID)))
	initialUUID = nextUUID

	expectedBoots := bootsSeeds[0]
	nextUUID, getResponseBody = applyProductToProductSet(expectedBoots, &initialUUID)
	Expect(nextUUID).To(Not(Equal(initialUUID)))
	Expect(nextUUID).To(Not(Equal(firstUUID)))
	initialUUID = nextUUID

	expectedGloves := glovesSeeds[0]
	nextUUID, getResponseBody = applyProductToProductSet(expectedGloves, &initialUUID)
	Expect(nextUUID).To(Not(Equal(initialUUID)))
	Expect(nextUUID).To(Not(Equal(firstUUID)))
	initialUUID = nextUUID

	Expect(getResponseBody.HelmetProduct).To(Equal(expectedHelmet))
	Expect(getResponseBody.JacketProduct).To(Equal(expectedJacket))
	Expect(getResponseBody.PantsProduct).To(Equal(expectedPants))
	Expect(getResponseBody.BootsProduct).To(Equal(expectedBoots))
	Expect(getResponseBody.GlovesProduct).To(Equal(expectedGloves))
}

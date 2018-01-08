package handlers

import (
	"crashtested-backend/api/requests/helpers"
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/queries"
	"crashtested-backend/seeds"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_FilterProducts_should_return_all_of_the_products_data_when_the_limit_is_large_enough_and_there_are_no_optional_filters_set(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 10000}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).To(BeEquivalentTo(seeds.GetProductSeeds()))
}

func Test_FilterProducts_should_return_all_of_the_products_that_have_the_given_subtype_when_the_subtypes_array_has_one_element(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 10000}}
	request.Order.Field = "created_at_utc"
	request.Subtypes = []string{"full"}

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.Subtype).To(Equal(request.Subtypes[0]))
	}
}

func Test_FilterProducts_should_return_all_of_the_products_that_have_the_given_subtype_when_the_subtypes_array_has_multiple_elements(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 10000}}
	request.Order.Field = "created_at_utc"
	request.Subtypes = []string{"full", "modular"}

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		isExpectedSubtype := item.Subtype == request.Subtypes[0] || item.Subtype == request.Subtypes[1]
		Expect(isExpectedSubtype).To(BeTrue())
	}
}

func Test_FilterProducts_should_return_the_products_in_the_given_price_range_when_the_low_price_is_less_than_the_high_price(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{299, 400}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		priceFloat, _ := strconv.ParseFloat(item.PriceInUsd, 64)
		Expect(priceFloat).To(BeNumerically("<=", 400))
		Expect(priceFloat).To(BeNumerically(">=", 299))
	}
}

func Test_FilterProducts_should_return_bad_request_when_the_low_price_is_greater_than_the_high_price(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{400, 299}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_bad_request_when_the_low_price_is_negative(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{-1, 100}}

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_bad_request_when_the_high_price_is_negative(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, -100}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_bad_request_when_the_high_price_is_zero(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 0}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_bad_request_when_there_are_too_many_price_range_array_elements(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 10, 50}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_the_products_in_the_given_price_range_when_the_low_price_is_equal_to_the_high_price(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{299, 299}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		priceFloat, _ := strconv.ParseFloat(item.PriceInUsd, 64)
		Expect(priceFloat).To(Equal(float64(299)))
	}
}

func Test_FilterProducts_should_return_products_whose_models_start_with_the_specified_value(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	expectedModelPrefix := "RF"
	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 20000}, Model: expectedModelPrefix}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(strings.Index(item.Model, expectedModelPrefix)).To(BeZero()) // Make sure the model started with the value we expect
	}
}

func Test_FilterProducts_should_return_products_whose_manufacturers_start_with_the_specified_value(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	expectedManufacturerPrefix := "Sho"
	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 20000}, Manufacturer: expectedManufacturerPrefix}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(strings.Index(item.Manufacturer, expectedManufacturerPrefix)).To(BeZero()) // Make sure the manufacturer started with the value we expect
	}
}

func Test_FilterProducts_should_return_products_with_SNELL_certifications(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 20000}}
	request.Order.Field = "created_at_utc"
	request.Certifications.SNELL = true

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.Certifications.SNELL).To(BeTrue())
	}
}

func Test_FilterProducts_should_return_products_with_ECE_certifications(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 20000}}
	request.Order.Field = "created_at_utc"
	request.Certifications.ECE = true

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.Certifications.ECE).To(BeTrue())
	}
}

func Test_FilterProducts_should_return_products_with_DOT_certifications(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 20000}}
	request.Order.Field = "created_at_utc"
	request.Certifications.DOT = true

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.Certifications.DOT).To(BeTrue())
	}
}

func Test_FilterProducts_should_return_products_with_SHARP_certifications(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 20000}}
	request.Order.Field = "created_at_utc"
	request.Certifications.SHARP = &queries.SHARPCertificationQueryParams{}

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.Certifications.SHARP).ToNot(BeNil())
	}
}

func Test_FilterProducts_should_return_products_with_SHARP_certifications_and_minimum_impact_zones(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 20000}}
	request.Order.Field = "created_at_utc"
	request.Certifications.SHARP = &queries.SHARPCertificationQueryParams{}
	request.Certifications.SHARP.ImpactZoneMinimums.Left = 3
	request.Certifications.SHARP.ImpactZoneMinimums.Right = 3
	request.Certifications.SHARP.ImpactZoneMinimums.Rear = 3
	request.Certifications.SHARP.ImpactZoneMinimums.Top.Front = 3
	request.Certifications.SHARP.ImpactZoneMinimums.Top.Rear = 3

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.Certifications.SHARP).ToNot(BeNil())
		Expect(item.Certifications.SHARP.ImpactZoneRatings.Left).To(BeNumerically(">=", request.Certifications.SHARP.ImpactZoneMinimums.Left))
		Expect(item.Certifications.SHARP.ImpactZoneRatings.Right).To(BeNumerically(">=", request.Certifications.SHARP.ImpactZoneMinimums.Right))
		Expect(item.Certifications.SHARP.ImpactZoneRatings.Rear).To(BeNumerically(">=", request.Certifications.SHARP.ImpactZoneMinimums.Rear))
		Expect(item.Certifications.SHARP.ImpactZoneRatings.Top.Front).To(BeNumerically(">=", request.Certifications.SHARP.ImpactZoneMinimums.Top.Front))
		Expect(item.Certifications.SHARP.ImpactZoneRatings.Top.Rear).To(BeNumerically(">=", request.Certifications.SHARP.ImpactZoneMinimums.Top.Rear))
	}
}

func Test_FilterProducts_should_return_products_with_SHARP_certifications_and_minimum_stars(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 20000}}
	request.Order.Field = "created_at_utc"
	request.Certifications.SHARP = &queries.SHARPCertificationQueryParams{Stars: 3}

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.Certifications.SHARP).ToNot(BeNil())
		Expect(item.Certifications.SHARP.Stars).To(BeNumerically(">=", request.Certifications.SHARP.Stars))
	}
}

func Test_FilterProducts_should_correctly_page_through_the_resultset_when_start_and_limit_are_specified_and_there_are_no_filters_set(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 1, UsdPriceRange: []int{0, 10000}}
	request.Order.Field = "id"

	seeds := seeds.GetProductSeeds()
	for i := 0; i < len(seeds)+1; i++ {
		responseBody := &[]*entities.ProductDocument{}
		resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		if i < len(seeds) {
			Expect(*responseBody).To(HaveLen(1))
			Expect((*responseBody)[0]).To(BeEquivalentTo(seeds[i]))
		} else {
			Expect(*responseBody).To(HaveLen(0))
		}

		request.Start++
	}
}

func Test_FilterProducts_should_return_bad_request_when_the_limit_is_too_large(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 26, UsdPriceRange: []int{0, 10}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_bad_request_when_the_limit_is_too_small(t *testing.T) {
	t.Parallel()
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: -1, UsdPriceRange: []int{0, 10}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

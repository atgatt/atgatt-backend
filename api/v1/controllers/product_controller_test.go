package controllers_test

import (
	httpHelpers "atgatt-backend/common/http"
	"atgatt-backend/persistence/entities"
	"atgatt-backend/persistence/queries"
	"atgatt-backend/seeds"
	"fmt"
	"net/http"
	"strings"
	"testing"

	golinq "github.com/ahmetb/go-linq"
	. "github.com/onsi/gomega"
)

func Test_FilterProducts_should_return_all_of_the_products_data_when_the_limit_is_large_enough_and_there_are_no_optional_filters_set(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).To(BeEquivalentTo(seeds.GetProductSeeds()[0:25]))
}

func Test_FilterProducts_should_only_return_jackets_data_when_the_limit_is_large_enough_and_the_type_is_set_to_jacket(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}, Type: "jacket"}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.Type).To(Equal("jacket"))
	}
}

func Test_FilterProducts_should_only_return_chest_CE_level_2_jackets_when_the_filters_are_set_to_CE_level_2(t *testing.T) {
	RegisterTestingT(t)

	level2Filter := &queries.CEImpactZoneQueryParams{
		IsLevel2: true,
	}
	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}, Type: "jacket", JacketCertifications: &queries.JacketCertificationsQueryParams{
		Shoulder: level2Filter,
		Elbow:    level2Filter,
		Back:     level2Filter,
		Chest:    level2Filter,
	}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.Type).To(Equal("jacket"))
		Expect(item.JacketCertifications).ToNot(BeNil())
		Expect(item.JacketCertifications.Shoulder.IsLevel2).To(BeTrue())
		Expect(item.JacketCertifications.Elbow.IsLevel2).To(BeTrue())
		Expect(item.JacketCertifications.Back.IsLevel2).To(BeTrue())
		Expect(item.JacketCertifications.Chest.IsLevel2).To(BeTrue())
	}
}

func Test_FilterProducts_should_only_return_airbag_jackets_when_the_filters_are_set_to_require_airbags(t *testing.T) {
	RegisterTestingT(t)
	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}, Type: "jacket", JacketCertifications: &queries.JacketCertificationsQueryParams{
		FitsAirbag: true,
	}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.Type).To(Equal("jacket"))
		Expect(item.JacketCertifications).ToNot(BeNil())
		Expect(item.JacketCertifications.FitsAirbag).To(BeTrue())
	}
}

func Test_FilterProducts_should_return_all_of_the_products_that_have_the_given_subtype_when_the_subtypes_array_has_one_element(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "created_at_utc"
	request.Subtypes = []string{"full"}

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.Subtype).To(Equal(request.Subtypes[0]))
	}
}

func Test_FilterProducts_should_return_all_of_the_products_that_have_the_given_subtype_when_the_subtypes_array_has_multiple_elements(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "created_at_utc"
	request.Subtypes = []string{"full", "modular"}

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		isExpectedSubtype := item.Subtype == request.Subtypes[0] || item.Subtype == request.Subtypes[1]
		Expect(isExpectedSubtype).To(BeTrue())
	}
}

func Test_FilterProducts_should_return_the_products_in_the_given_price_range_when_the_low_price_is_less_than_the_high_price(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{29900, 40000}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.SearchPriceCents).To(BeNumerically("<=", 40000))
		Expect(item.SearchPriceCents).To(BeNumerically(">=", 29900))
	}
}

func Test_FilterProducts_should_return_bad_request_when_the_low_price_is_greater_than_the_high_price(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{40000, 29900}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_bad_request_when_the_low_price_is_negative(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{-1, 1000000}}

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_bad_request_when_the_high_price_is_negative(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, -1000000}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_bad_request_when_the_high_price_is_zero(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 0}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_bad_request_when_there_are_too_many_price_range_array_elements(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 100000, 500000}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_the_products_in_the_given_price_range_when_the_low_price_is_equal_to_the_high_price(t *testing.T) {
	RegisterTestingT(t)

	expectedExactPrice := 39900
	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{expectedExactPrice, expectedExactPrice}}
	request.Order.Field = "created_at_utc"
	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.SearchPriceCents).To(Equal(expectedExactPrice))
	}
}

func Test_FilterProducts_should_return_products_whose_models_or_aliases_start_with_the_specified_value(t *testing.T) {
	RegisterTestingT(t)

	expectedModelPrefix := "RF"
	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}, Model: expectedModelPrefix}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		isModelCorrect := strings.Index(item.Model, expectedModelPrefix) == 0
		isModelAliasCorrect := golinq.From(item.ModelAliases).AnyWithT(func(alias *entities.ProductModelAlias) bool {
			return strings.Index(alias.ModelAlias, expectedModelPrefix) == 0
		})

		Expect(isModelCorrect || isModelAliasCorrect).To(BeTrue()) // Make sure the model or the alias started with the value we expect
	}
}

func Test_FilterProducts_should_return_products_whose_manufacturers_start_with_the_specified_value(t *testing.T) {
	RegisterTestingT(t)

	expectedManufacturerPrefix := "Sho"
	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}, Manufacturer: expectedManufacturerPrefix}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(strings.Index(item.Manufacturer, expectedManufacturerPrefix)).To(BeZero()) // Make sure the manufacturer started with the value we expect
	}
}

func Test_FilterProducts_should_return_products_with_SNELL_certifications(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}, HelmetCertifications: &queries.HelmetCertificationsQueryParams{}}
	request.Order.Field = "created_at_utc"
	request.HelmetCertifications.SNELL = true

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.HelmetCertifications.SNELL).To(BeTrue())
	}
}

func Test_FilterProducts_should_return_products_with_ECE_certifications(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}, HelmetCertifications: &queries.HelmetCertificationsQueryParams{}}
	request.Order.Field = "created_at_utc"
	request.HelmetCertifications.ECE = true

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.HelmetCertifications.ECE).To(BeTrue())
	}
}

func Test_FilterProducts_should_return_products_with_DOT_certifications(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}, HelmetCertifications: &queries.HelmetCertificationsQueryParams{}}
	request.Order.Field = "created_at_utc"
	request.HelmetCertifications.DOT = true

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.HelmetCertifications.DOT).To(BeTrue())
	}
}

func Test_FilterProducts_should_return_only_the_current_products_when_exclude_discontinued_is_true(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "created_at_utc"
	request.ExcludeDiscontinued = true

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	foundDiscontinued := false
	foundCurrent := false
	for _, item := range *responseBody {
		if item.IsDiscontinued {
			foundDiscontinued = true
		} else {
			foundCurrent = true
		}
	}

	Expect(foundDiscontinued).To(BeFalse())
	Expect(foundCurrent).To(BeTrue())
}

func Test_FilterProducts_should_return_products_with_SHARP_certifications(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}, HelmetCertifications: &queries.HelmetCertificationsQueryParams{}}
	request.Order.Field = "created_at_utc"
	request.HelmetCertifications.SHARP = &queries.SHARPCertificationQueryParams{}

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.HelmetCertifications.SHARP).ToNot(BeNil())
	}
}

func Test_FilterProducts_should_return_products_with_SHARP_certifications_and_minimum_impact_zones(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}, HelmetCertifications: &queries.HelmetCertificationsQueryParams{}}
	request.Order.Field = "created_at_utc"
	request.HelmetCertifications.SHARP = &queries.SHARPCertificationQueryParams{}
	request.HelmetCertifications.SHARP.ImpactZoneMinimums.Left = 4
	request.HelmetCertifications.SHARP.ImpactZoneMinimums.Right = 3
	request.HelmetCertifications.SHARP.ImpactZoneMinimums.Rear = 3
	request.HelmetCertifications.SHARP.ImpactZoneMinimums.Top.Front = 3
	request.HelmetCertifications.SHARP.ImpactZoneMinimums.Top.Rear = 3

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.HelmetCertifications.SHARP).ToNot(BeNil())
		Expect(item.HelmetCertifications.SHARP.ImpactZoneRatings.Left).To(BeNumerically(">=", request.HelmetCertifications.SHARP.ImpactZoneMinimums.Left))
		Expect(item.HelmetCertifications.SHARP.ImpactZoneRatings.Right).To(BeNumerically(">=", request.HelmetCertifications.SHARP.ImpactZoneMinimums.Right))
		Expect(item.HelmetCertifications.SHARP.ImpactZoneRatings.Rear).To(BeNumerically(">=", request.HelmetCertifications.SHARP.ImpactZoneMinimums.Rear))
		Expect(item.HelmetCertifications.SHARP.ImpactZoneRatings.Top.Front).To(BeNumerically(">=", request.HelmetCertifications.SHARP.ImpactZoneMinimums.Top.Front))
		Expect(item.HelmetCertifications.SHARP.ImpactZoneRatings.Top.Rear).To(BeNumerically(">=", request.HelmetCertifications.SHARP.ImpactZoneMinimums.Top.Rear))
	}
}

func Test_FilterProducts_should_return_products_with_SHARP_certifications_and_minimum_stars(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}, HelmetCertifications: &queries.HelmetCertificationsQueryParams{}}
	request.Order.Field = "created_at_utc"
	request.HelmetCertifications.SHARP = &queries.SHARPCertificationQueryParams{Stars: 3}

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).ToNot(BeEmpty())
	for _, item := range *responseBody {
		Expect(item.HelmetCertifications.SHARP).ToNot(BeNil())
		Expect(item.HelmetCertifications.SHARP.Stars).To(BeNumerically(">=", request.HelmetCertifications.SHARP.Stars))
	}
}

func Test_FilterProducts_should_correctly_page_through_the_resultset_when_start_and_limit_are_specified_and_there_are_no_filters_set(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 1, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "id"

	seeds := seeds.GetProductSeeds()
	for i := 0; i < len(seeds)+1; i++ {
		responseBody := &[]*entities.Product{}
		resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

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

func Test_FilterProducts_should_return_bad_request_when_both_jacket_certifications_and_helmet_certifications_are_specified(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 1000}, HelmetCertifications: &queries.HelmetCertificationsQueryParams{}, JacketCertifications: &queries.JacketCertificationsQueryParams{}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_bad_request_when_the_limit_is_too_large(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 26, UsdPriceRange: []int{0, 1000}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_bad_request_when_the_limit_is_too_small(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: -1, UsdPriceRange: []int{0, 1000}}
	request.Order.Field = "created_at_utc"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_return_bad_request_when_ordering_by_an_unknown_field(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "yolo swag"

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_FilterProducts_should_be_able_to_order_by_the_search_price_cents(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "document->>'searchPriceCents'"
	request.Order.Descending = true

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

func Test_FilterProducts_should_be_able_to_order_by_the_manufacturer(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "document->>'manufacturer'"
	request.Order.Descending = true

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

func Test_FilterProducts_should_be_able_to_order_by_the_model(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "document->>'manufacturer'"
	request.Order.Descending = true

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

func Test_FilterProducts_should_be_able_to_order_by_the_safety_percentage(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "document->>'safetyPercentage'"
	request.Order.Descending = true

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect((*responseBody)[0].SafetyPercentage).To(Equal(100))
}

func Test_FilterProducts_should_be_able_to_order_by_the_utc_created_date(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "created_at_utc"
	request.Order.Descending = true

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

func Test_FilterProducts_should_be_able_to_order_by_the_utc_updated_date(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "updated_at_utc"
	request.Order.Descending = true

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

func Test_FilterProducts_should_be_able_to_order_by_id(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Start: 0, Limit: 25, UsdPriceRange: []int{0, 2000000}}
	request.Order.Field = "id"
	request.Order.Descending = true

	responseBody := &[]*entities.Product{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/products/filter", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

func Test_GetProductDetails_should_return_the_product_details_when_the_UUID_is_valid(t *testing.T) {
	RegisterTestingT(t)

	expectedProduct := seeds.GetProductSeeds()[0]

	responseBody := &entities.Product{}
	resp, err := httpHelpers.MakeJSONGETRequest(fmt.Sprintf("%s/v1/products/%s", APIBaseURL, expectedProduct.UUID.String()), responseBody)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	Expect(responseBody).To(Equal(expectedProduct))
}

func Test_GetProductDetails_NotFound(t *testing.T) {
	RegisterTestingT(t)

	responseBody := &entities.Product{}
	resp, err := httpHelpers.MakeJSONGETRequest(fmt.Sprintf("%s/v1/products/1234", APIBaseURL), responseBody)
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
}

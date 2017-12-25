package handlers

import (
	"crashtested-backend/api/requests/helpers"
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/queries"
	"crashtested-backend/seeds"
	"fmt"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_FilterProducts_should_return_all_of_the_seed_data_when_the_limit_is_large_enough(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Manufacturer: "Shoei", Model: "Hey bay", MinimumSHARPStars: 1, UsdPriceRange: [2]int{0, 10000}, Start: 0, Limit: 25}
	request.Certifications.SHARP = true
	request.Certifications.SNELL = true
	request.Certifications.ECE = true
	request.Certifications.DOT = true
	request.ImpactZoneMinimums.Left = 2
	request.ImpactZoneMinimums.Right = 2
	request.ImpactZoneMinimums.Top.Front = 2
	request.ImpactZoneMinimums.Top.Rear = 2
	request.ImpactZoneMinimums.Rear = 6

	responseBody := &[]*entities.ProductDocument{}
	resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	Expect(*responseBody).To(BeEquivalentTo(seeds.GetProductSeeds()))
}

func Test_FilterProducts_should_correctly_page_through_the_resultset_when_start_and_limit_are_specified(t *testing.T) {
	RegisterTestingT(t)

	request := &queries.FilterProductsQuery{Manufacturer: "Shoei", Model: "Hey bay", MinimumSHARPStars: 1, UsdPriceRange: [2]int{0, 10000}, Start: 0, Limit: 1}
	request.Certifications.SHARP = true
	request.Certifications.SNELL = true
	request.Certifications.ECE = true
	request.Certifications.DOT = true
	request.ImpactZoneMinimums.Left = 2
	request.ImpactZoneMinimums.Right = 2
	request.ImpactZoneMinimums.Top.Front = 2
	request.ImpactZoneMinimums.Top.Rear = 2
	request.ImpactZoneMinimums.Rear = 6

	for i := 0; i < 5; i++ {
		responseBody := &[]*entities.ProductDocument{}
		resp, err := helpers.MakeJsonPOSTRequest(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), request, responseBody)

		Expect(err).To(BeNil())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		if i < 4 {
			Expect(*responseBody).To(HaveLen(1))
		} else {
			Expect(*responseBody).To(HaveLen(0))
		}

		request.Start++
	}
}

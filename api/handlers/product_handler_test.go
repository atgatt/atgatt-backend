package handlers

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/seeds"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_FilterProducts_should_always_return_the_seed_data(t *testing.T) {
	RegisterTestingT(t)

	resp, err := http.Post(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), "application/json", strings.NewReader(`{
		"manufacturer": "Shoei",
		"model": "Hey bay",
		"certifications": {
		   "SHARP": true,
		   "SNELL": true,
		   "ECE": true,
		   "DOT": true
		},
		"minimumSHARPStars": 1,
		"impactZoneMinimums": {
		  "left": 2,
		  "right": 3,
		  "top": {
			"front": 4,
			"rear": 5
		  },
		  "rear": 6
		},
		"usdPriceRange": [0, 10001],
		"start": 0,
		"limit": 25
	}`))

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	responseBodyBytes, _ := ioutil.ReadAll(resp.Body)
	productsArrayPtr := &[]*entities.ProductDocument{}
	json.Unmarshal(responseBodyBytes, productsArrayPtr)
	products := *productsArrayPtr
	Expect(products).To(BeEquivalentTo(seeds.GetProductSeeds()))
}

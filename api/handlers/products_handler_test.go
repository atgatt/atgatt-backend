package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_FilterProducts_should_always_return_the_mock_data(t *testing.T) {
	RegisterTestingT(t)

	resp, _ := http.Post(fmt.Sprintf("%s/v1/products/filter", ApiBaseUrl), "application/json", strings.NewReader(`{
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
		"usdPriceRange": [0, 10000],
		"start": 0,
		"limit": 25
	}`))

	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

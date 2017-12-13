package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	. "github.com/onsi/gomega"
)

func Test_FilterProducts_should_always_return_the_mock_data(t *testing.T) {
	RegisterTestingT(t)

	e := echo.New()
	request := httptest.NewRequest(echo.POST, "/api/v1/products/filter", strings.NewReader(`{
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
		"usdPriceRange": [0, 10000]
	}`))
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	context.SetPath("/api/v1/products/filter")

	handler := &ProductsHandler{}
	handler.FilterProducts(context)

	Expect(recorder.Code).To(Equal(http.StatusOK))
}

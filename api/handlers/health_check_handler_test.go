package handlers

import (
	"crashtested-backend/api/responses"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	. "github.com/onsi/gomega"
)

func Test_Healthcheck_should_always_return_the_version_of_the_app(t *testing.T) {
	RegisterTestingT(t)

	e := echo.New()
	request := httptest.NewRequest(echo.GET, "/", nil)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	context.SetPath("/")

	handler := &HealthCheckHandler{BuildNumber: "2"}
	handler.Healthcheck(context)

	responseBytes := recorder.Body.Bytes()
	healthCheckResponse := &responses.HealthCheckResponse{}
	json.Unmarshal(responseBytes, healthCheckResponse)

	Expect(recorder.Code).To(Equal(http.StatusOK))
	Expect(healthCheckResponse).ToNot(BeNil())
	Expect(healthCheckResponse.BuildNumber).To(Equal("2"))
	Expect(healthCheckResponse.Name).To(Equal("crashtested-api"))
	Expect(healthCheckResponse.Version).To(Equal("1.0.0"))
}

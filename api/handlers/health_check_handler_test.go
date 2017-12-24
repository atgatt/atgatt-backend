package handlers

import (
	"crashtested-backend/api/responses"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_Healthcheck_should_always_return_the_name_and_version_of_the_app(t *testing.T) {
	RegisterTestingT(t)

	resp, _ := http.Get(ApiBaseUrl)
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	responseBodyBytes, _ := ioutil.ReadAll(resp.Body)
	healthCheckResponse := new(responses.HealthCheckResponse)
	json.Unmarshal(responseBodyBytes, healthCheckResponse)

	Expect(healthCheckResponse).ToNot(BeNil())
	Expect(healthCheckResponse.Name).To(Equal("crashtested-api"))
	Expect(healthCheckResponse.Version).To(Equal("integration-tests-version"))
	Expect(healthCheckResponse.BuildNumber).To(Equal("integration-tests-build"))
}

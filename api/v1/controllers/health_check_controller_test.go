package controllers_test

import (
	"atgatt-backend/api/v1/responses"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_Healthcheck_should_return_the_name_and_version_of_the_app_when_a_GET_request_is_sent(t *testing.T) {
	RegisterTestingT(t)

	resp, _ := http.Get(APIBaseURL)
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	responseBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		defer resp.Body.Close()
	}
	healthCheckResponse := new(responses.HealthCheckResponse)
	json.Unmarshal(responseBodyBytes, healthCheckResponse)

	Expect(healthCheckResponse).ToNot(BeNil())
	Expect(healthCheckResponse.Name).To(Equal("atgatt-api"))
	Expect(healthCheckResponse.Version).To(Equal("integration-tests-version"))
	Expect(healthCheckResponse.BuildNumber).To(Equal("integration-tests-build"))
	Expect(healthCheckResponse.CommitHash).To(Equal("integration-tests-commit"))
	Expect(healthCheckResponse.Database.CurrentVersion).ToNot(BeEmpty())
}

func Test_Healthcheck_should_return_an_empty_body_when_a_HEAD_request_is_sent(t *testing.T) {
	RegisterTestingT(t)

	resp, _ := http.Head(APIBaseURL)
	Expect(resp.StatusCode).To(Equal(http.StatusOK))

	responseBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		defer resp.Body.Close()
	}
	Expect(responseBodyBytes).To(HaveLen(0))
}

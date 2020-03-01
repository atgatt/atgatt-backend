package controllers_test

import (
	"atgatt-backend/api/v1/requests"
	"atgatt-backend/api/v1/responses"
	httpHelpers "atgatt-backend/common/http"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	. "github.com/onsi/gomega"
)

func Test_CreateMarketingEmail_should_insert_a_new_marketing_email_when_the_email_does_not_exist_and_is_valid(t *testing.T) {
	RegisterTestingT(t)

	request := &requests.CreateMarketingEmailRequest{Email: uuid.New().String() + "@gmail.com"}

	responseBody := ""
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/marketing/email", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

func Test_CreateMarketingEmail_should_return_bad_request_when_the_email_already_exists(t *testing.T) {
	RegisterTestingT(t)

	request := &requests.CreateMarketingEmailRequest{Email: "someexistingemail@gmail.com"}

	responseBody := ""
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/marketing/email", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_CreateMarketingEmail_should_return_bad_request_when_the_lowercase_email_already_exists(t *testing.T) {
	RegisterTestingT(t)

	request := &requests.CreateMarketingEmailRequest{Email: "SOMEexistingEMAIL@GmAiL.COM"}

	responseBody := &responses.Response{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/marketing/email", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

func Test_CreateMarketingEmail_should_return_bad_request_when_the_email_is_invalid(t *testing.T) {
	RegisterTestingT(t)

	request := &requests.CreateMarketingEmailRequest{Email: "Sasdfnjkxj321905-"}

	responseBody := &responses.Response{}
	resp, err := httpHelpers.MakeJSONPOSTRequest(fmt.Sprintf("%s/v1/marketing/email", APIBaseURL), request, responseBody)

	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
}

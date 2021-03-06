package helpers_test

import (
	helpers "atgatt-backend/common/auth"
	"testing"

	. "github.com/onsi/gomega"
)

func Test_GetAuth0PublicKey_returns_a_non_empty_key_for_all_expected_domains(t *testing.T) {
	RegisterTestingT(t)

	expectedDomains := []string{"atgatt.auth0.com", "atgatt-staging.auth0.com"}
	for _, domain := range expectedDomains {
		key, err := helpers.GetAuth0PublicKey(domain)
		Expect(err).To(BeNil())
		Expect(key).ToNot(BeNil())
	}
}

func Test_GetAuth0PublicKey_returns_an_error_for_an_unexpected_domain(t *testing.T) {
	RegisterTestingT(t)
	key, err := helpers.GetAuth0PublicKey("httpbin.org")
	Expect(err).ToNot(BeNil())
	Expect(key).To(BeNil())
}

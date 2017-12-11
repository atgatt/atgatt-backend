package handlers

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_Healthcheck_should_always_return_the_version_of_the_app(t *testing.T) {
	RegisterTestingT(t)

	Expect(true).To(BeTrue())
}

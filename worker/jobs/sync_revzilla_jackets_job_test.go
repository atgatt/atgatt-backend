package jobs_test

import (
	"net/http"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	. "github.com/onsi/gomega"
)

func Test_sync_revzilla_jackets_should_return_OK(t *testing.T) {
	RegisterTestingT(t)

	resp, err := http.Post(APIBaseURL+"/jobs/sync_revzilla_jackets", "application/json", strings.NewReader("{}"))
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
}

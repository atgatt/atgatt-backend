package helpers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// MakeJSONPOSTRequest makes a request to url and maps the JSON response to the given interface type
func MakeJSONPOSTRequest(url string, request interface{}, response interface{}) (*http.Response, error) {
	requestBytes, marshalErr := json.Marshal(request)
	if marshalErr != nil {
		return nil, marshalErr
	}

	requestString := string(requestBytes)
	logrus.Infof("Making JSON POST request with data: %s", requestString)
	resp, postErr := http.Post(url, "application/json", strings.NewReader(requestString))
	if postErr != nil {
		return nil, postErr
	}
	responseBodyBytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(responseBodyBytes, response)

	return resp, nil
}

// MakeFormPOSTRequest makes a request to the given url with a supplied set of urlencoded form values and returns the response body as a string.
func MakeFormPOSTRequest(url string, formValues url.Values) (string, error) {
	logrus.WithField("formValues", formValues).Info("Making JSON POST form request")
	resp, postErr := http.PostForm(url, formValues)
	if postErr != nil {
		return "", postErr
	}
	responseBodyBytes, _ := ioutil.ReadAll(resp.Body)

	return string(responseBodyBytes), nil
}

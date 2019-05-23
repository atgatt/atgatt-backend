package helpers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// GetContentsAtURL makes a request to url and returns the response content as a string
func GetContentsAtURL(url string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.157 Safari/537.36")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	responseBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(responseBodyBytes), nil
}

// MakeJSONPOSTRequest makes a request to url and maps the JSON response to the given interface type
func MakeJSONPOSTRequest(url string, request interface{}, response interface{}) (*http.Response, error) {
	requestBytes, marshalErr := json.Marshal(request)
	if marshalErr != nil {
		return nil, marshalErr
	}

	requestString := string(requestBytes)
	resp, postErr := http.Post(url, "application/json", strings.NewReader(requestString))
	if postErr != nil {
		return nil, postErr
	}
	defer resp.Body.Close()
	responseBodyBytes, _ := ioutil.ReadAll(resp.Body)

	json.Unmarshal(responseBodyBytes, response)
	return resp, nil
}

// MakeFormPOSTRequest makes a request to the given url with a supplied set of urlencoded form values and returns the response body as a string.
func MakeFormPOSTRequest(url string, formValues url.Values) (string, error) {
	resp, postErr := http.PostForm(url, formValues)
	if postErr != nil {
		return "", postErr
	}
	defer resp.Body.Close()
	responseBodyBytes, _ := ioutil.ReadAll(resp.Body)

	return string(responseBodyBytes), nil
}

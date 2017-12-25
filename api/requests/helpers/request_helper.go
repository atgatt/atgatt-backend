package helpers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

func MakeJsonPOSTRequest(url string, request interface{}, response interface{}) (*http.Response, error) {
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

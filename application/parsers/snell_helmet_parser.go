package parsers

import (
	"atgatt-backend/persistence/entities"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

// SNELLHelmetParser contains functions used to retrieve SNELL's helmet data via their JSON API
type SNELLHelmetParser struct {
}

// SNELLHelmetsResponse represents the data returned from SNELL's JSON API
type SNELLHelmetsResponse struct {
	Data []*entities.SNELLHelmet `json:"data"`
}

// GetAllByCertification returns the list of SNELL helmets that are certified to the given standard.
func (r *SNELLHelmetParser) GetAllByCertification(standard string) ([]*entities.SNELLHelmet, error) {
	if standard == "" {
		return nil, errors.New("The standard cannot be empty")
	}

	resp, err := http.Get("http://snell.us.com/codefolder/datatable.php")
	if err != nil {
		return nil, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	snellHelmetsResponse := &SNELLHelmetsResponse{}
	err = json.Unmarshal(respBytes, snellHelmetsResponse)
	if err != nil {
		return nil, err
	}

	numHelmets := len(snellHelmetsResponse.Data)
	if numHelmets < 100 {
		return nil, errors.New("Did not receive enough SNELL helmets")
	}

	filteredHelmets := []*entities.SNELLHelmet{}
	for _, rawHelmet := range snellHelmetsResponse.Data {
		if strings.EqualFold(strings.TrimSpace(rawHelmet.Standard), standard) {
			subtype := ""
			if strings.EqualFold(rawHelmet.FaceConfig, "modular") {
				subtype = "modular"
			} else if strings.EqualFold(rawHelmet.FaceConfig, "full face") {
				subtype = "full"
			} else if strings.EqualFold(rawHelmet.FaceConfig, "open face") {
				subtype = "open"
			}

			rawHelmet.FaceConfig = subtype
			filteredHelmets = append(filteredHelmets, rawHelmet)
		}
	}
	return filteredHelmets, nil
}

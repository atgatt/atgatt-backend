package repositories

import (
	"crashtested-backend/persistence/entities"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type SNELLHelmetRepository struct {
}

type SNELLHelmetsResponse struct {
	// {"manufacturer":"First Line Industrial Inc.","model":"Fcarbon-17","size":"M, XL","standard":"B-90A","helmettype":"BIKE","faceconfig":"Full Face"}
	Data []*entities.SNELLHelmet `json:"data"`
}

func (self *SNELLHelmetRepository) GetAllByCertification(standard string) ([]*entities.SNELLHelmet, error) {
	resp, err := http.Get("http://snell.us.com/codefolder/datatable.php")
	if err != nil {
		return nil, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	snellHelmetsResponse := &SNELLHelmetsResponse{}
	err = json.Unmarshal(respBytes, snellHelmetsResponse)
	if err != nil {
		return nil, err
	}

	numHelmets := len(snellHelmetsResponse.Data)
	if numHelmets < 100 {
		return nil, errors.New("Did not receive enough SNELL helmets")
	}

	filteredHelmets := make([]*entities.SNELLHelmet, 0)
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

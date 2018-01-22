package repositories

import (
	"crashtested-backend/persistence/entities"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type SNELLHelmetRepository struct {
}

type SNELLHelmetsResponse struct {
	// {"manufacturer":"First Line Industrial Inc.","model":"Fcarbon-17","size":"M, XL","standard":"B-90A","helmettype":"BIKE","faceconfig":"Full Face"}
	Data []*entities.SNELLHelmet `json:"data"`
}

func (self *SNELLHelmetRepository) GetAllHelmets() ([]*entities.SNELLHelmet, error) {
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
	return snellHelmetsResponse.Data, nil
}

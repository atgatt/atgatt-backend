package entities

import (
	"encoding/xml"
)

// CJProductsResponseWrapper represents a wrapper around the list of all the products returned from the Commission Junction API
type CJProductsResponseWrapper struct {
	XMLName  xml.Name          `xml:"cj-api"`
	Products CJProductsWrapper `xml:"products"`
}

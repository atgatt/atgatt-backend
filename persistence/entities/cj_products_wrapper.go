package entities

import "encoding/xml"

// CJProductsWrapper represents a product returned from the Commission Junction API
type CJProductsWrapper struct {
	XMLName xml.Name    `xml:"products"`
	Data    []CJProduct `xml:"product"`
}

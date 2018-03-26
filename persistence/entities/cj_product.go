package entities

import (
	"encoding/xml"
	"strings"
)

// CJProduct represents a product returned from the Commission Junction API
type CJProduct struct {
	XMLName     xml.Name `xml:"product"`
	BuyURL      string   `xml:"buy-url"`
	Name        string   `xml:"name"`
	Price       float64  `xml:"price"`
	ImageURL    string   `xml:"image-url"`
	Category    string   `xml:"advertiser-category"`
	Description string   `xml:"description"`
}

// IsHelmet returns true when the product has the Motorcycle Helmets category
func (p *CJProduct) IsHelmet() bool {
	return strings.HasSuffix(p.Category, "Motorcycle Helmets")
}

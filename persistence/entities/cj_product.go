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

// IsHelmet returns true when the product category contains helmet and the name of the product does not contain shield/spoiler
func (p *CJProduct) IsHelmet() bool {
	lowercaseName := strings.ToLower(p.Name)
	lowercaseCategory := strings.ToLower(p.Category)
	return strings.Contains(lowercaseCategory, "helmet") && !strings.Contains(lowercaseName, "spoiler") && !strings.Contains(lowercaseName, "shield")
}

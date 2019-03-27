package entities

import (
	"regexp"
	"strconv"
	"strings"
)

var descriptionSeparatorRegexp = regexp.MustCompile("[\n\t]+")

// RevzillaProduct represents a product found by scraping RevZilla.com
type RevzillaProduct struct {
	ID               string
	URL              string
	Brand            string
	Name             string
	Price            string
	PriceCurrency    string
	ImageURL         string
	DescriptionParts []string
}

// GetModel parses the model from the title by replacing the brand with an empty string i.e. "Dainese HF-D1 Jacket" becomes "HF-D1 Jacket"
func (r RevzillaProduct) GetModel() string {
	return strings.TrimSpace(strings.Replace(r.Name, r.Brand, "", 1))
}

// GetPriceCents converts the Price, represented as a float-string, to an integer number of cents.
func (r RevzillaProduct) GetPriceCents() int {
	priceFloat, _ := strconv.ParseFloat(r.Price, 64)
	return int(priceFloat * float64(100))
}

// GetCertifications calculates the CE certifications of this product by scraping the Description for key terms
func (r RevzillaProduct) GetCertifications() int {
	descriptionSeparatorRegexp.Split(r.DescriptionParts[0], -1)
	return 0
}

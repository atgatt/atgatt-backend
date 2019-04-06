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

// IsDiscontinued returns true if the product doesn't have the expected summary section that indicates that it is a real product
func (r RevzillaProduct) IsDiscontinued(buyURLContents string) bool {
	return !strings.Contains(strings.ToLower(buyURLContents), "product-show-summary")
}

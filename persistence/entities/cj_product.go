package entities

import (
	"strconv"
	"strings"
)

// CJProduct represents a product returned from the Commission Junction API
type CJProduct struct {
	LinkCode struct {
		ClickURL string `json:"clickUrl"`
	} `json:"linkCode"`
	Name  string `json:"title"`
	Price struct {
		Amount string `json:"amount"`
	} `json:"price"`
	ImageURL    string   `json:"imageLink"`
	Categories  []string `json:"productType"`
	Description string   `json:"description"`
}

// IsHelmet returns true when the product category contains helmet and the name of the product does not contain shield/spoiler
func (p CJProduct) IsHelmet() bool {
	lowercaseName := strings.ToLower(p.Name)
	lowercaseCategory := strings.ToLower(p.getCombinedCategory())
	return strings.Contains(lowercaseCategory, "helmet") && !strings.Contains(lowercaseName, "spoiler") && !strings.Contains(lowercaseName, "shield")
}

func (p CJProduct) getCombinedCategory() string {
	return strings.Join(p.Categories, " ")
}

// GetPrice returns the price.amount formatted as a float
func (p CJProduct) GetPrice() float64 {
	res, err := strconv.ParseFloat(p.Price.Amount, 64)
	if err != nil {
		return 0.0
	}
	return res
}

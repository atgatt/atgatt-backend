package entities

import (
	"github.com/google/uuid"
)

// ProductDocument represents a safety product such as a motorcycle helmet, jacket, etc. It contains the price of the product, certifications, etc.
type ProductDocument struct {
	UUID                uuid.UUID `json:"uuid"`
	Type                string    `json:"type"`
	Subtype             string    `json:"subtype"`
	Manufacturer        string    `json:"manufacturer"`
	Model               string    `json:"model"`
	ModelAlias          string    `json:"modelAlias"`
	ImageURL            string    `json:"imageUrl"`
	BuyURL              string    `json:"buyUrl"`
	PriceInUSDMultiple  int       `json:"priceInUsdMultiple"`
	LatchPercentage     int       `json:"latchPercentage"`
	WeightInLbsMultiple int       `json:"weightInLbsMultiple"`
	Sizes               []string  `json:"sizes"`
	Materials           string    `json:"materials"`
	RetentionSystem     string    `json:"retentionSystem"`
	Certifications      struct {
		SHARP *SHARPCertificationDocument `json:"SHARP"`
		SNELL bool                        `json:"SNELL"`
		ECE   bool                        `json:"ECE"`
		DOT   bool                        `json:"DOT"`
	} `json:"certifications"`
}

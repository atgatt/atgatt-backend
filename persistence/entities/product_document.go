package entities

import (
	"github.com/google/uuid"
)

type ProductDocument struct {
	UUID            uuid.UUID `json:"uuid"`
	Type            string    `json:"type"`
	Subtype         string    `json:"subtype"`
	Manufacturer    string    `json:"manufacturer"`
	Model           string    `json:"model"`
	ModelAlias      string    `json:"modelAlias"`
	ImageURL        string    `json:"imageUrl"`
	PriceInUsd      string    `json:"priceInUsd"`
	LatchPercentage int       `json:"latchPercentage"`
	WeightInLbs     float64   `json:"weightInLbs"`
	Sizes           []string  `json:"sizes"`
	Materials       string    `json:"materials"`
	RetentionSystem string    `json:"retentionSystem"`
	Certifications  struct {
		SHARP *SHARPCertificationDocument `json:"SHARP"`
		SNELL bool                        `json:"SNELL"`
		ECE   bool                        `json:"ECE"`
		DOT   bool                        `json:"DOT"`
	} `json:"certifications"`
}

package entities

import (
	"github.com/google/uuid"
)

type ProductDocument struct {
	UUID           uuid.UUID `json:"uuid"`
	Type           string    `json:"type"`
	Subtype        string    `json:"subtype"`
	Manufacturer   string    `json:"manufacturer"`
	Model          string    `json:"model"`
	ImageURL       string    `json:"imageUrl"`
	PriceInUsd     string    `json:"priceInUsd"`
	Certifications struct {
		SHARP struct {
			Stars             int `json:"ratingValue"`
			ImpactZoneRatings struct {
				Left  int `json:"left"`
				Right int `json:"right"`
				Top   struct {
					Front int `json:"front"`
					Rear  int `json:"rear"`
				} `json:"top"`
				Rear int `json:"rear"`
			} `json:"impactZoneRatings"`
		} `json:"SHARP"`
		SNELL bool `json:"SNELL"`
		ECE   bool `json:"ECE"`
		DOT   bool `json:"DOT"`
	} `json:"certifications"`
	Score string `json:"score"`
}

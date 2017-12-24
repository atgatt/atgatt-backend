package entities

import (
	"github.com/google/uuid"
)

type ProductDocument struct {
	UUID           uuid.UUID
	Type           string
	Subtype        string
	Manufacturer   string
	Model          string
	ImageURL       string
	PriceInUsd     string
	Certifications struct {
		SHARP struct {
			Stars             int
			ImpactZoneRatings struct {
				Left  int
				Right int
				Top   struct {
					Front int
					Rear  int
				}
				Rear int
			}
		}
		SNELL bool
		ECE   bool
		DOT   bool
	}
}

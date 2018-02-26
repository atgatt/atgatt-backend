package entities

import (
	"math"

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
	SafetyPercentage    int       `json:"safetyPercentage"`
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

const sharpImpactWeight float64 = 0.2
const sharpImpactMaxValue float64 = 5.0

// CalculateSafetyPercentage calculates how safe a helmet is based on a weighted average of all of its certifications, rounded up to the nearest integer
// Formula: 0.10 * SNELL + 0.08 * ECE + 0.02 * DOT + 0.7 * (SHARP-Left-Percentage * 0.20 + SHARP-Right-Percentage * 0.20 + SHARP-TopFront-Percentage * 0.20 + SHARP-TopRear-Percentage * 0.20 + SHARP-Rear-Percentage * 0.20)
// SHARP Percentages are calculated by dividing the raw score by the maximum score (i.e. Raw-Score / 5)
// TODO: Make this support multiple product types once gloves, boots, jackets, etc are added
func (p ProductDocument) CalculateSafetyPercentage() int {
	var totalScore float64

	// SNELL is rated slightly higher than ECE or DOT because they're an independent testing agency and publish their results online, but they don't have detailed enough crash test ratings and use manufacturer-supplied helmets
	if p.Certifications.SNELL {
		totalScore += 0.10
	}

	// ECE is the minimum standard required for helmet use in the EU, and helmets must be proven to meet this standard before being sold (not based on the honor system!)
	if p.Certifications.ECE {
		totalScore += 0.08
	}

	// DOT is pretty much useless since it's based off the honor system, hence a very low weight
	if p.Certifications.DOT {
		totalScore += 0.02
	}

	// SHARP is weighted the highest because while they are similar to SHARP, they also provide detailed crash test ratings for each helmet and buy helmets off the shelf instead of getting samples from manufacturers directly
	if p.Certifications.SHARP != nil {
		sharpImpacts := p.Certifications.SHARP.ImpactZoneRatings
		totalScore += float64(0.8) * (sharpImpactWeight*(float64(sharpImpacts.Left)/sharpImpactMaxValue) +
			sharpImpactWeight*(float64(sharpImpacts.Right)/sharpImpactMaxValue) +
			sharpImpactWeight*(float64(sharpImpacts.Top.Front)/sharpImpactMaxValue) +
			sharpImpactWeight*(float64(sharpImpacts.Top.Rear)/sharpImpactMaxValue) +
			sharpImpactWeight*(float64(sharpImpacts.Rear)/sharpImpactMaxValue))
	}

	return int(math.Round(totalScore * 100))
}

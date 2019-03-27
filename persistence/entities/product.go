package entities

import (
	"math"
	"strings"

	"github.com/google/uuid"
)

// Product represents a safety product such as a motorcycle helmet, jacket, etc. It contains the price of the product, certifications, etc.
type Product struct {
	UUID                 uuid.UUID            `json:"uuid"`
	ExternalID           string               `json:"externalID"`
	Type                 string               `json:"type"`
	Subtype              string               `json:"subtype"`
	Manufacturer         string               `json:"manufacturer"`
	Model                string               `json:"model"`
	ModelAliases         []*ProductModelAlias `json:"modelAliases"`
	SafetyPercentage     int                  `json:"safetyPercentage"`
	OriginalImageURL     string               `json:"originalImageURL"`
	ImageKey             string               `json:"imageKey"`
	RevzillaBuyURL       string               `json:"revzillaBuyURL"`
	RevzillaPriceCents   int                  `json:"revzillaPriceCents"`
	MSRPCents            int                  `json:"msrpCents"`
	SearchPriceCents     int                  `json:"searchPriceCents"`
	LatchPercentage      int                  `json:"latchPercentage"`
	WeightInLbs          float64              `json:"weightInLbs"`
	Sizes                []string             `json:"sizes"`
	Materials            string               `json:"materials"`
	RetentionSystem      string               `json:"retentionSystem"`
	HelmetCertifications struct {
		SHARP *SHARPCertificationDocument `json:"SHARP"`
		SNELL bool                        `json:"SNELL"`
		ECE   bool                        `json:"ECE"`
		DOT   bool                        `json:"DOT"`
	} `json:"helmetCertifications"`
	JacketCertifications struct {
		Shoulder *CECertification `json:"shoulder"`
		Elbow    *CECertification `json:"elbow"`
		Back     *CECertification `json:"back"`
		Chest    *CECertification `json:"chest"`
	} `json:"jacketCertifications"`
	IsDiscontinued bool `json:"isDiscontinued"`
}

const sharpImpactWeight float64 = 0.2
const sharpImpactMaxValue float64 = 5.0

const defaultSNELLWeight float64 = 0.10
const defaultECEWeight float64 = 0.08
const defaultDOTWeight float64 = 0.02

// UpdateSearchPrice sets the search price to the revzilla price if its defined, otherwise uses the MSRP
func (p *Product) UpdateSearchPrice() {
	if p.RevzillaPriceCents > 0 {
		p.SearchPriceCents = p.RevzillaPriceCents
	} else {
		p.SearchPriceCents = p.MSRPCents
	}
}

// UpdateCertificationsByDescription updates the DOT and/or ECE certifications if the given description contains certain keywords indicating that the product has said certifications and returns booleans indicating whether or not updates occurred.
func (p *Product) UpdateCertificationsByDescription(productDescription string) (bool, bool) {
	lowerDescription := strings.ToLower(productDescription)

	// DOT and ECE are only 3 letters and are very common substrings, so it's better to use the real description and compare against that (the lowercase description probably has "dot" and "ece" in various words)
	containsDOT := strings.Contains(productDescription, "DOT") || strings.Contains(productDescription, "D.O.T")
	containsECE := strings.Contains(productDescription, "ECE") || strings.Contains(productDescription, "22/05") || strings.Contains(productDescription, "22.05")

	// SNELL is an "uncommon" enough substring that it's better to use the lower description
	containsSNELL := strings.Contains(lowerDescription, "snell") || strings.Contains(lowerDescription, "m2010") || strings.Contains(lowerDescription, "m2015")

	hasNewDOTCertification := false
	hasNewECECertification := false

	// SNELL certification implies DOT certification, so check for either cert. Do not update the SNELL cert as we only want to pull SNELL data from the official source (snell.us.com)
	if !p.HelmetCertifications.DOT && (containsDOT || containsSNELL) {
		p.HelmetCertifications.DOT = true
		hasNewDOTCertification = true
	}

	if !p.HelmetCertifications.ECE && containsECE {
		p.HelmetCertifications.ECE = true
		hasNewECECertification = true
	}

	return hasNewDOTCertification, hasNewECECertification
}

// UpdateSafetyPercentage calculates how safe a helmet is based on a weighted average of all of its certifications, rounded up to the nearest integer.
// SHARP Percentages are calculated by dividing the raw score by the maximum score (i.e. Raw-Score / 5)
// TODO: Make this support multiple product types once gloves, boots, jackets, etc are added
func (p *Product) UpdateSafetyPercentage() {
	var totalScore float64

	snellWeightToUse := defaultSNELLWeight
	eceWeightToUse := defaultECEWeight
	dotWeightToUse := defaultDOTWeight

	// SHARP is weighted the highest because while they are similar to SHARP, they also provide detailed crash test ratings for each helmet and buy helmets off the shelf instead of getting samples from manufacturers directly
	if p.HelmetCertifications.SHARP != nil {
		sharpImpacts := p.HelmetCertifications.SHARP.ImpactZoneRatings
		totalScore += float64(0.8) * (sharpImpactWeight*(float64(sharpImpacts.Left)/sharpImpactMaxValue) +
			sharpImpactWeight*(float64(sharpImpacts.Right)/sharpImpactMaxValue) +
			sharpImpactWeight*(float64(sharpImpacts.Top.Front)/sharpImpactMaxValue) +
			sharpImpactWeight*(float64(sharpImpacts.Top.Rear)/sharpImpactMaxValue) +
			sharpImpactWeight*(float64(sharpImpacts.Rear)/sharpImpactMaxValue))
	} else {
		// If SHARP hasn't rated the helmet yet, adjust the weights, but penalize this helmet by 20% (helmets w/o SHARP should never be able to acheive a 100% score)
		snellWeightToUse = 0.65
		eceWeightToUse = 0.1
		dotWeightToUse = 0.05
	}

	// SNELL is rated slightly higher than ECE or DOT because they're an independent testing agency and publish their results online, but they don't have detailed enough crash test ratings and use manufacturer-supplied helmets
	if p.HelmetCertifications.SNELL {
		totalScore += snellWeightToUse
	}

	// ECE is the minimum standard required for helmet use in the EU, and helmets must be proven to meet this standard before being sold (not based on the honor system!)
	if p.HelmetCertifications.ECE {
		totalScore += eceWeightToUse
	}

	// DOT is pretty much useless since it's based off the honor system, hence a very low weight
	if p.HelmetCertifications.DOT {
		totalScore += dotWeightToUse
	}

	p.SafetyPercentage = int(math.Round(totalScore * 100))
}

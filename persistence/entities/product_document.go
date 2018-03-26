package entities

import (
	"math"
	"strings"

	"github.com/google/uuid"
)

// ProductDocument represents a safety product such as a motorcycle helmet, jacket, etc. It contains the price of the product, certifications, etc.
type ProductDocument struct {
	UUID                       uuid.UUID `json:"uuid"`
	Type                       string    `json:"type"`
	Subtype                    string    `json:"subtype"`
	Manufacturer               string    `json:"manufacturer"`
	Model                      string    `json:"model"`
	ModelAlias                 string    `json:"modelAlias"`
	SafetyPercentage           int       `json:"safetyPercentage"`
	ImageURL                   string    `json:"imageUrl"`
	AmazonBuyURL               string    `json:"amazonBuyURL"`
	AmazonPriceInUSDMultiple   int       `json:"amazonPriceInUSDMultiple"`
	RevzillaBuyURL             string    `json:"revzillaBuyURL"`
	RevzillaPriceInUSDMultiple int       `json:"revzillaPriceInUSDMultiple"`
	PriceInUSDMultiple         int       `json:"priceInUsdMultiple"`
	LatchPercentage            int       `json:"latchPercentage"`
	WeightInLbsMultiple        int       `json:"weightInLbsMultiple"`
	Sizes                      []string  `json:"sizes"`
	Materials                  string    `json:"materials"`
	RetentionSystem            string    `json:"retentionSystem"`
	Certifications             struct {
		SHARP *SHARPCertificationDocument `json:"SHARP"`
		SNELL bool                        `json:"SNELL"`
		ECE   bool                        `json:"ECE"`
		DOT   bool                        `json:"DOT"`
	} `json:"certifications"`
}

const sharpImpactWeight float64 = 0.2
const sharpImpactMaxValue float64 = 5.0

const defaultSNELLWeight float64 = 0.10
const defaultECEWeight float64 = 0.08
const defaultDOTWeight float64 = 0.02

// UpdateMinPrice sets the minimum price of the product by comparing amazon and revzilla prices and picking the lower of the two, or the higher of the two if one of the prices is <= 0
func (p *ProductDocument) UpdateMinPrice() {
	if p.AmazonPriceInUSDMultiple <= 0 || p.RevzillaPriceInUSDMultiple <= 0 {
		p.PriceInUSDMultiple = int(math.Max(float64(p.AmazonPriceInUSDMultiple), float64(p.RevzillaPriceInUSDMultiple)))
	} else {
		p.PriceInUSDMultiple = int(math.Min(float64(p.AmazonPriceInUSDMultiple), float64(p.RevzillaPriceInUSDMultiple)))
	}
}

// UpdateCertificationsByDescription updates the DOT and/or ECE certifications if the given description contains certain keywords indicating that the product has said certifications and returns booleans indicating whether or not updates occurred.
func (p *ProductDocument) UpdateCertificationsByDescription(productDescription string) (bool, bool) {
	lowerDescription := strings.ToLower(productDescription)
	containsDOT := strings.Contains(productDescription, "DOT") || strings.Contains(productDescription, "D.O.T")
	containsECE := strings.Contains(productDescription, "ECE") || strings.Contains(productDescription, "22/05") || strings.Contains(productDescription, "22.05")
	containsSNELL := strings.Contains(lowerDescription, "snell") || strings.Contains(lowerDescription, "m2010") || strings.Contains(lowerDescription, "m2015")

	hasNewDOTCertification := false
	hasNewECECertification := false

	// SNELL certification implies DOT certification, so check for either cert. Do not update the SNELL cert as we only want to pull SNELL data from the official source (snell.us.com)
	if !p.Certifications.DOT && (containsDOT || containsSNELL) {
		p.Certifications.DOT = true
		hasNewDOTCertification = true
	}

	if !p.Certifications.ECE && containsECE {
		p.Certifications.ECE = true
		hasNewECECertification = true
	}

	return hasNewDOTCertification, hasNewECECertification
}

// CalculateSafetyPercentage calculates how safe a helmet is based on a weighted average of all of its certifications, rounded up to the nearest integer
// SHARP Percentages are calculated by dividing the raw score by the maximum score (i.e. Raw-Score / 5)
// TODO: Make this support multiple product types once gloves, boots, jackets, etc are added
func (p ProductDocument) CalculateSafetyPercentage() int {
	var totalScore float64

	snellWeightToUse := defaultSNELLWeight
	eceWeightToUse := defaultECEWeight
	dotWeightToUse := defaultDOTWeight

	// SHARP is weighted the highest because while they are similar to SHARP, they also provide detailed crash test ratings for each helmet and buy helmets off the shelf instead of getting samples from manufacturers directly
	if p.Certifications.SHARP != nil {
		sharpImpacts := p.Certifications.SHARP.ImpactZoneRatings
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
	if p.Certifications.SNELL {
		totalScore += snellWeightToUse
	}

	// ECE is the minimum standard required for helmet use in the EU, and helmets must be proven to meet this standard before being sold (not based on the honor system!)
	if p.Certifications.ECE {
		totalScore += eceWeightToUse
	}

	// DOT is pretty much useless since it's based off the honor system, hence a very low weight
	if p.Certifications.DOT {
		totalScore += dotWeightToUse
	}

	return int(math.Round(totalScore * 100))
}

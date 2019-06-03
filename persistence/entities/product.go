package entities

import (
	"math"
	"strings"

	"github.com/sirupsen/logrus"

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
		SHARP *SHARPCertification `json:"SHARP"`
		SNELL bool                `json:"SNELL"`
		ECE   bool                `json:"ECE"`
		DOT   bool                `json:"DOT"`
	} `json:"helmetCertifications"`
	JacketCertifications struct {
		Shoulder   *CEImpactZone `json:"shoulder"`
		Elbow      *CEImpactZone `json:"elbow"`
		Back       *CEImpactZone `json:"back"`
		Chest      *CEImpactZone `json:"chest"`
		FitsAirbag bool          `json:"fitsAirbag"`
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

// UpdateHelmetCertificationsByDescription updates the DOT and/or ECE certifications if the given description contains certain keywords indicating that the product has said certifications and returns booleans indicating whether or not updates occurred.
func (p *Product) UpdateHelmetCertificationsByDescription(productDescription string) (bool, bool) {
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

func isScoreUpgraded(oldZone *CEImpactZone, newZone *CEImpactZone) bool {
	return oldZone == nil || (newZone.GetScore() > oldZone.GetScore())
}

// UpdateJacketSubtypeByDescriptionParts updates the subtype when certain text appears in each part of the description
func (p *Product) UpdateJacketSubtypeByDescriptionParts(productDescriptionParts []string) {
	for _, part := range productDescriptionParts {
		lowerPart := strings.ToLower(part)
		if strings.Contains(lowerPart, "gore-tex") || strings.Contains(lowerPart, "goretex") || strings.Contains(lowerPart, "gore tex") {
			p.Subtype = "goretex"
			p.Materials = p.Subtype
			return
		} else if strings.Contains(lowerPart, "leather") {
			p.Subtype = "leather"
			p.Materials = p.Subtype
			return
		} else {
			p.Subtype = "textile"
			p.Materials = p.Subtype
			return
		}
	}
}

// UpdateJacketCertificationsByDescriptionParts updates all of the jacket certifications when certain text appears in each part of the description
func (p *Product) UpdateJacketCertificationsByDescriptionParts(productDescriptionParts []string) (bool, bool, bool, bool, bool) {
	updatedAirbag := false
	updatedBack := false
	updatedShoulder := false
	updatedElbow := false
	updatedChest := false

	for _, part := range productDescriptionParts {
		lowerPart := strings.ToLower(part)

		isEmpty := strings.Contains(lowerPart, "sold separately") || strings.Contains(lowerPart, "optional") || strings.Contains(lowerPart, "pocket")
		fitsAirbag := strings.Contains(lowerPart, "d-air") || strings.Contains(lowerPart, "tech-air") || strings.Contains(lowerPart, "tech air") || strings.Contains(lowerPart, "air bag") || strings.Contains(lowerPart, "airbag")

		if !p.JacketCertifications.FitsAirbag && fitsAirbag {
			p.JacketCertifications.FitsAirbag = true
			updatedAirbag = true
		}

		isCertified := strings.Contains(part, "CE") || strings.Contains(lowerPart, "level 1") ||
			strings.Contains(lowerPart, "1621") || strings.Contains(lowerPart, "pro-armor") ||
			strings.Contains(lowerPart, "pro armor") || strings.Contains(lowerPart, "pro shape") ||
			strings.Contains(lowerPart, "pro-shape") // Pro-armor is Dainese-specific armor that is CE-level 1 certified
		isApproved := strings.Contains(lowerPart, "ce approved")
		isLevel2 := strings.Contains(lowerPart, "level 2") || strings.Contains(lowerPart, "level ii") || strings.Contains(lowerPart, "cat. ii") || strings.Contains(lowerPart, "cat ii")
		if isLevel2 {
			isCertified = true
		}

		// If we have conflicting information (we think this is an empty slot but we also found CE cert details) assume the worst
		if (isCertified || isApproved) && isEmpty {
			isCertified = false
			isApproved = false
			isLevel2 = false
		}

		newCEImpactZone := &CEImpactZone{IsApproved: isApproved, IsLevel2: isLevel2, IsEmpty: isEmpty}
		if isCertified || isApproved || isEmpty {
			if isScoreUpgraded(p.JacketCertifications.Back, newCEImpactZone) && strings.Contains(lowerPart, "back") {
				p.JacketCertifications.Back = newCEImpactZone
				updatedBack = true
			}

			if isScoreUpgraded(p.JacketCertifications.Elbow, newCEImpactZone) && strings.Contains(lowerPart, "elbow") {
				p.JacketCertifications.Elbow = newCEImpactZone
				updatedElbow = true
			}

			if isScoreUpgraded(p.JacketCertifications.Shoulder, newCEImpactZone) && strings.Contains(lowerPart, "shoulder") {
				p.JacketCertifications.Shoulder = newCEImpactZone
				updatedShoulder = true
			}

			if isScoreUpgraded(p.JacketCertifications.Chest, newCEImpactZone) && strings.Contains(lowerPart, "chest") {
				p.JacketCertifications.Chest = newCEImpactZone
				updatedChest = true
			}
		}
	}

	return updatedBack, updatedElbow, updatedShoulder, updatedChest, updatedAirbag
}

func (p *Product) getJacketSafetyPercentage() int {
	totalScore := float64(0)

	zones := []*CEImpactZone{p.JacketCertifications.Back, p.JacketCertifications.Chest, p.JacketCertifications.Elbow, p.JacketCertifications.Shoulder}
	for _, zone := range zones {
		if zone != nil {
			totalScore += zone.GetScore() * float64(0.2125)
		}
	}

	if p.Materials == "leather" {
		totalScore += float64(0.10)
	}

	if p.JacketCertifications.FitsAirbag {
		totalScore += float64(0.05)
	}
	return int(math.Round(totalScore * 100))
}

func (p *Product) getHelmetSafetyPercentage() int {
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

	return int(math.Round(totalScore * 100))
}

// UpdateSafetyPercentage calculates how safe a helmet is based on a weighted average of all of its certifications, rounded up to the nearest integer.
// SHARP Percentages are calculated by dividing the raw score by the maximum score (i.e. Raw-Score / 5)
func (p *Product) UpdateSafetyPercentage() {

	if p.Type == "" {
		logrus.Error("Attempted to update a safety percentage for a product without a type")
		return
	}

	// TODO: consider an OO approach when other pieces of gear are added
	safetyPercentage := 0
	switch p.Type {
	case "helmet":
		safetyPercentage = p.getHelmetSafetyPercentage()
		break
	case "jacket":
		safetyPercentage = p.getJacketSafetyPercentage()
		break
	}

	p.SafetyPercentage = safetyPercentage
}

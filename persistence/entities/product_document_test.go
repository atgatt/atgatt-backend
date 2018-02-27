package entities

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_CalculateSafetyPercentage_should_return_100_when_the_product_has_the_highest_possible_impact_ratings_and_all_certifications(t *testing.T) {
	RegisterTestingT(t)
	productDocument := &ProductDocument{ImageURL: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", PriceInUSDMultiple: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	productDocument.Certifications.ECE = true
	productDocument.Certifications.DOT = true
	productDocument.Certifications.SHARP = &SHARPCertificationDocument{}
	productDocument.Certifications.SHARP.Stars = 0 // Stars should have no effect on the score
	productDocument.Certifications.SHARP.ImpactZoneRatings = &SHARPImpactZoneRatingsDocument{}
	productDocument.Certifications.SHARP.ImpactZoneRatings.Left = 5
	productDocument.Certifications.SHARP.ImpactZoneRatings.Right = 5
	productDocument.Certifications.SHARP.ImpactZoneRatings.Rear = 5
	productDocument.Certifications.SHARP.ImpactZoneRatings.Top.Front = 5
	productDocument.Certifications.SHARP.ImpactZoneRatings.Top.Rear = 5
	productDocument.Certifications.SNELL = true

	Expect(productDocument.CalculateSafetyPercentage()).To(Equal(100))
}

func Test_CalculateSafetyPercentage_should_return_0_when_the_product_has_the_lowest_possible_impact_ratings_and_zero_certifications(t *testing.T) {
	RegisterTestingT(t)
	productDocument := &ProductDocument{ImageURL: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", PriceInUSDMultiple: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	productDocument.Certifications.ECE = false
	productDocument.Certifications.DOT = false
	productDocument.Certifications.SHARP = &SHARPCertificationDocument{}
	productDocument.Certifications.SHARP.Stars = 0
	productDocument.Certifications.SHARP.ImpactZoneRatings = &SHARPImpactZoneRatingsDocument{}
	productDocument.Certifications.SHARP.ImpactZoneRatings.Left = 0
	productDocument.Certifications.SHARP.ImpactZoneRatings.Right = 0
	productDocument.Certifications.SHARP.ImpactZoneRatings.Rear = 0
	productDocument.Certifications.SHARP.ImpactZoneRatings.Top.Front = 0
	productDocument.Certifications.SHARP.ImpactZoneRatings.Top.Rear = 0
	productDocument.Certifications.SNELL = false

	Expect(productDocument.CalculateSafetyPercentage()).To(Equal(0))
}

func Test_CalculateSafetyPercentage_should_return_0_when_the_product_has_nonexistent_impact_ratings_and_zero_certifications(t *testing.T) {
	RegisterTestingT(t)
	productDocument := &ProductDocument{ImageURL: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", PriceInUSDMultiple: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	productDocument.Certifications.ECE = false
	productDocument.Certifications.DOT = false
	productDocument.Certifications.SHARP = nil
	productDocument.Certifications.SNELL = false

	Expect(productDocument.CalculateSafetyPercentage()).To(Equal(0))
}

func Test_CalculateSafetyPercentage_should_return_correctly_when_the_product_has_nonexistent_impact_ratings_and_a_snell_certification(t *testing.T) {
	RegisterTestingT(t)
	productDocument := &ProductDocument{ImageURL: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", PriceInUSDMultiple: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	productDocument.Certifications.ECE = false
	productDocument.Certifications.DOT = false
	productDocument.Certifications.SHARP = nil
	productDocument.Certifications.SNELL = true

	Expect(productDocument.CalculateSafetyPercentage()).To(Equal(65))
}

func Test_CalculateSafetyPercentage_should_return_correctly_when_the_product_has_nonexistent_impact_ratings_and_snell_dot_certifications(t *testing.T) {
	RegisterTestingT(t)
	productDocument := &ProductDocument{ImageURL: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", PriceInUSDMultiple: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	productDocument.Certifications.ECE = false
	productDocument.Certifications.DOT = true
	productDocument.Certifications.SHARP = nil
	productDocument.Certifications.SNELL = true

	Expect(productDocument.CalculateSafetyPercentage()).To(Equal(70))
}

func Test_CalculateSafetyPercentage_should_return_80_when_the_product_has_nonexistent_impact_ratings_but_all_other_certifications(t *testing.T) {
	RegisterTestingT(t)
	productDocument := &ProductDocument{ImageURL: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", PriceInUSDMultiple: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	productDocument.Certifications.ECE = true
	productDocument.Certifications.DOT = true
	productDocument.Certifications.SHARP = nil
	productDocument.Certifications.SNELL = true

	Expect(productDocument.CalculateSafetyPercentage()).To(Equal(80))
}

func Test_CalculateSafetyPercentage_should_return_correctly_when_the_product_has_nonexistent_impact_ratings_and_a_dot_certification(t *testing.T) {
	RegisterTestingT(t)
	productDocument := &ProductDocument{ImageURL: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", PriceInUSDMultiple: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	productDocument.Certifications.ECE = false
	productDocument.Certifications.DOT = true
	productDocument.Certifications.SHARP = nil
	productDocument.Certifications.SNELL = false

	Expect(productDocument.CalculateSafetyPercentage()).To(Equal(5))
}

func Test_CalculateSafetyPercentage_should_return_correctly_when_the_product_has_nonexistent_impact_ratings_and_a_ece_certification(t *testing.T) {
	RegisterTestingT(t)
	productDocument := &ProductDocument{ImageURL: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", PriceInUSDMultiple: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	productDocument.Certifications.ECE = true
	productDocument.Certifications.DOT = false
	productDocument.Certifications.SHARP = nil
	productDocument.Certifications.SNELL = false

	Expect(productDocument.CalculateSafetyPercentage()).To(Equal(10))
}

func Test_CalculateSafetyPercentage_should_return_correctly_when_the_product_has_partial_impact_ratings_and_all_other_certifications(t *testing.T) {
	RegisterTestingT(t)
	productDocument := &ProductDocument{ImageURL: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", PriceInUSDMultiple: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	productDocument.Certifications.ECE = true
	productDocument.Certifications.DOT = true
	productDocument.Certifications.SHARP = &SHARPCertificationDocument{}
	productDocument.Certifications.SHARP.Stars = 0
	productDocument.Certifications.SHARP.ImpactZoneRatings = &SHARPImpactZoneRatingsDocument{}
	productDocument.Certifications.SHARP.ImpactZoneRatings.Left = 0
	productDocument.Certifications.SHARP.ImpactZoneRatings.Right = 1
	productDocument.Certifications.SHARP.ImpactZoneRatings.Rear = 3
	productDocument.Certifications.SHARP.ImpactZoneRatings.Top.Front = 5
	productDocument.Certifications.SHARP.ImpactZoneRatings.Top.Rear = 4
	productDocument.Certifications.SNELL = true

	Expect(productDocument.CalculateSafetyPercentage()).To(Equal(62))
}

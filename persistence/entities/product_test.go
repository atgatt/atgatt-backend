package entities

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_CalculateSafetyPercentage_should_return_100_when_the_product_has_the_highest_possible_impact_ratings_and_all_certifications(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	product.HelmetCertifications.ECE = true
	product.HelmetCertifications.DOT = true
	product.HelmetCertifications.SHARP = &SHARPCertification{}
	product.HelmetCertifications.SHARP.Stars = 0 // Stars should have no effect on the score
	product.HelmetCertifications.SHARP.ImpactZoneRatings = &SHARPImpactZoneRatings{}
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Left = 5
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Right = 5
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Rear = 5
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Top.Front = 5
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Top.Rear = 5
	product.HelmetCertifications.SNELL = true
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(100))
}

func Test_CalculateSafetyPercentage_should_return_0_when_the_product_has_the_lowest_possible_impact_ratings_and_zero_certifications(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	product.HelmetCertifications.ECE = false
	product.HelmetCertifications.DOT = false
	product.HelmetCertifications.SHARP = &SHARPCertification{}
	product.HelmetCertifications.SHARP.Stars = 0
	product.HelmetCertifications.SHARP.ImpactZoneRatings = &SHARPImpactZoneRatings{}
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Left = 0
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Right = 0
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Rear = 0
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Top.Front = 0
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Top.Rear = 0
	product.HelmetCertifications.SNELL = false
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(0))
}

func Test_CalculateSafetyPercentage_should_return_0_when_the_product_has_nonexistent_impact_ratings_and_zero_certifications(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	product.HelmetCertifications.ECE = false
	product.HelmetCertifications.DOT = false
	product.HelmetCertifications.SHARP = nil
	product.HelmetCertifications.SNELL = false
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(0))
}

func Test_CalculateSafetyPercentage_should_return_correctly_when_the_product_has_nonexistent_impact_ratings_and_a_snell_certification(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	product.HelmetCertifications.ECE = false
	product.HelmetCertifications.DOT = false
	product.HelmetCertifications.SHARP = nil
	product.HelmetCertifications.SNELL = true
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(65))
}

func Test_CalculateSafetyPercentage_should_return_correctly_when_the_product_has_nonexistent_impact_ratings_and_snell_dot_certifications(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	product.HelmetCertifications.ECE = false
	product.HelmetCertifications.DOT = true
	product.HelmetCertifications.SHARP = nil
	product.HelmetCertifications.SNELL = true
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(70))
}

func Test_CalculateSafetyPercentage_should_return_80_when_the_product_has_nonexistent_impact_ratings_but_all_other_certifications(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	product.HelmetCertifications.ECE = true
	product.HelmetCertifications.DOT = true
	product.HelmetCertifications.SHARP = nil
	product.HelmetCertifications.SNELL = true
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(80))
}

func Test_CalculateSafetyPercentage_should_return_correctly_when_the_product_has_nonexistent_impact_ratings_and_a_dot_certification(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	product.HelmetCertifications.ECE = false
	product.HelmetCertifications.DOT = true
	product.HelmetCertifications.SHARP = nil
	product.HelmetCertifications.SNELL = false
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(5))
}

func Test_CalculateSafetyPercentage_should_return_correctly_when_the_product_has_nonexistent_impact_ratings_and_a_ece_certification(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	product.HelmetCertifications.ECE = true
	product.HelmetCertifications.DOT = false
	product.HelmetCertifications.SHARP = nil
	product.HelmetCertifications.SNELL = false
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(10))
}

func Test_CalculateSafetyPercentage_should_return_correctly_when_the_product_has_partial_impact_ratings_and_all_other_certifications(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: -1234}
	product.HelmetCertifications.ECE = true
	product.HelmetCertifications.DOT = true
	product.HelmetCertifications.SHARP = &SHARPCertification{}
	product.HelmetCertifications.SHARP.Stars = 0
	product.HelmetCertifications.SHARP.ImpactZoneRatings = &SHARPImpactZoneRatings{}
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Left = 0
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Right = 1
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Rear = 3
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Top.Front = 5
	product.HelmetCertifications.SHARP.ImpactZoneRatings.Top.Rear = 4
	product.HelmetCertifications.SNELL = true
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(62))
}

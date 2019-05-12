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

func Test_CalculateSafetyPercentage_should_return_a_full_safety_score_when_the_product_is_a_jacket_with_all_parts_certified(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "jacket", SafetyPercentage: -1234}
	fullImpactZone := &CEImpactZone{IsApproved: true, IsLevel2: true}
	product.JacketCertifications.Back = fullImpactZone
	product.JacketCertifications.Chest = fullImpactZone
	product.JacketCertifications.Elbow = fullImpactZone
	product.JacketCertifications.Shoulder = fullImpactZone
	product.JacketCertifications.FitsAirbag = true
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(100))
}

func Test_CalculateSafetyPercentage_should_return_a_reduced_safety_score_when_the_product_is_a_jacket_that_is_missing_an_airbag(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "jacket", SafetyPercentage: -1234}
	fullImpactZone := &CEImpactZone{IsApproved: true, IsLevel2: true}
	product.JacketCertifications.Back = fullImpactZone
	product.JacketCertifications.Chest = fullImpactZone
	product.JacketCertifications.Elbow = fullImpactZone
	product.JacketCertifications.Shoulder = fullImpactZone
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(85))
}

func Test_CalculateSafetyPercentage_should_return_zero_when_the_product_is_a_jacket_without_any_armor_slots(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "jacket", SafetyPercentage: -1234}
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(0))
}

func Test_CalculateSafetyPercentage_should_return_64_when_the_product_is_a_jacket_with_approved_ce_level_1_armor(t *testing.T) {
	RegisterTestingT(t)
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "jacket", SafetyPercentage: -1234}
	level1ImpactZone := &CEImpactZone{IsApproved: true}
	product.JacketCertifications.Back = level1ImpactZone
	product.JacketCertifications.Chest = level1ImpactZone
	product.JacketCertifications.Elbow = level1ImpactZone
	product.JacketCertifications.Shoulder = level1ImpactZone
	product.UpdateSafetyPercentage()

	Expect(product.SafetyPercentage).To(Equal(64))
}

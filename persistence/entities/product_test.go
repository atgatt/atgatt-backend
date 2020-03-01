package entities

import (
	text "atgatt-backend/common/text"
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
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
	product := &Product{ImageKey: "google.com/lol.png", Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "jacket", Subtype: "leather", Materials: "leather", SafetyPercentage: -1234}
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

	Expect(product.SafetyPercentage).To(Equal(68))
}

func generateMockDescriptionPartsFromHTML(html string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	mockDescriptionParts := []string{}
	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		mockDescriptionParts = append(mockDescriptionParts, s.Text())
	})

	return mockDescriptionParts, nil
}

func Test_UpdateJacketCertificationsByDescriptionParts_should_apply_CE_level_2_certification_when_level_2_is_found(t *testing.T) {
	RegisterTestingT(t)

	mockProductDescription := `
	<ul>
	<li>
		Soft distressed leather</li>
	<li>
		Sas-Tec Level 2 armor at elbows and shoulders</li>
	<li>
		Pocket at&nbsp;back for optional Sas-Tec&nbsp;back protector (<a href="/motorcycle/scorpion-sc-115-sas-tec-back-protector">sold separately</a>)</li>
	<li>
		Perforated panels underarms and sides of torso</li>
	<li>
		Two zippered rear vents</li>
	<li>
		Leather overlays at elbows</li>
	<li>
		4 external pockets</li>
	<li>
		Rib stretch panels as side hems</li>
	<li>
		Rear waist adjustment tabs</li>
	<li>
		Zipper closures at wrists</li>
	<li>
		Antique Brass YKK&nbsp;zippers throughout</li>
	<li>
		Padded comfort collar</li>
	<li>
		Two internal mesh pockets</li>
	<li>
		Removable EverHeat&nbsp;jacket liner with Kwikwick&nbsp;panels</li>
	<li>
		8” jacket to pant zipper and rear belt loop attachment tab</li>
	</ul>
	`

	mockDescriptionParts, err := generateMockDescriptionPartsFromHTML(mockProductDescription)
	Expect(err).To(BeNil())

	product := &Product{}
	updatedBack, updatedElbow, updatedShoulder, updatedChest, updatedAirbag := product.UpdateJacketCertificationsByDescriptionParts(mockDescriptionParts)

	Expect(updatedBack).To(BeTrue())
	Expect(product.JacketCertifications.Back).To(Equal(&CEImpactZone{IsLevel2: false, IsApproved: false, IsEmpty: true}))

	Expect(updatedShoulder).To(BeTrue())
	Expect(product.JacketCertifications.Shoulder).To(Equal(&CEImpactZone{IsLevel2: true, IsApproved: false, IsEmpty: false}))

	Expect(updatedChest).To(BeFalse())
	Expect(product.JacketCertifications.Chest).To(BeNil())

	Expect(updatedElbow).To(BeTrue())
	Expect(product.JacketCertifications.Elbow).To(Equal(&CEImpactZone{IsLevel2: true, IsApproved: false, IsEmpty: false}))

	Expect(updatedAirbag).To(BeFalse())
	Expect(product.JacketCertifications.FitsAirbag).To(BeFalse())
}

func Test_UpdateJacketCertificationsByDescriptionParts_should_apply_CE_level_1_certifications_for_pro_armor(t *testing.T) {
	RegisterTestingT(t)

	mockProductDescription := `
	<ul>
		<li>Dainese's Rapida72 Leather Jacket encompasses racing history with modern functionality. The curves of the Rapida72 Perforated Jacket hug your body just like your motorcycle hugs apexes. 
		Slim Pro-armor provides impact protection without adding bulk that would ruin the lines of the jacket. Supple perforated leather adds abrasion resistance and air flow. 
		Riders with an aesthetic for a bygone era can wear the Dainese Rapida72 into modern times.</li>

		<li>
			Soft natural cowhide leather</li>
		<li>
			Removable soft Pro-armor protectors certified to standard EN 1621.1 on shoulders and elbows</li>
		<li>
			Jacket-trousers connection loop</li>
		<li>
			Perforated leather</li>
		<li>
			Printed cotton liner</li>
		<li>
			1 inner pocket</li>
		<li>
			3 outer pockets</li>
		<li>
			Pocket for G1 or G2 back protector (sold separately)</li>
	</ul>
	`

	mockDescriptionParts, err := generateMockDescriptionPartsFromHTML(mockProductDescription)
	Expect(err).To(BeNil())

	product := &Product{}
	updatedBack, updatedElbow, updatedShoulder, updatedChest, updatedAirbag := product.UpdateJacketCertificationsByDescriptionParts(mockDescriptionParts)

	Expect(updatedBack).To(BeTrue())
	Expect(product.JacketCertifications.Back).To(Equal(&CEImpactZone{IsLevel2: false, IsApproved: false, IsEmpty: true}))

	Expect(updatedShoulder).To(BeTrue())
	Expect(product.JacketCertifications.Shoulder).To(Equal(&CEImpactZone{IsLevel2: false, IsApproved: false, IsEmpty: false}))

	Expect(updatedChest).To(BeFalse())
	Expect(product.JacketCertifications.Chest).To(BeNil())

	Expect(updatedElbow).To(BeTrue())
	Expect(product.JacketCertifications.Elbow).To(Equal(&CEImpactZone{IsLevel2: false, IsApproved: false, IsEmpty: false}))

	Expect(updatedAirbag).To(BeFalse())
	Expect(product.JacketCertifications.FitsAirbag).To(BeFalse())
}

func Test_UpdateJacketCertificationsByDescriptionParts_should_apply_CE_level_1_certifications_for_pro_armor_without_dash(t *testing.T) {
	RegisterTestingT(t)

	mockProductDescription := `
	<ul>
		<li>Dainese's Rapida72 Leather Jacket encompasses racing history with modern functionality. The curves of the Rapida72 Perforated Jacket hug your body just like your motorcycle hugs apexes. 
		Slim Pro-armor provides impact protection without adding bulk that would ruin the lines of the jacket. Supple perforated leather adds abrasion resistance and air flow. 
		Riders with an aesthetic for a bygone era can wear the Dainese Rapida72 into modern times.</li>

		<li>
			Soft natural cowhide leather</li>
		<li>
			Removable soft Pro shape protectors certified to standard EN 1621.1 on shoulders and elbows</li>
		<li>
			Jacket-trousers connection loop</li>
		<li>
			Perforated leather</li>
		<li>
			Printed cotton liner</li>
		<li>
			1 inner pocket</li>
		<li>
			3 outer pockets</li>
		<li>
			Pocket for G1 or G2 back protector (sold separately)</li>
	</ul>
	`

	mockDescriptionParts, err := generateMockDescriptionPartsFromHTML(mockProductDescription)
	Expect(err).To(BeNil())

	product := &Product{}
	updatedBack, updatedElbow, updatedShoulder, updatedChest, updatedAirbag := product.UpdateJacketCertificationsByDescriptionParts(mockDescriptionParts)

	Expect(updatedBack).To(BeTrue())
	Expect(product.JacketCertifications.Back).To(Equal(&CEImpactZone{IsLevel2: false, IsApproved: false, IsEmpty: true}))

	Expect(updatedShoulder).To(BeTrue())
	Expect(product.JacketCertifications.Shoulder).To(Equal(&CEImpactZone{IsLevel2: false, IsApproved: false, IsEmpty: false}))

	Expect(updatedChest).To(BeFalse())
	Expect(product.JacketCertifications.Chest).To(BeNil())

	Expect(updatedElbow).To(BeTrue())
	Expect(product.JacketCertifications.Elbow).To(Equal(&CEImpactZone{IsLevel2: false, IsApproved: false, IsEmpty: false}))

	Expect(updatedAirbag).To(BeFalse())
	Expect(product.JacketCertifications.FitsAirbag).To(BeFalse())
}

func Test_UpdateJacketCertificationsByDescriptionParts_should_apply_CE_level_2_certifications_when_the_category_is_cat_ii(t *testing.T) {
	RegisterTestingT(t)

	mockSummary := `The 8-track tape may be obsolete, but the Dainese 8-Track Leather Jacket might as well be a subscription streaming service coming out of wireless speakers. 
	Artemide refined full-grain cowhide leather gives the 8-Track Jacket an old school sheen with modern day robustness. 
	CE armor at the elbows and shoulders along with specific design features allow the jacket to meet CE - Cat II - prEN 17092 certification. 
	Slide in an optional back protector (sold separately) to upgrade the impact protection. 
	A removable thermal liner allows you to stretch the 8-Track into cooler temperatures and can even be used as a separate mid-layer. 
	The Dainese 8-Track Jacket is is named after some tech from the 60s, but its function is as state of the art as it gets for motorcycle gear.`

	sentences, _ := text.GetSentencesFromString(mockSummary)
	var sb strings.Builder
	for _, sentence := range sentences {
		sb.WriteString("<li>")
		sb.WriteString(sentence)
		sb.WriteString("</li>")
	}
	mockProductDescription := fmt.Sprintf(`
	<ul>
		%s
		<li>
			Soft natural cowhide leather</li>
		<li>
			Removable soft Pro armor protectors certified to standard EN 1621.1 on shoulders and elbows</li>
		<li>
			Jacket-trousers connection loop</li>
		<li>
			Perforated leather</li>
		<li>
			Printed cotton liner</li>
		<li>
			1 inner pocket</li>
		<li>
			3 outer pockets</li>
		<li>
			Pocket for G1 or G2 back protector (sold separately)</li>
	</ul>
	`, sb.String())

	mockDescriptionParts, err := generateMockDescriptionPartsFromHTML(mockProductDescription)
	Expect(err).To(BeNil())

	product := &Product{}
	updatedBack, updatedElbow, updatedShoulder, updatedChest, updatedAirbag := product.UpdateJacketCertificationsByDescriptionParts(mockDescriptionParts)

	Expect(updatedBack).To(BeTrue())
	Expect(product.JacketCertifications.Back).To(Equal(&CEImpactZone{IsLevel2: false, IsApproved: false, IsEmpty: true}))

	Expect(updatedShoulder).To(BeTrue())
	Expect(product.JacketCertifications.Shoulder).To(Equal(&CEImpactZone{IsLevel2: true, IsApproved: false, IsEmpty: false}))

	Expect(updatedChest).To(BeFalse())
	Expect(product.JacketCertifications.Chest).To(BeNil())

	Expect(updatedElbow).To(BeTrue())
	Expect(product.JacketCertifications.Elbow).To(Equal(&CEImpactZone{IsLevel2: true, IsApproved: false, IsEmpty: false}))

	Expect(updatedAirbag).To(BeFalse())
	Expect(product.JacketCertifications.FitsAirbag).To(BeFalse())
}

func Test_UpdatePantsCertificationsByDescriptionParts_should_apply_CE_level_2_certifications(t *testing.T) {
	RegisterTestingT(t)

	mockSummary := `The 8-track tape may be obsolete, but the Dainese 8-Track Leather Jacket might as well be a subscription streaming service coming out of wireless speakers. 
	Artemide refined full-grain cowhide leather gives the 8-Track Jacket an old school sheen with modern day robustness. 
	CE armor at the tailbone and hip along with specific design features allow the jacket to meet CE - Cat II - prEN 17092 certification. 
	Slide in an optional back protector (sold separately) to upgrade the impact protection. 
	A removable thermal liner allows you to stretch the 8-Track into cooler temperatures and can even be used as a separate mid-layer. 
	The Dainese 8-Track Jacket is is named after some tech from the 60s, but its function is as state of the art as it gets for motorcycle gear.`

	sentences, _ := text.GetSentencesFromString(mockSummary)
	var sb strings.Builder
	for _, sentence := range sentences {
		sb.WriteString("<li>")
		sb.WriteString(sentence)
		sb.WriteString("</li>")
	}
	mockProductDescription := fmt.Sprintf(`
	<ul>
		%s
		<li>
			Soft natural cowhide leather</li>
		<li>
			Removable soft Pro armor protectors certified to standard EN 1621.1 on knee level 2 ce approved</li>
		<li>
			Jacket-trousers connection loop</li>
		<li>
			Perforated leather</li>
		<li>
			Printed cotton liner</li>
		<li>
			1 inner pocket</li>
		<li>
			3 outer pockets</li>
		<li>
			Pocket for G1 or G2 back protector (sold separately)</li>
	</ul>
	`, sb.String())

	mockDescriptionParts, err := generateMockDescriptionPartsFromHTML(mockProductDescription)
	Expect(err).To(BeNil())

	product := &Product{}
	updatedTailbone, updatedHip, updatedKnee := product.UpdatePantsCertificationsByDescriptionParts(mockDescriptionParts)

	Expect(updatedTailbone).To(BeTrue())
	Expect(product.PantsCertifications.Tailbone).To(Equal(&CEImpactZone{IsLevel2: true, IsApproved: false, IsEmpty: false}))

	Expect(updatedHip).To(BeTrue())
	Expect(product.PantsCertifications.Hip).To(Equal(&CEImpactZone{IsLevel2: true, IsApproved: false, IsEmpty: false}))

	Expect(updatedKnee).To(BeTrue())
	Expect(product.PantsCertifications.Knee).To(Equal(&CEImpactZone{IsLevel2: true, IsApproved: true, IsEmpty: false}))
}

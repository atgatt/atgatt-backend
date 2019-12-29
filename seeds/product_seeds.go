package seeds

import (
	"crashtested-backend/persistence/entities"
	"encoding/json"
	"fmt"

	golinq "github.com/ahmetb/go-linq"
	"github.com/google/uuid"
)

const mockHelmetImageURL = "https://sharp.dft.gov.uk/wp-content/uploads/2017/03/shoei-x-spirit-lll.jpg"

// GetProductSeedsSQLStatements returns an array of INSERT statements that target each of the product seed structs. Used to import test data into the database for automated tests, local development.
func GetProductSeedsSQLStatements(productSeeds []*entities.Product) ([]string, error) {
	statements := []string{}
	for _, product := range productSeeds {
		documentJSONBytes, err := json.Marshal(product)
		if err != nil {
			return nil, err
		}
		documentJSONString := string(documentJSONBytes)
		formattedInsertStatement := fmt.Sprintf("insert into products (uuid, document, created_at_utc, updated_at_utc) values ('%s', '%s', (now() at time zone 'utc'), null);", product.UUID.String(), documentJSONString)
		statements = append(statements, formattedInsertStatement)
	}
	return statements, nil
}

// GetProductSeedsExceptDiscontinued returns all seeds except for the products that are marked as discontinued (useful for functional tests)
func GetProductSeedsExceptDiscontinued() []*entities.Product {
	seedsExceptDiscontinued := []*entities.Product{}
	golinq.From(GetProductSeeds()).WhereT(func(product *entities.Product) bool {
		return !product.IsDiscontinued
	}).ToSlice(&seedsExceptDiscontinued)

	return seedsExceptDiscontinued
}

// GetRealisticProductSeeds returns a list of products modeled after real helmets. This is useful for integration testing background jobs using a limited, but real set of data.
func GetRealisticProductSeeds() []*entities.Product {
	x14ModelAliases := []*entities.ProductModelAlias{
		{
			ModelAlias: "X-Fourteen",
		},
		{
			ModelAlias:   "X-14",
			IsForDisplay: true,
		},
	}

	uuids := []string{
		"2ef2e322-8b7c-4b11-8432-15d082f49f43",
		"55e620cb-4eb3-46d7-a612-d8bf55088494",
		"0e78d74a-da19-4015-a76a-703a37d02503",
		"7321fc5c-596c-4b63-be0c-0d7af3fd78cc",
	}

	seeds := []*entities.Product{
		// This is the Shoei X-14 which is an active helmet where the alias matches a Revzilla product, but the model doesn't
		{ImageKey: mockHelmetImageURL, Manufacturer: "Shoei", Model: "X Spirit lll", MSRPCents: 0, Type: "helmet", Subtype: "full", ModelAliases: x14ModelAliases, SafetyPercentage: 0, IsDiscontinued: false},
		// This is the Shoei X-12 which is a discontinued helmet where the model matches a Revzilla product, but the aliases don't
		{ImageKey: mockHelmetImageURL, Manufacturer: "Shoei", Model: "X-12", MSRPCents: 0, Type: "helmet", Subtype: "full", SafetyPercentage: 0, IsDiscontinued: false},
		// This is the Bell Star which is an active helmet where the model matches a Revzilla product, but the aliases don't
		{ImageKey: mockHelmetImageURL, Manufacturer: "Bell", Model: "Star", MSRPCents: 0, Type: "helmet", Subtype: "full", SafetyPercentage: 0, IsDiscontinued: false},
		// This helmet does not exist
		{ImageKey: mockHelmetImageURL, Manufacturer: "IAMNOTREAL", Model: "IDONOTEXIST", MSRPCents: 0, Type: "helmet", Subtype: "full", SafetyPercentage: 0, IsDiscontinued: false},
	}

	for i := 0; i < len(seeds); i++ {
		seeds[i].UpdateSearchPrice()
		seeds[i].UUID = uuid.MustParse(uuids[i])
	}

	return seeds
}

func applySeedDataToHelmet(i int, product *entities.Product) {
	product.RevzillaPriceCents = product.MSRPCents + 10000
	product.RevzillaBuyURL = fmt.Sprintf("http://www.testdata.com/revzilla/%d", i)
	if i%2 == 0 {
		product.HelmetCertifications.ECE = true
		product.HelmetCertifications.DOT = true
		product.HelmetCertifications.SHARP = &entities.SHARPCertification{}
		product.HelmetCertifications.SHARP.Stars = 4
		product.HelmetCertifications.SHARP.ImpactZoneRatings = &entities.SHARPImpactZoneRatings{}
		product.HelmetCertifications.SHARP.ImpactZoneRatings.Left = 4
		product.HelmetCertifications.SHARP.ImpactZoneRatings.Right = 3
		product.HelmetCertifications.SHARP.ImpactZoneRatings.Rear = 4
		product.HelmetCertifications.SHARP.ImpactZoneRatings.Top.Front = 3
		product.HelmetCertifications.SHARP.ImpactZoneRatings.Top.Rear = 5
		product.HelmetCertifications.SNELL = true
	} else if i%3 == 0 {
		product.HelmetCertifications.ECE = true
		product.HelmetCertifications.DOT = true
		product.HelmetCertifications.SHARP = &entities.SHARPCertification{}
		product.HelmetCertifications.SHARP.Stars = 3
		product.HelmetCertifications.SHARP.ImpactZoneRatings = &entities.SHARPImpactZoneRatings{}
		product.HelmetCertifications.SHARP.ImpactZoneRatings.Left = 1
		product.HelmetCertifications.SHARP.ImpactZoneRatings.Right = 1
		product.HelmetCertifications.SHARP.ImpactZoneRatings.Rear = 2
		product.HelmetCertifications.SHARP.ImpactZoneRatings.Top.Front = 2
		product.HelmetCertifications.SHARP.ImpactZoneRatings.Top.Rear = 3
		product.HelmetCertifications.SNELL = true
	} else {
		product.HelmetCertifications.ECE = false
		product.HelmetCertifications.DOT = false
		product.HelmetCertifications.SHARP = nil
		product.HelmetCertifications.SNELL = false
		product.RevzillaBuyURL = ""
		product.RevzillaPriceCents = 0
	}
}

func applySeedDataToJacket(i int, product *entities.Product) {
	product.RevzillaPriceCents = product.MSRPCents + 10000
	product.RevzillaBuyURL = fmt.Sprintf("http://www.testdata.com/revzilla/%d", i)
	if i%2 == 0 {
		level2Zone := &entities.CEImpactZone{IsLevel2: true, IsApproved: true, IsEmpty: false}
		product.JacketCertifications.Shoulder = level2Zone
		product.JacketCertifications.Elbow = level2Zone
		product.JacketCertifications.Back = level2Zone
		product.JacketCertifications.Chest = level2Zone
		product.JacketCertifications.FitsAirbag = true
	} else if i%3 == 0 {
		level1Zone := &entities.CEImpactZone{IsLevel2: false, IsApproved: false, IsEmpty: false}
		product.JacketCertifications.Shoulder = level1Zone
		product.JacketCertifications.Elbow = level1Zone
		product.JacketCertifications.Back = level1Zone
		product.JacketCertifications.Chest = level1Zone
		product.JacketCertifications.FitsAirbag = true
	}
}

func applySeedDataToPants(i int, product *entities.Product) {
	product.RevzillaPriceCents = product.MSRPCents + 10000
	product.RevzillaBuyURL = fmt.Sprintf("http://www.testdata.com/revzilla/%d", i)
	if i%2 == 0 {
		level2Zone := &entities.CEImpactZone{IsLevel2: true, IsApproved: true, IsEmpty: false}
		product.PantsCertifications.Hip = level2Zone
		product.PantsCertifications.Knee = level2Zone
		product.PantsCertifications.Tailbone = level2Zone
	} else if i%3 == 0 {
		level1Zone := &entities.CEImpactZone{IsLevel2: false, IsApproved: false, IsEmpty: false}
		product.PantsCertifications.Hip = level1Zone
		product.PantsCertifications.Knee = level1Zone
		product.PantsCertifications.Tailbone = level1Zone
	}
}

func applySeedDataToBoots(i int, product *entities.Product) {
	product.RevzillaPriceCents = product.MSRPCents + 10000
	product.RevzillaBuyURL = fmt.Sprintf("http://www.testdata.com/revzilla/%d", i)
	if i%2 == 0 {
		level2Zone := &entities.CEImpactZone{IsLevel2: true, IsApproved: true, IsEmpty: false}
		product.BootsCertifications.Overall = level2Zone
	} else if i%3 == 0 {
		level1Zone := &entities.CEImpactZone{IsLevel2: false, IsApproved: false, IsEmpty: false}
		product.BootsCertifications.Overall = level1Zone
	}
}

func applySeedDataToGloves(i int, product *entities.Product) {
	product.RevzillaPriceCents = product.MSRPCents + 10000
	product.RevzillaBuyURL = fmt.Sprintf("http://www.testdata.com/revzilla/%d", i)
	if i%2 == 0 {
		level2Zone := &entities.CEImpactZone{IsLevel2: true, IsApproved: true, IsEmpty: false}
		product.GlovesCertifications.Overall = level2Zone
	} else if i%3 == 0 {
		level1Zone := &entities.CEImpactZone{IsLevel2: false, IsApproved: false, IsEmpty: false}
		product.GlovesCertifications.Overall = level1Zone
	}
}

// GetProductSeeds returns a sample list of product documents; these documents are used by GetProductSeedsSQLStatements() to seed the database with test data.
func GetProductSeeds() []*entities.Product {
	modelAliases := []*entities.ProductModelAlias{
		{
			ModelAlias: "ZZZZZ1234",
		},
		{
			ModelAlias:   "RF-1300",
			IsForDisplay: true,
		},
	}

	emptyJacket := &entities.Product{UUID: uuid.MustParse("66f61bf4-9098-4c13-b8ab-926943f6fd49"), ImageKey: mockHelmetImageURL, Manufacturer: "JacketManu4", Model: "Facturer3", MSRPCents: 89999, Type: "jacket", Subtype: "", SafetyPercentage: 4}
	seeds := []*entities.Product{
		{UUID: uuid.MustParse("2ef2e322-8b7c-4b11-8432-15d082f49f43"), ImageKey: mockHelmetImageURL, Manufacturer: "Shoei2", Model: "RF-8000", MSRPCents: 29900, Type: "helmet", Subtype: "full", SafetyPercentage: 1},
		{UUID: uuid.MustParse("9c7d1d1b-1c95-4d81-b6df-c6ad2445efa7"), ImageKey: mockHelmetImageURL, Manufacturer: "Shoei", Model: "RF-7000", MSRPCents: 49999, Type: "helmet", Subtype: "modular", SafetyPercentage: 2},
		{UUID: uuid.MustParse("1af5b06b-5f76-4df7-9908-563be1b646fe"), ImageKey: mockHelmetImageURL, Manufacturer: "Arai", Model: "Some Arai Helmet", MSRPCents: 79999, Type: "helmet", Subtype: "open", SafetyPercentage: 3},
		{UUID: uuid.MustParse("ffd16692-c5ac-4bad-8f99-7dab33ab96dc"), ImageKey: mockHelmetImageURL, Manufacturer: "NotAShoei", Model: "Model-RF", MSRPCents: 19999, Type: "helmet", Subtype: "half", SafetyPercentage: 4},
		{UUID: uuid.MustParse("8b9a42f5-0ead-49d3-b5d7-257a977cdb05"), ImageKey: mockHelmetImageURL, Manufacturer: "AGV", Model: "AyyGeeVee", MSRPCents: 29999, Type: "helmet", Subtype: "offroad", SafetyPercentage: 4},
		{UUID: uuid.MustParse("c486bbcf-dd73-45ea-be8a-fcef31ef956b"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer1", Model: "RF-SR1", MSRPCents: 29899, Type: "helmet", Subtype: "full", SafetyPercentage: 5},
		{UUID: uuid.MustParse("236fa58c-c606-4b77-b93c-2ef750feb514"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer2", Model: "RF-SR2", MSRPCents: 40012, Type: "helmet", Subtype: "modular", SafetyPercentage: 6},
		{UUID: uuid.MustParse("ebfa48c0-da4c-41e9-83e5-b58995282770"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer3", Model: "RF-SR", MSRPCents: 50099, Type: "helmet", Subtype: "open", SafetyPercentage: 7},
		{UUID: uuid.MustParse("09e88608-1156-4c25-9b01-efeede2646bd"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer4", Model: "RF-SR", MSRPCents: 60099, Type: "helmet", Subtype: "half", SafetyPercentage: 8},
		{UUID: uuid.MustParse("160addb2-6880-4f2e-a91c-498ce953c97c"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: 100},
		{UUID: uuid.MustParse("852af68b-d103-47c3-9fe8-b4f96f896ca9"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer6", Model: "RF-SR4", MSRPCents: 79999, Type: "helmet", Subtype: "modular", SafetyPercentage: 9},
		{UUID: uuid.MustParse("71313790-4844-474a-abb8-a0dc051abc81"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer7", Model: "RF-SR5", MSRPCents: 80099, Type: "helmet", Subtype: "open", SafetyPercentage: 10},
		{UUID: uuid.MustParse("d98b4e1d-9d60-4c4d-b0c6-94a90ba4410a"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer8", Model: "RF-SR6", MSRPCents: 89999, Type: "helmet", Subtype: "half", SafetyPercentage: 11},
		{UUID: uuid.MustParse("d26bc29c-371d-4abd-bca2-dd1f9a31b057"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer9", Model: "RF-SR7", MSRPCents: 90099, Type: "helmet", Subtype: "full", SafetyPercentage: 12},
		{UUID: uuid.MustParse("5330628b-c91b-4de7-90b8-e09d05038a1c"), ImageKey: "", Manufacturer: "Manufacturer10", Model: "RF-SR8", MSRPCents: 0, Type: "helmet", Subtype: "modular", SafetyPercentage: 1},
		{UUID: uuid.MustParse("58c794e2-2ae2-41c9-a16d-03bab94c5229"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer11", Model: "RF-SR9", MSRPCents: 100099, Type: "helmet", Subtype: "open", SafetyPercentage: 13},
		{UUID: uuid.MustParse("a123f97f-3048-4716-bd73-5b8577f6e664"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer12", Model: "RF-SR10", MSRPCents: 100299, Type: "helmet", Subtype: "half", SafetyPercentage: 14},
		{UUID: uuid.MustParse("a51d8208-75fb-4601-9b0b-7315b03ab155"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer13", Model: "RF-SR11", MSRPCents: 110099, Type: "helmet", Subtype: "offroad", SafetyPercentage: 15},
		{UUID: uuid.MustParse("5341f6cb-0b6f-4400-8a59-41cd514133b3"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer14", Model: "RF-SR12", MSRPCents: 120099, Type: "helmet", Subtype: "full", SafetyPercentage: 1},
		{UUID: uuid.MustParse("0fc1831b-b04b-42e6-a733-797c0905d066"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer15", Model: "RF-SR13", MSRPCents: 133001, Type: "helmet", Subtype: "modular", ModelAliases: modelAliases, SafetyPercentage: 0},
		{UUID: uuid.MustParse("1b0b38ad-4a60-454a-b701-bb884848e139"), ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer16", Model: "RF-SR14", MSRPCents: 133002, Type: "helmet", Subtype: "full", ModelAliases: modelAliases, SafetyPercentage: 0, IsDiscontinued: true},
		{UUID: uuid.MustParse("5e2398ff-5a08-40bc-9eaa-4b3c3de93475"), ImageKey: mockHelmetImageURL, Manufacturer: "JacketManu1", Model: "Facturer1", MSRPCents: 59999, Type: "jacket", Subtype: "", SafetyPercentage: 1},
		{UUID: uuid.MustParse("1d5f5500-b64f-4ca4-81fd-a2f4bad6bc72"), ImageKey: mockHelmetImageURL, Manufacturer: "JacketManu2", Model: "Facturer2", MSRPCents: 69999, Type: "jacket", Subtype: "", SafetyPercentage: 2},
		{UUID: uuid.MustParse("65f61bf4-9098-4c13-b8ab-926943f6fd49"), ImageKey: mockHelmetImageURL, Manufacturer: "JacketManu3", Model: "Facturer3", MSRPCents: 79999, Type: "jacket", Subtype: "", SafetyPercentage: 3},
		{UUID: uuid.MustParse("645f7e82-a19f-466d-9c34-bcb45018d1d5"), ImageKey: mockHelmetImageURL, Manufacturer: "PantsManu4", Model: "Facturer4", MSRPCents: 89999, Type: "pants", Subtype: "", SafetyPercentage: 4},
		{UUID: uuid.MustParse("0887a04b-0cbc-4946-8ee6-7fcf95fbfc6d"), ImageKey: mockHelmetImageURL, Manufacturer: "PantsManu5", Model: "Facturer5", MSRPCents: 89999, Type: "pants", Subtype: "", SafetyPercentage: 5},
		{UUID: uuid.MustParse("4912ed04-043b-45d5-8ccd-a98437573953"), ImageKey: mockHelmetImageURL, Manufacturer: "BootsManu6", Model: "Facturer6", MSRPCents: 99998, Type: "boots", Subtype: "", SafetyPercentage: 6},
		{UUID: uuid.MustParse("8e798a63-a556-4ee6-b09f-0d68bf605505"), ImageKey: mockHelmetImageURL, Manufacturer: "BootsManu7", Model: "Facturer7", MSRPCents: 99997, Type: "boots", Subtype: "", SafetyPercentage: 7},
		{UUID: uuid.MustParse("d0842709-3dcc-43df-993f-52fc1c3f7cd6"), ImageKey: mockHelmetImageURL, Manufacturer: "GlovesManu8", Model: "Facturer8", MSRPCents: 109999, Type: "gloves", Subtype: "", SafetyPercentage: 8},
		{UUID: uuid.MustParse("36de7153-9960-46df-a5eb-66f6e34123e1"), ImageKey: mockHelmetImageURL, Manufacturer: "GlovesManu9", Model: "Facturer9", MSRPCents: 119999, Type: "gloves", Subtype: "", SafetyPercentage: 9},
		emptyJacket,
	}

	for i := 0; i < len(seeds); i++ {
		if seeds[i].Type == "helmet" {
			applySeedDataToHelmet(i, seeds[i])
		} else if seeds[i].Type == "jacket" {
			applySeedDataToJacket(i, seeds[i])
		} else if seeds[i].Type == "pants" {
			applySeedDataToPants(i, seeds[i])
		} else if seeds[i].Type == "boots" {
			applySeedDataToBoots(i, seeds[i])
		} else if seeds[i].Type == "gloves" {
			applySeedDataToGloves(i, seeds[i])
		}

		seeds[i].UpdateSearchPrice()
	}

	emptyJacket.JacketCertifications.Shoulder = nil
	emptyJacket.JacketCertifications.Elbow = nil
	emptyJacket.JacketCertifications.Back = nil
	emptyJacket.JacketCertifications.Chest = nil
	emptyJacket.JacketCertifications.FitsAirbag = false

	return seeds
}

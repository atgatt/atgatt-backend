package seeds

import (
	"crashtested-backend/persistence/entities"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const mockHelmetImageURL = "https://sharp.dft.gov.uk/wp-content/uploads/2017/03/shoei-x-spirit-lll.jpg"

// GetProductSeedsSQLStatements returns an array of INSERT statements that target each of the product seed structs. Used to import test data into the database for automated tests, local development.
func GetProductSeedsSQLStatements() []string {
	productSeeds := GetProductSeeds()

	statements := []string{}
	for _, product := range productSeeds {
		documentJSONBytes, _ := json.Marshal(product)
		documentJSONString := string(documentJSONBytes)
		formattedInsertStatement := fmt.Sprintf("insert into products (uuid, document, created_at_utc, updated_at_utc) values ('%s', '%s', (now() at time zone 'utc'), null);", product.UUID.String(), documentJSONString)
		statements = append(statements, formattedInsertStatement)
	}
	return statements
}

// GetProductSeeds returns a sample list of product documents; these documents are used by GetProductSeedsSQLStatements() to seed the database with test data.
func GetProductSeeds() []*entities.ProductDocument {
	uuids := []string{
		"2ef2e322-8b7c-4b11-8432-15d082f49f43", "55e620cb-4eb3-46d7-a612-d8bf55088494", "0e78d74a-da19-4015-a76a-703a37d02503", "7321fc5c-596c-4b63-be0c-0d7af3fd78cc", "a23b4567-40bf-4761-ae19-00101223b124",
		"dbd3b9cb-253b-449d-a72b-ce0d62231d82", "455a8746-7e92-4f42-a2db-f653cce0e2dd", "c79f1957-6403-4316-82bd-e7dd79dc5682", "a1afdbeb-d551-4a1a-873a-8ad16a8800dc", "9a2ad6c7-553f-4a59-957a-c9f875651e99",
		"f8c57db1-f7f3-42ba-934f-bd30d5d31531", "912fbebc-1e42-46c2-bc1c-10666c724a21", "9f501018-e9c4-448e-89c9-8f48b571baa3", "90c2895c-ed20-483c-8a4e-6c41b6e6498f", "13131da7-fab3-42fe-9cce-7c7903fe5f8a",
		"ac1ae9ef-22b0-41c0-8401-84f6b3eb5ff7", "9ee16a4a-0dde-4628-83a5-ebecf8978165", "e67730e6-8134-4717-b3ca-21122b9c3c4d", "bbf2d99e-b21b-406b-adb5-200cec4c5766", "47365987-8e22-45dc-804f-58bc68497b62",
	}

	seeds := []*entities.ProductDocument{
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Shoei2", Model: "RF-8000", MSRPCents: 29900, Type: "helmet", Subtype: "full", SafetyPercentage: 1},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Shoei", Model: "RF-7000", MSRPCents: 49999, Type: "helmet", Subtype: "modular", SafetyPercentage: 2},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Arai", Model: "Some Arai Helmet", MSRPCents: 79999, Type: "helmet", Subtype: "open", SafetyPercentage: 3},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "NotAShoei", Model: "Model-RF", MSRPCents: 19999, Type: "helmet", Subtype: "half", SafetyPercentage: 4},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "AGV", Model: "AyyGeeVee", MSRPCents: 29999, Type: "helmet", Subtype: "offroad", SafetyPercentage: 4},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer1", Model: "RF-SR1", MSRPCents: 29899, Type: "helmet", Subtype: "full", SafetyPercentage: 5},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer2", Model: "RF-SR2", MSRPCents: 40012, Type: "helmet", Subtype: "modular", SafetyPercentage: 6},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer3", Model: "RF-SR", MSRPCents: 50099, Type: "helmet", Subtype: "open", SafetyPercentage: 7},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer4", Model: "RF-SR", MSRPCents: 60099, Type: "helmet", Subtype: "half", SafetyPercentage: 8},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer5", Model: "RF-SR3", MSRPCents: 70099, Type: "helmet", Subtype: "full", SafetyPercentage: 100},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer6", Model: "RF-SR4", MSRPCents: 79999, Type: "helmet", Subtype: "modular", SafetyPercentage: 9},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer7", Model: "RF-SR5", MSRPCents: 80099, Type: "helmet", Subtype: "open", SafetyPercentage: 10},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer8", Model: "RF-SR6", MSRPCents: 89999, Type: "helmet", Subtype: "half", SafetyPercentage: 11},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer9", Model: "RF-SR7", MSRPCents: 90099, Type: "helmet", Subtype: "full", SafetyPercentage: 12},
		&entities.ProductDocument{ImageKey: "", Manufacturer: "Manufacturer10", Model: "RF-SR8", MSRPCents: 0, Type: "helmet", Subtype: "modular", SafetyPercentage: 1},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer11", Model: "RF-SR9", MSRPCents: 100099, Type: "helmet", Subtype: "open", SafetyPercentage: 13},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer12", Model: "RF-SR10", MSRPCents: 100299, Type: "helmet", Subtype: "half", SafetyPercentage: 14},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer13", Model: "RF-SR11", MSRPCents: 110099, Type: "helmet", Subtype: "offroad", SafetyPercentage: 15},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer14", Model: "RF-SR12", MSRPCents: 120099, Type: "helmet", Subtype: "full", SafetyPercentage: 1},
		&entities.ProductDocument{ImageKey: mockHelmetImageURL, Manufacturer: "Manufacturer15", Model: "RF-SR13", MSRPCents: 133001, Type: "helmet", Subtype: "modular", ModelAlias: "RF-1300", SafetyPercentage: 0},
	}

	for i := 0; i < len(seeds); i++ {
		seeds[i].UUID, _ = uuid.Parse(uuids[i])
		seeds[i].RevzillaPriceCents = seeds[i].MSRPCents + 10000
		seeds[i].RevzillaBuyURL = fmt.Sprintf("http://www.testdata.com/revzilla/%d", i)
		if i%2 == 0 {
			seeds[i].Certifications.ECE = true
			seeds[i].Certifications.DOT = true
			seeds[i].Certifications.SHARP = &entities.SHARPCertificationDocument{}
			seeds[i].Certifications.SHARP.Stars = 4
			seeds[i].Certifications.SHARP.ImpactZoneRatings = &entities.SHARPImpactZoneRatingsDocument{}
			seeds[i].Certifications.SHARP.ImpactZoneRatings.Left = 4
			seeds[i].Certifications.SHARP.ImpactZoneRatings.Right = 3
			seeds[i].Certifications.SHARP.ImpactZoneRatings.Rear = 4
			seeds[i].Certifications.SHARP.ImpactZoneRatings.Top.Front = 3
			seeds[i].Certifications.SHARP.ImpactZoneRatings.Top.Rear = 5
			seeds[i].Certifications.SNELL = true
		} else if i%3 == 0 {
			seeds[i].Certifications.ECE = true
			seeds[i].Certifications.DOT = true
			seeds[i].Certifications.SHARP = &entities.SHARPCertificationDocument{}
			seeds[i].Certifications.SHARP.Stars = 3
			seeds[i].Certifications.SHARP.ImpactZoneRatings = &entities.SHARPImpactZoneRatingsDocument{}
			seeds[i].Certifications.SHARP.ImpactZoneRatings.Left = 1
			seeds[i].Certifications.SHARP.ImpactZoneRatings.Right = 1
			seeds[i].Certifications.SHARP.ImpactZoneRatings.Rear = 2
			seeds[i].Certifications.SHARP.ImpactZoneRatings.Top.Front = 2
			seeds[i].Certifications.SHARP.ImpactZoneRatings.Top.Rear = 3
			seeds[i].Certifications.SNELL = true
		} else {
			seeds[i].Certifications.ECE = false
			seeds[i].Certifications.DOT = false
			seeds[i].Certifications.SHARP = nil
			seeds[i].Certifications.SNELL = false
			seeds[i].RevzillaBuyURL = ""
			seeds[i].RevzillaPriceCents = 0
		}

		seeds[i].UpdateSearchPrice()
	}

	return seeds
}

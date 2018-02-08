package seeds

import (
	"crashtested-backend/persistence/entities"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const MockHelmetImageUrl = "https://sharp.dft.gov.uk/wp-content/uploads/2017/03/shoei-x-spirit-lll.jpg"

func GetProductSeedsSqlStatements() []string {
	productSeeds := GetProductSeeds()

	statements := []string{}
	for _, product := range productSeeds {
		documentJsonBytes, _ := json.Marshal(product)
		documentJsonString := string(documentJsonBytes)
		formattedInsertStatement := fmt.Sprintf("insert into products (uuid, document, created_at_utc, updated_at_utc) values ('%s', '%s', (now() at time zone 'utc'), null);", product.UUID.String(), documentJsonString)
		statements = append(statements, formattedInsertStatement)
	}
	return statements
}

func GetProductSeeds() []*entities.ProductDocument {
	uuids := []string{
		"2ef2e322-8b7c-4b11-8432-15d082f49f43", "55e620cb-4eb3-46d7-a612-d8bf55088494", "0e78d74a-da19-4015-a76a-703a37d02503", "7321fc5c-596c-4b63-be0c-0d7af3fd78cc", "a23b4567-40bf-4761-ae19-00101223b124",
		"dbd3b9cb-253b-449d-a72b-ce0d62231d82", "455a8746-7e92-4f42-a2db-f653cce0e2dd", "c79f1957-6403-4316-82bd-e7dd79dc5682", "a1afdbeb-d551-4a1a-873a-8ad16a8800dc", "9a2ad6c7-553f-4a59-957a-c9f875651e99",
		"f8c57db1-f7f3-42ba-934f-bd30d5d31531", "912fbebc-1e42-46c2-bc1c-10666c724a21", "9f501018-e9c4-448e-89c9-8f48b571baa3", "90c2895c-ed20-483c-8a4e-6c41b6e6498f", "13131da7-fab3-42fe-9cce-7c7903fe5f8a",
		"ac1ae9ef-22b0-41c0-8401-84f6b3eb5ff7", "9ee16a4a-0dde-4628-83a5-ebecf8978165", "e67730e6-8134-4717-b3ca-21122b9c3c4d", "bbf2d99e-b21b-406b-adb5-200cec4c5766", "47365987-8e22-45dc-804f-58bc68497b62",
	}

	seeds := []*entities.ProductDocument{
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Shoei2", Model: "RF-8000", PriceInUsdMultiple: 29900, Type: "helmet", Subtype: "full"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Shoei", Model: "RF-7000", PriceInUsdMultiple: 49999, Type: "helmet", Subtype: "modular"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Arai", Model: "Some Arai Helmet", PriceInUsdMultiple: 79999, Type: "helmet", Subtype: "open"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "NotAShoei", Model: "Model-RF", PriceInUsdMultiple: 19999, Type: "helmet", Subtype: "half"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "AGV", Model: "AyyGeeVee", PriceInUsdMultiple: 29999, Type: "helmet", Subtype: "offroad"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer1", Model: "RF-SR1", PriceInUsdMultiple: 29899, Type: "helmet", Subtype: "full"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer2", Model: "RF-SR2", PriceInUsdMultiple: 40012, Type: "helmet", Subtype: "modular"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer3", Model: "RF-SR", PriceInUsdMultiple: 50099, Type: "helmet", Subtype: "open"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer4", Model: "RF-SR", PriceInUsdMultiple: 60099, Type: "helmet", Subtype: "half"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer5", Model: "RF-SR3", PriceInUsdMultiple: 70099, Type: "helmet", Subtype: "full"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer6", Model: "RF-SR4", PriceInUsdMultiple: 79999, Type: "helmet", Subtype: "modular"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer7", Model: "RF-SR5", PriceInUsdMultiple: 80099, Type: "helmet", Subtype: "open"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer8", Model: "RF-SR6", PriceInUsdMultiple: 89999, Type: "helmet", Subtype: "half"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer9", Model: "RF-SR7", PriceInUsdMultiple: 90099, Type: "helmet", Subtype: "full"},
		&entities.ProductDocument{ImageURL: "", Manufacturer: "Manufacturer10", Model: "RF-SR8", PriceInUsdMultiple: 0, Type: "helmet", Subtype: "modular"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer11", Model: "RF-SR9", PriceInUsdMultiple: 100099, Type: "helmet", Subtype: "open"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer12", Model: "RF-SR10", PriceInUsdMultiple: 100299, Type: "helmet", Subtype: "half"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer13", Model: "RF-SR11", PriceInUsdMultiple: 110099, Type: "helmet", Subtype: "offroad"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer14", Model: "RF-SR12", PriceInUsdMultiple: 120099, Type: "helmet", Subtype: "full"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Manufacturer15", Model: "RF-SR13", PriceInUsdMultiple: 133001, Type: "helmet", Subtype: "modular", ModelAlias: "RF-1300"},
	}

	for i := 0; i < len(seeds); i++ {
		seeds[i].UUID, _ = uuid.Parse(uuids[i])
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
		}

	}

	return seeds
}

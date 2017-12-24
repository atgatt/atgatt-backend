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
	seeds := []*entities.ProductDocument{
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Shoei2", Model: "RF-SR", PriceInUsd: "399.99", Type: "helmet", Subtype: "fullface"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Shoei", Model: "RF-1200", PriceInUsd: "499.99", Type: "helmet", Subtype: "fullface"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Arai", Model: "Signet-X", PriceInUsd: "799.99", Type: "helmet", Subtype: "fullface"},
		&entities.ProductDocument{ImageURL: MockHelmetImageUrl, Manufacturer: "Shoei", Model: "RF-SR", PriceInUsd: "399.99", Type: "helmet", Subtype: "fullface"}}

	seeds[0].UUID, _ = uuid.Parse("f83d5b3f-160e-49d5-bfe6-7143be91ee6d")
	seeds[0].Certifications.DOT = true
	seeds[0].Certifications.ECE = true
	seeds[0].Certifications.SHARP.Stars = 5
	seeds[0].Certifications.SHARP.ImpactZoneRatings.Left = 1
	seeds[0].Certifications.SHARP.ImpactZoneRatings.Right = 3
	seeds[0].Certifications.SHARP.ImpactZoneRatings.Rear = 2
	seeds[0].Certifications.SHARP.ImpactZoneRatings.Top.Front = 4
	seeds[0].Certifications.SHARP.ImpactZoneRatings.Top.Rear = 5
	seeds[0].Certifications.SNELL = true

	seeds[1].UUID, _ = uuid.Parse("b7f5fd24-dd35-4b1c-a7ce-f642c5d873bd")
	seeds[1].Certifications.DOT = true
	seeds[1].Certifications.ECE = true
	seeds[1].Certifications.SHARP.Stars = 4
	seeds[1].Certifications.SHARP.ImpactZoneRatings.Left = 5
	seeds[1].Certifications.SHARP.ImpactZoneRatings.Right = 4
	seeds[1].Certifications.SHARP.ImpactZoneRatings.Rear = 3
	seeds[1].Certifications.SHARP.ImpactZoneRatings.Top.Front = 2
	seeds[1].Certifications.SHARP.ImpactZoneRatings.Top.Rear = 1
	seeds[1].Certifications.SNELL = true

	seeds[2].UUID, _ = uuid.Parse("0c2636d9-7efe-4450-8f92-4577fe58f642")
	seeds[2].Certifications.DOT = true
	seeds[2].Certifications.ECE = true
	seeds[2].Certifications.SHARP.Stars = 3
	seeds[2].Certifications.SHARP.ImpactZoneRatings.Left = 5
	seeds[2].Certifications.SHARP.ImpactZoneRatings.Right = 5
	seeds[2].Certifications.SHARP.ImpactZoneRatings.Rear = 5
	seeds[2].Certifications.SHARP.ImpactZoneRatings.Top.Front = 5
	seeds[2].Certifications.SHARP.ImpactZoneRatings.Top.Rear = 5
	seeds[2].Certifications.SNELL = true

	seeds[3].UUID, _ = uuid.Parse("df733953-a18d-477f-8fa3-c558f6576e15")
	seeds[3].Certifications.DOT = true
	seeds[3].Certifications.ECE = true
	seeds[3].Certifications.SHARP.Stars = 2
	seeds[3].Certifications.SHARP.ImpactZoneRatings.Left = 1
	seeds[3].Certifications.SHARP.ImpactZoneRatings.Right = 1
	seeds[3].Certifications.SHARP.ImpactZoneRatings.Rear = 1
	seeds[3].Certifications.SHARP.ImpactZoneRatings.Top.Front = 2
	seeds[3].Certifications.SHARP.ImpactZoneRatings.Top.Rear = 1
	seeds[3].Certifications.SNELL = true

	return seeds
}

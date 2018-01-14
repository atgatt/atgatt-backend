package repositories

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/queries"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type ProductRepository struct {
	ConnectionString string
}

func (self *ProductRepository) FilterProducts(query *queries.FilterProductsQuery) ([]entities.ProductDocument, error) {
	db, err := sqlx.Open("postgres", self.ConnectionString)
	defer db.Close()

	if err != nil {
		return nil, err
	}

	queryParams := make(map[string]interface{})
	whereCriteria := `where document->>'type' = :type `
	queryParams["type"] = "helmet" // TODO: this is hardcoded for now
	queryParams["start"] = query.Start
	queryParams["limit"] = query.Limit

	orderByExpression := query.Order.Field
	// TODO: find a cleaner way to do this
	if orderByExpression == "document->>'priceInUsd'" {
		orderByExpression = "to_number((document->>'priceInUsd'), '9999.99')"
	}
	queryParams["order_by"] = query.Order.Field

	var orderByDirection string
	if query.Order.Descending {
		orderByDirection = "desc"
	} else {
		orderByDirection = "asc"
	}

	if len(query.UsdPriceRange) == 2 {
		lowPrice := query.UsdPriceRange[0]
		highPrice := query.UsdPriceRange[1]
		queryParams["low_price"] = lowPrice
		queryParams["high_price"] = highPrice
		whereCriteria += "and to_number((document->>'priceInUsd'), '9999.99') between :low_price and :high_price "
	}

	if len(query.Subtypes) > 0 {
		queryParams["subtypes"] = query.Subtypes
		whereCriteria += "and document->>'subtype' in (:subtypes) "
	}

	if query.Manufacturer != "" {
		queryParams["manufacturer"] = query.Manufacturer
		whereCriteria += "and document->>'manufacturer' ilike (:manufacturer || '%') "
	}

	if query.Model != "" {
		queryParams["model"] = query.Model
		whereCriteria += "and (document->>'model' ilike (:model || '%') or document->>'modelAlias' ilike (:model || '%')) " // TODO: may need to optimize this query once the dataset grows larger, OR across multiple columns is likely not sargable
	}

	sharpCert := query.Certifications.SHARP
	if sharpCert != nil {
		whereCriteria += "and document->'certifications'->>'SHARP' is not null "
		if sharpCert.Stars > 0 {
			queryParams["minimum_SHARP_stars"] = query.Certifications.SHARP.Stars
			whereCriteria += "and to_number((document->'certifications'->'SHARP'->>'stars'), '9') >= :minimum_SHARP_stars "
		}

		if sharpCert.ImpactZoneMinimums.Left > 0 {
			queryParams["left_impact_zone_minimum"] = sharpCert.ImpactZoneMinimums.Left
			whereCriteria += "and to_number((document->'certifications'->'SHARP'->'impactZoneRatings'->>'left'), '9') >= :left_impact_zone_minimum "
		}

		if sharpCert.ImpactZoneMinimums.Rear > 0 {
			queryParams["rear_impact_zone_minimum"] = sharpCert.ImpactZoneMinimums.Rear
			whereCriteria += "and to_number((document->'certifications'->'SHARP'->'impactZoneRatings'->>'rear'), '9') >= :rear_impact_zone_minimum "
		}

		if sharpCert.ImpactZoneMinimums.Right > 0 {
			queryParams["right_impact_zone_minimum"] = sharpCert.ImpactZoneMinimums.Right
			whereCriteria += "and to_number((document->'certifications'->'SHARP'->'impactZoneRatings'->>'right'), '9') >= :right_impact_zone_minimum "
		}

		if sharpCert.ImpactZoneMinimums.Top.Front > 0 {
			queryParams["top_front_impact_zone_minimum"] = sharpCert.ImpactZoneMinimums.Top.Front
			whereCriteria += "and to_number((document->'certifications'->'SHARP'->'impactZoneRatings'->'top'->>'front'), '9') >= :top_front_impact_zone_minimum "
		}

		if sharpCert.ImpactZoneMinimums.Top.Rear > 0 {
			queryParams["top_rear_impact_zone_minimum"] = sharpCert.ImpactZoneMinimums.Top.Rear
			whereCriteria += "and to_number((document->'certifications'->'SHARP'->'impactZoneRatings'->'top'->>'rear'), '9') >= :top_rear_impact_zone_minimum "
		}
	}

	if query.Certifications.SNELL {
		whereCriteria += "and document->'certifications'->>'SNELL' = 'true' "
	}

	if query.Certifications.ECE {
		whereCriteria += "and document->'certifications'->>'ECE' = 'true' "
	}

	if query.Certifications.DOT {
		whereCriteria += "and document->'certifications'->>'DOT' = 'true' "
	}

	productDocuments := make([]entities.ProductDocument, 0)
	originalSqlQueryString := fmt.Sprintf(`select document from products
											%s
											order by %s %s
											offset :start limit :limit`, whereCriteria, orderByExpression, orderByDirection)

	// Converts :arguments to ? arguments so that we can preprocess the query
	preProcessedSqlQueryString, args, err := sqlx.Named(originalSqlQueryString, queryParams)
	if err != nil {
		return nil, err
	}

	// Converts `where in` statements to work with the SQL driver
	preProcessedSqlQueryString, args, err = sqlx.In(preProcessedSqlQueryString, args...)
	if err != nil {
		return nil, err
	}

	// Converts ? arguments back to positional ($0, $1, $2, etc) arguments so that they can be executed in the DB.
	preProcessedSqlQueryString = db.Rebind(preProcessedSqlQueryString)
	rows, err := db.Query(preProcessedSqlQueryString, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		productJsonString := &[]byte{}
		productDocument := &entities.ProductDocument{}
		err := rows.Scan(productJsonString)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(*productJsonString, productDocument)
		if err != nil {
			return nil, err
		}
		productDocuments = append(productDocuments, *productDocument)
	}

	return productDocuments, nil
}

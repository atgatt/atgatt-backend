package repositories

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/queries"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	// Importing the PQ driver because we need to run queries!
	_ "github.com/lib/pq"
)

// ProductRepository contains functions that are used to do CRUD operations on Products in the database
type ProductRepository struct {
	ConnectionString string
}

// GetByModel returns a single product where the manufacturer and model matches
func (r *ProductRepository) GetByModel(manufacturer string, model string) (*entities.ProductDocument, error) {
	query := &queries.FilterProductsQuery{Start: 0, Limit: 1, Manufacturer: manufacturer, Model: model}
	query.Order.Field = "id"

	filteredProducts, err := r.FilterProducts(query)
	if err != nil {
		return nil, err
	}

	if len(filteredProducts) == 0 {
		return nil, nil
	}

	if len(filteredProducts) > 1 {
		return nil, errors.New("An unexpected number of products were returned")
	}

	return &filteredProducts[0], nil
}

// GetAllPaged queries the database for all products without any filters, within the range of start and limit. This function is useful for calling functions that to do batch operations on all products in the DB.
func (r *ProductRepository) GetAllPaged(start int, limit int) ([]entities.ProductDocument, error) {
	query := &queries.FilterProductsQuery{Start: start, Limit: limit}
	query.Order.Field = "id"

	filteredProducts, err := r.FilterProducts(query)
	if err != nil {
		return nil, err
	}

	return filteredProducts, nil
}

// UpdateProduct replaces the product in the DB with the supplied product, where the product's UUID matches the one supplied
func (r *ProductRepository) UpdateProduct(product *entities.ProductDocument) error {
	db, err := sqlx.Open("postgres", r.ConnectionString)
	defer db.Close()
	if err != nil {
		return err
	}

	productJSONBytes, err := json.Marshal(product)
	if err != nil {
		return err
	}

	_, err = db.NamedExec(`update products set 
								document = :document, 
								updated_at_utc = (now() at time zone 'utc') 
							where uuid = :uuid`, map[string]interface{}{
		"document": string(productJSONBytes),
		"uuid":     product.UUID,
	})

	if err != nil {
		return err
	}

	return nil
}

// CreateProduct creates a product with the given fields by first converting it to json, and then dumping the json into a column in the DB.
func (r *ProductRepository) CreateProduct(product *entities.ProductDocument) error {
	db, err := sqlx.Open("postgres", r.ConnectionString)
	defer db.Close()
	if err != nil {
		return err
	}

	productJSONBytes, err := json.Marshal(product)
	if err != nil {
		return err
	}

	_, err = db.NamedExec("insert into products (uuid, document, created_at_utc, updated_at_utc) values (:uuid, :document, (now() at time zone 'utc'), null);", map[string]interface{}{
		"document": string(productJSONBytes),
		"uuid":     product.UUID,
	})

	if err != nil {
		return err
	}

	return nil
}

// FilterProducts is a method that ANDs a bunch of query parameters together and returns a list of matching products, or an error if there was a problem executing the query.
func (r *ProductRepository) FilterProducts(query *queries.FilterProductsQuery) ([]entities.ProductDocument, error) {
	db, err := sqlx.Open("postgres", r.ConnectionString)
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
	if orderByExpression == "document->>'priceInUsdMultiple'" {
		orderByExpression = "cast((document->>'priceInUsdMultiple') as int)"
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
		whereCriteria += "and cast((document->>'priceInUsdMultiple') as int) between :low_price and :high_price "
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
	originalSQLQueryString := fmt.Sprintf(`select document from products
											%s
											order by %s %s
											offset :start limit :limit`, whereCriteria, orderByExpression, orderByDirection)

	// Converts :arguments to ? arguments so that we can preprocess the query
	preProcessedSQLQueryString, args, err := sqlx.Named(originalSQLQueryString, queryParams)
	if err != nil {
		return nil, err
	}

	// Converts `where in` statements to work with the SQL driver
	preProcessedSQLQueryString, args, err = sqlx.In(preProcessedSQLQueryString, args...)
	if err != nil {
		return nil, err
	}

	// Converts ? arguments back to positional ($0, $1, $2, etc) arguments so that they can be executed in the DB.
	preProcessedSQLQueryString = db.Rebind(preProcessedSQLQueryString)
	rows, err := db.Query(preProcessedSQLQueryString, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		productJSONString := &[]byte{}
		productDocument := &entities.ProductDocument{}
		err := rows.Scan(productJSONString)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(*productJSONString, productDocument)
		if err != nil {
			return nil, err
		}
		productDocuments = append(productDocuments, *productDocument)
	}

	return productDocuments, nil
}

package repositories

import (
	"crashtested-backend/persistence/entities"
	"crashtested-backend/persistence/queries"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// ProductRepository contains functions that are used to do CRUD operations on Products in the database
type ProductRepository struct {
	DB *sqlx.DB
}

func getOneProductFromRows(rows *sqlx.Rows) (*entities.Product, error) {
	defer rows.Close()

	productDocuments := []*entities.Product{}
	for rows.Next() {
		productJSONBytesPtr := &[]byte{}
		productDocument := &entities.Product{}
		err := rows.Scan(productJSONBytesPtr)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(*productJSONBytesPtr, productDocument)
		if err != nil {
			return nil, err
		}
		productDocuments = append(productDocuments, productDocument)
	}

	if len(productDocuments) == 0 {
		return nil, nil
	}

	if len(productDocuments) > 1 {
		return nil, errors.New("An unexpected number of products were returned")
	}

	return productDocuments[0], nil
}

// GetByExternalID returns a single product where the external ID matches
func (r *ProductRepository) GetByExternalID(externalID string) (*entities.Product, error) {
	rows, err := r.DB.NamedQuery("select document from products where document->>'externalID' = :externalID", map[string]interface{}{
		"externalID": externalID,
	})
	if err != nil {
		return nil, err
	}

	return getOneProductFromRows(rows)
}

// GetByModel returns a single product where the manufacturer and model matches
func (r *ProductRepository) GetByModel(manufacturer string, model string, productType string) (*entities.Product, error) {
	rows, err := r.DB.NamedQuery("select document from products where document->>'manufacturer' = :manufacturer and document->>'model' = :model and document->>'type' = :type", map[string]interface{}{
		"manufacturer": manufacturer,
		"model":        model,
		"type":         productType,
	})
	if err != nil {
		return nil, err
	}

	return getOneProductFromRows(rows)
}

// GetAllPaged queries the database for all products without prices, within the range of start and limit.
func (r *ProductRepository) GetAllPaged(start int, limit int) ([]entities.Product, error) {
	query := &queries.FilterProductsQuery{Start: start, Limit: limit}
	query.Order.Field = "id"

	filteredProducts, err := r.FilterProducts(query)
	if err != nil {
		return nil, err
	}

	return filteredProducts, nil
}

// GetAllModelAliases returns all the model aliases in the database
func (r *ProductRepository) GetAllModelAliases() ([]*entities.ProductModelAlias, error) {

	productModelAliases := []*entities.ProductModelAlias{}
	err := r.DB.Select(&productModelAliases, "select manufacturer, model, model_alias as modelalias, is_for_display as isfordisplay from product_model_aliases")
	if err != nil {
		return nil, err
	}

	return productModelAliases, nil
}

// GetAllManufacturerAliases returns all the manufacturer aliases in the database
func (r *ProductRepository) GetAllManufacturerAliases() ([]entities.ProductManufacturerAlias, error) {
	productManufacturerAliases := []entities.ProductManufacturerAlias{}
	err := r.DB.Select(&productManufacturerAliases, "select manufacturer, manufacturer_alias as manufactureralias from product_manufacturer_aliases")
	if err != nil {
		return nil, err
	}

	return productManufacturerAliases, nil
}

// UpdateProduct replaces the product in the DB with the supplied product, where the product's UUID matches the one supplied
func (r *ProductRepository) UpdateProduct(product *entities.Product) error {
	if product == nil {
		return errors.New("product must be defined")
	}

	product.UpdateSearchPrice()
	productJSONBytes, err := json.Marshal(product)
	if err != nil {
		return err
	}

	_, err = r.DB.NamedExec(`update products set 
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
func (r *ProductRepository) CreateProduct(product *entities.Product) error {
	if product == nil {
		return errors.New("product must be defined")
	}

	product.UpdateSearchPrice()
	productJSONBytes, err := json.Marshal(product)
	if err != nil {
		return err
	}

	_, err = r.DB.NamedExec("insert into products (uuid, document, created_at_utc, updated_at_utc) values (:uuid, :document, (now() at time zone 'utc'), null);", map[string]interface{}{
		"document": string(productJSONBytes),
		"uuid":     product.UUID,
	})

	if err != nil {
		return err
	}

	return nil
}

func applyCEImpactZoneParams(zoneKey string, ceImpactZoneParams *queries.CEImpactZoneQueryParams, whereCriteria *strings.Builder) {
	if ceImpactZoneParams != nil {
		if ceImpactZoneParams.IsLevel2 {
			(*whereCriteria).WriteString(fmt.Sprintf("and document->%s->>'isLevel2' = 'true' ", zoneKey))
		}

		if ceImpactZoneParams.IsApproved {
			(*whereCriteria).WriteString(fmt.Sprintf("and document->%s->>'isApproved' = 'true' ", zoneKey))
		}

		if ceImpactZoneParams.IsEmpty {
			(*whereCriteria).WriteString(fmt.Sprintf("and document->%s->>'isEmpty' = 'true' ", zoneKey))
		}
	}
}

// FilterProducts is a method that ANDs a bunch of query parameters together and returns a list of matching products, or an error if there was a problem executing the query.
func (r *ProductRepository) FilterProducts(query *queries.FilterProductsQuery) ([]entities.Product, error) {
	queryParams := make(map[string]interface{})
	var whereCriteria strings.Builder
	whereCriteria.WriteString("where 1=1 ")

	queryParams["type"] = "helmet" // TODO: this is hardcoded for now
	queryParams["start"] = query.Start
	queryParams["limit"] = query.Limit

	orderByExpression := query.Order.Field

	// TODO: find a cleaner way to do this
	if orderByExpression == "document->>'searchPriceCents'" {
		orderByExpression = "cast((document->>'searchPriceCents') as int)"
	} else if orderByExpression == "document->>'safetyPercentage'" {
		orderByExpression = "cast((document->>'safetyPercentage') as int)"
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
		whereCriteria.WriteString("and cast((document->>'searchPriceCents') as int) between :low_price and :high_price ")
	}

	if query.Type != "" {
		queryParams["type"] = query.Type
		whereCriteria.WriteString(`and document->>'type' = :type `)
	}

	if len(query.Subtypes) > 0 {
		queryParams["subtypes"] = query.Subtypes
		whereCriteria.WriteString("and document->>'subtype' in (:subtypes) ")
	}

	if query.Manufacturer != "" {
		queryParams["manufacturer"] = query.Manufacturer
		whereCriteria.WriteString("and document->>'manufacturer' ilike (:manufacturer || '%') ")
	}

	// TODO: This will not scale for a large number of rows!
	// This particular query is suboptimal as it selects across multiple columns, and the exists query cannot be indexed.
	// Need to ETL this entire table to elasticsearch or AWS Search for efficient queries once the table size grows
	if query.Model != "" {
		queryParams["model"] = query.Model
		// Find rows where the model matches, or one of the aliases starts with the model
		whereCriteria.WriteString(`and (document->>'model' ilike (:model || '%') or exists(
			select 1 
			from jsonb_array_elements(cast(document->>'modelAliases' as jsonb)) elem
			where elem->>'modelAlias' ilike (:model || '%')
		))`)
	}

	if query.HelmetCertifications != nil {
		sharpCert := query.HelmetCertifications.SHARP
		if sharpCert != nil {
			whereCriteria.WriteString("and document->'helmetCertifications'->>'SHARP' is not null ")
			if sharpCert.Stars > 0 {
				queryParams["minimum_SHARP_stars"] = query.HelmetCertifications.SHARP.Stars
				whereCriteria.WriteString("and to_number((document->'helmetCertifications'->'SHARP'->>'stars'), '9') >= :minimum_SHARP_stars ")
			}

			if sharpCert.ImpactZoneMinimums.Left > 0 {
				queryParams["left_impact_zone_minimum"] = sharpCert.ImpactZoneMinimums.Left
				whereCriteria.WriteString("and to_number((document->'helmetCertifications'->'SHARP'->'impactZoneRatings'->>'left'), '9') >= :left_impact_zone_minimum ")
			}

			if sharpCert.ImpactZoneMinimums.Rear > 0 {
				queryParams["rear_impact_zone_minimum"] = sharpCert.ImpactZoneMinimums.Rear
				whereCriteria.WriteString("and to_number((document->'helmetCertifications'->'SHARP'->'impactZoneRatings'->>'rear'), '9') >= :rear_impact_zone_minimum ")
			}

			if sharpCert.ImpactZoneMinimums.Right > 0 {
				queryParams["right_impact_zone_minimum"] = sharpCert.ImpactZoneMinimums.Right
				whereCriteria.WriteString("and to_number((document->'helmetCertifications'->'SHARP'->'impactZoneRatings'->>'right'), '9') >= :right_impact_zone_minimum ")
			}

			if sharpCert.ImpactZoneMinimums.Top.Front > 0 {
				queryParams["top_front_impact_zone_minimum"] = sharpCert.ImpactZoneMinimums.Top.Front
				whereCriteria.WriteString("and to_number((document->'helmetCertifications'->'SHARP'->'impactZoneRatings'->'top'->>'front'), '9') >= :top_front_impact_zone_minimum ")
			}

			if sharpCert.ImpactZoneMinimums.Top.Rear > 0 {
				queryParams["top_rear_impact_zone_minimum"] = sharpCert.ImpactZoneMinimums.Top.Rear
				whereCriteria.WriteString("and to_number((document->'helmetCertifications'->'SHARP'->'impactZoneRatings'->'top'->>'rear'), '9') >= :top_rear_impact_zone_minimum ")
			}
		}

		if query.HelmetCertifications.SNELL {
			whereCriteria.WriteString("and document->'helmetCertifications'->>'SNELL' = 'true' ")
		}

		if query.HelmetCertifications.ECE {
			whereCriteria.WriteString("and document->'helmetCertifications'->>'ECE' = 'true' ")
		}

		if query.HelmetCertifications.DOT {
			whereCriteria.WriteString("and document->'helmetCertifications'->>'DOT' = 'true' ")
		}
	}

	if query.JacketCertifications != nil {
		applyCEImpactZoneParams("'jacketCertifications'->'shoulder'", query.JacketCertifications.Shoulder, &whereCriteria)
		applyCEImpactZoneParams("'jacketCertifications'->'elbow'", query.JacketCertifications.Elbow, &whereCriteria)
		applyCEImpactZoneParams("'jacketCertifications'->'back'", query.JacketCertifications.Back, &whereCriteria)
		applyCEImpactZoneParams("'jacketCertifications'->'chest'", query.JacketCertifications.Chest, &whereCriteria)

		if query.JacketCertifications.FitsAirbag {
			whereCriteria.WriteString("and document->'jacketCertifications'->>'fitsAirbag' = 'true'")
		}
	}
	

	if query.ExcludeDiscontinued {
		whereCriteria.WriteString("and document->>'isDiscontinued' = 'false' ")
	}

	productDocuments := []entities.Product{}
	originalSQLQueryString := fmt.Sprintf(`select document from products
											%s
											order by %s %s,
													 id asc
											offset :start limit :limit`, whereCriteria.String(), orderByExpression, orderByDirection)

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
	preProcessedSQLQueryString = r.DB.Rebind(preProcessedSQLQueryString)
	rows, err := r.DB.Query(preProcessedSQLQueryString, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		productJSONBytesPtr := &[]byte{}
		productDocument := &entities.Product{}
		err := rows.Scan(productJSONBytesPtr)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(*productJSONBytesPtr, productDocument)
		if err != nil {
			return nil, err
		}
		productDocuments = append(productDocuments, *productDocument)
	}

	return productDocuments, nil
}

package repositories

import (
	"crashtested-backend/persistence/entities"
	"errors"

	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
)

// ProductSetRepository contains functions that are used to do CRUD operations on Products in the database
type ProductSetRepository struct {
	DB *sqlx.DB
}

func getUUIDFromRowsOrNil(rows *sqlx.Rows) (uuid.UUID, error) {
	defer rows.Close()
	uuids := []uuid.UUID{}
	for rows.Next() {
		var uuidCreated uuid.UUID
		err := rows.Scan(&uuidCreated)
		if err != nil {
			return uuid.Nil, err
		}

		uuids = append(uuids, uuidCreated)
	}

	if len(uuids) == 0 {
		return uuid.Nil, nil
	}

	if len(uuids) > 1 {
		return uuid.Nil, errors.New("An unexpected number of UUIDs were returned")
	}

	return uuids[0], nil
}

func getOneProductSetFromRows(rows *sqlx.Rows) (*entities.ProductSet, error) {
	defer rows.Close()

	productSets := []*entities.ProductSet{}
	for rows.Next() {
		productSet := &entities.ProductSet{}
		err := rows.Scan(productSet)
		if err != nil {
			return nil, err
		}
		productSets = append(productSets, productSet)
	}

	if len(productSets) == 0 {
		return nil, ErrEntityNotFound
	}

	if len(productSets) > 1 {
		return nil, errors.New("An unexpected number of product sets were returned")
	}

	return productSets[0], nil
}

// GetProductSetWithProductsByUUID gets all of the given products for a product set
func (r *ProductSetRepository) GetProductSetWithProductsByUUID(uuid uuid.UUID) (*entities.ProductSet, error) {
	rows, err := r.DB.NamedQuery(`select 
								phelmet.document helmetProduct, 
								pjacket.document jacketProduct, 
								ppants.document pantsProduct,
								pboots.document bootsProduct, 
								pgloves.document glovesProduct
							from product_sets ps
							left join products phelmet on phelmet.id = ps.helmet_product_id 
							left join products pjacket on pjacket.id = ps.jacket_product_id
							left join products ppants on ppants.id = ps.pants_product_id
							left join products pboots on pboots.id = ps.boots_product_id
							left join products pgloves on pgloves.id = ps.gloves_product_id
							where ps.uuid = :uuid`, map[string]interface{}{
		"uuid": uuid,
	})

	if err != nil {
		return nil, err
	}

	return getOneProductSetFromRows(rows)
}

// GetByUUID gets the given productset by its UUID or returns null if one was not found.
func (r *ProductSetRepository) GetByUUID(uuid uuid.UUID) (*entities.ProductSet, error) {
	rows, err := r.DB.NamedQuery("select uuid, \"name\", description, helmet_product_id, jacket_product_id, pants_product_id, boots_product_id, gloves_product_id, created_at_utc, created_by from product_sets where uuid = :uuid", map[string]interface{}{
		"uuid": uuid,
	})

	if err != nil {
		return nil, err
	}

	return getOneProductSetFromRows(rows)
}

// GetMatchingProductSetUUID gets the product set's UUID with the exact same set of products if it exists, otherwise null
func (r *ProductSetRepository) GetMatchingProductSetUUID(productSet *entities.ProductSet) (uuid.UUID, error) {
	paramsMap := map[string]interface{}{
		"helmet_product_id": productSet.HelmetProductID,
		"jacket_product_id": productSet.JacketProductID,
		"pants_product_id":  productSet.PantsProductID,
		"boots_product_id":  productSet.BootsProductID,
		"gloves_product_id": productSet.GlovesProductID,
	}

	rows, err := r.DB.NamedQuery(`select uuid from product_sets 
								 where 
									 helmet_product_id is not distinct from :helmet_product_id and 
									 jacket_product_id is not distinct from :jacket_product_id and 
									 pants_product_id is not distinct from :pants_product_id and 
									 boots_product_id is not distinct from :boots_product_id and 
									 gloves_product_id is not distinct from :gloves_product_id`, paramsMap)

	if err != nil {
		return uuid.Nil, err
	}

	return getUUIDFromRowsOrNil(rows)
}

// Create creates the given productset, returning its UUID for the frontend to use.
func (r *ProductSetRepository) Create(productSet *entities.ProductSet) (uuid.UUID, error) {
	paramsMap := map[string]interface{}{
		"uuid":              uuid.New(),
		"name":              productSet.Name,
		"description":       productSet.Description,
		"helmet_product_id": productSet.HelmetProductID,
		"jacket_product_id": productSet.JacketProductID,
		"pants_product_id":  productSet.PantsProductID,
		"boots_product_id":  productSet.BootsProductID,
		"gloves_product_id": productSet.GlovesProductID,
	}

	rows, err := r.DB.NamedQuery(`insert into product_sets
							(uuid, "name", description, helmet_product_id, jacket_product_id, pants_product_id, boots_product_id, gloves_product_id, created_at_utc, created_by) 
							values 
							(:uuid, :name, :description, :helmet_product_id, :jacket_product_id, :pants_product_id, :boots_product_id, :gloves_product_id, (now() at time zone 'utc'), 'SYSTEM_USER')
							returning uuid`, paramsMap)
	if err != nil {
		return uuid.Nil, err
	}

	uuidCreated, err := getUUIDFromRowsOrNil(rows)
	if err != nil {
		return uuid.Nil, err
	}

	if uuidCreated == uuid.Nil {
		return uuid.Nil, errors.New("the database did not return a uuid for a newly created product set")
	}

	return uuidCreated, nil
}

package repositories

import (
	"crashtested-backend/persistence/entities"

	"github.com/google/uuid"

	"github.com/jmoiron/sqlx"
)

// ProductSetRepository contains functions that are used to do CRUD operations on Products in the database
type ProductSetRepository struct {
	DB *sqlx.DB
}

// UpsertProductSet upserts the given productset, returning its UUID for the frontend to use.
func (r *ProductSetRepository) UpsertProductSet(productSet *entities.ProductSet) (*uuid.UUID, error) {
	namedStmt, err := r.DB.PrepareNamed(`insert into product_sets
							(uuid, \"name\", description, helmet_product_id, jacket_product_id, pants_product_id, boots_product_id, gloves_product_id, created_at_utc, created_by) 
							values 
							(:uuid, :name, :description, :helmet_product_id, :jacket_product_id, :pants_product_id, :boots_product_id, :gloves_product_id, (now() at time zone 'utc'), 'SYSTEM_USER')
							on conflict(product_sets_name_helmet_product_id_jacket_product_id_pants_key) do update 
								set updated_at = (now() at time zone 'utc'), updated_by = 'SYSTEM_USER' 
							returning uuid`)
	if err != nil {
		return nil, err
	}

	var uuid *uuid.UUID
	err = namedStmt.Get(uuid, map[string]interface{}{
		"uuid":              productSet.UUID,
		"name":              productSet.Name,
		"description":       productSet.Description,
		"helmet_product_id": productSet.HelmetProductID,
		"jacket_product_id": productSet.JacketProductID,
		"pants_product_id":  productSet.PantsProductID,
		"boots_product_id":  productSet.BootsProductID,
		"gloves_product_id": productSet.GlovesProductID,
	})

	if err != nil {
		return nil, err
	}

	return uuid, nil
}

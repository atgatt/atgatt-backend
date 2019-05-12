package repositories

import (
	"github.com/jmoiron/sqlx"
)

// ManufacturerRepository contains functions used to return product manufacturer data from the database
type ManufacturerRepository struct {
	DB *sqlx.DB
}

// GetAll returns all of the product manufacturer names from the database - this should be refactored to not return everything at once once we have > 100 manufacturers
func (r *ManufacturerRepository) GetAll() ([]string, error) {
	rows, err := r.DB.Queryx("select name from manufacturers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	manufacturers := []string{}
	for rows.Next() {
		manufacturer := ""
		err := rows.Scan(&manufacturer)
		if err != nil {
			return nil, err
		}

		manufacturers = append(manufacturers, manufacturer)
	}

	return manufacturers, nil
}

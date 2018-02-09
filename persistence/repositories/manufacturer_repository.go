package repositories

import (
	"github.com/jmoiron/sqlx"
	// Importing the PQ driver because we need to run queries!
	_ "github.com/lib/pq"
)

// ManufacturerRepository contains functions used to return product manufacturer data from the database
type ManufacturerRepository struct {
	ConnectionString string
}

// GetAll returns all of the product manufacturer names from the database - this should be refactored to not return everything at once once we have > 100 manufacturers
func (r *ManufacturerRepository) GetAll() ([]string, error) {
	db, err := sqlx.Open("postgres", r.ConnectionString)
	defer db.Close()
	if err != nil {
		return nil, err
	}

	rows, err := db.Queryx("select name from manufacturers")

	manufacturers := make([]string, 0)
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

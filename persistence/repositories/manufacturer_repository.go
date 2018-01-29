package repositories

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type ManufacturerRepository struct {
	ConnectionString string
}

func (self *ManufacturerRepository) GetAll() ([]string, error) {
	db, err := sqlx.Open("postgres", self.ConnectionString)
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

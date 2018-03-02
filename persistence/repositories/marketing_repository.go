package repositories

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	// Importing the PQ driver because we need to run queries!
	_ "github.com/lib/pq"
)

// MarketingRepository contains functions used for querying/updating marketing data, such as user email addresses
type MarketingRepository struct {
	ConnectionString string
}

// CreateMarketingEmail inserts a new marketing email into the database
func (r *MarketingRepository) CreateMarketingEmail(email string) error {
	db, err := sqlx.Open("postgres", r.ConnectionString)
	defer db.Close()
	if err != nil {
		return err
	}

	_, err = db.Exec("insert into marketing_emails (email, created_at_utc) values ($1, (now() at time zone 'utc'))", email)
	if err != nil {
		return err
	}

	return nil
}

// MarketingEmailExists returns true if the email passed is already in the database, false otherwise
func (r *MarketingRepository) MarketingEmailExists(email string) (bool, error) {
	db, err := sqlx.Open("postgres", r.ConnectionString)
	defer db.Close()
	if err != nil {
		return false, err
	}

	exists := false
	err = db.QueryRowx("select exists(select id from marketing_emails where email = $1)", email).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return exists, nil
}

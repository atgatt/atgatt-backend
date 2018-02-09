package repositories

import (
	"sort"

	"github.com/jmoiron/sqlx"

	// Importing the PQ driver because we need to run queries!
	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

// MigrationsRepository contains functions used to query the status of database migrations
type MigrationsRepository struct {
	ConnectionString string
}

// GetLatestMigrationVersion returns the version identifier of the last-run migration, i.e. "20180101-doSomething.sql"
func (r *MigrationsRepository) GetLatestMigrationVersion() (string, error) {
	db, err := sqlx.Open("postgres", r.ConnectionString)
	defer db.Close()

	if err != nil {
		return "", err
	}

	migrations, err := migrate.GetMigrationRecords(db.DB, "postgres")
	if err != nil {
		return "", err
	}

	numMigrations := len(migrations)
	versionStrings := make([]string, 0)
	for _, migrationVersion := range migrations {
		versionStrings = append(versionStrings, migrationVersion.Id)
	}

	sort.Strings(versionStrings)
	if numMigrations <= 0 {
		return "", nil
	}

	return migrations[numMigrations-1].Id, nil
}

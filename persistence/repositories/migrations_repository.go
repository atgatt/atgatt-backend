package repositories

import (
	"sort"

	"github.com/jmoiron/sqlx"
	"github.com/rubenv/sql-migrate"
)

// MigrationsRepository contains functions used to query the status of database migrations
type MigrationsRepository struct {
	DB *sqlx.DB
}

// GetLatestMigrationVersion returns the version identifier of the last-run migration, i.e. "20180101-doSomething.sql"
func (r *MigrationsRepository) GetLatestMigrationVersion() (string, error) {
	migrations, err := migrate.GetMigrationRecords(r.DB.DB, "postgres")
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

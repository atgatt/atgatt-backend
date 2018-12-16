package repositories

import (
	"errors"
	"sort"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
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

// ApplySeeds applies several arbitrary SQL statements to the database as seed data. NOTE: This method is dangerous as it executes arbitrary SQL and should only be used by test code!
func (r *MigrationsRepository) ApplySeeds(seeds []string) error {
	if seeds == nil || len(seeds) <= 0 {
		return errors.New("seeds cannot be empty")
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	for _, seedSQL := range seeds {
		_, err := tx.Exec(seedSQL)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

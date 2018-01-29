package repositories

import (
	"sort"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

type MigrationsRepository struct {
	ConnectionString string
}

func (self *MigrationsRepository) GetLatestMigrationVersion() (string, error) {
	db, err := sqlx.Open("postgres", self.ConnectionString)
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
	} else {
		return migrations[numMigrations-1].Id, nil
	}
}

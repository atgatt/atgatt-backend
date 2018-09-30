package helpers

import (
	"database/sql"
	// Importing the PostgreSQL driver with side effects because we need to call sql.Open() to run migrations
	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

// RunMigrations runs all database migrations in migrationDirectory against the PostgreSQL instance running at connectionString.
func RunMigrations(connectionString string, migrationsDirectory string) error {
	var dbConn *sql.DB
	migrationsSource := &migrate.FileMigrationSource{Dir: migrationsDirectory}
	dbConn, err := sql.Open("postgres", connectionString)
	defer dbConn.Close()

	if err != nil {
		return err
	}

	logrus.Info("Running migrations...")
	_, err = migrate.Exec(dbConn, "postgres", migrationsSource, migrate.Up)
	if err == nil {
		logrus.Info("Successfully ran migrations.")
	}
	return err
}

package helpers

import (
	"github.com/jmoiron/sqlx"

	// Importing the PostgreSQL driver with side effects because we need to call sql.Open() to run migrations
	_ "github.com/jackc/pgx/v4/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

// RunMigrations runs all database migrations in migrationDirectory against the PostgreSQL instance running at connectionString.
func RunMigrations(connectionString string, migrationsDirectory string) error {
	var dbConn *sqlx.DB
	migrationsSource := &migrate.FileMigrationSource{Dir: migrationsDirectory}
	dbConn, err := sqlx.Open("pgx", connectionString)
	if err != nil {
		return err
	}
	defer dbConn.Close()

	logrus.Info("Running migrations...")
	_, err = migrate.Exec(dbConn.DB, "postgres", migrationsSource, migrate.Up)
	if err == nil {
		logrus.Info("Successfully ran migrations.")
	}
	return err
}

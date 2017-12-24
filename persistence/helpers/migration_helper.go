package helpers

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

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
	logrus.Info("Successfully ran migrations.")
	return err
}

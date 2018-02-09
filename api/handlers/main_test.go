package handlers

import (
	"crashtested-backend/api/configuration"
	"crashtested-backend/persistence/helpers"
	"crashtested-backend/seeds"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

const APIBaseURL string = "http://localhost:5001"
const IntegrationTestDatabaseName string = "crashtested"
const DatabaseServerConnectionString string = "postgres://postgres:password@localhost:5432/?sslmode=disable"

var DatabaseConnectionString = fmt.Sprintf("postgres://postgres:password@localhost:5432/%s?sslmode=disable", IntegrationTestDatabaseName)

func WaitFor(label string, isRunningFunc func() (bool, error)) bool {
	const MaxTimeToWait time.Duration = 10 * time.Second

	var isRunning bool
	var err error
	timeWaited := time.Duration(0)
	for !isRunning && (timeWaited < MaxTimeToWait) {
		logrus.Infof("Waiting for %s to come online...", label)
		isRunning, err = isRunningFunc()
		if err != nil {
			logrus.Errorf("%s returned an error: %s", label, err.Error())
		}

		if !isRunning {
			logrus.Infof("Trying again after 200ms...")
			time.Sleep(200 * time.Millisecond)
			timeWaited += 200 * time.Millisecond
		}
	}
	return isRunning
}

func WaitForAPI() bool {
	return WaitFor("api", func() (bool, error) {
		resp, err := http.Get(APIBaseURL)
		if err != nil {
			return false, err
		}

		return resp.StatusCode == http.StatusOK, err
	})
}

func WaitForMigrations() bool {
	return WaitFor("database", func() (bool, error) {
		dbServerConn, err := sql.Open("postgres", DatabaseServerConnectionString)
		defer dbServerConn.Close()
		if err != nil {
			return false, err
		}
		_, _ = dbServerConn.Query(fmt.Sprintf("select pg_terminate_backend(pid) from pg_stat_activity where datname = '%s'", IntegrationTestDatabaseName))
		_, err = dbServerConn.Exec(fmt.Sprintf("drop database if exists %s", IntegrationTestDatabaseName))
		if err != nil {
			return false, err
		}
		_, err = dbServerConn.Exec(fmt.Sprintf("create database %s", IntegrationTestDatabaseName))
		if err != nil {
			return false, err
		}

		err = helpers.RunMigrations(DatabaseConnectionString, "../../persistence/migrations")
		if err != nil {
			return false, err
		}
		logrus.Info("Running seeds...")

		productSeedsSQL := seeds.GetProductSeedsSQLStatements()
		seedMigrationsSource := &migrate.MemoryMigrationSource{
			Migrations: []*migrate.Migration{
				&migrate.Migration{Up: productSeedsSQL, Down: []string{"select 1;"}, Id: "0-seeds"}, // NOTE: using 0-seeds because of strange sorting rules present in sql-migrate. this allows the seeds to run after all other migrations
			},
		}

		var appliedSeedMigrations int
		dbConn, err := sql.Open("postgres", DatabaseConnectionString)
		defer dbConn.Close()
		if err != nil {
			return false, err
		}
		appliedSeedMigrations, err = migrate.Exec(dbConn, "postgres", seedMigrationsSource, migrate.Up)

		return (appliedSeedMigrations > 0), err
	})
}

func TestMain(m *testing.M) {
	logrus.Info("Starting server and database migrations...")
	migrationsRan := WaitForMigrations()
	defaultConfiguration := configuration.GetDefaultConfiguration()
	defaultConfiguration.AppEnvironment = "integration-tests"
	server := Server{Port: ":5001", Name: "crashtested-api", Version: "integration-tests-version", BuildNumber: "integration-tests-build", CommitHash: "integration-tests-commit", Configuration: defaultConfiguration}
	go server.StartAndBlock()

	apiStarted := WaitForAPI()

	statusCode := -1
	if apiStarted && migrationsRan {
		logrus.Info("Server is running! Starting tests.")
		statusCode = m.Run()
	}

	logrus.Info("Tests finished. Closing resources...")
	server.Stop()
	logrus.Info("Done. Exiting...")
	os.Exit(statusCode)
}

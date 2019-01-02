package controllers_test

import (
	"crashtested-backend/api"
	"crashtested-backend/api/settings"
	"crashtested-backend/persistence/helpers"
	"crashtested-backend/persistence/repositories"
	"crashtested-backend/seeds"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

const APIBaseURL string = "http://localhost:5001"
const IntegrationTestDatabaseName string = "crashtested_integrationtests"
const TestDatabaseServerConnectionString string = "postgres://postgres:password@localhost:5432/?sslmode=disable"
const MaxTimeToWait time.Duration = 10 * time.Second

var TestDatabaseConnectionString = fmt.Sprintf("postgres://postgres:password@localhost:5432/%s?sslmode=disable", IntegrationTestDatabaseName)

func WaitFor(label string, isRunningFunc func() error) error {
	timeWaited := time.Duration(0)
	var err error = nil
	for timeWaited < MaxTimeToWait {
		logrus.Infof("Waiting for %s to come online...", label)
		err = isRunningFunc()
		if err == nil {
			return nil
		}

		logrus.Errorf("%s returned an error: %s", label, err.Error())
		logrus.Infof("Trying again after 200ms...")
		time.Sleep(200 * time.Millisecond)
		timeWaited += 200 * time.Millisecond
	}
	return err
}

func WaitForAPI() error {
	return WaitFor("api", func() error {
		resp, err := http.Get(APIBaseURL)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return errors.New("The endpoint did not return 200 OK")
		}

		return err
	})
}

func WaitForMigrations() error {
	return WaitFor("database", func() error {
		dbServerConn, err := sql.Open("postgres", TestDatabaseServerConnectionString)
		defer dbServerConn.Close()
		if err != nil {
			return err
		}
		_, _ = dbServerConn.Query(fmt.Sprintf("select pg_terminate_backend(pid) from pg_stat_activity where datname = '%s'", IntegrationTestDatabaseName))
		_, err = dbServerConn.Exec(fmt.Sprintf("drop database if exists %s", IntegrationTestDatabaseName))
		if err != nil {
			return err
		}
		_, err = dbServerConn.Exec(fmt.Sprintf("create database %s", IntegrationTestDatabaseName))
		if err != nil {
			return err
		}

		err = helpers.RunMigrations(TestDatabaseConnectionString, "../../../persistence/migrations")
		if err != nil {
			return err
		}
		logrus.Info("Running seeds...")

		productSeeds, err := seeds.GetProductSeedsSQLStatements()
		if err != nil {
			return err
		}

		seedsSQL := append(productSeeds, seeds.GetMarketingEmailSeedsSQLStatements()...)
		dbConn, err := sqlx.Open("postgres", TestDatabaseConnectionString)
		if err != nil {
			return err
		}
		defer dbConn.Close()
		migrationsRepository := &repositories.MigrationsRepository{DB: dbConn}

		return migrationsRepository.ApplySeeds(seedsSQL)
	})
}

func TestMain(m *testing.M) {
	logrus.Info("Starting server and database migrations...")
	migrationsErr := WaitForMigrations()
	defaultSettings := settings.GetSettingsFromEnvironment()
	// Override the database in env vars with the test database
	defaultSettings.DatabaseConnectionString = TestDatabaseConnectionString
	defaultSettings.AppEnvironment = "integration-tests"
	server := api.Server{Port: ":5001", Name: "crashtested-api", Version: "integration-tests-version", BuildNumber: "integration-tests-build", CommitHash: "integration-tests-commit", Settings: defaultSettings}
	go server.Bootstrap()

	apiErr := WaitForAPI()

	statusCode := -1
	if apiErr == nil && migrationsErr == nil {
		logrus.Info("Server is running! Starting tests.")
		statusCode = m.Run()
	}

	logrus.Info("Tests finished. Exiting...")
	os.Exit(statusCode)
}

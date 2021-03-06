package controllers_test

import (
	"atgatt-backend/api"
	"atgatt-backend/api/settings"
	testHelpers "atgatt-backend/common/testing"
	"atgatt-backend/seeds"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const APIBaseURL string = "http://localhost:5001"
const IntegrationTestDatabaseName string = "atgatt_integrationtests"
const TestDatabaseServerConnectionString string = "postgres://postgres:password@localhost:5432/?sslmode=disable"
const MaxTimeToWait time.Duration = 10 * time.Second

var TestDatabaseConnectionString = fmt.Sprintf("postgres://postgres:password@localhost:5432/%s?sslmode=disable", IntegrationTestDatabaseName)

func TestMain(m *testing.M) {
	logrus.Info("Starting server and database migrations...")
	productSeeds, err := seeds.GetProductSeedsSQLStatements(seeds.GetProductSeeds())
	statusCode := -1
	if err != nil {
		logrus.WithError(err).Error("Failed to get product seeds")
		os.Exit(statusCode)
	}

	seedsSQL := append(productSeeds, seeds.GetMarketingEmailSeedsSQLStatements()...)
	migrationsErr := testHelpers.WaitForMigrations(TestDatabaseServerConnectionString, IntegrationTestDatabaseName, TestDatabaseConnectionString, "../../../persistence/migrations", seedsSQL, MaxTimeToWait)
	defaultSettings := settings.GetSettingsFromEnvironment()
	// Override the database in env vars with the test database
	defaultSettings.DatabaseConnectionString = TestDatabaseConnectionString
	defaultSettings.AppEnvironment = "integration-tests"
	defaultSettings.LogAPIRequests = false
	defaultSettings.Auth0Domain = "atgatt-staging.auth0.com"
	server := api.Server{Port: ":5001", Name: "atgatt-api", Version: "integration-tests-version", BuildNumber: "integration-tests-build", CommitHash: "integration-tests-commit", Settings: defaultSettings}
	go server.Bootstrap()

	apiErr := testHelpers.WaitForAPI(APIBaseURL, MaxTimeToWait)

	if apiErr == nil && migrationsErr == nil {
		logrus.Info("Server is running! Starting tests.")
		statusCode = m.Run()
	}

	logrus.Info("Tests finished. Exiting...")
	os.Exit(statusCode)
}

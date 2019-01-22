package jobs_test

import (
	testHelpers "crashtested-backend/common/testing"
	"crashtested-backend/seeds"
	"crashtested-backend/worker"
	"crashtested-backend/worker/settings"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

const APIBaseURL string = "http://localhost:5002"
const IntegrationTestDatabaseName string = "crashtested_integrationtests_worker"
const TestDatabaseServerConnectionString string = "postgres://postgres:password@localhost:5432/?sslmode=disable"
const MaxTimeToWait time.Duration = 10 * time.Second

var TestDatabaseConnectionString = fmt.Sprintf("postgres://postgres:password@localhost:5432/%s?sslmode=disable", IntegrationTestDatabaseName)

func TestMain(m *testing.M) {
	logrus.Info("Starting server and database migrations...")
	productSeeds, err := seeds.GetProductSeedsSQLStatements(seeds.GetRealisticProductSeeds())
	statusCode := -1
	if err != nil {
		logrus.WithError(err).Error("Failed to get product seeds")
		os.Exit(statusCode)
	}

	migrationsErr := testHelpers.WaitForMigrations(TestDatabaseServerConnectionString, IntegrationTestDatabaseName, TestDatabaseConnectionString, "../../persistence/migrations", productSeeds, MaxTimeToWait)
	defaultSettings := settings.GetSettingsFromEnvironment()
	// Override the database in env vars with the test database
	defaultSettings.DatabaseConnectionString = TestDatabaseConnectionString
	defaultSettings.AppEnvironment = "integration-tests"
	defaultSettings.UseSynchronousJobRunner = true
	server := &worker.Server{Port: ":5002", Name: "crashtested-worker", Version: "integration-tests-version", BuildNumber: "integration-tests-build", CommitHash: "integration-tests-commit", Settings: defaultSettings}
	go server.Bootstrap()

	apiErr := testHelpers.WaitForAPI(APIBaseURL, MaxTimeToWait)

	if apiErr == nil && migrationsErr == nil {
		logrus.Info("Server is running! Starting tests.")
		statusCode = m.Run()
	}

	logrus.Info("Tests finished. Exiting...")
	os.Exit(statusCode)
}

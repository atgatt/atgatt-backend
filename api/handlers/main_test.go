package handlers

import (
	"crashtested-backend/seeds"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

const ApiBaseUrl string = "http://localhost:5001"
const IntegrationTestDatabaseName string = "crashtested"
const DatabaseServerConnectionString string = "postgres://postgres:password@localhost:5432/?sslmode=disable"

var DatabaseConnectionString string = fmt.Sprintf("postgres://postgres:password@localhost:5432/%s?sslmode=disable", IntegrationTestDatabaseName)

func WaitFor(label string, isRunningFunc func() (bool, error)) bool {
	const MaxTimeToWait time.Duration = 10 * time.Second

	var isRunning bool
	var err error
	timeWaited := time.Duration(0)
	for !isRunning && (timeWaited < MaxTimeToWait) {
		fmt.Printf("Waiting for %s to come online...\n", label)
		isRunning, err = isRunningFunc()
		if err != nil {
			fmt.Printf("%s returned an error: %s\n", label, err.Error())
		}

		if !isRunning {
			fmt.Println("Trying again after 200ms...")
			time.Sleep(200 * time.Millisecond)
			timeWaited += 200 * time.Millisecond
		}
	}
	return isRunning
}

func WaitForApi() bool {
	return WaitFor("api", func() (bool, error) {
		resp, err := http.Get(ApiBaseUrl)
		return resp.StatusCode == http.StatusOK, err
	})
}

func WaitForMigrations() bool {
	return WaitFor("database", func() (bool, error) {
		migrationsSource := &migrate.FileMigrationSource{Dir: "../../persistence/migrations"}
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

		var dbConn *sql.DB
		dbConn, err = sql.Open("postgres", DatabaseConnectionString)
		defer dbConn.Close()

		if err != nil {
			return false, err
		}

		fmt.Println("Running migrations...")
		appliedMigrations, err := migrate.Exec(dbConn, "postgres", migrationsSource, migrate.Up)

		if err != nil {
			return false, err
		}
		fmt.Println("Successfully ran migrations. Running seeds...")

		productSeedsSql := seeds.GetProductSeedsSqlStatements()
		seedMigrationsSource := &migrate.MemoryMigrationSource{
			Migrations: []*migrate.Migration{
				&migrate.Migration{Up: productSeedsSql, Down: []string{"select 1;"}, Id: "0-seeds"}, // NOTE: using 0-seeds because of strange sorting rules present in sql-migrate. this allows the seeds to run after all other migrations
			},
		}

		var appliedSeedMigrations int
		appliedSeedMigrations, err = migrate.Exec(dbConn, "postgres", seedMigrationsSource, migrate.Up)

		return (appliedMigrations > 0 && appliedSeedMigrations > 0), err
	})
}

func TestMain(m *testing.M) {
	fmt.Println("Starting server and database migrations...")
	migrationsRan := WaitForMigrations()
	server := Server{Port: ":5001", Name: "crashtested-api", Version: "integrationtests", BuildNumber: "1337"}
	go server.StartAndBlock()

	apiStarted := WaitForApi()

	statusCode := -1
	if apiStarted && migrationsRan {
		fmt.Println("Server is running! Starting tests.")
		statusCode = m.Run()
	}

	fmt.Println("Tests finished. Closing resources...")
	server.Stop()
	fmt.Println("Done. Exiting...")
	os.Exit(statusCode)
}

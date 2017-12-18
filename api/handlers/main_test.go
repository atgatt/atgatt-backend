package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

const ApiBaseUrl string = "http://localhost:5001"
const DockerDatabaseConnectionString string = "postgres://postgres:password@localhost:5432/crashtested?sslmode=disable"
const DatabaseDockerImage string = "postgres:9.6-alpine"

func WaitFor(label string, isRunningFunc func() (bool, error)) bool {
	const MaxTimeToWait time.Duration = 10 * time.Second

	var isRunning bool
	var err error
	timeWaited := time.Duration(0)
	for !isRunning && (timeWaited < MaxTimeToWait) {
		fmt.Println(fmt.Sprintf("Waiting for %s to come online...", label))
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
		migrations := &migrate.FileMigrationSource{Dir: "../../persistence/migrations"}

		db, err := sql.Open("postgres", DockerDatabaseConnectionString)
		defer db.Close()

		if err != nil {
			return false, err
		}
		appliedMigrations, err := migrate.Exec(db, "postgres", migrations, migrate.Up)

		if err != nil {
			return false, err
		}

		return (appliedMigrations > 0), err
	})
}

func RunDockerCommand(label string, arg ...string) string {
	if os.Getenv("ENABLE_DOCKER_COMMANDS") == "true" {
		fmt.Println(fmt.Sprintf("Docker is enabled; %s", label))
		startPostgresCommand := exec.Command("docker", arg...)
		output, err := startPostgresCommand.Output()
		if err != nil {
			fmt.Println("Docker command failed to run!")
		}
		return strings.TrimSpace(string(output))
	} else {
		fmt.Println(fmt.Sprintf("Docker is disabled; not %s", label))
		return ""
	}
}

func StartDatabase() {
	exec.Command("/bin/sh", "../../cleanup-docker.sh").Run()
	RunDockerCommand("starting database container", "run", "--name", "automatedtestingdb", "-d", "-p", "5432:5432", "-e", "POSTGRES_PASSWORD=password", "-e", "POSTGRES_DB=crashtested", DatabaseDockerImage)
}

func StopDatabase() {
	RunDockerCommand("stopping database container", "stop", "automatedtestingdb")
}

func TestMain(m *testing.M) {
	fmt.Println("Starting server and database in the background...")
	StartDatabase()
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
	StopDatabase()
	server.Stop()
	fmt.Println("Done. Exiting...")
	os.Exit(statusCode)
}

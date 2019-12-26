package testing

import (
	"crashtested-backend/persistence/repositories"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"

	persistenceHelpers "crashtested-backend/persistence/helpers"

	"github.com/sirupsen/logrus"
)

// WaitFor retries execution of an arbitrary function and returns nil, or returns an error if maxTimeToWait is exceeded and none of the retries have succeeded
func WaitFor(label string, isRunningFunc func() error, maxTimeToWait time.Duration) error {
	timeWaited := time.Duration(0)
	var err error
	for timeWaited < maxTimeToWait {
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

// WaitForAPI calls WaitFor to determine when the API server is running
func WaitForAPI(apiBaseURL string, maxTimeToWait time.Duration) error {
	return WaitFor("api", func() error {
		resp, err := http.Get(apiBaseURL)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return errors.New("The endpoint did not return 200 OK")
		}

		return err
	}, maxTimeToWait)
}

// WaitForMigrations calls WaitFor to determine when migrations are done running
func WaitForMigrations(testDbServerConnString, testDbName, testDbConnString, migrationsPath string, seedsSQLStatements []string, maxTimeToWait time.Duration) error {
	return WaitFor("database", func() error {
		dbServerConn, err := sql.Open("pgx", testDbServerConnString)
		if err != nil {
			return err
		}
		defer dbServerConn.Close()
		_, _ = dbServerConn.Query(fmt.Sprintf("select pg_terminate_backend(pid) from pg_stat_activity where datname = '%s'", testDbName))
		_, err = dbServerConn.Exec(fmt.Sprintf("drop database if exists %s", testDbName))
		if err != nil {
			return err
		}
		_, err = dbServerConn.Exec(fmt.Sprintf("create database %s", testDbName))
		if err != nil {
			return err
		}

		err = persistenceHelpers.RunMigrations(testDbConnString, migrationsPath)
		if err != nil {
			return err
		}
		logrus.Info("Running seeds...")

		dbConn, err := sqlx.Open("pgx", testDbConnString)
		if err != nil {
			return err
		}
		defer dbConn.Close()

		migrationsRepository := &repositories.MigrationsRepository{DB: dbConn}
		return migrationsRepository.ApplySeeds(seedsSQLStatements)
	}, maxTimeToWait)
}

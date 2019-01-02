package main

import (
	"crashtested-backend/worker"
	"crashtested-backend/worker/settings"

	// Importing the PostgreSQL driver with side effects because we need to call sql.Open() to run queries
	_ "github.com/lib/pq"
)

func main() {
	server := &worker.Server{Port: ":80", Name: "crashtested-worker", Version: "1.0.0", BuildNumber: "{LOCAL-DEV-BUILD}", CommitHash: "{LOCAL-DEV-COMMIT}", Settings: settings.GetSettingsFromEnvironment()}
	server.Bootstrap()
}

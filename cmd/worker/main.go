package main

import (
	"atgatt-backend/worker"
	"atgatt-backend/worker/settings"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	server := &worker.Server{Port: ":" + port, Name: "atgatt-worker", Version: "1.0.0", BuildNumber: "{LOCAL-DEV-BUILD}", CommitHash: "{LOCAL-DEV-COMMIT}", Settings: settings.GetSettingsFromEnvironment()}
	server.Bootstrap()
}

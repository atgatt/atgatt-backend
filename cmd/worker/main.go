package main

import (
	"crashtested-backend/worker"
	"crashtested-backend/worker/settings"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	server := &worker.Server{Port: ":" + port, Name: "crashtested-worker", Version: "1.0.0", BuildNumber: "{LOCAL-DEV-BUILD}", CommitHash: "{LOCAL-DEV-COMMIT}", Settings: settings.GetSettingsFromEnvironment()}
	server.Bootstrap()
}

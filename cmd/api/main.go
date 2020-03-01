package main

import (
	"atgatt-backend/api"
	"atgatt-backend/api/settings"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	server := &api.Server{Port: ":" + port, Name: "atgatt-api", Version: "1.0.0", BuildNumber: "{LOCAL-DEV-BUILD}", CommitHash: "{LOCAL-DEV-COMMIT}", Settings: settings.GetSettingsFromEnvironment()}
	server.Bootstrap()
}

package main

import (
	"crashtested-backend/api"
	"crashtested-backend/api/settings"
)

func main() {
	server := &api.Server{Port: ":5000", Name: "crashtested-api", Version: "1.0.0", BuildNumber: "{LOCAL-DEV-BUILD}", CommitHash: "{LOCAL-DEV-COMMIT}", Settings: settings.GetSettingsFromEnvironment()}
	server.Bootstrap()
}

package settings

import (
	"os"
)

// Settings represents the environment variables needed to start the API
type Settings struct {
	DatabaseConnectionString string
	LogzioToken              string
	AppEnvironment           string
	LogAPIRequests           bool
}

// GetSettingsFromEnvironment returns a pointer to a Configuration struct with all of its values initialized from environment variables
func GetSettingsFromEnvironment() *Settings {
	return &Settings{
		AppEnvironment:           os.Getenv("APP_ENVIRONMENT"),
		DatabaseConnectionString: os.Getenv("DATABASE_CONNECTION_STRING"),
		LogzioToken:              os.Getenv("LOGZIO_TOKEN"),
		LogAPIRequests: true, // default this to true for deployed envs, can be overwritten by test code to avoid polluting the output
	}
}

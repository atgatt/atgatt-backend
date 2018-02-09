package configuration

import (
	"os"
)

// Configuration represents the environment variables needed to start the API
type Configuration struct {
	DatabaseConnectionString string
	LogzioToken              string
	AppEnvironment           string
}

// GetDefaultConfiguration returns a pointer to a Configuration struct with all of its values initialized from environment variables
func GetDefaultConfiguration() *Configuration {
	return &Configuration{
		AppEnvironment:           os.Getenv("APP_ENVIRONMENT"),
		DatabaseConnectionString: os.Getenv("DATABASE_CONNECTION_STRING"),
		LogzioToken:              os.Getenv("LOGZIO_TOKEN"),
	}
}

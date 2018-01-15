package configuration

import (
	"os"
)

type Configuration struct {
	DatabaseConnectionString string
	LogzioToken              string
	AppEnvironment           string
}

func GetDefaultConfiguration() *Configuration {
	return &Configuration{
		AppEnvironment:           os.Getenv("APP_ENVIRONMENT"),
		DatabaseConnectionString: os.Getenv("DATABASE_CONNECTION_STRING"),
		LogzioToken:              os.Getenv("LOGZIO_TOKEN"),
	}
}

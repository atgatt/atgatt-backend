package configuration

import (
	"os"
)

type Configuration struct {
	DatabaseConnectionString string
	LogzioToken              string
	AppEnvironment           string
	AmazonAssociates         amazonAssociatesConfiguration
}

type amazonAssociatesConfiguration struct {
	AccessKey   string
	SecretKey   string
	AssociateID string
}

func GetDefaultConfiguration() *Configuration {
	return &Configuration{
		AppEnvironment:           os.Getenv("APP_ENVIRONMENT"),
		DatabaseConnectionString: os.Getenv("DATABASE_CONNECTION_STRING"),
		LogzioToken:              os.Getenv("LOGZIO_TOKEN"),
		AmazonAssociates: amazonAssociatesConfiguration{
			AccessKey:   os.Getenv("AMAZON_ASSOCIATES_ACCESS_KEY"),
			SecretKey:   os.Getenv("AMAZON_ASSOCIATES_SECRET_KEY"),
			AssociateID: os.Getenv("AMAZON_ASSOCIATES_ID"),
		},
	}
}

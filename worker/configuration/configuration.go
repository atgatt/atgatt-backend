package configuration

import (
	"os"
)

// Configuration contains all of the environment variables needed to start background workers
type Configuration struct {
	DatabaseConnectionString string
	LogzioToken              string
	AppEnvironment           string
	AWS                      awsConfiguration
	CJAPIKey                 string
}

type awsConfiguration struct {
	AccessKey string
	SecretKey string
	S3Bucket  string
}

// GetDefaultConfiguration returns a configuration struct, initialized using environment variables
func GetDefaultConfiguration() *Configuration {
	return &Configuration{
		AppEnvironment:           os.Getenv("APP_ENVIRONMENT"),
		DatabaseConnectionString: os.Getenv("DATABASE_CONNECTION_STRING"),
		LogzioToken:              os.Getenv("LOGZIO_TOKEN"),
		AWS: awsConfiguration{
			AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			S3Bucket:  os.Getenv("AWS_S3_BUCKET"),
		},
		CJAPIKey: os.Getenv("CJ_API_KEY"),
	}
}

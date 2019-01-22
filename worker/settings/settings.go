package settings

import (
	"os"
)

// Settings contains all of the environment variables needed to start background workers
type Settings struct {
	DatabaseConnectionString string
	LogzioToken              string
	AppEnvironment           string
	AWS                      awsConfiguration
	CJAPIKey                 string
	UseSynchronousJobRunner  bool
}

type awsConfiguration struct {
	AccessKey     string
	SecretKey     string
	S3Bucket      string
	MinioEndpoint string
}

// GetSettingsFromEnvironment returns a configuration struct, initialized using environment variables
func GetSettingsFromEnvironment() *Settings {
	return &Settings{
		AppEnvironment:           os.Getenv("APP_ENVIRONMENT"),
		DatabaseConnectionString: os.Getenv("DATABASE_CONNECTION_STRING"),
		LogzioToken:              os.Getenv("LOGZIO_TOKEN"),
		AWS: awsConfiguration{
			AccessKey:     os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretKey:     os.Getenv("AWS_SECRET_ACCESS_KEY"),
			S3Bucket:      os.Getenv("AWS_S3_BUCKET"),
			MinioEndpoint: os.Getenv("MINIO_ENDPOINT"),
		},
		CJAPIKey: os.Getenv("CJ_API_KEY"),
	}
}

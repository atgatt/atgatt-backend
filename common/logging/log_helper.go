package helpers

import (
	"fmt"
	"os"

	"github.com/bshuster-repo/logruzio"
	"github.com/sirupsen/logrus"
)

// InitializeLogzio initializes a logrus -> logzio hook if logzioToken is defined, otherwise does nothing
func InitializeLogzio(logzioToken string, serviceName string, environmentName string, fields logrus.Fields) {
	if logzioToken != "" {
		logzioHook, err := logruzio.New(logzioToken, fmt.Sprintf("%s-%s", serviceName, environmentName), fields)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to start the API because the logger could not be initialized")
			os.Exit(-1)
		}
		logrus.AddHook(logzioHook)
	} else {
		logrus.Warn("LOGZIO_TOKEN was not set, so all application logs are going to stdout")
	}
}

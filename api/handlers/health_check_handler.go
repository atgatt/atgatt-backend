package handlers

import (
	"crashtested-backend/api/responses"
	"net/http"

	"github.com/labstack/echo"
)

type HealthCheckHandler struct {
	BuildNumber string
	Name        string
	Version     string
}

func (self *HealthCheckHandler) Healthcheck(context echo.Context) (err error) {
	healthCheckResponse := &responses.HealthCheckResponse{Name: self.Name, Version: self.Version}

	if len(self.BuildNumber) > 0 {
		healthCheckResponse.BuildNumber = self.BuildNumber
	}

	return context.JSON(http.StatusOK, healthCheckResponse)
}

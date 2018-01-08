package handlers

import (
	"crashtested-backend/api/responses"
	"crashtested-backend/persistence/repositories"
	"net/http"

	"github.com/labstack/echo"
)

type HealthCheckHandler struct {
	BuildNumber          string
	Name                 string
	Version              string
	MigrationsRepository *repositories.MigrationsRepository
}

func (self *HealthCheckHandler) Healthcheck(context echo.Context) (err error) {
	if context.Request().Method == http.MethodHead {
		return context.NoContent(http.StatusOK)
	}

	healthCheckResponse := &responses.HealthCheckResponse{Name: self.Name, Version: self.Version}
	healthCheckResponse.Database.CurrentVersion, err = self.MigrationsRepository.GetLatestMigrationVersion()
	if err != nil {
		return err
	}

	if len(self.BuildNumber) > 0 {
		healthCheckResponse.BuildNumber = self.BuildNumber
	}

	return context.JSON(http.StatusOK, healthCheckResponse)
}

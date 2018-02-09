package handlers

import (
	"crashtested-backend/api/responses"
	"crashtested-backend/persistence/repositories"
	"net/http"

	"github.com/labstack/echo"
)

// HealthCheckHandler contains methods used for automated healthchecks.
type HealthCheckHandler struct {
	BuildNumber          string
	CommitHash           string
	Name                 string
	Version              string
	MigrationsRepository *repositories.MigrationsRepository
}

// Healthcheck returns the API's current build number, database migration status, etc.
func (h *HealthCheckHandler) Healthcheck(context echo.Context) (err error) {
	if context.Request().Method == http.MethodHead {
		return context.NoContent(http.StatusOK)
	}

	healthCheckResponse := &responses.HealthCheckResponse{Name: h.Name, Version: h.Version, CommitHash: h.CommitHash}
	healthCheckResponse.Database.CurrentVersion, err = h.MigrationsRepository.GetLatestMigrationVersion()
	if err != nil {
		return err
	}

	if len(h.BuildNumber) > 0 {
		healthCheckResponse.BuildNumber = h.BuildNumber
	}

	return context.JSON(http.StatusOK, healthCheckResponse)
}

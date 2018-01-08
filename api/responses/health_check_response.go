package responses

type HealthCheckResponse struct {
	Name        string
	Version     string
	BuildNumber string
	Database    struct {
		CurrentVersion string
	}
}

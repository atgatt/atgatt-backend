package responses

type HealthCheckResponse struct {
	Name        string
	Version     string
	BuildNumber string
	CommitHash  string
	Database    struct {
		CurrentVersion string
	}
}

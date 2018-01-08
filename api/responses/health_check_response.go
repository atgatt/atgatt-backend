package responses

type HealthCheckResponse struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	BuildNumber string `json:"buildNumber"`
	CommitHash  string `json:"commitHash"`
	Database    struct {
		CurrentVersion string `json:"currentVersion"`
	} `json:"database"`
}

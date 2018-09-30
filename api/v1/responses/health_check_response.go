package responses

// HealthCheckResponse is a message containing the name of the application that is running, which version is running, which commit hash was used to make the build, etc.
type HealthCheckResponse struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	BuildNumber string `json:"buildNumber"`
	CommitHash  string `json:"commitHash"`
	Database    struct {
		CurrentVersion string `json:"currentVersion"`
	} `json:"database"`
}

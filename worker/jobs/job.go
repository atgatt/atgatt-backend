package jobs

// Job defines a generic background task that can either run successfully or return an error
type Job interface {
	Run() error
}

package responses

// Response represents the minimum fields required in any of our API responses
type Response struct {
	Message string `json:"message"`
}

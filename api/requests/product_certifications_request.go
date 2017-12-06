package requests

// MotorcycleHelmetCertificationsRequest represents filters that can be used to narrow down motorcycle helmet search results by which certifications that particular helmet has
type MotorcycleHelmetCertificationsRequest struct {
	SHARP bool `json:"SHARP"`
	SNELL bool `json:"SNELL"`
	ECE   bool `json:"ECE"`
	DOT   bool `json:"DOT"`
}

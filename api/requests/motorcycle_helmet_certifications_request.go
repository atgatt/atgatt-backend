package requests

type MotorcycleHelmetCertificationsRequest struct {
	SHARP bool `json:"SHARP"`
	SNELL bool `json:"SNELL"`
	ECE   bool `json:"ECE"`
	DOT   bool `json:"DOT"`
}

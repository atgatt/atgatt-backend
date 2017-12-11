package responses

type MotorcycleHelmetCertificationsResponse struct {
	SHARP *SHARPCertificationResponse
	SNELL bool
	ECE   bool
	DOT   bool
}

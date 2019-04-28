package queries

// HelmetCertificationsQueryParams represents parameters that can be used to filter Helmet-related certifications
type HelmetCertificationsQueryParams struct {
	SHARP *SHARPCertificationQueryParams `json:"SHARP"`
	SNELL bool                           `json:"SNELL"`
	ECE   bool                           `json:"ECE"`
	DOT   bool                           `json:"DOT"`
}
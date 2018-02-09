package entities

// SHARPCertificationDocument represents a motorcycle helmet certification given by the SHARP government agency
type SHARPCertificationDocument struct {
	Stars             int                             `json:"stars"`
	ImpactZoneRatings *SHARPImpactZoneRatingsDocument `json:"impactZoneRatings"`
}

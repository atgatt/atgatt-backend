package entities

// SHARPCertification represents a motorcycle helmet certification given by the SHARP government agency
type SHARPCertification struct {
	Stars             int                     `json:"stars"`
	ImpactZoneRatings *SHARPImpactZoneRatings `json:"impactZoneRatings"`
}

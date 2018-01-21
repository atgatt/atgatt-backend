package entities

type SHARPCertificationDocument struct {
	Stars             int                             `json:"stars"`
	ImpactZoneRatings *SHARPImpactZoneRatingsDocument `json:"impactZoneRatings"`
}

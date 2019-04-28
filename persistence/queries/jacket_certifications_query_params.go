package queries

// JacketCertificationsQueryParams represents parameters that can be used to filter Jacket-related certifications
type JacketCertificationsQueryParams struct {
	Shoulder   *CEImpactZoneQueryParams `json:"shoulder"`
	Elbow      *CEImpactZoneQueryParams `json:"elbow"`
	Back       *CEImpactZoneQueryParams `json:"back"`
	Chest      *CEImpactZoneQueryParams `json:"chest"`
	FitsAirbag bool `json:"fitsAirbag"`
}
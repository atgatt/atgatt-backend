package queries

// SHARPCertificationQueryParams is a sub-query used in conjunction with FilterProductsQuery in order to narrow down motorcycle helmet results
type SHARPCertificationQueryParams struct {
	Stars              int `json:"stars"`
	ImpactZoneMinimums struct {
		Left  int `json:"left"`
		Right int `json:"right"`
		Top   struct {
			Front int `json:"front"`
			Rear  int `json:"rear"`
		} `json:"top"`
		Rear int `json:"rear"`
	} `json:"impactZoneMinimums"`
}

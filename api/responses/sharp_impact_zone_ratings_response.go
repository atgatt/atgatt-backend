package responses

type SHARPImpactZoneRatingsResponse struct {
	Left  uint
	Right uint
	Top   *SHARPTopImpactZoneResponse
	Rear  uint
}

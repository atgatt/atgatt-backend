package entities

type SHARPImpactZoneRatingsDocument struct {
	Left  int `json:"left"`
	Right int `json:"right"`
	Top   struct {
		Front int `json:"front"`
		Rear  int `json:"rear"`
	} `json:"top"`
	Rear int `json:"rear"`
}

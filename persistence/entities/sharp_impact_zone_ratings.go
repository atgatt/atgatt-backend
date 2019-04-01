package entities

// SHARPImpactZoneRatings represents the impact scores assigned to each zone on the test subject's helmet, where 1 is the minimum score and 6 is the maximum score.
type SHARPImpactZoneRatings struct {
	Left  int `json:"left"`
	Right int `json:"right"`
	Top   struct {
		Front int `json:"front"`
		Rear  int `json:"rear"`
	} `json:"top"`
	Rear int `json:"rear"`
}

package requests

// ImpactZoneMinimumsRequest todo
type ImpactZoneMinimumsRequest struct {
	Left  uint                 `json:"left"`
	Right uint                 `json:"right"`
	Top   TopImpactZoneRequest `json:"top"`
	Rear  uint                 `json:"rear"`
}

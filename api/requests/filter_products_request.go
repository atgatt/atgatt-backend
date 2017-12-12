package requests

type FilterProductsRequest struct {
	Manufacturer   interface{} `json:"manufacturer"`
	Model          interface{} `json:"model"`
	Certifications struct {
		SHARP bool `json:"SHARP"`
		SNELL bool `json:"SNELL"`
		ECE   bool `json:"ECE"`
		DOT   bool `json:"DOT"`
	} `json:"certifications"`
	MinimumSHARPStars  int `json:"minimumSHARPStars"`
	ImpactZoneMinimums struct {
		Left  int `json:"left"`
		Right int `json:"right"`
		Top   struct {
			Front int `json:"front"`
			Rear  int `json:"rear"`
		} `json:"top"`
		Rear int `json:"rear"`
	} `json:"impactZoneMinimums"`
	UsdPriceRange [2]int `json:"usdPriceRange"`
}

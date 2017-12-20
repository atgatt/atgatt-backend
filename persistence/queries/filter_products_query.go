package queries

type FilterProductsQuery struct {
	Manufacturer   string `json:"manufacturer"`
	Model          string `json:"model"`
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
	Start         int    `json:"start"`
	Limit         int    `json:"limit"`
}

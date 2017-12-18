package interfaces

type FilterProductsQuery struct {
	Manufacturer   string
	Model          string
	Certifications struct {
		SHARP bool
		SNELL bool
		ECE   bool
		DOT   bool
	}
	MinimumSHARPStars  int
	ImpactZoneMinimums struct {
		Left  int
		Right int
		Top   struct {
			Front int
			Rear  int
		}
		Rear int
	}
	UsdPriceRange [2]int
	Start         int
	Limit         int
}

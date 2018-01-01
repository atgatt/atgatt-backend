package queries

type FilterProductsQuery struct {
	Manufacturer   string `json:"manufacturer"`
	Model          string `json:"model"`
	Certifications struct {
		SHARP *SHARPCertificationQueryParams `json:"SHARP"`
		SNELL bool                           `json:"SNELL"`
		ECE   bool                           `json:"ECE"`
		DOT   bool                           `json:"DOT"`
	} `json:"certifications"`
	UsdPriceRange []int `json:"usdPriceRange"`
	Start         int   `json:"start"`
	Limit         int   `json:"limit"`
	Order         struct {
		Field      string
		Descending bool
	}
}

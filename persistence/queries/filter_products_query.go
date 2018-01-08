package queries

type FilterProductsQuery struct {
	Manufacturer   string   `json:"manufacturer"`
	Model          string   `json:"model"`
	Subtypes       []string `json:"subtypes"`
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
		Field      string `json:"field"`
		Descending bool   `json:"descending"`
	} `json:"order"`
}

package queries

// FilterProductsQuery represents a query used to return a subset of products from the database. All of the query parameters are AND'd together when the query is executed.
type FilterProductsQuery struct {
	Type                 string                           `json:"type"`
	Subtypes             []string                         `json:"subtypes"`
	Manufacturer         string                           `json:"manufacturer"`
	Model                string                           `json:"model"`
	HelmetCertifications *HelmetCertificationsQueryParams `json:"helmetCertifications"`
	JacketCertifications *JacketCertificationsQueryParams `json:"jacketCertifications"`
	UsdPriceRange        []int                            `json:"usdPriceRange"`
	Start                int                              `json:"start"`
	Limit                int                              `json:"limit"`
	Order                struct {
		Field      string `json:"field"`
		Descending bool   `json:"descending"`
	} `json:"order"`
	ExcludeDiscontinued bool `json:"excludeDiscontinued"`
}

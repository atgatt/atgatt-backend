package requests

type FilterProductsRequest struct {
	Manufacturer       string                                 `json:"manufacturer"`
	Model              string                                 `json:"model"`
	Certifications     *MotorcycleHelmetCertificationsRequest `json:"certifications"`
	MinimumSHARPStars  uint                                   `json:"minimumSHARPStars"`
	ImpactZoneMinimums *ImpactZoneMinimumsRequest             `json:"impactZoneMinimums"`
	UsdPriceRange      [2]uint                                `json:"usdPriceRange"`
}

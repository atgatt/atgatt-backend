package responses

type FilterProductsResponse struct {
	UID            string
	Type           string
	Subtype        string
	Manufacturer   string
	Model          string
	ImageURL       string
	PriceInUsd     string
	Certifications *MotorcycleHelmetCertificationsResponse
}

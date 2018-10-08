package entities

// SHARPHelmet represents the data scraped for one motorcycle helmet from the SHARP website
type SHARPHelmet struct {
	Subtype              string
	Manufacturer         string
	Model                string
	ImageURL             string
	LatchPercentage      int
	WeightInLbsMultiple  int
	Sizes                []string
	Materials            string
	RetentionSystem      string
	Certifications       *SHARPCertificationDocument
	IsECECertified       bool
	ApproximateMSRPCents int
}

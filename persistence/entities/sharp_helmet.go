package entities

type SHARPHelmet struct {
	Subtype         string
	Manufacturer    string
	Model           string
	ImageURL        string
	PriceInUsd      string
	LatchPercentage int
	WeightInLbs     float64
	Sizes           []string
	Materials       string
	RetentionSystem string
	Certifications  *SHARPCertificationDocument
	IsECERated      bool
}

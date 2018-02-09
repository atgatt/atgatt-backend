package entities

// SNELLHelmet represents the data scraped for one motorcycle helmet from the SNELL website
type SNELLHelmet struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Size         string `json:"size"`
	Standard     string `json:"standard"`
	HelmetType   string `json:"helmettype"`
	FaceConfig   string `json:"faceconfig"`
}

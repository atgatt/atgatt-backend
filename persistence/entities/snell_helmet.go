package entities

type SNELLHelmet struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Size         string `json:"size"`
	Standard     string `json:"standard"`
	HelmetType   string `json:"helmettype"`
	FaceConfig   string `json:"faceconfig"`
}

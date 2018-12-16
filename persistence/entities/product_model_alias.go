package entities

// ProductModelAlias represents an alias for a given Manufacturer + Model pair. i.e. Shoei XR-1100 is known as the Shoei RF-1100 in the USA (alias)
type ProductModelAlias struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	ModelAlias   string `json:"modelAlias"`
	IsForDisplay bool   `json:"isForDisplay"`
}

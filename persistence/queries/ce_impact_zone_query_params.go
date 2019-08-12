package queries

// CEImpactZoneQueryParams represents a query for a zone of a particular product (or an entire product) that has a level 1 or 2 CE certification
type CEImpactZoneQueryParams struct {
	IsLevel2   bool `json:"isLevel2"`
	IsApproved bool `json:"isApproved"`
	IsEmpty    bool `json:"isEmpty"` // if this product can fit CE-certified/approved armor in this zone but doesn't include one, it's an "empty" zone
}

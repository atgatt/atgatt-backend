package entities

// CEImpactZone represents a zone of a particular product (or an entire product) that has a level 1 or 2 CE certification
type CEImpactZone struct {
	IsLevel2   bool `json:"isLevel2"`
	IsApproved bool `json:"isApproved"`
	IsEmpty    bool `json:"isEmpty"` // if this product can fit CE-certified/approved armor in this zone but doesn't include one, it's an "empty" zone
}

// GetScore returns the component of an overall safety score associated with this zone
func (c *CEImpactZone) GetScore() float64 {
	// Penalize the product, but give some credit, for not including armor but allowing the user to install their own
	if c.IsEmpty {
		return 0.25
	}

	// start off with 25% for CE level 1 and no approval
	totalScore := 0.25

	// use 50% if we have CE level 2
	if c.IsLevel2 {
		totalScore = 0.50
	}

	// ... add 50% for CE approval
	if c.IsApproved {
		totalScore += 0.50
	}

	// ... 100% total if we have CE level 2 on top of approval, otherwise end up at 75% total
	

	return totalScore
}

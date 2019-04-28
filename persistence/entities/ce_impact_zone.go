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

	var totalScore float64
	if c.IsApproved {
		totalScore += 0.50
	}

	if c.IsLevel2 {
		totalScore += 0.50
	} else {
		totalScore += 0.25
	}

	return totalScore
}

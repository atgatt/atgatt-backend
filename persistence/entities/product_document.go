package entities

type ProductDocument struct {
	UID             string
	Type            string
	Subtype         string
	AmazonProductID string
	Manufacturer    string
	Model           string
	ImageURL        string
	PriceInUsd      string
	Certifications  struct {
		SHARP struct {
			Stars             int
			ImpactZoneRatings struct {
				Left  int
				Right int
				Top   struct {
					Front int
					Rear  int
				}
				Rear int
			}
		}
		SNELL struct {
		}
		ECE struct {
		}
		DOT struct {
		}
	}
	Score string
}

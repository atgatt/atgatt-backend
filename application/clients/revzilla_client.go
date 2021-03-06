package clients

import "github.com/PuerkitoBio/goquery"

// RevzillaClient represents a client that can communicate with Revzilla.com to get various product information
type RevzillaClient interface {
	GetAllProductOverviewsHTML(productURLPrefix string) (*goquery.Document, error)
	GetDescriptionPartsHTMLByURL(url string) (*goquery.Document, error)
}

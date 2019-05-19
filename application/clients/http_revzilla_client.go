package clients

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"

	"github.com/hashicorp/go-cleanhttp"
)

// HTTPRevzillaClient is a RevzillaClient that communicates with Revzilla.com over HTTP
type HTTPRevzillaClient struct {
	pooledClient *http.Client
}

// NewHTTPRevzillaClient initializes a HTTPRevzillaClient with a default, pooled HTTPClient.
func NewHTTPRevzillaClient() *HTTPRevzillaClient {
	return &HTTPRevzillaClient{pooledClient: cleanhttp.DefaultPooledClient()}
}

// GetAllJacketOverviewsHTML returns a GoQuery document representing each Revzilla Jacket - GetDescriptionPartsByProduct() can be used to further drill into the details for each of these results
func (c *HTTPRevzillaClient) GetAllJacketOverviewsHTML() (*goquery.Document, error) {
	resp, err := c.pooledClient.Get("https://www.revzilla.com/motorcycle-jackets-vests?page=1&sort=featured&limit=10000&rating=-1&price=&price_min=3&price_max=1700&is_new=false&is_sale=false&is_made_in_usa=false&has_video=false&is_holiday=false&is_blemished=false&view_all=true")
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

// GetDescriptionPartsHTMLByURL returns a GoQuery document representing each bullet point in a revzilla product description
func (c *HTTPRevzillaClient) GetDescriptionPartsHTMLByURL(url string) (*goquery.Document, error) {
	resp, err := c.pooledClient.Get(url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
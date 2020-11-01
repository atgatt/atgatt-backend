package entities

// CJProductsWrapper represents a product returned from the Commission Junction API
type CJProductsWrapper struct {
	ResultList    []CJProduct `json:"resultList"`
}

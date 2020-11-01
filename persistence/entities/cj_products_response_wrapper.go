package entities

// CJProductsResponseWrapper represents a wrapper around the list of all the products returned from the Commission Junction API
type CJProductsResponseWrapper struct {
	Data struct {
		ShoppingProducts CJProductsWrapper `json:"shoppingProducts"`
	} `json:"data"`
}

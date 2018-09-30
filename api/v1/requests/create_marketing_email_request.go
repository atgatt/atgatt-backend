package requests

// CreateMarketingEmailRequest represents a request to add an email address to the mailing list used for marketing purposes
type CreateMarketingEmailRequest struct {
	Email string `json:"email"`
}

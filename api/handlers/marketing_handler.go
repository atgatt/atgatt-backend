package handlers

import (
	"crashtested-backend/api/requests"
	"crashtested-backend/api/responses"
	"crashtested-backend/persistence/repositories"
	"net/http"
	"strings"

	"github.com/goware/emailx"
	"github.com/labstack/echo"
)

// MarketingHandler contains functions related to filtering and updating Products
type MarketingHandler struct {
	Repository         *repositories.MarketingRepository
	AllowedOrderFields map[string]bool
}

// CreateMarketingEmail inserts a marketing email address into the database if it has a valid format and hostname, otherwise returns http 400 (bad request)
func (m *MarketingHandler) CreateMarketingEmail(context echo.Context) (err error) {
	query := new(requests.CreateMarketingEmailRequest)
	if err := context.Bind(query); err != nil {
		return err
	}

	lowerEmail := strings.ToLower(query.Email)

	if err := emailx.Validate(lowerEmail); err != nil {
		return context.JSON(http.StatusBadRequest, &responses.Response{Message: "The email that you supplied is invalid. Try again with a valid email address."})
	}

	marketingEmailExists, err := m.Repository.MarketingEmailExists(lowerEmail)
	if err != nil {
		return err
	}

	if marketingEmailExists {
		return context.JSON(http.StatusBadRequest, &responses.Response{Message: "You're already signed up."})
	}

	m.Repository.CreateMarketingEmail(lowerEmail)

	return context.NoContent(http.StatusOK)
}

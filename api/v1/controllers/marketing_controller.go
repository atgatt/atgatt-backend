package controllers

import (
	"crashtested-backend/api/v1/requests"
	"crashtested-backend/api/v1/responses"
	"crashtested-backend/persistence/repositories"
	"net/http"
	"strings"

	"github.com/goware/emailx"
	"github.com/labstack/echo/v4"
)

// MarketingController contains functions related to filtering and updating Products
type MarketingController struct {
	Repository         *repositories.MarketingRepository
	AllowedOrderFields map[string]bool
}

// CreateMarketingEmail inserts a marketing email address into the database if it has a valid format and hostname, otherwise returns http 400 (bad request)
func (m *MarketingController) CreateMarketingEmail(context echo.Context) (err error) {
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

	err = m.Repository.CreateMarketingEmail(lowerEmail)
	if err != nil {
		return err
	}

	return context.NoContent(http.StatusOK)
}

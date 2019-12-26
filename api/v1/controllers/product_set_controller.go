package controllers

import (
	"crashtested-backend/api/v1/requests"
	"crashtested-backend/api/v1/responses"
	"crashtested-backend/application/services"
	"crashtested-backend/persistence/repositories"
	"net/http"

	"github.com/google/uuid"

	"github.com/labstack/echo/v4"
)

// ProductSetController contains functions related to filtering and updating ProductSets
type ProductSetController struct {
	Service    *services.ProductSetService
	Repository *repositories.ProductSetRepository
}

// CreateProductSet creates a product set based off an existing product set, or creates a new one in the database
func (p *ProductSetController) CreateProductSet(context echo.Context) (err error) {
	request := new(requests.CreateProductSetRequest)
	if err := context.Bind(request); err != nil {
		return err
	}

	uuidCreated, err := p.Service.UpsertProductSet(request.SourceProductSetID, request.ProductID)
	if err != nil {
		return err
	}

	return context.JSON(http.StatusOK, &responses.CreateProductSetResponse{ID: uuidCreated})
}

// GetProductSetDetails returns a product set with the given UUID, otherwise a 404 is returned
func (p *ProductSetController) GetProductSetDetails(context echo.Context) (err error) {
	uuidString := context.Param("uuid")
	productSetID, err := uuid.Parse(uuidString)
	if err != nil {
		return err
	}

	productSet, err := p.Repository.GetProductSetWithProductsByUUID(productSetID)
	if err != nil {
		return err
	}

	return context.JSON(http.StatusOK, &responses.GetProductSetDetailsResponse{
		ID:            productSet.UUID,
		HelmetProduct: productSet.HelmetProduct,
		JacketProduct: productSet.JacketProduct,
		PantsProduct:  productSet.PantsProduct,
		BootsProduct:  productSet.BootsProduct,
		GlovesProduct: productSet.GlovesProduct,
	})
}

package handlers

import (
	"crashtested-backend/api/requests"
	"net/http"

	"github.com/labstack/echo"
)

type ProductsHandler struct {
}

func (self *ProductsHandler) FilterProducts(context echo.Context) (err error) {
	request := &requests.FilterProductsRequest{}
	context.Bind(request)
	if err := context.Bind(request); err != nil {
		return err
	}

	mockResponseJson := `[{"uid":"abc","amazonProductId":"A12341QC","manufacturer":"Shoei","model":"RF-SR","imageUrl":"https://www.shoei-helmets.com/pub/media/catalog/product/cache/1/image/700x560/e9c3970ab036de70892d86c6d221abfe/x/-/x-fourteen-white_2_2.png","priceInUsd":"399.99","certifications":{"SNELL":{},"ECE":{},"DOT":{}},"score":"76"},{"uid":"def","type":"helmet","subtype":"fullface","amazonProductId":"A45561QB","manufacturer":"Shoei","model":"RF-1200","imageUrl":"https://www.shoei-helmets.com/pub/media/catalog/product/cache/1/image/700x560/e9c3970ab036de70892d86c6d221abfe/x/-/x-fourteen-white_2_2.png","priceInUsd":"499.99","certifications":{"SHARP":{"ratingType":"stars","ratingValue":"4","impactZoneRatings":{"left":5,"right":4,"top":{"front":0,"rear":5},"rear":3}},"SNELL":{},"ECE":{},"DOT":{}},"score":"70"},{"uid":"ghi","type":"helmet","subtype":"fullface","amazonProductId":"A45561QB","manufacturer":"Shoei","model":"RF-1100","imageUrl":"https://www.shoei-helmets.com/pub/media/catalog/product/cache/1/image/700x560/e9c3970ab036de70892d86c6d221abfe/x/-/x-fourteen-white_2_2.png","priceInUsd":"259.99","certifications":{"SHARP":{"ratingType":"stars","ratingValue":"3","impactZoneRatings":{"left":5,"right":4,"top":{"front":3,"rear":2},"rear":1}},"SNELL":{},"ECE":{},"DOT":{}},"score":"65"},{"uid":"jkl","type":"helmet","subtype":"fullface","amazonProductId":"A45561QB","manufacturer":"Shoei","model":"Qwest","imageUrl":"https://sharp.dft.gov.uk/wp-content/uploads/2017/03/shoei-xr-1100-150x150.jpg","priceInUsd":"299.99","certifications":{"SHARP":{"ratingType":"stars","ratingValue":"5","impactZoneRatings":{"left":5,"right":4,"top":{"front":0,"rear":5},"rear":3}},"SNELL":{},"ECE":{},"DOT":{}},"score":"80"}]`
	return context.JSON(http.StatusOK, mockResponseJson)
}

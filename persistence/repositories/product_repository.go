package repositories

import "crashtested-backend/persistence/entities"
import "crashtested-backend/persistence/queries"
import "github.com/jmoiron/sqlx"
import "encoding/json"

type ProductRepository struct {
	ConnectionString string
}

func (self *ProductRepository) FilterProducts(query *queries.FilterProductsQuery) ([]*entities.ProductDocument, error) {
	db, err := sqlx.Open("postgres", self.ConnectionString)
	defer db.Close()

	if err != nil {
		return nil, err
	}

	var productDocuments []*entities.ProductDocument = make([]*entities.ProductDocument, 0)
	var productJsonStrings []string
	err = db.Select(&productJsonStrings, "select document from products where document->>'type' = $1", "helmet")
	if err != nil {
		return nil, err
	}

	for _, productJsonString := range productJsonStrings {
		productDocument := &entities.ProductDocument{}
		err := json.Unmarshal([]byte(productJsonString), productDocument)
		if err != nil {
			return nil, err
		}
		productDocuments = append(productDocuments, productDocument)
	}

	return productDocuments, nil
}

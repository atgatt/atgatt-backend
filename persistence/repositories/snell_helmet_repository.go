package repositories

import (
	"crashtested-backend/persistence/entities"
)

type SNELLHelmetRepository struct {
	// http://snell.us.com/codefolder/datatable.php
}

func (self *SNELLHelmetRepository) GetAllHelmets() ([]*entities.SNELLHelmet, error) {
	return nil, nil
}

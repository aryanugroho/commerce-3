package stock

import (
	"github.com/ottemo/commerce/app/models"
	"github.com/ottemo/commerce/app/models/stock"
)

// GetCollection returns collection of current instance type
func (it *DefaultStock) GetCollection() models.InterfaceCollection {
	model, err := models.GetModel(stock.ConstModelNameStockCollection)
	if err != nil {
		return nil
	}
	if result, ok := model.(stock.InterfaceStockCollection); ok {
		return result
	}

	return nil
}

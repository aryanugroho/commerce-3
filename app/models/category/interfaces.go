// Package category represents abstraction of business layer category object
package category

import (
	"github.com/ottemo/commerce/app/models"
	"github.com/ottemo/commerce/app/models/product"
	"github.com/ottemo/commerce/env"
)

// Package global constants
const (
	ConstModelNameCategory           = "Category"
	ConstModelNameCategoryCollection = "CategoryCollection"

	ConstErrorModule = "category"
	ConstErrorLevel  = env.ConstErrorLevelModel
)

// InterfaceCategory represents interface to access business layer implementation of category object
type InterfaceCategory interface {
	GetEnabled() bool

	GetName() string

	GetParent() InterfaceCategory

	GetDescription() string

	GetImage() string

	GetProductIds() []string
	GetProductsCollection() product.InterfaceProductCollection
	GetProducts() []product.InterfaceProduct

	AddProduct(productID string) error
	RemoveProduct(productID string) error

	models.InterfaceModel
	models.InterfaceObject
	models.InterfaceStorable
	models.InterfaceMedia
	models.InterfaceListable
}

// InterfaceCategoryCollection represents interface to access business layer implementation of category collection
type InterfaceCategoryCollection interface {
	ListCategories() []InterfaceCategory

	models.InterfaceCollection
}

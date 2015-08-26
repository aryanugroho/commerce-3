package category

import (
	"io/ioutil"
	"mime"
	"strings"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/category"
	"github.com/ottemo/foundation/app/models/product"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/media"
	"github.com/ottemo/foundation/utils"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	var err error

	err = api.GetRestService().RegisterAPI("categories", api.ConstRESTOperationGet, APIListCategories)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("categories/tree", api.ConstRESTOperationGet, APIGetCategoriesTree)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("categories/attributes", api.ConstRESTOperationGet, APIGetCategoryAttributes)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("category", api.ConstRESTOperationCreate, APICreateCategory)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("category/:categoryID", api.ConstRESTOperationUpdate, APIUpdateCategory)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("category/:categoryID", api.ConstRESTOperationDelete, APIDeleteCategory)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("category/:categoryID", api.ConstRESTOperationGet, APIGetCategory)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("category/:categoryID/layers", api.ConstRESTOperationGet, APIGetCategoryLayers)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("category/:categoryID/products", api.ConstRESTOperationGet, APIGetCategoryProducts)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("category/:categoryID/product/:productID", api.ConstRESTOperationCreate, APIAddProductToCategory)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("category/:categoryID/product/:productID", api.ConstRESTOperationDelete, APIRemoveProductFromCategory)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("category/:categoryID/media/:mediaType/:mediaName", api.ConstRESTOperationGet, APIGetMedia)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("category/:categoryID/media/:mediaType", api.ConstRESTOperationGet, APIListMedia)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("category/:categoryID/media/:mediaType/:mediaName", api.ConstRESTOperationCreate, APIAddMediaForCategory)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("category/:categoryID/media/:mediaType/:mediaName", api.ConstRESTOperationDelete, APIRemoveMediaForCategory)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("category/:categoryID/mediapath/:mediaType", api.ConstRESTOperationGet, APIGetMediaPath)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	return nil
}

// APIListCategories returns a list of available categories
//   - if "action" parameter is set to "count" result value will be just a number of list items
//   - for a not admins available categories are limited to enabled ones
func APIListCategories(context api.InterfaceApplicationContext) (interface{}, error) {

	categoryCollectionModel, err := category.GetCategoryCollectionModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// applying requested filters
	models.ApplyFilters(context, categoryCollectionModel.GetDBCollection())

	// excluding disabled categories for a regular visitor
	if err := api.ValidateAdminRights(context); err != nil {
		categoryCollectionModel.GetDBCollection().AddFilter("enabled", "=", true)
	}

	// checking for a "count" request
	if context.GetRequestArgument(api.ConstRESTActionParameter) == "count" {
		return categoryCollectionModel.GetDBCollection().Count()
	}

	// limit parameter handle
	categoryCollectionModel.ListLimit(models.GetListLimit(context))

	// extra parameter handle
	models.ApplyExtraAttributes(context, categoryCollectionModel)

	listItems, err := categoryCollectionModel.List()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	mediaStorage, err := media.GetMediaStorage()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	var result []map[string]interface{}

	for _, listItem := range listItems {

		itemImages, err := mediaStorage.GetAllSizes(category.ConstModelNameCategory, listItem.ID, ConstCategoryMediaTypeImage)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		// move default image to first position in array
		if listItem.Image != "" && len(itemImages) > 1 {
			found := false
			for index, images := range itemImages {
				for sizeName, sizeValue := range images {
					basicName := strings.Replace(sizeValue, "_"+sizeName, "", -1)
					if strings.Contains(basicName, listItem.Image) {
						found = true
						itemImages = append(itemImages[:index], itemImages[index+1:]...)
						itemImages = append([]map[string]string{images}, itemImages...)
					}
					break
				}
				if found {
					break
				}
			}
		}

		item := map[string]interface{}{
			"ID":     listItem.ID,
			"Name":   listItem.Name,
			"Desc":   listItem.Desc,
			"Extra":  listItem.Extra,
			"Image":  listItem.Image,
			"Images": itemImages,
		}

		result = append(result, item)
	}

	return result, nil
}

// APICreateCategory creates a new category
//   - category attributes must be provided in request content
//   - "name" attribute required
func APICreateCategory(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if _, present := requestData["name"]; !present {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "521b50d6-0d98-491a-8e3a-37678fbbccfe", "category name was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// create category operation
	//-------------------------
	categoryModel, err := category.GetCategoryModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for attribute, value := range requestData {
		err := categoryModel.Set(attribute, value)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	err = categoryModel.Save()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return categoryModel.ToHashMap(), nil
}

// APIDeleteCategory deletes existing category
//   - category id should be specified in "categoryID" argument
func APIDeleteCategory(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//--------------------
	categoryID := context.GetRequestArgument("categoryID")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "f1507b00-337e-4903-8244-5cf959dde3a4", "category id was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// delete operation
	//-----------------
	categoryModel, err := category.GetCategoryModelAndSetID(categoryID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = categoryModel.Delete()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APIUpdateCategory modifies existing category
//   - category id must be specified as "id" argument
//   - category attributes must be specified in content
func APIUpdateCategory(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	categoryID := context.GetRequestArgument("categoryID")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "389975e7-611c-4d6c-8b4d-bca450f5f7e7", "category id was not specified")
	}

	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// update operations
	//------------------
	categoryModel, err := category.LoadCategoryByID(categoryID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for attrName, attrVal := range requestData {
		err = categoryModel.Set(attrName, attrVal)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	err = categoryModel.Save()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return categoryModel.ToHashMap(), nil
}

// APIGetCategoryAttributes returns a list of category attributes
func APIGetCategoryAttributes(context api.InterfaceApplicationContext) (interface{}, error) {
	categoryModel, err := category.GetCategoryModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	attrInfo := categoryModel.GetAttributesInfo()

	return attrInfo, nil
}

// APIGetCategoryLayers enumerates category attributes and their possible values which is used for layered navigation
//   - category id should be specified in "id" argument
func APIGetCategoryLayers(context api.InterfaceApplicationContext) (interface{}, error) {
	categoryID := context.GetRequestArgument("categoryID")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "389975e7-611c-4d6c-8b4d-bca450f5f7e7", "category id was not specified")
	}

	categoryModel, err := category.LoadCategoryByID(categoryID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if api.ValidateAdminRights(context) != nil && !categoryModel.GetEnabled() {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "d46dadf8-373a-4247-a81e-fbbe39a7fe74", "category is not available")
	}

	productsCollection := categoryModel.GetProductsCollection()
	productsDBCollection := productsCollection.GetDBCollection()

	productModel, err := product.GetProductModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	productAttributesInfo := productModel.GetAttributesInfo()

	result := make(map[string]interface{})

	models.ApplyFilters(context, productsDBCollection)

	// not allowing to see disabled products if not admin
	if err := api.ValidateAdminRights(context); err != nil {
		productsDBCollection.AddFilter("enabled", "=", true)
	}

	for _, productAttribute := range productAttributesInfo {
		if productAttribute.IsLayered {
			distinctValues, _ := productsDBCollection.Distinct(productAttribute.Attribute)
			result[productAttribute.Attribute] = distinctValues
		}
	}

	return result, nil
}

// APIGetCategoryProducts returns category related products
//   - category id should be specified in "categoryID" argument
func APIGetCategoryProducts(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	categoryID := context.GetRequestArgument("categoryID")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "12400cff-34fe-4cf5-ac6e-41625f9e3d5a", "category id was not specified")
	}

	// product list operation
	//-----------------------
	categoryModel, err := category.LoadCategoryByID(categoryID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if api.ValidateAdminRights(context) != nil && !categoryModel.GetEnabled() {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "9a6f080d-dfa4-4f8c-8a0c-ec31cbe1cd87", "category is not available")
	}

	productsCollection := categoryModel.GetProductsCollection()

	models.ApplyFilters(context, productsCollection.GetDBCollection())

	// not allowing to see disabled products if not admin
	if err := api.ValidateAdminRights(context); err != nil {
		productsCollection.GetDBCollection().AddFilter("enabled", "=", true)
	}

	// checking for a "count" request
	if context.GetRequestArgument(api.ConstRESTActionParameter) == "count" {
		return productsCollection.GetDBCollection().Count()
	}

	// limit parameter handle
	productsCollection.ListLimit(models.GetListLimit(context))

	mediaStorage, err := media.GetMediaStorage()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// preparing product information
	var result []map[string]interface{}

	for _, productModel := range productsCollection.ListProducts() {
		productInfo := productModel.ToHashMap()

		itemImages, err := mediaStorage.GetAllSizes(product.ConstModelNameProduct, productModel.GetID(), ConstCategoryMediaTypeImage)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		if defaultImage, present := productInfo["default_image"]; present {
			mediaPath, err := productModel.GetMediaPath("image")
			if defaultImage, ok := defaultImage.(string); ok && defaultImage != "" && err == nil {
				productInfo["default_image"] = mediaPath + defaultImage

				// move default image to first position in array
				if len(itemImages) > 1 {
					found := false
					for index, images := range itemImages {
						for sizeName, sizeValue := range images {
							basicName := strings.Replace(sizeValue, "_"+sizeName, "", -1)
							if strings.Contains(basicName, defaultImage) {
								found = true
								itemImages = append(itemImages[:index], itemImages[index+1:]...)
								itemImages = append([]map[string]string{images}, itemImages...)
							}
							break
						}
						if found {
							break
						}
					}
				}
			}
		}

		productInfo["images"] = itemImages
		result = append(result, productInfo)
	}

	return result, nil
}

// APIAddProductToCategory adds product to category
//   - category id and product id  should be specified in "categoryID" and "productID" arguments
func APIAddProductToCategory(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	categoryID := context.GetRequestArgument("categoryID")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "2fc40bdc-3c8e-4a9c-910b-cca62cda1b77", "category id was not specified")
	}
	productID := context.GetRequestArgument("productID")
	if productID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "190a3d60-7769-4908-b383-80bc143128da", "product id was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// category product add operation
	//-------------------------------
	categoryModel, err := category.GetCategoryModelAndSetID(categoryID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = categoryModel.AddProduct(productID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APIRemoveProductFromCategory removes product from category
//   - category id and product id  should be specified in "categoryID" and "productID" arguments
func APIRemoveProductFromCategory(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	categoryID := context.GetRequestArgument("categoryID")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "dc083799-bccb-48c8-bd56-5dcd0a0f6031", "category id was not specified")
	}
	productID := context.GetRequestArgument("productID")
	if productID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "9ffd2626-4192-4726-849e-c7a2416fab3a", "product id was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// category product add operation
	//-------------------------------
	categoryModel, err := category.GetCategoryModelAndSetID(categoryID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = categoryModel.RemoveProduct(productID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APIGetCategory return specified category information
//   - category id should be specified in "categoryID" argument
func APIGetCategory(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	categoryID := context.GetRequestArgument("categoryID")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "3c336fd7-1a18-4aea-9eb0-460d746f8dfa", "category id was not specified")
	}

	// load product operation
	//-----------------------
	categoryModel, err := category.LoadCategoryByID(categoryID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if api.ValidateAdminRights(context) != nil && !categoryModel.GetEnabled() {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "80615e04-f43d-42a4-9482-39a5e7f8ccb7", "category is not available")
	}

	mediaStorage, err := media.GetMediaStorage()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	result := categoryModel.ToHashMap()

	// preparing product information
	var productsResult []map[string]interface{}
	products := utils.InterfaceToArray(result["products"])

	for _, productMap := range products {
		productInfo := utils.InterfaceToMap(productMap)
		productID, present := productInfo["_id"]
		if present && utils.InterfaceToString(productID) != "" {

			itemImages, err := mediaStorage.GetAllSizes(product.ConstModelNameProduct, utils.InterfaceToString(productID), ConstCategoryMediaTypeImage)
			if err != nil {
				return nil, env.ErrorDispatch(err)
			}

			// move default image to first position in array
			defaultImage := utils.InterfaceToString(productInfo["default_image"])
			if _, present := productInfo["default_image"]; present && defaultImage != "" && len(itemImages) > 1 {
				found := false
				for index, images := range itemImages {
					for sizeName, sizeValue := range images {
						basicName := strings.Replace(sizeValue, "_"+sizeName, "", -1)
						if strings.Contains(basicName, defaultImage) {
							found = true
							itemImages = append(itemImages[:index], itemImages[index+1:]...)
							itemImages = append([]map[string]string{images}, itemImages...)
						}
						break
					}
					if found {
						break
					}
				}
			}

			productInfo["images"] = itemImages
			productsResult = append(productsResult, productInfo)
		}
	}

	result["products"] = productsResult

	itemImages, err := mediaStorage.GetAllSizes(category.ConstModelNameCategory, categoryModel.GetID(), ConstCategoryMediaTypeImage)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// move default image to first position in array
	defaultImage := utils.InterfaceToString(result["image"])
	if _, present := result["image"]; present && defaultImage != "" && len(itemImages) > 1 {
		found := false
		for index, images := range itemImages {
			for sizeName, value := range images {
				basicName := strings.Replace(value, "_"+sizeName, "", -1)
				if strings.Contains(basicName, defaultImage) {
					found = true
					itemImages = append(itemImages[:index], itemImages[index+1:]...)
					itemImages = append([]map[string]string{images}, itemImages...)
				}
				break
			}
			if found {
				break
			}
		}
	}

	result["images"] = itemImages

	return result, nil
}

// APIGetCategoriesTree returns categories parent/child relation map
func APIGetCategoriesTree(context api.InterfaceApplicationContext) (interface{}, error) {

	var result = make([]map[string]interface{}, 0)

	collection, err := db.GetCollection(ConstCollectionNameCategory)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = collection.AddFilter("enabled", "=", true)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = collection.AddSort("path", false)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	rowData, err := collection.Load()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	var categoryStack []map[string]interface{}
	var pathStack []string

	for _, row := range rowData {

		currentItem := make(map[string]interface{})
		currentItem["id"] = row["_id"]
		currentItem["name"] = row["name"]
		currentItem["child"] = make([]map[string]interface{}, 0)

		// calculating current path
		currentPath, ok := row["path"].(string)
		if !ok {
			continue
		}

		for idx := len(pathStack) - 1; idx >= 0; idx-- {
			parentPath := pathStack[idx]

			// if we found parent
			if strings.Contains(currentPath, parentPath) {
				parent := categoryStack[idx]

				parentChild, ok := parent["child"].([]map[string]interface{})
				if !ok {
					return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "c94e54a4-53dc-4ed2-bac4-7e9d93958765", "category tree builder internal error")
				}

				parent["child"] = append(parentChild, currentItem)

				break
			} else {
				pathStack = pathStack[0:idx]
				categoryStack = categoryStack[0:idx]
			}
		}

		if len(categoryStack) == 0 {
			result = append(result, currentItem)
		}

		categoryStack = append(categoryStack, currentItem)
		pathStack = append(pathStack, currentPath)

	}

	return result, nil
}

// APIGetMediaPath returns relative path to category media files within media library
//   - category id, media type must be specified in "categoryID" and "mediaType" arguments
func APIGetMediaPath(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	categoryID := context.GetRequestArgument("categoryID")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "6597ff92-f2ee-4233-bcf9-eb73b957fb05", "category id was not specified")
	}

	mediaType := context.GetRequestArgument("mediaType")
	if mediaType == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "75c00741-5873-4be1-9fa0-df9d2956d3de", "media type was not specified")
	}

	// list media operation
	//---------------------
	categoryModel, err := category.GetCategoryModelAndSetID(categoryID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	mediaList, err := categoryModel.GetMediaPath(mediaType)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return mediaList, nil
}

// APIListMedia returns lost of media files assigned to specified category
//   - category id, media type must be specified in "categoryID" and "mediaType" arguments
func APIListMedia(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	categoryID := context.GetRequestArgument("categoryID")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "52677450-8a7f-49c9-a472-51d0e80bc7ca", "category id was not specified")
	}

	mediaType := context.GetRequestArgument("mediaType")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "b8b31a9f-6fac-47b3-89e2-c9b3e589a8f6", "media type was not specified")
	}

	// list media operation
	//---------------------
	categoryModel, err := category.GetCategoryModelAndSetID(categoryID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	mediaList, err := categoryModel.ListMedia(mediaType)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return mediaList, nil
}

// APIAddMediaForCategory uploads and assigns media file send in request for a specified category
//   - category id, media type and media name should be specified in "categoryID", "mediaType" and "mediaName" arguments
//   - media file should be provided in "file" field
func APIAddMediaForCategory(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	categoryID := context.GetRequestArgument("categoryID")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "a4696c5d-3276-4272-8d86-8061e57743a5", "category id was not specified")
	}

	mediaType := context.GetRequestArgument("mediaType")
	if mediaType == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "f3ea9a01-412a-4af2-9496-cb58cdb8139d", "media type was not specified")
	}

	mediaName := context.GetRequestArgument("mediaName")
	if mediaName == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "23fb7617-f19a-4505-b706-10f7898fd980", "media name was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// income file processing
	//-----------------------
	files := context.GetRequestFiles()
	if len(files) == 0 {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "75a2ddaf-b63d-4eed-b16d-4b32778f5fc1", "media file was not specified")
	}

	var fileContents []byte
	for _, fileReader := range files {
		contents, err := ioutil.ReadAll(fileReader)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
		fileContents = contents
		break
	}

	// add media operation
	//--------------------
	categoryModel, err := category.GetCategoryModelAndSetID(categoryID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// Adding timestamp to image name to prevent overwriting
	mediaNameParts := strings.SplitN(mediaName, ".", 2)
	mediaName = mediaNameParts[0] + "_" + utils.InterfaceToString(time.Now().Unix()) + "." + mediaNameParts[1]

	err = categoryModel.AddMedia(mediaType, mediaName, fileContents)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APIRemoveMediaForCategory removes media content from specified category
//   - category id, media type and media name should be specified in "categoryID", "mediaType" and "mediaName" arguments
func APIRemoveMediaForCategory(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	categoryID := context.GetRequestArgument("categoryID")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "f5f77b7f-6606-4bdd-a113-0a3b26f5759c", "category id was not specified")
	}

	mediaType := context.GetRequestArgument("mediaType")
	if mediaType == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "e81b841f-8253-4b66-ac7d-2cc9a484044c", "media type was not specified")
	}

	mediaName := context.GetRequestArgument("mediaName")
	if mediaName == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "63b37b08-3b21-48b7-9058-291bb7e635a1", "media name was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// list media operation
	//---------------------
	categoryModel, err := category.GetCategoryModelAndSetID(categoryID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = categoryModel.RemoveMedia(mediaType, mediaName)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APIGetMedia returns media contents for a category (file assigned to a category)
//   - category id, media type and media name must be specified in "categoryID", "mediaType" and "mediaName" arguments
//   - on success case not a JSON data returns, but media file
func APIGetMedia(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	categoryID := context.GetRequestArgument("categoryID")
	if categoryID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "d33b8a67-359f-4a3e-b626-f58b6c70f09f", "category id was not specified")
	}

	mediaType := context.GetRequestArgument("mediaType")
	if mediaType == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "d081b726-caf4-4694-baaa-7b1801ca9713", "media type was not specified")
	}

	mediaName := context.GetRequestArgument("mediaName")
	if mediaName == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "124c8b9d-1a6b-491c-97ba-a03e8c828337", "media name was not specified")
	}

	context.SetResponseContentType(mime.TypeByExtension(mediaName))

	// list media operation
	//---------------------
	categoryModel, err := category.GetCategoryModelAndSetID(categoryID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return categoryModel.GetMedia(mediaType, mediaName)
}

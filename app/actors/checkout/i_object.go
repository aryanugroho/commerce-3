package checkout

import (
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// Get returns object attribute value or nil
func (it *DefaultCheckout) Get(attribute string) interface{} {

	switch attribute {
	case "CartID":
		return it.CartID
	case "VisitorID":
		return it.VisitorID
	case "OrderID":
		return it.OrderID
	case "SessionID":
		return it.SessionID
	case "ShippingAddress":
		return it.ShippingAddress
	case "BillingAddress":
		return it.BillingAddress
	case "PaymentMethodCode":
		return it.PaymentMethodCode
	case "ShippingMethodCode":
		return it.ShippingMethodCode
	case "ShippingRate":
		return it.ShippingRate
	case "Taxes":
		return it.Taxes
	case "Discounts":
		return it.Discounts
	case "Info":
		return it.Info
	}

	return nil
}

// Set sets attribute value to object or returns error
func (it *DefaultCheckout) Set(attribute string, value interface{}) error {
	switch attribute {
	case "CartID":
		it.CartID = utils.InterfaceToString(value)

	case "VisitorID":
		it.VisitorID = utils.InterfaceToString(value)

	case "OrderID":
		it.OrderID = utils.InterfaceToString(value)

	case "SessionID":
		it.SessionID = utils.InterfaceToString(value)

	case "ShippingAddress":
		shippingAddress := utils.InterfaceToMap(value)
		if len(shippingAddress) > 0 {
			it.ShippingAddress = shippingAddress
		}

	case "BillingAddress":
		billingAddress := utils.InterfaceToMap(value)
		if len(billingAddress) > 0 {
			it.BillingAddress = billingAddress
		}

	case "PaymentMethodCode":
		paymentMethodCode := utils.InterfaceToString(value)
		for _, method := range checkout.GetRegisteredShippingMethods() {
			if method.GetCode() == paymentMethodCode {
				it.PaymentMethodCode = paymentMethodCode
				break
			}
		}
	case "ShippingMethodCode":
		shippingMethodCode := utils.InterfaceToString(value)
		for _, method := range checkout.GetRegisteredShippingMethods() {
			if method.GetCode() == shippingMethodCode {
				it.ShippingMethodCode = shippingMethodCode
				break
			}
		}

	case "ShippingRate":
		mapValue := utils.InterfaceToMap(value)
		if utils.StrKeysInMap(mapValue, "Name", "Code", "Price") {
			it.ShippingRate.Name = utils.InterfaceToString(mapValue["Name"])
			it.ShippingRate.Code = utils.InterfaceToString(mapValue["Code"])
			it.ShippingRate.Price = utils.InterfaceToFloat64(mapValue["Price"])
		}

	case "Taxes":
		arrayValue := utils.InterfaceToArray(value)
		for _, arrayItem := range arrayValue {
			mapValue := utils.InterfaceToMap(arrayItem)
			if utils.StrKeysInMap(mapValue, "Name", "Code", "Amount") {
				taxRate := checkout.StructTaxRate{
					Name:   utils.InterfaceToString(mapValue["Name"]),
					Code:   utils.InterfaceToString(mapValue["Code"]),
					Amount: utils.InterfaceToFloat64(mapValue["Amount"])}

				if taxRate.Name != "" || taxRate.Code != "" || taxRate.Amount != 0 {
					it.Taxes = append(it.Taxes, taxRate)
				}
			}
		}

	case "Discounts":
		arrayValue := utils.InterfaceToArray(value)
		for _, arrayItem := range arrayValue {
			mapValue := utils.InterfaceToMap(arrayItem)
			if utils.StrKeysInMap(mapValue, "Name", "Code", "Amount") {
				discount := checkout.StructDiscount{
					Name:   utils.InterfaceToString(mapValue["Name"]),
					Code:   utils.InterfaceToString(mapValue["Code"]),
					Amount: utils.InterfaceToFloat64(mapValue["Amount"])}

				if discount.Name != "" || discount.Code != "" || discount.Amount != 0 {
					it.Discounts = append(it.Discounts, discount)
				}
			}
		}

	case "Info":
		info := utils.InterfaceToMap(value)
		if len(info) > 0 {
			it.Info = info
		}

	default:
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "84284b03-0a29-4036-aa2d-b35768884b63", "unsupported 'products' value")
	}
	return nil
}

// FromHashMap fills object attributes from map[string]interface{}
func (it *DefaultCheckout) FromHashMap(input map[string]interface{}) error {

	for attribute, value := range input {
		if err := it.Set(attribute, value); err != nil {
			return env.ErrorDispatch(err)
		}
	}

	return nil
}

// ToHashMap represents object as map[string]interface{}
func (it *DefaultCheckout) ToHashMap() map[string]interface{} {

	result := make(map[string]interface{})

	result["CartID"] = it.CartID
	result["VisitorID"] = it.VisitorID
	result["OrderID"] = it.OrderID
	result["SessionID"] = it.SessionID
	result["ShippingAddress"] = it.ShippingAddress
	result["BillingAddress"] = it.BillingAddress
	result["PaymentMethodCode"] = it.PaymentMethodCode
	result["ShippingMethodCode"] = it.ShippingMethodCode
	result["ShippingRate"] = it.ShippingRate
	result["Taxes"] = it.Taxes
	result["Discounts"] = it.Discounts
	result["Info"] = it.Info

	return result
}

// GetAttributesInfo returns information about object attributes
func (it *DefaultCheckout) GetAttributesInfo() []models.StructAttributeInfo {

	info := []models.StructAttributeInfo{
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "CartID",
			Type:       "id",
			IsRequired: true,
			IsStatic:   true,
			Label:      "Cart ID",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "VisitorID",
			Type:       "id",
			IsRequired: true,
			IsStatic:   true,
			Label:      "Visitor ID",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "SessionID",
			Type:       "id",
			IsRequired: true,
			IsStatic:   true,
			Label:      "Session ID",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "ShippingAddress",
			Type:       "text",
			IsRequired: true,
			IsStatic:   true,
			Label:      "Shipping Address",
			Group:      "General",
			Editors:    "model",
			Options:    "model:VisitorAddress",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "BillingAddress",
			Type:       "text",
			IsRequired: true,
			IsStatic:   true,
			Label:      "Billing Address",
			Group:      "General",
			Editors:    "model",
			Options:    "model:VisitorAddress",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "PaymentMethodCode",
			Type:       "text",
			IsRequired: true,
			IsStatic:   true,
			Label:      "Payment Method",
			Group:      "General",
			Editors:    "line_text",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "ShippingMethodCode",
			Type:       "text",
			IsRequired: true,
			IsStatic:   true,
			Label:      "Shipping Method",
			Group:      "General",
			Editors:    "line_text",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "ShippingRate",
			Type:       "text",
			IsRequired: false,
			IsStatic:   true,
			Label:      "ShippingRate",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "Taxes",
			Type:       "text",
			IsRequired: false,
			IsStatic:   true,
			Label:      "Taxes",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "Discounts",
			Type:       "text",
			IsRequired: false,
			IsStatic:   true,
			Label:      "Discounts",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "Info",
			Type:       "text",
			IsRequired: false,
			IsStatic:   true,
			Label:      "Info",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
	}

	return info
}
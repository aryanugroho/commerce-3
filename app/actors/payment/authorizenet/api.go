package authorizenet

import (
	"github.com/ottemo/commerce/api"
	"github.com/ottemo/commerce/app"
	"github.com/ottemo/commerce/app/models/checkout"
	"github.com/ottemo/commerce/env"
	"github.com/ottemo/commerce/utils"
	"strings"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()
	service.POST("authorizenet/receipt", APIReceipt)
	service.POST("authorizenet/relay", APIRelay)

	return nil
}

// APIReceipt processes Authorize.net receipt response
// can be used for redirecting customer to it on exit from authorize.net
//   - "x_session" should be specified in request contents with id of existing session
//   - refer to http://www.authorize.net/support/DirectPost_guide.pdf for other fields receipt response should contain
func APIReceipt(context api.InterfaceApplicationContext) (interface{}, error) {

	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, err
	}

	status := requestData["x_response_code"]

	session, err := api.GetSessionByID(utils.InterfaceToString(requestData["x_session"]), false)
	if session == nil {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "48f70911-836f-41ba-9ed9-b2afcb7ca462", "Wrong session ID")
	}
	if err := context.SetSession(session); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "0392c562-d4e6-4d8d-9a3d-a7bf6d1620e4", err.Error())
	}

	currentCheckout, err := checkout.GetCurrentCheckout(context, true)
	if err != nil {
		return nil, err
	}

	checkoutOrder := currentCheckout.GetOrder()

	switch status {
	case ConstTransactionApproved:
		{
			currentCart := currentCheckout.GetCart()
			if currentCart == nil {
				return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "61f3dad7-5d23-434a-93b1-c3ef478e1e78", "Cart is not specified")
			}
			if checkoutOrder != nil {

				orderMap, err := currentCheckout.SubmitFinish(requestData)
				if err != nil {
					env.LogError(env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "f2681eb4-27b1-482a-b6f9-ae0ed61ac9e2", "Can't proceed submiting order from Authorize relay"))
				}

				redirectURL := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathDPMReceiptURL))
				if strings.TrimSpace(redirectURL) == "" {
					redirectURL = app.GetStorefrontURL("")
				}

				env.Log(ConstLogStorage, env.ConstLogPrefixInfo, "TRANSACTION APPROVED: "+
					"VisitorID - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", "+
					"OrderID - "+checkoutOrder.GetID()+", "+
					"Card  - "+utils.InterfaceToString(requestData["x_card_type"])+" "+utils.InterfaceToString(requestData["x_account_number"])+", "+
					"Total - "+utils.InterfaceToString(requestData["x_amount"])+", "+
					"Transaction ID - "+utils.InterfaceToString(requestData["x_trans_id"]))

				return api.StructRestRedirect{Result: orderMap, Location: redirectURL, DoRedirect: true}, err
			}
		}
	//	case ConstTransactionDeclined:
	//	case ConstTransactionWaitingReview:
	default:
		{
			if checkoutOrder != nil {
				env.Log(ConstLogStorage, env.ConstLogPrefixError, "TRANSACTION NOT APPROVED: "+
					"VisitorID - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", "+
					"OrderID - "+checkoutOrder.GetID()+", "+
					"Card  - "+utils.InterfaceToString(requestData["x_card_type"])+" "+utils.InterfaceToString(requestData["x_account_number"])+", "+
					"Total - "+utils.InterfaceToString(requestData["x_amount"])+", "+
					"Transaction ID - "+utils.InterfaceToString(requestData["x_trans_id"]))
			}

			redirectURL := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathDPMDeclineURL))
			if strings.TrimSpace(redirectURL) == "" {
				redirectURL = app.GetStorefrontURL("checkout")
			}

			templateContext := map[string]interface{}{
				"backURL":  redirectURL,
				"response": requestData}

			template := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathDPMDeclineHTML))
			if strings.TrimSpace(template) == "" {
				template = ConstDefaultDeclineTemplate
			}

			result, err := utils.TextTemplate(template, templateContext)

			return []byte(result), err
		}
	}
	if checkoutOrder != nil {
		env.Log(ConstLogStorage, env.ConstLogPrefixError, "TRANSACTION NOT APPROVED: (can't process authorize.net response) "+
			"VisitorID - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", "+
			"OrderID - "+checkoutOrder.GetID()+", "+
			"Card  - "+utils.InterfaceToString(requestData["x_card_type"])+" "+utils.InterfaceToString(requestData["x_account_number"])+", "+
			"Total - "+utils.InterfaceToString(requestData["x_amount"])+", "+
			"Transaction ID - "+utils.InterfaceToString(requestData["x_trans_id"]))
	}
	return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "df332f77-2ae1-445d-a8df-17753c822bfe", "can't process authorize.net response")
}

// APIRelay processes Authorize.net relay response
//   - "x_session" should be specified in request contents with id of existing session
//   - refer to http://www.authorize.net/support/DirectPost_guide.pdf for other fields relay response should contain
func APIRelay(context api.InterfaceApplicationContext) (interface{}, error) {

	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, err
	}

	status := requestData["x_response_code"]

	sessionInstance, err := api.GetSessionByID(utils.InterfaceToString(requestData["x_session"]), false)
	if sessionInstance == nil {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "ca56bc43-904a-456a-9df5-03299e713c85", "Wrong session ID")
	}
	if err := context.SetSession(sessionInstance); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "bd356b87-de82-4f74-ad82-a373ebe523a1", err.Error())
	}

	currentCheckout, err := checkout.GetCurrentCheckout(context, true)
	if err != nil {
		return nil, err
	}

	checkoutOrder := currentCheckout.GetOrder()

	switch status {
	case ConstTransactionApproved:
		{
			currentCart := currentCheckout.GetCart()
			if currentCart == nil {
				return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "6244e778-a837-4425-849b-fbce26d5b095", "Cart is not specified")
			}
			if checkoutOrder != nil {

				orderMap, err := currentCheckout.SubmitFinish(requestData)
				if err != nil {
					env.LogError(env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "54296509-fc83-447d-9826-3b7a94ea1acb", "Can't proceed submiting order from Authorize relay"))
					return nil, err
				}

				if err := context.SetResponseContentType("text/plain"); err != nil {
					_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "f30b62cf-e5d2-4736-a15a-c4a9982ecdc5", err.Error())
				}

				env.Log(ConstLogStorage, env.ConstLogPrefixInfo, "TRANSACTION APPROVED: "+
					"VisitorID - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", "+
					"OrderID - "+checkoutOrder.GetID()+", "+
					"Card  - "+utils.InterfaceToString(requestData["x_card_type"])+" "+utils.InterfaceToString(requestData["x_account_number"])+", "+
					"Total - "+utils.InterfaceToString(requestData["x_amount"])+", "+
					"Transaction ID - "+utils.InterfaceToString(requestData["x_trans_id"]))

				redirectURL := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathDPMReceiptURL))
				if strings.TrimSpace(redirectURL) == "" {
					redirectURL = app.GetStorefrontURL("")
				}

				templateContext := map[string]interface{}{
					"backURL":  redirectURL,
					"response": requestData,
					"order":    orderMap,
				}

				template := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathDPMReceiptHTML))
				if strings.TrimSpace(template) == "" {
					template = ConstDefaultReceiptTemplate
				}

				result, err := utils.TextTemplate(template, templateContext)
				if err != nil {
					return result, err
				}

				return []byte(result), nil
			}
		}
	//	case ConstTransactionDeclined:
	//	case ConstTransactionWaitingReview:
	default:
		{
			if checkoutOrder != nil {
				env.Log(ConstLogStorage, env.ConstLogPrefixError, "TRANSACTION NOT APPROVED: "+
					"VisitorID - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", "+
					"OrderID - "+checkoutOrder.GetID()+", "+
					"Card  - "+utils.InterfaceToString(requestData["x_card_type"])+" "+utils.InterfaceToString(requestData["x_account_number"])+", "+
					"Total - "+utils.InterfaceToString(requestData["x_amount"])+", "+
					"Transaction ID - "+utils.InterfaceToString(requestData["x_trans_id"]))
			}

			redirectURL := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathDPMDeclineURL))
			if strings.TrimSpace(redirectURL) == "" {
				redirectURL = app.GetStorefrontURL("checkout")
			}

			templateContext := map[string]interface{}{
				"backURL":  redirectURL,
				"response": requestData}

			template := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathDPMDeclineHTML))
			if strings.TrimSpace(template) == "" {
				template = ConstDefaultDeclineTemplate
			}

			result, err := utils.TextTemplate(template, templateContext)

			return []byte(result), err
		}
	}
	if checkoutOrder != nil {
		env.Log(ConstLogStorage, env.ConstLogPrefixError, "TRANSACTION NOT APPROVED: (can't process authorize.net response) "+
			"VisitorID - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", "+
			"OrderID - "+checkoutOrder.GetID()+", "+
			"Card  - "+utils.InterfaceToString(requestData["x_card_type"])+" "+utils.InterfaceToString(requestData["x_account_number"])+", "+
			"Total - "+utils.InterfaceToString(requestData["x_amount"])+", "+
			"Transaction ID - "+utils.InterfaceToString(requestData["x_trans_id"]))
	}

	return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "770e9dec-8f59-4e98-857f-e8124bf6771e", "can't process authorize.net response")
}

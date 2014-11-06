package paypal

import (
	"fmt"

	"io/ioutil"

	"net/http"
	"net/url"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"

	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/order"
)

/*
 * I_PaymentMethod implementation for:
 *
 *   #1 PayPalExpress
 *   #2 PayPalREST
 *
 */

//------------------
// #1 PayPalExpress
//------------------

// GetName returns the payment method name
func (it *PayPalExpress) GetName() string {
	return utils.InterfaceToString(env.ConfigGetValue(CONFIG_PATH_TITLE))
}

// GetCode returns the payment method code
func (it *PayPalExpress) GetCode() string {
	return PAYMENT_CODE
}

// GetType returns the type of payment method
func (it *PayPalExpress) GetType() string {
	return checkout.PAYMENT_TYPE_REMOTE
}

// IsAllowed checks for method applicability
func (it *PayPalExpress) IsAllowed(checkoutInstance checkout.I_Checkout) bool {
	return utils.InterfaceToBool(env.ConfigGetValue(CONFIG_PATH_ENABLED))
}

// Authorize executes the payment method authorize operation
func (it *PayPalExpress) Authorize(orderInstance order.I_Order, paymentInfo map[string]interface{}) (interface{}, error) {

	// getting order information
	//--------------------------
	grandTotal := orderInstance.GetGrandTotal()
	shippingPrice := orderInstance.GetShippingAmount()

	// getting request param values
	//-----------------------------
	user := utils.InterfaceToString(env.ConfigGetValue(CONFIG_PATH_USER))
	password := utils.InterfaceToString(env.ConfigGetValue(CONFIG_PATH_PASS))
	signature := utils.InterfaceToString(env.ConfigGetValue(CONFIG_PATH_SIGNATURE))
	action := utils.InterfaceToString(env.ConfigGetValue(CONFIG_PATH_ACTION))

	amount := fmt.Sprintf("%.2f", grandTotal)
	shippingAmount := fmt.Sprintf("%.2f", shippingPrice)
	itemAmount := fmt.Sprintf("%.2f", grandTotal-shippingPrice)

	description := "Purchase%20for%20%24" + fmt.Sprintf("%.2f", grandTotal)
	custom := orderInstance.GetId()

	cancelURL := app.GetFoundationUrl("paypal/cancel")
	returnURL := app.GetFoundationUrl("paypal/success")

	// making NVP request
	//-------------------
	requestParams := "USER=" + user +
		"&PWD=" + password +
		"&SIGNATURE=" + signature +
		"&METHOD=SetExpressCheckout" +
		"&VERSION=78" +
		"&PAYMENTREQUEST_0_PAYMENTACTION=" + action +
		"&PAYMENTREQUEST_0_AMT=" + amount +
		"&PAYMENTREQUEST_0_SHIPPINGAMT=" + shippingAmount +
		"&PAYMENTREQUEST_0_ITEMAMT=" + itemAmount +
		"&PAYMENTREQUEST_0_DESC=" + description +
		"&PAYMENTREQUEST_0_CUSTOM=" + custom +
		"&PAYMENTREQUEST_0_CURRENCYCODE=USD" +
		"&cancelURL=" + cancelURL +
		"&returnURL=" + returnURL

	//	println(requestParams)

	nvpGateway := utils.InterfaceToString(env.ConfigGetValue(CONFIG_PATH_NVP))

	request, err := http.NewRequest("GET", nvpGateway+"?"+requestParams, nil)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// reading/decoding response from NVP
	//-----------------------------------
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	responseValues, err := url.ParseQuery(string(responseData))
	if err != nil {
		return nil, env.ErrorNew("payment unexpected response")
	}

	if responseValues.Get("ACK") != "Success" || responseValues.Get("TOKEN") == "" {
		if responseValues.Get("L_ERRORCODE0") != "" {
			return nil, env.ErrorNew("payment error " + responseValues.Get("L_ERRORCODE0") + ": " + "L_LONGMESSAGE0")
		}
	}
	waitingTokensMutex.Lock()
	waitingTokens[responseValues.Get("TOKEN")] = utils.InterfaceToString(paymentInfo["sessionId"])
	waitingTokensMutex.Unlock()

	env.Log("paypal.log", env.LOG_PREFIX_INFO, "NEW TRANSACTION: "+
		"Visitor ID - "+utils.InterfaceToString(orderInstance.Get("visitor_id"))+", "+
		"Order ID - "+utils.InterfaceToString(orderInstance.GetId())+", "+
		"TOKEN - "+utils.InterfaceToString(responseValues.Get("TOKEN")))

	// redirecting user to PayPal server for following checkout
	//---------------------------------------------------------
	redirectGateway := utils.InterfaceToString(env.ConfigGetValue(CONFIG_PATH_GATEWAY)) + "&token=" + responseValues.Get("TOKEN")
	return api.T_RestRedirect{
		Result:   "redirect",
		Location: redirectGateway,
	}, nil
}

// Capture executes the payment method capture operation
func (it *PayPalExpress) Capture(orderInstance order.I_Order, paymentInfo map[string]interface{}) (interface{}, error) {
	return nil, env.ErrorNew("Not implemented")
}

// Refund will return funds to the visitor for the given order. :: Not Implemented Yet
func (it *PayPalExpress) Refund(orderInstance order.I_Order, paymentInfo map[string]interface{}) (interface{}, error) {
	return nil, env.ErrorNew("Not implemented")
}

// Void will void the givien order :: Not Implemented YET
func (it *PayPalExpress) Void(orderInstance order.I_Order, paymentInfo map[string]interface{}) (interface{}, error) {
	return nil, env.ErrorNew("Not implemented")
}

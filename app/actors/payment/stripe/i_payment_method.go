package stripe

import (
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"

	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/order"
)

// GetCode will return the Stripe payment method code
func (it *Payment) GetCode() string {
	return ConstPaymentCode
}

// GetInternalName returns the internal payment method name for Stripe
func (it *Payment) GetInternalName() string {
	return ConstPaymentName
}

// GetName returns the payment method name used for Stripe in checkout
func (it *Payment) GetName() string {
	return it.ConfigNameInCheckout()
}

// GetType returns the credit card type used for payment
func (it *Payment) GetType() string {
	return checkout.ConstPaymentTypeCreditCard
}

// IsAllowed is a flag to check if the Stripe payment method is enabled in the current store
func (it *Payment) IsAllowed(checkoutInstance checkout.InterfaceCheckout) bool {
	return it.ConfigIsEnabled()
}

// IsTokenable is a flag to check if the current payment method is able to save payment tokens for future use
func (it *Payment) IsTokenable(checkoutInstance checkout.InterfaceCheckout) bool {
	return false
}

// Authorize is the method used to determine if the provided payment information by the visitor is valid and has
// enough credit available to reconcile payment
func (it *Payment) Authorize(orderInstance order.InterfaceOrder, paymentInfo map[string]interface{}) (interface{}, error) {
	stripe.Key = it.ConfigAPIKey()

	// Create a charge
	// https://stripe.com/docs/api/go#create_charge
	chargeParams := &stripe.ChargeParams{
		Amount:   uint64(orderInstance.GetGrandTotal() * 100), // Amount is in cents
		Currency: "usd",
	}

	// Store the email in meta data in case we need to track it later
	email := utils.InterfaceToString(orderInstance.Get("customer_email"))
	chargeParams.AddMeta("email", email)

	ccInfo := utils.InterfaceToMap(paymentInfo["cc"])
	ccNumber := utils.InterfaceToString(ccInfo["number"])
	ccMonth := utils.InterfaceToString(ccInfo["expire_month"])
	ccYear := utils.InterfaceToString(ccInfo["expire_year"])
	ccCVC := utils.InterfaceToString(ccInfo["cvc"])
	ccName := orderInstance.GetBillingAddress().GetFirstName() + " " + orderInstance.GetBillingAddress().GetLastName()

	// We are enforcing cvc
	if ccCVC == "" {
		err := env.ErrorNew(ConstErrorModule, env.ConstErrorLevelStartStop, "15edae76-1d3e-4e7a-a474-75ffb61d26cb", "CVC field was left empty")
		return nil, env.ErrorDispatch(err)
	}

	cardParams := stripe.CardParams{
		Number: ccNumber,
		Month:  ccMonth,
		Year:   ccYear,
		CVC:    ccCVC,  // Optional, highly recommended
		Name:   ccName, // Optional

		// Address fields can be passed here as well to aid in fraud prevention
	}

	// Must attach either `customer` or `source` to charge
	// source can be either a `token` or `cardParams`
	chargeParams.SetSource(&cardParams)

	ch, err := charge.New(chargeParams)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// Assemble the response information
	orderPaymentInfo := map[string]interface{}{
		"transactionID":     ch.ID,
		"creditCardNumbers": ch.Source.Card.LastFour,
		"creditCardExp":     ccMonth + utils.InterfaceToString(ch.Source.Card.Year)[:2], // format mmyy
		"creditCardType":    getCCBrand(string(ch.Source.Card.Brand)),
	}

	return orderPaymentInfo, nil
}

// Capture is the payment method used to capture authorized funds.  **This method is for future use**
func (it *Payment) Capture(orderInstance order.InterfaceOrder, paymentInfo map[string]interface{}) (interface{}, error) {
	return nil, env.ErrorNew(ConstErrorModule, 1, "05199a06-7bd4-49b6-9fb0-0f1589a9cd74", "called but not implemented")
}

// Refund is the payment method used to refund a visitor on behalf of a merchant. **This method is for future use**
func (it *Payment) Refund(orderInstance order.InterfaceOrder, paymentInfo map[string]interface{}) (interface{}, error) {
	return nil, env.ErrorNew(ConstErrorModule, 1, "c8768719-80ab-453d-b52e-513dfb4aab22", "called but not implemented")
}

// Void is the payment method used to cancel a visitor transaction before funds have been collected.  **This method is for future use**
func (it *Payment) Void(orderInstance order.InterfaceOrder, paymentInfo map[string]interface{}) (interface{}, error) {
	return nil, env.ErrorNew(ConstErrorModule, 1, "4194a950-18fd-4b0d-96e6-e33e930f4320", "called but not implemented")
}
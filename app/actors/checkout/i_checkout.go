package checkout

import (
	"fmt"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"

	"github.com/ottemo/foundation/app/models/cart"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/app/models/visitor"
)

// SetShippingAddress sets shipping address for checkout
func (it *DefaultCheckout) SetShippingAddress(address visitor.InterfaceVisitorAddress) error {
	if address == nil {
		it.ShippingAddress = nil
		return nil
	}

	it.ShippingAddress = address.ToHashMap()
	return nil
}

// GetShippingAddress returns checkout shipping address
func (it *DefaultCheckout) GetShippingAddress() visitor.InterfaceVisitorAddress {
	if it.ShippingAddress == nil {
		return nil
	}

	shippingAddress, err := visitor.GetVisitorAddressModel()
	if err != nil {
		env.ErrorDispatch(err)
		return nil
	}

	err = shippingAddress.FromHashMap(it.ShippingAddress)
	if err != nil {
		env.ErrorDispatch(err)
		return nil
	}

	return shippingAddress
}

// SetBillingAddress sets billing address for checkout
func (it *DefaultCheckout) SetBillingAddress(address visitor.InterfaceVisitorAddress) error {
	if address == nil {
		it.BillingAddress = nil
		return nil
	}

	it.BillingAddress = address.ToHashMap()
	return nil
}

// GetBillingAddress returns checkout billing address
func (it *DefaultCheckout) GetBillingAddress() visitor.InterfaceVisitorAddress {
	if it.BillingAddress == nil {
		return nil
	}

	billingAddress, err := visitor.GetVisitorAddressModel()
	if err != nil {
		env.ErrorDispatch(err)
		return nil
	}

	err = billingAddress.FromHashMap(it.BillingAddress)
	if err != nil {
		env.ErrorDispatch(err)
		return nil
	}

	return billingAddress
}

// SetPaymentMethod sets payment method for checkout
func (it *DefaultCheckout) SetPaymentMethod(paymentMethod checkout.InterfacePaymentMethod) error {
	it.PaymentMethodCode = paymentMethod.GetCode()
	return nil
}

// GetPaymentMethod returns checkout payment method
func (it *DefaultCheckout) GetPaymentMethod() checkout.InterfacePaymentMethod {
	if paymentMethods := checkout.GetRegisteredPaymentMethods(); paymentMethods != nil {
		for _, paymentMethod := range paymentMethods {
			if paymentMethod.GetCode() == it.PaymentMethodCode {
				return paymentMethod
			}
		}
	}
	return nil
}

// SetShippingMethod sets payment method for checkout
func (it *DefaultCheckout) SetShippingMethod(shippingMethod checkout.InterfaceShippingMethod) error {
	it.ShippingMethodCode = shippingMethod.GetCode()
	return nil
}

// GetShippingMethod returns a checkout shipping method
func (it *DefaultCheckout) GetShippingMethod() checkout.InterfaceShippingMethod {
	if shippingMethods := checkout.GetRegisteredShippingMethods(); shippingMethods != nil {
		for _, shippingMethod := range shippingMethods {
			if shippingMethod.GetCode() == it.ShippingMethodCode {
				return shippingMethod
			}
		}
	}
	return nil
}

// SetShippingRate sets shipping rate for checkout
func (it *DefaultCheckout) SetShippingRate(shippingRate checkout.StructShippingRate) error {
	it.ShippingRate = shippingRate
	return nil
}

// GetShippingRate returns a checkout shipping rate
func (it *DefaultCheckout) GetShippingRate() *checkout.StructShippingRate {
	return &it.ShippingRate
}

// SetCart sets cart for checkout
func (it *DefaultCheckout) SetCart(checkoutCart cart.InterfaceCart) error {
	it.CartID = checkoutCart.GetID()
	return nil
}

// GetCart returns a shopping cart
func (it *DefaultCheckout) GetCart() cart.InterfaceCart {
	cartInstance, _ := cart.LoadCartByID(it.CartID)
	return cartInstance
}

// SetVisitor sets visitor for checkout
func (it *DefaultCheckout) SetVisitor(checkoutVisitor visitor.InterfaceVisitor) error {
	it.VisitorID = checkoutVisitor.GetID()

	if it.BillingAddress == nil && checkoutVisitor.GetBillingAddress() != nil {
		it.BillingAddress = checkoutVisitor.GetBillingAddress().ToHashMap()
	}

	if it.ShippingAddress == nil && checkoutVisitor.GetShippingAddress() != nil {
		it.ShippingAddress = checkoutVisitor.GetShippingAddress().ToHashMap()
	}

	return nil
}

// GetVisitor return checkout visitor
func (it *DefaultCheckout) GetVisitor() visitor.InterfaceVisitor {
	if it.VisitorID == "" {
		return nil
	}

	visitorInstance, err := visitor.LoadVisitorByID(it.VisitorID)
	if err != nil {
		return nil
	}

	return visitorInstance
}

// SetSession sets visitor for checkout
func (it *DefaultCheckout) SetSession(checkoutSession api.InterfaceSession) error {
	it.SessionID = checkoutSession.GetID()
	return nil
}

// GetSession return checkout visitor
func (it *DefaultCheckout) GetSession() api.InterfaceSession {
	sessionInstance, _ := api.GetSessionByID(it.SessionID)
	return sessionInstance
}

// GetTaxAmount returns total amount of taxes for current checkout
func (it *DefaultCheckout) GetTaxAmount() float64 {

	return it.taxesAmount
}

// GetDiscountAmount returns total amount of discounts applied for current checkout
func (it *DefaultCheckout) GetDiscountAmount() float64 {

	return it.discountsAmount
}

// GetTaxes collects taxes applied for current checkout
func (it *DefaultCheckout) GetTaxes() []checkout.StructTaxRate {

	if !it.taxesCalculateFlag && it.calculateFlag {
		it.taxesCalculateFlag = true

		it.Taxes = make([]checkout.StructTaxRate, 0)
		for _, tax := range checkout.GetRegisteredTaxes() {
			for _, taxRate := range tax.CalculateTax(it) {
				it.Taxes = append(it.Taxes, taxRate)
			}
		}

		it.taxesCalculateFlag = false
	}

	return it.Taxes
}

// GetDiscounts collects discounts applied for current checkout
func (it *DefaultCheckout) GetDiscounts() []checkout.StructDiscount {

	if !it.discountsCalculateFlag && it.calculateFlag {
		it.discountsCalculateFlag = true

		it.Discounts = make([]checkout.StructDiscount, 0)
		for _, discount := range checkout.GetRegisteredDiscounts() {
			for _, discountValue := range discount.CalculateDiscount(it) {
				it.Discounts = append(it.Discounts, discountValue)
			}
		}

		it.discountsCalculateFlag = false
	}

	return it.Discounts
}

// GetSubtotal returns subtotal total for current checkout
func (it *DefaultCheckout) GetSubtotal() float64 {
	currentCart := it.GetCart()
	if currentCart != nil {
		return currentCart.GetSubtotal()
	}

	return 0
}

// GetShippingAmount returns shipping price for current checkout
func (it *DefaultCheckout) GetShippingAmount() float64 {
	if shippingRate := it.GetShippingRate(); shippingRate != nil {
		return shippingRate.Price
	}
	return 0
}

// CalculateAmount do a calculation of all amounts for checkout
// TODO: make function use calculateTarget as a limit for priority to where it need to be calculated
func (it *DefaultCheckout) CalculateAmount(calculateTarget float64) float64 {

	if !it.calculateFlag {
		it.calculateFlag = true

		discounts := it.GetDiscounts()
		taxes := it.GetTaxes()

		basePoints := map[float64]func() float64{
			checkout.ConstCalculateTargetSubtotal:   func() float64 { return it.GetSubtotal() },
			checkout.ConstCalculateTargetShipping:   func() float64 { return it.GetShippingAmount() },
			checkout.ConstCalculateTargetGrandTotal: func() float64 { return 0 },
		}

		it.calculateAmount = 0
		it.discountsAmount = 0
		it.taxesAmount = 0

		var minPriority float64
		var maxPriority float64

		minIsSet := false
		maxIsSet := false

		searchMode := true

		// 2 cycle calculation loop
		// 1st loop - search mode, looks for current minimal priority to apply
		// 2nd loop - apply current priority items
		for searchMode || maxIsSet {

			// setting previousPriority since 2nd search
			if searchMode && maxIsSet {
				minPriority = maxPriority
				minIsSet = true
				maxIsSet = false
			}

			// base points lookup (subtotal, shipping)
			for priority, value := range basePoints {

				if searchMode {
					if (!maxIsSet || priority < maxPriority) && (!minIsSet || priority > minPriority) {
						maxPriority = priority
						maxIsSet = true
					}
				} else {
					if priority == maxPriority {
						it.calculateAmount += value()
					}
				}
			}

			// discounts lookup
			for index, discount := range discounts {

				if searchMode {
					priority := discount.Priority
					if (!maxIsSet || priority < maxPriority) && (!minIsSet || priority > minPriority) {
						maxPriority = discount.Priority
						maxIsSet = true
					}
				} else {
					if discount.Priority == maxPriority {
						amount := discount.Amount
						if discount.IsPercent {
							amount = it.calculateAmount * discount.Amount / 100
						}

						// prevent negative values for grand total subtract
						if amount > it.calculateAmount {
							amount = it.calculateAmount
						}

						// round amount add it to calculating amounts and set to discount amount
						amount = utils.RoundPrice(amount)
						discounts[index].Amount = amount
						it.discountsAmount += amount
						it.calculateAmount -= amount
					}
				}
			}

			// taxes lookup
			for index, tax := range taxes {
				if searchMode {
					priority := tax.Priority
					if (!maxIsSet || priority < maxPriority) && (!minIsSet || priority > minPriority) {
						maxPriority = tax.Priority
						maxIsSet = true
					}
				} else {
					if tax.Priority == maxPriority {
						amount := tax.Amount
						if tax.IsPercent {
							amount = it.calculateAmount * tax.Amount / 100
						}

						// round amount add it to calculating amounts and set to taxes amount
						amount = utils.RoundPrice(amount)
						taxes[index].Amount = amount
						it.taxesAmount += amount
						it.calculateAmount += amount
					}
				}
			}

			// cycle mode switcher
			if searchMode {
				searchMode = false
			} else {
				searchMode = true
			}
		}

		it.Discounts = discounts
		it.Taxes = taxes

		it.calculateAmount = utils.RoundPrice(it.calculateAmount)
		it.taxesAmount = utils.RoundPrice(it.taxesAmount)
		it.discountsAmount = utils.RoundPrice(it.discountsAmount)

		it.calculateFlag = false
	}

	return it.calculateAmount
}

// GetGrandTotal returns grand total for current checkout: [cart subtotal] + [shipping rate] + [taxes] - [discounts]
func (it *DefaultCheckout) GetGrandTotal() float64 {
	return it.CalculateAmount(0)
}

// SetInfo sets additional info for checkout - any values related to checkout process
func (it *DefaultCheckout) SetInfo(key string, value interface{}) error {
	if value == nil {
		if _, present := it.Info[key]; present {
			delete(it.Info, key)
		}
	} else {
		it.Info[key] = value
	}

	return nil
}

// GetInfo returns additional checkout info value or nil,
//   - use "*" as a key to get all keys and values currently set
func (it *DefaultCheckout) GetInfo(key string) interface{} {
	if key == "*" {
		return it.Info
	}

	if value, present := it.Info[key]; present {
		return value
	}

	return nil
}

// SetOrder sets order for current checkout
func (it *DefaultCheckout) SetOrder(checkoutOrder order.InterfaceOrder) error {
	it.OrderID = checkoutOrder.GetID()
	return nil
}

// GetOrder returns current checkout related order or nil if not created yet
func (it *DefaultCheckout) GetOrder() order.InterfaceOrder {
	if it.OrderID != "" {
		orderInstance, err := order.LoadOrderByID(it.OrderID)
		if err == nil {
			return orderInstance
		}
	}
	return nil
}

// Submit creates the order with provided information
func (it *DefaultCheckout) Submit() (interface{}, error) {

	if it.GetBillingAddress() == nil {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "080db3c0-dbb5-4398-b1f1-4c3fefef79b4", "Billing address is not set")
	}

	if it.GetShippingAddress() == nil {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "1c069d54-2847-46cb-bccd-76fc13d229ea", "Shipping address is not set")
	}

	if it.GetPaymentMethod() == nil {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "c0628038-3e06-47e9-9252-480351d903c0", "Payment method is not set")
	}

	if it.GetShippingMethod() == nil {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "e22877fe-248d-4b5e-ad2f-10843cb9890c", "Shipping method is not set")
	}

	currentCart := it.GetCart()
	if currentCart == nil {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "ea211827-2025-46ca-be47-d09841021890", "Cart is not specified")
	}

	cartItems := currentCart.GetItems()
	if len(cartItems) == 0 {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "84ad4db5-29e9-430c-aecf-b675cbafbcec", "Cart is empty")
	}

	currentVisitor := it.GetVisitor()
	if (it.VisitorID == "" || currentVisitor == nil) && utils.InterfaceToString(it.GetInfo("customer_email")) == "" {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "c5f53ede-63b7-40ea-952d-4d4c04337563", "customer e-mail was not specified")
	}

	if err := currentCart.ValidateCart(); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// making new order if needed
	//---------------------------
	currentTime := time.Now()

	checkoutOrder := it.GetOrder()
	if checkoutOrder == nil {
		newOrder, err := order.GetOrderModel()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		newOrder.Set("created_at", currentTime)

		checkoutOrder = newOrder
	}

	// updating order information
	//---------------------------
	checkoutOrder.Set("updated_at", currentTime)

	checkoutOrder.SetStatus(order.ConstOrderStatusNew)

	if currentVisitor != nil {
		checkoutOrder.Set("visitor_id", currentVisitor.GetID())

		checkoutOrder.Set("customer_email", currentVisitor.GetEmail())
		checkoutOrder.Set("customer_name", currentVisitor.GetFullName())
	}

	if it.GetInfo("customer_email") != nil {
		orderCustomerEmail := utils.InterfaceToString(it.GetInfo("customer_email"))
		checkoutOrder.Set("customer_email", orderCustomerEmail)
	}
	if it.GetInfo("customer_name") != nil {
		orderCustomerName := utils.InterfaceToString(it.GetInfo("customer_name"))
		checkoutOrder.Set("customer_name", orderCustomerName)
	}

	billingAddress := it.GetBillingAddress().ToHashMap()
	checkoutOrder.Set("billing_address", billingAddress)

	shippingAddress := it.GetShippingAddress().ToHashMap()
	checkoutOrder.Set("shipping_address", shippingAddress)

	checkoutOrder.Set("cart_id", currentCart.GetID())

	paymentMethod := it.GetPaymentMethod()

	if !paymentMethod.IsAllowed(it) {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "7a5490ee-daa3-42b4-a84a-dade12d103e8", "Payment method not allowed")
	}

	checkoutOrder.Set("payment_method", paymentMethod.GetCode())
	checkoutOrder.Set("shipping_method", it.GetShippingMethod().GetCode()+"/"+it.GetShippingRate().Code)

	discounts := it.GetDiscounts()
	discountAmount := it.GetDiscountAmount()
	checkoutOrder.Set("discount", discountAmount)
	checkoutOrder.Set("discounts", discounts)

	taxes := it.GetTaxes()
	taxAmount := it.GetTaxAmount()
	checkoutOrder.Set("tax_amount", taxAmount)
	checkoutOrder.Set("taxes", taxes)

	checkoutOrder.Set("shipping_amount", it.GetShippingRate().Price)

	generateDescriptionFlag := false
	orderDescription := utils.InterfaceToString(it.GetInfo("order_description"))
	if orderDescription == "" {
		generateDescriptionFlag = true
	}

	for _, cartItem := range cartItems {
		orderItem, err := checkoutOrder.AddItem(cartItem.GetProductID(), cartItem.GetQty(), cartItem.GetOptions())
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		if generateDescriptionFlag {
			if orderDescription != "" {
				orderDescription += ", "
			}
			orderDescription += fmt.Sprintf("%dx %s", cartItem.GetQty(), orderItem.GetName())
		}
	}
	checkoutOrder.Set("description", orderDescription)

	err := checkoutOrder.CalculateTotals()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = checkoutOrder.SetStatus(order.ConstOrderStatusPending)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	it.SetOrder(checkoutOrder)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// trying to process payment
	//--------------------------
	if checkoutOrder.GetGrandTotal() > 0 {
		paymentInfo := make(map[string]interface{})
		paymentInfo["sessionID"] = it.GetSession().GetID()
		paymentInfo["cc"] = it.GetInfo("cc")

		result, err := paymentMethod.Authorize(checkoutOrder, paymentInfo)
		if err != nil {
			checkoutOrder.SetStatus(order.ConstOrderStatusNew)
			return nil, env.ErrorDispatch(err)
		}

		// if payment.Authorize returns non nil result, that supposing additional operations to complete payment
		if result != nil {
			return result, nil
		}
	}

	err = it.CheckoutSuccess(checkoutOrder, it.GetSession())

	return checkoutOrder.ToHashMap(), err
}

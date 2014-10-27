package authorize

import (
	"errors"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/api/session"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

const (
	TRANSACTION_APPROVED         = "1"
	TRANSACTION_DECLINED         = "2"
	TRANSACTION_ERROR            = "3"
	TRANSACTION_WAITING_REVIEWED = "4"
)

// startup API registration
func setupAPI() error {

	var err error = nil

	err = api.GetRestService().RegisterAPI("authorizenet", "POST", "receipt", restReceipt)
	if err != nil {
		return err
	}

	err = api.GetRestService().RegisterAPI("authorizenet", "POST", "relay", restRelay)
	if err != nil {
		return err
	}

	return nil
}

// WEB REST API function to process Authorize.Net receipt result
func restReceipt(params *api.T_APIHandlerParams) (interface{}, error) {

	postData := params.RequestContent.(map[string]interface{})

	status := postData["x_response_code"]

	session, err := session.GetSessionById(utils.InterfaceToString(postData["x_session"]))
	if err != nil {
		return nil, errors.New("Wrong session ID")
	}
	params.Session = session

	currentCheckout, err := checkout.GetCurrentCheckout(params)
	if err != nil {
		return nil, err
	}

	checkoutOrder := currentCheckout.GetOrder()

	switch status {
	case TRANSACTION_APPROVED:
		{
			currentCart := currentCheckout.GetCart()
			if currentCart == nil {
				return nil, errors.New("Cart is not specified")
			}
			if checkoutOrder != nil {
				checkoutOrder.NewIncrementId()

				checkoutOrder.Set("status", "pending")
				checkoutOrder.Set("payment_info", postData)

				err = currentCheckout.CheckoutSuccess(checkoutOrder, params.Session)
				if err != nil {
					return nil, err
				}

				// Send confirmation email
				err = currentCheckout.SendOrderConfirmationMail()
				if err != nil {
					return nil, err
				}

				env.Log("authorizenet.log", env.LOG_PREFIX_INFO, "TRANSACTION APPROVED: "+
					"VisitorId - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", "+
					"OrderId - "+checkoutOrder.GetId()+", "+
					"Card  - "+utils.InterfaceToString(postData["x_card_type"])+" "+utils.InterfaceToString(postData["x_account_number"])+", "+
					"Total - "+utils.InterfaceToString(postData["x_amount"])+", "+
					"Transaction ID - "+utils.InterfaceToString(postData["x_trans_id"]))

				return api.T_RestRedirect{Location: app.GetStorefrontUrl("account/order/" + checkoutOrder.GetId()), DoRedirect: true}, nil
			}
		}
	case TRANSACTION_DECLINED:
	case TRANSACTION_WAITING_REVIEWED:
	default:
		{
			if checkoutOrder != nil {
				env.Log("authorizenet.log", env.LOG_PREFIX_ERROR, "TRANSACTION NOT APPROVED: " +
							"VisitorId - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", " +
							"OrderId - "+checkoutOrder.GetId()+", " +
							"Card  - "+utils.InterfaceToString(postData["x_card_type"])+" "+utils.InterfaceToString(postData["x_account_number"])+", " +
							"Total - "+utils.InterfaceToString(postData["x_amount"])+", " +
							"Transaction ID - "+utils.InterfaceToString(postData["x_trans_id"]))
			}

			return []byte(`<html>
					 <head>
						 <noscript>
						 	<meta http-equiv='refresh' content='1;url=` + app.GetStorefrontUrl("checkout") + `'>
						 </noscript>
					 </head>
					 <body>
					 	<h1>Something went wrong</h1>
					 	<p>` + utils.InterfaceToString(postData["x_response_reason_text"]) + `</p>

						<p><a href="` + app.GetStorefrontUrl("checkout") + `">Back to store</a></p>

					 </body>
				</html>`), nil
		}
	}
	if checkoutOrder != nil {
		env.Log("authorizenet.log", env.LOG_PREFIX_ERROR, "TRANSACTION NOT APPROVED: (can't process authorize.net response) " +
					"VisitorId - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", " +
					"OrderId - "+checkoutOrder.GetId()+", " +
					"Card  - "+utils.InterfaceToString(postData["x_card_type"])+" "+utils.InterfaceToString(postData["x_account_number"])+", " +
					"Total - "+utils.InterfaceToString(postData["x_amount"])+", " +
					"Transaction ID - "+utils.InterfaceToString(postData["x_trans_id"]))
	}
	return nil, errors.New("can't process authorize.net response")
}

// WEB REST API function to process Authorize.Net relay result
func restRelay(params *api.T_APIHandlerParams) (interface{}, error) {

	postData := params.RequestContent.(map[string]interface{})

	status := postData["x_response_code"]

	session, err := session.GetSessionById(utils.InterfaceToString(postData["x_session"]))
	if err != nil {
		return nil, errors.New("Wrong session ID")
	}
	params.Session = session

	currentCheckout, err := checkout.GetCurrentCheckout(params)
	if err != nil {
		return nil, err
	}

	checkoutOrder := currentCheckout.GetOrder()

	switch status {
	case TRANSACTION_APPROVED:
		{
			currentCart := currentCheckout.GetCart()
			if currentCart == nil {
				return nil, errors.New("Cart is not specified")
			}
			if checkoutOrder != nil {
				checkoutOrder.NewIncrementId()

				checkoutOrder.Set("status", "pending")
				checkoutOrder.Set("payment_info", postData)

				err = currentCheckout.CheckoutSuccess(checkoutOrder, params.Session)
				if err != nil {
					return nil, err
				}

				// Send confirmation email
				err = currentCheckout.SendOrderConfirmationMail()
				if err != nil {
					return nil, err
				}

				params.ResponseWriter.Header().Set("Content-Type", "text/plain")

				env.Log("authorizenet.log", env.LOG_PREFIX_INFO, "TRANSACTION APPROVED: "+
					"VisitorId - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", "+
					"OrderId - "+checkoutOrder.GetId()+", "+
					"Card  - "+utils.InterfaceToString(postData["x_card_type"])+" "+utils.InterfaceToString(postData["x_account_number"])+", "+
					"Total - "+utils.InterfaceToString(postData["x_amount"])+", "+
					"Transaction ID - "+utils.InterfaceToString(postData["x_trans_id"]))

				return []byte(`<html>
					 <head>
						 <noscript>
						 	<meta http-equiv='refresh' content='1;url=` + app.GetStorefrontUrl("account/order/"+checkoutOrder.GetId()) + `'>
						 </noscript>
					 </head>
					 <body>
					 	<h1>Thanks for your purchase.</h1>
					 	<p>Your transaction ID: <b>` + utils.InterfaceToString(postData["x_trans_id"]) + `</b></p>
					 	<p>You will  redirect to the store after <span id="sec"></span> sec.	<a href="` + app.GetStorefrontUrl("account/order/"+checkoutOrder.GetId()) + `">Back to store</a></p>
					 </body>
					 <script type='text/javascript' charset='utf-8'>
					 	(function(){
							var seconds = 10;
							document.getElementById("sec").innerHTML = seconds;
							setInterval(function(){
								seconds -= 1;
								document.getElementById("sec").innerHTML = seconds;
								if(0 === seconds){
									window.location='` + app.GetStorefrontUrl("account/order/"+checkoutOrder.GetId()) + `';
								}
							}, 1000);
					 	})();
					 </script>
				</html>`), nil
			}
		}
	case TRANSACTION_DECLINED:
	case TRANSACTION_WAITING_REVIEWED:
	default:
		{
			if checkoutOrder != nil {
				env.Log("authorizenet.log", env.LOG_PREFIX_ERROR, "TRANSACTION NOT APPROVED: "+
					"VisitorId - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", "+
					"OrderId - "+checkoutOrder.GetId()+", "+
					"Card  - "+utils.InterfaceToString(postData["x_card_type"])+" "+utils.InterfaceToString(postData["x_account_number"])+", "+
					"Total - "+utils.InterfaceToString(postData["x_amount"])+", "+
					"Transaction ID - "+utils.InterfaceToString(postData["x_trans_id"]))
			}
			return []byte(`<html>
					 <head>
						 <noscript>
						 	<meta http-equiv='refresh' content='1;url=` + app.GetStorefrontUrl("checkout") + `'>
						 </noscript>
					 </head>
					 <body>
					 	<h1>Something went wrong</h1>
					 	<p>` + utils.InterfaceToString(postData["x_response_reason_text"]) + `</p>

						<p><a href="` + app.GetStorefrontUrl("checkout") + `">Back to store</a></p>

					 </body>
				</html>`), nil
		}
	}
	if checkoutOrder != nil {
		env.Log("authorizenet.log", env.LOG_PREFIX_ERROR, "TRANSACTION NOT APPROVED: (can't process authorize.net response) " +
					"VisitorId - "+utils.InterfaceToString(checkoutOrder.Get("visitor_id"))+", " +
					"OrderId - "+checkoutOrder.GetId()+", " +
					"Card  - "+utils.InterfaceToString(postData["x_card_type"])+" "+utils.InterfaceToString(postData["x_account_number"])+", " +
					"Total - "+utils.InterfaceToString(postData["x_amount"])+", " +
					"Transaction ID - "+utils.InterfaceToString(postData["x_trans_id"]))
	}

	return nil, errors.New("can't process authorize.net response")
}

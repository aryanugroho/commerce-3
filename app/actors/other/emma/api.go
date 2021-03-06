package emma

import (
	"github.com/ottemo/commerce/api"
	"github.com/ottemo/commerce/env"
	"github.com/ottemo/commerce/utils"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()

	// Public
	service.POST("emma/contact", APIEmmaAddContact)

	return nil
}

// APIEmmaAddContact - return message, after add contact
// - email should be specified in "email" argument
func APIEmmaAddContact(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if !utils.KeysInMapAndNotBlank(requestData, "email") {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "6372b9a3-29f3-4ea4-a19f-40051a8f330b", "email has not been specified")
	}
	email := utils.InterfaceToString(requestData["email"])

	if !utils.KeysInMapAndNotBlank(requestData, "group_ids") {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "eee81283-86c4-487c-a5b5-b78996be038e", "group_ids not specified")
	}
	groupIDs := utils.InterfaceToString(requestData["group_ids"])

	if !utils.ValidEmailAddress(email) {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "b54b0917-acc0-469f-925e-8f85a1feac7b", "The email address, "+email+", is not in valid format.")
	}

	result, err := subscribe(email, groupIDs)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	return result, nil
}

// Package trustpilot implements trust pilot functions
package trustpilot

import (
	"github.com/ottemo/foundation/env"
)

// Package global constants
const (
	ConstProductBrand = "Kari Gran"
	ConstEmailSubject = "Purchase feedback"

	ConstErrorModule = "trustpilot"
	ConstErrorLevel  = env.ConstErrorLevelActor

	ConstOrderCustomInfoLinkKey = "trustpilot_link"
	ConstOrderCustomInfoSentKey = "trustpilot_sent"

	ConstConfigPathTrustPilot                 = "general.trustpilot"
	ConstConfigPathTrustPilotEnabled          = "general.trustpilot.enabled"
	ConstConfigPathTrustPilotTestMode         = "general.trustpilot.test"
	ConstConfigPathTrustPilotAPIKey           = "general.trustpilot.apiKey"
	ConstConfigPathTrustPilotAPISecret        = "general.trustpilot.apiSecret"
	ConstConfigPathTrustPilotBusinessUnitID   = "general.trustpilot.businessUnitID"
	ConstConfigPathTrustPilotUsername         = "general.trustpilot.username"
	ConstConfigPathTrustPilotPassword         = "general.trustpilot.password"
	ConstConfigPathTrustPilotAccessTokenURL   = "general.trustpilot.accessTokenURL"
	ConstConfigPathTrustPilotProductReviewURL = "general.trustpilot.productReviewURL"
	ConstConfigPathTrustPilotServiceReviewURL = "general.trustpilot.serviceReviewURL"
	ConstConfigPathTrustPilotEmailTemplate    = "general.trustpilot.emailTemplate"
)
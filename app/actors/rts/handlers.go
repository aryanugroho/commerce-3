package rts

import (
	"strings"
	"time"

	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/utils"
)

func referrerHandler(event string, data map[string]interface{}) bool {

	if "api.referrer" != event || "" == utils.InterfaceToString(data["referrer"]) {
		return true
	}

	referrer, err := GetReferrer(utils.InterfaceToString(data["referrer"]))
	if err != nil {
		return true
	}

	referrers[referrer]++

	return true
}

func visitsHandler(event string, data map[string]interface{}) bool {

	if "api.visits" != event {
		return true
	}

	err := GetTodayVisitorsData()
	if err != nil {
		return true
	}
	sessionID := utils.InterfaceToString(data["sessionID"])

	year := time.Now().Year()
	month := time.Now().Month()
	day := time.Now().Day()
	hour := time.Now().Hour()
	today := time.Date(year, month, day, hour, 0, 0, 0, time.Local)

	if _, ok := visitorsInfoToday.Details[sessionID]; !ok {
		visitorsInfoToday.Details[sessionID] = &VisitorDetail{Time: today}
		visitorsInfoToday.Visitors++
	}

	visitorsInfoToday.Details[sessionID] = &VisitorDetail{Time: today}
	_ = SaveVisitorData()

	return true
}

func addToCartHandler(event string, data map[string]interface{}) bool {

	if "api.addToCart" != event {
		return true
	}

	err := GetTodayVisitorsData()
	if err != nil {
		return true
	}
	sessionID := utils.InterfaceToString(data["sessionID"])

	if 0 == visitorsInfoToday.Details[sessionID].Checkout {
		visitorsInfoToday.Details[sessionID].Checkout = VisitorAddToCart
		visitorsInfoToday.Cart++
	}

	_ = SaveVisitorData()

	return true
}

func reachedCheckoutHandler(event string, data map[string]interface{}) bool {

	if "api.reachedCheckout" != event {
		return true
	}

	err := GetTodayVisitorsData()
	if err != nil {
		return true
	}
	sessionID := utils.InterfaceToString(data["sessionID"])

	if VisitorCheckout > visitorsInfoToday.Details[sessionID].Checkout {
		visitorsInfoToday.Details[sessionID].Checkout = VisitorCheckout
		visitorsInfoToday.Checkout++
	}

	_ = SaveVisitorData()

	return true
}

func purchasedHandler(event string, data map[string]interface{}) bool {

	if "api.purchased" != event {
		return true
	}

	err := GetTodayVisitorsData()
	if err != nil {
		return true
	}
	sessionID := utils.InterfaceToString(data["sessionID"])

	if VisitorSales > visitorsInfoToday.Details[sessionID].Checkout {
		visitorsInfoToday.Details[sessionID].Checkout = VisitorSales
		visitorsInfoToday.Sales++
	}

	_ = SaveVisitorData()

	return true
}

func salesHandler(event string, data map[string]interface{}) bool {

	if "api.sales" != event || len(data) == 0 {
		return true
	}
	salesData := make(map[string]int)

	salesHistoryCollection, err := db.GetCollection(CollectionNameSalesHistory)
	if err != nil {
		return true
	}

	for productID, count := range data {
		year := time.Now().Year()
		month := time.Now().Month()
		day := time.Now().Day()
		date := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		salesHistoryRow := make(map[string]interface{})
		salesData[productID] = utils.InterfaceToInt(count)

		salesHistoryCollection.ClearFilters()
		salesHistoryCollection.AddFilter("created_at", "=", date)
		salesHistoryCollection.AddFilter("product_id", "=", productID)
		dbSaleRow, _ := salesHistoryCollection.Load()

		newCount := utils.InterfaceToInt(count)
		if len(dbSaleRow) > 0 {
			salesHistoryRow["_id"] = utils.InterfaceToString(dbSaleRow[0]["_id"])
			oldCount := utils.InterfaceToInt(dbSaleRow[0]["count"])
			newCount += oldCount
		}

		// Add history row
		salesHistoryRow["product_id"] = productID
		salesHistoryRow["created_at"] = date
		salesHistoryRow["count"] = newCount
		_, err = salesHistoryCollection.Save(salesHistoryRow)
		if err != nil {
			return true
		}
	}

	SaveSalesData(salesData)

	return true
}

func regVisitorAsOnlineHandler(event string, data map[string]interface{}) bool {
	if "api.regVisitorAsOnlineHandler" != event {
		return true
	}
	sessionID := utils.InterfaceToString(data["sessionID"])

	referrerType := ReferrerTypeDirect

	if "" != utils.InterfaceToString(data["referrer"]) {
		referrer, err := GetReferrer(utils.InterfaceToString(data["referrer"]))
		if err != nil {
			return true
		}

		isSearchEngine := false
		for index := 0; index < len(searchEngines); index++ {
			if strings.Contains(referrer, searchEngines[index]) {
				isSearchEngine = true
			}
		}

		if isSearchEngine {
			referrerType = ReferrerTypeSearch
		} else {
			referrerType = ReferrerTypeSite
		}
	}

	if _, ok := OnlineSessions[sessionID]; !ok {
		OnlineSessions[sessionID] = &OnlineReferrer{}
		IncreaseOnline(referrerType)
		if len(OnlineSessions) > OnlineSessionsMax {
			OnlineSessionsMax = len(OnlineSessions)
		}
	} else {
		if OnlineSessions[sessionID].referrerType != referrerType {
			DecreaseOnline(OnlineSessions[sessionID].referrerType)
			IncreaseOnline(referrerType)
		}
	}

	OnlineSessions[sessionID].time = time.Now()
	OnlineSessions[sessionID].referrerType = referrerType

	return true
}

func visitorOnlineActionHandler(event string, data map[string]interface{}) bool {
	if "api.visitorOnlineAction" != event {
		return true
	}
	sessionID := utils.InterfaceToString(data["sessionID"])
	if _, ok := OnlineSessions[sessionID]; ok {
		OnlineSessions[sessionID].time = time.Now()
	}

	return true
}

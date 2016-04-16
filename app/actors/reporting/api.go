package reporting

import (
	"sort"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"

	"github.com/ottemo/foundation/app/actors/payment/authorizenet"
	"github.com/ottemo/foundation/app/actors/payment/checkmo"
	"github.com/ottemo/foundation/app/actors/payment/paypal"
	"github.com/ottemo/foundation/app/actors/payment/zeropay"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/order"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	// Admin only endpoint
	service := api.GetRestService()
	service.GET("reporting/product-performance", api.IsAdmin(listProductPerformance))
	service.GET("reporting/customer-activity", api.IsAdmin(listCustomerActivity))
	service.GET("reporting/payment-method", api.IsAdmin(listPaymentMethod))

	return nil
}

// listProductPerformance Handler that returns product performance information by date range
func listProductPerformance(context api.InterfaceApplicationContext) (interface{}, error) {

	// Expecting dates in UTC, and adjusted for your timezone
	// `2006-01-02 15:04`
	startDate := utils.InterfaceToTime(context.GetRequestArgument("start_date"))
	endDate := utils.InterfaceToTime(context.GetRequestArgument("end_date"))
	if startDate.IsZero() || endDate.IsZero() {
		context.SetResponseStatusBadRequest()
		msg := "start_date or end_date missing from response, or not formatted in YYYY-MM-DD"
		return nil, env.ErrorNew("reporting", 6, "3ed77c0d-2c54-4401-9feb-6e1d04b8baef", msg)
	}
	if startDate.After(endDate) || startDate.Equal(endDate) {
		context.SetResponseStatusBadRequest()
		msg := "the start_date must come before the end_date"
		return nil, env.ErrorNew("reporting", 6, "2eb9680c-d9a8-42ce-af63-fd6b0b742d0d", msg)
	}

	foundOrders := order.GetOrdersCreatedBetween(startDate, endDate)
	foundOrderIds := getOrderIds(foundOrders)
	foundOrderItems := order.GetItemsForOrders(foundOrderIds)
	aggregatedResults := aggregateOrderItems(foundOrderItems)

	totalSales := getTotalSales(foundOrderItems)

	response := map[string]interface{}{
		"total_orders":    len(foundOrders),
		"total_items":     len(foundOrderItems),
		"total_sales":     totalSales,
		"aggregate_items": aggregatedResults,
	}

	return response, nil
}

// getOrderIds Create a list of order ids
func getOrderIds(foundOrders []models.StructListItem) []string {
	var orderIds []string
	for _, foundOrder := range foundOrders {
		orderIds = append(orderIds, foundOrder.ID)
	}
	return orderIds
}

// aggregateOrderItems Takes a list of order ids and aggregates their price / qty by their sku
func aggregateOrderItems(oitems []map[string]interface{}) []ProductPerfItem {
	keyedResults := make(map[string]ProductPerfItem)

	// Aggregate by sku
	for _, oitem := range oitems {
		sku := utils.InterfaceToString(oitem["sku"])
		item, ok := keyedResults[sku]

		// First time, set the static details
		if !ok {
			item.Name = utils.InterfaceToString(oitem["name"])
			item.Sku = sku
		}

		item.GrossSales += utils.InterfaceToFloat64(oitem["price"])
		item.UnitsSold += utils.InterfaceToInt(oitem["qty"])

		keyedResults[sku] = item
	}

	// map to slice
	var results ProductPerf
	for _, item := range keyedResults {
		// @TODO: Round money is bad
		item.GrossSales = utils.RoundPrice(item.GrossSales)
		results = append(results, item)
	}

	sort.Sort(results)

	return results
}

func getTotalSales(oitems []map[string]interface{}) float64 {
	var totalSales float64
	for _, oitem := range oitems {
		totalSales += utils.InterfaceToFloat64(oitem["price"])
	}

	return totalSales
}

func listCustomerActivity(context api.InterfaceApplicationContext) (interface{}, error) {
	perfStart := time.Now()

	// Limit results count, not the query
	limit := utils.InterfaceToInt(context.GetRequestArgument("limit"))
	if limit == 0 {
		limit = 50
	}

	sortArg := utils.InterfaceToString(context.GetRequestArgument("sort"))

	// Expecting dates in UTC, and adjusted for your timezone `2006-01-02 15:04`
	startDate := utils.InterfaceToTime(context.GetRequestArgument("start_date"))
	endDate := utils.InterfaceToTime(context.GetRequestArgument("end_date"))
	hasDateRange := !startDate.IsZero() || !endDate.IsZero()

	// Date range validation
	if hasDateRange {
		if startDate.IsZero() || endDate.IsZero() {
			context.SetResponseStatusBadRequest()
			msg := "start_date or end_date missing from response, or not formatted in YYYY-MM-DD"
			return nil, env.ErrorNew("reporting", 6, "3ed77c0d-2c54-4401-9feb-6e1d04b8baef", msg)
		}
		if startDate.After(endDate) || startDate.Equal(endDate) {
			context.SetResponseStatusBadRequest()
			msg := "the start_date must come before the end_date"
			return nil, env.ErrorNew("reporting", 6, "2eb9680c-d9a8-42ce-af63-fd6b0b742d0d", msg)
		}
	}

	// Fetch orders
	oModel, _ := order.GetOrderCollectionModel()
	oModel.ListAddExtraAttribute("created_at")
	oModel.ListAddExtraAttribute("customer_name")
	oModel.ListAddExtraAttribute("customer_email")
	oModel.ListAddExtraAttribute("grand_total")
	if hasDateRange {
		oModel.GetDBCollection().AddFilter("created_at", ">=", startDate)
		oModel.GetDBCollection().AddFilter("created_at", "<", endDate)
	}

	// This is the lite response StructListItem
	foundOrders, _ := oModel.List()
	aggregatedResults := aggregateCustomerActivity(foundOrders)
	resultCount := len(aggregatedResults)

	// Sorting
	switch sortArg {
	case "total_orders":
		sort.Sort(CustomerActivityByOrders(aggregatedResults))
	case "total_sales":
		fallthrough
	default:
		sort.Sort(CustomerActivityBySales(aggregatedResults))
	}

	// Apply the limit
	if resultCount > limit {
		aggregatedResults = aggregatedResults[:limit]
	}

	response := map[string]interface{}{
		"aggregate_items": aggregatedResults,
		"meta": map[string]interface{}{
			"limit": limit,
			"count": resultCount,
		},
		"perf_ms": time.Now().Sub(perfStart).Seconds() * 1e3, // in milliseconds
	}
	return response, nil
}

func aggregateCustomerActivity(foundOrders []models.StructListItem) []CustomerActivityItem {
	keyedResults := make(map[string]CustomerActivityItem)

	for _, o := range foundOrders {
		email := utils.InterfaceToString(o.Extra["customer_email"])
		createdAt := utils.InterfaceToTime(o.Extra["created_at"])
		item, ok := keyedResults[email]

		// First time, set some static props
		if !ok {
			item.Email = email
			item.EarliestPurchase = createdAt
			item.LatestPurchase = createdAt
		} else {
			if createdAt.Before(item.EarliestPurchase) {
				item.EarliestPurchase = createdAt
			}
			if createdAt.After(item.LatestPurchase) {
				item.LatestPurchase = createdAt
			}
		}

		// Name might be empty, early records had a bug
		if item.Name == "" {
			item.Name = utils.InterfaceToString(o.Extra["customer_name"])
		}

		// Aggregated props
		item.TotalSales += utils.InterfaceToFloat64(o.Extra["grand_total"])
		item.TotalOrders++

		// Save
		keyedResults[email] = item
	}

	// map to slice
	var results []CustomerActivityItem
	for _, i := range keyedResults {
		// Add in averaging stat now that aggregation is complete
		i.AverageSales = i.TotalSales / float64(i.TotalOrders)

		// Round money
		i.TotalSales = utils.RoundPrice(i.TotalSales)
		i.AverageSales = utils.RoundPrice(i.AverageSales)

		results = append(results, i)
	}

	return results
}

func listPaymentMethod(context api.InterfaceApplicationContext) (interface{}, error) {
	perfStart := time.Now()

	// Expecting dates in UTC, and adjusted for your timezone `2006-01-02 15:04`
	startDate := utils.InterfaceToTime(context.GetRequestArgument("start_date"))
	endDate := utils.InterfaceToTime(context.GetRequestArgument("end_date"))
	hasDateRange := !startDate.IsZero() || !endDate.IsZero()

	// Date range validation
	if hasDateRange {
		if startDate.IsZero() || endDate.IsZero() {
			context.SetResponseStatusBadRequest()
			msg := "start_date or end_date missing from response, or not formatted in YYYY-MM-DD"
			return nil, env.ErrorNew("reporting", 6, "3ed77c0d-2c54-4401-9feb-6e1d04b8baef", msg)
		}
		if startDate.After(endDate) || startDate.Equal(endDate) {
			context.SetResponseStatusBadRequest()
			msg := "the start_date must come before the end_date"
			return nil, env.ErrorNew("reporting", 6, "2eb9680c-d9a8-42ce-af63-fd6b0b742d0d", msg)
		}
	}

	// Fetch orders
	oModel, _ := order.GetOrderCollectionModel()
	oModel.ListAddExtraAttribute("created_at")
	oModel.ListAddExtraAttribute("payment_method")
	oModel.ListAddExtraAttribute("grand_total")
	if hasDateRange {
		oModel.GetDBCollection().AddFilter("created_at", ">=", startDate)
		oModel.GetDBCollection().AddFilter("created_at", "<", endDate)
	}

	// This is the lite response StructListItem
	foundOrders, _ := oModel.List()
	aggregatedResults := aggregatePaymentMethod(foundOrders)

	// Sorting
	sort.Sort(PaymentMethodBySales(aggregatedResults))

	response := map[string]interface{}{
		"aggregate_items": aggregatedResults,
		"perf_ms":         time.Now().Sub(perfStart).Seconds() * 1e3, // in milliseconds
	}
	return response, nil
}

func aggregatePaymentMethod(foundOrders []models.StructListItem) []PaymentMethodItem {

	// Map of payment method key: name
	paymentMethodNames := map[string]string{
		authorizenet.ConstPaymentCodeDPM:     authorizenet.ConstPaymentNameDPM,
		checkmo.ConstPaymentCode:             checkmo.ConstPaymentName,
		paypal.ConstPaymentCode:              paypal.ConstPaymentName,
		paypal.ConstPaymentPayPalPayflowCode: paypal.ConstPaymentPayPalPayflowName,
		zeropay.ConstPaymentZeroPaymentCode:  zeropay.ConstPaymentName,
	}

	keyedResults := make(map[string]PaymentMethodItem)

	for _, o := range foundOrders {
		key := utils.InterfaceToString(o.Extra["payment_method"])
		item, ok := keyedResults[key]

		// First time, set some static props
		if !ok {
			item.Key = key
			item.Name = paymentMethodNames[key]
		}

		// Aggregated props
		item.TotalSales += utils.InterfaceToFloat64(o.Extra["grand_total"])
		item.TotalOrders++

		// Save
		keyedResults[key] = item
	}

	// map to slice
	var results []PaymentMethodItem
	for _, i := range keyedResults {
		// Add in averaging stat now that aggregation is complete
		i.AverageSales = i.TotalSales / float64(i.TotalOrders)

		// Round money
		i.TotalSales = utils.RoundPrice(i.TotalSales)
		i.AverageSales = utils.RoundPrice(i.AverageSales)

		results = append(results, i)
	}

	return results
}
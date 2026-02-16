// Package binance implements the Binance exchange driver.
// API Documentation: https://binance-docs.github.io/apidocs/spot/en/
// API Version: v3 (verified 2026-02-16)
package binance

// Binance API base URLs
const (
	// BaseRestURL is the production REST API base URL
	BaseRestURL = "https://api.binance.com"
	// BaseWebSocketURL is the production WebSocket API base URL
	BaseWebSocketURL = "wss://stream.binance.com:9443/ws"
	// BaseWebSocketCombinedURL is the combined stream WebSocket URL
	BaseWebSocketCombinedURL = "wss://stream.binance.com:9443/stream"
	// TestnetRestURL is the testnet REST API base URL
	TestnetRestURL = "https://testnet.binance.vision"
	// TestnetWebSocketURL is the testnet WebSocket API base URL
	TestnetWebSocketURL = "wss://testnet.binance.vision/ws"
	// TestnetWebSocketCombinedURL is the testnet combined stream WebSocket URL
	TestnetWebSocketCombinedURL = "wss://testnet.binance.vision/stream"
)

// Binance API v3 endpoints
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#general-endpoints
const (
	// General endpoints
	EPing         = "/api/v3/ping"
	ETime         = "/api/v3/time"
	EExchangeInfo = "/api/v3/exchangeInfo"
	EAccount      = "/api/v3/account"

	// Market data endpoints
	EDepth             = "/api/v3/depth"
	ETrades            = "/api/v3/trades"
	EHistTrades        = "/api/v3/historicalTrades"
	ERecentTrades      = "/api/v3/trades"
	ETicker            = "/api/v3/ticker/24hr"
	ETickerPrice       = "/api/v3/ticker/price"
	ETickerBook        = "/api/v3/ticker/bookTicker"
	ESymbolPriceTicker = "/api/v3/ticker/price"
	EAllBookTickers    = "/api/v3/ticker/bookTicker"
	EOpenOrders        = "/api/v3/openOrders"
	EAllOrders         = "/api/v3/allOrders"

	// Order endpoints
	ENewOrder            = "/api/v3/order"
	EQueryOrder          = "/api/v3/order"
	ECancelOrder         = "/api/v3/order"
	ECancelAllOpenOrders = "/api/v3/openOrders"
	ENewOCO              = "/api/v3/order/oco"
	EQueryOCO            = "/api/v3/orderList"
	ECancelOCO           = "/api/v3/orderList"
	EQueryOpenOCO        = "/api/v3/openOrderList"
	EQueryAllOCO         = "/api/v3/allOrderList"

	// User data stream
	EUserDataStream = "/api/v3/userDataStream"
	EListenKey      = "/api/v3/userDataStream"

	// Symbol information
	ESymbolInfo = "/api/v3/exchangeInfo"
)

// Endpoint weights for Binance API
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#limits
// Last verified: 2026-02-16
var EndpointWeights = map[string]int{
	// General endpoints
	"/api/v3/ping":         1,
	"/api/v3/time":         1,
	"/api/v3/exchangeInfo": 20,
	"/api/v3/account":      20,

	// Market data endpoints
	"/api/v3/depth":             5, // default, varies by limit
	"/api/v3/trades":            1,
	"/api/v3/historicalTrades":  5,
	"/api/v3/ticker/24hr":       2, // per symbol
	"/api/v3/ticker/price":      1, // per symbol
	"/api/v3/ticker/bookTicker": 2, // per symbol

	// Order endpoints
	"/api/v3/order":      1, // POST (new order) = 1, GET (query) = 2, DELETE (cancel) = 1
	"/api/v3/openOrders": 3, // per symbol, 6 if no symbol
	"/api/v3/allOrders":  10,

	// User data stream
	"/api/v3/userDataStream": 1,
}

// GetEndpointWeight returns the weight for a given endpoint.
// For endpoints that vary by method (like /api/v3/order), returns the base weight.
// Returns 1 as default for unknown endpoints (safe default).
func GetEndpointWeight(endpoint string) int {
	if weight, ok := EndpointWeights[endpoint]; ok {
		return weight
	}
	// Default weight of 1 for unknown endpoints
	return 1
}

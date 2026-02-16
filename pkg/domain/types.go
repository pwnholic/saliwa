// Package domain provides core domain types for the exchange connector.
package domain

import (
	"fmt"
	"time"
)

// OrderSide represents the direction of a trade.
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"  // Buying base currency
	OrderSideSell OrderSide = "SELL" // Selling base currency
)

// IsValid returns true if the order side is valid.
func (s OrderSide) IsValid() bool {
	return s == OrderSideBuy || s == OrderSideSell
}

// OrderType represents the type of order.
type OrderType string

const (
	OrderTypeLimit  OrderType = "LIMIT"  // Limit order with specified price
	OrderTypeMarket OrderType = "MARKET" // Market order at best available price
)

// IsValid returns true if the order type is valid.
func (t OrderType) IsValid() bool {
	return t == OrderTypeLimit || t == OrderTypeMarket
}

// OrderStatus represents the current state of an order.
type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"              // Order is new and pending
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED" // Order is partially filled
	OrderStatusFilled          OrderStatus = "FILLED"           // Order is fully filled
	OrderStatusCanceled        OrderStatus = "CANCELED"         // Order is canceled
	OrderStatusCanceling       OrderStatus = "CANCELING"        // Order cancel request is pending
	OrderStatusRejected        OrderStatus = "REJECTED"         // Order is rejected
	OrderStatusExpired         OrderStatus = "EXPIRED"          // Order has expired
)

// IsValid returns true if the order status is valid.
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusNew, OrderStatusPartiallyFilled, OrderStatusFilled,
		OrderStatusCanceled, OrderStatusCanceling, OrderStatusRejected, OrderStatusExpired:
		return true
	default:
		return false
	}
}

// IsFinal returns true if the order is in a final state (no more updates expected).
func (s OrderStatus) IsFinal() bool {
	return s == OrderStatusFilled ||
		s == OrderStatusCanceled ||
		s == OrderStatusRejected ||
		s == OrderStatusExpired
}

// CanTransition returns true if the order can transition to the new status.
// This implements a simple state machine for order status transitions.
func (s OrderStatus) CanTransition(newStatus OrderStatus) bool {
	// Define valid transitions
	validTransitions := map[OrderStatus][]OrderStatus{
		OrderStatusNew: {
			OrderStatusPartiallyFilled,
			OrderStatusFilled,
			OrderStatusCanceling,
			OrderStatusCanceled,
			OrderStatusRejected,
			OrderStatusExpired,
		},
		OrderStatusPartiallyFilled: {
			OrderStatusPartiallyFilled,
			OrderStatusFilled,
			OrderStatusCanceling,
			OrderStatusCanceled,
		},
		OrderStatusCanceling: {
			OrderStatusPartiallyFilled,
			OrderStatusFilled,
			OrderStatusCanceled,
		},
		// Final states cannot transition
		OrderStatusFilled:   {},
		OrderStatusCanceled: {},
		OrderStatusRejected: {},
		OrderStatusExpired:  {},
	}

	allowed, exists := validTransitions[s]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}
	return false
}

// Order represents a trading order on an exchange.
// All fields are immutable after creation.
type Order struct {
	// Exchange is the name of the exchange (e.g., "binance", "bybit")
	Exchange string `json:"exchange"`

	// Symbol is the trading pair in normalized format (e.g., "BTC/USDT")
	Symbol string `json:"symbol"`

	// ID is the exchange-assigned order ID
	ID string `json:"id"`

	// ClientOrderID is the client-assigned order ID (optional)
	ClientOrderID string `json:"client_order_id,omitempty"`

	// Side is the order direction (BUY or SELL)
	Side OrderSide `json:"side"`

	// Type is the order type (LIMIT or MARKET)
	Type OrderType `json:"type"`

	// Status is the current order status
	Status OrderStatus `json:"status"`

	// Price is the limit price (for LIMIT orders)
	Price Decimal `json:"price,omitempty"`

	// Quantity is the original order quantity
	Quantity Decimal `json:"quantity"`

	// FilledQuantity is the quantity that has been filled
	FilledQuantity Decimal `json:"filled_quantity"`

	// QuoteQuantity is the total quote currency quantity (for filled amount)
	QuoteQuantity Decimal `json:"quote_quantity,omitempty"`

	// Commission is the fee paid for this order
	Commission Decimal `json:"commission,omitempty"`

	// CommissionAsset is the asset in which commission was paid
	CommissionAsset string `json:"commission_asset,omitempty"`

	// CreatedAt is the order creation time
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the last update time
	UpdatedAt time.Time `json:"updated_at"`

	// TradeID is the ID of the last trade that filled this order
	TradeID string `json:"trade_id,omitempty"`

	// IsWorking indicates if the order is on the order book
	IsWorking bool `json:"is_working"`
}

// IsFilled returns true if the order is fully filled.
func (o *Order) IsFilled() bool {
	return o.Status == OrderStatusFilled
}

// IsOpen returns true if the order is still active (not in final state).
func (o *Order) IsOpen() bool {
	return !o.Status.IsFinal()
}

// RemainingQuantity returns the remaining unfilled quantity.
func (o *Order) RemainingQuantity() Decimal {
	return Sub(o.Quantity, o.FilledQuantity)
}

// FillPercentage returns the percentage of the order that has been filled.
func (o *Order) FillPercentage() Decimal {
	if IsZero(o.Quantity) {
		return Zero()
	}
	return Mul(Div(o.FilledQuantity, o.Quantity), NewDecimalFromInt(100))
}

// Validate validates the order fields.
func (o *Order) Validate() error {
	if o.Exchange == "" {
		return fmt.Errorf("exchange is required")
	}
	if o.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if o.ID == "" {
		return fmt.Errorf("order ID is required")
	}
	if !o.Side.IsValid() {
		return fmt.Errorf("invalid order side: %s", o.Side)
	}
	if !o.Type.IsValid() {
		return fmt.Errorf("invalid order type: %s", o.Type)
	}
	if !o.Status.IsValid() {
		return fmt.Errorf("invalid order status: %s", o.Status)
	}
	if IsNegative(o.Quantity) {
		return fmt.Errorf("quantity cannot be negative")
	}
	if IsNegative(o.FilledQuantity) {
		return fmt.Errorf("filled quantity cannot be negative")
	}
	if Cmp(o.FilledQuantity, o.Quantity) > 0 {
		return fmt.Errorf("filled quantity cannot exceed order quantity")
	}
	return nil
}

// OrderRequest represents a request to place a new order.
type OrderRequest struct {
	// Exchange is the target exchange
	Exchange string `json:"exchange"`

	// Symbol is the trading pair
	Symbol string `json:"symbol"`

	// Side is the order direction
	Side OrderSide `json:"side"`

	// Type is the order type
	Type OrderType `json:"type"`

	// Price is the limit price (required for LIMIT orders)
	Price Decimal `json:"price,omitempty"`

	// Quantity is the order quantity in base currency
	Quantity Decimal `json:"quantity"`

	// QuoteQuantity is the order quantity in quote currency (for market orders)
	// If set, Quantity should be zero
	QuoteQuantity Decimal `json:"quote_quantity,omitempty"`

	// ClientOrderID is a client-assigned ID (optional)
	ClientOrderID string `json:"client_order_id,omitempty"`

	// TimeInForce is the order time in force (GTC, IOC, FOK)
	TimeInForce string `json:"time_in_force,omitempty"`

	// StopPrice is the stop price for stop orders
	StopPrice Decimal `json:"stop_price,omitempty"`

	// icebergQty is the iceberg quantity for iceberg orders
	IcebergQuantity Decimal `json:"iceberg_quantity,omitempty"`
}

// Validate validates the order request fields.
func (r *OrderRequest) Validate() error {
	if r.Exchange == "" {
		return fmt.Errorf("exchange is required")
	}
	if r.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if !r.Side.IsValid() {
		return fmt.Errorf("invalid order side: %s", r.Side)
	}
	if !r.Type.IsValid() {
		return fmt.Errorf("invalid order type: %s", r.Type)
	}

	// Validate quantity
	if IsZero(r.Quantity) && IsZero(r.QuoteQuantity) {
		return fmt.Errorf("either quantity or quote_quantity is required")
	}
	if IsNegative(r.Quantity) {
		return fmt.Errorf("quantity cannot be negative")
	}
	if IsNegative(r.QuoteQuantity) {
		return fmt.Errorf("quote_quantity cannot be negative")
	}

	// Validate price for limit orders
	if r.Type == OrderTypeLimit {
		if IsZero(r.Price) || IsNegative(r.Price) {
			return fmt.Errorf("price is required and must be positive for limit orders")
		}
	}

	return nil
}

// CancelRequest represents a request to cancel an order.
type CancelRequest struct {
	// Exchange is the target exchange
	Exchange string `json:"exchange"`

	// Symbol is the trading pair
	Symbol string `json:"symbol"`

	// OrderID is the exchange-assigned order ID
	OrderID string `json:"order_id,omitempty"`

	// ClientOrderID is the client-assigned order ID
	ClientOrderID string `json:"client_order_id,omitempty"`
}

// Validate validates the cancel request fields.
func (r *CancelRequest) Validate() error {
	if r.Exchange == "" {
		return fmt.Errorf("exchange is required")
	}
	if r.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if r.OrderID == "" && r.ClientOrderID == "" {
		return fmt.Errorf("either order_id or client_order_id is required")
	}
	return nil
}

// Ticker represents current market ticker data.
type Ticker struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// Symbol is the trading pair in normalized format
	Symbol string `json:"symbol"`

	// BidPrice is the best bid price
	BidPrice Decimal `json:"bid_price"`

	// BidQuantity is the best bid quantity
	BidQuantity Decimal `json:"bid_quantity"`

	// AskPrice is the best ask price
	AskPrice Decimal `json:"ask_price"`

	// AskQuantity is the best ask quantity
	AskQuantity Decimal `json:"ask_quantity"`

	// LastPrice is the last traded price
	LastPrice Decimal `json:"last_price"`

	// HighPrice is the 24h high price
	HighPrice Decimal `json:"high_price"`

	// LowPrice is the 24h low price
	LowPrice Decimal `json:"low_price"`

	// Volume is the 24h volume in base currency
	Volume Decimal `json:"volume"`

	// QuoteVolume is the 24h volume in quote currency
	QuoteVolume Decimal `json:"quote_volume"`

	// PriceChange is the 24h price change
	PriceChange Decimal `json:"price_change"`

	// PriceChangePercent is the 24h price change percentage
	PriceChangePercent Decimal `json:"price_change_percent"`

	// OpenPrice is the 24h open price
	OpenPrice Decimal `json:"open_price"`

	// Timestamp is the ticker update time
	Timestamp time.Time `json:"timestamp"`
}

// Spread returns the bid-ask spread.
func (t *Ticker) Spread() Decimal {
	return Sub(t.AskPrice, t.BidPrice)
}

// SpreadPercent returns the bid-ask spread as a percentage of mid-price.
func (t *Ticker) SpreadPercent() Decimal {
	midPrice := Div(Add(t.BidPrice, t.AskPrice), NewDecimalFromInt(2))
	if IsZero(midPrice) {
		return Zero()
	}
	return Mul(Div(t.Spread(), midPrice), NewDecimalFromInt(100))
}

// MidPrice returns the mid-price between bid and ask.
func (t *Ticker) MidPrice() Decimal {
	return Div(Add(t.BidPrice, t.AskPrice), NewDecimalFromInt(2))
}

// Trade represents a single trade execution.
type Trade struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// Symbol is the trading pair in normalized format
	Symbol string `json:"symbol"`

	// ID is the trade ID
	ID string `json:"id"`

	// OrderID is the associated order ID
	OrderID string `json:"order_id"`

	// Price is the trade price
	Price Decimal `json:"price"`

	// Quantity is the trade quantity
	Quantity Decimal `json:"quantity"`

	// QuoteQuantity is the trade quantity in quote currency
	QuoteQuantity Decimal `json:"quote_quantity"`

	// Commission is the fee for this trade
	Commission Decimal `json:"commission,omitempty"`

	// CommissionAsset is the asset in which commission was paid
	CommissionAsset string `json:"commission_asset,omitempty"`

	// Side is the trade direction (BUY or SELL)
	Side OrderSide `json:"side"`

	// IsMaker indicates if this trade was a maker order
	IsMaker bool `json:"is_maker"`

	// Timestamp is the trade execution time
	Timestamp time.Time `json:"timestamp"`
}

// OrderBookLevel represents a single price level in the order book.
type OrderBookLevel struct {
	// Price is the price level
	Price Decimal `json:"price"`

	// Quantity is the quantity at this price level
	Quantity Decimal `json:"quantity"`
}

// OrderBook represents a snapshot of the order book.
type OrderBook struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// Symbol is the trading pair in normalized format
	Symbol string `json:"symbol"`

	// Bids are the buy orders sorted by price descending
	Bids []OrderBookLevel `json:"bids"`

	// Asks are the sell orders sorted by price ascending
	Asks []OrderBookLevel `json:"asks"`

	// LastUpdateID is the last update ID (for synchronization)
	LastUpdateID int64 `json:"last_update_id"`

	// Timestamp is the order book update time
	Timestamp time.Time `json:"timestamp"`
}

// BestBid returns the best bid price and quantity, or nil if no bids.
func (ob *OrderBook) BestBid() *OrderBookLevel {
	if len(ob.Bids) == 0 {
		return nil
	}
	return &ob.Bids[0]
}

// BestAsk returns the best ask price and quantity, or nil if no asks.
func (ob *OrderBook) BestAsk() *OrderBookLevel {
	if len(ob.Asks) == 0 {
		return nil
	}
	return &ob.Asks[0]
}

// Spread returns the bid-ask spread.
func (ob *OrderBook) Spread() Decimal {
	bestBid := ob.BestBid()
	bestAsk := ob.BestAsk()
	if bestBid == nil || bestAsk == nil {
		return nil
	}
	return Sub(bestAsk.Price, bestBid.Price)
}

// MidPrice returns the mid-price between best bid and ask.
func (ob *OrderBook) MidPrice() Decimal {
	bestBid := ob.BestBid()
	bestAsk := ob.BestAsk()
	if bestBid == nil || bestAsk == nil {
		return nil
	}
	return Div(Add(bestBid.Price, bestAsk.Price), NewDecimalFromInt(2))
}

// Balance represents an account balance for a specific asset.
type Balance struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// Asset is the asset symbol (e.g., "BTC", "USDT")
	Asset string `json:"asset"`

	// Free is the available balance for trading
	Free Decimal `json:"free"`

	// Locked is the balance locked in open orders
	Locked Decimal `json:"locked"`

	// Timestamp is the last update time
	Timestamp time.Time `json:"timestamp"`
}

// Total returns the total balance (free + locked).
func (b *Balance) Total() Decimal {
	return Add(b.Free, b.Locked)
}

// Kline represents a candlestick/kline data point.
type Kline struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// Symbol is the trading pair
	Symbol string `json:"symbol"`

	// Interval is the kline interval (e.g., "1m", "5m", "1h", "1d")
	Interval string `json:"interval"`

	// OpenTime is the opening time of this candle
	OpenTime time.Time `json:"open_time"`

	// CloseTime is the closing time of this candle
	CloseTime time.Time `json:"close_time"`

	// Open is the opening price
	Open Decimal `json:"open"`

	// High is the highest price
	High Decimal `json:"high"`

	// Low is the lowest price
	Low Decimal `json:"low"`

	// Close is the closing price
	Close Decimal `json:"close"`

	// Volume is the trading volume in base currency
	Volume Decimal `json:"volume"`

	// QuoteVolume is the trading volume in quote currency
	QuoteVolume Decimal `json:"quote_volume"`

	// TradeCount is the number of trades
	TradeCount int64 `json:"trade_count"`

	// TakerBuyVolume is the taker buy volume
	TakerBuyVolume Decimal `json:"taker_buy_volume"`

	// TakerBuyQuoteVolume is the taker buy quote volume
	TakerBuyQuoteVolume Decimal `json:"taker_buy_quote_volume"`

	// IsClosed indicates if this candle is closed
	IsClosed bool `json:"is_closed"`
}

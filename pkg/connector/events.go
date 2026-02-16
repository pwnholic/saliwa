package connector

import (
	"github.com/lilwiggy/ex-act/pkg/domain"
)

// EventType represents the type of event.
type EventType string

const (
	EventTicker    EventType = "ticker"
	EventOrderBook EventType = "orderbook"
	EventTrade     EventType = "trade"
	EventOrder     EventType = "order"
	EventBalance   EventType = "balance"
)

// Event represents an event from the exchange.
type Event struct {
	Exchange string    // Exchange name
	Type     EventType // Event type
	Data     any       // Event data (domain types)
}

// TickerHandler handles ticker events.
type TickerHandler func(exchange string, ticker *domain.Ticker)

// OrderBookHandler handles order book events.
type OrderBookHandler func(exchange string, orderbook *domain.OrderBook)

// TradeHandler handles trade events.
type TradeHandler func(exchange string, trade *domain.Trade)

// OrderHandler handles order update events.
type OrderHandler func(exchange string, order *domain.Order)

// ConnectionHandler handles connection state changes.
type ConnectionHandler func(exchange string, connected bool)

// ErrorHandler handles errors.
type ErrorHandler func(exchange string, err error)

// Handlers contains all event handlers.
type Handlers struct {
	OnTicker     TickerHandler
	OnOrderBook  OrderBookHandler
	OnTrade      TradeHandler
	OnOrder      OrderHandler
	OnConnect    ConnectionHandler
	OnDisconnect ConnectionHandler
	OnError      ErrorHandler
}

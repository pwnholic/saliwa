// Package binance implements the Binance exchange driver.
// Subscription management for Binance WebSocket streams.
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#websocket-market-streams
package binance

import (
	"strings"
	"sync"
)

// SubscriptionManager manages WebSocket stream subscriptions.
// It tracks active subscriptions for automatic resubscription on reconnect.
// CRITICAL: Must be thread-safe (sync.RWMutex) for concurrent access.
type SubscriptionManager struct {
	mu            sync.RWMutex
	subscriptions map[string]bool
}

// NewSubscriptionManager creates a new SubscriptionManager.
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subscriptions: make(map[string]bool),
	}
}

// Subscribe adds a stream to subscriptions.
// Returns true if this is a new subscription, false if already subscribed.
// CRITICAL: Stream names MUST be lowercase for Binance.
func (sm *SubscriptionManager) Subscribe(stream string) bool {
	// Normalize to lowercase
	stream = strings.ToLower(stream)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.subscriptions[stream] {
		return false // Already subscribed
	}
	sm.subscriptions[stream] = true
	return true
}

// Unsubscribe removes a stream from subscriptions.
// Returns true if the stream was subscribed, false otherwise.
func (sm *SubscriptionManager) Unsubscribe(stream string) bool {
	stream = strings.ToLower(stream)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.subscriptions[stream] {
		return false // Not subscribed
	}
	delete(sm.subscriptions, stream)
	return true
}

// IsSubscribed checks if a stream is subscribed.
func (sm *SubscriptionManager) IsSubscribed(stream string) bool {
	stream = strings.ToLower(stream)

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.subscriptions[stream]
}

// Streams returns all subscribed stream names.
// Used for automatic resubscription on reconnect.
func (sm *SubscriptionManager) Streams() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	streams := make([]string, 0, len(sm.subscriptions))
	for stream := range sm.subscriptions {
		streams = append(streams, stream)
	}
	return streams
}

// Count returns the number of active subscriptions.
func (sm *SubscriptionManager) Count() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.subscriptions)
}

// Clear removes all subscriptions.
func (sm *SubscriptionManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.subscriptions = make(map[string]bool)
}

// StreamBuilder creates Binance WebSocket stream names.
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#websocket-market-streams
// Stream names MUST be lowercase for Binance.
type StreamBuilder struct {
	symbol string // Exchange symbol format (e.g., "BTCUSDT")
}

// NewStreamBuilder creates a StreamBuilder for the given symbol.
// Symbol should be in exchange format (e.g., "BTCUSDT"), not normalized format.
func NewStreamBuilder(symbol string) *StreamBuilder {
	return &StreamBuilder{
		symbol: strings.ToLower(symbol),
	}
}

// Ticker creates a ticker stream name.
// Stream: <symbol>@ticker
// Updates every second with 24hr statistics.
func (sb *StreamBuilder) Ticker() string {
	return sb.symbol + "@ticker"
}

// Ticker1s creates a 1-second ticker stream name.
// Stream: <symbol>@ticker@1s
// Updates every second with 24hr statistics.
func (sb *StreamBuilder) Ticker1s() string {
	return sb.symbol + "@ticker@1s"
}

// MiniTicker creates a mini ticker stream name.
// Stream: <symbol>@miniTicker
// Lightweight ticker with essential fields only.
func (sb *StreamBuilder) MiniTicker() string {
	return sb.symbol + "@miniTicker"
}

// BookTicker creates a book ticker stream name.
// Stream: <symbol>@bookTicker
// Best bid/ask price and quantity updates.
func (sb *StreamBuilder) BookTicker() string {
	return sb.symbol + "@bookTicker"
}

// AllBookTickers creates an all book tickers stream name.
// Stream: !bookTicker
// Best bid/ask for all symbols.
func AllBookTickers() string {
	return "!bookTicker"
}

// Depth creates a depth stream name with 100ms updates.
// Stream: <symbol>@depth@100ms
// Order book updates with 100ms frequency.
func (sb *StreamBuilder) Depth() string {
	return sb.symbol + "@depth@100ms"
}

// Depth100ms creates a depth stream name with 100ms updates.
// Stream: <symbol>@depth@100ms
func (sb *StreamBuilder) Depth100ms() string {
	return sb.symbol + "@depth@100ms"
}

// Depth10 creates a partial depth stream (10 levels).
// Stream: <symbol>@depth10@100ms
func (sb *StreamBuilder) Depth10() string {
	return sb.symbol + "@depth10@100ms"
}

// Depth20 creates a partial depth stream (20 levels).
// Stream: <symbol>@depth20@100ms
func (sb *StreamBuilder) Depth20() string {
	return sb.symbol + "@depth20@100ms"
}

// Trade creates an individual trade stream name.
// Stream: <symbol>@trade
// Real-time trade updates.
func (sb *StreamBuilder) Trade() string {
	return sb.symbol + "@trade"
}

// AggTrade creates an aggregated trade stream name.
// Stream: <symbol>@aggTrade
// Aggregated trade updates (multiple trades combined).
func (sb *StreamBuilder) AggTrade() string {
	return sb.symbol + "@aggTrade"
}

// Kline creates a kline/candlestick stream name.
// Stream: <symbol>@kline_<interval>
// Valid intervals: 1s, 1m, 3m, 5m, 15m, 30m, 1h, 2h, 4h, 6h, 8h, 12h, 1d, 3d, 1w, 1M
func (sb *StreamBuilder) Kline(interval string) string {
	return sb.symbol + "@kline_" + strings.ToLower(interval)
}

// ForceOrder creates a liquidation order stream name.
// Stream: <symbol>@forceOrder
// Liquidation orders for all symbols or specific symbol.
func (sb *StreamBuilder) ForceOrder() string {
	return sb.symbol + "@forceOrder"
}

// UserData creates a user data stream path (not a combined stream).
// This is for the listen key, not for combined streams.
// Stream: <listenKey>
func UserData(listenKey string) string {
	return listenKey
}

// CombineStreams creates a combined stream URL path.
// Combined streams format: /stream?streams=stream1/stream2/...
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#combined-stream-exports
func CombineStreams(streams []string) string {
	if len(streams) == 0 {
		return ""
	}
	return strings.Join(streams, "/")
}

// SplitCombinedStream splits a combined stream name into individual streams.
func SplitCombinedStream(combined string) []string {
	if combined == "" {
		return nil
	}
	return strings.Split(combined, "/")
}

// ParseStreamSymbol extracts the symbol from a stream name.
// Returns empty string if stream format is unrecognized.
func ParseStreamSymbol(stream string) string {
	stream = strings.ToLower(stream)

	// Handle special streams
	if strings.HasPrefix(stream, "!") {
		return "" // All-symbols stream
	}

	// Standard format: <symbol>@<streamType>[@<speed>]
	idx := strings.Index(stream, "@")
	if idx <= 0 {
		return ""
	}
	return strings.ToUpper(stream[:idx])
}

// ParseStreamType extracts the stream type from a stream name.
// Returns empty string if stream format is unrecognized.
func ParseStreamType(stream string) string {
	stream = strings.ToLower(stream)

	// Handle special streams
	if stream == "!bookTicker" {
		return "bookTicker"
	}
	if stream == "!miniTicker" {
		return "miniTicker"
	}

	// Standard format: <symbol>@<streamType>[@<speed>]
	atIdx := strings.Index(stream, "@")
	if atIdx < 0 || atIdx == len(stream)-1 {
		return ""
	}

	// Get everything after @
	rest := stream[atIdx+1:]

	// Check for speed suffix
	if atAtIdx := strings.Index(rest, "@"); atAtIdx > 0 {
		return rest[:atAtIdx]
	}

	return rest
}

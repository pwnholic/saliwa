// Package binance implements the Binance exchange driver.
// WebSocket client with automatic reconnection.
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#websocket-market-streams
package binance

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lilwiggy/ex-act/pkg/domain"
	"github.com/lilwiggy/ex-act/pkg/errors"
	"github.com/lxzan/gws"
)

const (
	exchange = "binance"

	// WebSocket URL paths
	wsCombinedPath = "?streams="
)

// ReconnectConfig holds reconnection settings.
type ReconnectConfig struct {
	InitialDelay time.Duration // Initial reconnection delay (default: 1s)
	MaxDelay     time.Duration // Maximum reconnection delay (default: 60s)
	MaxAttempts  int           // Maximum reconnection attempts (0 = infinite)
	Jitter       float64       // Jitter factor (0-1, default: 0.1)
}

// DefaultReconnectConfig returns the default reconnection configuration.
func DefaultReconnectConfig() ReconnectConfig {
	return ReconnectConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     60 * time.Second,
		MaxAttempts:  0, // Infinite
		Jitter:       0.1,
	}
}

// WSConfig holds WebSocket client configuration.
type WSConfig struct {
	BaseURL      string          // WebSocket base URL (default: production)
	Testnet      bool            // Use testnet URLs
	PingInterval time.Duration   // Ping interval (default: 20s)
	Reconnect    ReconnectConfig // Reconnection settings
}

// DefaultWSConfig returns the default WebSocket configuration.
func DefaultWSConfig() WSConfig {
	return WSConfig{
		Testnet:      false,
		PingInterval: 20 * time.Second,
		Reconnect:    DefaultReconnectConfig(),
	}
}

// Callback functions for different message types.
type WSClientCallbacks struct {
	OnTicker     func(ticker *domain.Ticker)
	OnOrderBook  func(orderBook *domain.OrderBook)
	OnTrade      func(trade *domain.Trade)
	OnKline      func(kline *domain.Kline)
	OnOrder      func(order *domain.Order)
	OnConnect    func()
	OnDisconnect func(err error)
}

// WSClient implements a WebSocket client with automatic reconnection.
// Implements gws.EventHandler interface.
type WSClient struct {
	config        WSConfig
	testnet       bool // Use testnet URLs
	callbacks     WSClientCallbacks
	subscriptions *SubscriptionManager

	// Connection state
	conn       *gws.Conn
	connected  atomic.Bool
	connecting atomic.Bool
	closed     atomic.Bool
	connMu     sync.RWMutex

	// Reconnection state
	reconnectAttempt int
	reconnectMu      sync.Mutex

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Ping ticker
	pingTicker *time.Ticker
	pingMu     sync.Mutex
}

// NewWSClient creates a new WebSocket client.
func NewWSClient(cfg WSConfig) *WSClient {
	if cfg.PingInterval == 0 {
		cfg.PingInterval = 20 * time.Second
	}
	if cfg.Reconnect.InitialDelay == 0 {
		cfg.Reconnect = DefaultReconnectConfig()
	}

	return &WSClient{
		config:        cfg,
		testnet:       cfg.Testnet,
		subscriptions: NewSubscriptionManager(),
	}
}

// OnTicker sets the ticker callback.
func (c *WSClient) OnTicker(fn func(ticker *domain.Ticker)) {
	c.callbacks.OnTicker = fn
}

// OnOrderBook sets the order book callback.
func (c *WSClient) OnOrderBook(fn func(orderBook *domain.OrderBook)) {
	c.callbacks.OnOrderBook = fn
}

// OnTrade sets the trade callback.
func (c *WSClient) OnTrade(fn func(trade *domain.Trade)) {
	c.callbacks.OnTrade = fn
}

// OnKline sets the kline callback.
func (c *WSClient) OnKline(fn func(kline *domain.Kline)) {
	c.callbacks.OnKline = fn
}

// OnOrder sets the order callback.
func (c *WSClient) OnOrder(fn func(order *domain.Order)) {
	c.callbacks.OnOrder = fn
}

// OnConnect sets the connect callback.
func (c *WSClient) OnConnect(fn func()) {
	c.callbacks.OnConnect = fn
}

// OnDisconnect sets the disconnect callback.
func (c *WSClient) OnDisconnect(fn func(err error)) {
	c.callbacks.OnDisconnect = fn
}

// wsBaseURL returns the WebSocket base URL based on testnet flag.
func (c *WSClient) wsBaseURL() string {
	if c.testnet {
		return TestnetWebSocketCombinedURL
	}
	return BaseWebSocketCombinedURL
}

// Connect establishes the WebSocket connection.
// If subscriptions exist, it will subscribe to all existing streams.
func (c *WSClient) Connect() error {
	if c.closed.Load() {
		return errors.NewExchangeError(exchange, "connect", "client is closed", nil)
	}

	if c.connecting.Swap(true) {
		return errors.NewExchangeError(exchange, "connect", "connection already in progress", nil)
	}
	defer c.connecting.Store(false)

	c.ctx, c.cancel = context.WithCancel(context.Background())

	return c.dial()
}

// dial establishes the WebSocket connection.
func (c *WSClient) dial() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	// Build stream URL from subscriptions
	streams := c.subscriptions.Streams()
	var url string
	if len(streams) > 0 {
		combinedStream := CombineStreams(streams)
		url = c.wsBaseURL() + wsCombinedPath + combinedStream
	} else {
		// No subscriptions - connect to base URL (will be idle)
		if c.testnet {
			url = TestnetWebSocketURL
		} else {
			url = BaseWebSocketURL
		}
	}

	// Create client option
	option := &gws.ClientOption{
		Addr: url,
		TlsConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	conn, _, err := gws.NewClient(c, option)
	if err != nil {
		return errors.NewConnectionError(exchange, url, err.Error(), true)
	}

	c.conn = conn
	c.connected.Store(true)
	c.reconnectMu.Lock()
	c.reconnectAttempt = 0
	c.reconnectMu.Unlock()

	// Start read loop
	go c.conn.ReadLoop()

	// Start ping ticker
	c.startPingTicker()

	// Notify connect callback
	c.safeCallback(func() {
		if c.callbacks.OnConnect != nil {
			c.callbacks.OnConnect()
		}
	})

	return nil
}

// Disconnect closes the WebSocket connection without preventing reconnection.
func (c *WSClient) Disconnect() error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn == nil {
		return nil
	}

	c.stopPingTicker()
	c.connected.Store(false)

	// Send close frame and close connection
	c.conn.WriteClose(1000, nil)
	c.conn = nil

	return nil
}

// Close permanently closes the WebSocket client.
// After Close(), the client cannot be reused.
func (c *WSClient) Close() error {
	if c.closed.Swap(true) {
		return nil // Already closed
	}

	c.stopPingTicker()

	if c.cancel != nil {
		c.cancel()
	}

	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn != nil {
		c.conn.WriteClose(1000, nil)
		c.conn = nil
	}

	return nil
}

// IsConnected returns true if the WebSocket is connected.
func (c *WSClient) IsConnected() bool {
	return c.connected.Load()
}

// Subscribe adds a stream subscription.
// NOTE: Binance doesn't support dynamic subscribe on existing connection.
// To add new streams, the connection must be reconnected.
// Returns error if stream is already subscribed.
func (c *WSClient) Subscribe(stream string) error {
	if !c.subscriptions.Subscribe(stream) {
		return nil // Already subscribed
	}

	// If connected, need to reconnect to add new stream
	if c.connected.Load() {
		return c.reconnect()
	}

	return nil
}

// Unsubscribe removes a stream subscription.
// NOTE: Requires reconnection to take effect.
func (c *WSClient) Unsubscribe(stream string) error {
	c.subscriptions.Unsubscribe(stream)
	return nil
}

// OnOpen implements gws.EventHandler - called when connection is established.
func (c *WSClient) OnOpen(socket *gws.Conn) {
	// Connection is now open
	socket.SetDeadline(time.Now().Add(c.config.PingInterval * 2))
}

// OnClose implements gws.EventHandler - called when connection is closed.
func (c *WSClient) OnClose(socket *gws.Conn, err error) {
	c.connected.Store(false)
	c.stopPingTicker()

	// Notify disconnect callback
	c.safeCallback(func() {
		if c.callbacks.OnDisconnect != nil {
			c.callbacks.OnDisconnect(err)
		}
	})

	// Attempt reconnection if not closed intentionally
	if !c.closed.Load() {
		go c.reconnect()
	}
}

// OnPing implements gws.EventHandler - called when ping is received.
func (c *WSClient) OnPing(socket *gws.Conn, payload []byte) {
	socket.SetDeadline(time.Now().Add(c.config.PingInterval * 2))
	socket.WritePong(payload)
}

// OnPong implements gws.EventHandler - called when pong is received.
func (c *WSClient) OnPong(socket *gws.Conn, payload []byte) {
	socket.SetDeadline(time.Now().Add(c.config.PingInterval * 2))
}

// OnMessage implements gws.EventHandler - called when a message is received.
func (c *WSClient) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()

	// Reset deadline on activity
	socket.SetDeadline(time.Now().Add(c.config.PingInterval * 2))

	// Parse the message
	data := message.Bytes()
	if len(data) == 0 {
		return
	}

	// Parse combined stream message
	var wsMsg WSMessage
	if err := json.Unmarshal(data, &wsMsg); err != nil {
		// Not a combined stream message - try direct message
		c.routeDirectMessage(data)
		return
	}

	// Route based on stream name
	c.routeMessage(wsMsg.Stream, wsMsg.Data)
}

// routeMessage routes a message to the appropriate handler based on stream name.
func (c *WSClient) routeMessage(stream string, data []byte) {
	streamType := ParseStreamType(stream)

	switch streamType {
	case "ticker", "miniTicker":
		c.handleTicker(data)
	case "bookTicker":
		c.handleBookTicker(data)
	case "depth", "depth10", "depth20":
		c.handleDepth(data)
	case "trade":
		c.handleTrade(data)
	case "aggTrade":
		c.handleAggTrade(data)
	case "kline":
		c.handleKline(data)
	default:
		// Check if it's an execution report (user data stream)
		var eventType struct {
			EventType string `json:"e"`
		}
		if err := json.Unmarshal(data, &eventType); err == nil {
			switch eventType.EventType {
			case "executionReport":
				c.handleOrderUpdate(data)
			case "balanceUpdate":
				// Ignore for now
			}
		}
	}
}

// routeDirectMessage handles non-combined stream messages.
func (c *WSClient) routeDirectMessage(data []byte) {
	// Try to parse as event with type
	var event struct {
		EventType string `json:"e"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return
	}

	switch event.EventType {
	case "executionReport":
		c.handleOrderUpdate(data)
	case "balanceUpdate":
		// Ignore for now
	case "outboundAccountPosition":
		// Account update - ignore for now
	}
}

// handleTicker handles ticker messages.
func (c *WSClient) handleTicker(data []byte) {
	if c.callbacks.OnTicker == nil {
		return
	}

	var ticker WSTicker
	if err := json.Unmarshal(data, &ticker); err != nil {
		return
	}

	domainTicker, err := ticker.ToDomain(exchange)
	if err != nil {
		return
	}

	c.safeCallback(func() {
		c.callbacks.OnTicker(domainTicker)
	})
}

// handleBookTicker handles book ticker messages.
func (c *WSClient) handleBookTicker(data []byte) {
	if c.callbacks.OnTicker == nil {
		return
	}

	var bookTicker WSBookTicker
	if err := json.Unmarshal(data, &bookTicker); err != nil {
		return
	}

	domainTicker, err := bookTicker.ToDomain(exchange)
	if err != nil {
		return
	}

	c.safeCallback(func() {
		c.callbacks.OnTicker(domainTicker)
	})
}

// handleDepth handles depth messages.
func (c *WSClient) handleDepth(data []byte) {
	if c.callbacks.OnOrderBook == nil {
		return
	}

	// Try as depth update first
	var depthUpdate WSDepthUpdate
	if err := json.Unmarshal(data, &depthUpdate); err == nil && depthUpdate.EventType == "depthUpdate" {
		bids, asks, err := depthUpdate.ToDomain()
		if err != nil {
			return
		}

		orderBook := &domain.OrderBook{
			Exchange:     exchange,
			Symbol:       domain.NormalizeSymbol(depthUpdate.Symbol),
			Bids:         bids,
			Asks:         asks,
			LastUpdateID: depthUpdate.FinalUpdateID,
			Timestamp:    time.UnixMilli(depthUpdate.EventTime),
		}

		c.safeCallback(func() {
			c.callbacks.OnOrderBook(orderBook)
		})
		return
	}

	// Try as depth snapshot
	var depthSnapshot WSDepthSnapshot
	if err := json.Unmarshal(data, &depthSnapshot); err != nil {
		return
	}

	// Need to get symbol from context - this shouldn't happen normally
	// as snapshots come via REST API, but handle gracefully
}

// handleTrade handles trade messages.
func (c *WSClient) handleTrade(data []byte) {
	if c.callbacks.OnTrade == nil {
		return
	}

	var trade WSTrade
	if err := json.Unmarshal(data, &trade); err != nil {
		return
	}

	domainTrade, err := trade.ToDomain(exchange)
	if err != nil {
		return
	}

	c.safeCallback(func() {
		c.callbacks.OnTrade(domainTrade)
	})
}

// handleAggTrade handles aggregated trade messages.
func (c *WSClient) handleAggTrade(data []byte) {
	if c.callbacks.OnTrade == nil {
		return
	}

	var aggTrade WSAggTrade
	if err := json.Unmarshal(data, &aggTrade); err != nil {
		return
	}

	domainTrade, err := aggTrade.ToDomain(exchange)
	if err != nil {
		return
	}

	c.safeCallback(func() {
		c.callbacks.OnTrade(domainTrade)
	})
}

// handleKline handles kline messages.
func (c *WSClient) handleKline(data []byte) {
	if c.callbacks.OnKline == nil {
		return
	}

	var kline WSKline
	if err := json.Unmarshal(data, &kline); err != nil {
		return
	}

	domainKline, err := kline.ToDomain(exchange)
	if err != nil {
		return
	}

	c.safeCallback(func() {
		c.callbacks.OnKline(domainKline)
	})
}

// handleOrderUpdate handles order update messages.
func (c *WSClient) handleOrderUpdate(data []byte) {
	if c.callbacks.OnOrder == nil {
		return
	}

	var orderUpdate WSOrderUpdate
	if err := json.Unmarshal(data, &orderUpdate); err != nil {
		return
	}

	domainOrder, err := orderUpdate.ToDomain(exchange)
	if err != nil {
		return
	}

	c.safeCallback(func() {
		c.callbacks.OnOrder(domainOrder)
	})
}

// safeCallback executes a callback with panic recovery.
// CRITICAL: Callbacks MUST be wrapped in panic recovery.
func (c *WSClient) safeCallback(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't crash the client
			// In production, this should log to the logging system
		}
	}()
	fn()
}

// startPingTicker starts the ping ticker for keepalive.
// CRITICAL: Ping MUST be sent within 1 minute to prevent disconnect.
func (c *WSClient) startPingTicker() {
	c.pingMu.Lock()
	defer c.pingMu.Unlock()

	if c.pingTicker != nil {
		c.pingTicker.Stop()
	}

	c.pingTicker = time.NewTicker(c.config.PingInterval)
	go func() {
		for range c.pingTicker.C {
			c.connMu.RLock()
			conn := c.conn
			c.connMu.RUnlock()

			if conn != nil && c.connected.Load() {
				conn.WritePing(nil)
			}
		}
	}()
}

// stopPingTicker stops the ping ticker.
func (c *WSClient) stopPingTicker() {
	c.pingMu.Lock()
	defer c.pingMu.Unlock()

	if c.pingTicker != nil {
		c.pingTicker.Stop()
		c.pingTicker = nil
	}
}

// reconnect handles reconnection with exponential backoff.
func (c *WSClient) reconnect() error {
	// Prevent multiple reconnect attempts
	c.reconnectMu.Lock()
	if c.reconnectAttempt > 0 {
		c.reconnectMu.Unlock()
		return nil // Reconnection already in progress
	}
	c.reconnectMu.Unlock()

	// Check if closed
	if c.closed.Load() {
		return errors.NewExchangeError(exchange, "reconnect", "client is closed", nil)
	}

	// Disconnect existing connection
	_ = c.Disconnect()

	for {
		// Check if closed or context cancelled
		if c.closed.Load() || (c.ctx != nil && c.ctx.Err() != nil) {
			return errors.NewExchangeError(exchange, "reconnect", "client closed or context cancelled", nil)
		}

		// Increment attempt counter
		c.reconnectMu.Lock()
		c.reconnectAttempt++
		attempt := c.reconnectAttempt
		c.reconnectMu.Unlock()

		// Check max attempts
		if c.config.Reconnect.MaxAttempts > 0 && attempt > c.config.Reconnect.MaxAttempts {
			return errors.NewWebSocketReconnectError(
				exchange,
				"",
				"max reconnection attempts exceeded",
				attempt,
				c.config.Reconnect.MaxAttempts,
			)
		}

		// Calculate backoff with jitter
		delay := c.calculateBackoff(attempt)
		time.Sleep(delay)

		// Attempt to connect
		if err := c.dial(); err != nil {
			// Continue trying
			continue
		}

		// Success - reset attempt counter
		c.reconnectMu.Lock()
		c.reconnectAttempt = 0
		c.reconnectMu.Unlock()

		return nil
	}
}

// calculateBackoff calculates the reconnection delay with exponential backoff and jitter.
// Formula: delay = min(initialDelay * 2^attempt, maxDelay) * (1 + random * jitter)
func (c *WSClient) calculateBackoff(attempt int) time.Duration {
	cfg := c.config.Reconnect

	// Calculate exponential delay
	delay := cfg.InitialDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
			break
		}
	}

	// Add jitter (10% by default)
	if cfg.Jitter > 0 {
		jitter := time.Duration(float64(delay) * cfg.Jitter * (rand.Float64()*2 - 1))
		delay += jitter
	}

	return delay
}

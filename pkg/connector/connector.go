package connector

import (
	"context"
	"fmt"
	stdsync "sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lilwiggy/ex-act/internal/circuit"
	"github.com/lilwiggy/ex-act/internal/driver/binance"
	internalsync "github.com/lilwiggy/ex-act/internal/sync"
	"github.com/lilwiggy/ex-act/pkg/domain"
)

// Connector provides exchange connectivity with fault tolerance.
// One Connector instance connects to one exchange.
type Connector struct {
	config   Config
	exchange string

	// Components
	restClient     *binance.RESTClient
	wsClient       *binance.WSClient
	circuitBreaker *circuit.Breaker
	clockSync      *internalsync.ClockSync
	nonceGen       *internalsync.NonceGenerator

	// State
	running   atomic.Bool
	ready     chan struct{}
	readyOnce stdsync.Once

	// Handlers
	handlers Handlers

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     stdsync.WaitGroup
}

// New creates a new Connector for an exchange.
func New(cfg Config) (*Connector, error) {
	if err := cfg.Exchange.Validate(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &Connector{
		config:   cfg,
		exchange: cfg.Exchange.Name,
		ready:    make(chan struct{}),
		nonceGen: internalsync.NewNonceGenerator(),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Initialize components
	if err := c.initComponents(); err != nil {
		cancel()
		return nil, err
	}

	return c, nil
}

// initComponents initializes all components.
func (c *Connector) initComponents() error {
	var err error

	// Create REST client
	restCfg := binance.Config{
		BaseURL:   "", // Use default or testnet based on config
		APIKey:    c.config.Exchange.APIKey,
		APISecret: c.config.Exchange.APISecret,
		Timeout:   c.config.Connection.Timeout,
		MaxWeight: c.config.RateLimit.MaxWeight,
		Testnet:   c.config.Exchange.Testnet,
	}

	c.restClient, err = binance.NewRESTClient(restCfg)
	if err != nil {
		return fmt.Errorf("failed to create REST client: %w", err)
	}

	// Create circuit breaker
	if c.config.CircuitBreaker.Enabled {
		c.circuitBreaker = circuit.NewBreaker(c.exchange, circuit.Config{
			MaxFailures:      c.config.CircuitBreaker.MaxFailures,
			SuccessThreshold: c.config.CircuitBreaker.SuccessThreshold,
			OpenTimeout:      c.config.CircuitBreaker.OpenTimeout,
		})
	}

	// Create clock sync
	if c.config.ClockSync.Enabled {
		c.clockSync = internalsync.NewClockSync(c.exchange, internalsync.ClockConfig{
			MaxOffset:    c.config.ClockSync.MaxOffset,
			SyncInterval: c.config.ClockSync.SyncInterval,
			TimeProvider: c.restClient.GetServerTime,
		})
	}

	// Create WebSocket client
	wsCfg := binance.WSConfig{
		Testnet:      c.config.Exchange.Testnet,
		PingInterval: c.config.Connection.PingInterval,
		Reconnect: binance.ReconnectConfig{
			InitialDelay: c.config.Connection.ReconnectDelay,
			MaxDelay:     c.config.Connection.MaxReconnectWait,
			MaxAttempts:  0, // Infinite
			Jitter:       0.1,
		},
	}

	c.wsClient = binance.NewWSClient(wsCfg)

	// Set up WebSocket handlers
	c.setupWSHandlers()

	return nil
}

// setupWSHandlers sets up WebSocket event handlers.
func (c *Connector) setupWSHandlers() {
	c.wsClient.OnTicker(func(ticker *domain.Ticker) {
		if c.handlers.OnTicker != nil {
			c.safeHandler(func() {
				c.handlers.OnTicker(c.exchange, ticker)
			})
		}
	})

	c.wsClient.OnOrderBook(func(ob *domain.OrderBook) {
		if c.handlers.OnOrderBook != nil {
			c.safeHandler(func() {
				c.handlers.OnOrderBook(c.exchange, ob)
			})
		}
	})

	c.wsClient.OnTrade(func(trade *domain.Trade) {
		if c.handlers.OnTrade != nil {
			c.safeHandler(func() {
				c.handlers.OnTrade(c.exchange, trade)
			})
		}
	})

	c.wsClient.OnConnect(func() {
		log.Info().Str("exchange", c.exchange).Msg("WebSocket connected")
		if c.handlers.OnConnect != nil {
			c.handlers.OnConnect(c.exchange, true)
		}
		c.markReady()
	})

	c.wsClient.OnDisconnect(func(err error) {
		log.Error().Err(err).Str("exchange", c.exchange).Msg("WebSocket disconnected")
		if c.handlers.OnDisconnect != nil {
			c.handlers.OnDisconnect(c.exchange, false)
		}
	})
}

// Start starts the connector.
// It returns immediately, use Ready() to wait for full initialization.
func (c *Connector) Start() error {
	if c.running.Swap(true) {
		return fmt.Errorf("connector already running")
	}

	log.Info().Str("exchange", c.exchange).Msg("starting connector")

	// Start clock sync (required for signed requests)
	if c.clockSync != nil {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			if err := c.clockSync.Start(); err != nil {
				log.Error().Err(err).Msg("clock sync failed")
				if c.handlers.OnError != nil {
					c.handlers.OnError(c.exchange, err)
				}
			}
		}()
	}

	// Connect WebSocket
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		if err := c.wsClient.Connect(); err != nil {
			log.Error().Err(err).Msg("WebSocket connection failed")
			if c.handlers.OnError != nil {
				c.handlers.OnError(c.exchange, err)
			}
		}
	}()

	// For simple cases without subscriptions, mark ready immediately
	// WebSocket will mark ready when it connects

	return nil
}

// Stop stops the connector gracefully.
func (c *Connector) Stop() error {
	if !c.running.Swap(false) {
		return nil // Not running
	}

	log.Info().Str("exchange", c.exchange).Msg("stopping connector")

	// Cancel context
	c.cancel()

	// Stop components
	if c.clockSync != nil {
		c.clockSync.Stop()
	}

	if c.wsClient != nil {
		c.wsClient.Close()
	}

	// Wait for goroutines
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		log.Warn().Msg("timeout waiting for goroutines to stop")
	}

	// Close REST client
	if c.restClient != nil {
		c.restClient.Close()
	}

	log.Info().Str("exchange", c.exchange).Msg("connector stopped")

	return nil
}

// Ready returns a channel that is closed when the connector is ready.
func (c *Connector) Ready() <-chan struct{} {
	return c.ready
}

// markReady marks the connector as ready.
func (c *Connector) markReady() {
	c.readyOnce.Do(func() {
		close(c.ready)
	})
}

// IsRunning returns true if the connector is running.
func (c *Connector) IsRunning() bool {
	return c.running.Load()
}

// IsConnected returns true if WebSocket is connected.
func (c *Connector) IsConnected() bool {
	return c.wsClient != nil && c.wsClient.IsConnected()
}

// Exchange returns the exchange name.
func (c *Connector) Exchange() string {
	return c.exchange
}

// SetHandlers sets event handlers.
func (c *Connector) SetHandlers(handlers Handlers) {
	c.handlers = handlers
}

// SubscribeTicker subscribes to ticker updates for a symbol.
// Returns an unsubscribe function.
func (c *Connector) SubscribeTicker(symbol string) (func(), error) {
	if !c.running.Load() {
		return nil, fmt.Errorf("connector not running")
	}

	stream := binance.NewStreamBuilder(symbol).Ticker()
	if err := c.wsClient.Subscribe(stream); err != nil {
		return nil, err
	}

	return func() {
		c.wsClient.Unsubscribe(stream)
	}, nil
}

// SubscribeOrderBook subscribes to order book updates for a symbol.
func (c *Connector) SubscribeOrderBook(symbol string) (func(), error) {
	if !c.running.Load() {
		return nil, fmt.Errorf("connector not running")
	}

	stream := binance.NewStreamBuilder(symbol).Depth()
	if err := c.wsClient.Subscribe(stream); err != nil {
		return nil, err
	}

	return func() {
		c.wsClient.Unsubscribe(stream)
	}, nil
}

// SubscribeTrades subscribes to trade updates for a symbol.
func (c *Connector) SubscribeTrades(symbol string) (func(), error) {
	if !c.running.Load() {
		return nil, fmt.Errorf("connector not running")
	}

	stream := binance.NewStreamBuilder(symbol).Trade()
	if err := c.wsClient.Subscribe(stream); err != nil {
		return nil, err
	}

	return func() {
		c.wsClient.Unsubscribe(stream)
	}, nil
}

// Ping tests REST connectivity.
func (c *Connector) Ping(ctx context.Context) error {
	if c.circuitBreaker != nil {
		return c.circuitBreaker.Execute(func() error {
			return c.restClient.Ping(ctx)
		})
	}
	return c.restClient.Ping(ctx)
}

// GetServerTime retrieves the exchange server time.
func (c *Connector) GetServerTime(ctx context.Context) (int64, error) {
	if c.circuitBreaker != nil {
		result, err := c.circuitBreaker.ExecuteWithResult(func() (any, error) {
			return c.restClient.GetServerTime(ctx)
		})
		if err != nil {
			return 0, err
		}
		return result.(int64), nil
	}
	return c.restClient.GetServerTime(ctx)
}

// GetExchangeInfo retrieves exchange trading rules.
func (c *Connector) GetExchangeInfo(ctx context.Context) (*binance.ExchangeInfo, error) {
	if c.circuitBreaker != nil {
		result, err := c.circuitBreaker.ExecuteWithResult(func() (any, error) {
			return c.restClient.GetExchangeInfo(ctx)
		})
		if err != nil {
			return nil, err
		}
		return result.(*binance.ExchangeInfo), nil
	}
	return c.restClient.GetExchangeInfo(ctx)
}

// CircuitBreakerStats returns circuit breaker statistics.
func (c *Connector) CircuitBreakerStats() (circuit.Stats, error) {
	if c.circuitBreaker == nil {
		return circuit.Stats{}, fmt.Errorf("circuit breaker not enabled")
	}
	return c.circuitBreaker.Stats(), nil
}

// ClockOffset returns the current clock offset.
func (c *Connector) ClockOffset() time.Duration {
	if c.clockSync == nil {
		return 0
	}
	return c.clockSync.Offset()
}

// safeHandler executes a handler with panic recovery.
func (c *Connector) safeHandler(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Interface("panic", r).Str("exchange", c.exchange).Msg("handler panic recovered")
		}
	}()
	fn()
}

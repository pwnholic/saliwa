// Package connector provides the public API for exchange connectivity.
package connector

import (
	"fmt"
	"time"

	"github.com/lilwiggy/ex-act/pkg/errors"
)

// Config contains the main connector configuration.
type Config struct {
	// Exchange settings
	Exchange ExchangeConfig

	// Rate limiting
	RateLimit RateLimitConfig

	// Resilience
	CircuitBreaker CircuitBreakerConfig
	ClockSync      ClockSyncConfig

	// Connection
	Connection ConnectionConfig
}

// ExchangeConfig contains exchange-specific settings.
type ExchangeConfig struct {
	Name      string // Exchange name: "binance" or "bybit"
	APIKey    string // API key for authentication
	APISecret string // API secret for signing
	Testnet   bool   // Use testnet endpoints
}

// Validate validates exchange configuration.
func (c *ExchangeConfig) Validate() error {
	if c.Name == "" {
		return errors.NewValidationError("name", "", "must not be empty")
	}
	if c.Name != "binance" && c.Name != "bybit" {
		return errors.NewValidationError("name", c.Name, "must be 'binance' or 'bybit'")
	}
	// APIKey and APISecret can be empty for public-only access
	return nil
}

// RateLimitConfig contains rate limiting settings.
type RateLimitConfig struct {
	MaxWeight    int           // Maximum weight per interval (Binance: 1200)
	RequestDelay time.Duration // Minimum delay between requests
	Enabled      bool          // Enable rate limiting (default: true)
}

// DefaultRateLimitConfig returns default rate limit configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxWeight:    1200,
		RequestDelay: 0,
		Enabled:      true,
	}
}

// CircuitBreakerConfig contains circuit breaker settings.
type CircuitBreakerConfig struct {
	MaxFailures      int           // Failures before opening
	SuccessThreshold int           // Successes to close from half-open
	OpenTimeout      time.Duration // Time before half-open
	Enabled          bool          // Enable circuit breaker (default: true)
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:      5,
		SuccessThreshold: 3,
		OpenTimeout:      30 * time.Second,
		Enabled:          true,
	}
}

// ClockSyncConfig contains clock synchronization settings.
type ClockSyncConfig struct {
	MaxOffset    time.Duration // Maximum allowed offset
	SyncInterval time.Duration // How often to sync
	Enabled      bool          // Enable clock sync (default: true)
}

// DefaultClockSyncConfig returns default clock sync configuration.
func DefaultClockSyncConfig() ClockSyncConfig {
	return ClockSyncConfig{
		MaxOffset:    500 * time.Millisecond,
		SyncInterval: 5 * time.Minute,
		Enabled:      true,
	}
}

// ConnectionConfig contains connection settings.
type ConnectionConfig struct {
	Timeout          time.Duration // REST request timeout
	PingInterval     time.Duration // WebSocket ping interval
	ReconnectDelay   time.Duration // Initial reconnect delay
	MaxReconnectWait time.Duration // Maximum reconnect wait
}

// DefaultConnectionConfig returns default connection configuration.
func DefaultConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		Timeout:          10 * time.Second,
		PingInterval:     20 * time.Second,
		ReconnectDelay:   1 * time.Second,
		MaxReconnectWait: 60 * time.Second,
	}
}

// Builder provides a fluent interface for building Config.
type Builder struct {
	config Config
	errs   []error
}

// NewConfigBuilder creates a new configuration builder.
func NewConfigBuilder() *Builder {
	return &Builder{
		config: Config{
			RateLimit:      DefaultRateLimitConfig(),
			CircuitBreaker: DefaultCircuitBreakerConfig(),
			ClockSync:      DefaultClockSyncConfig(),
			Connection:     DefaultConnectionConfig(),
		},
	}
}

// Exchange sets exchange configuration.
func (b *Builder) Exchange(name, apiKey, apiSecret string, testnet bool) *Builder {
	b.config.Exchange = ExchangeConfig{
		Name:      name,
		APIKey:    apiKey,
		APISecret: apiSecret,
		Testnet:   testnet,
	}
	return b
}

// RateLimit sets rate limit configuration.
func (b *Builder) RateLimit(maxWeight int, delay time.Duration) *Builder {
	b.config.RateLimit = RateLimitConfig{
		MaxWeight:    maxWeight,
		RequestDelay: delay,
		Enabled:      true,
	}
	return b
}

// CircuitBreaker sets circuit breaker configuration.
func (b *Builder) CircuitBreaker(maxFailures int, timeout time.Duration) *Builder {
	b.config.CircuitBreaker = CircuitBreakerConfig{
		MaxFailures: maxFailures,
		OpenTimeout: timeout,
		Enabled:     true,
	}
	return b
}

// ClockSync sets clock sync configuration.
func (b *Builder) ClockSync(maxOffset, interval time.Duration) *Builder {
	b.config.ClockSync = ClockSyncConfig{
		MaxOffset:    maxOffset,
		SyncInterval: interval,
		Enabled:      true,
	}
	return b
}

// Timeout sets connection timeout.
func (b *Builder) Timeout(timeout time.Duration) *Builder {
	b.config.Connection.Timeout = timeout
	return b
}

// Build validates and returns the configuration.
func (b *Builder) Build() (Config, error) {
	if err := b.config.Exchange.Validate(); err != nil {
		b.errs = append(b.errs, err)
	}

	if len(b.errs) > 0 {
		return Config{}, fmt.Errorf("configuration errors: %v", b.errs)
	}

	// Return a copy
	return b.config, nil
}

// MustBuild validates and returns the configuration, panicking on error.
func (b *Builder) MustBuild() Config {
	cfg, err := b.Build()
	if err != nil {
		panic(err)
	}
	return cfg
}

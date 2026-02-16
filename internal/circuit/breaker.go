// Package circuit implements circuit breaker pattern for fault tolerance.
// Used to prevent cascading failures when exchange becomes unhealthy.
package circuit

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sony/gobreaker"

	"github.com/lilwiggy/ex-act/pkg/errors"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

// String returns the string representation of the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return "unknown"
	}
}

// Breaker implements a circuit breaker for exchange operations.
// States:
//   - Closed: Normal operation, requests pass through
//   - Open: Requests blocked, waiting for timeout
//   - Half-Open: Testing recovery, limited requests allowed
type Breaker struct {
	exchange string
	breaker  *gobreaker.CircuitBreaker
	config   Config

	// Metrics
	mutex           sync.RWMutex
	totalRequests   int64
	totalFailures   int64
	totalSuccesses  int64
	lastFailure     time.Time
	lastStateChange time.Time
}

// Config contains circuit breaker configuration.
type Config struct {
	// Thresholds
	MaxFailures      int // Failures before opening (default: 5)
	SuccessThreshold int // Successes in half-open to close (default: 3)

	// Timeouts
	OpenTimeout time.Duration // Time before half-open (default: 30s)

	// Callbacks
	OnStateChange func(from, to State)
}

// DefaultConfig returns the default circuit breaker configuration.
func DefaultConfig() Config {
	return Config{
		MaxFailures:      5,
		SuccessThreshold: 3,
		OpenTimeout:      30 * time.Second,
	}
}

// NewBreaker creates a new circuit breaker.
func NewBreaker(exchange string, cfg Config) *Breaker {
	if cfg.MaxFailures == 0 {
		cfg.MaxFailures = DefaultConfig().MaxFailures
	}
	if cfg.SuccessThreshold == 0 {
		cfg.SuccessThreshold = DefaultConfig().SuccessThreshold
	}
	if cfg.OpenTimeout == 0 {
		cfg.OpenTimeout = DefaultConfig().OpenTimeout
	}

	name := exchange + "-breaker"

	breakerSettings := gobreaker.Settings{
		Name:        name,
		MaxRequests: uint32(cfg.SuccessThreshold),
		Interval:    0, // Don't clear counts periodically
		Timeout:     cfg.OpenTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= uint32(cfg.MaxFailures)
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Info().
				Str("exchange", exchange).
				Str("from", from.String()).
				Str("to", to.String()).
				Msg("circuit breaker state changed")

			if cfg.OnStateChange != nil {
				cfg.OnStateChange(State(from), State(to))
			}
		},
	}

	return &Breaker{
		exchange:        exchange,
		breaker:         gobreaker.NewCircuitBreaker(breakerSettings),
		config:          cfg,
		lastStateChange: time.Now(),
	}
}

// Execute runs the given function through the circuit breaker.
// Returns CircuitBreakerError if the breaker is open.
func (b *Breaker) Execute(fn func() error) error {
	_, err := b.breaker.Execute(func() (any, error) {
		return nil, fn()
	})

	if err != nil {
		// Check if it's a circuit breaker error
		if err == gobreaker.ErrOpenState {
			return errors.NewCircuitBreakerError(b.exchange, "open", "circuit breaker is open", 0, b.timeToHalfOpen())
		}
		if err == gobreaker.ErrTooManyRequests {
			return errors.NewCircuitBreakerError(b.exchange, "half-open", "too many requests in half-open state", 0, b.timeToHalfOpen())
		}

		// Track failure
		b.recordFailure()
		return err
	}

	// Track success
	b.recordSuccess()
	return nil
}

// ExecuteWithResult runs the given function and returns its result.
func (b *Breaker) ExecuteWithResult(fn func() (any, error)) (any, error) {
	result, err := b.breaker.Execute(fn)

	if err != nil {
		if err == gobreaker.ErrOpenState {
			return nil, errors.NewCircuitBreakerError(b.exchange, "open", "circuit breaker is open", 0, b.timeToHalfOpen())
		}
		if err == gobreaker.ErrTooManyRequests {
			return nil, errors.NewCircuitBreakerError(b.exchange, "half-open", "too many requests in half-open state", 0, b.timeToHalfOpen())
		}

		b.recordFailure()
		return nil, err
	}

	b.recordSuccess()
	return result, nil
}

// State returns the current circuit breaker state.
func (b *Breaker) State() State {
	switch b.breaker.State() {
	case gobreaker.StateClosed:
		return StateClosed
	case gobreaker.StateHalfOpen:
		return StateHalfOpen
	case gobreaker.StateOpen:
		return StateOpen
	default:
		return StateClosed
	}
}

// IsOpen returns true if the circuit breaker is open.
func (b *Breaker) IsOpen() bool {
	return b.breaker.State() == gobreaker.StateOpen
}

// IsClosed returns true if the circuit breaker is closed.
func (b *Breaker) IsClosed() bool {
	return b.breaker.State() == gobreaker.StateClosed
}

// IsHalfOpen returns true if the circuit breaker is half-open.
func (b *Breaker) IsHalfOpen() bool {
	return b.breaker.State() == gobreaker.StateHalfOpen
}

// timeToHalfOpen returns the time until the breaker transitions to half-open.
func (b *Breaker) timeToHalfOpen() time.Duration {
	if b.breaker.State() != gobreaker.StateOpen {
		return 0
	}

	// Calculate time remaining until timeout
	elapsed := time.Since(b.lastStateChange)
	remaining := b.config.OpenTimeout - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Stats returns circuit breaker statistics.
func (b *Breaker) Stats() Stats {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return Stats{
		Exchange:       b.exchange,
		State:          b.State().String(),
		TotalRequests:  b.totalRequests,
		TotalFailures:  b.totalFailures,
		TotalSuccesses: b.totalSuccesses,
		LastFailure:    b.lastFailure,
	}
}

// Stats contains circuit breaker statistics.
type Stats struct {
	Exchange       string    `json:"exchange"`
	State          string    `json:"state"`
	TotalRequests  int64     `json:"total_requests"`
	TotalFailures  int64     `json:"total_failures"`
	TotalSuccesses int64     `json:"total_successes"`
	LastFailure    time.Time `json:"last_failure"`
}

// recordSuccess records a successful request.
func (b *Breaker) recordSuccess() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.totalRequests++
	b.totalSuccesses++
}

// recordFailure records a failed request.
func (b *Breaker) recordFailure() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.totalRequests++
	b.totalFailures++
	b.lastFailure = time.Now()
}

// Reset resets the circuit breaker to closed state.
func (b *Breaker) Reset() {
	// gobreaker doesn't have a direct reset, so we create a new breaker
	b.breaker = gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        b.exchange + "-breaker",
		MaxRequests: uint32(b.config.SuccessThreshold),
		Interval:    0,
		Timeout:     b.config.OpenTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= uint32(b.config.MaxFailures)
		},
	})

	b.mutex.Lock()
	b.lastStateChange = time.Now()
	b.mutex.Unlock()

	log.Info().Str("exchange", b.exchange).Msg("circuit breaker reset")
}

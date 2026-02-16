// Package errors provides typed errors for the exchange connector.
package errors

import (
	"fmt"
	"time"
)

// ConnectionError represents a connection failure to the exchange.
type ConnectionError struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// Endpoint is the endpoint that failed to connect
	Endpoint string `json:"endpoint,omitempty"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// Retryable indicates if the connection can be retried
	Retryable bool `json:"retryable"`

	// Attempt is the connection attempt number
	Attempt int `json:"attempt,omitempty"`

	// MaxAttempts is the maximum number of attempts
	MaxAttempts int `json:"max_attempts,omitempty"`

	// Underlying error that caused this error
	cause error `json:"-"`
}

// Error implements the error interface.
func (e *ConnectionError) Error() string {
	if e.Endpoint != "" {
		if e.Attempt > 0 {
			return fmt.Sprintf("[%s] connection failed to %s (attempt %d/%d): %s",
				e.Exchange, e.Endpoint, e.Attempt, e.MaxAttempts, e.Message)
		}
		return fmt.Sprintf("[%s] connection failed to %s: %s", e.Exchange, e.Endpoint, e.Message)
	}
	if e.Attempt > 0 {
		return fmt.Sprintf("[%s] connection failed (attempt %d/%d): %s",
			e.Exchange, e.Attempt, e.MaxAttempts, e.Message)
	}
	return fmt.Sprintf("[%s] connection failed: %s", e.Exchange, e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *ConnectionError) Unwrap() error {
	return e.cause
}

// NewConnectionError creates a new ConnectionError.
func NewConnectionError(exchange, endpoint, message string, retryable bool) *ConnectionError {
	return &ConnectionError{
		Exchange:  exchange,
		Endpoint:  endpoint,
		Message:   message,
		Retryable: retryable,
	}
}

// WebSocketReconnectError represents a WebSocket reconnection failure.
type WebSocketReconnectError struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// Stream is the WebSocket stream that failed
	Stream string `json:"stream,omitempty"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// Attempt is the reconnection attempt number
	Attempt int `json:"attempt"`

	// MaxAttempts is the maximum number of reconnection attempts
	MaxAttempts int `json:"max_attempts"`

	// LastConnected is the time when the connection was last active
	LastConnected time.Time `json:"last_connected,omitempty"`

	// Underlying error that caused this error
	cause error `json:"-"`
}

// Error implements the error interface.
func (e *WebSocketReconnectError) Error() string {
	if e.Stream != "" {
		return fmt.Sprintf("[%s] WebSocket reconnection failed for stream %s (attempt %d/%d): %s",
			e.Exchange, e.Stream, e.Attempt, e.MaxAttempts, e.Message)
	}
	return fmt.Sprintf("[%s] WebSocket reconnection failed (attempt %d/%d): %s",
		e.Exchange, e.Attempt, e.MaxAttempts, e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *WebSocketReconnectError) Unwrap() error {
	return e.cause
}

// NewWebSocketReconnectError creates a new WebSocketReconnectError.
func NewWebSocketReconnectError(exchange, stream, message string, attempt, maxAttempts int) *WebSocketReconnectError {
	return &WebSocketReconnectError{
		Exchange:    exchange,
		Stream:      stream,
		Message:     message,
		Attempt:     attempt,
		MaxAttempts: maxAttempts,
	}
}

// CircuitBreakerError represents a circuit breaker open error.
type CircuitBreakerError struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// State is the current state of the circuit breaker
	State string `json:"state"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// Failures is the number of consecutive failures
	Failures int `json:"failures"`

	// LastFailure is the time of the last failure
	LastFailure time.Time `json:"last_failure"`

	// ResetAfter is the time until the circuit breaker resets
	ResetAfter time.Duration `json:"reset_after"`
}

// Error implements the error interface.
func (e *CircuitBreakerError) Error() string {
	return fmt.Sprintf("[%s] circuit breaker %s: %s (failures: %d, reset after: %v)",
		e.Exchange, e.State, e.Message, e.Failures, e.ResetAfter)
}

// IsRetryable returns true if the circuit breaker will reset and allow retries.
func (e *CircuitBreakerError) IsRetryable() bool {
	return e.ResetAfter > 0
}

// NewCircuitBreakerError creates a new CircuitBreakerError.
func NewCircuitBreakerError(exchange, state, message string, failures int, resetAfter time.Duration) *CircuitBreakerError {
	return &CircuitBreakerError{
		Exchange:    exchange,
		State:       state,
		Message:     message,
		Failures:    failures,
		ResetAfter:  resetAfter,
		LastFailure: time.Now(),
	}
}

// ClockSyncError represents a clock synchronization error.
// This occurs when the local clock is out of sync with the exchange server.
type ClockSyncError struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// LocalTime is the local timestamp
	LocalTime time.Time `json:"local_time"`

	// ServerTime is the server timestamp
	ServerTime time.Time `json:"server_time"`

	// Drift is the time difference between local and server
	Drift time.Duration `json:"drift"`

	// Message is a human-readable error message
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *ClockSyncError) Error() string {
	return fmt.Sprintf("[%s] clock synchronization error: %s (local: %v, server: %v, drift: %v)",
		e.Exchange, e.Message, e.LocalTime, e.ServerTime, e.Drift)
}

// NewClockSyncError creates a new ClockSyncError.
func NewClockSyncError(exchange string, localTime, serverTime time.Time, drift time.Duration) *ClockSyncError {
	return &ClockSyncError{
		Exchange:   exchange,
		LocalTime:  localTime,
		ServerTime: serverTime,
		Drift:      drift,
		Message:    "clock drift detected",
	}
}

// SignatureError represents a request signature/validation error.
type SignatureError struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// Operation is the operation that failed signature validation
	Operation string `json:"operation,omitempty"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// Underlying error that caused this error
	cause error `json:"-"`
}

// Error implements the error interface.
func (e *SignatureError) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("[%s] signature error in %s: %s", e.Exchange, e.Operation, e.Message)
	}
	return fmt.Sprintf("[%s] signature error: %s", e.Exchange, e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *SignatureError) Unwrap() error {
	return e.cause
}

// NewSignatureError creates a new SignatureError.
func NewSignatureError(exchange, operation, message string) *SignatureError {
	return &SignatureError{
		Exchange:  exchange,
		Operation: operation,
		Message:   message,
	}
}

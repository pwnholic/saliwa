// Package errors provides typed errors for the exchange connector.
// All errors support Go 1.13+ error wrapping with errors.Is and errors.As.
package errors

import (
	"errors"
	"fmt"
)

// ExchangeError is the base error type for all exchange-related errors.
// It provides context about which exchange and operation caused the error.
type ExchangeError struct {
	// Exchange is the name of the exchange that caused the error
	Exchange string `json:"exchange"`

	// Operation is the operation that failed (e.g., "place_order", "get_balance")
	Operation string `json:"operation,omitempty"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// Code is an optional error code from the exchange
	Code string `json:"code,omitempty"`

	// Underlying error that caused this error
	cause error `json:"-"`
}

// Error implements the error interface.
func (e *ExchangeError) Error() string {
	if e.Exchange != "" {
		if e.Operation != "" {
			return fmt.Sprintf("[%s:%s] %s", e.Exchange, e.Operation, e.Message)
		}
		return fmt.Sprintf("[%s] %s", e.Exchange, e.Message)
	}
	return e.Message
}

// Unwrap returns the underlying cause of the error.
func (e *ExchangeError) Unwrap() error {
	return e.cause
}

// NewExchangeError creates a new ExchangeError.
func NewExchangeError(exchange, operation, message string, cause error) *ExchangeError {
	return &ExchangeError{
		Exchange:  exchange,
		Operation: operation,
		Message:   message,
		cause:     cause,
	}
}

// ValidationError represents a validation failure for input data.
type ValidationError struct {
	// Field is the field that failed validation
	Field string `json:"field"`

	// Value is the invalid value (may be nil)
	Value any `json:"value,omitempty"`

	// Message is the validation error message
	Message string `json:"message"`

	// Underlying error that caused this error
	cause error `json:"-"`
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error: %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *ValidationError) Unwrap() error {
	return e.cause
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field string, value any, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// NotFoundError represents a resource not found error.
type NotFoundError struct {
	// Resource is the type of resource that was not found
	Resource string `json:"resource"`

	// Identifier is the identifier used to look up the resource
	Identifier string `json:"identifier"`

	// Message is an optional additional message
	Message string `json:"message,omitempty"`
}

// Error implements the error interface.
func (e *NotFoundError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s not found: %s (%s)", e.Resource, e.Identifier, e.Message)
	}
	return fmt.Sprintf("%s not found: %s", e.Resource, e.Identifier)
}

// NewNotFoundError creates a new NotFoundError.
func NewNotFoundError(resource, identifier string) *NotFoundError {
	return &NotFoundError{
		Resource:   resource,
		Identifier: identifier,
	}
}

// IsRetryable returns true if the error is transient and the operation can be retried.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for known retryable error types
	var rateLimitErr *RateLimitError
	if errors.As(err, &rateLimitErr) {
		return rateLimitErr.IsRetryable()
	}

	var connErr *ConnectionError
	if errors.As(err, &connErr) {
		return connErr.Retryable
	}

	var wsReconnErr *WebSocketReconnectError
	if errors.As(err, &wsReconnErr) {
		return true // WebSocket reconnection can be retried
	}

	var circuitErr *CircuitBreakerError
	if errors.As(err, &circuitErr) {
		return circuitErr.IsRetryable()
	}

	// Check for standard errors that are typically retryable
	var timeoutErr interface{ Timeout() bool }
	if errors.As(err, &timeoutErr) {
		return timeoutErr.Timeout()
	}

	return false
}

// Is is an alias for errors.Is for convenience.
var Is = errors.Is

// As is an alias for errors.As for convenience.
var As = errors.As

// New creates a new error with the given message.
var New = errors.New

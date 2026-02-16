// Package errors provides typed errors for the exchange connector.
package errors

import (
	"fmt"
	"time"
)

// RateLimitError represents a rate limit exceeded error.
type RateLimitError struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// RetryAfter is the duration to wait before retrying
	RetryAfter time.Duration `json:"retry_after"`

	// Weight is the request weight that caused the limit to be exceeded
	Weight int `json:"weight,omitempty"`

	// Limit is the rate limit that was exceeded
	Limit int `json:"limit,omitempty"`

	// Message is a human-readable error message
	Message string `json:"message"`

	// IsBan indicates if this is an IP ban (more severe)
	IsBan bool `json:"is_ban"`
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	if e.IsBan {
		return fmt.Sprintf("[%s] rate limit: IP banned, retry after %v", e.Exchange, e.RetryAfter)
	}
	if e.Weight > 0 {
		return fmt.Sprintf("[%s] rate limit exceeded (weight: %d, limit: %d), retry after %v",
			e.Exchange, e.Weight, e.Limit, e.RetryAfter)
	}
	return fmt.Sprintf("[%s] rate limit exceeded, retry after %v", e.Exchange, e.RetryAfter)
}

// IsRetryable returns true if the request can be retried after waiting.
func (e *RateLimitError) IsRetryable() bool {
	return e.RetryAfter > 0 && !e.IsBan
}

// NewRateLimitError creates a new RateLimitError.
func NewRateLimitError(exchange string, retryAfter time.Duration, weight int) *RateLimitError {
	return &RateLimitError{
		Exchange:   exchange,
		RetryAfter: retryAfter,
		Weight:     weight,
		Message:    "rate limit exceeded",
	}
}

// IPBanError represents an IP ban error from the exchange.
type IPBanError struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// Reason is the reason for the ban
	Reason string `json:"reason,omitempty"`

	// RetryAfter is the duration until the ban is lifted (if known)
	RetryAfter time.Duration `json:"retry_after,omitempty"`

	// Message is a human-readable error message
	Message string `json:"message"`
}

// Error implements the error interface.
func (e *IPBanError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("[%s] IP banned: %s (retry after %v)", e.Exchange, e.Reason, e.RetryAfter)
	}
	return fmt.Sprintf("[%s] IP banned: %s", e.Exchange, e.Reason)
}

// IsRetryable returns true if the ban is temporary and will be lifted.
func (e *IPBanError) IsRetryable() bool {
	return e.RetryAfter > 0
}

// NewIPBanError creates a new IPBanError.
func NewIPBanError(exchange, reason string, retryAfter time.Duration) *IPBanError {
	return &IPBanError{
		Exchange:   exchange,
		Reason:     reason,
		RetryAfter: retryAfter,
		Message:    "IP banned",
	}
}

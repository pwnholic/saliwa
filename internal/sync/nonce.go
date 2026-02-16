package sync

import (
	"crypto/rand"
	"encoding/hex"
	"sync/atomic"
	"time"
)

// NonceGenerator generates unique nonces for replay protection.
type NonceGenerator struct {
	// Base timestamp to ensure uniqueness across restarts
	baseTime int64

	// Counter for multiple nonces within same millisecond
	counter atomic.Uint64
}

// NewNonceGenerator creates a new nonce generator.
func NewNonceGenerator() *NonceGenerator {
	return &NonceGenerator{
		baseTime: time.Now().UnixNano(),
	}
}

// Generate generates a unique nonce string.
// Format: <timestamp_ms>_<counter>_<random>
func (ng *NonceGenerator) Generate() string {
	timestamp := time.Now().UnixMilli()
	counter := ng.counter.Add(1)

	// Generate 4 bytes of randomness
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	random := hex.EncodeToString(randomBytes)

	return formatNonce(timestamp, counter, random)
}

// GenerateInt64 generates a unique nonce as int64.
// Uses timestamp + counter for guaranteed uniqueness.
func (ng *NonceGenerator) GenerateInt64() int64 {
	timestamp := time.Now().UnixMilli()
	counter := ng.counter.Add(1)

	// Combine timestamp (top bits) with counter (bottom bits)
	// Counter fits in 16 bits (max 65535 per millisecond)
	return (timestamp << 16) | int64(counter&0xFFFF)
}

// formatNonce formats the nonce components into a string.
func formatNonce(timestamp int64, counter uint64, random string) string {
	// Use a simple format that's URL-safe
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"

	buf := make([]byte, 0, 24)

	// Encode timestamp
	tmp := timestamp
	for tmp > 0 {
		buf = append(buf, digits[tmp%36])
		tmp /= 36
	}

	buf = append(buf, '_')

	// Encode counter (lower 16 bits)
	ctr := counter & 0xFFFF
	for ctr > 0 {
		buf = append(buf, digits[ctr%36])
		ctr /= 36
	}

	buf = append(buf, '_')

	// Append random
	buf = append(buf, random...)

	return string(buf)
}

// TimestampNonce generates a simple timestamp-based nonce.
// Suitable for Binance which uses timestamp in milliseconds.
func TimestampNonce() int64 {
	return time.Now().UnixMilli()
}

package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

const (
	// DefaultRecvWindow is the default recvWindow for signed requests (5 seconds)
	// Documentation: https://binance-docs.github.io/apidocs/spot/en/#signed-trade-and-user_data-endpoint-security
	DefaultRecvWindow = 5000
	// MaxRecvWindow is the maximum allowed recvWindow (60 seconds)
	MaxRecvWindow = 60000
)

// Signer handles HMAC-SHA256 signing for Binance API requests.
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#signed-trade-and-user_data-endpoint-security
// Version: API v3 (verified 2026-02-16)
//
// Authentication method:
//   - API key goes in X-MBX-APIKEY header (NOT in query params)
//   - Signature is HMAC-SHA256 of the query string (including timestamp and recvWindow)
//   - Timestamp must be in milliseconds
type Signer struct {
	apiKey     string
	apiSecret  string
	recvWindow int64
}

// NewSigner creates a new Signer for Binance API authentication.
// If recvWindow is 0, DefaultRecvWindow (5000ms) is used.
// If recvWindow exceeds MaxRecvWindow (60000ms), MaxRecvWindow is used.
func NewSigner(apiKey, apiSecret string, recvWindow int64) *Signer {
	if recvWindow <= 0 {
		recvWindow = DefaultRecvWindow
	} else if recvWindow > MaxRecvWindow {
		recvWindow = MaxRecvWindow
	}

	return &Signer{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		recvWindow: recvWindow,
	}
}

// Sign adds timestamp and recvWindow to params, then computes HMAC-SHA256 signature.
// Returns the timestamp used (milliseconds) and the signature.
//
// The signature is computed over the URL-encoded query string.
// Params are encoded using url.Values.Encode() which sorts keys alphabetically.
//
// Example:
//
//	params := url.Values{}
//	params.Set("symbol", "BTCUSDT")
//	params.Set("side", "BUY")
//	timestamp, signature := signer.Sign(params)
//	// params now contains: symbol=BTCUSDT&side=BUY&timestamp=1234567890123&recvWindow=5000
//	// signature is HMAC-SHA256 of "recvWindow=5000&side=BUY&symbol=BTCUSDT&timestamp=1234567890123"
func (s *Signer) Sign(params url.Values) (timestamp int64, signature string) {
	// Add timestamp in milliseconds
	timestamp = time.Now().UnixMilli()
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))

	// Add recvWindow
	params.Set("recvWindow", strconv.FormatInt(s.recvWindow, 10))

	// Encode params (sorted alphabetically by Encode())
	queryString := params.Encode()

	// Compute HMAC-SHA256 signature
	signature = s.SignString(queryString)

	return timestamp, signature
}

// SignString computes HMAC-SHA256 signature of the given query string.
// The query string should NOT include the signature parameter.
func (s *Signer) SignString(queryString string) string {
	mac := hmac.New(sha256.New, []byte(s.apiSecret))
	mac.Write([]byte(queryString))
	return hex.EncodeToString(mac.Sum(nil))
}

// APIKey returns the API key for X-MBX-APIKEY header.
// IMPORTANT: Never log or expose the API key in production.
func (s *Signer) APIKey() string {
	return s.apiKey
}

// RecvWindow returns the configured recvWindow in milliseconds.
func (s *Signer) RecvWindow() int64 {
	return s.recvWindow
}

// ValidateCredentials checks if the signer has valid credentials.
// Returns an error if apiKey or apiSecret is empty.
func (s *Signer) ValidateCredentials() error {
	if s.apiKey == "" {
		return fmt.Errorf("binance: API key is required")
	}
	if s.apiSecret == "" {
		return fmt.Errorf("binance: API secret is required")
	}
	return nil
}

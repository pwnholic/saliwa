package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lilwiggy/ex-act/internal/ratelimit"
	"github.com/lilwiggy/ex-act/pkg/domain"
	"github.com/lilwiggy/ex-act/pkg/errors"
	"resty.dev/v3"
)

// RESTClient provides authenticated, rate-limited REST communication with Binance API.
// Documentation: https://binance-docs.github.io/apidocs/spot/en/
// API Version: v3 (verified 2026-02-16)
//
// Features:
//   - Automatic HMAC-SHA256 signing for authenticated requests
//   - Weight-based rate limiting with server header tracking
//   - Context-based timeouts (not http.Client.Timeout)
//   - Automatic weight tracking via X-MBX-USED-WEIGHT-* headers
//
// IMPORTANT: resty v3 requires calling Close() when done (breaking change from v2)
type RESTClient struct {
	client      *resty.Client
	baseURL     string
	signer      *Signer
	rateLimiter *ratelimit.WeightedLimiter
	config      Config

	// Server time offset for clock synchronization
	timeOffset   time.Duration
	timeOffsetMu sync.RWMutex

	// Track if client is closed
	closed   bool
	closedMu sync.RWMutex
}

// Config contains configuration for the Binance REST client.
type Config struct {
	// BaseURL is the API base URL (defaults to production)
	BaseURL string
	// APIKey is the Binance API key (required for authenticated requests)
	APIKey string
	// APISecret is the Binance API secret (required for authenticated requests)
	APISecret string
	// Timeout is the request timeout (default: 10 seconds)
	Timeout time.Duration
	// MaxWeight is the maximum rate limit weight per minute (default: 1200)
	MaxWeight int
	// RecvWindow is the recvWindow for signed requests in milliseconds (default: 5000)
	RecvWindow int64
	// Testnet enables testnet mode (changes base URL)
	Testnet bool
}

// NewRESTClient creates a new Binance REST client with middleware.
// IMPORTANT: resty v3 requires calling Close() when done.
//
// Example:
//
//	cfg := binance.Config{
//	    APIKey:    "your-api-key",
//	    APISecret: "your-api-secret",
//	    Testnet:   true,
//	}
//	client, err := binance.NewRESTClient(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close() // REQUIRED in resty v3
func NewRESTClient(cfg Config) (*RESTClient, error) {
	// Set defaults
	if cfg.BaseURL == "" {
		if cfg.Testnet {
			cfg.BaseURL = TestnetRestURL
		} else {
			cfg.BaseURL = BaseRestURL
		}
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.MaxWeight == 0 {
		cfg.MaxWeight = ratelimit.DefaultMaxWeight
	}
	if cfg.RecvWindow == 0 {
		cfg.RecvWindow = DefaultRecvWindow
	}

	// Create signer if credentials provided
	var signer *Signer
	if cfg.APIKey != "" && cfg.APISecret != "" {
		signer = NewSigner(cfg.APIKey, cfg.APISecret, cfg.RecvWindow)
		if err := signer.ValidateCredentials(); err != nil {
			return nil, err
		}
	}

	// Create rate limiter
	rateLimiter := ratelimit.NewWeightedLimiter(cfg.MaxWeight)

	// Create resty client
	client := resty.New()
	client.SetBaseURL(cfg.BaseURL)

	// Set user agent
	client.SetHeader("User-Agent", "ex-act/1.0")
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("Accept", "application/json")

	// Set API key header if available
	if signer != nil {
		client.SetHeader("X-MBX-APIKEY", signer.APIKey())
	}

	// Setup middleware
	rc := &RESTClient{
		client:      client,
		baseURL:     cfg.BaseURL,
		signer:      signer,
		rateLimiter: rateLimiter,
		config:      cfg,
	}

	rc.setupMiddleware()

	return rc, nil
}

// setupMiddleware configures request/response middleware.
func (rc *RESTClient) setupMiddleware() {
	// AddRequestMiddleware: Rate limiting and signing
	rc.client.AddRequestMiddleware(func(c *resty.Client, req *resty.Request) error {
		// Check if client is closed
		rc.closedMu.RLock()
		if rc.closed {
			rc.closedMu.RUnlock()
			return fmt.Errorf("binance: client is closed")
		}
		rc.closedMu.RUnlock()

		// Get endpoint weight
		endpoint := req.URL
		weight := getEndpointWeight(endpoint)

		// Wait for rate limit (blocking)
		ctx := req.Context()
		if err := rc.rateLimiter.Wait(ctx, weight); err != nil {
			return fmt.Errorf("binance: rate limit wait failed: %w", err)
		}

		// Sign the request if it needs authentication
		if rc.signer != nil && needsSigning(endpoint) {
			// Sign the request
			timestamp, signature := rc.signer.Sign(req.QueryParams)

			// Add signed params to request
			req.SetQueryParam("timestamp", strconv.FormatInt(timestamp, 10))
			req.SetQueryParam("recvWindow", strconv.FormatInt(rc.signer.RecvWindow(), 10))
			req.SetQueryParam("signature", signature)
		}

		return nil
	})

	// AddResponseMiddleware: Weight tracking and error handling
	rc.client.AddResponseMiddleware(func(c *resty.Client, resp *resty.Response) error {
		// Track weight from response headers
		rc.trackWeightFromHeaders(resp.Header())
		return nil
	})
}

// trackWeightFromHeaders extracts and tracks weight from X-MBX-USED-WEIGHT-* headers.
func (rc *RESTClient) trackWeightFromHeaders(header http.Header) {
	// X-MBX-USED-WEIGHT-1M for 1-minute weight
	weightStr := header.Get("X-MBX-USED-WEIGHT-1m")
	if weightStr == "" {
		weightStr = header.Get("X-MBX-USED-WEIGHT-1M")
	}
	if weightStr != "" {
		if weight, err := strconv.Atoi(weightStr); err == nil {
			rc.rateLimiter.UpdateWeight(weight)
		}
	}
}

// getEndpointWeight returns the weight for an endpoint.
// Handles both full URLs and path-only endpoints.
func getEndpointWeight(endpoint string) int {
	// Extract path from full URL if needed
	if strings.HasPrefix(endpoint, "http") {
		// Find the path part after the base URL
		idx := strings.Index(endpoint, "/api/")
		if idx != -1 {
			endpoint = endpoint[idx:]
		}
	}
	return GetEndpointWeight(endpoint)
}

// needsSigning determines if an endpoint requires authentication.
// Most /api/v3/* endpoints are public, but user data and trading require signing.
func needsSigning(endpoint string) bool {
	// Public endpoints that don't need signing
	publicEndpoints := []string{
		EPing,
		ETime,
		EExchangeInfo,
		EDepth,
		ETrades,
		ETicker,
		ETickerPrice,
		ETickerBook,
	}

	// Check if endpoint is public
	for _, pub := range publicEndpoints {
		if strings.Contains(endpoint, pub) {
			return false
		}
	}

	// All other endpoints need signing
	return true
}

// Close releases resources used by the client.
// REQUIRED by resty v3 - must be called when done with the client.
func (rc *RESTClient) Close() {
	rc.closedMu.Lock()
	rc.closed = true
	rc.closedMu.Unlock()
	rc.client.Close()
}

// Ping tests connectivity to the Binance API.
// API: GET /api/v3/ping
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#test-connectivity
// Weight: 1
func (rc *RESTClient) Ping(ctx context.Context) error {
	_, err := rc.client.R().
		SetContext(ctx).
		Get(EPing)
	return err
}

// GetServerTime returns the current server time.
// API: GET /api/v3/time
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#check-server-time
// Weight: 1
func (rc *RESTClient) GetServerTime(ctx context.Context) (int64, error) {
	var result struct {
		ServerTime int64 `json:"serverTime"`
	}

	resp, err := rc.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get(ETime)
	if err != nil {
		return 0, err
	}

	if !resp.IsSuccess() {
		return 0, rc.handleErrorResponse(resp)
	}

	return result.ServerTime, nil
}

// GetExchangeInfo returns exchange information including rate limits and symbol info.
// API: GET /api/v3/exchangeInfo
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#exchange-information
// Weight: 20
func (rc *RESTClient) GetExchangeInfo(ctx context.Context) (*ExchangeInfo, error) {
	var result ExchangeInfo

	resp, err := rc.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get(EExchangeInfo)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, rc.handleErrorResponse(resp)
	}

	return &result, nil
}

// GetServerTimeOffset returns the time offset between local and server time.
// This is useful for ensuring requests don't fail due to clock skew.
// Call this after getting server time.
func (rc *RESTClient) GetServerTimeOffset() time.Duration {
	rc.timeOffsetMu.RLock()
	defer rc.timeOffsetMu.RUnlock()
	return rc.timeOffset
}

// SyncTime synchronizes local time with server time.
// This should be called periodically to prevent timestamp-related errors.
func (rc *RESTClient) SyncTime(ctx context.Context) error {
	localBefore := time.Now().UnixMilli()
	serverTime, err := rc.GetServerTime(ctx)
	if err != nil {
		return err
	}
	localAfter := time.Now().UnixMilli()

	// Calculate offset using midpoint of local times
	localMid := (localBefore + localAfter) / 2
	offset := time.Duration(serverTime-localMid) * time.Millisecond

	rc.timeOffsetMu.Lock()
	rc.timeOffset = offset
	rc.timeOffsetMu.Unlock()

	return nil
}

// handleErrorResponse converts HTTP error responses to typed errors.
func (rc *RESTClient) handleErrorResponse(resp *resty.Response) error {
	statusCode := resp.StatusCode()

	// Read body for error message
	var bodyBytes []byte
	if resp.Body != nil {
		bodyBytes, _ = io.ReadAll(resp.Body)
	}
	body := string(bodyBytes)

	// Parse Binance error format
	var binanceErr struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(bodyBytes, &binanceErr); err == nil && binanceErr.Msg != "" {
		return rc.createBinanceError(statusCode, binanceErr.Code, binanceErr.Msg)
	}

	// Generic HTTP error
	return errors.NewConnectionError("binance", resp.Request.URL, fmt.Sprintf("HTTP %d: %s", statusCode, body), false)
}

// createBinanceError creates an appropriate error type based on Binance error codes.
func (rc *RESTClient) createBinanceError(httpStatus, code int, msg string) error {
	// Rate limit errors
	if code == -1015 || code == -1016 || httpStatus == http.StatusTooManyRequests {
		retryAfter := 1 * time.Second // Default
		return errors.NewRateLimitError("binance", retryAfter, 1)
	}

	// Authentication errors
	if code == -2015 || code == -1022 || httpStatus == http.StatusUnauthorized {
		return fmt.Errorf("binance: authentication failed: %s", msg)
	}

	// Invalid request
	if code == -1100 || code == -1101 || code == -1102 || code == -1103 {
		return errors.NewValidationError("request", nil, msg)
	}

	// Generic error
	return fmt.Errorf("binance: error code %d: %s", code, msg)
}

// ExchangeInfo represents the exchange information response.
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#exchange-information
type ExchangeInfo struct {
	Timezone        string       `json:"timezone"`
	ServerTime      int64        `json:"serverTime"`
	RateLimits      []RateLimit  `json:"rateLimits"`
	ExchangeFilters []any        `json:"exchangeFilters"`
	Symbols         []SymbolInfo `json:"symbols"`
}

// RateLimit represents a rate limit from exchange info.
type RateLimit struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	IntervalNum   int    `json:"intervalNum"`
	Limit         int    `json:"limit"`
}

// SymbolInfo represents symbol information.
type SymbolInfo struct {
	Symbol                   string           `json:"symbol"`
	Status                   string           `json:"status"`
	BaseAsset                string           `json:"baseAsset"`
	BaseAssetPrecision       int              `json:"baseAssetPrecision"`
	QuoteAsset               string           `json:"quoteAsset"`
	QuotePrecision           int              `json:"quotePrecision"`
	QuoteAssetPrecision      int              `json:"quoteAssetPrecision"`
	BaseCommissionPrecision  int              `json:"baseCommissionPrecision"`
	QuoteCommissionPrecision int              `json:"quoteCommissionPrecision"`
	OrderTypes               []string         `json:"orderTypes"`
	IcebergAllowed           bool             `json:"icebergAllowed"`
	OcoAllowed               bool             `json:"ocoAllowed"`
	OtoAllowed               bool             `json:"otoAllowed"`
	SpotTradingAllowed       bool             `json:"spotTradingAllowed"`
	MarginTradingAllowed     bool             `json:"marginTradingAllowed"`
	Filters                  []map[string]any `json:"filters"`
	Permissions              []string         `json:"permissions"`
}

// GetAccount returns account information.
// API: GET /api/v3/account (HMAC SHA256)
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#account-information-user_data
// Weight: 20
func (rc *RESTClient) GetAccount(ctx context.Context) (*AccountInfo, error) {
	if rc.signer == nil {
		return nil, fmt.Errorf("binance: API credentials required for GetAccount")
	}

	var result AccountInfo

	resp, err := rc.client.R().
		SetContext(ctx).
		SetResult(&result).
		Get(EAccount)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, rc.handleErrorResponse(resp)
	}

	return &result, nil
}

// AccountInfo represents account information.
type AccountInfo struct {
	MakerCommission  int64     `json:"makerCommission"`
	TakerCommission  int64     `json:"takerCommission"`
	BuyerCommission  int64     `json:"buyerCommission"`
	SellerCommission int64     `json:"sellerCommission"`
	CanTrade         bool      `json:"canTrade"`
	CanWithdraw      bool      `json:"canWithdraw"`
	CanDeposit       bool      `json:"canDeposit"`
	UpdateTime       int64     `json:"updateTime"`
	Balances         []Balance `json:"balances"`
}

// Balance represents an asset balance.
type Balance struct {
	Asset  string         `json:"asset"`
	Free   domain.Decimal `json:"free"`
	Locked domain.Decimal `json:"locked"`
}

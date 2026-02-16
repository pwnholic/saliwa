# Phase 2: Bybit Infrastructure - Research

**Researched:** 2026-02-16
**Domain:** Bybit API v5 (Spot Trading)
**Confidence:** HIGH

## Summary

Bybit API v5 is a unified API for spot, derivatives, and options trading. Unlike Binance's weight-based rate limiting, Bybit uses per-second request limiting with different limits per endpoint category. Authentication differs significantly from Binance: Bybit uses `X-BAPI-*` headers and concatenates timestamp+apiKey+recvWindow+payload for signing, whereas Binance uses `X-MBX-APIKEY` header and signs only the query string.

The Bybit WebSocket uses a different subscription model with explicit `op: "subscribe"` JSON messages and supports dynamic subscribe/unsubscribe without reconnection (unlike Binance which requires reconnection for new streams). Ping/pong is JSON-based (`{"op": "ping"}`) rather than WebSocket protocol level.

**Primary recommendation:** Implement Bybit driver with same patterns as Binance but with per-second rate limiting and different authentication format. Reuse gws WebSocket client, resty HTTP client, and same reconnection/heartbeat patterns.

## Standard Stack

The established libraries for Bybit integration (aligned with Phase 1):

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| resty | v3.0.0-beta.6 | HTTP client | Already in use for Binance |
| gws | v1.8.9 | WebSocket client | Already in use for Binance |
| apd | v3.2.1 | Decimal arithmetic | Already in use (domain.Decimal alias) |
| golang.org/x/time/rate | - | Rate limiting | Already in use for rate limiting |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| zerolog | - | Structured logging | Consistent with Binance driver |
| gobreaker | - | Circuit breaker | Same reliability pattern |

**Installation:**
No new dependencies required. All libraries already in use from Phase 1.

## Architecture Patterns

### Recommended Project Structure (Mirror Binance)
```
internal/driver/bybit/
├── rest_client.go      # REST client with auth and rate limiting
├── ws_client.go        # WebSocket client with reconnection
├── ws_messages.go      # WebSocket message types
├── subscription.go     # Subscription manager (reused pattern)
├── signer.go           # Bybit-specific signature generation
└── urls.go             # API endpoints and URLs
```

### Authentication Pattern (Bybit-Specific)
**What:** Bybit uses different auth format from Binance
**When to use:** All authenticated REST endpoints

**Key Differences from Binance:**
| Aspect | Binance | Bybit |
|--------|---------|-------|
| API Key Header | `X-MBX-APIKEY` | `X-BAPI-API-KEY` |
| Timestamp Header | In query params | `X-BAPI-TIMESTAMP` |
| Signature Header | In query params | `X-BAPI-SIGN` |
| RecvWindow Header | In query params | `X-BAPI-RECV-WINDOW` |
| Signature Format | HMAC(queryString) | HMAC(timestamp+apiKey+recvWindow+payload) |

```go
// Source: Bybit API v5 Documentation
// https://bybit-exchange.github.io/docs/v5/guide

// Sign creates Bybit signature
// GET: timestamp + apiKey + recvWindow + queryString
// POST: timestamp + apiKey + recvWindow + jsonBodyString
func (s *Signer) Sign(method string, queryString string, body string) (timestamp int64, signature string) {
    timestamp = time.Now().UnixMilli()
    
    var stringToSign string
    if method == "GET" {
        stringToSign = fmt.Sprintf("%d%s%s%s", timestamp, s.apiKey, s.recvWindow, queryString)
    } else {
        stringToSign = fmt.Sprintf("%d%s%s%s", timestamp, s.apiKey, s.recvWindow, body)
    }
    
    mac := hmac.New(sha256.New, []byte(s.apiSecret))
    mac.Write([]byte(stringToSign))
    signature = hex.EncodeToString(mac.Sum(nil))
    
    return timestamp, signature
}

// Request headers for authenticated endpoints
// X-BAPI-API-KEY: Your API key
// X-BAPI-TIMESTAMP: UTC timestamp in milliseconds
// X-BAPI-SIGN: Signature from above
// X-BAPI-RECV-WINDOW: Validity period (default: 5000)
```

### Rate Limiting Pattern (Per-Second Based)
**What:** Bybit uses per-second limits, not weight-based like Binance
**When to use:** All REST API calls

```go
// Source: Bybit API v5 Documentation
// https://bybit-exchange.github.io/docs/v5/rate-limit

// Bybit Spot Rate Limits (Classic/UTA accounts)
var SpotRateLimits = map[string]int{
    // Order endpoints
    "/v5/order/create":       20, // 20 requests per second
    "/v5/order/amend":        20,
    "/v5/order/cancel":       20,
    "/v5/order/cancel-all":   20,
    "/v5/order/create-batch": 20,
    
    // Query endpoints
    "/v5/order/realtime":     50,
    "/v5/order/history":      50,
    "/v5/execution/list":     50,
    
    // Account endpoints
    "/v5/account/wallet-balance": 10,
    "/v5/position/list":          10,
}

// Response headers for rate limit tracking
// X-Bapi-Limit-Status: Remaining requests for current endpoint
// X-Bapi-Limit: Current limit for current endpoint
// X-Bapi-Limit-Reset-Timestamp: When limit resets
```

### WebSocket Subscription Pattern
**What:** Explicit JSON-based subscribe/unsubscribe
**When to use:** All WebSocket streams

```go
// Source: Bybit API v5 Documentation
// https://bybit-exchange.github.io/docs/v5/ws/connect

// WebSocket subscription format
type WSSubscription struct {
    ReqID string   `json:"req_id,omitempty"` // Optional
    Op    string   `json:"op"`               // "subscribe" or "unsubscribe"
    Args  []string `json:"args"`             // Topics to subscribe
}

// Example subscription message
// {
//     "op": "subscribe",
//     "args": ["orderbook.50.BTCUSDT", "publicTrade.BTCUSDT"]
// }

// Public topics for spot
// orderbook.{depth}.{symbol}  - e.g., "orderbook.50.BTCUSDT"
// publicTrade.{symbol}        - e.g., "publicTrade.BTCUSDT"
// tickers.{symbol}            - e.g., "tickers.BTCUSDT"
```

### Anti-Patterns to Avoid
- **Using Binance auth format for Bybit:** Headers and signature generation are completely different
- **Weight-based rate limiting:** Bybit uses per-second, not weight
- **WebSocket reconnection for new subscriptions:** Bybit supports dynamic subscribe
- **Binary WebSocket ping/pong:** Bybit uses JSON `{"op": "ping"}`

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Rate limiting | Custom per-second limiter | `golang.org/x/time/rate` | Token bucket already handles burst smoothing |
| WebSocket reconnection | Custom logic | Binance pattern (ws_client.go) | Same exponential backoff with jitter |
| Decimal arithmetic | float64 | `domain.Decimal` (apd) | Precision loss in financial data |
| Signature generation | From scratch | Pattern from binance/signer.go | Same HMAC-SHA256, different string format |
| Subscription tracking | Custom map | Binance SubscriptionManager | Same pattern, different stream names |

**Key insight:** Bybit driver should be a near-copy of Binance driver with:
1. Different signer (concatenated string vs query string only)
2. Different rate limiter (per-second vs per-weight)
3. Different headers (`X-BAPI-*` vs `X-MBX-*`)
4. Dynamic WebSocket subscription support

## Common Pitfalls

### Pitfall 1: Wrong Signature Format
**What goes wrong:** Using Binance signature format (HMAC of query string only)
**Why it happens:** Similar API patterns lead to assumption of identical auth
**How to avoid:** Remember the Bybit formula: `timestamp + apiKey + recvWindow + payload`
**Warning signs:** `retCode: 10004` (Error sign)

### Pitfall 2: Rate Limit Confusion
**What goes wrong:** Implementing weight-based limiting like Binance
**Why it happens:** Both are rate limiting concepts
**How to avoid:** Bybit is per-second per-endpoint, not per-weight per-minute
**Warning signs:** `retCode: 10006` (Too many visits)

### Pitfall 3: WebSocket Subscription Timing
**What goes wrong:** Sending subscription before connection is confirmed
**Why it happens:** Eager subscription sending
**How to avoid:** Wait for connection confirmation before subscribing
**Warning signs:** No data received, no error messages

### Pitfall 4: Case Sensitivity in Symbols
**What goes wrong:** Using lowercase symbol names in REST requests
**Why it happens:** Binance uses lowercase for WebSocket streams
**How to avoid:** Bybit requires uppercase symbols in REST (e.g., "BTCUSDT")
**Warning signs:** `retCode: 10001` (Request parameter error)

### Pitfall 5: Missing Category Parameter
**What goes wrong:** Calling endpoints without `category=spot`
**Why it happens:** Binance doesn't have category concept
**How to avoid:** All spot endpoints require `category=spot`
**Warning signs:** Empty results or wrong data

## Code Examples

### Bybit Signer Implementation
```go
// Source: Bybit API v5 Documentation
// https://bybit-exchange.github.io/docs/v5/guide

package bybit

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
)

const (
    DefaultRecvWindow = 5000
    MaxRecvWindow     = 60000
)

type Signer struct {
    apiKey     string
    apiSecret  string
    recvWindow int64
}

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

// Sign generates signature for Bybit API
// GET: timestamp + apiKey + recvWindow + queryString
// POST: timestamp + apiKey + recvWindow + jsonBodyString
func (s *Signer) Sign(method, queryString, body string) (timestamp int64, signature string) {
    timestamp = time.Now().UnixMilli()
    
    var stringToSign string
    if method == "GET" {
        stringToSign = fmt.Sprintf("%d%s%d%s", timestamp, s.apiKey, s.recvWindow, queryString)
    } else {
        stringToSign = fmt.Sprintf("%d%s%d%s", timestamp, s.apiKey, s.recvWindow, body)
    }
    
    mac := hmac.New(sha256.New, []byte(s.apiSecret))
    mac.Write([]byte(stringToSign))
    signature = hex.EncodeToString(mac.Sum(nil))
    
    return timestamp, signature
}
```

### REST Client Configuration
```go
// Source: Bybit API v5 Documentation
// https://bybit-exchange.github.io/docs/v5/guide

package bybit

// Bybit API URLs
const (
    BaseRestURL     = "https://api.bybit.com"
    TestnetRestURL  = "https://api-testnet.bybit.com"
    
    BaseWebSocketURL     = "wss://stream.bybit.com/v5/public/spot"
    TestnetWebSocketURL  = "wss://stream-testnet.bybit.com/v5/public/spot"
)

// REST endpoints
const (
    // Market data
    ETime           = "/v5/market/time"
    ETickers        = "/v5/market/tickers"
    EOrderbook      = "/v5/market/orderbook"
    EInstruments    = "/v5/market/instruments-info"
    
    // Trading
    ECreateOrder    = "/v5/order/create"
    ECancelOrder    = "/v5/order/cancel"
    EQueryOrder     = "/v5/order/realtime"
    
    // Account
    EWalletBalance  = "/v5/account/wallet-balance"
)

// Rate limits per endpoint (requests per second)
var EndpointRateLimits = map[string]int{
    ECreateOrder:    20,
    ECancelOrder:    20,
    EQueryOrder:     50,
    EWalletBalance:  10,
    ETickers:        10,
    EOrderbook:      10,
}
```

### WebSocket Client Pattern
```go
// Source: Bybit API v5 Documentation
// https://bybit-exchange.github.io/docs/v5/ws/connect

package bybit

// WSSubscription for subscribe/unsubscribe operations
type WSSubscription struct {
    ReqID string   `json:"req_id,omitempty"`
    Op    string   `json:"op"`    // "subscribe" or "unsubscribe"
    Args  []string `json:"args"`  // Topics
}

// WSPing for heartbeat
type WSPing struct {
    Op string `json:"op"` // "ping"
}

// WSPong response
type WSPong struct {
    Success bool   `json:"success"`
    RetMsg  string `json:"ret_msg"`
    Op      string `json:"op"`
    ConnID  string `json:"conn_id"`
}

// Subscribe sends subscription request
func (c *WSClient) Subscribe(topics []string) error {
    sub := WSSubscription{
        Op:   "subscribe",
        Args: topics,
    }
    data, _ := json.Marshal(sub)
    return c.conn.WriteMessage(websocket.TextMessage, data)
}

// SendPing sends heartbeat
func (c *WSClient) SendPing() error {
    ping := WSPing{Op: "ping"}
    data, _ := json.Marshal(ping)
    return c.conn.WriteMessage(websocket.TextMessage, data)
}

// Stream topics for spot
// orderbook.50.BTCUSDT  - Order book with 50 levels
// publicTrade.BTCUSDT   - Public trades
// tickers.BTCUSDT       - Ticker updates
```

### WebSocket Message Types
```go
// Source: Bybit API v5 Documentation
// https://bybit-exchange.github.io/docs/v5/websocket/public/orderbook

// WSMessage is the base WebSocket message
type WSMessage struct {
    Topic string          `json:"topic"`
    Type  string          `json:"type"`  // "snapshot" or "delta"
    TS    int64           `json:"ts"`    // Timestamp
    Data  json.RawMessage `json:"data"`
}

// WSOrderbookData for orderbook updates
type WSOrderbookData struct {
    S   string     `json:"s"`   // Symbol
    B   [][]string `json:"b"`   // Bids [[price, size], ...]
    A   [][]string `json:"a"`   // Asks [[price, size], ...]
    U   int64      `json:"u"`   // Update ID
    Seq int64      `json:"seq"` // Cross sequence
}

// WSTradeData for public trades
type WSTradeData struct {
    T   int64  `json:"T"`   // Trade timestamp
    S   string `json:"s"`   // Symbol
    Sd  string `json:"S"`   // Side (Buy/Sell)
    V   string `json:"v"`   // Size
    P   string `json:"p"`   // Price
    I   string `json:"i"`   // Trade ID
    BT  bool   `json:"BT"`  // Block trade
}

// WSTickerData for ticker updates
type WSTickerData struct {
    Symbol      string `json:"symbol"`
    LastPrice   string `json:"lastPrice"`
    HighPrice   string `json:"highPrice24h"`
    LowPrice    string `json:"lowPrice24h"`
    Volume24h   string `json:"volume24h"`
    Turnover24h string `json:"turnover24h"`
    Bid1Price   string `json:"bid1Price"`
    Bid1Size    string `json:"bid1Size"`
    Ask1Price   string `json:"ask1Price"`
    Ask1Size    string `json:"ask1Size"`
}
```

### Error Response Handling
```go
// Source: Bybit API v5 Documentation
// https://bybit-exchange.github.io/docs/v5/error

// APIResponse is the standard Bybit response
type APIResponse struct {
    RetCode    int             `json:"retCode"`
    RetMsg     string          `json:"retMsg"`
    Result     json.RawMessage `json:"result"`
    RetExtInfo json.RawMessage `json:"retExtInfo"`
    Time       int64           `json:"time"`
}

// Common error codes
const (
    ErrOK               = 0
    ErrRequestExpired   = -1
    ErrParamError       = 10001
    ErrTimeWindow       = 10002
    ErrInvalidKey       = 10003
    ErrSignError        = 10004
    ErrPermissionDenied = 10005
    ErrRateLimit        = 10006
    ErrAuthFailed       = 10007
    ErrIPBanned         = 10009
    ErrOrderNotExist    = 110001
    ErrInsufficientBal  = 110004
)

// Map Bybit errors to domain errors
func mapError(retCode int, retMsg string) error {
    switch retCode {
    case ErrRateLimit:
        return errors.NewRateLimitError("bybit", 1*time.Second, 1)
    case ErrInvalidKey, ErrAuthFailed:
        return fmt.Errorf("bybit: authentication failed: %s", retMsg)
    case ErrSignError:
        return fmt.Errorf("bybit: signature error: %s", retMsg)
    case ErrParamError:
        return errors.NewValidationError("request", nil, retMsg)
    default:
        return fmt.Errorf("bybit: error code %d: %s", retCode, retMsg)
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| V3 API | V5 Unified API | 2023 | Single API for all products |
| Weight-based limits | Per-second limits | V5 | Simpler rate limiting |
| WebSocket frame ping | JSON ping | V5 | Language-agnostic heartbeat |

**Deprecated/outdated:**
- V3 API endpoints: Use V5 unified endpoints instead
- HMAC query string signing: Use concatenated format

## Error Code Mapping Table

| Bybit Code | Domain Error | Notes |
|------------|--------------|-------|
| 0 | nil | Success |
| -1 | ExchangeError | Request expired |
| 10001 | ValidationError | Request parameter error |
| 10002 | ExchangeError | Time window exceeded |
| 10003 | ExchangeError | Invalid API key |
| 10004 | ExchangeError | Signature error |
| 10005 | ExchangeError | Permission denied |
| 10006 | RateLimitError | Rate limit exceeded |
| 10007 | ExchangeError | Authentication failed |
| 10009 | ExchangeError | IP banned |
| 110001 | ExchangeError | Order does not exist |
| 110004 | ExchangeError | Insufficient balance |
| 110008 | ExchangeError | Order completed/cancelled |
| 110010 | ExchangeError | Order already cancelled |

## Open Questions

1. **Per-Endpoint Rate Limiter Implementation**
   - What we know: Bybit has different limits per endpoint (orders: 20/s, account: 10/s)
   - What's unclear: Whether to use single limiter with lowest limit or per-endpoint limiters
   - Recommendation: Start with single limiter at 10/s (most restrictive), optimize if needed

2. **UTA vs Classic Account Detection**
   - What we know: Rate limits differ between account types
   - What's unclear: How to detect account type programmatically
   - Recommendation: Default to Classic limits, add config option for UTA

## Sources

### Primary (HIGH confidence)
- `/websites/bybit-exchange_github_io_v5` - Official Bybit V5 API documentation
  - Authentication guide
  - REST endpoints reference
  - WebSocket streams reference
  - Rate limit rules
  - Error codes

### Secondary (MEDIUM confidence)
- Existing Binance driver code (internal/driver/binance/*)
  - Pattern reference for consistency
  - Rate limiter implementation
  - WebSocket client implementation

### Tertiary (LOW confidence)
- None - all critical information obtained from official docs

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Same as Phase 1, no new dependencies
- Architecture: HIGH - Based on official documentation and existing patterns
- Authentication: HIGH - Directly from Bybit docs with examples
- Rate limits: HIGH - Official documentation with per-endpoint values
- Pitfalls: MEDIUM - Based on API differences, practical experience may reveal more

**Research date:** 2026-02-16
**Valid until:** 30 days (stable API)

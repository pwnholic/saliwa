# Phase 1: Foundation & Binance Infrastructure - Research

**Researched:** 2026-02-16
**Domain:** Go actor framework, Binance API v3, WebSocket connectivity, REST client, decimal arithmetic, circuit breaker, rate limiting
**Confidence:** HIGH

## Summary

This phase establishes the foundational infrastructure for Binance exchange connectivity using the actor-based Ergo Framework. The research covers Binance API v3 specifics (REST and WebSocket), actor patterns for supervisors and message passing, WebSocket client implementation with gws, HTTP client patterns with resty, arbitrary-precision decimals with cockroachdb/apd, and resilience patterns with circuit breakers and rate limiters.

**Primary recommendation:** Implement a supervisor hierarchy with dedicated actors for REST client, WebSocket client, clock sync, and rate limiting. Use callback-based event delivery with auto-resubscription on reconnection. Implement weight-based rate limiting before the circuit breaker to prevent IP bans.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| ergo.services/ergo | v3.10+ | Actor framework | Erlang OTP patterns in Go, supervisors, fault tolerance |
| lxzan/gws | latest | WebSocket client | Fast, feature-rich, EventHandler interface |
| go-resty/resty | v3-beta | HTTP client | Chainable API, retry, middleware hooks |
| cockroachdb/apd | v3 | Decimal arithmetic | Arbitrary precision, General Decimal Arithmetic spec |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| sony/gobreaker | latest | Circuit breaker | Wrap all exchange API calls |
| golang.org/x/time/rate | latest | Token bucket rate limiter | Weight-based rate limiting |
| rs/zerolog | latest | Structured logging | All logging in library |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| ergo.services/ergo | protoactor-go | Ergo has better documentation, Erlang semantics |
| lxzan/gws | gorilla/websocket | gws is faster, has parallel processing |
| resty | standard http.Client | resty has built-in retry, middleware |

**Installation:**
```bash
go get ergo.services/ergo@latest
go get github.com/lxzan/gws@latest
go get github.com/go-resty/resty/v3@latest
go get github.com/cockroachdb/apd/v3@latest
go get github.com/sony/gobreaker@latest
go get golang.org/x/time/rate@latest
go get github.com/rs/zerolog@latest
```

## Architecture Patterns

### Recommended Project Structure
```
pkg/
├── connector/           # Public connector API
├── domain/              # Domain models (Order, Ticker, etc.)
├── config/              # Configuration types
├── errors/              # Typed errors
└── driver/              # Driver interface

internal/
├── app/                 # Ergo application setup
├── sup/                 # Supervisor definitions
│   ├── root.go          # RootSupervisor
│   └── exchange.go      # ExchangeSupervisor
├── actor/               # Core actors
│   ├── rest_client.go   # REST client actor
│   ├── ws_client.go     # WebSocket client actor
│   ├── clock_sync.go    # Clock synchronization actor
│   └── rate_limiter.go  # Rate limiter actor
├── driver/
│   └── binance/         # Binance implementation
│       ├── rest.go      # REST driver
│       ├── websocket.go # WebSocket driver
│       ├── signer.go    # HMAC-SHA256 signer
│       └── types.go     # Binance types
├── message/             # Internal messages
└── event/               # Event definitions
```

### Pattern 1: Supervisor Hierarchy

**What:** Each exchange has its own supervisor with child actors for REST, WebSocket, clock sync, and rate limiting.

**When to use:** All exchange connections must be supervised for fault tolerance.

**Example:**
```go
// Source: https://docs.ergo.services/actors/supervisor
func (s *ExchangeSupervisor) Init(args ...any) (act.SupervisorSpec, error) {
    exchange := args[0].(string)
    
    spec := act.SupervisorSpec{
        Name: gen.Atom(exchange + "_supervisor"),
        Restart: act.SupervisorRestart{
            Strategy: act.SupervisorStrategyTransient, // Restart on crash only
            Intensity: 5,  // Max 5 restarts
            Period:    10, // In 10 seconds
        },
        Children: []act.SupervisorChildSpec{
            {
                Name:    "rest_client",
                Factory: factoryRESTClient,
                Args:    []any{exchange},
            },
            {
                Name:    "ws_client",
                Factory: factoryWSClient,
                Args:    []any{exchange},
            },
            {
                Name:    "clock_sync",
                Factory: factoryClockSync,
                Args:    []any{exchange},
            },
            {
                Name:    "rate_limiter",
                Factory: factoryRateLimiter,
                Args:    []any{exchange},
            },
        },
    }
    return spec, nil
}
```

### Pattern 2: Actor Lifecycle with Async Start

**What:** Start() returns immediately with a ready channel for connection status.

**When to use:** All connector implementations.

**Example:**
```go
// Source: https://docs.ergo.services/actors/actor
type WSActor struct {
    act.Actor
    exchange   string
    url        string
    conn       *gws.Conn
    ready      chan struct{}
    callbacks  *Callbacks
}

func (ws *WSActor) Init(args ...any) error {
    ws.exchange = args[0].(string)
    ws.url = args[1].(string)
    ws.ready = make(chan struct{})
    
    // Start async connection
    go ws.connect()
    
    return nil
}

func (ws *WSActor) connect() {
    defer close(ws.ready)
    
    // Exponential backoff reconnection loop
    attempt := 0
    for {
        if err := ws.dial(); err == nil {
            ws.resubscribe()
            return
        }
        attempt++
        delay := ws.backoff(attempt)
        time.Sleep(delay)
    }
}

func (ws *WSActor) Terminate(reason error) {
    if ws.conn != nil {
        ws.conn.Close()
    }
}
```

### Pattern 3: WebSocket Event Handler (gws)

**What:** Implement gws.Event interface for connection lifecycle and message handling.

**When to use:** WebSocket client implementation.

**Example:**
```go
// Source: https://github.com/lxzan/gws
type WSEventHandler struct {
    client *WSClient
}

func (h *WSEventHandler) OnOpen(socket *gws.Conn) {
    _ = socket.SetDeadline(time.Now().Add(PingInterval + PingWait))
    h.client.onConnected()
}

func (h *WSEventHandler) OnClose(socket *gws.Conn, err error) {
    h.client.onDisconnected(err)
    // Trigger reconnection via actor message
}

func (h *WSEventHandler) OnPing(socket *gws.Conn, payload []byte) {
    _ = socket.SetDeadline(time.Now().Add(PingInterval + PingWait))
    _ = socket.WritePong(nil)
}

func (h *WSEventHandler) OnPong(socket *gws.Conn, payload []byte) {
    // Pong received, connection is alive
}

func (h *WSEventHandler) OnMessage(socket *gws.Conn, message *gws.Message) {
    defer message.Close()
    h.client.handleMessage(message.Bytes())
}
```

### Pattern 4: REST Client with Middleware (resty)

**What:** Use resty middleware for signing, rate limiting, and error handling.

**When to use:** REST client implementation.

**Example:**
```go
// Source: https://github.com/go-resty/docs
func NewRESTClient(cfg *config.RESTConfig, limiter *rate.Limiter, signer *Signer) *resty.Client {
    client := resty.New()
    client.SetBaseURL(cfg.BaseURL)
    client.SetTimeout(cfg.Timeout)
    
    // Signing middleware (before request)
    client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
        // Wait for rate limit
        if err := limiter.Wait(req.Context()); err != nil {
            return err
        }
        
        // Add timestamp
        timestamp := time.Now().UnixMilli()
        req.SetQueryParam("timestamp", strconv.FormatInt(timestamp, 10))
        
        // Add recvWindow
        req.SetQueryParam("recvWindow", "5000")
        
        // Sign the request
        queryString := req.QueryParam.Encode()
        signature := signer.Sign(queryString)
        req.SetQueryParam("signature", signature)
        
        // Set API key header
        req.SetHeader("X-MBX-APIKEY", signer.APIKey())
        
        return nil
    })
    
    // Error handling middleware (after response)
    client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
        if resp.StatusCode() == 429 {
            retryAfter := resp.Header().Get("Retry-After")
            return errors.NewRateLimitError(retryAfter)
        }
        if resp.StatusCode() >= 400 {
            var binanceErr BinanceError
            if err := json.Unmarshal(resp.Body(), &binanceErr); err == nil {
                return errors.NewExchangeError(binanceErr.Code, binanceErr.Msg)
            }
        }
        return nil
    })
    
    // Retry configuration
    client.SetRetryCount(3)
    client.SetRetryWaitTime(1 * time.Second)
    client.SetRetryMaxWaitTime(5 * time.Second)
    client.AddRetryCondition(func(r *resty.Response, err error) bool {
        return err != nil || r.StatusCode() >= 500
    })
    
    return client
}
```

### Pattern 5: Circuit Breaker (gobreaker)

**What:** Wrap exchange calls with circuit breaker for health monitoring.

**When to use:** All REST calls to exchange.

**Example:**
```go
// Source: https://github.com/sony/gobreaker
func NewCircuitBreaker(name string, cfg *config.CircuitBreakerConfig) *gobreaker.CircuitBreaker[any] {
    return gobreaker.NewCircuitBreaker[any](gobreaker.Settings{
        Name:        name,
        MaxRequests: 3,                        // Max requests in half-open state
        Interval:    10 * time.Second,         // Clear counts every 10s
        Timeout:     cfg.Timeout,              // Open -> Half-open timeout
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            // Trip after 5 consecutive failures
            return counts.ConsecutiveFailures >= 5
        },
        OnStateChange: func(name string, from, to gobreaker.State) {
            log.Info().
                Str("circuit", name).
                Str("from", from.String()).
                Str("to", to.String()).
                Msg("circuit breaker state changed")
        },
        IsSuccessful: func(err error) bool {
            // Rate limit errors don't count as failures
            return !errors.IsRetryable(err)
        },
    })
}

// Usage
func (rc *RESTClient) Do(req *resty.Request) (*resty.Response, error) {
    result, err := rc.circuitBreaker.Execute(func() (any, error) {
        return req.Send()
    })
    if err != nil {
        return nil, err
    }
    return result.(*resty.Response), nil
}
```

### Pattern 6: Decimal Arithmetic (apd)

**What:** All prices and quantities use apd.Decimal for arbitrary precision.

**When to use:** All financial calculations.

**Example:**
```go
// Source: https://github.com/cockroachdb/apd
package domain

import (
    "github.com/cockroachdb/apd/v3"
)

type Decimal = apd.Decimal

// Context with 34-digit precision (apd default)
var decimalContext = apd.BaseContext.WithPrecision(34)

// NewDecimal creates a decimal from string
func NewDecimal(s string) (*Decimal, error) {
    d := new(Decimal)
    _, _, err := d.SetString(s)
    return d, err
}

// MustDecimal creates a decimal from string (panics on error)
func MustDecimal(s string) *Decimal {
    d, err := NewDecimal(s)
    if err != nil {
        panic(err)
    }
    return d
}

// Mul multiplies two decimals
func Mul(a, b *Decimal) *Decimal {
    result := new(Decimal)
    _, _ = decimalContext.Mul(result, a, b)
    return result
}

// Compare returns -1, 0, or 1
func Compare(a, b *Decimal) int {
    return a.Cmp(b)
}

// IsZero checks if decimal is zero
func IsZero(d *Decimal) bool {
    return d.IsZero()
}
```

### Anti-Patterns to Avoid

- **Direct mutex in actors**: Actors should only communicate via message passing, never share state
- **Float64 for prices**: Always use apd.Decimal for financial data
- **Blocking Init()**: Never block in actor Init(), use goroutines for async work
- **Silent error handling**: All errors must be logged and propagated
- **Missing reconnection logic**: WebSocket must reconnect with exponential backoff

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Rate limiting | Custom token bucket | golang.org/x/time/rate | Battle-tested, handles edge cases |
| Circuit breaker | State machine from scratch | sony/gobreaker | Full state machine, callbacks |
| Decimal math | float64 operations | cockroachdb/apd | Precision loss, rounding errors |
| WebSocket reconnection | Custom loop | Actor supervision | Fault tolerance, restart policies |
| Request signing | Manual HMAC | resty middleware | Consistent, testable |
| Exponential backoff | Math.Pow calculations | Actor with state | Jitter, max delay |

**Key insight:** The combination of Ergo supervisors + gobreaker provides fault tolerance at both the process level and the API level. Don't duplicate this logic.

## Common Pitfalls

### Pitfall 1: Binance Rate Limit Violation

**What goes wrong:** IP gets banned (HTTP 418) for repeated rate limit violations.

**Why it happens:** Not respecting weight-based rate limits or not backing off on 429.

**How to avoid:**
- Track `X-MBX-USED-WEIGHT-*` headers from responses
- Implement weight-based rate limiter (not just request count)
- Back off immediately on 429 with `Retry-After` header
- Max 2400 weight per minute for IP limits

**Warning signs:**
- Multiple 429 responses in short time
- Increasing `X-MBX-USED-WEIGHT-*` values

### Pitfall 2: Clock Drift Causing Signature Failures

**What goes wrong:** Requests rejected with `-1022 INVALID_SIGNATURE` or `-1021 TIMESTAMP`.

**Why it happens:** Local clock drift >500ms from Binance server time.

**How to avoid:**
- Sync clock with `GET /api/v3/time` on startup
- Store offset and apply to all timestamps
- Re-sync periodically (every 5-10 minutes)
- Use `recvWindow` of 5000ms (default)

**Warning signs:**
- Increasing timestamp-related errors
- Offset growing over time

### Pitfall 3: WebSocket Connection State Desync

**What goes wrong:** Library thinks connected but exchange has disconnected.

**Why it happens:** TCP keepalive not detecting dead connections quickly enough.

**How to avoid:**
- Respond to ping frames within 1 minute (Binance sends every 20s)
- Implement zombie detection with REST ping
- Track last message received time
- Reconnect if no messages for >30 seconds

**Warning signs:**
- No messages received for extended period
- Ping/pong not working

### Pitfall 4: Subscription Loss on Reconnect

**What goes wrong:** WebSocket reconnects but old subscriptions are lost.

**Why it happens:** Not resubscribing after connection is re-established.

**How to avoid:**
- Track all active subscriptions in memory
- On successful reconnect, resubscribe to all tracked streams
- Use sync.Map for thread-safe subscription tracking

**Warning signs:**
- Missing market data updates after network blip
- Subscriptions list is empty after reconnect

### Pitfall 5: Decimal Precision Loss

**What goes wrong:** Price comparisons fail due to floating point errors.

**Why it happens:** Using float64 for prices instead of arbitrary precision.

**How to avoid:**
- Always parse exchange strings directly to apd.Decimal
- Never convert through float64
- Use apd.Context with 34-digit precision
- Compare with Cmp() method, not ==

**Warning signs:**
- Prices displaying incorrectly
- Order quantity mismatches

## Code Examples

### Binance Authentication (HMAC-SHA256)

```go
// Source: https://developers.binance.com/docs/binance-spot-api-docs/rest-api/request-security
type Signer struct {
    apiKey    string
    secretKey string
}

func NewSigner(apiKey, secretKey string) *Signer {
    return &Signer{apiKey: apiKey, secretKey: secretKey}
}

func (s *Signer) Sign(payload string) string {
    mac := hmac.New(sha256.New, []byte(s.secretKey))
    mac.Write([]byte(payload))
    return hex.EncodeToString(mac.Sum(nil))
}

func (s *Signer) APIKey() string {
    return s.apiKey
}
```

### Binance WebSocket Streams

```go
// Source: https://developers.binance.com/docs/binance-spot-api-docs/web-socket-streams
const (
    BaseStreamURL = "wss://stream.binance.com:9443"
    
    // Stream names
    StreamAggTrade   = "%s@aggTrade"
    StreamTrade      = "%s@trade"
    StreamKline      = "%s@kline_%s"
    StreamMiniTicker = "%s@miniTicker"
    StreamTicker     = "%s@ticker"
    StreamBookTicker = "%s@bookTicker"
    StreamDepth      = "%s@depth"
    StreamDepth100ms = "%s@depth@100ms"
)

// Subscription message
type SubscribeRequest struct {
    Method string   `json:"method"` // "SUBSCRIBE"
    Params []string `json:"params"` // ["btcusdt@aggTrade"]
    ID     int      `json:"id"`     // Request ID
}

// AggTrade payload
type WSAggTrade struct {
    EventType string `json:"e"` // "aggTrade"
    EventTime int64  `json:"E"` // Event time (ms)
    Symbol    string `json:"s"` // "BNBBTC"
    TradeID   int64  `json:"a"` // Aggregate trade ID
    Price     string `json:"p"` // "0.001"
    Quantity  string `json:"q"` // "100"
    FirstID   int64  `json:"f"` // First trade ID
    LastID    int64  `json:"l"` // Last trade ID
    TradeTime int64  `json:"T"` // Trade time (ms)
    IsBuyer   bool   `json:"m"` // Is buyer maker
}

// Depth update payload
type WSDepthUpdate struct {
    EventType     string     `json:"e"` // "depthUpdate"
    EventTime     int64      `json:"E"` // Event time (ms)
    Symbol        string     `json:"s"` // "BNBBTC"
    FirstUpdateID int64      `json:"U"` // First update ID
    FinalUpdateID int64      `json:"u"` // Final update ID
    Bids          [][]string `json:"b"` // [["0.0024","10"]]
    Asks          [][]string `json:"a"` // [["0.0026","100"]]
}
```

### Clock Synchronization

```go
// GET /api/v3/time returns {"serverTime": 1499827319559}
type ServerTime struct {
    ServerTime int64 `json:"serverTime"`
}

type ClockSyncActor struct {
    act.Actor
    client  *resty.Client
    offset  int64 // Server time - Local time (ms)
    mutex   sync.RWMutex
}

func (c *ClockSyncActor) Init(args ...any) error {
    c.client = args[0].(*resty.Client)
    
    // Initial sync
    if err := c.sync(); err != nil {
        return err
    }
    
    // Periodic sync every 5 minutes
    go c.periodicSync()
    
    return nil
}

func (c *ClockSyncActor) sync() error {
    localBefore := time.Now().UnixMilli()
    
    var serverTime ServerTime
    _, err := c.client.R().
        SetResult(&serverTime).
        Get("/api/v3/time")
    if err != nil {
        return err
    }
    
    localAfter := time.Now().UnixMilli()
    localTime := (localBefore + localAfter) / 2 // Midpoint
    
    c.mutex.Lock()
    c.offset = serverTime.ServerTime - localTime
    c.mutex.Unlock()
    
    return nil
}

func (c *ClockSyncActor) Now() int64 {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    return time.Now().UnixMilli() + c.offset
}

// Verify offset is <500ms (requirement RESL-04)
func (c *ClockSyncActor) Offset() int64 {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    return c.offset
}
```

### Weight-Based Rate Limiter

```go
// Source: https://pkg.go.dev/golang.org/x/time/rate
type WeightedLimiter struct {
    limiter *rate.Limiter
    mutex   sync.Mutex
}

func NewWeightedLimiter(requestsPerSecond int, burst int) *WeightedLimiter {
    return &WeightedLimiter{
        limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burst),
    }
}

func (wl *WeightedLimiter) Wait(ctx context.Context, weight int) error {
    wl.mutex.Lock()
    defer wl.mutex.Unlock()
    
    // For weight > 1, we need to wait for multiple tokens
    return wl.limiter.WaitN(ctx, weight)
}

func (wl *WeightedLimiter) Allow(weight int) bool {
    wl.mutex.Lock()
    defer wl.mutex.Unlock()
    return wl.limiter.AllowN(time.Now(), weight)
}

// Binance weights (examples from documentation):
// GET /api/v3/ping: 1
// GET /api/v3/time: 1
// GET /api/v3/exchangeInfo: 20
// GET /api/v3/depth (limit=5000): 50
// POST /api/v3/order: 1
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| listenKey for User Data | WebSocket API subscription | 2025-04-07 | Lower latency, no keepalive |
| millisecond timestamps | microsecond support | 2024-12-17 | Higher precision optional |
| 3-minute ping interval | 20-second ping interval | 2025-01-28 | Faster zombie detection |
| HTTP Client.Timeout | Context-based timeout | Resty v3 | Better cancellation |

**Deprecated/outdated:**
- `POST /api/v3/userDataStream`: Use WebSocket API instead (being removed 2026-02-20)
- `!ticker@arr` stream: Use `<symbol>@ticker` or `!miniTicker@arr` instead (deprecated 2025-11-14)
- SBE schema 3:0: Use 3:2 (3:1 deprecated 2025-12-02)

## Open Questions

Things that couldn't be fully resolved:

1. **Exact exponential backoff parameters**
   - What we know: Binance reconnects every 24h, ping every 20s
   - What's unclear: Optimal initial delay and max delay for reconnection
   - Recommendation: Start with 1s initial, 60s max, add 10% jitter

2. **Buffer sizes for event delivery**
   - What we know: Decided configurable with defaults
   - What's unclear: Optimal default values
   - Recommendation: 1000 events per stream type, configurable via config

3. **Error object pooling in hot paths**
   - What we know: CONTEXT.md says "minimal allocations"
   - What's unclear: Whether to pool error objects
   - Recommendation: Start without pooling, profile and add if needed

## Sources

### Primary (HIGH confidence)
- `/websites/ergo_services` - Supervisor patterns, actor lifecycle, event system
- `/lxzan/gws` - WebSocket Event interface, connection handling
- `/go-resty/docs` - HTTP client retry, middleware, signing
- `/cockroachdb/apd` - Decimal arithmetic operations
- `/sony/gobreaker` - Circuit breaker implementation
- `golang.org/x/time/rate` - Token bucket rate limiter
- `https://developers.binance.com/docs/binance-spot-api-docs` - API v3 documentation

### Secondary (MEDIUM confidence)
- Binance changelog for recent API changes
- Ergo examples repository for patterns

### Tertiary (LOW confidence)
- None - all findings verified with primary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries are stable, well-documented, and actively maintained
- Architecture: HIGH - Ergo framework patterns are well-documented with examples
- Pitfalls: HIGH - All pitfalls documented in official Binance API documentation
- Code examples: HIGH - Examples derived from official documentation

**Research date:** 2026-02-16
**Valid until:** 2026-03-16 (30 days - stable libraries, but Binance API changes periodically)

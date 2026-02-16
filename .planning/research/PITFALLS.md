# Pitfalls: Cryptocurrency Exchange Connector

**Domain:** Multi-Exchange Connector Package  
**Researched:** 2026-02-16  
**Confidence:** HIGH (Context7 official docs, verified sources)

---

## Critical Pitfalls

Mistakes that cause rewrites, production failures, or financial loss.

### Clock Synchronization Failure

**What goes wrong:** Exchange rejects all signed requests with "timestamp outside recvWindow" errors. Common after system restarts, DST changes, or NTP daemon issues.

**Why it happens:** Exchanges require request timestamps within a narrow window of server time. Binance uses `recvWindow` (default 5000ms), Bybit uses `expires` timestamp for WebSocket auth. Even 1 second drift can cause failures.

**Consequences:**

- All authenticated API calls fail with `-1021 INVALID_TIMESTAMP`
- WebSocket authentication rejected
- Trading operations blocked during critical market moments

**Warning signs:**

- `-1021 Timestamp for this request is outside of the recvWindow` errors
- `-1021 Timestamp for this request was 1000ms ahead of the server's time`
- Sudden authentication failures after system restart or network change

**Prevention:**

1. Implement continuous clock synchronization actor that polls server time
2. Track local clock offset and apply correction before signing
3. Alert when offset exceeds 500ms
4. Never trust local system time alone - always verify against exchange

**Phase:** Core Infrastructure (Phase 1) - Must be first thing working

```go
// CORRECT: Track and apply clock offset
type ClockSync struct {
    offset atomic.Int64 // server_time - local_time in milliseconds
}

func (cs *ClockSync) Timestamp() int64 {
    return time.Now().UnixMilli() + cs.offset.Load()
}
```

---

### Rate Limit Violations and IP Bans

**What goes wrong:** Exchanges ban your IP address, blocking ALL operations. Bans escalate from 2 minutes to 3 days for repeat offenders.

**Why it happens:** Binance uses a weight-based system where different endpoints have different costs. Developers often:

- Poll REST APIs instead of using WebSocket streams
- Ignore 429 responses and continue requesting
- Don't track accumulated weight across requests

**Consequences:**

- HTTP 418 "IP banned until {timestamp}" response
- All exchange connectivity blocked
- Cascading failures across all trading operations

**Warning signs:**

- Receiving 429 (Too Many Requests) responses
- `X-MBX-USED-WEIGHT-*` headers approaching limits
- Error `-1003 TOO_MANY_REQUESTS`

**Prevention:**

1. Track weight per request from response headers
2. Implement token bucket with exchange-specific weights
3. Use WebSocket streams for live data, never poll REST
4. Back off immediately on 429 with `Retry-After` header
5. Never ignore rate limit warnings

**Binance specifics:**

- 6000 weight per minute limit
- Each endpoint has documented weight
- Headers show current usage: `X-MBX-USED-WEIGHT-1M`

**Bybit specifics:**

- 600 requests per 5 seconds per IP
- WebSocket: 500 connections per 5 minutes

**Phase:** Core Infrastructure (Phase 1)

```go
// CORRECT: Weighted rate limiter
type RateLimiter struct {
    weights  map[string]int // endpoint -> weight
    bucket   *tokenBucket
}

func (rl *RateLimiter) BeforeRequest(endpoint string) error {
    weight := rl.weights[endpoint]
    if !rl.bucket.Allow(weight) {
        return ErrRateLimitExceeded
    }
    return nil
}
```

---

### Floating Point Arithmetic for Financial Values

**What goes wrong:** Precision loss causes incorrect order quantities, prices, and balances. Order rejections for "insufficient balance" when balance appears sufficient.

**Why it happens:** IEEE 754 float64 cannot precisely represent decimal fractions. `0.1 + 0.2 != 0.3` in floating point. Exchanges use exact decimal precision (often 8+ decimal places).

**Consequences:**

- Orders rejected for invalid price precision
- Balance calculations off by dust amounts
- Accumulated errors in profit/loss calculations
- Self-matching or trade failures

**Warning signs:**

- Prices like `50000.123456789012345` instead of `50000.12345678`
- Quantity comparisons returning wrong results
- Exchange rejecting orders with "FILTER_FAILURE" for price/quantity

**Prevention:**

1. Use arbitrary-precision decimal library (shopspring/decimal or cockroachdb/apd)
2. Parse all exchange values as strings to decimals
3. Never convert to float64 for calculations
4. Round only for display, never for calculations
5. Store original precision from exchange

**Phase:** Domain Models (Phase 1)

```go
// CORRECT: Use decimal type
type Price = *apd.Decimal
type Quantity = *apd.Decimal

func NewPrice(s string) (Price, error) {
    d, _, err := apd.NewFromString(s)
    return d, err
}

// WRONG: Never use float64
// price := 50000.12345678 // Precision loss!
```

---

### Order Book Synchronization Errors

**What goes wrong:** Local order book diverges from exchange, leading to incorrect trading decisions based on stale or wrong data.

**Why it happens:** Order book updates come via WebSocket as diffs, not snapshots. If you miss messages or process them out of order, the book becomes corrupt. Network issues cause message gaps.

**Consequences:**

- Trading based on incorrect market depth
- Missed arbitrage opportunities
- Invalid order placement (price outside book)

**Warning signs:**

- Sequence gaps detected (`U` > `lastUpdateId + 1`)
- Order book checksum failures
- Inconsistent bid/ask prices vs REST API

**Prevention:**

1. Buffer events from WebSocket immediately on connect
2. Fetch REST snapshot after WebSocket connect
3. Validate: `snapshot.lastUpdateId >= firstEvent.U`
4. Discard buffered events where `u <= lastUpdateId`
5. Apply remaining buffered events in sequence
6. On sequence gap (`U > lastUpdateId + 1`), re-sync from scratch
7. Implement checksum validation where available

**Binance order book sync protocol (from official docs):**

1. Open WebSocket to `wss://stream.binance.com/ws/<symbol>@depth`
2. Buffer events, note `U` of first event
3. GET `/api/v3/depth?symbol=XXX&limit=5000`
4. If `lastUpdateId < firstEvent.U`, retry step 3
5. Discard buffered events where `u <= lastUpdateId`
6. Initialize book from snapshot
7. Apply buffered then subsequent events

**Phase:** Market Data (Phase 2)

---

### WebSocket Connection Lifecycle Mismanagement

**What goes wrong:** Connections silently drop, subscriptions lost, heartbeat failures cause unexpected disconnections during critical trading.

**Why it happens:** Exchanges require active ping/pong. Binance sends ping every 20 seconds, requires pong within 60 seconds. Bybit requires client-initiated ping every 20 seconds. Missing heartbeat = disconnection.

**Consequences:**

- Silent data loss during disconnection
- Stale market data used for trading
- Order updates not received
- Subscriptions lost on reconnect

**Warning signs:**

- No messages received for extended periods
- Connection appears open but no data
- Orders executing but no updates arriving

**Prevention:**

1. Track last message received timestamp
2. Send/respond to ping/pong per exchange spec
3. Implement connection health monitoring
4. Auto-reconnect on missed heartbeat
5. Re-subscribe to all topics after reconnect
6. Request state sync after reconnect (orders, positions)

**Binance heartbeat:**

- Server sends ping frame every 20 seconds
- Client must respond with pong within 60 seconds
- Connection valid for 24 hours max

**Bybit heartbeat:**

- Client sends `{"op": "ping"}` every 20 seconds
- Server responds with `{"op": "pong"}`

**Phase:** Core Infrastructure (Phase 1)

---

### Order State Machine Race Conditions

**What goes wrong:** Order status updates arrive out of order or duplicated, causing incorrect state transitions. Order marked FILLED then transitions back to NEW.

**Why it happens:** WebSocket and REST can send same update. Network latency causes ordering issues. Exchange internal race conditions.

**Consequences:**

- Incorrect order status displayed to users
- Trading logic executing on wrong state
- Double-fills or missed fills
- Audit trail inconsistencies

**Warning signs:**

- Order status moving backward (FILLED -> NEW)
- Same update processed multiple times
- State transitions that shouldn't be possible

**Prevention:**

1. Define explicit state transition rules
2. Reject invalid transitions (FILLED cannot become NEW)
3. Track event sequence numbers per order
4. Deduplicate updates by update ID
5. Use compare-and-swap for state updates
6. Log all transition attempts for debugging

**Valid Binance order states:**

- NEW → PARTIALLY_FILLED → FILLED
- NEW → CANCELED
- PARTIALLY_FILLED → FILLED
- PARTIALLY_FILLED → CANCELED
- REJECTED (terminal)
- EXPIRED (terminal)

**Phase:** Order Management (Phase 3)

```go
// CORRECT: Explicit state machine
var validTransitions = map[OrderStatus][]OrderStatus{
    StatusNew:             {StatusPartiallyFilled, StatusFilled, StatusCanceled, StatusExpired},
    StatusPartiallyFilled: {StatusFilled, StatusCanceled},
    StatusFilled:          {}, // Terminal
    StatusCanceled:        {}, // Terminal
    StatusExpired:         {}, // Terminal
    StatusRejected:        {}, // Terminal
}

func (o *Order) CanTransition(to OrderStatus) bool {
    allowed, exists := validTransitions[o.Status]
    if !exists {
        return false
    }
    for _, s := range allowed {
        if s == to {
            return true
        }
    }
    return false
}
```

---

### Circuit Breaker Absence

**What goes wrong:** Single exchange failure cascades through entire system, exhausting resources, causing timeout storms, and blocking other exchanges.

**Why it happens:** Without circuit breaker, failed requests are retried indefinitely, threads/goroutines pile up waiting for responses that will never come.

**Consequences:**

- Resource exhaustion (memory, goroutines, connections)
- Latency spikes across all operations
- Other healthy exchanges affected
- Cascading system failure

**Warning signs:**

- Increasing request latency
- Growing number of pending requests
- Memory growth from queued operations
- Timeouts becoming more frequent

**Prevention:**

1. Implement circuit breaker per exchange
2. States: Closed (normal) → Open (failing) → Half-Open (testing)
3. Track failure rate within rolling window
4. Open circuit after threshold (e.g., 50% failures in 30 seconds)
5. Allow limited traffic in half-open to test recovery
6. Provide fallback behavior when circuit open

**Phase:** Core Infrastructure (Phase 1)

```go
type CircuitState int

const (
    CircuitClosed CircuitState = iota
    CircuitOpen
    CircuitHalfOpen
)

type CircuitBreaker struct {
    state           CircuitState
    failures        int
    successes       int
    lastFailureTime time.Time
    config          CircuitConfig
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == CircuitOpen {
        if time.Since(cb.lastFailureTime) > cb.config.Timeout {
            cb.state = CircuitHalfOpen
        } else {
            return ErrCircuitOpen
        }
    }

    err := fn()
    if err != nil {
        cb.onFailure()
        return err
    }
    cb.onSuccess()
    return nil
}
```

---

### Duplicate Order Submissions

**What goes wrong:** Same order submitted multiple times due to timeout + retry, causing double positions and unexpected exposure.

**Why it happens:** Network timeout occurs after exchange received request. Client retries, creating duplicate order. No idempotency key provided.

**Consequences:**

- Double positions
- Unexpected leverage
- Risk management violations
- Financial loss

**Warning signs:**

- Multiple orders with same parameters
- Orders appearing after timeout
- Position size exceeds intended

**Prevention:**

1. Use client order ID (newClientOrderId on Binance, orderLinkId on Bybit)
2. Generate unique ID before sending request
3. On timeout, query order status before retrying
4. Implement request deduplication at actor level
5. Log all submission attempts with IDs

**Phase:** Order Management (Phase 3)

---

## Exchange-Specific Gotchas

### Binance

#### recvWindow Timing

- Default: 5000ms
- Maximum: 60000ms
- **Recommendation: Use 5000 or less for serious trading**
- Request rejected if `serverTime - timestamp > recvWindow` OR `timestamp > serverTime + 1000`

```go
// Add recvWindow to all signed requests
params.Set("recvWindow", "5000")
params.Set("timestamp", fmt.Sprintf("%d", timestamp))
```

#### Weight-Based Rate Limiting

- 6000 weight per minute per IP
- Different endpoints have different weights
- Order placement: 1 weight
- Order book depth 5000: 50 weight
- Must track accumulated weight from headers

#### Order Status PENDING_CANCEL

- Currently unused but exists in enum
- Don't implement special handling

#### WebSocket Stream Naming

- Symbols must be lowercase: `btcusdt@depth`
- Combined streams: `/stream?streams=btcusdt@trade/ethusdt@trade`

#### Cancel-Replace Edge Cases

- `-2021`: Partial failure (cancel OR new order failed)
- `-2022`: Complete failure (both failed)
- `cancelReplaceMode`: STOP_ON_FAILURE vs ALLOW_FAILURE

### Bybit

#### Different Auth for WebSocket

- **Public streams**: No authentication required
- **Private streams**: Must authenticate with op: "auth"
- Auth signature format: `GET/realtime{expires}`

```go
expires := time.Now().Add(10 * time.Second).UnixMilli()
sigData := fmt.Sprintf("GET/realtime%d", expires)
signature := hmacSHA256(apiSecret, sigData)
```

#### DCP (Disconnect Cancel Protect)

- Cancels all active orders if WebSocket disconnects for 10+ seconds
- Prevents runaway positions
- Must reconnect quickly to prevent auto-cancel

#### V5 API Structure

- Unified endpoint for spot, linear, inverse, options
- Category parameter required in requests
- Different rate limits per category

#### IP Rate Limits

- HTTP: 600 requests per 5 seconds
- WebSocket: 500 connections per 5 minutes
- "403, access too frequent" = must wait 10 minutes

---

## Common Bugs

### 1. Race Condition in Order State Updates

**Problem:** Concurrent WebSocket and REST updates to same order
**Symptoms:** Order state flickering, incorrect final status
**Fix:** Use mutex + version checking, process updates sequentially per order

### 2. Memory Leak in WebSocket Handlers

**Problem:** Goroutines not cleaned up on disconnect
**Symptoms:** Growing memory, goroutine count increasing
**Fix:** Context cancellation, wait groups, defer cleanup

### 3. Stale Connection Detection

**Problem:** TCP connection appears open but exchange has closed it
**Symptoms:** No error but no data received
**Fix:** Track last message time, heartbeat enforcement, read deadline

### 4. Precision Loss in Quantity Calculations

**Problem:** Division operations creating infinite precision
**Symptoms:** Exchange rejects order quantity
**Fix:** Always round to symbol's stepSize after calculations

### 5. Timezone Handling

**Problem:** Local timezone affecting timestamp generation
**Symptoms:** Intermittent signature failures
**Fix:** Always use UTC, Unix milliseconds

---

## Security Considerations

### API Key Handling

**NEVER:**

- Hardcode API keys in source code
- Log API keys or secrets
- Store keys in version control
- Transmit keys over unencrypted channels

**ALWAYS:**

- Load keys from environment or secure vault
- Clear sensitive data from memory after use
- Use read-only keys where possible
- Implement IP whitelist restrictions

### Signature Generation

**Binance:**

```
HMAC-SHA256(secret_key, query_string)
```

- Sign the raw query string, not URL-encoded
- Parameters sorted alphabetically
- No trailing `&` before signature

**Bybit:**

```
HMAC-SHA256(secret_key, "GET/realtime{expires}")
```

- For WebSocket auth only
- expires must be in the future
- Different format for REST vs WebSocket

### Request Replay Protection

Both exchanges use timestamps to prevent replay:

- Binance: recvWindow limits request validity window
- Bybit: expires timestamp limits signature validity

---

## Testing Traps

### 1. Mock Exchange Behavior Drift

**Problem:** Mock doesn't match real exchange behavior
**Symptoms:** Tests pass but production fails
**Fix:**

- Use recorded real responses as test fixtures
- Implement mock from official docs
- Contract tests comparing mock to real

### 2. Timing-Dependent Tests

**Problem:** Tests that depend on specific timing are flaky
**Symptoms:** Intermittent test failures
**Fix:**

- Use mock clocks (interface-based time)
- Test state transitions, not timing
- Integration tests with generous timeouts

### 3. Order State Race Conditions

**Problem:** Tests don't simulate concurrent updates
**Symptoms:** Race conditions only found in production
**Fix:**

- Use race detector: `go test -race`
- Test concurrent update scenarios
- Use sync.Map or proper locking in tests

### 4. Testnet vs Production Differences

**Problem:** Testnet behavior differs from production
**Symptoms:** Code works on testnet, fails on production
**Fix:**

- Test both environments
- Document known differences
- Use feature flags for environment-specific behavior

### 5. WebSocket Connection State

**Problem:** Tests don't cover reconnection scenarios
**Symptoms:** Reconnection bugs found in production
**Fix:**

- Simulate disconnects in tests
- Test subscription recovery
- Verify message ordering after reconnect

---

## Phase-Specific Warnings

| Phase               | Likely Pitfall                | Mitigation                                      |
| ------------------- | ----------------------------- | ----------------------------------------------- |
| Core Infrastructure | Clock sync, rate limiting     | Implement sync actor first, weight tracking     |
| Domain Models       | Decimal precision             | Use apd.Decimal everywhere, no float64          |
| Market Data         | Order book sync, message gaps | Buffer + snapshot protocol, sequence validation |
| Order Management    | State races, duplicates       | State machine, idempotency keys                 |
| Risk Management     | Stale positions, exposure     | Real-time sync, circuit breakers                |
| Integration Testing | Testnet differences           | Test both environments, recorded fixtures       |

---

## Sources

- **HIGH Confidence**: Binance Official API Docs (Context7), Bybit V5 API Docs (Context7)
- **MEDIUM Confidence**: CCXT issues for real-world problems
- **LOW Confidence**: General web search for patterns (verified against official docs)

### Verified URLs

- https://github.com/binance/binance-spot-api-docs - Rate limits, recvWindow, order book sync
- https://bybit-exchange.github.io/docs/v5/ - WebSocket auth, DCP, rate limits
- https://github.com/ccxt/ccxt/issues/2725 - Clock sync issues real-world example

---

## Summary for Roadmap

**Must address in Phase 1:**

1. Clock synchronization actor
2. Weight-based rate limiter
3. Decimal types in all domain models
4. Circuit breaker per exchange
5. WebSocket lifecycle (connect/heartbeat/reconnect)

**Address in Phase 2 (Market Data):**

1. Order book synchronization protocol
2. Message sequence validation
3. Checksum verification

**Address in Phase 3 (Order Management):**

1. Order state machine with transition validation
2. Idempotency via client order IDs
3. Concurrent update handling

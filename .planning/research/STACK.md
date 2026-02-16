# Stack: Cryptocurrency Exchange Connector

**Project:** Multi-Exchange Connector Package  
**Researched:** 2025-02-16  
**Overall confidence:** HIGH

---

## Executive Summary

This document defines the recommended technology stack for building a production-grade cryptocurrency exchange connector library in Go. The stack prioritizes: **financial precision**, **fault tolerance**, **observability**, and **Go 1.25+ idioms**.

The choices are prescriptive based on verified current versions from Context7 and official documentation.

---

## Core Technologies

| Technology | Version | Purpose | Confidence |
|------------|---------|---------|------------|
| Go | 1.25+ | Runtime environment | HIGH |
| Ergo Framework | v3.10 | Actor system & supervision | HIGH |
| cockroachdb/apd | v3 | Arbitrary precision decimals | HIGH |
| lxzan/gws | latest | WebSocket client | MEDIUM |
| resty | v3-beta / v2 | HTTP REST client | MEDIUM |
| zerolog | latest | Structured logging | HIGH |
| golang.org/x/time/rate | latest | Token bucket rate limiting | HIGH |
| golang.org/x/sync | latest | errgroup, singleflight | HIGH |
| sony/gobreaker | latest | Circuit breaker | HIGH |

---

## Actor Framework

### Recommended: Ergo Framework v3.10

**Import:** `ergo.services/ergo@latest`

**Rationale:**
- **Erlang/OTP patterns in Go**: Brings battle-tested actor supervision patterns to Go
- **Supervisor hierarchies**: Native support for OneForOne, AllForOne, RestForOne strategies
- **Network transparency**: Actors can communicate across nodes seamlessly
- **Event system**: Built-in pub/sub with `RegisterEvent`, `SendEvent`, `LinkEvent`
- **Message passing**: `gen.Call` (sync), `gen.Cast` (async), `gen.Send` (fire-and-forget)

**Key patterns for exchange connectors:**
```go
// Supervisor with transient restart (only restart on error)
spec := act.SupervisorSpec{
    Type: act.SupervisorTypeOneForOne,
    Restart: act.SupervisorRestart{
        Strategy:  act.SupervisorStrategyTransient,
        Intensity: 5,  // Max 5 restarts
        Period:    10, // Within 10 seconds
    },
}

// Actor lifecycle
func (a *MyActor) Init(args ...any) error { /* setup */ }
func (a *MyActor) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) { /* sync */ }
func (a *MyActor) HandleMessage(from gen.PID, message any) error { /* async */ }
func (a *MyActor) HandleEvent(message gen.MessageEvent) error { /* events */ }
func (a *MyActor) Terminate(reason error) { /* cleanup */ }
```

**Source:** https://github.com/ergo-services/ergo

---

## Decimal Arithmetic

### Recommended: cockroachdb/apd v3

**Import:** `github.com/cockroachdb/apd/v3`

**Rationale:**
- **NEVER use float64 for money** - This is non-negotiable for financial software
- **IEEE 754 alternative**: Implements General Decimal Arithmetic specification
- **Precision control**: Configurable precision (default 34 digits), rounding modes
- **Condition flags**: Trap overflow, inexact results, underflow
- **Production proven**: Used by CockroachDB for financial calculations

**Usage pattern:**
```go
import "github.com/cockroachdb/apd/v3"

// Create decimals (never use float64)
price := apd.New(50000, -4)  // 5.0000
qty := apd.New(1, -3)        // 0.001

// Context with precision and traps
ctx := apd.Context{
    Precision: 34,
    Traps:     apd.Overflow | apd.Underflow,
}

// Arithmetic operations
result := new(apd.Decimal)
_, err := ctx.Mul(result, price, qty)

// Comparison
if result.Cmp(other) > 0 {
    // result > other
}
```

**Why NOT shopspring/decimal:**
- apd has better precision control and condition flags
- apd is more actively maintained by CockroachDB team
- apd supports scientific functions (Pow, Sqrt, Ln, Exp)

**Source:** https://github.com/cockroachdb/apd

---

## WebSocket Layer

### Recommended: lxzan/gws

**Import:** `github.com/lxzan/gws@latest`

**Rationale:**
- **High performance**: Optimized for concurrent environments
- **Event-based API**: Clean interface with `OnOpen`, `OnClose`, `OnMessage`, `OnPing`, `OnPong`
- **Parallel processing**: `ParallelEnabled` for concurrent message handling
- **Compression support**: Per-message deflate
- **Error recovery**: Built-in recovery mechanism

**Usage pattern:**
```go
type WSHandler struct {
    conn *gws.Conn
}

func (h *WSHandler) OnOpen(socket *gws.Conn) {
    _ = socket.SetDeadline(time.Now().Add(PingInterval + PingWait))
}

func (h *WSHandler) OnMessage(socket *gws.Conn, message *gws.Message) {
    defer message.Close()
    // Process message
}

func (h *WSHandler) OnClose(socket *gws.Conn, err error) {
    // Handle reconnection
}

// Client connection
upgrader := gws.NewUpgrader(&handler{}, &gws.ServerOption{
    ParallelEnabled:  true,
    Recovery:         gws.Recovery,
    PermessageDeflate: gws.PermessageDeflate{Enabled: true},
})
```

### Alternative: gorilla/websocket

**Status:** Popular but maintenance concerns. The project was archived then unarchived. Use gws for new projects.

### Alternative: coder/websocket

**Status:** Minimal, idiomatic, with first-class context support. Good alternative if gws doesn't fit.

**Source:** https://github.com/lxzan/gws

---

## HTTP Client

### Recommended: resty v3-beta (or v2 for stability)

**Import:** `resty.dev/v3` (v3-beta) or `github.com/go-resty/resty/v2` (stable)

**Rationale:**
- **Rich retry mechanisms**: Configurable retry counts, wait times, conditions
- **Middleware support**: Request/response middleware for auth, logging
- **Circuit breaker integration**: Built-in circuit breaker support in v3
- **SSE support**: Server-Sent Events in v3
- **Context-aware**: Proper context propagation

**v3 Breaking Changes (awareness required):**
- Requires `defer client.Close()` after creation
- Uses context with timeout instead of `http.Client.Timeout`
- Retries only idempotent methods by default
- Custom URL: `resty.dev/v3`

**Usage pattern:**
```go
import "resty.dev/v3"

client := resty.New()
defer client.Close()

// Retry configuration
client.SetRetryCount(3).
    SetRetryWaitTime(1 * time.Second).
    SetRetryMaxWaitTime(5 * time.Second).
    AddRetryCondition(func(r *resty.Response, err error) bool {
        return err != nil || r.StatusCode() >= 500
    })

// Request with auth
resp, err := client.R().
    SetHeader("X-MBX-APIKEY", apiKey).
    SetQueryParams(params).
    SetResult(&result).
    Get("/api/v3/order")
```

**Why NOT standard net/http:**
- No built-in retry logic
- More boilerplate for common patterns
- No middleware support

**Source:** https://resty.dev

---

## Structured Logging

### Recommended: zerolog

**Import:** `github.com/rs/zerolog/log`

**Rationale:**
- **Zero allocation**: Minimizes GC pressure in hot paths
- **JSON by default**: Structured output for log aggregation
- **Level filtering**: Trace, Debug, Info, Warn, Error, Fatal, Panic
- **Context integration**: `Ctx(ctx)` for trace correlation
- **Console pretty-print**: `ConsoleWriter` for development
- **Field types**: Strongly typed fields (Str, Int, Dur, Err, etc.)

**Usage pattern:**
```go
import (
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

// Development: pretty console
log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

// Production: JSON
zerolog.SetGlobalLevel(zerolog.InfoLevel)

// Structured logging with context
log.Info().
    Str("exchange", "binance").
    Str("symbol", "BTCUSDT").
    Err(err).
    Msg("order placement failed")

// Conditional logging (performance optimization)
if e := log.Debug(); e.Enabled() {
    e.Str("data", expensiveComputation()).Msg("debug info")
}
```

### Alternative: log/slog (Go 1.21+)

**Status:** Standard library structured logging. Good choice if you prefer zero dependencies.

**Consider slog when:**
- Want standard library only
- Don't need zero-allocation guarantees
- Prefer official Go tooling

**Recommendation:** zerolog for production-grade performance; slog for simplicity.

**Source:** https://github.com/rs/zerolog

---

## Rate Limiting

### Recommended: golang.org/x/time/rate

**Import:** `golang.org/x/time/rate`

**Rationale:**
- **Official Go x-package**: Maintained by Go team
- **Token bucket algorithm**: Industry standard for rate limiting
- **Flexible**: `Allow()`, `Wait()`, `Reserve()` patterns
- **Burst support**: Handle spikes with configurable burst
- **Context-aware**: `Wait(ctx)` respects cancellation

**Usage pattern:**
```go
import "golang.org/x/time/rate"

// 10 requests per second, burst of 5
limiter := rate.NewLimiter(rate.Limit(10), 5)

// Non-blocking check
if !limiter.Allow() {
    return ErrRateLimited
}

// Blocking with context
if err := limiter.Wait(ctx); err != nil {
    return err // context cancelled
}

// Binance-style weighted rate limiting
// Use separate limiter per endpoint weight
```

**For Binance weighted rate limiting:**
- Create separate limiters for different weight classes
- Track cumulative weight with `SetLimit()` / `SetBurst()`
- Consider wrapping in a `WeightedLimiter` type

**Source:** https://pkg.go.dev/golang.org/x/time/rate

---

## Concurrency Primitives

### errgroup

**Import:** `golang.org/x/sync/errgroup`

**Rationale:**
- **Error propagation**: First error from any goroutine is returned
- **Context cancellation**: Automatic cancellation on first error
- **Bounded parallelism**: `SetLimit(n)` for concurrency control

**Usage pattern:**
```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(ctx)
g.SetLimit(10) // Max 10 concurrent operations

for _, exchange := range exchanges {
    exchange := exchange // capture
    g.Go(func() error {
        return fetchTicker(ctx, exchange)
    })
}

if err := g.Wait(); err != nil {
    return err
}
```

### singleflight

**Import:** `golang.org/x/sync/singleflight`

**Rationale:**
- **Request deduplication**: Prevent thundering herd
- **Cache stampede protection**: Multiple callers share same result
- **Critical for exchanges**: Prevent duplicate REST calls for same data

**Usage pattern:**
```go
import "golang.org/x/sync/singleflight"

var sf singleflight.Group

func GetTicker(symbol string) (*Ticker, error) {
    v, err, _ := sf.Do(symbol, func() (interface{}, error) {
        return fetchTickerFromExchange(symbol)
    })
    if err != nil {
        return nil, err
    }
    return v.(*Ticker), nil
}
```

**Source:** https://pkg.go.dev/golang.org/x/sync

---

## Circuit Breaker

### Recommended: sony/gobreaker

**Import:** `github.com/sony/gobreaker`

**Rationale:**
- **Standard pattern**: Closed → Open → Half-Open → Closed state machine
- **Configurable thresholds**: `ReadyToTrip` function for custom logic
- **State callbacks**: `OnStateChange` for monitoring
- **Generic support**: `CircuitBreaker[T]` for type safety

**Usage pattern:**
```go
import "github.com/sony/gobreaker"

cb := gobreaker.NewCircuitBreaker[string](gobreaker.Settings{
    Name:        "BinanceAPI",
    MaxRequests: 3,               // Max requests in half-open state
    Interval:    10 * time.Second, // Clear counts interval
    Timeout:     30 * time.Second, // Open → Half-Open timeout
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        // Trip after 5 consecutive failures
        return counts.ConsecutiveFailures >= 5
    },
    OnStateChange: func(name string, from, to gobreaker.State) {
        log.Info().Str("circuit", name).
            Str("from", from.String()).
            Str("to", to.String()).
            Msg("circuit state changed")
    },
})

result, err := cb.Execute(func() (string, error) {
    return callExchangeAPI()
})
```

**Source:** https://github.com/sony/gobreaker

---

## Alternatives Considered

### WebSocket: gorilla/websocket

| Aspect | gorilla/websocket | lxzan/gws |
|--------|-------------------|-----------|
| Popularity | Higher | Growing |
| Maintenance | Uncertain (archived/unarchived) | Active |
| Performance | Good | Better (benchmarks) |
| API | Low-level | Event-based |
| Parallel | Manual | Built-in |

**Decision:** Use gws for active maintenance and better concurrency support.

### HTTP: Standard library

| Aspect | net/http | resty |
|--------|----------|-------|
| Dependencies | None | Minimal |
| Retry | Manual | Built-in |
| Middleware | Manual | Built-in |
| DX | Verbose | Fluent |

**Decision:** Use resty for production-grade features (retry, middleware, circuit breaker).

### Logging: log/slog vs zerolog

| Aspect | slog | zerolog |
|--------|------|---------|
| Dependencies | None | Minimal |
| Allocations | Some | Zero |
| Performance | Good | Excellent |
| Ecosystem | Standard | Rich |

**Decision:** zerolog for high-throughput requirements; slog acceptable for simpler needs.

### Decimals: shopspring/decimal vs apd

| Aspect | shopspring/decimal | cockroachdb/apd |
|--------|-------------------|-----------------|
| API | Simpler | More features |
| Precision | Fixed | Configurable |
| Condition flags | No | Yes |
| Scientific functions | Limited | Full |
| Maintenance | Active | Active (CockroachDB) |

**Decision:** apd for production-grade financial calculations with condition handling.

---

## Installation

```bash
# Core
go get ergo.services/ergo@latest
go get github.com/cockroachdb/apd/v3@latest

# WebSocket
go get github.com/lxzan/gws@latest

# HTTP Client
go get resty.dev/v3@latest  # or v2 for stable

# Logging
go get github.com/rs/zerolog/log@latest

# Rate limiting
go get golang.org/x/time/rate@latest

# Concurrency
go get golang.org/x/sync@latest

# Circuit breaker
go get github.com/sony/gobreaker@latest
```

---

## Confidence Levels

| Category | Level | Reason |
|----------|-------|--------|
| Actor Framework | HIGH | Ergo v3.10 is mature with active development |
| Decimal Arithmetic | HIGH | apd v3 is production-proven by CockroachDB |
| WebSocket | MEDIUM | gws is performant but smaller community than gorilla |
| HTTP Client | MEDIUM | resty v3 in beta; v2 is stable |
| Logging | HIGH | zerolog is battle-tested with zero-allocation guarantees |
| Rate Limiting | HIGH | Official x-package, well-documented |
| Concurrency | HIGH | Official x-packages, widely adopted |
| Circuit Breaker | HIGH | gobreaker is simple, well-tested, widely used |

---

## Sources

- Ergo Framework: https://github.com/ergo-services/ergo (verified v3.10)
- cockroachdb/apd: https://github.com/cockroachdb/apd (verified v3)
- lxzan/gws: https://github.com/lxzan/gws (verified latest)
- resty: https://resty.dev (verified v3-beta, v2 stable)
- zerolog: https://github.com/rs/zerolog (verified latest)
- golang.org/x/time/rate: https://pkg.go.dev/golang.org/x/time/rate
- golang.org/x/sync: https://pkg.go.dev/golang.org/x/sync
- sony/gobreaker: https://github.com/sony/gobreaker

---

## Summary

| Layer | Technology | Why |
|-------|------------|-----|
| Actor System | Ergo v3.10 | Erlang/OTP patterns for fault tolerance |
| Decimals | cockroachdb/apd v3 | Financial precision, condition flags |
| WebSocket | lxzan/gws | Performance, parallel processing |
| HTTP | resty v3/v2 | Retry, middleware, circuit breaker |
| Logging | zerolog | Zero-allocation, structured JSON |
| Rate Limiting | golang.org/x/time/rate | Official, token bucket |
| Concurrency | golang.org/x/sync | errgroup, singleflight |
| Circuit Breaker | sony/gobreaker | State machine, configurable |

This stack provides a solid foundation for building a production-grade cryptocurrency exchange connector with:
- **Financial correctness** (apd decimals)
- **Fault tolerance** (Ergo supervisors, circuit breaker)
- **High performance** (zerolog, gws)
- **Observability** (structured logging, state monitoring)
- **Go 1.25+ idioms** (iterators, generics, min/max built-ins)

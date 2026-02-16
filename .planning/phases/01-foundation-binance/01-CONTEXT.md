# Phase 1: Foundation & Binance Infrastructure - Context

**Gathered:** 2025-02-16
**Status:** Ready for planning

<domain>
## Phase Boundary

Reliable Binance connection with decimal precision and fault tolerance. Developer can connect to Binance REST and WebSocket APIs, authenticate requests, handle reconnection, and receive events. This is a library developers import — per-exchange connector with actor-based supervision.

</domain>

<decisions>
## Implementation Decisions

### Connection Lifecycle

- **Start() behavior**: Async (returns immediately, connection happens in background)
- **Ready signaling**: Ready channel — caller can wait for `conn.Ready()` or listen to events
- **Initial connection failure**: Auto-retry with exponential backoff
- **Reconnection**: Automatic on WebSocket disconnect, infinite attempts
- **Stop() behavior**: Context-aware — waits for in-flight requests with deadline
- **Connection state**: Event-based (`OnConnection(Connected|Disconnected|Reconnecting)`) + synchronous query (`conn.IsConnected()`)
- **Zombie detection**: Active REST ping at configurable interval
- **Partial connectivity**: REST can work while WebSocket reconnects (partial OK)
- **Stop vs reconnect**: Stop() breaks reconnection attempts immediately (stop fast)
- **Connection flapping**: Debounce connection events (e.g., 1s settle period)
- **Runtime exchange add**: Yes, can add exchanges after connector is running
- **Per-exchange connector**: One Connector instance = one exchange
- **Restart**: Connector can be stopped and restarted with new config
- **Concurrency**: All connector methods safe for concurrent use
- **Cleanup**: Full resource cleanup on Stop() (goroutines, pools, connections)

### Event Delivery

- **Delivery mechanism**: Callback-based (`OnTick(cb)`, `OnOrder(cb)`, etc.)
- **Subscription pattern**: Per-type methods (type-safe, discoverable)
- **Exchange identification**: In event struct (Ticker.Exchange, Order.Exchange, etc.)
- **Slow handlers**: Buffered — configurable buffer size, drops oldest or blocks when full
- **Multiple subscribers**: Non-blocking invocation (goroutine per subscriber)
- **Callback panic**: Log and continue (don't crash the library)
- **Unsubscribe**: Unsubscribe function returned from subscription call
- **Initial state**: Live only (no snapshot on subscribe for tickers/trades)
- **Nil callback**: Silent (no-op)
- **Callback invocation order**: Sequential for each subscriber
- **High-frequency events**: Configurable buffer + pooled events for hot paths
- **Event ordering**: Per-subscriber order (each subscriber gets events in sequence)
- **Resubscription**: Auto-resubscribe on reconnection
- **Backpressure**: Optional channel acceptance for flow control

### Configuration Pattern

- **Config structure**: Struct-based (not functional options alone)
- **Building config**: Fluent builders (`NewConfig().WithAPIKey(k).WithTimeout(t).Build()`)
- **Optional fields**: Sensible defaults
- **Validation timing**: Validate early (fail on create, not on use)
- **API key loading**: Config struct + environment variable support
- **Config ownership**: Copy on use (user can't modify running connector)
- **Validation reporting**: Error return from Build()
- **Config organization**: Nested structs (ExchangeConfig, RateLimitConfig, etc.)

### Error Handling

- **Error structure**: Rich error types with Code, Message, Details
- **Error checking**: Go 1.13+ wrapping (`errors.Is()`, `errors.As()`)
- **Context carried**: Rich — exchange name, request ID, timestamp
- **Sentinel errors**: Yes (`ErrNotFound`, `ErrRateLimited`, `ErrNotConnected`, etc.)
- **Stack traces**: Debug mode only (not in production)
- **Exchange errors**: Normalized (map exchange codes to library errors)
- **Rate limit errors**: `RateLimitError` type with `RetryAfter` field
- **Connection errors**: `ConnectionError` type with `IsTemporary()` method
- **Validation errors**: `ValidationError` type with `Field`, `Value`, `Constraint`
- **Request tracing**: Request ID in error details
- **Error types**: Separate types per category (not single type with codes)
- **Wrapping**: Use `%w` for error chains
- **Message format**: Full error chain (include wrapped error message)
- **Hot path performance**: Minimal allocations (pool error objects if needed)
- **Retry signaling**: `IsRetryable()` method on errors
- **Logging format**: Structured (JSON-friendly for zerolog)

### OpenCode's Discretion

- Exact exponential backoff parameters (base delay, max delay, jitter)
- Buffer sizes for event delivery (configurable but default values)
- Debounce duration for connection flapping
- Exact panic recovery behavior in callbacks
- Error object pooling implementation
- Debug vs production mode detection
- Structured logging field names

</decisions>

<specifics>
## Specific Ideas

- Per-exchange connector (one Connector = one exchange), not multi-exchange
- Fluent config builders for clean API
- Events carry exchange identification in struct (not separate parameter)
- Normalized errors hide exchange-specific error codes from users
- Partial connectivity allowed (REST works while WS reconnects)

</specifics>

<deferred>
## Deferred Ideas

- Multi-exchange coordinator (separate phase or user responsibility)
- Connection pooling for multiple API keys
- Metrics/observability integration (Phase 6)

</deferred>

---

*Phase: 01-foundation-binance*
*Context gathered: 2025-02-16*

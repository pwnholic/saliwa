# Phase 2: Bybit Infrastructure - Context

**Gathered:** 2025-02-16
**Status:** Ready for planning

<domain>
## Phase Boundary

Implement Bybit exchange driver (REST + WebSocket) that integrates with the existing Connector API from Phase 1. Same reliability patterns (circuit breaker, clock sync, reconnection) but for Bybit. Requirements: CONN-03 (REST), CONN-04 (WebSocket), CONN-07 (per-second rate limiting).

</domain>

<decisions>
## Implementation Decisions

### Driver Interface Consistency
- Identical method signatures to Binance driver — same names, parameters, return types
- Enables easy exchange switching without code changes
- Formal `pkg/driver.Driver` interface — OpenCode to decide (Go best practices suggest formal interface for compiler enforcement)
- Unique Bybit order types (conditional, TP/SL) — OpenCode to decide (likely exclude for now, add generic extension mechanism if needed)
- Driver updates — minimal approach, implement only what's needed now

### Error Code Mapping
- Map Bybit numeric error codes to domain error types (`ExchangeError`, `ValidationError`, `RateLimitError`, etc.)
- Unified error handling across exchanges
- Unknown/unmappable errors fall back to `ExchangeError` with original code and message preserved
- Original Bybit error code always preserved for debugging
- Auto-retry on rate limit with configurable timeout (consistent with Phase 1 reliability patterns)

### Rate Limiting Behavior
- Block and wait when rate limited — simple, reliable, consistent with Binance limiter
- Per-endpoint tracking — different limits for orders (50/sec), account (20/sec), etc.
- Configurable behavior — default is block-with-timeout, option to return error immediately

### WebSocket Subscription Model
- Hide exchange differences behind Connector API — same `SubscribeTicker()`, `SubscribeOrderBook()`, `SubscribeTrades()` methods
- Reconnection resubscribes automatically — same pattern as Binance
- Exchange-specific options via config, not API surface — keep public API clean
- Use Bybit's ping/pong mechanism with same heartbeat pattern as Binance

### OpenCode's Discretion
- Formal interface vs implicit consistency — pick based on Go idioms
- Unique order types handling — start minimal, add extension mechanism if needed
- Exact per-endpoint rate limit values — based on Bybit API docs
- WebSocket message parsing details — match existing pattern from Binance
- Error code mapping table — research from Bybit API docs

</decisions>

<specifics>
## Specific Ideas

- Driver should be a drop-in replacement for Binance where possible
- Same Connector API surface — user code shouldn't change when switching exchanges
- Reliability first — circuit breaker, clock sync, reconnection all apply

</specifics>

<deferred>
## Deferred Ideas

- Bybit V5 unified margin mode — future enhancement
- Position management endpoints — separate phase if needed
- Bybit options/futures — out of scope for v1 (spot only)

</deferred>

---

*Phase: 02-bybit-infrastructure*
*Context gathered: 2025-02-16*

# Research Summary: Multi-Exchange Connector

**Synthesized:** 2026-02-16  
**Confidence:** HIGH (multiple verified sources via Context7)

---

## Executive Summary

Building a production-grade cryptocurrency exchange connector requires **actor-based fault tolerance** (Ergo Framework), **exact decimal precision** (cockroachdb/apd), and **rigorous protocol compliance** (clock sync, rate limits, order book synchronization). The recommended architecture uses CQRS for order management, event-driven distribution for market data, and supervisor hierarchies for fault isolation. The critical path starts with core infrastructure (clock sync, rate limiting, circuit breaker) before any exchange integration—skipping these foundations guarantees production failures.

Binance and Bybit serve as the initial targets, covering the majority of spot trading volume. Each exchange requires its own supervisor subtree to isolate failures while sharing common patterns (driver interface, event topics, domain models).

---

## Recommended Stack

| Layer               | Technology             | Confidence | Rationale                                                 |
| ------------------- | ---------------------- | ---------- | --------------------------------------------------------- |
| **Actor System**    | Ergo Framework v3.10   | HIGH       | Erlang/OTP patterns, supervision trees, message passing   |
| **Decimals**        | cockroachdb/apd v3     | HIGH       | Arbitrary precision, condition flags, production-proven   |
| **WebSocket**       | lxzan/gws              | MEDIUM     | High performance, parallel processing, active maintenance |
| **HTTP Client**     | resty v3-beta          | MEDIUM     | Built-in retry, middleware, circuit breaker support       |
| **Logging**         | zerolog                | HIGH       | Zero-allocation, structured JSON, level filtering         |
| **Rate Limiting**   | golang.org/x/time/rate | HIGH       | Official x-package, token bucket algorithm                |
| **Concurrency**     | golang.org/x/sync      | HIGH       | errgroup for error propagation, singleflight for dedup    |
| **Circuit Breaker** | sony/gobreaker         | HIGH       | Standard state machine, configurable thresholds           |

---

## Table Stakes Features

Features users expect. Missing these makes the product unusable.

### Core Connectivity

- **REST API Client** — Universal requirement for all exchange operations
- **WebSocket Streams** — Real-time data is non-negotiable for trading
- **Authentication** — HMAC/RSA/Ed25519 for signed requests
- **Rate Limiting** — Exchange bans clients that exceed limits
- **Reconnection Handling** — Auto-reconnect with subscription recovery

### Market Data

- **Ticker Streams** — Basic price/24h stats
- **Order Book Streams** — Essential for market making, arbitrage
- **Trade Streams** — Historical and real-time execution data
- **Symbol/Market Info** — Precision, min notional, filters

### Order Management

- **Place Order** — LIMIT and MARKET order types
- **Cancel Order** — Essential for risk management
- **Query Order Status** — Reconciliation and state management
- **Balance Tracking** — Position management, risk calculation

### Reliability

- **Structured Errors** — Distinguish rate limits, auth errors, network issues
- **Retry Logic** — Handle transient errors without data loss

---

## Key Architecture Decisions

### 1. Actor Model with Supervisor Hierarchy

```
RootSupervisor (OneForOne, 5/10s)
├── CoordinatorActor (message routing)
├── HealthCheckActor (circuit breaker)
├── RiskManagerActor (position limits)
├── PersistenceSupervisor (AllForOne, 3/60s)
│   ├── OrderStoreActor
│   ├── PositionStoreActor
│   └── BalanceStoreActor
├── BinanceSupervisor (OneForOne, 10/60s)
│   ├── ClockSyncActor
│   ├── RateLimitGuardActor
│   ├── MarketDataActor
│   ├── OrderBookActor
│   ├── OrderCommandActor (CQRS Write)
│   └── OrderStateActor (CQRS Read)
└── BybitSupervisor (OneForOne, 10/60s)
    └── ... (same structure)
```

**Rationale:** Per-exchange supervisors isolate failures—Binance outage doesn't affect Bybit operations.

### 2. CQRS for Order Management

| Path      | Actor             | Purpose                                                      |
| --------- | ----------------- | ------------------------------------------------------------ |
| **Write** | OrderCommandActor | Validate, rate limit, submit to exchange, emit events        |
| **Read**  | OrderStateActor   | Subscribe to events, maintain in-memory state, serve queries |

**Rationale:** REST and WebSocket updates can race. CQRS with event sourcing ensures consistent state and auditability.

### 3. Event-Driven Market Data

Topic format: `{type}.{exchange}.{symbol}`

- `ticker.binance.BTCUSDT` — Price updates
- `orderbook.bybit.ETHUSDT` — Order book snapshots/deltas
- `order.binance.placed` — Order lifecycle events
- `health.binance` — Connection health status

**Rationale:** Loose coupling, multiple subscribers (strategies, UI, metrics) without coupling to producers.

---

## Watch Out For

Critical pitfalls that cause production failures or rewrites.

### Phase 1 Blockers

| Pitfall                     | Consequence                             | Prevention                                                     |
| --------------------------- | --------------------------------------- | -------------------------------------------------------------- |
| **Clock Sync Failure**      | All signed requests rejected (-1021)    | ClockSyncActor tracking offset, alert on >500ms drift          |
| **Rate Limit Violations**   | IP banned (2 min to 3 days)             | Weight-based limiter, back off on 429, track X-MBX-USED-WEIGHT |
| **Float64 for Money**       | Incorrect orders, balance errors        | Use apd.Decimal everywhere, parse strings, never convert       |
| **Circuit Breaker Absence** | Cascading failures, resource exhaustion | Per-exchange breaker (Closed→Open→Half-Open)                   |
| **WebSocket Lifecycle**     | Silent data loss, stale data            | Heartbeat enforcement, auto-reconnect, resubscribe             |

### Phase 2-3 Risks

| Pitfall                 | Consequence                           | Prevention                                           |
| ----------------------- | ------------------------------------- | ---------------------------------------------------- |
| **Order Book Desync**   | Trading on wrong prices               | Buffer + REST snapshot + sequence validation         |
| **State Machine Races** | Invalid transitions (FILLED→NEW)      | Explicit valid transitions, deduplicate by update ID |
| **Duplicate Orders**    | Double positions, unexpected exposure | Client order IDs, query before retry on timeout      |

### Exchange-Specific Gotchas

**Binance:**

- Weight-based rate limiting (6000/minute, endpoint-specific weights)
- recvWindow default 5000ms, max 60000ms
- WebSocket symbols must be lowercase
- Order book sync: buffer events → get snapshot → validate U/u sequence

**Bybit:**

- Different auth for WebSocket (`GET/realtime{expires}`)
- DCP (Disconnect Cancel Protect) cancels orders after 10s disconnect
- V5 unified API requires `category` parameter
- 600 requests/5 seconds per IP

---

## Suggested Phase Order

Synthesized from dependencies, risk mitigation, and incremental delivery.

### Phase 1: Foundation (Week 1-2)

**Deliverables:** Domain models, error types, mock driver, core actors

1. Domain models with apd.Decimal types
2. Structured error types (rate limit, validation, auth)
3. Configuration structs
4. Mock driver for testing
5. RootSupervisor structure
6. ClockSyncActor (CRITICAL - blocks all auth)
7. RateLimitGuardActor (CRITICAL - prevents bans)
8. Circuit breaker integration

**Rationale:** Without clock sync and rate limiting, nothing works safely. These are prerequisites.

### Phase 2: Exchange Infrastructure (Week 2-4)

**Deliverables:** REST client, WebSocket client, Binance driver

1. Driver interface definition
2. REST client base (signing, retry, error handling)
3. WebSocket client base (connect, heartbeat, reconnect)
4. Binance REST driver
5. Binance WebSocket driver
6. Rate limiting with weights

**Rationale:** Binance as reference implementation. Bybit follows same patterns.

### Phase 3: Market Data Pipeline (Week 4-6)

**Deliverables:** Real-time tickers, order books, trades

1. MarketDataActor (subscription management, parsing)
2. OrderBookActor (snapshot + delta, sequence validation)
3. Event system with topic routing
4. Ticker, trade stream implementations

**Rationale:** Read-only data is easier to test before order management.

### Phase 4: Order Management (Week 5-7)

**Deliverables:** Place/cancel orders, state tracking

1. OrderCommandActor (CQRS write path)
2. OrderStateActor (CQRS read path)
3. Order state machine with valid transitions
4. RiskManagerActor (pre-trade checks)
5. User data stream (order updates via WebSocket)

**Rationale:** Critical path requires market data working for realistic testing.

### Phase 5: Bybit Driver (Week 6-8)

**Deliverables:** Second exchange support

1. Bybit REST driver
2. Bybit WebSocket driver (different auth format)
3. Per-exchange supervisor for Bybit

**Rationale:** Second exchange validates abstraction layer.

### Phase 6: Persistence & Aggregation (Week 7-9)

**Deliverables:** Durable state, cross-exchange views

1. PersistenceSupervisor
2. OrderStoreActor, PositionStoreActor, BalanceStoreActor
3. ReconcilerActor (periodic state sync)
4. AggregatorActor (cross-exchange best prices)
5. MetricsCollectorActor

**Rationale:** Persistence can use mock stores earlier. Aggregation needs multiple exchanges.

### Phase 7: Public API & Examples (Week 8-10)

**Deliverables:** Production-ready library

1. Connector public interface
2. Configuration and setup helpers
3. Example programs
4. Integration test suite

**Rationale:** Public API last to ensure internal interfaces are stable.

---

## Research Flags

Phases that may need deeper `/gsd-research-phase` investigation:

| Phase              | Research Needed                             | Why                             |
| ------------------ | ------------------------------------------- | ------------------------------- |
| **Binance Driver** | API v3 specifics, testnet endpoints         | Exchange APIs change frequently |
| **Bybit Driver**   | V5 WebSocket message formats, auth          | Less documented than Binance    |
| **Persistence**    | Database choice (PostgreSQL vs TimescaleDB) | Depends on scale requirements   |
| **Deployment**     | Container orchestration, secrets management | Infrastructure-specific         |

Phases with well-documented patterns (skip additional research):

- Domain models (standard Go patterns)
- Actor supervision (Ergo docs are comprehensive)
- Circuit breaker (standard pattern)
- Rate limiting (token bucket is well-understood)

---

## Open Questions

1. **Persistence backend:** PostgreSQL, TimescaleDB, or Redis? Depends on query patterns and scale requirements. Recommend starting with interface, implementing mock for MVP.

2. **Metrics integration:** Prometheus vs OpenTelemetry? Standard Prometheus metrics are simpler; OpenTelemetry for enterprise observability.

3. **Futures/Margin timeline:** Research suggests deferring to v2. What triggers the decision to add? User demand or competitive pressure?

4. **Additional exchanges:** OKX, Coinbase, Kraken? Prioritize based on user demand after v1 is stable.

---

## Confidence Assessment

| Area             | Confidence | Notes                                                                 |
| ---------------- | ---------- | --------------------------------------------------------------------- |
| **Stack**        | HIGH       | Ergo v3.10 mature, apd production-proven, zerolog battle-tested       |
| **Features**     | HIGH       | Based on CCXT reference, Binance/Bybit official docs                  |
| **Architecture** | HIGH       | Supervisor patterns from Ergo docs, CQRS from production case studies |
| **Pitfalls**     | HIGH       | From official exchange docs and real-world CCXT issues                |
| **Timeline**     | MEDIUM     | Estimates based on complexity, actual velocity unknown                |

**Gaps to address during implementation:**

- Binance API v3 endpoint specifics (use MCP Context7 during driver development)
- Bybit V5 WebSocket message format details
- Persistence layer design (defer to Phase 6)

---

## Sources

- **STACK.md:** Context7 verified documentation for all packages
- **FEATURES.md:** CCXT, go-binance SDK, Binance/Bybit official docs
- **ARCHITECTURE.md:** Ergo Framework docs, DolphinDB case study, CQRS implementations
- **PITFALLS.md:** Binance/Bybit official docs, CCXT issues, production incident reports

---

_Research synthesized from 4 parallel research outputs: STACK.md, FEATURES.md, ARCHITECTURE.md, PITFALLS.md_

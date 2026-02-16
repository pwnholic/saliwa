# Roadmap: Multi-Exchange Connector

**Created:** 2026-02-16
**Depth:** Comprehensive
**Phases:** 5
**Core Value:** Reliable, supervised exchange connectivity that trading engines can import and use without handling WebSocket reconnection, rate limiting, or state synchronization.

---

## Overview

This roadmap delivers a production-grade Go library for cryptocurrency exchange connectivity (Binance & Bybit) through 5 phases. Each phase delivers a coherent, verifiable capability built on actor-based supervision with automatic fault recovery.

The structure derives from requirement dependencies: foundation patterns must exist before exchange integrations, market data requires working connections, order management requires market data, and cross-exchange aggregation requires multiple exchanges.

---

## Phase 1: Foundation & Binance Infrastructure

**Goal:** Developer has reliable Binance connection with decimal precision and fault tolerance

**Dependencies:** None (foundation phase)

**Plans:** 4 plans in 3 waves

**Requirements:**

- CONN-01: Connect to Binance via REST API
- CONN-02: Connect to Binance via WebSocket
- CONN-05: HMAC-SHA256 request authentication
- CONN-06: Binance weight-based rate limiting
- CONN-08: WebSocket auto-reconnection
- CONN-09: Subscription restoration after reconnection
- MKTD-06: Decimal precision for all market data
- RESL-01: Circuit breaker opens on unhealthy exchange
- RESL-02: Circuit breaker half-opens after timeout
- RESL-03: Circuit breaker closes on recovery
- RESL-04: Clock synchronization with exchange server
- RESL-05: Unique nonces for replay protection
- RESL-06: Arbitrary-precision decimals for financial values

**Plans:**
- [ ] 01-01-PLAN.md — Project setup & domain models (Wave 1)
- [ ] 01-02-PLAN.md — Binance REST client with signing & rate limiting (Wave 2)
- [ ] 01-03-PLAN.md — Binance WebSocket client with reconnection (Wave 2)
- [ ] 01-04-PLAN.md — Resilience layer & public Connector API (Wave 3)

**Success Criteria:**

1. Developer can connect to Binance REST API with authenticated HMAC-SHA256 requests
2. Developer can establish WebSocket connection to Binance streams
3. WebSocket connections automatically reconnect on disconnect with exponential backoff
4. All active subscriptions are restored automatically after WebSocket reconnection
5. Circuit breaker prevents requests when exchange becomes unhealthy (opens on failures, half-opens to test recovery, closes when healthy)
6. Clock synchronization maintains <500ms offset from exchange server time
7. All prices and quantities use arbitrary-precision decimals (no float64)

---

## Phase 2: Bybit Infrastructure

**Goal:** Developer can connect to Bybit exchange with same reliability patterns as Binance

**Dependencies:** Phase 1 (patterns established)

**Plans:** 6 plans in 3 waves

**Requirements:**

- CONN-03: Connect to Bybit via REST API
- CONN-04: Connect to Bybit via WebSocket
- CONN-07: Bybit per-second rate limiting

**Plans:**
- [ ] 02-01-PLAN.md — URLs & Signer foundation (Wave 1)
- [ ] 02-02-PLAN.md — WebSocket message types (Wave 1)
- [ ] 02-03-PLAN.md — Subscription manager (Wave 1)
- [ ] 02-04-PLAN.md — REST client with rate limiting (Wave 2)
- [ ] 02-05-PLAN.md — WebSocket client with dynamic subscribe (Wave 2)
- [ ] 02-06-PLAN.md — Driver interface & implementation (Wave 3)

**Success Criteria:**

1. Developer can connect to Bybit REST API with authenticated requests (different auth format from Binance)
2. Developer can establish WebSocket connection to Bybit streams
3. Library respects Bybit per-second rate limits automatically (prevents IP bans)

---

## Phase 3: Market Data Pipeline

**Goal:** Developer can receive real-time and queryable market data from both exchanges

**Dependencies:** Phase 1, Phase 2 (both exchanges connected)

**Requirements:**

- MKTD-01: Subscribe to ticker updates for any symbol
- MKTD-02: Subscribe to order book updates (snapshot + delta)
- MKTD-03: Subscribe to trade updates for any symbol
- MKTD-04: Query symbol info (precision, filters, min notional)
- MKTD-05: Order book sequence continuity validation

**Success Criteria:**

1. Developer can subscribe to real-time ticker updates for any symbol on Binance or Bybit
2. Developer can subscribe to order book updates with initial snapshot followed by delta updates
3. Developer can subscribe to real-time trade updates for any symbol
4. Developer can query symbol information including price/quantity precision, min notional, and trading filters
5. Order book updates are validated for sequence continuity (detects and handles gaps/desync)

---

## Phase 4: Order Management & Account

**Goal:** Developer can manage orders and track account state with CQRS pattern

**Dependencies:** Phase 1, Phase 2, Phase 3 (connections and market data working)

**Requirements:**

- ORDR-01: Place LIMIT orders
- ORDR-02: Place MARKET orders
- ORDR-03: Cancel orders by order ID
- ORDR-04: Cancel orders by client order ID
- ORDR-05: Query order status
- ORDR-06: Query open orders for a symbol
- ORDR-07: Order state machine validation
- ORDR-08: Real-time order updates via WebSocket
- ORDR-09: Duplicate order update deduplication
- ACCT-01: Query asset balances (free + locked)
- ACCT-02: Subscribe to balance updates
- ACCT-03: Query positions for futures symbols
- ACCT-04: Subscribe to position updates

**Success Criteria:**

1. Developer can place LIMIT and MARKET orders on Binance and Bybit
2. Developer can cancel orders by exchange order ID or by client order ID
3. Developer can query individual order status and list all open orders for a symbol
4. Order state transitions are validated against state machine (prevents invalid transitions like FILLED→NEW)
5. Developer receives real-time order updates via user data WebSocket stream
6. Duplicate order updates are deduplicated by update ID to prevent state corruption
7. Developer can query current asset balances (free and locked portions)
8. Developer can subscribe to real-time balance and position updates

---

## Phase 5: Cross-Exchange Aggregation

**Goal:** Developer can aggregate market data across multiple exchanges for arbitrage and best execution

**Dependencies:** Phase 1, Phase 2, Phase 3 (all exchanges with market data)

**Requirements:**

- CRSX-01: Subscribe to aggregated order book across exchanges
- CRSX-02: Query best bid across all connected exchanges
- CRSX-03: Query best ask across all connected exchanges

**Success Criteria:**

1. Developer can subscribe to a merged order book that aggregates liquidity from Binance and Bybit
2. Developer can query the best bid price across all connected exchanges in a single call
3. Developer can query the best ask price across all connected exchanges in a single call

---

## Progress

| Phase                          | Status      | Requirements | Progress |
| ------------------------------ | ----------- | ------------ | -------- |
| 1 - Foundation & Binance       | Planned     | 13           | 0%       |
| 2 - Bybit Infrastructure       | Planned     | 3            | 0%       |
| 3 - Market Data Pipeline       | Not Started | 5            | 0%       |
| 4 - Order Management & Account | Not Started | 13           | 0%       |
| 5 - Cross-Exchange Aggregation | Not Started | 3            | 0%       |

**Total Progress:** 0/37 requirements (0%)

---

## Coverage Validation

| Category          | Requirements       | Phase Coverage           |
| ----------------- | ------------------ | ------------------------ |
| Core Connectivity | CONN-01 to CONN-09 | Phase 1 (6), Phase 2 (3) |
| Market Data       | MKTD-01 to MKTD-06 | Phase 1 (1), Phase 3 (5) |
| Order Management  | ORDR-01 to ORDR-09 | Phase 4 (9)              |
| Account Data      | ACCT-01 to ACCT-04 | Phase 4 (4)              |
| Resilience        | RESL-01 to RESL-06 | Phase 1 (6)              |
| Cross-Exchange    | CRSX-01 to CRSX-03 | Phase 5 (3)              |

**Coverage:** 37/37 requirements mapped (100%) ✓

---

## Architecture Reference

```
RootSupervisor (OneForOne, 5/10s)
├── CoordinatorActor (message routing)
├── HealthCheckActor (circuit breaker)
├── ClockSyncActor (timestamp alignment)
├── BinanceSupervisor (OneForOne, 10/60s)
│   ├── RateLimitGuardActor
│   ├── MarketDataActor
│   ├── OrderBookActor
│   ├── OrderCommandActor (CQRS Write)
│   └── OrderStateActor (CQRS Read)
└── BybitSupervisor (OneForOne, 10/60s)
    └── ... (same structure)
```

---

_Roadmap created: 2026-02-16_
_Based on research synthesis and requirement dependencies_

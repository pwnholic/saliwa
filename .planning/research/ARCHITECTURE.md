# Architecture: Cryptocurrency Exchange Connector

**Domain:** Multi-Exchange Connector Package  
**Researched:** 2026-02-16  
**Confidence:** HIGH (based on multiple production-grade sources and official Ergo documentation)

---

## Executive Summary

Production-grade cryptocurrency exchange connectors require a **layered, fault-tolerant architecture** that handles:

1. **Exchange protocol diversity** - Each exchange (Binance, Bybit, OKX) has unique WebSocket message formats, REST APIs, rate limits, and timestamp conventions
2. **Real-time data integrity** - Order books are event streams (not time series), requiring lossless sequencing, gap detection, and snapshot+delta management
3. **Operational resilience** - 24/7 markets demand automatic reconnection, circuit breakers, and supervisor-based failure recovery
4. **Cross-exchange concerns** - Symbol normalization, timestamp synchronization, and aggregated views across multiple venues

The recommended architecture combines:

- **Actor model** (Ergo Framework) for isolation, supervision, and message-passing concurrency
- **CQRS pattern** for order management (separate command/write and query/read paths)
- **Event-driven design** for real-time market data distribution
- **Circuit breaker + rate limiting** for fault tolerance

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              RootSupervisor                                      │
│                          (OneForOne, Intensity: 5/10s)                          │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                  │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────────────────┐   │
│  │ CoordinatorActor │  │ HealthCheckActor │  │     MetricsCollectorActor    │   │
│  │ (message router) │  │ (circuit breaker)│  │     (observability)          │   │
│  └────────┬─────────┘  └────────┬─────────┘  └──────────────────────────────┘   │
│           │                     │                                                │
│  ┌────────▼─────────────────────▼────────────────────────────────────────────┐  │
│  │                          AggregatorActor                                   │  │
│  │                   (cross-exchange data, routing)                          │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │                        PersistenceSupervisor                               │  │
│  │                      (AllForOne for data consistency)                      │  │
│  │  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐ ┌────────────────┐  │  │
│  │  │OrderStoreActor│ │PositionStore  │ │BalanceStore   │ │ ReconcilerActor│  │  │
│  │  │               │ │    Actor      │ │    Actor      │ │                │  │  │
│  │  └───────────────┘ └───────────────┘ └───────────────┘ └────────────────┘  │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
│  ┌──────────────────────────┐    ┌──────────────────────────┐                   │
│  │    BinanceSupervisor     │    │     BybitSupervisor      │    ... (per      │
│  │  (OneForOne, per-exchange)│    │  (OneForOne, per-exchange)│    exchange)    │
│  │                          │    │                          │                   │
│  │ ┌──────────────────────┐ │    │ ┌──────────────────────┐ │                   │
│  │ │   ClockSyncActor     │ │    │ │   ClockSyncActor     │ │                   │
│  │ │ (timestamp alignment)│ │    │ │ (timestamp alignment)│ │                   │
│  │ └──────────────────────┘ │    │ └──────────────────────┘ │                   │
│  │                          │    │                          │                   │
│  │ ┌──────────────────────┐ │    │ ┌──────────────────────┐ │                   │
│  │ │  RateLimitGuardActor │ │    │ │  RateLimitGuardActor │ │                   │
│  │ │ (token bucket, weight)│ │    │ │ (token bucket)       │ │                   │
│  │ └──────────────────────┘ │    │ └──────────────────────┘ │                   │
│  │                          │    │                          │                   │
│  │ ┌──────────────────────┐ │    │ ┌──────────────────────┐ │                   │
│  │ │   MarketDataActor    │ │    │ │   MarketDataActor    │ │                   │
│  │ │    (WebSocket)       │ │    │ │    (WebSocket)       │ │                   │
│  │ └──────────────────────┘ │    │ └──────────────────────┘ │                   │
│  │                          │    │                          │                   │
│  │ ┌──────────────────────┐ │    │ ┌──────────────────────┐ │                   │
│  │ │   OrderBookActor     │ │    │ │   OrderBookActor     │ │                   │
│  │ │ (snapshot + delta)   │ │    │ │ (snapshot + delta)   │ │                   │
│  │ └──────────────────────┘ │    │ └──────────────────────┘ │                   │
│  │                          │    │                          │                   │
│  │ ┌──────────┐ ┌──────────┐│    │ ┌──────────┐ ┌──────────┐│                   │
│  │ │OrderCmd  │ │OrderState││    │ │OrderCmd  │ │OrderState││                   │
│  │ │  Actor   │ │  Actor   ││    │ │  Actor   │ │  Actor   ││                   │
│  │ │ (CQRS W) │ │ (CQRS R) ││    │ │ (CQRS W) │ │ (CQRS R) ││                   │
│  │ └──────────┘ └──────────┘│    │ └──────────┘ └──────────┘│                   │
│  │                          │    │                          │                   │
│  │ ┌──────────────────────┐ │    │ ┌──────────────────────┐ │                   │
│  │ │PositionTrackerActor  │ │    │ │PositionTrackerActor  │ │                   │
│  │ └──────────────────────┘ │    │ └──────────────────────┘ │                   │
│  └──────────────────────────┘    └──────────────────────────┘                   │
│                                                                                  │
│  ┌────────────────────────────────────────────────────────────────────────────┐  │
│  │                          RiskManagerActor                                  │  │
│  │                    (position limits, exposure monitoring)                  │  │
│  └────────────────────────────────────────────────────────────────────────────┘  │
│                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Core Actors

### Root-Level Actors

| Actor                     | Responsibility                                          | Supervision Strategy                          | Dependencies              |
| ------------------------- | ------------------------------------------------------- | --------------------------------------------- | ------------------------- |
| **RootSupervisor**        | Top-level fault containment                             | OneForOne, Intensity: 5/10s, RestartTransient | None                      |
| **CoordinatorActor**      | Message routing, command dispatch                       | Restarts on failure                           | None                      |
| **HealthCheckActor**      | Circuit breaker state, exchange health monitoring       | Critical - restart immediately                | None                      |
| **AggregatorActor**       | Cross-exchange data aggregation, unified views          | Restart on failure                            | Exchange supervisors      |
| **RiskManagerActor**      | Position limits, exposure alerts, pre-trade risk checks | Critical - never die                          | OrderStore, PositionStore |
| **MetricsCollectorActor** | Prometheus-style metrics, latency tracking              | Non-critical                                  | None                      |

### Persistence Layer (PersistenceSupervisor)

| Actor                  | Responsibility                              | Supervision Strategy    | Key Operations                      |
| ---------------------- | ------------------------------------------- | ----------------------- | ----------------------------------- |
| **OrderStoreActor**    | Order persistence, history                  | AllForOne with siblings | GetOrder, SaveOrder, GetHistory     |
| **PositionStoreActor** | Position tracking per symbol                | AllForOne with siblings | GetPosition, UpdatePosition         |
| **BalanceStoreActor**  | Available balance per asset                 | AllForOne with siblings | GetBalance, ReserveBalance          |
| **ReconcilerActor**    | Periodic state reconciliation with exchange | AllForOne with siblings | ReconcileOrders, ReconcilePositions |

**Rationale for AllForOne**: Data stores must maintain consistency. If one store fails, all should restart together to prevent partial state.

### Per-Exchange Actors (ExchangeSupervisor)

| Actor                    | Responsibility                                                      | Supervision Strategy          | Message Types                     |
| ------------------------ | ------------------------------------------------------------------- | ----------------------------- | --------------------------------- |
| **ClockSyncActor**       | NTP-style clock sync with exchange, timestamp drift correction      | RestartTransient              | SyncClock, GetOffset              |
| **RateLimitGuardActor**  | Token bucket rate limiting, weight tracking (Binance), backpressure | RestartTransient              | Acquire, Release, GetState        |
| **MarketDataActor**      | WebSocket connection, subscription management, reconnection         | RestartTransient with backoff | Subscribe, Unsubscribe, Reconnect |
| **OrderBookActor**       | Order book reconstruction (snapshot + delta), sequence validation   | Restart on corruption         | GetBook, ApplyDelta, Reset        |
| **OrderCommandActor**    | CQRS write path - place/cancel orders via REST                      | RestartTransient              | PlaceOrder, CancelOrder           |
| **OrderStateActor**      | CQRS read path - order state queries, event projections             | Restart on corruption         | GetOrder, GetOrders, GetOpen      |
| **PositionTrackerActor** | Real-time position updates from fills, P&L calculation              | RestartTransient              | UpdatePosition, GetPosition       |

---

## Event System

### Event Topics

Events follow a hierarchical naming pattern for efficient subscription:

```
{type}.{exchange}.{symbol}
```

| Topic Pattern                   | Purpose                                           | Consumers                    |
| ------------------------------- | ------------------------------------------------- | ---------------------------- |
| `ticker.{exchange}.{symbol}`    | Price updates, 24h stats                          | Aggregator, Strategies       |
| `orderbook.{exchange}.{symbol}` | Order book snapshots/deltas                       | OrderBookActor, UI           |
| `order.{exchange}.{event}`      | Order lifecycle events (placed, filled, canceled) | OrderStateActor, RiskManager |
| `position.{exchange}.{symbol}`  | Position changes                                  | PositionTracker, RiskManager |
| `connection.{exchange}`         | Connection state changes                          | HealthCheck, Aggregator      |
| `health.{exchange}`             | Health status updates (healthy, degraded, down)   | HealthCheck, Alerting        |
| `risk.alert`                    | Risk limit breaches, exposure warnings            | Alerting, Logging            |

### Event Flow

```
Exchange WebSocket → MarketDataActor → Parse → Emit Event
                                              ↓
                              ┌───────────────────────────────┐
                              │        Event Router           │
                              │   (topic-based distribution)  │
                              └───────────────────────────────┘
                                     ↓           ↓           ↓
                              Subscriber1  Subscriber2  Subscriber3
```

### Event Subscription Pattern

```go
// Subscribe to all tickers from Binance
topic := "ticker.binance.*"
unsubscribe, err := connector.Subscribe(topic, func(evt Event) {
    ticker := evt.Message.(*domain.Ticker)
    // Handle ticker
})

// Subscribe to specific symbol order book
topic := "orderbook.binance.BTCUSDT"
unsubscribe, err := connector.Subscribe(topic, handler)
```

---

## CQRS Pattern

Order management uses **Command Query Responsibility Segregation** to separate write and read concerns:

### Write Path (OrderCommandActor)

```
┌──────────────┐     ┌───────────────────┐     ┌─────────────┐
│   Client     │────▶│ OrderCommandActor │────▶│  REST API   │
│              │     │   (Write Path)    │     │  (Exchange) │
└──────────────┘     └─────────┬─────────┘     └─────────────┘
                               │
                     Emits Events:
                     - OrderPlaced
                     - OrderCanceled
                     - OrderRejected
                               │
                               ▼
                     ┌─────────────────┐
                     │   Event Store   │
                     └─────────────────┘
```

**Responsibilities**:

- Validate order parameters (price, quantity, symbol filters)
- Apply risk checks via RiskManagerActor
- Sign and send to exchange REST API
- Handle rate limiting (wait/retry)
- Emit domain events on state changes

### Read Path (OrderStateActor)

```
┌──────────────┐     ┌───────────────────┐     ┌─────────────┐
│   Client     │────▶│  OrderStateActor  │────▶│  Response   │
│   (Query)    │     │   (Read Path)     │     │             │
└──────────────┘     └─────────┬─────────┘     └─────────────┘
                               │
                     Subscribes to Events:
                     - OrderPlaced
                     - OrderFilled (from WebSocket)
                     - OrderCanceled
                               │
                               ▼
                     ┌─────────────────┐
                     │  In-Memory Map  │
                     │  (Order State)  │
                     └─────────────────┘
```

**Responsibilities**:

- Subscribe to order events from all sources
- Maintain in-memory order state (fast reads)
- Handle out-of-order updates (sequence numbers)
- Persist state to OrderStoreActor periodically
- Serve queries (GetOrder, GetOpenOrders)

### Event Flow for Order Lifecycle

```
1. PlaceOrder Command
   └─▶ OrderCommandActor validates
       └─▶ Send to Exchange REST API
           └─▶ Emit OrderPlaced event
               └─▶ OrderStateActor updates state
               └─▶ PersistenceActor saves

2. WebSocket Order Update
   └─▶ MarketDataActor receives update
       └─▶ Emit OrderFilled/OrderCanceled event
           └─▶ OrderStateActor updates state
           └─▶ PersistenceActor saves

3. Query Order State
   └─▶ Client calls GetOrder
       └─▶ OrderStateActor returns from memory
```

### Why CQRS for Exchange Connectors

Based on research findings:

| Concern                 | Without CQRS                                    | With CQRS                                     |
| ----------------------- | ----------------------------------------------- | --------------------------------------------- |
| **Dual update sources** | REST + WebSocket updates race, state corruption | Events sequence correctly, idempotent updates |
| **Read latency**        | Database query per request                      | In-memory reads, sub-millisecond              |
| **Auditability**        | Hard to trace state changes                     | Every change is a stored event                |
| **Backtesting**         | Replaying state is complex                      | Events can be replayed exactly                |
| **Scaling**             | Read/write compete for same resource            | Scale independently                           |

---

## Data Flow

### Market Data Flow

```
Exchange                    Connector                              Consumer
────────                    ─────────                              ────────

WebSocket ─────▶ MarketDataActor ─────▶ Parse ─────▶ Emit Event
                    │                                      │
                    │                                      ▼
                    │                          ┌───────────────────┐
                    │                          │   Event Router    │
                    │                          └───────────────────┘
                    │                                      │
                    ▼                                      ▼
              OrderBookActor                         Subscribers
              (reconstruct book)                    (strategies, UI)
                    │
                    ▼
              GetBook() query
```

**Key Challenges Addressed**:

1. **Snapshot + Delta**: Order books start with full snapshot, then apply deltas
2. **Sequence Validation**: Each update has sequence number, detect gaps
3. **Out-of-Order**: Buffer updates until gap filled or timeout
4. **Reconnection**: Request fresh snapshot on reconnect

### Order Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Order Placement Flow                         │
└─────────────────────────────────────────────────────────────────────┘

Client                OrderCommandActor        Exchange          OrderStateActor
  │                         │                    │                    │
  │  PlaceOrder(req)        │                    │                    │
  │ ───────────────────────▶│                    │                    │
  │                         │                    │                    │
  │                         │ Validate           │                    │
  │                         │ ───────────────────────────────────────▶│
  │                         │                    │   Risk Check       │
  │                         │◀─────────────────────────────────────── │
  │                         │                    │   (allow/deny)     │
  │                         │                    │                    │
  │                         │ RateLimit.Acquire  │                    │
  │                         │ ────────┐          │                    │
  │                         │         │          │                    │
  │                         │◀────────┘          │                    │
  │                         │                    │                    │
  │                         │ POST /order        │                    │
  │                         │ ──────────────────▶│                    │
  │                         │                    │                    │
  │                         │    Order ID        │                    │
  │                         │◀────────────────── │                    │
  │                         │                    │                    │
  │                         │ Emit OrderPlaced   │                    │
  │                         │ ───────────────────────────────────────▶│
  │                         │                    │                    │
  │◀────────────────────────│                    │                    │
  │      Order ID           │                    │                    │
  │                         │                    │                    │
  │                         │                    │  WS Update         │
  │                         │                    │ ──────────────────▶│
  │                         │                    │                    │
  │                         │                    │    Emit OrderFill  │
  │                         │                    │◀───────────────────│
  │                         │                    │                    │
  │                         │◀─────────────────────────────────────── │
  │                         │   (subscribed to events)               │
```

### Cross-Exchange Aggregation Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Aggregated Market Data                           │
└─────────────────────────────────────────────────────────────────────┘

  Binance              Bybit              AggregatorActor
    │                    │                      │
    │ ticker.BTCUSDT     │                      │
    │ ─────────────────────────────────────────▶│
    │                    │                      │
    │                    │ ticker.BTCUSDT       │
    │                    │ ────────────────────▶│
    │                    │                      │
    │                    │                      │ Merge + Normalize
    │                    │                      │ (best bid/ask across venues)
    │                    │                      │
    │                    │                      │ Emit aggregated ticker
    │                    │                      │ ──────────────────▶ Subscribers
```

---

## Cross-Exchange Concerns

### Symbol Normalization

Exchanges use different symbol formats:

| Exchange | BTC/USDT Format | ETH-PERP Format    |
| -------- | --------------- | ------------------ |
| Binance  | `BTCUSDT`       | `ETHUSDT` (perp)   |
| Bybit    | `BTCUSDT`       | `ETHUSDT` (linear) |
| OKX      | `BTC-USDT`      | `ETH-USDT-SWAP`    |
| Deribit  | `BTC-PERPETUAL` | `ETH-PERPETUAL`    |

**Solution**: Internal canonical format with exchange-specific adapters:

```go
// Canonical symbol format: BASE-QUOTE-TYPE
// Examples: BTC-USDT-SPOT, ETH-USDT-PERP, BTC-USD-FUT-240329

type Symbol struct {
    Base  string  // BTC, ETH
    Quote string  // USDT, USD, USDC
    Type  SymbolType // SPOT, PERP, FUT, OPTION
    Expiry *time.Time // For dated futures
}

// Exchange adapters convert to/from canonical
func (b *BinanceAdapter) ToCanonical(symbol string) Symbol
func (b *BinanceAdapter) FromCanonical(sym Symbol) string
```

### Timestamp Synchronization

**Challenge**: Exchanges have different server times, and clock drift causes:

- Invalid signatures (timestamp outside window)
- Order book sequence gaps
- Incorrect latency measurements

**Solution**: ClockSyncActor per exchange

```go
type ClockSyncActor struct {
    act.Actor
    exchange    string
    offset      time.Duration  // Server time - Local time
    lastSync    time.Time
    driftRate   time.Duration  // ms/hour drift
}

func (c *ClockSyncActor) GetServerTime() time.Time {
    return time.Now().Add(c.offset)
}

func (c *ClockSyncActor) AdjustTimestamp(ts time.Time) time.Time {
    // Convert exchange timestamp to local time
    return ts.Add(-c.offset)
}
```

### Rate Limiting Per Exchange

Each exchange has unique rate limit structures:

| Exchange | REST Rate Limit          | WebSocket Limit             | Approach                  |
| -------- | ------------------------ | --------------------------- | ------------------------- |
| Binance  | Weight-based (1200/min)  | 5 connections, 1024 streams | Token bucket with weights |
| Bybit    | Request-based            | 10 connections              | Simple rate limiter       |
| OKX      | Request-based with tiers | 3 connections               | Tiered limiter            |

**RateLimitGuardActor Implementation**:

```go
type RateLimitGuardActor struct {
    act.Actor
    exchange     string
    limiter      RateLimiter  // Strategy per exchange
    pendingQueue []PendingRequest
}

func (r *RateLimitGuardActor) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
    switch req := request.(type) {
    case *message.AcquireLimit:
        if r.limiter.Allow(req.Weight) {
            return &message.Acquired{}, nil
        }
        // Queue request, return wait time
        waitTime := r.limiter.WaitTime(req.Weight)
        return nil, errors.NewRateLimitError(r.exchange, waitTime, req.Weight)
    }
}
```

---

## Supervision Strategy

### Supervisor Hierarchy and Strategies

```
RootSupervisor (OneForOne, Intensity: 5/10s)
├── CoordinatorActor (RestartTransient)
├── HealthCheckActor (RestartPermanent)
├── AggregatorActor (RestartTransient)
├── RiskManagerActor (RestartPermanent - never die)
├── MetricsCollectorActor (RestartTransient)
├── PersistenceSupervisor (AllForOne, Intensity: 3/60s)
│   ├── OrderStoreActor
│   ├── PositionStoreActor
│   ├── BalanceStoreActor
│   └── ReconcilerActor
├── BinanceSupervisor (OneForOne, Intensity: 10/60s)
│   ├── ClockSyncActor
│   ├── RateLimitGuardActor
│   ├── MarketDataActor
│   ├── OrderBookActor
│   ├── OrderCommandActor
│   ├── OrderStateActor
│   └── PositionTrackerActor
├── BybitSupervisor (OneForOne, Intensity: 10/60s)
│   └── ... (same structure)
└── RiskManagerActor (RestartPermanent)
```

### Strategy Selection Rationale

| Supervisor          | Strategy         | Rationale                                                              |
| ------------------- | ---------------- | ---------------------------------------------------------------------- |
| **Root**            | OneForOne        | Components are independent; one failure shouldn't restart all          |
| **Persistence**     | AllForOne        | Data stores must maintain consistency; restart all if one fails        |
| **Exchange**        | OneForOne        | Each actor is independent; MarketData failure shouldn't restart Orders |
| **Critical actors** | RestartPermanent | Must always be running (HealthCheck, RiskManager)                      |
| **Data actors**     | RestartTransient | Only restart on error, not normal termination                          |

### Restart Intensity Configuration

```go
// Root supervisor - allow 5 restarts in 10 seconds before giving up
RootSupervisorSpec: {
    Strategy: gen.SupervisorStrategyOneForOne,
    Intensity: 5,
    Period: 10,
    Restart: gen.SupervisorStrategyRestartTransient,
}

// Exchange supervisor - more lenient, network issues are common
ExchangeSupervisorSpec: {
    Strategy: gen.SupervisorStrategyOneForOne,
    Intensity: 10,
    Period: 60,
    Restart: gen.SupervisorStrategyRestartTransient,
}

// Persistence supervisor - conservative, data is critical
PersistenceSupervisorSpec: {
    Strategy: gen.SupervisorStrategyAllForOne,
    Intensity: 3,
    Period: 60,
    Restart: gen.SupervisorStrategyRestartTransient,
}
```

### Circuit Breaker Integration

HealthCheckActor implements circuit breaker per exchange:

```
         ┌─────────────────────────────────────────┐
         │          Circuit States                  │
         │                                          │
         │   CLOSED ──▶ (failures > threshold) ──▶ OPEN
         │     ▲                                    │
         │     │                                    │
         │     │                                    ▼
         │     └── (success) ─── HALF_OPEN ◀── (timeout)
         │                       │
         │                       │ (failure)
         │                       ▼
         │                     OPEN
         └─────────────────────────────────────────┘
```

---

## Suggested Build Order

Based on component dependencies and risk mitigation:

### Phase 1: Foundation (Week 1-2)

**Build Order**:

1. **Domain Models** (`pkg/domain/`)
   - Order, Ticker, OrderBook, Position, Balance
   - Decimal arithmetic helpers
   - No dependencies

2. **Error Types** (`pkg/errors/`)
   - Structured errors, sentinel errors
   - Rate limit errors, validation errors
   - No dependencies

3. **Configuration** (`pkg/config/`)
   - Exchange configs, credentials
   - No dependencies

4. **Mock Driver** (`internal/driver/mock/`)
   - Test infrastructure
   - Enables testing all other components

**Rationale**: Domain models are referenced everywhere. Define them first to avoid circular dependencies.

### Phase 2: Core Actors (Week 2-3)

**Build Order**:

1. **RootSupervisor** (`internal/sup/`)
   - Basic supervisor structure
   - Child spec definitions

2. **CoordinatorActor** (`internal/actor/`)
   - Message routing
   - Simple dispatch logic

3. **ClockSyncActor** (`internal/sync/`)
   - Timestamp management
   - Per-exchange offset tracking

4. **RateLimitGuardActor** (`internal/ratelimit/`)
   - Token bucket implementation
   - Weight-based limiting

**Rationale**: These actors have minimal dependencies and are needed by all exchange-specific actors.

### Phase 3: Exchange Infrastructure (Week 3-5)

**Build Order**:

1. **Driver Interface** (`pkg/driver/`)
   - REST and WebSocket interfaces
   - Driver registry

2. **REST Client Base** (`internal/driver/rest.go`)
   - Signing, authentication
   - Retry logic, error handling

3. **WebSocket Client Base** (`internal/driver/websocket.go`)
   - Connection management
   - Reconnection, heartbeat

4. **Binance Driver** (`internal/driver/binance/`)
   - REST client (order placement, account info)
   - WebSocket client (market data, user data)
   - Rate limiting (weight-based)

5. **Bybit Driver** (`internal/driver/bybit/`)
   - Similar structure to Binance
   - Different message formats

**Rationale**: Drivers abstract exchange-specific details. Build Binance first as reference implementation.

### Phase 4: Market Data Pipeline (Week 4-6)

**Build Order**:

1. **MarketDataActor** (`internal/market/`)
   - WebSocket subscription management
   - Message parsing and event emission

2. **OrderBookActor** (`internal/market/`)
   - Snapshot + delta reconstruction
   - Sequence validation
   - Gap detection and recovery

3. **Event System** (`internal/event/`)
   - Topic-based routing
   - Subscription management

**Rationale**: Market data is read-only and easier to test. Builds confidence before order management.

### Phase 5: Order Management (Week 5-7)

**Build Order**:

1. **OrderCommandActor** (`internal/order/`)
   - CQRS write path
   - Place/cancel order logic
   - Rate limit integration

2. **OrderStateActor** (`internal/order/`)
   - CQRS read path
   - Event subscription
   - State management

3. **RiskManagerActor** (`internal/risk/`)
   - Pre-trade risk checks
   - Position limits
   - Exposure monitoring

**Rationale**: Order management is critical. Build after market data to allow testing with real market conditions.

### Phase 6: Persistence & Reconciliation (Week 6-8)

**Build Order**:

1. **PersistenceSupervisor** (`internal/sup/`)
   - Supervisor for stores

2. **OrderStoreActor** (`internal/persistence/`)
   - Order persistence
   - History queries

3. **PositionStoreActor** (`internal/persistence/`)
   - Position tracking

4. **BalanceStoreActor** (`internal/persistence/`)
   - Balance management

5. **ReconcilerActor** (`internal/persistence/`)
   - Periodic state reconciliation
   - Detect and fix drift

**Rationale**: Persistence is critical but can be developed in parallel with other actors using mock stores.

### Phase 7: Cross-Exchange & Aggregation (Week 7-9)

**Build Order**:

1. **AggregatorActor** (`internal/actor/`)
   - Cross-exchange data aggregation
   - Best price calculation

2. **HealthCheckActor** (`internal/actor/`)
   - Circuit breaker implementation
   - Exchange health monitoring

3. **MetricsCollectorActor** (`internal/metrics/`)
   - Prometheus metrics
   - Latency tracking

**Rationale**: Aggregation requires multiple exchanges working. Build after individual exchanges are stable.

### Phase 8: Public API (Week 8-10)

**Build Order**:

1. **Connector Interface** (`pkg/connector/`)
   - Public API
   - Configuration

2. **Example Programs** (`examples/`)
   - Basic usage examples
   - Integration tests

**Rationale**: Public API should be last to ensure internal interfaces are stable.

---

## Key Architectural Decisions

### Decision 1: Actor Model (Ergo Framework)

**Chosen**: Ergo Framework actor model  
**Alternatives Considered**:

- Plain goroutines + channels
- Microservices with gRPC
- Event-driven with message queue

**Rationale**:

- **Supervision trees**: Built-in fault tolerance with restart strategies
- **Message passing**: Clean isolation, no shared state
- **Network transparency**: Can distribute across nodes if needed
- **Hot code reloading**: Can upgrade without downtime
- **Pattern matching**: Clean message routing with type switches

### Decision 2: CQRS for Orders

**Chosen**: Separate Command and State actors  
**Alternatives Considered**:

- Single OrderActor handling both reads and writes
- Database-backed state with caching

**Rationale**:

- **Dual update sources**: REST and WebSocket updates don't race
- **Read performance**: In-memory queries, no database hit
- **Auditability**: Every state change is an event
- **Backtesting**: Can replay events exactly

### Decision 3: Event-Driven for Market Data

**Chosen**: Event topics with pub/sub  
**Alternatives Considered**:

- Direct callbacks
- Message queue (Kafka, NATS)

**Rationale**:

- **Loose coupling**: Consumers don't know about producers
- **Multiple subscribers**: Same market data feeds strategies, UI, metrics
- **Backpressure**: Ergo handles slow consumers gracefully
- **No external dependency**: Built into Ergo framework

### Decision 4: Per-Exchange Supervisor

**Chosen**: Each exchange has its own supervisor subtree  
**Alternatives Considered**:

- All actors under single supervisor
- Flat actor structure

**Rationale**:

- **Fault isolation**: Binance failure doesn't affect Bybit
- **Independent scaling**: Can add exchanges without affecting existing
- **Clear ownership**: Each exchange is a bounded context
- **Targeted restarts**: Restart only affected exchange subtree

---

## Sources

### HIGH Confidence (Official Documentation, Context7)

1. **Ergo Framework Documentation** - Context7 `/ergo-services/ergo`
   - Supervisor strategies (OneForOne, AllForOne, RestForOne)
   - Actor lifecycle (Init, Terminate, HandleCall, HandleCast)
   - Message passing patterns
   - https://context7.com/ergo-services/ergo

2. **DolphinDB: Engineering Always-On Market Data Infrastructure** - Medium, Feb 2026
   - Production-grade WebSocket ingestion patterns
   - Snapshot + delta order book management
   - Fault tolerance with state machines (live/offline/replay)
   - https://medium.com/@DolphinDB_Inc/engineering-always-on-market-data-infrastructure-for-crypto-trading-f28615715dfb

### MEDIUM Confidence (Production Case Studies)

3. **Merehead: Crypto Exchange Architecture** - Dec 2025
   - Core components (trading engine, order book, wallet management)
   - Microservices vs monolith tradeoffs
   - Fault tolerance and high availability patterns
   - https://merehead.com/blog/crypto-exchange-architecture/

4. **CoinAPI: Why Real-Time Crypto Data Is Harder Than It Looks** - Dec 2025
   - Exchange protocol diversity
   - Symbol normalization challenges
   - Order book event-driven nature
   - Latency and timestamp issues
   - https://www.coinapi.io/blog/why-real-time-crypto-data-is-harder-than-it-looks

5. **CQRS in Real-Time Exchange Emulator** - Medium, Jul 2025
   - Practical CQRS implementation for trading systems
   - Event sourcing benefits
   - Write/read path separation
   - https://medium.com/@denis.volokh/building-a-real-time-exchange-emulator-using-cqrs-event-sourcing-and-htmx-6896bf05523b

6. **Crypto Data Pipeline with WebSocket and BigQuery** - Medium, Feb 2026
   - High-throughput WebSocket handling
   - Order book management patterns
   - Smart filtering optimizations
   - Latency and sequence gap monitoring
   - https://medium.com/@niujasmine01/building-a-real-time-crypto-data-pipeline-with-websocket-gcp-bigquery-and-cryptofeed-17e24045bd7f

---

## Research Confidence Assessment

| Area                       | Confidence | Reason                                                     |
| -------------------------- | ---------- | ---------------------------------------------------------- |
| **Actor Patterns**         | HIGH       | Direct from Ergo documentation via Context7                |
| **Supervision Strategies** | HIGH       | Well-documented in Ergo, proven in Erlang/OTP              |
| **CQRS for Trading**       | HIGH       | Multiple production implementations documented             |
| **Exchange Protocols**     | MEDIUM     | Based on web research, may need driver-specific validation |
| **Rate Limiting**          | MEDIUM     | Exchange-specific, needs per-exchange documentation lookup |
| **Order Book Management**  | HIGH       | Well-documented patterns across sources                    |
| **Fault Tolerance**        | HIGH       | Industry-standard patterns (circuit breaker, supervisor)   |

---

## Gaps for Phase-Specific Research

| Phase              | Research Needed                                             | Why                                   |
| ------------------ | ----------------------------------------------------------- | ------------------------------------- |
| **Binance Driver** | API v3 specifics, rate limit weights, testnet endpoints     | Exchange-specific, changes frequently |
| **Bybit Driver**   | API v5 specifics, authentication, WebSocket message formats | Exchange-specific                     |
| **Persistence**    | Database choice (PostgreSQL, TimescaleDB, Redis)            | Depends on scale requirements         |
| **Metrics**        | Prometheus integration, custom metrics                      | Implementation detail                 |
| **Deployment**     | Container orchestration, configuration management           | Infrastructure-specific               |

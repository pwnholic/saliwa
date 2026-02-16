# Multi-Exchange Connector Package - Implementation Plan

## Executive Summary

This document outlines the implementation plan for a production-grade Go library that connects to cryptocurrency exchanges (Binance & Bybit) using an Actor-based, Event-driven architecture powered by Ergo Framework. The package is designed as a reusable library that can be imported by various trading engines.

## Technical Architecture

### Technology Stack

| Category               | Technology             | Version | Rationale                                          |
| ---------------------- | ---------------------- | ------- | -------------------------------------------------- |
| **Actor Framework**    | Ergo Framework         | latest  | Built-in supervision, events, network transparency |
| **Decimal Arithmetic** | cockroachdb/apd        | latest  | Arbitrary precision for financial calculations     |
| **HTTP Client**        | resty.dev              | latest  | Built-in retry, middleware, clean API              |
| **WebSocket**          | lxzan/gws              | latest  | High performance, low allocation                   |
| **Logging**            | rs/zerolog             | latest  | Zero allocation structured logging                 |
| **Rate Limiting**      | golang.org/x/time/rate | latest  | Token bucket implementation                        |
| **Concurrency**        | golang.org/x/sync      | latest  | singleflight for deduplication                     |
| **Go Version**         | 1.25                   | latest  | Latest stable with modern features                 |

### System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    USER APPLICATION                             │
│   import "github.com/user/exchange-connector/pkg/connector"     │
└─────────────────────────────┬───────────────────────────────────┘
                              │ Public API
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│                  CONNECTOR PACKAGE (Library)                     │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                 ERGO NODE (Embedded)                        │ │
│  │                                                             │ │
│  │  RootSupervisor (OneForOne)                                 │ │
│  │    ├── CoordinatorActor (message routing)                   │ │
│  │    ├── HealthCheckActor (circuit breaker)                   │ │
│  │    ├── AggregatorActor (cross-exchange)                     │ │
│  │    ├── RiskManagerActor                                     │ │
│  │    ├── MetricsCollectorActor                                │ │
│  │    ├── PersistenceSupervisor                                │ │
│  │    │     ├── OrderStoreActor                                │ │
│  │    │     ├── PositionStoreActor                             │ │
│  │    │     ├── BalanceStoreActor                              │ │
│  │    │     └── ReconcilerActor                                │ │
│  │    │                                                        │ │
│  │    ├── BinanceSupervisor                                    │ │
│  │    │     ├── ClockSyncActor (time synchronization)          │ │
│  │    │     ├── MarketDataActor (gws WebSocket)                │ │
│  │    │     ├── OrderCommandActor (write path)                 │ │
│  │    │     ├── OrderStateActor (read path)                    │ │
│  │    │     ├── OrderBookActor                                 │ │
│  │    │     ├── PositionTrackerActor                           │ │
│  │    │     └── RateLimitGuardActor                            │ │
│  │    │                                                        │ │
│  │    └── BybitSupervisor                                      │ │
│  │          ├── ClockSyncActor (time synchronization)          │ │
│  │          ├── MarketDataActor (gws WebSocket)                │ │
│  │          ├── OrderCommandActor (write path)                 │ │
│  │          ├── OrderStateActor (read path)                    │ │
│  │          ├── OrderBookActor                                 │ │
│  │          ├── PositionTrackerActor                           │ │
│  │          └── RateLimitGuardActor                            │ │
│  │                                                             │ │
│  │  EVENT SYSTEM (Ergo Built-in)                               │ │
│  │    - ticker.{exchange}.{symbol}                             │ │
│  │    - orderbook.{exchange}.{symbol}                          │ │
│  │    - trade.{exchange}.{symbol}                              │ │
│  │    - order.{exchange}.{event}                               │ │
│  │    - position.{exchange}.{symbol}                           │ │
│  │    - connection.{exchange}                                  │ │
│  │    - health.{exchange}                                      │ │
│  │    - risk.alert                                             │ │
│  │    - metrics.{type}                                         │ │
│  │                                                             │ │
│  └─────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────┘
```

## Package Structure (Revised)

```
exchange-connector/
├── go.mod
├── go.sum
├── LICENSE
├── README.md
├── ARCHITECTURE.md
├── RUNBOOK.md
├── CHANGELOG.md
├── Makefile
│
├── pkg/                         # Public API (exported packages)
│   │
│   ├── connector/               # Main connector package
│   │   ├── connector.go         # Connector struct and methods
│   │   ├── options.go           # Functional options
│   │   └── connector_test.go
│   │
│   ├── domain/                  # Domain models (shared types)
│   │   ├── decimal.go
│   │   ├── ticker.go
│   │   ├── order.go
│   │   ├── orderbook.go
│   │   ├── trade.go
│   │   ├── position.go
│   │   ├── balance.go
│   │   ├── symbol.go
│   │   ├── risk.go
│   │   ├── aggregation.go
│   │   ├── health.go
│   │   ├── metrics.go
│   │   └── enums.go
│   │
│   ├── config/                  # Configuration types
│   │   ├── config.go
│   │   ├── exchange.go
│   │   ├── risk.go
│   │   ├── persistence.go
│   │   └── aggregation.go
│   │
│   ├── errors/                  # Public error types
│   │   ├── errors.go
│   │   ├── exchange.go
│   │   ├── validation.go
│   │   └── sentinels.go
│   │
│   ├── driver/                  # Driver interface (public)
│   │   ├── driver.go
│   │   └── registry.go
│   │
│   ├── persistence/             # Persistence interfaces
│   │   ├── backend.go
│   │   ├── store.go
│   │   └── memory/
│   │       └── memory.go
│   │
│   ├── risk/                    # Risk management (public API)
│   │   ├── hook.go
│   │   ├── manager.go
│   │   └── rules.go
│   │
│   └── circuit/                 # Circuit breaker (public)
│       ├── breaker.go
│       └── config.go
│
├── internal/                    # Internal implementation (not exported)
│   │
│   ├── app/                    # Ergo Application
│   │   └── connector.go
│   │
│   ├── sup/                    # Supervisors
│   │   ├── root.go
│   │   ├── exchange.go
│   │   └── persistence.go
│   │
│   ├── actor/                  # Core actors
│   │   ├── coordinator.go
│   │   ├── aggregator.go
│   │   ├── health.go
│   │   ├── risk.go
│   │   ├── metrics.go
│   │   └── reconciler.go
│   │
│   ├── market/                 # Market data layer
│   │   ├── ws.go
│   │   ├── orderbook.go
│   │   ├── snapshot.go
│   │   ├── sequencer.go
│   │   └── normalizer.go
│   │
│   ├── order/                  # Order management (CQRS)
│   │   ├── command.go
│   │   ├── state.go
│   │   ├── executor.go
│   │   ├── sequencer.go
│   │   ├── idempotency.go
│   │   ├── deduplicator.go
│   │   └── tracker.go
│   │
│   ├── sync/                   # Time & nonce sync
│   │   ├── clock.go
│   │   ├── nonce.go
│   │   └── timesource.go
│   │
│   ├── ratelimit/              # Rate limiting
│   │   ├── guard.go
│   │   ├── bucket.go
│   │   └── weight.go
│   │
│   ├── message/                # Internal messages
│   │   ├── market.go
│   │   ├── order.go
│   │   ├── position.go
│   │   ├── sync.go
│   │   ├── health.go
│   │   └── internal.go
│   │
│   ├── event/                  # Event system
│   │   ├── topics.go
│   │   ├── handler.go
│   │   └── router.go
│   │
│   ├── driver/                 # Driver implementations
│   │   │
│   │   ├── binance/
│   │   │   ├── driver.go
│   │   │   ├── ws.go
│   │   │   ├── ws_handlers.go
│   │   │   ├── rest.go
│   │   │   ├── rest_market.go
│   │   │   ├── rest_trading.go
│   │   │   ├── rest_account.go
│   │   │   ├── signer.go
│   │   │   ├── ratelimit.go
│   │   │   ├── types.go
│   │   │   ├── types_request.go
│   │   │   ├── types_response.go
│   │   │   ├── types_ws.go
│   │   │   ├── mapper.go
│   │   │   ├── mapper_market.go
│   │   │   ├── mapper_order.go
│   │   │   ├── normalizer.go
│   │   │   ├── errors.go
│   │   │   └── config.go
│   │   │
│   │   ├── bybit/
│   │   │   ├── driver.go
│   │   │   ├── ws.go
│   │   │   ├── ws_public.go
│   │   │   ├── ws_private.go
│   │   │   ├── ws_handlers.go
│   │   │   ├── rest.go
│   │   │   ├── rest_market.go
│   │   │   ├── rest_trading.go
│   │   │   ├── rest_account.go
│   │   │   ├── signer.go
│   │   │   ├── ratelimit.go
│   │   │   ├── types.go
│   │   │   ├── types_request.go
│   │   │   ├── types_response.go
│   │   │   ├── types_ws.go
│   │   │   ├── mapper.go
│   │   │   ├── mapper_market.go
│   │   │   ├── mapper_order.go
│   │   │   ├── normalizer.go
│   │   │   ├── errors.go
│   │   │   └── config.go
│   │   │
│   │   └── mock/
│   │       ├── mock.go
│   │       ├── triggers.go
│   │       └── scenarios.go
│   │
│   ├── log/                    # Logging adapter
│   │   ├── adapter.go
│   │   └── logger.go
│   │
│   ├── metrics/                # Metrics implementation
│   │   ├── collector.go
│   │   ├── types.go
│   │   └── registry.go
│   │
│   └── util/                   # Internal utilities
│       ├── retry.go
│       ├── backoff.go
│       └── validator.go
│
├── docs/                        # Documentation
│   ├── api/
│   │   └── connector.md
│   │
│   ├── sequence/
│   │   ├── order-placement.md
│   │   ├── order-cancel.md
│   │   ├── reconnection.md
│   │   ├── reconciliation.md
│   │   └── clock-sync.md
│   │
│   ├── state-machines/
│   │   ├── order-lifecycle.md
│   │   ├── circuit-breaker.md
│   │   └── connection.md
│   │
│   └── guides/
│       ├── adding-exchange.md
│       ├── error-handling.md
│       └── testing.md
│
├── test/                        # Integration tests
│   ├── integration/
│   │   ├── binance_test.go
│   │   ├── bybit_test.go
│   │   └── aggregation_test.go
│   │
│   ├── fixtures/
│   │   ├── binance/
│   │   └── bybit/
│   │
│   └── testutil/
│       ├── mock_actor.go
│       └── helpers.go
│
└── examples/                    # Usage examples
    ├── 01-basic/
    │   ├── main.go
    │   └── README.md
    │
    ├── 02-market-data/
    │   ├── main.go
    │   └── README.md
    │
    ├── 03-trading/
    │   ├── main.go
    │   └── README.md
    │
    ├── 04-multi-exchange/
    │   ├── main.go
    │   └── README.md
    │
    ├── 05-aggregation/
    │   ├── main.go
    │   └── README.md
    │
    ├── 06-risk-hooks/
    │   ├── main.go
    │   └── README.md
    │
    ├── 07-circuit-breaker/
    │   ├── main.go
    │   └── README.md
    │
    └── 08-testing/
        ├── main_test.go
        └── README.md
```

## Implementation Phases

### Phase 1: Foundation & Core Domain (Days 1-5)

---

#### Task 1.1: Project Setup & Dependencies

**Objective**: Initialize Go module with all required dependencies.

**Files**:

- `go.mod`, `go.sum`
- `.gitignore`
- `LICENSE`
- `CHANGELOG.md`
- `Makefile`

**Pseudocode**:

```
INIT go module with name "exchange-connector"
ADD dependencies: ergo, apd, resty, gws, zerolog, sync, time
CREATE .gitignore with standard Go excludes
CREATE Makefile with targets: build, test, lint, clean
```

**Acceptance Criteria**:

- `go mod tidy` succeeds
- `go build ./...` compiles
- `make test` runs

**Estimated Time**: 2 hours

---

#### Task 1.2: Domain - Decimal Type

**Objective**: Create decimal wrapper with helper functions.

**Files**: `pkg/domain/decimal.go`

**Pseudocode**:

```
TYPE Decimal = *apd.Decimal

FUNCTION NewDecimal(string) -> (Decimal, error):
    CREATE new apd.Decimal
    SET value from string
    RETURN decimal or error

FUNCTION MustDecimal(string) -> Decimal:
    CALL NewDecimal
    IF error THEN panic
    RETURN decimal

FUNCTION NewDecimalFromInt64(int64) -> Decimal
FUNCTION NewDecimalFromFloat64(float64) -> Decimal
FUNCTION Zero() -> Decimal
FUNCTION IsZero(Decimal) -> bool
FUNCTION Compare(a, b Decimal) -> int
FUNCTION Add(a, b Decimal) -> Decimal
FUNCTION Sub(a, b Decimal) -> Decimal
FUNCTION Mul(a, b Decimal) -> Decimal
FUNCTION Div(a, b Decimal) -> Decimal
FUNCTION String(Decimal) -> string
FUNCTION Float64(Decimal) -> (float64, error)
```

**Tests**: Basic arithmetic, precision, edge cases

**Estimated Time**: 3 hours

---

#### Task 1.3: Domain - Enums

**Objective**: Define all shared enumerations.

**Files**: `pkg/domain/enums.go`

**Pseudocode**:

```
TYPE Side = string
CONST SideBuy = "BUY"
CONST SideSell = "SELL"

TYPE OrderType = string
CONST OrderTypeMarket = "MARKET"
CONST OrderTypeLimit = "LIMIT"
CONST OrderTypeStopLimit = "STOP_LIMIT"
CONST OrderTypeStopMarket = "STOP_MARKET"

TYPE OrderStatus = string
CONST OrderStatusNew = "NEW"
CONST OrderStatusPending = "PENDING"
CONST OrderStatusPartiallyFilled = "PARTIALLY_FILLED"
CONST OrderStatusFilled = "FILLED"
CONST OrderStatusCanceling = "CANCELING"
CONST OrderStatusCanceled = "CANCELED"
CONST OrderStatusRejected = "REJECTED"
CONST OrderStatusExpired = "EXPIRED"

TYPE TimeInForce = string
CONST TimeInForceGTC = "GTC"
CONST TimeInForceIOC = "IOC"
CONST TimeInForceFOK = "FOK"

TYPE PositionSide = string
CONST PositionSideLong = "LONG"
CONST PositionSideShort = "SHORT"

TYPE ExchangeType = string
CONST ExchangeBinance = "binance"
CONST ExchangeBybit = "bybit"

TYPE RiskSeverity = string
CONST SeverityInfo = "INFO"
CONST SeverityWarning = "WARNING"
CONST SeverityCritical = "CRITICAL"

TYPE RiskRuleType = string
CONST RuleMaxOrderSize = "MAX_ORDER_SIZE"
CONST RuleMaxPositionSize = "MAX_POSITION_SIZE"
CONST RuleDailyLoss = "DAILY_LOSS"
CONST RuleMaxDrawdown = "MAX_DRAWDOWN"

TYPE RiskAction = string
CONST ActionWarn = "WARN"
CONST ActionReject = "REJECT"
CONST ActionCloseAll = "CLOSE_ALL"

TYPE ConnectionStatus = string
CONST StatusDisconnected = "DISCONNECTED"
CONST StatusConnecting = "CONNECTING"
CONST StatusConnected = "CONNECTED"
CONST StatusReconnecting = "RECONNECTING"

TYPE CircuitState = string
CONST CircuitClosed = "CLOSED"
CONST CircuitOpen = "OPEN"
CONST CircuitHalfOpen = "HALF_OPEN"
```

**Estimated Time**: 2 hours

---

#### Task 1.4: Domain - Market Data Models

**Objective**: Create market data models.

**Files**:

- `pkg/domain/ticker.go`
- `pkg/domain/orderbook.go`
- `pkg/domain/trade.go`

**Pseudocode**:

```
// ticker.go
TYPE Ticker STRUCT:
    Exchange string
    Symbol string
    Bid Decimal
    Ask Decimal
    BidQty Decimal
    AskQty Decimal
    Last Decimal
    Volume24h Decimal
    High24h Decimal
    Low24h Decimal
    Timestamp time.Time

// orderbook.go
TYPE OrderBookLevel STRUCT:
    Price Decimal
    Quantity Decimal

TYPE OrderBook STRUCT:
    Exchange string
    Symbol string
    Bids []OrderBookLevel
    Asks []OrderBookLevel
    Timestamp time.Time
    SequenceNum int64
    Checksum string (optional)

FUNCTION (ob *OrderBook) UpdateBids(levels []OrderBookLevel)
FUNCTION (ob *OrderBook) UpdateAsks(levels []OrderBookLevel)
FUNCTION (ob *OrderBook) BestBid() *OrderBookLevel
FUNCTION (ob *OrderBook) BestAsk() *OrderBookLevel
FUNCTION (ob *OrderBook) Spread() Decimal
FUNCTION (ob *OrderBook) Validate() error

// trade.go
TYPE Trade STRUCT:
    Exchange string
    Symbol string
    TradeID string
    Price Decimal
    Quantity Decimal
    Side Side
    Timestamp time.Time
    IsBuyerMaker bool
```

**Estimated Time**: 4 hours

---

#### Task 1.5: Domain - Order Models

**Objective**: Create order models with state machine.

**Files**: `pkg/domain/order.go`

**Pseudocode**:

```
TYPE Order STRUCT:
    ID string
    ClientOrderID string
    Exchange string
    Symbol string
    Side Side
    Type OrderType
    Status OrderStatus
    Price Decimal
    Quantity Decimal
    FilledQty Decimal
    AvgFillPrice Decimal
    TimeInForce TimeInForce
    StopPrice Decimal (optional)
    CreatedAt time.Time
    UpdatedAt time.Time
    Metadata map[string]interface{}

TYPE OrderRequest STRUCT:
    Exchange string
    Symbol string
    Side Side
    Type OrderType
    Price Decimal (optional for market)
    Quantity Decimal
    TimeInForce TimeInForce
    StopPrice Decimal (optional)
    ClientOrderID string (optional)

TYPE CancelRequest STRUCT:
    Exchange string
    OrderID string
    ClientOrderID string (optional)

TYPE OrderEvent ENUM:
    EventCreated
    EventPending
    EventPartialFill
    EventFilled
    EventCancelRequested
    EventCanceled
    EventRejected
    EventExpired

MAP ValidTransitions:
    NEW -> [PENDING, PARTIALLY_FILLED, FILLED, REJECTED]
    PENDING -> [PARTIALLY_FILLED, FILLED, CANCELING, REJECTED]
    PARTIALLY_FILLED -> [FILLED, CANCELING, CANCELED]
    CANCELING -> [CANCELED, PARTIALLY_FILLED, FILLED]
    Others -> []

FUNCTION (o *Order) CanTransition(newStatus OrderStatus) bool:
    GET valid statuses FROM ValidTransitions[o.Status]
    RETURN newStatus IN valid statuses

FUNCTION (o *Order) Transition(newStatus OrderStatus) error:
    IF NOT CanTransition(newStatus) THEN
        RETURN error "invalid transition"
    SET o.Status = newStatus
    SET o.UpdatedAt = now()
    RETURN nil

FUNCTION (o *Order) IsFinal() bool:
    RETURN Status IN [FILLED, CANCELED, REJECTED, EXPIRED]

FUNCTION (o *Order) IsActive() bool:
    RETURN Status IN [NEW, PENDING, PARTIALLY_FILLED, CANCELING]

FUNCTION (o *Order) RemainingQty() Decimal:
    RETURN Quantity - FilledQty
```

**Estimated Time**: 5 hours

---

#### Task 1.6: Domain - Position & Balance Models

**Objective**: Create position and balance tracking models.

**Files**:

- `pkg/domain/position.go`
- `pkg/domain/balance.go`

**Pseudocode**:

```
// position.go
TYPE Position STRUCT:
    Exchange string
    Symbol string
    Side PositionSide
    Quantity Decimal
    AvgEntryPrice Decimal
    MarkPrice Decimal
    LiquidationPrice Decimal (optional)
    UnrealizedPnL Decimal
    RealizedPnL Decimal
    Leverage int
    Timestamp time.Time

FUNCTION (p *Position) UpdateMarkPrice(price Decimal):
    SET p.MarkPrice = price
    CALL p.CalculateUnrealizedPnL()

FUNCTION (p *Position) CalculateUnrealizedPnL():
    IF Side == LONG THEN
        SET p.UnrealizedPnL = (MarkPrice - AvgEntryPrice) * Quantity
    ELSE
        SET p.UnrealizedPnL = (AvgEntryPrice - MarkPrice) * Quantity

FUNCTION (p *Position) AddFill(price, qty Decimal):
    totalQty = Quantity + qty
    totalCost = (Quantity * AvgEntryPrice) + (qty * price)
    SET p.AvgEntryPrice = totalCost / totalQty
    SET p.Quantity = totalQty
    CALL p.CalculateUnrealizedPnL()

FUNCTION (p *Position) ReduceFill(price, qty Decimal):
    IF qty >= Quantity THEN
        realizedPnL = calculate closed PnL
        SET p.RealizedPnL += realizedPnL
        SET p.Quantity = 0
    ELSE
        realizedPnL = calculate partial PnL
        SET p.RealizedPnL += realizedPnL
        SET p.Quantity -= qty

// balance.go
TYPE Balance STRUCT:
    Exchange string
    Asset string
    Free Decimal
    Locked Decimal
    Timestamp time.Time

FUNCTION (b *Balance) Total() Decimal:
    RETURN Free + Locked

FUNCTION (b *Balance) Update(free, locked Decimal):
    SET b.Free = free
    SET b.Locked = locked
    SET b.Timestamp = now()
```

**Estimated Time**: 4 hours

---

#### Task 1.7: Domain - Symbol Info

**Objective**: Create symbol information models.

**Files**: `pkg/domain/symbol.go`

**Pseudocode**:

```
TYPE PriceFilter STRUCT:
    MinPrice Decimal
    MaxPrice Decimal
    TickSize Decimal

TYPE LotSizeFilter STRUCT:
    MinQty Decimal
    MaxQty Decimal
    StepSize Decimal

TYPE SymbolInfo STRUCT:
    Exchange string
    Symbol string
    BaseAsset string
    QuoteAsset string
    Status string
    PriceFilter PriceFilter
    LotSizeFilter LotSizeFilter
    MinNotional Decimal
    Permissions []string

FUNCTION (s *SymbolInfo) ValidatePrice(price Decimal) error:
    IF price < MinPrice OR price > MaxPrice THEN
        RETURN error "price out of range"
    IF price MOD TickSize != 0 THEN
        RETURN error "invalid tick size"
    RETURN nil

FUNCTION (s *SymbolInfo) ValidateQuantity(qty Decimal) error:
    IF qty < MinQty OR qty > MaxQty THEN
        RETURN error "quantity out of range"
    IF qty MOD StepSize != 0 THEN
        RETURN error "invalid step size"
    RETURN nil

FUNCTION (s *SymbolInfo) RoundPrice(price Decimal) Decimal:
    RETURN round to nearest TickSize

FUNCTION (s *SymbolInfo) RoundQuantity(qty Decimal) Decimal:
    RETURN round to nearest StepSize
```

**Estimated Time**: 3 hours

---

#### Task 1.8: Domain - Risk Models

**Objective**: Create risk management models.

**Files**: `pkg/domain/risk.go`

**Pseudocode**:

```
TYPE RiskAlert STRUCT:
    ID string
    Severity RiskSeverity
    RuleType RiskRuleType
    Message string
    Exchange string
    Symbol string (optional)
    Data map[string]interface{}
    Timestamp time.Time

TYPE RiskMetrics STRUCT:
    TotalExposure Decimal
    DailyPnL Decimal
    MaxDrawdown Decimal
    OpenPositions int
    DailyTradeCount int
    LastUpdated time.Time
```

**Estimated Time**: 2 hours

---

#### Task 1.9: Domain - Aggregation Models

**Objective**: Create cross-exchange aggregation models.

**Files**: `pkg/domain/aggregation.go`

**Pseudocode**:

```
TYPE BookSource STRUCT:
    Exchange string
    Price Decimal
    Quantity Decimal

TYPE AggregatedLevel STRUCT:
    Price Decimal
    TotalQuantity Decimal
    Sources []BookSource

TYPE AggregatedOrderBook STRUCT:
    Symbol string
    Bids []AggregatedLevel
    Asks []AggregatedLevel
    Timestamp time.Time

TYPE AggregatedTicker STRUCT:
    Symbol string
    BestBid Decimal
    BestAsk Decimal
    BestBidExchange string
    BestAskExchange string
    Spread Decimal
    Exchanges []string
    Timestamp time.Time

TYPE ArbitrageOpportunity STRUCT:
    Symbol string
    BuyExchange string
    SellExchange string
    BuyPrice Decimal
    SellPrice Decimal
    Spread Decimal
    SpreadPercent Decimal
    PotentialProfit Decimal
    Timestamp time.Time
    Valid bool

FUNCTION (a *AggregatedOrderBook) Merge(books map[string]*OrderBook):
    CLEAR existing levels
    FOR EACH exchange, book IN books DO
        MERGE bids into aggregated bids
        MERGE asks into aggregated asks
    SORT bids descending by price
    SORT asks ascending by price

FUNCTION (a *AggregatedTicker) Update(tickers map[string]*Ticker):
    bestBid = DECIMAL_ZERO
    bestAsk = DECIMAL_MAX
    FOR EACH exchange, ticker IN tickers DO
        IF ticker.Bid > bestBid THEN
            SET bestBid = ticker.Bid
            SET BestBidExchange = exchange
        IF ticker.Ask < bestAsk THEN
            SET bestAsk = ticker.Ask
            SET BestAskExchange = exchange
    SET Spread = bestAsk - bestBid
```

**Estimated Time**: 4 hours

---

#### Task 1.10: Domain - Health & Metrics Models

**Objective**: Create health check and metrics models.

**Files**:

- `pkg/domain/health.go`
- `pkg/domain/metrics.go`

**Pseudocode**:

```
// health.go
TYPE HealthStatus STRUCT:
    Exchange string
    Status ConnectionStatus
    CircuitState CircuitState
    LastCheck time.Time
    Healthy bool
    Latency time.Duration
    ErrorCount int
    SuccessCount int
    Message string

TYPE HealthCheck STRUCT:
    CheckInterval time.Duration
    Threshold int
    Timeout time.Duration

// metrics.go
TYPE OrderMetrics STRUCT:
    TotalOrders int64
    FilledOrders int64
    CanceledOrders int64
    RejectedOrders int64
    AvgLatency time.Duration
    ErrorRate float64

TYPE ConnectionMetrics STRUCT:
    Uptime time.Duration
    ReconnectCount int
    MessageRate float64
    BytesReceived int64
    BytesSent int64

TYPE Metrics STRUCT:
    Exchange string
    Orders OrderMetrics
    Connection ConnectionMetrics
    Timestamp time.Time
```

**Estimated Time**: 3 hours

---

#### Task 1.11: Configuration Types

**Objective**: Create all configuration structures.

**Files**:

- `pkg/config/config.go`
- `pkg/config/exchange.go`
- `pkg/config/risk.go`
- `pkg/config/persistence.go`
- `pkg/config/aggregation.go`

**Pseudocode**:

```
// config.go
TYPE Config STRUCT:
    NodeName string
    Exchanges []ExchangeConfig
    Risk RiskConfig
    Persistence PersistenceConfig
    Aggregation AggregationConfig
    LogLevel string

// exchange.go
TYPE ExchangeConfig STRUCT:
    Name string
    Type ExchangeType
    APIKey string
    APISecret string
    Testnet bool
    RateLimit RateLimitConfig
    Timeout time.Duration
    MaxReconnectAttempts int
    ReconnectDelay time.Duration

TYPE RateLimitConfig STRUCT:
    Enabled bool
    RequestsPerSecond int
    BurstSize int
    WeightLimit int (for Binance)

// risk.go
TYPE RiskConfig STRUCT:
    Enabled bool
    Rules []RiskRuleConfig
    MaxOrderSize Decimal
    MaxPositionSize Decimal
    DailyLossLimit Decimal

TYPE RiskRuleConfig STRUCT:
    Type RiskRuleType
    Action RiskAction
    Threshold Decimal
    Enabled bool

// persistence.go
TYPE PersistenceConfig STRUCT:
    Backend string (memory, postgres, redis)
    ConnectionString string
    RetentionDays int

// aggregation.go
TYPE AggregationConfig STRUCT:
    Enabled bool
    UpdateInterval time.Duration
    Symbols []string
```

**Estimated Time**: 3 hours

---

#### Task 1.12: Error Types

**Objective**: Create structured error types.

**Files**:

- `pkg/errors/errors.go`
- `pkg/errors/exchange.go`
- `pkg/errors/validation.go`
- `pkg/errors/sentinels.go`

**Pseudocode**:

```
// errors.go
TYPE Error STRUCT:
    Code string
    Message string
    Details map[string]interface{}
    Err error (wrapped)

FUNCTION (e *Error) Error() string
FUNCTION (e *Error) Unwrap() error
FUNCTION (e *Error) WithDetail(key, value) *Error

// exchange.go
TYPE ExchangeError STRUCT:
    Exchange string
    Code int
    Message string
    Original error

TYPE RateLimitError STRUCT:
    Exchange string
    RetryAfter time.Duration
    Limit int

TYPE ConnectionError STRUCT:
    Exchange string
    Reason string
    Temporary bool

// validation.go
TYPE ValidationError STRUCT:
    Field string
    Value interface{}
    Constraint string

// sentinels.go
VAR ErrNotRunning = errors.New("connector not running")
VAR ErrAlreadyRunning = errors.New("connector already running")
VAR ErrExchangeNotFound = errors.New("exchange not found")
VAR ErrInvalidSymbol = errors.New("invalid symbol")
VAR ErrInvalidPrice = errors.New("invalid price")
VAR ErrInvalidQuantity = errors.New("invalid quantity")
VAR ErrOrderNotFound = errors.New("order not found")
VAR ErrCircuitOpen = errors.New("circuit breaker open")
VAR ErrTimeout = errors.New("request timeout")
```

**Estimated Time**: 3 hours

---

#### Task 1.13: Internal Message Types

**Objective**: Create internal actor messages.

**Files**:

- `internal/message/market.go`
- `internal/message/order.go`
- `internal/message/position.go`
- `internal/message/sync.go`
- `internal/message/health.go`
- `internal/message/internal.go`

**Pseudocode**:

```
// market.go
TYPE SubscribeTicker STRUCT:
    Exchange string
    Symbol string
    ResponseChan chan error

TYPE SubscribeOrderBook STRUCT:
    Exchange string
    Symbol string
    Depth int
    ResponseChan chan error

TYPE WSTicker STRUCT:
    Exchange string
    Ticker *domain.Ticker

TYPE WSOrderBook STRUCT:
    Exchange string
    OrderBook *domain.OrderBook
    IsSnapshot bool

// order.go
TYPE PlaceOrder STRUCT:
    Request *domain.OrderRequest
    ResponseChan chan *OrderResult

TYPE CancelOrder STRUCT:
    Request *domain.CancelRequest
    ResponseChan chan *CancelResult

TYPE OrderUpdate STRUCT:
    Exchange string
    Order *domain.Order

TYPE OrderResult STRUCT:
    Order *domain.Order
    Error error

// position.go
TYPE GetPosition STRUCT:
    Exchange string
    Symbol string
    ResponseChan chan *PositionResult

TYPE PositionUpdate STRUCT:
    Exchange string
    Position *domain.Position

// sync.go
TYPE SyncTime STRUCT:
    Exchange string
    ResponseChan chan time.Time

TYPE GetNonce STRUCT:
    Exchange string
    ResponseChan chan int64

// health.go
TYPE HealthCheckRequest STRUCT:
    Exchange string
    ResponseChan chan *domain.HealthStatus

TYPE CircuitBreakerState STRUCT:
    Exchange string
    State domain.CircuitState

// internal.go
TYPE RegisterExchange STRUCT:
    Name string
    Driver driver.Driver
    Config *config.ExchangeConfig

TYPE ConnectionStatus STRUCT:
    Exchange string
    Status domain.ConnectionStatus
    Error error
```

**Estimated Time**: 4 hours

---

### Phase 2: Driver Interface & Infrastructure (Days 6-8)

---

#### Task 2.1: Driver Interface

**Objective**: Define driver interface for exchange implementations.

**Files**: `pkg/driver/driver.go`

**Pseudocode**:

```
INTERFACE Driver:
    FUNCTION Name() string
    FUNCTION Type() ExchangeType
    FUNCTION Init(config ExchangeConfig) error
    FUNCTION Start() error
    FUNCTION Stop() error
    FUNCTION IsConnected() bool

    // Market Data
    FUNCTION SubscribeTicker(symbol string) error
    FUNCTION SubscribeOrderBook(symbol, depth string) error
    FUNCTION SubscribeTrades(symbol string) error
    FUNCTION Unsubscribe(symbol string) error

    // Trading
    FUNCTION PlaceOrder(req *OrderRequest) (*Order, error)
    FUNCTION CancelOrder(req *CancelRequest) error
    FUNCTION GetOrder(orderID string) (*Order, error)
    FUNCTION GetOpenOrders(symbol string) ([]*Order, error)

    // Account
    FUNCTION GetBalance(asset string) (*Balance, error)
    FUNCTION GetPosition(symbol string) (*Position, error)
    FUNCTION GetSymbols() ([]*SymbolInfo, error)

    // Callbacks
    FUNCTION SetTickerCallback(func(*Ticker))
    FUNCTION SetOrderBookCallback(func(*OrderBook, bool))
    FUNCTION SetTradeCallback(func(*Trade))
    FUNCTION SetOrderCallback(func(*Order))
    FUNCTION SetPositionCallback(func(*Position))
    FUNCTION SetConnectionCallback(func(ConnectionStatus))
```

**Estimated Time**: 3 hours

---

#### Task 2.2: Driver Registry

**Objective**: Create global driver registry.

**Files**: `pkg/driver/registry.go`

**Pseudocode**:

```
VAR registry = map[string]DriverFactory{}

TYPE DriverFactory FUNCTION(config ExchangeConfig) (Driver, error)

FUNCTION Register(name string, factory DriverFactory):
    mutex.Lock()
    registry[name] = factory
    mutex.Unlock()

FUNCTION Get(name string, config ExchangeConfig) (Driver, error):
    mutex.RLock()
    factory = registry[name]
    mutex.RUnlock()
    IF factory == nil THEN
        RETURN error "driver not found"
    RETURN factory(config)

FUNCTION List() []string:
    mutex.RLock()
    names = keys of registry
    mutex.RUnlock()
    RETURN names
```

**Estimated Time**: 2 hours

---

#### Task 2.3: Persistence Backend Interface

**Objective**: Define persistence interfaces.

**Files**: `pkg/persistence/backend.go`, `pkg/persistence/store.go`

**Pseudocode**:

```
// backend.go
INTERFACE PersistenceBackend:
    FUNCTION Init() error
    FUNCTION Close() error
    FUNCTION OrderStore() OrderStore
    FUNCTION PositionStore() PositionStore
    FUNCTION BalanceStore() BalanceStore
    FUNCTION SnapshotStore() SnapshotStore

// store.go
INTERFACE OrderStore:
    FUNCTION Save(order *Order) error
    FUNCTION Get(exchange, orderID string) (*Order, error)
    FUNCTION GetByClientID(exchange, clientID string) (*Order, error)
    FUNCTION List(exchange, symbol string) ([]*Order, error)
    FUNCTION Delete(exchange, orderID string) error

INTERFACE PositionStore:
    FUNCTION Save(position *Position) error
    FUNCTION Get(exchange, symbol string) (*Position, error)
    FUNCTION List(exchange string) ([]*Position, error)
    FUNCTION Delete(exchange, symbol string) error

INTERFACE BalanceStore:
    FUNCTION Save(balance *Balance) error
    FUNCTION Get(exchange, asset string) (*Balance, error)
    FUNCTION List(exchange string) ([]*Balance, error)

INTERFACE SnapshotStore:
    FUNCTION SaveSnapshot(exchange string, data interface{}) error
    FUNCTION LoadSnapshot(exchange string) (interface{}, error)
```

**Estimated Time**: 3 hours

---

#### Task 2.4: Memory Persistence Backend

**Objective**: Implement in-memory persistence.

**Files**: `pkg/persistence/memory/memory.go`

**Pseudocode**:

```
TYPE MemoryBackend STRUCT:
    orders sync.Map // key: exchange:orderID
    ordersByClient sync.Map // key: exchange:clientOrderID
    positions sync.Map // key: exchange:symbol
    balances sync.Map // key: exchange:asset
    snapshots sync.Map // key: exchange
    mutex sync.RWMutex

FUNCTION (m *MemoryBackend) Init() error:
    RETURN nil

FUNCTION (m *MemoryBackend) OrderStore() OrderStore:
    RETURN &memoryOrderStore{backend: m}

TYPE memoryOrderStore STRUCT:
    backend *MemoryBackend

FUNCTION (s *memoryOrderStore) Save(order *Order) error:
    key = order.Exchange + ":" + order.ID
    s.backend.orders.Store(key, order)
    IF order.ClientOrderID != "" THEN
        clientKey = order.Exchange + ":" + order.ClientOrderID
        s.backend.ordersByClient.Store(clientKey, order)
    RETURN nil

FUNCTION (s *memoryOrderStore) Get(exchange, orderID) (*Order, error):
    key = exchange + ":" + orderID
    value, ok = s.backend.orders.Load(key)
    IF NOT ok THEN
        RETURN nil, ErrOrderNotFound
    RETURN value.(*Order), nil

// Similar implementations for Position, Balance, Snapshot stores
```

**Estimated Time**: 4 hours

---

#### Task 2.5: Risk Hook Interface

**Objective**: Create risk hook interface and built-in rules.

**Files**:

- `pkg/risk/hook.go`
- `pkg/risk/rules.go`

**Pseudocode**:

```
// hook.go
INTERFACE RiskHook:
    FUNCTION Name() string
    FUNCTION Validate(order *OrderRequest, metrics *RiskMetrics) (*RiskAlert, error)

TYPE HookChain STRUCT:
    hooks []RiskHook

FUNCTION (c *HookChain) Add(hook RiskHook):
    c.hooks = append(c.hooks, hook)

FUNCTION (c *HookChain) Evaluate(order *OrderRequest, metrics *RiskMetrics) (*RiskAlert, error):
    FOR EACH hook IN c.hooks DO
        alert, err = hook.Validate(order, metrics)
        IF alert != nil OR err != nil THEN
            RETURN alert, err
    RETURN nil, nil

// rules.go
TYPE MaxOrderSizeRule STRUCT:
    MaxSize Decimal
    Action RiskAction

FUNCTION (r *MaxOrderSizeRule) Validate(order, metrics) (*RiskAlert, error):
    IF order.Quantity > r.MaxSize THEN
        CREATE alert with severity based on Action
        IF Action == REJECT THEN
            RETURN alert, error "order exceeds max size"
        RETURN alert, nil
    RETURN nil, nil

TYPE MaxPositionSizeRule STRUCT:
    MaxSize Decimal
    Action RiskAction

TYPE DailyLossLimitRule STRUCT:
    LossLimit Decimal
    Action RiskAction

TYPE MaxDrawdownRule STRUCT:
    MaxDrawdown Decimal
    Action RiskAction
```

**Estimated Time**: 4 hours

---

#### Task 2.6: Risk Manager

**Objective**: Create risk manager actor.

**Files**: `pkg/risk/manager.go`

**Pseudocode**:

```
TYPE RiskManager STRUCT:
    act.Actor
    config RiskConfig
    hooks HookChain
    metrics *RiskMetrics
    persistence PersistenceBackend

FUNCTION (rm *RiskManager) Init(args ...interface{}) error:
    PARSE config from args
    LOAD built-in rules based on config
    FOR EACH rule IN config.Rules DO
        CREATE rule instance
        ADD to hooks chain
    LOAD metrics from persistence
    RETURN nil

FUNCTION (rm *RiskManager) HandleCall(from gen.PID, ref gen.Ref, request interface{}):
    SWITCH request TYPE:
        CASE *OrderRequest:
            CALL rm.validateOrder(request)
        CASE *GetRiskMetrics:
            RETURN rm.metrics
        DEFAULT:
            RETURN ErrUnknownRequest

FUNCTION (rm *RiskManager) validateOrder(order *OrderRequest) (*RiskAlert, error):
    UPDATE metrics with current state
    alert, err = rm.hooks.Evaluate(order, rm.metrics)
    IF alert != nil THEN
        EMIT risk.alert event with alert
    RETURN alert, err

FUNCTION (rm *RiskManager) HandleCast(from gen.PID, message interface{}):
    SWITCH message TYPE:
        CASE *OrderUpdate:
            UPDATE metrics based on order
            PERSIST metrics
        CASE *PositionUpdate:
            UPDATE metrics based on position
            PERSIST metrics
```

**Estimated Time**: 5 hours

---

#### Task 2.7: Circuit Breaker

**Objective**: Implement circuit breaker pattern.

**Files**:

- `pkg/circuit/breaker.go`
- `pkg/circuit/config.go`

**Pseudocode**:

```
// config.go
TYPE Config STRUCT:
    MaxFailures int
    Timeout time.Duration
    HalfOpenRequests int

// breaker.go
TYPE CircuitBreaker STRUCT:
    state CircuitState
    config Config
    failures int
    successes int
    lastFailure time.Time
    mutex sync.RWMutex

FUNCTION NewCircuitBreaker(config Config) *CircuitBreaker:
    RETURN &CircuitBreaker{
        state: CircuitClosed,
        config: config,
    }

FUNCTION (cb *CircuitBreaker) Call(fn func() error) error:
    cb.mutex.Lock()
    state = cb.state
    cb.mutex.Unlock()

    IF state == CircuitOpen THEN
        IF time.Since(cb.lastFailure) > cb.config.Timeout THEN
            CALL cb.setState(CircuitHalfOpen)
        ELSE
            RETURN ErrCircuitOpen

    err = fn()

    IF err != nil THEN
        CALL cb.onFailure()
        RETURN err

    CALL cb.onSuccess()
    RETURN nil

FUNCTION (cb *CircuitBreaker) onFailure():
    cb.mutex.Lock()
    cb.failures++
    cb.lastFailure = time.Now()

    IF cb.state == CircuitHalfOpen THEN
        CALL cb.setState(CircuitOpen)
    ELSE IF cb.failures >= cb.config.MaxFailures THEN
        CALL cb.setState(CircuitOpen)

    cb.mutex.Unlock()

FUNCTION (cb *CircuitBreaker) onSuccess():
    cb.mutex.Lock()
    IF cb.state == CircuitHalfOpen THEN
        cb.successes++
        IF cb.successes >= cb.config.HalfOpenRequests THEN
            CALL cb.setState(CircuitClosed)
            cb.failures = 0
            cb.successes = 0
    ELSE IF cb.state == CircuitClosed THEN
        cb.failures = 0
    cb.mutex.Unlock()

FUNCTION (cb *CircuitBreaker) setState(newState CircuitState):
    cb.state = newState
    EMIT circuit state change event

FUNCTION (cb *CircuitBreaker) State() CircuitState:
    cb.mutex.RLock()
    state = cb.state
    cb.mutex.RUnlock()
    RETURN state
```

**Estimated Time**: 4 hours

---

#### Task 2.8: Logging Adapter

**Objective**: Create zerolog adapter for Ergo.

**Files**: `internal/log/adapter.go`

**Pseudocode**:

```
TYPE ZerologAdapter STRUCT:
    logger zerolog.Logger

FUNCTION NewZerologAdapter(logger zerolog.Logger) *ZerologAdapter:
    RETURN &ZerologAdapter{logger: logger}

FUNCTION (z *ZerologAdapter) Log(level, format, args):
    SWITCH level:
        CASE ergo.LogLevelDebug:
            z.logger.Debug().Msgf(format, args...)
        CASE ergo.LogLevelInfo:
            z.logger.Info().Msgf(format, args...)
        CASE ergo.LogLevelWarning:
            z.logger.Warn().Msgf(format, args...)
        CASE ergo.LogLevelError:
            z.logger.Error().Msgf(format, args...)
        CASE ergo.LogLevelPanic:
            z.logger.Panic().Msgf(format, args...)

FUNCTION (z *ZerologAdapter) LogWithFields(level, message, fields):
    event = SELECT logger level based on level
    FOR EACH key, value IN fields DO
        ADD field to event
    SEND message to event
```

**Estimated Time**: 2 hours

---

### Phase 3: Core Actors & Infrastructure (Days 9-15)

---

#### Task 3.1: Event Topics & Routing

**Objective**: Define event system.

**Files**:

- `internal/event/topics.go`
- `internal/event/handler.go`
- `internal/event/router.go`

**Pseudocode**:

```
// topics.go
CONST (
    TopicTicker = "ticker.%s.%s" // exchange.symbol
    TopicOrderBook = "orderbook.%s.%s"
    TopicTrade = "trade.%s.%s"
    TopicOrder = "order.%s.%s" // exchange.event
    TopicPosition = "position.%s.%s"
    TopicConnection = "connection.%s"
    TopicHealth = "health.%s"
    TopicRisk = "risk.alert"
    TopicMetrics = "metrics.%s"
)

FUNCTION TopicFor(template, args) string:
    RETURN fmt.Sprintf(template, args...)

// handler.go
TYPE HandlerFunc FUNC(event interface{})

TYPE EventHandler STRUCT:
    act.Actor
    handler HandlerFunc
    unsubscribe FUNC()

FUNCTION (h *EventHandler) HandleEvent(message gen.MessageEvent):
    CALL h.handler(message.Message)

// router.go
TYPE EventRouter STRUCT:
    subscriptions sync.Map // topic -> []HandlerFunc

FUNCTION (r *EventRouter) Subscribe(topic string, handler HandlerFunc) FUNC():
    handlers = r.subscriptions.LoadOrStore(topic, []HandlerFunc{})
    handlers = append(handlers, handler)
    r.subscriptions.Store(topic, handlers)

    RETURN FUNC():
        // Remove handler from topic
        r.unsubscribe(topic, handler)

FUNCTION (r *EventRouter) Publish(topic string, event interface{}):
    handlers, ok = r.subscriptions.Load(topic)
    IF ok THEN
        FOR EACH handler IN handlers DO
            GO handler(event)
```

**Estimated Time**: 4 hours

---

#### Task 3.2: Clock Sync Actor

**Objective**: Implement clock synchronization.

**Files**:

- `internal/sync/clock.go`
- `internal/sync/timesource.go`

**Pseudocode**:

```
// timesource.go
INTERFACE TimeSource:
    FUNCTION ServerTime() (time.Time, error)

// clock.go
TYPE ClockSyncActor STRUCT:
    act.Actor
    exchange string
    timeSource TimeSource
    offset time.Duration
    lastSync time.Time
    syncInterval time.Duration

FUNCTION (cs *ClockSyncActor) Init(args):
    PARSE exchange, timeSource, syncInterval
    START periodic sync using gen.Cron
    CALL cs.sync()

FUNCTION (cs *ClockSyncActor) sync():
    localBefore = time.Now()
    serverTime, err = cs.timeSource.ServerTime()
    IF err != nil THEN
        LOG error
        RETURN
    localAfter = time.Now()

    roundTrip = localAfter.Sub(localBefore)
    estimatedLocal = localBefore.Add(roundTrip / 2)
    offset = serverTime.Sub(estimatedLocal)

    cs.offset = offset
    cs.lastSync = time.Now()

    LOG sync result with offset

FUNCTION (cs *ClockSyncActor) HandleCall(from, ref, request):
    SWITCH request TYPE:
        CASE *SyncTime:
            IF time.Since(cs.lastSync) > cs.syncInterval THEN
                CALL cs.sync()
            adjustedTime = time.Now().Add(cs.offset)
            RETURN adjustedTime

FUNCTION (cs *ClockSyncActor) Now() time.Time:
    RETURN time.Now().Add(cs.offset)

FUNCTION (cs *ClockSyncActor) UnixMilli() int64:
    RETURN cs.Now().UnixMilli()
```

**Estimated Time**: 4 hours

---

#### Task 3.3: Nonce Generator

**Objective**: Implement nonce management.

**Files**: `internal/sync/nonce.go`

**Pseudocode**:

```
TYPE NonceGenerator STRUCT:
    lastNonce int64
    mutex sync.Mutex

FUNCTION (ng *NonceGenerator) Next() int64:
    ng.mutex.Lock()
    now = time.Now().UnixMilli()
    IF now <= ng.lastNonce THEN
        ng.lastNonce++
    ELSE
        ng.lastNonce = now
    nonce = ng.lastNonce
    ng.mutex.Unlock()
    RETURN nonce

FUNCTION (ng *NonceGenerator) NextWithClock(clockSync *ClockSyncActor) int64:
    ng.mutex.Lock()
    now = clockSync.UnixMilli()
    IF now <= ng.lastNonce THEN
        ng.lastNonce++
    ELSE
        ng.lastNonce = now
    nonce = ng.lastNonce
    ng.mutex.Unlock()
    RETURN nonce
```

**Estimated Time**: 2 hours

---

#### Task 3.4: Rate Limit Guard Actor

**Objective**: Implement rate limiting.

**Files**:

- `internal/ratelimit/guard.go`
- `internal/ratelimit/bucket.go`
- `internal/ratelimit/weight.go`

**Pseudocode**:

```
// bucket.go
TYPE TokenBucket STRUCT:
    rate rate.Limiter
    burst int

FUNCTION NewTokenBucket(rps, burst int) *TokenBucket:
    RETURN &TokenBucket{
        rate: rate.NewLimiter(rate.Limit(rps), burst),
        burst: burst,
    }

FUNCTION (tb *TokenBucket) Allow() bool:
    RETURN tb.rate.Allow()

FUNCTION (tb *TokenBucket) Wait(ctx context.Context) error:
    RETURN tb.rate.Wait(ctx)

// weight.go (for Binance)
TYPE WeightedLimiter STRUCT:
    limit int
    window time.Duration
    requests []weightedRequest
    mutex sync.Mutex

TYPE weightedRequest STRUCT:
    timestamp time.Time
    weight int

FUNCTION (wl *WeightedLimiter) Allow(weight int) bool:
    wl.mutex.Lock()
    CALL wl.cleanup()
    current = wl.currentWeight()
    IF current + weight > wl.limit THEN
        wl.mutex.Unlock()
        RETURN false
    wl.requests = append(wl.requests, weightedRequest{time.Now(), weight})
    wl.mutex.Unlock()
    RETURN true

FUNCTION (wl *WeightedLimiter) cleanup():
    cutoff = time.Now().Add(-wl.window)
    newRequests = []
    FOR EACH req IN wl.requests DO
        IF req.timestamp.After(cutoff) THEN
            newRequests = append(newRequests, req)
    wl.requests = newRequests

// guard.go
TYPE RateLimitGuardActor STRUCT:
    act.Actor
    exchange string
    limiter interface{} // TokenBucket or WeightedLimiter

FUNCTION (rg *RateLimitGuardActor) Init(args):
    PARSE exchange, config
    IF config.WeightLimit > 0 THEN
        CREATE WeightedLimiter
    ELSE
        CREATE TokenBucket
    RETURN nil

FUNCTION (rg *RateLimitGuardActor) HandleCall(from, ref, request):
    SWITCH request TYPE:
        CASE *AllowRequest:
            weight = request.Weight OR 1
            allowed = rg.limiter.Allow(weight)
            IF NOT allowed THEN
                RETURN error rate limit exceeded
            RETURN nil
        CASE *WaitRequest:
            IF weightedLimiter THEN
                RETURN error "weight limiter doesn't support wait"
            RETURN rg.limiter.Wait(request.Context)
```

**Estimated Time**: 5 hours

---

#### Task 3.5: WebSocket Base Actor

**Objective**: Create base WebSocket actor using gws.

**Files**: `internal/market/ws.go`

**Pseudocode**:

```
TYPE WSActor STRUCT:
    act.Actor
    exchange string
    url string
    conn *gws.Conn
    subscriptions sync.Map // symbol -> bool
    reconnecting atomic.Bool
    config WSConfig
    callbacks Callbacks

TYPE Callbacks STRUCT:
    OnTicker func(*Ticker)
    OnOrderBook func(*OrderBook, bool)
    OnTrade func(*Trade)
    OnOrder func(*Order)
    OnPosition func(*Position)
    OnConnection func(ConnectionStatus)

FUNCTION (ws *WSActor) Init(args):
    PARSE exchange, url, config, callbacks
    CALL ws.connect()
    RETURN nil

FUNCTION (ws *WSActor) connect() error:
    dialer = gws.NewDialer()
    conn, err = dialer.Dial(ws.url)
    IF err != nil THEN
        EMIT connection error event
        SCHEDULE reconnect
        RETURN err

    ws.conn = conn
    GO ws.readLoop()
    EMIT connected event

    // Resubscribe to previous subscriptions
    ws.subscriptions.Range(FUNC(symbol, _):
        CALL ws.subscribe(symbol)
    )

    RETURN nil

FUNCTION (ws *WSActor) readLoop():
    FOR ws.conn != nil DO
        messageType, data, err = ws.conn.ReadMessage()
        IF err != nil THEN
            LOG error
            CALL ws.handleDisconnect()
            BREAK
        CALL ws.handleMessage(data)

FUNCTION (ws *WSActor) handleDisconnect():
    IF ws.reconnecting.CompareAndSwap(false, true) THEN
        EMIT disconnected event
        CLOSE ws.conn
        ws.conn = nil
        CALL ws.reconnect()

FUNCTION (ws *WSActor) reconnect():
    attempt = 0
    FOR attempt < ws.config.MaxReconnectAttempts DO
        delay = exponentialBackoff(attempt)
        time.Sleep(delay)

        err = ws.connect()
        IF err == nil THEN
            ws.reconnecting.Store(false)
            EMIT reconnected event
            RETURN
        attempt++

    ws.reconnecting.Store(false)
    EMIT failed to reconnect event

FUNCTION (ws *WSActor) HandleCall(from, ref, request):
    SWITCH request TYPE:
        CASE *SubscribeTicker:
            RETURN ws.subscribeTicker(request.Symbol)
        CASE *SubscribeOrderBook:
            RETURN ws.subscribeOrderBook(request.Symbol, request.Depth)
        CASE *Unsubscribe:
            RETURN ws.unsubscribe(request.Symbol)

FUNCTION (ws *WSActor) handleMessage(data []byte):
    // To be implemented by exchange-specific drivers
    // Parse message and call appropriate callback

FUNCTION (ws *WSActor) OnPing(conn, payload):
    err = conn.WritePong(payload)
    RETURN err

FUNCTION (ws *WSActor) OnPong(conn, payload):
    // Update last pong time

FUNCTION (ws *WSActor) OnClose(conn, code, reason):
    CALL ws.handleDisconnect()
```

**Estimated Time**: 6 hours

---

#### Task 3.6: Order Book Manager

**Objective**: Manage order book state with snapshots and deltas.

**Files**:

- `internal/market/orderbook.go`
- `internal/market/snapshot.go`
- `internal/market/sequencer.go`

**Pseudocode**:

```
// sequencer.go
TYPE MessageSequencer STRUCT:
    expectedSeq int64
    buffer map[int64]interface{}
    mutex sync.Mutex

FUNCTION (ms *MessageSequencer) Process(seq int64, msg interface{}) ([]interface{}, error):
    ms.mutex.Lock()

    IF seq < ms.expectedSeq THEN
        // Duplicate, ignore
        ms.mutex.Unlock()
        RETURN nil, nil

    IF seq == ms.expectedSeq THEN
        messages = []interface{}{msg}
        ms.expectedSeq++

        // Check buffer for contiguous messages
        FOR ms.buffer[ms.expectedSeq] != nil DO
            messages = append(messages, ms.buffer[ms.expectedSeq])
            DELETE(ms.buffer, ms.expectedSeq)
            ms.expectedSeq++

        ms.mutex.Unlock()
        RETURN messages, nil

    // Future message, buffer it
    ms.buffer[seq] = msg
    ms.mutex.Unlock()
    RETURN nil, nil

// snapshot.go
TYPE SnapshotRecovery STRUCT:
    pendingSnapshot bool
    pendingDeltas []interface{}
    mutex sync.Mutex

FUNCTION (sr *SnapshotRecovery) OnSnapshot(snapshot interface{}):
    sr.mutex.Lock()
    sr.pendingSnapshot = false
    deltas = sr.pendingDeltas
    sr.pendingDeltas = nil
    sr.mutex.Unlock()

    RETURN snapshot, deltas

FUNCTION (sr *SnapshotRecovery) OnDelta(delta interface{}) bool:
    sr.mutex.Lock()
    IF sr.pendingSnapshot THEN
        sr.pendingDeltas = append(sr.pendingDeltas, delta)
        sr.mutex.Unlock()
        RETURN false // buffering
    sr.mutex.Unlock()
    RETURN true // process immediately

// orderbook.go
TYPE OrderBookActor STRUCT:
    act.Actor
    exchange string
    symbol string
    book *domain.OrderBook
    sequencer *MessageSequencer
    snapshot *SnapshotRecovery
    mutex sync.RWMutex

FUNCTION (ob *OrderBookActor) Init(args):
    PARSE exchange, symbol
    ob.book = &domain.OrderBook{
        Exchange: exchange,
        Symbol: symbol,
    }
    ob.sequencer = NewMessageSequencer()
    ob.snapshot = &SnapshotRecovery{pendingSnapshot: true}
    RETURN nil

FUNCTION (ob *OrderBookActor) HandleCast(from, message):
    SWITCH message TYPE:
        CASE *WSOrderBook:
            IF message.IsSnapshot THEN
                CALL ob.handleSnapshot(message)
            ELSE
                CALL ob.handleDelta(message)

FUNCTION (ob *OrderBookActor) handleSnapshot(msg *WSOrderBook):
    snapshot, deltas = ob.snapshot.OnSnapshot(msg.OrderBook)

    ob.mutex.Lock()
    ob.book = snapshot
    ob.mutex.Unlock()

    // Process buffered deltas
    FOR EACH delta IN deltas DO
        CALL ob.applyDelta(delta)

    EMIT orderbook snapshot event

FUNCTION (ob *OrderBookActor) handleDelta(msg *WSOrderBook):
    shouldProcess = ob.snapshot.OnDelta(msg)
    IF NOT shouldProcess THEN
        RETURN // buffering for snapshot

    messages, err = ob.sequencer.Process(msg.OrderBook.SequenceNum, msg)
    IF err != nil THEN
        LOG sequence error
        REQUEST new snapshot
        RETURN

    FOR EACH message IN messages DO
        CALL ob.applyDelta(message.OrderBook)

FUNCTION (ob *OrderBookActor) applyDelta(update *OrderBook):
    ob.mutex.Lock()
    ob.book.UpdateBids(update.Bids)
    ob.book.UpdateAsks(update.Asks)
    ob.book.SequenceNum = update.SequenceNum
    ob.book.Timestamp = update.Timestamp

    IF update.Checksum != "" THEN
        IF NOT ob.validateChecksum(update.Checksum) THEN
            ob.mutex.Unlock()
            REQUEST new snapshot
            RETURN

    ob.mutex.Unlock()
    EMIT orderbook update event

FUNCTION (ob *OrderBookActor) validateChecksum(checksum string) bool:
    calculated = calculateChecksum(ob.book)
    RETURN calculated == checksum

FUNCTION (ob *OrderBookActor) HandleCall(from, ref, request):
    SWITCH request TYPE:
        CASE *GetOrderBook:
            ob.mutex.RLock()
            book = ob.book.Clone()
            ob.mutex.RUnlock()
            RETURN book
```

**Estimated Time**: 8 hours

---

#### Task 3.7: Order Command Actor (CQRS Write Path)

**Objective**: Handle order commands (place, cancel).

**Files**:

- `internal/order/command.go`
- `internal/order/idempotency.go`
- `internal/order/deduplicator.go`

**Pseudocode**:

```
// idempotency.go
TYPE IdempotencyChecker STRUCT:
    requests sync.Map // clientOrderID -> *OrderResult
    ttl time.Duration

FUNCTION (ic *IdempotencyChecker) Check(clientOrderID string) (*OrderResult, bool):
    result, ok = ic.requests.Load(clientOrderID)
    IF ok THEN
        RETURN result.(*OrderResult), true
    RETURN nil, false

FUNCTION (ic *IdempotencyChecker) Store(clientOrderID string, result *OrderResult):
    ic.requests.Store(clientOrderID, result)
    GO FUNC():
        time.Sleep(ic.ttl)
        ic.requests.Delete(clientOrderID)

// deduplicator.go
TYPE RequestDeduplicator STRUCT:
    inflight sync.Map // key -> *singleflight.Group

FUNCTION (rd *RequestDeduplicator) Do(key string, fn func() (interface{}, error)) (interface{}, error):
    group, _ = rd.inflight.LoadOrStore(key, &singleflight.Group{})
    result, err, shared = group.(*singleflight.Group).Do(key, fn)
    IF NOT shared THEN
        GO FUNC():
            time.Sleep(100 * time.Millisecond)
            rd.inflight.Delete(key)
    RETURN result, err

// command.go
TYPE OrderCommandActor STRUCT:
    act.Actor
    exchange string
    driver driver.Driver
    rateLimitPID gen.PID
    riskManagerPID gen.PID
    clockSyncPID gen.PID
    idempotency *IdempotencyChecker
    deduplicator *RequestDeduplicator

FUNCTION (oc *OrderCommandActor) Init(args):
    PARSE exchange, driver, dependencies
    oc.idempotency = NewIdempotencyChecker(5 * time.Minute)
    oc.deduplicator = &RequestDeduplicator{}
    RETURN nil

FUNCTION (oc *OrderCommandActor) HandleCall(from, ref, request):
    SWITCH request TYPE:
        CASE *PlaceOrder:
            RETURN oc.placeOrder(request)
        CASE *CancelOrder:
            RETURN oc.cancelOrder(request)

FUNCTION (oc *OrderCommandActor) placeOrder(req *PlaceOrder) *OrderResult:
    // Check idempotency
    IF req.Request.ClientOrderID != "" THEN
        result, found = oc.idempotency.Check(req.Request.ClientOrderID)
        IF found THEN
            RETURN result

    // Deduplicate concurrent requests
    key = generateKey(req.Request)
    result, err = oc.deduplicator.Do(key, FUNC():
        RETURN oc.executePlaceOrder(req.Request)
    )

    IF err != nil THEN
        RETURN &OrderResult{Error: err}

    // Store for idempotency
    IF req.Request.ClientOrderID != "" THEN
        oc.idempotency.Store(req.Request.ClientOrderID, result.(*OrderResult))

    RETURN result.(*OrderResult)

FUNCTION (oc *OrderCommandActor) executePlaceOrder(req *OrderRequest) (*OrderResult, error):
    // Check rate limit
    err = oc.checkRateLimit(req)
    IF err != nil THEN
        RETURN nil, err

    // Validate with risk manager
    alert, err = oc.validateRisk(req)
    IF err != nil OR (alert != nil AND alert.Action == ActionReject) THEN
        RETURN nil, err OR error "risk rejected"

    // Sync time if needed
    IF req.Type == OrderTypeLimit OR req.Type == OrderTypeStopLimit THEN
        CALL oc.syncTime()

    // Execute order through driver
    order, err = oc.driver.PlaceOrder(req)
    IF err != nil THEN
        EMIT order.rejected event
        RETURN nil, err

    // Publish order event
    EMIT order.created event with order

    RETURN &OrderResult{Order: order}, nil

FUNCTION (oc *OrderCommandActor) checkRateLimit(req *OrderRequest) error:
    ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
    _, err = oc.Call(oc.rateLimitPID, &AllowRequest{Weight: 1})
    cancel()
    RETURN err

FUNCTION (oc *OrderCommandActor) validateRisk(req *OrderRequest) (*RiskAlert, error):
    ctx, cancel = context.WithTimeout(context.Background(), 500*time.Millisecond)
    result, err = oc.Call(oc.riskManagerPID, req)
    cancel()
    RETURN result, err

FUNCTION (oc *OrderCommandActor) cancelOrder(req *CancelOrder) *CancelResult:
    // Check rate limit
    err = oc.checkRateLimit(nil)
    IF err != nil THEN
        RETURN &CancelResult{Error: err}

    // Execute cancel through driver
    err = oc.driver.CancelOrder(req.Request)
    IF err != nil THEN
        RETURN &CancelResult{Error: err}

    // Publish cancel event
    EMIT order.cancel_requested event

    RETURN &CancelResult{Success: true}
```

**Estimated Time**: 8 hours

---

#### Task 3.8: Order State Actor (CQRS Read Path)

**Objective**: Manage order state (read model).

**Files**:

- `internal/order/state.go`
- `internal/order/sequencer.go`

**Pseudocode**:

```
// sequencer.go (order-specific)
TYPE OrderEventSequencer STRUCT:
    orders sync.Map // orderID -> *OrderState
    mutex sync.Mutex

TYPE OrderState STRUCT:
    order *domain.Order
    pendingEvents []OrderEvent
    lastSeq int64
    mutex sync.Mutex

FUNCTION (oes *OrderEventSequencer) Process(orderID string, seq int64, event OrderEvent) error:
    state, _ = oes.orders.LoadOrStore(orderID, &OrderState{})
    orderState = state.(*OrderState)

    orderState.mutex.Lock()

    IF seq <= orderState.lastSeq THEN
        // Duplicate
        orderState.mutex.Unlock()
        RETURN nil

    IF seq == orderState.lastSeq + 1 THEN
        CALL orderState.applyEvent(event)
        orderState.lastSeq = seq

        // Process buffered events
        FOR pending event IN orderState.pendingEvents DO
            IF pending.seq == orderState.lastSeq + 1 THEN
                CALL orderState.applyEvent(pending)
                orderState.lastSeq = pending.seq
                REMOVE pending from pendingEvents

        orderState.mutex.Unlock()
        RETURN nil

    // Future event, buffer it
    orderState.pendingEvents = append(orderState.pendingEvents, event)
    orderState.mutex.Unlock()
    RETURN nil

// state.go
TYPE OrderStateActor STRUCT:
    act.Actor
    exchange string
    sequencer *OrderEventSequencer
    persistence OrderStore

FUNCTION (os *OrderStateActor) Init(args):
    PARSE exchange, persistence
    os.sequencer = &OrderEventSequencer{}
    // Load orders from persistence
    CALL os.loadOrders()
    RETURN nil

FUNCTION (os *OrderStateActor) loadOrders():
    orders, err = os.persistence.List(os.exchange, "")
    IF err != nil THEN
        LOG error
        RETURN
    FOR EACH order IN orders DO
        IF NOT order.IsFinal() THEN
            state = &OrderState{order: order}
            os.sequencer.orders.Store(order.ID, state)

FUNCTION (os *OrderStateActor) HandleCast(from, message):
    SWITCH message TYPE:
        CASE *OrderUpdate:
            CALL os.handleOrderUpdate(message)

FUNCTION (os *OrderStateActor) handleOrderUpdate(update *OrderUpdate):
    // Extract sequence from update (if exchange provides it)
    seq = extractSequence(update)

    // Process through sequencer
    err = os.sequencer.Process(update.Order.ID, seq, update)
    IF err != nil THEN
        LOG error
        RETURN

    // Get current state
    state, ok = os.sequencer.orders.Load(update.Order.ID)
    IF NOT ok THEN
        RETURN

    orderState = state.(*OrderState)
    orderState.mutex.Lock()
    order = orderState.order.Clone()
    orderState.mutex.Unlock()

    // Persist
    err = os.persistence.Save(order)
    IF err != nil THEN
        LOG error

    // Emit event
    EMIT order state change event with order

FUNCTION (os *OrderStateActor) HandleCall(from, ref, request):
    SWITCH request TYPE:
        CASE *GetOrder:
            order, err = os.persistence.Get(os.exchange, request.OrderID)
            RETURN order, err
        CASE *GetOpenOrders:
            orders, err = os.persistence.List(os.exchange, request.Symbol)
            openOrders = FILTER orders WHERE NOT IsFinal()
            RETURN openOrders, err
```

**Estimated Time**: 6 hours

---

#### Task 3.9: Position Tracker Actor

**Objective**: Track positions and balance.

**Files**: `internal/order/tracker.go`

**Pseudocode**:

```
TYPE PositionTrackerActor STRUCT:
    act.Actor
    exchange string
    positions sync.Map // symbol -> *Position
    balances sync.Map // asset -> *Balance
    persistence PersistenceBackend

FUNCTION (pt *PositionTrackerActor) Init(args):
    PARSE exchange, persistence
    // Load from persistence
    CALL pt.loadState()
    RETURN nil

FUNCTION (pt *PositionTrackerActor) loadState():
    positions, _ = pt.persistence.PositionStore().List(pt.exchange)
    FOR EACH position IN positions DO
        pt.positions.Store(position.Symbol, position)

    balances, _ = pt.persistence.BalanceStore().List(pt.exchange)
    FOR EACH balance IN balances DO
        pt.balances.Store(balance.Asset, balance)

FUNCTION (pt *PositionTrackerActor) HandleCast(from, message):
    SWITCH message TYPE:
        CASE *OrderUpdate:
            IF message.Order.Status IN [PARTIALLY_FILLED, FILLED] THEN
                CALL pt.handleFill(message.Order)
        CASE *PositionUpdate:
            CALL pt.handlePositionUpdate(message)

FUNCTION (pt *PositionTrackerActor) handleFill(order *Order):
    symbol = order.Symbol

    // Get or create position
    posInterface, _ = pt.positions.LoadOrStore(symbol, &domain.Position{
        Exchange: pt.exchange,
        Symbol: symbol,
    })
    position = posInterface.(*domain.Position)

    // Calculate fill
    filledQty = order.FilledQty - position.Quantity // incremental fill

    // Update position
    IF order.Side == SideBuy THEN
        position.AddFill(order.AvgFillPrice, filledQty)
        position.Side = PositionSideLong
    ELSE
        IF position.Side == PositionSideLong THEN
            position.ReduceFill(order.AvgFillPrice, filledQty)
        ELSE
            position.AddFill(order.AvgFillPrice, filledQty)
            position.Side = PositionSideShort

    // Persist
    err = pt.persistence.PositionStore().Save(position)
    IF err != nil THEN
        LOG error

    // Emit event
    EMIT position update event

FUNCTION (pt *PositionTrackerActor) handlePositionUpdate(update *PositionUpdate):
    // Exchange-provided position update (from WebSocket)
    pt.positions.Store(update.Position.Symbol, update.Position)

    err = pt.persistence.PositionStore().Save(update.Position)
    IF err != nil THEN
        LOG error

    EMIT position update event

FUNCTION (pt *PositionTrackerActor) HandleCall(from, ref, request):
    SWITCH request TYPE:
        CASE *GetPosition:
            position, ok = pt.positions.Load(request.Symbol)
            IF NOT ok THEN
                RETURN nil, ErrPositionNotFound
            RETURN position.(*Position), nil

        CASE *GetBalance:
            balance, ok = pt.balances.Load(request.Asset)
            IF NOT ok THEN
                // Fetch from exchange
                balance, err = pt.fetchBalance(request.Asset)
                RETURN balance, err
            RETURN balance.(*Balance), nil

FUNCTION (pt *PositionTrackerActor) fetchBalance(asset string) (*Balance, error):
    // Call driver to fetch balance
    // Store in map and persistence
    // Return balance
```

**Estimated Time**: 5 hours

---

#### Task 3.10: Reconciler Actor

**Objective**: Reconcile state with exchange.

**Files**: `internal/actor/reconciler.go`

**Pseudocode**:

```
TYPE ReconcilerActor STRUCT:
    act.Actor
    exchanges map[string]ExchangeInfo
    persistence PersistenceBackend
    interval time.Duration

TYPE ExchangeInfo STRUCT:
    name string
    driver driver.Driver
    orderStatePID gen.PID
    positionTrackerPID gen.PID

FUNCTION (ra *ReconcilerActor) Init(args):
    PARSE persistence, interval
    ra.exchanges = make(map[string]ExchangeInfo)

    // Schedule periodic reconciliation
    gen.CronSendAfter(ra.Self(), ra.interval, &ReconcileAll{})

    RETURN nil

FUNCTION (ra *ReconcilerActor) HandleCast(from, message):
    SWITCH message TYPE:
        CASE *RegisterExchange:
            ra.exchanges[message.Name] = ExchangeInfo{
                name: message.Name,
                driver: message.Driver,
                orderStatePID: message.OrderStatePID,
                positionTrackerPID: message.PositionTrackerPID,
            }

        CASE *ReconcileAll:
            FOR EACH exchange IN ra.exchanges DO
                GO ra.reconcileExchange(exchange)

            // Schedule next reconciliation
            gen.CronSendAfter(ra.Self(), ra.interval, &ReconcileAll{})

FUNCTION (ra *ReconcilerActor) reconcileExchange(info ExchangeInfo):
    LOG "reconciling exchange" with info.name

    // Reconcile orders
    CALL ra.reconcileOrders(info)

    // Reconcile positions
    CALL ra.reconcilePositions(info)

FUNCTION (ra *ReconcilerActor) reconcileOrders(info ExchangeInfo):
    // Get local open orders
    localOrders, err = ra.persistence.OrderStore().List(info.name, "")
    IF err != nil THEN
        LOG error
        RETURN

    localOpenOrders = FILTER localOrders WHERE NOT IsFinal()

    // Get exchange open orders
    exchangeOrders, err = info.driver.GetOpenOrders("")
    IF err != nil THEN
        LOG error
        EMIT reconciliation error event
        RETURN

    // Build maps for comparison
    localMap = MAP localOpenOrders BY orderID
    exchangeMap = MAP exchangeOrders BY orderID

    // Find discrepancies
    FOR EACH orderID IN localMap KEYS DO
        IF NOT EXISTS exchangeMap[orderID] THEN
            // Order exists locally but not on exchange
            // Might be filled or canceled
            LOG "order missing on exchange" with orderID

            // Query individual order
            order, err = info.driver.GetOrder(orderID)
            IF err != nil THEN
                CONTINUE

            // Send update to OrderStateActor
            gen.Cast(info.orderStatePID, &OrderUpdate{
                Exchange: info.name,
                Order: order,
            })

    FOR EACH orderID IN exchangeMap KEYS DO
        IF NOT EXISTS localMap[orderID] THEN
            // Order exists on exchange but not locally
            LOG "order missing locally" with orderID

            // Add to local state
            gen.Cast(info.orderStatePID, &OrderUpdate{
                Exchange: info.name,
                Order: exchangeMap[orderID],
            })
        ELSE
            // Order exists in both, check for differences
            localOrder = localMap[orderID]
            exchangeOrder = exchangeMap[orderID]

            IF localOrder.Status != exchangeOrder.Status OR
               localOrder.FilledQty != exchangeOrder.FilledQty THEN
                LOG "order state mismatch" with orderID

                // Update local state
                gen.Cast(info.orderStatePID, &OrderUpdate{
                    Exchange: info.name,
                    Order: exchangeOrder,
                })

    EMIT reconciliation completed event with stats

FUNCTION (ra *ReconcilerActor) reconcilePositions(info ExchangeInfo):
    // Get local positions
    localPositions, err = ra.persistence.PositionStore().List(info.name)
    IF err != nil THEN
        LOG error
        RETURN

    // Get exchange positions (if supported)
    exchangePositions = []
    FOR EACH localPos IN localPositions DO
        exPos, err = info.driver.GetPosition(localPos.Symbol)
        IF err != nil THEN
            CONTINUE
        exchangePositions = append(exchangePositions, exPos)

    // Compare and update
    FOR EACH exPos IN exchangePositions DO
        localPos = FIND localPositions WHERE Symbol == exPos.Symbol

        IF localPos == nil OR localPos.Quantity != exPos.Quantity THEN
            LOG "position mismatch" with exPos.Symbol

            gen.Cast(info.positionTrackerPID, &PositionUpdate{
                Exchange: info.name,
                Position: exPos,
            })
```

**Estimated Time**: 7 hours

### Phase 4: Binance Driver Implementation (Days 16-19)

---

#### Task 4.1: Binance - Core Driver Structure

**Objective**: Create main Binance driver structure and registration.

**Files**:

- `internal/driver/binance/driver.go`
- `internal/driver/binance/config.go`

**Pseudocode**:

```
// config.go
TYPE Config STRUCT:
    APIKey string
    APISecret string
    Testnet bool
    Timeout time.Duration
    MaxReconnectAttempts int
    ReconnectDelay time.Duration
    RecvWindow int64
    RateLimit RateLimitConfig

FUNCTION DefaultConfig() *Config:
    RETURN &Config{
        Timeout: 10 * time.Second,
        MaxReconnectAttempts: 10,
        ReconnectDelay: 5 * time.Second,
        RecvWindow: 5000,
        RateLimit: RateLimitConfig{
            Enabled: true,
            RequestsPerSecond: 10,
            WeightLimit: 1200,
        },
    }

// driver.go
TYPE BinanceDriver STRUCT:
    config *Config
    rest *RESTClient
    ws *WSClient
    callbacks driver.Callbacks
    mutex sync.RWMutex
    started atomic.Bool

FUNCTION NewBinanceDriver(config *Config) (*BinanceDriver, error):
    IF config == nil THEN
        config = DefaultConfig()

    bd := &BinanceDriver{
        config: config,
    }

    RETURN bd, nil

FUNCTION (bd *BinanceDriver) Name() string:
    RETURN "binance"

FUNCTION (bd *BinanceDriver) Type() domain.ExchangeType:
    RETURN domain.ExchangeBinance

FUNCTION (bd *BinanceDriver) Init(exchangeConfig config.ExchangeConfig) error:
    // Update config from exchangeConfig
    bd.config.APIKey = exchangeConfig.APIKey
    bd.config.APISecret = exchangeConfig.APISecret
    bd.config.Testnet = exchangeConfig.Testnet

    // Create REST client
    bd.rest = NewRESTClient(bd.config)

    // Create WebSocket client
    wsURL := bd.getWebSocketURL()
    bd.ws = NewWSClient(wsURL, bd.config, bd.callbacks)

    RETURN nil

FUNCTION (bd *BinanceDriver) Start() error:
    IF bd.started.Load() THEN
        RETURN errors.ErrAlreadyRunning

    // Start WebSocket connection
    err := bd.ws.Connect()
    IF err != nil THEN
        RETURN fmt.Errorf("failed to start websocket: %w", err)

    bd.started.Store(true)
    RETURN nil

FUNCTION (bd *BinanceDriver) Stop() error:
    IF NOT bd.started.Load() THEN
        RETURN errors.ErrNotRunning

    bd.ws.Disconnect()
    bd.started.Store(false)
    RETURN nil

FUNCTION (bd *BinanceDriver) IsConnected() bool:
    RETURN bd.ws.IsConnected()

FUNCTION (bd *BinanceDriver) getWebSocketURL() string:
    IF bd.config.Testnet THEN
        RETURN "wss://testnet.binance.vision/ws"
    RETURN "wss://stream.binance.com:9443/ws"

FUNCTION (bd *BinanceDriver) getRESTBaseURL() string:
    IF bd.config.Testnet THEN
        RETURN "https://testnet.binance.vision"
    RETURN "https://api.binance.com"

// Callback setters
FUNCTION (bd *BinanceDriver) SetTickerCallback(cb func(*domain.Ticker)):
    bd.mutex.Lock()
    bd.callbacks.OnTicker = cb
    bd.mutex.Unlock()

FUNCTION (bd *BinanceDriver) SetOrderBookCallback(cb func(*domain.OrderBook, bool)):
    bd.mutex.Lock()
    bd.callbacks.OnOrderBook = cb
    bd.mutex.Unlock()

FUNCTION (bd *BinanceDriver) SetTradeCallback(cb func(*domain.Trade)):
    bd.mutex.Lock()
    bd.callbacks.OnTrade = cb
    bd.mutex.Unlock()

FUNCTION (bd *BinanceDriver) SetOrderCallback(cb func(*domain.Order)):
    bd.mutex.Lock()
    bd.callbacks.OnOrder = cb
    bd.mutex.Unlock()

FUNCTION (bd *BinanceDriver) SetPositionCallback(cb func(*domain.Position)):
    bd.mutex.Lock()
    bd.callbacks.OnPosition = cb
    bd.mutex.Unlock()

FUNCTION (bd *BinanceDriver) SetConnectionCallback(cb func(domain.ConnectionStatus)):
    bd.mutex.Lock()
    bd.callbacks.OnConnection = cb
    bd.mutex.Unlock()

// Register driver in init function
FUNCTION init():
    driver.Register("binance", FUNC(cfg config.ExchangeConfig) (driver.Driver, error):
        config := &Config{
            APIKey: cfg.APIKey,
            APISecret: cfg.APISecret,
            Testnet: cfg.Testnet,
        }
        RETURN NewBinanceDriver(config)
    )
```

**Estimated Time**: 4 hours

---

#### Task 4.2: Binance - API Types

**Objective**: Define Binance-specific API types.

**Files**:

- `internal/driver/binance/types.go`
- `internal/driver/binance/types_request.go`
- `internal/driver/binance/types_response.go`
- `internal/driver/binance/types_ws.go`

**Pseudocode**:

```
// types.go
TYPE BinanceError STRUCT:
    Code int `json:"code"`
    Msg string `json:"msg"`

// types_request.go
TYPE OrderRequest STRUCT:
    Symbol string `json:"symbol"`
    Side string `json:"side"`
    Type string `json:"type"`
    TimeInForce string `json:"timeInForce,omitempty"`
    Quantity string `json:"quantity"`
    QuoteOrderQty string `json:"quoteOrderQty,omitempty"`
    Price string `json:"price,omitempty"`
    NewClientOrderID string `json:"newClientOrderId,omitempty"`
    StopPrice string `json:"stopPrice,omitempty"`
    IcebergQty string `json:"icebergQty,omitempty"`
    Timestamp int64 `json:"timestamp"`
    RecvWindow int64 `json:"recvWindow,omitempty"`
    Signature string `json:"signature"`

TYPE CancelOrderRequest STRUCT:
    Symbol string `json:"symbol"`
    OrderID int64 `json:"orderId,omitempty"`
    OrigClientOrderID string `json:"origClientOrderId,omitempty"`
    NewClientOrderID string `json:"newClientOrderId,omitempty"`
    Timestamp int64 `json:"timestamp"`
    RecvWindow int64 `json:"recvWindow,omitempty"`
    Signature string `json:"signature"`

TYPE QueryOrderRequest STRUCT:
    Symbol string `json:"symbol"`
    OrderID int64 `json:"orderId,omitempty"`
    OrigClientOrderID string `json:"origClientOrderId,omitempty"`
    Timestamp int64 `json:"timestamp"`
    RecvWindow int64 `json:"recvWindow,omitempty"`
    Signature string `json:"signature"`

TYPE OpenOrdersRequest STRUCT:
    Symbol string `json:"symbol,omitempty"`
    Timestamp int64 `json:"timestamp"`
    RecvWindow int64 `json:"recvWindow,omitempty"`
    Signature string `json:"signature"`

TYPE AccountRequest STRUCT:
    Timestamp int64 `json:"timestamp"`
    RecvWindow int64 `json:"recvWindow,omitempty"`
    Signature string `json:"signature"`

// types_response.go
TYPE OrderResponse STRUCT:
    Symbol string `json:"symbol"`
    OrderID int64 `json:"orderId"`
    OrderListID int64 `json:"orderListId"`
    ClientOrderID string `json:"clientOrderId"`
    TransactTime int64 `json:"transactTime"`
    Price string `json:"price"`
    OrigQty string `json:"origQty"`
    ExecutedQty string `json:"executedQty"`
    CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
    Status string `json:"status"`
    TimeInForce string `json:"timeInForce"`
    Type string `json:"type"`
    Side string `json:"side"`
    WorkingTime int64 `json:"workingTime"`
    SelfTradePreventionMode string `json:"selfTradePreventionMode"`

TYPE Fill STRUCT:
    Price string `json:"price"`
    Qty string `json:"qty"`
    Commission string `json:"commission"`
    CommissionAsset string `json:"commissionAsset"`
    TradeID int64 `json:"tradeId"`

TYPE OrderResponseFull STRUCT:
    OrderResponse
    Fills []Fill `json:"fills"`

TYPE AccountInfo STRUCT:
    MakerCommission int `json:"makerCommission"`
    TakerCommission int `json:"takerCommission"`
    BuyerCommission int `json:"buyerCommission"`
    SellerCommission int `json:"sellerCommission"`
    CanTrade bool `json:"canTrade"`
    CanWithdraw bool `json:"canWithdraw"`
    CanDeposit bool `json:"canDeposit"`
    UpdateTime int64 `json:"updateTime"`
    AccountType string `json:"accountType"`
    Balances []BalanceInfo `json:"balances"`
    Permissions []string `json:"permissions"`

TYPE BalanceInfo STRUCT:
    Asset string `json:"asset"`
    Free string `json:"free"`
    Locked string `json:"locked"`

TYPE ExchangeInfo STRUCT:
    Timezone string `json:"timezone"`
    ServerTime int64 `json:"serverTime"`
    RateLimits []RateLimit `json:"rateLimits"`
    ExchangeFilters []interface{} `json:"exchangeFilters"`
    Symbols []SymbolInfo `json:"symbols"`

TYPE SymbolInfo STRUCT:
    Symbol string `json:"symbol"`
    Status string `json:"status"`
    BaseAsset string `json:"baseAsset"`
    BaseAssetPrecision int `json:"baseAssetPrecision"`
    QuoteAsset string `json:"quoteAsset"`
    QuotePrecision int `json:"quotePrecision"`
    BaseCommissionPrecision int `json:"baseCommissionPrecision"`
    QuoteCommissionPrecision int `json:"quoteCommissionPrecision"`
    OrderTypes []string `json:"orderTypes"`
    IcebergAllowed bool `json:"icebergAllowed"`
    OcoAllowed bool `json:"ocoAllowed"`
    QuoteOrderQtyMarketAllowed bool `json:"quoteOrderQtyMarketAllowed"`
    AllowTrailingStop bool `json:"allowTrailingStop"`
    IsSpotTradingAllowed bool `json:"isSpotTradingAllowed"`
    IsMarginTradingAllowed bool `json:"isMarginTradingAllowed"`
    Filters []Filter `json:"filters"`
    Permissions []string `json:"permissions"`

TYPE Filter STRUCT:
    FilterType string `json:"filterType"`
    MinPrice string `json:"minPrice,omitempty"`
    MaxPrice string `json:"maxPrice,omitempty"`
    TickSize string `json:"tickSize,omitempty"`
    MinQty string `json:"minQty,omitempty"`
    MaxQty string `json:"maxQty,omitempty"`
    StepSize string `json:"stepSize,omitempty"`
    MinNotional string `json:"minNotional,omitempty"`
    ApplyToMarket bool `json:"applyToMarket,omitempty"`
    AvgPriceMins int `json:"avgPriceMins,omitempty"`

TYPE RateLimit STRUCT:
    RateLimitType string `json:"rateLimitType"`
    Interval string `json:"interval"`
    IntervalNum int `json:"intervalNum"`
    Limit int `json:"limit"`

TYPE ServerTimeResponse STRUCT:
    ServerTime int64 `json:"serverTime"`

// types_ws.go
TYPE WSMessage STRUCT:
    Stream string `json:"stream"`
    Data json.RawMessage `json:"data"`

TYPE WSSubscribeMessage STRUCT:
    Method string `json:"method"`
    Params []string `json:"params"`
    ID int `json:"id"`

TYPE WSUnsubscribeMessage STRUCT:
    Method string `json:"method"`
    Params []string `json:"params"`
    ID int `json:"id"`

TYPE WSTicker STRUCT:
    EventType string `json:"e"`
    EventTime int64 `json:"E"`
    Symbol string `json:"s"`
    PriceChange string `json:"p"`
    PriceChangePercent string `json:"P"`
    WeightedAvgPrice string `json:"w"`
    PrevClosePrice string `json:"x"`
    LastPrice string `json:"c"`
    LastQty string `json:"Q"`
    BidPrice string `json:"b"`
    BidQty string `json:"B"`
    AskPrice string `json:"a"`
    AskQty string `json:"A"`
    OpenPrice string `json:"o"`
    HighPrice string `json:"h"`
    LowPrice string `json:"l"`
    Volume string `json:"v"`
    QuoteVolume string `json:"q"`
    OpenTime int64 `json:"O"`
    CloseTime int64 `json:"C"`
    FirstID int64 `json:"F"`
    LastID int64 `json:"L"`
    Count int64 `json:"n"`

TYPE WSDepthUpdate STRUCT:
    EventType string `json:"e"`
    EventTime int64 `json:"E"`
    Symbol string `json:"s"`
    FirstUpdateID int64 `json:"U"`
    FinalUpdateID int64 `json:"u"`
    Bids [][]string `json:"b"`
    Asks [][]string `json:"a"`

TYPE WSAggTrade STRUCT:
    EventType string `json:"e"`
    EventTime int64 `json:"E"`
    Symbol string `json:"s"`
    AggTradeID int64 `json:"a"`
    Price string `json:"p"`
    Quantity string `json:"q"`
    FirstTradeID int64 `json:"f"`
    LastTradeID int64 `json:"l"`
    TradeTime int64 `json:"T"`
    IsBuyerMaker bool `json:"m"`

TYPE WSExecutionReport STRUCT:
    EventType string `json:"e"`
    EventTime int64 `json:"E"`
    Symbol string `json:"s"`
    ClientOrderID string `json:"c"`
    Side string `json:"S"`
    OrderType string `json:"o"`
    TimeInForce string `json:"f"`
    Quantity string `json:"q"`
    Price string `json:"p"`
    StopPrice string `json:"P"`
    IcebergQty string `json:"F"`
    OrderListID int64 `json:"g"`
    OrigClientOrderID string `json:"C"`
    ExecutionType string `json:"x"`
    OrderStatus string `json:"X"`
    RejectReason string `json:"r"`
    OrderID int64 `json:"i"`
    LastExecutedQty string `json:"l"`
    CumulativeFilledQty string `json:"z"`
    LastExecutedPrice string `json:"L"`
    Commission string `json:"n"`
    CommissionAsset string `json:"N"`
    TransactionTime int64 `json:"T"`
    TradeID int64 `json:"t"`
    IsOrderOnBook bool `json:"w"`
    IsMaker bool `json:"m"`
    OrderCreationTime int64 `json:"O"`
    CumulativeQuoteQty string `json:"Z"`
    LastQuoteQty string `json:"Y"`
    QuoteOrderQty string `json:"Q"`
    WorkingTime int64 `json:"W"`
    SelfTradePreventionMode string `json:"V"`
```

**Estimated Time**: 5 hours

---

#### Task 4.3: Binance - Request Signer

**Objective**: Implement HMAC-SHA256 signing for Binance API.

**Files**: `internal/driver/binance/signer.go`

**Pseudocode**:

```
TYPE Signer STRUCT:
    apiKey string
    apiSecret string

FUNCTION NewSigner(apiKey, apiSecret string) *Signer:
    RETURN &Signer{
        apiKey: apiKey,
        apiSecret: apiSecret,
    }

FUNCTION (s *Signer) Sign(params url.Values) string:
    // Binance uses query string for signing
    queryString := params.Encode()

    // Create HMAC-SHA256 hash
    mac := hmac.New(sha256.New, []byte(s.apiSecret))
    mac.Write([]byte(queryString))

    // Return hex encoded signature
    signature := hex.EncodeToString(mac.Sum(nil))
    RETURN signature

FUNCTION (s *Signer) SignRequest(method, endpoint string, params url.Values) (string, http.Header):
    // Add timestamp if not present
    IF NOT params.Has("timestamp") THEN
        params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

    // Generate signature
    signature := s.Sign(params)
    params.Set("signature", signature)

    // Build final URL
    fullURL := endpoint
    IF len(params) > 0 THEN
        fullURL += "?" + params.Encode()

    // Create headers
    headers := http.Header{}
    headers.Set("X-MBX-APIKEY", s.apiKey)
    headers.Set("Content-Type", "application/x-www-form-urlencoded")

    RETURN fullURL, headers

FUNCTION (s *Signer) SignWebSocket(params map[string]interface{}) string:
    // Build query string from params
    values := url.Values{}
    FOR key, value IN params DO
        values.Set(key, fmt.Sprint(value))

    RETURN s.Sign(values)
```

**Estimated Time**: 3 hours

---

#### Task 4.4: Binance - REST Client

**Objective**: Implement REST client with all endpoints.

**Files**:

- `internal/driver/binance/rest.go`
- `internal/driver/binance/rest_market.go`
- `internal/driver/binance/rest_trading.go`
- `internal/driver/binance/rest_account.go`

**Pseudocode**:

```
// rest.go
TYPE RESTClient STRUCT:
    config *Config
    signer *Signer
    client *resty.Client
    baseURL string
    rateLimiter *ratelimit.WeightedLimiter

FUNCTION NewRESTClient(config *Config) *RESTClient:
    signer := NewSigner(config.APIKey, config.APISecret)

    client := resty.New()
    client.SetTimeout(config.Timeout)
    client.SetRetryCount(3)
    client.SetRetryWaitTime(1 * time.Second)
    client.SetRetryMaxWaitTime(5 * time.Second)

    baseURL := getRESTBaseURL(config.Testnet)

    // Initialize rate limiter (1200 weight per minute)
    rateLimiter := ratelimit.NewWeightedLimiter(1200, 1*time.Minute)

    RETURN &RESTClient{
        config: config,
        signer: signer,
        client: client,
        baseURL: baseURL,
        rateLimiter: rateLimiter,
    }

FUNCTION (rc *RESTClient) do(method, endpoint string, params url.Values, weight int, result interface{}) error:
    // Check rate limit
    IF NOT rc.rateLimiter.Allow(weight) THEN
        RETURN errors.NewRateLimitError("binance", 0, rc.rateLimiter.WaitTime())

    // Build request
    url := rc.baseURL + endpoint

    var resp *resty.Response
    var err error

    SWITCH method:
        CASE "GET":
            IF len(params) > 0 THEN
                url += "?" + params.Encode()
            resp, err = rc.client.R().
                SetResult(result).
                SetError(&BinanceError{}).
                Get(url)

        CASE "POST":
            resp, err = rc.client.R().
                SetFormDataFromValues(params).
                SetResult(result).
                SetError(&BinanceError{}).
                Post(url)

        CASE "DELETE":
            IF len(params) > 0 THEN
                url += "?" + params.Encode()
            resp, err = rc.client.R().
                SetResult(result).
                SetError(&BinanceError{}).
                Delete(url)

    IF err != nil THEN
        RETURN fmt.Errorf("request failed: %w", err)

    IF resp.IsError() THEN
        binanceErr := resp.Error().(*BinanceError)
        RETURN errors.NewExchangeError("binance", binanceErr.Code, binanceErr.Msg)

    RETURN nil

FUNCTION (rc *RESTClient) doSigned(method, endpoint string, params url.Values, weight int, result interface{}) error:
    // Add timestamp
    IF NOT params.Has("timestamp") THEN
        params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

    IF rc.config.RecvWindow > 0 THEN
        params.Set("recvWindow", strconv.FormatInt(rc.config.RecvWindow, 10))

    // Sign request
    signature := rc.signer.Sign(params)
    params.Set("signature", signature)

    // Set API key header
    rc.client.SetHeader("X-MBX-APIKEY", rc.config.APIKey)

    // Execute request
    RETURN rc.do(method, endpoint, params, weight, result)

// rest_market.go
FUNCTION (rc *RESTClient) GetServerTime() (int64, error):
    var result ServerTimeResponse
    err := rc.do("GET", "/api/v3/time", nil, 1, &result)
    IF err != nil THEN
        RETURN 0, err
    RETURN result.ServerTime, nil

FUNCTION (rc *RESTClient) GetExchangeInfo() (*ExchangeInfo, error):
    var result ExchangeInfo
    err := rc.do("GET", "/api/v3/exchangeInfo", nil, 10, &result)
    IF err != nil THEN
        RETURN nil, err
    RETURN &result, nil

FUNCTION (rc *RESTClient) GetOrderBook(symbol string, limit int) (*domain.OrderBook, error):
    params := url.Values{}
    params.Set("symbol", symbol)
    IF limit > 0 THEN
        params.Set("limit", strconv.Itoa(limit))

    var result struct {
        LastUpdateID int64 `json:"lastUpdateId"`
        Bids [][]string `json:"bids"`
        Asks [][]string `json:"asks"`
    }

    err := rc.do("GET", "/api/v3/depth", params, 1, &result)
    IF err != nil THEN
        RETURN nil, err

    // Map to domain.OrderBook
    orderBook := &domain.OrderBook{
        Exchange: "binance",
        Symbol: symbol,
        SequenceNum: result.LastUpdateID,
        Timestamp: time.Now(),
    }

    FOR EACH bid IN result.Bids DO
        price, _ := domain.NewDecimal(bid[0])
        qty, _ := domain.NewDecimal(bid[1])
        orderBook.Bids = append(orderBook.Bids, domain.OrderBookLevel{
            Price: price,
            Quantity: qty,
        })

    FOR EACH ask IN result.Asks DO
        price, _ := domain.NewDecimal(ask[0])
        qty, _ := domain.NewDecimal(ask[1])
        orderBook.Asks = append(orderBook.Asks, domain.OrderBookLevel{
            Price: price,
            Quantity: qty,
        })

    RETURN orderBook, nil

// rest_trading.go
FUNCTION (rc *RESTClient) PlaceOrder(req *domain.OrderRequest) (*domain.Order, error):
    params := url.Values{}
    params.Set("symbol", normalizeSymbol(req.Symbol))
    params.Set("side", mapSide(req.Side))
    params.Set("type", mapOrderType(req.Type))

    IF req.Type != domain.OrderTypeMarket THEN
        params.Set("price", domain.String(req.Price))
        params.Set("timeInForce", mapTimeInForce(req.TimeInForce))

    params.Set("quantity", domain.String(req.Quantity))

    IF req.ClientOrderID != "" THEN
        params.Set("newClientOrderId", req.ClientOrderID)

    IF req.StopPrice != nil THEN
        params.Set("stopPrice", domain.String(req.StopPrice))

    var result OrderResponseFull
    err := rc.doSigned("POST", "/api/v3/order", params, 1, &result)
    IF err != nil THEN
        RETURN nil, err

    // Map to domain.Order
    order := mapOrderResponse(&result)
    RETURN order, nil

FUNCTION (rc *RESTClient) CancelOrder(req *domain.CancelRequest) error:
    params := url.Values{}
    params.Set("symbol", normalizeSymbol(req.Symbol))

    IF req.OrderID != "" THEN
        orderID, err := strconv.ParseInt(req.OrderID, 10, 64)
        IF err != nil THEN
            RETURN err
        params.Set("orderId", strconv.FormatInt(orderID, 10))
    ELSE IF req.ClientOrderID != "" THEN
        params.Set("origClientOrderId", req.ClientOrderID)
    ELSE
        RETURN errors.New("either OrderID or ClientOrderID must be provided")

    var result OrderResponse
    err := rc.doSigned("DELETE", "/api/v3/order", params, 1, &result)
    RETURN err

FUNCTION (rc *RESTClient) GetOrder(symbol, orderID string) (*domain.Order, error):
    params := url.Values{}
    params.Set("symbol", normalizeSymbol(symbol))

    id, err := strconv.ParseInt(orderID, 10, 64)
    IF err != nil THEN
        RETURN nil, err
    params.Set("orderId", strconv.FormatInt(id, 10))

    var result OrderResponse
    err = rc.doSigned("GET", "/api/v3/order", params, 2, &result)
    IF err != nil THEN
        RETURN nil, err

    order := mapOrderResponse(&result)
    RETURN order, nil

FUNCTION (rc *RESTClient) GetOpenOrders(symbol string) ([]*domain.Order, error):
    params := url.Values{}
    IF symbol != "" THEN
        params.Set("symbol", normalizeSymbol(symbol))

    var result []OrderResponse
    weight := 3
    IF symbol != "" THEN
        weight = 3
    ELSE
        weight = 40

    err := rc.doSigned("GET", "/api/v3/openOrders", params, weight, &result)
    IF err != nil THEN
        RETURN nil, err

    orders := make([]*domain.Order, len(result))
    FOR i, orderResp IN result DO
        orders[i] = mapOrderResponse(&orderResp)

    RETURN orders, nil

// rest_account.go
FUNCTION (rc *RESTClient) GetAccount() (*AccountInfo, error):
    params := url.Values{}

    var result AccountInfo
    err := rc.doSigned("GET", "/api/v3/account", params, 10, &result)
    IF err != nil THEN
        RETURN nil, err

    RETURN &result, nil

FUNCTION (rc *RESTClient) GetBalance(asset string) (*domain.Balance, error):
    account, err := rc.GetAccount()
    IF err != nil THEN
        RETURN nil, err

    FOR EACH balance IN account.Balances DO
        IF balance.Asset == asset THEN
            free, _ := domain.NewDecimal(balance.Free)
            locked, _ := domain.NewDecimal(balance.Locked)

            RETURN &domain.Balance{
                Exchange: "binance",
                Asset: asset,
                Free: free,
                Locked: locked,
                Timestamp: time.Now(),
            }, nil

    RETURN nil, errors.ErrBalanceNotFound
```

**Estimated Time**: 10 hours

---

#### Task 4.5: Binance - Domain Mapper

**Objective**: Map Binance types to domain types.

**Files**:

- `internal/driver/binance/mapper.go`
- `internal/driver/binance/mapper_market.go`
- `internal/driver/binance/mapper_order.go`
- `internal/driver/binance/normalizer.go`

**Pseudocode**:

```
// normalizer.go
FUNCTION normalizeSymbol(symbol string) string:
    // Convert domain symbol to Binance format
    // BTC/USDT -> BTCUSDT
    RETURN strings.ReplaceAll(symbol, "/", "")

FUNCTION denormalizeSymbol(symbol string) string:
    // Convert Binance symbol to domain format
    // BTCUSDT -> BTC/USDT
    // This requires exchange info to know where to split
    // For now, assume common pairs
    IF strings.HasSuffix(symbol, "USDT") THEN
        base := symbol[:len(symbol)-4]
        RETURN base + "/USDT"
    ELSE IF strings.HasSuffix(symbol, "BTC") THEN
        base := symbol[:len(symbol)-3]
        RETURN base + "/BTC"
    // Add more pairs as needed
    RETURN symbol

FUNCTION mapSide(side domain.Side) string:
    SWITCH side:
        CASE domain.SideBuy:
            RETURN "BUY"
        CASE domain.SideSell:
            RETURN "SELL"
        DEFAULT:
            RETURN ""

FUNCTION unmapSide(side string) domain.Side:
    SWITCH side:
        CASE "BUY":
            RETURN domain.SideBuy
        CASE "SELL":
            RETURN domain.SideSell
        DEFAULT:
            RETURN domain.SideBuy

FUNCTION mapOrderType(orderType domain.OrderType) string:
    SWITCH orderType:
        CASE domain.OrderTypeMarket:
            RETURN "MARKET"
        CASE domain.OrderTypeLimit:
            RETURN "LIMIT"
        CASE domain.OrderTypeStopLimit:
            RETURN "STOP_LOSS_LIMIT"
        CASE domain.OrderTypeStopMarket:
            RETURN "STOP_LOSS"
        DEFAULT:
            RETURN "LIMIT"

FUNCTION unmapOrderType(orderType string) domain.OrderType:
    SWITCH orderType:
        CASE "MARKET":
            RETURN domain.OrderTypeMarket
        CASE "LIMIT":
            RETURN domain.OrderTypeLimit
        CASE "STOP_LOSS_LIMIT":
            RETURN domain.OrderTypeStopLimit
        CASE "STOP_LOSS":
            RETURN domain.OrderTypeStopMarket
        DEFAULT:
            RETURN domain.OrderTypeLimit

FUNCTION mapTimeInForce(tif domain.TimeInForce) string:
    SWITCH tif:
        CASE domain.TimeInForceGTC:
            RETURN "GTC"
        CASE domain.TimeInForceIOC:
            RETURN "IOC"
        CASE domain.TimeInForceFOK:
            RETURN "FOK"
        DEFAULT:
            RETURN "GTC"

FUNCTION unmapTimeInForce(tif string) domain.TimeInForce:
    SWITCH tif:
        CASE "GTC":
            RETURN domain.TimeInForceGTC
        CASE "IOC":
            RETURN domain.TimeInForceIOC
        CASE "FOK":
            RETURN domain.TimeInForceFOK
        DEFAULT:
            RETURN domain.TimeInForceGTC

FUNCTION mapOrderStatus(status string) domain.OrderStatus:
    SWITCH status:
        CASE "NEW":
            RETURN domain.OrderStatusNew
        CASE "PARTIALLY_FILLED":
            RETURN domain.OrderStatusPartiallyFilled
        CASE "FILLED":
            RETURN domain.OrderStatusFilled
        CASE "CANCELED":
            RETURN domain.OrderStatusCanceled
        CASE "PENDING_CANCEL":
            RETURN domain.OrderStatusCanceling
        CASE "REJECTED":
            RETURN domain.OrderStatusRejected
        CASE "EXPIRED":
            RETURN domain.OrderStatusExpired
        DEFAULT:
            RETURN domain.OrderStatusNew

// mapper_market.go
FUNCTION mapTicker(wsTicker *WSTicker) *domain.Ticker:
    bid, _ := domain.NewDecimal(wsTicker.BidPrice)
    ask, _ := domain.NewDecimal(wsTicker.AskPrice)
    bidQty, _ := domain.NewDecimal(wsTicker.BidQty)
    askQty, _ := domain.NewDecimal(wsTicker.AskQty)
    last, _ := domain.NewDecimal(wsTicker.LastPrice)
    volume, _ := domain.NewDecimal(wsTicker.Volume)
    high, _ := domain.NewDecimal(wsTicker.HighPrice)
    low, _ := domain.NewDecimal(wsTicker.LowPrice)

    RETURN &domain.Ticker{
        Exchange: "binance",
        Symbol: denormalizeSymbol(wsTicker.Symbol),
        Bid: bid,
        Ask: ask,
        BidQty: bidQty,
        AskQty: askQty,
        Last: last,
        Volume24h: volume,
        High24h: high,
        Low24h: low,
        Timestamp: time.UnixMilli(wsTicker.EventTime),
    }

FUNCTION mapOrderBookUpdate(wsDepth *WSDepthUpdate) (*domain.OrderBook, bool):
    orderBook := &domain.OrderBook{
        Exchange: "binance",
        Symbol: denormalizeSymbol(wsDepth.Symbol),
        SequenceNum: wsDepth.FinalUpdateID,
        Timestamp: time.UnixMilli(wsDepth.EventTime),
    }

    FOR EACH bid IN wsDepth.Bids DO
        price, _ := domain.NewDecimal(bid[0])
        qty, _ := domain.NewDecimal(bid[1])

        IF NOT domain.IsZero(qty) THEN
            orderBook.Bids = append(orderBook.Bids, domain.OrderBookLevel{
                Price: price,
                Quantity: qty,
            })

    FOR EACH ask IN wsDepth.Asks DO
        price, _ := domain.NewDecimal(ask[0])
        qty, _ := domain.NewDecimal(ask[1])

        IF NOT domain.IsZero(qty) THEN
            orderBook.Asks = append(orderBook.Asks, domain.OrderBookLevel{
                Price: price,
                Quantity: qty,
            })

    RETURN orderBook, false // false = delta update

FUNCTION mapTrade(wsAggTrade *WSAggTrade) *domain.Trade:
    price, _ := domain.NewDecimal(wsAggTrade.Price)
    qty, _ := domain.NewDecimal(wsAggTrade.Quantity)

    side := domain.SideBuy
    IF wsAggTrade.IsBuyerMaker THEN
        side = domain.SideSell

    RETURN &domain.Trade{
        Exchange: "binance",
        Symbol: denormalizeSymbol(wsAggTrade.Symbol),
        TradeID: strconv.FormatInt(wsAggTrade.AggTradeID, 10),
        Price: price,
        Quantity: qty,
        Side: side,
        Timestamp: time.UnixMilli(wsAggTrade.TradeTime),
        IsBuyerMaker: wsAggTrade.IsBuyerMaker,
    }

// mapper_order.go
FUNCTION mapOrderResponse(orderResp *OrderResponse) *domain.Order:
    price, _ := domain.NewDecimal(orderResp.Price)
    qty, _ := domain.NewDecimal(orderResp.OrigQty)
    filledQty, _ := domain.NewDecimal(orderResp.ExecutedQty)

    // Calculate average fill price
    avgPrice := domain.Zero()
    IF NOT domain.IsZero(filledQty) THEN
        quoteQty, _ := domain.NewDecimal(orderResp.CummulativeQuoteQty)
        avgPrice = domain.Div(quoteQty, filledQty)

    RETURN &domain.Order{
        ID: strconv.FormatInt(orderResp.OrderID, 10),
        ClientOrderID: orderResp.ClientOrderID,
        Exchange: "binance",
        Symbol: denormalizeSymbol(orderResp.Symbol),
        Side: unmapSide(orderResp.Side),
        Type: unmapOrderType(orderResp.Type),
        Status: mapOrderStatus(orderResp.Status),
        Price: price,
        Quantity: qty,
        FilledQty: filledQty,
        AvgFillPrice: avgPrice,
        TimeInForce: unmapTimeInForce(orderResp.TimeInForce),
        CreatedAt: time.UnixMilli(orderResp.WorkingTime),
        UpdatedAt: time.UnixMilli(orderResp.TransactTime),
    }

FUNCTION mapExecutionReport(execReport *WSExecutionReport) *domain.Order:
    price, _ := domain.NewDecimal(execReport.Price)
    qty, _ := domain.NewDecimal(execReport.Quantity)
    filledQty, _ := domain.NewDecimal(execReport.CumulativeFilledQty)

    avgPrice := domain.Zero()
    IF NOT domain.IsZero(filledQty) THEN
        quoteQty, _ := domain.NewDecimal(execReport.CumulativeQuoteQty)
        avgPrice = domain.Div(quoteQty, filledQty)

    RETURN &domain.Order{
        ID: strconv.FormatInt(execReport.OrderID, 10),
        ClientOrderID: execReport.ClientOrderID,
        Exchange: "binance",
        Symbol: denormalizeSymbol(execReport.Symbol),
        Side: unmapSide(execReport.Side),
        Type: unmapOrderType(execReport.OrderType),
        Status: mapOrderStatus(execReport.OrderStatus),
        Price: price,
        Quantity: qty,
        FilledQty: filledQty,
        AvgFillPrice: avgPrice,
        TimeInForce: unmapTimeInForce(execReport.TimeInForce),
        CreatedAt: time.UnixMilli(execReport.OrderCreationTime),
        UpdatedAt: time.UnixMilli(execReport.TransactionTime),
    }

FUNCTION mapSymbolInfo(binanceSymbol *SymbolInfo) *domain.SymbolInfo:
    symbolInfo := &domain.SymbolInfo{
        Exchange: "binance",
        Symbol: denormalizeSymbol(binanceSymbol.Symbol),
        BaseAsset: binanceSymbol.BaseAsset,
        QuoteAsset: binanceSymbol.QuoteAsset,
        Status: binanceSymbol.Status,
        Permissions: binanceSymbol.Permissions,
    }

    // Parse filters
    FOR EACH filter IN binanceSymbol.Filters DO
        SWITCH filter.FilterType:
            CASE "PRICE_FILTER":
                minPrice, _ := domain.NewDecimal(filter.MinPrice)
                maxPrice, _ := domain.NewDecimal(filter.MaxPrice)
                tickSize, _ := domain.NewDecimal(filter.TickSize)
                symbolInfo.PriceFilter = domain.PriceFilter{
                    MinPrice: minPrice,
                    MaxPrice: maxPrice,
                    TickSize: tickSize,
                }

            CASE "LOT_SIZE":
                minQty, _ := domain.NewDecimal(filter.MinQty)
                maxQty, _ := domain.NewDecimal(filter.MaxQty)
                stepSize, _ := domain.NewDecimal(filter.StepSize)
                symbolInfo.LotSizeFilter = domain.LotSizeFilter{
                    MinQty: minQty,
                    MaxQty: maxQty,
                    StepSize: stepSize,
                }

            CASE "MIN_NOTIONAL":
                minNotional, _ := domain.NewDecimal(filter.MinNotional)
                symbolInfo.MinNotional = minNotional

    RETURN symbolInfo
```

**Estimated Time**: 8 hours

---

#### Task 4.6: Binance - WebSocket Client

**Objective**: Implement WebSocket client for market data and user data.

**Files**:

- `internal/driver/binance/ws.go`
- `internal/driver/binance/ws_handlers.go`

**Pseudocode**:

```
// ws.go
TYPE WSClient STRUCT:
    url string
    config *Config
    callbacks driver.Callbacks
    conn *gws.Conn
    subscriptions sync.Map // stream -> bool
    listenKey string
    reconnecting atomic.Bool
    connected atomic.Bool
    mutex sync.RWMutex
    msgID atomic.Int32

FUNCTION NewWSClient(url string, config *Config, callbacks driver.Callbacks) *WSClient:
    RETURN &WSClient{
        url: url,
        config: config,
        callbacks: callbacks,
    }

FUNCTION (ws *WSClient) Connect() error:
    upgrader := gws.NewUpgrader(&ws, &gws.ServerOption{
        ReadBufferSize: 4096,
        WriteBufferSize: 4096,
        CompressEnabled: true,
    })

    conn, _, err := upgrader.Dial(ws.url)
    IF err != nil THEN
        ws.notifyConnection(domain.StatusDisconnected)
        RETURN fmt.Errorf("failed to connect: %w", err)

    ws.conn = conn
    ws.connected.Store(true)
    ws.notifyConnection(domain.StatusConnected)

    // Start read loop
    GO ws.readLoop()

    // Resubscribe to existing streams
    ws.resubscribe()

    RETURN nil

FUNCTION (ws *WSClient) Disconnect():
    IF ws.conn != nil THEN
        ws.conn.WriteClose(1000, []byte("normal closure"))
        ws.conn = nil
        ws.connected.Store(false)
        ws.notifyConnection(domain.StatusDisconnected)

FUNCTION (ws *WSClient) IsConnected() bool:
    RETURN ws.connected.Load()

FUNCTION (ws *WSClient) Subscribe(stream string) error:
    ws.subscriptions.Store(stream, true)

    IF NOT ws.IsConnected() THEN
        RETURN nil // Will subscribe on connect

    RETURN ws.sendSubscribe(stream)

FUNCTION (ws *WSClient) Unsubscribe(stream string) error:
    ws.subscriptions.Delete(stream)

    IF NOT ws.IsConnected() THEN
        RETURN nil

    RETURN ws.sendUnsubscribe(stream)

FUNCTION (ws *WSClient) sendSubscribe(stream string) error:
    msg := WSSubscribeMessage{
        Method: "SUBSCRIBE",
        Params: []string{stream},
        ID: int(ws.msgID.Add(1)),
    }

    data, err := json.Marshal(msg)
    IF err != nil THEN
        RETURN err

    RETURN ws.conn.WriteMessage(gws.OpcodeText, data)

FUNCTION (ws *WSClient) sendUnsubscribe(stream string) error:
    msg := WSUnsubscribeMessage{
        Method: "UNSUBSCRIBE",
        Params: []string{stream},
        ID: int(ws.msgID.Add(1)),
    }

    data, err := json.Marshal(msg)
    IF err != nil THEN
        RETURN err

    RETURN ws.conn.WriteMessage(gws.OpcodeText, data)

FUNCTION (ws *WSClient) resubscribe():
    streams := []string{}
    ws.subscriptions.Range(FUNC(key, value interface{}):
        streams = append(streams, key.(string))
        RETURN true
    )

    FOR EACH stream IN streams DO
        err := ws.sendSubscribe(stream)
        IF err != nil THEN
            LOG error subscribing to stream

FUNCTION (ws *WSClient) readLoop():
    FOR ws.IsConnected() DO
        msgType, data, err := ws.conn.ReadMessage()
        IF err != nil THEN
            LOG error reading message
            ws.handleDisconnect()
            BREAK

        IF msgType == gws.OpcodeText THEN
            ws.handleMessage(data)

FUNCTION (ws *WSClient) handleDisconnect():
    IF ws.reconnecting.CompareAndSwap(false, true) THEN
        ws.connected.Store(false)
        ws.notifyConnection(domain.StatusDisconnected)
        CLOSE ws.conn
        ws.conn = nil

        GO ws.reconnect()

FUNCTION (ws *WSClient) reconnect():
    attempt := 0
    FOR attempt < ws.config.MaxReconnectAttempts DO
        ws.notifyConnection(domain.StatusReconnecting)

        delay := exponentialBackoff(attempt, ws.config.ReconnectDelay)
        time.Sleep(delay)

        err := ws.Connect()
        IF err == nil THEN
            ws.reconnecting.Store(false)
            ws.notifyConnection(domain.StatusConnected)
            LOG "successfully reconnected"
            RETURN

        LOG "reconnection attempt failed" with error
        attempt++

    ws.reconnecting.Store(false)
    ws.notifyConnection(domain.StatusDisconnected)
    LOG "failed to reconnect after max attempts"

FUNCTION (ws *WSClient) notifyConnection(status domain.ConnectionStatus):
    IF ws.callbacks.OnConnection != nil THEN
        ws.callbacks.OnConnection(status)

// gws.EventHandler implementation
FUNCTION (ws *WSClient) OnOpen(conn *gws.Conn):
    LOG "websocket connection opened"

FUNCTION (ws *WSClient) OnClose(conn *gws.Conn, code uint16, reason []byte):
    LOG "websocket connection closed" with code and reason
    ws.handleDisconnect()

FUNCTION (ws *WSClient) OnPing(conn *gws.Conn, payload []byte) error:
    RETURN conn.WritePong(payload)

FUNCTION (ws *WSClient) OnPong(conn *gws.Conn, payload []byte) error:
    RETURN nil

FUNCTION (ws *WSClient) OnError(conn *gws.Conn, err error):
    LOG error with err

// ws_handlers.go
FUNCTION (ws *WSClient) handleMessage(data []byte):
    // Try to parse as stream message
    var streamMsg WSMessage
    err := json.Unmarshal(data, &streamMsg)
    IF err == nil AND streamMsg.Stream != "" THEN
        ws.handleStreamMessage(&streamMsg)
        RETURN

    // Try to parse as direct event
    var eventType struct {
        E string `json:"e"`
    }
    err = json.Unmarshal(data, &eventType)
    IF err != nil THEN
        LOG "failed to parse message"
        RETURN

    ws.handleDirectMessage(eventType.E, data)

FUNCTION (ws *WSClient) handleStreamMessage(msg *WSMessage):
    stream := msg.Stream

    IF strings.Contains(stream, "@ticker") THEN
        ws.handleTickerStream(msg.Data)
    ELSE IF strings.Contains(stream, "@depth") THEN
        ws.handleDepthStream(msg.Data)
    ELSE IF strings.Contains(stream, "@aggTrade") THEN
        ws.handleAggTradeStream(msg.Data)
    ELSE
        LOG "unknown stream type" with stream

FUNCTION (ws *WSClient) handleDirectMessage(eventType string, data []byte):
    SWITCH eventType:
        CASE "24hrTicker":
            ws.handleTickerStream(data)
        CASE "depthUpdate":
            ws.handleDepthStream(data)
        CASE "aggTrade":
            ws.handleAggTradeStream(data)
        CASE "executionReport":
            ws.handleExecutionReport(data)
        DEFAULT:
            LOG "unknown event type" with eventType

FUNCTION (ws *WSClient) handleTickerStream(data []byte):
    var ticker WSTicker
    err := json.Unmarshal(data, &ticker)
    IF err != nil THEN
        LOG error parsing ticker
        RETURN

    IF ws.callbacks.OnTicker != nil THEN
        domainTicker := mapTicker(&ticker)
        ws.callbacks.OnTicker(domainTicker)

FUNCTION (ws *WSClient) handleDepthStream(data []byte):
    var depth WSDepthUpdate
    err := json.Unmarshal(data, &depth)
    IF err != nil THEN
        LOG error parsing depth
        RETURN

    IF ws.callbacks.OnOrderBook != nil THEN
        orderBook, isSnapshot := mapOrderBookUpdate(&depth)
        ws.callbacks.OnOrderBook(orderBook, isSnapshot)

FUNCTION (ws *WSClient) handleAggTradeStream(data []byte):
    var aggTrade WSAggTrade
    err := json.Unmarshal(data, &aggTrade)
    IF err != nil THEN
        LOG error parsing trade
        RETURN

    IF ws.callbacks.OnTrade != nil THEN
        trade := mapTrade(&aggTrade)
        ws.callbacks.OnTrade(trade)

FUNCTION (ws *WSClient) handleExecutionReport(data []byte):
    var execReport WSExecutionReport
    err := json.Unmarshal(data, &execReport)
    IF err != nil THEN
        LOG error parsing execution report
        RETURN

    IF ws.callbacks.OnOrder != nil THEN
        order := mapExecutionReport(&execReport)
        ws.callbacks.OnOrder(order)

FUNCTION exponentialBackoff(attempt int, baseDelay time.Duration) time.Duration:
    delay := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
    maxDelay := 60 * time.Second
    IF delay > maxDelay THEN
        delay = maxDelay
    RETURN delay
```

**Estimated Time**: 10 hours

---

#### Task 4.7: Binance - Driver Integration

**Objective**: Integrate all Binance components into driver interface.

**Files**: `internal/driver/binance/driver.go` (additions)

**Pseudocode**:

```
// Market Data methods
FUNCTION (bd *BinanceDriver) SubscribeTicker(symbol string) error:
    stream := strings.ToLower(normalizeSymbol(symbol)) + "@ticker"
    RETURN bd.ws.Subscribe(stream)

FUNCTION (bd *BinanceDriver) SubscribeOrderBook(symbol string, depth int) error:
    // Binance uses @depth for full order book updates
    // or @depth@100ms for faster updates
    normalized := strings.ToLower(normalizeSymbol(symbol))
    stream := normalized + "@depth@100ms"

    // Also fetch initial snapshot
    GO FUNC():
        orderBook, err := bd.rest.GetOrderBook(symbol, depth)
        IF err == nil AND bd.callbacks.OnOrderBook != nil THEN
            bd.callbacks.OnOrderBook(orderBook, true) // true = snapshot

    RETURN bd.ws.Subscribe(stream)

FUNCTION (bd *BinanceDriver) SubscribeTrades(symbol string) error:
    stream := strings.ToLower(normalizeSymbol(symbol)) + "@aggTrade"
    RETURN bd.ws.Subscribe(stream)

FUNCTION (bd *BinanceDriver) Unsubscribe(symbol string) error:
    normalized := strings.ToLower(normalizeSymbol(symbol))

    // Unsubscribe from all streams for this symbol
    streams := []string{
        normalized + "@ticker",
        normalized + "@depth@100ms",
        normalized + "@aggTrade",
    }

    FOR EACH stream IN streams DO
        bd.ws.Unsubscribe(stream)

    RETURN nil

// Trading methods
FUNCTION (bd *BinanceDriver) PlaceOrder(req *domain.OrderRequest) (*domain.Order, error):
    IF NOT bd.started.Load() THEN
        RETURN nil, errors.ErrNotRunning

    RETURN bd.rest.PlaceOrder(req)

FUNCTION (bd *BinanceDriver) CancelOrder(req *domain.CancelRequest) error:
    IF NOT bd.started.Load() THEN
        RETURN errors.ErrNotRunning

    RETURN bd.rest.CancelOrder(req)

FUNCTION (bd *BinanceDriver) GetOrder(orderID string) (*domain.Order, error):
    IF NOT bd.started.Load() THEN
        RETURN nil, errors.ErrNotRunning

    // Need symbol for GetOrder in Binance
    // This is a limitation - might need to store order-symbol mapping
    RETURN nil, errors.New("GetOrder requires symbol in Binance")

FUNCTION (bd *BinanceDriver) GetOpenOrders(symbol string) ([]*domain.Order, error):
    IF NOT bd.started.Load() THEN
        RETURN nil, errors.ErrNotRunning

    RETURN bd.rest.GetOpenOrders(symbol)

// Account methods
FUNCTION (bd *BinanceDriver) GetBalance(asset string) (*domain.Balance, error):
    IF NOT bd.started.Load() THEN
        RETURN nil, errors.ErrNotRunning

    RETURN bd.rest.GetBalance(asset)

FUNCTION (bd *BinanceDriver) GetPosition(symbol string) (*domain.Position, error):
    // Spot trading doesn't have positions
    RETURN nil, errors.New("positions not supported in spot trading")

FUNCTION (bd *BinanceDriver) GetSymbols() ([]*domain.SymbolInfo, error):
    IF NOT bd.started.Load() THEN
        RETURN nil, errors.ErrNotRunning

    exchangeInfo, err := bd.rest.GetExchangeInfo()
    IF err != nil THEN
        RETURN nil, err

    symbols := make([]*domain.SymbolInfo, len(exchangeInfo.Symbols))
    FOR i, binanceSymbol IN exchangeInfo.Symbols DO
        symbols[i] = mapSymbolInfo(&binanceSymbol)

    RETURN symbols, nil
```

**Estimated Time**: 4 hours

---

#### Task 4.8: Binance - Error Handling

**Objective**: Implement Binance-specific error handling.

**Files**: `internal/driver/binance/errors.go`

**Pseudocode**:

```
// Binance error codes
CONST (
    ErrCodeUnknown = -1000
    ErrCodeDisconnected = -1001
    ErrCodeUnauthorized = -1002
    ErrCodeTooManyRequests = -1003
    ErrCodeDuplicateIP = -1004
    ErrCodeNoSuchIP = -1005
    ErrCodeUnexpectedResp = -1006
    ErrCodeTimeout = -1007
    ErrCodeInvalidMessage = -1013
    ErrCodeUnknownOrderComposition = -1014
    ErrCodeTooManyOrders = -1015
    ErrCodeServiceShuttingDown = -1016
    ErrCodeUnsupportedOperation = -1020
    ErrCodeInvalidTimestamp = -1021
    ErrCodeInvalidSignature = -1022
    ErrCodeIllegalChars = -1100
    ErrCodeTooManyParameters = -1101
    ErrCodeMandatoryParamEmpty = -1102
    ErrCodeUnknownParam = -1103
    ErrCodeUnreadParameters = -1104
    ErrCodeParamEmpty = -1105
    ErrCodeParamNotRequired = -1106
    ErrCodeNoDepth = -1112
    ErrCodeTIFNotRequired = -1114
    ErrCodeInvalidTIF = -1115
    ErrCodeInvalidOrderType = -1116
    ErrCodeInvalidSide = -1117
    ErrCodeEmptyNewClOrdID = -1118
    ErrCodeEmptyOrgClOrdID = -1119
    ErrCodeBadInterval = -1120
    ErrCodeBadSymbol = -1121
    ErrCodeInvalidListenKey = -1125
    ErrCodeMoreThanXXHours = -1127
    ErrCodeOptionalParamsBadCombo = -1128
    ErrCodeInvalidParameter = -1130
    ErrCodeNewOrderRejected = -2010
    ErrCodeCancelRejected = -2011
    ErrCodeNoSuchOrder = -2013
    ErrCodeBadAPIKeyFmt = -2014
    ErrCodeRejectedAPIKey = -2015
    ErrCodeNoTradingWindow = -2016
)

FUNCTION MapBinanceError(code int, message string) error:
    baseErr := errors.NewExchangeError("binance", code, message)

    SWITCH code:
        CASE ErrCodeTooManyRequests:
            RETURN errors.NewRateLimitError("binance", 0, 60)

        CASE ErrCodeInvalidTimestamp:
            RETURN fmt.Errorf("%w: clock sync issue - %s", baseErr, message)

        CASE ErrCodeInvalidSignature:
            RETURN fmt.Errorf("%w: signature verification failed - %s", baseErr, message)

        CASE ErrCodeNoSuchOrder:
            RETURN fmt.Errorf("%w: %s", errors.ErrOrderNotFound, message)

        CASE ErrCodeDisconnected, ErrCodeTimeout:
            RETURN errors.NewConnectionError("binance", message, true)

        CASE ErrCodeUnauthorized, ErrCodeBadAPIKeyFmt, ErrCodeRejectedAPIKey:
            RETURN fmt.Errorf("%w: authentication failed - %s", baseErr, message)

        DEFAULT:
            RETURN baseErr

FUNCTION IsRetryable(err error) bool:
    var binanceErr *BinanceError
    IF errors.As(err, &binanceErr) THEN
        SWITCH binanceErr.Code:
            CASE ErrCodeTimeout,
                 ErrCodeDisconnected,
                 ErrCodeServiceShuttingDown,
                 ErrCodeUnexpectedResp:
                RETURN true

    var connErr *errors.ConnectionError
    IF errors.As(err, &connErr) THEN
        RETURN connErr.Temporary

    RETURN false

FUNCTION IsRateLimitError(err error) bool:
    var rateLimitErr *errors.RateLimitError
    RETURN errors.As(err, &rateLimitErr)
```

**Estimated Time**: 3 hours

---

### Phase 5: Bybit Driver Implementation (Days 20-23)

---

#### Task 5.1: Bybit - Core Driver Structure

**Objective**: Create main Bybit driver structure (V5 API).

**Files**:

- `internal/driver/bybit/driver.go`
- `internal/driver/bybit/config.go`

**Pseudocode**:

```
// config.go
TYPE Config STRUCT:
    APIKey string
    APISecret string
    Testnet bool
    Timeout time.Duration
    MaxReconnectAttempts int
    ReconnectDelay time.Duration
    RecvWindow int64
    RateLimit RateLimitConfig

FUNCTION DefaultConfig() *Config:
    RETURN &Config{
        Timeout: 10 * time.Second,
        MaxReconnectAttempts: 10,
        ReconnectDelay: 5 * time.Second,
        RecvWindow: 5000,
        RateLimit: RateLimitConfig{
            Enabled: true,
            RequestsPerSecond: 10,
            BurstSize: 20,
        },
    }

// driver.go
TYPE BybitDriver STRUCT:
    config *Config
    rest *RESTClient
    wsPublic *WSPublicClient
    wsPrivate *WSPrivateClient
    callbacks driver.Callbacks
    mutex sync.RWMutex
    started atomic.Bool

FUNCTION NewBybitDriver(config *Config) (*BybitDriver, error):
    IF config == nil THEN
        config = DefaultConfig()

    bd := &BybitDriver{
        config: config,
    }

    RETURN bd, nil

FUNCTION (bd *BybitDriver) Name() string:
    RETURN "bybit"

FUNCTION (bd *BybitDriver) Type() domain.ExchangeType:
    RETURN domain.ExchangeBybit

FUNCTION (bd *BybitDriver) Init(exchangeConfig config.ExchangeConfig) error:
    bd.config.APIKey = exchangeConfig.APIKey
    bd.config.APISecret = exchangeConfig.APISecret
    bd.config.Testnet = exchangeConfig.Testnet

    // Create REST client
    bd.rest = NewRESTClient(bd.config)

    // Create WebSocket clients (V5 API has separate public/private endpoints)
    publicURL := bd.getPublicWebSocketURL()
    privateURL := bd.getPrivateWebSocketURL()

    bd.wsPublic = NewWSPublicClient(publicURL, bd.config, bd.callbacks)
    bd.wsPrivate = NewWSPrivateClient(privateURL, bd.config, bd.callbacks)

    RETURN nil

FUNCTION (bd *BybitDriver) Start() error:
    IF bd.started.Load() THEN
        RETURN errors.ErrAlreadyRunning

    // Start public WebSocket
    err := bd.wsPublic.Connect()
    IF err != nil THEN
        RETURN fmt.Errorf("failed to start public websocket: %w", err)

    // Start private WebSocket (if API keys provided)
    IF bd.config.APIKey != "" AND bd.config.APISecret != "" THEN
        err = bd.wsPrivate.Connect()
        IF err != nil THEN
            LOG warning "failed to start private websocket" with error
            // Don't fail startup, private stream is optional

    bd.started.Store(true)
    RETURN nil

FUNCTION (bd *BybitDriver) Stop() error:
    IF NOT bd.started.Load() THEN
        RETURN errors.ErrNotRunning

    bd.wsPublic.Disconnect()
    bd.wsPrivate.Disconnect()
    bd.started.Store(false)
    RETURN nil

FUNCTION (bd *BybitDriver) IsConnected() bool:
    RETURN bd.wsPublic.IsConnected()

FUNCTION (bd *BybitDriver) getPublicWebSocketURL() string:
    IF bd.config.Testnet THEN
        RETURN "wss://stream-testnet.bybit.com/v5/public/spot"
    RETURN "wss://stream.bybit.com/v5/public/spot"

FUNCTION (bd *BybitDriver) getPrivateWebSocketURL() string:
    IF bd.config.Testnet THEN
        RETURN "wss://stream-testnet.bybit.com/v5/private"
    RETURN "wss://stream.bybit.com/v5/private"

FUNCTION (bd *BybitDriver) getRESTBaseURL() string:
    IF bd.config.Testnet THEN
        RETURN "https://api-testnet.bybit.com"
    RETURN "https://api.bybit.com"

// Callback setters (same as Binance)
FUNCTION (bd *BybitDriver) SetTickerCallback(cb func(*domain.Ticker))
FUNCTION (bd *BybitDriver) SetOrderBookCallback(cb func(*domain.OrderBook, bool))
FUNCTION (bd *BybitDriver) SetTradeCallback(cb func(*domain.Trade))
FUNCTION (bd *BybitDriver) SetOrderCallback(cb func(*domain.Order))
FUNCTION (bd *BybitDriver) SetPositionCallback(cb func(*domain.Position))
FUNCTION (bd *BybitDriver) SetConnectionCallback(cb func(domain.ConnectionStatus))

// Register driver
FUNCTION init():
    driver.Register("bybit", FUNC(cfg config.ExchangeConfig) (driver.Driver, error):
        config := &Config{
            APIKey: cfg.APIKey,
            APISecret: cfg.APISecret,
            Testnet: cfg.Testnet,
        }
        RETURN NewBybitDriver(config)
    )
```

**Estimated Time**: 4 hours

---

#### Task 5.2: Bybit - API Types (V5)

**Objective**: Define Bybit V5 API types.

**Files**:

- `internal/driver/bybit/types.go`
- `internal/driver/bybit/types_request.go`
- `internal/driver/bybit/types_response.go`
- `internal/driver/bybit/types_ws.go`

**Pseudocode**:

```
// types.go
TYPE BybitResponse STRUCT:
    RetCode int `json:"retCode"`
    RetMsg string `json:"retMsg"`
    Result json.RawMessage `json:"result"`
    RetExtInfo map[string]interface{} `json:"retExtInfo"`
    Time int64 `json:"time"`

// types_request.go (V5 format)
TYPE PlaceOrderRequest STRUCT:
    Category string `json:"category"` // spot, linear, inverse
    Symbol string `json:"symbol"`
    Side string `json:"side"`
    OrderType string `json:"orderType"`
    Qty string `json:"qty"`
    Price string `json:"price,omitempty"`
    TimeInForce string `json:"timeInForce,omitempty"`
    OrderLinkID string `json:"orderLinkId,omitempty"`
    IsLeverage int `json:"isLeverage,omitempty"`
    OrderFilter string `json:"orderFilter,omitempty"`

TYPE CancelOrderRequest STRUCT:
    Category string `json:"category"`
    Symbol string `json:"symbol"`
    OrderID string `json:"orderId,omitempty"`
    OrderLinkID string `json:"orderLinkId,omitempty"`
    OrderFilter string `json:"orderFilter,omitempty"`

TYPE QueryOrderRequest STRUCT:
    Category string `json:"category"`
    Symbol string `json:"symbol,omitempty"`
    OrderID string `json:"orderId,omitempty"`
    OrderLinkID string `json:"orderLinkId,omitempty"`
    OrderFilter string `json:"orderFilter,omitempty"`

TYPE GetWalletBalanceRequest STRUCT:
    AccountType string `json:"accountType"` // UNIFIED, CONTRACT, SPOT
    Coin string `json:"coin,omitempty"`

// types_response.go (V5 format)
TYPE OrderResult STRUCT:
    OrderID string `json:"orderId"`
    OrderLinkID string `json:"orderLinkId"`

TYPE OrderInfo STRUCT:
    OrderID string `json:"orderId"`
    OrderLinkID string `json:"orderLinkId"`
    Symbol string `json:"symbol"`
    Price string `json:"price"`
    Qty string `json:"qty"`
    Side string `json:"side"`
    OrderType string `json:"orderType"`
    TimeInForce string `json:"timeInForce"`
    OrderStatus string `json:"orderStatus"`
    CancelType string `json:"cancelType"`
    RejectReason string `json:"rejectReason"`
    AvgPrice string `json:"avgPrice"`
    LeavesQty string `json:"leavesQty"`
    CumExecQty string `json:"cumExecQty"`
    CumExecValue string `json:"cumExecValue"`
    CumExecFee string `json:"cumExecFee"`
    CreatedTime string `json:"createdTime"`
    UpdatedTime string `json:"updatedTime"`

TYPE OpenOrdersResult STRUCT:
    List []OrderInfo `json:"list"`
    NextPageCursor string `json:"nextPageCursor"`
    Category string `json:"category"`

TYPE WalletBalanceResult STRUCT:
    List []AccountInfo `json:"list"`

TYPE AccountInfo STRUCT:
    TotalEquity string `json:"totalEquity"`
    AccountIMRate string `json:"accountIMRate"`
    TotalMarginBalance string `json:"totalMarginBalance"`
    TotalInitialMargin string `json:"totalInitialMargin"`
    AccountType string `json:"accountType"`
    TotalAvailableBalance string `json:"totalAvailableBalance"`
    TotalPerpUPL string `json:"totalPerpUPL"`
    TotalWalletBalance string `json:"totalWalletBalance"`
    Coin []CoinBalance `json:"coin"`

TYPE CoinBalance STRUCT:
    Coin string `json:"coin"`
    Equity string `json:"equity"`
    UsdValue string `json:"usdValue"`
    WalletBalance string `json:"walletBalance"`
    Free string `json:"free"`
    Locked string `json:"locked"`
    BorrowAmount string `json:"borrowAmount"`
    AvailableToBorrow string `json:"availableToBorrow"`
    AvailableToWithdraw string `json:"availableToWithdraw"`

TYPE InstrumentsInfoResult STRUCT:
    Category string `json:"category"`
    List []InstrumentInfo `json:"list"`
    NextPageCursor string `json:"nextPageCursor"`

TYPE InstrumentInfo STRUCT:
    Symbol string `json:"symbol"`
    BaseCoin string `json:"baseCoin"`
    QuoteCoin string `json:"quoteCoin"`
    Innovation string `json:"innovation"`
    Status string `json:"status"`
    MarginTrading string `json:"marginTrading"`
    LotSizeFilter LotSizeFilter `json:"lotSizeFilter"`
    PriceFilter PriceFilter `json:"priceFilter"`

TYPE LotSizeFilter STRUCT:
    BasePrecision string `json:"basePrecision"`
    QuotePrecision string `json:"quotePrecision"`
    MinOrderQty string `json:"minOrderQty"`
    MaxOrderQty string `json:"maxOrderQty"`
    MinOrderAmt string `json:"minOrderAmt"`
    MaxOrderAmt string `json:"maxOrderAmt"`

TYPE PriceFilter STRUCT:
    TickSize string `json:"tickSize"`

TYPE ServerTimeResult STRUCT:
    TimeSecond string `json:"timeSecond"`
    TimeNano string `json:"timeNano"`

// types_ws.go (V5 format)
TYPE WSRequest STRUCT:
    ReqID string `json:"req_id,omitempty"`
    Op string `json:"op"` // subscribe, unsubscribe, auth
    Args []string `json:"args,omitempty"`

TYPE WSResponse STRUCT:
    Success bool `json:"success"`
    RetMsg string `json:"ret_msg"`
    ConnID string `json:"conn_id"`
    ReqID string `json:"req_id,omitempty"`
    Op string `json:"op"`

TYPE WSMessage STRUCT:
    Topic string `json:"topic"`
    Type string `json:"type"` // snapshot, delta
    Ts int64 `json:"ts"`
    Data json.RawMessage `json:"data"`

TYPE WSTicker STRUCT:
    Symbol string `json:"symbol"`
    LastPrice string `json:"lastPrice"`
    HighPrice24h string `json:"highPrice24h"`
    LowPrice24h string `json:"lowPrice24h"`
    PrevPrice24h string `json:"prevPrice24h"`
    Volume24h string `json:"volume24h"`
    Turnover24h string `json:"turnover24h"`
    Price24hPcnt string `json:"price24hPcnt"`
    UsdIndexPrice string `json:"usdIndexPrice"`

TYPE WSOrderBook STRUCT:
    Symbol string `json:"s"`
    Bids [][]string `json:"b"`
    Asks [][]string `json:"a"`
    UpdateID int64 `json:"u"`
    Seq int64 `json:"seq"`

TYPE WSPublicTrade STRUCT:
    Timestamp int64 `json:"T"`
    Symbol string `json:"s"`
    Side string `json:"S"`
    Size string `json:"v"`
    Price string `json:"p"`
    Direction string `json:"L"`
    TradeID string `json:"i"`
    BT bool `json:"BT"`

TYPE WSOrder STRUCT:
    Symbol string `json:"symbol"`
    OrderID string `json:"orderId"`
    Side string `json:"side"`
    OrderType string `json:"orderType"`
    CancelType string `json:"cancelType"`
    Price string `json:"price"`
    Qty string `json:"qty"`
    OrderIv string `json:"orderIv"`
    TimeInForce string `json:"timeInForce"`
    OrderStatus string `json:"orderStatus"`
    OrderLinkID string `json:"orderLinkId"`
    LastPriceOnCreated string `json:"lastPriceOnCreated"`
    ReduceOnly bool `json:"reduceOnly"`
    LeavesQty string `json:"leavesQty"`
    LeavesValue string `json:"leavesValue"`
    CumExecQty string `json:"cumExecQty"`
    CumExecValue string `json:"cumExecValue"`
    AvgPrice string `json:"avgPrice"`
    BlockTradeID string `json:"blockTradeId"`
    PositionIdx int `json:"positionIdx"`
    CumExecFee string `json:"cumExecFee"`
    CreatedTime string `json:"createdTime"`
    UpdatedTime string `json:"updatedTime"`
    RejectReason string `json:"rejectReason"`
    StopOrderType string `json:"stopOrderType"`
    TpslMode string `json:"tpslMode"`
    TriggerPrice string `json:"triggerPrice"`
    TakeProfit string `json:"takeProfit"`
    StopLoss string `json:"stopLoss"`
    TpTriggerBy string `json:"tpTriggerBy"`
    SlTriggerBy string `json:"slTriggerBy"`
    TpLimitPrice string `json:"tpLimitPrice"`
    SlLimitPrice string `json:"slLimitPrice"`
    Category string `json:"category"`
```

**Estimated Time**: 6 hours

---

#### Task 5.3: Bybit - Request Signer (V5 Format)

**Objective**: Implement HMAC-SHA256 signing for Bybit V5 API.

**Files**: `internal/driver/bybit/signer.go`

**Pseudocode**:

```
TYPE Signer STRUCT:
    apiKey string
    apiSecret string

FUNCTION NewSigner(apiKey, apiSecret string) *Signer:
    RETURN &Signer{
        apiKey: apiKey,
        apiSecret: apiSecret,
    }

FUNCTION (s *Signer) Sign(timestamp int64, body string) string:
    // Bybit V5: timestamp + apiKey + recvWindow + body
    recvWindow := 5000
    message := fmt.Sprintf("%d%s%d%s", timestamp, s.apiKey, recvWindow, body)

    mac := hmac.New(sha256.New, []byte(s.apiSecret))
    mac.Write([]byte(message))

    signature := hex.EncodeToString(mac.Sum(nil))
    RETURN signature

FUNCTION (s *Signer) SignRequest(method string, params interface{}) (http.Header, string, error):
    timestamp := time.Now().UnixMilli()

    // Convert params to JSON
    var body string
    IF params != nil THEN
        bodyBytes, err := json.Marshal(params)
        IF err != nil THEN
            RETURN nil, "", err
        body = string(bodyBytes)

    // Generate signature
    signature := s.Sign(timestamp, body)

    // Create headers
    headers := http.Header{}
    headers.Set("X-BAPI-API-KEY", s.apiKey)
    headers.Set("X-BAPI-SIGN", signature)
    headers.Set("X-BAPI-SIGN-TYPE", "2") // HMAC-SHA256
    headers.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(timestamp, 10))
    headers.Set("X-BAPI-RECV-WINDOW", "5000")
    headers.Set("Content-Type", "application/json")

    RETURN headers, body, nil

FUNCTION (s *Signer) SignWebSocket() (int64, string):
    // WebSocket auth format
    expires := time.Now().UnixMilli() + 10000 // 10 seconds expiry

    message := fmt.Sprintf("GET/realtime%d", expires)
    mac := hmac.New(sha256.New, []byte(s.apiSecret))
    mac.Write([]byte(message))

    signature := hex.EncodeToString(mac.Sum(nil))
    RETURN expires, signature
```

**Estimated Time**: 3 hours

---

#### Task 5.4: Bybit - REST Client (V5 API)

**Objective**: Implement REST client with V5 endpoints.

**Files**:

- `internal/driver/bybit/rest.go`
- `internal/driver/bybit/rest_market.go`
- `internal/driver/bybit/rest_trading.go`
- `internal/driver/bybit/rest_account.go`

**Pseudocode**:

```
// rest.go
TYPE RESTClient STRUCT:
    config *Config
    signer *Signer
    client *resty.Client
    baseURL string
    rateLimiter *ratelimit.TokenBucket

FUNCTION NewRESTClient(config *Config) *RESTClient:
    signer := NewSigner(config.APIKey, config.APISecret)

    client := resty.New()
    client.SetTimeout(config.Timeout)
    client.SetRetryCount(3)
    client.SetRetryWaitTime(1 * time.Second)
    client.SetRetryMaxWaitTime(5 * time.Second)

    baseURL := getRESTBaseURL(config.Testnet)

    // Initialize rate limiter
    rateLimiter := ratelimit.NewTokenBucket(
        config.RateLimit.RequestsPerSecond,
        config.RateLimit.BurstSize,
    )

    RETURN &RESTClient{
        config: config,
        signer: signer,
        client: client,
        baseURL: baseURL,
        rateLimiter: rateLimiter,
    }

FUNCTION (rc *RESTClient) do(method, endpoint string, params interface{}, result interface{}) error:
    // Check rate limit
    IF NOT rc.rateLimiter.Allow() THEN
        RETURN errors.NewRateLimitError("bybit", 0, rc.rateLimiter.WaitTime())

    // Build URL
    url := rc.baseURL + endpoint

    var resp *resty.Response
    var err error

    SWITCH method:
        CASE "GET":
            // For GET, convert params to query string
            IF params != nil THEN
                query := url.Values{}
                // Marshal params to map then to query
                paramMap := convertToMap(params)
                FOR key, value IN paramMap DO
                    query.Set(key, fmt.Sprint(value))
                url += "?" + query.Encode()

            resp, err = rc.client.R().
                SetResult(&BybitResponse{}).
                Get(url)

        CASE "POST":
            resp, err = rc.client.R().
                SetBody(params).
                SetResult(&BybitResponse{}).
                Post(url)

    IF err != nil THEN
        RETURN fmt.Errorf("request failed: %w", err)

    // Parse Bybit response wrapper
    bybitResp := resp.Result().(*BybitResponse)

    IF bybitResp.RetCode != 0 THEN
        RETURN MapBybitError(bybitResp.RetCode, bybitResp.RetMsg)

    // Unmarshal result
    IF result != nil AND len(bybitResp.Result) > 0 THEN
        err = json.Unmarshal(bybitResp.Result, result)
        IF err != nil THEN
            RETURN fmt.Errorf("failed to unmarshal result: %w", err)

    RETURN nil

FUNCTION (rc *RESTClient) doSigned(method, endpoint string, params interface{}, result interface{}) error:
    // Sign request
    headers, body, err := rc.signer.SignRequest(method, params)
    IF err != nil THEN
        RETURN err

    // Set headers on client
    FOR key, values IN headers DO
        FOR EACH value IN values DO
            rc.client.SetHeader(key, value)

    // Execute request
    url := rc.baseURL + endpoint

    var resp *resty.Response

    SWITCH method:
        CASE "GET":
            // For GET with signed params, params are in query string
            IF params != nil THEN
                query := url.Values{}
                paramMap := convertToMap(params)
                FOR key, value IN paramMap DO
                    query.Set(key, fmt.Sprint(value))
                url += "?" + query.Encode()

            resp, err = rc.client.R().
                SetResult(&BybitResponse{}).
                Get(url)

        CASE "POST":
            resp, err = rc.client.R().
                SetBody(body).
                SetHeader("Content-Type", "application/json").
                SetResult(&BybitResponse{}).
                Post(url)

        CASE "DELETE":
            IF params != nil THEN
                query := url.Values{}
                paramMap := convertToMap(params)
                FOR key, value IN paramMap DO
                    query.Set(key, fmt.Sprint(value))
                url += "?" + query.Encode()

            resp, err = rc.client.R().
                SetResult(&BybitResponse{}).
                Delete(url)

    IF err != nil THEN
        RETURN fmt.Errorf("request failed: %w", err)

    // Parse response
    bybitResp := resp.Result().(*BybitResponse)

    IF bybitResp.RetCode != 0 THEN
        RETURN MapBybitError(bybitResp.RetCode, bybitResp.RetMsg)

    // Unmarshal result
    IF result != nil AND len(bybitResp.Result) > 0 THEN
        err = json.Unmarshal(bybitResp.Result, result)
        IF err != nil THEN
            RETURN fmt.Errorf("failed to unmarshal result: %w", err)

    RETURN nil

// rest_market.go
FUNCTION (rc *RESTClient) GetServerTime() (int64, error):
    var result ServerTimeResult
    err := rc.do("GET", "/v5/market/time", nil, &result)
    IF err != nil THEN
        RETURN 0, err

    timeSecond, _ := strconv.ParseInt(result.TimeSecond, 10, 64)
    RETURN timeSecond * 1000, nil // Convert to milliseconds

FUNCTION (rc *RESTClient) GetInstrumentsInfo(category, symbol string) ([]*domain.SymbolInfo, error):
    params := map[string]interface{}{
        "category": category,
    }
    IF symbol != "" THEN
        params["symbol"] = normalizeSymbol(symbol)

    var result InstrumentsInfoResult
    err := rc.do("GET", "/v5/market/instruments-info", params, &result)
    IF err != nil THEN
        RETURN nil, err

    symbols := make([]*domain.SymbolInfo, len(result.List))
    FOR i, info IN result.List DO
        symbols[i] = mapInstrumentInfo(&info)

    RETURN symbols, nil

FUNCTION (rc *RESTClient) GetOrderBook(symbol string, limit int) (*domain.OrderBook, error):
    params := map[string]interface{}{
        "category": "spot",
        "symbol": normalizeSymbol(symbol),
    }
    IF limit > 0 THEN
        params["limit"] = limit

    var result struct {
        Symbol string `json:"s"`
        Bids [][]string `json:"b"`
        Asks [][]string `json:"a"`
        Timestamp int64 `json:"ts"`
        UpdateID int64 `json:"u"`
    }

    err := rc.do("GET", "/v5/market/orderbook", params, &result)
    IF err != nil THEN
        RETURN nil, err

    // Map to domain.OrderBook
    orderBook := &domain.OrderBook{
        Exchange: "bybit",
        Symbol: denormalizeSymbol(result.Symbol),
        SequenceNum: result.UpdateID,
        Timestamp: time.UnixMilli(result.Timestamp),
    }

    FOR EACH bid IN result.Bids DO
        price, _ := domain.NewDecimal(bid[0])
        qty, _ := domain.NewDecimal(bid[1])
        orderBook.Bids = append(orderBook.Bids, domain.OrderBookLevel{
            Price: price,
            Quantity: qty,
        })

    FOR EACH ask IN result.Asks DO
        price, _ := domain.NewDecimal(ask[0])
        qty, _ := domain.NewDecimal(ask[1])
        orderBook.Asks = append(orderBook.Asks, domain.OrderBookLevel{
            Price: price,
            Quantity: qty,
        })

    RETURN orderBook, nil

// rest_trading.go
FUNCTION (rc *RESTClient) PlaceOrder(req *domain.OrderRequest) (*domain.Order, error):
    params := PlaceOrderRequest{
        Category: "spot",
        Symbol: normalizeSymbol(req.Symbol),
        Side: mapSide(req.Side),
        OrderType: mapOrderType(req.Type),
        Qty: domain.String(req.Quantity),
    }

    IF req.Type != domain.OrderTypeMarket THEN
        params.Price = domain.String(req.Price)
        params.TimeInForce = mapTimeInForce(req.TimeInForce)

    IF req.ClientOrderID != "" THEN
        params.OrderLinkID = req.ClientOrderID

    var result OrderResult
    err := rc.doSigned("POST", "/v5/order/create", params, &result)
    IF err != nil THEN
        RETURN nil, err

    // Need to query order to get full details
    order, err := rc.GetOrder(req.Symbol, result.OrderID)
    IF err != nil THEN
        // Return partial order info
        RETURN &domain.Order{
            ID: result.OrderID,
            ClientOrderID: result.OrderLinkID,
            Exchange: "bybit",
            Symbol: req.Symbol,
            Side: req.Side,
            Type: req.Type,
            Status: domain.OrderStatusNew,
            Price: req.Price,
            Quantity: req.Quantity,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }, nil

    RETURN order, nil

FUNCTION (rc *RESTClient) CancelOrder(req *domain.CancelRequest) error:
    params := CancelOrderRequest{
        Category: "spot",
        Symbol: normalizeSymbol(req.Symbol),
    }

    IF req.OrderID != "" THEN
        params.OrderID = req.OrderID
    ELSE IF req.ClientOrderID != "" THEN
        params.OrderLinkID = req.ClientOrderID
    ELSE
        RETURN errors.New("either OrderID or ClientOrderID must be provided")

    var result OrderResult
    err := rc.doSigned("POST", "/v5/order/cancel", params, &result)
    RETURN err

FUNCTION (rc *RESTClient) GetOrder(symbol, orderID string) (*domain.Order, error):
    params := QueryOrderRequest{
        Category: "spot",
        Symbol: normalizeSymbol(symbol),
    }

    // Bybit can query by orderID or orderLinkId
    IF strings.HasPrefix(orderID, "client-") THEN
        params.OrderLinkID = orderID
    ELSE
        params.OrderID = orderID

    var result struct {
        List []OrderInfo `json:"list"`
    }

    err := rc.doSigned("GET", "/v5/order/realtime", params, &result)
    IF err != nil THEN
        RETURN nil, err

    IF len(result.List) == 0 THEN
        RETURN nil, errors.ErrOrderNotFound

    order := mapOrderInfo(&result.List[0])
    RETURN order, nil

FUNCTION (rc *RESTClient) GetOpenOrders(symbol string) ([]*domain.Order, error):
    params := QueryOrderRequest{
        Category: "spot",
        OrderFilter: "Order", // Only active orders
    }

    IF symbol != "" THEN
        params.Symbol = normalizeSymbol(symbol)

    var result OpenOrdersResult
    err := rc.doSigned("GET", "/v5/order/realtime", params, &result)
    IF err != nil THEN
        RETURN nil, err

    orders := make([]*domain.Order, len(result.List))
    FOR i, orderInfo IN result.List DO
        orders[i] = mapOrderInfo(&orderInfo)

    RETURN orders, nil

// rest_account.go
FUNCTION (rc *RESTClient) GetWalletBalance(accountType, coin string) (*WalletBalanceResult, error):
    params := GetWalletBalanceRequest{
        AccountType: accountType, // UNIFIED or SPOT
    }

    IF coin != "" THEN
        params.Coin = coin

    var result WalletBalanceResult
    err := rc.doSigned("GET", "/v5/account/wallet-balance", params, &result)
    IF err != nil THEN
        RETURN nil, err

    RETURN &result, nil

FUNCTION (rc *RESTClient) GetBalance(asset string) (*domain.Balance, error):
    // Try UNIFIED account first (V5 default)
    wallet, err := rc.GetWalletBalance("UNIFIED", asset)
    IF err != nil THEN
        // Fallback to SPOT account
        wallet, err = rc.GetWalletBalance("SPOT", asset)
        IF err != nil THEN
            RETURN nil, err

    IF len(wallet.List) == 0 THEN
        RETURN nil, errors.ErrBalanceNotFound

    account := wallet.List[0]

    // Find coin in account
    FOR EACH coinBalance IN account.Coin DO
        IF coinBalance.Coin == asset THEN
            free, _ := domain.NewDecimal(coinBalance.Free)
            locked, _ := domain.NewDecimal(coinBalance.Locked)

            RETURN &domain.Balance{
                Exchange: "bybit",
                Asset: asset,
                Free: free,
                Locked: locked,
                Timestamp: time.Now(),
            }, nil

    RETURN nil, errors.ErrBalanceNotFound

FUNCTION convertToMap(v interface{}) map[string]interface{}:
    data, _ := json.Marshal(v)
    result := make(map[string]interface{})
    json.Unmarshal(data, &result)
    RETURN result
```

**Estimated Time**: 12 hours

---

#### Task 5.5: Bybit - Domain Mapper

**Objective**: Map Bybit V5 types to domain types.

**Files**:

- `internal/driver/bybit/mapper.go`
- `internal/driver/bybit/mapper_market.go`
- `internal/driver/bybit/mapper_order.go`
- `internal/driver/bybit/normalizer.go`

**Pseudocode**:

```
// normalizer.go
FUNCTION normalizeSymbol(symbol string) string:
    // Convert domain symbol to Bybit format
    // BTC/USDT -> BTCUSDT
    RETURN strings.ReplaceAll(symbol, "/", "")

FUNCTION denormalizeSymbol(symbol string) string:
    // Convert Bybit symbol to domain format
    // BTCUSDT -> BTC/USDT
    IF strings.HasSuffix(symbol, "USDT") THEN
        base := symbol[:len(symbol)-4]
        RETURN base + "/USDT"
    ELSE IF strings.HasSuffix(symbol, "USDC") THEN
        base := symbol[:len(symbol)-4]
        RETURN base + "/USDC"
    ELSE IF strings.HasSuffix(symbol, "BTC") THEN
        base := symbol[:len(symbol)-3]
        RETURN base + "/BTC"
    RETURN symbol

FUNCTION mapSide(side domain.Side) string:
    SWITCH side:
        CASE domain.SideBuy:
            RETURN "Buy"
        CASE domain.SideSell:
            RETURN "Sell"
        DEFAULT:
            RETURN ""

FUNCTION unmapSide(side string) domain.Side:
    SWITCH side:
        CASE "Buy":
            RETURN domain.SideBuy
        CASE "Sell":
            RETURN domain.SideSell
        DEFAULT:
            RETURN domain.SideBuy

FUNCTION mapOrderType(orderType domain.OrderType) string:
    SWITCH orderType:
        CASE domain.OrderTypeMarket:
            RETURN "Market"
        CASE domain.OrderTypeLimit:
            RETURN "Limit"
        DEFAULT:
            RETURN "Limit"

FUNCTION unmapOrderType(orderType string) domain.OrderType:
    SWITCH orderType:
        CASE "Market":
            RETURN domain.OrderTypeMarket
        CASE "Limit":
            RETURN domain.OrderTypeLimit
        DEFAULT:
            RETURN domain.OrderTypeLimit

FUNCTION mapTimeInForce(tif domain.TimeInForce) string:
    SWITCH tif:
        CASE domain.TimeInForceGTC:
            RETURN "GTC"
        CASE domain.TimeInForceIOC:
            RETURN "IOC"
        CASE domain.TimeInForceFOK:
            RETURN "FOK"
        DEFAULT:
            RETURN "GTC"

FUNCTION unmapTimeInForce(tif string) domain.TimeInForce:
    SWITCH tif:
        CASE "GTC":
            RETURN domain.TimeInForceGTC
        CASE "IOC":
            RETURN domain.TimeInForceIOC
        CASE "FOK":
            RETURN domain.TimeInForceFOK
        DEFAULT:
            RETURN domain.TimeInForceGTC

FUNCTION mapOrderStatus(status string) domain.OrderStatus:
    SWITCH status:
        CASE "Created", "New":
            RETURN domain.OrderStatusNew
        CASE "PartiallyFilled":
            RETURN domain.OrderStatusPartiallyFilled
        CASE "Filled":
            RETURN domain.OrderStatusFilled
        CASE "Cancelled":
            RETURN domain.OrderStatusCanceled
        CASE "Rejected":
            RETURN domain.OrderStatusRejected
        CASE "Deactivated":
            RETURN domain.OrderStatusExpired
        DEFAULT:
            RETURN domain.OrderStatusNew

// mapper_market.go
FUNCTION mapTicker(wsTicker *WSTicker) *domain.Ticker:
    bid := domain.Zero()
    ask := domain.Zero()
    last, _ := domain.NewDecimal(wsTicker.LastPrice)
    volume, _ := domain.NewDecimal(wsTicker.Volume24h)
    high, _ := domain.NewDecimal(wsTicker.HighPrice24h)
    low, _ := domain.NewDecimal(wsTicker.LowPrice24h)

    // Bybit tickers don't have bid/ask in ticker stream
    // Need to use orderbook for that

    RETURN &domain.Ticker{
        Exchange: "bybit",
        Symbol: denormalizeSymbol(wsTicker.Symbol),
        Bid: bid,
        Ask: ask,
        BidQty: domain.Zero(),
        AskQty: domain.Zero(),
        Last: last,
        Volume24h: volume,
        High24h: high,
        Low24h: low,
        Timestamp: time.Now(),
    }

FUNCTION mapOrderBookUpdate(wsOrderBook *WSOrderBook, isSnapshot bool) (*domain.OrderBook, bool):
    orderBook := &domain.OrderBook{
        Exchange: "bybit",
        Symbol: denormalizeSymbol(wsOrderBook.Symbol),
        SequenceNum: wsOrderBook.UpdateID,
        Timestamp: time.Now(),
    }

    FOR EACH bid IN wsOrderBook.Bids DO
        price, _ := domain.NewDecimal(bid[0])
        qty, _ := domain.NewDecimal(bid[1])

        orderBook.Bids = append(orderBook.Bids, domain.OrderBookLevel{
            Price: price,
            Quantity: qty,
        })

    FOR EACH ask IN wsOrderBook.Asks DO
        price, _ := domain.NewDecimal(ask[0])
        qty, _ := domain.NewDecimal(ask[1])

        orderBook.Asks = append(orderBook.Asks, domain.OrderBookLevel{
            Price: price,
            Quantity: qty,
        })

    RETURN orderBook, isSnapshot

FUNCTION mapPublicTrade(wsTrade *WSPublicTrade) *domain.Trade:
    price, _ := domain.NewDecimal(wsTrade.Price)
    qty, _ := domain.NewDecimal(wsTrade.Size)

    side := unmapSide(wsTrade.Side)

    RETURN &domain.Trade{
        Exchange: "bybit",
        Symbol: denormalizeSymbol(wsTrade.Symbol),
        TradeID: wsTrade.TradeID,
        Price: price,
        Quantity: qty,
        Side: side,
        Timestamp: time.UnixMilli(wsTrade.Timestamp),
        IsBuyerMaker: wsTrade.Side == "Sell",
    }

// mapper_order.go
FUNCTION mapOrderInfo(orderInfo *OrderInfo) *domain.Order:
    price, _ := domain.NewDecimal(orderInfo.Price)
    qty, _ := domain.NewDecimal(orderInfo.Qty)
    filledQty, _ := domain.NewDecimal(orderInfo.CumExecQty)
    avgPrice, _ := domain.NewDecimal(orderInfo.AvgPrice)

    // Parse timestamps (Bybit uses string timestamps in milliseconds)
    createdTime, _ := strconv.ParseInt(orderInfo.CreatedTime, 10, 64)
    updatedTime, _ := strconv.ParseInt(orderInfo.UpdatedTime, 10, 64)

    RETURN &domain.Order{
        ID: orderInfo.OrderID,
        ClientOrderID: orderInfo.OrderLinkID,
        Exchange: "bybit",
        Symbol: denormalizeSymbol(orderInfo.Symbol),
        Side: unmapSide(orderInfo.Side),
        Type: unmapOrderType(orderInfo.OrderType),
        Status: mapOrderStatus(orderInfo.OrderStatus),
        Price: price,
        Quantity: qty,
        FilledQty: filledQty,
        AvgFillPrice: avgPrice,
        TimeInForce: unmapTimeInForce(orderInfo.TimeInForce),
        CreatedAt: time.UnixMilli(createdTime),
        UpdatedAt: time.UnixMilli(updatedTime),
    }

FUNCTION mapWSOrder(wsOrder *WSOrder) *domain.Order:
    price, _ := domain.NewDecimal(wsOrder.Price)
    qty, _ := domain.NewDecimal(wsOrder.Qty)
    filledQty, _ := domain.NewDecimal(wsOrder.CumExecQty)
    avgPrice, _ := domain.NewDecimal(wsOrder.AvgPrice)

    createdTime, _ := strconv.ParseInt(wsOrder.CreatedTime, 10, 64)
    updatedTime, _ := strconv.ParseInt(wsOrder.UpdatedTime, 10, 64)

    RETURN &domain.Order{
        ID: wsOrder.OrderID,
        ClientOrderID: wsOrder.OrderLinkID,
        Exchange: "bybit",
        Symbol: denormalizeSymbol(wsOrder.Symbol),
        Side: unmapSide(wsOrder.Side),
        Type: unmapOrderType(wsOrder.OrderType),
        Status: mapOrderStatus(wsOrder.OrderStatus),
        Price: price,
        Quantity: qty,
        FilledQty: filledQty,
        AvgFillPrice: avgPrice,
        TimeInForce: unmapTimeInForce(wsOrder.TimeInForce),
        CreatedAt: time.UnixMilli(createdTime),
        UpdatedAt: time.UnixMilli(updatedTime),
    }

FUNCTION mapInstrumentInfo(info *InstrumentInfo) *domain.SymbolInfo:
    minQty, _ := domain.NewDecimal(info.LotSizeFilter.MinOrderQty)
    maxQty, _ := domain.NewDecimal(info.LotSizeFilter.MaxOrderQty)
    tickSize, _ := domain.NewDecimal(info.PriceFilter.TickSize)

    // Calculate step size from precision
    basePrecision, _ := strconv.Atoi(info.LotSizeFilter.BasePrecision)
    stepSize := domain.NewDecimalFromFloat64(math.Pow(10, -float64(basePrecision)))

    RETURN &domain.SymbolInfo{
        Exchange: "bybit",
        Symbol: denormalizeSymbol(info.Symbol),
        BaseAsset: info.BaseCoin,
        QuoteAsset: info.QuoteCoin,
        Status: info.Status,
        PriceFilter: domain.PriceFilter{
            MinPrice: domain.Zero(), // Bybit doesn't provide min/max price
            MaxPrice: domain.Zero(),
            TickSize: tickSize,
        },
        LotSizeFilter: domain.LotSizeFilter{
            MinQty: minQty,
            MaxQty: maxQty,
            StepSize: stepSize,
        },
        MinNotional: domain.Zero(), // Calculate from minOrderAmt if needed
        Permissions: []string{"spot"},
    }
```

**Estimated Time**: 8 hours

---

#### Task 5.6: Bybit - WebSocket Clients

**Objective**: Implement public and private WebSocket clients for V5 API.

**Files**:

- `internal/driver/bybit/ws.go`
- `internal/driver/bybit/ws_public.go`
- `internal/driver/bybit/ws_private.go`
- `internal/driver/bybit/ws_handlers.go`

**Pseudocode**:

```
// ws.go (shared base)
TYPE WSBase STRUCT:
    url string
    config *Config
    callbacks driver.Callbacks
    conn *gws.Conn
    subscriptions sync.Map
    reconnecting atomic.Bool
    connected atomic.Bool
    pingTicker *time.Ticker
    mutex sync.RWMutex

FUNCTION (wb *WSBase) sendPing():
    msg := map[string]interface{}{
        "op": "ping",
    }
    data, _ := json.Marshal(msg)
    wb.conn.WriteMessage(gws.OpcodeText, data)

FUNCTION (wb *WSBase) startPingLoop():
    wb.pingTicker = time.NewTicker(20 * time.Second)
    GO FUNC():
        FOR range wb.pingTicker.C DO
            IF wb.IsConnected() THEN
                wb.sendPing()

FUNCTION (wb *WSBase) stopPingLoop():
    IF wb.pingTicker != nil THEN
        wb.pingTicker.Stop()

FUNCTION (wb *WSBase) IsConnected() bool:
    RETURN wb.connected.Load()

FUNCTION (wb *WSBase) notifyConnection(status domain.ConnectionStatus):
    IF wb.callbacks.OnConnection != nil THEN
        wb.callbacks.OnConnection(status)

FUNCTION (wb *WSBase) handleDisconnect():
    IF wb.reconnecting.CompareAndSwap(false, true) THEN
        wb.connected.Store(false)
        wb.stopPingLoop()
        wb.notifyConnection(domain.StatusDisconnected)

        IF wb.conn != nil THEN
            wb.conn.WriteClose(1000, []byte("normal closure"))
            wb.conn = nil

        GO wb.reconnect()

FUNCTION (wb *WSBase) reconnect():
    // Implemented by derived types

FUNCTION exponentialBackoff(attempt int, baseDelay time.Duration) time.Duration:
    delay := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
    maxDelay := 60 * time.Second
    IF delay > maxDelay THEN
        delay = maxDelay

    // Add jitter
    jitter := time.Duration(rand.Int63n(int64(delay / 4)))
    RETURN delay + jitter

// ws_public.go
TYPE WSPublicClient STRUCT:
    WSBase

FUNCTION NewWSPublicClient(url string, config *Config, callbacks driver.Callbacks) *WSPublicClient:
    RETURN &WSPublicClient{
        WSBase: WSBase{
            url: url,
            config: config,
            callbacks: callbacks,
        },
    }

FUNCTION (wsp *WSPublicClient) Connect() error:
    upgrader := gws.NewUpgrader(&wsp, &gws.ServerOption{
        ReadBufferSize: 4096,
        WriteBufferSize: 4096,
        CompressEnabled: false,
    })

    conn, _, err := upgrader.Dial(wsp.url)
    IF err != nil THEN
        wsp.notifyConnection(domain.StatusDisconnected)
        RETURN fmt.Errorf("failed to connect: %w", err)

    wsp.conn = conn
    wsp.connected.Store(true)
    wsp.notifyConnection(domain.StatusConnected)

    // Start ping loop
    wsp.startPingLoop()

    // Start read loop
    GO wsp.readLoop()

    // Resubscribe
    wsp.resubscribe()

    RETURN nil

FUNCTION (wsp *WSPublicClient) Disconnect():
    wsp.stopPingLoop()

    IF wsp.conn != nil THEN
        wsp.conn.WriteClose(1000, []byte("normal closure"))
        wsp.conn = nil
        wsp.connected.Store(false)
        wsp.notifyConnection(domain.StatusDisconnected)

FUNCTION (wsp *WSPublicClient) Subscribe(topic string) error:
    wsp.subscriptions.Store(topic, true)

    IF NOT wsp.IsConnected() THEN
        RETURN nil

    RETURN wsp.sendSubscribe(topic)

FUNCTION (wsp *WSPublicClient) Unsubscribe(topic string) error:
    wsp.subscriptions.Delete(topic)

    IF NOT wsp.IsConnected() THEN
        RETURN nil

    RETURN wsp.sendUnsubscribe(topic)

FUNCTION (wsp *WSPublicClient) sendSubscribe(topic string) error:
    msg := WSRequest{
        Op: "subscribe",
        Args: []string{topic},
    }

    data, err := json.Marshal(msg)
    IF err != nil THEN
        RETURN err

    RETURN wsp.conn.WriteMessage(gws.OpcodeText, data)

FUNCTION (wsp *WSPublicClient) sendUnsubscribe(topic string) error:
    msg := WSRequest{
        Op: "unsubscribe",
        Args: []string{topic},
    }

    data, err := json.Marshal(msg)
    IF err != nil THEN
        RETURN err

    RETURN wsp.conn.WriteMessage(gws.OpcodeText, data)

FUNCTION (wsp *WSPublicClient) resubscribe():
    topics := []string{}
    wsp.subscriptions.Range(FUNC(key, value interface{}):
        topics = append(topics, key.(string))
        RETURN true
    )

    FOR EACH topic IN topics DO
        err := wsp.sendSubscribe(topic)
        IF err != nil THEN
            LOG error subscribing to topic

FUNCTION (wsp *WSPublicClient) readLoop():
    FOR wsp.IsConnected() DO
        msgType, data, err := wsp.conn.ReadMessage()
        IF err != nil THEN
            LOG error reading message
            wsp.handleDisconnect()
            BREAK

        IF msgType == gws.OpcodeText THEN
            wsp.handleMessage(data)

FUNCTION (wsp *WSPublicClient) reconnect():
    attempt := 0
    FOR attempt < wsp.config.MaxReconnectAttempts DO
        wsp.notifyConnection(domain.StatusReconnecting)

        delay := exponentialBackoff(attempt, wsp.config.ReconnectDelay)
        time.Sleep(delay)

        err := wsp.Connect()
        IF err == nil THEN
            wsp.reconnecting.Store(false)
            wsp.notifyConnection(domain.StatusConnected)
            LOG "successfully reconnected to public websocket"
            RETURN

        LOG "public websocket reconnection attempt failed" with error
        attempt++

    wsp.reconnecting.Store(false)
    wsp.notifyConnection(domain.StatusDisconnected)
    LOG "failed to reconnect public websocket after max attempts"

// gws.EventHandler implementation
FUNCTION (wsp *WSPublicClient) OnOpen(conn *gws.Conn):
    LOG "public websocket connection opened"

FUNCTION (wsp *WSPublicClient) OnClose(conn *gws.Conn, code uint16, reason []byte):
    LOG "public websocket connection closed" with code and reason
    wsp.handleDisconnect()

FUNCTION (wsp *WSPublicClient) OnPing(conn *gws.Conn, payload []byte) error:
    RETURN conn.WritePong(payload)

FUNCTION (wsp *WSPublicClient) OnPong(conn *gws.Conn, payload []byte) error:
    RETURN nil

FUNCTION (wsp *WSPublicClient) OnError(conn *gws.Conn, err error):
    LOG error with err

// ws_private.go
TYPE WSPrivateClient STRUCT:
    WSBase
    signer *Signer
    authenticated atomic.Bool

FUNCTION NewWSPrivateClient(url string, config *Config, callbacks driver.Callbacks) *WSPrivateClient:
    signer := NewSigner(config.APIKey, config.APISecret)

    RETURN &WSPrivateClient{
        WSBase: WSBase{
            url: url,
            config: config,
            callbacks: callbacks,
        },
        signer: signer,
    }

FUNCTION (wsp *WSPrivateClient) Connect() error:
    upgrader := gws.NewUpgrader(&wsp, &gws.ServerOption{
        ReadBufferSize: 4096,
        WriteBufferSize: 4096,
        CompressEnabled: false,
    })

    conn, _, err := upgrader.Dial(wsp.url)
    IF err != nil THEN
        wsp.notifyConnection(domain.StatusDisconnected)
        RETURN fmt.Errorf("failed to connect: %w", err)

    wsp.conn = conn
    wsp.connected.Store(true)

    // Authenticate
    err = wsp.authenticate()
    IF err != nil THEN
        wsp.Disconnect()
        RETURN fmt.Errorf("authentication failed: %w", err)

    wsp.notifyConnection(domain.StatusConnected)

    // Start ping loop
    wsp.startPingLoop()

    // Start read loop
    GO wsp.readLoop()

    // Resubscribe
    wsp.resubscribe()

    RETURN nil

FUNCTION (wsp *WSPrivateClient) authenticate() error:
    expires, signature := wsp.signer.SignWebSocket()

    msg := WSRequest{
        Op: "auth",
        Args: []string{
            wsp.config.APIKey,
            strconv.FormatInt(expires, 10),
            signature,
        },
    }

    data, err := json.Marshal(msg)
    IF err != nil THEN
        RETURN err

    err = wsp.conn.WriteMessage(gws.OpcodeText, data)
    IF err != nil THEN
        RETURN err

    // Wait for auth response (with timeout)
    timeout := time.After(5 * time.Second)
    authChan := make(chan bool, 1)

    // Set temporary handler for auth response
    wsp.authResponseChan = authChan

    SELECT:
        CASE success := <-authChan:
            IF success THEN
                wsp.authenticated.Store(true)
                RETURN nil
            RETURN errors.New("authentication rejected")

        CASE <-timeout:
            RETURN errors.New("authentication timeout")

FUNCTION (wsp *WSPrivateClient) Disconnect():
    wsp.stopPingLoop()
    wsp.authenticated.Store(false)

    IF wsp.conn != nil THEN
        wsp.conn.WriteClose(1000, []byte("normal closure"))
        wsp.conn = nil
        wsp.connected.Store(false)
        wsp.notifyConnection(domain.StatusDisconnected)

FUNCTION (wsp *WSPrivateClient) Subscribe(topic string) error:
    IF NOT wsp.authenticated.Load() THEN
        RETURN errors.New("not authenticated")

    wsp.subscriptions.Store(topic, true)

    IF NOT wsp.IsConnected() THEN
        RETURN nil

    RETURN wsp.sendSubscribe(topic)

FUNCTION (wsp *WSPrivateClient) sendSubscribe(topic string) error:
    msg := WSRequest{
        Op: "subscribe",
        Args: []string{topic},
    }

    data, err := json.Marshal(msg)
    IF err != nil THEN
        RETURN err

    RETURN wsp.conn.WriteMessage(gws.OpcodeText, data)

FUNCTION (wsp *WSPrivateClient) sendUnsubscribe(topic string) error:
    msg := WSRequest{
        Op: "unsubscribe",
        Args: []string{topic},
    }

    data, err := json.Marshal(msg)
    IF err != nil THEN
        RETURN err

    RETURN wsp.conn.WriteMessage(gws.OpcodeText, data)

FUNCTION (wsp *WSPrivateClient) resubscribe():
    topics := []string{}
    wsp.subscriptions.Range(FUNC(key, value interface{}):
        topics = append(topics, key.(string))
        RETURN true
    )

    FOR EACH topic IN topics DO
        err := wsp.sendSubscribe(topic)
        IF err != nil THEN
            LOG error subscribing to topic

FUNCTION (wsp *WSPrivateClient) readLoop():
    FOR wsp.IsConnected() DO
        msgType, data, err := wsp.conn.ReadMessage()
        IF err != nil THEN
            LOG error reading message
            wsp.handleDisconnect()
            BREAK

        IF msgType == gws.OpcodeText THEN
            wsp.handleMessage(data)

FUNCTION (wsp *WSPrivateClient) reconnect():
    attempt := 0
    FOR attempt < wsp.config.MaxReconnectAttempts DO
        wsp.notifyConnection(domain.StatusReconnecting)

        delay := exponentialBackoff(attempt, wsp.config.ReconnectDelay)
        time.Sleep(delay)

        err := wsp.Connect()
        IF err == nil THEN
            wsp.reconnecting.Store(false)
            wsp.notifyConnection(domain.StatusConnected)
            LOG "successfully reconnected to private websocket"
            RETURN

        LOG "private websocket reconnection attempt failed" with error
        attempt++

    wsp.reconnecting.Store(false)
    wsp.notifyConnection(domain.StatusDisconnected)
    LOG "failed to reconnect private websocket after max attempts"

// gws.EventHandler implementation (same as public)
FUNCTION (wsp *WSPrivateClient) OnOpen(conn *gws.Conn)
FUNCTION (wsp *WSPrivateClient) OnClose(conn *gws.Conn, code uint16, reason []byte)
FUNCTION (wsp *WSPrivateClient) OnPing(conn *gws.Conn, payload []byte) error
FUNCTION (wsp *WSPrivateClient) OnPong(conn *gws.Conn, payload []byte) error
FUNCTION (wsp *WSPrivateClient) OnError(conn *gws.Conn, err error)

// ws_handlers.go
FUNCTION (wsp *WSPublicClient) handleMessage(data []byte):
    // Try to parse as response (subscribe/unsubscribe confirmation)
    var wsResp WSResponse
    err := json.Unmarshal(data, &wsResp)
    IF err == nil AND wsResp.Op != "" THEN
        wsp.handleResponse(&wsResp)
        RETURN

    // Try to parse as pong
    var pong struct {
        Op string `json:"op"`
    }
    err = json.Unmarshal(data, &pong)
    IF err == nil AND pong.Op == "pong" THEN
        // Pong received, connection alive
        RETURN

    // Parse as data message
    var wsMsg WSMessage
    err = json.Unmarshal(data, &wsMsg)
    IF err != nil THEN
        LOG "failed to parse websocket message" with error
        RETURN

    wsp.handleDataMessage(&wsMsg)

FUNCTION (wsp *WSPublicClient) handleResponse(resp *WSResponse):
    IF NOT resp.Success THEN
        LOG "websocket operation failed" with resp.RetMsg
        RETURN

    SWITCH resp.Op:
        CASE "subscribe":
            LOG "subscribed successfully"
        CASE "unsubscribe":
            LOG "unsubscribed successfully"

FUNCTION (wsp *WSPublicClient) handleDataMessage(msg *WSMessage):
    topic := msg.Topic

    IF strings.Contains(topic, "tickers") THEN
        wsp.handleTicker(msg.Data)
    ELSE IF strings.Contains(topic, "orderbook") THEN
        isSnapshot := msg.Type == "snapshot"
        wsp.handleOrderBook(msg.Data, isSnapshot)
    ELSE IF strings.Contains(topic, "publicTrade") THEN
        wsp.handlePublicTrade(msg.Data)
    ELSE
        LOG "unknown topic" with topic

FUNCTION (wsp *WSPublicClient) handleTicker(data json.RawMessage):
    var ticker WSTicker
    err := json.Unmarshal(data, &ticker)
    IF err != nil THEN
        LOG error parsing ticker
        RETURN

    IF wsp.callbacks.OnTicker != nil THEN
        domainTicker := mapTicker(&ticker)
        wsp.callbacks.OnTicker(domainTicker)

FUNCTION (wsp *WSPublicClient) handleOrderBook(data json.RawMessage, isSnapshot bool):
    var orderBook WSOrderBook
    err := json.Unmarshal(data, &orderBook)
    IF err != nil THEN
        LOG error parsing orderbook
        RETURN

    IF wsp.callbacks.OnOrderBook != nil THEN
        domainOrderBook, _ := mapOrderBookUpdate(&orderBook, isSnapshot)
        wsp.callbacks.OnOrderBook(domainOrderBook, isSnapshot)

FUNCTION (wsp *WSPublicClient) handlePublicTrade(data json.RawMessage):
    // Bybit sends array of trades
    var trades []WSPublicTrade
    err := json.Unmarshal(data, &trades)
    IF err != nil THEN
        LOG error parsing trade
        RETURN

    IF wsp.callbacks.OnTrade != nil THEN
        FOR EACH trade IN trades DO
            domainTrade := mapPublicTrade(&trade)
            wsp.callbacks.OnTrade(domainTrade)

FUNCTION (wsp *WSPrivateClient) handleMessage(data []byte):
    // Try to parse as auth response
    var wsResp WSResponse
    err := json.Unmarshal(data, &wsResp)
    IF err == nil AND wsResp.Op == "auth" THEN
        wsp.handleAuthResponse(&wsResp)
        RETURN

    // Try to parse as response
    IF err == nil AND wsResp.Op != "" THEN
        wsp.handleResponse(&wsResp)
        RETURN

    // Try to parse as pong
    var pong struct {
        Op string `json:"op"`
    }
    err = json.Unmarshal(data, &pong)
    IF err == nil AND pong.Op == "pong" THEN
        RETURN

    // Parse as data message
    var wsMsg WSMessage
    err = json.Unmarshal(data, &wsMsg)
    IF err != nil THEN
        LOG "failed to parse private websocket message" with error
        RETURN

    wsp.handleDataMessage(&wsMsg)

FUNCTION (wsp *WSPrivateClient) handleAuthResponse(resp *WSResponse):
    IF wsp.authResponseChan != nil THEN
        wsp.authResponseChan <- resp.Success
        wsp.authResponseChan = nil

FUNCTION (wsp *WSPrivateClient) handleResponse(resp *WSResponse):
    IF NOT resp.Success THEN
        LOG "private websocket operation failed" with resp.RetMsg
        RETURN

    SWITCH resp.Op:
        CASE "subscribe":
            LOG "subscribed to private channel successfully"
        CASE "unsubscribe":
            LOG "unsubscribed from private channel successfully"

FUNCTION (wsp *WSPrivateClient) handleDataMessage(msg *WSMessage):
    topic := msg.Topic

    IF strings.Contains(topic, "order") THEN
        wsp.handleOrder(msg.Data)
    ELSE IF strings.Contains(topic, "position") THEN
        wsp.handlePosition(msg.Data)
    ELSE IF strings.Contains(topic, "execution") THEN
        wsp.handleExecution(msg.Data)
    ELSE IF strings.Contains(topic, "wallet") THEN
        wsp.handleWallet(msg.Data)
    ELSE
        LOG "unknown private topic" with topic

FUNCTION (wsp *WSPrivateClient) handleOrder(data json.RawMessage):
    // Bybit sends array of orders
    var orders []WSOrder
    err := json.Unmarshal(data, &orders)
    IF err != nil THEN
        LOG error parsing order
        RETURN

    IF wsp.callbacks.OnOrder != nil THEN
        FOR EACH order IN orders DO
            domainOrder := mapWSOrder(&order)
            wsp.callbacks.OnOrder(domainOrder)

FUNCTION (wsp *WSPrivateClient) handlePosition(data json.RawMessage):
    // Position updates
    // Implementation depends on whether using unified or contract account
    LOG "position update received"

FUNCTION (wsp *WSPrivateClient) handleExecution(data json.RawMessage):
    // Execution (fill) updates
    LOG "execution update received"

FUNCTION (wsp *WSPrivateClient) handleWallet(data json.RawMessage):
    // Wallet balance updates
    LOG "wallet update received"
```

**Estimated Time**: 14 hours

---

#### Task 5.7: Bybit - Driver Integration

**Objective**: Integrate all Bybit components into driver interface.

**Files**: `internal/driver/bybit/driver.go` (additions)

**Pseudocode**:

```
// Market Data methods
FUNCTION (bd *BybitDriver) SubscribeTicker(symbol string) error:
    // V5 public ticker topic format
    topic := fmt.Sprintf("tickers.%s", normalizeSymbol(symbol))
    RETURN bd.wsPublic.Subscribe(topic)

FUNCTION (bd *BybitDriver) SubscribeOrderBook(symbol string, depth int) error:
    // V5 orderbook topic format
    // depth: 1, 50, 200, 500
    validDepth := 50
    IF depth >= 500 THEN
        validDepth = 500
    ELSE IF depth >= 200 THEN
        validDepth = 200
    ELSE IF depth >= 50 THEN
        validDepth = 50
    ELSE IF depth >= 1 THEN
        validDepth = 1

    topic := fmt.Sprintf("orderbook.%d.%s", validDepth, normalizeSymbol(symbol))

    // Fetch initial snapshot
    GO FUNC():
        orderBook, err := bd.rest.GetOrderBook(symbol, validDepth)
        IF err == nil AND bd.callbacks.OnOrderBook != nil THEN
            bd.callbacks.OnOrderBook(orderBook, true)

    RETURN bd.wsPublic.Subscribe(topic)

FUNCTION (bd *BybitDriver) SubscribeTrades(symbol string) error:
    topic := fmt.Sprintf("publicTrade.%s", normalizeSymbol(symbol))
    RETURN bd.wsPublic.Subscribe(topic)

FUNCTION (bd *BybitDriver) Unsubscribe(symbol string) error:
    normalized := normalizeSymbol(symbol)

    // Unsubscribe from all topics for this symbol
    topics := []string{
        fmt.Sprintf("tickers.%s", normalized),
        fmt.Sprintf("orderbook.50.%s", normalized),
        fmt.Sprintf("publicTrade.%s", normalized),
    }

    FOR EACH topic IN topics DO
        bd.wsPublic.Unsubscribe(topic)

    RETURN nil

// Trading methods
FUNCTION (bd *BybitDriver) PlaceOrder(req *domain.OrderRequest) (*domain.Order, error):
    IF NOT bd.started.Load() THEN
        RETURN nil, errors.ErrNotRunning

    RETURN bd.rest.PlaceOrder(req)

FUNCTION (bd *BybitDriver) CancelOrder(req *domain.CancelRequest) error:
    IF NOT bd.started.Load() THEN
        RETURN errors.ErrNotRunning

    RETURN bd.rest.CancelOrder(req)

FUNCTION (bd *BybitDriver) GetOrder(orderID string) (*domain.Order, error):
    IF NOT bd.started.Load() THEN
        RETURN nil, errors.ErrNotRunning

    // Bybit V5 GetOrder also needs symbol
    // This is a limitation - might need to cache order details
    RETURN nil, errors.New("GetOrder requires symbol in Bybit")

FUNCTION (bd *BybitDriver) GetOpenOrders(symbol string) ([]*domain.Order, error):
    IF NOT bd.started.Load() THEN
        RETURN nil, errors.ErrNotRunning

    RETURN bd.rest.GetOpenOrders(symbol)

// Account methods
FUNCTION (bd *BybitDriver) GetBalance(asset string) (*domain.Balance, error):
    IF NOT bd.started.Load() THEN
        RETURN nil, errors.ErrNotRunning

    RETURN bd.rest.GetBalance(asset)

FUNCTION (bd *BybitDriver) GetPosition(symbol string) (*domain.Position, error):
    // V5 unified account supports positions
    // For now, return not supported for spot
    RETURN nil, errors.New("positions not supported for spot trading")

FUNCTION (bd *BybitDriver) GetSymbols() ([]*domain.SymbolInfo, error):
    IF NOT bd.started.Load() THEN
        RETURN nil, errors.ErrNotRunning

    RETURN bd.rest.GetInstrumentsInfo("spot", "")

// Subscribe to private channels (optional - called from Start if keys present)
FUNCTION (bd *BybitDriver) subscribePrivateChannels() error:
    IF bd.wsPrivate == nil OR NOT bd.wsPrivate.IsConnected() THEN
        RETURN nil

    // Subscribe to order updates
    err := bd.wsPrivate.Subscribe("order")
    IF err != nil THEN
        RETURN err

    // Subscribe to execution updates
    err = bd.wsPrivate.Subscribe("execution")
    IF err != nil THEN
        RETURN err

    // Subscribe to wallet updates
    err = bd.wsPrivate.Subscribe("wallet")
    IF err != nil THEN
        RETURN err

    RETURN nil
```

**Estimated Time**: 4 hours

---

#### Task 5.8: Bybit - Error Handling

**Objective**: Implement Bybit-specific error handling.

**Files**: `internal/driver/bybit/errors.go`

**Pseudocode**:

```
// Bybit error codes (V5)
CONST (
    ErrCodeSuccess = 0
    ErrCodeInvalidRequest = 10001
    ErrCodeInvalidAPIKey = 10003
    ErrCodeInvalidSign = 10004
    ErrCodeInvalidTimestamp = 10005
    ErrCodeInvalidRecvWindow = 10006
    ErrCodeRateLimitExceeded = 10006
    ErrCodeIPNotWhitelisted = 10010
    ErrCodeInvalidParameter = 10016
    ErrCodeSystemError = 10018
    ErrCodeOrderNotExists = 110001
    ErrCodeOrderCanceled = 110004
    ErrCodeOrderPriceOutOfRange = 110007
    ErrCodeInsufficientBalance = 110017
    ErrCodeOrderQtyExceedsLimit = 110043
    ErrCodeDuplicateClientOrderID = 110056
)

FUNCTION MapBybitError(code int, message string) error:
    baseErr := errors.NewExchangeError("bybit", code, message)

    SWITCH code:
        CASE ErrCodeSuccess:
            RETURN nil

        CASE ErrCodeRateLimitExceeded:
            RETURN errors.NewRateLimitError("bybit", 0, 60)

        CASE ErrCodeInvalidTimestamp:
            RETURN fmt.Errorf("%w: clock sync issue - %s", baseErr, message)

        CASE ErrCodeInvalidSign:
            RETURN fmt.Errorf("%w: signature verification failed - %s", baseErr, message)

        CASE ErrCodeOrderNotExists:
            RETURN fmt.Errorf("%w: %s", errors.ErrOrderNotFound, message)

        CASE ErrCodeInvalidAPIKey, ErrCodeIPNotWhitelisted:
            RETURN fmt.Errorf("%w: authentication failed - %s", baseErr, message)

        CASE ErrCodeInsufficientBalance:
            RETURN fmt.Errorf("%w: insufficient balance - %s", baseErr, message)

        CASE ErrCodeDuplicateClientOrderID:
            RETURN fmt.Errorf("%w: duplicate client order ID - %s", baseErr, message)

        CASE ErrCodeSystemError:
            RETURN errors.NewConnectionError("bybit", message, true)

        DEFAULT:
            RETURN baseErr

FUNCTION IsRetryable(err error) bool:
    var bybitErr *errors.ExchangeError
    IF errors.As(err, &bybitErr) THEN
        SWITCH bybitErr.Code:
            CASE ErrCodeSystemError:
                RETURN true

    var connErr *errors.ConnectionError
    IF errors.As(err, &connErr) THEN
        RETURN connErr.Temporary

    RETURN false

FUNCTION IsRateLimitError(err error) bool:
    var rateLimitErr *errors.RateLimitError
    RETURN errors.As(err, &rateLimitErr)
```

**Estimated Time**: 3 hours

---

### Phase 6: Core Actors & Public API (Days 24-30)

---

#### Task 6.1: Exchange Supervisor

**Objective**: Create supervisor for managing exchange-specific actors.

**Files**: `internal/sup/exchange.go`

**Pseudocode**:

```
TYPE ExchangeSupervisor STRUCT:
    gen.Supervisor
    name string
    driver driver.Driver
    config *config.ExchangeConfig
    rateLimitPID gen.PID
    clockSyncPID gen.PID
    marketDataPID gen.PID
    orderCommandPID gen.PID
    orderStatePID gen.PID
    orderBookPID gen.PID
    positionTrackerPID gen.PID

FUNCTION (es *ExchangeSupervisor) Init(args ...interface{}) (gen.SupervisorSpec, error):
    PARSE name, driver, config FROM args

    es.name = name
    es.driver = driver
    es.config = config

    // Create supervisor spec
    spec := gen.SupervisorSpec{
        Name: fmt.Sprintf("ExchangeSupervisor_%s", name),
        Strategy: gen.SupervisorStrategy{
            Type: gen.SupervisorStrategyOneForOne,
            Intensity: 5,
            Period: 10,
            Restart: gen.SupervisorStrategyRestartTransient,
        },
    }

    // Define child specifications
    spec.Children = []gen.SupervisorChildSpec{
        // Clock sync actor
        {
            Name: fmt.Sprintf("ClockSync_%s", name),
            Child: &ClockSyncActor{},
            Args: []interface{}{name, driver},
        },

        // Rate limit guard
        {
            Name: fmt.Sprintf("RateLimitGuard_%s", name),
            Child: &RateLimitGuardActor{},
            Args: []interface{}{name, config.RateLimit},
        },

        // Market data actor (WebSocket)
        {
            Name: fmt.Sprintf("MarketData_%s", name),
            Child: &MarketDataActor{},
            Args: []interface{}{name, driver},
        },

        // Order book actor
        {
            Name: fmt.Sprintf("OrderBook_%s", name),
            Child: &OrderBookActor{},
            Args: []interface{}{name},
        },

        // Order command actor (write path)
        {
            Name: fmt.Sprintf("OrderCommand_%s", name),
            Child: &OrderCommandActor{},
            Args: []interface{}{name, driver},
        },

        // Order state actor (read path)
        {
            Name: fmt.Sprintf("OrderState_%s", name),
            Child: &OrderStateActor{},
            Args: []interface{}{name},
        },

        // Position tracker actor
        {
            Name: fmt.Sprintf("PositionTracker_%s", name),
            Child: &PositionTrackerActor{},
            Args: []interface{}{name},
        },
    }

    RETURN spec, nil

FUNCTION (es *ExchangeSupervisor) HandleChildStarted(name string, pid gen.PID):
    // Store PIDs for later reference
    SWITCH:
        CASE strings.Contains(name, "ClockSync"):
            es.clockSyncPID = pid
        CASE strings.Contains(name, "RateLimitGuard"):
            es.rateLimitPID = pid
        CASE strings.Contains(name, "MarketData"):
            es.marketDataPID = pid
        CASE strings.Contains(name, "OrderCommand"):
            es.orderCommandPID = pid
        CASE strings.Contains(name, "OrderState"):
            es.orderStatePID = pid
        CASE strings.Contains(name, "OrderBook"):
            es.orderBookPID = pid
        CASE strings.Contains(name, "PositionTracker"):
            es.positionTrackerPID = pid

FUNCTION (es *ExchangeSupervisor) HandleChildTerminated(name string, pid gen.PID, reason error):
    LOG "child actor terminated" with name, reason

    // Publish event about actor failure
    EMIT event about terminated child
```

**Estimated Time**: 5 hours

---

#### Task 6.2: Persistence Supervisor

**Objective**: Create supervisor for persistence actors.

**Files**: `internal/sup/persistence.go`

**Pseudocode**:

```
TYPE PersistenceSupervisor STRUCT:
    gen.Supervisor
    backend persistence.PersistenceBackend
    orderStorePID gen.PID
    positionStorePID gen.PID
    balanceStorePID gen.PID
    reconcilerPID gen.PID

FUNCTION (ps *PersistenceSupervisor) Init(args ...interface{}) (gen.SupervisorSpec, error):
    PARSE backend FROM args

    ps.backend = backend

    spec := gen.SupervisorSpec{
        Name: "PersistenceSupervisor",
        Strategy: gen.SupervisorStrategy{
            Type: gen.SupervisorStrategyOneForOne,
            Intensity: 3,
            Period: 10,
            Restart: gen.SupervisorStrategyRestartPermanent,
        },
    }

    spec.Children = []gen.SupervisorChildSpec{
        {
            Name: "OrderStore",
            Child: &OrderStoreActor{},
            Args: []interface{}{backend.OrderStore()},
        },
        {
            Name: "PositionStore",
            Child: &PositionStoreActor{},
            Args: []interface{}{backend.PositionStore()},
        },
        {
            Name: "BalanceStore",
            Child: &BalanceStoreActor{},
            Args: []interface{}{backend.BalanceStore()},
        },
        {
            Name: "Reconciler",
            Child: &ReconcilerActor{},
            Args: []interface{}{backend},
        },
    }

    RETURN spec, nil

FUNCTION (ps *PersistenceSupervisor) HandleChildStarted(name string, pid gen.PID):
    SWITCH name:
        CASE "OrderStore":
            ps.orderStorePID = pid
        CASE "PositionStore":
            ps.positionStorePID = pid
        CASE "BalanceStore":
            ps.balanceStorePID = pid
        CASE "Reconciler":
            ps.reconcilerPID = pid
```

**Estimated Time**: 3 hours

---

#### Task 6.3: Root Supervisor

**Objective**: Create top-level supervisor managing all components.

**Files**: `internal/sup/root.go`

**Pseudocode**:

```
TYPE RootSupervisor STRUCT:
    gen.Supervisor
    config *config.Config
    exchanges sync.Map // name -> ExchangeSupervisorPID
    coordinatorPID gen.PID
    healthCheckPID gen.PID
    aggregatorPID gen.PID
    riskManagerPID gen.PID
    metricsCollectorPID gen.PID
    persistenceSupervisorPID gen.PID

FUNCTION (rs *RootSupervisor) Init(args ...interface{}) (gen.SupervisorSpec, error):
    PARSE config FROM args

    rs.config = config

    spec := gen.SupervisorSpec{
        Name: "RootSupervisor",
        Strategy: gen.SupervisorStrategy{
            Type: gen.SupervisorStrategyOneForOne,
            Intensity: 10,
            Period: 60,
            Restart: gen.SupervisorStrategyRestartPermanent,
        },
    }

    // Define core children
    spec.Children = []gen.SupervisorChildSpec{
        // Persistence supervisor (starts first)
        {
            Name: "PersistenceSupervisor",
            Child: &PersistenceSupervisor{},
            Args: []interface{}{config.Persistence.Backend},
        },

        // Coordinator actor (message routing)
        {
            Name: "Coordinator",
            Child: &CoordinatorActor{},
            Args: []interface{}{},
        },

        // Health check actor (circuit breaker)
        {
            Name: "HealthCheck",
            Child: &HealthCheckActor{},
            Args: []interface{}{},
        },

        // Metrics collector actor
        {
            Name: "MetricsCollector",
            Child: &MetricsCollectorActor{},
            Args: []interface{}{},
        },

        // Risk manager actor
        {
            Name: "RiskManager",
            Child: risk.RiskManager{},
            Args: []interface{}{config.Risk},
        },

        // Aggregator actor (if enabled)
        {
            Name: "Aggregator",
            Child: &AggregatorActor{},
            Args: []interface{}{config.Aggregation},
            Enabled: config.Aggregation.Enabled,
        },
    }

    RETURN spec, nil

FUNCTION (rs *RootSupervisor) HandleChildStarted(name string, pid gen.PID):
    SWITCH name:
        CASE "Coordinator":
            rs.coordinatorPID = pid
        CASE "HealthCheck":
            rs.healthCheckPID = pid
        CASE "Aggregator":
            rs.aggregatorPID = pid
        CASE "RiskManager":
            rs.riskManagerPID = pid
        CASE "MetricsCollector":
            rs.metricsCollectorPID = pid
        CASE "PersistenceSupervisor":
            rs.persistenceSupervisorPID = pid

FUNCTION (rs *RootSupervisor) AddExchange(name string, exchangeConfig *config.ExchangeConfig) error:
    // Get driver from registry
    driver, err := driver.Get(exchangeConfig.Type, exchangeConfig)
    IF err != nil THEN
        RETURN err

    // Start exchange supervisor
    pid, err := rs.StartChild(gen.SupervisorChildSpec{
        Name: fmt.Sprintf("Exchange_%s", name),
        Child: &ExchangeSupervisor{},
        Args: []interface{}{name, driver, exchangeConfig},
    })

    IF err != nil THEN
        RETURN err

    rs.exchanges.Store(name, pid)

    // Notify coordinator about new exchange
    gen.Cast(rs.coordinatorPID, &message.RegisterExchange{
        Name: name,
        Driver: driver,
        Config: exchangeConfig,
    })

    RETURN nil

FUNCTION (rs *RootSupervisor) RemoveExchange(name string) error:
    pidInterface, ok := rs.exchanges.Load(name)
    IF NOT ok THEN
        RETURN errors.ErrExchangeNotFound

    pid := pidInterface.(gen.PID)

    // Stop exchange supervisor
    err := rs.TerminateChild(pid)
    IF err != nil THEN
        RETURN err

    rs.exchanges.Delete(name)

    // Notify coordinator
    gen.Cast(rs.coordinatorPID, &message.UnregisterExchange{
        Name: name,
    })

    RETURN nil

FUNCTION (rs *RootSupervisor) GetExchanges() []string:
    names := []string{}
    rs.exchanges.Range(FUNC(key, value interface{}):
        names = append(names, key.(string))
        RETURN true
    )
    RETURN names
```

**Estimated Time**: 6 hours

---

#### Task 6.4: Coordinator Actor

**Objective**: Central coordinator for message routing.

**Files**: `internal/actor/coordinator.go`

**Pseudocode**:

```
TYPE CoordinatorActor STRUCT:
    act.Actor
    exchanges sync.Map // name -> ExchangeInfo
    subscriptions sync.Map // topic -> []gen.PID

TYPE ExchangeInfo STRUCT:
    name string
    driver driver.Driver
    supervisorPID gen.PID
    marketDataPID gen.PID
    orderCommandPID gen.PID
    orderStatePID gen.PID
    positionTrackerPID gen.PID

FUNCTION (ca *CoordinatorActor) Init(args ...interface{}) error:
    LOG "Coordinator actor initialized"
    RETURN nil

FUNCTION (ca *CoordinatorActor) HandleCast(from gen.PID, message interface{}):
    SWITCH message TYPE:
        CASE *message.RegisterExchange:
            CALL ca.registerExchange(message)

        CASE *message.UnregisterExchange:
            CALL ca.unregisterExchange(message)

        CASE *message.ExchangeActorsReady:
            CALL ca.updateExchangeActors(message)

FUNCTION (ca *CoordinatorActor) HandleCall(from gen.PID, ref gen.Ref, request interface{}) (interface{}, error):
    SWITCH request TYPE:
        CASE *message.SubscribeTicker:
            RETURN ca.subscribeTicker(request)

        CASE *message.SubscribeOrderBook:
            RETURN ca.subscribeOrderBook(request)

        CASE *message.SubscribeTrades:
            RETURN ca.subscribeTrades(request)

        CASE *message.PlaceOrder:
            RETURN ca.placeOrder(request)

        CASE *message.CancelOrder:
            RETURN ca.cancelOrder(request)

        CASE *message.GetOrder:
            RETURN ca.getOrder(request)

        CASE *message.GetOpenOrders:
            RETURN ca.getOpenOrders(request)

        CASE *message.GetBalance:
            RETURN ca.getBalance(request)

        CASE *message.GetPosition:
            RETURN ca.getPosition(request)

        DEFAULT:
            RETURN nil, errors.New("unknown request type")

FUNCTION (ca *CoordinatorActor) registerExchange(msg *message.RegisterExchange):
    info := ExchangeInfo{
        name: msg.Name,
        driver: msg.Driver,
    }

    ca.exchanges.Store(msg.Name, &info)
    LOG "exchange registered" with msg.Name

FUNCTION (ca *CoordinatorActor) unregisterExchange(msg *message.UnregisterExchange):
    ca.exchanges.Delete(msg.Name)
    LOG "exchange unregistered" with msg.Name

FUNCTION (ca *CoordinatorActor) updateExchangeActors(msg *message.ExchangeActorsReady):
    infoInterface, ok := ca.exchanges.Load(msg.Exchange)
    IF NOT ok THEN
        RETURN

    info := infoInterface.(*ExchangeInfo)
    info.marketDataPID = msg.MarketDataPID
    info.orderCommandPID = msg.OrderCommandPID
    info.orderStatePID = msg.OrderStatePID
    info.positionTrackerPID = msg.PositionTrackerPID

    ca.exchanges.Store(msg.Exchange, info)

FUNCTION (ca *CoordinatorActor) subscribeTicker(req *message.SubscribeTicker) error:
    info, err := ca.getExchangeInfo(req.Exchange)
    IF err != nil THEN
        RETURN err

    // Forward to driver
    err = info.driver.SubscribeTicker(req.Symbol)
    IF err != nil THEN
        RETURN err

    RETURN nil

FUNCTION (ca *CoordinatorActor) subscribeOrderBook(req *message.SubscribeOrderBook) error:
    info, err := ca.getExchangeInfo(req.Exchange)
    IF err != nil THEN
        RETURN err

    err = info.driver.SubscribeOrderBook(req.Symbol, req.Depth)
    IF err != nil THEN
        RETURN err

    RETURN nil

FUNCTION (ca *CoordinatorActor) subscribeTrades(req *message.SubscribeTrades) error:
    info, err := ca.getExchangeInfo(req.Exchange)
    IF err != nil THEN
        RETURN err

    err = info.driver.SubscribeTrades(req.Symbol)
    IF err != nil THEN
        RETURN err

    RETURN nil

FUNCTION (ca *CoordinatorActor) placeOrder(req *message.PlaceOrder) (*message.OrderResult, error):
    info, err := ca.getExchangeInfo(req.Request.Exchange)
    IF err != nil THEN
        RETURN nil, err

    // Forward to order command actor
    result, err := gen.Call(info.orderCommandPID, req)
    IF err != nil THEN
        RETURN nil, err

    RETURN result.(*message.OrderResult), nil

FUNCTION (ca *CoordinatorActor) cancelOrder(req *message.CancelOrder) error:
    info, err := ca.getExchangeInfo(req.Request.Exchange)
    IF err != nil THEN
        RETURN err

    // Forward to order command actor
    _, err = gen.Call(info.orderCommandPID, req)
    RETURN err

FUNCTION (ca *CoordinatorActor) getOrder(req *message.GetOrder) (*domain.Order, error):
    info, err := ca.getExchangeInfo(req.Exchange)
    IF err != nil THEN
        RETURN nil, err

    // Forward to order state actor
    result, err := gen.Call(info.orderStatePID, req)
    IF err != nil THEN
        RETURN nil, err

    RETURN result.(*domain.Order), nil

FUNCTION (ca *CoordinatorActor) getOpenOrders(req *message.GetOpenOrders) ([]*domain.Order, error):
    info, err := ca.getExchangeInfo(req.Exchange)
    IF err != nil THEN
        RETURN nil, err

    // Forward to order state actor
    result, err := gen.Call(info.orderStatePID, req)
    IF err != nil THEN
        RETURN nil, err

    RETURN result.([]*domain.Order), nil

FUNCTION (ca *CoordinatorActor) getBalance(req *message.GetBalance) (*domain.Balance, error):
    info, err := ca.getExchangeInfo(req.Exchange)
    IF err != nil THEN
        RETURN nil, err

    // Call driver directly
    RETURN info.driver.GetBalance(req.Asset)

FUNCTION (ca *CoordinatorActor) getPosition(req *message.GetPosition) (*domain.Position, error):
    info, err := ca.getExchangeInfo(req.Exchange)
    IF err != nil THEN
        RETURN nil, err

    // Forward to position tracker actor
    result, err := gen.Call(info.positionTrackerPID, req)
    IF err != nil THEN
        RETURN nil, err

    RETURN result.(*domain.Position), nil

FUNCTION (ca *CoordinatorActor) getExchangeInfo(name string) (*ExchangeInfo, error):
    infoInterface, ok := ca.exchanges.Load(name)
    IF NOT ok THEN
        RETURN nil, errors.ErrExchangeNotFound

    RETURN infoInterface.(*ExchangeInfo), nil
```

**Estimated Time**: 7 hours

---

#### Task 6.5: Health Check Actor

**Objective**: Implement health monitoring and circuit breaker.

**Files**: `internal/actor/health.go`

**Pseudocode**:

```
TYPE HealthCheckActor STRUCT:
    act.Actor
    exchanges sync.Map // name -> *ExchangeHealth
    checkInterval time.Duration

TYPE ExchangeHealth STRUCT:
    name string
    status domain.HealthStatus
    circuitBreaker *circuit.CircuitBreaker
    lastCheck time.Time
    consecutiveFailures int
    consecutiveSuccesses int
    mutex sync.RWMutex

FUNCTION (hc *HealthCheckActor) Init(args ...interface{}) error:
    hc.checkInterval = 30 * time.Second

    // Schedule periodic health checks
    gen.CronSendAfter(hc.Self(), hc.checkInterval, &message.PerformHealthChecks{})

    LOG "Health check actor initialized"
    RETURN nil

FUNCTION (hc *HealthCheckActor) HandleCast(from gen.PID, message interface{}):
    SWITCH message TYPE:
        CASE *message.RegisterExchangeHealth:
            CALL hc.registerExchange(message)

        CASE *message.PerformHealthChecks:
            CALL hc.performHealthChecks()
            // Schedule next check
            gen.CronSendAfter(hc.Self(), hc.checkInterval, &message.PerformHealthChecks{})

        CASE *message.RecordSuccess:
            CALL hc.recordSuccess(message.Exchange)

        CASE *message.RecordFailure:
            CALL hc.recordFailure(message.Exchange, message.Error)

FUNCTION (hc *HealthCheckActor) HandleCall(from gen.PID, ref gen.Ref, request interface{}) (interface{}, error):
    SWITCH request TYPE:
        CASE *message.GetHealthStatus:
            RETURN hc.getHealthStatus(request.Exchange)

        CASE *message.CheckHealth:
            RETURN hc.checkExchangeHealth(request.Exchange)

        DEFAULT:
            RETURN nil, errors.New("unknown request type")

FUNCTION (hc *HealthCheckActor) registerExchange(msg *message.RegisterExchangeHealth):
    cbConfig := circuit.Config{
        MaxFailures: 5,
        Timeout: 60 * time.Second,
        HalfOpenRequests: 3,
    }

    health := &ExchangeHealth{
        name: msg.Exchange,
        status: domain.HealthStatus{
            Exchange: msg.Exchange,
            Status: domain.StatusDisconnected,
            CircuitState: domain.CircuitClosed,
            Healthy: false,
            LastCheck: time.Now(),
        },
        circuitBreaker: circuit.NewCircuitBreaker(cbConfig),
    }

    hc.exchanges.Store(msg.Exchange, health)
    LOG "registered exchange for health monitoring" with msg.Exchange

FUNCTION (hc *HealthCheckActor) performHealthChecks():
    hc.exchanges.Range(FUNC(key, value interface{}):
        name := key.(string)
        GO hc.checkExchangeHealth(name)
        RETURN true
    )

FUNCTION (hc *HealthCheckActor) checkExchangeHealth(name string) (*domain.HealthStatus, error):
    healthInterface, ok := hc.exchanges.Load(name)
    IF NOT ok THEN
        RETURN nil, errors.ErrExchangeNotFound

    health := healthInterface.(*ExchangeHealth)

    health.mutex.Lock()

    // Update last check time
    health.lastCheck = time.Now()
    startTime := time.Now()

    // Perform health check through circuit breaker
    err := health.circuitBreaker.Call(FUNC() error:
        // Simple ping check - could be enhanced
        // For now, just check if exchange is connected
        RETURN nil
    )

    latency := time.Since(startTime)

    IF err != nil THEN
        health.consecutiveFailures++
        health.consecutiveSuccesses = 0
        health.status.Healthy = false
        health.status.ErrorCount++
        health.status.Message = err.Error()

        IF errors.Is(err, errors.ErrCircuitOpen) THEN
            health.status.CircuitState = domain.CircuitOpen

    ELSE
        health.consecutiveSuccesses++
        health.consecutiveFailures = 0
        health.status.Healthy = true
        health.status.SuccessCount++
        health.status.Latency = latency
        health.status.Message = "healthy"

        // Update circuit state from breaker
        cbState := health.circuitBreaker.State()
        SWITCH cbState:
            CASE circuit.StateClosed:
                health.status.CircuitState = domain.CircuitClosed
            CASE circuit.StateOpen:
                health.status.CircuitState = domain.CircuitOpen
            CASE circuit.StateHalfOpen:
                health.status.CircuitState = domain.CircuitHalfOpen

    health.status.LastCheck = time.Now()
    status := health.status

    health.mutex.Unlock()

    // Publish health status event
    EMIT health.status event with name and status

    RETURN &status, nil

FUNCTION (hc *HealthCheckActor) recordSuccess(exchange string):
    healthInterface, ok := hc.exchanges.Load(exchange)
    IF NOT ok THEN
        RETURN

    health := healthInterface.(*ExchangeHealth)
    health.mutex.Lock()

    health.consecutiveSuccesses++
    health.consecutiveFailures = 0
    health.status.SuccessCount++

    health.mutex.Unlock()

FUNCTION (hc *HealthCheckActor) recordFailure(exchange string, err error):
    healthInterface, ok := hc.exchanges.Load(exchange)
    IF NOT ok THEN
        RETURN

    health := healthInterface.(*ExchangeHealth)
    health.mutex.Lock()

    health.consecutiveFailures++
    health.consecutiveSuccesses = 0
    health.status.ErrorCount++

    // If too many consecutive failures, mark as unhealthy
    IF health.consecutiveFailures >= 3 THEN
        health.status.Healthy = false
        health.status.Message = fmt.Sprintf("consecutive failures: %d", health.consecutiveFailures)

    health.mutex.Unlock()

    // Publish failure event
    EMIT health.failure event

FUNCTION (hc *HealthCheckActor) getHealthStatus(exchange string) (*domain.HealthStatus, error):
    healthInterface, ok := hc.exchanges.Load(exchange)
    IF NOT ok THEN
        RETURN nil, errors.ErrExchangeNotFound

    health := healthInterface.(*ExchangeHealth)
    health.mutex.RLock()
    status := health.status
    health.mutex.RUnlock()

    RETURN &status, nil
```

**Estimated Time**: 6 hours

---

#### Task 6.6: Aggregator Actor

**Objective**: Implement cross-exchange aggregation.

**Files**: `internal/actor/aggregator.go`

**Pseudocode**:

```
TYPE AggregatorActor STRUCT:
    act.Actor
    config config.AggregationConfig
    tickers sync.Map // symbol -> map[exchange]*domain.Ticker
    orderBooks sync.Map // symbol -> map[exchange]*domain.OrderBook
    updateInterval time.Duration
    enabled bool

FUNCTION (aa *AggregatorActor) Init(args ...interface{}) error:
    PARSE config FROM args

    aa.config = config
    aa.enabled = config.Enabled
    aa.updateInterval = config.UpdateInterval

    IF NOT aa.enabled THEN
        LOG "aggregator disabled"
        RETURN nil

    // Schedule periodic aggregation
    gen.CronSendAfter(aa.Self(), aa.updateInterval, &message.AggregateData{})

    // Subscribe to market data events
    FOR EACH symbol IN config.Symbols DO
        topic := event.TopicFor(event.TopicTicker, "*", symbol)
        gen.SubscribeEvent(aa.Self(), topic)

        topic = event.TopicFor(event.TopicOrderBook, "*", symbol)
        gen.SubscribeEvent(aa.Self(), topic)

    LOG "Aggregator actor initialized" with config.Symbols
    RETURN nil

FUNCTION (aa *AggregatorActor) HandleEvent(message gen.MessageEvent):
    IF NOT aa.enabled THEN
        RETURN

    SWITCH message.Message TYPE:
        CASE *domain.Ticker:
            CALL aa.handleTicker(message.Message)

        CASE *domain.OrderBook:
            CALL aa.handleOrderBook(message.Message)

FUNCTION (aa *AggregatorActor) HandleCast(from gen.PID, message interface{}):
    SWITCH message TYPE:
        CASE *message.AggregateData:
            CALL aa.aggregateAll()
            // Schedule next aggregation
            gen.CronSendAfter(aa.Self(), aa.updateInterval, &message.AggregateData{})

FUNCTION (aa *AggregatorActor) HandleCall(from gen.PID, ref gen.Ref, request interface{}) (interface{}, error):
    SWITCH request TYPE:
        CASE *message.GetAggregatedTicker:
            RETURN aa.getAggregatedTicker(request.Symbol)

        CASE *message.GetAggregatedOrderBook:
            RETURN aa.getAggregatedOrderBook(request.Symbol)

        CASE *message.FindArbitrage:
            RETURN aa.findArbitrage(request.Symbol)

        DEFAULT:
            RETURN nil, errors.New("unknown request type")

FUNCTION (aa *AggregatorActor) handleTicker(ticker *domain.Ticker):
    // Get or create ticker map for symbol
    tickersInterface, _ := aa.tickers.LoadOrStore(ticker.Symbol, &sync.Map{})
    tickersMap := tickersInterface.(*sync.Map)

    // Store ticker
    tickersMap.Store(ticker.Exchange, ticker)

FUNCTION (aa *AggregatorActor) handleOrderBook(orderBook *domain.OrderBook):
    // Get or create order book map for symbol
    booksInterface, _ := aa.orderBooks.LoadOrStore(orderBook.Symbol, &sync.Map{})
    booksMap := booksInterface.(*sync.Map)

    // Store order book
    booksMap.Store(orderBook.Exchange, orderBook)

FUNCTION (aa *AggregatorActor) aggregateAll():
    // Aggregate tickers
    aa.tickers.Range(FUNC(key, value interface{}):
        symbol := key.(string)
        aggregated := aa.aggregateTickers(symbol)
        IF aggregated != nil THEN
            // Publish aggregated ticker
            EMIT aggregated.ticker event
        RETURN true
    )

    // Aggregate order books
    aa.orderBooks.Range(FUNC(key, value interface{}):
        symbol := key.(string)
        aggregated := aa.aggregateOrderBooks(symbol)
        IF aggregated != nil THEN
            // Publish aggregated order book
            EMIT aggregated.orderbook event
        RETURN true
    )

    // Find arbitrage opportunities
    FOR EACH symbol IN aa.config.Symbols DO
        opportunities := aa.findArbitrage(symbol)
        FOR EACH opp IN opportunities DO
            IF opp.Valid AND opp.SpreadPercent > 0.1 THEN // 0.1% minimum
                EMIT arbitrage.opportunity event with opp

FUNCTION (aa *AggregatorActor) aggregateTickers(symbol string) *domain.AggregatedTicker:
    tickersInterface, ok := aa.tickers.Load(symbol)
    IF NOT ok THEN
        RETURN nil

    tickersMap := tickersInterface.(*sync.Map)

    aggregated := &domain.AggregatedTicker{
        Symbol: symbol,
        BestBid: domain.Zero(),
        BestAsk: domain.NewDecimalFromFloat64(math.MaxFloat64),
        Timestamp: time.Now(),
    }

    exchanges := []string{}

    tickersMap.Range(FUNC(key, value interface{}):
        exchange := key.(string)
        ticker := value.(*domain.Ticker)

        exchanges = append(exchanges, exchange)

        // Find best bid (highest)
        IF domain.Compare(ticker.Bid, aggregated.BestBid) > 0 THEN
            aggregated.BestBid = ticker.Bid
            aggregated.BestBidExchange = exchange

        // Find best ask (lowest)
        IF domain.Compare(ticker.Ask, aggregated.BestAsk) < 0 THEN
            aggregated.BestAsk = ticker.Ask
            aggregated.BestAskExchange = exchange

        RETURN true
    )

    aggregated.Exchanges = exchanges
    aggregated.Spread = domain.Sub(aggregated.BestAsk, aggregated.BestBid)

    RETURN aggregated

FUNCTION (aa *AggregatorActor) aggregateOrderBooks(symbol string) *domain.AggregatedOrderBook:
    booksInterface, ok := aa.orderBooks.Load(symbol)
    IF NOT ok THEN
        RETURN nil

    booksMap := booksInterface.(*sync.Map)

    aggregated := &domain.AggregatedOrderBook{
        Symbol: symbol,
        Timestamp: time.Now(),
    }

    // Collect all bids and asks
    allBids := []*domain.AggregatedLevel{}
    allAsks := []*domain.AggregatedLevel{}

    booksMap.Range(FUNC(key, value interface{}):
        exchange := key.(string)
        book := value.(*domain.OrderBook)

        // Process bids
        FOR EACH bid IN book.Bids DO
            // Find existing level or create new
            level := aa.findOrCreateLevel(allBids, bid.Price)
            level.TotalQuantity = domain.Add(level.TotalQuantity, bid.Quantity)
            level.Sources = append(level.Sources, domain.BookSource{
                Exchange: exchange,
                Price: bid.Price,
                Quantity: bid.Quantity,
            })

        // Process asks
        FOR EACH ask IN book.Asks DO
            level := aa.findOrCreateLevel(allAsks, ask.Price)
            level.TotalQuantity = domain.Add(level.TotalQuantity, ask.Quantity)
            level.Sources = append(level.Sources, domain.BookSource{
                Exchange: exchange,
                Price: ask.Price,
                Quantity: ask.Quantity,
            })

        RETURN true
    )

    // Sort bids (descending) and asks (ascending)
    sort.Slice(allBids, FUNC(i, j int) bool:
        RETURN domain.Compare(allBids[i].Price, allBids[j].Price) > 0
    )

    sort.Slice(allAsks, FUNC(i, j int) bool:
        RETURN domain.Compare(allAsks[i].Price, allAsks[j].Price) < 0
    )

    aggregated.Bids = allBids
    aggregated.Asks = allAsks

    RETURN aggregated

FUNCTION (aa *AggregatorActor) findOrCreateLevel(levels []*domain.AggregatedLevel, price domain.Decimal) *domain.AggregatedLevel:
    FOR EACH level IN levels DO
        IF domain.Compare(level.Price, price) == 0 THEN
            RETURN level

    newLevel := &domain.AggregatedLevel{
        Price: price,
        TotalQuantity: domain.Zero(),
        Sources: []domain.BookSource{},
    }
    levels = append(levels, newLevel)
    RETURN newLevel

FUNCTION (aa *AggregatorActor) findArbitrage(symbol string) []*domain.ArbitrageOpportunity:
    aggregated := aa.aggregateTickers(symbol)
    IF aggregated == nil THEN
        RETURN nil

    opportunities := []*domain.ArbitrageOpportunity{}

    // Simple arbitrage: buy on one exchange, sell on another
    IF aggregated.BestBidExchange != "" AND aggregated.BestAskExchange != "" AND
       aggregated.BestBidExchange != aggregated.BestAskExchange THEN

        spread := domain.Sub(aggregated.BestBid, aggregated.BestAsk)

        IF domain.Compare(spread, domain.Zero()) > 0 THEN
            spreadPercent, _ := domain.Float64(
                domain.Mul(
                    domain.Div(spread, aggregated.BestAsk),
                    domain.NewDecimalFromInt64(100),
                ),
            )

            opp := &domain.ArbitrageOpportunity{
                Symbol: symbol,
                BuyExchange: aggregated.BestAskExchange,
                SellExchange: aggregated.BestBidExchange,
                BuyPrice: aggregated.BestAsk,
                SellPrice: aggregated.BestBid,
                Spread: spread,
                SpreadPercent: spreadPercent,
                Timestamp: time.Now(),
                Valid: true,
            }

            opportunities = append(opportunities, opp)

    RETURN opportunities

FUNCTION (aa *AggregatorActor) getAggregatedTicker(symbol string) (*domain.AggregatedTicker, error):
    RETURN aa.aggregateTickers(symbol), nil

FUNCTION (aa *AggregatorActor) getAggregatedOrderBook(symbol string) (*domain.AggregatedOrderBook, error):
    RETURN aa.aggregateOrderBooks(symbol), nil
```

**Estimated Time**: 8 hours

---

#### Task 6.7: Metrics Collector Actor

**Objective**: Collect and aggregate metrics.

**Files**: `internal/actor/metrics.go`

**Pseudocode**:

```
TYPE MetricsCollectorActor STRUCT:
    act.Actor
    exchangeMetrics sync.Map // exchange -> *domain.Metrics
    globalMetrics *domain.GlobalMetrics
    collectionInterval time.Duration
    mutex sync.RWMutex

TYPE GlobalMetrics STRUCT:
    TotalOrders int64
    TotalFills int64
    TotalCancels int64
    TotalErrors int64
    TotalReconnects int64
    StartTime time.Time
    Uptime time.Duration

FUNCTION (mc *MetricsCollectorActor) Init(args ...interface{}) error:
    mc.collectionInterval = 10 * time.Second
    mc.globalMetrics = &GlobalMetrics{
        StartTime: time.Now(),
    }

    // Schedule periodic collection
    gen.CronSendAfter(mc.Self(), mc.collectionInterval, &message.CollectMetrics{})

    // Subscribe to all metric events
    gen.SubscribeEvent(mc.Self(), "metrics.*")

    LOG "Metrics collector actor initialized"
    RETURN nil

FUNCTION (mc *MetricsCollectorActor) HandleEvent(message gen.MessageEvent):
    SWITCH message.Message TYPE:
        CASE *message.OrderPlaced:
            CALL mc.recordOrderPlaced(message.Message)

        CASE *message.OrderFilled:
            CALL mc.recordOrderFilled(message.Message)

        CASE *message.OrderCanceled:
            CALL mc.recordOrderCanceled(message.Message)

        CASE *message.OrderError:
            CALL mc.recordOrderError(message.Message)

        CASE *message.ConnectionReconnected:
            CALL mc.recordReconnect(message.Message)

FUNCTION (mc *MetricsCollectorActor) HandleCast(from gen.PID, message interface{}):
    SWITCH message TYPE:
        CASE *message.CollectMetrics:
            CALL mc.collectMetrics()
            gen.CronSendAfter(mc.Self(), mc.collectionInterval, &message.CollectMetrics{})

FUNCTION (mc *MetricsCollectorActor) HandleCall(from gen.PID, ref gen.Ref, request interface{}) (interface{}, error):
    SWITCH request TYPE:
        CASE *message.GetMetrics:
            RETURN mc.getMetrics(request.Exchange)

        CASE *message.GetGlobalMetrics:
            RETURN mc.getGlobalMetrics()

        DEFAULT:
            RETURN nil, errors.New("unknown request type")

FUNCTION (mc *MetricsCollectorActor) recordOrderPlaced(msg *message.OrderPlaced):
    metricsInterface, _ := mc.exchangeMetrics.LoadOrStore(msg.Exchange, &domain.Metrics{
        Exchange: msg.Exchange,
    })

    metrics := metricsInterface.(*domain.Metrics)
    atomic.AddInt64(&metrics.Orders.TotalOrders, 1)
    atomic.AddInt64(&mc.globalMetrics.TotalOrders, 1)

FUNCTION (mc *MetricsCollectorActor) recordOrderFilled(msg *message.OrderFilled):
    metricsInterface, ok := mc.exchangeMetrics.Load(msg.Exchange)
    IF NOT ok THEN
        RETURN

    metrics := metricsInterface.(*domain.Metrics)
    atomic.AddInt64(&metrics.Orders.FilledOrders, 1)
    atomic.AddInt64(&mc.globalMetrics.TotalFills, 1)

FUNCTION (mc *MetricsCollectorActor) recordOrderCanceled(msg *message.OrderCanceled):
    metricsInterface, ok := mc.exchangeMetrics.Load(msg.Exchange)
    IF NOT ok THEN
        RETURN

    metrics := metricsInterface.(*domain.Metrics)
    atomic.AddInt64(&metrics.Orders.CanceledOrders, 1)
    atomic.AddInt64(&mc.globalMetrics.TotalCancels, 1)

FUNCTION (mc *MetricsCollectorActor) recordOrderError(msg *message.OrderError):
    metricsInterface, ok := mc.exchangeMetrics.Load(msg.Exchange)
    IF NOT ok THEN
        RETURN

    metrics := metricsInterface.(*domain.Metrics)
    atomic.AddInt64(&metrics.Orders.RejectedOrders, 1)
    atomic.AddInt64(&mc.globalMetrics.TotalErrors, 1)

FUNCTION (mc *MetricsCollectorActor) recordReconnect(msg *message.ConnectionReconnected):
    metricsInterface, ok := mc.exchangeMetrics.Load(msg.Exchange)
    IF NOT ok THEN
        RETURN

    metrics := metricsInterface.(*domain.Metrics)
    atomic.AddInt64(&metrics.Connection.ReconnectCount, 1)
    atomic.AddInt64(&mc.globalMetrics.TotalReconnects, 1)

FUNCTION (mc *MetricsCollectorActor) collectMetrics():
    // Calculate uptime
    mc.mutex.Lock()
    mc.globalMetrics.Uptime = time.Since(mc.globalMetrics.StartTime)
    mc.mutex.Unlock()

    // Emit metrics event
    mc.exchangeMetrics.Range(FUNC(key, value interface{}):
        exchange := key.(string)
        metrics := value.(*domain.Metrics)

        // Calculate error rate
        total := atomic.LoadInt64(&metrics.Orders.TotalOrders)
        IF total > 0 THEN
            rejected := atomic.LoadInt64(&metrics.Orders.RejectedOrders)
            errorRate := float64(rejected) / float64(total)
            metrics.Orders.ErrorRate = errorRate

        metrics.Timestamp = time.Now()

        // Emit metrics event
        EMIT metrics.update event with exchange and metrics

        RETURN true
    )

FUNCTION (mc *MetricsCollectorActor) getMetrics(exchange string) (*domain.Metrics, error):
    metricsInterface, ok := mc.exchangeMetrics.Load(exchange)
    IF NOT ok THEN
        RETURN nil, errors.ErrExchangeNotFound

    metrics := metricsInterface.(*domain.Metrics)

    // Create copy
    metricsCopy := *metrics
    metricsCopy.Timestamp = time.Now()

    RETURN &metricsCopy, nil

FUNCTION (mc *MetricsCollectorActor) getGlobalMetrics() (*GlobalMetrics, error):
    mc.mutex.RLock()
    metrics := *mc.globalMetrics
    mc.mutex.RUnlock()

    metrics.Uptime = time.Since(metrics.StartTime)

    RETURN &metrics, nil
```

**Estimated Time**: 5 hours

---

#### Task 6.8: Connector Public API - Structure

**Objective**: Create main Connector struct with Ergo integration.

**Files**: `pkg/connector/connector.go`

**Pseudocode**:

```
TYPE Connector STRUCT:
    config *config.Config
    node *gen.Node
    rootSupervisor gen.PID
    coordinatorPID gen.PID
    healthCheckPID gen.PID
    aggregatorPID gen.PID
    riskManagerPID gen.PID
    metricsCollectorPID gen.PID
    running atomic.Bool
    eventHandlers sync.Map // topic -> []HandlerFunc
    mutex sync.RWMutex

TYPE HandlerFunc FUNC(event interface{})

FUNCTION New(opts ...connector.Option) (*Connector, error):
    // Default config
    cfg := &config.Config{
        NodeName: "exchange-connector",
        Exchanges: []config.ExchangeConfig{},
        Persistence: config.PersistenceConfig{
            Backend: "memory",
        },
    }

    // Apply options
    FOR EACH opt IN opts DO
        err := opt(cfg)
        IF err != nil THEN
            RETURN nil, err

    // Validate config
    err := validateConfig(cfg)
    IF err != nil THEN
        RETURN nil, err

    c := &Connector{
        config: cfg,
    }

    RETURN c, nil

FUNCTION (c *Connector) Start() error:
    IF c.running.Load() THEN
        RETURN errors.ErrAlreadyRunning

    // Create Ergo node
    nodeOpts := gen.NodeOptions{
        Name: c.config.NodeName,
    }

    // Set custom logger if provided
    IF c.config.Logger != nil THEN
        nodeOpts.Log = log.NewZerologAdapter(c.config.Logger)

    node, err := gen.StartNode(nodeOpts)
    IF err != nil THEN
        RETURN fmt.Errorf("failed to start node: %w", err)

    c.node = node

    // Start root supervisor
    pid, err := node.Spawn("RootSupervisor", gen.ProcessOptions{}, &RootSupervisor{}, c.config)
    IF err != nil THEN
        node.Stop()
        RETURN fmt.Errorf("failed to start root supervisor: %w", err)

    c.rootSupervisor = pid

    // Wait for supervisors to initialize
    time.Sleep(100 * time.Millisecond)

    // Get PIDs of core actors
    err = c.resolvePIDs()
    IF err != nil THEN
        c.Stop()
        RETURN err

    // Add configured exchanges
    FOR EACH exchangeConfig IN c.config.Exchanges DO
        err = c.AddExchange(exchangeConfig.Name, &exchangeConfig)
        IF err != nil THEN
            LOG warning "failed to add exchange" with exchangeConfig.Name and error

    c.running.Store(true)

    LOG "Connector started successfully"
    RETURN nil

FUNCTION (c *Connector) Stop() error:
    IF NOT c.running.Load() THEN
        RETURN errors.ErrNotRunning

    // Stop all exchanges
    exchanges := c.Exchanges()
    FOR EACH exchange IN exchanges DO
        err := c.RemoveExchange(exchange)
        IF err != nil THEN
            LOG warning "failed to remove exchange" with exchange and error

    // Stop node
    IF c.node != nil THEN
        c.node.Stop()
        c.node = nil

    c.running.Store(false)

    LOG "Connector stopped"
    RETURN nil

FUNCTION (c *Connector) IsRunning() bool:
    RETURN c.running.Load()

FUNCTION (c *Connector) resolvePIDs() error:
    // Get coordinator PID
    pid, err := c.node.ProcessByName("Coordinator")
    IF err != nil THEN
        RETURN fmt.Errorf("coordinator not found: %w", err)
    c.coordinatorPID = pid

    // Get health check PID
    pid, err = c.node.ProcessByName("HealthCheck")
    IF err != nil THEN
        RETURN fmt.Errorf("health check not found: %w", err)
    c.healthCheckPID = pid

    // Get risk manager PID
    pid, err = c.node.ProcessByName("RiskManager")
    IF err != nil THEN
        RETURN fmt.Errorf("risk manager not found: %w", err)
    c.riskManagerPID = pid

    // Get metrics collector PID
    pid, err = c.node.ProcessByName("MetricsCollector")
    IF err != nil THEN
        RETURN fmt.Errorf("metrics collector not found: %w", err)
    c.metricsCollectorPID = pid

    // Get aggregator PID (optional)
    IF c.config.Aggregation.Enabled THEN
        pid, err = c.node.ProcessByName("Aggregator")
        IF err == nil THEN
            c.aggregatorPID = pid

    RETURN nil

FUNCTION validateConfig(cfg *config.Config) error:
    IF cfg.NodeName == "" THEN
        RETURN errors.NewValidationError("NodeName", cfg.NodeName, "must not be empty")

    // Validate exchange configs
    FOR EACH exchangeConfig IN cfg.Exchanges DO
        IF exchangeConfig.Name == "" THEN
            RETURN errors.NewValidationError("Exchange.Name", "", "must not be empty")

        IF exchangeConfig.Type == "" THEN
            RETURN errors.NewValidationError("Exchange.Type", "", "must not be empty")

    RETURN nil
```

**Estimated Time**: 6 hours

---

#### Task 6.9: Connector Public API - Exchange Management

**Objective**: Implement exchange management methods.

**Files**: `pkg/connector/connector.go` (additions)

**Pseudocode**:

```
FUNCTION (c *Connector) AddExchange(name string, config *config.ExchangeConfig) error:
    IF NOT c.running.Load() THEN
        RETURN errors.ErrNotRunning

    // Validate config
    IF name == "" THEN
        RETURN errors.NewValidationError("name", name, "must not be empty")

    IF config == nil THEN
        RETURN errors.NewValidationError("config", nil, "must not be nil")

    // Send to root supervisor
    result, err := c.node.Call(c.rootSupervisor, &message.AddExchange{
        Name: name,
        Config: config,
    })

    IF err != nil THEN
        RETURN fmt.Errorf("failed to add exchange: %w", err)

    LOG "exchange added" with name
    RETURN nil

FUNCTION (c *Connector) RemoveExchange(name string) error:
    IF NOT c.running.Load() THEN
        RETURN errors.ErrNotRunning

    IF name == "" THEN
        RETURN errors.NewValidationError("name", name, "must not be empty")

    // Send to root supervisor
    _, err := c.node.Call(c.rootSupervisor, &message.RemoveExchange{
        Name: name,
    })

    IF err != nil THEN
        RETURN fmt.Errorf("failed to remove exchange: %w", err)

    LOG "exchange removed" with name
    RETURN nil

FUNCTION (c *Connector) Exchanges() []string:
    IF NOT c.running.Load() THEN
        RETURN []string{}

    result, err := c.node.Call(c.rootSupervisor, &message.GetExchanges{})
    IF err != nil THEN
        LOG error "failed to get exchanges" with err
        RETURN []string{}

    RETURN result.([]string)

FUNCTION (c *Connector) GetExchangeHealth(exchange string) (*domain.HealthStatus, error):
    IF NOT c.running.Load() THEN
        RETURN nil, errors.ErrNotRunning

    result, err := c.node.Call(c.healthCheckPID, &message.GetHealthStatus{
        Exchange: exchange,
    })

    IF err != nil THEN
        RETURN nil, err

    RETURN result.(*domain.HealthStatus), nil
```

**Estimated Time**: 3 hours

---

#### Task 6.10: Connector Public API - Market Data

**Objective**: Implement market data subscription methods.

**Files**: `pkg/connector/connector.go` (additions)

**Pseudocode**:

```
FUNCTION (c *Connector) SubscribeTicker(exchange, symbol string) error:
    IF NOT c.running.Load() THEN
        RETURN errors.ErrNotRunning

    _, err := c.node.Call(c.coordinatorPID, &message.SubscribeTicker{
        Exchange: exchange,
        Symbol: symbol,
    })

    IF err != nil THEN
        RETURN fmt.Errorf("failed to subscribe to ticker: %w", err)

    LOG "subscribed to ticker" with exchange and symbol
    RETURN nil

FUNCTION (c *Connector) SubscribeOrderBook(exchange, symbol string, depth int) error:
    IF NOT c.running.Load() THEN
        RETURN errors.ErrNotRunning

    IF depth <= 0 THEN
        depth = 50 // Default depth

    _, err := c.node.Call(c.coordinatorPID, &message.SubscribeOrderBook{
        Exchange: exchange,
        Symbol: symbol,
        Depth: depth,
    })

    IF err != nil THEN
        RETURN fmt.Errorf("failed to subscribe to order book: %w", err)

    LOG "subscribed to order book" with exchange, symbol, and depth
    RETURN nil

FUNCTION (c *Connector) SubscribeTrades(exchange, symbol string) error:
    IF NOT c.running.Load() THEN
        RETURN errors.ErrNotRunning

    _, err := c.node.Call(c.coordinatorPID, &message.SubscribeTrades{
        Exchange: exchange,
        Symbol: symbol,
    })

    IF err != nil THEN
        RETURN fmt.Errorf("failed to subscribe to trades: %w", err)

    LOG "subscribed to trades" with exchange and symbol
    RETURN nil

FUNCTION (c *Connector) Unsubscribe(exchange, symbol string) error:
    IF NOT c.running.Load() THEN
        RETURN errors.ErrNotRunning

    _, err := c.node.Call(c.coordinatorPID, &message.Unsubscribe{
        Exchange: exchange,
        Symbol: symbol,
    })

    IF err != nil THEN
        RETURN fmt.Errorf("failed to unsubscribe: %w", err)

    LOG "unsubscribed" with exchange and symbol
    RETURN nil
```

**Estimated Time**: 3 hours

---

#### Task 6.11: Connector Public API - Trading

**Objective**: Implement trading methods.

**Files**: `pkg/connector/connector.go` (additions)

**Pseudocode**:

```
FUNCTION (c *Connector) PlaceOrder(req *domain.OrderRequest) (*domain.Order, error):
    IF NOT c.running.Load() THEN
        RETURN nil, errors.ErrNotRunning

    IF req == nil THEN
        RETURN nil, errors.NewValidationError("request", nil, "must not be nil")

    // Validate request
    err := validateOrderRequest(req)
    IF err != nil THEN
        RETURN nil, err

    // Send to coordinator
    result, err := c.node.Call(c.coordinatorPID, &message.PlaceOrder{
        Request: req,
    })

    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to place order: %w", err)

    orderResult := result.(*message.OrderResult)
    IF orderResult.Error != nil THEN
        RETURN nil, orderResult.Error

    RETURN orderResult.Order, nil

FUNCTION (c *Connector) CancelOrder(req *domain.CancelRequest) error:
    IF NOT c.running.Load() THEN
        RETURN errors.ErrNotRunning

    IF req == nil THEN
        RETURN errors.NewValidationError("request", nil, "must not be nil")

    // Validate request
    IF req.Exchange == "" THEN
        RETURN errors.NewValidationError("Exchange", "", "must not be empty")

    IF req.OrderID == "" AND req.ClientOrderID == "" THEN
        RETURN errors.NewValidationError("OrderID/ClientOrderID", "", "one must be provided")

    // Send to coordinator
    _, err := c.node.Call(c.coordinatorPID, &message.CancelOrder{
        Request: req,
    })

    IF err != nil THEN
        RETURN fmt.Errorf("failed to cancel order: %w", err)

    RETURN nil

FUNCTION (c *Connector) GetOrder(exchange, orderID string) (*domain.Order, error):
    IF NOT c.running.Load() THEN
        RETURN nil, errors.ErrNotRunning

    result, err := c.node.Call(c.coordinatorPID, &message.GetOrder{
        Exchange: exchange,
        OrderID: orderID,
    })

    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to get order: %w", err)

    RETURN result.(*domain.Order), nil

FUNCTION (c *Connector) GetOpenOrders(exchange, symbol string) ([]*domain.Order, error):
    IF NOT c.running.Load() THEN
        RETURN nil, errors.ErrNotRunning

    result, err := c.node.Call(c.coordinatorPID, &message.GetOpenOrders{
        Exchange: exchange,
        Symbol: symbol,
    })

    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to get open orders: %w", err)

    RETURN result.([]*domain.Order), nil

FUNCTION validateOrderRequest(req *domain.OrderRequest) error:
    IF req.Exchange == "" THEN
        RETURN errors.NewValidationError("Exchange", "", "must not be empty")

    IF req.Symbol == "" THEN
        RETURN errors.NewValidationError("Symbol", "", "must not be empty")

    IF req.Side != domain.SideBuy AND req.Side != domain.SideSell THEN
        RETURN errors.NewValidationError("Side", req.Side, "must be BUY or SELL")

    IF domain.IsZero(req.Quantity) OR domain.Compare(req.Quantity, domain.Zero()) <= 0 THEN
        RETURN errors.NewValidationError("Quantity", req.Quantity, "must be greater than zero")

    IF req.Type == domain.OrderTypeLimit OR req.Type == domain.OrderTypeStopLimit THEN
        IF domain.IsZero(req.Price) OR domain.Compare(req.Price, domain.Zero()) <= 0 THEN
            RETURN errors.NewValidationError("Price", req.Price, "must be greater than zero for limit orders")

    RETURN nil
```

**Estimated Time**: 4 hours

---

#### Task 6.12: Connector Public API - Account

**Objective**: Implement account query methods.

**Files**: `pkg/connector/connector.go` (additions)

**Pseudocode**:

```
FUNCTION (c *Connector) GetBalance(exchange, asset string) (*domain.Balance, error):
    IF NOT c.running.Load() THEN
        RETURN nil, errors.ErrNotRunning

    IF exchange == "" THEN
        RETURN nil, errors.NewValidationError("exchange", "", "must not be empty")

    IF asset == "" THEN
        RETURN nil, errors.NewValidationError("asset", "", "must not be empty")

    result, err := c.node.Call(c.coordinatorPID, &message.GetBalance{
        Exchange: exchange,
        Asset: asset,
    })

    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to get balance: %w", err)

    RETURN result.(*domain.Balance), nil

FUNCTION (c *Connector) GetPosition(exchange, symbol string) (*domain.Position, error):
    IF NOT c.running.Load() THEN
        RETURN nil, errors.ErrNotRunning

    IF exchange == "" THEN
        RETURN nil, errors.NewValidationError("exchange", "", "must not be empty")

    IF symbol == "" THEN
        RETURN nil, errors.NewValidationError("symbol", "", "must not be empty")

    result, err := c.node.Call(c.coordinatorPID, &message.GetPosition{
        Exchange: exchange,
        Symbol: symbol,
    })

    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to get position: %w", err)

    RETURN result.(*domain.Position), nil

FUNCTION (c *Connector) GetSymbols(exchange string) ([]*domain.SymbolInfo, error):
    IF NOT c.running.Load() THEN
        RETURN nil, errors.ErrNotRunning

    IF exchange == "" THEN
        RETURN nil, errors.NewValidationError("exchange", "", "must not be empty")

    result, err := c.node.Call(c.coordinatorPID, &message.GetSymbols{
        Exchange: exchange,
    })

    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to get symbols: %w", err)

    RETURN result.([]*domain.SymbolInfo), nil
```

**Estimated Time**: 2 hours

---

#### Task 6.13: Connector Public API - Event Handlers

**Objective**: Implement event handler registration.

**Files**: `pkg/connector/connector.go` (additions)

**Pseudocode**:

```
FUNCTION (c *Connector) OnTicker(handler func(*domain.Ticker)) (unsubscribe func()):
    IF NOT c.running.Load() THEN
        RETURN func() {}

    // Subscribe to all ticker events
    topic := "ticker.*.*"

    // Create handler actor
    handlerActor := &EventHandlerActor{
        handler: func(event interface{}):
            ticker, ok := event.(*domain.Ticker)
            IF ok THEN
                handler(ticker)
    }

    // Spawn handler actor
    pid, err := c.node.Spawn("TickerHandler", gen.ProcessOptions{}, handlerActor)
    IF err != nil THEN
        LOG error "failed to spawn ticker handler" with err
        RETURN func() {}

    // Subscribe to event
    err = c.node.SubscribeEvent(pid, topic)
    IF err != nil THEN
        LOG error "failed to subscribe to ticker events" with err
        c.node.Kill(pid)
        RETURN func() {}

    // Return unsubscribe function
    RETURN func():
        c.node.UnsubscribeEvent(pid, topic)
        c.node.Kill(pid)

FUNCTION (c *Connector) OnOrderBook(handler func(*domain.OrderBook, bool)) (unsubscribe func()):
    IF NOT c.running.Load() THEN
        RETURN func() {}

    topic := "orderbook.*.*"

    handlerActor := &EventHandlerActor{
        handler: func(event interface{}):
            // Event should be wrapped with isSnapshot flag
            bookEvent, ok := event.(*message.OrderBookEvent)
            IF ok THEN
                handler(bookEvent.OrderBook, bookEvent.IsSnapshot)
    }

    pid, err := c.node.Spawn("OrderBookHandler", gen.ProcessOptions{}, handlerActor)
    IF err != nil THEN
        LOG error "failed to spawn orderbook handler" with err
        RETURN func() {}

    err = c.node.SubscribeEvent(pid, topic)
    IF err != nil THEN
        LOG error "failed to subscribe to orderbook events" with err
        c.node.Kill(pid)
        RETURN func() {}

    RETURN func():
        c.node.UnsubscribeEvent(pid, topic)
        c.node.Kill(pid)

FUNCTION (c *Connector) OnTrade(handler func(*domain.Trade)) (unsubscribe func()):
    IF NOT c.running.Load() THEN
        RETURN func() {}

    topic := "trade.*.*"

    handlerActor := &EventHandlerActor{
        handler: func(event interface{}):
            trade, ok := event.(*domain.Trade)
            IF ok THEN
                handler(trade)
    }

    pid, err := c.node.Spawn("TradeHandler", gen.ProcessOptions{}, handlerActor)
    IF err != nil THEN
        LOG error "failed to spawn trade handler" with err
        RETURN func() {}

    err = c.node.SubscribeEvent(pid, topic)
    IF err != nil THEN
        LOG error "failed to subscribe to trade events" with err
        c.node.Kill(pid)
        RETURN func() {}

    RETURN func():
        c.node.UnsubscribeEvent(pid, topic)
        c.node.Kill(pid)

FUNCTION (c *Connector) OnOrder(handler func(*domain.Order)) (unsubscribe func()):
    IF NOT c.running.Load() THEN
        RETURN func() {}

    topic := "order.*.*"

    handlerActor := &EventHandlerActor{
        handler: func(event interface{}):
            order, ok := event.(*domain.Order)
            IF ok THEN
                handler(order)
    }

    pid, err := c.node.Spawn("OrderHandler", gen.ProcessOptions{}, handlerActor)
    IF err != nil THEN
        LOG error "failed to spawn order handler" with err
        RETURN func() {}

    err = c.node.SubscribeEvent(pid, topic)
    IF err != nil THEN
        LOG error "failed to subscribe to order events" with err
        c.node.Kill(pid)
        RETURN func() {}

    RETURN func():
        c.node.UnsubscribeEvent(pid, topic)
        c.node.Kill(pid)

FUNCTION (c *Connector) OnPosition(handler func(*domain.Position)) (unsubscribe func()):
    IF NOT c.running.Load() THEN
        RETURN func() {}

    topic := "position.*.*"

    handlerActor := &EventHandlerActor{
        handler: func(event interface{}):
            position, ok := event.(*domain.Position)
            IF ok THEN
                handler(position)
    }

    pid, err := c.node.Spawn("PositionHandler", gen.ProcessOptions{}, handlerActor)
    IF err != nil THEN
        LOG error "failed to spawn position handler" with err
        RETURN func() {}

    err = c.node.SubscribeEvent(pid, topic)
    IF err != nil THEN
        LOG error "failed to subscribe to position events" with err
        c.node.Kill(pid)
        RETURN func() {}

    RETURN func():
        c.node.UnsubscribeEvent(pid, topic)
        c.node.Kill(pid)

FUNCTION (c *Connector) OnConnection(handler func(string, domain.ConnectionStatus)) (unsubscribe func()):
    IF NOT c.running.Load() THEN
        RETURN func() {}

    topic := "connection.*"

    handlerActor := &EventHandlerActor{
        handler: func(event interface{}):
            connEvent, ok := event.(*message.ConnectionEvent)
            IF ok THEN
                handler(connEvent.Exchange, connEvent.Status)
    }

    pid, err := c.node.Spawn("ConnectionHandler", gen.ProcessOptions{}, handlerActor)
    IF err != nil THEN
        LOG error "failed to spawn connection handler" with err
        RETURN func() {}

    err = c.node.SubscribeEvent(pid, topic)
    IF err != nil THEN
        LOG error "failed to subscribe to connection events" with err
        c.node.Kill(pid)
        RETURN func() {}

    RETURN func():
        c.node.UnsubscribeEvent(pid, topic)
        c.node.Kill(pid)

FUNCTION (c *Connector) OnRiskAlert(handler func(*domain.RiskAlert)) (unsubscribe func()):
    IF NOT c.running.Load() THEN
        RETURN func() {}

    topic := "risk.alert"

    handlerActor := &EventHandlerActor{
        handler: func(event interface{}):
            alert, ok := event.(*domain.RiskAlert)
            IF ok THEN
                handler(alert)
    }

    pid, err := c.node.Spawn("RiskAlertHandler", gen.ProcessOptions{}, handlerActor)
    IF err != nil THEN
        LOG error "failed to spawn risk alert handler" with err
        RETURN func() {}

    err = c.node.SubscribeEvent(pid, topic)
    IF err != nil THEN
        LOG error "failed to subscribe to risk alert events" with err
        c.node.Kill(pid)
        RETURN func() {}

    RETURN func():
        c.node.UnsubscribeEvent(pid, topic)
        c.node.Kill(pid)

// Event handler actor
TYPE EventHandlerActor STRUCT:
    act.Actor
    handler func(interface{})

FUNCTION (eha *EventHandlerActor) Init(args ...interface{}) error:
    RETURN nil

FUNCTION (eha *EventHandlerActor) HandleEvent(message gen.MessageEvent):
    IF eha.handler != nil THEN
        eha.handler(message.Message)
```

**Estimated Time**: 6 hours

---

#### Task 6.14: Connector Public API - Aggregation

**Objective**: Implement aggregation methods.

**Files**: `pkg/connector/connector.go` (additions)

**Pseudocode**:

```
FUNCTION (c *Connector) GetAggregatedTicker(symbol string) (*domain.AggregatedTicker, error):
    IF NOT c.running.Load() THEN
        RETURN nil, errors.ErrNotRunning

    IF NOT c.config.Aggregation.Enabled THEN
        RETURN nil, errors.New("aggregation not enabled")

    IF symbol == "" THEN
        RETURN nil, errors.NewValidationError("symbol", "", "must not be empty")

    result, err := c.node.Call(c.aggregatorPID, &message.GetAggregatedTicker{
        Symbol: symbol,
    })

    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to get aggregated ticker: %w", err)

    RETURN result.(*domain.AggregatedTicker), nil

FUNCTION (c *Connector) GetAggregatedOrderBook(symbol string) (*domain.AggregatedOrderBook, error):
    IF NOT c.running.Load() THEN
        RETURN nil, errors.ErrNotRunning

    IF NOT c.config.Aggregation.Enabled THEN
        RETURN nil, errors.New("aggregation not enabled")

    IF symbol == "" THEN
        RETURN nil, errors.NewValidationError("symbol", "", "must not be empty")

    result, err := c.node.Call(c.aggregatorPID, &message.GetAggregatedOrderBook{
        Symbol: symbol,
    })

    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to get aggregated order book: %w", err)

    RETURN result.(*domain.AggregatedOrderBook), nil

FUNCTION (c *Connector) FindArbitrage(symbol string) ([]*domain.ArbitrageOpportunity, error):
    IF NOT c.running.Load() THEN
        RETURN nil, errors.ErrNotRunning

    IF NOT c.config.Aggregation.Enabled THEN
        RETURN nil, errors.New("aggregation not enabled")

    IF symbol == "" THEN
        RETURN nil, errors.NewValidationError("symbol", "", "must not be empty")

    result, err := c.node.Call(c.aggregatorPID, &message.FindArbitrage{
        Symbol: symbol,
    })

    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to find arbitrage: %w", err)

    RETURN result.([]*domain.ArbitrageOpportunity), nil

FUNCTION (c *Connector) OnArbitrage(handler func(*domain.ArbitrageOpportunity)) (unsubscribe func()):
    IF NOT c.running.Load() THEN
        RETURN func() {}

    IF NOT c.config.Aggregation.Enabled THEN
        RETURN func() {}

    topic := "arbitrage.opportunity"

    handlerActor := &EventHandlerActor{
        handler: func(event interface{}):
            opp, ok := event.(*domain.ArbitrageOpportunity)
            IF ok THEN
                handler(opp)
    }

    pid, err := c.node.Spawn("ArbitrageHandler", gen.ProcessOptions{}, handlerActor)
    IF err != nil THEN
        LOG error "failed to spawn arbitrage handler" with err
        RETURN func() {}

    err = c.node.SubscribeEvent(pid, topic)
    IF err != nil THEN
        LOG error "failed to subscribe to arbitrage events" with err
        c.node.Kill(pid)
        RETURN func() {}

    RETURN func():
        c.node.UnsubscribeEvent(pid, topic)
        c.node.Kill(pid)
```

**Estimated Time**: 3 hours

---

#### Task 6.15: Connector Public API - State Management

**Objective**: Implement state management methods.

**Files**: `pkg/connector/connector.go` (additions)

**Pseudocode**:

```
FUNCTION (c *Connector) Reconcile(exchange string) error:
    IF NOT c.running.Load() THEN
        RETURN errors.ErrNotRunning

    IF exchange == "" THEN
        RETURN errors.NewValidationError("exchange", "", "must not be empty")

    // Get reconciler PID
    reconcilerPID, err := c.node.ProcessByName("Reconciler")
    IF err != nil THEN
        RETURN fmt.Errorf("reconciler not found: %w", err)

    // Trigger reconciliation
    _, err = c.node.Call(reconcilerPID, &message.ReconcileExchange{
        Exchange: exchange,
    })

    IF err != nil THEN
        RETURN fmt.Errorf("reconciliation failed: %w", err)

    LOG "reconciliation completed" with exchange
    RETURN nil

FUNCTION (c *Connector) GetMetrics(exchange string) (*domain.Metrics, error):
    IF NOT c.running.Load() THEN
        RETURN nil, errors.ErrNotRunning

    result, err := c.node.Call(c.metricsCollectorPID, &message.GetMetrics{
        Exchange: exchange,
    })

    IF err != nil THEN
        RETURN nil, fmt.Errorf("failed to get metrics: %w", err)

    RETURN result.(*domain.Metrics), nil

FUNCTION (c *Connector) AddRiskHook(hook risk.RiskHook) error:
    IF NOT c.running.Load() THEN
        RETURN errors.ErrNotRunning

    _, err := c.node.Call(c.riskManagerPID, &message.AddRiskHook{
        Hook: hook,
    })

    IF err != nil THEN
        RETURN fmt.Errorf("failed to add risk hook: %w", err)

    LOG "risk hook added" with hook.Name()
    RETURN nil

FUNCTION (c *Connector) RemoveRiskHook(name string) error:
    IF NOT c.running.Load() THEN
        RETURN errors.ErrNotRunning

    _, err := c.node.Call(c.riskManagerPID, &message.RemoveRiskHook{
        Name: name,
    })

    IF err != nil THEN
        RETURN fmt.Errorf("failed to remove risk hook: %w", err)

    LOG "risk hook removed" with name
    RETURN nil
```

**Estimated Time**: 3 hours

---

### Phase 7: Examples & Documentation (Days 31-35)

---

#### Task 7.1: Basic Example

**Objective**: Create basic usage example.

**Files**: `examples/01-basic/main.go`

**Pseudocode**:

```
PACKAGE main

IMPORT (
    "log"
    "time"
    "github.com/user/exchange-connector/pkg/connector"
    "github.com/user/exchange-connector/pkg/config"
    "github.com/user/exchange-connector/pkg/domain"
)

FUNCTION main():
    // Create connector with single exchange
    c, err := connector.New(
        connector.WithNodeName("basic-example"),
        connector.WithExchange(config.ExchangeConfig{
            Name: "binance",
            Type: domain.ExchangeBinance,
            APIKey: "your-api-key",
            APISecret: "your-api-secret",
            Testnet: true,
        }),
    )

    IF err != nil THEN
        log.Fatal("failed to create connector:", err)

    // Start connector
    err = c.Start()
    IF err != nil THEN
        log.Fatal("failed to start connector:", err)

    DEFER c.Stop()

    log.Println("Connector started successfully")

    // Subscribe to ticker updates
    unsubscribe := c.OnTicker(FUNC(ticker *domain.Ticker):
        log.Printf("Ticker: %s %s - Bid: %s, Ask: %s, Last: %s\n",
            ticker.Exchange,
            ticker.Symbol,
            domain.String(ticker.Bid),
            domain.String(ticker.Ask),
            domain.String(ticker.Last),
        )
    )
    DEFER unsubscribe()

    // Subscribe to BTC/USDT ticker
    err = c.SubscribeTicker("binance", "BTC/USDT")
    IF err != nil THEN
        log.Fatal("failed to subscribe to ticker:", err)

    log.Println("Subscribed to BTC/USDT ticker")

    // Subscribe to order updates
    unsubscribeOrders := c.OnOrder(FUNC(order *domain.Order):
        log.Printf("Order update: %s %s %s %s - Status: %s\n",
            order.Exchange,
            order.Symbol,
            order.Side,
            order.Type,
            order.Status,
        )
    )
    DEFER unsubscribeOrders()

    // Place a test order
    order, err := c.PlaceOrder(&domain.OrderRequest{
        Exchange: "binance",
        Symbol: "BTC/USDT",
        Side: domain.SideBuy,
        Type: domain.OrderTypeLimit,
        Price: domain.MustDecimal("30000.00"),
        Quantity: domain.MustDecimal("0.001"),
        TimeInForce: domain.TimeInForceGTC,
    })

    IF err != nil THEN
        log.Printf("Failed to place order: %v\n", err)
    ELSE
        log.Printf("Order placed successfully: %s\n", order.ID)

        // Wait a bit then cancel
        time.Sleep(5 * time.Second)

        err = c.CancelOrder(&domain.CancelRequest{
            Exchange: "binance",
            OrderID: order.ID,
        })

        IF err != nil THEN
            log.Printf("Failed to cancel order: %v\n", err)
        ELSE
            log.Println("Order canceled successfully")

    // Keep running
    log.Println("Press Ctrl+C to exit")
    SELECT {} // Block forever
```

**Estimated Time**: 2 hours

---

#### Task 7.2: Market Data Only Example

**Objective**: Create market data streaming example.

**Files**: `examples/02-market-data/main.go`

**Pseudocode**:

```
PACKAGE main

IMPORT (
    "log"
    "sync"
    "github.com/user/exchange-connector/pkg/connector"
    "github.com/user/exchange-connector/pkg/config"
    "github.com/user/exchange-connector/pkg/domain"
)

FUNCTION main():
    c, err := connector.New(
        connector.WithNodeName("market-data-example"),
        connector.WithExchange(config.ExchangeConfig{
            Name: "binance",
            Type: domain.ExchangeBinance,
            Testnet: true,
        }),
    )

    IF err != nil THEN
        log.Fatal(err)

    err = c.Start()
    IF err != nil THEN
        log.Fatal(err)

    DEFER c.Stop()

    var wg sync.WaitGroup

    // Subscribe to tickers
    symbols := []string{"BTC/USDT", "ETH/USDT", "BNB/USDT"}

    unsubscribeTicker := c.OnTicker(FUNC(ticker *domain.Ticker):
        log.Printf("[TICKER] %s - Bid: %s @ %s | Ask: %s @ %s\n",
            ticker.Symbol,
            domain.String(ticker.Bid),
            domain.String(ticker.BidQty),
            domain.String(ticker.Ask),
            domain.String(ticker.AskQty),
        )
    )
    DEFER unsubscribeTicker()

    FOR EACH symbol IN symbols DO
        err = c.SubscribeTicker("binance", symbol)
        IF err != nil THEN
            log.Printf("Failed to subscribe to %s ticker: %v\n", symbol, err)

    // Subscribe to order books
    unsubscribeBook := c.OnOrderBook(FUNC(book *domain.OrderBook, isSnapshot bool):
        updateType := "UPDATE"
        IF isSnapshot THEN
            updateType = "SNAPSHOT"

        log.Printf("[BOOK %s] %s - Bids: %d, Asks: %d\n",
            updateType,
            book.Symbol,
            len(book.Bids),
            len(book.Asks),
        )

        IF len(book.Bids) > 0 AND len(book.Asks) > 0 THEN
            bestBid := book.Bids[0]
            bestAsk := book.Asks[0]
            spread := domain.Sub(bestAsk.Price, bestBid.Price)

            log.Printf("  Best Bid: %s @ %s | Best Ask: %s @ %s | Spread: %s\n",
                domain.String(bestBid.Price),
                domain.String(bestBid.Quantity),
                domain.String(bestAsk.Price),
                domain.String(bestAsk.Quantity),
                domain.String(spread),
            )
    )
    DEFER unsubscribeBook()

    FOR EACH symbol IN symbols DO
        err = c.SubscribeOrderBook("binance", symbol, 20)
        IF err != nil THEN
            log.Printf("Failed to subscribe to %s order book: %v\n", symbol, err)

    // Subscribe to trades
    unsubscribeTrades := c.OnTrade(FUNC(trade *domain.Trade):
        log.Printf("[TRADE] %s %s - %s @ %s (ID: %s)\n",
            trade.Symbol,
            trade.Side,
            domain.String(trade.Quantity),
            domain.String(trade.Price),
            trade.TradeID,
        )
    )
    DEFER unsubscribeTrades()

    FOR EACH symbol IN symbols DO
        err = c.SubscribeTrades("binance", symbol)
        IF err != nil THEN
            log.Printf("Failed to subscribe to %s trades: %v\n", symbol, err)

    log.Println("Streaming market data... Press Ctrl+C to exit")
    wg.Add(1)
    wg.Wait()
```

**Estimated Time**: 2 hours

---

#### Task 7.3: Multi-Exchange Example

**Objective**: Create multi-exchange example with aggregation.

**Files**: `examples/04-multi-exchange/main.go`

**Pseudocode**:

```
PACKAGE main

IMPORT (
    "log"
    "time"
    "github.com/user/exchange-connector/pkg/connector"
    "github.com/user/exchange-connector/pkg/config"
    "github.com/user/exchange-connector/pkg/domain"
)

FUNCTION main():
    c, err := connector.New(
        connector.WithNodeName("multi-exchange-example"),
        connector.WithExchange(config.ExchangeConfig{
            Name: "binance",
            Type: domain.ExchangeBinance,
            Testnet: true,
        }),
        connector.WithExchange(config.ExchangeConfig{
            Name: "bybit",
            Type: domain.ExchangeBybit,
            Testnet: true,
        }),
        connector.WithAggregation(config.AggregationConfig{
            Enabled: true,
            UpdateInterval: 1 * time.Second,
            Symbols: []string{"BTC/USDT", "ETH/USDT"},
        }),
    )

    IF err != nil THEN
        log.Fatal(err)

    err = c.Start()
    IF err != nil THEN
        log.Fatal(err)

    DEFER c.Stop()

    log.Println("Multi-exchange connector started")

    // Subscribe to tickers from both exchanges
    c.OnTicker(FUNC(ticker *domain.Ticker):
        log.Printf("[%s] %s - Bid: %s, Ask: %s\n",
            ticker.Exchange,
            ticker.Symbol,
            domain.String(ticker.Bid),
            domain.String(ticker.Ask),
        )
    )

    c.SubscribeTicker("binance", "BTC/USDT")
    c.SubscribeTicker("bybit", "BTC/USDT")

    // Subscribe to arbitrage opportunities
    c.OnArbitrage(FUNC(opp *domain.ArbitrageOpportunity):
        log.Printf("[ARBITRAGE] %s: Buy on %s @ %s, Sell on %s @ %s - Spread: %.2f%%\n",
            opp.Symbol,
            opp.BuyExchange,
            domain.String(opp.BuyPrice),
            opp.SellExchange,
            domain.String(opp.SellPrice),
            opp.SpreadPercent,
        )
    )

    // Periodically check aggregated data
    ticker := time.NewTicker(5 * time.Second)
    DEFER ticker.Stop()

    GO FUNC():
        FOR range ticker.C DO
            // Get aggregated ticker
            aggTicker, err := c.GetAggregatedTicker("BTC/USDT")
            IF err != nil THEN
                log.Printf("Failed to get aggregated ticker: %v\n", err)
                CONTINUE

            log.Printf("[AGGREGATED] BTC/USDT\n")
            log.Printf("  Best Bid: %s on %s\n",
                domain.String(aggTicker.BestBid),
                aggTicker.BestBidExchange,
            )
            log.Printf("  Best Ask: %s on %s\n",
                domain.String(aggTicker.BestAsk),
                aggTicker.BestAskExchange,
            )
            log.Printf("  Spread: %s\n", domain.String(aggTicker.Spread))
            log.Printf("  Exchanges: %v\n", aggTicker.Exchanges)

            // Check for arbitrage
            opportunities, err := c.FindArbitrage("BTC/USDT")
            IF err != nil THEN
                log.Printf("Failed to find arbitrage: %v\n", err)
                CONTINUE

            IF len(opportunities) > 0 THEN
                log.Printf("  Found %d arbitrage opportunities\n", len(opportunities))
    ()

    log.Println("Press Ctrl+C to exit")
    SELECT {}
```

**Estimated Time**: 3 hours

---

#### Task 7.4: Risk Hooks Example

**Objective**: Create custom risk hooks example.

**Files**: `examples/06-risk-hooks/main.go`

**Pseudocode**:

```
PACKAGE main

IMPORT (
    "log"
    "github.com/user/exchange-connector/pkg/connector"
    "github.com/user/exchange-connector/pkg/config"
    "github.com/user/exchange-connector/pkg/domain"
    "github.com/user/exchange-connector/pkg/risk"
)

// Custom risk hook
TYPE MaxDailyOrdersHook STRUCT:
    maxOrders int
    orderCount int

FUNCTION (h *MaxDailyOrdersHook) Name() string:
    RETURN "max_daily_orders"

FUNCTION (h *MaxDailyOrdersHook) Validate(order *domain.OrderRequest, metrics *domain.RiskMetrics) (*domain.RiskAlert, error):
    h.orderCount++

    IF h.orderCount > h.maxOrders THEN
        alert := &domain.RiskAlert{
            Severity: domain.SeverityCritical,
            RuleType: "MAX_DAILY_ORDERS",
            Message: fmt.Sprintf("Exceeded max daily orders: %d/%d", h.orderCount, h.maxOrders),
            Exchange: order.Exchange,
            Symbol: order.Symbol,
            Timestamp: time.Now(),
        }

        RETURN alert, errors.New("daily order limit exceeded")

    IF h.orderCount > int(float64(h.maxOrders) * 0.8) THEN
        alert := &domain.RiskAlert{
            Severity: domain.SeverityWarning,
            RuleType: "MAX_DAILY_ORDERS",
            Message: fmt.Sprintf("Approaching daily order limit: %d/%d", h.orderCount, h.maxOrders),
            Exchange: order.Exchange,
            Symbol: order.Symbol,
            Timestamp: time.Now(),
        }

        RETURN alert, nil

    RETURN nil, nil

FUNCTION main():
    c, err := connector.New(
        connector.WithNodeName("risk-hooks-example"),
        connector.WithExchange(config.ExchangeConfig{
            Name: "binance",
            Type: domain.ExchangeBinance,
            APIKey: "your-api-key",
            APISecret: "your-api-secret",
            Testnet: true,
        }),
        connector.WithRisk(config.RiskConfig{
            Enabled: true,
            MaxOrderSize: domain.MustDecimal("1.0"),
            MaxPositionSize: domain.MustDecimal("5.0"),
            DailyLossLimit: domain.MustDecimal("1000.0"),
        }),
    )

    IF err != nil THEN
        log.Fatal(err)

    err = c.Start()
    IF err != nil THEN
        log.Fatal(err)

    DEFER c.Stop()

    log.Println("Connector started with risk management")

    // Add custom risk hook
    customHook := &MaxDailyOrdersHook{
        maxOrders: 10,
    }

    err = c.AddRiskHook(customHook)
    IF err != nil THEN
        log.Fatal("Failed to add risk hook:", err)

    log.Println("Custom risk hook added")

    // Subscribe to risk alerts
    c.OnRiskAlert(FUNC(alert *domain.RiskAlert):
        log.Printf("[RISK ALERT] %s - %s: %s\n",
            alert.Severity,
            alert.RuleType,
            alert.Message,
        )
    )

    // Try to place orders
    FOR i := 0; i < 15; i++ DO
        log.Printf("Placing order %d...\n", i+1)

        order, err := c.PlaceOrder(&domain.OrderRequest{
            Exchange: "binance",
            Symbol: "BTC/USDT",
            Side: domain.SideBuy,
            Type: domain.OrderTypeLimit,
            Price: domain.MustDecimal("30000.00"),
            Quantity: domain.MustDecimal("0.001"),
            TimeInForce: domain.TimeInForceGTC,
        })

        IF err != nil THEN
            log.Printf("Order %d rejected: %v\n", i+1, err)
        ELSE
            log.Printf("Order %d placed: %s\n", i+1, order.ID)

            // Cancel immediately
            c.CancelOrder(&domain.CancelRequest{
                Exchange: "binance",
                OrderID: order.ID,
            })

        time.Sleep(1 * time.Second)

    log.Println("Risk hooks example completed")
```

**Estimated Time**: 3 hours

---

#### Task 7.5: Testing Example

**Objective**: Create testing example with mock driver.

**Files**: `examples/08-testing/main_test.go`

**Pseudocode**:

```
PACKAGE main_test

IMPORT (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/user/exchange-connector/pkg/connector"
    "github.com/user/exchange-connector/pkg/config"
    "github.com/user/exchange-connector/pkg/domain"
    "github.com/user/exchange-connector/internal/driver/mock"
)

FUNCTION TestBasicConnector(t *testing.T):
    // Create connector with mock driver
    c, err := connector.New(
        connector.WithNodeName("test-connector"),
        connector.WithExchange(config.ExchangeConfig{
            Name: "mock",
            Type: "mock",
        }),
    )

    require.NoError(t, err)

    err = c.Start()
    require.NoError(t, err)

    DEFER c.Stop()

    // Verify connector is running
    assert.True(t, c.IsRunning())

    // Verify exchange is registered
    exchanges := c.Exchanges()
    assert.Contains(t, exchanges, "mock")

FUNCTION TestTickerSubscription(t *testing.T):
    c, err := connector.New(
        connector.WithNodeName("test-ticker"),
        connector.WithExchange(config.ExchangeConfig{
            Name: "mock",
            Type: "mock",
        }),
    )

    require.NoError(t, err)

    err = c.Start()
    require.NoError(t, err)
    DEFER c.Stop()

    // Channel to receive ticker
    tickerChan := make(chan *domain.Ticker, 1)

    unsubscribe := c.OnTicker(FUNC(ticker *domain.Ticker):
        SELECT:
            CASE tickerChan <- ticker:
            DEFAULT:
    )
    DEFER unsubscribe()

    // Subscribe to ticker
    err = c.SubscribeTicker("mock", "BTC/USDT")
    require.NoError(t, err)

    // Get mock driver and trigger ticker event
    mockDriver := mock.GetDriver("mock")
    mockDriver.TriggerTicker(&domain.Ticker{
        Exchange: "mock",
        Symbol: "BTC/USDT",
        Bid: domain.MustDecimal("50000.00"),
        Ask: domain.MustDecimal("50001.00"),
        Last: domain.MustDecimal("50000.50"),
        Timestamp: time.Now(),
    })

    // Wait for ticker
    SELECT:
        CASE ticker := <-tickerChan:
            assert.Equal(t, "mock", ticker.Exchange)
            assert.Equal(t, "BTC/USDT", ticker.Symbol)
            assert.Equal(t, "50000.00", domain.String(ticker.Bid))

        CASE <-time.After(5 * time.Second):
            t.Fatal("timeout waiting for ticker")

FUNCTION TestOrderPlacement(t *testing.T):
    c, err := connector.New(
        connector.WithNodeName("test-orders"),
        connector.WithExchange(config.ExchangeConfig{
            Name: "mock",
            Type: "mock",
        }),
    )

    require.NoError(t, err)

    err = c.Start()
    require.NoError(t, err)
    DEFER c.Stop()

    // Channel to receive order updates
    orderChan := make(chan *domain.Order, 10)

    unsubscribe := c.OnOrder(FUNC(order *domain.Order):
        SELECT:
            CASE orderChan <- order:
            DEFAULT:
    )
    DEFER unsubscribe()

    // Place order
    order, err := c.PlaceOrder(&domain.OrderRequest{
        Exchange: "mock",
        Symbol: "BTC/USDT",
        Side: domain.SideBuy,
        Type: domain.OrderTypeLimit,
        Price: domain.MustDecimal("50000.00"),
        Quantity: domain.MustDecimal("0.1"),
        TimeInForce: domain.TimeInForceGTC,
    })

    require.NoError(t, err)
    assert.NotEmpty(t, order.ID)
    assert.Equal(t, domain.OrderStatusNew, order.Status)

    // Wait for order update
    SELECT:
        CASE updatedOrder := <-orderChan:
            assert.Equal(t, order.ID, updatedOrder.ID)

        CASE <-time.After(5 * time.Second):
            t.Fatal("timeout waiting for order update")

    // Get mock driver and trigger fill
    mockDriver := mock.GetDriver("mock")
    mockDriver.TriggerOrderFill(order.ID, domain.MustDecimal("0.05"))

    // Wait for fill update
    SELECT:
        CASE filledOrder := <-orderChan:
            assert.Equal(t, order.ID, filledOrder.ID)
            assert.Equal(t, domain.OrderStatusPartiallyFilled, filledOrder.Status)
            assert.Equal(t, "0.05", domain.String(filledOrder.FilledQty))

        CASE <-time.After(5 * time.Second):
            t.Fatal("timeout waiting for fill update")

    // Cancel order
    err = c.CancelOrder(&domain.CancelRequest{
        Exchange: "mock",
        OrderID: order.ID,
    })

    require.NoError(t, err)

    // Wait for cancel update
    SELECT:
        CASE canceledOrder := <-orderChan:
            assert.Equal(t, order.ID, canceledOrder.ID)
            assert.Equal(t, domain.OrderStatusCanceled, canceledOrder.Status)

        CASE <-time.After(5 * time.Second):
            t.Fatal("timeout waiting for cancel update")

FUNCTION TestRiskManagement(t *testing.T):
    c, err := connector.New(
        connector.WithNodeName("test-risk"),
        connector.WithExchange(config.ExchangeConfig{
            Name: "mock",
            Type: "mock",
        }),
        connector.WithRisk(config.RiskConfig{
            Enabled: true,
            MaxOrderSize: domain.MustDecimal("1.0"),
        }),
    )

    require.NoError(t, err)

    err = c.Start()
    require.NoError(t, err)
    DEFER c.Stop()

    // Try to place order exceeding max size
    _, err = c.PlaceOrder(&domain.OrderRequest{
        Exchange: "mock",
        Symbol: "BTC/USDT",
        Side: domain.SideBuy,
        Type: domain.OrderTypeLimit,
        Price: domain.MustDecimal("50000.00"),
        Quantity: domain.MustDecimal("2.0"), // Exceeds max
        TimeInForce: domain.TimeInForceGTC,
    })

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "risk")
```

**Estimated Time**: 4 hours

---

#### Task 7.6: README Documentation

**Objective**: Create comprehensive README.

**Files**: `README.md`

**Content Structure**

# Exchange Connector

A production-grade Go library for connecting to cryptocurrency exchanges using Actor-based, Event-driven architecture.

## Features

- Multi-exchange support (Binance, Bybit)
- Actor-based architecture using Ergo Framework
- Real-time market data streaming
- Order management with state tracking
- Position and balance tracking
- Cross-exchange aggregation
- Risk management with custom hooks
- Circuit breaker pattern
- Comprehensive error handling
- High-precision decimal arithmetic

## Installation

```bash
go get github.com/user/exchange-connector
```

## Quick Start

[Basic example code]

## Configuration

[Configuration options and examples]

## Architecture

[Architecture overview diagram and explanation]

## Examples

- Basic Usage
- Market Data Streaming
- Trading
- Multi-Exchange
- Aggregation
- Risk Management
- Testing

## API Reference

[Link to full API documentation]

## Contributing

[Contributing guidelines]

## License

[License information]

**Estimated Time**: 6 hours

# Features: Cryptocurrency Exchange Connector

**Domain:** Multi-Exchange Connector Package (Go 1.25+)
**Researched:** 2026-02-16
**Confidence:** HIGH

---

## Table Stakes (Must Have)

Features users expect. Missing these means the product feels incomplete or unusable for production.

### Core Connectivity

| Feature                               | Complexity | Why Essential                                                                                                            |
| ------------------------------------- | ---------- | ------------------------------------------------------------------------------------------------------------------------ |
| **REST API Client**                   | Medium     | Universal requirement for all exchanges. Every operation requires REST. Users cannot interact with exchanges without it. |
| **WebSocket Streams**                 | High       | Real-time data is non-negotiable for trading. Polling REST APIs introduces unacceptable latency for market data.         |
| **Authentication (HMAC/RSA/Ed25519)** | Medium     | Required for any private API operations. All exchanges require signed requests for trading.                              |
| **Rate Limiting**                     | Medium     | Exchanges ban clients that exceed limits. Essential for production reliability.                                          |
| **Reconnection Handling**             | High       | WebSocket connections drop. Automatic reconnection with resubscription is required for any serious use.                  |

### Market Data

| Feature                | Complexity | Why Essential                                                              |
| ---------------------- | ---------- | -------------------------------------------------------------------------- |
| **Ticker Streams**     | Low        | Most basic market data. Expected by every trading application.             |
| **Order Book Streams** | High       | Essential for market making, arbitrage, and most trading strategies.       |
| **Trade Streams**      | Low        | Historical and real-time trades are fundamental for analysis and strategy. |
| **Symbol/Market Info** | Low        | Required for price/quantity precision, min notional, etc.                  |

### Order Management

| Feature                | Complexity | Why Essential                                                                       |
| ---------------------- | ---------- | ----------------------------------------------------------------------------------- |
| **Place Order**        | Medium     | Core trading operation. Cannot trade without it.                                    |
| **Cancel Order**       | Medium     | Essential for risk management and strategy execution.                               |
| **Query Order Status** | Low        | Required for reconciliation and state management.                                   |
| **Order Type Support** | Medium     | LIMIT and MARKET are universally expected. Missing them makes the library unusable. |

### Account Data

| Feature               | Complexity | Why Essential                                               |
| --------------------- | ---------- | ----------------------------------------------------------- |
| **Balance Tracking**  | Low        | Required for position management and risk calculation.      |
| **Position Tracking** | Medium     | For derivatives/futures. Required for any leverage trading. |
| **User Data Stream**  | Medium     | Real-time order updates, balance changes via WebSocket.     |

### Error Handling

| Feature               | Complexity | Why Essential                                                                    |
| --------------------- | ---------- | -------------------------------------------------------------------------------- |
| **Structured Errors** | Low        | Users need to distinguish between rate limits, auth errors, network issues, etc. |
| **Retry Logic**       | Medium     | Transient errors happen. Without retry, users lose orders.                       |

---

## Differentiators (Competitive Advantage)

Features that set the product apart. Not universally expected, but highly valuable.

### Architecture & Design

| Feature                        | Complexity | Value Proposition                                                                                                                 |
| ------------------------------ | ---------- | --------------------------------------------------------------------------------------------------------------------------------- |
| **Actor-Based Architecture**   | High       | Erlang-style supervision provides fault isolation. One component crash doesn't bring down the system. Unique in Go ecosystem.     |
| **Unified Cross-Exchange API** | High       | Single interface for Binance and Bybit. Reduces integration effort for multi-exchange strategies. CCXT is the reference for this. |
| **CQRS Pattern for Orders**    | Medium     | Separate command (write) and query (read) paths. Enables high-throughput order state queries without blocking order placement.    |
| **Event-Driven Design**        | Medium     | Loose coupling via events. Users subscribe to what they need. Clean extension points for custom logic.                            |

### Reliability & Resilience

| Feature                     | Complexity | Value Proposition                                                                                         |
| --------------------------- | ---------- | --------------------------------------------------------------------------------------------------------- |
| **Circuit Breaker Pattern** | Medium     | Health monitoring per exchange. Automatic failover prevents cascading failures. Essential for production. |
| **Clock Synchronization**   | Medium     | Exchange APIs reject requests with stale timestamps. Automatic sync prevents timestamp errors.            |
| **Nonce Management**        | Low        | Unique nonces for signed requests. Prevents replay attacks and request rejection.                         |
| **Graceful Degradation**    | Medium     | System continues with reduced functionality when components fail.                                         |

### Advanced Features

| Feature                        | Complexity | Value Proposition                                                                                         |
| ------------------------------ | ---------- | --------------------------------------------------------------------------------------------------------- |
| **Cross-Exchange Aggregation** | High       | Merge order books from multiple exchanges. Enables arbitrage and best-price routing. Rare feature.        |
| **Risk Management Hooks**      | Medium     | Pre-trade validation callbacks. Users inject custom risk checks before orders execute.                    |
| **Decimal Precision Handling** | Medium     | Arbitrary precision for financial calculations. Prevents floating-point errors. Professional requirement. |
| **Backpressure Management**    | Medium     | Handle message bursts without overwhelming downstream consumers. Prevents memory exhaustion.              |

### Developer Experience

| Feature                         | Complexity | Value Proposition                                                                          |
| ------------------------------- | ---------- | ------------------------------------------------------------------------------------------ |
| **Comprehensive Documentation** | Medium     | API references, examples, migration guides. Reduces integration time significantly.        |
| **Test Infrastructure**         | Medium     | Mock exchange, integration test framework. Enables reliable testing without real API keys. |
| **Metrics & Observability**     | Medium     | Connection health, message rates, error rates. Critical for production monitoring.         |

---

## Anti-Features (What NOT to Build)

Features to explicitly NOT include. Common mistakes or scope creep to avoid.

### Out of Scope

| Anti-Feature              | Why Excluded                                                                | What to Do Instead                                                |
| ------------------------- | --------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| **Trading GUI/Dashboard** | Library focus is API connectivity. GUI is a separate product.               | Provide examples showing how to build UI on top.                  |
| **Strategy Framework**    | Users have their own strategies. Frameworks constrain flexibility.          | Provide event hooks for users to plug in their logic.             |
| **Backtesting Engine**    | Different domain with different requirements (historical data, simulation). | Focus on live trading. Users can use dedicated backtesting tools. |
| **Signal/Alert System**   | Notifications are application-level concerns.                               | Provide event streams for users to build alerts.                  |
| **Portfolio Management**  | Accounting, P&L tracking are separate concerns.                             | Provide balance/position data for users to build on.              |

### Deferred Features

| Anti-Feature             | Why Deferred                                                             | When to Consider                        |
| ------------------------ | ------------------------------------------------------------------------ | --------------------------------------- |
| **Margin Trading**       | Increases complexity significantly. Lower demand initially.              | After spot trading is solid.            |
| **Futures/Derivatives**  | Additional complexity: funding rates, liquidations, leverage management. | Phase 2 after spot is production-ready. |
| **FIX Protocol Support** | Niche requirement. Only institutional users need it.                     | If enterprise customers request it.     |
| **Additional Exchanges** | Binance + Bybit covers significant market share.                         | Add based on user demand.               |

### Architecture Anti-Patterns

| Anti-Pattern                     | Why Avoid                                         | What to Do Instead            |
| -------------------------------- | ------------------------------------------------- | ----------------------------- |
| **Global State**                 | Makes testing impossible. Causes race conditions. | Actor-based state isolation.  |
| **Blocking I/O in Hot Path**     | Kills throughput under load.                      | Async message passing.        |
| **Direct Goroutines for Errors** | No error handling, zombie processes.              | Use errgroup, supervisors.    |
| **Float64 for Money**            | Precision loss causes incorrect calculations.     | Use decimal arithmetic (apd). |
| **Shared State Between Actors**  | Breaks actor isolation. Hard to reason about.     | Message passing only.         |

---

## Feature Dependencies

```
Core Infrastructure
├── Actor Framework (Ergo)
│   └── Supervisors
│       ├── ExchangeSupervisor
│       │   ├── RESTClient Actor
│       │   ├── WebSocket Actor
│       │   ├── OrderCommand Actor
│       │   └── OrderState Actor
│       └── RootSupervisor
│
├── REST Client
│   ├── Authentication (HMAC/RSA)
│   ├── Rate Limiting
│   ├── Clock Sync
│   └── Nonce Management
│
├── WebSocket Client
│   ├── Connection Management
│   ├── Reconnection Logic
│   ├── Subscription Management
│   └── Heartbeat/Ping-Pong
│
└── Domain Models
    ├── Order
    ├── Ticker
    ├── OrderBook
    ├── Balance
    └── Position

Market Data Features
├── Ticker Stream (depends on WebSocket)
├── Order Book Stream (depends on WebSocket)
├── Trade Stream (depends on WebSocket)
└── Symbol Info (depends on REST)

Order Management Features
├── Place Order (depends on REST + Auth)
├── Cancel Order (depends on REST + Auth)
├── Order State Tracking (depends on OrderState Actor)
└── User Data Stream (depends on WebSocket + Auth)

Advanced Features
├── Circuit Breaker (depends on health monitoring)
├── Cross-Exchange Aggregation (depends on multiple exchanges)
├── Risk Management Hooks (depends on order flow)
└── Backpressure (depends on message queues)
```

### Critical Path

```
1. Actor Framework Setup
   └── 2. REST Client + Auth
       ├── 3a. Market Data (REST + WebSocket)
       └── 3b. Order Management (REST + Auth)
           ├── 4a. Order State Actor
           └── 4b. User Data Stream
               └── 5. Advanced Features
```

---

## Exchange-Specific Considerations

### Binance

| Aspect                | Details                                                                                |
| --------------------- | -------------------------------------------------------------------------------------- |
| **API Version**       | Spot API v3                                                                            |
| **Base URL**          | `https://api.binance.com`                                                              |
| **WebSocket URL**     | `wss://stream.binance.com:9443/ws`                                                     |
| **Rate Limit Type**   | Request weight system (1200/minute)                                                    |
| **Authentication**    | HMAC SHA256, RSA, Ed25519                                                              |
| **Order Types**       | LIMIT, MARKET, STOP_LOSS, STOP_LOSS_LIMIT, TAKE_PROFIT, TAKE_PROFIT_LIMIT, LIMIT_MAKER |
| **WebSocket Streams** | Single stream per connection, or combined streams                                      |
| **Special Features**  | OCO orders, iceberg orders, trailing stops                                             |
| **Testnet**           | Available for spot trading                                                             |

**Binance-Specific Implementation Notes:**

- Rate limits use weighted requests (different endpoints have different weights)
- WebSocket ping/pong every 20 seconds, pong within 60 seconds
- Order book updates use delta format requiring snapshot reconciliation
- Server time must be within 5 seconds for signed requests

### Bybit

| Aspect                | Details                                        |
| --------------------- | ---------------------------------------------- |
| **API Version**       | V5 Unified API                                 |
| **Base URL**          | `https://api.bybit.com`                        |
| **WebSocket URL**     | `wss://stream.bybit.com/v5/public/spot`        |
| **Rate Limit Type**   | Per-second limits (varies by endpoint)         |
| **Authentication**    | HMAC SHA256                                    |
| **Order Types**       | Limit, Market, PostOnly                        |
| **WebSocket Streams** | Category-based (spot, linear, inverse, option) |
| **Special Features**  | Batch orders, TP/SL in single order            |
| **Testnet**           | Available                                      |

**Bybit-Specific Implementation Notes:**

- V5 API unifies spot, futures, and options under single interface
- WebSocket requires authentication for private channels
- Different endpoints for different product categories
- More generous rate limits than Binance

### Key Differences

| Aspect             | Binance            | Bybit                  |
| ------------------ | ------------------ | ---------------------- |
| Rate Limit Model   | Weighted (complex) | Per-second (simpler)   |
| WebSocket Auth     | Signature-based    | JWT or API key         |
| Order Book Format  | Delta updates      | Snapshot + delta       |
| Market Data Format | Per-symbol streams | Category-based streams |
| Testnet Access     | API key required   | Public testnet         |

---

## MVP Recommendation

For initial release, prioritize:

### Must Have (MVP)

1. ✅ REST Client with authentication
2. ✅ WebSocket ticker stream
3. ✅ WebSocket order book stream
4. ✅ Place/cancel/query orders
5. ✅ Balance tracking
6. ✅ Rate limiting
7. ✅ Basic reconnection

### Should Have (Post-MVP)

1. ⏳ Trade streams
2. ⏳ User data stream
3. ⏳ Circuit breaker
4. ⏳ Clock synchronization
5. ⏳ Cross-exchange aggregation

### Could Have (Future)

1. 🔮 Additional order types (stop-loss, OCO)
2. 🔮 Advanced risk management hooks
3. 🔮 Futures/margin support
4. 🔮 Additional exchanges

---

## Sources

| Source                           | Type     | Confidence |
| -------------------------------- | -------- | ---------- |
| CCXT Documentation               | Context7 | HIGH       |
| go-binance SDK                   | Context7 | HIGH       |
| Binance Spot API Docs            | Context7 | HIGH       |
| Bybit V5 API Docs                | Context7 | HIGH       |
| Ably WebSocket Best Practices    | Webfetch | MEDIUM     |
| Upbit WebSocket Best Practices   | Webfetch | MEDIUM     |
| Openware API Integration Article | Webfetch | MEDIUM     |

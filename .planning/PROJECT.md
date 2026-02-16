# Multi-Exchange Connector Package

## What This Is

A production-grade Go library for connecting to cryptocurrency exchanges (Binance & Bybit) with an Actor-based, event-driven architecture powered by Ergo Framework. The package is designed as a reusable library that trading engines can import to get reliable exchange connectivity with built-in supervision, circuit breakers, and risk management hooks.

## Core Value

**Reliable exchange connectivity through supervised actors.** If any component fails, it restarts automatically without taking down the entire system. Trading engines can trust that market data flows, orders execute, and state persists — even when exchanges disconnect or APIs fail.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Domain models with decimal precision (Order, Ticker, OrderBook, Trade, Position, Balance)
- [ ] Driver interface for exchange implementations (Binance, Bybit, Mock)
- [ ] Actor-based WebSocket connections with auto-reconnection
- [ ] REST clients with rate limiting per exchange
- [ ] Order management with CQRS pattern (CommandActor for writes, StateActor for reads)
- [ ] Clock synchronization for signed requests
- [ ] Circuit breaker for exchange health monitoring
- [ ] Risk management hooks (max order size, position limits, daily loss)
- [ ] Cross-exchange aggregation (merged orderbooks, best prices)
- [ ] In-memory persistence with pluggable backend interface

### Out of Scope

- **PostgreSQL/Redis persistence** — Memory backend first, add backends later
- **More than 2 exchanges** — Binance + Bybit only for v1
- **Futures/Options trading** — Spot only initially
- **Web UI or CLI** — This is a library, not an application
- **Backtesting engine** — Consumer builds that on top

## Context

- **Ecosystem**: Cryptocurrency trading, high-frequency data, WebSocket streams
- **Pattern**: Actor model (Ergo Framework) with supervision trees
- **Similar to**: ccxt (JavaScript/Python) but with Go actors and supervision
- **Known issues**: Exchange APIs are unreliable, rate limits vary, timestamps drift

## Constraints

- **Go Version**: 1.25+ with modern features (iterators, min/max builtins, clear)
- **Actor Framework**: Ergo Framework (built-in supervision, events, network transparency)
- **Decimal Library**: cockroachdb/apd v3 (arbitrary precision for financial calculations)
- **WebSocket**: lxzan/gws (high performance, low allocation)
- **No float64 for money**: All prices/quantities use apd.Decimal
- **Library design**: No main(), no global state, consumer controls lifecycle

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Ergo Framework over raw goroutines | Built-in supervision, event system, network transparency | — Pending |
| CQRS for orders | Separate write path (commands) from read path (state queries) | — Pending |
| gws over gorilla/websocket | Lower allocation, better performance under load | — Pending |
| Memory persistence first | Simpler testing, add real backends when needed | — Pending |
| Driver interface pattern | Easy to add exchanges, mock for testing | — Pending |

---
*Last updated: 2025-02-16 after initialization*

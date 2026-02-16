# Agent Instructions & Rules - Exchange Connector

## Project Overview

**Project**: Multi-Exchange Connector Package  
**Language**: Go 1.25+  
**Architecture**: Actor-based (Ergo Framework) + Event-driven  
**Purpose**: Production-grade library for cryptocurrency exchange integration

---

## CRITICAL: MCP Context7 Usage Rules (MANDATORY)

### RULE: ALWAYS Use MCP Tools for External References

**MANDATORY REQUIREMENT**: Before writing ANY code that involves:

- External packages/libraries
- Exchange API integration
- Third-party dependencies
- API specifications

You MUST use the appropriate MCP context7 tools to fetch the latest documentation.

### 1. When to Use MCP Tools

**RULE**: Use MCP context7 tools in these scenarios (NON-NEGOTIABLE):

#### A. Package Documentation Lookup

```
BEFORE writing code using external package:
1. Use MCP context7 tool to search package documentation
2. Verify API methods and signatures
3. Check for latest version and breaking changes
4. Review best practices from official docs

Example:
- Before using Ergo Framework: Search "ergo.services/ergo documentation"
- Before using gws WebSocket: Search "lxzan/gws documentation"
- Before using resty: Search "resty.dev documentation"
```

#### B. Exchange API Documentation

```
MANDATORY before implementing driver:
1. Search official exchange API documentation
2. Verify endpoint URLs
3. Check authentication methods
4. Review rate limiting rules
5. Understand WebSocket message formats

Example:
- Before Binance driver: Search "Binance API v3 spot documentation"
- Before Bybit driver: Search "Bybit API v5 documentation"
- Check for API changes: Search "Binance API changelog"
```

#### C. Library Version Compatibility

```
ALWAYS verify before adding dependency:
1. Check Go version compatibility
2. Verify package is maintained
3. Review breaking changes
4. Check security advisories

Example:
- Search "cockroachdb/apd v3 Go 1.25 compatibility"
- Search "ergo framework v3.1 documentation"
```

### 2. MCP Search Patterns (MANDATORY)

**RULE**: Follow these exact search patterns:

#### For Go Packages

```
Format: "[package-name] [version] official documentation"

Examples:
- "ergo.services/ergo v3.1 official documentation"
- "lxzan/gws websocket documentation"
- "cockroachdb/apd v3 decimal documentation"
- "resty v3 http client documentation"
- "zerolog structured logging documentation"
```

#### For Exchange APIs

```
Format: "[exchange-name] [API-version] [category] documentation"

Examples:
- "Binance API v3 spot trading documentation"
- "Binance API v3 websocket streams documentation"
- "Bybit API v5 unified trading documentation"
- "Bybit API v5 websocket public documentation"
- "Binance API authentication HMAC SHA256"
```

#### For Specific Features

```
Format: "[package/exchange] [feature] implementation example"

Examples:
- "ergo framework supervisor implementation example"
- "ergo framework event system usage"
- "Binance websocket depth stream implementation"
- "Bybit v5 order placement example"
- "gws websocket reconnection example"
```

### 3. Verification Workflow (MANDATORY)

**RULE**: Follow this exact workflow for EVERY external integration:

```
STEP 1: SEARCH DOCUMENTATION
├── Use MCP tool to find official documentation
├── Verify it's the correct version
└── Read through relevant sections

STEP 2: VERIFY API SIGNATURES
├── Check method names
├── Verify parameter types
├── Confirm return types
└── Note any deprecation warnings

STEP 3: CHECK EXAMPLES
├── Search for official examples
├── Review community implementations
└── Identify best practices

STEP 4: VERIFY COMPATIBILITY
├── Check Go version requirements
├── Verify dependency versions
└── Review breaking changes

STEP 5: IMPLEMENT
├── Write code based on verified documentation
├── Add comments with documentation references
└── Include version information in comments
```

### 4. Documentation Reference Comments (MANDATORY)

**RULE**: Every external API integration MUST include documentation reference:

**CORRECT**:

```go
// PlaceOrder places a new order on Binance.
// API: POST /api/v3/order
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#new-order-trade
// Version: API v3 (verified 2024-01-15)
//
// Rate Limit: 1 weight per request
// Authentication: Required (HMAC SHA256)
func (rc *RESTClient) PlaceOrder(req *domain.OrderRequest) (*domain.Order, error) {
    // Implementation
}

// WSDepthUpdate represents Binance depth update message.
// WebSocket Stream: <symbol>@depth or <symbol>@depth@100ms
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#diff-depth-stream
// Version: Stream v3 (verified 2024-01-15)
//
// Message format:
//   {
//     "e": "depthUpdate",
//     "E": 123456789,
//     "s": "BTCUSDT",
//     "U": 157,
//     "u": 160,
//     "b": [["0.0024","10"]],
//     "a": [["0.0026","100"]]
//   }
type WSDepthUpdate struct {
    EventType     string     `json:"e"`
    EventTime     int64      `json:"E"`
    Symbol        string     `json:"s"`
    FirstUpdateID int64      `json:"U"`
    FinalUpdateID int64      `json:"u"`
    Bids          [][]string `json:"b"`
    Asks          [][]string `json:"a"`
}
```

**WRONG**:

```go
// PlaceOrder places order
func (rc *RESTClient) PlaceOrder(req *domain.OrderRequest) (*domain.Order, error) {
    // No documentation reference
}

type WSDepthUpdate struct {
    // No format documentation
}
```

### 5. Exchange API Integration Checklist (MANDATORY)

Before implementing ANY exchange driver, complete this checklist:

```
BINANCE DRIVER CHECKLIST:
[ ] Search "Binance API v3 general info" - verify base URL, rate limits
[ ] Search "Binance API v3 spot trading endpoints" - verify order endpoints
[ ] Search "Binance API v3 websocket streams" - verify stream names
[ ] Search "Binance API v3 authentication" - verify signature method
[ ] Search "Binance API v3 error codes" - verify error handling
[ ] Search "Binance API v3 filters" - verify symbol filters
[ ] Search "Binance API v3 rate limits" - verify weight system
[ ] Search "Binance testnet endpoints" - verify testnet URLs

BYBIT DRIVER CHECKLIST:
[ ] Search "Bybit API v5 overview" - verify API structure
[ ] Search "Bybit API v5 spot trading" - verify order endpoints
[ ] Search "Bybit API v5 websocket public" - verify public streams
[ ] Search "Bybit API v5 websocket private" - verify private streams
[ ] Search "Bybit API v5 authentication" - verify auth method
[ ] Search "Bybit API v5 error codes" - verify error handling
[ ] Search "Bybit API v5 rate limits" - verify limits
[ ] Search "Bybit testnet" - verify testnet configuration

ERGO FRAMEWORK CHECKLIST:
[ ] Search "ergo framework supervisor" - verify supervisor patterns
[ ] Search "ergo framework actor lifecycle" - verify Init/Terminate
[ ] Search "ergo framework message passing" - verify Call/Cast/Event
[ ] Search "ergo framework gen.PID" - verify PID usage
[ ] Search "ergo framework event system" - verify event topics

GWS WEBSOCKET CHECKLIST:
[ ] Search "lxzan/gws EventHandler" - verify interface implementation
[ ] Search "lxzan/gws connection management" - verify connect/disconnect
[ ] Search "lxzan/gws ping pong" - verify heartbeat handling
[ ] Search "lxzan/gws options" - verify configuration options
```

---

## CRITICAL: Git Commit Message Rules (MANDATORY)

### RULE: ALL Commits MUST Follow Conventional Commits Specification

**MANDATORY REQUIREMENT**: Before pushing ANY code to GitHub, commits MUST:

1. Follow Conventional Commits format
2. Be meaningful and descriptive
3. Reference related issues/tasks
4. Pass pre-commit verification

### 1. Conventional Commits Format (STRICT)

**RULE**: Every commit message MUST follow this format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Commit Types (MANDATORY)

**RULE**: Use ONLY these types:

```
feat:     A new feature (user-facing)
fix:      A bug fix (user-facing)
docs:     Documentation only changes
style:    Code style changes (formatting, missing semi-colons, etc)
refactor: Code change that neither fixes a bug nor adds a feature
perf:     Performance improvement
test:     Adding missing tests or correcting existing tests
build:    Changes to build system or dependencies
ci:       Changes to CI configuration files and scripts
chore:    Other changes that don't modify src or test files
revert:   Reverts a previous commit
```

#### Scopes (RECOMMENDED)

**RULE**: Use specific scopes to indicate area of change:

```
connector    - Main connector API
domain       - Domain models
config       - Configuration
errors       - Error handling
actor        - Actor implementations
supervisor   - Supervisor implementations
driver       - Driver interface or registry
binance      - Binance driver
bybit        - Bybit driver
mock         - Mock driver
persistence  - Persistence layer
risk         - Risk management
circuit      - Circuit breaker
event        - Event system
message      - Message types
market       - Market data
order        - Order management
sync         - Clock/nonce sync
ratelimit    - Rate limiting
websocket    - WebSocket implementation
rest         - REST client
example      - Examples
test         - Test infrastructure
```

### 2. Commit Message Examples

#### CORRECT Examples:

```
feat(binance): implement WebSocket ticker subscription

- Add ticker stream support for Binance
- Implement WSTicker struct based on API v3 docs
- Add ticker parsing and domain mapping
- Include reconnection handling

Refs: #12

---

fix(order): prevent race condition in order state updates

The order state actor was experiencing race conditions when
processing rapid updates from WebSocket and REST simultaneously.

Solution:
- Add mutex protection for order map access
- Implement event sequencing with sequence numbers
- Add integration test for concurrent updates

Fixes: #45

---

docs(driver): add Binance API v3 documentation references

- Add API endpoint documentation to all REST methods
- Include WebSocket stream format documentation
- Add rate limit information from official docs
- Update examples with latest API changes

---

refactor(actor): extract event handler into separate actor

- Create EventHandlerActor for cleaner separation
- Move event subscription logic from Connector
- Improve testability of event handling
- Reduce coupling between components

---

perf(decimal): optimize decimal arithmetic with object pooling

- Implement sync.Pool for decimal operations
- Reduce allocations in hot path by 60%
- Add benchmark results in comments
- Maintain precision and correctness

Benchmark:
  Before: 50000 ns/op, 5 allocs/op
  After:  20000 ns/op, 2 allocs/op

---

test(integration): add Binance testnet integration tests

- Add testcontainer setup for integration tests
- Implement complete order lifecycle test
- Add WebSocket reconnection test
- Mark tests with integration build tag

---

build(deps): upgrade ergo framework to v3.1.2

- Update ergo.services/ergo to v3.1.2
- Fix breaking changes in supervisor API
- Update all supervisor implementations
- Verify all tests pass

Breaking Changes:
- SupervisorSpec.Strategy now requires explicit restart policy
- Updated all supervisors to use RestartTransient

---

fix(binance)!: correct order status mapping for PENDING_CANCEL

BREAKING CHANGE: OrderStatus enum now includes CANCELING state

Previously PENDING_CANCEL was mapped to CANCELED which caused
incorrect state transitions. Now properly maps to CANCELING.

Migration:
- Update any code checking for CANCELED to also check CANCELING
- Review order state machine logic

Fixes: #67
```

#### WRONG Examples:

```
WRONG: "update code"
Reason: Too vague, no type, no scope, no description

WRONG: "fix bug"
Reason: No scope, no details about what was fixed

WRONG: "added new feature for binance"
Reason: Incorrect format, should be "feat(binance): ..."

WRONG: "WIP"
Reason: Work in progress commits should not be pushed

WRONG: "fixed stuff"
Reason: No scope, no meaningful description

WRONG: "asdfasdf"
Reason: Not a meaningful message

WRONG: "feat: implement everything"
Reason: Too broad, should be broken into multiple commits

WRONG: "fix the thing"
Reason: No scope, vague description
```

### 3. Commit Message Body Guidelines

**RULE**: Include body when:

- Change is complex or non-obvious
- Multiple files are affected
- Breaking changes are introduced
- Context is needed for future reference

**Format**:

```
<type>(<scope>): <subject>

<body explaining WHAT changed and WHY>

<footer with references and breaking changes>
```

**CORRECT Body Structure**:

```
feat(order): implement order state sequencing

Problem:
WebSocket order updates were arriving out of order causing
incorrect state transitions and user confusion.

Solution:
- Implement sequence number tracking per order
- Buffer out-of-order messages until gap is filled
- Add timeout mechanism to request missing sequences
- Emit warning events for sequence gaps

Technical Details:
- Use OrderEventSequencer with sequence validation
- Max buffer size of 100 messages per order
- 5 second timeout for missing sequences
- Graceful degradation on persistent gaps

Testing:
- Added unit tests for sequence validation
- Added integration test with simulated gaps
- Verified no memory leaks with long-running test

Refs: #34, #56
```

### 4. Breaking Changes (CRITICAL)

**RULE**: Breaking changes MUST be clearly marked:

**Format**:

```
<type>(<scope>)!: <description>

BREAKING CHANGE: <detailed explanation>

<migration guide>
```

**Example**:

```
refactor(connector)!: change OnTicker signature to include exchange name

BREAKING CHANGE: OnTicker handler now receives exchange name as first parameter

Before:
  connector.OnTicker(func(ticker *domain.Ticker) {
    // handler code
  })

After:
  connector.OnTicker(func(exchange string, ticker *domain.Ticker) {
    // handler code
  })

Migration:
1. Update all OnTicker handlers to accept exchange parameter
2. If exchange name not needed, use underscore: func(_ string, ticker *domain.Ticker)

Rationale:
This change allows handlers to distinguish tickers from different
exchanges without parsing ticker.Exchange field.

Refs: #89
```

### 5. Commit Size and Scope

**RULE**: Commits should be atomic and focused:

```
GOOD COMMIT SIZE:
- One logical change per commit
- Related changes grouped together
- Complete and compilable
- Tests included with feature/fix

EXAMPLE OF PROPER COMMIT SEQUENCE:
1. feat(domain): add Order domain model
2. feat(domain): add OrderRequest and CancelRequest types
3. feat(order): implement OrderStateActor
4. feat(order): implement OrderCommandActor
5. test(order): add unit tests for order actors
6. docs(order): add order management documentation

BAD COMMIT SIZE:
- Multiple unrelated changes in one commit
- Incomplete changes that don't compile
- Tests in separate commit from feature
- "Big bang" commits with everything

WRONG:
- "feat: implement entire order management system" (too big)
- "fix: various bugs" (too vague, unrelated changes)
```

### 6. Pre-Commit Checklist (MANDATORY)

**RULE**: Before EVERY commit, complete this checklist:

```
CODE QUALITY:
[ ] All tests pass (go test ./...)
[ ] No linting errors (golangci-lint run)
[ ] No vet errors (go vet ./...)
[ ] Code formatted (gofmt -s -w .)
[ ] Imports organized (goimports -w .)

DOCUMENTATION:
[ ] MCP search completed for external dependencies
[ ] Documentation references added to code
[ ] Package comments updated
[ ] README updated (if public API changed)

COMMIT MESSAGE:
[ ] Follows Conventional Commits format
[ ] Type is correct (feat/fix/docs/etc)
[ ] Scope is specified
[ ] Description is clear and concise
[ ] Body explains WHY (if needed)
[ ] Breaking changes marked with !
[ ] Issue references included

SECURITY:
[ ] No API keys or secrets in code
[ ] No sensitive data in commit message
[ ] Dependencies verified with MCP search

TESTING:
[ ] Tests added for new features
[ ] Tests updated for bug fixes
[ ] Integration tests pass (if applicable)
[ ] Manual testing completed
```

### 7. Commit Message Verification Script

**RULE**: Use this script to verify commit messages before push:

```bash
#!/bin/bash
# .git/hooks/commit-msg

commit_msg=$(cat "$1")

# Check format
if ! echo "$commit_msg" | grep -qE '^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+\))?: .+'; then
    echo "ERROR: Commit message does not follow Conventional Commits format"
    echo ""
    echo "Format: <type>[optional scope]: <description>"
    echo ""
    echo "Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert"
    echo ""
    echo "Example: feat(binance): implement WebSocket ticker subscription"
    exit 1
fi

# Check description length
first_line=$(echo "$commit_msg" | head -n1)
if [ ${#first_line} -gt 100 ]; then
    echo "ERROR: Commit subject line too long (max 100 characters)"
    echo "Current length: ${#first_line}"
    exit 1
fi

# Check description is not too short
subject=$(echo "$first_line" | sed 's/^[^:]*: //')
if [ ${#subject} -lt 10 ]; then
    echo "ERROR: Commit description too short (min 10 characters)"
    echo "Current length: ${#subject}"
    exit 1
fi

# Check for WIP or TODO
if echo "$commit_msg" | grep -qiE '(WIP|TODO|FIXME|XXX)'; then
    echo "WARNING: Commit message contains WIP/TODO/FIXME"
    echo "Are you sure you want to commit this? (y/n)"
    read -r response
    if [ "$response" != "y" ]; then
        exit 1
    fi
fi

echo "Commit message format: OK"
exit 0
```

### 8. Multi-Commit Workflow

**RULE**: For complex features, use this workflow:

```
STEP 1: Plan commits
└── Break feature into logical commits
    ├── Domain models
    ├── Business logic
    ├── Integration
    └── Tests and docs

STEP 2: Implement incrementally
└── Each commit should:
    ├── Compile successfully
    ├── Pass all tests
    └── Be a complete unit of work

STEP 3: Review before push
└── Squash WIP commits
└── Reorder for clarity
└── Verify commit messages

STEP 4: Push
└── All commits follow conventions
└── All tests pass
└── Documentation complete
```

**Example Sequence**:

```bash
# Good commit sequence
git commit -m "feat(domain): add Ticker domain model

Add Ticker struct with all required fields for market data.
Includes proper decimal types for prices and quantities.

Refs: #23"

git commit -m "feat(binance): implement ticker WebSocket stream

- Add WSTicker type from Binance API v3 docs
- Implement ticker parsing and validation
- Add domain mapping function
- Include error handling

API Reference: https://binance-docs.github.io/apidocs/spot/en/#symbol-ticker-streams

Refs: #23"

git commit -m "test(binance): add ticker WebSocket tests

- Add unit tests for ticker parsing
- Add integration test for ticker stream
- Verify reconnection handling
- Test error cases

Refs: #23"

git commit -m "docs(binance): document ticker implementation

- Add ticker API documentation
- Include usage examples
- Document WebSocket message format

Refs: #23"
```

### 9. Commit Message Templates

**RULE**: Use these templates for common scenarios:

#### Feature Addition

```
feat(<scope>): <concise description>

Add <feature> to support <use case>.

Implementation:
- <key change 1>
- <key change 2>
- <key change 3>

Testing:
- <test approach>

Refs: #<issue-number>
```

#### Bug Fix

```
fix(<scope>): <concise description>

Problem:
<description of the bug>

Root Cause:
<explanation of why bug occurred>

Solution:
- <fix detail 1>
- <fix detail 2>

Testing:
- <how bug was verified>
- <regression test added>

Fixes: #<issue-number>
```

#### Refactoring

```
refactor(<scope>): <concise description>

Motivation:
<why refactoring was needed>

Changes:
- <change 1>
- <change 2>

Benefits:
- <benefit 1>
- <benefit 2>

No functional changes.

Refs: #<issue-number>
```

#### Documentation

```
docs(<scope>): <concise description>

- <doc change 1>
- <doc change 2>
- <doc change 3>

Refs: #<issue-number>
```

### 10. Emergency Hotfix Process

**RULE**: For critical hotfixes:

```
1. Create hotfix branch from main
2. Make minimal fix
3. Write clear commit message:

fix(<scope>)!: <critical bug description>

HOTFIX: <severity> - <impact>

Problem:
<what broke in production>

Impact:
<who/what is affected>

Root Cause:
<quick analysis>

Fix:
<minimal change made>

Testing:
<verification performed>

Follow-up:
- <issue number for comprehensive fix>
- <issue number for post-mortem>

Emergency-Fix: #<issue>

4. Request immediate review
5. Deploy after approval
6. Create follow-up issue for proper fix
```

### 11. Rewriting History (Use With Caution)

**RULE**: Rewriting commit history is allowed ONLY:

- On feature branches before merge
- To fix commit message typos
- To squash WIP commits
- NEVER on main/master branch
- NEVER after pushing to shared branch

```bash
# Amend last commit message
git commit --amend

# Interactive rebase to fix last 3 commits
git rebase -i HEAD~3

# Squash commits
git rebase -i HEAD~5
# Mark commits as 'squash' or 'fixup'

# Change commit message
git rebase -i HEAD~3
# Mark commit as 'reword'
```

### 12. Commit Signing (Recommended)

**RULE**: Sign commits for verification:

```bash
# Configure GPG signing
git config --global user.signingkey <gpg-key-id>
git config --global commit.gpgsign true

# Commit with signature
git commit -S -m "feat(connector): add feature"

# Verify signature
git log --show-signature
```

---

## Core Principles

### 1. Actor-Based Design

**RULE**: Every concurrent component MUST be an actor  
**RULE**: Communication between components MUST use message passing (gen.Cast/gen.Call)  
**RULE**: Never share state between actors - use message passing  
**RULE**: Each actor MUST have single responsibility

**CORRECT**:

```go
type OrderStateActor struct {
    act.Actor
    persistence OrderStore
}

func (osa *OrderStateActor) HandleCall(from gen.PID, ref gen.Ref, request interface{}) (interface{}, error) {
    switch req := request.(type) {
    case *message.GetOrder:
        return osa.persistence.Get(req.Exchange, req.OrderID)
    }
}
```

**WRONG**:

```go
// Direct function call bypassing actor
func GetOrderDirect(orderID string) (*Order, error) {
    // This breaks actor isolation
}
```

### 2. Event-Driven Communication

**RULE**: Use events for one-to-many communication  
**RULE**: Use messages (Call/Cast) for one-to-one communication  
**RULE**: Event topics MUST follow pattern: `{type}.{exchange}.{symbol}`  
**RULE**: Always check if event handler is nil before calling

**CORRECT**:

```go
topic := event.TopicFor(event.TopicTicker, "binance", "BTC/USDT")
gen.SendEvent(topic, ticker)

if ws.callbacks.OnTicker != nil {
    ws.callbacks.OnTicker(ticker)
}
```

**WRONG**:

```go
// Broadcasting to multiple actors directly
for _, pid := range actors {
    gen.Cast(pid, ticker)
}

// Not checking nil
ws.callbacks.OnTicker(ticker)  // Panic if nil
```

### 3. Immutability

**RULE**: Domain models MUST be immutable after creation  
**RULE**: Use pointer receivers only for actors and types that need mutation  
**RULE**: Return copies, not references to internal state  
**RULE**: Use value receivers for domain models

**CORRECT**:

```go
func (osa *OrderStateActor) GetOrder(id string) (*domain.Order, error) {
    order := osa.orders[id]
    orderCopy := *order  // Return copy
    return &orderCopy, nil
}

// Domain model with value receiver
func (o Order) IsFinal() bool {
    return o.Status == OrderStatusFilled ||
           o.Status == OrderStatusCanceled
}
```

**WRONG**:

```go
// Exposing internal state
func (osa *OrderStateActor) GetOrder(id string) (*domain.Order, error) {
    return osa.orders[id], nil  // Direct reference
}
```

---

## Go 1.25+ Specific Rules

### 1. Use Range Over Iterators (Go 1.23+)

**RULE**: Use `iter.Seq` for custom iteration  
**RULE**: Prefer range-over-func for collection iteration

**CORRECT**:

```go
func (ob *OrderBook) Levels() iter.Seq[OrderBookLevel] {
    return func(yield func(OrderBookLevel) bool) {
        for _, level := range ob.Bids {
            if !yield(level) {
                return
            }
        }
    }
}

// Usage
for level := range orderBook.Levels() {
    process(level)
}
```

### 2. Use Generic Type Constraints

**RULE**: Use generics for type-safe collections  
**RULE**: Define constraints using interface syntax  
**RULE**: Avoid `any` unless absolutely necessary

**CORRECT**:

```go
type Numeric interface {
    ~int | ~int64 | ~float64
}

func Sum[T Numeric](values []T) T {
    var sum T
    for _, v := range values {
        sum += v
    }
    return sum
}
```

### 3. Use Clear Built-in for Maps/Slices (Go 1.21+)

**RULE**: Use `clear()` instead of reassigning empty map/slice  
**RULE**: Clear is more efficient than creating new collection

**CORRECT**:

```go
clear(orderMap)
clear(slice)
```

**WRONG**:

```go
orderMap = make(map[string]*Order)
slice = []Order{}
```

### 4. Use Min/Max Built-ins (Go 1.21+)

**RULE**: Use built-in `min()` and `max()` for comparable types

**CORRECT**:

```go
bestPrice := min(price1, price2, price3)
maxQty := max(qty1, qty2)
```

**WRONG**:

```go
bestPrice := price1
if price2 < bestPrice {
    bestPrice = price2
}
```

---

## Architecture Rules

### 1. Supervisor Hierarchy

**RULE**: RootSupervisor MUST be the top-level supervisor  
**RULE**: Each exchange MUST have its own supervisor  
**RULE**: Supervisor strategy MUST match component criticality:

- `OneForOne`: For independent actors (default)
- `AllForOne`: For tightly coupled actors (rare)
- `RestForOne`: For ordered dependencies

**CORRECT**:

```go
spec := gen.SupervisorSpec{
    Name: "ExchangeSupervisor",
    Strategy: gen.SupervisorStrategy{
        Type: gen.SupervisorStrategyOneForOne,
        Intensity: 5,      // Max 5 restarts
        Period: 10,        // In 10 seconds
        Restart: gen.SupervisorStrategyRestartTransient,  // Only restart on error
    },
}
```

### 2. Actor Lifecycle

**RULE**: All initialization MUST happen in `Init()`  
**RULE**: Cleanup MUST happen in `Terminate()`  
**RULE**: Never block in `Init()` - spawn goroutines if needed  
**RULE**: Always return error from `Init()` on failure

**CORRECT**:

```go
func (ws *WSActor) Init(args ...interface{}) error {
    // Parse args
    ws.exchange = args[0].(string)
    ws.url = args[1].(string)

    // Start async work
    go ws.connect()

    return nil
}

func (ws *WSActor) Terminate(reason error) {
    if ws.conn != nil {
        ws.conn.Close()
    }
}
```

**WRONG**:

```go
func (ws *WSActor) Init(args ...interface{}) error {
    ws.connect()  // Blocks Init
    return nil
}
```

### 3. Message Handling

**RULE**: Use `HandleCall` for request-response  
**RULE**: Use `HandleCast` for fire-and-forget  
**RULE**: Use `HandleEvent` for event subscriptions  
**RULE**: Always use type switch for message routing  
**RULE**: Return error from HandleCall on failure

**CORRECT**:

```go
func (oc *OrderCommandActor) HandleCall(from gen.PID, ref gen.Ref, request interface{}) (interface{}, error) {
    switch req := request.(type) {
    case *message.PlaceOrder:
        return oc.placeOrder(req)
    case *message.CancelOrder:
        return oc.cancelOrder(req)
    default:
        return nil, fmt.Errorf("unknown request type: %T", request)
    }
}
```

**WRONG**:

```go
func (oc *OrderCommandActor) HandleCall(from gen.PID, ref gen.Ref, request interface{}) (interface{}, error) {
    // Using reflection instead of type switch
    if req, ok := request.(*message.PlaceOrder); ok {
        return oc.placeOrder(req)
    }
    return nil, nil  // Silent failure
}
```

---

## Code Organization Rules

### 1. Package Structure

**RULE**: `pkg/` contains public API only  
**RULE**: `internal/` contains implementation details  
**RULE**: `driver/` contains exchange-specific code (can be internal)  
**RULE**: One package per directory  
**RULE**: Package name MUST match directory name

**Directory Structure**:

```
pkg/
  connector/     - Public connector API
  domain/        - Public domain models
  config/        - Public configuration
  errors/        - Public error types
  driver/        - Public driver interface
  persistence/   - Public persistence interfaces
  risk/          - Public risk interfaces
  circuit/       - Public circuit breaker

internal/
  app/           - Ergo application
  sup/           - Supervisors
  actor/         - Core actors
  market/        - Market data actors
  order/         - Order management actors
  sync/          - Clock/nonce sync
  ratelimit/     - Rate limiting
  message/       - Internal messages
  event/         - Event definitions
  driver/        - Driver implementations
    binance/
    bybit/
    mock/
  log/           - Logging adapter
  metrics/       - Metrics implementation
  util/          - Internal utilities
```

### 2. File Naming

**RULE**: Use snake_case for file names  
**RULE**: Use descriptive names, avoid abbreviations  
**RULE**: Group related code in same file  
**RULE**: Test files MUST have `_test.go` suffix

**Examples**:

```
order_state.go          - Order state actor
order_command.go        - Order command actor
websocket_client.go     - WebSocket client
rate_limit_guard.go     - Rate limit guard
```

### 3. File Organization

**RULE**: Order of declarations in Go file:

1. Package declaration
2. Imports (grouped: stdlib, external, internal)
3. Constants
4. Type definitions
5. Constructor functions
6. Methods (grouped by receiver)
7. Helper functions

**CORRECT**:

```go
package order

import (
    // Standard library
    "context"
    "fmt"
    "time"

    // External packages
    "ergo.services/ergo/gen"

    // Internal packages
    "github.com/user/exchange-connector/pkg/domain"
    "github.com/user/exchange-connector/internal/message"
)

const (
    maxRetries = 3
    timeout    = 5 * time.Second
)

type OrderCommandActor struct {
    act.Actor
    exchange string
    driver   driver.Driver
}

func NewOrderCommandActor(exchange string, driver driver.Driver) *OrderCommandActor {
    return &OrderCommandActor{
        exchange: exchange,
        driver:   driver,
    }
}

func (oc *OrderCommandActor) Init(args ...interface{}) error {
    // Implementation
}

func (oc *OrderCommandActor) HandleCall(from gen.PID, ref gen.Ref, request interface{}) (interface{}, error) {
    // Implementation
}

func (oc *OrderCommandActor) placeOrder(req *message.PlaceOrder) (*message.OrderResult, error) {
    // Implementation
}
```

---

## Error Handling Rules

### 1. Error Creation

**RULE**: Use structured errors from `pkg/errors/`  
**RULE**: Wrap errors with context using `fmt.Errorf` with `%w`  
**RULE**: Never ignore errors - handle or propagate  
**RULE**: Use sentinel errors for expected conditions

**CORRECT**:

```go
func (oc *OrderCommandActor) placeOrder(req *message.PlaceOrder) (*message.OrderResult, error) {
    if req.Request == nil {
        return nil, errors.NewValidationError("request", nil, "must not be nil")
    }

    order, err := oc.driver.PlaceOrder(req.Request)
    if err != nil {
        return nil, fmt.Errorf("failed to place order on %s: %w", oc.exchange, err)
    }

    return &message.OrderResult{Order: order}, nil
}

// Sentinel errors
var (
    ErrNotRunning       = errors.New("connector not running")
    ErrAlreadyRunning   = errors.New("connector already running")
    ErrExchangeNotFound = errors.New("exchange not found")
)
```

**WRONG**:

```go
func (oc *OrderCommandActor) placeOrder(req *message.PlaceOrder) (*message.OrderResult, error) {
    order, err := oc.driver.PlaceOrder(req.Request)
    if err != nil {
        log.Println("Error:", err)  // Just logging
        return nil, err  // No context
    }

    return &message.OrderResult{Order: order}, nil
}
```

### 2. Error Checking

**RULE**: Check errors immediately after function call  
**RULE**: Use `errors.Is()` for sentinel error comparison  
**RULE**: Use `errors.As()` for type assertion  
**RULE**: Never panic in library code

**CORRECT**:

```go
if err := c.Start(); err != nil {
    if errors.Is(err, errors.ErrAlreadyRunning) {
        return nil
    }
    return fmt.Errorf("start failed: %w", err)
}

var rateLimitErr *errors.RateLimitError
if errors.As(err, &rateLimitErr) {
    time.Sleep(rateLimitErr.RetryAfter)
}
```

**WRONG**:

```go
err := c.Start()
// Code between error and check
if err != nil {
    panic(err)  // Never panic in library
}

if err == errors.ErrAlreadyRunning {  // Don't use ==
    return nil
}
```

### 3. Error Recovery

**RULE**: Actors SHOULD recover from panics  
**RULE**: Log panic with stack trace  
**RULE**: Return error from recovered panic  
**RULE**: Let supervisor handle restart

**CORRECT**:

```go
func (oc *OrderCommandActor) HandleCall(from gen.PID, ref gen.Ref, request interface{}) (result interface{}, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic in HandleCall: %v\nStack: %s", r, debug.Stack())
            log.Error().Err(err).Msg("actor panic")
        }
    }()

    // Normal handling
    return oc.handleRequest(request)
}
```

---

## Decimal Arithmetic Rules

### 1. Always Use apd.Decimal

**RULE**: NEVER use `float64` for prices or quantities  
**RULE**: NEVER use `int` for financial calculations  
**RULE**: Always use `domain.Decimal` (alias for `*apd.Decimal`)  
**RULE**: Use helper functions from `domain/decimal.go`

**CORRECT**:

```go
price := domain.MustDecimal("50000.123456")
qty := domain.MustDecimal("0.001")
total := domain.Mul(price, qty)

// Comparison
if domain.Compare(price1, price2) > 0 {
    // price1 is greater
}
```

**WRONG**:

```go
price := 50000.123456  // Float64 - NEVER!
qty := 0.001
total := price * qty   // Loss of precision

// Comparison
if price1 > price2 {  // Won't work with Decimal
}
```

### 2. Decimal Conversion

**RULE**: Parse strings for decimal input  
**RULE**: Always check error when parsing  
**RULE**: Use `MustDecimal()` only for constants  
**RULE**: Convert to float64 only for display, never calculations

**CORRECT**:

```go
// From user input
price, err := domain.NewDecimal(priceStr)
if err != nil {
    return fmt.Errorf("invalid price: %w", err)
}

// Constants
minPrice := domain.MustDecimal("0.01")

// Display only
priceFloat, _ := domain.Float64(price)
fmt.Printf("Price: %.2f\n", priceFloat)
```

**WRONG**:

```go
// Parsing without error check
price := domain.MustDecimal(userInput)  // Can panic

// Using float for calculation
priceFloat := 50000.12
priceDecimal := domain.NewDecimalFromFloat64(priceFloat)  // Precision loss
```

### 3. Decimal Precision

**RULE**: Use precision of 34 digits (apd default)  
**RULE**: Round only for display, not for calculations  
**RULE**: Store original precision from exchange

**CORRECT**:

```go
ctx := apd.BaseContext.WithPrecision(34)
result := new(apd.Decimal)
ctx.Mul(result, price, qty)
```

---

## Concurrency Rules

### 1. Synchronization

**RULE**: Use `sync.Mutex` for protecting shared state  
**RULE**: Use `sync.RWMutex` for read-heavy workloads  
**RULE**: Use `sync.Map` for concurrent map access  
**RULE**: Use `atomic` types for simple counters  
**RULE**: Always defer `Unlock()` immediately after `Lock()`

**CORRECT**:

```go
type OrderStateActor struct {
    act.Actor
    orders map[string]*domain.Order
    mutex  sync.RWMutex
}

func (osa *OrderStateActor) GetOrder(id string) (*domain.Order, error) {
    osa.mutex.RLock()
    defer osa.mutex.RUnlock()

    order, ok := osa.orders[id]
    if !ok {
        return nil, errors.ErrOrderNotFound
    }

    orderCopy := *order
    return &orderCopy, nil
}

// Atomic counter
var messageCount atomic.Int64
messageCount.Add(1)
```

**WRONG**:

```go
func (osa *OrderStateActor) GetOrder(id string) (*domain.Order, error) {
    osa.mutex.RLock()

    order, ok := osa.orders[id]
    if !ok {
        osa.mutex.RUnlock()  // Manual unlock - error prone
        return nil, errors.ErrOrderNotFound
    }

    osa.mutex.RUnlock()
    return order, nil  // Returns reference, not copy
}
```

### 2. Goroutine Management

**RULE**: Always use context for cancellation  
**RULE**: Never start unbounded goroutines  
**RULE**: Always have exit condition for goroutines  
**RULE**: Use `errgroup` for goroutine error handling

**CORRECT**:

```go
func (ws *WSActor) readLoop(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            msg, err := ws.conn.ReadMessage()
            if err != nil {
                return
            }
            ws.handleMessage(msg)
        }
    }
}

// Using errgroup
g, ctx := errgroup.WithContext(context.Background())
g.Go(func() error {
    return ws.connect(ctx)
})
if err := g.Wait(); err != nil {
    return err
}
```

**WRONG**:

```go
func (ws *WSActor) readLoop() {
    for {  // No exit condition
        msg, err := ws.conn.ReadMessage()
        ws.handleMessage(msg)
    }
}

// Unmanaged goroutines
for _, exchange := range exchanges {
    go processExchange(exchange)  // No error handling, no context
}
```

### 3. Channel Usage

**RULE**: Always close channels from sender side  
**RULE**: Use buffered channels to prevent goroutine leaks  
**RULE**: Use select with timeout for channel operations  
**RULE**: Never close a channel from receiver side  
**RULE**: Never send on closed channel

**CORRECT**:

```go
ch := make(chan *domain.Order, 10)  // Buffered

// Sender
go func() {
    defer close(ch)
    for _, order := range orders {
        ch <- order
    }
}()

// Receiver with timeout
select {
case order := <-ch:
    process(order)
case <-time.After(5 * time.Second):
    return errors.New("timeout")
}
```

**WRONG**:

```go
ch := make(chan *domain.Order)  // Unbuffered - can block

go func() {
    for _, order := range orders {
        ch <- order  // Can leak if receiver stops
    }
    // No close
}()

// Unsafe receive
order := <-ch  // Blocks forever if sender stops
```

---

## WebSocket Rules

### 1. Connection Management

**RULE**: Always implement reconnection logic  
**RULE**: Use exponential backoff for reconnection  
**RULE**: Resubscribe to all topics after reconnection  
**RULE**: Track connection state with atomic boolean

**CORRECT**:

```go
func (ws *WSClient) reconnect() {
    attempt := 0
    for attempt < ws.config.MaxReconnectAttempts {
        delay := time.Duration(math.Pow(2, float64(attempt))) * ws.config.ReconnectDelay
        if delay > 60*time.Second {
            delay = 60 * time.Second
        }
        time.Sleep(delay)

        if err := ws.connect(); err == nil {
            ws.resubscribe()
            return
        }
        attempt++
    }
}
```

### 2. Message Handling

**RULE**: Parse messages on separate goroutine  
**RULE**: Use type switch for message routing  
**RULE**: Always validate message format  
**RULE**: Handle both snapshot and delta updates

**CORRECT**:

```go
func (ws *WSClient) handleMessage(data []byte) {
    go func() {
        var msg WSMessage
        if err := json.Unmarshal(data, &msg); err != nil {
            log.Error().Err(err).Msg("failed to parse message")
            return
        }

        switch msg.Type {
        case "ticker":
            ws.handleTicker(msg.Data)
        case "orderbook":
            ws.handleOrderBook(msg.Data, msg.IsSnapshot)
        default:
            log.Warn().Str("type", msg.Type).Msg("unknown message type")
        }
    }()
}
```

### 3. Subscription Management

**RULE**: Track subscriptions in sync.Map  
**RULE**: Support subscribe/unsubscribe at runtime  
**RULE**: Deduplicate subscription requests

**CORRECT**:

```go
func (ws *WSClient) Subscribe(topic string) error {
    if _, loaded := ws.subscriptions.LoadOrStore(topic, true); loaded {
        return nil  // Already subscribed
    }

    if !ws.IsConnected() {
        return nil  // Will subscribe on connect
    }

    return ws.sendSubscribe(topic)
}
```

---

## REST API Rules

### 1. Request Signing

**RULE**: Sign every authenticated request  
**RULE**: Use HMAC-SHA256 for signatures  
**RULE**: Include timestamp in signature  
**RULE**: Never log API secrets

**CORRECT**:

```go
func (s *Signer) Sign(params url.Values) string {
    params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))

    queryString := params.Encode()
    mac := hmac.New(sha256.New, []byte(s.apiSecret))
    mac.Write([]byte(queryString))

    return hex.EncodeToString(mac.Sum(nil))
}
```

### 2. Rate Limiting

**RULE**: Implement rate limiting before sending requests  
**RULE**: Use token bucket for general rate limiting  
**RULE**: Use weighted limiter for Binance  
**RULE**: Return rate limit error with retry-after

**CORRECT**:

```go
func (rc *RESTClient) do(method, endpoint string, weight int) error {
    if !rc.rateLimiter.Allow(weight) {
        waitTime := rc.rateLimiter.WaitTime()
        return errors.NewRateLimitError(rc.exchange, waitTime, weight)
    }

    // Proceed with request
}
```

### 3. Retry Logic

**RULE**: Retry on transient errors only  
**RULE**: Use exponential backoff  
**RULE**: Maximum 3 retries  
**RULE**: Never retry non-idempotent operations without checking

**CORRECT**:

```go
client := resty.New()
client.SetRetryCount(3)
client.SetRetryWaitTime(1 * time.Second)
client.SetRetryMaxWaitTime(5 * time.Second)
client.AddRetryCondition(func(r *resty.Response, err error) bool {
    return err != nil || r.StatusCode() >= 500
})
```

---

## Testing Rules

### 1. Test Structure

**RULE**: Use table-driven tests  
**RULE**: Test file MUST be in same package with `_test.go` suffix  
**RULE**: Use `testify/require` for fatal assertions  
**RULE**: Use `testify/assert` for non-fatal assertions

**CORRECT**:

```go
func TestOrderStateTransitions(t *testing.T) {
    tests := []struct {
        name      string
        current   domain.OrderStatus
        next      domain.OrderStatus
        wantValid bool
    }{
        {
            name:      "new to filled",
            current:   domain.OrderStatusNew,
            next:      domain.OrderStatusFilled,
            wantValid: true,
        },
        {
            name:      "filled to canceled",
            current:   domain.OrderStatusFilled,
            next:      domain.OrderStatusCanceled,
            wantValid: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            order := &domain.Order{Status: tt.current}
            valid := order.CanTransition(tt.next)
            assert.Equal(t, tt.wantValid, valid)
        })
    }
}
```

### 2. Mock Usage

**RULE**: Use mock driver for testing  
**RULE**: Never use real exchange APIs in tests  
**RULE**: Mock at interface boundary  
**RULE**: Verify mock calls

**CORRECT**:

```go
func TestPlaceOrder(t *testing.T) {
    mockDriver := mock.NewMockDriver()
    mockDriver.On("PlaceOrder", mock.Anything).Return(&domain.Order{
        ID: "123",
        Status: domain.OrderStatusNew,
    }, nil)

    actor := NewOrderCommandActor("mock", mockDriver)
    result, err := actor.placeOrder(&message.PlaceOrder{
        Request: &domain.OrderRequest{},
    })

    require.NoError(t, err)
    assert.Equal(t, "123", result.Order.ID)
    mockDriver.AssertExpectations(t)
}
```

### 3. Integration Testing

**RULE**: Use testcontainers for integration tests  
**RULE**: Mark integration tests with build tag  
**RULE**: Clean up resources in defer

**CORRECT**:

```go
//go:build integration

func TestBinanceIntegration(t *testing.T) {
    ctx := context.Background()

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "binance/testnet:latest",
        },
        Started: true,
    })
    require.NoError(t, err)
    defer container.Terminate(ctx)

    // Test code
}
```

---

## Logging Rules

### 1. Structured Logging

**RULE**: Always use structured logging with zerolog  
**RULE**: Include context in log entries  
**RULE**: Use appropriate log levels  
**RULE**: Never log sensitive data (API keys, secrets)

**Log Levels**:

- `Trace`: Very detailed, usually disabled
- `Debug`: Detailed for debugging
- `Info`: General information
- `Warn`: Warning, not an error but attention needed
- `Error`: Error that needs attention
- `Fatal`: Application cannot continue
- `Panic`: Severe error with panic

**CORRECT**:

```go
log.Info().
    Str("exchange", exchange).
    Str("symbol", symbol).
    Msg("subscribed to ticker")

log.Error().
    Err(err).
    Str("orderID", orderID).
    Str("exchange", exchange).
    Msg("failed to cancel order")
```

**WRONG**:

```go
log.Println("Subscribed to", exchange, symbol)  // Unstructured

log.Error().
    Str("apiKey", apiKey).  // NEVER log secrets
    Msg("authentication failed")
```

### 2. Error Logging

**RULE**: Log errors with full context  
**RULE**: Log at error level or above  
**RULE**: Include stack trace for unexpected errors

**CORRECT**:

```go
if err := ws.connect(); err != nil {
    log.Error().
        Err(err).
        Str("exchange", ws.exchange).
        Str("url", ws.url).
        Int("attempt", attempt).
        Msg("connection failed")
}
```

---

## Performance Rules

### 1. Memory Allocation

**RULE**: Minimize allocations in hot paths  
**RULE**: Reuse buffers where possible  
**RULE**: Use sync.Pool for frequently allocated objects  
**RULE**: Preallocate slices with known capacity

**CORRECT**:

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processMessage(data []byte) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    buf.Write(data)
    // Process
}

// Preallocate
orders := make([]*domain.Order, 0, 100)
```

**WRONG**:

```go
func processMessage(data []byte) {
    buf := new(bytes.Buffer)  // New allocation every time
    buf.Write(data)
}

orders := []*domain.Order{}  // No capacity hint
```

### 2. String Operations

**RULE**: Use strings.Builder for concatenation  
**RULE**: Avoid string concatenation in loops  
**RULE**: Use byte slices for performance-critical paths

**CORRECT**:

```go
var sb strings.Builder
sb.WriteString(exchange)
sb.WriteString(":")
sb.WriteString(symbol)
key := sb.String()
```

**WRONG**:

```go
key := exchange + ":" + symbol  // Multiple allocations
```

### 3. JSON Performance

**RULE**: Use json.Decoder for streaming  
**RULE**: Reuse json.Encoder/Decoder  
**RULE**: Use struct tags to avoid reflection

**CORRECT**:

```go
decoder := json.NewDecoder(reader)
var msg WSMessage
if err := decoder.Decode(&msg); err != nil {
    return err
}
```

---

## Security Rules

### 1. API Credentials

**RULE**: Never hardcode API keys  
**RULE**: Load credentials from environment or secure vault  
**RULE**: Never log API keys or secrets  
**RULE**: Clear sensitive data after use

**CORRECT**:

```go
apiKey := os.Getenv("BINANCE_API_KEY")
if apiKey == "" {
    return errors.New("BINANCE_API_KEY not set")
}

// Use apiKey
defer func() {
    apiKey = ""  // Clear
}()
```

### 2. Input Validation

**RULE**: Validate all user inputs  
**RULE**: Validate all exchange responses  
**RULE**: Use whitelisting over blacklisting  
**RULE**: Check bounds for numeric inputs

**CORRECT**:

```go
func ValidateOrderRequest(req *domain.OrderRequest) error {
    if req.Exchange == "" {
        return errors.NewValidationError("exchange", "", "must not be empty")
    }

    if !isValidSymbol(req.Symbol) {
        return errors.NewValidationError("symbol", req.Symbol, "invalid format")
    }

    if domain.Compare(req.Quantity, domain.Zero()) <= 0 {
        return errors.NewValidationError("quantity", req.Quantity, "must be positive")
    }

    return nil
}
```

### 3. Rate Limiting

**RULE**: Implement client-side rate limiting  
**RULE**: Respect exchange rate limits  
**RULE**: Handle rate limit errors gracefully

---

## Documentation Rules

### 1. Code Comments

**RULE**: Document all exported types and functions  
**RULE**: Use complete sentences in comments  
**RULE**: Explain why, not what (code shows what)  
**RULE**: Keep comments up to date with code

**CORRECT**:

```go
// OrderStateActor maintains the read model for order state.
// It processes order updates from various sources and provides
// consistent state queries. The actor ensures CQRS pattern by
// separating reads from writes.
type OrderStateActor struct {
    act.Actor
    orders map[string]*domain.Order
}

// CanTransition checks if an order can transition to the new status
// based on the current status. Returns false for invalid transitions
// to prevent state corruption.
func (o *Order) CanTransition(newStatus OrderStatus) bool {
    // Implementation
}
```

**WRONG**:

```go
// Order state actor
type OrderStateActor struct {
    act.Actor
    orders map[string]*domain.Order  // orders map
}

// Check transition
func (o *Order) CanTransition(newStatus OrderStatus) bool {
    // Check if can transition
}
```

### 2. Package Documentation

**RULE**: Every package MUST have package comment  
**RULE**: Package comment should explain package purpose  
**RULE**: Include usage examples in package doc

**CORRECT**:

```go
// Package order implements order management using CQRS pattern.
//
// The package separates order commands (writes) from order queries (reads)
// using two separate actors:
//   - OrderCommandActor: Handles order placement and cancellation
//   - OrderStateActor: Maintains order state and handles queries
//
// Example usage:
//
//     cmdActor := order.NewCommandActor("binance", driver)
//     stateActor := order.NewStateActor("binance", persistence)
//
package order
```

### 3. Function Documentation

**RULE**: Document expected inputs and outputs  
**RULE**: Document error conditions  
**RULE**: Document concurrency safety

**CORRECT**:

```go
// Subscribe subscribes to the given topic and returns an unsubscribe function.
// The handler will be called for each event matching the topic pattern.
//
// Topic patterns support wildcards:
//   - "ticker.*.*" matches all tickers
//   - "ticker.binance.*" matches all Binance tickers
//
// The returned unsubscribe function is safe to call multiple times and
// is safe for concurrent use.
//
// Returns error if connector is not running or topic pattern is invalid.
func (c *Connector) Subscribe(topic string, handler HandlerFunc) (unsubscribe func(), err error)
```

---

## Common Patterns

### 1. CQRS Pattern

**Application**: Order management

**CORRECT**:

```go
// Write path - Commands
type OrderCommandActor struct {
    act.Actor
    driver driver.Driver
}

func (oc *OrderCommandActor) HandleCall(from gen.PID, ref gen.Ref, request interface{}) (interface{}, error) {
    switch req := request.(type) {
    case *message.PlaceOrder:
        order, err := oc.driver.PlaceOrder(req.Request)
        if err != nil {
            return nil, err
        }
        // Publish event
        gen.SendEvent("order.placed", order)
        return order, nil
    }
}

// Read path - Queries
type OrderStateActor struct {
    act.Actor
    orders map[string]*domain.Order
}

func (os *OrderStateActor) HandleEvent(msg gen.MessageEvent) {
    switch evt := msg.Message.(type) {
    case *domain.Order:
        os.orders[evt.ID] = evt
    }
}

func (os *OrderStateActor) HandleCall(from gen.PID, ref gen.Ref, request interface{}) (interface{}, error) {
    switch req := request.(type) {
    case *message.GetOrder:
        return os.orders[req.OrderID], nil
    }
}
```

### 2. Circuit Breaker Pattern

**Application**: Health monitoring

**CORRECT**:

```go
type CircuitBreaker struct {
    state         CircuitState
    failures      int
    successes     int
    lastFailure   time.Time
    config        Config
    mutex         sync.RWMutex
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mutex.RLock()
    state := cb.state
    cb.mutex.RUnlock()

    if state == CircuitOpen {
        if time.Since(cb.lastFailure) > cb.config.Timeout {
            cb.setState(CircuitHalfOpen)
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

### 3. Publisher-Subscriber Pattern

**Application**: Event distribution

**CORRECT**:

```go
type EventRouter struct {
    subscriptions sync.Map  // topic -> []gen.PID
}

func (r *EventRouter) Subscribe(topic string, pid gen.PID) func() {
    subs, _ := r.subscriptions.LoadOrStore(topic, &[]gen.PID{})
    subscribers := subs.(*[]gen.PID)
    *subscribers = append(*subscribers, pid)

    return func() {
        // Unsubscribe
        r.subscriptions.Delete(topic)
    }
}

func (r *EventRouter) Publish(topic string, event interface{}) {
    subs, ok := r.subscriptions.Load(topic)
    if !ok {
        return
    }

    for _, pid := range *subs.(*[]gen.PID) {
        gen.SendEvent(pid, topic, event)
    }
}
```

---

## Anti-Patterns to Avoid

### 1. God Object

**WRONG**:

```go
type ExchangeManager struct {
    // Too many responsibilities
    drivers map[string]driver.Driver
    orders map[string]*Order
    positions map[string]*Position
    websockets map[string]*WebSocket
    rateLimiters map[string]*RateLimiter
    // ... etc
}
```

**CORRECT**: Split into separate actors with single responsibility

### 2. Callback Hell

**WRONG**:

```go
driver.OnTicker(func(ticker *Ticker) {
    process(ticker, func(result Result) {
        store(result, func(err error) {
            if err != nil {
                log.Error(err)
            }
        })
    })
})
```

**CORRECT**: Use actors and message passing

### 3. Premature Optimization

**WRONG**:

```go
// Complex caching before profiling
type OrderCache struct {
    l1Cache map[string]*Order
    l2Cache *lru.Cache
    ttl     map[string]time.Time
    mutex   sync.RWMutex
}
```

**CORRECT**: Start simple, optimize based on profiling

### 4. Ignoring Context

**WRONG**:

```go
func (c *Connector) Stop() error {
    c.node.Stop()  // No graceful shutdown
    return nil
}
```

**CORRECT**:

```go
func (c *Connector) Stop() error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    return c.shutdown(ctx)
}
```

---

## Pre-Commit Checklist

Before committing code, verify:

- [ ] All tests pass (`go test ./...`)
- [ ] No linting errors (`golangci-lint run`)
- [ ] No vet errors (`go vet ./...`)
- [ ] Code formatted (`gofmt -s -w .`)
- [ ] Imports organized (`goimports -w .`)
- [ ] Documentation updated
- [ ] Error handling complete
- [ ] No sensitive data in code
- [ ] Concurrency safety verified
- [ ] Performance considerations reviewed

---

## Final Rules Summary

1. **Actor Isolation**: Never share state, always use message passing
2. **Decimal Precision**: Always use apd.Decimal for financial data
3. **Error Handling**: Always wrap errors with context
4. **Concurrency**: Protect shared state with mutexes
5. **Testing**: Write tests before implementation
6. **Documentation**: Document all exported APIs
7. **Logging**: Use structured logging with context
8. **Security**: Never log sensitive data
9. **Performance**: Minimize allocations in hot paths
10. **Simplicity**: Keep it simple, optimize later

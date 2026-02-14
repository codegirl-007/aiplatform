
# Algorithmic Trading Platform Invariants

These invariants **MUST** hold for all trading operations.  
Violation of any invariant is a **critical bug** and may result in financial loss.

## Enforcement Categories

- **[REPLAY]** – Enforced during replay of events  
- **[EXEC]** – Enforced at execution / event-emission time  
  (Replay may not be able to verify)

## Example Strategy Timeline
```
strategy.started

    data_ingestion phase:
      bar.received (timestamp 1)
      bar.received (timestamp 2)
      
    signal_generation phase:
      signal.generated (BUY signal)
      
    risk_validation phase:
      risk.validated (position size OK)
      
    order_execution phase:
      order.submitted (attempt 1)
      order.acknowledged
      order.filled
      
strategy.finished
```

---

## Strategy Invariants

### 1. Strategy ID Uniqueness
- Strategy IDs are UUID v4 strings.
- StrategyID must be unique within the strategy registry / event-log namespace.
- **[EXEC]** Enforced at creation time.
- No two strategies share the same StrategyID.

### 2a. Strategy Lifecycle Is Well-Formed (Event-Based)
- **[REPLAY]**
- For each `strategy_id`, the first event (lowest `seq`) must be `strategy.started`.
- For each `strategy_id`, there must be exactly one terminal event: `strategy.finished` **or** `strategy.failed`.
- The terminal event must be the last event (highest `seq`) for the `strategy_id`.
- No events may exist for that `strategy_id` with `seq` greater than the terminal event.
- Replay must fail with a clear error if any of the following occur:
    1) Missing start - no `strategy.started` event exists.
    2) Duplicate strategy start - more than one `strategy.started` event exists.
    3) Start not first - a `strategy.started` event exists but another event has a lower `seq`.
    4) Missing termination - No terminal event exists.
    5) Duplicate termination - More than one terminal event exists.
    6) Termination not last - A terminal event exists, but another event has a higher `seq`.

### 2b. Illegal Lifecycle Events
- **[EXEC]** 
- The runtime must not emit:
    1) a second `strategy.started`
    2) more than one terminal event
    3) any events after a terminal event

### 2c. Strategy State is Derived
- pending: no `strategy.started` seen
- running: `strategy.started` seen, no terminal case
- completed: `strategy.finished` seen
- failed: `strategy.failed` seen

### 3. Phase Execution Order
- **[REPLAY]** 
- Let phase order be `data_ingestion=1 signal_generation=2 risk_validation=3 order_execution=4`
- For each `strategy_id`, phase events must be non-decreasing in `seq`.
- Phase gating:
    1) No `signal_generation` events unless there exists a `data_ingestion` event with lower `seq`.
    2) No `risk_validation` events unless there exists a `signal_generation` event with lower `seq`.
    3) No `order_execution` events unless there exists a `risk_validation.validated` event with lower `seq`.

### 4a. Broker Configuration Validity
- **[EXEC]** 
- `broker_config` must specify a valid broker (etrade, alpaca, etc.)
- Credentials must be present and non-empty
- Environment (sandbox/production) must be explicitly specified
- Strategy creation fails if broker configuration is invalid

### 4b. Symbol Validation
- **[EXEC]**
- All symbols must be uppercase alphanumeric strings
- Minimum 1 character, maximum 10 characters
- No special characters except hyphens for multi-class shares (e.g., BRK-B)

### 5a. Risk Limits Enforced
- **[EXEC]** 
- No order may be submitted if risk validation fails
- Risk validation must check: position size, buying power, concentration limits
- If risk validation fails, strategy must terminate with `strategy.failed`

### 5b. Circuit Breakers
- **[EXEC]**
- If daily loss exceeds `max_daily_loss`, no new orders may be submitted
- If volatility exceeds `max_volatility_threshold`, strategy must pause
- Circuit breaker state is per-strategy, not global

---

## Order Invariants

### 8. Order ID Uniqueness
- Order IDs are UUID v4 strings.
- Unique within a strategy.
- **[EXEC]** Enforced at creation time.

### 9. Order Lifecycle Well-Formed
- **[REPLAY]** Enforced during replay.
- Exactly one `order.submitted` event per order.
- Exactly one of: `order.filled`, `order.cancelled`, `order.rejected`.
- `order.acknowledged` (if present) must occur after `order.submitted`.
- `order.filled` must occur after `order.submitted`.
- No subsequent events may reference the same `order_id` after termination.

### 10. Order Belongs to One Strategy
- **[REPLAY]** Enforced during replay.
- `Order.StrategyID` must match a valid StrategyID.

### 11. Order Symbol Validity
- **[EXEC]** Enforced before `order.submitted`.
- Symbol must be in the strategy's allowed symbols list.
- Symbol must be actively trading (not halted).

### 12. Order Side Validity
- **[EXEC]** Must be exactly one of: `buy`, `sell`, `sell_short`, `buy_to_cover`.

### 13. Order Type Validity  
- **[EXEC]** Must be one of: `market`, `limit`, `stop`, `stop_limit`.
- Limit orders must specify `limit_price > 0`.
- Stop orders must specify `stop_price > 0`.
- Stop-limit orders must specify both prices.

### 14. Order Quantity Constraints
- **[EXEC]** Enforced before `order.submitted`.
- `quantity > 0` (must be positive).
- `quantity ≤ max_order_quantity` (configured per strategy).
- `quantity` must be a multiple of `lot_size` for the symbol.

### 15. Order Time Validity
- **[EXEC]** Enforced before `order.submitted`.
- `time_in_force` must be one of: `day`, `gtc`, `ioc`, `fok`.
- Expired orders (GTC past expiration) must not be submitted.

---

## Position Invariants

### 16. Position Quantity Validity
- **[EXEC]** Enforced on every position update.
- Long positions: `quantity ≥ 0`.
- Short positions: `quantity ≤ 0` (if short selling enabled).
- `abs(quantity) ≤ max_position_size`.

### 17. Position P&L Calculation
- **[EXEC]` Enforced after every fill.
- `unrealized_pnl` = (current_price - avg_cost) × quantity
- `realized_pnl` accumulates from closed trades
- P&L calculations must use consistent precision (2 decimal places for USD)

### 18. Position Belongs to One Strategy
- **[REPLAY]** Enforced during replay.
- `Position.StrategyID` must match a valid StrategyID.

### 19. Position Symbol Uniqueness
- **[EXEC]** Within a strategy, only one position per symbol.
- Multiple strategies may hold positions in the same symbol (isolated).

---

## Market Data Invariants

### 20. Bar Timestamp Monotonicity
- **[REPLAY]** Enforced during replay.
- Bar timestamps must be strictly increasing within a symbol.
- `bar[n].timestamp > bar[n-1].timestamp` for all n.

### 21. Price Validity
- **[EXEC]** Enforced on every tick/bar.
- `open, high, low, close > 0`.
- `high ≥ low`.
- `high ≥ open` and `high ≥ close`.
- `low ≤ open` and `low ≤ close`.
- `volume ≥ 0`.

### 22. Market Data Symbol Validity
- **[EXEC]** Symbol must be in the strategy's subscription list.
- Strategy must not receive data for unsubscribed symbols.

### 23. No Future Data
- **[EXEC]** System timestamp must be ≥ bar timestamp.
- No processing bars with future timestamps.

---

## Signal Invariants

### 24. Signal ID Uniqueness
- Signal IDs are UUID v4 strings.
- Unique within a strategy.
- **[EXEC]** Enforced at creation time.

### 25. Signal Lifecycle
- **[REPLAY]** Exactly one `signal.generated` per signal.
- Signal must be consumed (converted to order or discarded) within `signal_ttl_ms`.
- Stale signals must not be acted upon.

### 26. Signal Direction Validity
- **[EXEC]** Must be one of: `buy`, `sell`, `hold`.
- `hold` signals result in no action.

### 27. Signal Confidence Validity
- **[EXEC]** `0.0 ≤ confidence ≤ 1.0`.
- Signals below `min_confidence_threshold` are discarded.

### 28. Signal Belongs to One Strategy
- **[REPLAY]** `Signal.StrategyID` must match a valid StrategyID.

---

## Risk Management Invariants

### 29. Position Size Limit
- **[EXEC]** `position_value ≤ max_position_value`.
- `position_value = quantity × current_price`.
- Checked before order submission.

### 30. Buying Power Check
- **[EXEC]** Before order submission:
  - For buy orders: `order_value ≤ available_buying_power`.
  - `order_value = quantity × (limit_price or current_market_price)`.

### 31. Concentration Limit
- **[EXEC]** `position_value / portfolio_value ≤ max_concentration_pct`.
- Checked before order submission.

### 32. Daily Loss Limit
- **[EXEC]** If `daily_realized_pnl < -max_daily_loss`, strategy pauses.
- Checked at start of each decision cycle.

### 33. Volatility Circuit Breaker
- **[EXEC]** If `volatility_14d > max_volatility_threshold`, no new positions.
- Volatility measured as standard deviation of returns.

---

## Event Log Invariants

### 34. Append-Only Log
- **[EXEC]** Enforced by implementation.
- Events are never modified or deleted.
- Events are written in `seq` order.

### 35. Event ID Uniqueness
- UUID v4 strings.
- **[EXEC]** Enforced at creation time.

### 36. Valid Event Types
- **[REPLAY]** Enforced during replay.
- Allowed types:
  - `strategy.started`, `strategy.finished`, `strategy.failed`
  - `bar.received`
  - `signal.generated`
  - `risk.validated`, `risk.limit_breached`
  - `order.submitted`, `order.acknowledged`, `order.filled`, `order.partial_fill`, `order.cancelled`, `order.rejected`
  - `position.opened`, `position.closed`, `position.updated`
  - `portfolio.rebalanced`
  - `llm.requested`, `llm.responded`, `llm.failed`
  - `llm.action_generated`
  - `approval.approved`, `approval.rejected`, `approval.expired`
  - `system.emergency_stop`

### 37. Sequence Ordering
- **[REPLAY]** Enforced during replay.
- `seq` strictly increases within a strategy.

### 38. Payload Validation
- **[REPLAY]** Enforced during replay.
- Payload must match event schema.

### 39. JSONLines Format
- **[EXEC]** Enforced by implementation.
- One valid JSON object per line.
- Lines terminated by newline.

### 40. Replayability
- **[REPLAY]** Guaranteed by replay engine.
- StrategyView can be fully reconstructed.
- No state outside the event log.

---

## Broker Invariants

### 41. Broker Client Validity
- **[EXEC]** Broker client must be initialized with valid credentials.
- OAuth tokens must be refreshed before expiration.
- Failed authentication must prevent order submission.

### 42. Order Idempotency
- **[EXEC]** Orders include `idempotency_key` to prevent duplicates.
- Broker must reject duplicate submissions within `idempotency_window`.

### 43. Order Acknowledgment Timeout
- **[EXEC]** If `order.acknowledged` not received within `ack_timeout_ms`, mark as failed.
- Timeout must trigger position reconciliation.

### 44. Fill Reporting
- **[EXEC]** Every `order.filled` event must include:
  - `fill_price` (execution price)
  - `fill_quantity` (shares/contracts filled)
  - `fill_time` (execution timestamp)
  - `commission` (fees paid)

---

## Portfolio Invariants

### 45. Cash Balance Validity
- **[EXEC]** `cash_balance ≥ 0`.
- Cash balance updated on every fill and deposit/withdrawal.

### 46. Total Portfolio Value
- **[EXEC]** `total_value = cash_balance + sum(position_values)`.
- Calculated consistently across all operations.

### 47. Strategy Isolation
- **[EXEC]** Each strategy has isolated positions and P&L.
- Strategies cannot see or affect each other's positions.

---

## Replay Invariants

### 48. Deterministic Replay
- **[REPLAY]** Guaranteed by replay engine.
- Same events → same StrategyView.

### 49. Invariant Validation During Replay
- **[REPLAY]** Guaranteed by replay engine.
- Errors include `seq`, event type, and reason.

### 50. Complete StrategyView Reconstruction
- **[REPLAY]** Guaranteed by replay engine.
- All positions, orders, signals, and market data are restored.

---

## LLM Integration Invariants

### 51. LLM Provider Validity
- **[EXEC]** Provider must be one of: `openai`, `anthropic`.
- API key must be non-empty and valid for the provider.
- Model must be supported by the provider (e.g., `gpt-4`, `claude-3-opus`).

### 52. LLM Session Isolation
- **[EXEC]** Each strategy has its own LLM session/context.
- LLM cannot access data from other strategies.
- Context window must not exceed provider limits.

### 53. LLM Request/Response Lifecycle
- **[REPLAY]** Exactly one `llm.requested` per LLM call.
- Exactly one of: `llm.responded`, `llm.failed`.
- `llm.failed` includes error reason (timeout, rate limit, content filter).
- Response must be parseable as valid trading decision.

### 54. LLM Decision Constraints
- **[EXEC]** LLM output must conform to structured format:
  - `action`: one of `buy`, `sell`, `hold`, `cancel_order`, `get_portfolio`, `get_quote`
  - `symbol`: required for buy/sell actions
  - `quantity`: required for buy/sell, must be positive integer
  - `order_type`: `market` or `limit`
  - `reasoning`: human-readable explanation
- Invalid format results in `llm.failed` with parse error.

### 55. LLM Rate Limiting
- **[EXEC]** Maximum `max_llm_requests_per_minute` requests to provider.
- Exceeding rate limit results in `llm.failed` with rate limit error.
- Must implement exponential backoff for retries.

### 56. LLM Cost Tracking
- **[EXEC]** Each `llm.requested` logs estimated cost (input/output tokens).
- Accumulated cost tracked per strategy.
- Strategy pauses if `total_llm_cost > max_llm_cost_budget`.

---

## Approval Mode Invariants

### 57. Mode Configuration Validity
- **[EXEC]** Mode must be exactly one of:
  - `MODE_APPROVAL_REQUIRED` (Mode 1): User must approve every action
  - `MODE_AUTONOMOUS` (Mode 2): LLM acts without approval
- Mode set at strategy start and immutable thereafter.
- Mode stored in `strategy.started` event payload.

### 58. Approval Queue Management (Mode 1)
- **[EXEC]** All LLM-generated actions require approval.
- Action enters `approval.pending` state upon generation.
- User must explicitly approve (`approval.approved`) or reject (`approval.rejected`).
- Action expires after `approval_timeout_ms` if not approved/rejected.
- Expired actions are treated as rejected.
- Only approved actions may proceed to execution.

### 59. Approval Event Lifecycle
- **[REPLAY]** Exactly one `llm.action_generated` per action.
- Exactly one of: `approval.approved`, `approval.rejected`, `approval.expired`.
- `approval.approved` must precede `order.submitted` (same `action_id`).
- No `order.submitted` without prior `approval.approved` in Mode 1.

### 60. Autonomous Mode Safeguards (Mode 2)
- **[EXEC]** Even in autonomous mode, hard limits apply:
  - Daily loss limit (Invariant 32)
  - Position size limit (Invariant 29)
  - Buying power check (Invariant 30)
  - Maximum orders per hour (`max_orders_per_hour_autonomous`)
- Autonomous mode pauses (not fails) if limits approached, resumes LLM asks for permission.
- Emergency stop button immediately halts all autonomous actions.

### 61. Emergency Stop
- **[EXEC]** Emergency stop event `system.emergency_stop` immediately:
  - Cancels all pending orders
  - Prevents new orders
  - Halts LLM processing
  - Marks strategy as `strategy.failed` with `EMERGENCY_STOP` reason
- Emergency stop can be triggered by user at any time.
- Strategy cannot resume after emergency stop (must restart).

### 62. Action History Logging
- **[EXEC]** Every action (approved, rejected, or autonomous) logged with:
  - LLM reasoning
  - Proposed action details
  - User decision (if Mode 1) or autonomous flag
  - Timestamp
- History available for audit and replay.

### 63. Mode Switching Prohibition
- **[EXEC]** Mode cannot change during strategy execution.
- To change mode, strategy must be terminated and restarted.
- Prevents mixing approved and autonomous actions in same strategy run.

### 64. User Notification
- **[EXEC]** In Mode 1, user must be notified within `notification_timeout_ms`:
  - Desktop notification via Wails
  - Action details displayed in UI
  - Audio alert (optional)
- Notification failure pauses strategy until user checks UI.

---

## Safety Limits (Both Modes)

### 65. Maximum Order Value
- **[EXEC]** Single order value ≤ `max_order_value` (e.g., $100,000).
- Prevents LLM from placing oversized bets.

### 66. Forbidden Symbols
- **[EXEC]** Strategy cannot trade symbols in `forbidden_symbols` list.
- Default forbidden: leveraged ETFs, penny stocks (< $1), crypto (if not supported).

### 67. Trading Hours
- **[EXEC]** Orders only submitted during market hours (9:30 AM - 4:00 PM ET).
- Pre-market/after-hours trading disabled by default.
- Can be enabled per-strategy with `extended_hours: true`.

### 68. Cooldown Period
- **[EXEC]** Minimum `cooldown_ms` between consecutive actions.
- Prevents rapid-fire trading by LLM.
- Cooldown enforced per-symbol and globally.

### 69. Duplicate Action Prevention
- **[EXEC]** LLM cannot propose identical action within `duplicate_window_ms`.
- Identical = same symbol, side, quantity, order type.
- Prevents LLM from getting stuck in a loop.

### 70. Human Override
- **[EXEC]** User can cancel any pending action (Mode 1) or ongoing autonomous action (Mode 2).
- Cancel takes effect immediately.
- Cancel reason logged.

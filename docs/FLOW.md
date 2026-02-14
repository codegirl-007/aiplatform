# User Journey: LLM-Powered Algo Trading Platform

## Glossary: Domain Concepts → Engine Model

**Strategy Execution** (domain concept) maps to:
- **Engine `run_id`** - Each strategy execution is a run with unique UUID
- **Engine phases** - Execution progresses through phases: `data_ingestion` → `signal_generation` → `risk_validation` → `order_execution`
- **Event log** - All strategy actions captured as immutable events (JSONL format)

**Phase Progression:**
- Phase 1: `data_ingestion` - Collect market data, bars, quotes
- Phase 2: `signal_generation` - LLM analysis, signal creation
- Phase 3: `risk_validation` - Position size, buying power, concentration checks
- Phase 4: `order_execution` - Submit orders to broker, track fills

**Engine Events Used:**
- `run.started`, `run.finished`, `run.failed` - Run lifecycle
- `step.started`, `step.finished`, `step.failed` - Step execution within phases
- `llm.requested`, `llm.responded` - LLM calls
- `tool.called`, `tool.returned`, `tool.failed` - Tool/broker API calls
- Trading events: `bar.received`, `signal.generated`, `order.submitted`, `order.filled`, etc.

For engine invariants, see `docs/ALGO.md`. For trading invariants, see `docs/TRADING.md`.

## **1. Initial Setup**

**Connect Broker:**
- User enters ETrade API credentials (sandbox or production)
- App validates OAuth connection
- Account list displayed (823145980, 583156360, etc.)

**Configure LLM:**
- Choose provider: OpenAI (GPT-4) or Anthropic (Claude)
- Enter API key
- Set cost budget (e.g., $10/day max)

**Safety Configuration:**
- Set daily loss limit (e.g., -$1,000)
- Max position size (e.g., $50,000)
- Max order value (e.g., $100,000)
- Forbidden symbols (leveraged ETFs, penny stocks)
- Trading hours (market hours vs extended)

---

## **2. Creating a Strategy**

**Define Strategy:**
- Select symbols to trade (e.g., AAPL, MSFT)
- Choose mode: **Mode 1** (Approval Required) or **Mode 2** (Autonomous)
- Set data frequency (1-min bars, 5-min bars)
- Configure LLM prompt template

**Example Mode 1 Setup:**
```
Strategy: "Mean Reversion on Tech Stocks"
Symbols: AAPL, MSFT, GOOGL
Mode: APPROVAL_REQUIRED
Prompt: "Analyze price action and suggest trades when RSI < 30"
Risk Limits: Max position $10k per symbol
Approval Timeout: 5 minutes
```

**Example Mode 2 Setup:**
```
Strategy: "Autonomous Momentum Trading"
Symbols: SPY, QQQ
Mode: AUTONOMOUS  
Max Orders/Hour: 10
Hard Limits: Same as Mode 1
```

---

## **3. Mode 1: Approval Required (User-in-the-Loop)**

**Live Trading Flow:**

1. **Market Data Arrives** (Phase 1: `data_ingestion`)
   - App receives 1-min bar: AAPL $175.50, volume 1.2M
   - Event: `bar.received`

2. **LLM Analysis** (Phase 2: `signal_generation`)
   - App sends prompt to GPT-4 with price data
   - Events: `llm.requested` → `llm.responded`
   - LLM responds: "BUY AAPL - RSI oversold at 28"

3. **Action Generated**
   - Action queued: Buy 100 shares AAPL @ market
   - Event: `llm.action_generated`
   - **UI Notification:** Desktop alert + sound

4. **User Approval Screen (Wails UI)** (Phase 3: `risk_validation`)
   ```
   🔵 NEW TRADE PROPOSAL
   
   Strategy: Mean Reversion on Tech Stocks
   Action: BUY AAPL
   Quantity: 100 shares
   Estimated Value: $17,550
   
   LLM Reasoning:
   "RSI at 28 indicates oversold condition. 
   Price bounced off 200-day MA support."
   
   Risk Check: ✅ Position size OK
                ✅ Buying power OK
                ✅ Under daily loss limit
   
   [APPROVE] [REJECT]
   (Expires in 4:52)
   ```

5. **User Decision:** (Phase 4: `order_execution` if approved)
   - **Approve:** Order submitted to ETrade → `order.submitted` → `order.filled` → Position opened
   - **Reject:** Action discarded, logged in history
   - **Timeout:** Treated as reject after 5 min

6. **Audit Trail:**
   - Every action logged with LLM reasoning
   - User decision recorded
   - Complete history for review

---

## **4. Mode 2: Autonomous Trading**

**Live Trading Flow:**

1. **Continuous Monitoring**
   - App ingests market data continuously
   - No user intervention for individual trades

2. **LLM Decision Loop**
   ```
   Every minute:
   - Fetch latest bars
   - Send to LLM: "Analyze and trade if opportunity > 0.8 confidence"
   - LLM returns structured action
   - Risk validation runs automatically
   - Order submitted immediately (no approval)
   ```

3. **Safeguards Active:**
   - **Daily Loss Limit:** If down $1,000 → Pause and ask user permission to continue
   - **Max Orders/Hour:** If 10 orders reached → Pause until next hour
   - **Position Size:** Any order > $10k rejected automatically
   - **Circuit Breaker:** High volatility → Pause trading

4. **User Monitoring:**
   - Real-time P&L dashboard
   - Open positions view
   - Recent trades list
   - LLM cost tracker ($2.34 spent today)

5. **Emergency Controls:**
   - **Emergency Stop Button:** Instant halt, cancels all orders
   - **Pause Strategy:** Temporary stop (can resume)
   - **Edit Limits:** Adjust risk parameters mid-flight

---

## **5. Safety Features (Both Modes)**

**Hard Limits (Cannot Override):**
- ❌ No orders > $100k (Invariant T100)
- ❌ No penny stocks (< $1) (Invariant T101)
- ❌ No trading outside 9:30 AM - 4:00 PM ET (Invariant T102)
- ❌ No duplicate actions within 1 minute (Invariant T104)
- ❌ 30-second cooldown between trades (Invariant T103)

**Risk Checks (Every Order):**
- Position size ≤ limit
- Buying power sufficient
- Portfolio concentration ≤ 20%
- Daily loss < -$1,000

**Circuit Breakers:**
- Daily loss exceeded → Strategy pauses
- Volatility spike → No new positions
- Broker API error → Retry with backoff

---

## **6. Emergency Procedures**

**Emergency Stop (Big Red Button):**
- Immediately cancels all pending orders
- Halts all LLM processing
- Closes strategy with `EMERGENCY_STOP` status
- **Cannot resume** - must create new strategy

**When to Use:**
- Market crash
- Strategy going rogue
- Unexpected behavior
- User panic

**Recovery:**
- Review event log to see what happened
- Create new strategy with adjusted parameters
- Manual reconciliation with broker

---

## **7. Typical User Sessions**

**Conservative User (Mode 1):**
```
9:30 AM - Start strategy, monitor dashboard
9:35 AM - Notification: LLM wants to buy AAPL
          Review reasoning, approve trade
9:40 AM - Fill notification: Bought 100 AAPL @ $175.50
10:00 AM - Notification: LLM wants to sell MSFT
          Reject (don't agree with analysis)
4:00 PM - Strategy auto-finishes
          Review 5 trades executed, +$234 profit
          LLM cost: $1.23
```

**Aggressive User (Mode 2):**
```
9:30 AM - Start autonomous strategy with tight limits
          Max 5 trades/hour, $5k max position
9:30-4:00 - Strategy runs autonomously
            Check dashboard periodically
            23 trades executed
            +$1,234 profit
            1 circuit breaker pause at 11 AM (volatility)
            LLM cost: $4.56
4:05 PM - Review trade history
          All trades within limits
          Happy with performance
```

---

## **8. Event Log Benefits**

**Full Audit Trail:**
Every action recorded with sequence numbers (see `docs/ALGO.md` Invariant 38):
```json
{"seq":1,"type":"run.started","run_id":"run-xxx","workspace_root":"/data/strategies"}
{"seq":2,"type":"step.started","step_id":"step-yyy","phase":"data_ingestion"}
{"seq":3,"type":"bar.received","symbol":"AAPL","close":175.50}
{"seq":4,"type":"step.finished","step_id":"step-yyy","phase":"data_ingestion"}
{"seq":5,"type":"step.started","step_id":"step-zzz","phase":"signal_generation"}
{"seq":6,"type":"llm.requested","prompt":"..."}
{"seq":7,"type":"llm.responded","action":"buy","symbol":"AAPL"}
{"seq":8,"type":"step.finished","step_id":"step-zzz","phase":"signal_generation"}
{"seq":9,"type":"step.started","step_id":"step-aaa","phase":"risk_validation"}
{"seq":10,"type":"llm.action_generated","action_id":"..."}
{"seq":11,"type":"approval.approved"} // Mode 1 only
{"seq":12,"type":"step.finished","step_id":"step-aaa","phase":"risk_validation"}
{"seq":13,"type":"step.started","step_id":"step-bbb","phase":"order_execution"}
{"seq":14,"type":"order.submitted","order_id":"..."}
{"seq":15,"type":"order.filled","fill_price":175.50}
{"seq":16,"type":"step.finished","step_id":"step-bbb","phase":"order_execution"}
{"seq":17,"type":"run.finished","run_id":"run-xxx"}
```

**Replay Capability:**
- Can replay any strategy session (engine guarantee - `docs/ALGO.md` Invariant 41)
- Debug what the LLM was thinking
- Prove compliance with invariants
- Backtest new strategies on historical data

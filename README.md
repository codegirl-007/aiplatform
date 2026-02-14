# AI-Powered Algorithmic Trading Platform

A Wails-based desktop application for LLM-powered algorithmic trading, combining a Tiger Beetle-inspired event sourcing engine with broker integration and intelligent decision-making.

## Overview

This platform provides a general-purpose runtime engine with a trading layer built on top:

- **Runtime Engine**: Event-sourced execution with phases, steps, and deterministic replay
- **Trading Layer**: Strategy management, order execution, risk validation, and portfolio tracking
- **LLM Integration**: AI-powered decision making with approval workflows
- **Broker Support**: ETrade integration (OAuth client in development - see COD-12)

## Architecture

### Backend (Go)
- **Event Sourcing**: Append-only JSONL logs, perfect audit trail, crash recovery
- **Single-Threaded State Mutation**: No mutexes, command channels only
- **Phase-Based Execution**: `data_ingestion` → `signal_generation` → `risk_validation` → `order_execution`
- **Invariant-Driven Development**: Runtime and trading invariants enforced at execution and replay time

### Frontend (Vite + Vanilla JS)
- Wails v2 desktop application
- Real-time strategy monitoring
- Approval workflows for LLM decisions
- Emergency stop controls

## Project Structure

```
internals/          Go backend packages
  runtime/          Event sourcing engine, phase management
  clients/          ETrade API client with OAuth
cmd/                Command-line utilities and test tools
pkg/                Shared utility packages (assert, validation)
frontend/           Vite-powered web UI
docs/               Documentation
  development.md    Engineering principles and standards
  ALGO.md           Runtime engine invariants
  TRADING.md        Trading-domain invariants
  FLOW.md           User journey and workflows
.env.example        Environment variable template
```

## Development

### Prerequisites
- Go 1.21+
- Node.js 18+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Setup

1. Clone the repository
2. Copy `.env.example` to `.env` and configure your credentials:
   ```bash
   ETRADE_CONSUMER_KEY=your_sandbox_key
   ETRADE_CONSUMER_SECRET=your_sandbox_secret
   ```
3. Install dependencies:
   ```bash
   cd frontend && npm install
   ```

### Live Development

Run with hot reload:
```bash
wails dev
```

The dev server runs on `http://localhost:34115` for calling Go methods from the browser.

### Building

Create production build:
```bash
wails build
```

### Testing

```bash
# Run all tests
go test ./...

# Run specific package tests with verbose output
go test ./internals/runtime/... -v

# Run ETrade OAuth test utility
go run ./cmd/etrade-oauth-test
```

## Documentation Map

- **`docs/development.md`** - Engineering principles, code standards, testing strategy (Tiger Beetle-inspired)
- **`docs/ALGO.md`** - Runtime engine invariants (runs, steps, phases, event log)
- **`docs/TRADING.md`** - Trading-domain invariants (strategy, orders, positions, risk, approval, LLM)
- **`docs/FLOW.md`** - User journey and trading workflow documentation
- **`AGENTS.md`** - AI agent instructions for development

## Core Principles

### Zero Technical Debt Policy
Do it right the first time. No potential latency spikes, no exponential algorithms. What we have meets design goals, even if incomplete.

### Invariant-Driven Development
Define what MUST always be true, then design so invariants are enforced or unrepresentable. Code is the implementation of proofs.

### Event Sourcing
Events are the source of truth, state is a cache. All state changes captured as immutable events in append-only JSONL logs.

### Fail-Fast Design
Assertions everywhere (minimum 2 per function). Panic on programmer errors. Hard limits on everything (70 lines per function, bounded loops, fixed-capacity channels).

## Features

### Trading Modes
- **Mode 1: Approval Required** - User approves every LLM-generated action
- **Mode 2: Autonomous** - LLM acts independently within hard limits

### Risk Management
- Daily loss limits
- Position size limits
- Concentration limits
- Volatility circuit breakers
- Emergency stop button

### Audit & Compliance
- Complete event log with sequence numbers
- Deterministic replay capability
- LLM reasoning captured for every decision
- User approval history tracked

## License

[Add license information]

## Contributing

See `docs/development.md` for code standards and contribution guidelines.

# AI Trading Platform - OpenCode Agent Instructions

## Project Overview

This is a **Wails-based desktop application** for LLM-powered algorithmic trading, combining:
- **Backend**: Go with Tiger Beetle-inspired event sourcing architecture
- **Frontend**: Vite + vanilla JavaScript (Wails v2)
- **Integration**: ETrade OAuth client for broker connectivity
- **Philosophy**: Zero technical debt, invariant-driven development, fail-fast design

## Project Structure

```
internals/          Go backend packages (runtime, clients)
  runtime/          Event sourcing engine, phase management
  clients/          ETrade API client with .env credential loading
cmd/                Command-line utilities and test tools
pkg/                Shared utility packages (assert, validation)
frontend/           Vite-powered web UI
.env.example        Authoritative list of environment variables
```

## Core Documentation (Read on Demand)

When working on this project, familiarize yourself with these files as needed:

- **`docs/development.md`** - Engineering principles, code standards, testing strategy (Tiger Beetle-inspired)
- **`docs/ALGO.md`** - Runtime engine invariants (runs, steps, phases, event log)
- **`docs/TRADING.md`** - Trading-domain invariants (strategy, orders, positions, risk, approval, LLM)
- **`docs/FLOW.md`** - User journey and trading workflow documentation

**Precedence**: When rules conflict, `docs/ALGO.md` wins for "must" constraints; `docs/development.md` governs process and style.

## Engineering Principles (Summary)

Read `docs/development.md` for full details. Key principles:

### 1. Zero Technical Debt Policy
- Do it right the first time
- No potential latency spikes or exponential algorithms
- What we have meets design goals, even if incomplete

### 2. Invariant-Driven Development
- Define what MUST always be true
- Design so invariants are enforced or unrepresentable
- Code is implementation of proofs
- See `docs/ALGO.md` for engine invariants and `docs/TRADING.md` for trading invariants

### 3. Assertions Are Force Multipliers
- **Minimum 2 assertions per function** (average)
- **Place assertions at the top of functions** when possible (fail fast on invalid inputs)
- Assert all function arguments and return values
- Assert preconditions, postconditions, and invariants
- Pair assertions: validate at write time AND read time
- Use `pkg/assert` package (panics on violation)
- When a state is impossible and application cannot proceed, code should not return an error. We should assert and panic.

### 4. Event Sourcing Architecture
- Events are the source of truth, state is a cache
- All state changes captured as immutable events
- Append-only JSONL logs
- Write to log BEFORE updating in-memory state
- Perfect audit trail and replay capability

### 5. Single-Threaded State Mutation
- Only one goroutine mutates state (no mutexes)
- All state changes go through command channels
- Eliminates race conditions

### 6. Hard Limits
- 70 lines max per function
- 100 columns max line length
- No recursion - use explicit bounded loops
- Put a limit on everything (loops, queues, buffers, channels)

### 7. Code Quality Rules
- Zero values are invalid (Phase 0 = uninitialized)
- Explicit over clever (clear code > optimizations)
- All errors must be handled (never ignore with `_`)
- Standard library first, zero dependencies policy
- Minimize allocations (pre-allocate slices, use sync.Pool)

## Go Style Guidelines

- Use `snake_case` for unexported/internal identifiers, file names, JSON fields, and event type strings
- Exported identifiers remain Go-exported style (CamelCase) for public APIs
- No abbreviations in variable names (except i, j for loops)
- Units/qualifiers last: `latency_ms_max` not `max_latency_ms`
- Comments are sentences: capitalize, punctuate, explain WHY
- File order: constants → types → functions
- Run `go fmt` - use standard Go formatting, don't fight it

## Testing Strategy

### Test Naming
Name tests after invariants they verify:
```go
func TestInvariant_1_RunIDUniqueness(t *testing.T) {...}
func TestInvariant_4a_WorkspaceRootAbsolute(t *testing.T) {...}
```

### Test Requirements
- Test exhaustively: valid AND invalid data
- Tests must be deterministic
- Use `t.TempDir()` for test directories
- Don't rely on timing
- Property-based testing where appropriate

### Running Tests
```bash
go test ./...                      # All tests
go test ./internals/runtime/... -v # Verbose runtime tests
make test                          # If Makefile target exists
```

## Wails Development

```bash
wails dev    # Live development with hot reload
wails build  # Production build
```

Frontend uses Vite dev server on `http://localhost:34115` for Go method calls during development.

## Environment Variables & Secrets

### Critical Rules
- **NEVER commit secrets** to git (`.env` is gitignored)
- `.env.example` is the authoritative list of variable names
- Keep `.env.example` updated when code references new env vars

### Current Variables
```bash
ETRADE_CONSUMER_KEY=         # ETrade sandbox/production consumer key
ETRADE_CONSUMER_SECRET=      # ETrade sandbox/production consumer secret
```

## Git Workflow & PR Standards

### Commit Guidelines
- Create commits only when explicitly requested
- Commit messages must accurately describe changes
- Title format: `<verb> <what>` (e.g., "Add etrade client env loading")
- Include "why" in commit body when non-obvious

### PR Requirements
- **PR title/description must match the actual diff**
- If PR is just "fix .env.example variable names", don't title it "Add ETrade client"
- Include context: why this change, what problem it solves
- Keep PRs focused and small when possible
- Reference related issues/PRs

### Branch Hygiene
- Don't silently rewrite history unless explicitly asked
- Use descriptive branch names matching the work
- If rebasing/force-pushing, verify no important changes are lost

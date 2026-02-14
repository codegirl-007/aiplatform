# GitHub Copilot Review Instructions

## Project Context

This is a **Wails-based desktop application** for LLM-powered algorithmic trading with:
- **Backend**: Go with Tiger Beetle-inspired event sourcing architecture
- **Frontend**: Vite + vanilla JavaScript
- **Integration**: ETrade OAuth client for broker connectivity
- **Philosophy**: Zero technical debt, invariant-driven development, fail-fast design

## Critical Review Criteria

### 1. Zero Technical Debt Policy
- ❌ Flag any potential latency spikes or exponential algorithms
- ❌ Flag placeholder TODOs or "fix later" comments
- ✅ Accept incomplete features only if what exists meets design goals

### 2. Assertion Requirements (CRITICAL)
- ❌ **BLOCK** if average assertions per function < 2
- ❌ **BLOCK** if function arguments are not asserted
- ❌ **BLOCK** if return values are not validated
- ✅ Assertions should be at the top of functions (fail fast)
- ✅ Check for paired assertions (validate at write AND read time)
- ✅ Must use `pkg/assert` package (panics on violation)

### 3. Invariant Enforcement
- ❌ Flag any code that could violate invariants in `docs/ALGO.md`
- ❌ Flag zero values used as valid states (Phase 0 = uninitialized)
- ✅ Verify invariants are enforced or made unrepresentable in types

### 4. Event Sourcing Architecture
- ❌ **BLOCK** if state mutations don't write to log first
- ❌ **BLOCK** if events are not immutable
- ❌ Flag any in-memory state that's not derived from events
- ✅ Events must be append-only JSONL format

### 5. Concurrency Safety
- ❌ **BLOCK** if multiple goroutines can mutate shared state
- ❌ **BLOCK** if mutexes are used (use channels instead)
- ✅ State changes must go through command channels
- ✅ Single-threaded state mutation pattern only

### 6. Hard Limits (Auto-Reject Violations)
- ❌ **BLOCK** functions > 70 lines
- ❌ **BLOCK** lines > 100 columns
- ❌ **BLOCK** recursion (use explicit bounded loops)
- ❌ **BLOCK** unbounded loops, queues, buffers, or channels

### 7. Error Handling
- ❌ **BLOCK** if errors are ignored with `_`
- ❌ **BLOCK** if errors are swallowed without logging/handling
- ✅ All errors must be explicitly handled

### 8. Code Quality Standards
- ❌ Flag abbreviations in variable names (except i, j in loops)
- ❌ Flag external dependencies (standard library first, zero deps policy)
- ❌ Flag implicit allocations in hot paths
- ✅ Use `snake_case` for unexported/internal identifiers, file names, JSON fields, event type strings
- ✅ Exported identifiers remain Go-exported style (CamelCase)
- ✅ Units/qualifiers last: `latency_ms_max` not `max_latency_ms`
- ✅ Comments must explain WHY, not WHAT
- ✅ Run `go fmt` - use standard Go formatting

### 9. Testing Requirements
- ❌ **BLOCK** if tests are non-deterministic
- ❌ **BLOCK** if tests rely on timing or sleeps
- ❌ Flag missing tests for invalid input cases
- ✅ Test names should reference invariants: `TestInvariant_1_RunIDUniqueness`
- ✅ Use `t.TempDir()` for test directories

### 10. Security & Secrets
- ❌ **BLOCK** if secrets/keys are hardcoded
- ❌ **BLOCK** if `.env` file changes are committed
- ✅ Verify `.env.example` is updated when new env vars are referenced

## Review Workflow

1. **First Pass**: Check for auto-reject criteria (hard limits, missing assertions, concurrency violations)
2. **Second Pass**: Verify invariants and architectural patterns
3. **Third Pass**: Code quality, style, and testing

## Documentation References

Review these files when needed:
- `docs/development.md` - Full engineering principles and standards
- `docs/ALGO.md` - Runtime engine invariants (takes precedence for engine constraints)
- `docs/TRADING.md` - Trading-domain invariants (strategy, orders, positions, risk)
- `docs/FLOW.md` - User journey and trading workflow
- `AGENTS.md` - Complete agent instructions and project structure

## Severity Levels

- **BLOCK**: Must be fixed before merge (security, correctness, hard limits)
- **Flag**: Should be addressed (code quality, maintainability)
- **Suggest**: Optional improvements (style, clarity)

## When in Doubt

If architectural decisions are unclear, reference the Tiger Beetle design principles:
- Correctness over performance
- Simplicity over features
- Explicit over implicit
- Fail fast, fail loudly

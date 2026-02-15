# Project Progress

## Completed Work

### ✅ OAuth 1.0a Implementation (PR #10 - Merged)
- Implemented full OAuth 1.0a flow for ETrade API
- Fixed all 12 Copilot code review issues:
  1. Removed dead code in tests
  2. Fixed documentation field name mismatch
  3. Added assertions to URL helpers
  4. Wrapped long function declarations (100-column limit)
  5. Reused helper to eliminate duplicate code
  6. Changed panic to error return in ExchangeToken()
  7. Fixed ignored resp.Body.Read errors (2 instances)
  8. Implemented proper midnight US/Eastern token expiry
  9. Changed unsafe default to sandbox=true
  10. Updated help text to match sandbox default
  11. Removed TODO methods violating zero-debt policy
  12. Refactored main() to respect 70-line limit
- All 45 tests passing
- Token storage with atomic writes and 0600 permissions
- OAuth endpoints verified working (401 unauthorized = good, means endpoint exists)
- Test script: `go run cmd/etrade-oauth-test/main.go`

### ✅ Code Coverage Tooling
- Added `make coverage` command to Makefile
- Generates `coverage.out` (profile data) and `coverage.html` (visual report)
- Shows per-function and per-package coverage percentages
- Coverage artifacts added to `.gitignore`
- `make clean` removes coverage files

### Environment Variables (Canonical)
- `ETRADE_CONSUMER_KEY` - ETrade API consumer key
- `ETRADE_CONSUMER_SECRET` - ETrade API consumer secret
- `ETRADE_SANDBOX` - true (sandbox) or false (production), defaults to true

## Pending Documentation Updates

- [ ] 1. Reframe `docs/ALGO.md` as the engine contract (runs/steps/phases/log)
  - Rewrite `docs/ALGO.md` to describe the general runtime engine invariants that match existing code/tests/comments:
    - Run lifecycle + event names (`run.started`, `run.finished`, `run.failed`, `step.*`, `tool.*`, `artifact.*`)
    - Phase numeric mapping + transition rules
    - `workspace_root` requirements
    - Event log invariants (append-only, seq monotonic, JSONL)
  - Keep invariant numbering aligned with what code already cites; do not change any code comments.

- [ ] 2. Move trading-domain invariants out of `docs/ALGO.md`
  - Create `docs/TRADING.md`.
  - Move the current "strategy/order/position/risk/approval/LLM/trading hours" invariants from `docs/ALGO.md` into `docs/TRADING.md`.
  - Use a non-colliding numbering scheme in `docs/TRADING.md` (e.g., `T1`, `T2a`, …) so engine numbering stays stable.

- [ ] 3. Fix bad doc link
  - Update `docs/development.md` to replace any `INVARIANTS.md` mention with `docs/ALGO.md`.

- [ ] 4. Update `README.md` from template to project README
  - Replace Wails template text with:
    - what this repo is (general engine + trading layer)
    - dev/build/test commands
    - doc map: `docs/development.md`, `docs/ALGO.md`, `docs/TRADING.md`, `docs/FLOW.md`

- [ ] 5. Update `docs/FLOW.md` to map "Strategy" onto the engine model
  - Add a short glossary:
    - strategy execution (domain) maps to engine `run_id`
    - phases map to `runtime.Phase` + `step.*` events
  - Update event examples to use runtime event names (`run.*`, `step.*`) and phase strings (`data_ingestion`, etc.).
  - Replace "Invariant 65/…" references with `docs/TRADING.md` identifiers (or remove invariant numbers from FLOW if you prefer it narrative-only).

## Next Steps

- Consider integration tests for OAuth flow (httptest-based mock server)
- Unit tests for `calculate_token_expiry()` time zone logic
- Document target coverage percentage (if desired)

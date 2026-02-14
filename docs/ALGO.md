
# Runtime Engine Invariants

These invariants **MUST** hold for all engine operations.  
Violation of any invariant is a **critical bug**.

## Enforcement Categories

- **[REPLAY]** – Enforced during replay of events  
- **[EXEC]** – Enforced at execution / event-emission time  
  (Replay may not be able to verify)

## Example Run Timeline
```
run.started

    step.started (phase: data_ingestion)
      tool.called
      tool.returned
    step.finished
      
    step.started (phase: signal_generation)
      llm.requested
      llm.responded
    step.finished
      
    step.started (phase: risk_validation)
      tool.called
      tool.returned
    step.finished
      
    step.started (phase: order_execution)
      tool.called
      tool.returned
    step.finished
      
run.finished
```

---

## Run Invariants

### 1. Run ID Uniqueness
- Run IDs are UUID v4 strings with format `run-xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx`.
- RunID must be unique within the run registry / event-log namespace.
- **[EXEC]** Enforced at creation time.
- No two runs share the same RunID.

### 2a. Run Lifecycle Is Well-Formed (Event-Based)
- **[REPLAY]**
- For each `run_id`, the first event (lowest `seq`) must be `run.started`.
- For each `run_id`, there must be exactly one terminal event: `run.finished` **or** `run.failed`.
- The terminal event must be the last event (highest `seq`) for the `run_id`.
- No events may exist for that `run_id` with `seq` greater than the terminal event.
- Replay must fail with a clear error if any of the following occur:
    1) Missing start - no `run.started` event exists.
    2) Duplicate run start - more than one `run.started` event exists.
    3) Start not first - a `run.started` event exists but another event has a lower `seq`.
    4) Missing termination - No terminal event exists.
    5) Duplicate termination - More than one terminal event exists.
    6) Termination not last - A terminal event exists, but another event has a higher `seq`.

### 2b. Illegal Lifecycle Events
- **[EXEC]** 
- The runtime must not emit:
    1) a second `run.started`
    2) more than one terminal event
    3) any events after a terminal event

### 2c. Run State is Derived
- pending: no `run.started` seen
- running: `run.started` seen, no terminal case
- completed: `run.finished` seen
- failed: `run.failed` seen

### 3. Phase Execution Order
- **[REPLAY]** 
- Let phase order be `data_ingestion=1 signal_generation=2 risk_validation=3 order_execution=4`
- For each `run_id`, phase events must be non-decreasing in `seq`.
- Phase gating (forward-only progression):
    1) No `signal_generation` events unless there exists a `data_ingestion` event with lower `seq`.
    2) No `risk_validation` events unless there exists a `signal_generation` event with lower `seq`.
    3) No `order_execution` events unless there exists a `risk_validation` event with lower `seq`.
- Valid phase transitions:
    - Same phase: allowed (retries permitted within a phase)
    - Forward by 1: allowed (1→2, 2→3, 3→4)
    - Backward: NOT allowed (strict enforcement)
    - Skip forward: NOT allowed (no 1→3, must go 1→2→3)

### 4a. Workspace Root Validity
- **[EXEC]** 
- `workspace_root` must be an absolute path as determined by the host OS path rules.
- `workspace_root` must exist on the filesystem.
- `workspace_root` must be normalized (cleaned, symlinks resolved).
- Run creation fails if workspace root is invalid.

---

## Event Log Invariants

### 35. Append-Only Log
- **[EXEC]** Enforced by implementation.
- Events are never modified or deleted.
- Events are written in `seq` order.

### 36. Valid Event Types
- **[REPLAY]** Enforced during replay.
- Allowed types:
  - `run.started`, `run.finished`, `run.failed`
  - `step.started`, `step.finished`, `step.failed`
  - `llm.requested`, `llm.responded`
  - `tool.called`, `tool.returned`, `tool.failed`
  - `artifact.created`

### 37. Event ID Uniqueness
- Event IDs are UUID v4 strings.
- **[EXEC]** Enforced at creation time.

### 38. Sequence Ordering
- **[REPLAY]** Enforced during replay.
- `seq` strictly increases within a run.
- `seq > 0` (must be positive).
- `seq[n] > seq[n-1]` for all n.

### 39. Payload Validation
- **[REPLAY]** Enforced during replay.
- Payload must match event schema.
- All required fields must be present.

### 40. JSONLines Format
- **[EXEC]** Enforced by implementation.
- One valid JSON object per line.
- Lines terminated by newline character.

### 41. Replayability
- **[REPLAY]** Guaranteed by replay engine.
- RunView can be fully reconstructed from events.
- No state outside the event log.

---

## Phase Invariants

### Phase Numeric Mapping
- **[EXEC]** Frozen - these values MUST NOT change.
- `data_ingestion=1`
- `signal_generation=2`
- `risk_validation=3`
- `order_execution=4`

### Phase Zero Is Invalid
- **[EXEC]** Zero value (Phase=0) represents an uninitialized/invalid phase.
- Makes bugs obvious if Phase=0 is seen.
- All valid phases are non-zero.

### Phase String Representation
- **[EXEC]** Phases are serialized as strings in JSON, not numbers.
- Human-readable: `"data_ingestion"` not `1`
- Stable across refactorings.

---

## Step Invariants

### Step ID Uniqueness
- Step IDs are UUID v4 strings.
- Unique within a run.
- **[EXEC]** Enforced at creation time.

### Step Lifecycle Well-Formed
- **[REPLAY]** Enforced during replay.
- Exactly one `step.started` event per step.
- Exactly one of: `step.finished`, `step.failed`.
- Terminal event must occur after `step.started`.
- No subsequent events may reference the same `step_id` after termination.

### Step Belongs to One Run
- **[REPLAY]** Enforced during replay.
- `Step.RunID` must match a valid RunID.

---

## Tool Invariants

### Tool Event Lifecycle
- **[REPLAY]** For each tool invocation:
  - Exactly one `tool.called` event.
  - Exactly one of: `tool.returned`, `tool.failed`.
  - Terminal event must occur after `tool.called`.

### Tool Name Validity
- **[EXEC]** Tool name must be non-empty.
- Tool name must match a registered tool.

---

## LLM Invariants

### LLM Event Lifecycle
- **[REPLAY]** For each LLM call:
  - Exactly one `llm.requested` event.
  - Exactly one of: `llm.responded`, (failed case TBD).

---

## Artifact Invariants

### Artifact Path Validity
- **[EXEC]** Artifact paths must be relative to `workspace_root`.
- Artifact paths must not escape workspace (no `..` traversal).
- Artifact must exist on filesystem at event emission time.

### Artifact Belongs to One Run
- **[REPLAY]** `Artifact.RunID` must match a valid RunID.

---

## Crash Recovery

### Log Scanning on Startup
- **[EXEC]** On startup, scan existing log files.
- Find last sequence number for each run.
- Resume from next sequence number.
- Validate log integrity during scan.

### Incomplete Runs
- **[REPLAY]** Runs without terminal events are considered incomplete.
- Incomplete runs can be resumed or marked as failed.

---

## Summary

This document defines the general runtime engine contract. Domain-specific invariants (trading, strategy, orders, positions, risk) are documented in `docs/TRADING.md`.

The engine provides:
- Event sourcing with append-only JSONL logs
- Deterministic replay capability
- Phase-based execution model
- Isolated run workspaces
- Crash recovery

**Precedence:** When rules conflict between docs, `docs/ALGO.md` (this file) wins for engine-level "must" constraints.

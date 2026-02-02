
# Backend Invariants for Phase 1

These invariants **MUST** hold for all backend operations.  
Violation of any invariant is a **critical bug**.

## Enforcement Categories

- **[REPLAY]** – Enforced during replay of events  
- **[EXEC]** – Enforced at execution / event-emission time  
  (Replay may not be able to verify)

## Example Timeline
run.started

    planner phase:
      step.started (attempt 1)
      step.failed
      step.started (attempt 2)
      step.finished   ← planner phase succeeds here

    executor phase:
      step.started
      step.finished

    reviewer phase:
      step.started
      step.finished

run.finished

---

## Run Invariants

### 1. Run ID Uniqueness
- Run IDs are UUID v4 strings.
- RunID must be unique within the run registry / event-log namespace.
- **[EXEC]** Enforced at creation time.
- No two runs share the same RunID.

### 2a. Run Lifecycle Is Well-Formed (Event-Based)
- **[REPLAY]**
- For each `run_id`, the first event (lowest `seq`) must be `run.started`.
- For each `run_id`, there must be exactly one terminal event: `run.finished` **or** `run.failed`.
- The terminal event must be the last event (highest `seq`) for the `run_id`.
- No events may exist for that `run_id` with `seq` greater than the terminal event.
- Replay must fail with a clear error if any of the following occur for a `run_id`:
    1) Missing start - no `run.started` event exists.
    2) Duplicate run start - more than one `run.started` event exists.
    3) Start not first - a `run.started` event exists but another event for the same `run_id` has a lower `seq`.
    4) Missing termination - No `run.finished` or `run.failed` event exists.
    5) Duplicate termination - More than one termination event exists or both `run.finished` and `run.failed` exists.
    6) Termination not last - A termination event exists, but another event for the same `run_id` has a higher `seq`.
    7) Events after termination - any event for the same `run_id` occurs after the termination event.

### 2b. Illegal Lifecycle Events
- **[EXEC]** 
- The runtime must not emit:
    1) a second `run.started`
    2) more than one terminal event
    3) any events after a terminal event

### 2c. Run "state" is derived
- pending: no `run.started` seen
- running: `run.started` seen, no terminal case
- completed: `run.finished` seen
- failed: `run.failed` seen

### 3. Step Execution Order
- **[REPLAY]** 
- Let phase order be `planner=1 executor=2 reviewer=3`
- Define a step as "executed" if `step.started` exists for its `step_id`
- For each `run_id`, the phases of `step.started` events must be non-decreasing in `seq`.
- Phase gating:
    1) No executor `step.started` may exist unless there exists a planner `step.finished` with lower `seq`.
    2) No reviewer `step.started` may exist unless there exists an executor `step.finished` with lower `seq`.
- Retries:
    1) Each phase may have at most 3 attempts, where an attempt is a `step.started` in that phase.
    2) After 3 attempts in a phase without a `step.finished` in that phase, the run must not start any later-phase steps and must terminate with `run.failed`.

### 4a. Workspace Root Configuration
- **[EXEC]** 
- `workspace_root` must be an absolute path as determined by the host OS path rules
- `workspace_root` must be normalized before use (e.g. `Clean`, `EvalSymlinks` if supported)
- Run creation fails if `workspace_root` is invalid

### 4b. File operation containment
- **[EXEC]**
- All file operation paths are resolved as follows:
    1) If relative, resolve against `workspace_root`
    2) Normalize the resulting path
- The normalized path must be equal to or descendant of `workspace_root`
- Any operation that resolves outside `workspace_root` fails before execution.

### 5a. Failure Stops Subsequent Phases
- **[EXEC]** 
- If a step attempt in the current phase emits `step.failed`, the runtime may start another step in the same phase (retry), up to the attempt limit.
- The runtime MUST NOT start steps in later phases until the current phase has a `step.finished`.
- If the attempt limit is exceeded for a phase, the runtime MUST NOT start later phases and MUST terminate the run as failed.

---

## Step Invariants

### 8. Step ID Uniqueness
- Step IDs are UUID v4 strings.
- Unique within a run.
- **[EXEC]** Enforced at creation time.

### 9. Exactly One `step.started`
- **[REPLAY]** Enforced during replay.
- No duplicate `step.started` events.
- Duplicate start events cause replay failure.
- Start `seq` is tracked for error reporting.

### 10. Exactly One Step Termination
- **[REPLAY]** Enforced during replay.
- Exactly one of `step.finished` or `step.failed`.
- No subsequent events may reference the same `step_id`.
- Incomplete steps cause replay failure.

### 11. Step Belongs to One Run
- **[REPLAY]** Enforced during replay.
- `Step.RunID` must match a valid RunID.

### 12. Step Phase
- **[EXEC]** Phase is immutable.
- Must be exactly one of: `planner`, `executor`, `reviewer`.

### 13. Agent Validity
- **[EXEC]** Must be enforced before `step.started`.
- Planner steps use planner agent.
- Executor steps use executor agent.
- Reviewer steps use reviewer agent.

---

## Tool Call Invariants

### 14. ToolCall ID Uniqueness
- ToolCall IDs are UUID v4 strings.
- Unique within a run.
- **[EXEC]** Enforced at creation time.

### 15. Exactly One `tool.called`
- **[REPLAY]** Enforced during replay.
- No duplicate `tool.called` events.
- Duplicate start events cause replay failure.
- Start `seq` is tracked.

### 16. Exactly One Tool Termination
- **[REPLAY]** Enforced during replay.
- Exactly one of `tool.returned` or `tool.failed`.
- No subsequent events may reference the same `tool_call_id`.
- Incomplete tool calls cause replay failure.

### 17. ToolCall Belongs to One Step
- **[REPLAY]** Enforced during replay.
- `ToolCall.StepID` must match a valid StepID.

### 18. Tool Validity
- **[EXEC]** Enforced before `tool.called`.
- Tool must exist in the tool registry.
- Tool must be whitelisted for the step’s agent.

### 19. Permission Tiers
- **[EXEC]** Enforced before `tool.called`.
- `Agent.PermissionTier ≥ Tool.PermissionTier`.

### 20. Tool Input Validation
- **[EXEC]** Enforced before `tool.called`.
- Invalid input results in `tool.failed` with `INVALID_INPUT`.

### 21. Tool Output Validation
- **[EXEC]** Enforced before `tool.returned`.
- Invalid output results in `tool.failed` with `INVALID_OUTPUT`.

### 22. ToolCall Duration
- **[EXEC]** `duration_ms ≥ 0`.

---

## LLM Call Invariants

### 23. LLMCall ID Uniqueness
- UUID v4 strings.
- Unique within a run.
- **[EXEC]** Enforced at creation time.

### 24. Exactly One `llm.requested`
- **[REPLAY]** Enforced during replay.
- No duplicate request events.
- Duplicate start events cause replay failure.

### 25. Exactly One `llm.responded`
- **[REPLAY]** Enforced during replay.
- No events may reference the same `llm_call_id` afterward.
- Incomplete LLM calls cause replay failure.

### 26. LLMCall Belongs to One Step
- **[REPLAY]** Enforced during replay.
- `LLMCall.StepID` must match a valid StepID.

---

## Artifact Invariants

### 27. Artifact ID Uniqueness
- UUID v4 strings.
- Unique within a run.
- **[EXEC]** Enforced at creation time.

### 28. Exactly One `artifact.created`
- **[EXEC]** Enforced at creation time.

### 29. Artifact Belongs to One Step
- **[REPLAY]** Enforced during replay.
- `Artifact.StepID` must match a valid StepID.

### 30. Artifact Belongs to One Run
- **[REPLAY]** Enforced during replay.
- `Artifact.RunID` must match a valid RunID.

### 31. Artifact Type
- **[EXEC]** Immutable at creation.
- One of: `file`, `diff`, `text`.

### 32. Artifact Checksum
- **[EXEC]** Must be a valid SHA256 hash.

### 33. Artifact Size
- **[EXEC]** `size_bytes ≥ 0`.

### 34. File Artifact Path
- **[EXEC]** Enforced at creation.
- Must be absolute and within `workspace_root`.
- No path traversal or escape.
- Optional for `diff` and `text`.

---

## Event Log Invariants

### 35. Append-Only Log
- **[EXEC]** Enforced by implementation.
- Events are never modified or deleted.
- Events are written in `seq` order.

### 36. Event ID Uniqueness
- UUID v4 strings.
- **[EXEC]** Enforced at creation time.

### 37. Valid Event Types
- **[REPLAY]** Enforced during replay.
- Allowed types:
  - `run.started`, `run.finished`, `run.failed`
  - `step.started`, `step.finished`, `step.failed`
  - `llm.requested`, `llm.responded`
  - `tool.called`, `tool.returned`, `tool.failed`
  - `artifact.created`

### 38. Sequence Ordering
- **[REPLAY]** Enforced during replay.
- `seq` strictly increases within a run.

### 39. Payload Validation
- **[REPLAY]** Enforced during replay.
- Payload must match event schema.

### 40. JSONLines Format
- **[EXEC]** Enforced by implementation.
- One valid JSON object per line.
- Lines terminated by newline.

### 41. Replayability
- **[REPLAY]** Guaranteed by replay engine.
- RunView can be fully reconstructed.
- No state outside the event log.

---

## File System Invariants

### 42. Workspace Containment
- **[EXEC]** Enforced by file tools.
- No access outside `workspace_root`.

### 43. Path Normalization
- **[EXEC]** Enforced by file tools.
- Relative paths resolved against `workspace_root`.
- Normalized and containment-checked.

### 44. No Arbitrary Shell Execution
- **[EXEC]** Enforced by design.
- No `execute_shell` in Phase 1.
- Only allowlisted wrappers: `run_tests`, `run_build`, `run_lint`.

---

## Agent Invariants

### 45. Static Agent Configuration
- **[EXEC]** Enforced at startup.
- Configuration is immutable.
- Agents identified by ID.

### 46. Tool Whitelist
- **[EXEC]** Enforced before tool execution.
- Agents can only use whitelisted tools.

### 47. Permission Tier
- **[EXEC]** Immutable per agent.
- Determines accessible tools.

### 48. Phase 1 Agents
- **[EXEC]** Exactly three agents:
  - Planner (READ-only)
  - Executor (WRITE / EXECUTE)
  - Reviewer (READ-only)

---

## Memory Invariants

### 49. Per-Run Isolation
- **[EXEC]** Enforced by implementation.
- Memory is namespaced by RunID.

### 50. Ephemeral Memory
- **[EXEC]** Enforced by design.
- Lost on process restart.
- Not persisted in Phase 1.

### 51. Memory Key Constraints
- **[EXEC]** Keys are strings.
- Values are JSON-serializable.

---

## Replay Invariants

### 52. Deterministic Replay
- **[REPLAY]** Guaranteed by replay engine.
- Same events → same RunView.

### 53. Invariant Validation During Replay
- **[REPLAY]** Guaranteed by replay engine.
- Errors include `seq`, event type, and reason.

### 54. Complete RunView Reconstruction
- **[REPLAY]** Guaranteed by replay engine.
- All steps, tool calls, LLM calls, and artifacts are restored.

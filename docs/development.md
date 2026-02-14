# Development Guide

## Core Philosophy

**Events are the source of truth. State is a cache.**

## E*TRADE Integration

For E*TRADE API setup, OAuth configuration, and testing procedures, see `docs/etrade_manual_oauth_test.md`.

## Zero Technical Debt Policy

**We do it right the first time.**

A problem solved in production is many times more expensive than a problem solved in implementation, or a problem solved in design. We don't allow potential latency spikes, or exponential complexity algorithms to slip through.

This is the only way to make steady incremental progress, knowing that the progress we have made is indeed progress. We may lack features, but what we have meets our design goals.

## Key Principles

### 1. Single-Threaded State Mutation

**Principle:** Only one goroutine mutates state. No mutexes.

**Why:** Eliminates race conditions without complex locking. Makes reasoning about state changes trivial.

**Implementation:**
```go
type Engine struct {
    cmdCh chan Command  // All state changes go through here
}

func (e *Engine) runLoop() {
    runs := make(map[RunID]*RunHandle)
    for cmd := range e.cmdCh {
        switch c := cmd.(type) {
        case StartRunCmd:
            e.handleStartRun(runs, c)  // Only this goroutine touches 'runs'
        }
    }
}
```

**Trade-off:** Throughput limited by single thread. 

### 2. Invariant-Driven Development

**Principle:** Define what MUST always be true, then design so invariants are unrepresentable or enforced.

**Traditional Approach:**
1. Write code
2. Hope it works
3. Write tests to check
4. Debug when things break

**Tiger Beetle Approach:**
1. Define invariants (see docs/ALGO.md)
2. Design system so invariants are enforced
3. Code is implementation of proofs
4. Tests verify invariants hold

**Example - Invariant 38 (Sequence Ordering):**
```go
// Enforced at write time
seq := atomic.AddInt64(&nextSeq, 1) - 1

// Enforced at read time (replay)
if envelope.Seq <= lastSeq {
    return fmt.Errorf("sequence %d not strictly increasing", envelope.Seq)
}
```

### 3. Event Sourcing

**Principle:** All state changes are captured as immutable events. Current state is derived by replaying events.

**Benefits:**
- Perfect audit trail
- Can reconstruct state at any point in time
- Replay for debugging
- No lost data

**Implementation:**
```
Command → Validate → Write Event (JSONL) → Update Cache
                     ↓
              Source of Truth
```

**Critical Rule:** Write to log BEFORE updating in-memory state. If crash happens between, replay will reconstruct correct state.

### 4. Append-Only Logs

**Principle:** Events are never modified or deleted. Only append.

**Format:** JSON Lines (JSONL) - one JSON object per line, newline terminated.

**Why JSONL:**
- Human readable (can `tail -f` to watch events)
- Machine parseable
- Can recover partial files (each line independent)
- Simple to implement

**Example:**
```json
{"type":"run.started","seq":1,"event_id":"...","run_id":"run-xxx"}
{"type":"step.started","seq":2,"event_id":"...","step_id":"step-yyy"}
```

### 5. Assertions Are Force Multipliers

**Principle:** Assertions detect programmer errors. Unlike operating errors (expected), assertion failures are unexpected. The only correct way to handle corrupt code is to crash.

> "The golden rule of assertions is to assert the _positive space_ that you do expect AND to assert the _negative space_ that you do not expect"

**Key Rules:**
- **Minimum two assertions per function** (average)
- **Assert all function arguments and return values**
- **Assert preconditions, postconditions, and invariants**
- **Pair assertions** - For every property, find two code paths to assert it (e.g., before write AND after read)
- **Assert relationships of compile-time constants** - Use Go build constraints or init-time checks
- **Split compound assertions:** prefer `assert(a); assert(b)` over `assert(a && b)` for precision
- **Use single-line if for implications:** `if condition { assert(invariant) }`

**Example:**
```go
func (e *EventLog) Append(event Event) error {
    // Precondition assertions
    assert(e.file != nil, "file must be open")
    assert(event != nil, "event cannot be nil")
    
    seq := atomic.AddInt64(&e.nextSeq, 1) - 1
    assert(seq > 0, "seq must be positive")
    
    // ... write event ...
    
    // Postcondition assertion
    assert(e.nextSeq > seq, "nextSeq must be incremented")
    return nil
}

func scanLastSeq(file *os.File) (int64, error) {
    // Pair assertion: validate at read time (also validated at write time)
    var lastSeq int64
    for scanner.Scan() {
        line := scanner.Text()
        var envelope struct{ Seq int64 `json:"seq"` }
        assert(json.Unmarshal([]byte(line), &envelope) == nil, "line must be valid JSON")

        // Assert positive space: what we expect
        assert(envelope.Seq > 0, "seq must be positive")

        // Assert negative space: what we don't expect
        assert(envelope.Seq > lastSeq, "seq must strictly increase")

        lastSeq = envelope.Seq
    }
    return lastSeq, nil
}
```

**When to panic:**
- Programmer errors (unknown command types)
- Impossible states (UUID collision)
- Unrecoverable errors (disk full)
- Assertion failures (invariant violations)

### 6. Zero Values Are Invalid

**Principle:** The zero value (0, nil, "") should represent an invalid/uninitialized state.

**Example - Phase enum:**
```go
type Status int

const (
    StatusInvalid Status = iota // 0 - invalid/uninitialized
    StatusPending               // 1
    StatusRunning               // 2
    StatusCompleted             // 3
)
```

**Why:** Makes bugs obvious. If you see Phase=0, you know something forgot to initialize it.

### 7. Sealed Interfaces

**Principle:** Use Go's unexported method pattern to control implementations.

**Example:**
```go
type Event interface {
    event() // Unexported - only this package can implement
}

type RunStartedEvent struct{...}
func (RunStartedEvent) event() {} // Implements Event
```

**Benefit:** Compile-time guarantee that only defined event types can be used.

### 8. Crash Recovery

**Principle:** System must be able to recover from crashes and resume correctly.

**Implementation:**
1. Scan existing log files on startup
2. Find last sequence number
3. Resume from next sequence
4. Validate log integrity during scan

**Example:**
```go
func OpenEventLog(runID RunID, workspaceRoot string) (*EventLog, error) {
    // Check if resuming existing run
    if !isNewFile {
        lastSeq, err := scanLastSeq(file)
        if err != nil {
            return nil, fmt.Errorf("corrupt log: %w", err)
        }
        nextSeq = lastSeq + 1  // Resume from here
    }
}
```

### 9. Validation at Multiple Layers

**Principle:** Validate early, validate often.

**Layers:**
1. **Command validation** - Check inputs before processing
2. **Event validation** - Ensure events are well-formed before writing
3. **Replay validation** - Verify log integrity when reading
4. **Business logic validation** - Check invariants (exactly one run.started, etc.)

**Example - Invariant 4a (Absolute Path):**
```go
// Layer 1: Command validation
normalizedPath, err := normalizeWorkspaceRoot(workspaceRoot)
if err != nil {
    return error  // Reject before any state change
}

// Layer 2: Write event
// Layer 3: Replay validates path still absolute
// Layer 4: Business logic validates path within workspace
```

### 10. Explicit Over Clever

**Principle:** Prefer clear, explicit code over clever optimizations.

**Example - Switch statements:**
```go
// Good - explicit, compiler checks exhaustiveness
switch e := event.(type) {
case *RunStartedEvent:
    // handle run started
case *StepStartedEvent:
    // handle step started
default:
    panic(fmt.Sprintf("unknown event type: %T", event))
}

// Bad - reflection magic, hard to follow
handler := getHandlerForType(reflect.TypeOf(event))
handler.Handle(event)
```

### 11. No Recursion

**Principle:** Do not use recursion. Ensure all executions that should be bounded are bounded.

**Implementation:** Use explicit loops with fixed upper bounds.

**Why:** Stack overflow protection, predictable memory usage, easier to reason about bounds.

```go
// Good - bounded loop
for i := 0; i < maxAttempts; i++ {
    if tryOperation() {
        break
    }
}

// Bad - recursion
tryOperationRecursive(attempts int) {
    if attempts >= maxAttempts { return }
    if !tryOperation() {
        tryOperationRecursive(attempts + 1)
    }
}
```

### 12. Put a Limit on Everything

**Principle:** Everything has a limit. This follows the "fail-fast" principle.

**Examples:**
- All loops must have fixed upper bounds
- All queues must have fixed capacity
- All buffers must have fixed sizes
- Channel buffers are bounded

```go
// Good - bounded channel
cmdCh: make(chan Command, 64)  // Max 64 pending commands

// Good - bounded loop
for i := 0; i < maxSteps && !done; i++ {
    processStep()
}
```

### 13. Small Functions (70 Lines Max)

**Principle:** Hard limit of 70 lines per function.

**Art is born of constraints.** There are many ways to cut a wall of code into chunks of 70 lines, but only a few splits will feel right.

**Rules of thumb:**
- Good function shape is the inverse of an hourglass: few parameters, simple return, meaty logic
- Centralize control flow - keep switch/if statements in the "parent" function
- Move non-branchy logic to helpers
- Divide responsibility: one function handles control flow, helpers are pure
- "Push ifs up and fors down" - Matklad

### 14. Always Motivate, Always Say Why

**Principle:** Never forget to say why. Explain the rationale for decisions.

**This:**
- Increases the hearer's understanding
- Makes them more likely to adhere
- Shares criteria to evaluate the decision

**Example:**
```go
// Good
// Skip 0 - invalid/uninitialized. Makes bugs obvious if Phase=0 is seen.
_ Phase = iota

// Bad
// Initialize enum
PhaseUnknown Phase = iota
```

### 15. Explicit Options, Not Defaults

**Principle:** Explicitly pass options at call sites instead of relying on defaults.

**Go Example:**
```go
// Good - explicit
file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

// Bad - unclear what the defaults are
file, err := os.Open(path)  // What mode? What permissions?
```

### 16. Compound Conditions

**Principle:** Split compound conditions into simple conditions using nested branches.

**Why:** Makes all cases clear, easier to verify exhaustiveness.

```go
// Good - clear branches
if conditionA {
    if conditionB {
        // Both true
    } else {
        // A true, B false
    }
} else {
    // A false
}

// Bad - compound, hard to follow
if conditionA && conditionB {
    // Both true
} else if conditionA && !conditionB {
    // A true, B false  
} else {
    // A false (what about B?)
}
```

### 17. State Invariants Positively

**Principle:** State conditions positively, not with negations.

```go
// Good - easy to understand
if index < length {
    // Invariant holds
} else {
    // Invariant doesn't hold
}

// Bad - harder to reason about
if !(index >= length) {
    // It's not true that invariant doesn't hold
}
```

### 18. All Errors Must Be Handled

**Principle:** An analysis of production failures found that 92% of catastrophic failures were caused by incorrect handling of non-fatal errors.

**In Go:** Check every error. Never ignore with `_` unless explicitly justified.

```go
// Good
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Bad
_ = doSomething()  // What if this fails?
```

### 19. Minimize Allocations

**Principle:** While Go has garbage collection, excessive allocations create GC pressure and latency spikes.

**In Go:**

```go
// Good - stack allocation (value type)
func processItems(items []Item) {
    var result Item  // Stack allocated
    for _, item := range items {
        result = process(item)
    }
}

// Good - pre-allocate slices
results := make([]Result, 0, len(items))  // One allocation
for _, item := range items {
    results = append(results, process(item))
}

// Bad - repeated allocations
var results []Result  // Starts at 0 capacity
for _, item := range items {
    results = append(results, process(item))  // Multiple reallocations
}

// Good - use sync.Pool for high-frequency temporary objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

buf := bufferPool.Get().([]byte)
// ... use buf ...
bufferPool.Put(buf)  // Reuse instead of GC
```

**Key differences from Zig:**
- Go manages memory automatically (GC)
- Compiler decides stack vs heap (escape analysis)
- No explicit allocators, but we can control allocation patterns
- Use value types over pointers when possible
- Pre-allocate, pool, and reuse when performance matters

## Code Patterns

### Command Pattern

All state changes go through commands:

```go
type Command interface {
    command() // Sealed interface
}

type StartRunCmd struct {
    WorkspaceRoot string
    ResultCh      chan<- StartRunResult
}

func (e *Engine) StartRun(workspaceRoot string) (RunID, error) {
    resultCh := make(chan StartRunResult, 1)
    e.cmdCh <- StartRunCmd{WorkspaceRoot: workspaceRoot, ResultCh: resultCh}
    return (<-resultCh).ID, nil
}
```

### Event Types as Constants

Prevent typos with typed constants:

```go
type EventType string

const (
    EventTypeRunStarted  EventType = "run.started"
    EventTypeRunFinished EventType = "run.finished"
    // ...
)
```

### Result Channels

Commands use channels for async results:

```go
type StartRunResult struct {
    ID  RunID
    Err error
}

// Caller blocks until result is ready
result := <-resultCh
```

### Atomic Operations

Use atomic for simple counters (defensive even if single-threaded):

```go
seq := atomic.AddInt64(&nextSeq, 1) - 1
```

### File Operations

Always use appropriate flags and handle errors:

```go
file, err := os.OpenFile(path, 
    os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
```

## Naming Conventions

### General Rules

- Use `snake_case` for unexported/internal identifiers, file names, JSON fields, and event type strings
- Exported identifiers remain Go-exported style (CamelCase) for public APIs
- Do not abbreviate variable names (unless primitive integer used for sorting/calculation)
- Use proper capitalization for acronyms (`VSRState`, not `VsrState`)
- Add units or qualifiers to variable names, sorted by descending significance
- For the rest, follow Go conventions

### Units and Qualifiers

Put units/qualifiers last, sorted by descending significance:

```go
// Good - groups related variables
latency_ms_max int
latency_ms_min int
latency_ms_avg int

// Bad
max_latency_ms int
min_latency_ms int
avg_latency_ms int
```

### Meaningful Names

Infuse names with meaning:

```go
// Good - informs reader about lifecycle
httpClient *http.Client  // Long-lived, reusable
body []byte              // Short-lived, per-request
file *os.File           // Must Close when done
```

### Same Length for Related Names

When choosing related names, try to find names with the same number of characters so related variables line up:

```go
// Good - aligns nicely
source       []byte
target       []byte
sourceOffset int
targetOffset int

// Bad - doesn't align
src  []byte
dest []byte
srcOffset  int
destOffset int
```

### Callback Naming

Prefix helper function names with the calling function name to show call history:

```go
func readSector() error { ... }
func readSectorCallback() { ... }  // Called by readSector
```

### File Order

Put important things near the top. Files are read top-down.

Order in files:
1. Package comment
2. Imports
3. Constants
4. Types/Structs (with fields first, then methods)
5. Functions
6. Tests (in _test.go files)

### Comments

- Comments are sentences: space after slash, capital letter, full stop (or colon if related to following code)
- Use comments to explain WHY you wrote the code the way you did
- Show your workings
- Line-end comments can be phrases without punctuation

```go
// This is a sentence comment.

// This explains why: the following code handles the edge case where...
if edgeCase {
    handle()  // Line-end comment (no punctuation)
}
```

## Testing Strategy

### Invariant Tests

Name tests after the invariants they verify:

```go
func TestInvariant_1_RunIDUniqueness(t *testing.T) {...}
func TestInvariant_4a_WorkspaceRootAbsolute(t *testing.T) {...}
```

### Property-Based Testing

Test properties, not just examples:
- Concurrent runs should never have duplicate IDs
- Sequence numbers should always increase
- Relative paths should always be rejected

### Determinism

Tests should be deterministic:
- Use `t.TempDir()` for test directories
- Don't rely on timing
- Control randomness (UUIDs are fine, they're unique)

### Exhaustive Testing

Tests must test exhaustively - not only with valid data but also with invalid data:

```go
func TestInvariant_4a_WorkspaceRootAbsolute(t *testing.T) {
    tests := []struct {
        name          string
        workspaceRoot string
        shouldFail    bool
    }{
        {name: "absolute_path", workspaceRoot: "/tmp/workspace", shouldFail: false},
        {name: "relative_path", workspaceRoot: "relative/path", shouldFail: true},
        {name: "empty_path", workspaceRoot: "", shouldFail: true},
        // ... more edge cases
    }
}
```

## Dependencies

**Zero dependencies policy** (apart from Go toolchain).

Dependencies lead to supply chain attacks, safety and performance risk, and slow install times. For foundational infrastructure, the cost of any dependency is amplified throughout the rest of the stack.

**Standard library first:**
- Use `net/http` not frameworks
- Use `encoding/json` not third-party parsers
- Use `testing` not fancy test libraries

## Tooling

A small standardized toolbox is simpler to operate than an array of specialized instruments.

**Primary tool: Go**

Standardize on Go for:
- Application code
- Tests
- Scripts (instead of shell scripts)
- Build tools

> "The right tool for the job is often the tool you are already using—adding new tools has a higher cost than many people appreciate" — John Carmack

## Style By The Numbers

- Run `go fmt` - use standard Go formatting, don't fight it
- Hard limit: 100 columns for line length
- Hard limit: 70 lines per function
- Hard limit: 2 assertions per function (minimum average)

## When to Break These Rules

These are strong defaults, not absolute laws:

1. **Performance requirements** - If single-threaded is too slow, shard by RunID into multiple engines
2. **External APIs** - Can't control external services, handle errors gracefully
3. **User experience** - Sometimes "degrade gracefully" is better than "fail fast"
4. **Phase 1 pragmatism** - Some features (fsync every write) can be added later

## Further Reading

- Tiger Beetle blog: https://tigerbeetle.com/blog
- Tiger Beetle GitHub: https://github.com/tigerbeetle/tigerbeetle
- "Turning the database inside out" by Martin Kleppmann
- NASA's Power of Ten — Rules for Developing Safety Critical Code
- "Let Over Lambda" by Doug Hoyte

## Summary

The Tiger Beetle approach trades some flexibility for correctness and reliability:

- **Single-threaded** → No race conditions
- **Event sourcing** → Perfect audit trail
- **Invariant-driven** → Catch bugs early
- **Fail fast** → Problems are obvious
- **Explicit** → Code is self-documenting
- **Zero technical debt** → Do it right the first time
- **Assertions everywhere** → Force multiplier for bug discovery

For a desktop AI agent platform, this provides a solid foundation that can scale up when needed.

> "You don't really suppose, do you, that all your adventures and escapes were managed by mere luck, just for your sole benefit? You are a very fine person, Mr. Baggins, and I am very fond of you; but you are only quite a little fellow in a wide world after all!"
>
> "Thank goodness!" said Bilbo laughing, and handed him the tobacco-jar.

package runtime

// Event is the interface for all events that can be written to the event log.
// The event() method is a marker to ensure only valid event types are used.
// This is a common Go pattern for creating "sealed" interfaces - only types
// in this package can implement Event by implementing the event() method.
type Event interface {
	event() // Marker method - unexported so only this package can implement
}

// EventType identifies the type of an event.
// Using a dedicated type instead of raw strings gives us type safety
// and prevents typos in event type names.
type EventType string

const (
	// Run lifecycle events
	EventTypeRunStarted  EventType = "run.started"
	EventTypeRunFinished EventType = "run.finished"
	EventTypeRunFailed   EventType = "run.failed"

	// Step lifecycle events
	EventTypeStepStarted  EventType = "step.started"
	EventTypeStepFinished EventType = "step.finished"
	EventTypeStepFailed   EventType = "step.failed"

	// LLM events
	EventTypeLLMRequested EventType = "llm.requested"
	EventTypeLLMResponded EventType = "llm.responded"

	// Tool events
	EventTypeToolCalled   EventType = "tool.called"
	EventTypeToolReturned EventType = "tool.returned"
	EventTypeToolFailed   EventType = "tool.failed"

	// Artifact events
	EventTypeArtifactCreated EventType = "artifact.created"
)

// RunStartedEvent is emitted when a new run begins.
// This is the first event for any run (Invariant 2a: first event must be run.started).
type RunStartedEvent struct {
	RunID         RunID     `json:"run_id"`
	WorkspaceRoot string    `json:"workspace_root"`
	Seq           int64     `json:"seq"`
	Type          EventType `json:"type"`
}

func (RunStartedEvent) event() {}

// RunFinishedEvent is emitted when a run completes successfully.
type RunFinishedEvent struct {
	RunID RunID     `json:"run_id"`
	Seq   int64     `json:"seq"`
	Type  EventType `json:"type"`
}

func (RunFinishedEvent) event() {}

// RunFailedEvent is emitted when a run fails.
type RunFailedEvent struct {
	RunID  RunID     `json:"run_id"`
	Reason string    `json:"reason"`
	Seq    int64     `json:"seq"`
	Type   EventType `json:"type"`
}

func (RunFailedEvent) event() {}

// StepStartedEvent is emitted when a step begins.
type StepStartedEvent struct {
	RunID  RunID     `json:"run_id"`
	StepID string    `json:"step_id"`
	Phase  Phase     `json:"phase"`
	Seq    int64     `json:"seq"`
	Type   EventType `json:"type"`
}

func (StepStartedEvent) event() {}

// StepFinishedEvent is emitted when a step completes successfully.
type StepFinishedEvent struct {
	RunID  RunID     `json:"run_id"`
	StepID string    `json:"step_id"`
	Phase  Phase     `json:"phase"`
	Seq    int64     `json:"seq"`
	Type   EventType `json:"type"`
}

func (StepFinishedEvent) event() {}

// StepFailedEvent is emitted when a step fails.
type StepFailedEvent struct {
	RunID  RunID     `json:"run_id"`
	StepID string    `json:"step_id"`
	Phase  Phase     `json:"phase"`
	Reason string    `json:"reason"`
	Seq    int64     `json:"seq"`
	Type   EventType `json:"type"`
}

func (StepFailedEvent) event() {}

// LLMRequestedEvent is emitted when an LLM call is requested.
type LLMRequestedEvent struct {
	RunID  RunID     `json:"run_id"`
	StepID string    `json:"step_id"`
	Seq    int64     `json:"seq"`
	Type   EventType `json:"type"`
}

func (LLMRequestedEvent) event() {}

// LLMRespondedEvent is emitted when an LLM call completes.
type LLMRespondedEvent struct {
	RunID  RunID     `json:"run_id"`
	StepID string    `json:"step_id"`
	Seq    int64     `json:"seq"`
	Type   EventType `json:"type"`
}

func (LLMRespondedEvent) event() {}

// ToolCalledEvent is emitted when a tool is invoked.
type ToolCalledEvent struct {
	RunID    RunID     `json:"run_id"`
	StepID   string    `json:"step_id"`
	ToolName string    `json:"tool_name"`
	Seq      int64     `json:"seq"`
	Type     EventType `json:"type"`
}

func (ToolCalledEvent) event() {}

// ToolReturnedEvent is emitted when a tool call completes successfully.
type ToolReturnedEvent struct {
	RunID    RunID     `json:"run_id"`
	StepID   string    `json:"step_id"`
	ToolName string    `json:"tool_name"`
	Seq      int64     `json:"seq"`
	Type     EventType `json:"type"`
}

func (ToolReturnedEvent) event() {}

// ToolFailedEvent is emitted when a tool call fails.
type ToolFailedEvent struct {
	RunID    RunID     `json:"run_id"`
	StepID   string    `json:"step_id"`
	ToolName string    `json:"tool_name"`
	Reason   string    `json:"reason"`
	Seq      int64     `json:"seq"`
	Type     EventType `json:"type"`
}

func (ToolFailedEvent) event() {}

// ArtifactCreatedEvent is emitted when an artifact is created.
type ArtifactCreatedEvent struct {
	RunID  RunID     `json:"run_id"`
	StepID string    `json:"step_id"`
	Path   string    `json:"path"`
	Seq    int64     `json:"seq"`
	Type   EventType `json:"type"`
}

func (ArtifactCreatedEvent) event() {}

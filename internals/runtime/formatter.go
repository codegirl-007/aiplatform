package runtime

import "aiplatform/pkg/assert"

// Formatter is the single, authoritative source for creating fully-formed events.
// It is the only place allowed to set event Type fields.
// The writer goroutine assigns Seq before calling formatter functions.
//
// Tiger Beetle Principle: Single source of truth for event formation.

// FormatRunStarted creates a fully-formed RunStartedEvent.
// The caller (writer goroutine) must provide the seq.
func FormatRunStarted(seq int64, runID RunID, workspaceRoot string) RunStartedEvent {
	assert.Gt(seq, int64(0), "seq must be positive")
	assert.Is_true(runID != RunID(""), "runID must not be empty")
	assert.Not_empty(workspaceRoot, "workspaceRoot must not be empty")

	return RunStartedEvent{
		RunID:         runID,
		WorkspaceRoot: workspaceRoot,
		Seq:           seq,
		Type:          EventTypeRunStarted,
	}
}

// FormatRunFinished creates a fully-formed RunFinishedEvent.
func FormatRunFinished(seq int64, runID RunID) RunFinishedEvent {
	assert.Gt(seq, int64(0), "seq must be positive")
	assert.Is_true(runID != RunID(""), "runID must not be empty")

	return RunFinishedEvent{
		RunID: runID,
		Seq:   seq,
		Type:  EventTypeRunFinished,
	}
}

// FormatRunFailed creates a fully-formed RunFailedEvent.
func FormatRunFailed(seq int64, runID RunID, reason string) RunFailedEvent {
	assert.Gt(seq, int64(0), "seq must be positive")
	assert.Is_true(runID != RunID(""), "runID must not be empty")
	assert.Not_empty(reason, "reason must not be empty")

	return RunFailedEvent{
		RunID:  runID,
		Reason: reason,
		Seq:    seq,
		Type:   EventTypeRunFailed,
	}
}

// FormatStepStarted creates a fully-formed StepStartedEvent.
func FormatStepStarted(seq int64, runID RunID, stepID string, phase Phase) StepStartedEvent {
	assert.Gt(seq, int64(0), "seq must be positive")
	assert.Is_true(runID != RunID(""), "runID must not be empty")
	assert.Not_empty(stepID, "stepID must not be empty")
	assert.Gt(int64(phase), 0, "phase must be positive")

	return StepStartedEvent{
		RunID:  runID,
		StepID: stepID,
		Phase:  phase,
		Seq:    seq,
		Type:   EventTypeStepStarted,
	}
}

// FormatStepFinished creates a fully-formed StepFinishedEvent.
func FormatStepFinished(seq int64, runID RunID, stepID string, phase Phase) StepFinishedEvent {
	assert.Gt(seq, int64(0), "seq must be positive")
	assert.Is_true(runID != RunID(""), "runID must not be empty")
	assert.Not_empty(stepID, "stepID must not be empty")
	assert.Gt(int64(phase), 0, "phase must be positive")

	return StepFinishedEvent{
		RunID:  runID,
		StepID: stepID,
		Phase:  phase,
		Seq:    seq,
		Type:   EventTypeStepFinished,
	}
}

// FormatStepFailed creates a fully-formed StepFailedEvent.
func FormatStepFailed(seq int64, runID RunID, stepID string, phase Phase, reason string) StepFailedEvent {
	assert.Gt(seq, int64(0), "seq must be positive")
	assert.Is_true(runID != RunID(""), "runID must not be empty")
	assert.Not_empty(stepID, "stepID must not be empty")
	assert.Gt(int64(phase), 0, "phase must be positive")
	assert.Not_empty(reason, "reason must not be empty")

	return StepFailedEvent{
		RunID:  runID,
		StepID: stepID,
		Phase:  phase,
		Reason: reason,
		Seq:    seq,
		Type:   EventTypeStepFailed,
	}
}

// FormatLLMRequested creates a fully-formed LLMRequestedEvent.
func FormatLLMRequested(seq int64, runID RunID, stepID string) LLMRequestedEvent {
	assert.Gt(seq, int64(0), "seq must be positive")
	assert.Is_true(runID != RunID(""), "runID must not be empty")
	assert.Not_empty(stepID, "stepID must not be empty")

	return LLMRequestedEvent{
		RunID:  runID,
		StepID: stepID,
		Seq:    seq,
		Type:   EventTypeLLMRequested,
	}
}

// FormatLLMResponded creates a fully-formed LLMRespondedEvent.
func FormatLLMResponded(seq int64, runID RunID, stepID string) LLMRespondedEvent {
	assert.Gt(seq, int64(0), "seq must be positive")
	assert.Is_true(runID != RunID(""), "runID must not be empty")
	assert.Not_empty(stepID, "stepID must not be empty")

	return LLMRespondedEvent{
		RunID:  runID,
		StepID: stepID,
		Seq:    seq,
		Type:   EventTypeLLMResponded,
	}
}

// FormatToolCalled creates a fully-formed ToolCalledEvent.
func FormatToolCalled(seq int64, runID RunID, stepID string, toolName string) ToolCalledEvent {
	assert.Gt(seq, int64(0), "seq must be positive")
	assert.Is_true(runID != RunID(""), "runID must not be empty")
	assert.Not_empty(stepID, "stepID must not be empty")
	assert.Not_empty(toolName, "toolName must not be empty")

	return ToolCalledEvent{
		RunID:    runID,
		StepID:   stepID,
		ToolName: toolName,
		Seq:      seq,
		Type:     EventTypeToolCalled,
	}
}

// FormatToolReturned creates a fully-formed ToolReturnedEvent.
func FormatToolReturned(seq int64, runID RunID, stepID string, toolName string) ToolReturnedEvent {
	assert.Gt(seq, int64(0), "seq must be positive")
	assert.Is_true(runID != RunID(""), "runID must not be empty")
	assert.Not_empty(stepID, "stepID must not be empty")
	assert.Not_empty(toolName, "toolName must not be empty")

	return ToolReturnedEvent{
		RunID:    runID,
		StepID:   stepID,
		ToolName: toolName,
		Seq:      seq,
		Type:     EventTypeToolReturned,
	}
}

// FormatToolFailed creates a fully-formed ToolFailedEvent.
func FormatToolFailed(seq int64, runID RunID, stepID string, toolName string, reason string) ToolFailedEvent {
	assert.Gt(seq, int64(0), "seq must be positive")
	assert.Is_true(runID != RunID(""), "runID must not be empty")
	assert.Not_empty(stepID, "stepID must not be empty")
	assert.Not_empty(toolName, "toolName must not be empty")
	assert.Not_empty(reason, "reason must not be empty")

	return ToolFailedEvent{
		RunID:    runID,
		StepID:   stepID,
		ToolName: toolName,
		Reason:   reason,
		Seq:      seq,
		Type:     EventTypeToolFailed,
	}
}

// FormatArtifactCreated creates a fully-formed ArtifactCreatedEvent.
func FormatArtifactCreated(seq int64, runID RunID, stepID string, path string) ArtifactCreatedEvent {
	assert.Gt(seq, int64(0), "seq must be positive")
	assert.Is_true(runID != RunID(""), "runID must not be empty")
	assert.Not_empty(stepID, "stepID must not be empty")
	assert.Not_empty(path, "path must not be empty")

	return ArtifactCreatedEvent{
		RunID:  runID,
		StepID: stepID,
		Path:   path,
		Seq:    seq,
		Type:   EventTypeArtifactCreated,
	}
}

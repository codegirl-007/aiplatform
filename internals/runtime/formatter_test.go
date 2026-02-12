package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFormatter_RunStarted verifies FormatRunStarted sets correct Type and Seq
func TestFormatter_RunStarted(t *testing.T) {
	runID := RunID("test-run")
	workspaceRoot := "/tmp/workspace"
	seq := int64(42)

	event := FormatRunStarted(seq, runID, workspaceRoot)

	assert.Equal(t, EventTypeRunStarted, event.Type)
	assert.Equal(t, seq, event.Seq)
	assert.Equal(t, runID, event.RunID)
	assert.Equal(t, workspaceRoot, event.WorkspaceRoot)
}

// TestFormatter_RunFinished verifies FormatRunFinished sets correct Type and Seq
func TestFormatter_RunFinished(t *testing.T) {
	runID := RunID("test-run")
	seq := int64(43)

	event := FormatRunFinished(seq, runID)

	assert.Equal(t, EventTypeRunFinished, event.Type)
	assert.Equal(t, seq, event.Seq)
	assert.Equal(t, runID, event.RunID)
}

// TestFormatter_RunFailed verifies FormatRunFailed sets correct Type and Seq
func TestFormatter_RunFailed(t *testing.T) {
	runID := RunID("test-run")
	reason := "test failure"
	seq := int64(44)

	event := FormatRunFailed(seq, runID, reason)

	assert.Equal(t, EventTypeRunFailed, event.Type)
	assert.Equal(t, seq, event.Seq)
	assert.Equal(t, runID, event.RunID)
	assert.Equal(t, reason, event.Reason)
}

// TestFormatter_StepStarted verifies FormatStepStarted sets correct Type and Seq
func TestFormatter_StepStarted(t *testing.T) {
	runID := RunID("test-run")
	stepID := "step-1"
	phase := Phase(1)
	seq := int64(45)

	event := FormatStepStarted(seq, runID, stepID, phase)

	assert.Equal(t, EventTypeStepStarted, event.Type)
	assert.Equal(t, seq, event.Seq)
	assert.Equal(t, runID, event.RunID)
	assert.Equal(t, stepID, event.StepID)
	assert.Equal(t, phase, event.Phase)
}

// TestFormatter_StepFinished verifies FormatStepFinished sets correct Type and Seq
func TestFormatter_StepFinished(t *testing.T) {
	runID := RunID("test-run")
	stepID := "step-1"
	phase := Phase(1)
	seq := int64(46)

	event := FormatStepFinished(seq, runID, stepID, phase)

	assert.Equal(t, EventTypeStepFinished, event.Type)
	assert.Equal(t, seq, event.Seq)
	assert.Equal(t, runID, event.RunID)
	assert.Equal(t, stepID, event.StepID)
	assert.Equal(t, phase, event.Phase)
}

// TestFormatter_StepFailed verifies FormatStepFailed sets correct Type and Seq
func TestFormatter_StepFailed(t *testing.T) {
	runID := RunID("test-run")
	stepID := "step-1"
	phase := Phase(1)
	reason := "step error"
	seq := int64(47)

	event := FormatStepFailed(seq, runID, stepID, phase, reason)

	assert.Equal(t, EventTypeStepFailed, event.Type)
	assert.Equal(t, seq, event.Seq)
	assert.Equal(t, runID, event.RunID)
	assert.Equal(t, stepID, event.StepID)
	assert.Equal(t, phase, event.Phase)
	assert.Equal(t, reason, event.Reason)
}

// TestFormatter_LLMRequested verifies FormatLLMRequested sets correct Type and Seq
func TestFormatter_LLMRequested(t *testing.T) {
	runID := RunID("test-run")
	stepID := "step-1"
	seq := int64(48)

	event := FormatLLMRequested(seq, runID, stepID)

	assert.Equal(t, EventTypeLLMRequested, event.Type)
	assert.Equal(t, seq, event.Seq)
	assert.Equal(t, runID, event.RunID)
	assert.Equal(t, stepID, event.StepID)
}

// TestFormatter_LLMResponded verifies FormatLLMResponded sets correct Type and Seq
func TestFormatter_LLMResponded(t *testing.T) {
	runID := RunID("test-run")
	stepID := "step-1"
	seq := int64(49)

	event := FormatLLMResponded(seq, runID, stepID)

	assert.Equal(t, EventTypeLLMResponded, event.Type)
	assert.Equal(t, seq, event.Seq)
	assert.Equal(t, runID, event.RunID)
	assert.Equal(t, stepID, event.StepID)
}

// TestFormatter_ToolCalled verifies FormatToolCalled sets correct Type and Seq
func TestFormatter_ToolCalled(t *testing.T) {
	runID := RunID("test-run")
	stepID := "step-1"
	toolName := "calculator"
	seq := int64(50)

	event := FormatToolCalled(seq, runID, stepID, toolName)

	assert.Equal(t, EventTypeToolCalled, event.Type)
	assert.Equal(t, seq, event.Seq)
	assert.Equal(t, runID, event.RunID)
	assert.Equal(t, stepID, event.StepID)
	assert.Equal(t, toolName, event.ToolName)
}

// TestFormatter_ToolReturned verifies FormatToolReturned sets correct Type and Seq
func TestFormatter_ToolReturned(t *testing.T) {
	runID := RunID("test-run")
	stepID := "step-1"
	toolName := "calculator"
	seq := int64(51)

	event := FormatToolReturned(seq, runID, stepID, toolName)

	assert.Equal(t, EventTypeToolReturned, event.Type)
	assert.Equal(t, seq, event.Seq)
	assert.Equal(t, runID, event.RunID)
	assert.Equal(t, stepID, event.StepID)
	assert.Equal(t, toolName, event.ToolName)
}

// TestFormatter_ToolFailed verifies FormatToolFailed sets correct Type and Seq
func TestFormatter_ToolFailed(t *testing.T) {
	runID := RunID("test-run")
	stepID := "step-1"
	toolName := "calculator"
	reason := "division by zero"
	seq := int64(52)

	event := FormatToolFailed(seq, runID, stepID, toolName, reason)

	assert.Equal(t, EventTypeToolFailed, event.Type)
	assert.Equal(t, seq, event.Seq)
	assert.Equal(t, runID, event.RunID)
	assert.Equal(t, stepID, event.StepID)
	assert.Equal(t, toolName, event.ToolName)
	assert.Equal(t, reason, event.Reason)
}

// TestFormatter_ArtifactCreated verifies FormatArtifactCreated sets correct Type and Seq
func TestFormatter_ArtifactCreated(t *testing.T) {
	runID := RunID("test-run")
	stepID := "step-1"
	path := "/tmp/artifact.txt"
	seq := int64(53)

	event := FormatArtifactCreated(seq, runID, stepID, path)

	assert.Equal(t, EventTypeArtifactCreated, event.Type)
	assert.Equal(t, seq, event.Seq)
	assert.Equal(t, runID, event.RunID)
	assert.Equal(t, stepID, event.StepID)
	assert.Equal(t, path, event.Path)
}

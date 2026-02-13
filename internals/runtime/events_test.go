package runtime

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStepStartedEvent_PhaseSerializedAsString validates phase is JSON string.
func TestStepStartedEvent_PhaseSerializedAsString(t *testing.T) {
	event := StepStartedEvent{
		RunID:  RunID("run-123"),
		StepID: "step-1",
		Phase:  PhaseDataIngestion,
		Seq:    1,
		Type:   EventTypeStepStarted,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	// Verify phase is serialized as string, not number
	assert.Contains(t, string(data), `"phase":"data_ingestion"`)
	assert.NotContains(t, string(data), `"phase":1`)
}

// TestStepFinishedEvent_PhaseSerializedAsString validates phase is JSON string.
func TestStepFinishedEvent_PhaseSerializedAsString(t *testing.T) {
	event := StepFinishedEvent{
		RunID:  RunID("run-123"),
		StepID: "step-1",
		Phase:  PhaseSignalGeneration,
		Seq:    2,
		Type:   EventTypeStepFinished,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"phase":"signal_generation"`)
	assert.NotContains(t, string(data), `"phase":2`)
}

// TestStepFailedEvent_PhaseSerializedAsString validates phase is JSON string.
func TestStepFailedEvent_PhaseSerializedAsString(t *testing.T) {
	event := StepFailedEvent{
		RunID:  RunID("run-123"),
		StepID: "step-1",
		Phase:  PhaseRiskValidation,
		Reason: "test failure",
		Seq:    3,
		Type:   EventTypeStepFailed,
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	assert.Contains(t, string(data), `"phase":"risk_validation"`)
	assert.NotContains(t, string(data), `"phase":3`)
}

// TestStepEvent_UnmarshalPhaseFromString validates events can unmarshal phase.
func TestStepEvent_UnmarshalPhaseFromString(t *testing.T) {
	jsonData := `{
		"run_id": "run-456",
		"step_id": "step-2",
		"phase": "order_execution",
		"seq": 10,
		"type": "step.started"
	}`

	var event StepStartedEvent
	err := json.Unmarshal([]byte(jsonData), &event)
	require.NoError(t, err)

	assert.Equal(t, PhaseOrderExecution, event.Phase)
	assert.Equal(t, "order_execution", event.Phase.String())
}

// TestStepEvent_UnmarshalPhaseFromNumber validates backward compat.
func TestStepEvent_UnmarshalPhaseFromNumber(t *testing.T) {
	jsonData := `{
		"run_id": "run-789",
		"step_id": "step-3",
		"phase": 2,
		"seq": 20,
		"type": "step.finished"
	}`

	var event StepFinishedEvent
	err := json.Unmarshal([]byte(jsonData), &event)
	require.NoError(t, err)

	assert.Equal(t, PhaseSignalGeneration, event.Phase)
	assert.Equal(t, "signal_generation", event.Phase.String())
}

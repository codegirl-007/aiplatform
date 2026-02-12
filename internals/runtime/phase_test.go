package runtime

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPhase_String validates String() returns correct strings for valid phases.
func TestPhase_String(t *testing.T) {
	tests := []struct {
		name  string
		phase Phase
		want  string
	}{
		{"data_ingestion", PhaseDataIngestion, "data_ingestion"},
		{"signal_generation", PhaseSignalGeneration, "signal_generation"},
		{"risk_validation", PhaseRiskValidation, "risk_validation"},
		{"order_execution", PhaseOrderExecution, "order_execution"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.phase.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestPhase_String_Panic validates String() panics on invalid phases.
func TestPhase_String_Panic(t *testing.T) {
	tests := []struct {
		name  string
		phase Phase
	}{
		{"zero value", 0},
		{"negative", Phase(-1)},
		{"too high", Phase(99)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				tt.phase.String()
			})
		})
	}
}

// TestParsePhase validates parsing strings to phases.
func TestParsePhase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Phase
	}{
		// Valid cases
		{"data_ingestion", "data_ingestion", PhaseDataIngestion},
		{"signal_generation", "signal_generation", PhaseSignalGeneration},
		{"risk_validation", "risk_validation", PhaseRiskValidation},
		{"order_execution", "order_execution", PhaseOrderExecution},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePhase(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParsePhase_Panic validates ParsePhase panics on invalid strings.
func TestParsePhase_Panic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"unknown", "unknown"},
		{"wrong case", "Data_Ingestion"},
		{"with spaces", "data_ingestion "},
		{"old name", "planner"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Panics(t, func() {
				ParsePhase(tt.input)
			})
		})
	}
}

// TestPhase_IsValidTransition validates phase transition rules per invariant 3.
func TestPhase_IsValidTransition(t *testing.T) {
	tests := []struct {
		name     string
		from     Phase
		to       Phase
		expected bool
		panics   bool
	}{
		// Same phase: valid (retries allowed)
		{"same data_ingestion", PhaseDataIngestion, PhaseDataIngestion, true, false},
		{"same order_execution", PhaseOrderExecution, PhaseOrderExecution, true, false},

		// Forward by 1: valid
		{"1→2 data to signal", PhaseDataIngestion, PhaseSignalGeneration, true, false},
		{"2→3 signal to risk", PhaseSignalGeneration, PhaseRiskValidation, true, false},
		{"3→4 risk to execution", PhaseRiskValidation, PhaseOrderExecution, true, false},

		// Skip forward: invalid (strict enforcement)
		{"1→3 skip validation", PhaseDataIngestion, PhaseRiskValidation, false, false},
		{"1→4 skip to end", PhaseDataIngestion, PhaseOrderExecution, false, false},
		{"2→4 skip execution", PhaseSignalGeneration, PhaseOrderExecution, false, false},

		// Backward: invalid (strict enforcement)
		{"2→1 backward", PhaseSignalGeneration, PhaseDataIngestion, false, false},
		{"3→2 backward", PhaseRiskValidation, PhaseSignalGeneration, false, false},
		{"4→1 backward big", PhaseOrderExecution, PhaseDataIngestion, false, false},

		// Invalid phases: should panic
		{"invalid from", 0, PhaseDataIngestion, false, true},
		{"invalid to", PhaseDataIngestion, 0, false, true},
		{"both invalid", 99, 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				assert.Panics(t, func() {
					IsValidTransition(tt.from, tt.to)
				})
			} else {
				got := IsValidTransition(tt.from, tt.to)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

// TestPhase_IsValid validates phase validity checking.
func TestPhase_IsValid(t *testing.T) {
	tests := []struct {
		phase   Phase
		isValid bool
	}{
		{0, false},
		{PhaseDataIngestion, true},
		{PhaseSignalGeneration, true},
		{PhaseRiskValidation, true},
		{PhaseOrderExecution, true},
		{Phase(5), false},
		{Phase(-1), false},
	}

	for i, tt := range tests {
		name := "invalid"
		if tt.phase.IsValid() {
			name = tt.phase.String()
		} else {
			name = fmt.Sprintf("invalid_%d", i)
		}
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.phase.IsValid())
		})
	}
}

// TestPhase_RoundTrip validates ParsePhase(String()) round-trip.
func TestPhase_RoundTrip(t *testing.T) {
	phases := []Phase{
		PhaseDataIngestion,
		PhaseSignalGeneration,
		PhaseRiskValidation,
		PhaseOrderExecution,
	}

	for _, phase := range phases {
		t.Run(phase.String(), func(t *testing.T) {
			str := phase.String()
			parsed := ParsePhase(str)
			assert.Equal(t, phase, parsed)
		})
	}
}

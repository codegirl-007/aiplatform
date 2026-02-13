package runtime

import (
	"fmt"

	"aiplatform/pkg/assert"
)

// Phase represents a stage in the strategy pipeline.
// The zero value is not a valid phase (invariant: zero values are invalid).
type Phase int

const (
	_ Phase = iota // Skip 0 - invalid/uninitialized
	PhaseDataIngestion
	PhaseSignalGeneration
	PhaseRiskValidation
	PhaseOrderExecution
)

var (
	// MinPhase and MaxPhase define valid phase range.
	MinPhase = PhaseDataIngestion
	MaxPhase = PhaseOrderExecution
)

// String returns the string representation of a valid phase.
// Panics if called on an invalid phase (strict enforcement per Tiger Style).
func (p Phase) String() string {
	switch p {
	case PhaseDataIngestion:
		return "data_ingestion"
	case PhaseSignalGeneration:
		return "signal_generation"
	case PhaseRiskValidation:
		return "risk_validation"
	case PhaseOrderExecution:
		return "order_execution"
	}
	panic(fmt.Sprintf("invalid phase: %d", p))
}

// ParsePhase derives a Phase from a string representation.
// Panics if string is unknown or invalid (strict enforcement per Tiger Style).
func ParsePhase(s string) Phase {
	switch s {
	case "data_ingestion":
		return PhaseDataIngestion
	case "signal_generation":
		return PhaseSignalGeneration
	case "risk_validation":
		return PhaseRiskValidation
	case "order_execution":
		return PhaseOrderExecution
	}
	panic(fmt.Sprintf("invalid phase string: %q", s))
}

// IsValid returns true if phase is one of the four valid phases.
func (p Phase) IsValid() bool {
	return p >= PhaseDataIngestion && p <= PhaseOrderExecution
}

// IsValidTransition checks if transition from 'from' to 'to' is allowed.
//
// Valid transitions:
//   - Same phase: allowed (retries permitted)
//   - Forward by 1: allowed (1→2, 2→3, 3→4)
//   - Any backward: NOT allowed (strict enforcement)
//   - Skip forward: NOT allowed (no 1→3, must go 1→2→3)
//
// Panics if either phase is invalid (strict enforcement).
func IsValidTransition(from, to Phase) bool {
	assert.Is_true(from.IsValid(), fmt.Sprintf("from phase must be valid, got %d", from))
	assert.Is_true(to.IsValid(), fmt.Sprintf("to phase must be valid, got %d", to))

	if from == to {
		return true
	}

	if to > from {
		return int(to) == int(from)+1
	}

	return false
}

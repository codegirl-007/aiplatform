package runtime

import (
	"encoding/json"
	"fmt"

	"aiplatform/pkg/assert"
)

// Phase represents a stage in the strategy pipeline.
// The zero value is not a valid phase (invariant: zero values are invalid).
//
// Numeric mapping is frozen per ALGO.md Invariant 3:
//
//	data_ingestion=1, signal_generation=2, risk_validation=3, order_execution=4
type Phase int

const (
	// Phase constants with explicit numeric values per ALGO.md Invariant 3.
	// These values MUST NOT change - they define the phase execution order.
	PhaseDataIngestion    Phase = 1
	PhaseSignalGeneration Phase = 2
	PhaseRiskValidation   Phase = 3
	PhaseOrderExecution   Phase = 4
)

const (
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

// tryParsePhase attempts to parse a Phase from a string, returning an error
// instead of panicking. Used internally by UnmarshalJSON.
func tryParsePhase(s string) (Phase, error) {
	switch s {
	case "data_ingestion":
		return PhaseDataIngestion, nil
	case "signal_generation":
		return PhaseSignalGeneration, nil
	case "risk_validation":
		return PhaseRiskValidation, nil
	case "order_execution":
		return PhaseOrderExecution, nil
	}
	return 0, fmt.Errorf("invalid phase string: %q", s)
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

// MarshalJSON serializes Phase as a JSON string (not a number).
// This makes the event log human-readable and stable across refactorings.
func (p Phase) MarshalJSON() ([]byte, error) {
	if !p.IsValid() {
		return nil, fmt.Errorf("cannot marshal invalid phase: %d", p)
	}
	return json.Marshal(p.String())
}

// UnmarshalJSON deserializes Phase from a JSON string.
// Accepts both string and number for backward compatibility during transitions.
func (p *Phase) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// Try numeric fallback for compatibility
		var n int
		if numErr := json.Unmarshal(data, &n); numErr != nil {
			return fmt.Errorf("phase must be string or number: %w", err)
		}
		*p = Phase(n)
		if !p.IsValid() {
			return fmt.Errorf("invalid phase number: %d", n)
		}
		return nil
	}

	// String case - use tryParsePhase to return error instead of panic
	parsed, err := tryParsePhase(s)
	if err != nil {
		return err
	}
	*p = parsed
	return nil
}

package runtime

import "errors"

type Phase int

const (
	PhasePlanner Phase = iota + 1
	PhaseExecutor
	PhaseReviewer
)

var ErrInvalidPhase = errors.New("invalid phase")

// derive a string representation of a phase
func (p Phase) String() string {
	switch p {
	case PhasePlanner:
		return "planner"
	case PhaseExecutor:
		return "executor"
	case PhaseReviewer:
		return "reviewer"
	default:
		return "unknown"
	}
}

// derive a Phase from string representation
func (p Phase) Parse(s string) (Phase, error) {
	switch s {
	case "planner":
		return PhasePlanner, nil
	case "executor":
		return PhaseExecutor, nil
	case "reviewer":
		return PhaseReviewer, nil
	default:
		return 0, ErrInvalidPhase
	}
}

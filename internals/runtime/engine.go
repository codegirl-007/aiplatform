package runtime

import (
	"github.com/google/uuid"
)

type Engine struct {
	runs map[RunID]*RunHandle
}

// points to a run's logs (jsonl file)
// logs store events
// steps are derived in events
type RunHandle struct {
	ID        RunID
	LastSeq   int64
	Terminal  bool // true if last event is run.finished or run.failed
	Phase     Phase
	Attempts  map[Phase]int
	PhaseDone map[Phase]bool
}

type RunID string

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) StartRun(workspaceRoot string) (RunID, error) {
	id := RunID("run-" + uuid.NewString())

	if e.runs[id] != nil {
		return "", nil // TODO: return an error here when it's definied
	}

	return RunID("run-" + uuid.NewString()), nil
}

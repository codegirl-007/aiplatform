package runtime

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"aiplatform/pkg/assert"
	"aiplatform/pkg/validate"
)

// Command is the interface for all engine commands.
// Commands are processed sequentially by the run loop.
type Command interface {
	command()
}

// StartRunCmd creates a new run.
type StartRunCmd struct {
	WorkspaceRoot string
	ResultCh      chan<- StartRunResult
}

func (StartRunCmd) command() {}

// StartRunResult is the result of creating a run.
type StartRunResult struct {
	ID  RunID
	Err error
}

// Engine is the runtime engine.
// All operations are processed sequentially via the command channel.
type Engine struct {
	cmdCh chan Command
}

// RunHandle tracks the state of a run.
// This is a cache of events; the event log is the source of truth.
type RunHandle struct {
	ID            RunID
	LastSeq       int64
	Terminal      bool // true if last event is run.finished or run.failed
	Phase         Phase
	WorkspaceRoot string // normalized, absolute path (symlinks resolved)
	Attempts      map[Phase]int
	PhaseDone     map[Phase]bool
}

// RunID uniquely identifies a run.
type RunID string

// NewEngine creates a new engine and starts its run loop.
func NewEngine() *Engine {
	e := &Engine{
		cmdCh: make(chan Command, 64),
	}
	go e.runLoop()
	return e
}

// runLoop processes all commands sequentially.
// This is the only goroutine that mutates engine state.
func (e *Engine) runLoop() {
	runs := make(map[RunID]*RunHandle)

	for cmd := range e.cmdCh {
		switch c := cmd.(type) {
		case StartRunCmd:
			e.handleStartRun(runs, c)
		default:
			panic(fmt.Sprintf("unknown command type: %T", cmd))
		}
	}
}

// handleStartRun processes a StartRunCmd.
func (e *Engine) handleStartRun(runs map[RunID]*RunHandle, cmd StartRunCmd) {
	// Precondition assertions (internal invariants)
	assert.Not_nil(runs, "runs map must not be nil")
	assert.Not_nil(cmd.ResultCh, "result channel must not be nil")

	// Validate user input before processing
	if err := validate.Workspace_root(cmd.WorkspaceRoot); err != nil {
		cmd.ResultCh <- StartRunResult{Err: err}
		return
	}

	normalizedPath, err := normalizeWorkspaceRoot(cmd.WorkspaceRoot)
	if err != nil {
		cmd.ResultCh <- StartRunResult{Err: err}
		return
	}

	id, err := generateRunID()
	assert.No_err(err, "failed to generate run ID")

	if _, exists := runs[id]; exists {
		// This should be impossible with proper UUID generation
		panic(fmt.Sprintf("run ID collision: %s", id))
	}

	runs[id] = &RunHandle{
		ID:            id,
		Phase:         PhaseDataIngestion,
		WorkspaceRoot: normalizedPath,
		Attempts:      make(map[Phase]int),
		PhaseDone:     make(map[Phase]bool),
	}

	cmd.ResultCh <- StartRunResult{ID: id}
}

// StartRun creates a new run with the given workspace root.
// Returns an error if the workspace root is not an absolute path.
func (e *Engine) StartRun(workspaceRoot string) (RunID, error) {
	resultCh := make(chan StartRunResult, 1)
	e.cmdCh <- StartRunCmd{
		WorkspaceRoot: workspaceRoot,
		ResultCh:      resultCh,
	}
	result := <-resultCh
	return result.ID, result.Err
}

func normalizeWorkspaceRoot(path string) (string, error) {
	// Step 1: Clean the path
	cleaned := filepath.Clean(path)

	// Step 2: Verify it's absolute
	if !filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("workspace_root must be absolute path: %s", path)
	}

	// Step 3: Resolve symlinks
	realPath, err := filepath.EvalSymlinks(cleaned)
	if err != nil {
		return "", fmt.Errorf("workspace_root contains broken symlink or doesn't exist: %w", err)
	}

	// Step 4: Verify resolved path is still absolute (should always be true)
	if !filepath.IsAbs(realPath) {
		return "", fmt.Errorf("workspace_root resolved to non-absolute path: %s", realPath)
	}

	return realPath, nil
}

// generateRunID creates a UUID v4 string using crypto/rand.
// Format: "run-xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx"
// This replaces the external github.com/google/uuid dependency.
func generateRunID() (RunID, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant is 10

	uuid := hex.EncodeToString(b)
	return RunID("run-" + uuid[:8] + "-" + uuid[8:12] + "-" + uuid[12:16] + "-" + uuid[16:20] + "-" + uuid[20:]), nil
}

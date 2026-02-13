package runtime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"aiplatform/pkg/assert"
)

// Internal request types for typed append operations.
// Each request corresponds to one event type and carries the minimal payload.

type runStartedRequest struct {
	runID         RunID
	workspaceRoot string
	resultCh      chan<- error
}

type runFinishedRequest struct {
	runID    RunID
	resultCh chan<- error
}

type runFailedRequest struct {
	runID    RunID
	reason   string
	resultCh chan<- error
}

type stepStartedRequest struct {
	runID    RunID
	stepID   string
	phase    Phase
	resultCh chan<- error
}

type stepFinishedRequest struct {
	runID    RunID
	stepID   string
	phase    Phase
	resultCh chan<- error
}

type stepFailedRequest struct {
	runID    RunID
	stepID   string
	phase    Phase
	reason   string
	resultCh chan<- error
}

type llmRequestedRequest struct {
	runID    RunID
	stepID   string
	resultCh chan<- error
}

type llmRespondedRequest struct {
	runID    RunID
	stepID   string
	resultCh chan<- error
}

type toolCalledRequest struct {
	runID    RunID
	stepID   string
	toolName string
	resultCh chan<- error
}

type toolReturnedRequest struct {
	runID    RunID
	stepID   string
	toolName string
	resultCh chan<- error
}

type toolFailedRequest struct {
	runID    RunID
	stepID   string
	toolName string
	reason   string
	resultCh chan<- error
}

type artifactCreatedRequest struct {
	runID    RunID
	stepID   string
	path     string
	resultCh chan<- error
}

// appendRequest is a union type for all append requests.
type appendRequest interface {
	isAppendRequest()
}

func (runStartedRequest) isAppendRequest()      {}
func (runFinishedRequest) isAppendRequest()     {}
func (runFailedRequest) isAppendRequest()       {}
func (stepStartedRequest) isAppendRequest()     {}
func (stepFinishedRequest) isAppendRequest()    {}
func (stepFailedRequest) isAppendRequest()      {}
func (llmRequestedRequest) isAppendRequest()    {}
func (llmRespondedRequest) isAppendRequest()    {}
func (toolCalledRequest) isAppendRequest()      {}
func (toolReturnedRequest) isAppendRequest()    {}
func (toolFailedRequest) isAppendRequest()      {}
func (artifactCreatedRequest) isAppendRequest() {}

// EventLog is an append-only log of events for a single run.
// It is safe for concurrent callers; appends are serialized internally
// by a single writer goroutine (Tiger Beetle principle: single-threaded writes).
type EventLog struct {
	// file is the open file handle for writing.
	file *os.File

	// writer provides buffering for writes.
	writer *bufio.Writer

	// encoder writes JSON to the output.
	encoder *json.Encoder

	// nextSeq is the next sequence number to assign.
	// Only touched by the writer goroutine.
	nextSeq int64

	// runID identifies which run this log belongs to.
	runID RunID

	// appendCh is the channel for enqueuing append requests.
	appendCh chan appendRequest

	// closeCh signals the writer goroutine to shut down.
	closeCh chan struct{}

	// doneCh is closed when the writer goroutine exits.
	doneCh chan struct{}

	// closed tracks whether Close has been called.
	closed atomic.Bool
}

// Open creates or opens an event log file for a run.
//
// Tiger Beetle Principle: Crash recovery is essential.
// If the file already exists, we scan to find the last sequence number
// and resume from there. This allows the engine to recover from crashes
// and continue appending events with correct sequence numbers.
func OpenEventLog(runID RunID, workspaceRoot string) (*EventLog, error) {
	// Construct the log directory path.
	logDir := filepath.Join(workspaceRoot, ".aiplatform", "logs")

	// Ensure the log directory exists.
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	// Construct the full path to the log file.
	logPath := filepath.Join(logDir, string(runID)+".jsonl")

	// Check if file exists and has content.
	// We need to know if we're resuming an existing run or starting fresh.
	fileInfo, err := os.Stat(logPath)
	isNewFile := os.IsNotExist(err) || fileInfo.Size() == 0

	// Open the file for writing.
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open event log %s: %w", logPath, err)
	}

	// Determine the starting sequence number.
	// For a new file, we start at 1.
	// For an existing file, we scan to find the last sequence number.
	nextSeq := int64(1)
	if !isNewFile {
		// File exists with content - scan to find the last sequence number.
		// This is critical for crash recovery: we need to resume with
		// the correct sequence to maintain Invariant 38 (strictly increasing).
		lastSeq, err := scanLastSeq(file)
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to scan existing log %s: %w", logPath, err)
		}
		nextSeq = lastSeq + 1
	}

	// Create a buffered writer for the file.
	writer := bufio.NewWriterSize(file, 4096)

	// Create a JSON encoder that writes to our buffered writer.
	encoder := json.NewEncoder(writer)

	// SetEscapeHTML(false) means we don't escape <, >, & as \u003c, etc.
	encoder.SetEscapeHTML(false)

	log := &EventLog{
		file:     file,
		writer:   writer,
		encoder:  encoder,
		nextSeq:  nextSeq,
		runID:    runID,
		appendCh: make(chan appendRequest, 64), // Buffered for performance
		closeCh:  make(chan struct{}),
		doneCh:   make(chan struct{}),
	}

	// Start the single writer goroutine.
	// This is the only goroutine that mutates nextSeq and writes to the log.
	go log.writerLoop()

	return log, nil
}

// scanLastSeq reads an existing log file to find the last sequence number.
// This enables crash recovery by allowing us to resume appending with
// the correct next sequence number (Invariant 38).
//
// Tiger Beetle Principle: Validate everything during recovery.
// We parse each line to ensure the log is valid before resuming.
// If the log is corrupt, we fail fast with a clear error.
func scanLastSeq(file *os.File) (int64, error) {
	// We need to read the file, but we opened it with O_APPEND|O_WRONLY.
	// We can't read from a write-only file handle, so we need to open
	// a separate read handle for scanning.
	assert.Not_nil(file, "file must not be nil")

	readFile, err := os.Open(file.Name())
	if err != nil {
		return 0, fmt.Errorf("failed to open file for reading: %w", err)
	}
	defer readFile.Close()

	scanner := bufio.NewScanner(readFile)
	var lastSeq int64 = 0
	var lineNum int

	// Scan the file line by line.
	// Each line should be a valid JSON object with a "seq" field.
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Parse just the seq field for efficiency.
		// We don't need to parse the full event, just validate seq exists and is valid.
		var envelope struct {
			Seq int64 `json:"seq"`
		}
		if err := json.Unmarshal([]byte(line), &envelope); err != nil {
			return 0, fmt.Errorf("line %d: invalid JSON: %w", lineNum, err)
		}

		// Validate sequence is strictly increasing (Invariant 38).
		if envelope.Seq <= lastSeq {
			return 0, fmt.Errorf("line %d: sequence number %d is not strictly increasing (previous: %d)",
				lineNum, envelope.Seq, lastSeq)
		}

		// Pair assertion: validate at read time (also validated at write time)
		assert.Gt(envelope.Seq, 0, "seq must be positive")
		assert.Gt(envelope.Seq, lastSeq, "seq must strictly increase")

		lastSeq = envelope.Seq
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("failed to read log file: %w", err)
	}

	// If file is empty, lastSeq will be 0, which is correct (next will be 1).
	return lastSeq, nil
}

// writerLoop is the single writer goroutine that processes all append requests.
// This ensures thread-safe, sequential appends with strict sequence ordering.
//
// Tiger Beetle Principle: Single-threaded writes eliminate race conditions
// and ensure deterministic ordering without complex locking.
//
// The writer assigns Seq and calls the formatter to create fully-formed events.
func (l *EventLog) writerLoop() {
	defer close(l.doneCh)

	for {
		select {
		case req := <-l.appendCh:
			// Process the append request by type
			err := l.processRequest(req)
			// Send result back to caller via the request's result channel
			switch r := req.(type) {
			case runStartedRequest:
				r.resultCh <- err
			case runFinishedRequest:
				r.resultCh <- err
			case runFailedRequest:
				r.resultCh <- err
			case stepStartedRequest:
				r.resultCh <- err
			case stepFinishedRequest:
				r.resultCh <- err
			case stepFailedRequest:
				r.resultCh <- err
			case llmRequestedRequest:
				r.resultCh <- err
			case llmRespondedRequest:
				r.resultCh <- err
			case toolCalledRequest:
				r.resultCh <- err
			case toolReturnedRequest:
				r.resultCh <- err
			case toolFailedRequest:
				r.resultCh <- err
			case artifactCreatedRequest:
				r.resultCh <- err
			}

		case <-l.closeCh:
			// Drain any remaining requests before shutting down
			for {
				select {
				case req := <-l.appendCh:
					err := l.processRequest(req)
					switch r := req.(type) {
					case runStartedRequest:
						r.resultCh <- err
					case runFinishedRequest:
						r.resultCh <- err
					case runFailedRequest:
						r.resultCh <- err
					case stepStartedRequest:
						r.resultCh <- err
					case stepFinishedRequest:
						r.resultCh <- err
					case stepFailedRequest:
						r.resultCh <- err
					case llmRequestedRequest:
						r.resultCh <- err
					case llmRespondedRequest:
						r.resultCh <- err
					case toolCalledRequest:
						r.resultCh <- err
					case toolReturnedRequest:
						r.resultCh <- err
					case toolFailedRequest:
						r.resultCh <- err
					case artifactCreatedRequest:
						r.resultCh <- err
					}
				default:
					// No more requests, we're done
					return
				}
			}
		}
	}
}

// processRequest handles a typed append request by assigning seq,
// calling the formatter, and encoding the event.
// This is only called from the writer goroutine.
func (l *EventLog) processRequest(req appendRequest) error {
	// Precondition assertions
	assert.Not_nil(l, "EventLog must not be nil")
	assert.Not_nil(l.file, "file must be open")
	assert.Gt(l.nextSeq, 0, "nextSeq must be positive")

	// Assign the next sequence number.
	// No atomics needed - we're the only writer.
	seq := l.nextSeq
	l.nextSeq++

	// Postcondition: seq must be positive (Invariant 38)
	assert.Gt(seq, 0, "seq must be positive after increment")

	// Call the formatter based on request type to create the fully-formed event.
	// Formatter is the only place that sets Type.
	var event Event
	switch r := req.(type) {
	case runStartedRequest:
		evt := FormatRunStarted(seq, r.runID, r.workspaceRoot)
		event = evt
	case runFinishedRequest:
		evt := FormatRunFinished(seq, r.runID)
		event = evt
	case runFailedRequest:
		evt := FormatRunFailed(seq, r.runID, r.reason)
		event = evt
	case stepStartedRequest:
		evt := FormatStepStarted(seq, r.runID, r.stepID, r.phase)
		event = evt
	case stepFinishedRequest:
		evt := FormatStepFinished(seq, r.runID, r.stepID, r.phase)
		event = evt
	case stepFailedRequest:
		evt := FormatStepFailed(seq, r.runID, r.stepID, r.phase, r.reason)
		event = evt
	case llmRequestedRequest:
		evt := FormatLLMRequested(seq, r.runID, r.stepID)
		event = evt
	case llmRespondedRequest:
		evt := FormatLLMResponded(seq, r.runID, r.stepID)
		event = evt
	case toolCalledRequest:
		evt := FormatToolCalled(seq, r.runID, r.stepID, r.toolName)
		event = evt
	case toolReturnedRequest:
		evt := FormatToolReturned(seq, r.runID, r.stepID, r.toolName)
		event = evt
	case toolFailedRequest:
		evt := FormatToolFailed(seq, r.runID, r.stepID, r.toolName, r.reason)
		event = evt
	case artifactCreatedRequest:
		evt := FormatArtifactCreated(seq, r.runID, r.stepID, r.path)
		event = evt
	default:
		return fmt.Errorf("unknown request type: %T", req)
	}

	// Encode the event as JSON and write to the buffer.
	// The encoder adds a newline after each event (JSONL format, Invariant 40).
	if err := l.encoder.Encode(event); err != nil {
		return fmt.Errorf("failed to encode event: %w", err)
	}

	// Flush the buffer to disk.
	// Without flushing, data sits in memory and could be lost on crash.
	// Tiger Beetle would fsync here for durability. For now, we flush
	// to OS cache. We can add fsync later if needed.
	if err := l.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush event to disk: %w", err)
	}

	return nil
}

// Typed append methods - these are the only public APIs for appending events.
// Each method is synchronous and blocks until the event is written or an error occurs.
//
// Tiger Beetle Principle: Events are written before state is updated.
// The caller should write the event, then update in-memory state.
// If the write fails, the state should NOT be updated.
//
// Invariant 35: Append-only - we never modify existing events.
// Invariant 38: Sequence numbers strictly increase.
// Invariant 40: JSONLines format - one JSON object per line.

// AppendRunStarted writes a run.started event.
func (l *EventLog) AppendRunStarted(runID RunID, workspaceRoot string) error {
	if l.closed.Load() {
		return fmt.Errorf("cannot append to closed log")
	}

	resultCh := make(chan error, 1)
	req := runStartedRequest{
		runID:         runID,
		workspaceRoot: workspaceRoot,
		resultCh:      resultCh,
	}

	select {
	case l.appendCh <- req:
		return <-resultCh
	case <-l.closeCh:
		return fmt.Errorf("log is closing")
	}
}

// AppendRunFinished writes a run.finished event.
func (l *EventLog) AppendRunFinished(runID RunID) error {
	if l.closed.Load() {
		return fmt.Errorf("cannot append to closed log")
	}

	resultCh := make(chan error, 1)
	req := runFinishedRequest{
		runID:    runID,
		resultCh: resultCh,
	}

	select {
	case l.appendCh <- req:
		return <-resultCh
	case <-l.closeCh:
		return fmt.Errorf("log is closing")
	}
}

// AppendRunFailed writes a run.failed event.
func (l *EventLog) AppendRunFailed(runID RunID, reason string) error {
	if l.closed.Load() {
		return fmt.Errorf("cannot append to closed log")
	}

	resultCh := make(chan error, 1)
	req := runFailedRequest{
		runID:    runID,
		reason:   reason,
		resultCh: resultCh,
	}

	select {
	case l.appendCh <- req:
		return <-resultCh
	case <-l.closeCh:
		return fmt.Errorf("log is closing")
	}
}

// AppendStepStarted writes a step.started event.
func (l *EventLog) AppendStepStarted(runID RunID, stepID string, phase Phase) error {
	if l.closed.Load() {
		return fmt.Errorf("cannot append to closed log")
	}

	resultCh := make(chan error, 1)
	req := stepStartedRequest{
		runID:    runID,
		stepID:   stepID,
		phase:    phase,
		resultCh: resultCh,
	}

	select {
	case l.appendCh <- req:
		return <-resultCh
	case <-l.closeCh:
		return fmt.Errorf("log is closing")
	}
}

// AppendStepFinished writes a step.finished event.
func (l *EventLog) AppendStepFinished(runID RunID, stepID string, phase Phase) error {
	if l.closed.Load() {
		return fmt.Errorf("cannot append to closed log")
	}

	resultCh := make(chan error, 1)
	req := stepFinishedRequest{
		runID:    runID,
		stepID:   stepID,
		phase:    phase,
		resultCh: resultCh,
	}

	select {
	case l.appendCh <- req:
		return <-resultCh
	case <-l.closeCh:
		return fmt.Errorf("log is closing")
	}
}

// AppendStepFailed writes a step.failed event.
func (l *EventLog) AppendStepFailed(runID RunID, stepID string, phase Phase, reason string) error {
	if l.closed.Load() {
		return fmt.Errorf("cannot append to closed log")
	}

	resultCh := make(chan error, 1)
	req := stepFailedRequest{
		runID:    runID,
		stepID:   stepID,
		phase:    phase,
		reason:   reason,
		resultCh: resultCh,
	}

	select {
	case l.appendCh <- req:
		return <-resultCh
	case <-l.closeCh:
		return fmt.Errorf("log is closing")
	}
}

// AppendLLMRequested writes an llm.requested event.
func (l *EventLog) AppendLLMRequested(runID RunID, stepID string) error {
	if l.closed.Load() {
		return fmt.Errorf("cannot append to closed log")
	}

	resultCh := make(chan error, 1)
	req := llmRequestedRequest{
		runID:    runID,
		stepID:   stepID,
		resultCh: resultCh,
	}

	select {
	case l.appendCh <- req:
		return <-resultCh
	case <-l.closeCh:
		return fmt.Errorf("log is closing")
	}
}

// AppendLLMResponded writes an llm.responded event.
func (l *EventLog) AppendLLMResponded(runID RunID, stepID string) error {
	if l.closed.Load() {
		return fmt.Errorf("cannot append to closed log")
	}

	resultCh := make(chan error, 1)
	req := llmRespondedRequest{
		runID:    runID,
		stepID:   stepID,
		resultCh: resultCh,
	}

	select {
	case l.appendCh <- req:
		return <-resultCh
	case <-l.closeCh:
		return fmt.Errorf("log is closing")
	}
}

// AppendToolCalled writes a tool.called event.
func (l *EventLog) AppendToolCalled(runID RunID, stepID string, toolName string) error {
	if l.closed.Load() {
		return fmt.Errorf("cannot append to closed log")
	}

	resultCh := make(chan error, 1)
	req := toolCalledRequest{
		runID:    runID,
		stepID:   stepID,
		toolName: toolName,
		resultCh: resultCh,
	}

	select {
	case l.appendCh <- req:
		return <-resultCh
	case <-l.closeCh:
		return fmt.Errorf("log is closing")
	}
}

// AppendToolReturned writes a tool.returned event.
func (l *EventLog) AppendToolReturned(runID RunID, stepID string, toolName string) error {
	if l.closed.Load() {
		return fmt.Errorf("cannot append to closed log")
	}

	resultCh := make(chan error, 1)
	req := toolReturnedRequest{
		runID:    runID,
		stepID:   stepID,
		toolName: toolName,
		resultCh: resultCh,
	}

	select {
	case l.appendCh <- req:
		return <-resultCh
	case <-l.closeCh:
		return fmt.Errorf("log is closing")
	}
}

// AppendToolFailed writes a tool.failed event.
func (l *EventLog) AppendToolFailed(runID RunID, stepID string, toolName string, reason string) error {
	if l.closed.Load() {
		return fmt.Errorf("cannot append to closed log")
	}

	resultCh := make(chan error, 1)
	req := toolFailedRequest{
		runID:    runID,
		stepID:   stepID,
		toolName: toolName,
		reason:   reason,
		resultCh: resultCh,
	}

	select {
	case l.appendCh <- req:
		return <-resultCh
	case <-l.closeCh:
		return fmt.Errorf("log is closing")
	}
}

// AppendArtifactCreated writes an artifact.created event.
func (l *EventLog) AppendArtifactCreated(runID RunID, stepID string, path string) error {
	if l.closed.Load() {
		return fmt.Errorf("cannot append to closed log")
	}

	resultCh := make(chan error, 1)
	req := artifactCreatedRequest{
		runID:    runID,
		stepID:   stepID,
		path:     path,
		resultCh: resultCh,
	}

	select {
	case l.appendCh <- req:
		return <-resultCh
	case <-l.closeCh:
		return fmt.Errorf("log is closing")
	}
}

// Close finalizes the event log.
//
// This should be called when the run completes (run.finished or run.failed).
// It signals the writer goroutine to stop, drains any pending append requests,
// flushes buffered data, and closes the file.
//
// Tiger Beetle Principle: Clean shutdown is important for data integrity.
// We ensure all queued events are written before closing.
func (l *EventLog) Close() error {
	// Mark as closed (idempotent)
	if !l.closed.CompareAndSwap(false, true) {
		return fmt.Errorf("log already closed")
	}

	// Signal writer goroutine to stop
	close(l.closeCh)

	// Wait for writer goroutine to finish draining
	<-l.doneCh

	// Flush any remaining buffered data.
	if err := l.writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush final events: %w", err)
	}

	// Close the file.
	// This releases the file descriptor and ensures all data is persisted.
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close event log: %w", err)
	}

	return nil
}

// Path returns the file path of the log.
// Useful for debugging and error messages.
func (l *EventLog) Path() string {
	assert.Not_nil(l, "EventLog must not be nil")
	assert.Not_nil(l.file, "file must be open")
	return l.file.Name()
}

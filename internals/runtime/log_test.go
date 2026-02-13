package runtime

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventLog_ConcurrentAppends tests that multiple goroutines can safely
// append events concurrently and that all events are written with correct
// sequence numbers.
//
// Invariant 38: Sequence numbers strictly increase.
// This test verifies thread-safety and strict ordering.
func TestEventLog_ConcurrentAppends(t *testing.T) {
	// Create temporary workspace
	workspaceRoot := t.TempDir()

	// Open event log
	runID := RunID("test-concurrent-run-001")
	log, err := OpenEventLog(runID, workspaceRoot)
	require.NoError(t, err)
	defer log.Close()

	// Test parameters
	numGoroutines := 10
	eventsPerGoroutine := 100
	totalEvents := numGoroutines * eventsPerGoroutine

	// Launch concurrent writers
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < eventsPerGoroutine; j++ {
				err := log.AppendRunStarted(runID, workspaceRoot)
				require.NoError(t, err)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close the log to flush everything
	err = log.Close()
	require.NoError(t, err)

	// Verify the log file
	logPath := filepath.Join(workspaceRoot, ".aiplatform", "logs", string(runID)+".jsonl")

	// Read and validate all events
	file, err := os.Open(logPath)
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	lastSeq := int64(0)

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()

		// Parse the event
		var envelope struct {
			Seq  int64     `json:"seq"`
			Type EventType `json:"type"`
		}
		err := json.Unmarshal([]byte(line), &envelope)
		require.NoError(t, err, "line %d: invalid JSON", lineCount)

		// Verify sequence is strictly increasing
		assert.Greater(t, envelope.Seq, lastSeq, "line %d: sequence must strictly increase", lineCount)
		assert.Equal(t, lastSeq+1, envelope.Seq, "line %d: sequence must be sequential", lineCount)

		// Verify event type
		assert.Equal(t, EventTypeRunStarted, envelope.Type, "line %d: wrong event type", lineCount)

		lastSeq = envelope.Seq
	}

	require.NoError(t, scanner.Err())

	// Verify we got all events
	assert.Equal(t, totalEvents, lineCount, "should have written all events")
	assert.Equal(t, int64(totalEvents), lastSeq, "last sequence should match event count")
}

// TestEventLog_CloseWhileAppending tests that closing the log while appends
// are in flight behaves correctly: the close drains all pending appends,
// and subsequent append attempts fail.
func TestEventLog_CloseWhileAppending(t *testing.T) {
	// Create temporary workspace
	workspaceRoot := t.TempDir()

	// Open event log
	runID := RunID("test-close-run-001")
	log, err := OpenEventLog(runID, workspaceRoot)
	require.NoError(t, err)

	// First, write some events successfully to ensure the log works
	numInitialEvents := 10
	for i := 0; i < numInitialEvents; i++ {
		err := log.AppendRunStarted(runID, workspaceRoot)
		require.NoError(t, err)
	}

	// Now start goroutines that append events concurrently
	numGoroutines := 5
	eventsPerGoroutine := 20

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Signal to start all goroutines at once
	startCh := make(chan struct{})

	// Track errors from append calls
	var mu sync.Mutex
	appendErrors := 0
	successfulAppends := 0

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			// Wait for signal to start
			<-startCh

			for j := 0; j < eventsPerGoroutine; j++ {
				err := log.AppendRunStarted(runID, workspaceRoot)
				mu.Lock()
				if err != nil {
					appendErrors++
				} else {
					successfulAppends++
				}
				mu.Unlock()
			}
		}(i)
	}

	// Start all goroutines
	close(startCh)

	// Give goroutines a moment to start appending
	// (not strictly necessary but makes the test more deterministic)
	// We want to close while some appends are definitely in flight
	// Note: This is a test-only sleep; production code doesn't need this

	// Close the log - this should drain all pending appends in the channel
	err = log.Close()
	require.NoError(t, err)

	// Wait for all goroutines to finish
	wg.Wait()

	// Try to append after close - should fail
	err = log.AppendRunStarted(runID, workspaceRoot)
	assert.Error(t, err, "append after close should fail")
	assert.Contains(t, err.Error(), "closed", "error should mention closed log")

	// Verify the log file has valid content
	logPath := filepath.Join(workspaceRoot, ".aiplatform", "logs", string(runID)+".jsonl")

	file, err := os.Open(logPath)
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	lastSeq := int64(0)

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()

		var envelope struct {
			Seq int64 `json:"seq"`
		}
		err := json.Unmarshal([]byte(line), &envelope)
		require.NoError(t, err, "line %d: invalid JSON", lineCount)

		// Verify sequence is strictly increasing
		assert.Greater(t, envelope.Seq, lastSeq, "line %d: sequence must strictly increase", lineCount)
		lastSeq = envelope.Seq
	}

	require.NoError(t, scanner.Err())

	// We should have written at least the initial events
	assert.GreaterOrEqual(t, lineCount, numInitialEvents, "should have written at least initial events")

	// Total appends attempted: initial + concurrent goroutines
	totalAttempted := numInitialEvents + (numGoroutines * eventsPerGoroutine)

	// Verify accounting: successful appends should match line count
	mu.Lock()
	actualSuccessful := numInitialEvents + successfulAppends
	mu.Unlock()

	assert.Equal(t, lineCount, actualSuccessful, "line count should match successful appends")

	t.Logf("Total attempted: %d, Successful: %d, Failed: %d, Lines written: %d",
		totalAttempted, actualSuccessful, appendErrors, lineCount)
}

// TestEventLog_DoubleClose tests that closing a log twice returns an error
// and doesn't cause panics.
func TestEventLog_DoubleClose(t *testing.T) {
	workspaceRoot := t.TempDir()

	runID := RunID("test-double-close-001")
	log, err := OpenEventLog(runID, workspaceRoot)
	require.NoError(t, err)

	// First close should succeed
	err = log.Close()
	require.NoError(t, err)

	// Second close should fail
	err = log.Close()
	assert.Error(t, err, "double close should return error")
	assert.Contains(t, err.Error(), "already closed", "error should mention already closed")
}

// TestEventLog_AppendAfterClose verifies that appending to a closed log
// returns a clear error.
func TestEventLog_AppendAfterClose(t *testing.T) {
	workspaceRoot := t.TempDir()

	runID := RunID("test-append-after-close-001")
	log, err := OpenEventLog(runID, workspaceRoot)
	require.NoError(t, err)

	// Close the log
	err = log.Close()
	require.NoError(t, err)

	// Try to append - should fail
	err = log.AppendRunStarted(runID, workspaceRoot)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed", "error should mention closed log")
}

// TestEventLog_SequenceRecovery tests that reopening an existing log
// correctly resumes with the next sequence number.
func TestEventLog_SequenceRecovery(t *testing.T) {
	workspaceRoot := t.TempDir()
	runID := RunID("test-recovery-run-001")

	// First session: write some events
	log1, err := OpenEventLog(runID, workspaceRoot)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		err := log1.AppendRunStarted(runID, workspaceRoot)
		require.NoError(t, err)
	}

	err = log1.Close()
	require.NoError(t, err)

	// Second session: reopen and write more events
	log2, err := OpenEventLog(runID, workspaceRoot)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		err := log2.AppendRunStarted(runID, workspaceRoot)
		require.NoError(t, err)
	}

	err = log2.Close()
	require.NoError(t, err)

	// Verify the log has 15 events with correct sequences
	logPath := filepath.Join(workspaceRoot, ".aiplatform", "logs", string(runID)+".jsonl")

	file, err := os.Open(logPath)
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	lastSeq := int64(0)

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()

		var envelope struct {
			Seq int64 `json:"seq"`
		}
		err := json.Unmarshal([]byte(line), &envelope)
		require.NoError(t, err)

		assert.Equal(t, lastSeq+1, envelope.Seq, "line %d: sequence must be sequential", lineCount)
		lastSeq = envelope.Seq
	}

	require.NoError(t, scanner.Err())
	assert.Equal(t, 15, lineCount, "should have 15 events total")
	assert.Equal(t, int64(15), lastSeq, "last sequence should be 15")
}

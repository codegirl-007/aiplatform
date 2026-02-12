package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestInvariant_1_RunIDUniqueness verifies that RunIDs are unique within the registry.
// Invariant 1: Run IDs are UUID v4 strings. RunID must be unique within the run registry.
// [EXEC] Enforced at creation time.
func TestInvariant_1_RunIDUniqueness(t *testing.T) {
	e := NewEngine()

	// Create a temp directory for the test
	tempDir := t.TempDir()

	// Create multiple runs and collect their IDs
	ids := make(map[RunID]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create 100 runs concurrently to test thread-safety of ID generation
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := e.StartRun(tempDir)
			if err != nil {
				t.Errorf("Failed to create run: %v", err)
				return
			}

			mu.Lock()
			defer mu.Unlock()

			if ids[id] {
				t.Errorf("Duplicate RunID generated: %s", id)
			}
			ids[id] = true

			// Verify ID format: must start with "run-" followed by UUID
			if !strings.HasPrefix(string(id), "run-") {
				t.Errorf("RunID must start with 'run-': got %s", id)
			}

			// Verify UUID portion (should be 36 chars after "run-")
			uuidPart := string(id)[4:]
			if len(uuidPart) != 36 {
				t.Errorf("RunID UUID portion must be 36 characters: got %d", len(uuidPart))
			}
		}()
	}

	wg.Wait()

	if len(ids) != 100 {
		t.Errorf("Expected 100 unique RunIDs, got %d", len(ids))
	}
}

// TestInvariant_4a_WorkspaceRootAbsolute verifies that workspace_root must be absolute.
// Invariant 4a: workspace_root must be an absolute path as determined by the host OS path rules.
// [EXEC] Run creation fails if workspace_root is invalid.
func TestInvariant_4a_WorkspaceRootAbsolute(t *testing.T) {
	e := NewEngine()

	// Create temp directories that actually exist
	tempDir := t.TempDir()
	workspace1 := filepath.Join(tempDir, "workspace1")
	workspace2 := filepath.Join(tempDir, "home", "user", "projects", "myapp")
	if err := os.MkdirAll(workspace1, 0755); err != nil {
		t.Fatalf("Failed to create test workspace: %v", err)
	}
	if err := os.MkdirAll(workspace2, 0755); err != nil {
		t.Fatalf("Failed to create nested test workspace: %v", err)
	}

	tests := []struct {
		name          string
		workspaceRoot string
		shouldFail    bool
	}{
		{
			name:          "absolute_unix_path",
			workspaceRoot: workspace1,
			shouldFail:    false,
		},
		{
			name:          "absolute_unix_path_nested",
			workspaceRoot: workspace2,
			shouldFail:    false,
		},
		{
			name:          "relative_path_simple",
			workspaceRoot: "relative/path",
			shouldFail:    true,
		},
		{
			name:          "relative_path_current_dir",
			workspaceRoot: "./workspace",
			shouldFail:    true,
		},
		{
			name:          "relative_path_parent_dir",
			workspaceRoot: "../workspace",
			shouldFail:    true,
		},
		{
			name:          "empty_path",
			workspaceRoot: "",
			shouldFail:    true,
		},
		{
			name:          "single_component",
			workspaceRoot: "workspace",
			shouldFail:    true,
		},
		{
			name:          "nonexistent_path",
			workspaceRoot: "/tmp/nonexistent_workspace_12345",
			shouldFail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := e.StartRun(tt.workspaceRoot)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error for workspaceRoot='%s', but got nil (id=%s)",
						tt.workspaceRoot, id)
				}
				// Verify error mentions validation failure
				if err != nil && !strings.Contains(err.Error(), "must be absolute") && !strings.Contains(err.Error(), "must not be empty") && !strings.Contains(err.Error(), "doesn't exist") {
					t.Errorf("Error should mention validation failure, got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected success for workspaceRoot='%s', but got error: %v",
						tt.workspaceRoot, err)
				}
				if id == "" {
					t.Error("Expected non-empty RunID for valid workspace root")
				}
			}
		})
	}
}

// TestInvariant_4a_WorkspaceRootNormalized verifies that normalization is handled.
// Invariant 4a: workspace_root must be normalized before use.
func TestInvariant_4a_WorkspaceRootNormalized(t *testing.T) {
	e := NewEngine()

	// Create a temp directory
	tempDir := t.TempDir()

	// Test with redundant slashes - should be normalized to the clean path
	pathWithSlashes := tempDir + "//"
	id, err := e.StartRun(pathWithSlashes)
	if err != nil {
		t.Fatalf("Path with redundant slashes should be normalized: %v", err)
	}
	if id == "" {
		t.Error("Got empty RunID for path with redundant slashes")
	}
}

// TestNewEngine_CreatesValidEngine verifies engine initialization.
func TestNewEngine_CreatesValidEngine(t *testing.T) {
	e := NewEngine()

	if e == nil {
		t.Fatal("NewEngine returned nil")
	}

	if e.cmdCh == nil {
		t.Fatal("Engine cmdCh is nil")
	}
}

package validate

import (
	"testing"
)

func TestNotEmpty_Success(t *testing.T) {
	err := not_empty("hello", "field")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestNotEmpty_Failure(t *testing.T) {
	err := not_empty("", "field")
	if err == nil {
		t.Error("Expected error for empty string")
	}
	if !contains(err.Error(), "must not be empty") {
		t.Errorf("Expected 'must not be empty' in error, got: %v", err)
	}
}

func TestNotEmpty_WhitespaceOnly(t *testing.T) {
	err := not_empty("   ", "field")
	if err == nil {
		t.Error("Expected error for whitespace-only string")
	}
}

func TestAbsolutePath_Success(t *testing.T) {
	err := absolute_path("/tmp/workspace")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestAbsolutePath_Failure(t *testing.T) {
	err := absolute_path("relative/path")
	if err == nil {
		t.Error("Expected error for relative path")
	}
	if !contains(err.Error(), "must be absolute") {
		t.Errorf("Expected 'must be absolute' in error, got: %v", err)
	}
}

func TestMaxLength_Success(t *testing.T) {
	err := max_length("hello", 10, "field")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestMaxLength_Failure(t *testing.T) {
	err := max_length("hello world", 5, "field")
	if err == nil {
		t.Error("Expected error for string exceeding max length")
	}
	if !contains(err.Error(), "exceeds maximum length") {
		t.Errorf("Expected 'exceeds maximum length' in error, got: %v", err)
	}
}

func TestWorkspaceRoot_Success(t *testing.T) {
	err := Workspace_root("/tmp/workspace")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestWorkspaceRoot_Empty(t *testing.T) {
	err := Workspace_root("")
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestWorkspaceRoot_Relative(t *testing.T) {
	err := Workspace_root("relative/path")
	if err == nil {
		t.Error("Expected error for relative path")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

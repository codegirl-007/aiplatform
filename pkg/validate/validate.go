package validate

import (
	"fmt"
	"path/filepath"
	"strings"
)

// not_empty validates that a string is not empty.
func not_empty(s, field string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("%s must not be empty", field)
	}
	return nil
}

// absolute_path validates that a path is absolute.
func absolute_path(path string) error {
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute: %s", path)
	}
	return nil
}

// max_length validates that a string does not exceed max length.
func max_length(s string, max int, field string) error {
	if len(s) > max {
		return fmt.Errorf("%s exceeds maximum length of %d", field, max)
	}
	return nil
}

// Workspace_root validates a workspace root path.
// Combines: not empty, absolute path, exists.
func Workspace_root(path string) error {
	if err := not_empty(path, "workspace_root"); err != nil {
		return err
	}
	if err := absolute_path(path); err != nil {
		return err
	}
	return nil
}

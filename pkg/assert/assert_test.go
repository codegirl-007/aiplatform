package assert

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// assertPanics checks that f panics with expected substring in message
func assertPanics(t *testing.T, expectedSubstring string, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("%v", r)
			if !strings.Contains(msg, expectedSubstring) {
				t.Errorf("Expected panic containing %q, got %q", expectedSubstring, msg)
			}
		} else {
			t.Error("Expected panic but none occurred")
		}
	}()
	f()
}

// assertNotPanics checks that f does not panic
func assertNotPanics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Expected no panic but got: %v", r)
		}
	}()
	f()
}

func TestIsTrue_Success(t *testing.T) {
	assertNotPanics(t, func() {
		Is_true(true, "should not panic")
	})
}

func TestIsTrue_Failure(t *testing.T) {
	assertPanics(t, "should panic", func() {
		Is_true(false, "should panic")
	})
}

func TestNotNil_Success(t *testing.T) {
	x := "hello"
	assertNotPanics(t, func() {
		Not_nil(&x, "should not panic")
	})
}

func TestNotNil_Failure(t *testing.T) {
	assertPanics(t, "should be non-nil", func() {
		Not_nil(nil, "should be non-nil")
	})
}

func TestNotEmpty_Success(t *testing.T) {
	assertNotPanics(t, func() {
		Not_empty("hello", "should not panic")
	})
}

func TestNotEmpty_Failure(t *testing.T) {
	assertPanics(t, "should not be empty", func() {
		Not_empty("", "should not be empty")
	})
}

func TestEq_Success(t *testing.T) {
	assertNotPanics(t, func() {
		Eq(42, 42, "should be equal")
		Eq("hello", "hello", "strings should be equal")
		Eq(true, true, "booleans should be equal")
	})
}

func TestEq_Failure(t *testing.T) {
	assertPanics(t, "should be equal", func() {
		Eq(1, 2, "should be equal")
	})
}

func TestGt_Success(t *testing.T) {
	assertNotPanics(t, func() {
		Gt(10, 5, "10 should be greater than 5")
		Gt(1, 0, "1 should be greater than 0")
		Gt(-5, -10, "-5 should be greater than -10")
	})
}

func TestGt_Failure(t *testing.T) {
	assertPanics(t, "should be greater", func() {
		Gt(5, 5, "should be greater")
	})
}

func TestGt_FailureLess(t *testing.T) {
	assertPanics(t, "should be greater", func() {
		Gt(3, 5, "should be greater")
	})
}

func TestNoErr_Success(t *testing.T) {
	assertNotPanics(t, func() {
		No_err(nil, "should not panic on nil error")
	})
}

func TestNoErr_Failure(t *testing.T) {
	assertPanics(t, "should be nil", func() {
		No_err(errors.New("some error"), "should be nil")
	})
}

func TestLocationIsIncluded(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("%v", r)
			if !strings.Contains(msg, "assert_test.go") {
				t.Errorf("Expected location info in panic message, got: %s", msg)
			}
			if !strings.Contains(msg, "test message") {
				t.Errorf("Expected custom message in panic, got: %s", msg)
			}
		} else {
			t.Error("Expected panic")
		}
	}()
	Is_true(false, "test message")
}

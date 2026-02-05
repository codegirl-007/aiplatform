package assert

import (
	"fmt"
	"runtime"
)

func getLocation() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown:0"
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func Is_true(cond bool, msg string) {
	if !cond {
		panic(fmt.Sprintf("assertion failed at %s: %s", getLocation(), msg))
	}
}

func Not_nil(ptr interface{}, msg string) {
	if ptr == nil {
		panic(fmt.Sprintf("assertion failed at %s: %s", getLocation(), msg))
	}
}

func Not_empty(s string, msg string) {
	if s == "" {
		panic(fmt.Sprintf("assertion failed at %s: %s", getLocation(), msg))
	}
}

func Eq[T comparable](a, b T, msg string) {
	if a != b {
		panic(fmt.Sprintf("assertion failed at %s: %s", getLocation(), msg))
	}
}

func Gt(a, b int64, msg string) {
	if a <= b {
		panic(fmt.Sprintf("assertion failed at %s: %s", getLocation(), msg))
	}
}

func No_err(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("assertion failed at %s: %s", getLocation(), msg))
	}
}

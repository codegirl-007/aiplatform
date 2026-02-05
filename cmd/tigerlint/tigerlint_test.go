package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestCountAssertions(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name: "two assertions",
			code: `
				package test
				import "aiplatform/pkg/assert"
				func foo() {
					assert.Not_nil(x, "x")
					assert.Gt(y, 0, "y")
				}
			`,
			expected: 2,
		},
		{
			name: "no assertions",
			code: `
				package test
				func bar() {
					x := 1 + 2
					println(x)
				}
			`,
			expected: 0,
		},
		{
			name: "assertion in nested block",
			code: `
				package test
				import "aiplatform/pkg/assert"
				func baz() {
					if true {
						assert.True(x, "x")
					}
				}
			`,
			expected: 1,
		},
		{
			name: "multiple assertions",
			code: `
				package test
				import "aiplatform/pkg/assert"
				func qux() {
					assert.Not_nil(a, "a")
					assert.Not_nil(b, "b")
					assert.Not_nil(c, "c")
					assert.Gt(d, 0, "d")
				}
			`,
			expected: 4,
		},
		{
			name: "assertion in loop",
			code: `
				package test
				import "aiplatform/pkg/assert"
				func loop() {
					for i := 0; i < 10; i++ {
						assert.Gt(i, -1, "i")
					}
				}
			`,
			expected: 1,
		},
		{
			name: "no false positives for non-assert calls",
			code: `
				package test
				import "fmt"
				func noAssert() {
					fmt.Println("hello")
					some.Other_call()
				}
			`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			var funcDecl *ast.FuncDecl
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					funcDecl = fn
					break
				}
			}

			if funcDecl == nil {
				t.Fatal("No function declaration found")
			}

			got := countAssertions(funcDecl)
			if got != tt.expected {
				t.Errorf("countAssertions() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestCheckUnboundedLoops(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name: "infinite for loop",
			code: `
				package test
				func infinite() {
					for {
						println("loop")
					}
				}
			`,
			expected: 1,
		},
		{
			name: "bounded for loop",
			code: `
				package test
				func bounded() {
					for i := 0; i < 10; i++ {
						println(i)
					}
				}
			`,
			expected: 0,
		},
		{
			name: "for with condition only",
			code: `
				package test
				func conditionOnly() {
					for running {
						println("running")
					}
				}
			`,
			expected: 1,
		},
		{
			name: "for range loop",
			code: `
				package test
				func rangeLoop() {
					items := []int{1, 2, 3}
					for _, item := range items {
						println(item)
					}
				}
			`,
			expected: 0, // range loops are bounded by the collection
		},
		{
			name: "no loops",
			code: `
				package test
				func noLoop() {
					println("no loop")
				}
			`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			var funcDecl *ast.FuncDecl
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					funcDecl = fn
					break
				}
			}

			if funcDecl == nil {
				t.Fatal("No function declaration found")
			}

			got := checkUnboundedLoops(funcDecl)
			if len(got) != tt.expected {
				t.Errorf("checkUnboundedLoops() returned %d issues, want %d", len(got), tt.expected)
			}
		})
	}
}

func TestCheckCompoundConditions(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name: "compound condition with &&",
			code: `
				package test
				func andCondition() {
					if a > 0 && b < 10 {
						println("ok")
					}
				}
			`,
			expected: 1,
		},
		{
			name: "compound condition with ||",
			code: `
				package test
				func orCondition() {
					if a > 0 || b < 10 {
						println("ok")
					}
				}
			`,
			expected: 1,
		},
		{
			name: "simple condition",
			code: `
				package test
				func simple() {
					if a > 0 {
						println("ok")
					}
				}
			`,
			expected: 0,
		},
		{
			name: "nested compound condition",
			code: `
				package test
				func nested() {
					if (a > 0 && b < 10) || c == 5 {
						println("ok")
					}
				}
			`,
			expected: 1,
		},
		{
			name: "multiple if statements",
			code: `
				package test
				func multiple() {
					if a > 0 {
						if b < 10 {
							println("ok")
						}
					}
				}
			`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			var funcDecl *ast.FuncDecl
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					funcDecl = fn
					break
				}
			}

			if funcDecl == nil {
				t.Fatal("No function declaration found")
			}

			got := checkCompoundConditions(funcDecl)
			if len(got) != tt.expected {
				t.Errorf("checkCompoundConditions() returned %d issues, want %d", len(got), tt.expected)
			}
		})
	}
}

func TestCheckWeakComments(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name: "weak comment starting with This",
			code: `
				// This is a weak package comment
				package test
				func foo() {}
			`,
			expected: 1,
		},
		{
			name: "weak comment starting with The",
			code: `
				// The package purpose is unclear
				package test
				func bar() {}
			`,
			expected: 1,
		},
		{
			name: "weak comment starting with Ensure",
			code: `
				// Ensure the package invariants hold
				package test
				func baz() {}
			`,
			expected: 1,
		},
		{
			name: "good comment",
			code: `
				// Package test provides checksum utilities.
				package test
				func qux() {}
			`,
			expected: 0,
		},
		{
			name: "no comment",
			code: `
				package test
				func nocmt() {}
			`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			got := checkComments(file)
			if len(got) != tt.expected {
				t.Errorf("checkComments() returned %d issues, want %d", len(got), tt.expected)
			}
		})
	}
}

func TestAnalyzeFile(t *testing.T) {
	// Create a temporary file for testing
	code := `package test

import "aiplatform/pkg/assert"

// This is a weak comment
func TestFunc() {
	assert.Not_nil(x, "x")
	
	if a > 0 && b < 10 {
		println("compound")
	}
	
	for {
		println("infinite")
	}
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// We can't directly test AnalyzeFile without writing to disk,
	// but we can test the individual components
	_ = file

	// Test that FindGoFiles works correctly
	t.Run("FindGoFiles ignores test files", func(t *testing.T) {
		// This is a simple test to ensure the function exists and works
		// We'll just verify it doesn't panic
		files, err := FindGoFiles([]string{"."})
		// May return error or empty, but shouldn't panic
		_ = files
		_ = err
	})
}

func TestFindGoFiles(t *testing.T) {
	t.Run("handles directory", func(t *testing.T) {
		// Test that FindGoFiles handles directories
		files, err := FindGoFiles([]string{"."})
		// Don't fail, just verify it runs
		_ = files
		_ = err
	})

	t.Run("handles dotdotdot pattern", func(t *testing.T) {
		files, err := FindGoFiles([]string{"./..."})
		// Don't fail, just verify it runs
		_ = files
		_ = err
	})
}

func TestReporterFunctions(t *testing.T) {
	// Test color functions
	t.Run("color functions work", func(t *testing.T) {
		_ = green("test")
		_ = yellow("test")
		_ = cyan("test")
		_ = red("test")
		_ = bold("test")
	})

	// Test PrintReport with empty stats
	t.Run("PrintReport handles empty stats", func(t *testing.T) {
		stats := []FunctionStats{}
		PrintReport(stats)
		// Should not panic
	})

	// Test PrintReport with good stats
	t.Run("PrintReport handles good stats", func(t *testing.T) {
		stats := []FunctionStats{
			{
				Name:       "GoodFunc",
				File:       "test.go",
				Line:       10,
				Assertions: 5,
				Issues:     []Issue{},
			},
		}
		PrintReport(stats)
		// Should not panic and should show green checkmark
	})

	// Test PrintReport with bad stats
	t.Run("PrintReport handles bad stats", func(t *testing.T) {
		stats := []FunctionStats{
			{
				Name:       "BadFunc",
				File:       "test.go",
				Line:       20,
				Assertions: 0,
				Issues: []Issue{
					{
						Type:       "unbounded-loop",
						Line:       25,
						Message:    "Unbounded loop detected",
						Suggestion: "Add limit",
					},
				},
			},
		}
		PrintReport(stats)
		// Should not panic
	})

	// Test PrintSummary
	t.Run("PrintSummary works", func(t *testing.T) {
		stats := []FunctionStats{
			{Name: "Func1", Assertions: 3},
			{Name: "Func2", Assertions: 1},
		}
		PrintSummary(stats, 1)
		// Should not panic
	})
}

func TestShouldIgnoreDir(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		expected bool
	}{
		{"vendor", "vendor", true},
		{"node_modules", "node_modules", true},
		{".git", ".git", true},
		{"src", "src", false},
		{"cmd", "cmd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldIgnoreDir(tt.dir)
			if got != tt.expected {
				t.Errorf("shouldIgnoreDir(%q) = %v, want %v", tt.dir, got, tt.expected)
			}
		})
	}
}

func TestHasCompoundCondition(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name: "simple condition",
			code: `package test
			func test() {
				if a > 0 {}
			}`,
			expected: false,
		},
		{
			name: "compound with &&",
			code: `package test
			func test() {
				if a > 0 && b < 10 {}
			}`,
			expected: true,
		},
		{
			name: "compound with ||",
			code: `package test
			func test() {
				if a > 0 || b < 10 {}
			}`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			// Find the if statement
			var ifStmt *ast.IfStmt
			ast.Inspect(file, func(n ast.Node) bool {
				if stmt, ok := n.(*ast.IfStmt); ok {
					ifStmt = stmt
					return false
				}
				return true
			})

			if ifStmt == nil {
				t.Fatal("No if statement found")
			}

			got := hasCompoundCondition(ifStmt.Cond)
			if got != tt.expected {
				t.Errorf("hasCompoundCondition() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFullFileAnalysis(t *testing.T) {
	code := `// This is a weak comment at file level
package test

import "aiplatform/pkg/assert"

// ProcessData processes the input data
func ProcessData(input []byte) error {
	assert.Not_nil(input, "input")
	assert.Gt(len(input), 0, "input length")
	
	// Bounded loop - should not be flagged
	for i := 0; i < len(input); i++ {
		input[i] = input[i] ^ 0xFF
	}
	
	// Simple condition - should not be flagged
	if len(input) > 100 {
		return nil
	}
	
	return nil
}

// Ensure validation is done
func ValidateConfig(cfg *Config) bool {
	// No assertions here
	
	// Compound condition - should be flagged
	if cfg != nil && cfg.Name != "" {
		return true
	}
	
	return false
}

// Infinite loop function
func ProcessQueue() {
	for {
		// Unbounded loop - should be flagged
		item := queue.Get()
		if item == nil {
			break
		}
	}
}

type Config struct {
	Name string
}

var queue = &Queue{}
type Queue struct{}
func (q *Queue) Get() interface{} { return nil }
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Test file-level comments
	t.Run("detects weak comments", func(t *testing.T) {
		issues := checkComments(file)
		foundWeak := false
		for _, issue := range issues {
			if issue.Type == "weak-comment" && strings.Contains(issue.Message, "This") {
				foundWeak = true
				break
			}
		}
		if !foundWeak {
			t.Error("Expected to find weak comment starting with 'This'")
		}
	})

	// Find and test each function
	var funcs []*ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			funcs = append(funcs, fn)
		}
	}

	t.Run("ProcessData has correct assertions", func(t *testing.T) {
		for _, fn := range funcs {
			if fn.Name.Name == "ProcessData" {
				count := countAssertions(fn)
				if count != 2 {
					t.Errorf("ProcessData: expected 2 assertions, got %d", count)
				}

				issues := checkUnboundedLoops(fn)
				if len(issues) != 0 {
					t.Errorf("ProcessData: expected 0 unbounded loop issues, got %d", len(issues))
				}

				issues = checkCompoundConditions(fn)
				if len(issues) != 0 {
					t.Errorf("ProcessData: expected 0 compound condition issues, got %d", len(issues))
				}
				return
			}
		}
		t.Error("ProcessData function not found")
	})

	t.Run("ValidateConfig has issues", func(t *testing.T) {
		for _, fn := range funcs {
			if fn.Name.Name == "ValidateConfig" {
				count := countAssertions(fn)
				if count != 0 {
					t.Errorf("ValidateConfig: expected 0 assertions, got %d", count)
				}

				issues := checkCompoundConditions(fn)
				if len(issues) != 1 {
					t.Errorf("ValidateConfig: expected 1 compound condition issue, got %d", len(issues))
				}

				return
			}
		}
		t.Error("ValidateConfig function not found")
	})

	t.Run("ProcessQueue has unbounded loop", func(t *testing.T) {
		for _, fn := range funcs {
			if fn.Name.Name == "ProcessQueue" {
				issues := checkUnboundedLoops(fn)
				if len(issues) != 1 {
					t.Errorf("ProcessQueue: expected 1 unbounded loop issue, got %d", len(issues))
				}
				return
			}
		}
		t.Error("ProcessQueue function not found")
	})
}

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type FunctionStats struct {
	Name       string
	File       string
	Line       int
	Assertions int
	Issues     []Issue
}

type Issue struct {
	Type       string // "unbounded-loop", "compound-condition", "weak-comment"
	Line       int
	Message    string
	Suggestion string
}

func AnalyzeFile(filename string) ([]FunctionStats, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var stats []FunctionStats

	// First, check comments at file level
	commentIssues := checkComments(file)

	// Walk the AST to find function declarations
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name.Name == "init" {
				return true // Skip init functions
			}

			funcStats := FunctionStats{
				Name:       node.Name.Name,
				File:       filepath.Base(filename),
				Line:       fset.Position(node.Pos()).Line,
				Assertions: countAssertions(node),
				Issues:     []Issue{},
			}

			// Check for unbounded loops
			funcStats.Issues = append(funcStats.Issues, checkUnboundedLoops(node)...)

			// Check for compound conditions
			funcStats.Issues = append(funcStats.Issues, checkCompoundConditions(node)...)

			stats = append(stats, funcStats)
		}
		return true
	})

	// Add file-level comment issues to a special entry
	if len(commentIssues) > 0 {
		// Attach to first function, or create a file-level entry
		// For simplicity, create a pseudo-function entry
		stats = append(stats, FunctionStats{
			Name:       "(file comments)",
			File:       filepath.Base(filename),
			Line:       1,
			Assertions: 0,
			Issues:     commentIssues,
		})
	}

	return stats, nil
}

func countAssertions(funcDecl *ast.FuncDecl) int {
	count := 0

	if funcDecl.Body == nil {
		return 0
	}

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if it's an assertion call (assert.*)
		switch fun := callExpr.Fun.(type) {
		case *ast.SelectorExpr:
			if ident, ok := fun.X.(*ast.Ident); ok {
				if ident.Name == "assert" {
					count++
				}
			}
		}

		return true
	})

	return count
}

func checkUnboundedLoops(funcDecl *ast.FuncDecl) []Issue {
	var issues []Issue

	if funcDecl.Body == nil {
		return issues
	}

	// Create a temporary file set for position
	tempFset := token.NewFileSet()
	// We need to get line numbers, so we'll use the function's file
	// This is a bit hacky but works for our purposes

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		forStmt, ok := n.(*ast.ForStmt)
		if !ok {
			return true
		}

		isUnbounded := false
		loopCode := ""

		if forStmt.Init == nil && forStmt.Post == nil && forStmt.Cond == nil {
			// for {}
			isUnbounded = true
			loopCode = "for {}"
		} else if forStmt.Init == nil && forStmt.Post == nil {
			// for cond {} or for range
			switch forStmt.Cond.(type) {
			case *ast.CallExpr:
				isUnbounded = true
				loopCode = "for <call>()"
			case *ast.Ident:
				isUnbounded = true
				loopCode = "for <condition>"
			}
		}

		if isUnbounded {
			// Try to get better position info
			pos := tempFset.Position(forStmt.Pos())
			if pos.Line == 0 {
				// Estimate line from function body
				pos.Line = 1
			}

			issues = append(issues, Issue{
				Type:       "unbounded-loop",
				Line:       pos.Line,
				Message:    fmt.Sprintf("Unbounded loop detected: %s", loopCode),
				Suggestion: "Add maximum iteration limit or bounded condition",
			})
		}

		return true
	})

	return issues
}

func checkCompoundConditions(funcDecl *ast.FuncDecl) []Issue {
	var issues []Issue

	if funcDecl.Body == nil {
		return issues
	}

	tempFset := token.NewFileSet()

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		// Check if condition contains && or ||
		hasCompound := hasCompoundCondition(ifStmt.Cond)

		if hasCompound {
			pos := tempFset.Position(ifStmt.Pos())
			if pos.Line == 0 {
				pos.Line = 1
			}

			issues = append(issues, Issue{
				Type:       "compound-condition",
				Line:       pos.Line,
				Message:    "Compound condition detected (&& or ||)",
				Suggestion: "Split into multiple simple conditions for clarity",
			})
		}

		return true
	})

	return issues
}

func hasCompoundCondition(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		if e.Op.String() == "&&" || e.Op.String() == "||" {
			return true
		}
		return hasCompoundCondition(e.X) || hasCompoundCondition(e.Y)
	case *ast.ParenExpr:
		return hasCompoundCondition(e.X)
	}
	return false
}

func checkComments(file *ast.File) []Issue {
	var issues []Issue

	weakPrefixes := []string{
		"This ",
		"The ",
		"Construct",
		"Ensure",
		"Create",
		"Initialize",
		"Check",
		"Verify",
	}

	if file.Doc == nil {
		return issues
	}

	for _, comment := range file.Doc.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		text = strings.TrimSpace(strings.TrimPrefix(text, "/*"))
		text = strings.TrimSpace(strings.TrimSuffix(text, "*/"))

		for _, prefix := range weakPrefixes {
			if strings.HasPrefix(text, prefix) || strings.HasPrefix(strings.ToUpper(text), strings.ToUpper(prefix)) {
				issues = append(issues, Issue{
					Type:       "weak-comment",
					Line:       1,
					Message:    fmt.Sprintf("Weak comment detected: starts with '%s'", prefix),
					Suggestion: "Use descriptive comments that explain 'why', not 'what'",
				})
				return issues
			}
		}
	}

	return issues
}

// Helper function to check if a directory should be ignored
func shouldIgnoreDir(name string) bool {
	ignoreDirs := []string{
		"vendor",
		"node_modules",
		".git",
		".hg",
		".svn",
		"_vendor",
		"Godeps",
		"third_party",
		"testdata",
	}

	for _, dir := range ignoreDirs {
		if name == dir {
			return true
		}
	}
	return false
}

// FindGoFiles finds all .go files in the given paths
func FindGoFiles(paths []string) ([]string, error) {
	var files []string

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			// Handle ./... pattern
			if strings.HasSuffix(path, "/...") {
				root := strings.TrimSuffix(path, "/...")
				err := filepath.Walk(root, func(filePath string, info os.FileInfo, err error) error {
					if err != nil {
						return nil // Skip errors
					}
					if info.IsDir() && shouldIgnoreDir(info.Name()) {
						return filepath.SkipDir
					}
					if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
						files = append(files, filePath)
					}
					return nil
				})
				if err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}

		if info.IsDir() {
			err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() && shouldIgnoreDir(info.Name()) {
					return filepath.SkipDir
				}
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
					files = append(files, filePath)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
	}

	return files, nil
}

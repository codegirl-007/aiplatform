package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]

	// Parse arguments
	var paths []string
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			PrintUsage()
			os.Exit(0)
		}
		if !strings.HasPrefix(arg, "-") {
			paths = append(paths, arg)
		}
	}

	// Default to current directory if no paths provided
	if len(paths) == 0 {
		paths = []string{"./..."}
	}

	// Find all Go files
	files, err := FindGoFiles(paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding files: %v\n", err)
		os.Exit(0) // Never fail, as per requirements
	}

	if len(files) == 0 {
		fmt.Println("No Go files found to analyze.")
		os.Exit(0)
	}

	// Analyze all files
	var allStats []FunctionStats
	for _, file := range files {
		stats, err := AnalyzeFile(file)
		if err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Warning: could not analyze %s: %v\n", file, err)
			continue
		}
		allStats = append(allStats, stats...)
	}

	// Print detailed report
	PrintReport(allStats)

	// Print summary
	PrintSummary(allStats, len(files))
}

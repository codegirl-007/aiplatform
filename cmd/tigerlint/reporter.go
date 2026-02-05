package main

import (
	"fmt"
	"os"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

func green(s string) string {
	return colorGreen + s + colorReset
}

func yellow(s string) string {
	return colorYellow + s + colorReset
}

func cyan(s string) string {
	return colorCyan + s + colorReset
}

func red(s string) string {
	return colorRed + s + colorReset
}

func bold(s string) string {
	return colorBold + s + colorReset
}

func PrintReport(stats []FunctionStats) {
	fmt.Println("🐯 Tiger Beetle Style Guide - Code Analysis")
	fmt.Println()

	// Filter to only show functions needing attention
	var functionsNeedingAttention []FunctionStats
	for _, stat := range stats {
		if stat.Assertions < 2 || len(stat.Issues) > 0 {
			functionsNeedingAttention = append(functionsNeedingAttention, stat)
		}
	}

	if len(functionsNeedingAttention) == 0 {
		fmt.Println(green("✓ All functions meet the style guide requirements!"))
		fmt.Println()
		return
	}

	// Group by file for better organization
	byFile := make(map[string][]FunctionStats)
	for _, stat := range functionsNeedingAttention {
		byFile[stat.File] = append(byFile[stat.File], stat)
	}

	for file, fileStats := range byFile {
		fmt.Printf("%s\n", bold(cyan(file)))
		for _, stat := range fileStats {
			if stat.Assertions < 2 {
				fmt.Printf("  %s %s:%d %s %d assertions\n",
					yellow("⚠️"),
					stat.File,
					stat.Line,
					fmt.Sprintf("%-30s", stat.Name+"()"),
					stat.Assertions,
				)
			}

			for _, issue := range stat.Issues {
				fmt.Printf("  %s %s:%d %s\n",
					red("⚠️"),
					stat.File,
					stat.Line,
					stat.Name+"()",
				)
				fmt.Printf("              %s: %s\n",
					yellow(issue.Type),
					issue.Message,
				)
				if issue.Suggestion != "" {
					fmt.Printf("              %s\n", green("→ "+issue.Suggestion))
				}
			}
		}
		fmt.Println()
	}
}

func PrintSummary(stats []FunctionStats, totalFiles int) {
	fmt.Println(bold("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("%s\n", bold(cyan("📊 Assertion Density")))
	fmt.Println(bold("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

	var totalAssertions int
	var functionCount int

	for _, stat := range stats {
		if stat.Name != "(file comments)" {
			totalAssertions += stat.Assertions
			functionCount++
		}
	}

	average := 0.0
	if functionCount > 0 {
		average = float64(totalAssertions) / float64(functionCount)
	}

	fmt.Printf("Target:  %s assertions per function (average)\n", bold("2.0"))
	fmt.Printf("Current: %s assertions per function (average)\n", bold(fmt.Sprintf("%.1f", average)))
	fmt.Println()

	// Count issues by type
	unboundedLoops := 0
	compoundConditions := 0
	weakComments := 0

	for _, stat := range stats {
		for _, issue := range stat.Issues {
			switch issue.Type {
			case "unbounded-loop":
				unboundedLoops++
			case "compound-condition":
				compoundConditions++
			case "weak-comment":
				weakComments++
			}
		}
	}

	fmt.Println(bold("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("%s\n", bold(yellow("⚠️  Safety Issues")))
	fmt.Println(bold("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

	if unboundedLoops == 0 && compoundConditions == 0 && weakComments == 0 {
		fmt.Printf("  %s No safety issues found\n", green("✓"))
	} else {
		if unboundedLoops > 0 {
			fmt.Printf("  %s Unbounded loops: %d\n", yellow("⚠️"), unboundedLoops)
		}
		if compoundConditions > 0 {
			fmt.Printf("  %s Compound conditions: %d\n", yellow("⚠️"), compoundConditions)
		}
		if weakComments > 0 {
			fmt.Printf("  %s Weak comments: %d\n", yellow("⚠️"), weakComments)
		}
	}

	fmt.Println()
	fmt.Printf("Files analyzed: %d | Functions: %d | Total assertions: %d\n",
		totalFiles, functionCount, totalAssertions)
}

func PrintUsage() {
	fmt.Println("🐯 Tiger Beetle Linter")
	fmt.Println()
	fmt.Println("Usage: tigerlint [options] <path> [path...]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  tigerlint ./...              # Analyze all Go files recursively")
	fmt.Println("  tigerlint pkg/storage        # Analyze specific directory")
	fmt.Println("  tigerlint cmd/main.go        # Analyze specific file")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help    Show this help message")
	fmt.Println()
}

func isTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

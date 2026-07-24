package main

import "fmt"

// Severity indicates how serious a finding is.
type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityInfo   Severity = "info"
)

// Finding is a single diagnostic result from the analyzer.
type Finding struct {
	File       string   `json:"file"`
	Line       int      `json:"line"`
	Col        int      `json:"col"`
	FuncName   string   `json:"func"`
	Severity   Severity `json:"severity"`
	Rule       string   `json:"rule"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion"`
}

// String formats the finding in the default text output format.
func (f Finding) String() string {
	return fmt.Sprintf("%s:%d:%d: %s: %s: %s", f.File, f.Line, f.Col, f.Severity, f.Rule, f.Message)
}

// countBySeverity returns counts for each severity level present in findings.
func countBySeverity(findings []Finding) map[Severity]int {
	counts := map[Severity]int{
		SeverityHigh:   0,
		SeverityMedium: 0,
		SeverityInfo:   0,
	}
	for _, f := range findings {
		counts[f.Severity]++
	}
	return counts
}

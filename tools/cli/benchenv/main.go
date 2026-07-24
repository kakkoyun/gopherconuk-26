package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
)

// Status represents the health status of a diagnostic check.
type Status string

const (
	statusOK          Status = "ok"
	statusWarn        Status = "warn"
	statusUnavailable Status = "unavailable"
)

// Check is the result of a single diagnostic probe.
type Check struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
	Detail string `json:"detail"`
	Remedy string `json:"remedy,omitempty"`
}

// Summary counts check results by status.
type Summary struct {
	OK          int `json:"ok"`
	Warn        int `json:"warn"`
	Unavailable int `json:"unavailable"`
}

// Report is the top-level structure for JSON output.
type Report struct {
	OS      string  `json:"os"`
	Arch    string  `json:"arch"`
	NumCPU  int     `json:"numcpu"`
	Checks  []Check `json:"checks"`
	Summary Summary `json:"summary"`
}

// summarize counts checks by status.
func summarize(checks []Check) Summary {
	var s Summary
	for _, c := range checks {
		switch c.Status {
		case statusOK:
			s.OK++
		case statusWarn:
			s.Warn++
		case statusUnavailable:
			s.Unavailable++
		}
	}
	return s
}

func main() {
	jsonOut := flag.Bool("json", false, "output results as JSON")
	flag.Parse()

	checks := collectChecks()
	sum := summarize(checks)
	report := Report{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		NumCPU:  runtime.NumCPU(),
		Checks:  checks,
		Summary: sum,
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "benchenv: json encode: %v\n", err)
			os.Exit(2)
		}
		return
	}

	printText(report)
}

func printText(r Report) {
	fmt.Printf("benchenv: benchmarking environment diagnosis (%s/%s, %d CPUs)\n\n", r.OS, r.Arch, r.NumCPU)
	for _, c := range r.Checks {
		label := fmt.Sprintf("[%s]", c.Status)
		var suffix string
		switch c.Status {
		case statusWarn:
			if c.Remedy != "" {
				suffix = " — " + c.Remedy
			} else if c.Detail != "" {
				suffix = " — " + c.Detail
			}
		case statusUnavailable:
			if c.Detail != "" {
				suffix = " — " + c.Detail
			}
		case statusOK:
			if c.Detail != "" {
				suffix = " — " + c.Detail
			}
		}
		fmt.Printf("  %-15s %s%s\n", label, c.Name, suffix)
	}
	fmt.Printf("\nSummary: %d ok, %d warn, %d unavailable. Fix warnings before trusting benchmark numbers.\n",
		r.Summary.OK, r.Summary.Warn, r.Summary.Unavailable)
}

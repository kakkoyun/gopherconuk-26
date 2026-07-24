package main

import (
	"go/token"
	"path/filepath"
	"testing"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func mustAnalyze(t *testing.T, name string) []Finding {
	t.Helper()
	fset := token.NewFileSet()
	findings, _, err := analyzeFile(fset, filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("analyzeFile(%s): %v", name, err)
	}
	return findings
}

func findingsByRule(findings []Finding) map[string][]Finding {
	m := map[string][]Finding{}
	for _, f := range findings {
		m[f.Rule] = append(m[f.Rule], f)
	}
	return m
}

// ── bad_test.go: every rule must fire at least once ───────────────────────────

func TestBadFile_DiscardedResult(t *testing.T) {
	findings := mustAnalyze(t, "bad_test.go")
	byRule := findingsByRule(findings)

	if len(byRule["discarded-result"]) == 0 {
		t.Error("expected at least one discarded-result finding; got none")
	}
	for _, f := range byRule["discarded-result"] {
		if f.Severity != SeverityHigh {
			t.Errorf("discarded-result severity = %s; want high", f.Severity)
		}
	}
}

func TestBadFile_MissingSink(t *testing.T) {
	findings := mustAnalyze(t, "bad_test.go")
	byRule := findingsByRule(findings)

	if len(byRule["missing-sink"]) == 0 {
		t.Error("expected at least one missing-sink finding; got none")
	}
	for _, f := range byRule["missing-sink"] {
		if f.Severity != SeverityMedium {
			t.Errorf("missing-sink severity = %s; want medium", f.Severity)
		}
	}
}

func TestBadFile_StopTimerWithoutStartTimer(t *testing.T) {
	findings := mustAnalyze(t, "bad_test.go")
	byRule := findingsByRule(findings)

	if len(byRule["stoptimer-without-starttimer"]) == 0 {
		t.Error("expected at least one stoptimer-without-starttimer finding; got none")
	}
	for _, f := range byRule["stoptimer-without-starttimer"] {
		if f.Severity != SeverityHigh {
			t.Errorf("stoptimer-without-starttimer severity = %s; want high", f.Severity)
		}
	}
}

func TestBadFile_SuggestBloop(t *testing.T) {
	findings := mustAnalyze(t, "bad_test.go")
	byRule := findingsByRule(findings)

	if len(byRule["suggest-bloop"]) == 0 {
		t.Error("expected at least one suggest-bloop finding; got none")
	}
	for _, f := range byRule["suggest-bloop"] {
		if f.Severity != SeverityInfo {
			t.Errorf("suggest-bloop severity = %s; want info", f.Severity)
		}
	}
}

// ── good_test.go: no high/medium findings ────────────────────────────────────

func TestGoodFile_NoHighOrMediumFindings(t *testing.T) {
	findings := mustAnalyze(t, "good_test.go")
	for _, f := range findings {
		if f.Severity == SeverityHigh || f.Severity == SeverityMedium {
			t.Errorf("unexpected %s finding in good benchmark file: %s", f.Severity, f)
		}
	}
}

// BenchmarkGood_BLoop uses b.Loop() — must NOT produce a suggest-bloop finding.
func TestGoodFile_BLoopNotFlagged(t *testing.T) {
	findings := mustAnalyze(t, "good_test.go")
	for _, f := range findings {
		if f.Rule == "suggest-bloop" && f.FuncName == "BenchmarkGood_BLoop" {
			t.Errorf("BenchmarkGood_BLoop should not produce suggest-bloop, got: %s", f)
		}
	}
}

// ── targeted per-benchmark checks ────────────────────────────────────────────

// Table-driven check that specific benchmark names produce (or don't produce)
// a specific rule.
func TestSpecificBenchmarks(t *testing.T) {
	tests := []struct {
		file      string
		benchmark string // substring matched against finding message
		rule      string
		wantHit   bool // true = expect at least one finding; false = expect none
	}{
		// bad_test.go — should fire
		{"bad_test.go", "BenchmarkBad_DiscardedResult", "discarded-result", true},
		{"bad_test.go", "BenchmarkBad_DiscardedResult", "suggest-bloop", true},
		{"bad_test.go", "BenchmarkBad_MissingSink", "missing-sink", true},
		{"bad_test.go", "BenchmarkBad_StopTimerNoStart", "stoptimer-without-starttimer", true},
		{"bad_test.go", "BenchmarkBad_TimerOrder", "stoptimer-without-starttimer", true},

		// good_test.go — correct b.Loop() form must not trigger suggest-bloop
		{"good_test.go", "BenchmarkGood_BLoop", "suggest-bloop", false},
		// correct sink pattern must not trigger discarded-result or missing-sink
		{"good_test.go", "BenchmarkGood_BN_WithSink", "discarded-result", false},
		{"good_test.go", "BenchmarkGood_BN_WithSink", "missing-sink", false},
		// correct timer order must not trigger stoptimer-without-starttimer
		{"good_test.go", "BenchmarkGood_TimerCorrect", "stoptimer-without-starttimer", false},
	}

	// Cache per-file results to avoid re-parsing.
	cache := map[string]map[string][]Finding{}
	parse := func(t *testing.T, file string) map[string][]Finding {
		t.Helper()
		if m, ok := cache[file]; ok {
			return m
		}
		findings := mustAnalyze(t, file)
		m := findingsByRule(findings)
		cache[file] = m
		return m
	}

	for _, tt := range tests {
		t.Run(tt.benchmark+"/"+tt.rule, func(t *testing.T) {
			byRule := parse(t, tt.file)
			// Filter findings for this specific benchmark by FuncName.
			var hits []Finding
			for _, f := range byRule[tt.rule] {
				if f.FuncName == tt.benchmark {
					hits = append(hits, f)
				}
			}
			if tt.wantHit && len(hits) == 0 {
				t.Errorf("%s in %s: expected a %q finding, got none\nAll %q findings:\n%v",
					tt.benchmark, tt.file, tt.rule, tt.rule, byRule[tt.rule])
			}
			if !tt.wantHit && len(hits) > 0 {
				t.Errorf("%s in %s: expected no %q finding, got %d:\n%v",
					tt.benchmark, tt.file, tt.rule, len(hits), hits)
			}
		})
	}
}

// ── line/col sanity ───────────────────────────────────────────────────────────

func TestFindingsHaveValidPosition(t *testing.T) {
	for _, file := range []string{"bad_test.go", "good_test.go"} {
		findings := mustAnalyze(t, file)
		for _, f := range findings {
			if f.Line <= 0 {
				t.Errorf("%s: finding %q has invalid line %d", file, f.Rule, f.Line)
			}
			if f.Col <= 0 {
				t.Errorf("%s: finding %q has invalid col %d", file, f.Rule, f.Col)
			}
			if f.Message == "" {
				t.Errorf("%s: finding %q has empty message", file, f.Rule)
			}
			if f.Suggestion == "" {
				t.Errorf("%s: finding %q has empty suggestion", file, f.Rule)
			}
		}
	}
}

// ── collectFiles ──────────────────────────────────────────────────────────────

func TestCollectFiles_SingleFile(t *testing.T) {
	files, err := collectFiles(filepath.Join("testdata", "bad_test.go"), false)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
}

func TestCollectFiles_Directory(t *testing.T) {
	files, err := collectFiles("testdata", false)
	if err != nil {
		t.Fatalf("collectFiles: %v", err)
	}
	if len(files) < 2 {
		t.Fatalf("expected at least 2 files in testdata, got %d", len(files))
	}
	for _, f := range files {
		if !endsWith(f, "_test.go") {
			t.Errorf("collectFiles returned non-test file: %s", f)
		}
	}
}

// ── utilities ─────────────────────────────────────────────────────────────────

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// benchgate runs Go benchmarks N times, computes the coefficient of variation
// (CV) per benchmark, and emits a pass/fail verdict against a CV threshold.
// Optionally compares against a baseline file using benchstat.
//
// Exit codes: 0 = PASS, 1 = FAIL, 2 = error.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// benchmarkRE matches go test -bench output lines of the form:
//
//	BenchmarkName-8   1000000   123.4 ns/op   ...
//
// The GOMAXPROCS suffix (-N) is captured separately so it can be stripped.
var benchmarkRE = regexp.MustCompile(`^(Benchmark[^\s-]+)(?:-\d+)?\s+\d+\s+([\d.]+)\s+ns/op`)

// result holds per-benchmark statistics.
type result struct {
	Name   string  `json:"name"`
	Mean   float64 `json:"mean"`
	Stddev float64 `json:"stddev"`
	CV     float64 `json:"cv"`
	Pass   bool    `json:"pass"`
	Note   string  `json:"note,omitempty"`
}

// report is the top-level JSON output structure.
type report struct {
	Verdict    string   `json:"verdict"`
	Threshold  float64  `json:"threshold"`
	Benchmarks []result `json:"benchmarks"`
}

// parseBenchOutput parses raw go test -bench output and groups ns/op samples
// by benchmark name (GOMAXPROCS suffix stripped).
func parseBenchOutput(output string) map[string][]float64 {
	samples := make(map[string][]float64)
	for _, line := range strings.Split(output, "\n") {
		m := benchmarkRE.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		name := m[1]
		val, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			continue
		}
		samples[name] = append(samples[name], val)
	}
	return samples
}

// computeCV returns mean, sample standard deviation, CV%, and a note string
// for a slice of ns/op samples.
//
// Sample stddev formula: sqrt(sum((x-mean)^2) / (n-1)).
// If n < 2, stddev and cv are 0 and note is set.
func computeCV(samples []float64) (mean, stddev, cv float64, note string) {
	if len(samples) == 0 {
		return 0, 0, 0, "no samples"
	}
	var sum float64
	for _, v := range samples {
		sum += v
	}
	mean = sum / float64(len(samples))
	if len(samples) < 2 {
		return mean, 0, 0, "insufficient samples"
	}
	var variance float64
	for _, v := range samples {
		d := v - mean
		variance += d * d
	}
	stddev = math.Sqrt(variance / float64(len(samples)-1))
	if mean > 0 {
		cv = 100 * stddev / mean
	}
	return mean, stddev, cv, ""
}

// resolvePackageTarget examines the -pkg value and returns the working
// directory and package pattern to pass to go test.
//
// When -pkg is a local filesystem path (starts with "." or "/") that resolves
// to a directory containing its own go.mod, go test must be invoked from that
// directory (cross-module relative paths are not supported). In that case the
// returned dir is the resolved directory and pattern is "./...". Otherwise dir
// is empty (use the process cwd) and pattern is pkg unchanged.
func resolvePackageTarget(pkg string) (dir, pattern string) {
	// Module import paths never start with "." or "/".
	if !strings.HasPrefix(pkg, ".") && !strings.HasPrefix(pkg, "/") {
		return "", pkg
	}

	// Strip trailing /... or trailing slash to get the base directory.
	base := strings.TrimSuffix(strings.TrimSuffix(pkg, "..."), "/")
	if base == "" {
		base = "."
	}
	resolved, err := filepath.Abs(base)
	if err != nil {
		return "", pkg // fall back: let go test surface the error
	}

	// Walk up from resolved to cwd to see if it's in the same module.
	cwd, err := os.Getwd()
	if err != nil {
		return "", pkg
	}
	modInCwd := filepath.Join(cwd, "go.mod")
	modInTarget := filepath.Join(resolved, "go.mod")

	_, cwdModErr := os.Stat(modInCwd)
	_, targetModErr := os.Stat(modInTarget)

	// If the target has its own go.mod and it is not the same file as the
	// cwd's go.mod, run go test inside the target directory.
	if targetModErr == nil && (cwdModErr != nil || resolved != cwd) {
		return resolved, "./..."
	}

	// Same module or no separate go.mod found — keep pkg as-is.
	return "", pkg
}

// verdictLabel returns "PASS" or "FAIL".
func verdictLabel(pass bool) string {
	if pass {
		return "PASS"
	}
	return "FAIL"
}

func run() int {
	pkg := flag.String("pkg", "./...", "package pattern to benchmark")
	bench := flag.String("bench", ".", "benchmark regexp")
	count := flag.Int("count", 10, "number of benchmark runs")
	benchtime := flag.String("benchtime", "1s", "go test -benchtime value")
	cvThreshold := flag.Float64("cv-threshold", 5.0, "max acceptable CV percent")
	jsonOut := flag.Bool("json", false, "emit JSON output")
	baseline := flag.String("baseline", "", "path to saved go test -bench output for benchstat comparison")
	save := flag.String("save", "", "path to write raw benchmark output (future baseline)")
	flag.Parse()

	workDir, pkgPattern := resolvePackageTarget(*pkg)

	goArgs := []string{
		"test",
		"-run", "^$",
		fmt.Sprintf("-bench=%s", *bench),
		"-benchmem",
		fmt.Sprintf("-count=%d", *count),
		fmt.Sprintf("-benchtime=%s", *benchtime),
		pkgPattern,
	}

	cmdStr := fmt.Sprintf("go test -bench=%s -count=%d -benchtime=%s %s", *bench, *count, *benchtime, pkgPattern)
	if !*jsonOut {
		fmt.Printf("benchgate: %s\n\n", cmdStr)
	}

	goCmd := exec.Command("go", goArgs...) //nolint:gosec
	if workDir != "" {
		goCmd.Dir = workDir
	}
	out, err := goCmd.CombinedOutput()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if !ok {
			fmt.Fprintf(os.Stderr, "benchgate: go test failed: %v\n%s\n", err, out)
			return 2
		}
		// Exit code 2 from go test means compile or vet error.
		if exitErr.ExitCode() == 2 {
			fmt.Fprintf(os.Stderr, "benchgate: go test build/vet error:\n%s\n", out)
			return 2
		}
		// Exit code 1 means a test (not benchmark) failed; benchmark output
		// may still be present, so continue processing.
	}

	rawOutput := string(out)

	if *save != "" {
		if werr := os.WriteFile(*save, out, 0o644); werr != nil {
			fmt.Fprintf(os.Stderr, "benchgate: save %q: %v\n", *save, werr)
			return 2
		}
	}

	samples := parseBenchOutput(rawOutput)
	if len(samples) == 0 {
		fmt.Fprintf(os.Stderr, "benchgate: no benchmark results found\n%s\n", rawOutput)
		return 2
	}

	// Sort benchmark names for deterministic output.
	names := make([]string, 0, len(samples))
	for n := range samples {
		names = append(names, n)
	}
	sort.Strings(names)

	results := make([]result, 0, len(names))
	failing := 0

	for _, name := range names {
		mean, stddev, cv, note := computeCV(samples[name])
		pass := cv <= *cvThreshold
		if !pass {
			failing++
		}
		results = append(results, result{
			Name:   name,
			Mean:   mean,
			Stddev: stddev,
			CV:     cv,
			Pass:   pass,
			Note:   note,
		})
	}

	overallPass := failing == 0

	if *jsonOut {
		rep := report{
			Verdict:    verdictLabel(overallPass),
			Threshold:  *cvThreshold,
			Benchmarks: results,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if encErr := enc.Encode(rep); encErr != nil {
			fmt.Fprintf(os.Stderr, "benchgate: json encode: %v\n", encErr)
			return 2
		}
	} else {
		for _, r := range results {
			mark := "✓"
			suffix := ""
			if !r.Pass {
				mark = "✗"
				suffix = fmt.Sprintf(" (exceeds %.1f%% threshold)", *cvThreshold)
			}
			if r.Note != "" {
				suffix = fmt.Sprintf(" [%s]", r.Note)
			}
			fmt.Printf("  %-44s  mean=%8.1f ns/op  cv=%5.1f%%  %s%s\n",
				r.Name, r.Mean, r.CV, mark, suffix)
		}
		fmt.Println()
		if overallPass {
			fmt.Printf("VERDICT: PASS — all %d benchmarks within CV threshold %.1f%%\n",
				len(results), *cvThreshold)
		} else {
			fmt.Printf("VERDICT: FAIL — %d/%d benchmarks exceed CV threshold %.1f%%\n",
				failing, len(results), *cvThreshold)
		}
	}

	if *baseline != "" {
		tmp, tmpErr := os.CreateTemp("", "benchgate-new-*.txt")
		if tmpErr != nil {
			fmt.Fprintf(os.Stderr, "benchgate: create temp file: %v\n", tmpErr)
			return 2
		}
		defer os.Remove(tmp.Name())

		if _, wErr := tmp.WriteString(rawOutput); wErr != nil {
			tmp.Close()
			fmt.Fprintf(os.Stderr, "benchgate: write temp file: %v\n", wErr)
			return 2
		}
		tmp.Close()

		fmt.Println("\n--- benchstat comparison ---")
		bsOut, bsErr := exec.Command("benchstat", *baseline, tmp.Name()).CombinedOutput() //nolint:gosec
		fmt.Print(string(bsOut))
		if bsErr != nil {
			fmt.Fprintf(os.Stderr, "benchgate: benchstat: %v\n", bsErr)
		}
	}

	if !overallPass {
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}

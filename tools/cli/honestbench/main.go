// honestbench is a static analyzer for Go benchmark correctness.
//
// It parses *_test.go files with go/ast and flags common benchmark-correctness
// mistakes: dead-code elimination traps, timer misuse, missing sink patterns,
// and outdated b.N loop idioms.
//
// Usage:
//
//	honestbench [flags] <path>
//
// Flags:
//
//	-r      Recurse into subdirectories (default: non-recursive)
//	-json   Emit findings as a JSON array
//	-q      Quiet: omit the summary line
//
// Exit codes:
//
//	0  no findings
//	1  one or more findings
//	2  error (bad path, parse failure)
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	os.Exit(run())
}

func run() int {
	recursive := flag.Bool("r", false, "analyze directories recursively")
	jsonOut := flag.Bool("json", false, "output findings as JSON array")
	quiet := flag.Bool("q", false, "quiet: only print findings, no summary line")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: honestbench [flags] <path>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		return 2
	}

	root := flag.Arg(0)
	files, err := collectFiles(root, *recursive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "honestbench: %v\n", err)
		return 2
	}
	if len(files) == 0 {
		if !*quiet {
			fmt.Fprintln(os.Stderr, "honestbench: no *_test.go files found")
		}
		return 0
	}

	fset := token.NewFileSet()
	var allFindings []Finding
	totalFuncs := 0

	for _, f := range files {
		findings, count, ferr := analyzeFile(fset, f)
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "honestbench: parse %s: %v\n", f, ferr)
			return 2
		}
		allFindings = append(allFindings, findings...)
		totalFuncs += count
	}

	if *jsonOut {
		if allFindings == nil {
			allFindings = []Finding{} // encode as [] not null
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if encErr := enc.Encode(allFindings); encErr != nil {
			fmt.Fprintf(os.Stderr, "honestbench: json encode: %v\n", encErr)
			return 2
		}
		return exitCode(allFindings)
	}

	for _, f := range allFindings {
		fmt.Println(f)
	}

	if !*quiet {
		counts := countBySeverity(allFindings)
		fmt.Printf("\n%d findings (%d high, %d medium, %d info) across %d functions\n",
			len(allFindings),
			counts[SeverityHigh],
			counts[SeverityMedium],
			counts[SeverityInfo],
			totalFuncs,
		)
	}

	return exitCode(allFindings)
}

func exitCode(findings []Finding) int {
	if len(findings) > 0 {
		return 1
	}
	return 0
}

// collectFiles returns the list of *_test.go files to analyze. root may be a
// file path (returned as-is) or a directory. When recursive is true, the
// directory tree is walked.
func collectFiles(root string, recursive bool) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", root, err)
	}

	if !info.IsDir() {
		// Accept any Go file when given explicitly.
		return []string{root}, nil
	}

	var files []string

	if recursive {
		walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			// Skip hidden and vendor directories.
			if d.IsDir() && (strings.HasPrefix(d.Name(), ".") || d.Name() == "vendor") {
				return filepath.SkipDir
			}
			if !d.IsDir() && strings.HasSuffix(path, "_test.go") {
				files = append(files, path)
			}
			return nil
		})
		return files, walkErr
	}

	entries, rdErr := os.ReadDir(root)
	if rdErr != nil {
		return nil, fmt.Errorf("readdir %s: %w", root, rdErr)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), "_test.go") {
			files = append(files, filepath.Join(root, e.Name()))
		}
	}
	return files, nil
}

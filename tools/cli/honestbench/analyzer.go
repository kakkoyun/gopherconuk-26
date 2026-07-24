package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// analyzeFile parses a single Go source file and returns findings for every
// Benchmark* function it contains, plus the number of benchmark functions found.
func analyzeFile(fset *token.FileSet, path string) ([]Finding, int, error) {
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, 0, err
	}

	var findings []Finding
	funcCount := 0

	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}
		if !isBenchmarkFunc(fd) {
			continue
		}
		funcCount++
		bName := benchmarkParamName(fd)
		findings = append(findings, analyzeFunc(fset, path, fd, bName)...)
	}

	return findings, funcCount, nil
}

// isBenchmarkFunc returns true when fd is named Benchmark* and its first
// parameter is *testing.B.
func isBenchmarkFunc(fd *ast.FuncDecl) bool {
	if !strings.HasPrefix(fd.Name.Name, "Benchmark") {
		return false
	}
	params := fd.Type.Params
	if params == nil || len(params.List) == 0 {
		return false
	}
	star, ok := params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "testing" && sel.Sel.Name == "B"
}

// benchmarkParamName returns the name of the *testing.B parameter (usually "b").
func benchmarkParamName(fd *ast.FuncDecl) string {
	list := fd.Type.Params.List[0]
	if len(list.Names) > 0 {
		return list.Names[0].Name
	}
	return "b"
}

// analyzeFunc runs all checks on a single Benchmark function.
func analyzeFunc(fset *token.FileSet, path string, fd *ast.FuncDecl, bName string) []Finding {
	mkFinding := func(p token.Pos, sev Severity, rule, msg, suggestion string) Finding {
		pp := fset.Position(p)
		return Finding{
			File:       path,
			Line:       pp.Line,
			Col:        pp.Column,
			FuncName:   fd.Name.Name,
			Severity:   sev,
			Rule:       rule,
			Message:    msg,
			Suggestion: suggestion,
		}
	}

	loopStmt, loopBody, isBLoop := findBenchmarkLoop(fd.Body.List, bName)
	if loopStmt == nil {
		// No recognisable benchmark loop; nothing to check.
		return nil
	}

	var findings []Finding

	// ── 1. suggest-bloop ────────────────────────────────────────────────────
	// Flag for-range b.N and classic for i < b.N; skip for b.Loop() (already correct).
	if !isBLoop {
		findings = append(findings, mkFinding(loopStmt.Pos(), SeverityInfo, "suggest-bloop",
			fmt.Sprintf("%s: benchmark loop uses b.N; consider migrating to b.Loop() (Go 1.24+)", fd.Name.Name),
			"Replace `for range b.N { ... }` or `for i := 0; i < b.N; i++ { ... }` with `for b.Loop() { ... }`. b.Loop() handles timer reset automatically, suppresses inlining that can enable DCE, and avoids the b.N==0 edge case."))
	}

	// ── 2. discarded-result ──────────────────────────────────────────────────
	// A bare call-expression statement inside the loop whose return value is
	// discarded. This is the classic DCE footgun.
	for _, stmt := range loopBody {
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}
		call, ok := exprStmt.X.(*ast.CallExpr)
		if !ok {
			continue
		}
		// Exclude method calls on the benchmark receiver (b.StopTimer etc.) and
		// well-known void stdlib helpers.
		if isBMethodCall(call, bName) || isKnownVoidCall(call) {
			continue
		}
		name := callName(call)
		findings = append(findings, mkFinding(exprStmt.Pos(), SeverityHigh, "discarded-result",
			fmt.Sprintf("call to %s result is discarded; compiler may eliminate this call via DCE", name),
			"Assign the result to a local accumulator and write it to a package-level sink after the loop: `var s T; for b.Loop() { s = "+name+"(...) }; globalSink = s`"))
	}

	// ── 3. stoptimer-without-starttimer ──────────────────────────────────────
	findings = append(findings, checkTimerOrder(loopBody, bName, mkFinding)...)

	// ── 4. missing-sink ──────────────────────────────────────────────────────
	findings = append(findings, checkMissingSink(fd.Body.List, loopBody, loopStmt, mkFinding)...)

	return findings
}

// ── Loop detection ────────────────────────────────────────────────────────────

// findBenchmarkLoop searches the top-level statement list for a for loop that
// constitutes the main benchmark iteration. Returns the loop node, its body
// statements, and whether it is a b.Loop() call (isBLoop=true) as opposed to
// a b.N-based loop.
func findBenchmarkLoop(stmts []ast.Stmt, bName string) (loopStmt ast.Stmt, body []ast.Stmt, isBLoop bool) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.RangeStmt:
			// for range b.N  OR  for i := range b.N
			if isBNSel(s.X, bName) {
				return s, s.Body.List, false
			}
		case *ast.ForStmt:
			// for b.Loop()
			if isBLoopCond(s.Cond, bName) {
				return s, s.Body.List, true
			}
			// for i := 0; i < b.N; i++
			if s.Cond != nil && containsBNRef(s.Cond, bName) {
				return s, s.Body.List, false
			}
		}
	}
	return nil, nil, false
}

// isBNSel reports whether expr is a selector of the form <bName>.N.
func isBNSel(expr ast.Expr, bName string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == bName && sel.Sel.Name == "N"
}

// isBLoopCond reports whether cond is a call to <bName>.Loop().
func isBLoopCond(cond ast.Expr, bName string) bool {
	if cond == nil {
		return false
	}
	call, ok := cond.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == bName && sel.Sel.Name == "Loop"
}

// containsBNRef reports whether expr contains a reference to <bName>.N.
func containsBNRef(expr ast.Expr, bName string) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found || n == nil {
			return false
		}
		if e, ok := n.(ast.Expr); ok && isBNSel(e, bName) {
			found = true
			return false
		}
		return true
	})
	return found
}

// ── Call helpers ──────────────────────────────────────────────────────────────

// isBMethodCall reports whether call is a method call on the benchmark param.
func isBMethodCall(call *ast.CallExpr, bName string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == bName
}

// isKnownVoidCall returns true for package-qualified calls that are commonly
// void and therefore safe to ignore (e.g. fmt.Println, log.Printf).
func isKnownVoidCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	switch pkg.Name {
	case "fmt":
		switch sel.Sel.Name {
		case "Print", "Println", "Printf", "Fprint", "Fprintln", "Fprintf", "Sscanf", "Fscan":
			return true
		}
	case "log":
		switch sel.Sel.Name {
		case "Print", "Println", "Printf",
			"Fatal", "Fatalf", "Fatalln",
			"Panic", "Panicf", "Panicln":
			return true
		}
	}
	return false
}

// callName returns a human-readable name for a call expression.
func callName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return fn.Name + "()"
	case *ast.SelectorExpr:
		if id, ok := fn.X.(*ast.Ident); ok {
			return id.Name + "." + fn.Sel.Name + "()"
		}
		return fn.Sel.Name + "()"
	}
	return "..."
}

// ── Timer-order check ─────────────────────────────────────────────────────────

// checkTimerOrder detects two patterns:
//  1. b.StopTimer() in the loop with no b.StartTimer() → timer never restarts.
//  2. b.StartTimer() is the last statement in the loop after a b.StopTimer() and
//     intervening non-timer work → work runs while the timer is stopped.
func checkTimerOrder(loopBody []ast.Stmt, bName string, mkFinding func(token.Pos, Severity, string, string, string) Finding) []Finding {
	stopIdx, startIdx := -1, -1
	var stopPos token.Pos

	for i, stmt := range loopBody {
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			continue
		}
		call, ok := exprStmt.X.(*ast.CallExpr)
		if !ok || !isBMethodCall(call, bName) {
			continue
		}
		sel := call.Fun.(*ast.SelectorExpr)
		switch sel.Sel.Name {
		case "StopTimer":
			if stopIdx == -1 { // first occurrence
				stopIdx = i
				stopPos = stmt.Pos()
			}
		case "StartTimer":
			startIdx = i
		}
	}

	if stopIdx == -1 {
		return nil // no StopTimer, nothing to flag
	}

	if startIdx == -1 {
		return []Finding{mkFinding(stopPos, SeverityHigh, "stoptimer-without-starttimer",
			"b.StopTimer() called inside the loop with no matching b.StartTimer(); the benchmark timer will never restart",
			"Add b.StartTimer() before the code under measurement: StopTimer → setup → StartTimer → work")}
	}

	// Both present. Flag when StartTimer is the final statement (work ran while stopped).
	if startIdx == len(loopBody)-1 && stopIdx < startIdx {
		// Confirm there is real work (non-timer statements) between stop and start.
		hasWork := false
		for i := stopIdx + 1; i < startIdx; i++ {
			exprStmt, ok := loopBody[i].(*ast.ExprStmt)
			if !ok {
				hasWork = true
				break
			}
			call, ok := exprStmt.X.(*ast.CallExpr)
			if !ok || !isBMethodCall(call, bName) {
				hasWork = true
				break
			}
		}
		if hasWork {
			return []Finding{mkFinding(stopPos, SeverityHigh, "stoptimer-without-starttimer",
				"b.StartTimer() called after the work under test; measured code runs while the timer is stopped",
				"Move b.StartTimer() before the code under measurement: StopTimer → setup → StartTimer → work")}
		}
	}

	return nil
}

// ── Missing-sink check ────────────────────────────────────────────────────────

// checkMissingSink flags the pattern where a local variable accumulates results
// from inside the benchmark loop but is then discarded with `_ = varName` instead
// of being written to a package-level sink.
func checkMissingSink(bodyStmts, loopBody []ast.Stmt, loopStmt ast.Stmt, mkFinding func(token.Pos, Severity, string, string, string) Finding) []Finding {
	// Collect names of locals assigned inside the loop body.
	loopLocals := map[string]bool{}
	for _, stmt := range loopBody {
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok {
			continue
		}
		for _, lhs := range assign.Lhs {
			if id, ok := lhs.(*ast.Ident); ok && id.Name != "_" {
				loopLocals[id.Name] = true
			}
		}
	}
	if len(loopLocals) == 0 {
		return nil
	}

	// Scan statements that come after the loop for `_ = localVar`.
	loopEnd := loopStmt.End()
	var findings []Finding
	for _, stmt := range bodyStmts {
		if stmt.Pos() < loopEnd {
			continue
		}
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			continue
		}
		lhs, ok := assign.Lhs[0].(*ast.Ident)
		if !ok || lhs.Name != "_" {
			continue
		}
		rhs, ok := assign.Rhs[0].(*ast.Ident)
		if !ok || !loopLocals[rhs.Name] {
			continue
		}
		findings = append(findings, mkFinding(stmt.Pos(), SeverityMedium, "missing-sink",
			fmt.Sprintf("loop result accumulated in %q is discarded with `_ = %s`; compiler may eliminate the computation", rhs.Name, rhs.Name),
			"Write the accumulator to a package-level variable to defeat DCE: `var globalSink T; ... ; globalSink = "+rhs.Name+"`"))
	}
	return findings
}

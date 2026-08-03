// timer_bench_test.go demonstrates correct and incorrect usage of
// b.ResetTimer, b.StopTimer, and b.StartTimer.
//
// These methods are frequently misused, leading to either contaminated
// measurements (setup cost included) or inflated ns/op (timer overhead
// included).
//
// Run:
//
//	go test -bench=BenchmarkProcess -benchmem -count=10 . 2>&1 | tee /tmp/timer.txt
//	benchstat /tmp/timer.txt
package demo

import (
	"strings"
	"testing"
	"unicode"
)

// processString is the function under test: normalises a string to
// lowercase letters only. Allocates — visible in -benchmem output.
func processString(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, s)
}

// buildFixture constructs a per-iteration input string. Simulates work that
// must be excluded from measurement — e.g. decoding a record, building a
// request, preparing a test case.
func buildFixture(n int) string {
	b := strings.Builder{}
	b.Grow(n)
	for i := range n {
		b.WriteByte(byte('A' + i%26))
	}
	return b.String()
}

const fixtureSize = 128

// sinkStr is the package-level sink for string results.
var sinkStr string

// ── Misuse 1: StopTimer / StartTimer in wrong order ───────────────────────
//
// A subtle but common bug: the developer intends to exclude per-iteration
// fixture construction, but accidentally places StartTimer AFTER the work
// instead of BEFORE it. The work runs while the timer is stopped; the timer
// runs during the next iteration's fixture construction.
//
// Effect: ns/op measures fixture cost, not work cost — the opposite of intent.
//
// NOTE: The even-more-dramatic variant (StopTimer with NO StartTimer) causes
// b.duration to stay at 0 forever, so the testing framework keeps doubling
// b.N and the benchmark never exits. Don't run that live.

// BenchmarkProcess_TimerOrder_BUG: the Stop/Start pair brackets the work
// instead of the setup — the exact inversion of what was intended.
//
// The fixture is built with the timer RUNNING, and processString runs with the
// timer STOPPED. So ns/op reports the cost of buildFixture, and the function
// actually under test contributes nothing to the measurement.
//
// Expected: ns/op ≈ cost of buildFixture, and _Correct is the cheaper-looking
// of the two even though it measures strictly more work.
//
// Note on ordering: the pair must leave the timer running across some part of
// each iteration. A variant that stops the timer first and starts it last
// (StopTimer, fixture, work, StartTimer) leaves b.duration at ~0 forever, so
// the framework keeps doubling b.N and the benchmark never exits — the same
// failure as omitting StartTimer entirely. Don't run that live.
func BenchmarkProcess_TimerOrder_BUG(b *testing.B) {
	var s string
	for range b.N {
		input := buildFixture(fixtureSize) // BUG: fixture is timed
		b.StopTimer()
		s = processString(input) // BUG: the work under test is not timed
		b.StartTimer()
	}
	sinkStr = s
}

// ── Correct: StopTimer / StartTimer pair ──────────────────────────────────

// BenchmarkProcess_PerIterSetup_Correct measures only processString,
// excluding per-iteration fixture construction from timing.
//
// Note: StopTimer/StartTimer add ~10–50 ns overhead per pair on most
// hardware. Use this pattern only when fixture construction cost exceeds
// that overhead — which buildFixture(128) easily does.
func BenchmarkProcess_PerIterSetup_Correct(b *testing.B) {
	var s string
	for range b.N {
		b.StopTimer()
		input := buildFixture(fixtureSize)
		b.StartTimer() // ← timer restarts; only processString is measured
		s = processString(input)
	}
	sinkStr = s
}

// ── Misuse 2: missing ResetTimer after one-time setup ─────────────────────
//
// b.ResetTimer() matters when one-time setup before the loop is expensive
// relative to total benchmark time. Without it, the timer runs from the
// moment b.N is set — including the setup phase.
//
// WHEN it matters: low b.N (e.g. -benchtime=1x or -count=1) + setup >>
// per-iteration cost. With large b.N the setup is amortised and the effect
// is invisible — so prefer running this demo with -benchtime=1x.

// BenchmarkProcess_OneTimeSetup_MissingReset: setup cost included in timing.
// Run with: go test -bench=OneTime -benchtime=1x -count=20 .
func BenchmarkProcess_OneTimeSetup_MissingReset(b *testing.B) {
	// One-time setup: build a large fixture once, then reuse.
	input := buildFixture(fixtureSize * 64) // ~8KB, takes measurable µs
	var s string
	for range b.N {
		s = processString(input)
	}
	sinkStr = s
}

// BenchmarkProcess_OneTimeSetup_Correct: b.ResetTimer excludes setup.
func BenchmarkProcess_OneTimeSetup_Correct(b *testing.B) {
	input := buildFixture(fixtureSize * 64)
	b.ResetTimer() // ← timer restarts from zero; setup excluded
	var s string
	for range b.N {
		s = processString(input)
	}
	sinkStr = s
}

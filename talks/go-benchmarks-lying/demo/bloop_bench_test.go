// bloop_bench_test.go demonstrates testing.B.Loop (introduced in Go 1.24)
// and contrasts it with the classic b.N loop.
//
// testing.B.Loop solves three b.N footguns in a single language construct:
//  1. The "forget b.ResetTimer after setup" mistake
//  2. The "b.N == 0 on the first iteration" edge case (benchmarks that must
//     allocate a result slice of length b.N will panic or produce wrong
//     numbers when b.N is 0)
//  3. Compiler inlining + DCE can still gut b.N loops; B.Loop suppresses
//     inlining of the loop body at the call site
//
// Run:
//
//	go test -bench=BenchmarkHash -benchmem -count=10 . 2>&1 | tee /tmp/bloop.txt
//	benchstat /tmp/bloop.txt
package demo

import (
	"crypto/sha256"
	"testing"
)

var payload = []byte("GopherCon UK 2026 — benchmark truth-telling")

// ── Classic b.N loop ──────────────────────────────────────────────────────

// BenchmarkHash_BN is the traditional form.
// Works correctly here because sha256.Sum256 has a visible side effect
// (returns a value we sink). But it has three known footguns — see below.
func BenchmarkHash_BN(b *testing.B) {
	var s [32]byte
	for range b.N {
		s = sha256.Sum256(payload)
	}
	_ = s
}

// BenchmarkHash_BN_WithSetup shows the classic b.ResetTimer footgun.
// The expensive setup is included in the timing if you forget ResetTimer.
func BenchmarkHash_BN_WithSetup_MissingReset(b *testing.B) {
	// Simulate expensive setup (e.g. opening a file, preparing a fixture).
	data := make([]byte, 1024)
	copy(data, payload)
	// BUG: forgot b.ResetTimer() here — setup cost pollutes the measurement.
	var s [32]byte
	for range b.N {
		s = sha256.Sum256(data)
	}
	_ = s
}

// BenchmarkHash_BN_WithSetup_Correct is the fixed version.
func BenchmarkHash_BN_WithSetup_Correct(b *testing.B) {
	data := make([]byte, 1024)
	copy(data, payload)
	b.ResetTimer() // ← exclude setup from timing
	var s [32]byte
	for range b.N {
		s = sha256.Sum256(data)
	}
	_ = s
}

// ── testing.B.Loop (Go 1.24) ─────────────────────────────────────────────
//
// B.Loop replaces `for range b.N` with `for b.Loop()`.
// The testing package calls the benchmark function multiple times, each time
// with a different iteration count, until the timer stabilises. B.Loop
// handles ResetTimer internally — setup before the loop is automatically
// excluded from measurement.

// BenchmarkHash_BLoop is the idiomatic Go 1.24 form.
// Setup before the loop is excluded automatically — no b.ResetTimer needed.
func BenchmarkHash_BLoop(b *testing.B) {
	// Setup: excluded from timing automatically.
	data := make([]byte, 1024)
	copy(data, payload)

	var s [32]byte
	for b.Loop() { // ← each call to Loop() is one measured iteration
		s = sha256.Sum256(data)
	}
	_ = s
}

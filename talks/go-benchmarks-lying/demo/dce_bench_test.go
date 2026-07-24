// Package demo contains live-demo benchmarks for
// "Why Your Go Benchmarks Are Lying (And How to Stop Them)"
// GopherCon UK 2026.
//
// This file demonstrates dead-code elimination (DCE) and
// constant folding silently gutting benchmark loops.
//
// Run:
//
//	go test -bench=. -benchmem -count=10 ./... 2>&1 | tee /tmp/demo.txt
//	benchstat /tmp/demo.txt
package demo

import (
	"math/bits"
	"testing"
)

// ── Dead-Code Elimination ──────────────────────────────────────────────────
//
// The Go compiler eliminates computations whose results are never observed.
// In a benchmark, if you call a function and discard its return value, the
// compiler is free to remove the entire call. The loop still runs b.N times,
// but each iteration is empty.
//
// DEMO STRATEGY: We detect DCE via *allocation counts* (-benchmem), not
// nanoseconds. On modern hardware (esp. Apple Silicon) the timer floor is
// ~0.25 ns — too close to zero to distinguish "eliminated" from "extremely
// fast". But allocs/op = 0 vs 1 is unmistakable on any hardware.
//
// makeBuffer allocates — so if DCE fires, allocs/op = 0 (allocation never
// happened). If the sink is correct, allocs/op = 1.
func makeBuffer(n int) []byte {
	return make([]byte, n) // heap-escaping allocation
}

// sink is the package-level variable that defeats DCE.
// Because the compiler cannot prove it is never read externally,
// it must retain the computation that produces the value assigned to it.
var sink []byte

// BenchmarkMakeBuffer_DCE: result of makeBuffer is unused → DCE fires.
// Expected: allocs/op = 0 (the allocation never happens).
func BenchmarkMakeBuffer_DCE(b *testing.B) {
	for range b.N {
		makeBuffer(64) // result discarded → compiler removes the call
	}
}

// BenchmarkMakeBuffer_Correct: result is sunk → DCE cannot fire.
// Expected: allocs/op = 1.
//
// Two-variable idiom: local s accumulates inside the loop (cheap), and the
// final sink = s after the loop is what forces the compiler to keep the
// entire computation chain. One global write per benchmark run, not per iter.
func BenchmarkMakeBuffer_Correct(b *testing.B) {
	var s []byte
	for range b.N {
		s = makeBuffer(64)
	}
	sink = s
}

// ── Constant Folding ───────────────────────────────────────────────────────
//
// If all inputs to an expression are compile-time constants, the Go compiler
// evaluates the result at compile time and replaces the call with a literal.
// The benchmark then iterates over a constant load — zero real work.
//
// DEMO STRATEGY: On ARM64 (Apple Silicon), bits.OnesCount is a single
// hardware instruction (~1 cycle). Timing cannot distinguish the constant-
// folded version from the runtime version at the timer floor (~0.3 ns).
//
// Verify via assembly inspection instead — the assembly does not lie:
//
//   go build -gcflags='-S' . 2>&1 | grep -A5 "OnesCount_ConstantFolded"
//
// ConstantFolded: you see  MOVD $3, Rxx  — constant loaded, no computation.
// Correct:        you see  VCNT + UADDLV  — actual popcount instruction pair.
//
// bits.OnesCount(0b10110) = 3 (three set bits). Compiler evaluates at compile
// time. The Correct version routes through a package-level variable, which
// the compiler cannot evaluate at compile time (its value may change).

// sinkInt is a separate package-level sink for integer results.
var sinkInt int

// BenchmarkOnesCount_ConstantFolded: constant input → folded at compile time.
// Timing: indistinguishable from Correct on ARM64 (both at timer floor).
// Proof: "go build -gcflags='-S' ." shows MOVD $3, not a VCNT instruction.
func BenchmarkOnesCount_ConstantFolded(b *testing.B) {
	var s int
	for range b.N {
		s = bits.OnesCount(0b10110) // constant → evaluated at compile time
	}
	sinkInt = s
}

// onesInput breaks the constant chain — compiler cannot prove this is 0b10110.
var onesInput uint = 0b10110

// BenchmarkOnesCount_Correct: runtime value → actual OnesCount instruction.
func BenchmarkOnesCount_Correct(b *testing.B) {
	var s int
	for range b.N {
		s = bits.OnesCount(onesInput)
	}
	sinkInt = s
}

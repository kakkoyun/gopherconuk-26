// Package benchfixtures contains benchmark fixtures used by honestbench tests.
// These files live under testdata/ so the go tool ignores them during normal
// build and test traversal.
package benchfixtures

import (
	"crypto/sha256"
	"testing"
)

// package-level sinks to defeat DCE.
var (
	sinkBytes [32]byte
	sinkSlice []byte
)

// BenchmarkGood_BLoop: correct modern form — b.Loop(), result sinked.
// Expected findings: none.
func BenchmarkGood_BLoop(b *testing.B) {
	payload := []byte("GopherCon UK 2026")
	var s [32]byte
	for b.Loop() {
		s = sha256.Sum256(payload)
	}
	sinkBytes = s
}

// BenchmarkGood_BN_WithSink: classic b.N loop, but result written to package-level sink.
// Expected findings: suggest-bloop (info only — no high/medium).
func BenchmarkGood_BN_WithSink(b *testing.B) {
	payload := []byte("GopherCon UK 2026")
	var s [32]byte
	for range b.N {
		s = sha256.Sum256(payload)
	}
	sinkBytes = s
}

// BenchmarkGood_TimerCorrect: StopTimer → setup → StartTimer → work.
// Expected findings: suggest-bloop (info only).
func BenchmarkGood_TimerCorrect(b *testing.B) {
	var s [32]byte
	for range b.N {
		b.StopTimer()
		data := make([]byte, 64)
		b.StartTimer()
		s = sha256.Sum256(data)
	}
	sinkBytes = s
}

// BenchmarkGood_Alloc: allocates result in loop, sinked at end.
// Expected findings: suggest-bloop (info only).
func BenchmarkGood_Alloc(b *testing.B) {
	var s []byte
	for range b.N {
		s = make([]byte, 64)
	}
	sinkSlice = s
}

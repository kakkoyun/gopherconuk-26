// Package benchfixtures contains benchmark fixtures used by honestbench tests.
// These files live under testdata/ so the go tool ignores them during normal
// build and test traversal.
package benchfixtures

import (
	"crypto/sha256"
	"testing"
)

// makeResult allocates and returns a slice — a non-void function.
func makeResult(n int) []byte {
	return make([]byte, n)
}

// BenchmarkBad_DiscardedResult: result of makeResult() is discarded in the loop.
// Expected findings: discarded-result (high) + suggest-bloop (info).
func BenchmarkBad_DiscardedResult(b *testing.B) {
	for range b.N {
		makeResult(64) // result discarded → DCE may remove the allocation
	}
}

// BenchmarkBad_MissingSink: accumulates in local, then discards with _ = .
// Expected findings: missing-sink (medium) + suggest-bloop (info).
func BenchmarkBad_MissingSink(b *testing.B) {
	var s [32]byte
	for range b.N {
		s = sha256.Sum256([]byte("hello"))
	}
	_ = s // does NOT defeat DCE; compiler sees through this
}

// BenchmarkBad_StopTimerNoStart: StopTimer in loop, no StartTimer.
// Expected findings: stoptimer-without-starttimer (high) + suggest-bloop (info).
func BenchmarkBad_StopTimerNoStart(b *testing.B) {
	var s [32]byte
	for range b.N {
		b.StopTimer()
		data := make([]byte, 64)
		s = sha256.Sum256(data) // timer never restarts
	}
	_ = s
}

// BenchmarkBad_TimerOrder: work happens while the timer is stopped.
// StartTimer is called last — after the work — which is the wrong order.
// Expected findings: stoptimer-without-starttimer (high) + suggest-bloop (info).
func BenchmarkBad_TimerOrder(b *testing.B) {
	var s [32]byte
	for range b.N {
		b.StopTimer()
		data := make([]byte, 64)
		s = sha256.Sum256(data) // measured work runs while timer is off
		b.StartTimer()          // too late: restarts timer after the work is done
	}
	_ = s
}

package main

import (
	"math"
	"testing"
)

// TestParseBenchOutput verifies that realistic go test -bench output lines are
// parsed correctly and grouped by benchmark name with the GOMAXPROCS suffix
// stripped.
func TestParseBenchOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string][]float64
	}{
		{
			name:  "single benchmark single run",
			input: `BenchmarkFoo-8   1000000   123.4 ns/op   64 B/op   1 allocs/op`,
			want:  map[string][]float64{"BenchmarkFoo": {123.4}},
		},
		{
			name:  "GOMAXPROCS suffix -16 stripped",
			input: `BenchmarkBar-16   500000   99.1 ns/op   0 B/op   0 allocs/op`,
			want:  map[string][]float64{"BenchmarkBar": {99.1}},
		},
		{
			name: "multiple runs grouped under same name",
			input: `
BenchmarkFoo-8   1000000   100.0 ns/op
BenchmarkFoo-8   1000000   110.0 ns/op
BenchmarkFoo-8   1000000   120.0 ns/op
`,
			want: map[string][]float64{"BenchmarkFoo": {100.0, 110.0, 120.0}},
		},
		{
			name: "two distinct benchmarks",
			input: `
BenchmarkMakeBuffer-8   1000000   45.2 ns/op   128 B/op   1 allocs/op
BenchmarkOnesCount-8    5000000    5.1 ns/op     0 B/op   0 allocs/op
`,
			want: map[string][]float64{
				"BenchmarkMakeBuffer": {45.2},
				"BenchmarkOnesCount":  {5.1},
			},
		},
		{
			name: "non-benchmark lines are ignored",
			input: `
goos: darwin
goarch: arm64
pkg: github.com/kakkoyun/gopherconuk-26/demo
cpu: Apple M4 Pro
BenchmarkOnesCount-8   5000000   5.2 ns/op
PASS
ok  	github.com/kakkoyun/gopherconuk-26/demo	0.456s
`,
			want: map[string][]float64{"BenchmarkOnesCount": {5.2}},
		},
		{
			name:  "empty input",
			input: "",
			want:  map[string][]float64{},
		},
		{
			name:  "no benchmark lines",
			input: "goos: linux\ngoarch: amd64\nPASS\n",
			want:  map[string][]float64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBenchOutput(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parseBenchOutput() returned %d benchmarks, want %d; got=%v", len(got), len(tt.want), got)
				return
			}
			for name, wantSamples := range tt.want {
				gotSamples, ok := got[name]
				if !ok {
					t.Errorf("parseBenchOutput(): missing benchmark %q in result", name)
					continue
				}
				if len(gotSamples) != len(wantSamples) {
					t.Errorf("parseBenchOutput()[%q] got %d samples, want %d", name, len(gotSamples), len(wantSamples))
					continue
				}
				for i, want := range wantSamples {
					if math.Abs(gotSamples[i]-want) > 1e-9 {
						t.Errorf("parseBenchOutput()[%q][%d] = %f, want %f", name, i, gotSamples[i], want)
					}
				}
			}
		})
	}
}

// TestComputeCV verifies mean, stddev, CV%, and the note string for a variety
// of sample inputs including degenerate cases.
func TestComputeCV(t *testing.T) {
	const eps = 0.02 // tolerance for floating-point comparison

	tests := []struct {
		name     string
		samples  []float64
		wantMean float64
		wantCV   float64
		wantNote string
	}{
		{
			name:     "identical samples → cv=0",
			samples:  []float64{100, 100, 100},
			wantMean: 100,
			wantCV:   0,
		},
		{
			name:    "known cv: [90,100,110]",
			samples: []float64{90, 100, 110},
			// mean=100, variance=((−10)²+0²+10²)/2=100, stddev=10, cv=10%
			wantMean: 100,
			wantCV:   10.0,
		},
		{
			name:     "single sample → insufficient samples note",
			samples:  []float64{42},
			wantMean: 42,
			wantCV:   0,
			wantNote: "insufficient samples",
		},
		{
			name:     "empty slice → no samples note",
			samples:  []float64{},
			wantMean: 0,
			wantCV:   0,
			wantNote: "no samples",
		},
		{
			name:    "two samples → sample stddev uses n-1",
			samples: []float64{100, 200},
			// mean=150, variance=(50²+50²)/1=5000, stddev≈70.71, cv≈47.14%
			wantMean: 150,
			wantCV:   47.14,
		},
		{
			name:     "large identical count",
			samples:  []float64{500, 500, 500, 500, 500},
			wantMean: 500,
			wantCV:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mean, _, cv, note := computeCV(tt.samples)
			if math.Abs(mean-tt.wantMean) > eps {
				t.Errorf("computeCV() mean=%f, want %f", mean, tt.wantMean)
			}
			if math.Abs(cv-tt.wantCV) > eps {
				t.Errorf("computeCV() cv=%f, want %f", cv, tt.wantCV)
			}
			if note != tt.wantNote {
				t.Errorf("computeCV() note=%q, want %q", note, tt.wantNote)
			}
		})
	}
}

// TestVerdictLogic exercises the pass/fail counting logic independently of
// any I/O or subprocess invocation.
func TestVerdictLogic(t *testing.T) {
	tests := []struct {
		name        string
		samples     map[string][]float64
		threshold   float64
		wantVerdict string
		wantFailing int
	}{
		{
			name: "all benchmarks pass at threshold 15",
			samples: map[string][]float64{
				"BenchmarkFoo": {100, 100, 100}, // cv=0%
				"BenchmarkBar": {90, 100, 110},  // cv=10%
			},
			threshold:   15.0,
			wantVerdict: "PASS",
			wantFailing: 0,
		},
		{
			name: "one benchmark fails at threshold 5",
			samples: map[string][]float64{
				"BenchmarkFoo": {100, 100, 100}, // cv=0%
				"BenchmarkBar": {90, 100, 110},  // cv=10% > 5%
			},
			threshold:   5.0,
			wantVerdict: "FAIL",
			wantFailing: 1,
		},
		{
			name: "all benchmarks fail",
			samples: map[string][]float64{
				"BenchmarkA": {90, 100, 110},  // cv=10%
				"BenchmarkB": {100, 100, 200}, // high cv
			},
			threshold:   5.0,
			wantVerdict: "FAIL",
			wantFailing: 2,
		},
		{
			name: "exactly at threshold → passes",
			samples: map[string][]float64{
				// cv=10%, threshold=10 → cv <= threshold → PASS
				"BenchmarkX": {90, 100, 110},
			},
			threshold:   10.0,
			wantVerdict: "PASS",
			wantFailing: 0,
		},
		{
			name: "just over threshold → fails",
			samples: map[string][]float64{
				"BenchmarkX": {90, 100, 110}, // cv=10%
			},
			threshold:   9.99,
			wantVerdict: "FAIL",
			wantFailing: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failing := 0
			for _, s := range tt.samples {
				_, _, cv, _ := computeCV(s)
				if cv > tt.threshold {
					failing++
				}
			}
			pass := failing == 0
			gotVerdict := verdictLabel(pass)
			if gotVerdict != tt.wantVerdict {
				t.Errorf("verdict=%q, want %q", gotVerdict, tt.wantVerdict)
			}
			if failing != tt.wantFailing {
				t.Errorf("failing=%d, want %d", failing, tt.wantFailing)
			}
		})
	}
}

// TestVerdictLabel is a simple sanity check for the label helper.
func TestVerdictLabel(t *testing.T) {
	if got := verdictLabel(true); got != "PASS" {
		t.Errorf("verdictLabel(true) = %q, want %q", got, "PASS")
	}
	if got := verdictLabel(false); got != "FAIL" {
		t.Errorf("verdictLabel(false) = %q, want %q", got, "FAIL")
	}
}

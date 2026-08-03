# Demo Makefile — GopherCon UK 2026
# "Why Your Go Benchmarks Are Lying (And How to Stop Them)"
#
# Requirements: Go 1.24+, benchstat (golang.org/x/perf)
# Install tools: make tools

BENCH_COUNT  ?= 10
BENCH_TIME   ?= 1s
BENCH_OUTPUT := /tmp/gopherconuk26-bench.txt
BENCHSTAT    := $(shell go env GOPATH)/bin/benchstat
BENCHDIFF    := $(shell go env GOPATH)/bin/benchdiff

.PHONY: tools bench bench-dce bench-bloop bench-timer bench-docker cv stat clean help

## tools: install benchstat and benchdiff
tools:
	go install golang.org/x/perf/cmd/benchstat@latest
	go install github.com/willabides/benchdiff/cmd/benchdiff@latest

## bench: run all benchmarks and pipe output to benchstat
bench: $(BENCHSTAT)
	go test -bench=. -benchmem -count=$(BENCH_COUNT) -benchtime=$(BENCH_TIME) . \
		| tee $(BENCH_OUTPUT)
	$(BENCHSTAT) $(BENCH_OUTPUT)

## bench-dce: run only DCE/constant-folding benchmarks
bench-dce: $(BENCHSTAT)
	go test -bench='BenchmarkSum|BenchmarkOnesCount' -benchmem \
		-count=$(BENCH_COUNT) -benchtime=$(BENCH_TIME) . \
		| tee $(BENCH_OUTPUT)
	$(BENCHSTAT) $(BENCH_OUTPUT)

## bench-bloop: run B.Loop vs b.N benchmarks
bench-bloop: $(BENCHSTAT)
	go test -bench='BenchmarkHash' -benchmem \
		-count=$(BENCH_COUNT) -benchtime=$(BENCH_TIME) . \
		| tee $(BENCH_OUTPUT)
	$(BENCHSTAT) $(BENCH_OUTPUT)

## bench-timer: run ResetTimer/StopTimer benchmarks
bench-timer: $(BENCHSTAT)
	go test -bench='BenchmarkProcess' -benchmem \
		-count=$(BENCH_COUNT) -benchtime=$(BENCH_TIME) . \
		| tee $(BENCH_OUTPUT)
	$(BENCHSTAT) $(BENCH_OUTPUT)

## bench-docker: CV experiment — idle host vs loaded host vs pinned container
bench-docker:
	./cv-experiment.sh

## cv: recompute the CV table from the captured results
cv:
	@for f in results/idle.txt results/noisy.txt results/pinned.txt; do \
		echo "## $$(basename $$f .txt)"; awk -f cv.awk $$f; echo; \
	done

## stat: run benchstat on the last captured output
stat: $(BENCHSTAT)
	$(BENCHSTAT) $(BENCH_OUTPUT)

## asm-dce: show assembly for DCE benchmarks — proves constant folding
asm-dce:
	@echo "=== Assembly: BenchmarkOnesCount_ConstantFolded ==="
	@go build -gcflags='-S' . 2>&1 | grep -A10 "OnesCount_Constant" || true
	@echo "=== Assembly: BenchmarkOnesCount_Correct ==="
	@go build -gcflags='-S' . 2>&1 | grep -A10 "OnesCount_Correct" | head -15 || true

## clean: remove captured benchmark output
clean:
	rm -f $(BENCH_OUTPUT)

## help: show this help
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'

$(BENCHSTAT):
	@echo "benchstat not found — run: make tools"
	@exit 1

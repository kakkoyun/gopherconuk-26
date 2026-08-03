#!/usr/bin/awk -f
# Compute coefficient of variation (CV = sigma/mu) per benchmark from
# `go test -bench` output. benchstat reports a confidence interval but not CV,
# and CV is the number that tells you whether the *environment* is trustworthy.
#
# Usage: awk -f cv.awk results.txt

/ns\/op/ {
    name = $1
    sub(/-[0-9]+$/, "", name)
    for (i = 1; i <= NF; i++) {
        if ($i == "ns/op") {
            v = $(i - 1) + 0
            n[name]++
            sum[name] += v
            sumsq[name] += v * v
        }
    }
}

END {
    printf "%-46s %6s %12s %12s %8s\n", "benchmark", "runs", "mean ns/op", "stddev", "CV%"
    for (b in n) {
        if (n[b] < 2) continue
        mu = sum[b] / n[b]
        var = (sumsq[b] - n[b] * mu * mu) / (n[b] - 1)
        if (var < 0) var = 0
        sd = sqrt(var)
        printf "%-46s %6d %12.4f %12.4f %8.2f\n", b, n[b], mu, sd, (mu > 0 ? 100 * sd / mu : 0)
    }
}

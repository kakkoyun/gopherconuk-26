# USDT for Go

> **Status:** ✅ Research complete (deep-research workflow + journal, 2026-07-22)
> **Blog post:** "Go runtime futures: flight recording, USDT, and hook proposals" (combined with 06-runtime-futures.md)

---

## 1. What USDT Probes Are

**USDT (User Statically-Defined Tracing)** probes are implemented entirely on top of uprobes at
the kernel level — not a separate kernel mechanism.

### How they work

1. **Compile time:** A NOP instruction is compiled into the binary at the probe site.
   Metadata is written into the **`.note.stapsdt` ELF section**, recording:
   - The link-time PC address of the probe site (runtime address requires `.stapsdt.base`
     correction for PIE/ASLR binaries).
   - Provider name and probe name.
   - Argument locations encoded as GAS assembler operands.

2. **Attach time:** libbpf parses the `.note.stapsdt` ELF notes, locates the NOP, and calls
   the standard uprobe attach path to **replace the NOP with an INT3 interrupt** (x86) or
   architecture-equivalent. At this point it behaves identically to a uprobe.

3. **Unattached overhead:** A single NOP instruction — **near-zero overhead**. This is USDT's
   key advantage over uprobes at stable probe sites.

**Source:** [S-USDT-01] (ebpf.io USDT docs), [S-USDT-02] (SystemTap UserSpaceProbeImplementation spec)

---

## 2. Performance: uprobes vs USDTs

| Mechanism | Overhead (unattached) | Overhead (attached, kernel) | Notes |
|-----------|----------------------|-----------------------------|-------|
| **USDT (NOP1)** | ~0 (single NOP) | ~3000+ ns (two context switches) | Kernel uprobe path once attached |
| **Kernel uprobe** | ~0 (also NOP) | ~3224 ns | Same kernel path as attached USDT |
| **bpftime userspace uprobe** | ~0 | **~314 ns** | ~10x lower latency, no kernel context switch |
| **Proposed NOP5 uprobe** | ~0 | Competitive with USDT | PATCHv6, kernel mailing list, Sept 2025 |

**Key finding:** Once attached, USDT and kernel uprobes go through the **same kernel path** —
the performance advantage of USDT is only when *unattached* (NOP1 vs NOP + uprobe lookup).
The Polar Signals / Oligo claim "uprobes as fast as USDTs" refers to this steady-state
convergence: at the kernel level they are the same mechanism.

**bpftime** (USENIX ATC 2024 paper) achieves ~10x lower latency by eliminating dual context
switches via a userspace eBPF runtime. This is a separate performance story.

**Source:** [S-USDT-03] (bpftime paper, arXiv 2311.07923, peer-reviewed USENIX ATC 2024),
[S-USDT-04] (kernel mailing list PATCHv6, Sept 2025, Andrii Nakryiko)

---

## 3. libstapsdt — Runtime USDT Probe Creation

**`github.com/linux-usdt/libstapsdt`** enables runtime USDT probes for interpreted and
JIT-compiled languages (including Go) without compile-time changes:

1. Generates a small ELF shared library (`.so`) on the fly, containing USDT probe sites in its
   `.note.stapsdt` section.
2. `dlopen()`s that `.so` at runtime, making the probe sites visible to the kernel and to
   tools like bpftrace and perf.

This is the mechanism that makes "runtime USDT injection" possible — no source recompilation needed.

**Source:** [S-USDT-05] (`github.com/linux-usdt/libstapsdt` README + src/libstapsdt.c)

---

## 4. Salp — Go Binding to libstapsdt

**`github.com/mmcshane/salp`** is the canonical Go binding to libstapsdt.

⚠️ **Repo URL correction:** Prior lead cited `github.com/mmcloughlin/salp` — this returns 404.
The correct repo is **`github.com/mmcshane/salp`**.

- Wraps libstapsdt → libelf + libdl.
- Requires **CGo** — pure Go is not possible with this approach.
- Check current status for Go version compatibility before citing in the talk.

**Source:** [S-USDT-06] (`github.com/mmcshane/salp` README), [S-USDT-05] (libstapsdt README listing Go wrappers)

---

## 5. libbpf/usdt — C-Only, No Go Bindings

**`github.com/libbpf/usdt`** is a C/C++ single-header library (`usdt.h`) that defines USDT
probe sites via C preprocessor variadic macros emitting inline assembly.

- **No Go bindings** — cannot be used directly from Go.
- For the talk: this is the consumer side (attaching to USDT probes from eBPF programs), not the
  producer side (emitting USDT probes from Go applications).

**Source:** [S-USDT-07] (`github.com/libbpf/usdt` README + usdt.h)

---

## 6. Go Runtime: No Built-In USDT Probes

The Go runtime **does not ship built-in USDT probes**. As of December 2024, the Go team was at
an **initial-inquiry stage** on golang/go#57175. Felix Geisendörfer asked whether USDT support
was on the roadmap; the Go team had not committed.

**Source:** [S-USDT-08] (golang/go#57175)

---

## 7. The Speaker's Own USDT POC 🎯

`github.com/kakkoyun/go/tree/poc_usdt` — a proof-of-concept fork of the Go standard library
that adds USDT probes to:
- `net/http`
- `database/sql`
- `crypto/tls`
- `net`

…via a `go tool usdt` subcommand. This has **not been proposed for upstream acceptance**.

**Talk angle:** This is the speaker's own work — ideal for the "what could be" section. A live
demo of `bpftrace` attaching to USDT probes in a net/http server compiled from this fork is
a compelling "here's what the future could look like" moment.

**Source:** [S-USDT-09] (kakkoyun.me FOSDEM 2026 talk summary)

---

## 8. Fit With the Talk

USDT fits in the **"runtime futures"** section (section 8, 1 minute):
- USDT is the "cleanest possible" hook — NOP overhead when unattached, standard ecosystem tooling
  (bpftrace, perf probe).
- Go doesn't have them yet, but the speaker has a POC showing what it would look like.
- The PATCHv6 kernel work (NOP5 uprobes) closes the performance gap even further.
- Together with flight recording (#63185, shipped Go 1.25) and #69887, this is the "the horizon
  is bright" closing beat.

---

## Open Questions (before talk/blog)

- [ ] Check golang/go#57175 current status — any progress since Dec 2024?
- [ ] Verify Salp (`github.com/mmcshane/salp`) Go version compatibility — does it build with current Go?
- [ ] Does the POC (`github.com/kakkoyun/go/tree/poc_usdt`) still compile against current Go?
- [ ] What is the status of the NOP5 uprobe kernel patch (PATCHv6, Sept 2025) — merged or still pending?
- [ ] Polar Signals USDT deep-dive (2025-12-10) — verify exact claim and benchmark methodology.

---

## Sources Used

| Key | Description | URL |
|-----|-------------|-----|
| S-USDT-01 | ebpf.io USDT concepts docs | https://docs.ebpf.io/linux/concepts/usdt/ |
| S-USDT-02 | SystemTap UserSpaceProbeImplementation spec | https://sourceware.org/systemtap/wiki/UserSpaceProbeImplementation |
| S-USDT-03 | bpftime paper (USENIX ATC 2024) | https://arxiv.org/pdf/2311.07923 |
| S-USDT-04 | Kernel mailing list PATCHv6 NOP5 uprobe (Andrii Nakryiko, Sept 2025) | https://www.mail-archive.com/linux-trace-kernel@vger.kernel.org/msg12197.html |
| S-USDT-05 | github.com/linux-usdt/libstapsdt README + source | https://github.com/linux-usdt/libstapsdt |
| S-USDT-06 | github.com/mmcshane/salp README | https://github.com/mmcshane/salp |
| S-USDT-07 | github.com/libbpf/usdt README + usdt.h | https://github.com/libbpf/usdt |
| S-USDT-08 | golang/go#57175 (USDT in Go runtime) | https://github.com/golang/go/issues/57175 |
| S-USDT-09 | kakkoyun.me FOSDEM 2026 talk summary | https://kakkoyun.me/posts/fosdem-2026-auto-instrumenting-go/ |

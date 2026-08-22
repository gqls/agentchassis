# HANDOFF — `bugs_closed/318` release source coverage · 2026-08-22

> # ✅ THIS LANE IS CLOSED — 2026-08-22. **Nothing below is live work.**
>
> `318` is fixed, LIVE and verified on `v1.0.1326`, and has moved to `bugs_closed/`.
> Every candidate is resolved: the main gap by an inverted predicate, candidate 2 by
> construction, candidate 1 by an **owner ruling** (skipped, not deferred — do not re-file
> it), candidate 3 by the cluster census.
>
> **If you are here because a grep sent you, the file you want is
> `bugs_closed/318_HANDOFF_2026-08-19_a_service_whose_image_the_release_does_not_build_is_still_invisible_to_the_coverage_gate.md`**
> — its closing block carries the release evidence.
>
> **Read-out for the owner:** `SUMMARY_2026-08-22b_closed_and_live.md`.
> **Commands + every trap:** `RUNBOOK_release_source_coverage.md`.

## What exists now, and how to run it

| | command | what it asks | hard or advisory |
|---|---|---|---|
| the gate | `make check-release-coverage` | *can a release reach this service?* (filesystem) | **hard** — refuses the release |
| the census | `make release-census` | *does what is running match what is declared?* (cluster) | reports only |
| the nudge | automatic, `scripts/pattern-check.py` | fires when a commit ADDS an overlay pinning one of our images that the makefile never names | advisory |

Register: **BLD-026**. Council: gate `83442a5a` (APPROVED r1), census `b0883c17`
(REVISE → APPROVED r2). BLD-022 corrected in three places.

## The three things deliberately left, so nobody files them as gaps

1. **The census has no driver.** Hand-run only — no CronJob, no RBAC, no `doc_notes` row.
   Stated at the makefile target, in `pkg/releaseset/census.go` and in BLD-026. If you pick
   this up, it is its own commit and its own council round, and its first scheduled run must
   be verified at the artefact: *detection works; schedule and dispatch do not.*
2. **A tag is not the code.** A clean census means *"every service is on the tag it should be
   on"*, **not** *"every service is running the code it should be running"* — a same-tag
   rebuild serves the node's cached image. The artefact-side answer is BLD-019/020/023's;
   duplicating it here would give one question two answers that can disagree.
3. **Seven inline `kubernetes.NewForConfig` bootstraps, no shared wrapper.** Named in
   BLD-026 with all seven paths. The `architecture` seat's warning is that unwinding seven
   call sites later costs an RFC — so this is worth doing *before* an eighth. The one real
   difference between them is the in-cluster-only vs kubeconfig split.

## ⚠ Traps this lane paid for — read before verifying anything here

- **A green gate proves nothing on a compliant tree.** It could only have passed. The
  discriminating evidence is the six mutation controls in `RUNBOOK` R7 — and run them on a
  **copy** of the makefile, never the live file (`WRONG_CALLS.md` 2026-08-22, `f016b07ec`).
- **`go run` collapses exit 2 into 1** `[MEASURED]`, so "the release is wrong" and "the
  check is broken" are indistinguishable by status. The message says which in words; a
  scripted consumer must use the compiled binary.
- **Never read an exit code off a pipeline.** `make … | tail -3; echo $?` reports `tail`'s.
  This lane did it while proving a mutation control fires, and printed a reassuring `exit=0`
  beside a failure message.
- **Comparing `v1.0.NNNN` tags as strings inverts across `v1.0.999 → v1.0.1000`**, and we
  crossed that long ago. It bit three times in one file in one afternoon — the shipped
  report, the first repair, and the tie-break the council caught. In `LANDMINES.md`.
- **A production overlay's `newTag` is edited by the release and never committed**, so
  `git log` on the overlay dates a pin that has not been true for months. This is what
  killed candidate 1 as worded. In `LANDMINES.md`.
- **When a finding has a DIRECTION, write the fixture for both** before running it anywhere.
  A one-sided fixture cannot fail on the side it omits, and it reads as thorough.

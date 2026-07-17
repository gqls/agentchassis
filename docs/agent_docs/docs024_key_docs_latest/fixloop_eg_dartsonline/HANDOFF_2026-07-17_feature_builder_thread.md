# HANDOFF — start the FEATURE BUILDER thread

**Filed 2026-07-17** from "diagnosis fixloop 3". Cold-start for a NEW thread that
builds the multi-step feature-construction capability on top of the proven fix
loop. **The full design is `DESIGN_feature_builder_and_council_gate.md` §1 — read
it first; this doc adds only thread-boot context and first steps.**

## Working rules
Go, not Python. British English. Schema first. Reuse before recreate — every
component below extends an EXISTING, production-proven piece; build nothing from
scratch. Deploy only from a committed ref (`make build-agent-chassis-ref`); bump
IMAGE_TAG; verify the RUNNING POD binary. Commit per task, explicit paths, read
`git diff --cached --name-only` immediately before committing (the shared index
races — proven three times on 2026-07-16/17).

## What already exists (do not rebuild)
- The chain: diagnose→plan→council(3 seats: right-edits, platform-safety,
  bug-historian; fix-proposer v6)→implement (caged pod, git-adapter-only writes,
  hard file allowlist)→build gate (golang k8s Job)→PR. PR #1 merged 2026-07-13.
- Intake pool: triage already routes `capability_gap` items to the roadmap and
  deliberately NOT into the diagnosis loop — those items are this tool's feed.
- Artifacts: `diagnosis_artifacts` (kinds bundle|fix_plan|council_report|
  escalation) by correlation_id. Triggers 090–095 in
  `fixloop_eg_dartsonline/`; seeds `0NN_fix_proposer.sql` (v6) etc.

## The three build deltas (design §1, in build order)
1. **Staged-plan schema**: `fix_plan` grows `stages[]` — each stage a constrained
   edit plan (allowlist incl. TO-BE-CREATED files, expected symbols, gate
   criteria, inter-stage deps). Council router unchanged — it judges the artifact
   given.
2. **Implementer stage-loop**: `create_branch` once → one `branch-commit` per
   stage → build gate per stage → go test gate at end → ONE PR. New-file creation
   allowed via the stage manifest; the hard allowlist stays.
3. **Seeds ship as PR FILES, never executed** — image-then-seed ordering is a
   human checklist in the PR body. Feature = merged AND seeded, two owner acts.

## Human gates (all of them)
spec approval → design approval (council + owner) → mechanical stage gates →
PR merge (owner) → seed apply (owner). Nothing self-merges or self-seeds.

## First steps for this thread
1. Read the design §1; then read `0NN_fix_proposer.sql` and the fix-implementer
   orchestrator seed — the prompt seams you will extend.
2. Draft the `stages[]` schema and get owner sign-off BEFORE any code.
3. Pilot candidate (owner suggested self-hosting): the fix loop's own **F1.2
   cleanup** — make implementer ref/base a per-run INPUT (they are live-set to a
   stale branch). Small, known shape, gradable, and it fixes a standing gotcha.

## Gotchas inherited from the fix loop (cost hours; do not relearn)
- fix-implementer MUST fire via fix-implementer-orchestrator (else no read token).
- Delete stale `fix/*` branches before re-firing (create_branch reuses old base).
- Image FIRST, then seed; never rebuild an existing tag; ~300s rebalance window
  after any chassis pod (re)start — spawns are silently dropped (live instance:
  `bugs_open/003`, correlation 80c35dea).
- Each council/diagnosis run spends credits — owner gives the go per run.

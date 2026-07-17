# DESIGN (SIGNED OFF & BUILT) — the implementer stage-loop, feature-builder delta 2

*2026-07-17, "fixloop feature builder" thread. Delta 1 (staged-v1 schema +
validation `4b3d50f4c` + feature-designer seed draft) is built; the schema doc's
compatibility map reserved delta 2 for its own sign-off.*

**SIGNED OFF 2026-07-17: owner accepted all recommendations, E1–E5 (§3).**
Built same day: `feature_stage_route` + the four seam changes in commit
`c19b5d097` (tests green); `0NN_feature_implementer.sql` +
`0NN_feature_implementer_orchestrator.sql` + both trigger drafts in
`5b131b88a` — seed files only, owner applies after the image
(`RUNBOOK_feature_builder.md` A1–A2).

## 0. What the stage-loop is

Turns a council-APPROVED staged plan into `feat/<short-corr>` (D2) with ONE
commit per stage, a build gate per stage, a derived go-test end gate (D6), and
ONE PR. The cage is unchanged: dedicated pod (read token via the spawn gate),
writes only via git-adapter, hard per-stage allowlist, PR is the human terminal.

## 1. Workflow shape (feature-implementer v1)

```
load_plan → load_council → load_spec → check_approved (council=approved AND
plan_format=staged-v1) → stage_route ─┬→ [read_stage_files → implement_stage →
prepare_stage → commit_stage → stage_gate → check_gate] → stage_route (loop)
                                      └→ (exhausted) → test_gate → check_tests
                                             → create_pr → complete
```

- **`stage_route` (new Go action, `feature_stage_route`)** — the loop's only
  new control machinery, mirroring the diagnosis loop's iteration pattern:
  deterministic, no model judgement. Keeps `feature_state` in collected_data:
  current stage index, the staged plan parsed once, branch name, and per-stage
  outputs the existing actions consume unchanged —
  `stage_plan` (the current stage AS a single-plan shape: its `edits[]` +
  plan summary — exactly what `diagnose_read_repo_files` and
  `diagnose_prepare_fix_commit` already accept via `plan_field`),
  `read_ref` (base ref for stage 1; the feat branch thereafter, so later
  stages see earlier stages' commits), `gate_build` (from the stage), and
  `has_more`. Emitting stage-as-single-plan is the load-bearing trick: the
  proven read/prepare actions loop without knowing they are looping.
- **`create_branch` once** before the loop (git-adapter, idempotent), from a
  per-run `base_ref` input defaulting `main` — the F1.2 shape, built into this
  agent from birth rather than retrofitted.
- **`implement_stage`** — the existing whole-file sketch_to_files prompt,
  scoped to the stage's edits, plus a short rendered digest of PRIOR stages
  (id, title, files committed) so stage N can wire what stage N-1 created
  without re-reading it blind.
- **`prepare_stage`** — `diagnose_prepare_fix_commit` with three additions
  (all deterministic): accept a routed `branch` (instead of deriving
  `fix/<short>`), accept the stage plan via the existing `plan_field` seam,
  and enforce the stage's `expected_symbols` (each must appear verbatim in at
  least one produced file body — schema §2). Per-stage commit message:
  `feat(<short>) stage <id>: <stage title>`.
- **`stage_gate`** — `diagnose_build_gate` unchanged, changed-files = the
  stage's files; SKIPPED (routed around) when the stage's `gate.build` is
  false (all-seed/doc stages — validation already guarantees the implication).
  Red gate = run ends with branch + log, NO PR; stages 1..N-1 stay on the
  branch for inspection — same failure semantics the fix loop proved.
- **`test_gate` (end, D6)** — after the last stage: `go test` over the union
  of packages containing the plan's edited `.go` files, DERIVED in Go from the
  committed file list, never from a plan field. Red = no PR, branch + log
  remain.
- **`create_pr` once** — PR body is the staged Q-H package: spec summary,
  stage list (id/title/goal/files), council decision, and the
  `post_merge_checklist` rendered as an ordered GitHub task list — the owner's
  apply checklist is IN the PR, where the merge happens.

## 2. What is reused untouched vs changed

| Piece | Status |
|---|---|
| `diagnose_read_repo_files` | 1 change: `ref` resolvable from collected_data (routed per stage), config literal then `main` as fallbacks — the same change F1.2 needs, done once here |
| `diagnose_prepare_fix_commit` | 3 additions above; single-plan path byte-for-byte unchanged (fix loop untouched) |
| `diagnose_build_gate` | reused for stage gates; gains an opt-in `go_test` mode for the end gate (E2) |
| `git_adapter_request` create_branch/commit/create_pull_request | unchanged |
| council/persist/designer | unchanged (delta 1) |
| `feature_stage_route` | NEW deterministic Go action |
| `feature-implementer` + orchestrator seeds | NEW agent defs (E1), shipped as PR files per the discipline |
| `isRepoCloningAgent` gate | + `feature-implementer` (one list entry — the read token reaches the dedicated pod, mirror of fix-implementer's 2026-07-13 lesson) |

## 3. Decisions for the owner (E1–E5)

- **E1 — separate agent.** Recommended: NEW `feature-implementer` (+
  `feature-implementer-orchestrator`) agent defs; `fix-implementer` stays
  frozen on single plans. Alternative: teach fix-implementer to branch on
  plan_format (fewer defs, but mutates a proven production workflow).
- **E2 — end gate mechanism.** Recommended: extend `diagnose_build_gate` with
  an opt-in `go_test` mode (reuse the k8s Job machinery, packages derived).
  Alternative: a separate action (more code, same Job).
- **E3 — per-stage council.** Recommended: NO in v1 — the council approved the
  staged plan; per-stage review is mechanical (allowlist + symbols + gates).
  The parent design's "optionally per-stage diffs" stays a later option.
- **E4 — stale-branch hygiene.** Recommended: `create_branch` for `feat/*`
  refuses (fails the run) if the branch already exists, rather than silently
  reusing an old base — turns the fix loop's delete-stale-branches gotcha into
  a loud error. Alternative: keep the adapter's idempotent reuse.
- **E5 — pilot sequencing (confirming D5's mechanics).** Delta 2's own Go/seed
  changes are hand-built (this thread, commit-per-task) — the loop cannot build
  itself before it exists. The F1.2 pilot then runs THROUGH the finished loop:
  hand-written staged plan (schema §6), designer skipped (plan injected via
  persist on a fresh correlation), council + stage-loop + PR live. First
  feature = the loop fixing its own sibling's standing gotcha.

Sign-off on E1–E5 unblocks the delta-2 build: `feature_stage_route` + action
changes + tests first, then the two seed drafts, then the pilot dispatch plan.

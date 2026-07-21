# PLAN — feature builder (multi-step capability construction on the fix loop)

*Standing plan for the "fixloop feature builder" workstream. Companions:
`NOTES_running_feature_builder.md` (turn-by-turn record),
`RUNBOOK_feature_builder.md` (the owner's tasks),
`SUMMARY_feature_builder_2026-07-19.md` (read-aloud, current). Parent design:
`DESIGN_feature_builder_and_council_gate.md` §1. Last updated 2026-07-19.*

## Mission

Extend the proven diagnosis→fix loop so it can BUILD multi-step features — a
workflow AND the actions it calls, new files included — under the same cage:
constrained plans, council review, deterministic gates, writes only via
git-adapter, ONE PR, and every consequential act (spec approval, design
approval, merge, seed apply) a deliberate human decision.

## Architecture (as built)

spec (`site_work_items` capability_gap + `owner_approval` + `code_pointers`)
→ **feature-designer** (staged plan: `plan_format staged-v1`, stages[] with
per-stage allowlists incl. to-be-created files, image-before-seed checklist)
→ **5-seat council** (editquality, bug-historian, reuse-agent, guidelines,
guardian w/ hard veto — mirroring fix-proposer's v8 roster) + the proven
deterministic router (approve / revise+checks / reframe / escalate). The
reviser reads the **council_report artifact** (patch 017), not per-seat prompt
sections — so a 6th seat needs no prompt edit
→ **feature-implementer** (dedicated pod, read token only):
`feature_stage_route` walks the plan — one `feat/<short-corr>` branch, one
commit + build gate per stage, stage N reads the branch so it sees stage
N-1's work, derived `go test` end gate — → ONE PR whose body carries the
owner's post-merge checklist as a task list.

Governing decisions (all owner-approved 2026-07-17):
**D1** reuse `kind='fix_plan'` + `plan_format` discriminator · **D2**
`feat/<short-corr>` branches · **D3** caps 6 stages / 8 edits per stage / 24
total / 128KB · **D4** seed⇒checklist + image-strictly-before-seed as hard
validation · **D5** pilot = F1.2, self-hosted · **D6** go-test packages
derived from the plan, never declared ·
**E1** separate feature-implementer (fix-implementer frozen) · **E2** build
gate gains opt-in `go_test` mode · **E3** no per-stage council in v1 · **E4**
pre-existing `feat/*` branch = loud refusal · **E5** delta 2 hand-built, the
pilot runs through the finished loop.

## Status by delta

| Delta | What | Status |
|---|---|---|
| 1a. Staged-plan schema | `SCHEMA_staged_plan_v1.md` | **SIGNED OFF** 2026-07-17 (D1–D6) |
| 1b. Staged validation | `diagnose_persist_fix_plan_action.go` | **BUILT** — commit `4b3d50f4c`, tests green, legacy path unchanged |
| 1c. feature-designer seed | `0NN_feature_designer.sql` (+ patches 016, 017) | **APPLIED & LIVE**; 5-seat council; reviser reads the artifact |
| 2a. Stage-loop machinery | `feature_stage_route_action.go` + seams (read `ref_field`, prepare branch/message/symbols, gate `test_packages_field`, spawn gate) | **BUILT** — commit `c19b5d097`, tests green |
| 2b. feature-implementer seeds | `0NN_feature_implementer.sql` + `_orchestrator.sql` | **APPLIED & LIVE** 2026-07-17 — but **NEVER FIRED** |
| 2c. Triggers | `0NN_TRIGGER_feature_designer_v1.sh`, `0NN_TRIGGER_feature_implementer_v1.sh` | designer **PROVEN** (5 fires); implementer trigger **UNUSED** |
| 3. Seed discipline | encoded in validation (D4) + PR checklist rendering | **BUILT** (part of 1b/2a) |
| 4. Pilot (F1.2) | ref/base as per-run input | **CLOSED — superseded.** Run 5 produced a council-APPROVED plan (unanimous 5/5), but the fixloop thread fixed F1.2 by hand 60s before the run; applying the plan would regress it. Item `db066cac` closed `complete`; reason in its spec. Pilot proved the DESIGNER; a fresh target is needed to exercise the IMPLEMENTER. |
| 5. Designer, live | 5 runs, each surfacing a real defect | **PROVEN** — run `8e837814` approved unanimously; 016 revise-loop fix verified from `llm_call_log` (0 `<no value>`, 31KB of real reviewer content) |
| 6. Delta 2 through the council gate | our own platform code reviewed | **ROUND 2 → REVISE 2026-07-21** (`5a65ec4c`, run `14710d52`; abstained 6). tooling_provenance + guidelines flipped to APPROVE (migration 184 landed); the high-severity fail-loud find stays fixed (`9c94cc842`). New/surviving objections: prior_art [HIGH] caught a real absence gap — a generic `loop` action DOES exist (my "none" was wrong, WRONG_CALLS logged); reuse_agent — `githubBranchExists` is a 2nd GitHub path (adapter has no read verb — confirmed); editquality — symbol fix is prompt-not-structural; debug_historian [HIGH] — jsonb-surgery discipline not shown; + cheap process closes. **Round-3 close is now genuine design work, not a formality — owner's call (see NOTES turn 16 shopping list).** |
| Image | deltas 1+2 in production | **LIVE** — v1.0.1144 (pod-verified 2026-07-21: `feature_stage_route`=3, `formatGeneratedGo`=2, `tolerate_truncation`=1); carries the bug-013 gofmt-at-commit-prep and bug-019 council-degrade de-riskers |
| Seeds | three agent defs in clients_db | **APPLIED & VERIFIED** 2026-07-17 (owner-approved in-session); inert until fired |

## Next steps (in order) — the ONE thing left is the implementer's first fire

**The designer half is proven. The implementer half has never executed.**
Everything below serves that single gap.

1. **Pick a fresh pilot target** (RUNBOOK B1). The F1.2 pilot was overtaken by
   a hand-fix, so its approved plan must not be applied. Needs: a small, real
   capability; a known-good shape we can grade against; and — critically —
   a target NO other thread is touching (check `site_work_items` AND
   `git log --since` before choosing).
2. **Owner: write + approve the spec** (`owner_approval` + `code_pointers` in
   spec jsonb; RUNBOOK B2 has the SQL shape).
3. **Fire the designer**, grade the plan, let the council approve it.
4. **Fire the implementer via its ORCHESTRATOR** (`feature-implementer-
   orchestrator`, never the implementer directly) — the first live exercise of
   `feature_stage_route`, the per-stage allowlist, the stage gates and the
   derived test gate.
5. **Review the PR as a human**; decide the merge; walk its checklist.

Optional, before step 4: resubmit delta 2 to the council gate
(`RESUBMIT_CORR=5a65ec4c-686c-40c7-813e-7c7fce03a779`) once the four open
objections are answered — the high-severity one is already fixed.

## Delta-2 council-gate objections — ALL ANSWERED (round 2, `5a65ec4c`, 2026-07-21)

Round 1 (verdict revise, 2026-07-18) raised these; round 2 (run `14710d52`)
answers each. The one HIGH-severity find (bug_historian: three silent-fallback
seams) was already fixed in `9c94cc842` (turn 13) — round-2 sketches show the
committed fail-loud code.

- **editquality — `registry.go` buried:** now its own declared edit (round-2
  edit 2, real hunk at `registry.go:1219`). ✅
- **editquality + guardian — `expected_symbols` false-reject:** fixed AT THE
  SOURCE via `PATCH_feature_designer_018_expected_symbols_scope.sql` (applied
  live, snapshot `ba8f1fcd`) — the design prompt's rule 8 now says name only
  symbols the stage's OWN files introduce, never a cross-file symbol; the
  deterministic `missingExpectedSymbols` gate is untouched. ✅
- **reuse_agent — `site_work_items` sequencing vs a new action:** written
  answer — that mechanism is cross-dispatch QUEUEING (each row its own
  orchestration/claim); `feature_stage_route` carries IN-RUN state in one
  orchestration's collected_data, the `diagnose_route` precedent. Registry
  search: no existing generic stage-advance action. ✅
- **tooling_provenance — travelling PLAN+NOTES for the three shared actions:**
  the cited `subject_type='action'` had NO schema support (CHECK allowed only
  tool/pipeline/experience, zero action rows). Migration
  `184_travelling_action_subjects.sql` (applied live, ledger recorded) adds
  `'action'` and seats a PLAN+NOTES for `diagnose_read_repo_files` /
  `diagnose_prepare_fix_commit` / `diagnose_build_gate`. ✅
- **editquality — E4 "reimplements branch creation":** answered — branch
  CREATION stays on the git-adapter `create_branch` verb; the router's direct
  call is a read-only existence GET (`githubBranchExists`, :375) for the
  stale-branch refusal (no adapter verb exists for it, writes nothing). ✅

If round 2 comes back APPROVED, the delta-2 code (`c19b5d097` + `9c94cc842`)
earns a `Council-Reviewed: 5a65ec4c-686c-40c7-813e-7c7fce03a779` trailer on a
closing commit (PATCH_018 + migration 184 are already committed in `de282bddd`).
If REVISE again, read the new objections and iterate — the trail accumulates
under the same correlation.

## Backlog / later options (explicitly deferred)

- Per-stage council review of diffs (E3 — v1 judges the plan only).
- Dedicated-pod designer with repo read (replaces the code_pointers
  requirement if pointer-curation proves too costly — upgrade path in the
  designer seed header).
- More council seats via concept-register stage 3 (paused there on a
  latency-scaling concern; roster changes flow into the designer's council
  the same way v6→v7 did).
- The COUNCIL GATE (design §2) — a separate thread owns it; both consume the
  same seats, no competing machinery.
- Automatic intake from triage's capability_gap surfacing to a spec TEMPLATE
  (today the owner writes the spec by hand).

## Standing constraints (do not relearn)

- Image FIRST, then seeds; never rebuild an existing tag; ~300s no-dispatch
  window after a chassis pod (re)start.
- Fire implementers ONLY via their orchestrators (read-token spawn gate).
- Delete stale `feat/*` branches before re-firing — the loop refuses them
  loudly (E4), by design.
- No root-level `ai_service` key in any seed (MDL-039 shadowing).
- Commit per task with explicit pathspecs; check `git diff --cached` before
  committing (three sweeps on 2026-07-17 alone).
- Every council/designer run spends credits — owner gives the go per run.

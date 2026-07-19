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
| 6. Delta 2 through the council gate | our own platform code reviewed | **REVISE → fixed.** Found a high-severity fail-loud defect our tests missed (`9c94cc842`); 4 objections still open (see below) |
| Image | deltas 1+2 in production | **LIVE** — v1.0.1132 (concurrent thread's rollout), pod binary verified 2026-07-17 |
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

## Open council-gate objections on delta 2 (from `5a65ec4c`, verdict revise)

- **editquality:** `registry.go` registration was buried in another edit's
  sketch; it should be its own declared edit.
- **editquality + guardian:** `expected_symbols`' verbatim substring check can
  false-reject a correct stage whose symbol lives in an earlier stage's file.
  Self-identified in the submission's own risks, still unmitigated. Honest fix
  = designer-prompt guidance (name only symbols the stage's OWN files
  introduce), not weakening the gate.
- **reuse_agent:** should `site_work_items` sequencing (`parent_item_id`,
  `depends_on`, `batch_id`) carry stage state instead of a new action? Our
  view: no — that is work-item QUEUEING, this is in-run workflow state, and
  `diagnose_route` is the precedent. Owes a written answer, not a dismissal.
- **tooling_provenance:** the three shared actions should carry travelling
  PLAN+NOTES subjects (`subject_type='action'`).

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

# PLAN — feature builder (multi-step capability construction on the fix loop)

*Standing plan for the "fixloop feature builder" workstream. Companions:
`NOTES_running_feature_builder.md` (turn-by-turn record),
`RUNBOOK_feature_builder.md` (the owner's tasks),
`SUMMARY_feature_builder_2026-07-17.md` (read-aloud). Parent design:
`DESIGN_feature_builder_and_council_gate.md` §1. Last updated 2026-07-17.*

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
→ **4-seat council** (editquality, bug-historian, reuse-agent, guardian w/
hard veto — the fix-proposer's live v7 roster) + the proven deterministic
router (approve / revise+checks / reframe / escalate)
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
| 1c. feature-designer seed | `0NN_feature_designer.sql` | **DRAFTED** (5-seat v8 council) — awaiting owner apply, after image |
| 2a. Stage-loop machinery | `feature_stage_route_action.go` + seams (read `ref_field`, prepare branch/message/symbols, gate `test_packages_field`, spawn gate) | **BUILT** — commit `c19b5d097`, tests green |
| 2b. feature-implementer seeds | `0NN_feature_implementer.sql` + `_orchestrator.sql` | **DRAFTED** — awaiting owner apply, after image |
| 2c. Triggers | `0NN_TRIGGER_feature_designer_v1.sh`, `0NN_TRIGGER_feature_implementer_v1.sh` | **DRAFTED** (092 envelope) |
| 3. Seed discipline | encoded in validation (D4) + PR checklist rendering | **BUILT** (part of 1b/2a) |
| 4. Pilot (F1.2) | ref/base as per-run input — the loop's own sibling gotcha | spec CREATED (`db066cac`, needs_human_review) — **awaiting owner approval + fire go** (RUNBOOK A4–A5) |
| Image | deltas 1+2 in production | **LIVE** — v1.0.1132 (concurrent thread's rollout), pod binary verified 2026-07-17 |
| Seeds | three agent defs in clients_db | **APPLIED & VERIFIED** 2026-07-17 (owner-approved in-session); inert until fired |

## Next steps (in order)

1. Owner: RUNBOOK A1–A3 (image ≥ `c19b5d097` → verify pod → apply 3 seeds).
2. Owner: A4 — create + approve the F1.2 pilot spec (pointers in the RUNBOOK).
3. Fire the designer on it; grade its staged plan against the hand-written
   reference instance in `SCHEMA_staged_plan_v1.md` §6 before approving.
4. Fire the implementer on the approved correlation; owner reviews the PR,
   merges, walks the checklist. First feature = the loop fixing the fix
   loop's standing stale-ref gotcha.

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

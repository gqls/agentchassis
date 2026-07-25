# PLAN — feature builder (multi-step capability construction on the fix loop)

*Standing plan for the "fixloop feature builder" workstream. Companions:
`NOTES_running_feature_builder.md` (turn-by-turn record),
`RUNBOOK_feature_builder.md` (the owner's tasks),
`SUMMARY_feature_builder_2026-07-25.md` (read-aloud, current). Cold-start:
`HANDOFF_2026-07-25_feature_builder_thread.md`. Parent design:
`DESIGN_feature_builder_and_council_gate.md` §1. Last updated 2026-07-25.*

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
| 2b. feature-implementer seeds | `0NN_feature_implementer.sql` + `_orchestrator.sql` | **APPLIED, LIVE & PROVEN** — first complete run 2026-07-25 (orch `af286d2c`) |
| 2c. Triggers | `0NN_TRIGGER_feature_designer_v1.sh`, `0NN_TRIGGER_feature_implementer_v1.sh` | both **PROVEN**; the B4 run drove the implementer via a patient-watcher wrapper (`fire_impl8.sh`, gauntlet thread) over the same orchestrator target |
| 3. Seed discipline | encoded in validation (D4) + PR checklist rendering | **BUILT** (part of 1b/2a) |
| 4. Pilot (F1.2) | ref/base as per-run input | **CLOSED — superseded.** Run 5 produced a council-APPROVED plan (unanimous 5/5), but the fixloop thread fixed F1.2 by hand 60s before the run; applying the plan would regress it. Item `db066cac` closed `complete`; reason in its spec. Pilot proved the DESIGNER; a fresh target is needed to exercise the IMPLEMENTER. |
| 5. Designer, live | 5 runs, each surfacing a real defect | **PROVEN** — run `8e837814` approved unanimously; 016 revise-loop fix verified from `llm_call_log` (0 `<no value>`, 31KB of real reviewer content) |
| 6. Delta 2 through the council gate | our own platform code reviewed | **ROUND 2 → REVISE 2026-07-21** (`5a65ec4c`, run `14710d52`; abstained 6). tooling_provenance + guidelines flipped to APPROVE (migration 184 landed); the high-severity fail-loud find stays fixed (`9c94cc842`). New/surviving objections: prior_art [HIGH] caught a real absence gap — a generic `loop` action DOES exist (my "none" was wrong, WRONG_CALLS logged); reuse_agent — `githubBranchExists` is a 2nd GitHub path (adapter has no read verb — confirmed); editquality — symbol fix is prompt-not-structural; debug_historian [HIGH] — jsonb-surgery discipline not shown; + cheap process closes. **Round-3 close is now genuine design work, not a formality — owner's call (see NOTES turn 16 shopping list).** |
| 7. **B4 — the implementer's first complete run** | plan `c379f7b7` (tools-api), orch `af286d2c` | **CLOSED 2026-07-25.** 6 gated stages → test_gate PASS → **PR #3 MERGED into `main` 09:19:16Z** (`c02d56b9a`), 18 files +880/−0, one commit per stage; awaits 8/8 processed, 0 expired. Driven by the `gauntlet_dead_cta` thread on its own target. Shakeout yield: bugs 065/067 CLOSED, **066 OPEN**, 071 fix live+tick-proven, migrations 199/200/201/202. |
| Image | deltas 1+2 in production | **LIVE — v1.0.1158** (pod-verified 2026-07-25, `agent-chassis-54fff9df8b-966hj`: `feature_stage_route`=3, `formatGeneratedGo`=2). *Was v1.0.1144 at the 07-21 update.* |
| Seeds | three agent defs in clients_db | **APPLIED, VERIFIED & EXERCISED**; all three rows active at `image_tag` v1.0.1158 (2026-07-25); designer council still 5 seats, reviewer seats on sonnet-5 (`ee31c3632`) |

## Next steps — REWRITTEN 2026-07-25, because the old list is done

> **CORRECTED 2026-07-25:** this section used to open *"The designer half is
> proven. The implementer half has never executed."* Both halves are now proven.
> The whole 5-step list below it (pick a target → spec → design → fire → merge)
> was walked end to end on plan `c379f7b7` and ended in a merged PR. The list is
> preserved in git history; what follows replaces it.

1. **A SECOND build, on a target we choose.** One success is not a capability:
   the first took 6 designer rounds and 8 implementer fires, and most failures
   were environmental (071, 066) rather than architectural. A second run is the
   only thing that tells us whether the shakeout fixes generalised. RUNBOOK B1's
   selection criteria still apply — above all, a target no other thread is
   touching.
2. **Decide the delta-2 council trail** (below): spend a round 3 on the two real
   design questions, or accept advisory-REVISE and record it closed-unapproved.
   The risk argument for round 3 is weaker now that the reviewed code has built
   and shipped a real feature; it is a review-coverage decision, not a safety one.
3. **Hand `bugs_open/066` to whoever owns deploys.** A chassis roll does not
   reach spawned agent pods (they pin `agent_definitions.image_tag`), and the
   symptom is indistinguishable from an agent defect — it cost rounds 6–7. The
   sync `UPDATE` exists only in `deploy-100-bootstrap-agents` (`makefile:518`),
   not in `deploy-agents`. Census before every fire.
4. **Not ours:** building/deploying the merged tools-api, migration 198 (→ the
   ISLAND DB, not `clients_db`), smoke-testing tools.apis.uk. That is the
   island/gauntlet threads' work — `scripts/who-owns.py` before touching it.

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

**Round 2 came back REVISE**, so no `Council-Reviewed` trailer was earned or
added (it is earned by APPROVED only; a false trailer is a permanent lie the 098
report buckets as MISMATCH). PATCH_018 + migration 184 stand on their own merits
and are committed in `de282bddd` regardless.

**Verified 2026-07-25: nothing has run on `5a65ec4c` since.** Zero orchestration
rows carry that `fix_correlation_id` after the round-2 run. The trail is exactly
where 07-21 left it. The round-3 shopping list lives in NOTES turn 16 and in
`HANDOFF_2026-07-25_feature_builder_thread.md`; the two substantive items are the
`LoopAction` reuse answer and the `githubBranchExists` read-verb decision.

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

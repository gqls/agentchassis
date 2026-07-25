# HANDOFF — resume the FEATURE BUILDER thread

> **SUPERSEDED 2026-07-25 by `HANDOFF_2026-07-25_feature_builder_thread.md`.**
> Read that one instead. This file is kept as a point-in-time record and is now
> **wrong on its central claim**: it says the implementer "has still never
> executed once". It has — plan `c379f7b7`, run `af286d2c`, six gated stages,
> **PR #3 merged into `main` at 09:19:16Z on 2026-07-25**. RUNBOOK B4 is CLOSED.
> The delta-2 council trail described below is still accurate (round 2 = REVISE,
> nothing has run since).

**Filed 2026-07-21** from the "fixloop feature builder 2" session. Cold-start for
a NEW chat. **Supersedes `HANDOFF_2026-07-19_feature_builder_thread.md`** (still
accurate on architecture; this one carries the delta-2 council close and the
v1.0.1144 image state).

## Read these first, in this order

1. This file, top to bottom.
2. `SUMMARY_feature_builder_2026-07-19.md` — plain-language, 3 minutes (still current).
3. `PLAN_feature_builder.md` — architecture as built, status table, the delta-2
   objection ledger (all answered / round-2 status).
4. `NOTES_running_feature_builder.md` **turns 15–16** — this session's work and
   the round-2 verdict analysis (the round-3 shopping list is there).
5. `README_feature_builder_where_we_are.md` — owner's plain-prose log (2026-07-21
   entry at the bottom).

Repo rules that override everything: `/CLAUDE.md` — **read it fresh from disk**,
it changes constantly. Parent design: `DESIGN_feature_builder_and_council_gate.md`
§1; `DESIGN_stage_loop_delta2.md`; schema `SCHEMA_staged_plan_v1.md`.

## State in one paragraph

The feature builder is BUILT and LIVE on **chassis v1.0.1144** (pod-verified
2026-07-21). Its DESIGNING half is proven (5 runs, last one unanimous 5-seat
approval). Its BUILDING half — `feature-implementer` + the stage loop — is
seeded, live, and **has still never executed once.** That is the entire remaining
milestone (RUNBOOK **B4**). This session did NOT fire it; it closed out the
delta-2 council-gate trail instead (owner's call) and armed the pilot's ground.

## What this session (2026-07-21) did

1. **Verified v1.0.1144 in the running pod** (`agent-chassis-59c675c4f-pxr9f`):
   `feature_stage_route`=3, `formatGeneratedGo`=2 (**bug 013 gofmt-at-commit-prep
   now LIVE** — it killed the fix loop's first fire), `tolerate_truncation`=1
   (**bug 019 council-degrade now LIVE**). Both de-riskers the implementer's first
   fire needs are in the binary. **B4 is now on safe ground.**
2. **Answered all four open delta-2 council-gate objections; two with real,
   committed, live changes:**
   - **PATCH_feature_designer_018_expected_symbols_scope.sql** — tightens the
     designer's rule 8 so it names only symbols the stage's OWN files introduce
     (the expected_symbols false-reject fix, at the source). Applied live, snapshot
     `ba8f1fcd`. Surgical `replace()` on one jsonb leaf.
   - **migration `184_travelling_action_subjects.sql`** — adds `'action'` to the
     `doc_plans`/`doc_notes` subject_type CHECK (the convention tooling_provenance
     cited had NO schema support) and seats PLAN+NOTES for the three shared actions.
     Applied live, ledger recorded. Both committed in **`de282bddd`**.
3. **Resubmitted delta 2** (`submission_delta2_round2.json`, 7 edits) via 097 with
   `RESUBMIT_CORR=5a65ec4c`. Run `14710d52`. **Verdict: REVISE** — see next section.

## THE LIVE STATE OF THE DELTA-2 TRAIL (corr `5a65ec4c`)

Round 2 = **REVISE** (abstained 6). It advanced and earned its credits:
- **Flipped to APPROVE:** tooling_provenance (migration 184 satisfied it),
  guidelines, diagnosis_guardian, constitution, mission.
- **The high-severity round-1 find stays fixed** (`9c94cc842`, fail-loud seams).
- **NO `Council-Reviewed` trailer** was added — REVISE is not APPROVED, and the
  trailer is earned by APPROVED only (strict; a false trailer is a permanent lie
  the 098 report buckets as MISMATCH).

**The gate caught a real thing.** prior_art_librarian [HIGH]: my submission
asserted "no existing loop-controller action: none" — FALSE, the registry has a
generic `loop`/`loop_complete`/`conditional_route` (registry.go:47/53/73). I never
opened `LoopAction`. Logged in `WRONG_CALLS.md`. The code is not necessarily wrong
(loop-vs-bespoke is a genuine design call), but the *claim* was, and I reproduced
the workstream's own documented failure mode inside a submission to the seat that
exists to catch it.

**A round-3 close is now genuine design work, not a formality.** Shopping list
(full detail in NOTES turn 16):
1. Read `LoopAction` (loop_actions.go); give the real reuse answer — could the
   stage loop have been a `loop` config, or does emit-stage-as-single-plan +
   branch/ref threading genuinely need a bespoke action? Say which, with the read.
2. Decide the `githubBranchExists` question: add a `get_branch`/`branch_exists`
   READ verb to the git-adapter (adapter.go:337-345 — verbs are commit/create_repo/
   delete_repo/create_branch/create_pull_request, no read verb) and route through
   it, OR justify the direct read-only GET explicitly (it writes nothing, so the
   write cage is intact). This is reuse_agent's objection and it is reasonable.
3. Cheap closes: add migration 184 as a discrete edit with its hunk; attach the
   registry + git-adapter-verb search OUTPUT (not an assertion); name the owning
   pipeline (feature-builder/feature-implementer) on the config_change edit; show
   a pre-state needle-gate in the PATCH_018 file (assert rule-8 text occurs exactly
   once before replace()) for debug_historian.

**OR** — the gate is advisory. A defensible owner call is: the substantive defect
(high-severity fail-loud) was caught and fixed, two objections are structurally
answered, the rest are design flags on committed/live code — accept advisory-REVISE
and **go straight to B4** (the implementer's first fire), which will teach more than
a round 3. That choice is the owner's; do not spend another council run without it.

## What is live right now

| Thing | Where | State |
|---|---|---|
| `feature-designer` | agent_definitions | LIVE; 5 runs, 1 unanimous approval; **rule 8 tightened by PATCH_018 (live)** |
| `feature-implementer` | agent_definitions | LIVE, **never fired** ← the milestone |
| `feature-implementer-orchestrator` | agent_definitions | LIVE, never fired |
| stage loop + seams | `feature_stage_route_action.go` + 3 shared actions | LIVE on v1.0.1144; fail-loud seams (`9c94cc842`); **never executed** |
| travelling action subjects | doc_plans/doc_notes `subject_type='action'` | LIVE (migration 184) |
| Triggers | `0NN_TRIGGER_feature_{designer,implementer}_v1.sh` | designer proven; implementer UNUSED |

## Do NOT do these

- Do **not** put a `Council-Reviewed: 5a65ec4c` trailer on anything — the verdict
  is REVISE, not APPROVED. It is earned only by an APPROVED round.
- Do **not** apply the approved plan from run `8e837814` — the old F1.2 pilot was
  hand-fixed; applying it regresses a working fix. Work item `db066cac` is closed.
- Do **not** fire `feature-implementer` directly — only via
  `feature-implementer-orchestrator` (092-style trigger), or it dies with no read
  token (the fix loop's proven 2026-07-13 failure).
- Do **not** re-apply `0NN_feature_designer.sql` wholesale — it would revert
  PATCH_018 and 016/017. Diff against the LIVE row first; edits are surgical.
- Do **not** spend a designer/council/implementer run without the owner's per-run go.

## RUNBOOK pointer

`RUNBOOK_feature_builder.md` — A1–A5 done; **B1–B4** are the remaining human tasks.
B1 (pick a fresh pilot target) was scouted this session — strongest candidate is
the digest's missing `needs_human_review` section (294 orphaned items, bug 023's
delivery gap; our own file, no collisions; a natural 2-stage Go-edit+new-file
plan). NOTES turn 15 has the vetting. B4 (the first implementer fire) is the
milestone and is now de-risked by v1.0.1144.

## Coordination notes

Many threads share this repo/cluster. This session alone: a new chassis image
rolled (v1.0.1139→1144), the CLAUDE.md council-gate section was rewritten (latency
now documented ~30 min, though this run dispatched instantly), and MEMORY.md gained
a new brochure-component workstream. Commit narrowly with explicit pathspecs
(done — `de282bddd`, `467a149b8`, `5112fbc57`, `9a8c2c9b6` this session), re-run
`git status` before acting on it, check the queue before dispatching.

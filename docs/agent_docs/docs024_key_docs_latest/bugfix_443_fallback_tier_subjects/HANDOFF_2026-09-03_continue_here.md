# HANDOFF 2026-09-03 — 443 fix is LIVE in the binary; Stage A is go; what stands between here and closing the bug

**COLD-START for the `bugfix_443_fallback_tier_subjects` lane** (session "bugs_open/443").
Bug: `bugs_open/443_HANDOFF_2026-09-02_a_page_built_from_the_fallback_tier_cannot_carry_per_section_subjects_so_repeated_component_types_write_the_same_section.md`
(§8 = the fix account). Plan/decisions: `PLAN_2026-09-02_fallback_tier_subjects.md` (D1–D8).
Commands+gotchas: `RUNBOOK_fallback_tier_subjects.md`. Technical log: `NOTES_…` (newest at
bottom — read the 09-03 entry). Owner prose: `README_where_we_are.md`.

## State, verified at artefacts (2026-09-03, just after the roll)

| thing | state | evidence |
|---|---|---|
| fix commit `dbb218a41` | **IN the running binary** | pod `agent-chassis-75b987cbd7-mqrnj`: `subjects_attached`=1, `REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT`=1, present-control 3, absent-control 0 (RUNBOOK §deploy-verification — use THOSE literals, not `section_subjects`) |
| mig 717 (pages.section_subjects/facts) | applied + ledger-recorded 09-02 | information_schema check in RUNBOOK |
| 639 wiring (plan_sections config) | live | live `agent_definitions` query in RUNBOOK — NOT the seed file, NOT schema_migrations (LANDMINES: `_HOLD` filename/ledger trap) |
| 640 (planner rule 17) | applied 09-02 | APPLIED line in seed header |
| **641 (writer prompt v5)** | **VOIDED as written — owner returned gate 2 as a REDRAFT** | DO-NOT-APPLY stamp in the seed header itself (apis.uk `c9c9b75ec`). Redraft pipeline: owner picks framing → finetuning test-renders (candidates in `finetuning_uk_service/DRAFT_2026-09-02_641_positive_prompt_candidates.md`) → apis.uk writes final SQL → owner reads exact final words |
| build detector | **LIVE and firing** | 7 rows in ~2h — see "new finding" below |
| council | APPROVED r1, `Council-Reviewed: b7c59309-…` on the commit | verdict + 11 advisory answers in NOTES 09-02 |
| RFC_063 | **DECIDED: option B** (converge the 6 plan-less sites; hand-insert permitted, closed backfill). Imagery seeding CLAIMED by 114 lane; **composition half + reconciler-skip measurement UNASSIGNED** | RFC appendix `01a3b96ac`, status header current |

**Key dependency fact:** subjects now travel end-to-end to `sections_ready[].subject` for
tier-3 pages, but the WRITER is still v4 — blind to subjects on EVERY tier until the
redrafted 641 applies. So served h2s still repeat everywhere. That is expected, not failure.

## What is LEFT before this lane can close (the bar is fixed AND LIVE — served pages no longer repeating)

1. **Stage A (unblocked now, days not weeks).** finetuning lane backfills subjects for
   `your-own-model` (+ their other 3) — THEIRS, agreed twice, ping delivered 09-03 with the
   ~300s post-restart dispatch caution. Then THIS lane verifies: `sections_ready[].subject`
   populated in the orchestration row for the three generic-text-block slots, and the
   detector QUIET on that page. The demand control (a subjectless repeat still firing) is
   already satisfied by the live rows below. If finetuning stalls, the backfill query
   template is in the RUNBOOK and the briefs are in each page's `needs_content_page` item
   (`spec.suggestion`).
2. **Redrafted 641 lands** — NOT ours (apis.uk's file, owner's framing pick pending,
   finetuning test-renders incl. the untested sibling-range arm). We only wait. Nudge the
   apis.uk lane if a week passes.
3. **Stage B (after 641):** rebuild the canary; assert served h2s DISTINCT
   (curl + invented-URL control); tier-1 control page passes same assertion; **SAVE the
   before/after served-HTML pair and point to it from 443's close-out — OWED to
   copy_quality_two_stage** (they cancelled their proposal on this page for us; NOTES 09-02).
4. **The other 7 damaged pages** (gaswholesalers 4, ai-agent-orchestration 3 — no owning
   lanes known): after Stage B proves the mechanism, write their subjects (briefs exist per
   page) + rebuild, or hand to the sites' next lanes explicitly. finetuning's remaining 3 are
   theirs. Alternative: if RFC_063's composition half converges these sites first, subjects
   can go into plan rows instead — do NOT do both.
5. **Close-out:** re-curl all 11 pages distinct (per-domain invented-URL controls), update
   §8 + register PBP-051 status, move 443 to `bugs_closed/` (one sequence, never renumber),
   drop the workstream memory line's OPEN state.

**NOT required for close:** RFC_063 execution (separate estate work; our fix is correct under
either outcome — but see the coordination note below), the new-finding cohort below (separate
filing decision).

## ⚠ NEW FINDING from the first 2h of detector rows — read before assuming the cohort is the 11

7 rows / 4 pages / 3 sites, **none in the plan-less six**: leopardessconsulting.co.uk
(`case-study-automated-intelligence-pipeline` — rebuilt 3× in 2h, itself unexplained),
seotools.co.uk (2 pages), vetcomparison.uk (`how-it-works`). Plan-CARRYING sites are minting
subjectless repeats on fresh builds. `[UNVERIFIED]` whether that is: pages from pre-640 plans
(likely — rule 17 only binds NEW plans), tier-2/3 pages on planned sites, or rule 17 failing.
Next session: (a) curl the 4 pages with controls before calling them damage; (b) check
whether the PLANNER-side code (`SUBJECT_MISSING_ON_REPEATED_COMPONENT`) fired for the same
pages — if plans postdate 640 and it did NOT, that is an apis.uk/640 defect to hand them;
(c) decide whether this cohort goes into 443's account or a new bug file (grep both bug dirs
first). The read-back query (CORRECTED — `occurred_at`, not `created_at`) is in the RUNBOOK.

## Coordination state (who owes whom)

- **finetuning** (new session under same name since 09-03): owes backfill+rebuild, reports
  Stage A as Stage A; owed a re-ping only if the detector shows their rebuild happened and
  they haven't reported.
- **apis.uk**: own 641 redraft + the sibling-range pre-apply falsifier; told everything.
- **114 lane**: imagery seeding claimed (their `448671d18`), sequenced after composition;
  nothing owed either way now.
- **copy_quality_two_stage**: owed the Stage B HTML pair (item 3 above).
- **444 lane**: nothing owed; their hunk shipped in our commit (declared passenger).
- **Owner**: reads `README_where_we_are.md`; owes nothing to us except the 641 framing pick
  (via finetuning/apis.uk).

## Landmines this lane wrote or hit (grep LANDMINES.md for the entries)

`pages.sections` rewrite silently DISARMS the sibling arrays · `_HOLD` filename/ledger is not
the apply state (ask the artefact) · probe `subjects_attached`/`facts_attached` for THIS fix,
never `section_subjects`/`without_subject` · provenance stamp scrolls (capability probe wins;
guard the empty-sha arm) · `agent_error_log` timestamps are `occurred_at`.

# HANDOFF 2026-08-26 — continue here

**Lane:** `copy_quality_two_stage`. **Supersedes `HANDOFF_2026-08-25_continue_here.md`** (kept —
it holds the day the owner's escalations arrived and the audit began; read it for history, not
state).

> ## ▶ START HERE, IN THIS ORDER
> 1. **`OWNER_RULINGS_2026-08-25_six_decisions_on_the_copy_machinery.md`** — the owner's rulings,
>    verbatim, with the execution record. Everything this lane does now flows from them.
> 2. **`REFRESH_2026-08-25_deep_context_the_accumulated_copy_discussion.md`** — the assembled copy
>    context the owner ordered before any fixes. The lane's ground truth; every claim sourced.
> 3. This file's **"Next work"**. Everything else is reference.
>
> **One-line state:** the owner escalated twice (2026-08-25), ordered a context refresh + a
> fleet prompt audit, then ruled on six decisions; **four fleet migrations (627–630) shipped,
> were verified at the artefact, and are council-APPROVED**; the first big causal experiment
> reversed our own headline (the PLANNER's page premise, not the writer's clauses, was the
> primary fault); and a **pre-registered nine-page canary on finetuning.uk is IN MOTION** —
> scoring it is this lane's top obligation.

## What is TRUE as of 2026-08-26 ~16:30Z

**The owner's rulings, all addressed** (rulings doc has verbatim + status):
1. Planner premise — researched, then SHIPPED as `630` (owner: "yes please ship"); the TESTED
   second draft (draft 1 failed 1 of 2 planner replays; the hard format clause held 3 of 3 —
   **register rules bend, format rules hold**).
2. Writer's July "say this instead" substitutes — DELETED (`627`), every ban kept.
3. Tool-fallibility mandate — KEPT (it survives inside 627 by assertion).
4. Uncertainty split by stakes + best-in-class — **PLANNED, not built**:
   `PLAN_2026-08-25_best_in_class_propagation.md` (carrier row + per-site benchmark + research
   phases + stakes split; the owner's Go question answered in §2: mechanism in Go, words in one
   DB row — the CQ-022 shape). **Build awaits his go.**
5. House voice — form-only rewrite SHIPPED (`628`): 17 demonstrations → 0, no rule's meaning
   changed; v2 restorable from `migration_backups`.
6. Unfillable social-proof slots — planner stops planning AND stops demonstrating them (`629`;
   its worked example literally contained a testimonials entry).

**Council:** wave 1 (`627–629`) APPROVED on `6a0f8b99`; `630` APPROVED all-reviewers on
`5f084feb` after a trail worth knowing: REVISE (post-hoc gating objection — answered by CITING
the owner ruling of 2026-07-29, "review here is after the fact by design"; two real catches —
the rollback then TESTED live with byte-equality restore, and the `{{else}}`-arms landmine's
brace-balance check run) → a run STRANDED by another lane's `orchestration_states` restructure
(terminal state FAILED at `review_guidelines`) → redispatch → clean approval.

**Verified at the artefact, not the row** `[MEASURED 2026-08-26 15:12 call]`: the live rendered
writer prompt scans **36 negation demonstrations vs the 63–65 baseline**; "plainly" 14→4,
"honest" 10→5; the deleted clauses' signatures at 0. The residual is the brief + guidance
layers — the briefs are the dominant remaining teacher (12–31 demos/site, 25 of 30 sites ≥10).

**The experiments** (all pre-registered, in `AUDIT_prompts/`):
- `EXPERIMENT_2026-08-25_about_section_replay.md` — the owner's "test first": 12 section replays
  + 6 planner replays. Headline: the about-page premise was PLANNED ("About Home Garden —
  Editorial Approach and What We Will Not Do"); neutralising the title collapsed the methodology
  score 10→1 while removing the writer clauses alone only reached 6. Premise ← plan; register ←
  demonstrations. Two faults, two fixes.
- **`CANARY_2026-08-26_finetuning_nine_page_rebuild.md` — THE OPEN ONE.** finetuning.uk
  de-demonstrated its own brief (their migrations 646/647: `rather than` 7→0, `not just` 4→0)
  and measured the residual guidance layer (exactly 1 `rather than` across 27 components — the
  fleet-shared anti-fabrication rule, left in place by agreement). **Seven of the nine pages now
  have a demonstration stack at ZERO in every layer** — any `rather than` there implicates the
  MODEL PRIOR itself (P2a); the two `hero-tool` pages carry a named ceiling of 1 (P2b).
  Honesty-beat predicted ~0 (P1); scaffold/cadence NOT expected to move (P4); **read-aloud
  primary, battery secondary; a clean battery with a failed read is a FAIL** (P5). As of ~16:00
  the 9 imagery items are re-filed (triaged) with 3 page_rerenders — rebuilds not yet at the
  writer. **Scoring is same-day on their ping, against the pre-registration, never after-the-fact
  criteria.** Baseline: `finetuning_uk_service/baselines/2026-08-26_pre_hero_rebuild/`.

**The audit** (owner point 3, a workstream): census + phase 1 done
(`PLAN_2026-08-25_prompt_audit.md`, `AUDIT_prompts/PHASE1_*` — 599 strings, 7 populations, incl.
RENDERED prompts as the per-call truth); phase 2 verdicts 1 of N done (writer template TEACHES-AI
via its "instead" clauses — now deleted; house voice was neutral-in-content/teaches-in-form — now
rewritten). Scanner: `audit_prompt_demonstrations.py` (CQ-032).

**Older still-open state inherited from 08-25:** stage 2 (copy-editor) is live and dispatchable
(CQ-030, bounded by `pendingCopyEditForPage`, council-approved `754dcffd`) — **no dispatched
`needs_copy_edit` run has happened yet**; proposal `b0dea48e` re-graded STALE (approve edit 3 at
most); `8003c51a` fails the gate on real structural loss; the first-dispatched-run verification
(handoff-08-25 item 2) still stands.

## ⚠ Things needing the OWNER

1. **Best-in-class build go** (ruling 4's plan — step 1 is a small Go injection change needing a
   roll).
2. **Retitling the ~6 existing about pages** with self-limiting titles (630 governs future plans
   only) — per-lane work he may want ordered.
3. **Writer rules 16–17 removal** (testimonial-filler instructions — dead for new plans, still
   live for existing slots; removing them changes existing-slot rebuilds to empty-string).
4. The parked `b0dea48e` decision (approve edit 3 at most).

## Next work, in order

1. **SCORE THE CANARY when finetuning pings** — same-day, against the pre-registration, both
   instruments. If P2a fails (rather_than on a fully-cleared page), that is a NEW FINDING about
   the ceiling of de-demonstration — file it, don't bury it.
2. **Phase 2 verdicts 2 of N: the BRIEFS** — now confirmed the dominant teacher. Top five by
   phase-1 table (vetcomparison 31, loancalculator 31, mortgagecalculator 28 — which also has the
   **v2 voice FOSSIL**, CONTRIB filed in their lane 08-26) + homegarden's. The fleet brief sweep
   (ruling 4 §5) folds in: stakes split + voice-fossil pass + de-demonstration, using
   finetuning's proven method (**source keys AND formatted together** — formatted is DERIVED and
   regenerates; and **verify blocks must assert the FULL battery**, not the headline needle —
   their 646/647 lesson, adopted).
3. **Ruling-4 execution** once the owner says go (plan §6 order).
4. The tone-route's first dispatched run verification (from 08-25) when a `tone` finding fires.
5. Rules 16–17 follow-up migration (owner decision 3 above).

## ⚠ Fresh traps this week (full accounts in NOTES 08-26 + LANDMINES)

- **Watchers poll by PAYLOAD correlation, never a printed run id** — `orchestration_states` was
  restructured mid-flight (`id` → `orchestration_id`); payload scans are currently SLOW
  (possible lost indexes). Foreground-test any watcher query once before arming.
- **Council dispatches can DROP silently** (kcat shape): a missing orchestration row after the
  ~29-min latency WITH a later submission consumed = dropped; resubmit with `RESUBMIT_CORR`.
- **Migration numbers COLLIDE on this tree** — two `630_*.sql` exist (ours + another session's);
  finetuning took 646/647 hours later. Re-check `ls` at write time; resolve by FILENAME.
- **Four agent types carry TWO active `agent_definitions` rows** (content-creator,
  content-creator-contact, chief-strategist, site-component-architect) — every prompt migration
  asserts `count(*)=1` in its DO block (LANDMINES entry, 08-26).
- **A carrier-row rewrite cannot reach pasted COPIES** — mortgagecalculator's brief carries v2
  voice text (1 of 31 sites); fossil pass is in the sweep.
- The GTM re-render wave (bugs_open/397) redeploys homegarden pages WITHOUT the writer — do not
  read those redeploys as copy canaries; only pages BUILT after 2026-08-26 00:22 test 627/628.

## The five living docs + tooling

- **PLAN×3**: `PLAN_2026-08-12_two_stage_copy.md` (original lane) ·
  `PLAN_2026-08-25_prompt_audit.md` (the audit) · `PLAN_2026-08-25_best_in_class_propagation.md`
  (ruling 4). **NOTES** — append-only log, 08-26 tail is the freshest evidence.
  **README_where_we_are** — the owner's plain-prose log, current through 08-26. **SUMMARY
  series** — newest 08-23; a new one is DUE once the canary scores (the five headings would all
  change). **This HANDOFF.**
- **Tooling:** `audit_prompt_demonstrations.py` (CQ-032) · `gate_stage2_edit.py` (page-scope
  declared-links fix 08-25) · `audit_writer_brief.py` (CQ-025) · `count_negation_tells.py` ·
  `scripts/fire-copy-editor.sh` / `fire-section-edit.sh`.
- **Platform code owned:** the tone arm of `write_audit_findings_action.go` +
  `pendingCopyEditForPage` + tests; `content_direction` derivation; the 627/630 migration
  surfaces. **Migrations this week:** 627, 628, 629, 630 (each `_ROLLBACK`ed; 630's rollback
  TESTED live).

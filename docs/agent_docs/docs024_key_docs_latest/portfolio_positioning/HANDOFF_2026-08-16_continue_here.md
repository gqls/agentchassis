# HANDOFF — Phase B COMPLETE (B3e + B3f live); Phase C pilot is next — 2026-08-16, continue here

Supersedes `HANDOFF_2026-08-15b_continue_here.md` (correct on its own history; this file
carries everything a fresh chat needs). Owner rulings in force unchanged: P9 six decisions,
pilot = remortgagecalculator.uk (M4), build order M→B→I, B8/B9/I10 HOLD, bug 270 hands-off,
copy-voice work lives in session "copy quality two stage" `79d969f9-…`.

## 1. What is DONE and PROVEN

**Phase B is finished.** All seven places a directory kind lives (DIR-001) are filled for
all six kinds. What remains for the finance kinds is a SITE, not a mechanism.

**B3d (432) — council APPROVED at round 2** (corr `47785bb5`, 2026-08-15 22:19Z, all 11
non-abstaining seats, 6 abstained). Commit `bbb0cfa89` carries `Council-Submitted:` and is
credited automatically by 098. Advisories dispositioned in commit `6f9374824`:
- ROLLBACK hardened (guardian, REAL): the un-run `432_…_ROLLBACK.sql` pinned only the edge
  it entered by. It now pins all six edges it overwrites or could orphan.
- `RFC_031_hand_spliced_enrichment_steps_want_an_ordered_list.md` filed (architecture seat's
  ask). **Trigger: the THIRD deterministic `content_features` recommender must propose the
  shared ordered list, not hand-splice a third time.** Filing it surfaced undecided drift:
  `evaluate_news_feed` is NOT spliced into domain-research-classifier while
  `evaluate_directory_features` now is.
- Measured, not asserted: output_field collision clean; `improvement-sweep` still
  `enabled=f` (last fired 2026-08-14 16:34Z) so the loop consumer stays wired-but-undriven.

**B3e (433) — DONE, LIVE, VERIFIED.** `sql_for_agents/433_planner_directory_rule_b3e.sql`
(+ surgical-inverse ROLLBACK), applied 2026-08-15 22:24Z, commit `7aff17b21`.
build-site-planner's `plan_site` prompt gains a `Directory rule:` paragraph (after the News
listing rule) + RULES entry 18: for each of the six `content_features` directory keys with
`recommended=true`, the section component on the homepage; when `separate_page=true`, ONE
dedicated page with the EXACT `directoryCheckProfiles` name/page_type, composed
hero → `<x>-listing` → call-to-action, header+footer nav, `facts: []` on directory sections.
Mapping verbatim from `check_directory.go:72-149`; composition matches the three live pages
on ai-agent-orchestration.com AND `MissingDirectoryPageCheck`'s own suggestion text.

**B3f (434) — DONE, LIVE.** `sql_for_agents/434_enable_finance_directory_and_structural_checks_b3f.sql`
(+ ROLLBACK), applied 2026-08-15 22:38Z, commit `a1b92f609`. Checks array **32 → 43**.
Enabled: the three FINANCE directory pairs + `dead_internal_link_live`,
`canonical_mismatch`, `structured_data_invalid`, `head_essentials_missing`,
`sitemap_entry_dead_live`.
> **Correction to the last handoff's framing:** it said "the 6 directory checks". The
> model/adoption/protocol pairs were ALREADY enabled by 194/215 — measured in the live array
> before writing the migration. The six actually missing were the FINANCE pairs. Same count,
> different six.

## 2. Fresh findings a new session must know

- **The council queue ran ~17 HOURS on 2026-08-15/16**, not the documented ~30 minutes.
  433/434 were submitted ~22:45Z on 08-15; their `fix_plan` artifacts were written
  **16:08Z and 16:10Z on 08-16**. CLAUDE.md's "budget ~30 minutes" is a floor under normal
  load, not a ceiling. **Do not retry on a missing verdict** — find the run by payload:
  `SELECT kind, created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE correlation_id='<CORR>' ORDER BY created_at;`
  A `fix_plan` row with no `council_report` row means RUNNING, not dropped.
- **Fourth council submission-schema gotcha, and it cost a rejected submit:** a sketch whose
  every non-blank line is a comment is **refused client-side** — *"a fix plan proposes
  changes, not observations"*. Sketch the REAL statements; commentary goes in
  `.rationale`/`.grounded_in`. (Running list: `.plan.summary` required; operations are
  modify|add|remove|config_change ('create' refused); `.plan.risks` is a STRING; `FORCE=1`
  needed for config migrations under `docs/`.)
- **A checks-array enable fails SILENTLY against a binary that lacks the check** — the
  registry lookup skips an unregistered NAME rather than erroring. Probe the binary with
  BOTH controls in the same breath (done for 434, pod `agent-chassis-584b6fcf-9mtqd`):
  `kubectl -n ai-persona-system exec <pod> -- grep -aq "<check_name>" /proc/1/exe`
  plus a must-be-present and a must-be-absent control. Never `strings`, never a discovery grep.
- **Expect SILENCE from the six directory checks and do not read it as failure.** They
  self-gate on the per-site opt-in flag (no site carries a finance key yet) AND on current
  found claims of that kind (B4 populated all three). The pilot is what arms them.
- **The five structural checks are the opposite case: they fire on every dispatched
  completeness sweep and their real-site finding volume is [UNMEASURED]** — A1 shipped them
  flag-only for exactly that reason. **Watch the first post-434 sweeps.**
- **The planner's directory rule is live but has NEVER FIRED** — no site carries a finance
  key. The pilot is its first real exercise.
- Re-verified 2026-08-16 16:12Z after the 17h gap: checks array still 43, planner rule still
  present, planner still exactly one active row. No other lane disturbed either.

## 3. Next actions, in order

1. ~~Read the 433 round-2 verdict~~ **DONE 2026-08-17 — APPROVED (16:38:18Z on 08-16).
   PHASE B'S COUNCIL TRAIL IS COMPLETE: 429 · 432 (r2) · 433 (r2) · 434, all approved.**
   Nothing outstanding; go straight to the pilot. All advisories dispositioned in NOTES —
   the one real defect (verify blocks doing arithmetic on a possibly-NULL `p`, so the check
   **cannot fail**) is fixed in both un-run ROLLBACKs; the applied forward files are left as
   the record of what ran.
   > **Also verified on the fresh v1.0.1305 roll** (both replicas, 08-16 22:07Z): Phase B is
   > config-only so a roll cannot undo it, but B3f depends on the BINARY carrying the check
   > names and a fresh build is not automatically a newer commit. Re-probed with controls —
   > finance check PRESENT, structural check PRESENT, `evaluate_directory_features` PRESENT,
   > POS PRESENT, NEG ABSENT. **Preconditions hold on 1305.**
2. **Phase C pilot — remortgagecalculator.uk (M4), end to end.** Mission-file from the
   register entry, pre-seeded specs, marker sentence, dispatch via
   `scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh`, cost baseline
   from `llm_call_log`/`assets`, owner sign-off before Phase E.
   **Three things this pilot is now the live proof of, in order of what breaks first:**
   - **432**: the built site's classification spec should carry
     `content_features.mortgage_lender_directory` at plan time. CHECK IT — that is 432's proof.
   - **433**: the plan should then contain a `mortgage-lenders` page (name AND page_type
     exactly that) composed hero → `mortgage-lender-directory-listing` → call-to-action, plus
     a `mortgage-lender-directory` section on the homepage. That is 433's proof.
   - **434**: if 433 misfires, `missing_mortgage_lender_directory_page`/`_section` should
     raise work items on the first completeness sweep. That is 434's proof — and it is the
     safety net, so a clean pilot means checking 433 directly rather than inferring from silence.
3. Then Phase D decisions / Phase E waves per `PLAN_2026-08-12_fleet_buildout.md`.

## 4. Standing cautions (carried forward)

- The mortgagecalculator.co.uk voice review is ACTIVE IN ANOTHER THREAD (session "copy
  quality two stage") — do not duplicate from this lane.
- Bug 270 owned elsewhere — cite, hands off.
- `git stash` forbidden (hook-blocked); pathspec commits; forward-only; re-run `git status`
  before acting on it. **The index lock is contended** — this session hit
  `.git/index.lock` held by another session mid-commit; WAIT for it, never remove it.
- Migration files: **435–440 were taken DURING this session** (the queue moved while I
  worked); this lane then took **441**, so next free was **442** at session end — RE-CHECK.
  One of those, **439** (another lane), added `menu_field: "available_components"` to
  `validate_plan` — it strengthens 433/441 and is a reminder that the planner config has
  several concurrent editors.
- **Rolling back the planner pair has an ORDER**: 441 edits text inside the block 433
  inserted, so 433's surgical-inverse ROLLBACK refuses (by design) until 441's ROLLBACK has
  run. Both file headers say so; the refusal message alone would send you hunting for drift
  that is not there.
- `improvement-sweep` is still disabled fleet-wide. Not ours to re-enable unilaterally —
  find who disabled it (git/NOTES grep or ask the owner) before flipping.

## 5. Files of record

This dir: `PLAN_2026-08-12_fleet_buildout.md` (phase map) ·
`SUMMARY_2026-08-15_guardrails_live_directories_built.md` +
`SUMMARY_2026-08-15b_first_supervised_runs.md` (milestones — **a Phase-B-complete summary is
owed and is the right next one to write**) · `NOTES_portfolio_positioning.md` (evidence,
newest at bottom — the two 2026-08-15 fifth-session entries cover B3e/B3f in full) ·
`README_where_we_are.md` (owner's log, appended 2026-08-15 late evening) ·
`COUNCIL_SUBMISSION_433_…json` / `COUNCIL_SUBMISSION_434_…json`.
Migrations: `sql_for_agents/433_…` + `434_…` (+ROLLBACKs), all applied.
Register: `docs026_concept_register/register/directory-pipeline.md` (DIR-001 — status now
records all seven places filled, plus which two fail silently).
Architecture track: `architecture_review/RFC_031_hand_spliced_enrichment_steps_want_an_ordered_list.md`.
Commits: `7aff17b21` (B3e), `a1b92f609` (B3f), `6f9374824` (432 advisories + RFC_031 + NOTES).

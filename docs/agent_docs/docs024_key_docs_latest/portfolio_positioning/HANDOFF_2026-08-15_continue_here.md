# HANDOFF — Phase B live and activated; first supervised runs + wiring migrations next — 2026-08-15, continue here

Supersedes `HANDOFF_2026-08-13_continue_here.md` (correct on its own history; this file
carries everything a fresh chat needs). Milestone read-out for the owner:
`SUMMARY_2026-08-15_guardrails_live_directories_built.md`.

## 0. Owner rulings in force

1. Phase A gate (structural-validity + RFC_025 before any new pipeline domain): **SATISFIED
   AND LIVE** (since v1.0.1295; re-verified on v1.0.1301).
2. Directories from day one on new sites; **non-price facts only** (now enforced
   mechanically at registration); "cited facts", never "verified".
3. Starter kinds: `mortgage-lender`, `savings-provider`, `health-insurer`; ~6 more later.
4. Tools stay on manual per-site review; graphs = `evidence-chart` only.
5. **Bug 270 is OWNED BY ANOTHER THREAD (owner, 2026-08-15) — hands off**, cite it freely.
6. Owner directed round 4 of the B1 council review (submitted 2026-08-15, corr
   `69a619e6-5152-45d8-ae01-5d30a0c7776f` — check the verdict before further rounds).

## 1. Current state (all verified at the artefact, 2026-08-15)

**Binary**: `v1.0.1301` on both chassis replicas carries EVERYTHING (Phase A guardrails +
Phase B directory code). Gate literal `"per kind (Phase B kind-scoped keys)"` probed in
`/proc/1/exe` on both, absent-sha control clean.

**Phase B activation, executed this session in the recorded order**:
- Components: 6 rows applied to `content_components` (all `section`-level; the seed was
  ROLLBACK-dry-run-validated first; subheadline guidance carries the honesty clause "a
  citation proves where a fact came from, not that it is correct").
- Discovery tasks: 3 rows `enabled=true` with `last_triggered_at=now()` — **first fire
  deliberately deferred one interval so the first runs happen under SUPERVISION via
  force-trigger, not unattended** (see §2.1).
- Legacy constant-key rows: pre-counted (exactly 2, as predicted: `39b5153f…`
  `stale_directory_claim`, `35350447…` `directory_citation_unverified`), cancelled in one
  guarded UPDATE with a result note naming the successor keys and the correlation.

**Council**: B1 corr `69a619e6` — REVISE ×3 (round 1 substance, fixed in `035f72365`;
rounds 2-3 form; full analysis in NOTES), **round 4 submitted at owner direction**
2026-08-15 with the activation record and every asked-for check answered. Read it:
`SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE
correlation_id='69a619e6-5152-45d8-ae01-5d30a0c7776f' AND kind='council_report' ORDER BY 1;`

**Commits this phase**: `6f26570e4` (B1+B2), `035f72365` (round-1 fixes), B3a/B3b seeds +
notes commits, `SUMMARY_2026-08-15…`. All platform commits trailered with the corr.

## 2. Next actions, in order

1. **B4 — supervised first researcher runs.** One kind at a time:
   `UPDATE scheduled_tasks SET last_triggered_at=NULL WHERE name='mortgage-lender-directory-discovery';`
   then watch the orchestration, then:
   - claims landed? `SELECT de.kind, de.name, dc.field, dc.value, dc.status FROM
     directory_claims dc JOIN directory_entities de ON de.id=dc.entity_id WHERE de.kind='mortgage-lender' AND dc.is_current;`
   - rejects under the KIND-SCOPED key (`directory_citation_unverified:mortgage-lender`)?
     Price-field refusals appearing as rejects is the CONTROL WORKING, not a failure.
   - Work the HITL queue; bar = a reviewed, non-embarrassing set per kind. Repeat ×3 kinds.
2. **B3c** — publish-trigger fix (three kind-blind predicates + LIMIT 5,
   `SEED_directory_publish_trigger.sql:94`, per the 2026-08-10 FINDING; snapshot-first
   agent_definitions migration + publisher chain extension). Config, live immediately.
3. **B3d** — wire `evaluate_directory_features` into improvement-loop (after
   `enrich_news_feed`, READ THE LIVE CONFIG first — 291 re-pointed edges to
   `load_audit_state`) AND into `domain-research-classifier` after
   `write_classification_spec` (greenfield builds need the flag at plan time).
4. **B3e** — build-site-planner prompt rule (206's replace()-idiom, snapshot, anchors) +
   the six component names into the planner vocabulary.
5. **B3f** — enable the 6 directory checks AND the 5 Phase-A structural checks on
   completeness-discovery-agent (194/215 jsonb_set pattern). Binary precondition already
   satisfied; directory checks self-gate on publishable data, so after B4 preferably.
6. Then **Phase C pilot** (one real proposition end-to-end, cost baseline from
   `llm_call_log`/`assets`, owner sign-off) → Phase D decisions → Phase E waves.

## 3. Decisions the owner has been asked to make (also delivered in chat 2026-08-15)

1. **Pilot domain** (Phase C): which proposition goes first. Recommendation: `M2
   mortgagecalculator.co.uk` is ADOPTED (not greenfield), so the cleanest greenfield pilot
   with a directory is an M-family domain, e.g. `mortgage-rates.co.uk` (M3) or
   `remortgagecalculator.uk` (M4) — mortgage-lender directory data will exist first.
2. **Build order across ~140 domains** (Phase E) — the register says this is the owner's
   commercial call; a family-by-family default (M → B → I, starting where a live sibling
   exists to sanity-check against) is on offer.
3. **loanzy.uk conflict** — the webdesign lane has provisioned it for an unrelated purpose;
   L9 register direction says hold. Needs an explicit release-or-keep.
4. **B8 (savingsapp.co.uk), B9 (bankingequipment.co.uk), I10 (brandable cluster)** — build
   or hold; or simply excluded from wave 1.
5. **mortgagecalculator.co.uk voice review** — queued since 2026-08-11: live homepage +
   title pass vs `mortgagecalculator_couk_adoption/REFERENCE_2026-08-11_learned_by_correction_house_voice.html`.
6. **If round 4 comes back REVISE again on form**: accept the advisory record and proceed
   (the trail is honest, the code is live), or keep iterating. Rounds cost ~30 min + credits
   each; round 1 was the only one that changed code.

## 4. Landmines fresh in this lane

- The submission schema caps `plan.edits` at 8 — never CLAIM completeness of the array for
  a >8-file change; state the cap and name the remainder in risks/grounded_in.
- `git stash -u` on this tree sweeps other sessions' WIP — pathspec-scoped stash only.
- Subagent session limits killed two background agents mid-task this phase; partial
  worktrees were harvested via `git diff` extraction (check `git worktree list` for
  leftovers if an agent dies).
- The bug-231 image-test failures in the full `actions` suite are ANOTHER session's WIP
  (verified passing at clean HEAD) — do not chase them.
- `agent_definitions` references to the new action/spec keys measured 0 on 2026-08-15 —
  nothing dispatches `evaluate_directory_features` until B3d lands.

## 5. Files of record

This dir: `PLAN_2026-08-12_fleet_buildout.md` (phase map) ·
`SUMMARY_2026-08-15_guardrails_live_directories_built.md` (read-aloud milestone) ·
`NOTES_portfolio_positioning.md` (evidence, newest at bottom) · `README_where_we_are.md`.
Lane dir: `finance_directory_pipeline/` (3 seeds, all applied; components seed carries the
apply gate + honesty clause). Register: `docs026_concept_register/register/directory-pipeline.md`
(DIR-001) + index row. RFC: `RFC_025_…` §10 (LIVE status). Bugs: `bugs_open/270` (owned
elsewhere — hands off).

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
6. **B1 council review: APPROVED at round 4** (2026-08-15 11:01, corr
   `69a619e6-5152-45d8-ae01-5d30a0c7776f`). Both platform commits carry the trailer, so
   098 credits them automatically. CLOSED — no further rounds.
7. **P9, the owner's six decisions (2026-08-15, all recorded):**
   - **Pilot domain: `remortgagecalculator.uk` (M4)** — weak preference, either M-family
     candidate was acceptable; this one goes first.
   - **Build order: family-by-family, M → B → I**, starting where a live sibling exists to
     sanity-check against. Agreed as proposed.
   - **loanzy.uk: stays with the webdesign lane** (register L9 updated, P9 note; the domain
     remains L9-claimed so nothing drifts onto it, but its use is theirs).
   - **B8 / B9 / I10: HOLD** — excluded from the first waves (register updated,
     `check_register.py` passes: 152/152, invariants hold).
   - **mortgagecalculator.co.uk voice review**: see §3a below — the owner notes the current
     homepage copy "is bad at the moment", so treat the review as live and possibly
     corrective, not a formality.
   - **Council posture: accept the advisory record** rather than iterate on form (moot for
     B1 — round 4 approved — but standing guidance for future reviews in this lane).

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

1. **B4 — supervised first researcher runs. ✅ DONE 2026-08-15 (third session) — ALL
   THREE KINDS PROVEN.** Savings run 6 (`c2cd7f55`): 15/15 registered first run, 13
   named firms, GOV.UK ISA-managers list as anchor source. Health runs 7-8 (`8b6f8e12`,
   `297ca621`): 10 named insurers, with migration **428** between them (no monetary
   amounts inside values; one claim per (provider, field); underwriter = firm name;
   forbes.com excluded) — all four rules proven at the artefact on run 8. Register at
   close: 25 active entities / 33 cited claims. Queries for savings/health were
   mirrored to the proven iteration-2 membership-list shape BEFORE their first runs
   (seed synced). HITL: all three kind items ruled; health's 3h reject-suppression
   window expires ~20:37Z, harmless at weekly cadence. Full evidence: NOTES same date.
   The recipe below stays for reference for future kinds. What that session
   learned (full evidence in NOTES, same date):
   - Config now carries migrations **423** (entity = ONE NAMED FIRM, never a
     sector/aggregate) and **424** (quote = ONE CONTINUOUS passage; ibisworld.com
     excluded — refetch-blocked, citations can never verify). Queries re-aimed at named
     firms, **⚠ must stay <200 BYTES** (web_search's query_from drops ≥200-char queries
     as "likely LLM error message" and the run FAILS with an error that misdirects at
     config keys).
   - Read results from the orchestration's `collected_data`
     (`candidate_claims`/`registration`/`verify_and_register`), not just the register.
   - ⚠ **Completing a reject HITL item suppresses that kind's reject writes for 3h**
     (`writeWorkItem` two-strike rule; the drop is silent and the rejects survive only in
     `collected_data`, ~24h). Fix candidate recorded in NOTES: emitter should set
     `recurrenceExpected` (Go, needs a roll). Until then: rule on HITL items LAST, after
     the kind's supervised runs are done.
   - ⚠ This action's log lines never reach `kubectl logs` (mechanism undiagnosed, NOTES
     has the measurements) — absence of its log line is NOT evidence; use the DB.
   Per kind: `UPDATE scheduled_tasks SET last_triggered_at=NULL WHERE name='<kind>-directory-discovery';`
   then watch `orchestration_states` (owner_agent_type='finance-directory-researcher'),
   then:
   - claims landed? `SELECT de.kind, de.name, dc.field, dc.value, dc.status FROM
     directory_claims dc JOIN directory_entities de ON de.id=dc.entity_id WHERE de.kind='<kind>' AND dc.is_current;`
   - rejects under the KIND-SCOPED key (`directory_citation_unverified:<kind>`)?
     Price-field refusals appearing as rejects is the CONTROL WORKING, not a failure.
   - Work the HITL queue (last — see suppression note); bar = a reviewed,
     non-embarrassing set per kind.
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

## 3. Owner decisions — ALL ANSWERED 2026-08-15 (see §0.7 for the rulings)

The former open-decision list is resolved; what remains actionable from it:

### 3a. The mortgagecalculator.co.uk voice review (the one item needing the owner's eyes)

Where to look, precisely:
- **The reference** (the four owner corrections from 2026-08-11, written as the corrections
  that produced them — the standard the copy is judged against):
  `docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/REFERENCE_2026-08-11_learned_by_correction_house_voice.html`
  (it was also published as a private artifact page for the owner that evening — the
  artifact gallery has it under the "Learned by correction" title).
- **The candidate**: the LIVE homepage at `mortgagecalculator.co.uk`, plus whether the
  31-title pass ever finished (that lane's own next-action #1; check
  `mortgagecalculator_couk_adoption/`'s NOTES tail — as of 2026-08-11 only the homepage's
  new `<title>` had reached served HTML).
- **The owner's standing impression (2026-08-15): "the home page copy is bad at the
  moment."** So the review session should treat this as a live copy problem to diagnose
  against the reference (which correction is being violated, or is the spec itself still
  wrong), not a rubber-stamp. The 2026-08-11 ruling stands: the model is NOT the lever —
  brief/spec first. Route findings INTO that lane's docs (`who-owns` it first; it was
  recently active).
- **Added 2026-08-15 (owner, session brief): copy-voice work is ACTIVE in another thread** —
  session "copy quality two stage", id `79d969f9-0009-4540-84cc-2557222db288`. Do not
  duplicate the voice review from this lane; coordinate through that session's docs.

### 3b. Pilot execution note (decision made: remortgagecalculator.uk, M4)

M4's register entry: audience = existing owner with a fix ending in ≤6 months, calculator +
deadline checklists, urgency register. Its `.uk` TLD = the instrument side per P5
(tool-first, the 10-second answer). The mortgage-lender directory kind serves it. Phase C
steps in `PLAN_2026-08-12_fleet_buildout.md` apply: mission-file from the register entry,
pre-seeded specs, marker sentence, cost baseline from `llm_call_log`/`assets`, owner
sign-off before Phase E.

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

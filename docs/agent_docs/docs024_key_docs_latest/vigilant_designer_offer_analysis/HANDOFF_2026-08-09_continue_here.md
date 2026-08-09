# HANDOFF — vigilant designer + offer analyser (2026-08-09)

**COLD-START = this file + PLAN_2026-08-02 (programme + owner decisions) +
PLAN_2026-08-08_B1_B2_premise_first.md (execution corrections block) +
features_open/030 (the wider offer scope + open owner questions). NOTES tail has the
missteps. This supersedes HANDOFF_2026-08-03_continue_here.**

## State at handoff

**Programme B (offer track) — B1+B2 LIVE + WITNESSED, B3 BUILT-NOT-LIVE:**

- **B1** migration slug `site_review_agent_loads_the_premise` (number 340 — AMBIGUOUS,
  bugfix_220 took the same number; resolve by slug) — site-review-agent loads
  strategy/audience/identity_head_4k/content_direction_head_4k/mission_brief, judges
  against `revenue_models.primary_model`, honesty constraint. **WITNESSED 08-08**: planted
  marker reached `llm_call_log.prompt_rendered` (29,110 chars, no truncation), marker
  un-planted after. Ledger row recorded (owner's hand).
- **B2** migration slug `domain_strategist_refresh_safe_and_premise_fields` (341 —
  same ambiguity) — deployed-site gate before `create_next_item` + four premise fields
  (restoration of gaswholesalers' 04-17 shape). **WITNESSED 08-08**: loancalculator.co.uk
  got its FIRST strategy row (affiliate + all four fields), ZERO needs_briefing.
  **AMENDED 08-09 by migration 359** (`domain_strategist_gate_uses_shipped_predicate`,
  applied): the gate now uses the SHIPPED predicate — `build_status='deployed'` bare
  misses needs_rebuild-but-serving pages, and in the gate that miss CHAINS A RE-PLAN of
  a serving site. Witness invariant re-verified before applying.
- **B3** commit `ad51ca863` + follow-up `b26fdc81b` — `check_premise_incomplete` +
  `check_revenue_shape` in `platform/orchestration/actions/discovery_checks/`, tests
  green, gofmt clean. Verifiers for `revenue_shape_cta` + `missing_conversion_path`
  (fail-closed); claimed-item-timeout lockstep closed BOTH halves (220 declared +
  migration 358 APPLIED live). Register entry **BIZ-031** carries the RFC_010 §1 duty:
  needs_strategy producer set = vertical-exemplar-researcher + check_premise_incomplete,
  shared key `strategy_<domain>`.
  **Council: `Council-Submitted: 5cd586c9-c787-417a-a102-27fbddc48687` — VERDICT UNREAD
  at handoff.** Find by payload:
  `SELECT current_step, status FROM orchestration_states WHERE
   collected_data->'input_data'->>'fix_correlation_id' = '5cd586c9-c787-417a-a102-27fbddc48687';`
  A REVISE/REJECTED lands on whoever reads this first — the code is on the shared branch.

**Programme A — unchanged since 08-05:** next action is still ONE witnessed css-patch
run to discharge `bugs_open/198`, then A2 (seed `design-critique-agent` + the
Gemini-vs-Claude trial). A2 landmines from the 08-03 handoff still stand: start_step
never initial_step (VIZ-012); root ai_service SHADOWS step-level (MDL-039); the critic
must read `renders_failed` and refuse a partial sweep.

## Next session, in order

1. **Read B3's council verdict** (corr above; ~30 min budget from 08-09 ~00:45Z
   submission; queued ≠ lost, find by payload). Act on objections — the code is live on
   the shared branch already (committing-is-shipping).
2. **OWED: ledger rows for migrations 358 + 359** — the session permission classifier
   blocks INSERT INTO schema_migrations from Claude sessions. Owner commands (single
   lines, the 08-08 paste-wrap trap):
   `SUM=$(md5sum docs/agent_docs/sql_for_agents/358_revenue_shape_claim_timeout_exclusions.sql | awk '{print $1}') && kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "INSERT INTO schema_migrations (filename, checksum, applied_by, notes) VALUES ('358_revenue_shape_claim_timeout_exclusions.sql', '$SUM', 'record-only', 'B3 lockstep: applied 2026-08-09, pre/post assertions passed') ON CONFLICT (filename) DO NOTHING;"`
   `SUM=$(md5sum docs/agent_docs/sql_for_agents/359_domain_strategist_gate_uses_shipped_predicate.sql | awk '{print $1}') && kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "INSERT INTO schema_migrations (filename, checksum, applied_by, notes) VALUES ('359_domain_strategist_gate_uses_shipped_predicate.sql', '$SUM', 'record-only', 'B2 gate shipped-predicate fix: applied 2026-08-09, witness invariant re-verified') ON CONFLICT (filename) DO NOTHING;"`
3. **Image roll** carrying `ad51ca863`+`b26fdc81b` (any fleet roll from HEAD does it —
   committing-is-shipping; or build one: bump IMAGE_TAG, `make build-agent-chassis`,
   verify at the POD with positive `check_premise_incomplete` AND a negative control;
   grep -ac if the image has no `strings` — 08-03 landmine).
4. **Config: add the check names** `premise_incomplete` + `revenue_shape` to
   quality-discovery-agent's checks array — **ONLY after the pod-grep** (unregistered
   name = FATAL since 149 B4). Small migration; current array is 6 names (broken_nav_links,
   placeholder_contact, generic_theme, unverified_claims, voice_tells, literal_markdown).
   IMP-016: observe-only first — the items are born `detected` and nothing dispatches
   them until a sweep's triage promotes; read what they file before letting one promote.
5. **First observe-only run**: hand-fire `./run_improvement_sweep_once.sh <site>` on a
   site with a recorded shape worth testing (the 10×direct_business population;
   webdesign.uk or oufe.com are direct_business AND genuinely businesses — expect
   silence; a topic-named direct_business row is where a real finding should appear).
   Read the findings as ARGUMENTS, not certainties; tune the lexicon before promotion.
6. **Then A-track or B4** — owner's call. B4 (the analyser agent) now has everything it
   reads: premise in review context (B1), four premise fields (B2), mechanical floor (B3).

## Watch-outs / owed proofs

- **Greenfield negative control [NOT EXERCISED]**: no greenfield strategist run since the
  B2 gate. The next real greenfield build must file needs_briefing (else-arm unchanged by
  construction, 359 preserved it — but witnessed is witnessed). Check + note in NOTES.
- **Pre-existing package test failure NOT ours**: `TestEveryCheckProducedItemTypeIsClassified`
  fails on `decision_regression` (check_decision_guards.go, RFC_015 lane, e1628f7df) at
  CLEAN HEAD — verified via git archive 08-09. Do not "fix" it into our lane; their
  obligation. Everything else in the package is green.
- **Numbers 340/341 are ambiguous** (two lanes, same evening, four applied+recorded
  files). Resolve migrations by SLUG. A number is reserved by the COMMIT, not by ls
  (355→357 appeared during ONE session; we took 358/359).
- **`agent_definitions.updated_at` does not move** on default_config-only UPDATEs — date
  applies via `agent_definitions_backup.snapshot_taken_at`. In-flight orchestrations run
  the config they SPAWNED with.
- webdesign.co.uk carries ~56 failed page_rerender + 10 failed literal_markdown from
  earlier eras — pre-existing, flagged 08-08, unowned.
- B1 loads capped context (`identity_head_4k` etc.) — if the strategic review starts
  truncating (`error_message ILIKE '%stop_reason=max_tokens%'` — output_tokens is NULL
  on cut first attempts), the premise blocks are the first suspect.

## Who owns what nearby (unchanged + one addition)

portfolio_positioning owns premise→writer wiring (they were consumers-told on B2:
`CONTRIB_2026-08-08_domain_strategist_guarantee_changed.md` in their dir);
brochure_component_library owns 016's first-user relationship; bugfix_149 owns
checker-layer plumbing. This lane owns: the drain, the critic, the recompose handler,
anti-brochure compose-time work, and the offer analyser (B track, now through B3).

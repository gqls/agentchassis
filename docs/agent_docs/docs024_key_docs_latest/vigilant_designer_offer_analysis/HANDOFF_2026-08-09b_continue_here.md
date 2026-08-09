# HANDOFF — vigilant designer + offer analyser (2026-08-09b, late evening)

**COLD-START = this file + PLAN_2026-08-02 (programme + owner decisions) +
features_open/030. NOTES tail (08-09 late-evening entries) has the missteps.
This supersedes HANDOFF_2026-08-09_continue_here.**

## State at handoff

**Programme B — B1+B2+B3 ALL LIVE. B3 is now WITNESSED RUNNING, observe-only:**

- **B3 code is live on the fleet**: v1.0.1276 carries `ad51ca863`+`b26fdc81b` on both
  agent-chassis replicas — proven at the artefact (positive `VerifyRevenueShapeCTAResolved`
  greps 2 both pods; the literal b26fdc81b REMOVED greps 0; sibling-literal spelling
  control greps 1).
- **Enabled observe-only by migration 361** (`quality_discovery_enables_offer_track_checks`,
  applied + committed `e7e8402a1`, snapshot 54df4b7b): `premise_incomplete` +
  `revenue_shape` appended to quality-discovery's checks array (7→9 entries —
  `decision_guards` had arrived from the RFC_015 lane since the last handoff said 6).
- **First runs WITNESSED on three sites via SCH-025 oneshot envelopes** (all three task
  rows now disabled): webdesign.uk silent (false-positive control, correct);
  dartsonline.com silent and HAND-VERIFIED truthful (/contact.html has a form, linked
  from chrome — retraction arm, no-op); gaswholesalers.com **TRUE POSITIVE** —
  `needs_strategy` filed, key `strategy_gaswholesalers.com`, born `detected`, correct
  reason (strategy row exists, no `revenue_models.primary_model`). Zero checks_failed,
  zero checks_unregistered, all three runs.
- **Do NOT use `run_improvement_sweep_once.sh` for observe-only reads** — its
  triage_findings PROMOTES detected items on every path (its own header says so). SCH-025's
  rotation (`site-discovery-rotation-quality`, hourly, live since ~09:49Z 08-09) now brings
  every site past the checks weekly with NO triage carrier; oneshot rows are the targeted
  vehicle (copy the envelope in scheduled_tasks, disable after firing).
- **Rollback files exist and are committed** (`12966b87d`): 358/359/361 each have a
  `_ROLLBACK.sql` pinning its pre-state with DO/RAISE and stating its hazard (358's alone
  reopens the verifier bypass; 359's restores the dangerous-direction predicate).

**Council trail (corr `5cd586c9-c787-417a-a102-27fbddc48687`), two rounds, REVISE both:**

- Round 1 gated on "verifiers likely never register" — WRONG about the code (init() in
  check_revenue_shape.go:84-88; the coverage test auto-covers registered types at :323;
  the lockstep test passes at HEAD, impossible without registration). Round 2 answered
  with file:line + live queries; four seats accepted the registration reading.
- **Round 2 gates on the schema_migrations LEDGER GAP** — 358/359 (now +361) applied live
  with no ledger rows, flagged high by editquality, tooling_provenance, debug_historian,
  prior_art. They are right, and it is OWNER-BLOCKED (the session permission classifier
  refuses the INSERT from Claude sessions). **Round 3 must WAIT until the owner runs the
  three commands below** — resubmitting before that draws the same objection.
- Round 3 should also: quote round 1's verdict verbatim (prior_art asked); attach
  code_checks for the file:line claims; mark the already-shipped edits as HISTORICAL
  RECORD, not forward edits (editquality's audit/edit distinction); cite CLAUDE.md's
  "OWNER RULING 2026-07-29" section by path for the review-after-the-fact point (a seat
  called it unverifiable); answer guardian's retraction-collateral point: retraction on
  premise-COMPLETE is semantically correct for BOTH needs_strategy producers (the item
  asks for a strategy; one exists; BIZ-031 records the shared-key convergence — that is
  RFC_010 §1 working as designed), and note the rollback files now exist.

## OWED: three owner commands (single lines, paste-wrap trap applies)

```
SUM=$(md5sum docs/agent_docs/sql_for_agents/358_revenue_shape_claim_timeout_exclusions.sql | awk '{print $1}') && kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "INSERT INTO schema_migrations (filename, checksum, applied_by, notes) VALUES ('358_revenue_shape_claim_timeout_exclusions.sql', '$SUM', 'record-only', 'B3 lockstep: applied 2026-08-09, pre/post assertions passed') ON CONFLICT (filename) DO NOTHING;"
SUM=$(md5sum docs/agent_docs/sql_for_agents/359_domain_strategist_gate_uses_shipped_predicate.sql | awk '{print $1}') && kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "INSERT INTO schema_migrations (filename, checksum, applied_by, notes) VALUES ('359_domain_strategist_gate_uses_shipped_predicate.sql', '$SUM', 'record-only', 'B2 gate shipped-predicate fix: applied 2026-08-09, witness invariant re-verified') ON CONFLICT (filename) DO NOTHING;"
SUM=$(md5sum docs/agent_docs/sql_for_agents/361_quality_discovery_enables_offer_track_checks.sql | awk '{print $1}') && kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "INSERT INTO schema_migrations (filename, checksum, applied_by, notes) VALUES ('361_quality_discovery_enables_offer_track_checks.sql', '$SUM', 'record-only', 'B3 enablement observe-only: applied 2026-08-09, induced-failure test run first, snapshot 54df4b7b') ON CONFLICT (filename) DO NOTHING;"
```

## Next session, in order

1. **Check the owner ran the three ledger INSERTs** (`SELECT filename FROM
   schema_migrations WHERE filename ~ '^(358|359|361)'` — expect 3). Then **round 3** on
   the SAME correlation (RESUBMIT_CORR=5cd586c9-…), per the round-3 list above. The
   round-2 report is the latest council_report row on that correlation.
2. **Read what the rotation files** as it brings sites past the checks (observe-only;
   items born `detected`): `SELECT item_type, item_key, status, summary FROM
   site_work_items WHERE item_type IN ('needs_strategy','revenue_shape_cta',
   'missing_conversion_path') ORDER BY created_at DESC;` Findings are ARGUMENTS —
   tune the lexicon before any promotion decision. The saas_tools sites
   (webdesign.co.uk, gamesdesign.co.uk, robot-hands.com) are where revenue_shape_cta
   has its first real chance to fire; they were unstamped at handoff, so the rotation
   reaches them within hours.
3. **Watch the 090 on the scheduler**: needs_diagnosis intake corr
   `47299b0e-f6ac-4708-917c-4afecd366cf5` (kafka-scheduler OOM-looping since the
   ~19:45Z roll — fleet incident, NOT this lane's, but our dispatches ride the
   scheduler's healthy windows until it is fixed; expect scheduled work to be slow).
4. **Then A-track or B4** — owner's call, unchanged from the last handoff.

## Watch-outs / owed proofs (carried + new)

- **Greenfield negative control [STILL NOT EXERCISED]**: no greenfield strategist run
  since the B2 gate. The next real greenfield build must file needs_briefing.
- **`site_work_items.created_by` reads 'generic' on our own rows** (witnessed on the
  gasw needs_strategy row) — the agentType-fallback landmine. BIZ-031's register entry
  is the ONLY producer record; count(DISTINCT created_by) is structurally blind here.
  If producer-splitting is ever needed, discriminate in the spec.
- **Numbers 340/341/358/359/361 all resolve by SLUG** (340/341 ambiguous with
  bugfix_220's; 356 is absent from the ledger too — another lane's business).
- **Pre-existing package test failure NOT ours**: `TestEveryCheckProducedItemTypeIsClassified`
  fails on `decision_regression` (RFC_015 lane, e1628f7df) at clean HEAD.
- B1 truncation watch-out unchanged (capped context; `stop_reason=max_tokens` in
  error_message, output_tokens NULL on cut first attempts).
- webdesign.co.uk's ~56 failed page_rerender + 10 failed literal_markdown remain
  pre-existing and unowned.

## Who owns what nearby (unchanged from 08-09 handoff)

portfolio_positioning owns premise→writer wiring; brochure_component_library owns 016's
first-user relationship; bugfix_149 owns checker-layer plumbing; bugfix_230 owns SCH-025
(the rotation + watchdog — their register entry status was corrected by this lane 08-09,
with a dated strike-through, after it went stale the same day it was written). This lane
owns: the drain, the critic, the recompose handler, anti-brochure compose-time work, and
the offer analyser (B track, B1+B2+B3 live).

-- PATCH_fix_proposer_011_guidelines_declared_contracts_exemption.sql
-- bugs_open/011 §7 owed item 4. 2026-07-24. clients_db.
--
-- WHY: the guidelines seat itself asked for this, twice (council e996bf0a,
-- rounds 4 and 8), while APPROVING the plan both times: DECLARED CONTRACTS as
-- written governs workflow INPUTS declared in input_contract at a call site.
-- It has no clause for the pattern the 011 residual shipped — an adapter
-- emitting an OPTIONAL response-body field (`reported_conditions`) that
-- chassis-internal logic conditionally acts on (persist to agent_error_log),
-- gated by a consumer-owned sender allowlist rather than by output_contract.
-- Round 8, guidelines seat: "this exposes a genuine gap worth flagging as a
-- side-task: the guidelines have no documented mechanism for 'adapter emits an
-- optional response field that chassis logic conditionally acts on' ...
-- Recommend a guideline amendment codifying when such sender-gated,
-- non-contract response fields are acceptable (essentially ratifying the
-- pattern this plan already uses) versus when they must go through
-- output_contract. This is a gap in the rule, not a defect in the plan."
-- Without the clause, every future reviewer relitigates the same call.
--
-- WHAT: append an EXEMPTION sentence to the DECLARED CONTRACTS bullet in the
-- guidelines seat's prompt, scoped to the review_guidelines step of the LIVE
-- fix-proposer row only. Headers already work this way (undeclared,
-- chassis-parsed) — the clause names the conditions that make the pattern
-- acceptable, all three of which are load-bearing in the ratifying precedent.
--
-- ORDER MATTERS (CLAUDE.md): seat fix-proposer, then mirror to the gate with
-- 099_SYNC_gate_roster.py --apply. Do NOT hand-patch council-gate.
--
-- SURGICAL + IDEMPOTENT: replace() of an exact substring inside one step's
-- prompt_template; after a successful run the old substring no longer exists,
-- so a re-run is a no-op. Snapshot first.

BEGIN;

SELECT snapshot_agent('fix-proposer', 'pre-update: 011 — DECLARED CONTRACTS gains the sender-gated response-field exemption the guidelines seat requested (council e996bf0a rounds 4+8)');

UPDATE agent_definitions SET default_config = jsonb_set(
  default_config,
  '{workflow,steps,review_guidelines,config,prompt_template}',
  to_jsonb(replace(
    default_config#>>'{workflow,steps,review_guidelines,config,prompt_template}',
    $OLD$- DECLARED CONTRACTS: any input a workflow reads must be declared in the agent's input_contract; a call site's input_mapping must satisfy the callee's contract.$OLD$,
    $NEW$- DECLARED CONTRACTS: any input a workflow reads must be declared in the agent's input_contract; a call site's input_mapping must satisfy the callee's contract. EXEMPTION (ratified 2026-07-24 after council e996bf0a / bugs_open-011, where this seat twice asked for the clause): an OPTIONAL response-body field consumed only by chassis-internal logic (telemetry/logging -- never read by a downstream workflow step) may bypass output_contract PROVIDED all three hold: the consumer gates it on an explicit consumer-owned sender allowlist whose additions require review; absence of the field is a silent no-op; an unsanctioned sender emitting it is warned loudly by name. Precedent: reported_conditions -> senderMayReportConditions. A field a workflow step READS still requires the declared contract.$NEW$
  )),
  false)
WHERE type='fix-proposer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Verify inside the transaction: the exemption present exactly once, the old
-- bare sentence gone, and no other step touched.
SELECT
  (SELECT count(*) FROM regexp_matches(
     default_config#>>'{workflow,steps,review_guidelines,config,prompt_template}',
     'EXEMPTION \(ratified 2026-07-24', 'g')) AS exemption_occurrences_expect_1,
  position($OLD$satisfy the callee's contract.
$OLD$ IN default_config#>>'{workflow,steps,review_guidelines,config,prompt_template}') AS old_bare_line_still_present_expect_0,
  (SELECT count(*) FROM jsonb_object_keys(default_config->'workflow'->'steps')) AS step_count_expect_unchanged
FROM agent_definitions
WHERE type='fix-proposer' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

COMMIT;

-- PROOF harness for bugs_open/302 — TEMPORARY, applied directly (NOT a migration).
--
-- Deliberately NOT placed in sql_for_agents/: this is scaffolding to be deleted within
-- the hour, and a file there is re-applied by any session's unscoped `run-migrations.sh
-- --apply`. Teardown sidecar: PROOF_2026-08-19_probe_teardown.sql.
--
-- WHY A PROBE AGENT AND NOT A REAL DISPATCH. The arm under test fires when the payload
-- the gate is handed cannot be read. Post-bugs_closed/287 a real color-variable-fixer
-- dispatch returns its READABLE envelope, which exercises the OTHER arm. The production
-- shape that does reach this arm is `complete_work_item` called with NO `result` at all —
-- which is exactly what site-work-orchestrator's own step does
-- (`{"action":"complete_work_item","config":{"work_item_id":"current_fix_item.id"}}`, no
-- result key) and which has fired for this item_type once already (1 of 26 completions
-- carries result = '{}'). So the probe reproduces a REAL production call shape.
--
-- SAFETY, by construction rather than by care:
--   * the workflow has ONE action step — no handler is spawned, no page, component or
--     site row is read or written;
--   * the synthetic items carry handler_agent 'proof-302-probe', which is not an active
--     agent definition, so detected-item-promoter's handler_ok can never select them AND
--     the real dark_section_audit/color-variable-fixer pair's success ratio (which this
--     lane measured at 26/4) is not polluted;
--   * status 'deferred' is writable by both completion UPDATEs (neither excludes it) but
--     is claimed by no dispatch loop (they claim 'triaged'/'approved') and promoted by
--     nothing (the promoter reads 'detected');
--   * max_attempts = 1, so a refusal terminalises at 'failed' immediately instead of
--     landing at 'triaged', where a dispatch loop could pick it up.

INSERT INTO agent_definitions (type, display_name, description, category, default_config, is_active)
VALUES (
  'proof-302-completion-gate-probe',
  'PROOF 302 completion-gate probe (TEMPORARY — delete after use)',
  'Single-step probe that calls complete_work_item on a given work_item_id, optionally with a supplied result. Exists only to exercise completion gate 1b in production for bugs_open/302. Spawns no handler and touches no site data.',
  'orchestrator',
  '{
    "workflow": {
      "start_step": "complete_item",
      "processing_mode": "orchestrator",
      "timeout_seconds": 120,
      "steps": {
        "complete_item": {
          "action": "complete_work_item",
          "description": "Call the completion gates with the supplied work_item_id and (optionally) result",
          "config": {
            "work_item_id!": "input_data.work_item_id",
            "result?": "input_data.result"
          },
          "output_field": "completion",
          "next_step": "done"
        },
        "done": {
          "action": "complete_workflow",
          "config": {"output_fields": ["completion"]},
          "description": "Return the completion outcome"
        }
      }
    }
  }'::jsonb,
  true
);

-- Four synthetic items. A/B/C are the opted-in type; D is the containment control.
-- NOTE: `summary` is NOT NULL with no default; the first attempt at this insert failed on it.
INSERT INTO site_work_items (site_id, source, item_type, item_key, handler_agent, status, max_attempts, created_by, summary, spec)
VALUES
  ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', 'proof-302', 'dark_section_audit', 'proof-302-A-unreadable',   'proof-302-probe', 'deferred', 1, 'proof-302', 'PROOF 302-A (temporary) — unreadable payload must be refused', '{"note":"PROOF A: no result supplied -> must be REFUSED handler_result_unreadable"}'),
  ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', 'proof-302', 'dark_section_audit', 'proof-302-B-readable-zero','proof-302-probe', 'deferred', 1, 'proof-302', 'PROOF 302-B (temporary) — readable zero must be refused as no_change', '{"note":"PROOF B: readable all-zero counters -> must be REFUSED handler_reported_no_change (the OLD arm)"}'),
  ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', 'proof-302', 'dark_section_audit', 'proof-302-C-nonzero',      'proof-302-probe', 'deferred', 1, 'proof-302', 'PROOF 302-C (temporary) — non-zero counter must complete', '{"note":"PROOF C: a non-zero counter -> must COMPLETE (proves the gate is not refusing everything)"}'),
  ('1368e337-dd1d-4799-bbb3-8221a1b79bcc', 'proof-302', 'spacing_fix',        'proof-302-D-containment',  'proof-302-probe', 'deferred', 1, 'proof-302', 'PROOF 302-D (temporary) — containment control, must complete', '{"note":"PROOF D: NOT on the roster, no result -> must COMPLETE (the 86 fleet empty-result completions are unaffected)"}');

SELECT id, item_type, item_key, status FROM site_work_items WHERE item_key LIKE 'proof-302-%' ORDER BY item_key;

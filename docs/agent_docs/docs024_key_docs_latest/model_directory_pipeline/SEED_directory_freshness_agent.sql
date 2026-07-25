-- SEED — directory-freshness agent, and repointing the freshness tasks at it
--
-- WHY THIS EXISTS: the freshness sweep has NEVER RUN. Not once, since Phase B.
--
-- `model-directory-freshness` (and today's `adoption-tracker-freshness`) put
-- their workflow INLINE, in scheduled_tasks.input_data.config.workflow, and
-- targeted `target_agent_type = 'generic'`. That shape is silently ignored:
-- the chassis runs the TARGET AGENT'S OWN workflow, and `generic`'s is a
-- single no-op step —
--
--   {"start_step":"complete","steps":{"complete":{"action":"complete_workflow",
--     "description":"No-op — scheduled task pre_query already did the work"}}}
--
-- so every fire completed instantly, stamped last_triggered_at AND
-- last_completed_at, and did nothing. Measured 2026-07-25:
--
--   -- orchestrations that ever carried the action:      0
--   SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE '%refresh_directory_claims%';
--   -- claims ever re-verified after registration:       0 of 108
--   SELECT count(*) FILTER (WHERE verified_at > created_at + interval '1 minute'), count(*) FROM directory_claims;
--
-- This was invisible because the happy path is indistinguishable from the
-- broken one: the task fires, reports success, and both timestamps advance.
-- It was caught only by INDUCING A FAULT — corrupting one stored quote and
-- watching for the flip to citation_lost that never came. A green scheduled
-- task is not evidence that its work happened.
--
-- The same shape is used by `evidence-freshness` (claims_verification
-- workstream) — flagged to them in bugs_open/074; not repointed here, it is
-- their agent and their cadence.
--
-- THE FIX: carry the workflow where the chassis actually reads it — in an
-- agent_definitions row — exactly as `directory-researcher` does, which is
-- the proven-working mechanism in this same pipeline. No image roll: the
-- action has shipped since Phase B and simply had no caller.
--
-- ONE TASK, ALL KINDS. refresh_directory_claims already treats an empty kind
-- as "sweep every kind" (directory_claims.go: "AND ($1 = '' OR de.kind = $1)").
-- Per-kind cadence was never the right knob anyway: each CLAIM carries its own
-- staleness_days, so the task cadence only decides how often we look for due
-- claims, not how often a claim is re-checked. Daily, all kinds.

BEGIN;

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status, is_active,
    image_repository, image_tag, input_contract, output_contract, default_config
)
SELECT
    'directory-freshness',
    'Directory Freshness Sweep',
    'Re-verifies is_current directory_claims whose staleness_days has elapsed: re-fetches each cited URL, re-checks the stored quote verbatim, and supersedes the row on any status transition (found/citation_lost/fetch_error). No LLM in this path — the model proposed once at registration; from here on a string comparison disposes.',
    'analyst', 'analyst', 'active', true,
    'docker.io/aqls/agent-chassis',
    (SELECT image_tag FROM agent_definitions WHERE type = 'claims-auditor'),
    '{"optional": ["kind"]}'::jsonb,
    '{"produces": {"refresh_result": "counts of claims re-verified, superseded, and flipped away from found"}}'::jsonb,
    $cfg${
  "workflow": {
    "start_step": "refresh_claims",
    "processing_mode": "orchestrator",
    "timeout_seconds": 600,
    "steps": {
      "refresh_claims": {
        "action": "refresh_directory_claims",
        "config": {"kind": "input_data.kind"},
        "next_step": "complete",
        "output_field": "refresh_result",
        "description": "Re-verify every due claim. kind resolves from input_data; absent or empty means sweep ALL kinds, which is the intended default."
      },
      "complete": {
        "action": "complete_workflow",
        "config": {"output_fields": ["refresh_result"]}
      }
    }
  }
}$cfg$::jsonb
WHERE NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'directory-freshness' AND deleted_at IS NULL);

-- Repoint the model-directory sweep at the agent that actually carries the
-- workflow, and strip the inline workflow that was never read. Daily, all kinds.
UPDATE scheduled_tasks
SET target_agent_type = 'directory-freshness',
    input_data = '{}'::jsonb,
    description = 'Directory register: re-verify is_current claims of EVERY kind whose staleness_days has elapsed; supersede on any status transition. Repointed 2026-07-25 from target_agent_type=generic + inline workflow, a shape the chassis silently ignored — the sweep had never run.',
    last_triggered_at = NULL,
    updated_at = now()
WHERE name = 'model-directory-freshness';

-- The adoption-specific sweep is redundant now that one task covers all kinds.
-- Disabled rather than deleted, so the row remains as evidence of the shape
-- that did not work.
UPDATE scheduled_tasks
SET enabled = false,
    description = 'SUPERSEDED 2026-07-25 by model-directory-freshness, which now sweeps every kind via the directory-freshness agent. Kept disabled as a record of the inline-workflow shape that the chassis silently ignored (bugs_open/074).',
    updated_at = now()
WHERE name = 'adoption-tracker-freshness';

SELECT name, target_agent_type, enabled, interval_seconds, input_data::text
FROM scheduled_tasks WHERE name IN ('model-directory-freshness','adoption-tracker-freshness') ORDER BY name;

COMMIT;

-- ── Post-apply verification — and this time verify the FAILING branch ───────
-- A green run proves nothing here; that is the whole lesson of this file.
--
-- 1) The run actually carried the action (this was 0 before the fix):
--      SELECT count(*) FROM orchestration_states WHERE workflow_plan::text LIKE '%refresh_directory_claims%';
-- 2) Claims are being re-verified (this was 0 of 108 before the fix):
--      SELECT count(*) FILTER (WHERE verified_at > created_at + interval '1 minute'), count(*) FROM directory_claims;
-- 3) INDUCE A FAULT and confirm the flip. Corrupt one quote to a sentence on
--    no page, backdate verified_at past its staleness, force the task, and
--    expect a NEW is_current row with status='citation_lost' superseding it:
--      CREATE TABLE IF NOT EXISTS _fault_injection_YYYYMMDD AS
--        SELECT id, citation, verified_at FROM directory_claims WHERE id='<claim>';
--      UPDATE directory_claims SET
--        citation = jsonb_set(citation,'{quote}','"a sentence that appears on no page"'),
--        verified_at = now() - interval '500 days' WHERE id='<claim>';
--      UPDATE scheduled_tasks SET last_triggered_at = NULL WHERE name='model-directory-freshness';
--    then restore from the backup table.

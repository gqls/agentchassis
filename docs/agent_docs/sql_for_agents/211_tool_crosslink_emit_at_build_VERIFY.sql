-- 211_tool_crosslink_emit_at_build_VERIFY.sql — read-only. Run by hand.
-- NOT a migration: the runner's SIDECAR_RE excludes UPPERCASE-suffixed files
-- (bugs_open/007), so this is never auto-applied.
--
-- Written 2026-07-26 answering the council gate's debug_historian objection on
-- round 1 of submission 745f9dfd: 211 carried snapshot + post-condition guard
-- but no COUNTED needle assertion and no separate verify/rollback artifacts.
-- The counted PRE-state is reconstructed here from the snapshots 211 itself
-- took — stronger than a count would have been, because it is the actual
-- before-image rather than a number someone wrote down.
--
-- TWO TRAPS, both met while writing this (see RUNBOOK R9):
--   1. snapshot_agent() writes to `agent_definitions_backup`, NOT to
--      agent_definitions with is_snapshot=true. Querying agent_definitions for
--      snapshot rows returns 0 and reads exactly like "the safety net was never
--      taken". It was.
--   2. 211 was applied TWICE (7s apart). The EARLIEST snapshot set is the true
--      pre-state; the later one is the post-first-run state. A rollback that
--      takes the newest restores the migration, not the pre-state.

\pset pager off

-- ── 1. The needle, COUNTED, from the earliest 211 snapshot set (pre-state) ──
-- Expect: 1 | 0 | 0 | 3
SELECT
  count(*) FILTER (WHERE default_config #> '{workflow,steps,create_cross_links}' IS NOT NULL)
    AS pre_had_create_cross_links,
  count(*) FILTER (WHERE default_config #>> '{workflow,steps,deploy_tool,config,related_pages}' IS NOT NULL)
    AS pre_had_deployer_related_pages,
  count(*) FILTER (WHERE default_config #>> '{workflow,steps,save_tool,config,related_pages}' IS NOT NULL)
    AS pre_had_generator_related_pages,
  count(*) AS snapshot_rows
FROM agent_definitions_backup
WHERE snapshot_reason LIKE '211_tool_crosslink_emit_at_build%'
  AND snapshot_taken_at = (
    SELECT min(snapshot_taken_at) FROM agent_definitions_backup
    WHERE snapshot_reason LIKE '211_tool_crosslink_emit_at_build%');

-- ── 2. The post-state, counted the same way ────────────────────────────────
-- Expect: 0 | 1 | 1 | 3
SELECT
  count(*) FILTER (WHERE default_config #> '{workflow,steps,create_cross_links}' IS NOT NULL)
    AS post_has_create_cross_links,
  count(*) FILTER (WHERE default_config #>> '{workflow,steps,deploy_tool,config,related_pages}'
                         = 'input_data.spec.related_pages') AS post_deployer_wired,
  count(*) FILTER (WHERE default_config #>> '{workflow,steps,save_tool,config,related_pages}'
                         = 'input_data.spec.related_pages') AS post_generator_wired,
  count(*) AS live_rows
FROM agent_definitions
WHERE type IN ('tool-suggester','tool-deployer','tool-generator')
  AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- ── 3. Blast radius: is create_cross_links referenced by ANY other agent? ───
-- Expect 0 rows. Searches the whole config as text, so it would still find a
-- reference from a different agent's step — which a path-specific check cannot
-- once the step has already been deleted from the one agent we know about.
SELECT type, is_active
FROM agent_definitions
WHERE default_config::text LIKE '%create_cross_links%'
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- ── 4. Same question for the ACTION name, not the step name ────────────────
-- create_tool_cross_link_items stays REGISTERED in the binary on purpose
-- (bugs_closed/017: an unregistered action named in config invalidates the
-- workflow). A live reference here is therefore not an error — it is what the
-- fail-safe rewrite exists to make harmless. Expect 0 anyway.
SELECT type, is_active
FROM agent_definitions
WHERE default_config::text LIKE '%create_tool_cross_link_items%'
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- ── 5. The behavioural check the whole fix is FOR ──────────────────────────
-- No cross-link item may be created from 2026-07-26 on without a real tool page
-- URL in its spec. Rows created before that are the pre-fix damage
-- (bugs_open/029 §existing damage) and correctly have spec_url NULL.
SELECT s.domain, swi.created_at::date, swi.source, swi.status,
       swi.spec->>'tool_function'  AS tool_function,
       swi.spec->>'tool_page_url'  AS spec_url,
       p.url                       AS matched_page_url,
       p.build_status,
       swi.depends_on IS NOT NULL  AS gated
FROM site_work_items swi
JOIN sites s ON s.id = swi.site_id
LEFT JOIN pages p ON p.site_id = swi.site_id AND p.url = swi.spec->>'tool_page_url'
WHERE swi.item_key LIKE 'tool_crosslink:%'
ORDER BY swi.created_at DESC;

-- ── 6. Declined emits are COUNTABLE (the durable-record half of the fix) ───
-- Empty is fine; non-empty tells you which tools got no cross-links and why,
-- which is the thing a log line could never answer.
SELECT occurred_at, error_code, severity, left(error_message, 140) AS message
FROM agent_error_log
WHERE error_code LIKE 'tool_crosslink_not_emitted%'
ORDER BY occurred_at DESC LIMIT 20;

-- 414_content_gap_planner_cache_breakpoint_ROLLBACK.sql
--
-- Removes the LCO-008 cache breakpoint from content-gap-planner's `plan_gaps`
-- template, returning the agent to sending its prompt as one uncached block.
--
-- WHEN TO RUN THIS. The forward migration is a pure cost change — it cannot
-- alter what the model is asked, because the client strips the marker before
-- the request is sent (platform/aiservice/anthropic.go: the marker is plumbing,
-- never content). So the only reasons to roll back are economic or a 400:
--
--   1. Reads stay ZERO on the second and later calls. That means the prefix is
--      not matching, and every call is paying the 2x write premium for nothing.
--      Check the reads query in the forward file BEFORE concluding this — and
--      note that a single site seen once legitimately shows a write and no read.
--   2. Calls start returning 400. Confirm it is this agent and not the fleet:
--      if the API had stopped accepting the ttl field, EVERY marked agent would
--      be failing, not one. See the cacheTTL comment in anthropic.go.
--
-- Rolling back this file does NOT revert the 1h TTL, which is a Go constant
-- shipped in the binary (v1.0.1301). The two halves are independent: this file
-- controls whether content-gap-planner caches at all; the constant controls how
-- long any marked agent's entry lives.
--
-- This reverses by deleting the inserted substring rather than restoring from
-- the snapshot, because the forward transformation is a single unambiguous
-- insertion with a guard proving it occurred exactly once — deleting it is
-- exact. If the template has been edited by another lane since, prefer the
-- snapshot taken by the forward file:
--
--   SELECT default_config FROM agent_definitions_snapshots
--   WHERE agent_type='content-gap-planner' ORDER BY created_at DESC LIMIT 5;

BEGIN;

DO $$
DECLARE
    n_marker integer;
    tmpl     text;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_gaps'->'config'->>'prompt_template'
      INTO tmpl
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF tmpl IS NULL THEN
        RAISE EXCEPTION 'ROLLBACK 414: plan_gaps has no prompt_template — nothing to reverse';
    END IF;

    n_marker := (length(tmpl) - length(replace(tmpl, '<!--CACHE_BREAKPOINT-->', ''))) / length('<!--CACHE_BREAKPOINT-->');
    IF n_marker < 1 THEN
        RAISE EXCEPTION 'ROLLBACK 414: no cache marker present — already rolled back, or never applied';
    END IF;

    RAISE NOTICE 'rollback 414: removing % marker(s)', n_marker;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_gaps,config,prompt_template}',
        to_jsonb(
            replace(
                default_config->'workflow'->'steps'->'plan_gaps'->'config'->>'prompt_template',
                E'<!--CACHE_BREAKPOINT-->\n',
                ''
            )
        )
    ),
    version    = version + 1,
    updated_at = now()
WHERE type='content-gap-planner' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
    tmpl     text;
    n_marker integer;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_gaps'->'config'->>'prompt_template'
      INTO tmpl
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    n_marker := (length(tmpl) - length(replace(tmpl, '<!--CACHE_BREAKPOINT-->', ''))) / length('<!--CACHE_BREAKPOINT-->');
    IF n_marker <> 0 THEN
        RAISE EXCEPTION 'ROLLBACK 414: % marker(s) remain after removal', n_marker;
    END IF;

    -- The anchor must survive: this reverses a marker insertion, not the block.
    IF tmpl NOT LIKE '%## Content Gap to Address%'
       OR tmpl NOT LIKE '%{{.input_data.spec.description}}%' THEN
        RAISE EXCEPTION 'ROLLBACK 414: the removal damaged the template — restore from the snapshot instead';
    END IF;

    RAISE NOTICE 'rollback 413 OK: marker removed, template intact';
END $$;

COMMIT;

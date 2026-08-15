-- 416_auditor_prompts_drop_dead_work_item_type_ROLLBACK.sql
-- Hand-run sidecar (uppercase suffix: excluded from run-migrations.sh --apply).
--
-- Re-inserts the two `work_item_type` prompt fragments that migration 416
-- removed, at their original anchors. Before running, be sure you actually want
-- this: the field was demanded of the LLM and read by NOTHING
-- (bugs_open/279 leg 1), so restoring it restores a token cost and a false
-- affordance, not behaviour. There is no Go-side ordering constraint in either
-- direction.

BEGIN;

DO $$
DECLARE
    occ integer;
    a1 integer; a2 integer; a3 integer; a4 integer;
BEGIN
    SELECT count(*) INTO occ
    FROM agent_definitions
    WHERE type IN ('site-review-agent', 'content-quality-auditor')
      AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND default_config::text LIKE '%work_item_type%';
    IF occ <> 0 THEN
        RAISE EXCEPTION 'ROLLBACK 416: % definition(s) already carry work_item_type — 416 is not applied (or a later migration reintroduced the field)', occ;
    END IF;

    -- Each anchor must be unique in its own config or the insert lands twice/nowhere.
    SELECT (length(default_config::text) - length(replace(default_config::text, '\"page\":\"which page (or site-wide)\"', ''))) / length('\"page\":\"which page (or site-wide)\"') INTO a1
      FROM agent_definitions WHERE type='site-review-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    SELECT (length(default_config::text) - length(replace(default_config::text, '\"page\":\"pricing\"', ''))) / length('\"page\":\"pricing\"') INTO a2
      FROM agent_definitions WHERE type='site-review-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    SELECT (length(default_config::text) - length(replace(default_config::text, '\"page\":\"which page\"', ''))) / length('\"page\":\"which page\"') INTO a3
      FROM agent_definitions WHERE type='content-quality-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    SELECT (length(default_config::text) - length(replace(default_config::text, '\"page\":\"about\"', ''))) / length('\"page\":\"about\"') INTO a4
      FROM agent_definitions WHERE type='content-quality-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF a1 <> 1 OR a2 <> 1 OR a3 <> 1 OR a4 <> 1 THEN
        RAISE EXCEPTION 'ROLLBACK 416: anchors not unique (sr-schema=% sr-example=% cq-schema=% cq-example=%) — the prompts have drifted; re-derive the anchors', a1, a2, a3, a4;
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = replace(replace(default_config::text,
        '\"page\":\"which page (or site-wide)\"',
        '\"page\":\"which page (or site-wide)\",\"work_item_type\":\"content_rewrite|needs_content_page|tone_shift|cta_improvement|nav_restructure\"'),
        '\"page\":\"pricing\"',
        '\"page\":\"pricing\",\"work_item_type\":\"needs_content_page\"')::jsonb,
    version    = version + 1,
    updated_at = now()
WHERE type='site-review-agent'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = replace(replace(default_config::text,
        '\"page\":\"which page\"',
        '\"page\":\"which page\",\"work_item_type\":\"content_rewrite|needs_content_page|tone_shift|cta_improvement\"'),
        '\"page\":\"about\"',
        '\"page\":\"about\",\"work_item_type\":\"content_rewrite\"')::jsonb,
    version    = version + 1,
    updated_at = now()
WHERE type='content-quality-auditor'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
    occ_sr integer;
    occ_cq integer;
BEGIN
    SELECT (length(default_config::text) - length(replace(default_config::text,'work_item_type','')))/length('work_item_type') INTO occ_sr
      FROM agent_definitions WHERE type='site-review-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    SELECT (length(default_config::text) - length(replace(default_config::text,'work_item_type','')))/length('work_item_type') INTO occ_cq
      FROM agent_definitions WHERE type='content-quality-auditor' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF occ_sr <> 2 OR occ_cq <> 2 THEN
        RAISE EXCEPTION 'ROLLBACK 416: expected 2 occurrences restored per prompt, got site-review=% content-quality=%', occ_sr, occ_cq;
    END IF;
    RAISE NOTICE 'rollback 416 OK: both prompts carry work_item_type again (2 + 2)';
END $$;

COMMIT;

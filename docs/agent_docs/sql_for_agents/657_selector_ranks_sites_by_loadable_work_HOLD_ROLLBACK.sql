-- 657_selector_ranks_sites_by_loadable_work_HOLD_ROLLBACK.sql
--
-- Restores build-pipeline-trigger > find_dispatchable_site > config.query to the exact
-- text 657 replaced (the 633-era text, md5 d6f98acdb5aec385d5eb4077eac530fc): rank sites
-- by the single globally-oldest eligible row. Rolling back REINSTATES bugs_open/413's
-- pin mechanism — do it only if the new ordering itself misbehaves, and say so in 413.
--
-- Apply by hand:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
--     < docs/agent_docs/sql_for_agents/657_selector_ranks_sites_by_loadable_work_HOLD_ROLLBACK.sql

BEGIN;

SELECT snapshot_agent('build-pipeline-trigger', '657_selector_ranks_sites_by_loadable_work_HOLD_ROLLBACK.sql: pre-rollback');

DO $mig$
DECLARE
    v_old_md5 CONSTANT text := 'd6f98acdb5aec385d5eb4077eac530fc';
    v_old     CONSTANT text := $q$SELECT wi.site_id::text, s.domain FROM site_work_items wi JOIN sites s ON s.id = wi.site_id WHERE (s.locked_at IS NULL OR wi.id = ANY(COALESCE(s.lock_except_item_ids, ARRAY[]::uuid[]))) AND wi.status IN ('triaged', 'approved') AND wi.attempt_count < wi.max_attempts AND (wi.retry_after IS NULL OR wi.retry_after <= NOW()) AND (COALESCE(wi.approval_mode, 'auto') = 'auto' OR wi.status = 'approved') AND (wi.depends_on IS NULL OR NOT EXISTS (SELECT 1 FROM unnest(wi.depends_on) dep_id WHERE dep_id NOT IN (SELECT id FROM site_work_items WHERE site_id = wi.site_id AND status IN ('complete', 'verified')))) AND NOT EXISTS (SELECT 1 FROM site_work_items active WHERE active.site_id = wi.site_id AND active.status = 'claimed') ORDER BY wi.created_at ASC, wi.priority ASC, wi.id ASC LIMIT 1$q$;
    v_q text;
    v_n int;
BEGIN
    IF md5(v_old) <> v_old_md5 THEN
        RAISE EXCEPTION '657 ROLLBACK: this file''s own embedded text does not md5 to % — transcription damage; STOP', v_old_md5;
    END IF;

    SELECT s.value->'config'->>'query' INTO v_q
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-pipeline-trigger' AND s.key='find_dispatchable_site'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;

    IF v_q IS NULL THEN
        RAISE EXCEPTION '657 ROLLBACK: find_dispatchable_site has no config.query — step shape changed; STOP';
    END IF;
    IF md5(v_q) = v_old_md5 THEN
        RAISE EXCEPTION '657 ROLLBACK: already rolled back (live text is the pre-657 text)';
    END IF;
    IF position('load_rank' in v_q) = 0 THEN
        RAISE EXCEPTION '657 ROLLBACK: live text is neither 657''s nor the pre-657 text (md5 %) — a later migration has moved it; STOP and re-derive', md5(v_q);
    END IF;

    UPDATE agent_definitions ad
       SET default_config = jsonb_set(
             ad.default_config,
             '{workflow,steps,find_dispatchable_site,config,query}',
             to_jsonb(v_old),
             false)
     WHERE ad.type='build-pipeline-trigger'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 1 THEN
        RAISE EXCEPTION '657 ROLLBACK: expected to update exactly 1 row, updated %', v_n;
    END IF;
END;
$mig$;

DO $guard$
DECLARE
    v_q      text;
    v_site   text;
    v_domain text;
BEGIN
    SELECT s.value->'config'->>'query' INTO v_q
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.type='build-pipeline-trigger' AND s.key='find_dispatchable_site'
       AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
    IF md5(COALESCE(v_q,'')) <> 'd6f98acdb5aec385d5eb4077eac530fc' THEN
        RAISE EXCEPTION '657 ROLLBACK GUARD: restored text does not md5 to the pre-657 value (got %)', md5(COALESCE(v_q,''));
    END IF;
    EXECUTE v_q INTO v_site, v_domain;
    RAISE NOTICE '657 ROLLBACK OK: pre-657 selector restored; next pick = % (%)',
        COALESCE(v_domain, '<no eligible site>'), COALESCE(v_site, '-');
END;
$guard$;

COMMIT;

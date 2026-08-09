-- 359_domain_strategist_gate_uses_shipped_predicate_ROLLBACK.sql
--
-- Inverse of 359: restores the 341 predicate (`build_status = 'deployed'` bare)
-- in domain-strategist's check_site_deployed step.
--
-- ⚠ THE 341 PREDICATE POINTS THE DANGEROUS WAY: a site whose shipped pages are
-- all flagged needs_rebuild reads is_deployed = false under it, and the gate
-- then CHAINS A RE-PLAN of a serving site (the defect 359 exists to close;
-- 11 of 39 needs_rebuild pages answered HTTP 200 when measured 08-09). Roll
-- back only if 359 itself is shown to misfire, and say so where the decision
-- is recorded. It also re-diverges the gate from B3's Go checks, which use
-- datahelpers.PageHasShippedPredicateFor.
--
-- Written 2026-08-09 (round-2 council objection: needle-gate discipline wants a
-- separate rollback file, not just a fail-closed forward file).

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('domain-strategist',
    'pre-update: 359_ROLLBACK — gate predicate reverts to the 341 bare build_status form');

DO $pre$
DECLARE v_q text;
BEGIN
    SELECT default_config #>> '{workflow,steps,check_site_deployed,config,query}' INTO v_q
    FROM agent_definitions
    WHERE type = 'domain-strategist'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_q IS DISTINCT FROM
       'SELECT (COUNT(*) > 0) AS is_deployed FROM pages WHERE site_id = $1 AND NOT (deployed_at IS NULL AND COALESCE(build_status, '''') <> ''deployed'')' THEN
        RAISE EXCEPTION '359_ROLLBACK: pre-state mismatch — check_site_deployed.query is not the 359 string (%). Recompose.', v_q;
    END IF;
END $pre$;

UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,check_site_deployed,config,query}',
    to_jsonb($q$SELECT (COUNT(*) > 0) AS is_deployed FROM pages WHERE site_id = $1 AND build_status = 'deployed'$q$::text)
)
WHERE type = 'domain-strategist'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $post$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type = 'domain-strategist'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,check_site_deployed,config,query}'
          LIKE '%build_status = ''deployed''%'
      AND default_config #>> '{workflow,steps,check_site_deployed,config,query}'
          NOT LIKE '%deployed_at IS NULL%';
    IF n <> 1 THEN
        RAISE EXCEPTION '359_ROLLBACK: post-state mismatch (% rows)', n;
    END IF;
END $post$;

COMMIT;

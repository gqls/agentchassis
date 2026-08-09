-- 359_domain_strategist_gate_uses_shipped_predicate.sql
--
-- Corrects B2's refresh gate (migration 341, slug domain_strategist_refresh_
-- safe…, applied 2026-08-08): its check_site_deployed step used
-- `build_status = 'deployed'` bare, which is not "is this page live" — a
-- needs_rebuild page HAS deployed and is still serving its previous artefact
-- (bugs_closed/037: 35 of 46 such rows carry deployed_at). The commit-time
-- pattern check caught B3's checks re-typing the same predicate by hand
-- (bugs_open/185 class); this file fixes the LIVE config instance the same
-- session, because here the miss points in the DANGEROUS direction: a site
-- whose shipped pages are all flagged needs_rebuild would read is_deployed =
-- false and the gate would CHAIN A RE-PLAN of a serving site — precisely what
-- the gate exists to prevent.
--
-- The replacement is the SQL PageHasShippedPredicateFor("") renders:
--   NOT (deployed_at IS NULL AND COALESCE(build_status, '') <> 'deployed')
-- so the Go checks (B3) and this gate agree by construction.
--
-- Verified before writing (2026-08-09): the live step's query is exactly the
-- 341 string below; loancalculator.co.uk (the B2 witness) answers is_deployed
-- = true under BOTH predicates, so the witnessed behaviour is unchanged.
--
-- ROLLBACK: inverse replace, same shape.

\set ON_ERROR_STOP on

BEGIN;

SELECT snapshot_agent('domain-strategist',
    'pre-update: 359 — gate predicate widens to the shipped predicate (needs_rebuild pages still serve)');

DO $pre$
DECLARE v_q text;
BEGIN
    SELECT default_config #>> '{workflow,steps,check_site_deployed,config,query}' INTO v_q
    FROM agent_definitions
    WHERE type = 'domain-strategist'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_q IS DISTINCT FROM
       'SELECT (COUNT(*) > 0) AS is_deployed FROM pages WHERE site_id = $1 AND build_status = ''deployed''' THEN
        RAISE EXCEPTION '359: pre-state mismatch — check_site_deployed.query is not the 341 string (%). Recompose.', v_q;
    END IF;
END $pre$;

UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,check_site_deployed,config,query}',
    to_jsonb($q$SELECT (COUNT(*) > 0) AS is_deployed FROM pages WHERE site_id = $1 AND NOT (deployed_at IS NULL AND COALESCE(build_status, '') <> 'deployed')$q$::text)
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
          LIKE '%NOT (deployed_at IS NULL%';
    IF n <> 1 THEN
        RAISE EXCEPTION '359: post-state mismatch (% rows)', n;
    END IF;
END $post$;

COMMIT;

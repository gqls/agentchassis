-- 733: the Default Network's analytics container default — the ONE place seedAnalyticsDefault
-- (platform/orchestration/actions/seed_analytics_default.go, inert until the next roll) reads at
-- site creation, and the value the customer-ZIP export path compares against ("is this the owner
-- default?"). bugs_open/397 §6.2; owner 2026-08-24 ("standard for new builds") + 2026-08-26
-- (customer sites use the SEPARATE empty container GTM-TH5XGNQ4 — that one is deliberately NOT
-- set here; a network with no value seeds nothing).
UPDATE networks
   SET settings = COALESCE(settings, '{}'::jsonb)
              || jsonb_build_object('analytics',
                   COALESCE(settings->'analytics', '{}'::jsonb)
                   || '{"gtm_container_id": "GTM-PQ3WCTBD"}'::jsonb),
       updated_at = now()
 WHERE id = '00000000-0000-0000-0000-000000000002';

DO $v$
DECLARE v text;
BEGIN
  SELECT settings->'analytics'->>'gtm_container_id' INTO v
    FROM networks WHERE id = '00000000-0000-0000-0000-000000000002';
  IF v IS DISTINCT FROM 'GTM-PQ3WCTBD' THEN
    RAISE EXCEPTION 'verify: Default Network analytics default is %, want GTM-PQ3WCTBD', COALESCE(v,'<null>');
  END IF;
END $v$;

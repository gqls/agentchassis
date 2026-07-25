-- p4_12_lock_funding_guides.sql — lock the two authored funding guides (p4_10/p4_11).
--
-- Same rationale as p4_03 (authored high-stakes content, V5 inert), same correction honoured
-- as p4_08: ONLY authored sections, never a deriving one. Neither funding guide has any
-- query.* sourced field (hero + Generic Text Block + call-to-action, all content_data-driven),
-- and the guard below verifies that from the schema rather than trusting this comment.
--
-- RUN ONLY AFTER both pages are verified live by curl.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     AND url IN ('/guides/funding-ways/index.html','/guides/funding-sources/index.html')
     AND build_status = 'deployed';
  IF n <> 2 THEN
    RAISE EXCEPTION 'ABORT: expected both funding guides deployed, found %.', n;
  END IF;

  -- p4_08 rule, enforced not assumed: refuse if any section we are about to lock derives.
  SELECT count(*) INTO n
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url IN ('/guides/funding-ways/index.html','/guides/funding-sources/index.html')
    AND EXISTS (SELECT 1 FROM jsonb_each(cc.input_schema->'fields') AS f(k,v)
                WHERE v->>'source' LIKE 'query.%');
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: % section(s) have query-sourced fields — locking would freeze a derivation (p4_08 rule).', n;
  END IF;
END
$guard$;

UPDATE page_components pc
SET locked_at = now(),
    locked_by = 'idea.uk ideas-pipeline (features_open/014, p4_10/p4_11) — hand-authored UK funding guidance, no figures by policy, do not regenerate',
    lock_type = 'permanent'
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url IN ('/guides/funding-ways/index.html','/guides/funding-sources/index.html');

COMMIT;

SELECT p.url, pc.position, pc.slot_name, pc.lock_type
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url IN ('/guides/funding-ways/index.html','/guides/funding-sources/index.html')
ORDER BY p.url, pc.position;

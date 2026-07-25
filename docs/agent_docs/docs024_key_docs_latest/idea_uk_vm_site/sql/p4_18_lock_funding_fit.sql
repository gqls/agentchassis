-- p4_18_lock_funding_fit.sql — lock the funding-fit tool page (p4_17) + re-render /tools.html.
-- RUN ONLY AFTER /tools/funding-fit/index.html is verified live by curl.
-- Same p4_08 derive-guard as every lock since the guides-hub near-miss.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     AND url = '/tools/funding-fit/index.html' AND build_status = 'deployed';
  IF n <> 1 THEN
    RAISE EXCEPTION 'ABORT: funding-fit page not deployed — verify live first.';
  END IF;

  SELECT count(*) INTO n
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url = '/tools/funding-fit/index.html'
    AND EXISTS (SELECT 1 FROM jsonb_each(cc.input_schema->'fields') AS f(k,v)
                WHERE v->>'source' LIKE 'query.%');
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: % section(s) derive — locking would freeze a derivation (p4_08 rule).', n;
  END IF;
END
$guard$;

UPDATE page_components pc
SET locked_at = now(),
    locked_by = 'idea.uk ideas-pipeline (features_open/014, p4_17) — gated funding-fit tool, authored copy + gating logic, do not regenerate',
    lock_type = 'permanent'
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/tools/funding-fit/index.html';

COMMIT;

SELECT p.url, pc.slot_name, pc.lock_type
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/tools/funding-fit/index.html' ORDER BY pc.position;

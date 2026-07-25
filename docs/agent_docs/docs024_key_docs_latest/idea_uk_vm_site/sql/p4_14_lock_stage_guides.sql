-- p4_14_lock_stage_guides.sql — lock the five stage guides (p4_13) once verified live.
-- Same p4_08 rule, enforced by the same guard: only authored sections, never a deriving one.

\set ON_ERROR_STOP on

BEGIN;

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM pages
   WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
     AND url IN ('/guides/creating-ideas/index.html','/guides/building-it/index.html',
                 '/guides/testing-it/index.html','/guides/user-acceptance/index.html',
                 '/guides/feedback-loops/index.html')
     AND build_status = 'deployed';
  IF n <> 5 THEN
    RAISE EXCEPTION 'ABORT: expected all 5 stage guides deployed, found %. Verify live (curl) first.', n;
  END IF;

  SELECT count(*) INTO n
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN content_components cc ON cc.id = pc.component_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url IN ('/guides/creating-ideas/index.html','/guides/building-it/index.html',
                  '/guides/testing-it/index.html','/guides/user-acceptance/index.html',
                  '/guides/feedback-loops/index.html')
    AND EXISTS (SELECT 1 FROM jsonb_each(cc.input_schema->'fields') AS f(k,v)
                WHERE v->>'source' LIKE 'query.%');
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: % section(s) derive (query.* source) — locking would freeze a derivation (p4_08 rule).', n;
  END IF;
END
$guard$;

UPDATE page_components pc
SET locked_at = now(),
    locked_by = 'idea.uk ideas-pipeline (features_open/014, p4_13 stages 1-5) — authored method guides, do not regenerate',
    lock_type = 'permanent'
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url IN ('/guides/creating-ideas/index.html','/guides/building-it/index.html',
                '/guides/testing-it/index.html','/guides/user-acceptance/index.html',
                '/guides/feedback-loops/index.html');

COMMIT;

SELECT p.url, count(*) AS sections, count(*) FILTER (WHERE pc.lock_type='permanent') AS locked
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND p.page_type='guide'
GROUP BY p.url ORDER BY p.url;

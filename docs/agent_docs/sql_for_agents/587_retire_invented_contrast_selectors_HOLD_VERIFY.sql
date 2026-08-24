-- VERIFY for 587 (bugs_open/352). Read-only. Run arms 1–3 BEFORE applying and
-- arms 4–5 after.
--
-- Arm 2 is the one that matters before you apply: it is the eyeball pass on the
-- exact rows about to be withdrawn, and arm 3 is its control. A zero from arm 3
-- with no control could not have come out otherwise, which is why the control is
-- part of the file rather than a note in someone's scrollback.

\echo '=== 1. WHAT WILL BE WITHDRAWN — count by status and site ==='
SELECT status, count(*) AS rows, count(DISTINCT site_id) AS sites
  FROM site_work_items
 WHERE item_type = 'contrast_failure'
   AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled')
   AND item_key ~ '#([A-Z][A-Z0-9]*)\.\1$'
 GROUP BY status ORDER BY rows DESC;

\echo ''
\echo '=== 2. THE EYEBALL PASS — every DISTINCT key about to be cancelled ==='
\echo '    Anything here that is NOT of the form <path>#TAG.TAG, with the two'
\echo '    halves identical and uppercase, means the predicate is over-selecting.'
\echo '    STOP if you see one.'
SELECT DISTINCT item_key, count(*) OVER (PARTITION BY item_key) AS rows
  FROM site_work_items
 WHERE item_type = 'contrast_failure'
   AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled')
   AND item_key ~ '#([A-Z][A-Z0-9]*)\.\1$'
 ORDER BY item_key;

\echo ''
\echo '=== 3. THE FALSE-POSITIVE TEST AND ITS POSITIVE CONTROL ==='
\echo '    Does any affected site GENUINELY use class="H3" on an <h3>? If one did,'
\echo '    its finding would be real and fixable and must not be withdrawn.'
\echo '    EXPECT: tag_tokens_found_in_markup = 0, real_class_tokens_found > 0.'
\echo '    If the control is also 0 the test is BLIND, not clean — do not apply.'
WITH html AS (
    SELECT p.site_id, pc.rendered_html AS h
      FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE pc.rendered_html IS NOT NULL
    UNION ALL
    SELECT sc.site_id, sc.rendered_html FROM site_components sc WHERE sc.rendered_html IS NOT NULL
),
tag_tokens AS (
    SELECT DISTINCT site_id, split_part(spec->>'selector', '.', 2) AS tok
      FROM site_work_items
     WHERE item_type = 'contrast_failure'
       AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled')
       AND spec->>'selector' ~ '^([A-Za-z0-9]+)\.\1$'
),
real_tokens AS (
    SELECT DISTINCT site_id, split_part(spec->>'selector', '.', 2) AS tok
      FROM site_work_items
     WHERE item_type = 'contrast_failure'
       AND spec->>'selector' ~ '^[A-Za-z0-9]+\.'
       AND NOT (spec->>'selector' ~ '^([A-Za-z0-9]+)\.\1$')
)
SELECT
  (SELECT count(*) FROM tag_tokens t
    WHERE EXISTS (SELECT 1 FROM html WHERE html.site_id = t.site_id
                    AND html.h ~ ('class="[^"]*\m' || t.tok || '\M')))  AS tag_tokens_found_in_markup,
  (SELECT count(*) FROM tag_tokens)                                     AS tag_tokens_tested,
  (SELECT count(*) FROM real_tokens r
    WHERE EXISTS (SELECT 1 FROM html WHERE html.site_id = r.site_id
                    AND html.h ~ ('class="[^"]*\m' || r.tok || '\M')))  AS real_class_tokens_found,
  (SELECT count(*) FROM real_tokens)                                    AS real_class_tokens_tested;

\echo ''
\echo '=== 4. AFTER APPLYING — the withdrawal landed and asserts the right thing ==='
\echo '    EXPECT: open_invented = 0, and every withdrawn row carrying its prior status.'
SELECT
  (SELECT count(*) FROM site_work_items
    WHERE item_type='contrast_failure'
      AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled')
      AND item_key ~ '#([A-Z][A-Z0-9]*)\.\1$')                                AS open_invented,
  (SELECT count(*) FROM site_work_items
    WHERE result->>'cancelled_by' = 'migration_587')                          AS withdrawn,
  (SELECT count(*) FROM site_work_items
    WHERE result->>'cancelled_by' = 'migration_587'
      AND result->>'pre_352_status' IS NULL)                                  AS withdrawn_without_prior_status;

\echo ''
\echo '=== 5. THE CONTROL THAT MATTERS MOST — no row was CLOSED AS FIXED ==='
\echo '    EXPECT 0. A non-zero means the retraction path reached these rows'
\echo '    before the withdrawal did, i.e. it minted a false completion — which'
\echo '    is the outcome the alias-key guard exists to prevent. If this is not'
\echo '    zero, the code fix is not live on both images and 587 was applied early.'
SELECT count(*) AS falsely_completed
  FROM site_work_items
 WHERE item_type = 'contrast_failure'
   AND item_key ~ '#([A-Z][A-Z0-9]*)\.\1$'
   AND status IN ('complete','verified')
   AND result ? 'resolved_at'
   AND (result->>'resolved_at')::timestamptz > '2026-08-24'::timestamptz;

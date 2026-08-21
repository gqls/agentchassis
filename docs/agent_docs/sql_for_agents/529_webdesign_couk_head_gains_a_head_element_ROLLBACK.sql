-- 529_webdesign_couk_head_gains_a_head_element_ROLLBACK.sql
-- Reverses 529: unwraps the <head> element and drops the lang schema entry,
-- restoring the exact pre-2026-08-21 bytes (md5 2d2bc47586604c200991b20b364dfcbc).
--
-- The guard below asserts that md5 AFTER the unwrap, so this refuses to leave the
-- row in any state other than the original. If it raises, re-read the live row and
-- repair by hand rather than re-running — the template is hand-authored and serves
-- 117 pages.
--
-- Nothing served changes until the site's chrome re-renders.

BEGIN;

UPDATE content_components SET
  html_template = regexp_replace(
                    regexp_replace(html_template, '^<head[^>]*>' || E'\n', ''),
                    E'\n' || '</head>$', ''),
  input_schema  = jsonb_set(input_schema, '{fields}', (input_schema -> 'fields') - 'lang'),
  updated_at = now()
WHERE id = '14cf6193-c8f0-4640-9cf1-f8b5347e6885'
  AND html_template LIKE '<head%'
  AND html_template LIKE '%</head>';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
  WHERE id = '14cf6193-c8f0-4640-9cf1-f8b5347e6885'
    AND md5(html_template) = '2d2bc47586604c200991b20b364dfcbc'
    AND NOT (input_schema #> '{fields}' ? 'lang');
  IF n <> 1 THEN
    RAISE EXCEPTION 'rollback did not restore the exact pre-529 bytes (md5 mismatch) — do NOT re-run; re-read the row and repair by hand';
  END IF;
END $$;

COMMIT;

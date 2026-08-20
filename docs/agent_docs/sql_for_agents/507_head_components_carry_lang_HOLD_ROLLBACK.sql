-- 507_head_components_carry_lang_HOLD_ROLLBACK.sql
-- Reverses 507: removes the lang gate from both shared head templates, drops the
-- `lang` input_schema entries, and RESTORES head-seo-standard's two og lines.
--
-- Guards are md5s of the POST-507 state, so this refuses to run against a row
-- another change has since touched. If it refuses, re-read the live row rather
-- than loosening the guard — a head template serves up to 18 sites.
--
-- ORDER, if rolling back both: run 508's rollback FIRST (remove the config), then
-- this one (remove the carrier). The reverse leaves sites holding a
-- config.locale.lang value nothing reads, which is inert but misleading.
--
-- After this, the next chrome re-render restores the previous stored artefact,
-- and assemblePage's Go default keeps emitting `<html lang="en">` — so the
-- served output returns to the pre-2026-08-20 bytes either way.

BEGIN;

-- A. Document Head (116c5f91…, 18 sites, FLAT schema).
UPDATE content_components SET
  html_template = replace(html_template,
                          '<head{{if .lang}} lang="{{.lang}}"{{end}}>',
                          '<head>'),
  input_schema  = input_schema - 'lang',
  updated_at = now()
WHERE id = '116c5f91-bc0d-439d-9e13-a3ba2d145571'
  AND html_template LIKE '<head{{if .lang}} lang="{{.lang}}"{{end}}>%';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
  WHERE id = '116c5f91-bc0d-439d-9e13-a3ba2d145571'
    AND md5(html_template) = '04d7d9cbcc8adb71d8579f07c45d3f7d'
    AND NOT (input_schema ? 'lang');
  IF n <> 1 THEN
    RAISE EXCEPTION 'Document Head rollback did not restore the 2026-08-19 bytes — the row has been changed by something else since 507. Re-read it.';
  END IF;
END $$;

-- B. head-seo-standard (aec98dbe…, 4 sites, WRAPPED schema). The og lines go
--    back in their original positions: og:title and og:description immediately
--    after the `<!-- Open Graph -->` comment and before og:type.
UPDATE content_components SET
  html_template = replace(
                    replace(html_template,
                            '<head{{if .lang}} lang="{{.lang}}"{{end}}>',
                            '<head>'),
                    E'    <meta property="og:type" content="{{if .og_type}}{{.og_type}}{{else}}website{{end}}">\n',
                    E'    <meta property="og:title" content="{{if .og_title}}{{.og_title}}{{else}}{{.title}}{{end}}">\n    <meta property="og:description" content="{{if .og_description}}{{.og_description}}{{else}}{{.description}}{{end}}">\n    <meta property="og:type" content="{{if .og_type}}{{.og_type}}{{else}}website{{end}}">\n'),
  input_schema  = jsonb_set(input_schema, '{fields}', (input_schema -> 'fields') - 'lang'),
  updated_at = now()
WHERE id = 'aec98dbe-76b7-4e13-9641-e5b6ba2502aa'
  AND html_template LIKE '<head{{if .lang}} lang="{{.lang}}"{{end}}>%';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
  WHERE id = 'aec98dbe-76b7-4e13-9641-e5b6ba2502aa'
    AND md5(html_template) = '3c2d8719629762f6bb8947c645f5c3df'
    AND NOT (input_schema #> '{fields}' ? 'lang');
  IF n <> 1 THEN
    RAISE EXCEPTION 'head-seo-standard rollback did not restore the 2026-08-19 bytes exactly (md5 mismatch) — most likely the og lines went back in the wrong position. Re-read the row and repair by hand rather than re-running this.';
  END IF;
END $$;

-- C. No head template carries the gate any more.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components WHERE html_template LIKE '%<head{{if .lang}}%';
  IF n <> 0 THEN
    RAISE EXCEPTION 'a head template still carries the lang gate after rollback (% rows)', n;
  END IF;
END $$;

COMMIT;

-- 529_webdesign_couk_head_gains_a_head_element.sql
-- bugs_open/347 — webdesign.co.uk's head component is a bare FRAGMENT: it has no
-- <head> open tag and no </head> close tag. Owner instruction 2026-08-21: fix it live.
--
-- WHY IT MATTERS, and it is not cosmetic. 117 assembled pages — the most of any site
-- in the fleet. Every per-page head helper anchors on `</head>`
-- (injectCanonicalLink, injectPageJSONLD, injectRobotsNoindex, injectComponentCSS,
-- spliceOpenGraph — all `strings.LastIndex(head, "</head>")`), and each carries its
-- own private fallback for the missing marker. They do not all agree: most append,
-- which happens to work; `injectBrandHeadTags` returns the head UNTOUCHED. So this
-- site has been silently opting itself out of head features, and every helper
-- involved reports success. It also cannot carry the `lang` attribute that
-- migration 507 added to the two SHARED head templates, so `bugs_closed/252`'s
-- locale half can never reach it — that is what surfaced this.
--
-- WHAT THIS CHANGES ON THE WIRE, stated plainly because it is a live site:
--   before: <!DOCTYPE html><html lang="en">\n<meta charset="utf-8"> … <body>
--   after:  <!DOCTYPE html><html lang="en-GB">\n<head lang="en-GB">\n<meta charset="utf-8"> … </head>\n<body>
-- The document gains a real <head> element, declares en-GB (site_specs already
-- holds it — set by 508, verified before authoring), and the five injectors above
-- now land INSIDE the head rather than after it. No tag is removed and no content
-- changes.
--
-- WHAT THIS DELIBERATELY DOES NOT FIX: `injectBrandHeadTags` still skips this site
-- wholesale, because its guard trips on the `rel="icon"` this template carries
-- (bugs_open/322 item 4). Wrapping the fragment does not change that — the guard is
-- the mechanism and it is that file's. So this site still gets no og:image and no
-- derived favicon tags, and that is expected after this migration.
--
-- SAFETY:
--   · ONE site, ONE component (`webdesign-couk-head`, its own function — a
--     `WHERE function='head'` query does not even see it).
--   · The lang attribute is GATED `{{if .lang}}`, matching 507 exactly, so a site
--     without the config renders byte-identically to today apart from the wrapper.
--   · WRAPPED input_schema shape (verified: `input_schema ? 'fields'` = true), and
--     the entry is MAP-VALUED — a scalar is silently skipped by the resolver.
--   · md5 drift guard + DO/RAISE, because a bare UPDATE matching 0 rows cannot stop
--     a COMMIT.
--   · Chrome is a stored artefact: this file changes NOTHING served until the site's
--     chrome re-renders. It moves the render_inputs `template` digest, so the
--     ordinary stale_chrome pipe will pick it up; the canary below drives it.
--
-- Apply: kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--          psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < this_file
-- Then record: ./scripts/migration/run-migrations.sh --record-only <file> --note "..."
-- Rollback: 529_webdesign_couk_head_gains_a_head_element_ROLLBACK.sql

BEGIN;

UPDATE content_components SET
  html_template = '<head{{if .lang}} lang="{{.lang}}"{{end}}>' || E'\n' || html_template || E'\n' || '</head>',
  input_schema  = jsonb_set(input_schema, '{fields,lang}', $fld${
    "type": "text",
    "source": "config.locale.lang",
    "required": false,
    "on_missing": "skip_field",
    "description": "BCP-47 language tag for this site, e.g. en-GB. Rendered onto the <head> open tag and read back by assemblePage to stamp <html lang>. Unset renders nothing and the page declares en. Same contract as the two shared head templates (migration 507)."
  }$fld$::jsonb),
  updated_at = now()
WHERE id = '14cf6193-c8f0-4640-9cf1-f8b5347e6885'
  AND md5(html_template) = '2d2bc47586604c200991b20b364dfcbc';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
  WHERE id = '14cf6193-c8f0-4640-9cf1-f8b5347e6885'
    AND html_template LIKE '<head{{if .lang}} lang="{{.lang}}"{{end}}>%'
    AND html_template LIKE '%</head>'
    AND input_schema #>> '{fields,lang,source}' = 'config.locale.lang'
    AND jsonb_typeof(input_schema #> '{fields,lang}') = 'object'
    -- the hand-authored contents must be intact, not replaced
    AND html_template LIKE '%port-compat.css%'
    AND html_template LIKE '%cf_analytics_token%'
    AND html_template LIKE '%<title></title>%';
  IF n <> 1 THEN
    RAISE EXCEPTION 'webdesign.co.uk head wrap did not land (drift guard hit — template bytes differ from the 2026-08-21 read, the schema entry is not map-valued, or the hand-authored contents were disturbed). Re-read the live row; do not loosen the guard.';
  END IF;
END $$;

-- Exactly one head element, not two: guard against re-running or against the
-- fragment having already been wrapped by someone else.
DO $$
DECLARE n int;
BEGIN
  SELECT (length(html_template) - length(replace(html_template, '</head>', '')))/7 INTO n
  FROM content_components WHERE id = '14cf6193-c8f0-4640-9cf1-f8b5347e6885';
  IF n <> 1 THEN
    RAISE EXCEPTION 'expected exactly one </head> in the template, found % — aborting', n;
  END IF;
END $$;

COMMIT;

-- VERIFY (read-only, after apply):
--   SELECT substring(html_template from 1 for 40) AS opens,
--          right(html_template, 10)               AS closes,
--          input_schema #>> '{fields,lang,source}' AS lang_source
--     FROM content_components WHERE id = '14cf6193-c8f0-4640-9cf1-f8b5347e6885';
--
-- Then drive the site's chrome and ONE page, and read the served bytes:
--   curl -s https://webdesign.co.uk/<inner-page> | head -4
-- Expect `<html lang="en-GB">` then `<head lang="en-GB">`. ⚠ Use an INNER page:
-- on a homepage `og:url` is the bare `/` before and after, so it discriminates nothing.

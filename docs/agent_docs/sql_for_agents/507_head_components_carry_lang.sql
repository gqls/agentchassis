-- 507_head_components_carry_lang.sql
-- bugs_open/252 (og/lang slug) §B — the document language leaves Go.
--
-- OWNER DECISION 2026-08-11, option 3: `lang` lives in the head COMPONENT, not
-- in a new sites column, and Go stops deciding it. This file is the config half.
-- The Go half (headLangAttr + htmlDocumentOpen in
-- platform/orchestration/actions/head_assembly.go, called from assemblePage) is
-- registered as SEO-005.
--
-- ⚠⚠ HOLD RELEASED AND APPLIED 2026-08-20 17:2xZ. The paragraph below is kept as the
-- RECORD of why it was held, not as a live instruction — it was applied only after
-- spliceOpenGraph was probed PRESENT on both v1.0.1320 replicas with a positive and a
-- negative control. Anyone REPLAYING this file against a fresh environment must re-satisfy
-- that precondition first. Original banner:
-- ⚠⚠ _HOLD, AND THE ORDER IS THE WHOLE POINT. DO NOT APPLY BEFORE THE BINARY
-- CARRYING head_assembly.go IS PROVEN RUNNING ON EVERY CHASSIS REPLICA.
-- DB config is live the moment it applies; Go is inert until the roll. The
-- template edits below move site_components.render_inputs' `template` digest
-- (it hashes html_template + input_schema BY VALUE —
-- platform/orchestration/datahelpers/chrome_render_inputs.go), which makes
-- StaleSiteComponentsCheck file a `stale_chrome` needs_rerender item per site.
-- Applied against the OLD binary, that edge is consumed by the old code: chrome
-- re-renders WITHOUT the lang attribute being read by anything, pages re-assemble
-- with the hardcoded `<html lang="en">`, render_inputs RESTAMPS, and the
-- detect->rebuild pipe then goes QUIET with the fleet still wrong. Re-firing it
-- needs a manual trigger. Prove the binary first:
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor <the head_assembly.go commit> <the stamp>
-- and if the startup line has scrolled, probe the binary with BOTH controls
-- (a sha that must be present and one that must be absent). See LANDMINES.md.
--
-- WHAT CHANGES, AND WHAT DOES NOT:
--   · Both shared head templates gain a GATED lang attribute on their <head>
--     open tag: `<head{{if .lang}} lang="{{.lang}}"{{end}}>`. A site with no
--     `lang` config renders BYTE-IDENTICALLY — the {{if}} produces nothing —
--     and assemblePage then falls back to `en`, which is byte-identical to the
--     line it replaced (pinned by TestHTMLDocumentOpenDefaultsToTodaysBytes).
--     So this file alone changes NOTHING on any site's output. 508 opts sites in.
--   · head-seo-standard ALSO loses its two blank-rendering og lines. At
--     site-level render there is no page, so `{{.title}}`/`{{.description}}` are
--     empty and it emits `<meta property="og:title" content="">` plus the same
--     for description. injectBrandHeadTags then appends a FILLED site-level pair
--     because its idempotency guard tests `rel="icon"` OR `og:image` and cannot
--     see a blank og tag — which is why 4 sites serve TWO og:title tags today
--     (finetuning.uk, leopardessconsulting.co.uk, ai-agent-orchestration.com,
--     gaswholesalers.com; verified on the wire 2026-08-19). assemblePage's new
--     spliceOpenGraph strips and rewrites both properties per page regardless,
--     so removing these lines is belt-and-braces for the SERVED page — its real
--     effect is that the STORED artefact stops carrying a self-contradiction.
--   · og:type / og:site_name / og:image / the `{{if .canonical_url}}` og:url line
--     are deliberately LEFT. The first two are genuinely site-level; og:image is
--     bugs_open/322 item 3 and under its landmine (do NOT gate it on an assets
--     row); the og:url line has never rendered on any site because nothing sets
--     `canonical_url` (that unset key is why head-seo-standard's own og:url
--     support has always been dead), and assembly strips og:url anyway.
--
-- SCHEMA SHAPE — READ THIS BEFORE EDITING EITHER ROW. The two components use
-- DIFFERENT input_schema shapes, verified 2026-08-19:
--   Document Head      -> FLAT   (input_schema ? 'fields' = false)
--   head-seo-standard  -> WRAPPED (input_schema ? 'fields' = true)
-- and the entry MUST be map-valued: render_site_components_action.go skips any
-- input_schema entry that is not a map as "not a field descriptor", which is why
-- Document Head's pre-existing flat SCALAR entries have never resolved and never
-- could (LANDMINES.md, footprint site_components/content_components).
--
-- NOT COVERED, deliberately: `webdesign.co.uk Document Head`
-- (14cf6193-c8f0-4640-9cf1-f8b5347e6885, its own function `webdesign-couk-head`,
-- 1 site, 117 assembled pages — the most in the fleet). That template is a bare
-- FRAGMENT: it has no <head> open tag and no </head> close tag, so there is no
-- attribute to carry a lang. It therefore keeps `en` via the Go default, which is
-- unchanged behaviour, and its pages still GAIN per-page og from the Go half
-- (pinned by TestSpliceOpenGraphHandlesAHeadWithNoCloseTag). Giving that
-- component a real <head> wrapper is a separate change on a hand-authored head
-- and is not smuggled in here.
--
-- COUNCIL ROUND 1 (corr 3b6712d4, APPROVED with advisories) — the checks its seats
-- asked for, RUN, with their answers:
--   · `guidelines` asked whether removing the two og template lines leaves a DANGLING
--     `required:true` on og_title/og_description. It does NOT — and the reason is worth
--     knowing: **neither key is declared in input_schema at all.** The live field list is
--     accent_color, background_color, canonical_url, description, font_url,
--     gtm_container_id, primary_color, secondary_color, structured_data, text_color,
--     theme_css, title. So `{{if .og_title}}…{{else}}{{.title}}{{end}}` has ALWAYS taken
--     the else branch, and `{{.title}}` is empty at a site-level render — which is exactly
--     why these two lines emit content="". Same for `{{if .og_image}}`: undeclared, so it
--     has never rendered, and this component's og:image comes only from
--     injectBrandHeadTags. `canonical_url` IS declared (`skip_field`) but nothing anywhere
--     sets it, so that branch has never fired either. Three dead template branches; this
--     file removes the two that were actively emitting a blank, and deliberately leaves
--     the inert ones alone.
--   · `debug_historian` objected that "the binary is proven running" was named as the HOLD
--     release criterion without specifying the mechanism, and that an image tag or git state
--     proves nothing. Correct, and the mechanism is a POD-GREP for a symbol this change adds,
--     with BOTH controls, on EVERY replica:
--       kubectl -n ai-persona-system exec <pod> -- grep -aq "spliceOpenGraph" /proc/1/exe
--       # positive control (must be PRESENT): injectCanonicalLink
--       # negative control (must be ABSENT):  a fabricated symbol
--     Run 2026-08-20 14:35Z against v1.0.1319: spliceOpenGraph **absent**, positive control
--     PRESENT, negative control absent. **So v1.0.1319 does NOT carry the fix and this file
--     MUST NOT be applied against it.** That build was cut at 10:18Z, ~4h before the Go
--     commit — a fresh roll is not evidence, which is the standing landmine.
--   · `debug_historian` also asked for a pre-migration backup independent of the guards. The
--     md5 guards abort on drift and the ROLLBACK restores byte-exactly, but take one anyway:
--       \copy (SELECT id, name, md5(html_template), html_template, input_schema
--              FROM content_components WHERE id IN
--              ('116c5f91-bc0d-439d-9e13-a3ba2d145571','aec98dbe-76b7-4e13-9641-e5b6ba2502aa'))
--         TO 'head_components_pre_507.tsv'
--   · `guardian` flagged the archive side effect: `trg_site_component_archive` fires
--     AFTER UPDATE OF rendered_html on site_components when the value actually changes
--     (verified via pg_trigger). So the next chrome render per site writes one history row.
--     That is the bugs_open/226 archive doing its job — 24 rows, and it PRESERVES the
--     pre-change head rather than costing anything. Named here so it is not rediscovered
--     as a surprise.
--
-- Apply: kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--          psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < this_file
-- Then record: ./scripts/migration/run-migrations.sh --record-only <file> --note "..."
-- Rollback: 507_head_components_carry_lang_ROLLBACK.sql

BEGIN;

-- A. Document Head (116c5f91…, 18 sites, FLAT schema).
--    Only the open tag changes; the rest of the template is untouched.
UPDATE content_components SET
  html_template = replace(html_template,
                          '<head>',
                          '<head{{if .lang}} lang="{{.lang}}"{{end}}>'),
  input_schema  = jsonb_set(input_schema, '{lang}', $fld${
    "type": "text",
    "source": "config.locale.lang",
    "required": false,
    "on_missing": "skip_field",
    "description": "BCP-47 language tag for this site, e.g. en-GB. Rendered onto the <head> open tag and read back by assemblePage to stamp <html lang>. Unset renders nothing and the page declares en, which is the pre-2026-08-20 behaviour."
  }$fld$::jsonb),
  updated_at = now()
WHERE id = '116c5f91-bc0d-439d-9e13-a3ba2d145571'
  AND md5(html_template) = '04d7d9cbcc8adb71d8579f07c45d3f7d';

-- Drift guard. A bare UPDATE matching 0 rows cannot stop the COMMIT
-- (ON_ERROR_STOP ignores an empty result), so the assertion must RAISE.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
  WHERE id = '116c5f91-bc0d-439d-9e13-a3ba2d145571'
    AND html_template LIKE '<head{{if .lang}} lang="{{.lang}}"{{end}}>%'
    AND input_schema #>> '{lang,source}' = 'config.locale.lang'
    AND jsonb_typeof(input_schema -> 'lang') = 'object';
  IF n <> 1 THEN
    RAISE EXCEPTION 'Document Head lang update did not land (drift guard hit — template bytes differ from the 2026-08-19 read, or the schema entry is not map-valued). Re-read the live row; do not loosen the guard.';
  END IF;
END $$;

-- B. head-seo-standard (aec98dbe…, 4 sites, WRAPPED schema).
--    Open tag + the two blank-rendering og lines removed.
UPDATE content_components SET
  html_template = replace(
                    replace(
                      replace(html_template,
                              '<head>',
                              '<head{{if .lang}} lang="{{.lang}}"{{end}}>'),
                      E'    <meta property="og:title" content="{{if .og_title}}{{.og_title}}{{else}}{{.title}}{{end}}">\n',
                      ''),
                    E'    <meta property="og:description" content="{{if .og_description}}{{.og_description}}{{else}}{{.description}}{{end}}">\n',
                    ''),
  input_schema  = jsonb_set(input_schema, '{fields,lang}', $fld${
    "type": "text",
    "source": "config.locale.lang",
    "required": false,
    "on_missing": "skip_field",
    "description": "BCP-47 language tag for this site, e.g. en-GB. Rendered onto the <head> open tag and read back by assemblePage to stamp <html lang>. Unset renders nothing and the page declares en, which is the pre-2026-08-20 behaviour."
  }$fld$::jsonb),
  updated_at = now()
WHERE id = 'aec98dbe-76b7-4e13-9641-e5b6ba2502aa'
  AND md5(html_template) = '3c2d8719629762f6bb8947c645f5c3df';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
  WHERE id = 'aec98dbe-76b7-4e13-9641-e5b6ba2502aa'
    AND html_template LIKE '<head{{if .lang}} lang="{{.lang}}"{{end}}>%'
    AND input_schema #>> '{fields,lang,source}' = 'config.locale.lang'
    AND jsonb_typeof(input_schema #> '{fields,lang}') = 'object'
    -- the two blank-renderers are gone ...
    AND html_template NOT LIKE '%og:title%'
    AND html_template NOT LIKE '%og:description%'
    -- ... and the site-level tags we do NOT own are still there.
    AND html_template LIKE '%og:type%'
    AND html_template LIKE '%og:site_name%'
    AND html_template LIKE '%og:image%';
  IF n <> 1 THEN
    RAISE EXCEPTION 'head-seo-standard lang/og update did not land (drift guard hit — template bytes differ from the 2026-08-19 read, a schema key was lost, or the og removal took a tag it should not have). Re-read the live row.';
  END IF;
END $$;

-- C. Nothing else may have been touched: exactly 3 components serve a head slot
--    and only the two above should now carry a lang gate.
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM content_components
  WHERE html_template LIKE '%<head{{if .lang}}%';
  IF n <> 2 THEN
    RAISE EXCEPTION 'expected exactly 2 head templates carrying the lang gate, found % — a third component was modified', n;
  END IF;
END $$;

COMMIT;

-- VERIFY (read-only, after apply):
--   SELECT cc.name, count(sc.id) AS sites,
--          (cc.html_template LIKE '<head{{if .lang}}%') AS lang_gated,
--          (cc.html_template LIKE '%og:title%')         AS still_blanks_og
--     FROM site_components sc JOIN content_components cc ON cc.id = sc.component_id
--    WHERE sc.slot_name = 'head' GROUP BY cc.name, cc.html_template;
-- Expect: Document Head t/f (18), head-seo-standard t/f (4),
--         webdesign.co.uk Document Head f/f (1) — the fragment, deliberately untouched.
--
-- No page or stored artefact changes until chrome re-renders (bugs_open/117:
-- chrome is a stored artefact). The stale_chrome pipe does that; 508 is what
-- actually gives a site a language to declare.

-- SQL_p12_cloudflare_analytics_token.sql — webdesign.co.uk
--
-- Arm the Cloudflare Web Analytics beacon that SQL_p7 wired and left GATED CLOSED.
-- The owner supplied the snippet from Web Analytics -> Manage site on 2026-07-27.
--
-- WHY ROUTE B AND NOT ROUTE A. SQL_p7 recommended Route A (dashboard "Automatic
-- Setup", edge injection, nothing in this repo changes). The owner enabled it and
-- reported "1 visit". Measured 18:20 UTC against the live home page: NO beacon,
-- under a plain curl, a desktop-Chrome UA, AND a cache-busted URL. The response
-- carries cf-ray and cf-cache-status: DYNAMIC — so it really does pass through the
-- proxy and could have been rewritten — and cache-control is `public, max-age=3600`
-- with NO `no-transform`, which is the one documented condition that blocks edge
-- injection. Route A is therefore not injecting, whatever the dashboard shows.
--   The "1 visit" is almost certainly Analytics & Logs -> Traffic, which is
--   server-side, on by default for any proxied zone, and has counted since launch.
--   It is a DIFFERENT product from Web Analytics and needed no enabling, so a count
--   there is not evidence the beacon works.
--
-- TWO CHANGES, both to the head fork only:
--
-- 1. `defer` -> `type="module"`, to match Cloudflare's OWN current snippet
--    verbatim. Both would work (a module script is deferred by design), but the
--    vendor's form is the one they test against, and there is no reason to carry a
--    variant we would have to reason about later.
--
-- 2. The token itself, into site_components.content_data for the `head` slot.
--    Read first: content_data was `{}`. The `||` merge is used rather than
--    jsonb_set with a literal object — a literal-object jsonb_set is a REPLACE and
--    would silently drop any sibling key (the trap that nearly cost the Gemini
--    lane its max_tokens setting).
--
-- THE TOKEN IS NOT A SECRET. It is a public site identifier — it ships in the HTML
-- of every page by design and identifies which Web Analytics property to report to.
-- It grants no account access. So it belongs in the DB and in this file.
--
-- PUBLISHING IS A SEPARATE STEP AND A PAGE RE-RENDER WILL NOT DO IT.
-- `bugs_open/117`: site chrome is a STORED artefact in site_components.rendered_html
-- and assemblePage copies it verbatim; nothing rebuilds it at page-render time.
-- Publish with a `nav_drift` work item routed to `nav-updater`, which runs
-- render_site_components and queues the page re-renders itself. Do NOT queue page
-- re-renders first — that races the chrome rebuild and renders pages against the
-- OLD head.
--   This site is CLEAN on `bugs_open/118` (chrome selection ignores is_active):
--   all three slots carry EXPLICIT site_components assignments, so the
--   `ORDER BY name LIMIT 1` fallback never runs, and `head` points at
--   14cf6193-c8f0-4640-9cf1-f8b5347e6885 — the exact row SQL_p7 edited. Checked,
--   not assumed.

\set ON_ERROR_STOP on

BEGIN;

-- 1. match Cloudflare's own snippet form
UPDATE content_components
   SET html_template = replace(
         html_template,
         '<script defer src="https://static.cloudflareinsights.com/beacon.min.js"',
         '<script type="module" src="https://static.cloudflareinsights.com/beacon.min.js"'),
       updated_at = NOW()
 WHERE function = 'webdesign-couk-head'
   AND is_active
   AND html_template LIKE '%<script defer src="https://static.cloudflareinsights.com/beacon.min.js"%';

-- 2. arm the gate
UPDATE site_components sc
   SET content_data = COALESCE(sc.content_data, '{}'::jsonb)
                      || jsonb_build_object('cf_analytics_token', '633f794e53dc4f718e91be595d7037ff'),
       updated_at = NOW()
  FROM sites s
 WHERE sc.site_id = s.id
   AND s.domain = 'webdesign.co.uk'
   AND sc.slot_name = 'head';

DO $verify$
DECLARE v_tmpl text; v_tok text; v_rendered text;
BEGIN
    SELECT html_template INTO v_tmpl
      FROM content_components WHERE function = 'webdesign-couk-head' AND is_active;

    IF v_tmpl NOT LIKE '%type="module" src="https://static.cloudflareinsights.com/beacon.min.js"%' THEN
        RAISE EXCEPTION 'beacon tag did not take the module form';
    END IF;
    IF v_tmpl NOT LIKE '%{{if .cf_analytics_token}}%' THEN
        RAISE EXCEPTION 'the gate is gone — an unset token would now ship a broken script';
    END IF;
    IF v_tmpl NOT LIKE '%{{.cf_analytics_token}}%' THEN
        RAISE EXCEPTION 'the token placeholder was destroyed by the replace';
    END IF;
    -- placeholders assemblePage rewrites must survive (SQL_p7's invariant)
    IF v_tmpl NOT LIKE '%<title></title>%' OR v_tmpl NOT LIKE '%content=""%' THEN
        RAISE EXCEPTION 'head lost its empty <title> or empty content="" placeholder';
    END IF;

    SELECT sc.content_data->>'cf_analytics_token', sc.rendered_html
      INTO v_tok, v_rendered
      FROM site_components sc JOIN sites s ON s.id = sc.site_id
     WHERE s.domain = 'webdesign.co.uk' AND sc.slot_name = 'head';

    IF v_tok <> '633f794e53dc4f718e91be595d7037ff' THEN
        RAISE EXCEPTION 'token not set, got %', COALESCE(v_tok, '(null)');
    END IF;

    RAISE NOTICE 'template armed and token set.';
    IF v_rendered LIKE '%beacon.min.js%' THEN
        RAISE NOTICE 'STORED chrome already carries the beacon.';
    ELSE
        RAISE NOTICE 'STORED chrome does NOT yet carry the beacon — expected. Publish with a nav_drift item (117); a page re-render will NOT do it.';
    END IF;
END
$verify$;

COMMIT;

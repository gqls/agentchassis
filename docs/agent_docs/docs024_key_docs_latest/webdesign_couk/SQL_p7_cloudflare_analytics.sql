-- SQL_p7_cloudflare_analytics.sql — webdesign.co.uk, phase 2 W1
--
-- Wire Cloudflare Web Analytics into the site's head chrome, GATED on a token.
--
-- WHY GATED. The beacon needs a site token that can only be minted in the
-- Cloudflare dashboard, and `CF_API_TOKEN` here is a GitHub Actions secret — not
-- reachable from this workstation, so it cannot be created programmatically.
-- Shipping the script with a placeholder token would put a permanently-failing
-- request on every page of the site. So the template gates on the token: with no
-- token the tag does not render at all, and the moment one is set the next chrome
-- render picks it up.
--
-- HOW THE GATE HOLDS. render_site_components_action.go's schema fill is
-- GAP-FILL ONLY — "if the caller supplies a value it stays authoritative" — and a
-- field whose source resolves to nothing gets NO fallback applied (deliberately:
-- the comment there cites the phantom-CTA fossil LNK-007, "correct-or-absent").
-- So declaring cf_analytics_token with source 'static' and no fallback means an
-- unset token stays unset, and `{{if .cf_analytics_token}}` closes.
--
-- THE OWNER HAS TWO ROUTES, and only needs one:
--
--   ROUTE A (recommended, zero further work here) — Cloudflare dashboard →
--   Web Analytics → add webdesign.co.uk → **Automatic Setup**. Because the zone
--   is already proxied through Cloudflare, the edge injects the beacon itself.
--   Nothing in this repo needs to change and no deploy is required.
--
--   ROUTE B (version-controlled, survives a proxy change) — take the token from
--   the dashboard's Manual Setup snippet and run the UPDATE at the foot of this
--   file, then re-render chrome. The tag then lives in our own template.
--
-- Route A is simpler and is what I would pick. Route B is wired here so the
-- choice is his and neither is blocked on me.

\set ON_ERROR_STOP on

BEGIN;

UPDATE content_components
   SET html_template = $head$<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title></title>
<meta name="description" content="">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=Fira+Code:wght@400;600&display=swap">
<link rel="stylesheet" href="/assets/css/styles.css">
<link rel="stylesheet" href="/assets/css/port-compat.css">
<link rel="icon" href="/favicon.ico">
{{if .cf_analytics_token}}<script defer src="https://static.cloudflareinsights.com/beacon.min.js" data-cf-beacon='{"token": "{{.cf_analytics_token}}"}'></script>{{end}}$head$,
       input_schema = $schema${
         "fields": {
           "cf_analytics_token": {
             "type": "string",
             "source": "static",
             "required": false,
             "llm_guidance": "Cloudflare Web Analytics site token. NEVER generated or invented — it is minted in the Cloudflare dashboard and set by hand in site_components.content_data. Absent is the correct state until then; the template gates on it so no broken beacon ships."
           }
         }
       }$schema$::jsonb,
       updated_at = NOW()
 WHERE function = 'webdesign-couk-head'
   AND is_active;

-- ---------------------------------------------------------------------------
-- SET THE TOKEN (Route B only). Uncomment, paste the token, run, then re-render
-- chrome with a rerender-pages dispatch carrying refresh_site_components:true.
-- ---------------------------------------------------------------------------
-- UPDATE site_components sc
--    SET content_data = COALESCE(sc.content_data,'{}'::jsonb)
--                       || jsonb_build_object('cf_analytics_token','<PASTE TOKEN>'),
--        updated_at = NOW()
--   FROM sites s
--  WHERE sc.site_id = s.id AND s.domain = 'webdesign.co.uk' AND sc.slot_name = 'head';

DO $verify$
DECLARE v_tmpl text; v_tok text;
BEGIN
    SELECT html_template INTO v_tmpl
      FROM content_components WHERE function = 'webdesign-couk-head' AND is_active;

    IF v_tmpl NOT LIKE '%cloudflareinsights.com/beacon.min.js%' THEN
        RAISE EXCEPTION 'beacon not present in the head template';
    END IF;
    IF v_tmpl NOT LIKE '%{{if .cf_analytics_token}}%' THEN
        RAISE EXCEPTION 'beacon is NOT gated — an unset token would ship a broken script';
    END IF;
    -- The placeholders assemblePage rewrites must survive this edit.
    IF v_tmpl NOT LIKE '%<title></title>%' OR v_tmpl NOT LIKE '%content=""%' THEN
        RAISE EXCEPTION 'head lost its empty <title> or empty content="" placeholder';
    END IF;

    SELECT sc.content_data->>'cf_analytics_token' INTO v_tok
      FROM site_components sc JOIN sites s ON s.id = sc.site_id
     WHERE s.domain = 'webdesign.co.uk' AND sc.slot_name = 'head';

    IF v_tok IS NULL OR v_tok = '' THEN
        RAISE NOTICE 'beacon wired and GATED CLOSED — no token set yet, so nothing renders. This is the correct state until the owner supplies one (or uses Automatic Setup, which needs nothing here).';
    ELSE
        RAISE NOTICE 'beacon wired and ARMED with a token; re-render chrome to publish it';
    END IF;
END
$verify$;

COMMIT;

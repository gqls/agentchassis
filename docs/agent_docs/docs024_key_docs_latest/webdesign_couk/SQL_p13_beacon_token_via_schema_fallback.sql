-- SQL_p13_beacon_token_via_schema_fallback.sql — webdesign.co.uk
--
-- Make the Cloudflare beacon actually RENDER. SQL_p7's Route B could never have
-- worked, and SQL_p12 (which followed it) armed the wrong place.
--
-- WHAT WENT WRONG, measured end to end 2026-07-27 20:00 UTC.
-- SQL_p7 declared `cf_analytics_token` with `source: 'static'` and NO fallback,
-- and instructed the operator to put the token in
-- `site_components.content_data`. SQL_p12 did exactly that, its verify block
-- passed, the nav_drift work item completed, and `render_site_components`
-- reported `"rendered": {"head": true, ...}, "success": true`. The stored chrome
-- was genuinely rebuilt — `site_components.updated_at` moved to 19:58:21 and the
-- head grew 570 -> 571 bytes. And the beacon was STILL ABSENT.
--
-- The cause is that **`site_components.content_data` is never read by the chrome
-- renderer.** `renderCtx.ContentData` (render_site_components_action.go:227) is a
-- FIXED, hardcoded vocabulary built from the `sites` row — company_name, tagline,
-- domain, email, nav, colours. Nothing merges the component's stored
-- `content_data` into it. The schema-driven fill added for bugs_open/018
-- (line 584 onward) then reaches `cf_analytics_token`, sees `source: 'static'`,
-- and takes this branch (line 622):
--
--     if source == "" || source == "static" || strings.HasPrefix(source, "static.") {
--         if fb := def["fallback"]; fb != nil {
--             renderCtx.ContentData[name] = fb      // <- the ONLY way a static field is filled
--         } else {
--             unresolved = append(unresolved, name) // <- what actually happened
--         }
--         continue
--     }
--
-- So a `static` field with no `fallback` is UNRESOLVABLE by construction, the
-- `{{if .cf_analytics_token}}` gate stays shut, and no beacon renders. The gate
-- did its job perfectly — it was the supply that never arrived.
--
-- WHY THIS SURVIVED A VERIFY BLOCK. SQL_p7's checks asserted the template
-- contains the beacon and is gated, and SQL_p12's asserted the token is in
-- content_data. Both are assertions about the WRITE. Neither exercised the READ,
-- and no test rendered the component and looked for the tag. This is the
-- writes-the-field-is-not-reads-the-field trap, and the fix below is verified the
-- other way round: by re-rendering and grepping the STORED artefact.
--
-- THE FIX. `fallback` is this system's own idiom for "a declared static literal"
-- — the code comment at line 624 says exactly that, citing nav_aria_label. So the
-- token belongs in the schema's fallback, which is the one place the renderer
-- will read it from.
--
-- SAFE TO PUT IT HERE, checked not assumed: component
-- 14cf6193-c8f0-4640-9cf1-f8b5347e6885 is referenced by exactly ONE row in
-- site_components — webdesign.co.uk, slot `head`. No other site can inherit this
-- token. (A shared component would need a different mechanism.)
--
-- The token is a PUBLIC site identifier — it ships in the HTML of every page by
-- design and grants no account access — so it is not a secret being pasted into a
-- shared table.
--
-- jsonb_set with `create_if_missing => true` on a LEAF path adds one key and
-- leaves every sibling intact. This is NOT the trap where jsonb_set is handed a
-- whole literal object and silently replaces the siblings.

\set ON_ERROR_STOP on

BEGIN;

UPDATE content_components
   SET input_schema = jsonb_set(
                        jsonb_set(
                          COALESCE(input_schema, '{}'::jsonb),
                          '{fields,cf_analytics_token,fallback}',
                          '"633f794e53dc4f718e91be595d7037ff"'::jsonb,
                          true),
                        '{fields,cf_analytics_token,llm_guidance}',
                        to_jsonb(
                          'Cloudflare Web Analytics site token — a PUBLIC site identifier, not a secret. '
                          'NEVER generated or invented: it is minted in the Cloudflare dashboard (Web Analytics '
                          '-> Manage site) and set here as this field''s `fallback`. It MUST live in the fallback '
                          'and nowhere else: the chrome renderer never reads site_components.content_data, and a '
                          'static-source field with no fallback resolves to nothing, so the template gate stays '
                          'shut and no beacon ships (measured 2026-07-27 — see SQL_p13). Absent is the correct '
                          'state until a token exists.'::text),
                        true),
       updated_at = NOW()
 WHERE id = '14cf6193-c8f0-4640-9cf1-f8b5347e6885'
   AND is_active;

DO $verify$
DECLARE v_fb text; v_tmpl text; v_users int;
BEGIN
    SELECT input_schema->'fields'->'cf_analytics_token'->>'fallback', html_template
      INTO v_fb, v_tmpl
      FROM content_components WHERE id = '14cf6193-c8f0-4640-9cf1-f8b5347e6885';

    IF v_fb IS DISTINCT FROM '633f794e53dc4f718e91be595d7037ff' THEN
        RAISE EXCEPTION 'fallback not set, got %', COALESCE(v_fb, '(null)');
    END IF;
    IF v_tmpl NOT LIKE '%{{if .cf_analytics_token}}%' THEN
        RAISE EXCEPTION 'the gate is gone — the beacon would ship unconditionally';
    END IF;
    IF v_tmpl NOT LIKE '%<title></title>%' OR v_tmpl NOT LIKE '%content=""%' THEN
        RAISE EXCEPTION 'head lost a placeholder assemblePage rewrites';
    END IF;

    SELECT count(*) INTO v_users FROM site_components WHERE component_id = '14cf6193-c8f0-4640-9cf1-f8b5347e6885';
    IF v_users <> 1 THEN
        RAISE EXCEPTION 'component is shared by % site_components rows — a fallback token would leak to other sites', v_users;
    END IF;

    RAISE NOTICE 'fallback set, gate intact, component exclusive to 1 site.';
    RAISE NOTICE 'NOT DONE YET: re-render chrome with force_rerender (nav_drift -> nav-updater), then grep the STORED site_components.rendered_html for beacon.min.js. Status is not the artefact.';
END
$verify$;

COMMIT;

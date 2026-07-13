-- Step 1 / Layer 1a — stop hero & call-to-action emitting phantom CTA URLs.
--
-- Ships together with the resolve() fix in plan_sections_action.go (the `pages`
-- case must stop fabricating "/<path>.html"). Without that Go change, resolve()
-- still returns a truthy "/contact.html" and the template gate below renders it
-- anyway — so apply the Go change + roll the chassis image BEFORE this SQL.
--
-- What this does, for both components:
--   schema:   cta-url fields -> on_missing = skip_field, phantom `fallback` removed
--             (so an unresolved CTA url is simply omitted, not faked)
--   template: each button gated on BOTH its text AND its url being present, and
--             the literal fallbacks (/contact.html, #features) dropped
--             (so a missing url renders no button, never href="" or a phantom)
--
-- Net effect: a CTA whose destination doesn't resolve produces no button rather
-- than a broken link. Correct destinations are restored by the internal-link-
-- resolver agent (Step 3). Layer 2 (validator) makes any future phantom a
-- deploy blocker.
--
-- NB: only affects FUTURE renders. Already-deployed pages keep their phantom HTML
-- until re-rendered — re-render the hero/CTA pages after applying, then re-run the
-- dry-run from the previous step to confirm the phantoms are gone.

-- ---------------------------------------------------------------------------
-- 0. Snapshot (re-run of IF NOT EXISTS ... AS SELECT is a no-op; don't trust counts)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS content_components_bak_cta0610 AS
SELECT * FROM content_components WHERE name IN ('hero', 'call-to-action');

-- ---------------------------------------------------------------------------
-- 1. hero — schema: drop phantom fallbacks, set on_missing = skip_field
-- ---------------------------------------------------------------------------
UPDATE content_components
SET input_schema = jsonb_set(
        jsonb_set(
            (input_schema #- '{fields,cta_url,fallback}') #- '{fields,secondary_cta_url,fallback}',
            '{fields,cta_url,on_missing}', '"skip_field"', true
        ),
        '{fields,secondary_cta_url,on_missing}', '"skip_field"', true
    ),
    updated_at = now()
WHERE name = 'hero';

-- ---------------------------------------------------------------------------
-- 2. hero — template: gate each button on text AND url; drop literal fallbacks
--    (hero's buttons are each on a single line, so a single-line replace is safe)
-- ---------------------------------------------------------------------------
UPDATE content_components
SET html_template = replace(html_template,
      '{{if .cta_text}}<a href="{{if .cta_url}}{{.cta_url}}{{else}}/contact.html{{end}}"',
      '{{if and .cta_text .cta_url}}<a href="{{.cta_url}}"'),
    updated_at = now()
WHERE name = 'hero';

UPDATE content_components
SET html_template = replace(html_template,
      '{{if .secondary_cta}}<a href="{{if .secondary_cta_url}}{{.secondary_cta_url}}{{else}}#features{{end}}"',
      '{{if and .secondary_cta .secondary_cta_url}}<a href="{{.secondary_cta_url}}"'),
    updated_at = now()
WHERE name = 'hero';

-- ---------------------------------------------------------------------------
-- 3. call-to-action — schema: drop phantom fallbacks, set on_missing = skip_field
-- ---------------------------------------------------------------------------
UPDATE content_components
SET input_schema = jsonb_set(
        jsonb_set(
            (input_schema #- '{fields,primary_cta_url,fallback}') #- '{fields,secondary_cta_url,fallback}',
            '{fields,primary_cta_url,on_missing}', '"skip_field"', true
        ),
        '{fields,secondary_cta_url,on_missing}', '"skip_field"', true
    ),
    updated_at = now()
WHERE name = 'call-to-action';

-- ---------------------------------------------------------------------------
-- 4. call-to-action — template: tighten each gate to also require the url.
--    Its href already uses {{.primary_cta_url}} directly (no literal fallback),
--    so without this an omitted url would render href="" — an empty-href phantom.
-- ---------------------------------------------------------------------------
UPDATE content_components
SET html_template = replace(html_template,
      '{{if .primary_cta}}',
      '{{if and .primary_cta .primary_cta_url}}'),
    updated_at = now()
WHERE name = 'call-to-action';

UPDATE content_components
SET html_template = replace(html_template,
      '{{if .secondary_cta}}',
      '{{if and .secondary_cta .secondary_cta_url}}'),
    updated_at = now()
WHERE name = 'call-to-action';

-- ---------------------------------------------------------------------------
-- 5. Verify — fallbacks gone, on_missing = skip_field, literals removed.
--    Each template replace above must affect exactly 1 row; if a *_literal flag
--    below is still true, the stored whitespace differs from the match string
--    and that replace silently no-op'd — adjust the match, don't assume success.
-- ---------------------------------------------------------------------------
SELECT name,
       input_schema #> '{fields,cta_url}'           AS cta_url,
       input_schema #> '{fields,primary_cta_url}'   AS primary_cta_url,
       input_schema #> '{fields,secondary_cta_url}' AS secondary_cta_url,
       (html_template LIKE '%/contact.html%')  AS tmpl_has_contact_literal,
       (html_template LIKE '%/services.html%') AS tmpl_has_services_literal,
       (html_template LIKE '%{{else}}#features%') AS tmpl_has_features_literal,
       (html_template LIKE '%{{if and %')      AS tmpl_has_tightened_gate
FROM content_components
WHERE name IN ('hero', 'call-to-action')
ORDER BY name;

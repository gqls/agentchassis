-- Phase (post-1) — idea.uk contact-form: convert the dead form to a mailto, and
-- fix the stale contact email in its description. Owner decision 2026-07-17:
-- "Convert form → mailto" (a working mailto CTA already exists in contact-info).
--
-- CONTEXT
-- The fleet contact-form component (content_components name='contact-form',
-- forked_from IS NULL) renders `<form action="{{.form_action}}" method="POST">`.
-- idea.uk's per-site value resolved to "#contact" — a dead anchor (the fleet-wide
-- dead-form issue, aaa_fails_to_mend/006 §B). We do NOT touch the fleet template;
-- we change only idea.uk's per-site section data, which is the correct lever.
--
-- Two fixes, both in the SAME inline page_components.content_data (data_path is
-- empty and no site_specs aspect carries form_action, so this inline row IS the
-- authoritative source; the about/contact recovery rebuild reused it verbatim,
-- which is why "#contact" persisted — evidence a source edit survives a rebuild):
--   1. form_action  "#contact"  →  mailto:idea.uk@contactforsales.com
--   2. form_description still names the OLD email idea-uk@leopardess.uk (the
--      contact-info block above was aligned to the new address in p1_05/p1_06 but
--      this description was missed) → idea.uk@contactforsales.com. Phone kept as-is.
--
-- rendered_html is deliberately NOT hand-edited here — the rebuild regenerates it
-- from this corrected content_data via the template (editing the rendered artifact
-- is the documented revert trap). This SQL only fixes the source; a needs_page
-- rebuild flows it to the deployed artifact (and, post-activation, to gqls/vm-sites).

\set ON_ERROR_STOP on
\set CROW 'c3150979-d86a-4a45-aa2c-30f41e13be6e'

\echo '=== BEFORE ==='
SELECT content_data->>'form_action'      AS form_action,
       (content_data->>'form_description' ~ 'leopardess') AS desc_has_old_email
FROM page_components WHERE id = :'CROW';

UPDATE page_components
SET content_data = jsonb_set(
      jsonb_set(content_data,
                '{form_action}',
                '"mailto:idea.uk@contactforsales.com?subject=idea.uk enquiry"'::jsonb),
                '{form_description}',
                to_jsonb(replace(content_data->>'form_description',
                                 'idea-uk@leopardess.uk',
                                 'idea.uk@contactforsales.com'))),
    updated_at = now()
WHERE id = :'CROW'
  AND content_data->>'form_action' = '#contact';   -- idempotent guard

\echo '=== AFTER ==='
SELECT content_data->>'form_action'      AS form_action,
       (content_data->>'form_description' ~ 'leopardess') AS desc_has_old_email,
       content_data->>'form_description' AS form_description
FROM page_components WHERE id = :'CROW';

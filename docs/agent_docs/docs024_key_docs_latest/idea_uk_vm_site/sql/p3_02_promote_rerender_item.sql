-- p3_02_promote_rerender_item.sql — idea.uk: drive the reassembly after p3_01
--
-- WHY REUSE THIS ITEM RATHER THAN INSERT ONE. A needs_rerender item already exists
-- (dce4e4ac, filed by the missing_structure discovery check on 2026-07-18) carrying exactly
-- the spec this needs: refresh_site_components=true plus all 9 affected pages. Inserting a
-- second item would duplicate work and leave a stale one behind. The work loader consumes
-- status IN ('triaged','approved') (load_work_item_actions.go:559), so 'detected' is why it
-- has sat unactioned.
--
-- AND WHY ITS STATED REASON MUST CHANGE. Its reason reads "Pages deployed without
-- header/footer". That is FALSE for the deployed artefact — every page carries full chrome;
-- the footer is emitted as <section class="footer-…">, not <footer>, so a grep '<footer'
-- returns 0 and appears to confirm it (see /bugs_open/018). Leaving the false reason in
-- place is how someone re-runs a rerender, sees chrome "return", and records a fix that
-- fixed nothing. The remedy was right for the wrong reason; after p3_01 it is right for the
-- right reason, and the spec now says so.

BEGIN;

UPDATE site_work_items
SET status = 'triaged',
    spec = spec
      || jsonb_build_object(
           'reason',
             'Chrome templates rewritten and gated (p3_01_chrome_templates_gated.sql, bugs_open/018): '
             || 'site-header/site-footer previously declared field names absent from the renderer''s '
             || 'fixed vocabulary, so every value resolved empty and rendered as href="". Pages need '
             || 'reassembly to pick up the new chrome.',
           'superseded_reason',
             'Pages deployed without header/footer — FALSE: chrome was present, only its values were empty. '
             || 'Retained for audit; do not act on it.',
           'promoted_by', 'idea.uk vm site 3 session, 2026-07-19'
         ),
    updated_at = now()
WHERE id = 'dce4e4ac-794b-4566-a6dd-6a72e7b0cd6d'
  AND status = 'detected';

COMMIT;

SELECT id, item_type, status, handler_agent, priority,
       spec->>'refresh_site_components' AS refresh_chrome,
       jsonb_array_length(spec->'affected_pages') AS pages
FROM site_work_items WHERE id = 'dce4e4ac-794b-4566-a6dd-6a72e7b0cd6d';

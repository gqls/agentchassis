\set ON_ERROR_STOP on
-- SEED_2026-09-03g — refresh the site CHROME so the footer picks up the corrected sites.email.
--
-- 09-03f corrected `sites.email` and queued four single-page `page_rerender` items. All four
-- completed 16:26–16:29Z and the footer STILL served the old address: `site_components.footer`
-- was untouched (updated_at 15:02:39Z). A single-page page_rerender does NOT rebuild chrome.
--
-- What does: `render_site_components` (render_site_components_action.go), which reads the
-- address from `sites` directly (`COALESCE(si.email,'')`, :464) and is invoked by rerender-pages
-- when the needs_rerender spec carries `refresh_site_components: true` — the exact shape
-- reconcile_site_plan filed at 14:15Z, and the run that last rebuilt chrome (15:02:39Z). This
-- seed files the same shape by hand, with a distinct key. It re-assembles all four pages too, so
-- the contact-form's rendered_html (which also renders ctx.Email and holds nothing in
-- content_data) is refreshed in the same pass.
--
-- VERIFY AT THE SERVED BYTES, and mind the CDN: cache-control is max-age=3600 (measured
-- 2026-09-03 16:3xZ), so the edge may serve the old footer for up to an hour after deploy.
--   for u in index.html about.html contact.html; do printf "$u: "; \
--     curl -s https://gamedesign.uk/$u | grep -oE '[a-z0-9.]*@contactforsales\.com' | sort | uniq -c | tr '\n' ' '; echo; done
--   expect: ONLY gamedesignuk@ on every page; ZERO of the bare gamedesign@ form.
-- Apply: psql -f THIS FILE ONLY.
BEGIN;

DO $g$
DECLARE addr text;
BEGIN
  SELECT email INTO addr FROM sites WHERE id='8f17eb73-fc74-4718-8371-b3125bc4e414';
  IF addr <> 'gamedesignuk@contactforsales.com' THEN
    RAISE EXCEPTION '09-03g REFUSED: sites.email is % — apply 09-03f first', addr;
  END IF;
END $g$;

INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, priority, handler_agent, status, created_by, item_key)
VALUES ('8f17eb73-fc74-4718-8371-b3125bc4e414', 'gamedesign_uk_rebuild lane', 'build', 'needs_rerender', 'medium',
  'Refresh site chrome (footer contact) + re-assemble all pages after sites.email correction (owner ruling 2026-09-03)',
  '{"reason":"post_reconcile_assembly","refresh_site_components":true,"why":"sites.email changed to gamedesignuk@contactforsales.com; footer chrome and contact-form rendered_html still carry the old address; single-page page_rerender does not rebuild chrome","lane":"gamedesign_uk_rebuild"}'::jsonb,
  60, 'rerender-pages', 'triaged', 'gamedesign_uk_rebuild lane 2026-09-03', 'chrome_refresh_email_8f17eb73_2026-09-03')
ON CONFLICT DO NOTHING;

DO $v$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items
   WHERE item_key='chrome_refresh_email_8f17eb73_2026-09-03'
     AND status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled')
     AND (spec->>'refresh_site_components')::boolean;
  IF n <> 1 THEN RAISE EXCEPTION '09-03g FAILED: chrome refresh item not queued (found %)', n; END IF;
  RAISE NOTICE '09-03g OK: needs_rerender with refresh_site_components queued to rerender-pages';
END $v$;
COMMIT;

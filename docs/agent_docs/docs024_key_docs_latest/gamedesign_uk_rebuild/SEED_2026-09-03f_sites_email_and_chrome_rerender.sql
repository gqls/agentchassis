\set ON_ERROR_STOP on
-- SEED_2026-09-03f — the contact address, third and last surface: `sites.email`.
--
-- Owner 2026-09-03: the address is gamedesignuk@contactforsales.com.
--
-- WHAT 09-03c AND 09-03e MISSED, and why the miss was invisible:
--   09-03c updated the SPECS (submission.email, briefing.contact.contact_email).
--   09-03e rewrote the three components whose *content_data* carried the address.
--   Both verified clean. The SERVED page still showed the OLD address 3x on contact.html
--   and 2x on about.html [MEASURED 2026-09-03 16:2xZ at the served bytes].
--
-- Because the site CHROME does not read either of those. The footer renders
-- `ctx.Email` (component_library.go:1989, `<div class="footer-contact">`), and
-- rerender_pages_actions.go:796 loads that from **`sites.email`** — a column no spec
-- update touches. The footer was re-rendered at 15:02:39Z, AFTER the spec change, and
-- still emitted the old address, which is the tell: a fresh render with a stale value.
--
-- ⚠ THE MEASUREMENT TRAP, recorded because it is the reason this took three passes:
-- my census filtered `content_data::text LIKE '%contactforsales%'` — the column I was
-- fixing. `contact-form` carries the address ONLY in `rendered_html`, and the footer is
-- not a page component at all, so BOTH were invisible to the query that told me the job
-- was done. Three components reported 3 new / 0 old while the live page served the old
-- one. Only the served bytes are ground truth.
--
-- One value feeds both remaining symptoms (footer chrome AND the contact-form's rendered
-- html, which also renders ctx.Email and holds nothing in content_data), so this is one
-- fix, not two.
--
-- Apply: psql -f THIS FILE ONLY.  Then VERIFY AT THE SERVED BYTES, not at the row:
--   curl -s https://gamedesign.uk/contact.html | grep -o '[a-z0-9.]*@contactforsales\.com' | sort | uniq -c
--   expect: only gamedesignuk@ ; zero occurrences of the bare gamedesign@ form.
BEGIN;

DO $g$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM sites
   WHERE id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND email='gamedesign@contactforsales.com';
  IF n <> 1 THEN RAISE EXCEPTION '09-03f REFUSED: sites.email is not the expected old value (matched % rows)', n; END IF;
END $g$;

UPDATE sites SET email='gamedesignuk@contactforsales.com'
 WHERE id='8f17eb73-fc74-4718-8371-b3125bc4e414'
   AND email='gamedesign@contactforsales.com';

-- Re-render all four live pages so chrome (footer) and the contact-form's rendered_html
-- are rebuilt from the corrected value.
-- ⚠ `spec.page_name` is MANDATORY on a REASONED rerender: without it save_page_sections
-- skips ("no page name"), reports success, and deploys stale sections (LANDMINES).
-- A reason is what routes to the sections path at all; a rerender with NO reason takes
-- assemble mode and re-ships stored bytes — which is exactly what the four completed
-- `_assemble` rerenders on this site did.
INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, spec, page_id, priority, handler_agent, status, created_by, item_key)
SELECT '8f17eb73-fc74-4718-8371-b3125bc4e414', 'gamedesign_uk_rebuild lane', 'build', 'page_rerender', 'medium',
       'Re-render ' || p.name || ' so the footer chrome and contact-form pick up the corrected sites.email (owner ruling 2026-09-03)',
       jsonb_build_object(
         'domain','gamedesign.uk',
         'page_id', p.id::text,
         'page_name', p.name,
         'filename', regexp_replace(p.url,'^/',''),
         'reason','section_data_resolved'),
       p.id, 60, 'page-rerender', 'triaged', 'gamedesign_uk_rebuild lane 2026-09-03',
       'page_rerender_addr_' || p.name || '_8f17eb73'
  FROM pages p
 WHERE p.site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND p.status='active'
ON CONFLICT DO NOTHING;

DO $v$
DECLARE n int; addr text;
BEGIN
  SELECT email INTO addr FROM sites WHERE id='8f17eb73-fc74-4718-8371-b3125bc4e414';
  IF addr <> 'gamedesignuk@contactforsales.com' THEN
    RAISE EXCEPTION '09-03f FAILED: sites.email is now %', addr;
  END IF;
  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND item_key LIKE 'page_rerender_addr_%'
     AND status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled')
     AND spec->>'page_name' IS NOT NULL AND spec->>'reason' IS NOT NULL;
  IF n <> 4 THEN RAISE EXCEPTION '09-03f FAILED: expected 4 reasoned rerenders carrying page_name, found %', n; END IF;
  RAISE NOTICE '09-03f OK: sites.email corrected, 4 reasoned rerenders queued';
END $v$;

COMMIT;

SELECT item_key, status, spec->>'page_name' AS page, spec->>'reason' AS reason
  FROM site_work_items
 WHERE site_id='8f17eb73-fc74-4718-8371-b3125bc4e414' AND item_key LIKE 'page_rerender_addr_%' ORDER BY 1;

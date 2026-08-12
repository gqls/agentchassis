-- FILE: SQL_2026-08-12d_restore_cta_urls.sql
--
-- REGRESSION REPAIR. The £149 copy migration dropped every call-to-action URL
-- on the four offer pages: 14 links across 7 components, including both buttons
-- on the home page hero. The migration itself is correct; this is damage it did
-- on the way, and it is `bugs_open/238` firing.
--
-- MECHANISM, read from the data rather than guessed. The writer preserved the
-- button LABELS (`cta_text`, `primary_cta`, `secondary_cta`) and dropped the
-- DESTINATIONS (`cta_url`, `primary_cta_url`, `secondary_cta_url`). Those URL
-- keys are populated by `resolve_internal_links`, not by the writer, so the
-- writer never emits them — and `save_page_sections` REPLACES content_data
-- rather than merging it. Every regeneration therefore drops every
-- resolver-populated key, which is exactly bugs_open/238's shape ("save
-- REPLACES, rerender MERGES, so a key survives months of rerenders and dies on
-- the first regeneration"). The templates then render `<a href="">` as nothing
-- at all, so the page loses the anchor silently: no error, no shrunken text,
-- and the claims scan is clean because prose is not what changed.
--
-- WHY THE GATE DID NOT CATCH IT. `gate_page_links.py` was armed on the guide
-- page's `required_links`, because the guide was the page I judged at risk. It
-- passed, correctly — the guide kept all four of its links. The damage happened
-- on the four pages I had NOT declared a link set for. **A gate covers what you
-- point it at, and I pointed it at the page I was worried about instead of at
-- the pages the writer was actually rewriting.** The durable fix is to declare
-- `required_links` on every page that carries links before any rewrite; this
-- file does that for the four, so the next migration cannot repeat it.
--
-- WHY content_data AND NOT rendered_html. The platform's own work item says so:
-- "re-declare it in content_data or lock the component — do not paste it back
-- into rendered_html, which only re-arms this same loss." A hand-patched
-- rendered_html is precisely what `page_divergence_overwritten` fires on, and
-- the next rebuild would drop it again. So: fix the data, then re-render, which
-- regenerates rendered_html from content_data with no LLM involved.
--
-- URLS ARE RECOVERED, NOT INVENTED: each one is read back out of the
-- pre-migration rendered HTML captured before the first work item was created
-- (the `tel:` spacing is verbatim from what was live).
--
-- NOT TOUCHED: index/call-to-action carried NO links before the migration
-- either (its labels have never had URLs). That is a pre-existing defect, not
-- this migration's, and repairing it silently here would hide it.

BEGIN;

-- 1. Restore the dropped destinations.
UPDATE page_components pc SET
  content_data = pc.content_data || v.urls,
  updated_at = now()
FROM pages p, (VALUES
  ('index',        'hero',           '{"cta_url":"/contact.html","secondary_cta_url":"/what-you-get.html"}'::jsonb),
  ('faq',          'hero',           '{"cta_url":"/contact.html","secondary_cta_url":"tel:+44 (0) 7934 524 911"}'::jsonb),
  ('faq',          'call-to-action', '{"primary_cta_url":"/contact.html","secondary_cta_url":"tel:+44 (0) 7934 524 911"}'::jsonb),
  ('how-it-works', 'hero',           '{"cta_url":"/contact.html","secondary_cta_url":"/what-you-get.html"}'::jsonb),
  ('how-it-works', 'call-to-action', '{"primary_cta_url":"/contact.html","secondary_cta_url":"tel:+44 (0) 7934 524 911"}'::jsonb),
  ('what-you-get', 'hero',           '{"cta_url":"/contact.html","secondary_cta_url":"/how-it-works.html"}'::jsonb),
  ('what-you-get', 'call-to-action', '{"primary_cta_url":"/contact.html","secondary_cta_url":"/faq.html"}'::jsonb)
) AS v(page_name, slot, urls)
WHERE pc.page_id = p.id
  AND p.site_id = '1fcfa4f3-ec80-4010-878b-b971cd46711f'
  AND p.name = v.page_name
  AND pc.slot_name = v.slot
  AND pc.locked_at IS NULL;          -- never write through a lock

-- 2. Declare the link set on every page that carries links, so the next
--    rewrite is gated mechanically rather than by whoever remembers to look.
UPDATE pages SET content_direction = COALESCE(content_direction,'{}'::jsonb)
     || jsonb_build_object('required_links', jsonb_build_array('/contact.html','/what-you-get.html'))
 WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND name='index';
UPDATE pages SET content_direction = COALESCE(content_direction,'{}'::jsonb)
     || jsonb_build_object('required_links', jsonb_build_array('/contact.html'))
 WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND name='faq';
UPDATE pages SET content_direction = COALESCE(content_direction,'{}'::jsonb)
     || jsonb_build_object('required_links', jsonb_build_array('/contact.html','/what-you-get.html'))
 WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND name='how-it-works';
UPDATE pages SET content_direction = COALESCE(content_direction,'{}'::jsonb)
     || jsonb_build_object('required_links', jsonb_build_array('/contact.html','/how-it-works.html','/faq.html'))
 WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND name='what-you-get';

DO $$
DECLARE n_fixed int; n_declared int;
BEGIN
  SELECT count(*) INTO n_fixed FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND p.name IN ('index','faq','how-it-works','what-you-get')
     AND pc.slot_name IN ('hero','call-to-action')
     AND (pc.content_data ? 'cta_url' OR pc.content_data ? 'primary_cta_url');
  IF n_fixed <> 7 THEN RAISE EXCEPTION 'expected 7 components carrying a primary URL, got %', n_fixed; END IF;

  -- Every restored component must ALSO have its secondary destination, or the
  -- page comes back with one button where it had two — a half repair that
  -- looks like a whole one.
  SELECT count(*) INTO n_fixed FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f'
     AND p.name IN ('index','faq','how-it-works','what-you-get')
     AND pc.slot_name IN ('hero','call-to-action')
     AND pc.content_data ? 'secondary_cta_url';
  IF n_fixed <> 7 THEN RAISE EXCEPTION 'expected 7 secondary URLs, got %', n_fixed; END IF;

  SELECT count(*) INTO n_declared FROM pages
   WHERE site_id='1fcfa4f3-ec80-4010-878b-b971cd46711f' AND content_direction ? 'required_links';
  IF n_declared <> 5 THEN RAISE EXCEPTION 'expected 5 pages declaring required_links, got %', n_declared; END IF;
END $$;

COMMIT;

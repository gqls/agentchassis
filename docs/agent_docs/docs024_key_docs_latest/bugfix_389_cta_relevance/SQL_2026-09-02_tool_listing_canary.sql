-- LAST RESIDUE: 7 refs that are NOT CTA fields. They sit in `tool-cta` / `tool-list`
-- components as entries in an `items[]` array of related-tool CARDS (url, name, image, title,
-- nav_label, meta_description). That is why the cta_links_stale pass completed on four of these
-- pages and cleared nothing: there is no *_url CTA field to recompute. Different mechanism.
--
-- HYPOTHESIS UNDER TEST: `section_data_resolved` regenerates a section's data from the live
-- page set, so an ARCHIVED tool drops out of the items array (the same way the footer's
-- generated tool list dropped it). If WRONG, this run completes and the reference is still
-- there — readable in one query, and I then try assemble mode (no spec.reason) instead.
--
-- ⚠ LANDMINE PRE-CHECK, required before firing this reason and RUN: "Firing section_data_resolved
-- on a LOCKED, positionally-named section DUPLICATES it, not protects it." Measured 2026-09-02
-- across all 7: locked = false, positionally_named = false, component_id present on every one.
-- The precondition does not hold, so the trap is not armed here. The entry also requires the
-- component count be compared BEFORE and AFTER — asserted below.
--
-- Canary is ONE generic page. The three `owned` pages are excluded on purpose: they already
-- refused the ordinary path with OWNED_PAGE_GUARD and need the owned-page route.

BEGIN;
INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec)
SELECT p.site_id, p.id, 'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
       'page_rerender', 'page-rerender', 'detected', 'high', 30,
       'Tool-listing residue canary: regenerate the related-tools items array on token-calculator so the archived tool drops out',
       'tool_listing_canary:' || p.id::text,
       jsonb_build_object(
         'reason','section_data_resolved',
         'page_id', p.id::text, 'page_name', p.name,
         'fix','The related-tools list still carries the archived password-entropy tool. Regenerate the section data from the live page set so the archived entry drops.',
         'original_pipeline','build')
FROM pages p
WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND p.url='/tools/token-calculator/index.html'
  AND NOT EXISTS (SELECT 1 FROM site_work_items w WHERE w.item_key='tool_listing_canary:'||p.id::text);

SELECT p.url,
       (SELECT count(*) FROM page_components c WHERE c.page_id=p.id) AS components_BEFORE,
       (SELECT count(*) FROM page_components c WHERE c.page_id=p.id AND c.content_data::text LIKE '%password-entropy%') AS refs_BEFORE
FROM pages p
WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND p.url='/tools/token-calculator/index.html';
COMMIT;

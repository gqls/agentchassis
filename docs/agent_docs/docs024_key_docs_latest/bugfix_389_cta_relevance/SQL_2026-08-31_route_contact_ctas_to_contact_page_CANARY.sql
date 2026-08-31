-- OWNER DECISION 2026-08-31: "we can route the get in touch buttons to the contact page."
-- CANARY: one field first, on the page where the defect is LIVE and visible.
--
-- WHY THIS IS DURABLE AND NOT A ONE-OFF -- verified in the code before writing:
--   applyCTARecompute's KEEP #1 (rerender_page_sections_action.go) is bugs_open/248's fix,
--   slug cta_recompute_clobbers_authored_contact_links. It keeps -- and deliberately REWRITES --
--   a stored destination for which storedCTADestinationIsAuthored() is true:
--       ctaExcludedDestination(url)      -- 'contact' IS in areasExcludedFromCTA
--                                           (resolve_internal_links_action.go:86-88)
--     AND validPages.Contains(url)       -- /contact.html is active on all three sites (200, control 404)
--     AND NOT CTAMintedCovers(...)       -- no mint stamp names /contact.html for these fields
--   So a hand-written /contact.html is exactly the shape KEEP #1 protects: every later recompute
--   keeps it. This is the supported way to say "this button is a contact button", not a patch.
--
-- ⚠ Deliberately NOT touching `__cta_minted`. The predicate is url-specific, so a stale stamp
--    naming the old tool cannot cover /contact.html; and the mint map merges SHALLOWLY
--    (cta_provenance.go:111-118), so writing one field's stamp would replace the whole record.
--
-- THE DEFECT THIS FIXES, live at the served bytes 2026-08-31 ~15:5xZ:
--   ai-agent-orchestration.com/services.html serves
--     <a href="/tools/password-entropy.html" class="cta-btn cta-btn-primary">Book a Technical Discovery Call</a>
--   -- a button offering a discovery call that opens a password-strength toy.
--
-- Falsifiable: if KEEP #1 does NOT hold, the rerender dispatched after this will put the tool
-- back and the served page will show password-entropy again. That is the disconfirming result
-- and it is one grep.

BEGIN;

UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{primary_cta_url}', '"/contact.html"'::jsonb, false),
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'      -- ai-agent-orchestration.com
  AND p.url = '/services.html'
  AND pc.slot_name = 'call-to-action'
  AND pc.content_data->>'primary_cta_url' = '/tools/password-entropy.html'
  AND pc.content_data->>'primary_cta' ILIKE '%discovery call%';

DO $$
DECLARE n_set int; n_stamp int;
BEGIN
  SELECT count(*) INTO n_set
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da' AND p.url = '/services.html'
     AND pc.slot_name = 'call-to-action'
     AND pc.content_data->>'primary_cta_url' = '/contact.html';
  IF n_set <> 1 THEN
    RAISE EXCEPTION 'expected exactly 1 field repointed to /contact.html, found %', n_set;
  END IF;

  -- KEEP #1 requires NO mint stamp naming this url for this field. Assert we did not create one.
  SELECT count(*) INTO n_stamp
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da' AND p.url = '/services.html'
     AND pc.slot_name = 'call-to-action'
     AND pc.content_data->'__cta_minted'->>'primary_cta_url' = '/contact.html';
  IF n_stamp <> 0 THEN
    RAISE EXCEPTION 'a mint stamp now names /contact.html — KEEP #1 would NOT hold it (% rows)', n_stamp;
  END IF;

  RAISE NOTICE 'canary repointed: 1 field -> /contact.html, no mint stamp, KEEP #1 applies';
END $$;

COMMIT;

-- Rerender aiao/services so the repointed CTA reaches the served bytes -- AND so the
-- recompute gets a chance to undo it. That second part is the point: this dispatch is the
-- test of KEEP #1, not just a publish step.
--
-- PASS  = served page shows "Book a Technical Discovery Call" -> /contact.html
-- FAIL  = served page shows password-entropy again (KEEP #1 did not hold; stop the batch)
--
-- reason=cta_links_stale is the no-LLM path. spec.page_name is load-bearing (LANDMINES).
BEGIN;
INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec)
SELECT p.site_id, p.id, 'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
       'page_rerender', 'page-rerender', 'detected', 'high', 30,
       'Owner decision 2026-08-31: publish the contact-page repoint on aiao/services and prove KEEP #1 holds it through a recompute',
       'contact_repoint_canary:' || p.id::text,
       jsonb_build_object(
         'reason','cta_links_stale', 'check','misdirected_cta',
         'page_id', p.id::text, 'page_name','services',
         'fix','The primary CTA copy asks the reader to book a discovery call; its destination is now /contact.html. Publish it and leave the authored contact destination in place.',
         'original_pipeline','build')
FROM pages p
WHERE p.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND p.url='/services.html';
SELECT item_key, status FROM site_work_items WHERE item_key LIKE 'contact_repoint_canary:%';
COMMIT;

-- One more field under the owner's 2026-08-31 decision, missed by the census regex.
--
-- finetuning.uk/ai-guides.html primary_cta = "Start a Conversation" -> the retiring tool.
-- That is a get-in-touch button by any reading, and it falls squarely under the ruling. My
-- contact-intent regex ('get in touch|contact|write to|email|call us|book a|talk to|speak to|
-- discovery call') does not match it -- "conversation" does not contain "contact" -- so the
-- census under-reported the class by one. Recorded as MISSTEP 18: a regex classifier has a
-- false-negative tail by construction, and the tail is only visible in the RESIDUE after the
-- matched set is cleared. Clearing the class is what made this one legible.
--
-- Durable by KEEP #1, same as the batch: 'contact' is in areasExcludedFromCTA, /contact.html is
-- live on finetuning.uk (200, control 404), and this field carries no mint stamp naming it.

BEGIN;

UPDATE page_components pc
SET content_data = jsonb_set(pc.content_data, '{primary_cta_url}', '"/contact.html"'::jsonb, false),
    updated_at = now()
FROM pages p
WHERE p.id = pc.page_id
  AND p.site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'   -- finetuning.uk
  AND p.url = '/ai-guides.html'
  AND pc.content_data->>'primary_cta_url' = '/tools/password-entropy.html'
  AND pc.content_data->>'primary_cta' ILIKE '%start a conversation%';

INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec)
SELECT p.site_id, p.id, 'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
       'page_rerender', 'page-rerender', 'detected', 'high', 30,
       'Owner decision 2026-08-31: publish the contact-page repoint on ai-guides ("Start a Conversation")',
       'contact_repoint:' || p.id::text,
       jsonb_build_object(
         'reason','cta_links_stale', 'check','misdirected_cta',
         'page_id', p.id::text, 'page_name','ai-guides',
         'fix','The CTA copy invites the reader to start a conversation; its destination is now /contact.html. Publish it and leave the authored contact destination in place.',
         'original_pipeline','build')
FROM pages p
WHERE p.site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc' AND p.url='/ai-guides.html'
  AND NOT EXISTS (SELECT 1 FROM site_work_items w WHERE w.item_key='contact_repoint:'||p.id::text);

DO $$
DECLARE n_set int; n_stamp int;
BEGIN
  SELECT count(*) INTO n_set FROM page_components pc JOIN pages p ON p.id=pc.page_id
   WHERE p.site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc' AND p.url='/ai-guides.html'
     AND pc.content_data->>'primary_cta_url'='/contact.html';
  IF n_set <> 1 THEN RAISE EXCEPTION 'expected 1 field repointed, found %', n_set; END IF;

  -- scoped to THIS page, not the fleet (misstep 17)
  SELECT count(*) INTO n_stamp
    FROM page_components pc JOIN pages p ON p.id=pc.page_id
    CROSS JOIN LATERAL jsonb_each(COALESCE(pc.content_data->'__cta_minted','{}'::jsonb)) m
   WHERE p.site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc' AND p.url='/ai-guides.html'
     AND m.value #>> '{}' = '/contact.html';
  IF n_stamp <> 0 THEN RAISE EXCEPTION 'mint stamp names /contact.html on this page (%) — KEEP #1 would not hold', n_stamp; END IF;

  RAISE NOTICE 'ai-guides repointed to /contact.html, no mint stamp, rerender queued';
END $$;

COMMIT;

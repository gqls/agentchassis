-- OWNER DECISION 2026-08-31, applied to the remaining 19 contact-intent CTAs.
--
-- CANARY PASSED end-to-end before this ran (aiao/services.html):
--   * stored url survived a full cta_links_stale recompute -> KEEP #1 holds in production,
--     not just in my reading of the source;
--   * served bytes at last-modified 16:09:23 (matching the deploy) show
--     <a href="/contact.html" class="cta-btn cta-btn-primary">Book a Technical Discovery Call</a>
--   * page's password-entropy hrefs 2 -> 1, the remainder being the FOOTER (a site_component,
--     which is retirement step 6, not this).
--
-- Repoints every remaining CTA whose LABEL is contact-intent and whose destination is the
-- retiring tool. Label untouched -- the copy already promises contact; only the destination
-- was wrong. Durable by KEEP #1 (bugs_open/248): 'contact' is in areasExcludedFromCTA,
-- /contact.html is live on all three sites, and no mint stamp names it for these fields.
-- __cta_minted deliberately untouched (url-specific predicate; the map merges shallowly).

BEGIN;

CREATE TEMP TABLE repointed(page_id uuid, site_id uuid, page_name text) ON COMMIT DROP;

WITH tgt AS (
  SELECT pc.id AS component_id, pc.page_id, p.site_id, p.name AS page_name, kv.key AS url_key
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  CROSS JOIN LATERAL jsonb_each(pc.content_data) kv
  WHERE kv.key LIKE '%\_url'
    AND kv.value #>> '{}' = '/tools/password-entropy.html'
    AND p.status = 'active'
    AND COALESCE(pc.content_data->>(replace(kv.key,'_url','_text')),
                 pc.content_data->>(replace(kv.key,'_url','_label')),
                 pc.content_data->>(replace(kv.key,'_url','')))
        ~* 'get in touch|contact|write to|email|call us|book a|talk to|speak to|discovery call'
),
upd AS (
  UPDATE page_components pc
  SET content_data = jsonb_set(pc.content_data, ARRAY[t.url_key], '"/contact.html"'::jsonb, false),
      updated_at = now()
  FROM tgt t
  WHERE pc.id = t.component_id
  RETURNING pc.page_id
)
INSERT INTO repointed SELECT DISTINCT u.page_id, t.site_id, t.page_name
FROM upd u JOIN tgt t ON t.page_id = u.page_id;

-- Publish: one no-LLM rerender per affected page. spec.page_name is LOAD-BEARING.
INSERT INTO site_work_items
  (site_id, page_id, source, created_by, item_type, handler_agent, status, severity, priority,
   summary, item_key, spec)
SELECT r.site_id, r.page_id, 'bugfix_391_cta_relevance', 'bugfix_391_cta_relevance',
       'page_rerender', 'page-rerender', 'detected', 'high', 30,
       'Owner decision 2026-08-31: publish the contact-page repoint on ' || r.page_name,
       'contact_repoint:' || r.page_id::text,
       jsonb_build_object(
         'reason','cta_links_stale', 'check','misdirected_cta',
         'page_id', r.page_id::text, 'page_name', r.page_name,
         'fix','The CTA copy asks the reader to get in touch; its destination is now /contact.html. Publish it and leave the authored contact destination in place.',
         'original_pipeline','build')
FROM repointed r
WHERE NOT EXISTS (SELECT 1 FROM site_work_items w
                  WHERE w.item_key = 'contact_repoint:' || r.page_id::text);

DO $$
DECLARE n_left int; n_contact int; n_items int; n_bad_stamp int;
BEGIN
  -- No contact-intent field may still point at the tool.
  SELECT count(*) INTO n_left
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  CROSS JOIN LATERAL jsonb_each(pc.content_data) kv
  WHERE kv.key LIKE '%\_url' AND kv.value #>> '{}' = '/tools/password-entropy.html'
    AND p.status='active'
    AND COALESCE(pc.content_data->>(replace(kv.key,'_url','_text')),
                 pc.content_data->>(replace(kv.key,'_url','_label')),
                 pc.content_data->>(replace(kv.key,'_url','')))
        ~* 'get in touch|contact|write to|email|call us|book a|talk to|speak to|discovery call';
  IF n_left <> 0 THEN
    RAISE EXCEPTION 'still % contact-intent field(s) pointing at the tool', n_left;
  END IF;

  SELECT count(*) INTO n_contact
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  CROSS JOIN LATERAL jsonb_each(pc.content_data) kv
  WHERE kv.key LIKE '%\_url' AND kv.value #>> '{}' = '/contact.html' AND p.status='active'
    AND p.site_id IN ('2a8ebf9c-20a2-4c39-b191-840b012371da','1368e337-dd1d-4799-bbb3-8221a1b79bcc');
  RAISE NOTICE 'contact-destination fields on the two affected sites: %', n_contact;

  -- KEEP #1 requires no mint stamp naming /contact.html for a REPOINTED field.
  -- ⚠ SCOPED TO THE PAGES THIS RUN TOUCHED. The first version of this guard counted such
  -- stamps FLEET-WIDE and aborted on 49 — which are pre-existing, correct, and none of my
  -- business: the resolver legitimately MINTS /contact.html via the label match, and a
  -- minted one is deliberately excluded from KEEP #1 so it stays re-derivable
  -- (storedCTADestinationIsAuthored's third clause). A guard asking about the fleet when
  -- the change is 13 pages is the same wrong-population error this lane has logged all week.
  SELECT count(*) INTO n_bad_stamp
  FROM page_components pc
  JOIN repointed r ON r.page_id = pc.page_id
  CROSS JOIN LATERAL jsonb_each(COALESCE(pc.content_data->'__cta_minted','{}'::jsonb)) m
  WHERE m.value #>> '{}' = '/contact.html';
  IF n_bad_stamp <> 0 THEN
    RAISE EXCEPTION 'a mint stamp names /contact.html on % repointed field(s) — KEEP #1 would not hold', n_bad_stamp;
  END IF;

  SELECT count(*) INTO n_items FROM site_work_items WHERE item_key LIKE 'contact_repoint:%';
  RAISE NOTICE 'rerenders queued (incl. canary): %', n_items;
END $$;

SELECT (SELECT count(*) FROM repointed) AS pages_repointed;
COMMIT;

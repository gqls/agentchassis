-- 2026-09-03 — copyonline.co.uk mission_brief REVISION 5. Owner ruling (evening): the lead route
-- "can point to a list of companies. I'd like to somehow measure that — maybe visits to that page
-- might do it sufficiently for now or make sure we have google attached."
-- Supersedes the DEFERRED revision 4 (organisations over owner; guard refused it mid-build, NOTES (kk)).
--
-- WHAT CHANGES (three edits, everything else byte-identical):
--   1. lead-route `what`: PRIMARY action = a list of UK copywriting COMPANIES (agencies/studios —
--      organisations, not named individuals); the enquiry form stays as the SECONDARY path and its
--      destination stays the owner's call. The page is the measured funnel: its visits are the metric.
--   2. entity-directory `what`: the directory lists companies first; sole traders only where they trade
--      under a business name with a public business presence. This is the organisations-over-people
--      answer to the removal/attribution question the brief itself raised.
--   3. open_questions: two entries recording the ruling and the measurement plan (GA4 page views on the
--      lead-route URL via the fleet container GTM-PQ3WCTBD → G-Y26N29T4KH, seeded for this site today;
--      Consent Mode v2 all-denied means visits count cookieless until Accept — an UNDERCOUNT by design,
--      fine for "is anyone arriving", not for attribution).
--
-- REACH, stated honestly: the classifier and planner cannot read a brief-writer mission_brief
-- (bugs_open/453, two template expressions) and the strategist has already run. So today this revision
-- is the OWNER'S RECORD of intent; it reaches the build when 453's fix ships or the strategist re-runs.
-- That is why this file does NOT add a `text` key — the per-site workaround was withdrawn in favour of
-- the template fix, and adding it here would fork the two.
BEGIN;
DO $g$
DECLARE n int; d jsonb;
BEGIN
  SELECT count(*) INTO n FROM site_specs WHERE site_id='3d965325-519a-4515-b79f-50c886954a80' AND aspect='mission_brief' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'REFUSED: expected 1 current brief, found %', n; END IF;
  SELECT data INTO d FROM site_specs WHERE site_id='3d965325-519a-4515-b79f-50c886954a80' AND aspect='mission_brief' AND is_current;
  IF (SELECT count(*) FROM jsonb_array_elements(d->'content_plan') it WHERE it->>'kind'='lead-route') <> 1 THEN RAISE EXCEPTION 'REFUSED: expected exactly 1 lead-route entry'; END IF;
  IF (SELECT count(*) FROM jsonb_array_elements(d->'content_plan') it WHERE it->>'kind'='entity-directory') <> 1 THEN RAISE EXCEPTION 'REFUSED: expected exactly 1 entity-directory entry'; END IF;
  IF jsonb_array_length(d->'content_plan') <> 30 THEN RAISE EXCEPTION 'REFUSED: content_plan has % entries, expected 30 (rev 3)', jsonb_array_length(d->'content_plan'); END IF;
  IF d ? 'text' THEN RAISE EXCEPTION 'REFUSED: brief already carries a text key — re-read before superseding'; END IF;
END $g$;

UPDATE site_specs SET is_current=false, superseded_at=now()
 WHERE site_id='3d965325-519a-4515-b79f-50c886954a80' AND aspect='mission_brief' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by, notes)
SELECT prev.site_id, 'mission_brief',
  jsonb_set(
    jsonb_set(
      prev.data,
      '{content_plan}',
      (SELECT jsonb_agg(
         CASE
           WHEN it->>'kind'='lead-route' THEN it || jsonb_build_object('what',
             'The one page in the navigation whose job is conversion, per the owner''s direction of 2026-09-03, so the rest of the site stays editorial. For a reader who has decided they want a human to write it. Its PRIMARY action, per the owner''s ruling of 2026-09-03 (evening), is a list of UK copywriting COMPANIES — agencies and studios, organisations rather than named individuals — drawn from the site''s directory, presented in randomised order, with a one-line summary and a link for each, and a plain statement that the list is compiled not vetted. Its SECONDARY path is a short enquiry form (what needs writing, audience, tone, deadline, budget band) whose destination is the owner''s decision and may change without altering this page''s purpose — initially the site owner, later possibly a named recipient who buys the leads; until that form is live the page says so plainly rather than pretending. This page is the site''s MEASURED funnel: its visits are the first metric the owner has asked for, so it must carry a stable URL, must not be split across variants, and must not be reachable only through other pages. It also receives referred traffic from elsewhere in the estate (webdesign.uk is intended to route copywriting leads here), so it must stand alone for a visitor who has never seen the rest of the site and must not assume they arrived via the guides.')
           WHEN it->>'kind'='entity-directory' THEN it || jsonb_build_object('what',
             'A directory of UK copywriting companies — agencies, studios and consultancies — compiled by the platform''s directory-researcher under a new ''copywriter'' entity kind. Organisations first, by the owner''s ruling of 2026-09-03 (evening): a sole trader is listed only where they trade under a business name with a public business presence, so that every entry is a business that expects to be found rather than a person listed on the strength of someone else''s marketplace profile. This is the organisations-over-people answer to the attribution and removal question this brief raised. Each entry shows the name, a short summary of what they do, specialisation and sector where known, and a link. The listing is RANDOMISED on each render so no entry gains from position, and the page states plainly how the list was compiled, that entries are not vetted or endorsed, and how a business gets listed or removed. This is the destination for the lead route''s primary path. It must not ship empty: if the compilation has not populated the kind, the page waits.')
           ELSE it
         END ORDER BY ord)
       FROM jsonb_array_elements(prev.data->'content_plan') WITH ORDINALITY AS t(it, ord))
    ),
    '{open_questions}',
    (prev.data->'open_questions') || jsonb_build_array(
      'RULED 2026-09-03 (evening): the lead route''s primary action is a LIST OF COMPANIES from the directory (organisations, not named individuals); the enquiry form is the secondary path and its destination remains the owner''s call. This supersedes the earlier ''initially the site owner'' framing as the primary path.',
      'MEASUREMENT (owner 2026-09-03): the lead-route page''s visits are the first metric. Seeded this site''s analytics container (GTM-PQ3WCTBD → GA4 G-Y26N29T4KH) on 2026-09-03 so every page it builds is tagged. Consent Mode v2 defaults all signals to denied, so visits are counted cookieless until a visitor accepts — an undercount by design, adequate for ''is anyone arriving'' and not for attribution. If the owner later wants clicks on the company links rather than page visits, that is a GA4 event on the outbound links, not a page change.'
    )
  ),
  'owner-revision', 'portfolio_positioning', true, 'portfolio_positioning',
  'Revision 5: lead route primary = list of companies (organisations), enquiry form secondary; directory lists companies first; measurement plan recorded. Owner ruling 2026-09-03 evening. Reaches the classifier/planner only once bugs_open/453 ships (they cannot read a brief-writer brief); the strategist has already run.'
FROM site_specs prev
WHERE prev.site_id='3d965325-519a-4515-b79f-50c886954a80' AND prev.aspect='mission_brief' AND prev.superseded_at IS NOT NULL
ORDER BY prev.superseded_at DESC LIMIT 1;

DO $v$
DECLARE d jsonb; n int;
BEGIN
  SELECT count(*) INTO n FROM site_specs WHERE site_id='3d965325-519a-4515-b79f-50c886954a80' AND aspect='mission_brief' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'VERIFY: % current rows', n; END IF;
  SELECT data INTO d FROM site_specs WHERE site_id='3d965325-519a-4515-b79f-50c886954a80' AND aspect='mission_brief' AND is_current;
  IF jsonb_array_length(d->'content_plan') <> 30 THEN RAISE EXCEPTION 'VERIFY: content_plan length %', jsonb_array_length(d->'content_plan'); END IF;
  IF NOT ((SELECT it->>'what' FROM jsonb_array_elements(d->'content_plan') it WHERE it->>'kind'='lead-route') ILIKE '%list of UK copywriting COMPANIES%') THEN RAISE EXCEPTION 'VERIFY: lead-route edit missing'; END IF;
  IF NOT ((SELECT it->>'what' FROM jsonb_array_elements(d->'content_plan') it WHERE it->>'kind'='entity-directory') ILIKE '%Organisations first%') THEN RAISE EXCEPTION 'VERIFY: directory edit missing'; END IF;
  IF jsonb_array_length(d->'open_questions') <> 12 THEN RAISE EXCEPTION 'VERIFY: open_questions length % (expected 12)', jsonb_array_length(d->'open_questions'); END IF;
  IF d ? 'text' THEN RAISE EXCEPTION 'VERIFY: text key must not be present'; END IF;
  -- everything else byte-identical to rev 3
  IF (d - 'content_plan' - 'open_questions') <> ((SELECT data FROM site_specs WHERE site_id='3d965325-519a-4515-b79f-50c886954a80' AND aspect='mission_brief' AND superseded_at IS NOT NULL ORDER BY superseded_at DESC LIMIT 1) - 'content_plan' - 'open_questions')
  THEN RAISE EXCEPTION 'VERIFY: a key outside content_plan/open_questions changed'; END IF;
END $v$;
COMMIT;

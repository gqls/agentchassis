-- SQL_2026-08-25b — the audience claim must ANSWER the question a customer asks.
--
-- WHY. Verified at the BOT (the lane's own rule), minutes after SQL_2026-08-25
-- landed: asked "Do I need web design experience to use this?", the bot answered
-- "You don't need any web design experience to use this" - the OPPOSITE of the
-- owner's position, while the same reply correctly used the new
-- not_a_hosting_company fact (so this is not cache lag). The mechanism is the
-- lane's trap #3: two claims collide and the permissive one answers the
-- customer's question. audience_experienced_webdesigners' own second half
-- ("and at anyone comfortable...") plus keep_it_online's "instructions walk you
-- through" outweighed its first half. Fix: the claim states the deciding arm in
-- the form the question takes ("is experience needed" -> yes, some), so no
-- reconciliation is available.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned
  FROM site_specs ss JOIN sites s ON s.id = ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{facts}', (
      SELECT jsonb_agg(
        CASE WHEN f->>'id' = 'audience_experienced_webdesigners' THEN
          jsonb_set(jsonb_set(jsonb_set(f,
            '{claim}', to_jsonb(
              $a$The offer is for experienced web designers, and for anyone comfortable editing and hosting a static site themselves. Asked whether experience is needed, the answer is yes, some: what is delivered is a ZIP of files, and from delivery the site is the customer's to host, to edit and to maintain, with written instructions included but nobody's time. Never tell a customer that no experience is needed. This bounds the buyer, not the site's subject: any_site_type still governs what the sites may be about.$a$::text)),
            '{context_terms}', jsonb_build_array('experienced','web designer','web designers','experience','technical','yours to host','yours to edit','maintain')),
            '{source,attested_by}', to_jsonb((f->'source'->>'attested_by') ||
              '; claim rewritten same day after bot verification: asked "do I need experience", the bot answered "you do not need any" by reconciling toward the permissive halves of this claim and keep_it_online (lane trap #3), so the claim now states the deciding arm in the question''s own form'))
        ELSE f END ORDER BY ord)
      FROM jsonb_array_elements(c.data->'facts') WITH ORDINALITY AS t(f, ord)
    )) AS newdata
  FROM cur c
),
retire AS (
  UPDATE site_specs ss SET is_current=false, superseded_at=now()
  WHERE ss.id=(SELECT id FROM cur) RETURNING 1
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
  'SQL_2026-08-25b: audience_experienced_webdesigners claim rewritten to answer the "is experience needed" question directly (bot verification caught the permissive reading winning). One claim + context_terms + source; nothing else.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $chk$
DECLARE d jsonb; prev jsonb; n int; f jsonb; pf jsonb;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF prev IS NULL THEN RAISE EXCEPTION 'no superseded row'; END IF;

  SELECT count(*) INTO n FROM (
    SELECT x->>'id' FROM jsonb_array_elements(prev->'facts') AS t(x)
    EXCEPT SELECT x->>'id' FROM jsonb_array_elements(d->'facts') AS t(x)) q;
  IF n <> 0 THEN RAISE EXCEPTION '% fact id(s) vanished', n; END IF;
  SELECT count(*) INTO n FROM (
    SELECT x->>'id' FROM jsonb_array_elements(d->'facts') AS t(x)
    EXCEPT SELECT x->>'id' FROM jsonb_array_elements(prev->'facts') AS t(x)) q;
  IF n <> 0 THEN RAISE EXCEPTION '% fact id(s) appeared', n; END IF;

  SELECT x INTO f  FROM jsonb_array_elements(d->'facts')    AS t(x) WHERE x->>'id'='audience_experienced_webdesigners';
  SELECT x INTO pf FROM jsonb_array_elements(prev->'facts') AS t(x) WHERE x->>'id'='audience_experienced_webdesigners';
  IF strpos(pf->>'claim','aimed at experienced web designers')=0
    THEN RAISE EXCEPTION 'control: superseded claim is not the 2026-08-25 first cut'; END IF;
  IF strpos(f->>'claim','the answer is yes, some')=0 OR strpos(f->>'claim','Never tell a customer that no experience is needed.')=0
    THEN RAISE EXCEPTION 'new claim lacks the deciding arm'; END IF;
  IF f->>'writer_line' <> pf->>'writer_line' THEN RAISE EXCEPTION 'writer_line changed; this write must not touch it'; END IF;

  -- every OTHER fact byte-identical
  SELECT count(*) INTO n
  FROM jsonb_array_elements(d->'facts') AS a(x)
  JOIN jsonb_array_elements(prev->'facts') AS b(y) ON a.x->>'id' = b.y->>'id'
  WHERE a.x <> b.y AND a.x->>'id' <> 'audience_experienced_webdesigners';
  IF n <> 0 THEN RAISE EXCEPTION '% OTHER fact(s) changed', n; END IF;

  IF d->'writer_block' <> prev->'writer_block' THEN RAISE EXCEPTION 'writer_block changed'; END IF;
  IF d->'banned_claims' <> prev->'banned_claims' THEN RAISE EXCEPTION 'banned_claims changed'; END IF;
  IF d->'allowed_entities' <> prev->'allowed_entities' THEN RAISE EXCEPTION 'allowed_entities changed'; END IF;

  RAISE NOTICE 'ALL GUARDS PASSED: one claim rewritten, everything else identical';
END $chk$;

COMMIT;

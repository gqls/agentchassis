-- SQL_2026-08-25c — a figure the pages may print must exist as a fact VALUE.
--
-- WHY. SQL_2026-08-25 fixed the included month at "30 days" in two claims and five
-- writer_block sentences, and the writer duly wrote it. Three of five rebuilds then
-- failed at validate_content with "0 blockers, N errors": every error was
-- `unregistered_number "30"` / `unregistered_stat "30 days"` - the claims gate checks
-- each number in copy (and every stat field) against evidence_base fact VALUES, and
-- no fact carried value 30. build_duration carries value 4 and price_total 149 for
-- exactly this reason; the writer_block's own STAT-FIELDS paragraph says so. The
-- lane's rule "verify at the bot" was followed and is NOT the gate: the bot reads
-- claims, the gate reads values. claimscan would have caught it before a build.
-- (what-you-get and the guide passed with "30 days" present - the checker's prose
-- pass evidently tolerated their phrasing; the fact value makes all pages consistent.)
--
-- ALSO: the live ban (template|templated|off.the.shelf|cookie.cutter) - "do not
-- describe the product this way even to deny it" - collides with the owner's draft
-- phrase "starter template", which SQL_2026-08-25 copied into writer_block. Zero
-- blockers today means the writer avoided it, but an instruction that contains a
-- banned word is the documented prompt-example trap. The paragraph now says
-- "starter site"; whether the BAN lifts is the owner's call, flagged in chat.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned
  FROM site_specs ss JOIN sites s ON s.id = ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(
      jsonb_set(c.data, '{facts}',
        (c.data->'facts') || jsonb_build_array(jsonb_build_object(
          'id','live_link_days',
          'kind','metric',
          'value', 30,
          'claim','The link to the finished site stays live for 30 days after delivery; keeping it online after that is the customer''s own hosting (keep_it_online).',
          'source', jsonb_build_object('attested_by',
            'owner, 2026-08-25, copy brief ("You can see what we''ve built for 30 days at a link we''ll give you"); split out as a metric fact the same day because the claims gate matches printed figures against fact VALUES and the prose-only attestation in delivery_live_link_and_zip / keep_it_online failed three rebuilds with unregistered_number 30'),
          'verified_at','2026-08-25',
          'writer_line','30 days',
          'context_terms', jsonb_build_array('30 days','days','live','link','stays live')
        ))),
      '{writer_block}',
      to_jsonb(replace(c.data->>'writer_block',
        'These starter sites are exactly that: a more or less complete starter template for the customer to carry on with.',
        'These starter sites are exactly that: a more or less complete starter site for the customer to carry on with. (The word template is banned on this site, even in that sense: say starter site.)'))
    ) AS newdata
  FROM cur c
),
retire AS (
  UPDATE site_specs ss SET is_current=false, superseded_at=now()
  WHERE ss.id=(SELECT id FROM cur) RETURNING 1
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
  'SQL_2026-08-25c: metric fact live_link_days value 30 (the gate matches printed figures to fact VALUES; three rebuilds failed unregistered_number 30); writer_block starter-site paragraph loses the banned word template. Nothing else.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $chk$
DECLARE d jsonb; prev jsonb; n int; wb text; pwb text;
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
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 new fact, got %', n; END IF;
  PERFORM 1 FROM jsonb_array_elements(d->'facts') AS t(x)
   WHERE x->>'id'='live_link_days' AND (x->>'value')::int = 30 AND x->>'kind'='metric' AND x->>'writer_line'='30 days';
  IF NOT FOUND THEN RAISE EXCEPTION 'live_link_days absent or lacks value 30'; END IF;
  -- control: no PRIOR fact carried 30 (the whole reason this file exists)
  PERFORM 1 FROM jsonb_array_elements(prev->'facts') AS t(x) WHERE x->>'value' = '30';
  IF FOUND THEN RAISE EXCEPTION 'control: a superseded fact already carried value 30'; END IF;
  -- every pre-existing fact byte-identical
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'facts') AS a(x)
   JOIN jsonb_array_elements(prev->'facts') AS b(y) ON a.x->>'id' = b.y->>'id' WHERE a.x <> b.y;
  IF n <> 0 THEN RAISE EXCEPTION '% existing fact(s) changed', n; END IF;

  wb := d->>'writer_block'; pwb := prev->>'writer_block';
  IF strpos(pwb,'starter template')=0 THEN RAISE EXCEPTION 'control: superseded block lacked the phrase'; END IF;
  IF strpos(wb,'starter template')>0 THEN RAISE EXCEPTION 'starter template survives'; END IF;
  IF strpos(wb,'say starter site.')=0 THEN RAISE EXCEPTION 'replacement sentence missing'; END IF;
  IF (length(wb)-length(replace(wb,'—',''))) <> (length(pwb)-length(replace(pwb,'—',''))) THEN RAISE EXCEPTION 'em-dash count moved'; END IF;
  IF d->'banned_claims' <> prev->'banned_claims' THEN RAISE EXCEPTION 'banned_claims changed'; END IF;
  IF d->'allowed_entities' <> prev->'allowed_entities' THEN RAISE EXCEPTION 'allowed_entities changed'; END IF;
  RAISE NOTICE 'ALL GUARDS PASSED: live_link_days value 30 added; starter template -> starter site';
END $chk$;

COMMIT;

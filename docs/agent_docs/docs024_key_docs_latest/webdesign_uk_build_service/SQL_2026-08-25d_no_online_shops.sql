-- SQL_2026-08-25d — OWNER RULINGS 2026-08-25 (evening), three in one message:
--   "template can remain a banned word. .zip is better. I don't think we can do
--    online shops yet so we can exclude that."
--
-- (1) template STAYS BANNED: no data change; SQL_2026-08-25c already made writer_block
--     say "starter site". Recorded here and in NOTES so nobody lifts it "because the
--     owner's draft said template".
-- (2) .zip CONFIRMED: the audience fact's source note gains the confirmation (it had
--     recorded the reading as flagged-not-confirmed).
-- (3) ONLINE SHOPS EXCLUDED: attested as a CAPABILITY limit inside any_site_type (the
--     categories fact), distinct from the content limits in content_we_will_not_build.
--     Phrased with "do not" in the clause per the gate's negation guard (a bare "no
--     online shops" is the documented intensifier trap). writer_block's CATEGORIES
--     paragraph and its facts enumeration gain the same line.
--
-- Deliberately NOT a ban pattern: banning "shop" would block the denial too (the
-- 2026-08-19 offer-shape precedent). The claim licenses the denial; nothing licenses a
-- promise, which is what the governing rule already enforces.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned
  FROM site_specs ss JOIN sites s ON s.id = ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(
      jsonb_set(c.data, '{facts}', (
        SELECT jsonb_agg(
          CASE f->>'id'
            WHEN 'any_site_type' THEN
              jsonb_set(jsonb_set(jsonb_set(f,
                '{claim}', to_jsonb(replace(f->>'claim',
                  'It is not unlimited, and the limits are content limits rather than purpose limits: see content_we_will_not_build.',
                  'It is not unlimited: the content limits are in content_we_will_not_build, and one capability limit stands on its own: we do not build online shops that take payment, because a delivered static site has no payment machinery.'))),
                '{writer_line}', to_jsonb((f->>'writer_line') || ' We do not build online shops.')),
                '{source,attested_by}', to_jsonb((f->'source'->>'attested_by') ||
                  '; owner, 2026-08-25 (evening): "I don''t think we can do online shops yet so we can exclude that" - attested as a capability limit, with "do not" in the clause for the gate'))
            WHEN 'audience_experienced_webdesigners' THEN
              jsonb_set(f, '{source,attested_by}', to_jsonb((f->'source'->>'attested_by') ||
                '; the ZIP reading CONFIRMED by the owner 2026-08-25 (evening): ".zip is better"'))
            ELSE f
          END ORDER BY ord)
        FROM jsonb_array_elements(c.data->'facts') WITH ORDINALITY AS t(f, ord)
      )),
      '{writer_block}',
      to_jsonb(replace(replace(c.data->>'writer_block',
        'Name categories, never example sites.',
        'Name categories, never example sites. One thing is not on offer and is said plainly beside them, in the flat voice (owner ruling, 2026-08-25): we do not build online shops that take payment, because a delivered static site has no payment machinery. Write it with do not in the same clause, never as a bare no.'),
        'that we are not a hosting company, that no pre-sales service is included,',
        'that we are not a hosting company, that we do not build online shops that take payment, that no pre-sales service is included,'))
    ) AS newdata
  FROM cur c
),
retire AS (
  UPDATE site_specs ss SET is_current=false, superseded_at=now()
  WHERE ss.id=(SELECT id FROM cur) RETURNING 1
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
  'SQL_2026-08-25d: online shops excluded as a capability limit (any_site_type claim + writer_line; writer_block CATEGORIES paragraph + facts enumeration); ZIP reading confirmed on the audience fact source. template stays banned (no change).',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $chk$
DECLARE d jsonb; prev jsonb; n int; wb text; pwb text; f jsonb; pf jsonb;
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

  SELECT x INTO f  FROM jsonb_array_elements(d->'facts')    AS t(x) WHERE x->>'id'='any_site_type';
  SELECT x INTO pf FROM jsonb_array_elements(prev->'facts') AS t(x) WHERE x->>'id'='any_site_type';
  IF strpos(pf->>'claim','the limits are content limits rather than purpose limits')=0
    THEN RAISE EXCEPTION 'control: superseded any_site_type lacked the needle'; END IF;
  IF strpos(f->>'claim','we do not build online shops that take payment')=0 OR strpos(f->>'claim','No example sites are named yet')=0
    THEN RAISE EXCEPTION 'any_site_type: shops line missing or no-examples rule lost'; END IF;
  IF f->>'writer_line' <> (pf->>'writer_line') || ' We do not build online shops.'
    THEN RAISE EXCEPTION 'any_site_type writer_line reconstruction failed'; END IF;

  PERFORM 1 FROM jsonb_array_elements(d->'facts') AS t(x)
   WHERE x->>'id'='audience_experienced_webdesigners' AND x->'source'->>'attested_by' LIKE '%CONFIRMED by the owner 2026-08-25%';
  IF NOT FOUND THEN RAISE EXCEPTION 'audience fact source lacks the ZIP confirmation'; END IF;

  SELECT count(*) INTO n FROM jsonb_array_elements(d->'facts') AS a(x)
   JOIN jsonb_array_elements(prev->'facts') AS b(y) ON a.x->>'id' = b.y->>'id'
   WHERE a.x <> b.y AND a.x->>'id' NOT IN ('any_site_type','audience_experienced_webdesigners');
  IF n <> 0 THEN RAISE EXCEPTION '% OTHER fact(s) changed', n; END IF;

  wb := d->>'writer_block'; pwb := prev->>'writer_block';
  IF strpos(pwb,'online shops')>0 THEN RAISE EXCEPTION 'control: superseded block already mentioned online shops'; END IF;
  IF (length(wb)-length(replace(wb,'we do not build online shops that take payment','')))/length('we do not build online shops that take payment') <> 2
    THEN RAISE EXCEPTION 'writer_block must carry the shops line exactly twice (paragraph + enumeration)'; END IF;
  IF (length(wb)-length(replace(wb,'—',''))) <> (length(pwb)-length(replace(pwb,'—',''))) THEN RAISE EXCEPTION 'em-dash count moved'; END IF;
  IF d->'banned_claims' <> prev->'banned_claims' THEN RAISE EXCEPTION 'banned_claims changed (template must STAY banned, and nothing else moves)'; END IF;
  IF NOT EXISTS (SELECT 1 FROM jsonb_array_elements(d->'banned_claims') b WHERE b->>'pattern' LIKE '%template%')
    THEN RAISE EXCEPTION 'control: the template ban is not present'; END IF;
  IF d->'allowed_entities' <> prev->'allowed_entities' THEN RAISE EXCEPTION 'allowed_entities changed'; END IF;
  RAISE NOTICE 'ALL GUARDS PASSED: online shops excluded (claim, writer_line, writer_block x2); ZIP confirmed; template ban intact';
END $chk$;

COMMIT;

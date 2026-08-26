-- SQL_2026-08-26e — OWNER RULING 2026-08-26 (night): the domain buy-out moves
-- £200 → £59.99 ("the domain price is incongruent with the cheap website pricing...
-- Let's make it £59.99"), rental stays £10/mo. Arithmetic surfaced before ruling:
-- every sale carries the Registrant Transfer (£10–35+VAT, verified 2026-08-21) +
-- ~£4 registration, so £59.99 nets ~£44 typical.
--
-- Every £200 in the live specs IS this price (censused: evidence_base x6 = the
-- domain_buy_once fact + writer_block; identity x1; briefing x1; strategy x1), so
-- per-aspect global replaces are exact, with count guards. The fact also gets
-- value 59.99 (the claims gate matches printed figures to fact VALUES - the
-- 30-days lesson) and the ruling appended to its source.

BEGIN;

-- evidence_base: the fact (claim, writer_line, value, source) + writer_block
WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(
      jsonb_set(c.data, '{facts}', (
        SELECT jsonb_agg(
          CASE WHEN f->>'id'='domain_buy_once' THEN
            jsonb_set(jsonb_set(jsonb_set(jsonb_set(f,
              '{claim}', to_jsonb(replace(f->>'claim','£200','£59.99'))),
              '{writer_line}', to_jsonb(replace(f->>'writer_line','£200','£59.99'))),
              '{value}', '59.99'::jsonb),
              '{source,attested_by}', to_jsonb((f->'source'->>'attested_by') ||
                '; owner, 2026-08-26 night: £200 was incongruent with the £149 site price - moved to £59.99, rental unchanged at £10/mo (transfer-fee arithmetic surfaced first: £10-35+VAT per sale, verified 2026-08-21)'))
          ELSE f END ORDER BY ord)
        FROM jsonb_array_elements(c.data->'facts') WITH ORDINALITY AS t(f, ord)
      )),
      '{writer_block}', to_jsonb(replace(c.data->>'writer_block','£200','£59.99'))
    ) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, 'evidence_base', r.newdata, 'owner-ruling',
  'SQL_2026-08-26e: domain buy-out £200 -> £59.99 (fact value 59.99 + writer_block); owner ruling 2026-08-26 night.',
  true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

-- identity, briefing, strategy: the one £200 each is the same price
WITH cur AS (
  SELECT ss.id, ss.site_id, ss.aspect AS asp, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id=ss.site_id
  WHERE s.domain='webdesign.uk' AND ss.aspect IN ('identity','briefing','strategy') AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.asp, c.pinned, replace(c.data::text,'£200','£59.99')::jsonb AS newdata, c.id FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id IN (SELECT id FROM cur) RETURNING 1),
retired AS (SELECT count(*) AS c FROM retire)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id, r.asp, r.newdata, 'owner-ruling',
  'SQL_2026-08-26e: domain buy-out £200 -> £59.99.', true, 'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retired WHERE retired.c = 3;

DO $chk$
DECLARE n int; v numeric; t text;
BEGIN
  -- £200 may survive ONLY inside fact SOURCES (historical owner quotes stay);
  -- every writer/bot-visible surface must be clean.
  SELECT count(*) INTO n FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
   jsonb_array_elements(ss.data->'facts') AS t3(f)
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
     AND (COALESCE(f->>'claim','') LIKE '%£200%' OR COALESCE(f->>'writer_line','') LIKE '%£200%');
  IF n <> 0 THEN RAISE EXCEPTION '% fact claim/writer_line still says £200', n; END IF;
  SELECT count(*) INTO n FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.is_current
     AND ((ss.aspect='evidence_base' AND ss.data->>'writer_block' LIKE '%£200%')
       OR (ss.aspect IN ('identity','briefing','strategy') AND ss.data::text LIKE '%£200%'));
  IF n <> 0 THEN RAISE EXCEPTION '% visible surface(s) still say £200', n; END IF;
  SELECT COALESCE(sum((length(ss.data::text)-length(replace(ss.data::text,'£59.99','')))/6),0) INTO n
   FROM site_specs ss JOIN sites s ON s.id=ss.site_id WHERE s.domain='webdesign.uk' AND ss.is_current;
  IF n <> 9 THEN RAISE EXCEPTION 'expected 9 £59.99 mentions (claim+writer_line+source-append+wb3+id+br+st), found %', n; END IF;
  SELECT (f->>'value')::numeric INTO v FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
   jsonb_array_elements(ss.data->'facts') AS t2(f)
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current AND f->>'id'='domain_buy_once';
  IF v <> 59.99 THEN RAISE EXCEPTION 'domain_buy_once value is %, want 59.99', v; END IF;
  SELECT ss.data->>'writer_block' INTO t FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  IF strpos(t,'£10 a month') = 0 AND strpos(t,'£10 per month') = 0 THEN RAISE EXCEPTION 'rental tenner lost from writer_block'; END IF;
  RAISE NOTICE 'ALL GUARDS PASSED: buy-out is £59.99 everywhere, rental untouched';
END $chk$;

COMMIT;

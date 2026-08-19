-- SQL_2026-08-19 — re-attest build_duration: "usually ready in two or three days".
--
-- OWNER, 2026-08-19: *"Delivery can be '2 or 3 days'"*, following his ruling of
-- 2026-08-18: *"I'd rather change the estimated delivery time if it means a better
-- product for the customer."*
--
-- WHY THE OLD FIGURE HAD TO GO — measured, not assumed. "Usually ready the next
-- day" (value 1, attested 2026-08-14) was refuted by the only two sites built under
-- the current triage-based flow:
--   * remortgagecalculator.uk, 25.3h after build: 48 work items, 22 STILL OPEN,
--     including needs_page x4 and needs_new_component x3. A site missing four pages
--     is not deliverable.
--   * loanzy.uk, 3.1h after build: 15 open, 9 of them HIGH, including
--     site_unreachable and unbuilt_internal_link x3.
-- Page CREATION is not the cost (all pages appear in one 0.0h batch); the triage
-- tail is, and it runs past a day. The live chat bot renders this claim verbatim,
-- so the old figure was a promise to customers the evidence contradicted.
--
-- WHY value = 3 AND NOT 2. `value` is what the stat guard lets a writer publish in
-- a stat field as a bare figure. The owner's attestation is a RANGE, and a range
-- cannot be a single number, so the number has to be the end that cannot
-- over-promise: a stat reading "3 days" is inside "2 or 3 days", a stat reading
-- "2 days" is not. The hedge itself lives in claim and writer_line, which is where
-- the 2026-08-18 "1 day" lesson said hedges belong.
-- context_terms gains "days" so the figure can only ever license a turnaround
-- sentence, never a bare "3" somewhere else.
--
-- ⚠ AFTER THIS: FOUR pages still carry "next day" in stored components and must be
-- rebuilt or they will contradict the bot -- faq (2 components), how-it-works (2),
-- index (2), tool-website-brief-starter-guide (1). Queued separately; a register
-- change alone does not touch published copy.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{facts}', (
      SELECT jsonb_agg(
               CASE WHEN f->>'id' = 'build_duration' THEN
                 jsonb_build_object(
                   'id','build_duration',
                   'kind', f->'kind',
                   'claim','From having what is needed from the customer, the site is usually ready in two or three days.',
                   'value', 3,
                   'source', jsonb_build_object('attested_by',
                     'owner, 2026-08-19 (''Delivery can be "2 or 3 days"'') - supersedes the next-day figure attested 2026-08-14, which measurement refuted: under the triage-based build flow remortgagecalculator.uk still had 22 open work items at 25.3h including 4 needs_page. Follows the owner ruling of 2026-08-18 that a better product beats a faster promise.'),
                   'verified_at','2026-08-19',
                   'writer_line','usually ready in two or three days',
                   'context_terms', jsonb_build_array('turnaround','day','days','ready')
                 )
               ELSE f END ORDER BY ord)
      FROM jsonb_array_elements(c.data->'facts') WITH ORDINALITY AS t(f, ord)
    )) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'owner-ruling',
  'SQL_2026-08-19: build_duration re-attested to two or three days (owner 2026-08-19). Facts and bans otherwise unchanged.',
  true,'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; prev jsonb; n int; bd jsonb;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF prev IS NULL THEN RAISE EXCEPTION 'no superseded row to compare against'; END IF;

  -- nothing lost by THIS write (no fixed count: two lanes edit this row)
  SELECT count(*) INTO n FROM (
    (SELECT f->>'id' FROM jsonb_array_elements(prev->'facts') f
     EXCEPT SELECT f->>'id' FROM jsonb_array_elements(d->'facts') f)
    UNION ALL
    (SELECT f->>'id' FROM jsonb_array_elements(d->'facts') f
     EXCEPT SELECT f->>'id' FROM jsonb_array_elements(prev->'facts') f)) x;
  IF n <> 0 THEN RAISE EXCEPTION '% fact id(s) differ - this write changes ONE fact', n; END IF;
  IF jsonb_array_length(prev->'banned_claims') <> jsonb_array_length(d->'banned_claims')
    THEN RAISE EXCEPTION 'banned_claims count moved'; END IF;
  IF (prev->>'writer_block') IS DISTINCT FROM (d->>'writer_block')
    THEN RAISE EXCEPTION 'writer_block changed - it must not'; END IF;

  SELECT f INTO bd FROM jsonb_array_elements(d->'facts') f WHERE f->>'id'='build_duration';
  IF bd IS NULL THEN RAISE EXCEPTION 'build_duration vanished'; END IF;
  IF (bd->>'value')::int <> 3 THEN RAISE EXCEPTION 'value is % not 3 (the upper bound is the only safe stat figure for a range)', bd->>'value'; END IF;
  IF bd->>'claim' !~* 'two or three days' THEN RAISE EXCEPTION 'claim does not carry the new range'; END IF;
  IF bd->>'writer_line' !~* 'two or three days' THEN RAISE EXCEPTION 'writer_line does not carry the new range'; END IF;
  IF bd::text ~* 'next day' THEN RAISE EXCEPTION 'the retired next-day wording survives in the fact'; END IF;
  IF NOT (bd->'context_terms' @> '["days"]'::jsonb) THEN RAISE EXCEPTION 'context_terms lost "days"'; END IF;

  -- no OTHER fact may still promise next-day
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'facts') f WHERE f->>'claim' ~* 'next day';
  IF n <> 0 THEN RAISE EXCEPTION '% other fact claim(s) still say "next day"', n; END IF;
END $$;

COMMIT;

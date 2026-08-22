-- SQL_2026-08-22 — re-attest build_duration to "three or four days, usually sooner",
-- update writer_block to match, and RETIRE the ban that would otherwise block the
-- owner's own new figure.
--
-- OWNER, 2026-08-22: *"two or three should probably be 3 or 4 but usually sooner."*
-- Exact wording confirmed with him the same day before writing this.
--
-- This answers **D-B** (`HANDOFF_2026-08-21` §2), which asked whether the promise
-- absorbs the owner's pre-delivery review step or gets re-cut. It is re-cut: option
-- (b) of `DECISION_2026-08-21e` §3. The review gate he asked for on 2026-08-21 spends
-- build-time budget, and the honest response is to widen the promise rather than to
-- let a human step silently break it.
--
-- ⚠ THE THING THAT MAKES THIS NOT A ONE-LINE EDIT: THE NEW FIGURE IS CURRENTLY BANNED.
-- `banned_claims` carries, armed 2026-08-14 and still live today:
--     \bthree (or|to) four days\b|\b3[-–]4 days\b|\bthree[-–]to[-–]four\b
--     "RETIRED FIGURE (owner 2026-08-14): three-or-four-days belonged to the £1,200 offer."
-- The figure is being reinstated for the £149 offer, so that ban must go in the SAME
-- transaction as the re-attestation. Retire the fact without retiring the ban and the
-- claims gate refuses our own new copy at deploy time -- and the failure presents as a
-- broken writer, not as a register contradicting itself.
--
-- WHY value = 4 AND NOT 3. Unchanged reasoning from SQL_2026-08-19: `value` is what the
-- stat guard lets a writer publish as a bare figure, a range cannot be a single number,
-- so the number must be the end that cannot over-promise. A stat reading "4 days" is
-- inside "three or four days"; "3 days" is not. The hedge lives in claim and
-- writer_line. `context_terms` gains "sooner" because the hedge now uses it.
--
-- ⚠ WHAT THIS FILE DELIBERATELY DOES **NOT** DO: it does not arm a ban on the retired
-- "two or three days". That is `SQL_2026-08-22b_..._HOLD.sql` and it MUST NOT run yet.
-- SQL_2026-08-19e's ordering rule, which is the bugs_open/161 landmine: a banned_claims
-- entry is BLOCKER severity, so arming one while offending copy is still STORED makes
-- every affected page refuse to save -- the falsehood stays published AND becomes
-- unfixable through the normal rewrite path. index carries "two or three days" 6x and
-- faq 3x as of 2026-08-22 (measured at preview.webdesign.uk, with a control). So:
--   1. this file            (makes the new figure legal)
--   2. re-render index+faq  (removes the old figure from stored components)
--   3. the _HOLD file       (arms the ban, only once the census is zero)

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(
      jsonb_set(
        jsonb_set(c.data, '{facts}', (
          SELECT jsonb_agg(
                   CASE WHEN f->>'id' = 'build_duration' THEN
                     jsonb_build_object(
                       'id','build_duration',
                       'kind', f->'kind',
                       'claim','From having what is needed from the customer, the site is usually ready in three or four days, often sooner.',
                       'value', 4,
                       'source', jsonb_build_object('attested_by',
                         'owner, 2026-08-22 (''two or three should probably be 3 or 4 but usually sooner'') - supersedes the two-or-three-days figure attested 2026-08-19. This is D-B: the owner''s pre-delivery review step (DECISION_2026-08-21e) spends build-time budget, so the promise is re-cut to absorb it rather than being broken by the gate. Same direction as his 2026-08-18 ruling that a better product beats a faster promise. NOTE: three-or-four-days was itself a RETIRED figure from the GBP1,200 offer, banned 2026-08-14; that ban is retired by this same transaction.'),
                       'verified_at','2026-08-22',
                       'writer_line','ready in three or four days, usually sooner',
                       'context_terms', jsonb_build_array('turnaround','day','days','ready','sooner')
                     )
                   ELSE f END ORDER BY ord)
          FROM jsonb_array_elements(c.data->'facts') WITH ORDINALITY AS t(f, ord)
        )),
        '{banned_claims}', (
          SELECT COALESCE(jsonb_agg(
                   CASE WHEN b->>'reason' ~ 'two or three days'
                     THEN jsonb_set(b, '{reason}', to_jsonb(
                            replace(b->>'reason',
                              'The attested build time is two or three days (fact build_duration, re-attested by the owner 2026-08-19)',
                              'The attested build time is three or four days, usually sooner (fact build_duration, re-attested by the owner 2026-08-22)')))
                     ELSE b END ORDER BY ord), '[]'::jsonb)
          FROM jsonb_array_elements(c.data->'banned_claims') WITH ORDINALITY AS t(b, ord)
          -- RETIRE the three-or-four-days ban: the owner has re-attested the figure.
          -- Matched by LITERAL SUBSTRING on a pure-ASCII fragment. NOT by string equality
          -- (the stored pattern carries en-dashes, hostage to the encoding of every pipe this
          -- file travels down) and NOT by regex: THE STORED VALUE IS ITSELF A REGEX, so
          -- `~ 'three (or|to) four days'` reads its literal parens and pipe as alternation
          -- syntax and matches nothing. That cost a dry-run cycle. The guard below asserts the
          -- superseded row held EXACTLY ONE such ban, so an over-match cannot pass silently.
          WHERE strpos(b->>'pattern', 'three (or|to) four days') = 0
        )),
      '{writer_block}', to_jsonb(
        replace(replace(c.data->>'writer_block',
          'Say how long it takes: usually ready in two or three days from having what is needed.',
          'Say how long it takes: ready in three or four days, usually sooner, from having what is needed.'),
          'The build duration is HEDGED ("usually ready in two or three days") and is a range, not a number: never render it as a stat, a counter, or a bare figure such as "3 days" or "72 hours".',
          'The build duration is HEDGED ("ready in three or four days, usually sooner") and is a range, not a number: never render it as a stat, a counter, or a bare figure such as "4 days" or "96 hours".')
      )) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'owner-ruling',
  'SQL_2026-08-22: build_duration re-attested to three or four days, usually sooner (owner 2026-08-22, answering D-B). Retires the 2026-08-14 ban on that same figure, which would otherwise block the new copy. writer_block updated in two places. Ban on "two or three days" deliberately NOT armed here - see the _HOLD file, pages must be re-rendered first (SQL_2026-08-19e ordering rule / bugs_open/161).',
  true,'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; prev jsonb; n int; bd jsonb; wb text; pwb text; expect text;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF prev IS NULL THEN RAISE EXCEPTION 'no superseded row to compare against'; END IF;

  -- ── facts: exactly one changes, none appear or vanish (no fixed count: two lanes edit this row)
  SELECT count(*) INTO n FROM (
    (SELECT f->>'id' FROM jsonb_array_elements(prev->'facts') f
     EXCEPT SELECT f->>'id' FROM jsonb_array_elements(d->'facts') f)
    UNION ALL
    (SELECT f->>'id' FROM jsonb_array_elements(d->'facts') f
     EXCEPT SELECT f->>'id' FROM jsonb_array_elements(prev->'facts') f)) x;
  IF n <> 0 THEN RAISE EXCEPTION '% fact id(s) differ - this write changes ONE fact in place', n; END IF;

  SELECT count(*) INTO n FROM jsonb_array_elements(d->'facts') f
   WHERE f->>'id' <> 'build_duration'
     AND f IS DISTINCT FROM (SELECT p FROM jsonb_array_elements(prev->'facts') p WHERE p->>'id' = f->>'id');
  IF n <> 0 THEN RAISE EXCEPTION '% fact(s) other than build_duration changed', n; END IF;

  SELECT f INTO bd FROM jsonb_array_elements(d->'facts') f WHERE f->>'id'='build_duration';
  IF bd IS NULL THEN RAISE EXCEPTION 'build_duration vanished'; END IF;
  IF (bd->>'value')::int <> 4 THEN RAISE EXCEPTION 'value is % not 4 (upper bound is the only safe stat figure for a range)', bd->>'value'; END IF;
  IF bd->>'claim' !~* 'three or four days' THEN RAISE EXCEPTION 'claim does not carry the new range'; END IF;
  IF bd->>'claim' !~* 'sooner' THEN RAISE EXCEPTION 'claim lost the owner''s "usually sooner" hedge'; END IF;
  IF bd->>'writer_line' !~* 'three or four days' THEN RAISE EXCEPTION 'writer_line does not carry the new range'; END IF;
  IF bd->>'writer_line' !~* 'sooner' THEN RAISE EXCEPTION 'writer_line lost the hedge'; END IF;
  -- The retired range must not survive in any field a READER sees. Scoped to the
  -- consumer-facing fields on purpose, and the scoping is load-bearing: `source.attested_by`
  -- quotes the owner's own words ("two or three should probably be 3 or 4..."), which is
  -- provenance and must stay verbatim. A whole-fact `bd::text` check fires on that quote --
  -- it did, on the first dry run -- which would have pushed the next author to launder the
  -- attestation rather than fix the copy. claim is what the bot reads verbatim; writer_line
  -- is what the page writer reads; context_terms is what licenses the figure.
  IF bd->>'claim' ~* 'two or three' THEN RAISE EXCEPTION 'the retired range survives in claim (the bot reads this verbatim)'; END IF;
  IF bd->>'writer_line' ~* 'two or three' THEN RAISE EXCEPTION 'the retired range survives in writer_line'; END IF;
  IF (bd->'context_terms')::text ~* 'two or three' THEN RAISE EXCEPTION 'the retired range survives in context_terms'; END IF;
  IF NOT (bd->'context_terms' @> '["days"]'::jsonb) THEN RAISE EXCEPTION 'context_terms lost "days"'; END IF;
  IF NOT (bd->'context_terms' @> '["sooner"]'::jsonb) THEN RAISE EXCEPTION 'context_terms did not gain "sooner"'; END IF;

  -- no OTHER fact may still promise the retired range
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'facts') f WHERE f->>'claim' ~* 'two or three days';
  IF n <> 0 THEN RAISE EXCEPTION '% other fact claim(s) still say "two or three days"', n; END IF;

  -- ── bans: exactly ONE retired, and it is the RIGHT one
  IF jsonb_array_length(d->'banned_claims') <> jsonb_array_length(prev->'banned_claims') - 1
    THEN RAISE EXCEPTION 'banned_claims moved by % , expected exactly -1',
      jsonb_array_length(d->'banned_claims') - jsonb_array_length(prev->'banned_claims'); END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'banned_claims') b
   WHERE strpos(b->>'pattern', 'three (or|to) four days') > 0;
  IF n <> 0 THEN RAISE EXCEPTION 'the three-or-four-days ban survives - the owner''s new figure is still blocked'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(prev->'banned_claims') b
   WHERE strpos(b->>'pattern', 'three (or|to) four days') > 0;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 three-or-four ban in the superseded row, found % - the target moved', n; END IF;

  -- the retired figure must NOT be armed here (that is the _HOLD file's job, after re-render)
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'banned_claims') b
   WHERE b->>'pattern' ~ 'two or three';
  IF n <> 0 THEN RAISE EXCEPTION 'a ban on "two or three days" was armed here - it must wait for the re-render (bugs_open/161)'; END IF;

  -- the stale ban REASON was corrected, and no reason still quotes the retired figure
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'banned_claims') b
   WHERE b->>'reason' ~ 'attested build time is two or three days';
  IF n <> 0 THEN RAISE EXCEPTION '% ban reason(s) still quote the retired build time', n; END IF;

  -- ── writer_block: EXACTLY the two intended edits, proven by reconstruction
  wb := d->>'writer_block'; pwb := prev->>'writer_block';
  expect := replace(replace(pwb,
    'Say how long it takes: usually ready in two or three days from having what is needed.',
    'Say how long it takes: ready in three or four days, usually sooner, from having what is needed.'),
    'The build duration is HEDGED ("usually ready in two or three days") and is a range, not a number: never render it as a stat, a counter, or a bare figure such as "3 days" or "72 hours".',
    'The build duration is HEDGED ("ready in three or four days, usually sooner") and is a range, not a number: never render it as a stat, a counter, or a bare figure such as "4 days" or "96 hours".');
  IF wb IS DISTINCT FROM expect
    THEN RAISE EXCEPTION 'writer_block is not the old text plus exactly the two named edits'; END IF;

  -- the OUTCOME, asserted independently of how it was reached
  IF wb ~* 'two or three' THEN RAISE EXCEPTION 'the retired range survives somewhere in writer_block'; END IF;
  IF wb !~* 'three or four days' THEN RAISE EXCEPTION 'writer_block does not state the attested turnaround'; END IF;
  IF position(bd->>'writer_line' in wb) = 0
    THEN RAISE EXCEPTION 'writer_block does not contain build_duration''s own writer_line (%)', bd->>'writer_line'; END IF;
END $$;

COMMIT;

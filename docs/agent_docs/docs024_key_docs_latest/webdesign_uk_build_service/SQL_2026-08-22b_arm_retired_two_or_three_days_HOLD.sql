-- SQL_2026-08-22b — arms a promise-shape ban on the RETIRED "two or three days"
-- turnaround, so the class cannot come back after the pages are rebuilt.
--
-- ⚠⚠ _HOLD: DO NOT RUN THIS YET. It is held back deliberately, and the file
-- enforces that rather than asking you to remember it -- but read why first.
--
-- SQL_2026-08-22 (the owner's D-B answer) re-attested build_duration to "three or
-- four days, usually sooner". This file removes the OLD figure's ability to return.
-- It must run THIRD:
--   1. SQL_2026-08-22  -- makes the new figure legal (DONE 2026-08-22)
--   2. re-render index + faq -- removes "two or three days" from stored components
--   3. THIS FILE -- arms the ban, once the census below returns zero
--
-- THE ORDER IS THE WHOLE POINT (SQL_2026-08-19e's rule; the bugs_open/161 landmine).
-- A banned_claims entry is BLOCKER severity. Arm one while offending copy is still
-- STORED and every affected page refuses to save: the retired figure stays published
-- AND becomes unfixable through the normal rewrite path.
--
-- ⚠ THE RE-RENDER SCOPE IS FOUR PAGES, NOT TWO -- and the SERVED pages under-report it.
-- Counting at preview.webdesign.uk gives index 6x and faq 3x (measured 2026-08-22 WITH A
-- CONTROL; the apex 302s to webdesign.co.uk and returns a confident, meaningless zero).
-- But the census below reads STORED components, which is what the claims gate actually
-- blocks on, and it finds **10 components across 4 pages**:
--     faq/{call-to-action,faq,hero}, how-it-works/{call-to-action,generic-text-block,hero},
--     index/{brief-explanation,call-to-action,hero}, tool-website-brief-starter-guide/article-body
-- The same four pages the 2026-08-19 next-day change had to rebuild. Two of them are not
-- reachable by curling the two obvious URLs, so SIZE THE RE-RENDER FROM THIS CENSUS, never
-- from the served pages.
--
-- So the transaction opens with that census, scoped to ACTIVE pages, and REFUSES on
-- non-zero. Repair first, arm second.
--
-- THE PATTERN IS A PROMISE SHAPE, NOT A BARE TOKEN, and that is deliberate. This lane
-- has had bare-token bans block a page for DENYING the banned thing three times over
-- (\brefunds?\b, `whenever you like`, `round of changes`). A bare \btwo or three days\b
-- would do it a fourth time the first time a page writes "it is not two or three days
-- any more, allow three or four" -- which is exactly the sentence the re-render may
-- want. The negation guard scans backwards and covers "not"/"never"/contractions, but
-- NOT a bare "no".
--
-- THE CENSUS REGEX IS DELIBERATELY BROADER than the armed pattern (bare "two or three
-- days" with no promise shape required). Erring toward refusing is the safe direction:
-- if the broad census returns zero, the narrower armed pattern certainly matches
-- nothing. It is written in Postgres ARE, not copied from the Go RE2 pattern, because
-- \b is a word boundary in one and a backspace in the other and a silent translation
-- slip would disarm the census.
--
-- BEFORE RUNNING, also do what 19e did and MEASURE with the real engine, over both
-- halves -- the shapes that must block AND the innocent shapes that must pass:
--   go run ./cmd/claimscan -evidence <candidate_eb.json> -components <tsv>
-- A must-pass-only probe set cannot tell a clean scan from a dead one.

BEGIN;

DO $$
DECLARE offending int; detail text;
BEGIN
  SELECT count(*), string_agg(DISTINCT p.name || '/' || pc.slot_name, ', ')
    INTO offending, detail
    FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
   WHERE s.domain='webdesign.uk' AND p.status <> 'archived'
     AND (pc.rendered_html      ~* '(two or three days|2[-–]3 days|two[- ]to[- ]three)'
       OR pc.content_data::text ~* '(two or three days|2[-–]3 days|two[- ]to[- ]three)');
  IF offending > 0 THEN
    RAISE EXCEPTION 'REFUSING TO ARM: % component(s) still carry the retired turnaround (%). Repair the copy first, or arming makes them unfixable through the rewrite path.', offending, detail;
  END IF;
END $$;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{banned_claims}',
      (c.data->'banned_claims') || jsonb_build_object(
        'pattern', '\b(ready|built|finished|delivered|live|done)\b[^.!?]{0,40}\btwo or three days\b|\btwo or three days\b[^.!?]{0,25}\b(turnaround|delivery|build|service|site|website)\b|\b2[-–]3 days\b|\btwo[-–]to[-–]three days\b',
        'reason', 'RETIRED FIGURE (owner 2026-08-22): the build time was "usually ready in two or three days" from 2026-08-19 until 2026-08-22, and is now "three or four days, usually sooner" (fact build_duration), re-cut so the owner''s pre-delivery review step does not break the promise. Armed to stop the class returning, as the next-day class was. Written as a PROMISE SHAPE, not a bare token: you may deny it in normal English ("it is not two or three days any more, allow three or four"), and the negation guard covers "not"/"never"/contractions but NOT a bare "no". What is banned is stating or implying a two-or-three-day turnaround.'
      )
    ) AS newdata
  FROM cur c
),
retire AS (
  UPDATE site_specs ss SET is_current=false, superseded_at=now()
   WHERE ss.id=(SELECT id FROM cur) RETURNING 1
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'owner-ruling',
 'SQL_2026-08-22b: arms a promise-shape ban on the retired two-or-three-days turnaround, after the last offending component was repaired. Census guard refused on non-zero.',
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

  IF prev->'facts'         IS DISTINCT FROM d->'facts'         THEN RAISE EXCEPTION 'facts changed, they must not'; END IF;
  IF prev->>'writer_block' IS DISTINCT FROM d->>'writer_block' THEN RAISE EXCEPTION 'writer_block changed, it must not'; END IF;

  -- Exactly one ban added, and every previous one carried through untouched.
  IF jsonb_array_length(d->'banned_claims') <> jsonb_array_length(prev->'banned_claims') + 1
    THEN RAISE EXCEPTION 'expected exactly one ban added: % -> %',
      jsonb_array_length(prev->'banned_claims'), jsonb_array_length(d->'banned_claims'); END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(prev->'banned_claims') e
   WHERE NOT (d->'banned_claims' @> jsonb_build_array(e));
  IF n <> 0 THEN RAISE EXCEPTION '% pre-existing ban(s) were altered or dropped', n; END IF;

  -- The new one is present, exactly once, and is the promise shape not a bare token.
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'banned_claims') e
   WHERE position('two or three days' in COALESCE(e->>'pattern','')) > 0;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 two-or-three ban, found %', n; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'banned_claims') e
   WHERE position('two or three days' in COALESCE(e->>'pattern','')) > 0
     AND position('[^.!?]' in COALESCE(e->>'pattern','')) = 0;
  IF n <> 0 THEN RAISE EXCEPTION 'the two-or-three ban is a BARE TOKEN - it will block the denial too'; END IF;

  -- The ban must not contradict the fact it protects.
  SELECT f INTO bd FROM jsonb_array_elements(d->'facts') f WHERE f->>'id'='build_duration';
  IF bd->>'writer_line' !~* 'three or four days'
    THEN RAISE EXCEPTION 'build_duration is not on the three-or-four figure - arming this ban would block the LIVE promise'; END IF;
END $$;

COMMIT;

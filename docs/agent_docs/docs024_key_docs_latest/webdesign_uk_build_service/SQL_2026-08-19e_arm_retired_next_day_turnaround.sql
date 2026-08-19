-- SQL_2026-08-19e — arms a banned_claims entry on the RETIRED next-day
-- turnaround, so the class cannot come back after four pages were rebuilt to
-- remove it.
--
-- ⚠ ORDER IS LOAD-BEARING, AND THIS FILE ENFORCES IT RATHER THAN ASKING YOU TO
-- REMEMBER IT. A banned_claims entry is BLOCKER severity. Arming one while
-- offending copy is still STORED makes every affected page refuse to save: the
-- falsehood stays published AND becomes unfixable through the normal rewrite
-- path. That is the bugs_open/161 landmine. So the transaction opens with a
-- census of still-offending components on ACTIVE pages and refuses on non-zero.
-- Repair first, arm second.
--
-- THE PATTERN IS A PROMISE SHAPE, NOT A BARE TOKEN, and that is deliberate: this
-- lane has now had THREE bare-token bans block a page for DENYING the banned
-- thing (\brefunds?\b, narrowed 08-18; `whenever you like`, owner ruled the ban
-- right; `round of changes`, still the owner's call). A bare \bnext.day\b would
-- do it a fourth time the first time a page wrote "it is not ready the next day,
-- allow two or three".
--
-- MEASURED BEFORE ARMING, with cmd/claimscan (the same engine the deploy gate
-- runs), over ALL 27 current components plus eight probe sentences:
--   * sentence probes: the three offending shapes fire ("usually ready the next
--     day", "next-day turnaround", "ready by tomorrow"); five must-pass
--     sentences are clean, including the current attested copy and an explicit
--     denial (the denial is suppressed by the negation guard, which recognises
--     "not" - reported as 1 suppressed match, so the suppression is observable
--     rather than assumed);
--   * whole corpus, candidate register vs live register: +1 finding, -0. The one
--     new finding is the known offender. Every live page is otherwise clean
--     against the live register; the 20 baseline findings all sit on
--     `index-rejected-v1-20260806`, which is `pages.status='archived'` and is why
--     the census below is scoped to active pages.
--
-- The census regex is deliberately BROADER than the armed pattern (bare
-- `next[- ]day` with no promise shape required). A guard that errs toward
-- refusing is the safe direction: if the broad census returns zero, the narrower
-- armed pattern certainly matches nothing. It is written in Postgres ARE, not
-- copied from the Go RE2 pattern, because \b means word-boundary in one and
-- backspace in the other and a silent translation slip would disarm the census.

BEGIN;

DO $$
DECLARE offending int; detail text;
BEGIN
  SELECT count(*), string_agg(DISTINCT p.name || '/' || pc.slot_name, ', ')
    INTO offending, detail
    FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
   WHERE s.domain='webdesign.uk' AND p.status <> 'archived'
     AND (pc.rendered_html      ~* '(next[- ]day|ready by tomorrow|in (just )?(a|one) day|within (a|one|24) (day|hour))'
       OR pc.content_data::text ~* '(next[- ]day|ready by tomorrow|in (just )?(a|one) day|within (a|one|24) (day|hour))');
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
        'pattern', '\b(ready|built|finished|delivered|live|done)\b[^.!?]{0,40}\bnext[- ]day\b|\bnext[- ]day\b[^.!?]{0,25}\b(turnaround|delivery|build|service|site|website)\b|\bready\s+(by\s+)?tomorrow\b|\bin\s+(just\s+)?(a|one)\s+day\b|\bwithin\s+(a|one|24)\s+(day|hours?)\b',
        'reason', 'RETIRED FIGURE (owner 2026-08-19): the build time was "usually ready the next day" from 2026-08-14 until 2026-08-19, and is now two or three days (fact build_duration). Four pages had to be rebuilt to remove the next-day figure, so this is armed to stop the class returning. Written as a PROMISE SHAPE, not a bare token: you may deny it in normal English ("it is not ready the next day, allow two or three"), and the negation guard covers "not"/"never"/contractions but NOT a bare "no". What is banned is stating or implying next-day, by-tomorrow, one-day or within-24-hours delivery.'
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
 'SQL_2026-08-19e: arms a promise-shape ban on the retired next-day turnaround, after the last offending component was repaired. Census guard refused on non-zero.',
 true,'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; prev jsonb; n int;
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

  -- The new one is present and is the promise shape, not a bare token.
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'banned_claims') e
   WHERE position('next[- ]day' in COALESCE(e->>'pattern','')) > 0;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 next-day ban, found %', n; END IF;
END $$;

COMMIT;

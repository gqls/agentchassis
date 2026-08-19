-- SQL_2026-08-19h — the testimonial-shape ban could never do what its own reason
-- says, because EVERY banned_claims pattern is compiled case-INSENSITIVELY and
-- this one is the only pattern on the estate that depends on capitalisation.
--
-- THE PATTERN:  "[^"]{20,}" ?[—,-]? ?[A-Z][a-z]+ [A-Z]
-- ITS REASON:   "Testimonial shape: a long quotation followed by an attributed
--                name. There are no customers and therefore no testimonials."
--
-- The `[A-Z][a-z]+ [A-Z]` tail is the attributed name: "Sarah T". It is the only
-- thing separating a testimonial from any other quotation. But `claims.go:296`
-- compiles every site pattern as `regexp.Compile("(?i)" + p)` (and
-- `claims_global.go:223` does the same for the fleet-wide set), so `[A-Z]`
-- matches any letter and `[a-z]+` matches any word. The tail degrades to
-- "a word, then a letter" - which is simply "the quotation is followed by more
-- prose".
--
-- SO THE BAN HAS BEEN STOPPING ANY PAGE THAT QUOTES ANYTHING. Measured with
-- cmd/claimscan against the live register, six probes:
--   BEFORE: all three real testimonials blocked, AND all three innocent
--           quotations blocked - including a quoted QUESTION ("Ask yourself
--           \"what should someone be able to do on this site\" before you start")
--           and a quoted anti-example ("\"Modern and professional\" does not tell
--           the tool anything").
--   AFTER:  the three testimonials blocked, the three innocent quotations clean.
-- Over the 27 live components the fix removes nothing and adds nothing, because
-- a page blocked at save never becomes a stored component - which is exactly why
-- this went unnoticed. The damage is an ABSENCE.
--
-- HOW IT SURFACED: the Website Brief Starter guide was stopped at 16:09Z today
-- on *"A joiner in Leeds who wants a one-page site with a contact form and photos
-- of finished jobs" tells t*. A guide about writing briefs quotes example briefs;
-- it could not have avoided this. Third blocker on the same page today, and the
-- only one of the three that was not about commercial terms at all.
--
-- THE FIX is four characters: `(?-i)` before the name part, restoring case
-- sensitivity for that tail only. Go's RE2 supports inline flag groups, so the
-- forced `(?i)` prefix can be locally reversed without touching platform code.
-- Nothing else about the pattern changes.
--
-- ⚠ THE GENERAL TRAP, which is not confined to this site: a banned_claims pattern
-- that leans on capitalisation is silently disarmed, and it fails in the
-- direction that BLOCKS pages rather than the direction that lets a falsehood
-- through, so it looks like a strict gate rather than a broken one. Fleet census
-- 2026-08-19: `b->>'pattern' ~ '\[A-Z\]'` over every current evidence_base
-- returns exactly ONE row, this one, and the fleet-wide Go set has none. Filed in
-- LANDMINES.md.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{banned_claims}', (
      SELECT jsonb_agg(
        CASE WHEN b.elem->>'pattern' = '"[^"]{20,}" ?[—,-]? ?[A-Z][a-z]+ [A-Z]'
          THEN jsonb_set(
                 jsonb_set(b.elem, '{pattern}', to_jsonb('"[^"]{20,}" ?[—,-]? ?(?-i)[A-Z][a-z]+ [A-Z]'::text)),
                 '{reason}', to_jsonb(
                   'Testimonial shape: a long quotation followed by an attributed name. There are no customers and therefore no testimonials. CASE SENSITIVITY RESTORED 2026-08-19: every pattern is compiled with a forced (?i) prefix (claims.go:296), so the [A-Z][a-z]+ [A-Z] tail that identifies an attributed NAME was matching any word followed by any letter - i.e. any quotation followed by more prose. Measured: it blocked a quoted question and a quoted anti-example as readily as a real testimonial, and it stopped the Website Brief Starter guide for quoting an example brief. The (?-i) group restores case sensitivity for the name only.'::text))
          ELSE b.elem END ORDER BY b.ord)
        FROM jsonb_array_elements(c.data->'banned_claims') WITH ORDINALITY AS b(elem, ord)
    )) AS newdata
  FROM cur c
),
retire AS (
  UPDATE site_specs ss SET is_current=false, superseded_at=now()
   WHERE ss.id=(SELECT id FROM cur) RETURNING 1
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'lane-correction',
 'SQL_2026-08-19h: the testimonial-shape ban depended on capitalisation and every pattern is compiled (?i), so it blocked any quotation followed by prose. (?-i) restores case sensitivity for the attributed-name tail.',
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
  IF jsonb_array_length(prev->'banned_claims') <> jsonb_array_length(d->'banned_claims')
    THEN RAISE EXCEPTION 'ban count moved'; END IF;

  SELECT count(*) INTO n
    FROM jsonb_array_elements(prev->'banned_claims') WITH ORDINALITY a(e,o)
    JOIN jsonb_array_elements(d->'banned_claims')    WITH ORDINALITY b(e,o) USING (o)
   WHERE a.e->>'pattern' IS DISTINCT FROM b.e->>'pattern';
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 pattern to change, got %', n; END IF;

  SELECT count(*) INTO n FROM jsonb_array_elements(d->'banned_claims') e
   WHERE position('(?-i)[A-Z][a-z]+ [A-Z]' in COALESCE(e->>'pattern','')) > 0;
  IF n <> 1 THEN RAISE EXCEPTION 'the case-sensitive group did not land (found %)', n; END IF;

  -- No OTHER pattern may now depend on capitalisation without saying so.
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'banned_claims') e
   WHERE COALESCE(e->>'pattern','') ~ '\[A-Z\]' AND position('(?-i)' in e->>'pattern') = 0;
  IF n <> 0 THEN RAISE EXCEPTION '% pattern(s) still depend on case without a (?-i) group', n; END IF;
END $$;

COMMIT;

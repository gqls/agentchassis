-- SQL_2026-08-19f — OWNER RULING 2026-08-19 (evening): narrow the "round of
-- changes" ban to OFFER SHAPES, so a page may deny changes in normal English.
--
-- WHAT WAS WRONG. The ban's first alternative was a bare token,
-- `\brounds? of (revisions?|changes)\b`. It was armed on 2026-08-12 to kill the
-- retired "one round of changes included" offer, and it did that. It also
-- stopped any page that DENIED changes, because the checker matches a phrase and
-- what is banned is an assertion. It blocked the Website Brief Starter guide at
-- 15:25Z today on: "There is no approval stage, and no round of changes once the
-- build is done." That sentence states fact `no_changes_included` exactly.
--
-- WHY THE NEGATION GUARD DID NOT SAVE IT. `negatedClaimMatch` scans BACKWARDS
-- within the clause for a cue, and bare "no" is EXCLUDED by design (claims.go:
-- "There are no exceptions: every claim is verified" would otherwise disarm a
-- real overclaim). "not"/"never"/contractions are cues; "no" is not. So
-- "we do not include a round of changes" was already fine and
-- "no round of changes" was not.
--
-- THE OWNER'S OWN TEST, applied. Three bans in this family have blocked a denial
-- in two days and they have NOT all gone the same way:
--   \brefunds?\b        register attests no_refund          -> over-broad, narrowed 08-18
--   `whenever you like` register attests nothing unbounded  -> ban right, copy wrong (owner)
--   `round of changes`  register attests no_changes_included -> over-broad, narrowed HERE
-- The discriminator is whether the register attests the thing being denied. It
-- does, so the copy must be able to say so.
--
-- MEASURED, with cmd/claimscan (the engine the deploy gate runs), five offer
-- shapes and six denials, live register vs candidate:
--   before -> 5 offers blocked, and 3 of the 6 denials ALSO blocked
--             ("...and no round of changes once the build is done", "No rounds of
--              changes are included", "There is no round of revisions")
--             the other 3 were already suppressed by the negation guard ("do not")
--   after  -> 5 offers blocked, 0 denials blocked
-- Over the whole live corpus (27 components): the narrowing loses NOTHING and
-- gains nothing. The 20 baseline findings all sit on the archived
-- `index-rejected-v1-20260806` page and are raised by other patterns.
--
-- An earlier candidate let "We give you a round of changes after the build."
-- through, because the object pronoun sits between the verb and the quantifier.
-- Caught by the BLOCK set, which is the half of the test that is easy to skip:
-- a narrowing that only checks its denials pass has not checked it still bans.
--
-- ONLY the first alternative changes. The other three arms of this pattern, and
-- every other ban, are untouched.

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
        CASE WHEN position('\brounds? of (revisions?|changes)\b|' in COALESCE(b.elem->>'pattern','')) = 1
          THEN jsonb_set(
                 jsonb_set(b.elem, '{pattern}', to_jsonb(
                   replace(b.elem->>'pattern',
                     '\brounds? of (revisions?|changes)\b',
                     '\b(includes?|including)\b[^.!?]{0,15}\brounds? of (revisions?|changes)\b|\b(one|1|a|two|2|three|3)\s+rounds?\s+of\s+(revisions?|changes)\b[^.!?]{0,25}\b(included|free|yours|at no (extra )?(cost|charge))\b|\b(you|we)\s+(will |can |may |shall |also |always )*(get|give|offer|provide)\s+(you |them |the customer )?(one |1 |a |two |2 )?rounds? of (revisions?|changes)\b'
                   ))),
                 '{reason}', to_jsonb(
                   'RETIRED OFFER (owner ruling 2026-08-12, PLAN §1d): no changes are included at £149. This supersedes BOTH the two-rounds term of 2026-08-09 and the one-set-of-changes term of 2026-08-11. NARROWED 2026-08-19 on the owner ruling: the first alternative was the bare token "rounds? of (revisions|changes)", which also blocked the DENIAL and stopped a page for saying "there is no approval stage, and no round of changes once the build is done" - which is fact no_changes_included stated exactly. The negation guard scans backwards for "not"/"never"/contractions and deliberately excludes a bare "no", so the denial was unprotected. What is banned now is the OFFER: including a round of changes, a round of changes being included or free, or us giving/offering one. Denying changes in any natural wording is allowed and correct.'::text))
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
SELECT r.site_id,'evidence_base', r.newdata,'owner-ruling',
 'SQL_2026-08-19f: owner ruling 2026-08-19 - narrow the changes ban to offer shapes so a page may deny changes in plain English. One pattern edited; no facts, no writer_block, no other ban touched.',
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

  -- EXACTLY ONE pattern may differ.
  SELECT count(*) INTO n
    FROM jsonb_array_elements(prev->'banned_claims') WITH ORDINALITY a(e,o)
    JOIN jsonb_array_elements(d->'banned_claims')    WITH ORDINALITY b(e,o) USING (o)
   WHERE a.e->>'pattern' IS DISTINCT FROM b.e->>'pattern';
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 pattern to change, got %', n; END IF;

  -- The bare token must be gone, and the offer shapes present.
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'banned_claims') e
   WHERE position('\brounds? of (revisions?|changes)\b|' in COALESCE(e->>'pattern','')) = 1;
  IF n <> 0 THEN RAISE EXCEPTION 'the bare-token alternative survives'; END IF;
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'banned_claims') e
   WHERE position('(includes?|including)\b[^.!?]{0,15}\brounds? of' in COALESCE(e->>'pattern','')) > 0;
  IF n <> 1 THEN RAISE EXCEPTION 'the narrowed alternative did not land (found %)', n; END IF;
END $$;

COMMIT;

-- SQL_2026-08-21d — CORRECTING SQL_2026-08-21c, twenty minutes later, caught by
-- asking the bot rather than by reading the register.
--
-- WHAT I GOT WRONG. 21c added the content policy and narrowed `any_site_type` so the
-- two attested facts could not contradict each other. It narrowed the **writer_line**
-- and left the **claim** saying "The service builds any sort of site". Those two
-- fields have DIFFERENT CONSUMERS:
--
--   * writer_line -> the page writer, via writer_block.
--   * claim       -> THE CHAT BOT, verbatim. renderSystemPrompt (box/chat-service/
--                    facts.go) writes "- " + f.Claim for every fact and nothing else.
--                    It never sees writer_line at all.
--
-- So I narrowed the half that steers the pages nobody has rebuilt yet, and left the
-- half that the only live customer-facing surface reads aloud. Asked "I want a site
-- for my adult entertainment business, quite explicit. Can you do that?", the bot
-- answered: **"Yes, we can build a site for that. The system builds any sort of
-- site."** That is the un-narrowed claim being read back, with the brand-new content
-- policy sitting in the same prompt and losing to it.
--
-- WHY IT LOST rather than merely competing: two contradictory claims in one prompt is
-- not a 50/50. The permissive one answers the customer's question directly and the
-- restrictive one reads as being about something else, so the model takes the one
-- that fits the sentence in front of it. A collision between attested facts is not a
-- style problem; it is the register failing at the exact moment it matters.
--
-- THE GENERAL LESSON, and it is the reverse of the one this lane learned this morning:
-- earlier today the trap was that `writer_block` had been left behind while the facts
-- moved. This is the same seam from the other side. **Narrowing a fact means narrowing
-- CLAIM, writer_line and context_terms together, and the check is not "does the
-- register look right" but "what does the bot now say".** Verified at the artefact,
-- never at the row. Logged in WRONG_CALLS.md.
--
-- The claim keeps `any_site_type`'s original point, which was breadth of PURPOSE
-- (personal, community and project sites, not only business ones) and never content.

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
        CASE WHEN f.elem->>'id' = 'any_site_type' THEN
          f.elem || jsonb_build_object(
            'claim', 'The service builds sites for anyone, not just businesses: personal, community and project sites are all welcome. It is not unlimited, and the limits are content limits rather than purpose limits: see content_we_will_not_build. No example sites are named yet: the sites on the owner''s other domains were not produced by this one-shot route, and examples will be added once the owner has produced sites through it.',
            'verified_at', '2026-08-21')
        ELSE f.elem END ORDER BY f.ord)
        FROM jsonb_array_elements(c.data->'facts') WITH ORDINALITY AS f(elem, ord)
    )) AS newdata
  FROM cur c
),
retire AS (
  UPDATE site_specs ss SET is_current=false, superseded_at=now()
   WHERE ss.id=(SELECT id FROM cur) RETURNING 1
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'correction',
 'SQL_2026-08-21d: corrects 21c. any_site_type''s CLAIM still said "any sort of site" and the chat bot reads claim verbatim, so it answered yes to an explicit adult brief minutes after the content policy went live. Claim narrowed to match the writer_line.',
 true,'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; prev jsonb; n int; fact jsonb;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF prev IS NULL THEN RAISE EXCEPTION 'no superseded row to compare against'; END IF;

  IF prev->'banned_claims' IS DISTINCT FROM d->'banned_claims' THEN RAISE EXCEPTION 'banned_claims changed'; END IF;
  IF prev->>'writer_block' IS DISTINCT FROM d->>'writer_block' THEN RAISE EXCEPTION 'writer_block changed, it must not'; END IF;
  IF jsonb_array_length(prev->'facts') <> jsonb_array_length(d->'facts') THEN RAISE EXCEPTION 'fact count moved'; END IF;

  SELECT count(*) INTO n
    FROM jsonb_array_elements(prev->'facts') WITH ORDINALITY a(e,o)
    JOIN jsonb_array_elements(d->'facts')    WITH ORDINALITY b(e,o) USING (o)
   WHERE a.e IS DISTINCT FROM b.e AND b.e->>'id' <> 'any_site_type';
  IF n <> 0 THEN RAISE EXCEPTION 'a fact other than any_site_type changed'; END IF;

  SELECT e INTO fact FROM jsonb_array_elements(d->'facts') e WHERE e->>'id'='any_site_type';
  -- The exact sentence the bot read back must be gone from the CLAIM.
  IF position('builds any sort of site' in fact->>'claim') > 0
    THEN RAISE EXCEPTION 'the claim still says "builds any sort of site", which is what the bot read aloud'; END IF;
  IF position('not just businesses' in fact->>'claim') = 0
    THEN RAISE EXCEPTION 'the claim lost its original point (breadth of purpose)'; END IF;
  IF position('content_we_will_not_build' in fact->>'claim') = 0
    THEN RAISE EXCEPTION 'the claim does not point at the fact that bounds it, so a reader of one need never see the other'; END IF;

  -- BOTH consumer-facing fields must now agree. This is the check whose absence
  -- caused 21c: claim and writer_line have different readers and only one was fixed.
  IF position('any sort of site' in fact->>'claim') > 0
     OR position('Any sort of site' in fact->>'writer_line') > 0
    THEN RAISE EXCEPTION 'claim and writer_line have not been narrowed together'; END IF;

  -- And the policy fact it defers to must actually exist, or the pointer is a lie.
  IF NOT EXISTS (SELECT 1 FROM jsonb_array_elements(d->'facts') e WHERE e->>'id'='content_we_will_not_build')
    THEN RAISE EXCEPTION 'any_site_type points at content_we_will_not_build and that fact is not present'; END IF;
END $$;

COMMIT;

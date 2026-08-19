-- SQL_2026-08-19c — CORRECTS my own writer_block carve-out from SQL_2026-08-19b,
-- three hours old, on the owner's clarification.
--
-- OWNER, 2026-08-19: *"the zip is theirs forever, our hosting is temporally
-- limited. The domain will need to be moved to their registrar."*
--
-- WHAT I GOT WRONG. `SQL_2026-08-19b` bounded the timing of moving the SITE and
-- then carved out the domain as the one genuinely open-ended thing:
--     "a domain the customer BUYS outright for £200 is their property, so where and
--      when they move THAT is their business for ever."
-- I reasoned from ownership: they own it, so nobody can make them move it. That is
-- true about the PROPERTY and false about the ARRANGEMENT — a bought domain still
-- sits in our registrar account until it is transferred out, so it does need moving,
-- and telling a writer it is open-ended invites exactly the unbounded promise the
-- 2026-08-09 caps ban exists to stop. Second time in one day that reasoning from
-- what is TRUE about ownership produced the wrong answer about TIMING.
--
-- THE CORRECT SHAPE, and it is simpler than what I wrote: what is permanent is
-- OWNERSHIP (the ZIP is theirs for ever; a bought domain is theirs for ever). What
-- is temporary is anything WE OPERATE — the hosting and the registrar account. So
-- nothing we run is open-ended, and there is no carve-out at all.
--
-- ⚠ ONE THING NOT CHANGED HERE, DELIBERATELY, because it is a customer promise and
-- the owner's to word: fact `domain_buy_once` says the customer is *"then FREE to
-- transfer it to their own registrar or host"*. "Free to" reads as an option they
-- may decline; the owner says it "will need to be moved", which is an obligation.
-- That is a real difference in what a buyer is being told. Flagged in the handoff
-- for his ruling; the fact is untouched.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{writer_block}', to_jsonb(
      replace(
        c.data->>'writer_block',
        'One thing is genuinely open-ended and may be written that way: a domain the customer BUYS outright for £200 is their property, so where and when they move THAT is their business for ever. Bind the move, never the ownership.',
        'Separate OWNERSHIP from ARRANGEMENT, because only one of them is permanent. Permanent: the ZIP of the finished site is theirs for ever, and a domain bought outright for £200 is theirs for ever. Temporary: everything WE operate. The address we host the site at does not stay up indefinitely, and a bought domain sits in our registrar account until the customer transfers it out, so THAT needs moving to their own registrar as well. Nothing we run is open-ended, so no time phrase about anything we run may be. Write ownership as permanent and timing as bounded, and never let a true statement about what they own become an unbounded promise about how long we will keep running it.'
      )
    )) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'owner-ruling',
  'SQL_2026-08-19c: corrects 19b - the bought domain must move to their own registrar too, so there is no open-ended carve-out. writer_block only.',
  true,'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; prev jsonb; wb text; n int;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF prev IS NULL THEN RAISE EXCEPTION 'no superseded row to compare against'; END IF;
  IF prev->'facts' IS DISTINCT FROM d->'facts' THEN RAISE EXCEPTION 'facts changed - they must not'; END IF;
  IF prev->'banned_claims' IS DISTINCT FROM d->'banned_claims' THEN RAISE EXCEPTION 'banned_claims changed - they must not'; END IF;

  wb := d->>'writer_block';
  IF position('their business for ever' in wb) <> 0
    THEN RAISE EXCEPTION 'the old open-ended carve-out survives'; END IF;
  SELECT count(*) INTO n FROM regexp_matches(wb, 'Separate OWNERSHIP from ARRANGEMENT', 'g');
  IF n <> 1 THEN RAISE EXCEPTION 'the replacement landed % times, expected exactly 1', n; END IF;
  IF position('needs moving to their own registrar as well' in wb) = 0
    THEN RAISE EXCEPTION 'the registrar obligation did not land'; END IF;
  -- 19b''s own instruction, and earlier wires, must survive
  IF position('THE TIMING OF THAT MOVE IS BOUNDED' in wb) = 0 THEN RAISE EXCEPTION '19b bound lost'; END IF;
  IF position('helpful assistant, not a marketing bot' in wb) = 0 THEN RAISE EXCEPTION 'voice brief lost'; END IF;
  IF position('THE WORD PREVIEW IS FOR BEFORE PAYMENT ONLY' in wb) = 0 THEN RAISE EXCEPTION 'preview rule lost'; END IF;
  IF position('Lead by showing the work' in wb) = 0 THEN RAISE EXCEPTION 'lead instruction lost'; END IF;
END $$;

COMMIT;

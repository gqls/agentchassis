-- SQL_2026-08-19b — the customer's move-it-yourself timing is BOUNDED, and the
-- writer must say so instead of reaching for an open-ended phrase.
--
-- OWNER, 2026-08-19: *"Whenever you like should be within the next month."*
--
-- WHAT THIS FIXES, and why it is a writer_block edit and NOT a ban change.
-- `881c95ef` (the Website Brief Starter guide) failed TWICE, the second time on
-- banned claim "whenever you like" at:
--     "...it's yours to move to any registrar or host you like, whenever you like."
-- I had it queued up as the refund-ban defect in a new place -- a bare-phrase ban
-- catching an attested freedom -- and was going to propose narrowing the 2026-08-09
-- caps ban. **The owner's ruling resolves it the other way: the ban is RIGHT and the
-- copy was wrong.** The hosting we provide is not indefinite, so an unbounded
-- "whenever you like" is exactly the open-ended time promise that ban exists to
-- stop. Nothing about the ban changes.
--
-- The writer was not inventing, though, which is why an instruction is the remedy
-- rather than a re-triage: writer_block already says the ZIP is theirs "to keep and
-- host wherever they like" (LOCATION, unbounded and true) and the writer added a
-- TIME clause on top. The block never told it the timing has a bound. Now it does.
--
-- CAREFUL DISTINCTION, kept explicit because getting it wrong would make a false
-- claim in the other direction: a domain the customer BUYS outright for £200 is
-- their property and is genuinely theirs to transfer whenever, for ever. What is
-- bounded is the window in which they must move the SITE off our hosting. The
-- instruction below is worded to bind the move, not the ownership.
--
-- Facts, bans and every other writer_block paragraph unchanged: this appends one
-- sentence to one existing paragraph, by verbatim anchor, exactly once.

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
        'The ZIP means they can equally host it anywhere themselves.',
        'The ZIP means they can equally host it anywhere themselves. THE TIMING OF THAT MOVE IS BOUNDED AND MUST BE STATED AS SUCH (owner ruling, 2026-08-19): the customer should move the site to their own hosting WITHIN THE NEXT MONTH, because the address we provide does not stay up indefinitely. Never write an open-ended time phrase about anything we host or provide - not "whenever you like", not "at any time", not "no rush" - because that is a promise this offer does not make and the claims gate will stop the page. Say "within the next month" and let the reader plan. One thing is genuinely open-ended and may be written that way: a domain the customer BUYS outright for £200 is their property, so where and when they move THAT is their business for ever. Bind the move, never the ownership.'
      )
    )) AS newdata
  FROM cur c
),
retire AS (UPDATE site_specs ss SET is_current=false, superseded_at=now() WHERE ss.id=(SELECT id FROM cur) RETURNING 1)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'owner-ruling',
  'SQL_2026-08-19b: the move-it-yourself window is bounded to within the month (owner 2026-08-19). writer_block only; facts and bans unchanged.',
  true,'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; prev jsonb; n int; wb text;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF prev IS NULL THEN RAISE EXCEPTION 'no superseded row to compare against'; END IF;

  -- facts and bans must be untouched by THIS write
  SELECT count(*) INTO n FROM (
    (SELECT f->>'id' FROM jsonb_array_elements(prev->'facts') f
     EXCEPT SELECT f->>'id' FROM jsonb_array_elements(d->'facts') f)
    UNION ALL
    (SELECT f->>'id' FROM jsonb_array_elements(d->'facts') f
     EXCEPT SELECT f->>'id' FROM jsonb_array_elements(prev->'facts') f)) x;
  IF n <> 0 THEN RAISE EXCEPTION '% fact id(s) differ - this write touches writer_block only', n; END IF;
  IF prev->'facts' IS DISTINCT FROM d->'facts' THEN RAISE EXCEPTION 'facts changed - they must not'; END IF;
  IF prev->'banned_claims' IS DISTINCT FROM d->'banned_claims' THEN RAISE EXCEPTION 'banned_claims changed - they must not'; END IF;

  wb := d->>'writer_block';
  -- landed exactly once
  SELECT count(*) INTO n FROM regexp_matches(wb, 'THE TIMING OF THAT MOVE IS BOUNDED', 'g');
  IF n <> 1 THEN RAISE EXCEPTION 'the new instruction landed % times, expected exactly 1', n; END IF;
  IF position('Bind the move, never the ownership.' in wb) = 0
    THEN RAISE EXCEPTION 'the ownership carve-out did not land'; END IF;
  -- earlier wires must survive
  IF position('pays before the site is built' in wb) = 0 THEN RAISE EXCEPTION 'payment sentence lost'; END IF;
  IF position('helpful assistant, not a marketing bot' in wb) = 0 THEN RAISE EXCEPTION 'voice brief lost'; END IF;
  IF position('THE WORD PREVIEW IS FOR BEFORE PAYMENT ONLY' in wb) = 0 THEN RAISE EXCEPTION 'preview rule lost'; END IF;
  IF position('Lead by showing the work' in wb) = 0 THEN RAISE EXCEPTION 'lead instruction lost'; END IF;
  IF length(wb) <= length(prev->>'writer_block') THEN RAISE EXCEPTION 'writer_block did not grow - the anchor did not match'; END IF;
END $$;

COMMIT;

-- SQL_2026-08-21b — the writer_block had been left behind by TWO owner rulings
-- it was supposed to carry, and one of the leftovers instructs copy that the
-- deploy gate now REFUSES. Facts and bans are untouched here; this is the
-- steering text catching up with attestations already made.
--
-- ── HOW IT HAPPENED, because the mechanism is worth more than the four fixes ──
--
-- A fact edit on this register is written with a guard asserting
-- `writer_block` UNCHANGED. That guard is CORRECT: it is what makes a one-fact
-- change reviewable, and SQL_2026-08-19 (build_duration) and SQL_2026-08-19g
-- (domain_buy_once) both carry it. But writer_block is the WIRE, not
-- bookkeeping — this lane's own NOTES say so at 2026-08-10: *"a fact not copied
-- into writer_block does not exist for the writer"*. So the same line that
-- proves a fact edit was careful is the line that leaves the writer steered by
-- the fact's RETIRED value. The guard reads as rigour and its effect is drift.
-- Filed in LANDMINES.md.
--
-- ── WHAT WAS STALE, all four MEASURED at the live row on 2026-08-21 ──
--
-- 1. TURNAROUND (owner ruling 2026-08-19, applied to facts as SQL_2026-08-19).
--    `build_duration` has said "usually ready in two or three days" (value 3)
--    since 2026-08-19. The writer_block still said *"Say how long it takes:
--    usually ready the next day"* and, in the same breath, *"never a range of
--    days"* — which forbids the attested phrasing outright. The writer was
--    therefore told the opposite of the fact, twice, in the one text that
--    reaches it.
--
--    THIS IS NOT A TIDY-UP. The next-day shape was ARMED as a ban on 2026-08-20
--    (SQL_2026-08-19e). Fed the live writer_block's own instruction sentence,
--    cmd/claimscan — the same engine as the deploy gate — returns:
--      BANNED  "ready the next day"  …Say how long it takes: usually ready the
--              next day from having what is needed.…
--    So the steering text instructs a sentence that stops the page it steers.
--    The pages are clean today only because they were rebuilt by hand on
--    2026-08-19; the next rebuild of any page that states the turnaround would
--    have walked straight into it. That is the fourth attempt the guide rebuild
--    needed, waiting to happen again.
--
-- 2. The stat rule quoted the retired figure too: *The build duration is HEDGED
--    ("usually ready the next day")*. Re-quoted, and the reason restated as a
--    RANGE. The prohibition itself is deliberately NOT relaxed: the register
--    would permit a bare "3 days" in a stat (value 3, the safe end), and
--    writer_block choosing to be stricter than the gate is an editorial call
--    that was already made. Only the retired quotation moves.
--
-- 3./4. DOMAIN TRANSFER (owner ruling 2026-08-19, applied to facts as
--    SQL_2026-08-19g). That ruling replaced an OPTION with an OBLIGATION and
--    19g's own header records why in detail: *"free to" is an OPTION. The
--    owner's position is an OBLIGATION*, and the old wording had *generated
--    banned copy* — the 15:56Z blocker was the writer elaborating it into
--    "whenever you like". The fact was fixed. The writer_block was not, and
--    still carried the option in both places: *"they are free to transfer the
--    domain to their own registrar or host"* and, in the may-state list,
--    *"and then transferred freely"*. Two days on, the sentence that has already
--    blocked a page once was still the instruction.
--
-- ── ONE THING DELIBERATELY WORDED THE LONG WAY ──
--
-- The new turnaround rule names the banned shape WITHOUT containing a live
-- instance of it. The first draft read *"Never promise it for the next day, for
-- tomorrow, in one day or within 24 hours"* and claimscan BANNED it on "in one
-- day": the negation guard scans backwards a short way, so "Never" reached the
-- first item in the list and not the third. A page writer echoing that
-- instruction would have been refused. Prompt text is read as an example
-- (box/chat-service/facts_test.go makes the same point about the em dash rule),
-- so the rule now describes the shape in prose the gate has no opinion about:
-- "on the following day, by tomorrow, or inside twenty-four hours" — 0 findings.
-- The writer does not need the literal; when the gate refuses, it prints the
-- ban's own reason, which carries it.
--
-- MEASURED: every replacement string scanned through cmd/claimscan against the
-- live register — 0 findings, in a run whose three controls all fired.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{writer_block}', to_jsonb(
      replace(replace(replace(replace(c.data->>'writer_block',
        -- 1. the turnaround rule
        'Say how long it takes: usually ready the next day from having what is needed. Give it plainly, as a normal fact about how the work goes. Never say three or four days, and never a range of days: that was the old offer''s figure and it is retired.',
        'Say how long it takes: usually ready in two or three days from having what is needed. Give it plainly, as a normal fact about how the work goes, and keep both ends of the range and the hedge. Two older figures are retired and neither may come back. The single-day turnaround this offer promised until 2026-08-19 is now banned outright, so a sentence promising delivery on the following day, by tomorrow, or inside twenty-four hours is refused at the gate rather than merely reading oddly. The three-or-four-day figure from the offer before that is retired as well.'),
        -- 2. the stat rule's stale quotation
        'The build duration is HEDGED ("usually ready the next day") and is not a number: never render it as a stat, a counter, or a bare figure such as "1 day" or "24 hours".',
        'The build duration is HEDGED ("usually ready in two or three days") and is a range, not a number: never render it as a stat, a counter, or a bare figure such as "3 days" or "72 hours".'),
        -- 3. the domain paragraph's option wording
        'and after that they are free to transfer the domain to their own registrar or host.',
        'and after that the domain is theirs and must be moved to their own registrar, because we do not stay their registrar.'),
        -- 4. the same option wording in the may-state list
        'for a one-off £200 and then transferred freely,',
        'for a one-off £200 and must then be moved to the customer''s own registrar,')
    )) AS newdata
  FROM cur c
),
retire AS (
  UPDATE site_specs ss SET is_current=false, superseded_at=now()
   WHERE ss.id=(SELECT id FROM cur) RETURNING 1
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'correction',
 'SQL_2026-08-21b: writer_block catches up with the owner rulings of 2026-08-19 (turnaround two or three days; the £200 domain must be moved, it is not an option). Facts and bans deliberately unchanged. The stale turnaround instruction was measured to be BANNED copy by cmd/claimscan.',
 true,'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; prev jsonb; wb text; pwb text; expect text; fact jsonb;
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF prev IS NULL THEN RAISE EXCEPTION 'no superseded row to compare against'; END IF;

  wb := d->>'writer_block'; pwb := prev->>'writer_block';

  -- This file changes the WIRE ONLY. Facts and bans must be byte-identical.
  IF prev->'facts' IS DISTINCT FROM d->'facts' THEN RAISE EXCEPTION 'facts changed, they must not'; END IF;
  IF prev->'banned_claims' IS DISTINCT FROM d->'banned_claims' THEN RAISE EXCEPTION 'banned_claims changed, they must not'; END IF;

  -- EXACTLY the four intended edits, proven by reconstruction.
  expect := replace(replace(replace(replace(pwb,
    'Say how long it takes: usually ready the next day from having what is needed. Give it plainly, as a normal fact about how the work goes. Never say three or four days, and never a range of days: that was the old offer''s figure and it is retired.',
    'Say how long it takes: usually ready in two or three days from having what is needed. Give it plainly, as a normal fact about how the work goes, and keep both ends of the range and the hedge. Two older figures are retired and neither may come back. The single-day turnaround this offer promised until 2026-08-19 is now banned outright, so a sentence promising delivery on the following day, by tomorrow, or inside twenty-four hours is refused at the gate rather than merely reading oddly. The three-or-four-day figure from the offer before that is retired as well.'),
    'The build duration is HEDGED ("usually ready the next day") and is not a number: never render it as a stat, a counter, or a bare figure such as "1 day" or "24 hours".',
    'The build duration is HEDGED ("usually ready in two or three days") and is a range, not a number: never render it as a stat, a counter, or a bare figure such as "3 days" or "72 hours".'),
    'and after that they are free to transfer the domain to their own registrar or host.',
    'and after that the domain is theirs and must be moved to their own registrar, because we do not stay their registrar.'),
    'for a one-off £200 and then transferred freely,',
    'for a one-off £200 and must then be moved to the customer''s own registrar,');
  IF expect IS DISTINCT FROM wb
    THEN RAISE EXCEPTION 'writer_block is not the old text plus exactly the four named edits'; END IF;

  -- The OUTCOME, asserted independently of how it was reached. A reconstruction
  -- guard proves no unintended edit; it cannot prove the intended ones mattered.
  IF position('next day' in wb) > 0
    THEN RAISE EXCEPTION 'the retired next-day figure survives somewhere in writer_block'; END IF;
  IF position('free to transfer' in wb) > 0 OR position('transferred freely' in wb) > 0
    THEN RAISE EXCEPTION 'the option wording survives in writer_block'; END IF;
  IF position('never a range of days' in wb) > 0
    THEN RAISE EXCEPTION 'writer_block still forbids the attested range'; END IF;
  IF position('two or three days' in wb) = 0
    THEN RAISE EXCEPTION 'writer_block does not state the attested turnaround'; END IF;
  IF position('must be moved to their own registrar' in wb) = 0
    THEN RAISE EXCEPTION 'writer_block does not state the transfer obligation'; END IF;

  -- And it must agree with the fact it is meant to carry, not merely differ
  -- from the old text: this is the check that would have caught the drift.
  SELECT e INTO fact FROM jsonb_array_elements(d->'facts') e WHERE e->>'id'='build_duration';
  IF position(fact->>'writer_line' in wb) = 0
    THEN RAISE EXCEPTION 'writer_block does not contain build_duration''s own writer_line (%)', fact->>'writer_line'; END IF;
END $$;

COMMIT;

-- SQL_2026-08-19d — corrects SEVEN stale `reason` strings on webdesign.uk's
-- banned_claims. PATTERNS UNCHANGED, FACTS UNCHANGED, WRITER_BLOCK UNCHANGED.
--
-- WHAT IS WRONG. A banned_claims entry carries a `pattern` (what is stopped) and
-- a `reason` (why). The patterns are owner rulings and are not touched here. The
-- reasons were each written on the day their ban was armed, and five of the seven
-- below go further than explaining a retirement: they state the CURRENT position
-- of the business, and the position has since moved. Two more cite `facts` ids
-- that no longer exist. Measured 2026-08-19 against the live row:
--
--   pattern                                  the reason says            actually
--   ---------------------------------------- -------------------------- ------------------
--   (100%|fully) (guaranteed|satisfaction)   "a refund IS available      no refund at all
--                                             until the customer         (fact no_refund,
--                                             accepts the site"          owner 2026-08-11)
--   \bdeposits?\b                            payment comes AFTER         payment is up front
--                                             approval                   (fact payment_upfront)
--   \byou only pay if you ...\b              cites fact                  fact absent; the
--                                             payment_after_approval     switch it warned
--                                                                        about has flipped
--   (unlimited|no limit to|...)              cites fact                  fact absent; NO
--                                             revision_rounds_included   changes are included
--   (at any point|whenever you like|...)     "the review window is       window retired
--                                             the bound"                 2026-08-11
--   (instant|...|same day)                   "about three or four days"  two or three days
--   \bthree (or|to) four days\b              "usually ready the          two or three days
--                                             next day"                  (owner 2026-08-19)
--
-- WHO ACTUALLY READS THESE, because it changes how much this matters and the
-- lane's NOTES had it wrong. Checked 2026-08-19 at the live agent rows:
--   SELECT type FROM agent_definitions WHERE is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
--     AND default_config::text ILIKE '%banned_claims%';   -> fix-proposer, council-gate
-- and `page-content-writer` matches neither '%banned%' nor '%banned_claims%'; its
-- template pulls `.site_specs.specs.evidence_base.writer_block` and nothing else.
-- So the SITE WRITER NEVER SEES A BAN REASON. NOTES_webdesign_uk_build_service.md
-- said of two of these "the patterns are correct; only the prose the writer reads
-- is out of date" - the prose is out of date, but the writer does not read it.
--
-- It still matters, for two readers that do:
--   1. the blocker message. `checkBannedClaims` copies `reason` verbatim into the
--      ValidationIssue description, which lands in agent_error_log and is the
--      whole of what a session sees when triaging a stopped page. A blocker that
--      explains itself with "a refund is available until the customer accepts the
--      site" hands the next session the retired commercial model as fact.
--   2. the register is its own audit trail, and these are the rows that record
--      WHY each retired term is banned.
--
-- Each replacement states the current position, cites the fact that attests it,
-- and keeps a dated note of what the reason used to say, so the correction is
-- visible rather than silently applied.

BEGIN;

WITH cur AS (
  SELECT ss.id, ss.site_id, ss.data, ss.pinned
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current
),
-- Match on a distinctive substring of the CURRENT reason, never on the pattern:
-- the patterns carry backslashes and pipes, and an escaping slip there would
-- silently match nothing. Never on the array index either: another lane writes
-- this row and an index is not a stable key.
repl(marker, newreason) AS (VALUES
 ('a refund is available until the customer accepts the site',
  'Overclaims the guarantee. There is no guarantee to overclaim: no refund is offered (fact no_refund) and no changes are included (fact no_changes_included). State the actual terms plainly and never a satisfaction slogan. REASON CORRECTED 2026-08-19: it previously read "the real term is narrower and specific: a refund is available until the customer accepts the site", which described the retired £1,200 preview-then-refund model and has been false since the owner ruling of 2026-08-11.'),
 ('The attested figure is about three or four days',
  'Speed class. The attested build time is two or three days (fact build_duration, re-attested by the owner 2026-08-19); anything faster is unsupported and also fights the positioning. REASON CORRECTED 2026-08-19: it previously quoted "about three or four days", a figure retired on 2026-08-14.'),
 ('see facts revision_rounds_included',
  'Caps ruling, owner 2026-08-09, and the offer has since moved further from an uncapped promise, not closer: NO changes are included at £149 (fact no_changes_included, owner ruling 2026-08-12). REASON CORRECTED 2026-08-19: it previously cited fact revision_rounds_included, which does not exist and did not survive the 2026-08-12 ruling. The pattern is unchanged and is still right.'),
 ('no open-ended time promises; the review window is the bound',
  'Caps ruling, owner 2026-08-09, RE-AFFIRMED by the owner on 2026-08-19 after this lane was about to propose narrowing it as an over-block: no open-ended time promise about anything WE operate. What we operate is temporary. The hosting is not indefinite, and a bought domain sits in our registrar account until the customer transfers it out, so the stated bound for moving off us is "within the next month". Ownership may be written as permanent; timing may not. REASON CORRECTED 2026-08-19: it previously named the fourteen-day review window as the bound, and that window was retired on 2026-08-11.'),
 ('Payment is taken once, after the customer approves the site',
  'RETIRED OFFER (owner ruling 2026-08-11): there is no deposit. Payment is taken once, in full, BEFORE the build starts (fact payment_upfront). REASON CORRECTED 2026-08-19: it previously said payment is taken "after the customer approves the site", which was true of the retired model and was reversed by the owner directive of 2026-08-18.'),
 ('stated by fact payment_after_approval',
  'RETIRED OFFER (owner ruling 2026-08-11): these promises belonged to the £1,200 preview-then-refund model. Payment is now taken up front (fact payment_upfront), so they are not merely unsupported, they are the opposite of what happens. REASON CORRECTED 2026-08-19: it previously cited fact payment_after_approval, which no longer exists. The switch that reason warned might flip has flipped.'),
 ('the build time is ''usually ready the next day''',
  'RETIRED FIGURE (owner 2026-08-14): three-or-four-days belonged to the £1,200 offer. The current attested build time is two or three days (fact build_duration, re-attested by the owner 2026-08-19). REASON CORRECTED 2026-08-19: it previously named "usually ready the next day", which was itself retired on 2026-08-19, so the reason for one dead figure was pointing at another.')
),
rebuilt AS (
  SELECT c.site_id, c.pinned,
    jsonb_set(c.data, '{banned_claims}', (
      SELECT jsonb_agg(
               CASE WHEN r.newreason IS NULL THEN b.elem
                    ELSE jsonb_set(b.elem, '{reason}', to_jsonb(r.newreason))
               END
               ORDER BY b.ord)
        FROM jsonb_array_elements(c.data->'banned_claims') WITH ORDINALITY AS b(elem, ord)
        LEFT JOIN repl r ON position(r.marker in COALESCE(b.elem->>'reason','')) > 0
    )) AS newdata
  FROM cur c
),
retire AS (
  UPDATE site_specs ss SET is_current=false, superseded_at=now()
   WHERE ss.id=(SELECT id FROM cur) RETURNING 1
)
INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by, pinned)
SELECT r.site_id,'evidence_base', r.newdata,'lane-correction',
 'SQL_2026-08-19d: corrects seven stale banned_claims REASONS that stated retired commercial terms as current, or cited facts that no longer exist. Patterns, facts and writer_block untouched.',
 true,'webdesign_uk_build_service lane', r.pinned
FROM rebuilt r, retire;

DO $$
DECLARE d jsonb; prev jsonb; n int; pat_before text[]; pat_after text[];
BEGIN
  SELECT ss.data INTO d FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND ss.is_current;
  SELECT ss.data INTO prev FROM site_specs ss JOIN sites s ON s.id=ss.site_id
   WHERE s.domain='webdesign.uk' AND ss.aspect='evidence_base' AND NOT ss.is_current
   ORDER BY ss.superseded_at DESC NULLS LAST LIMIT 1;
  IF prev IS NULL THEN RAISE EXCEPTION 'no superseded row to compare against'; END IF;

  -- Nothing but reasons may move.
  IF prev->'facts'        IS DISTINCT FROM d->'facts'        THEN RAISE EXCEPTION 'facts changed, they must not'; END IF;
  IF prev->>'writer_block' IS DISTINCT FROM d->>'writer_block' THEN RAISE EXCEPTION 'writer_block changed, it must not'; END IF;
  IF jsonb_array_length(prev->'banned_claims') <> jsonb_array_length(d->'banned_claims')
    THEN RAISE EXCEPTION 'ban COUNT moved: % -> %',
      jsonb_array_length(prev->'banned_claims'), jsonb_array_length(d->'banned_claims'); END IF;

  -- Patterns must be identical, in the same order.
  SELECT array_agg(e->>'pattern' ORDER BY o) INTO pat_before
    FROM jsonb_array_elements(prev->'banned_claims') WITH ORDINALITY t(e,o);
  SELECT array_agg(e->>'pattern' ORDER BY o) INTO pat_after
    FROM jsonb_array_elements(d->'banned_claims') WITH ORDINALITY t(e,o);
  IF pat_before IS DISTINCT FROM pat_after THEN RAISE EXCEPTION 'a banned_claims PATTERN changed, none may'; END IF;

  -- Exactly seven reasons must differ. Not "at least": a wider match would mean
  -- a marker matched a row it was not written for.
  SELECT count(*) INTO n
    FROM jsonb_array_elements(prev->'banned_claims') WITH ORDINALITY a(e,o)
    JOIN jsonb_array_elements(d->'banned_claims')    WITH ORDINALITY b(e,o) USING (o)
   WHERE a.e->>'reason' IS DISTINCT FROM b.e->>'reason';
  IF n <> 7 THEN RAISE EXCEPTION 'expected exactly 7 reasons to change, got %', n; END IF;

  -- Every corrected row must SAY it was corrected, and there must be exactly seven.
  SELECT count(*) INTO n FROM jsonb_array_elements(d->'banned_claims') e
   WHERE position('REASON CORRECTED 2026-08-19' in COALESCE(e->>'reason','')) > 0;
  IF n <> 7 THEN RAISE EXCEPTION 'expected exactly 7 reasons to carry the correction note, got %', n; END IF;

  -- And no stale string may survive OUTSIDE a corrected reason. Note the first
  -- version of this check asserted the stale strings were gone altogether, and it
  -- fired on the probe run: each new reason deliberately QUOTES the wording it
  -- replaces, so a visible correction and a bare absence test are incompatible.
  -- Kept as evidence the assertion is live: it has already failed once on real data.
  FOR n IN SELECT 1 FROM jsonb_array_elements(d->'banned_claims') e
    WHERE position('REASON CORRECTED 2026-08-19' in COALESCE(e->>'reason','')) = 0
      AND (position('a refund is available until the customer accepts the site' in COALESCE(e->>'reason','')) > 0
       OR position('The attested figure is about three or four days' in COALESCE(e->>'reason','')) > 0
       OR position('see facts revision_rounds_included' in COALESCE(e->>'reason','')) > 0
       OR position('the review window is the bound' in COALESCE(e->>'reason','')) > 0
       OR position('Payment is taken once, after the customer approves the site' in COALESCE(e->>'reason','')) > 0
       OR position('stated by fact payment_after_approval' in COALESCE(e->>'reason','')) > 0
       OR position('the build time is ''usually ready the next day''' in COALESCE(e->>'reason','')) > 0)
  LOOP
    RAISE EXCEPTION 'a stale reason string survives on an UNcorrected row';
  END LOOP;
END $$;

COMMIT;

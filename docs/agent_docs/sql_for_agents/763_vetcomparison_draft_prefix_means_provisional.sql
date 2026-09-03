-- 763_vetcomparison_draft_prefix_means_provisional.sql
--
-- Answers the council's MEDIUM objection on migration 759 (corr b6cbdcd3, APPROVED with two
-- advisory objections; editquality). The objection is CORRECT and this is the forward fix.
--
-- ─── THE OBJECTION, AND WHY IT IS RIGHT ──────────────────────────────────────────────────────
--
-- 759's header says "EVERY CMA FACT IS DRAFT AND CARRIES `draft_status`", and its verify block
-- asserts `IF ndraft <> 5`. But SIX fact ids match `CMA-DRAFT-%`. Measured after apply:
--
--   CMA-DRAFT-ORDER-CONSULTATION-2026-07-21          draft_status: NO
--   CMA-DRAFT-ADDITIONAL-PRESCRIPTION-FEE-CAP-12-50  draft_status: yes
--   CMA-DRAFT-LARGE-BUSINESS-THRESHOLD-15            draft_status: yes
--   CMA-DRAFT-PRICE-LIST-CATEGORIES-5                draft_status: yes
--   CMA-DRAFT-PRICE-LIST-SERVICES-36                 draft_status: yes
--   CMA-DRAFT-PRIMARY-PRESCRIPTION-FEE-CAP-21        draft_status: yes
--
-- The verify passed only because it required BOTH `LIKE 'CMA-DRAFT-%'` AND `? 'draft_status'`,
-- so it counted 5 and never noticed the sixth. The SUBSTANCE was right and stays right: the
-- consultation fact records a SETTLED historical event — the draft Order really was published
-- for consultation on 21 July 2026, and that does not become untrue when the Order is made — so
-- it correctly carries no `draft_status`. What is wrong is the NAME and the header sentence:
-- a reader who greps `CMA-DRAFT-%` gets 6 against a guard asserting 5 and reasonably concludes
-- one of them is broken.
--
-- **This is the class where the prefix is doing semantic work it was never defined to do.** The
-- fix is to make the prefix mean exactly one thing: `CMA-DRAFT-` = this figure is PROVISIONAL and
-- changes when the Order is made. Everything else about the draft Order that is already settled
-- gets a different prefix.
--
-- After this migration, `CMA-DRAFT-%` and `? 'draft_status'` select the SAME five facts, and the
-- verify below asserts that equivalence directly rather than asserting a bare count — a count
-- cannot notice a sixth member, which is exactly how this got through.
--
-- ⚠ 759's file header remains as written; it is the record of what ran. Forward-only. This file
-- is where the correction lives, and the lane's NOTES carry it too.
--
-- Rollback: 763_..._ROLLBACK.sql

BEGIN;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM sites WHERE id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND domain = 'vetcomparison.uk') THEN
    RAISE EXCEPTION '763 ABORT: site_id does not resolve to vetcomparison.uk';
  END IF;
END $$;

DO $$
DECLARE nold int; nnew int;
BEGIN
  SELECT count(*) FILTER (WHERE f->>'id' = 'CMA-DRAFT-ORDER-CONSULTATION-2026-07-21'),
         count(*) FILTER (WHERE f->>'id' = 'CMA-ORDER-CONSULTATION-PUBLISHED-2026-07-21')
    INTO nold, nnew
    FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current;
  IF nnew > 0 THEN
    RAISE EXCEPTION '763 ABORT: the renamed id already exists - this migration has already run';
  END IF;
  IF nold <> 1 THEN
    RAISE EXCEPTION '763 ABORT: expected exactly 1 fact with the old id, found % - the register has moved on, look before renaming', nold;
  END IF;
END $$;

-- Rebuild the facts array with the one id renamed. jsonb_set cannot address an array element by
-- value, so the array is reconstructed in order; ORDER BY ordinality keeps it stable.
UPDATE site_specs ss
   SET data = jsonb_set(ss.data, '{facts}', (
         SELECT jsonb_agg(
                  CASE WHEN f->>'id' = 'CMA-DRAFT-ORDER-CONSULTATION-2026-07-21'
                       THEN jsonb_set(f, '{id}', '"CMA-ORDER-CONSULTATION-PUBLISHED-2026-07-21"'::jsonb)
                       ELSE f END
                  ORDER BY ord)
           FROM jsonb_array_elements(ss.data->'facts') WITH ORDINALITY AS t(f, ord)))
 WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d'
   AND ss.aspect = 'evidence_base'
   AND ss.is_current;

DO $$
DECLARE nfacts int; nban int; ndraftid int; ndrafttag int; nboth int; posture text;
BEGIN
  SELECT jsonb_array_length(data->'facts'), jsonb_array_length(data->'banned_claims'),
         data->'posture'->>'rung'
    INTO nfacts, nban, posture
    FROM site_specs
   WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;

  -- IS DISTINCT FROM, not <>: on the destructive path these go NULL and `NULL <> 21` is NULL,
  -- not TRUE, so a two-way comparison passes the very disaster it exists to catch. Learned the
  -- hard way on migration 761, by mutation test.
  IF nfacts IS DISTINCT FROM 21 OR nban IS DISTINCT FROM 6 THEN
    RAISE EXCEPTION '763 VERIFY: the register lost content - expected 21 facts and 6 banned_claims, found % and %',
      coalesce(nfacts::text,'NULL'), coalesce(nban::text,'NULL');
  END IF;
  IF posture IS DISTINCT FROM 'relied_upon' THEN
    RAISE EXCEPTION '763 VERIFY: the posture record was lost - expected relied_upon, found %', coalesce(posture,'NULL');
  END IF;

  -- THE POINT OF THIS MIGRATION: the prefix and the tag must now select the SAME set.
  -- Asserted as an EQUIVALENCE, not as a count - a count is what failed to notice the sixth.
  SELECT count(*) FILTER (WHERE f->>'id' LIKE 'CMA-DRAFT-%'),
         count(*) FILTER (WHERE f ? 'draft_status'),
         count(*) FILTER (WHERE (f->>'id' LIKE 'CMA-DRAFT-%') = (f ? 'draft_status'))
    INTO ndraftid, ndrafttag, nboth
    FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
   WHERE ss.site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND ss.aspect = 'evidence_base' AND ss.is_current;

  IF nboth IS DISTINCT FROM nfacts THEN
    RAISE EXCEPTION '763 VERIFY: CMA-DRAFT-%% and draft_status disagree on % of % facts - the prefix must mean exactly "provisional"', nfacts - nboth, nfacts;
  END IF;
  IF ndraftid IS DISTINCT FROM 5 OR ndrafttag IS DISTINCT FROM 5 THEN
    RAISE EXCEPTION '763 VERIFY: expected 5 provisional facts by BOTH prefix and tag, found % and %', ndraftid, ndrafttag;
  END IF;

  RAISE NOTICE '763 OK: CMA-DRAFT-%% now means exactly "provisional" - % ids and % draft_status tags select the same set; register intact at % facts / % banned_claims, posture=%',
    ndraftid, ndrafttag, nfacts, nban, posture;
END $$;

COMMIT;

-- ============================================================================
-- agritec.uk — remove four DOMESTIC energy facts from the register, keep one
-- Written 2026-08-22. Applied out of band (psql -f), per-site setup.
--
-- WHY. Phase 2 run 2 asked for "Ofgem average UK non-domestic electricity price
-- pence per kWh, and UK grid carbon intensity". The run COMPLETED and registered
-- five facts. Four of them are unusable on this site, and one of those is
-- actively misleading. This is bugs_open/161 in practice rather than in
-- principle: verify_and_register proved every quote is really on its page, and
-- that is ALL it proved. It is a PROVENANCE check. It is not a relevance check
-- and it is not a reading check. A registered fact is simultaneously the
-- whitelist the writer is instructed from and the authority the gate gives it,
-- so leaving these in place would have caused the claim and then vouched for it.
--
-- WHAT IS WRONG WITH EACH
--
-- CIT-f8aa28d909b16c00 — "26.11 pence per kWh". Its own quote is a TABLE SCRAPE:
--     "Electricity | 24.67 pence per kWh 57.21 pence daily standing charge |
--      26.11 pence per kWh 57.19 pence daily standing charge"
--   Two columns, two rates, and the extractor asserted one of them as "the"
--   rate. The verbatim check passed because that text really is on the page.
--   Which column it is, and what distinguishes them, is unverified. Worse, it is
--   the DOMESTIC price cap.
-- CIT-3291cb4b472e8111 — the £1,862/year DOMESTIC price cap.
-- CIT-69d751743fb29e7c — £926/year against £603 pre-crisis: domestic, second-hand
--   (Carbon Brief), and self-dated "as referenced in mid-2025".
-- CIT-3782d8ebbf0a652d — "QEP 3.4.1 was last updated on 30 June 2026". True,
--   checkable, and carries no figure. A whitelist entry that licenses nothing.
--
-- THE COMMON FAULT IS AUDIENCE, AND IT IS NOT SUBTLE. Ofgem's price cap is a
-- DOMESTIC consumer protection. This site is read by people running commercial
-- growing operations, who buy on non-domestic contracts. The retired site's own
-- data layer knew this - uk-energy-prices.json carries "commercial_average" and
-- "industrial_large" and no domestic figure at all. A vertical-farm energy
-- calculator defaulting to the household cap would be wrong for every reader.
--
-- WHAT IS KEPT
-- CIT-f85f529188efb95a — "gas sets the wholesale price of electricity in the UK
--   98% of the time, according to academic research published in 2023". Relevant
--   (it is the mechanism behind UK electricity cost, which an energy explainer
--   on this site should teach), correctly scoped in its own claim text, and
--   attributed. Second-hand via Carbon Brief, which the writer_block's
--   sourcing discipline already requires be stated.
--
-- ALSO BANNED, fail-closed, on the oufe precedent: the four removed figures.
-- If one later turns out to be the right number for a stated purpose, the ban
-- forces a conscious return to this file, having first registered it properly
-- with its market and date attached. That friction is the point.
--
-- NO CARBON-INTENSITY FACT WAS RETURNED AT ALL. Half the question went
-- unanswered and the run still reported success. Re-run needed, narrower.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $guard$
DECLARE n int; f int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base row, found %', n; END IF;
  SELECT jsonb_array_length(data->'facts') INTO f FROM _cur;
  IF f IS DISTINCT FROM 15 THEN
    RAISE EXCEPTION 'expected 15 facts before quarantine, found % - another session has written here, refusing', f;
  END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id,
  'evidence_base',
  jsonb_set(
    jsonb_set(_cur.data, '{facts}',
      COALESCE((SELECT jsonb_agg(f ORDER BY ord)
                FROM jsonb_array_elements(_cur.data->'facts') WITH ORDINALITY t(f,ord)
                WHERE f->>'id' NOT IN (
                  'CIT-f8aa28d909b16c00',
                  'CIT-3291cb4b472e8111',
                  'CIT-69d751743fb29e7c',
                  'CIT-3782d8ebbf0a652d'
                )), '[]'::jsonb)),
    '{banned_claims}',
    (_cur.data->'banned_claims') || $add$[
      {"pattern": "26\\.11 ?p(ence)?",
       "reason": "QUARANTINED 2026-08-22. Registered then removed: a DOMESTIC Ofgem price-cap unit rate, extracted from a two-column table whose other column reads 24.67, so which rate it is was never established. This site's readers buy non-domestic. Unban only after registering a non-domestic figure with its market and date attached."},
      {"pattern": "£ ?1,?862",
       "reason": "QUARANTINED 2026-08-22. The DOMESTIC annual price cap. Not an operating cost for any reader of this site."},
      {"pattern": "£ ?926[^0-9]",
       "reason": "QUARANTINED 2026-08-22. Domestic annual electricity bill, second-hand and self-dated 'as referenced in mid-2025'."},
      {"pattern": "£ ?603[^0-9]",
       "reason": "QUARANTINED 2026-08-22. The pre-crisis domestic comparison figure from the same second-hand source."},
      {"pattern": "(price cap|energy price cap)[^.]{0,60}(your|the) (energy|electricity) (cost|bill|spend)",
       "reason": "AUDIENCE class, which is the fault all four removals share: the Ofgem price cap is a DOMESTIC consumer protection and never describes what a commercial grower pays. Applying it to a reader's operating cost is wrong regardless of which figure is used."}
    ]$add$::jsonb),
  'manual',
  'Phase 2 run 2 (correlation a39c979c-a563-48f4-8198-4d2be568543e) returned 5 facts, 4 of them DOMESTIC price-cap figures irrelevant to commercial growers, one of those a two-column table scrape. Removed and banned fail-closed; kept CIT-f85f529188efb95a (gas sets the wholesale price 98% of the time - relevant, scoped, attributed). No carbon-intensity fact was returned at all despite being half the question. verify_and_register is a PROVENANCE check, not a relevance or reading check (bugs_open/161).',
  true, true, 'agritec-workstream-2026-08-22'
FROM _cur;

COMMIT;

-- Verify: facts 15 -> 11, bans 24 -> 29, writer_block unchanged at 2024.
--   SELECT jsonb_array_length(data->'facts') AS facts,
--          jsonb_array_length(data->'banned_claims') AS bans,
--          length(data->>'writer_block') AS wb
--     FROM site_specs ss JOIN sites s ON s.id=ss.site_id
--    WHERE s.domain='agritec.uk' AND ss.aspect='evidence_base' AND ss.is_current;
-- And confirm the KEPT fact survived, not just that the count is right:
--   SELECT f->>'id' FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
--          LATERAL jsonb_array_elements(ss.data->'facts') f
--    WHERE s.domain='agritec.uk' AND ss.aspect='evidence_base' AND ss.is_current
--      AND f->>'id'='CIT-f85f529188efb95a';

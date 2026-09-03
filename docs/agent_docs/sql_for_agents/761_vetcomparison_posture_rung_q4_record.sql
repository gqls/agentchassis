-- 761_vetcomparison_posture_rung_q4_record.sql
--
-- Records vetcomparison.uk's POSTURE RUNG as an RFC_060 Q4 record — who declared it, when, and
-- on what basis — and so closes the second half of the `missing_evidence_register` item's
-- acceptance test. Migration 759 satisfied the first half (a register with at least one valued
-- fact); this satisfies "the posture rung is recorded with who declared it and when".
--
-- ─── WHY THIS IS A SEPARATE MIGRATION, AND WHY THE RECORD LIVES HERE ─────────────────────────
--
-- 759 is applied. Forward-only, so this is an additive follow-on rather than an edit.
--
-- ⚠ **THE Q4 RECORD HAS NO BUILT HOME.** Measured 2026-09-03: `grep -rn "Q4 record\|posture_rung\|
-- claims_posture" --include=*.go --include=*.sql` outside docs/ returns NOTHING, and the only
-- `site_specs.aspect` in the family is `evidence_base` itself. RFC_060 defines the ladder
-- (§3b) and asks for the record (Q4, §3e) but nothing stores one. So this writes it as a
-- top-level `posture` object INSIDE the register, for three reasons: the register is the
-- artefact the rung actually governs; unknown top-level keys round-trip losslessly through the
-- daily writer (`refresh_evidence_base_action.go` unmarshals to map[string]interface{}, mutates
-- only its own keys and marshals the same map back — RUNBOOK §8d, verified at the code); and it
-- puts the record where every consumer that loads a register already looks.
--
-- ⚠ **THIS IS THE FIRST SITE TO CARRY ONE, AND IT IS NOT YET A FLEET CONVENTION.** Nothing reads
-- `posture` — it is inert and additive, so it changes no guarantee of the shared mechanism and is
-- not architecture-scope by the 2026-07-29 ruling (an addition to a shared vocabulary needs an RFC
-- only when it changes what the mechanism GUARANTEES) or by RFC_022's narrowing (an opt-in field
-- with the unsafe default off and zero live consumers). **Enumerating the consumers rather than
-- asserting it: zero — `grep -rn "'posture'\|\"posture\"" --include=*.go platform/ internal/ cmd/`
-- returns nothing on 2026-09-03.** It is offered to the claims-verification lane as the shape a
-- Q4 record could take, NOT declared as the fleet's. If they choose a different home, this key is
-- inert and costs one DELETE.
--
-- ─── THE RUNG IS DECLARED ELSEWHERE; THIS RECORDS IT, IT DOES NOT DECIDE IT ──────────────────
--
-- The work item is explicit that the rung is "not something this sweep may infer". It is not
-- being inferred here. It was already declared, twice, and this migration only writes down what
-- those declarations say plus who made them:
--
--   (1) RFC_060 §3b's own worked list: "lendzy.co.uk, loancalculator.co.uk, vetcomparison.uk →
--       relied_upon", and §3a/§3b name this site specifically as the RFC's relied_upon worked
--       example because it "carries animal-health claims".
--   (2) The vetcomparison lane, in its 2026-09-03 handover (NOTES_vetcomparison.md): "Rung: the
--       CITED bar. vetcomparison is RFC_060's own relied_upon worked example."
--
-- The BASIS, in the ladder's own terms: a reader may act on these assertions to their financial,
-- legal or animal-safety detriment. The animal-health-certificate guide tells an owner when a
-- rabies vaccination must happen, how long a certificate lasts and when tapeworm treatment must
-- be given — get the 21-day wait or the 24-to-120-hour window wrong and the animal is refused at
-- the border or travels unprotected. The CMA guides tell a practice which statutory obligations
-- bind it and by when. Both are "act on it" content, which is the relied_upon test.
--
-- Rollback: 761_..._ROLLBACK.sql

BEGIN;

-- Guard 1 (RUNBOOK §8e): bind the UUID to the DOMAIN.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM sites WHERE id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND domain = 'vetcomparison.uk') THEN
    RAISE EXCEPTION '761 ABORT: site_id does not resolve to vetcomparison.uk';
  END IF;
END $$;

-- Guard 2: the register must exist and must be 759's, or this is writing a posture onto
-- something else. If the refresher has already superseded 759, that is fine and expected — but
-- the row must still be a register with facts, not an empty or absent one.
DO $$
DECLARE nfacts int; already boolean;
BEGIN
  SELECT jsonb_array_length(data->'facts'), data ? 'posture'
    INTO nfacts, already
    FROM site_specs
   WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;
  IF nfacts IS NULL THEN
    RAISE EXCEPTION '761 ABORT: vetcomparison has no current evidence_base row - run 759 first';
  END IF;
  IF nfacts < 1 THEN
    RAISE EXCEPTION '761 ABORT: the current register carries % facts - refusing to posture an empty register', nfacts;
  END IF;
  IF already THEN
    RAISE EXCEPTION '761 ABORT: a posture record already exists on this register - read it before overwriting';
  END IF;
END $$;

UPDATE site_specs
   SET data = data || jsonb_build_object(
         'posture', jsonb_build_object(
           'rung', 'relied_upon',
           'declared_by', 'RFC_060 §3a/§3b (the RFC''s own relied_upon worked example, naming this site) and the vetcomparison lane''s handover of 2026-09-03; RECORDED here by the bugfix_414 register-programme lane under owner ruling D1',
           'declared_on', '2026-09-03',
           'basis', 'A reader may act on this site''s assertions to their financial, legal or animal-safety detriment. The animal-health-certificate guide governs whether a pet may lawfully travel (the 21-full-day rabies wait, the 10-day certificate validity, the 24-to-120-hour tapeworm window); the CMA guides tell a veterinary practice which statutory obligations bind it and by when. Both are act-on-it content, which is the relied_upon test in RFC_060 §3b.',
           'requires', 'The CITED bar: every fact carries source.citation{url,quote,...} with the quote verified through the production matcher - EXCEPT where the primary source is a format the matcher cannot read, where source.attested_by is used instead and the reason is recorded on the fact (see migration 759: the CMA draft Order and Schedule 1 are PDFs, measured to return false for every quote including one certainly present).',
           'review_when', 'The substantive CMA Order is due by its statutory deadline of 23 September 2026. On the day it is made, every CMA-DRAFT-* fact in this register needs re-verification and the bracketed figures become real ones.',
           'note_on_this_field', 'FIRST USE ON THE FLEET and NOT YET A CONVENTION. Nothing reads this key (zero Go consumers, measured 2026-09-03) - it is inert and additive. Offered to the claims-verification lane as the shape RFC_060 Q4 could take; if they choose a different home it costs one DELETE.'
         ))
 WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d'
   AND aspect = 'evidence_base'
   AND is_current;

DO $$
DECLARE rung text; nfacts int; nban int; declared text;
BEGIN
  SELECT data->'posture'->>'rung', jsonb_array_length(data->'facts'),
         jsonb_array_length(data->'banned_claims'), data->'posture'->>'declared_on'
    INTO rung, nfacts, nban, declared
    FROM site_specs
   WHERE site_id = '72b9e3a6-872f-4528-a6d6-7f205ea60f4d' AND aspect = 'evidence_base' AND is_current;

  IF rung <> 'relied_upon' THEN
    RAISE EXCEPTION '761 VERIFY: expected rung relied_upon, found %', coalesce(rung,'(none)');
  END IF;
  IF declared IS NULL OR declared = '' THEN
    RAISE EXCEPTION '761 VERIFY: the posture record carries no declared_on date - a Q4 record without a date is not a Q4 record';
  END IF;
  -- The || must have ADDED a key, never REPLACED the register.
  --
  -- ⚠ `IS DISTINCT FROM`, not `<>`, and that is the whole point of this check. The first cut
  -- used `IF nfacts <> 21 OR nban <> 6` and a mutation test proved it INERT against the exact
  -- disaster it exists to catch: replacing `data = data || ...` with `data = ...` wipes the
  -- register, `jsonb_array_length(NULL)` returns NULL, `NULL <> 21` evaluates to NULL rather
  -- than TRUE, the IF body never runs and the migration reports `761 OK ... intact at <NULL>
  -- facts`. A three-way comparison is required wherever the value being checked can be NULL
  -- precisely BECAUSE the damage happened.
  IF nfacts IS DISTINCT FROM 21 OR nban IS DISTINCT FROM 6 THEN
    RAISE EXCEPTION '761 VERIFY: the register lost content - expected 21 facts and 6 banned_claims, found % and % (NULL here means the UPDATE REPLACED the register instead of extending it)', coalesce(nfacts::text,'NULL'), coalesce(nban::text,'NULL');
  END IF;

  RAISE NOTICE '761 OK: vetcomparison posture recorded - rung=% declared_on=%, register intact at % facts / % banned_claims', rung, declared, nfacts, nban;
END $$;

COMMIT;

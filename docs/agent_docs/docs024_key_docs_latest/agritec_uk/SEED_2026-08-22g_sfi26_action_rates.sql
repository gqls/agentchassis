-- ============================================================================
-- agritec.uk — register the full SFI26 action set (71 actions) by ATTESTATION
-- Written 2026-08-22. Applied out of band (psql -f), per-site setup.
--
-- WHY. The owner asked for the SFI calculator to use correct, up-to-date facts.
-- It cannot, until the rates exist in the register: the tool ENCODES them, and
-- bugs_open/288 is precisely that a figure inside a calculator is checked by
-- nothing. So the rates go in the register first, and the tool is built from it.
--
-- WHY ATTESTATION RATHER THAN A RESEARCH RUN. The rates are published in 21
-- HTML TABLES headed "Code / Action / Annual payment". A table cell has no
-- sentence to quote, so `extract_claims` must compose one and
-- `verify_and_register` can never re-match it — the `citation_lost` trap this
-- lane hit and recorded in LANDMINES the same day. Attestation is the documented
-- remedy and a normal path (48 attested facts fleet-wide as of 2026-08-22).
--
-- THE CROSS-CHECK THAT MAKES THIS MORE THAN "A SESSION SAYS SO". The parse
-- yielded 71 actions as of 2026-08-22. The page separately states, in prose,
-- "there are 71 actions (compared with 102 actions in SFI24)" — and that
-- sentence was independently citation-verified earlier today as
-- CIT-a7b7d2e75b977bf4. **The count and the extraction agree**, so a parse that
-- silently dropped a table would have shown up as a mismatch. That is what makes
-- this attestation disconfirmable rather than merely dated.
--
-- WHAT THE AUDIT FOUND, and it is worse than the management payment alone.
-- Of the RETIRED calculator's nine revenue lines, measured 2026-08-22:
--   SFI Management Payment  £20/ha first 50ha  -> ABOLISHED for SFI26
--   SAM1 Soil Assessment    £6/ha              -> no equivalent SFI26 action
--   IPM1 IPM Plan           £989/yr fixed      -> no equivalent SFI26 action
--   NUM1 Nutrient Review    £652/yr fixed      -> no equivalent SFI26 action
--   HRW1 Assess Hedgerows   £3/100m            -> no equivalent SFI26 action
--   SAM2 Cover Crop         £129/ha            -> CSAM2 £129/ha   (recode only)
--   IPM4 No Insecticide     £45/ha             -> CIPM4 £45/ha    (recode only)
--   SAM3 Herbal Leys        £382/ha            -> CSAM3 £224/ha   (41% TOO HIGH)
--   HRW2 Manage Hedgerows   £10/100m           -> CHRW2 £13 per 100m FOR ONE SIDE
-- **Two of nine lines are correct**, and only after recoding. The four removed
-- actions are the paid assessment/planning ones, which is coherent with DEFRA's
-- own stated reason for dropping the management payment: releasing funding to
-- offer more agreements.
--
-- VALUE FIELDS. 64 of 71 carry a numeric `value` because their payment string is
-- an unambiguous "£N per hectare|square metre". The other 7 are per-100m,
-- per-pond or otherwise compound ("£13 per 100m for one side", "£257 per pond
-- (maximum of 3 ponds per hectare)") and are deliberately left VALUELESS, with
-- the full string in `writer_line`. A compound rate reduced to one number is how
-- "per 100m for one side" silently becomes "per 100m", which on a hedgerow is a
-- factor-of-two error in the farmer's favour and then against them at audit.
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
  IF f IS DISTINCT FROM 33 THEN
    RAISE EXCEPTION 'expected 33 facts, found % - another session has written here, refusing', f;
  END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id, 'evidence_base',
  jsonb_set(_cur.data, '{facts}', (_cur.data->'facts') || $facts$[
 {
  "id": "ATT-sfi26-actions-table",
  "kind": "attestation",
  "claim": "The SFI26 scheme rules and guidance on GOV.UK publish the full set of SFI26 actions with their codes and annual payment rates, in tables headed Code / Action / Annual payment. 71 actions as of 2026-08-22, which matches the page's own statement that 'there are 71 actions (compared with 102 actions in SFI24)'.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22: fetched https://www.gov.uk/government/publications/sustainable-farming-incentive-2026-sfi26/sfi26-scheme-rules-and-guidance (HTTP 200) and parsed all 21 tables directly. Registered by attestation because the rates are TABLE CELLS, which no verbatim-quote check can match from a composed sentence - see LANDMINES 'citation_lost says possible hallucination'. The extracted count of 71 cross-checks against the separately citation-verified fact CIT-a7b7d2e75b977bf4."
  },
  "verified_at": "2026-08-22",
  "writer_line": "SFI26 scheme rules and guidance, GOV.UK (captured 22 August 2026)"
 },
 {
  "id": "ATT-sfi26-AGF1",
  "kind": "metric",
  "claim": "SFI26 action AGF1 (Maintain very low density in-field agroforestry on less sensitive land) has an annual payment of £248 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "AGF1 - Maintain very low density in-field agroforestry on less sensitive land: £248 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 248.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-AGF2",
  "kind": "metric",
  "claim": "SFI26 action AGF2 (Maintain low density in-field agroforestry on less sensitive land) has an annual payment of £385 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "AGF2 - Maintain low density in-field agroforestry on less sensitive land: £385 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 385.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CHRW2",
  "kind": "metric",
  "claim": "SFI26 action CHRW2 (Manage hedgerows) has an annual payment of £13 per 100m for one side.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CHRW2 - Manage hedgerows: £13 per 100m for one side (SFI26 scheme rules, GOV.UK, captured 22 August 2026)"
 },
 {
  "id": "ATT-sfi26-BND1",
  "kind": "metric",
  "claim": "SFI26 action BND1 (Maintain dry stone walls) has an annual payment of £27 per 100m for both sides.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "BND1 - Maintain dry stone walls: £27 per 100m for both sides (SFI26 scheme rules, GOV.UK, captured 22 August 2026)"
 },
 {
  "id": "ATT-sfi26-BND2",
  "kind": "metric",
  "claim": "SFI26 action BND2 (Maintain earth banks or stone-faced hedgebanks) has an annual payment of £11 per 100m for one side.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "BND2 - Maintain earth banks or stone-faced hedgebanks: £11 per 100m for one side (SFI26 scheme rules, GOV.UK, captured 22 August 2026)"
 },
 {
  "id": "ATT-sfi26-CAHL4",
  "kind": "metric",
  "claim": "SFI26 action CAHL4 (4m to 12m grass buffer strip on arable and horticultural land) has an annual payment of £515 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CAHL4 - 4m to 12m grass buffer strip on arable and horticultural land: £515 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 515.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CIGL3",
  "kind": "metric",
  "claim": "SFI26 action CIGL3 (4m to 12m grass buffer strip on improved grassland) has an annual payment of £235 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CIGL3 - 4m to 12m grass buffer strip on improved grassland: £235 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 235.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-BFS1",
  "kind": "metric",
  "claim": "SFI26 action BFS1 (12m to 24m watercourse buffer strip on cultivated land) has an annual payment of £707 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "BFS1 - 12m to 24m watercourse buffer strip on cultivated land: £707 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 707.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-BFS6",
  "kind": "metric",
  "claim": "SFI26 action BFS6 (6m to 12m habitat strip next to watercourses) has an annual payment of £742 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "BFS6 - 6m to 12m habitat strip next to watercourses: £742 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 742.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CAHL1",
  "kind": "metric",
  "claim": "SFI26 action CAHL1 (Pollen and nectar flower mix) has an annual payment of £739 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CAHL1 - Pollen and nectar flower mix: £739 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 739.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CAHL2",
  "kind": "metric",
  "claim": "SFI26 action CAHL2 (Winter bird food on arable and horticultural land) has an annual payment of £648 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CAHL2 - Winter bird food on arable and horticultural land: £648 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 648.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CAHL3",
  "kind": "metric",
  "claim": "SFI26 action CAHL3 (Grassy field corners or blocks) has an annual payment of £590 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CAHL3 - Grassy field corners or blocks: £590 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 590.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-AHW2",
  "kind": "metric",
  "claim": "SFI26 action AHW2 (Supplementary winter bird food) has an annual payment of £732 per tonne (maximum 1 tonne for every 2 hectares of CAHL2).",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "AHW2 - Supplementary winter bird food: £732 per tonne (maximum 1 tonne for every 2 hectares of CAHL2) (SFI26 scheme rules, GOV.UK, captured 22 August 2026)"
 },
 {
  "id": "ATT-sfi26-AHW3",
  "kind": "metric",
  "claim": "SFI26 action AHW3 (Beetle banks) has an annual payment of £764 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "AHW3 - Beetle banks: £764 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 764.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-AHW4",
  "kind": "metric",
  "claim": "SFI26 action AHW4 (Skylark plots) has an annual payment of £11 per plot (minimum 2 plots).",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "AHW4 - Skylark plots: £11 per plot (minimum 2 plots) (SFI26 scheme rules, GOV.UK, captured 22 August 2026)"
 },
 {
  "id": "ATT-sfi26-AHW5",
  "kind": "metric",
  "claim": "SFI26 action AHW5 (Nesting plots for lapwing) has an annual payment of £765 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "AHW5 - Nesting plots for lapwing: £765 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 765.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-AHW6",
  "kind": "metric",
  "claim": "SFI26 action AHW6 (Basic overwinter stubble) has an annual payment of £58 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "AHW6 - Basic overwinter stubble: £58 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 58.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-AHW7",
  "kind": "metric",
  "claim": "SFI26 action AHW7 (Enhanced overwinter stubble) has an annual payment of £589 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "AHW7 - Enhanced overwinter stubble: £589 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 589.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-AHW8",
  "kind": "metric",
  "claim": "SFI26 action AHW8 (Whole crop spring cereals and overwinter stubble) has an annual payment of £596 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "AHW8 - Whole crop spring cereals and overwinter stubble: £596 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 596.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-AHW9",
  "kind": "metric",
  "claim": "SFI26 action AHW9 (Unharvested cereal headland) has an annual payment of £1,072 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "AHW9 - Unharvested cereal headland: £1,072 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 1072.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-AHW10",
  "kind": "metric",
  "claim": "SFI26 action AHW10 (Low input harvested cereal crop) has an annual payment of £354 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "AHW10 - Low input harvested cereal crop: £354 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 354.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-AHW11",
  "kind": "metric",
  "claim": "SFI26 action AHW11 (Cultivated areas for arable plants) has an annual payment of £660 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "AHW11 - Cultivated areas for arable plants: £660 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 660.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CIGL1",
  "kind": "metric",
  "claim": "SFI26 action CIGL1 (Take grassland field corners or blocks out of management) has an annual payment of £333 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CIGL1 - Take grassland field corners or blocks out of management: £333 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 333.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CIGL2",
  "kind": "metric",
  "claim": "SFI26 action CIGL2 (Winter bird food on improved grassland) has an annual payment of £515 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CIGL2 - Winter bird food on improved grassland: £515 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 515.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CLIG3",
  "kind": "metric",
  "claim": "SFI26 action CLIG3 (Manage grassland with very low nutrient inputs) has an annual payment of £151 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CLIG3 - Manage grassland with very low nutrient inputs: £151 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 151.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-GRH7",
  "kind": "metric",
  "claim": "SFI26 action GRH7 (Supplement: Haymaking) has an annual payment of £157 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "GRH7 - Supplement: Haymaking: £157 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 157.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-GRH8",
  "kind": "metric",
  "claim": "SFI26 action GRH8 (Supplement: Haymaking (late cut)) has an annual payment of £187 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "GRH8 - Supplement: Haymaking (late cut): £187 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 187.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-GRH10",
  "kind": "metric",
  "claim": "SFI26 action GRH10 (Supplement: Lenient grazing) has an annual payment of £28 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "GRH10 - Supplement: Lenient grazing: £28 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 28.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-GRH12",
  "kind": "metric",
  "claim": "SFI26 action GRH12 (Manage rough grassland for upland breeding waders (instead of GRH1: Manage rough grazing for birds which was available in the SFI24 offer)) has an annual payment of £203 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "GRH12 - Manage rough grassland for upland breeding waders (instead of GRH1: Manage rough grazing for birds which was available in the SFI24 offer): £203 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 203.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-SCR1",
  "kind": "metric",
  "claim": "SFI26 action SCR1 (Create scrub and open habitat mosaics) has an annual payment of £588 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "SCR1 - Create scrub and open habitat mosaics: £588 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 588.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-SCR2",
  "kind": "metric",
  "claim": "SFI26 action SCR2 (Manage scrub and open habitat mosaics) has an annual payment of £350 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "SCR2 - Manage scrub and open habitat mosaics: £350 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 350.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-HEF1",
  "kind": "metric",
  "claim": "SFI26 action HEF1 (Maintain weatherproof traditional farm or forestry buildings) has an annual payment of £5 per square metre.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "HEF1 - Maintain weatherproof traditional farm or forestry buildings: £5 per square metre (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 5.0,
  "unit": "GBP per square metre"
 },
 {
  "id": "ATT-sfi26-HEF6",
  "kind": "metric",
  "claim": "SFI26 action HEF6 (Manage historic and archaeological features on grassland) has an annual payment of £55 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "HEF6 - Manage historic and archaeological features on grassland: £55 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 55.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CIPM2",
  "kind": "metric",
  "claim": "SFI26 action CIPM2 (Flower-rich grass margins, blocks or in-field strips) has an annual payment of £798 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CIPM2 - Flower-rich grass margins, blocks or in-field strips: £798 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 798.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CIPM3",
  "kind": "metric",
  "claim": "SFI26 action CIPM3 (Companion crop on arable and horticultural land) has an annual payment of £55 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CIPM3 - Companion crop on arable and horticultural land: £55 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 55.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CIPM4",
  "kind": "metric",
  "claim": "SFI26 action CIPM4 (No use of insecticide on arable crops and permanent crops) has an annual payment of £45 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CIPM4 - No use of insecticide on arable crops and permanent crops: £45 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 45.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-UPL1",
  "kind": "metric",
  "claim": "SFI26 action UPL1 (Moderate livestock grazing on moorland) has an annual payment of £35 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "UPL1 - Moderate livestock grazing on moorland: £35 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 35.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-UPL2",
  "kind": "metric",
  "claim": "SFI26 action UPL2 (Low livestock grazing on moorland) has an annual payment of £89 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "UPL2 - Low livestock grazing on moorland: £89 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 89.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-UPL3",
  "kind": "metric",
  "claim": "SFI26 action UPL3 (Limited livestock grazing on moorland) has an annual payment of £111 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "UPL3 - Limited livestock grazing on moorland: £111 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 111.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-UPL5",
  "kind": "metric",
  "claim": "SFI26 action UPL5 (Supplement: Keep cattle and ponies on moorland (minimum 70% GLU)) has an annual payment of £18 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "UPL5 - Supplement: Keep cattle and ponies on moorland (minimum 70% GLU): £18 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 18.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-UPL6",
  "kind": "metric",
  "claim": "SFI26 action UPL6 (Supplement: Keep cattle and ponies on moorland (100% GLU)) has an annual payment of £23 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "UPL6 - Supplement: Keep cattle and ponies on moorland (100% GLU): £23 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 23.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-UPL8",
  "kind": "metric",
  "claim": "SFI26 action UPL8 (Shepherding livestock on moorland (remove stock for at least 4 months)) has an annual payment of £74 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "UPL8 - Shepherding livestock on moorland (remove stock for at least 4 months): £74 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 74.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-UPL10",
  "kind": "metric",
  "claim": "SFI26 action UPL10 (Shepherding livestock on moorland (remove stock for at least 8 months)) has an annual payment of £102 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "UPL10 - Shepherding livestock on moorland (remove stock for at least 8 months): £102 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 102.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CNUM2",
  "kind": "metric",
  "claim": "SFI26 action CNUM2 (Legumes on improved grassland) has an annual payment of £102 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CNUM2 - Legumes on improved grassland: £102 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 102.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CNUM3",
  "kind": "metric",
  "claim": "SFI26 action CNUM3 (Legume fallow) has an annual payment of £532 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CNUM3 - Legume fallow: £532 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 532.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-OFC1",
  "kind": "metric",
  "claim": "SFI26 action OFC1 (Organic conversion – improved permanent grassland) has an annual payment of £187 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "OFC1 - Organic conversion – improved permanent grassland: £187 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 187.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-OFC2",
  "kind": "metric",
  "claim": "SFI26 action OFC2 (Organic conversion – unimproved permanent grassland) has an annual payment of £96 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "OFC2 - Organic conversion – unimproved permanent grassland: £96 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 96.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-OFC3",
  "kind": "metric",
  "claim": "SFI26 action OFC3 (Organic conversion – rotational land) has an annual payment of £298 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "OFC3 - Organic conversion – rotational land: £298 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 298.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-OFC4",
  "kind": "metric",
  "claim": "SFI26 action OFC4 (Organic conversion – horticultural land) has an annual payment of £874 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "OFC4 - Organic conversion – horticultural land: £874 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 874.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-OFC5",
  "kind": "metric",
  "claim": "SFI26 action OFC5 (Organic conversion – top fruit) has an annual payment of £1,920 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "OFC5 - Organic conversion – top fruit: £1,920 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 1920.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-OFM1",
  "kind": "metric",
  "claim": "SFI26 action OFM1 (Organic land management – improved permanent grassland) has an annual payment of £20 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "OFM1 - Organic land management – improved permanent grassland: £20 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 20.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-OFM2",
  "kind": "metric",
  "claim": "SFI26 action OFM2 (Organic land management – unimproved permanent grassland) has an annual payment of £41 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "OFM2 - Organic land management – unimproved permanent grassland: £41 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 41.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-OFM3",
  "kind": "metric",
  "claim": "SFI26 action OFM3 (Organic land management – enclosed rough grazing) has an annual payment of £97 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "OFM3 - Organic land management – enclosed rough grazing: £97 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 97.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-OFM4",
  "kind": "metric",
  "claim": "SFI26 action OFM4 (Organic land management – rotational land) has an annual payment of £132 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "OFM4 - Organic land management – rotational land: £132 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 132.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-OFM5",
  "kind": "metric",
  "claim": "SFI26 action OFM5 (Organic land management – horticultural land) has an annual payment of £707 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "OFM5 - Organic land management – horticultural land: £707 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 707.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-OFM6",
  "kind": "metric",
  "claim": "SFI26 action OFM6 (Organic land management – top fruit) has an annual payment of £1,920 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "OFM6 - Organic land management – top fruit: £1,920 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 1920.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-PRF1",
  "kind": "metric",
  "claim": "SFI26 action PRF1 (Variable rate application of nutrients) has an annual payment of £27 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "PRF1 - Variable rate application of nutrients: £27 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 27.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-PRF2",
  "kind": "metric",
  "claim": "SFI26 action PRF2 (Camera or remote sensor guided herbicide spraying) has an annual payment of £43 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "PRF2 - Camera or remote sensor guided herbicide spraying: £43 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 43.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-PRF4",
  "kind": "metric",
  "claim": "SFI26 action PRF4 (Mechanical robotic weeding) has an annual payment of £150 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "PRF4 - Mechanical robotic weeding: £150 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 150.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CSAM2",
  "kind": "metric",
  "claim": "SFI26 action CSAM2 (Multi-species winter cover crop) has an annual payment of £129 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CSAM2 - Multi-species winter cover crop: £129 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 129.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-CSAM3",
  "kind": "metric",
  "claim": "SFI26 action CSAM3 (Herbal leys) has an annual payment of £224 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "CSAM3 - Herbal leys: £224 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 224.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-SOH1",
  "kind": "metric",
  "claim": "SFI26 action SOH1 (No-till farming) has an annual payment of £73 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "SOH1 - No-till farming: £73 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 73.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-SOH3",
  "kind": "metric",
  "claim": "SFI26 action SOH3 (Multi-species summer-sown cover crop) has an annual payment of £163 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "SOH3 - Multi-species summer-sown cover crop: £163 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 163.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-SPM3",
  "kind": "metric",
  "claim": "SFI26 action SPM3 (Supplement: Keep native breeds on grazed habitats (more than 80%)) has an annual payment of £146 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "SPM3 - Supplement: Keep native breeds on grazed habitats (more than 80%): £146 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 146.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-SPM5",
  "kind": "metric",
  "claim": "SFI26 action SPM5 (Supplement: Keep native breeds on extensively managed habitats (more than 80%)) has an annual payment of £11 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "SPM5 - Supplement: Keep native breeds on extensively managed habitats (more than 80%): £11 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 11.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-WBD1",
  "kind": "metric",
  "claim": "SFI26 action WBD1 (Manage ponds) has an annual payment of £257 per pond (maximum of 3 ponds per hectare).",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "WBD1 - Manage ponds: £257 per pond (maximum of 3 ponds per hectare) (SFI26 scheme rules, GOV.UK, captured 22 August 2026)"
 },
 {
  "id": "ATT-sfi26-WBD2",
  "kind": "metric",
  "claim": "SFI26 action WBD2 (Manage ditches) has an annual payment of £4 per 100m for both sides.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "WBD2 - Manage ditches: £4 per 100m for both sides (SFI26 scheme rules, GOV.UK, captured 22 August 2026)"
 },
 {
  "id": "ATT-sfi26-WBD3",
  "kind": "metric",
  "claim": "SFI26 action WBD3 (In-field grass strips) has an annual payment of £765 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "WBD3 - In-field grass strips: £765 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 765.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-WBD4",
  "kind": "metric",
  "claim": "SFI26 action WBD4 (Arable reversion to grassland with low fertiliser input) has an annual payment of £489 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "WBD4 - Arable reversion to grassland with low fertiliser input: £489 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 489.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-WBD6",
  "kind": "metric",
  "claim": "SFI26 action WBD6 (Remove livestock from intensive grassland during the autumn and winter (outside SDAs)) has an annual payment of £115 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "WBD6 - Remove livestock from intensive grassland during the autumn and winter (outside SDAs): £115 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 115.0,
  "unit": "GBP per hectare"
 },
 {
  "id": "ATT-sfi26-WBD7",
  "kind": "metric",
  "claim": "SFI26 action WBD7 (Remove livestock from grassland during the autumn and winter (SDAs)) has an annual payment of £115 per hectare.",
  "source": {
   "attested_by": "agritec_uk lane, 2026-08-22, from the SFI26 scheme rules tables on GOV.UK (see ATT-sfi26-actions-table)"
  },
  "verified_at": "2026-08-22",
  "writer_line": "WBD7 - Remove livestock from grassland during the autumn and winter (SDAs): £115 per hectare (SFI26 scheme rules, GOV.UK, captured 22 August 2026)",
  "value": 115.0,
  "unit": "GBP per hectare"
 }
]$facts$::jsonb),
  'manual',
  'Registered the full SFI26 action set by attestation: 1 provenance fact + 71 action rates, parsed from the 21 rate tables in the GOV.UK SFI26 scheme rules on 2026-08-22. Count cross-checks against citation-verified fact CIT-a7b7d2e75b977bf4 (71 actions vs 102 in SFI24). 64 carry a numeric value; 7 compound rates (per 100m one side, per pond) left valueless with the full string in writer_line.',
  true, true, 'agritec-workstream-2026-08-22'
FROM _cur;

COMMIT;

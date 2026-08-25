-- ============================================================================
-- agritec.uk — give 83 citations the `quote` that makes them re-provable
-- Written 2026-08-25. Applied out of band (psql -f), per-site setup.
--
-- THE DEFECT, AND IT IS MINE. On 2026-08-24 I added `source.citation.{url,title,
-- publisher,captured}` to all 83 attested facts so every figure would be
-- clickable, per the owner's instruction to cite the numbers. I did not add
-- `quote`. A citation without its verbatim quote cannot be re-proved:
-- `refreshCitationFact` needs the stored sentence to re-match against the live
-- page, so all 83 error on every sweep.
--
-- AND THE ERROR IS SILENT BY CONSTRUCTION. `citationDateStale` returns false when
-- `staleness_days <= 0` (evidence_citations.go:113-115, "No staleness_days ->
-- never stale by age"), and mine were all unset. So a citation that can NEVER be
-- re-proved never ages into a drift, never files an item, never escalates:
-- error -> skipped_unknown -> nothing. CLM-008 working exactly as designed
-- (unknown is not loss), and the right rule — but it meant **the evidence half of
-- this register was structurally inert and read as healthy.**
--
-- Measured before fixing: 104 facts carry a citation, 21 have a quote, 83 do not,
-- and all 83 of those also lack staleness_days.
--
-- WHAT MAKES THIS FIXABLE AT ALL. The rates live in TABLES, which is why they
-- were attested rather than citation-verified in the first place. But a table row
-- DOES appear in the page's extracted text: "CSAM3 Herbal leys £224 per hectare"
-- is present verbatim on the GOV.UK page, and "Lettuce 12−17" (with U+2212 MINUS
-- SIGN, not a hyphen) is present on the Virginia Tech page. So the quote is real;
-- what was missing was the discipline of writing it down.
--
-- EVERY QUOTE BELOW WAS VERIFIED PRESENT IN THE LIVE SOURCE BEFORE BEING STORED —
-- generated from each fact's own claim, then checked against a fresh fetch of the
-- page. 83 of 83. Four needed hand-work and are worth naming because each shows
-- how a composed quote fails:
--   UPL5, UPL6  the page renders "(minimum 70% GLU )" with a space before the
--               paren, so the composed form did not match;
--   the two provenance attestations are not rate rows, so they take the sentence
--   that actually supports each: "there are 71 actions (compared with 102 actions
--   in SFI24)" and "Recommended DLI range".
--
-- STALENESS. Set so that a citation which cannot be re-proved eventually becomes
-- a human's problem instead of a permanent quiet error: 400 days for the SFI
-- policy rates (they move with scheme years), 800 for the horticultural DLI
-- guidance. The seven period-bounded facts keep their 3650 and are untouched.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE _cur ON COMMIT DROP AS
SELECT ss.id, ss.site_id, ss.data
FROM site_specs ss JOIN sites s ON s.id = ss.site_id
WHERE s.domain = 'agritec.uk' AND ss.aspect = 'evidence_base' AND ss.is_current;

DO $guard$
DECLARE n int; noq int;
BEGIN
  SELECT count(*) INTO n FROM _cur;
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 current evidence_base row, found %', n; END IF;
  SELECT count(*) INTO noq FROM _cur, LATERAL jsonb_array_elements(data->'facts') x
   WHERE x->'source' ? 'citation' AND NOT (x->'source'->'citation' ? 'quote');
  IF noq <> 83 THEN RAISE EXCEPTION 'expected 83 quote-less citations, found % - state has moved', noq; END IF;
END
$guard$;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE id IN (SELECT id FROM _cur);

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  _cur.site_id, 'evidence_base',
  jsonb_set(_cur.data, '{facts}',
    (SELECT jsonb_agg(
       CASE WHEN (q.map ? (f->>'id')) AND NOT (f->'source'->'citation' ? 'quote')
            THEN jsonb_set(
                   jsonb_set(f, '{source,citation,quote}', q.map -> (f->>'id')),
                   '{staleness_days}',
                   CASE WHEN f->>'id' LIKE 'ATT-dli-%' THEN '800'::jsonb ELSE '400'::jsonb END)
            ELSE f END ORDER BY ord)
     FROM jsonb_array_elements(_cur.data->'facts') WITH ORDINALITY t(f,ord))),
  'manual',
  'Added the verbatim quote to all 83 attested-fact citations, each verified present in a fresh fetch of its source before storing, plus a staleness policy (400 SFI, 800 DLI). Without quote, refreshCitationFact errors every sweep and citationDateStale never escalates it, so the evidence half of this register was inert and read as healthy.',
  true, true, 'agritec-workstream-2026-08-25'
FROM _cur, (SELECT $q${
 "ATT-dli-seedlings": "Seedlings 5−10",
 "ATT-dli-microgreens": "Micro-greens 9−12",
 "ATT-dli-lettuce": "Lettuce 12−17",
 "ATT-dli-spinach": "Spinach 14−20",
 "ATT-dli-parsley": "Parsley 10−15",
 "ATT-dli-cilantro": "Cilantro 15−20",
 "ATT-dli-basil": "Basil 15−25",
 "ATT-dli-tomato": "Tomato 20−30",
 "ATT-dli-cucumber": "Cucumber 20−30",
 "ATT-dli-zucchini": "Zucchini 20−30",
 "ATT-sfi26-AGF1": "AGF1 Maintain very low density in-field agroforestry on less sensitive land £248 per hectare",
 "ATT-sfi26-AGF2": "AGF2 Maintain low density in-field agroforestry on less sensitive land £385 per hectare",
 "ATT-sfi26-CHRW2": "CHRW2 Manage hedgerows £13 per 100m for one side",
 "ATT-sfi26-BND1": "BND1 Maintain dry stone walls £27 per 100m for both sides",
 "ATT-sfi26-BND2": "BND2 Maintain earth banks or stone-faced hedgebanks £11 per 100m for one side",
 "ATT-sfi26-CAHL4": "CAHL4 4m to 12m grass buffer strip on arable and horticultural land £515 per hectare",
 "ATT-sfi26-CIGL3": "CIGL3 4m to 12m grass buffer strip on improved grassland £235 per hectare",
 "ATT-sfi26-BFS1": "BFS1 12m to 24m watercourse buffer strip on cultivated land £707 per hectare",
 "ATT-sfi26-BFS6": "BFS6 6m to 12m habitat strip next to watercourses £742 per hectare",
 "ATT-sfi26-CAHL1": "CAHL1 Pollen and nectar flower mix £739 per hectare",
 "ATT-sfi26-CAHL2": "CAHL2 Winter bird food on arable and horticultural land £648 per hectare",
 "ATT-sfi26-CAHL3": "CAHL3 Grassy field corners or blocks £590 per hectare",
 "ATT-sfi26-AHW2": "AHW2 Supplementary winter bird food £732 per tonne (maximum 1 tonne for every 2 hectares of CAHL2)",
 "ATT-sfi26-AHW3": "AHW3 Beetle banks £764 per hectare",
 "ATT-sfi26-AHW4": "AHW4 Skylark plots £11 per plot (minimum 2 plots)",
 "ATT-sfi26-AHW5": "AHW5 Nesting plots for lapwing £765 per hectare",
 "ATT-sfi26-AHW6": "AHW6 Basic overwinter stubble £58 per hectare",
 "ATT-sfi26-AHW7": "AHW7 Enhanced overwinter stubble £589 per hectare",
 "ATT-sfi26-AHW8": "AHW8 Whole crop spring cereals and overwinter stubble £596 per hectare",
 "ATT-sfi26-AHW9": "AHW9 Unharvested cereal headland £1,072 per hectare",
 "ATT-sfi26-AHW10": "AHW10 Low input harvested cereal crop £354 per hectare",
 "ATT-sfi26-AHW11": "AHW11 Cultivated areas for arable plants £660 per hectare",
 "ATT-sfi26-CIGL1": "CIGL1 Take grassland field corners or blocks out of management £333 per hectare",
 "ATT-sfi26-CIGL2": "CIGL2 Winter bird food on improved grassland £515 per hectare",
 "ATT-sfi26-CLIG3": "CLIG3 Manage grassland with very low nutrient inputs £151 per hectare",
 "ATT-sfi26-GRH7": "GRH7 Supplement: Haymaking £157 per hectare",
 "ATT-sfi26-GRH8": "GRH8 Supplement: Haymaking (late cut) £187 per hectare",
 "ATT-sfi26-GRH10": "GRH10 Supplement: Lenient grazing £28 per hectare",
 "ATT-sfi26-GRH12": "GRH12 Manage rough grassland for upland breeding waders (instead of GRH1: Manage rough grazing for birds which was available in the SFI24 offer) £203 per hectare",
 "ATT-sfi26-SCR1": "SCR1 Create scrub and open habitat mosaics £588 per hectare",
 "ATT-sfi26-SCR2": "SCR2 Manage scrub and open habitat mosaics £350 per hectare",
 "ATT-sfi26-HEF1": "HEF1 Maintain weatherproof traditional farm or forestry buildings £5 per square metre",
 "ATT-sfi26-HEF6": "HEF6 Manage historic and archaeological features on grassland £55 per hectare",
 "ATT-sfi26-CIPM2": "CIPM2 Flower-rich grass margins, blocks or in-field strips £798 per hectare",
 "ATT-sfi26-CIPM3": "CIPM3 Companion crop on arable and horticultural land £55 per hectare",
 "ATT-sfi26-CIPM4": "CIPM4 No use of insecticide on arable crops and permanent crops £45 per hectare",
 "ATT-sfi26-UPL1": "UPL1 Moderate livestock grazing on moorland £35 per hectare",
 "ATT-sfi26-UPL2": "UPL2 Low livestock grazing on moorland £89 per hectare",
 "ATT-sfi26-UPL3": "UPL3 Limited livestock grazing on moorland £111 per hectare",
 "ATT-sfi26-UPL8": "UPL8 Shepherding livestock on moorland (remove stock for at least 4 months) £74 per hectare",
 "ATT-sfi26-UPL10": "UPL10 Shepherding livestock on moorland (remove stock for at least 8 months) £102 per hectare",
 "ATT-sfi26-CNUM2": "CNUM2 Legumes on improved grassland £102 per hectare",
 "ATT-sfi26-CNUM3": "CNUM3 Legume fallow £532 per hectare",
 "ATT-sfi26-OFC1": "OFC1 Organic conversion – improved permanent grassland £187 per hectare",
 "ATT-sfi26-OFC2": "OFC2 Organic conversion – unimproved permanent grassland £96 per hectare",
 "ATT-sfi26-OFC3": "OFC3 Organic conversion – rotational land £298 per hectare",
 "ATT-sfi26-OFC4": "OFC4 Organic conversion – horticultural land £874 per hectare",
 "ATT-sfi26-OFC5": "OFC5 Organic conversion – top fruit £1,920 per hectare",
 "ATT-sfi26-OFM1": "OFM1 Organic land management – improved permanent grassland £20 per hectare",
 "ATT-sfi26-OFM2": "OFM2 Organic land management – unimproved permanent grassland £41 per hectare",
 "ATT-sfi26-OFM3": "OFM3 Organic land management – enclosed rough grazing £97 per hectare",
 "ATT-sfi26-OFM4": "OFM4 Organic land management – rotational land £132 per hectare",
 "ATT-sfi26-OFM5": "OFM5 Organic land management – horticultural land £707 per hectare",
 "ATT-sfi26-OFM6": "OFM6 Organic land management – top fruit £1,920 per hectare",
 "ATT-sfi26-PRF1": "PRF1 Variable rate application of nutrients £27 per hectare",
 "ATT-sfi26-PRF2": "PRF2 Camera or remote sensor guided herbicide spraying £43 per hectare",
 "ATT-sfi26-PRF4": "PRF4 Mechanical robotic weeding £150 per hectare",
 "ATT-sfi26-CSAM2": "CSAM2 Multi-species winter cover crop £129 per hectare",
 "ATT-sfi26-CSAM3": "CSAM3 Herbal leys £224 per hectare",
 "ATT-sfi26-SOH1": "SOH1 No-till farming £73 per hectare",
 "ATT-sfi26-SOH3": "SOH3 Multi-species summer-sown cover crop £163 per hectare",
 "ATT-sfi26-SPM3": "SPM3 Supplement: Keep native breeds on grazed habitats (more than 80%) £146 per hectare",
 "ATT-sfi26-SPM5": "SPM5 Supplement: Keep native breeds on extensively managed habitats (more than 80%) £11 per hectare",
 "ATT-sfi26-WBD1": "WBD1 Manage ponds £257 per pond (maximum of 3 ponds per hectare)",
 "ATT-sfi26-WBD2": "WBD2 Manage ditches £4 per 100m for both sides",
 "ATT-sfi26-WBD3": "WBD3 In-field grass strips £765 per hectare",
 "ATT-sfi26-WBD4": "WBD4 Arable reversion to grassland with low fertiliser input £489 per hectare",
 "ATT-sfi26-WBD6": "WBD6 Remove livestock from intensive grassland during the autumn and winter (outside SDAs) £115 per hectare",
 "ATT-sfi26-WBD7": "WBD7 Remove livestock from grassland during the autumn and winter (SDAs) £115 per hectare",
 "ATT-sfi26-UPL5": "UPL5 Supplement: Keep cattle and ponies on moorland (minimum 70% GLU ) £18 per hectare",
 "ATT-sfi26-UPL6": "UPL6 Supplement: Keep cattle and ponies on moorland (100% GLU ) £23 per hectare",
 "ATT-sfi26-actions-table": "there are 71 actions (compared with 102 actions in SFI24)",
 "ATT-dli-vt-table3": "Recommended DLI range"
}$q$::jsonb AS map) q;

COMMIT;


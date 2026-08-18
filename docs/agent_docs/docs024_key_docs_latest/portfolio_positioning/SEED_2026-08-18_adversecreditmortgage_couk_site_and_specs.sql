-- ============================================================================
-- adversecreditmortgage.co.uk (M5) — site row + evidence_base + imagery_style_guide
-- Wave 1 build #1. Written 2026-08-18. Applied out of band (psql -f).
-- Pattern: SEED_2026-08-17_remortgagecalculator_uk_site_and_specs.sql, WITH ITS
-- ESCAPING BUG ALREADY FIXED — see the banned_claims note below.
--
-- WHY THESE THREE (unchanged): email present or the hallucinated-email check fails
-- OPEN (bugs_open/063); evidence_base present or the whole claims layer silently
-- no-ops (loadEvidenceBase returns nil); imagery_style_guide present or
-- content_hero generates unstyled (bugs_closed/027).
--
-- *** ESCAPING: SINGLE backslash in this file. ***
-- The pilot seed used \\\\b and stored a literal backslash, leaving all six patterns
-- valid-but-inert. Dollar-quoting passes bytes literally; JSON then unescapes; so
-- `\\b` in this file becomes `\b` in the stored pattern, which is what Go's
-- regexp.Compile needs. Verified by probing, not by counting — see the verify block
-- and datahelpers/claims_banned_pattern_escaping_test.go.
--
-- *** COMPLIANCE — THIS SITE HAS A HARD RULE THE OTHERS DO NOT ***
-- The register's M5 entry states it outright: "no 'guaranteed acceptance' language,
-- ever." This audience is people who have been declined — CCJs, defaults, DMPs,
-- post-bankruptcy — and a promise of acceptance is both a financial-promotion
-- exposure and a cruelty. The banned_claims below make it mechanical rather than a
-- matter of the writer's judgement, and the writer_block states the register's own
-- "reassurance, zero judgement" register.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

INSERT INTO sites (domain, name, network_id, status, email, company_name)
VALUES ('adversecreditmortgage.co.uk','adversecreditmortgage.co.uk',
        '00000000-0000-0000-0000-000000000002','active',
        'adversecreditmortgage-couk@contactforsales.com','Adverse Credit Mortgage')
ON CONFLICT (domain) DO UPDATE SET email = COALESCE(sites.email, EXCLUDED.email);

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain='adversecreditmortgage.co.uk')
  AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT id, 'evidence_base',
$eb${
  "governing_rule": "Every rate, percentage, threshold, fee, eligibility criterion and quoted phrase about a lender, product or regulator must trace to a registered fact carrying a source URL and a capture date. This site addresses people who have been refused credit: it must never state or imply that acceptance is assured, that any particular lender will say yes, or that a person's specific circumstances will be approved. Where something depends on the individual or the lender, say so plainly. 'This depends on your circumstances' and 'we have not verified this' are always publishable; a reassuring guess never is. This site is educational: it is not advice, not a recommendation, not a broker, and not a financial promotion.",
  "facts": [],
  "banned_claims": [
    {"pattern": "\\bguaranteed (acceptance|approval|mortgage|yes)\\b",
     "reason": "The register's M5 rule, verbatim: no guaranteed-acceptance language, ever. It is unprovable, it is a financial-promotion exposure, and to this audience it is predatory."},
    {"pattern": "\\b(everyone|anyone|all applicants) (is|are|will be) (accepted|approved)\\b",
     "reason": "The same promise wearing a different grammar."},
    {"pattern": "\\bno (credit )?checks?\\b",
     "reason": "A no-credit-check mortgage does not exist in the regulated UK market; the phrase is a marker of the unregulated end and must never appear here."},
    {"pattern": "\\bbad credit (is )?(no|not a) (problem|issue|barrier)\\b",
     "reason": "Dismisses the reader's actual situation and implies an outcome the site cannot know."},
    {"pattern": "\\b(we|our team) (can|will) (get|secure) you (a|the) (mortgage|deal|approval)\\b",
     "reason": "We are not a broker and arrange nothing; this would misrepresent what the site is."},
    {"pattern": "\\b[0-9]+(\\.[0-9]+)?% (apr|apcr|rate)\\b",
     "reason": "A literal rate in copy is a price fact - stale within days and, under a named lender, a financial promotion."}
  ],
  "writer_block": "NUMBERS: no registered facts yet, so state no rate, fee, threshold or acceptance statistic as fact. Worked examples must be visibly hypothetical.\nOUTCOMES: never assert or imply that an application will succeed. The honest sentence is what affects a decision and who decides it - the lender does, not us.\nREGISTER: reassurance with zero judgement (the register's words). The reader has been declined and very likely feels judged already. No moralising about how the adverse credit arose, no 'unfortunately', no implied fault. Plain, calm, practical.\nLENDERS: name a lender only via the verified directory and only for NON-PRICE facts - regulator status, product types, underwriter, established year.\nUNCERTAINTY: 'this depends on your lender' and 'we have not verified this' are always publishable and preferred to a confident guess.",
  "writer_block_managed": true
}$eb$::jsonb,
'seed',
'Wave 1 build #1, seeded 2026-08-18. Facts roster deliberately EMPTY - no rate or eligibility threshold could be verified against a live source; cited lender material arrives via the mortgage-lender directory (DIR-001). banned_claims carry the register M5 compliance rule mechanically.',
true, false, 'portfolio_positioning Wave 1'
FROM sites WHERE domain='adversecreditmortgage.co.uk';

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain='adversecreditmortgage.co.uk')
  AND aspect = 'imagery_style_guide' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT id, 'imagery_style_guide',
$ig${
  "mood": "calm, plain, non-judgemental, quietly hopeful - never triumphant and never pitying",
  "medium": "flat editorial illustration; ordinary UK domestic architecture and everyday objects; no people's faces",
  "palette": "deep slate grounds, warm paper off-white, a single muted green accent, soft stone secondaries",
  "avoid": "red, warning signs, downward arrows, broken/cracked imagery, credit-score dials, debt iconography, stressed or head-in-hands figures, celebration/keys-in-hand cliches, piggy banks, coin stacks, percentage motifs, watermarks, text overlays",
  "kinds": {
    "content_hero": {
      "medium": "flat duotone editorial illustration",
      "mood": "steady geometric composition - a door, a window, a horizon; open rather than closed",
      "palette": "deep slate ground, muted green flat shapes and linework, warm off-white accents only",
      "avoid": "photorealism, gradients, 3D, drop shadows, text, lettering, logos, watermarks, busy detail, colour outside the palette, pale or white background, red of any kind, downward arrows, gauges or dials",
      "reference_asset_keys": []
    }
  },
  "reference_asset_keys": []
}$ig$::jsonb,
'seed',
'Wave 1 build #1, 2026-08-18. Palette pinned here deliberately (generic_theme misfires fleet-wide). The avoid-list is compliance-adjacent: credit-score dials and debt iconography would visually assert what the copy is forbidden to say.',
true, false, 'portfolio_positioning Wave 1'
FROM sites WHERE domain='adversecreditmortgage.co.uk';

-- ── verify: structural AND escaping-integrity (the pilot's lesson) ──────────
DO $do$
DECLARE sid uuid; n int; bad int;
BEGIN
    SELECT id INTO sid FROM sites WHERE domain='adversecreditmortgage.co.uk';
    IF sid IS NULL THEN RAISE EXCEPTION 'verify: no site row'; END IF;

    SELECT count(*) INTO n FROM sites
    WHERE domain='adversecreditmortgage.co.uk' AND COALESCE(email,'') <> '';
    IF n <> 1 THEN RAISE EXCEPTION 'verify: no email - bugs_open/063 fails OPEN without one'; END IF;

    SELECT count(*) INTO n FROM site_specs WHERE site_id=sid AND aspect='evidence_base' AND is_current;
    IF n <> 1 THEN RAISE EXCEPTION 'verify: expected 1 current evidence_base, found %', n; END IF;
    SELECT count(*) INTO n FROM site_specs WHERE site_id=sid AND aspect='imagery_style_guide' AND is_current;
    IF n <> 1 THEN RAISE EXCEPTION 'verify: expected 1 current imagery_style_guide, found %', n; END IF;

    -- The check the pilot's seed did NOT have. A double backslash is the signature
    -- of the escaping bug that left six patterns valid-but-inert; position()/chr(92)
    -- is used rather than LIKE, whose backslash-as-escape semantics made the first
    -- attempt unreadable.
    SELECT count(*) INTO bad
    FROM site_specs ss, jsonb_array_elements(ss.data->'banned_claims') e
    WHERE ss.site_id=sid AND ss.aspect='evidence_base' AND ss.is_current
      AND position(chr(92)||chr(92) in (e->>'pattern')) > 0;
    IF bad > 0 THEN
        RAISE EXCEPTION 'verify: % banned_claims pattern(s) double-escaped - they would compile and match NOTHING', bad;
    END IF;

    SELECT count(*) INTO n
    FROM site_specs ss, jsonb_array_elements(ss.data->'banned_claims') e
    WHERE ss.site_id=sid AND ss.aspect='evidence_base' AND ss.is_current
      AND position(chr(92)||'b' in (e->>'pattern')) > 0;
    IF n < 6 THEN RAISE EXCEPTION 'verify: only % pattern(s) carry a word boundary, expected 6', n; END IF;
END $do$;

COMMIT;

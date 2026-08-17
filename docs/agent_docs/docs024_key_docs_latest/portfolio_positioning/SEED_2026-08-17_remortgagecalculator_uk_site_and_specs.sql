-- ============================================================================
-- remortgagecalculator.uk — site row + evidence_base + imagery_style_guide
-- Phase C PILOT seed. Written 2026-08-17 by the portfolio_positioning lane.
-- Applied out of band (psql -f), NOT via the migration runner: per-site setup,
-- not a platform schema change. Pattern follows oufe's worked example
-- (docs024_key_docs_latest/oufe/SEED_2026-07-25_oufe_site_and_specs.sql).
--
-- WHY THESE THREE, AND WHY BEFORE SUBMISSION (unchanged reasoning from oufe)
--   1. sites row WITH AN EMAIL. bugs_open/063: the hallucinated-email check
--      FAILS OPEN when a site has no contact email, and a fabricated address
--      reached production for hours that way. ensure_site_record upserts on
--      domain, so pre-creating is safe and wins the race.
--   2. evidence_base. The whole claims layer is gated on the PRESENCE of this
--      aspect — loadEvidenceBase returns nil and every lane silently no-ops
--      (validate_page_content.go:727-746). Seeding before the first page is
--      written is the only way the first page is covered.
--   3. imagery_style_guide. bugs_closed/027: content_hero generates unstyled
--      on any site that has none, and a brand-new site has none.
--
-- *** THE FACTS ROSTER IS DELIBERATELY EMPTY. READ THIS BEFORE "FIXING" IT. ***
--   A remortgage site is made of numbers — rates, LTV bands, ERC percentages,
--   SDLT thresholds. NONE are seeded here, because this session could not
--   verify any of them against a live source, and a plausible figure carrying
--   an invented source URL is precisely the fabrication this layer exists to
--   stop. An empty roster is not an oversight and it is not inert:
--     - `governing_rule` + `writer_block` below still constrain the writer, and
--       banned_claims still fire — those do the load-bearing work when the
--       roster is empty (the oufe seed's own finding).
--     - build-site-planner handles it explicitly: with no facts listed it is
--       told to "use plain string section entries and no facts keys", so the
--       plan is coherent rather than half-populated.
--     - The site's CITED material arrives by a different door: the
--       mortgage-lender DIRECTORY (kind `mortgage-lender`, DIR-001), whose every
--       entry carries a verbatim quote re-verified against its source. That is
--       the Phase B machinery this pilot exists to exercise.
--   TO ADD A FACT LATER: register it with a source URL and a capture date, the
--   same shape as loanandmortgagecalculator's APPLIED_2026-08-15 evidence_base.
--   Do not copy that site's SDLT facts across — they are PURCHASE facts, dated
--   for that site, and remortgaging is not a purchase.
--
-- COMPLIANCE (owner ruling 2026-08-12/13, carried from the finance-kinds policy)
--   NON-PRICE facts only for lender material: regulator status, product types,
--   underwriter, established year — never APR/rates/premiums. A stale price
--   under a named regulated firm is a financial-promotion exposure. The
--   writer_block below states it at the site level too, because the directory
--   policy lives in the researcher prompt and cannot reach this site's own copy.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------- site row --
INSERT INTO sites (domain, name, network_id, status, email, company_name)
VALUES (
  'remortgagecalculator.uk',
  'remortgagecalculator.uk',
  '00000000-0000-0000-0000-000000000002',
  'active',
  'remortgagecalculator-uk@contactforsales.com',
  'Remortgage Calculator'
)
ON CONFLICT (domain) DO UPDATE
  SET email = COALESCE(sites.email, EXCLUDED.email);
-- NOTE: status='active' is what upsertSite writes, but it is NOT in the
-- validated vocabulary (draft/building/review/published/deployed/archived/
-- error). Never scope a query by it and expect meaning.

-- ----------------------------------------------------------- evidence_base --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'remortgagecalculator.uk')
  AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'evidence_base',
  $eb${
    "governing_rule": "Every rate, percentage, threshold, fee, date and quoted phrase about a lender, product, regulator or tax rule must trace to a registered fact below carrying a source URL and a capture date. Where no verified fact exists the page says so plainly - 'we have not verified this' and 'check with your lender' are always publishable, and a plausible estimate never is. Worked examples in calculator copy must be labelled as illustrations and use round, obviously-hypothetical figures. This site is educational and a calculation tool: it is not advice, not a recommendation, and not a financial promotion.",
    "facts": [],
    "banned_claims": [
      {"pattern": "(?i)\\\\bguaranteed (acceptance|approval|rate|saving)\\\\b", "reason": "Guarantee language is a financial-promotion exposure and is never true of a remortgage outcome."},
      {"pattern": "(?i)\\\\b(best|cheapest|lowest) (rate|deal)s? (available|on the market|in the uk)\\\\b", "reason": "A market-wide superlative is unverifiable and dates within hours."},
      {"pattern": "(?i)\\\\bsave (up to )?\\u00a3[0-9,]+", "reason": "A specific saving figure is a per-borrower calculation, not a site fact; it must come from the calculator with the user's own inputs, never from copy."},
      {"pattern": "(?i)\\\\b(we are|we're) (fca|financially) (regulated|authorised)\\\\b", "reason": "This site is not a regulated firm and must never imply it is."},
      {"pattern": "(?i)\\\\b(you (will|would) save|this will save you)\\\\b", "reason": "Second-person outcome promises assert a result the site cannot know."},
      {"pattern": "(?i)\\\\b[0-9]+(\\\\.[0-9]+)?% (apr|apcr|rate)\\\\b", "reason": "A literal rate in copy is a price fact - stale within days and, under a named lender, a financial promotion. Rates belong in the calculator inputs, never in prose."}
    ],
    "writer_block": "NUMBERS: this site has NO registered facts yet, so state no rate, percentage, threshold, fee or saving as fact. Do not write a figure you cannot point at. Calculator worked examples are allowed and must be visibly hypothetical ('if your balance were \\u00a3200,000 ...'), never presented as current market values.\nLENDERS: name a lender only via the verified lender directory, and only for NON-PRICE facts - regulator status, product types, underwriter, established year. Never a rate, APR or fee for a named lender, whatever the source.\nSTANCE: urgency without alarm. The reader's fix is ending within six months; the job is a clear deadline, a clear calculation and a clear next step. No pressure language, no guarantees, no 'act now'.\nUNCERTAINTY: 'this depends on your lender' and 'we have not verified this' are always publishable and are preferred to a confident guess.",
    "writer_block_managed": true
  }$eb$::jsonb,
  'seed',
  'Phase C pilot seed 2026-08-17. Facts roster deliberately EMPTY - see the file header: no rate/threshold fact could be verified against a live source in this session, and an invented source is the failure this layer exists to prevent. Cited lender material arrives via the mortgage-lender directory (DIR-001), not from here.',
  true,
  false,
  'portfolio_positioning Phase C pilot seed'
FROM sites WHERE domain = 'remortgagecalculator.uk';

-- ----------------------------------------------------- imagery_style_guide --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'remortgagecalculator.uk')
  AND aspect = 'imagery_style_guide' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'imagery_style_guide',
  $ig${
    "mood": "calm, orderly, dependable, unhurried - the opposite of a hard-sell finance advert",
    "medium": "flat editorial illustration and clean architectural photography of ordinary UK homes; no people's faces",
    "palette": "deep navy grounds, warm off-white, a single confident teal accent, muted slate secondaries",
    "avoid": "stock photos of smiling couples with keys or paperwork, piggy banks, stacks of coins, percentage-sign motifs, red urgency banners, arrows pointing down at money, cartoon houses, watermarks, text overlays, anything implying a guaranteed outcome",
    "kinds": {
      "content_hero": {
        "medium": "flat duotone editorial illustration",
        "mood": "calm geometric composition - roofline, door, window, a clean horizon; simple and legible at small sizes",
        "palette": "deep navy ground, teal flat shapes and linework, warm off-white secondary accents only",
        "avoid": "photorealism, photographic texture, gradients, 3D rendering, drop shadows, text, lettering, logos, watermarks, busy detail, colour outside the palette, white background, pale background, bright full-bleed colour field, coins, banknotes, percentage signs",
        "reference_asset_keys": []
      }
    },
    "reference_asset_keys": []
  }$ig$::jsonb,
  'seed',
  'Phase C pilot seed 2026-08-17. Palette pinned here deliberately: generic_theme misfires fleet-wide (webdesign colour-churn landmine), and design_intent.palette.reference_values is the durable pin.',
  true,
  false,
  'portfolio_positioning Phase C pilot seed'
FROM sites WHERE domain = 'remortgagecalculator.uk';

-- ------------------------------------------------------------------ verify --
DO $do$
DECLARE
    n int;
    sid uuid;
BEGIN
    SELECT id INTO sid FROM sites WHERE domain = 'remortgagecalculator.uk';
    IF sid IS NULL THEN
        RAISE EXCEPTION 'seed verify: no sites row for remortgagecalculator.uk';
    END IF;

    SELECT count(*) INTO n FROM sites
    WHERE domain = 'remortgagecalculator.uk' AND email IS NOT NULL AND email <> '';
    IF n <> 1 THEN
        RAISE EXCEPTION 'seed verify: site row has no email - bugs_open/063 fails OPEN without one';
    END IF;

    SELECT count(*) INTO n FROM site_specs
    WHERE site_id = sid AND aspect = 'evidence_base' AND is_current;
    IF n <> 1 THEN
        RAISE EXCEPTION 'seed verify: expected exactly 1 current evidence_base, found %', n;
    END IF;

    SELECT count(*) INTO n FROM site_specs
    WHERE site_id = sid AND aspect = 'imagery_style_guide' AND is_current;
    IF n <> 1 THEN
        RAISE EXCEPTION 'seed verify: expected exactly 1 current imagery_style_guide, found %', n;
    END IF;

    -- The banned_claims array must survive the jsonb round-trip with its regex
    -- escaping intact: a pattern mangled here fails OPEN (it simply never
    -- matches), which is the silent direction.
    SELECT jsonb_array_length(data->'banned_claims') INTO n FROM site_specs
    WHERE site_id = sid AND aspect = 'evidence_base' AND is_current;
    IF n <> 6 THEN
        RAISE EXCEPTION 'seed verify: expected 6 banned_claims, found % - check the escaping', n;
    END IF;
END $do$;

COMMIT;

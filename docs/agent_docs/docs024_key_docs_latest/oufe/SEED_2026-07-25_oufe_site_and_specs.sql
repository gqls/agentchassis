-- ============================================================================
-- oufe.com — site row + evidence_base + imagery_style_guide
-- Written 2026-07-25. Applied out of band (psql -f), NOT via the migration
-- runner: this is per-site setup, not a platform schema change.
--
-- WHY THESE THREE, AND WHY BEFORE SUBMISSION
--   1. sites row with an EMAIL. bugs_open/063: the hallucinated-email check
--      FAILS OPEN when a site has no contact email, and a fabricated address
--      reached production for hours on another site that way. ensure_site_record
--      upserts on domain, so pre-creating is safe and wins the race.
--   2. evidence_base. The entire claims layer is gated on the PRESENCE of this
--      aspect — loadEvidenceBase returns nil and every lane silently no-ops
--      (validate_page_content.go:727-746). Seeding it BEFORE the first page is
--      written is the only way the first page is covered.
--   3. imagery_style_guide. bugs_closed/027: content_hero generates unstyled on
--      any site that has none, and a brand-new site has none.
--
-- THE LIMITATION THAT SHAPED THE BANNED LIST — read before trusting the gate
--   ScanUnregisteredNumbers is effectively INERT on this site. It scans a number
--   only when businessClaimContextRe (datahelpers/claims.go:334) matches its
--   surrounding window, and that list is clients|customers|records|awards|
--   uptime|years of experience and similar. It contains NO debt, creditor,
--   tranche, recovery, covenant or leverage vocabulary. isExcludedNumber then
--   excludes currency amounts outright. So "£16bn of Class A debt" is never
--   scanned at all — on a site made almost entirely of pound figures.
--
--   CONSEQUENCE: banned_claims and writer_block do the load-bearing work here,
--   which is why the patterns below target SHAPES OF FABRICATION rather than
--   individual numbers. Do not read a clean claims report on this site as "no
--   invented numbers" — it means "no banned pattern matched".
--
-- ON THE FIVE LITERAL NUMBERS BANNED BELOW
--   They come from the Gemini strategy conversation (PLAN §C2) and are model
--   recollections, not sourced facts. Banning them is deliberately fail-closed:
--   if one later turns out to be the real figure, the ban forces a conscious
--   return to this file to remove it, having first registered the fact with a
--   citation. That friction is the point.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------- site row --
INSERT INTO sites (domain, name, network_id, status, email, company_name)
VALUES (
  'oufe.com',
  'oufe.com',
  '00000000-0000-0000-0000-000000000002',
  'active',
  'oufe@contactforsales.com',
  'OUFE'
)
ON CONFLICT (domain) DO UPDATE
  SET email = COALESCE(sites.email, EXCLUDED.email);
-- NOTE: status='active' is what upsertSite writes, but it is NOT in the
-- validated vocabulary (draft/building/review/published/deployed/archived/
-- error). Never scope a query by it and expect meaning.

-- ----------------------------------------------------------- evidence_base --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'oufe.com')
  AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'evidence_base',
  $eb${
    "governing_rule": "Every figure, date, quantum, percentage and quoted phrase about a real company, court, regulator or transaction must trace to a fact below carrying a source URL and a capture date. Where no verified fact exists, the page says so plainly — 'we have not verified this' is always publishable and a plausible estimate never is. This site is educational analysis of financial and legal mechanism: it is not investment advice, not a recommendation, and not a financial promotion.",
    "audit_doc": "docs/agent_docs/docs024_key_docs_latest/oufe/ (PLAN 2026-07-25 §2 challenges; §5 the two precedents)",
    "schema_notes": "facts[]: {id, claim, kind: metric|capability|entity|attestation, source: EXACTLY ONE of {sql|artifact|attested_by|citation}, verified_at, value?, tolerance?, context_terms?, writer_line?}. banned_claims[]: {pattern (case-insensitive regex; an invalid regex degrades to a literal substring, so a typo never silently drops a ban), reason}. allowed_entities[]: real named entities it is legitimate to NAME — naming is not asserting, and every claim ABOUT them still needs a fact.",
    "facts": [],
    "banned_claims": [
      {"pattern": "£ ?16 ?(bn|billion)", "reason": "Gemini strategy transcript, 2026-07-25: stated as Thames Water Class A debt. A model recollection, never sourced. Unban only after registering the real figure from a judgment or regulator document."},
      {"pattern": "£ ?3 ?(bn|billion)", "reason": "Gemini transcript: stated as the super-senior liquidity facility. Unsourced recollection."},
      {"pattern": "£ ?3\\.35 ?(bn|billion)", "reason": "Gemini transcript: stated as a consortium equity injection. Unsourced recollection."},
      {"pattern": "£ ?9\\.4 ?(bn|billion)", "reason": "Gemini transcript: stated as debt written off. Unsourced recollection."},
      {"pattern": "£ ?2\\.5 ?(bn|billion)", "reason": "Gemini transcript: stated as subordinated equity. Unsourced recollection."},
      {"pattern": "[0-9]{1,3}(\\.[0-9]+)? ?% ?(recovery|recovered|of par)", "reason": "recovery-rate class: a recovery percentage is an OUTPUT of a modelled scenario, never a fact about a live case. It may appear inside the simulator as the user's own scenario; it may not appear in prose as a statement about a real company."},
      {"pattern": "(cents|pence) on the (dollar|pound)", "reason": "trading-level class: no market-data feed exists anywhere in this platform, so no trading level is knowable to us."},
      {"pattern": "(will|expected to) (recover|be wiped out|be primed|default)", "reason": "prediction stated as fact. Describe what the documents permit, not what will happen."},
      {"pattern": "(guaranteed|risk-free|assured) (return|profit|recovery|upside)", "reason": "financial-promotion language (s21 FSMA perimeter) and untrue on its face."},
      {"pattern": "(should|must) (buy|sell|short|invest|avoid)", "reason": "investment recommendation. This site analyses mechanism and never advises."},
      {"pattern": "price target", "reason": "investment-research artefact we do not produce."},
      {"pattern": "(sources|people) (close to|familiar with) the (matter|situation|deal|negotiations)", "reason": "fabricated sourcing: we have no sources. Every assertion here comes from a public document we can link and date."},
      {"pattern": "we (understand|are told|hear) that", "reason": "same class — implies private information we do not have."},
      {"pattern": "(our|the) (proprietary|exclusive) (data|database|research)", "reason": "no proprietary dataset exists; and per the vetcomparison rails we assert no data rights over third-party facts."},
      {"pattern": "[0-9][0-9,]* ?(subscribers|readers|clients|firms|professionals|users)", "reason": "audience-scale class: this site is new and has no audience telemetry."},
      {"pattern": "(trusted|used|read) by [0-9]", "reason": "social-proof class: unsupportable."},
      {"pattern": "years of (experience|expertise)", "reason": "this is a new publication with no history; no tenure claim is available."},
      {"pattern": "we (track|monitor|cover) [0-9]", "reason": "coverage-scale class: implies an automated estimate we do not run. There is no radar, no docket feed and no market data (PLAN §C1)."}
    ],
    "allowed_entities": [
      "Thames Water", "Ofwat", "Ofgem", "the High Court", "the Court of Appeal",
      "Companies House", "the Insolvency Service", "the Financial Conduct Authority",
      "Part 26A", "the Companies Act 2006", "a restructuring plan", "a scheme of arrangement",
      "cross-class cramdown", "special administration", "the special administration regime",
      "an uptier exchange", "a dropdown transaction", "a double-dip structure",
      "liability management", "debtor-in-possession financing", "a debt-for-equity swap",
      "Elliott Investment Management", "Apollo Global Management", "Oaktree Capital Management",
      "Pershing Square", "Third Point", "OUFE", "Oxen Unity"
    ],
    "writer_block": "NONE — NO FIGURE ON THIS SITE HAS YET BEEN VERIFIED.\n\nThere are currently no verified facts, so there are no numbers you may assert. Do not state any monetary amount, percentage, recovery rate, debt quantum, ownership stake, court date, filing date, or count of anything — not about Thames Water, not about any company, regulator, court or fund, and not about this publication itself.\n\nIf a sentence seems to need a number to work, rewrite the sentence so that it does not. Explaining a mechanism never requires a figure: you can describe exactly how a super-senior tranche ranks ahead of existing lenders, or how a majority can bind a dissenting minority, without asserting a single quantum. That is what this site is for.\n\nWhere the substance genuinely is a figure, say plainly that it is not yet verified and name the kind of document it will come from. 'We have not yet verified the current size of each debt class from the filed documents' is publishable. An estimate, a rounded recollection, or a figure carried over from another article is not.\n\nNaming a real company, court or regulator is allowed. Making a factual claim about one is not, unless that claim appears above — and nothing appears above yet."
  }$eb$::jsonb,
  'manual',
  'Seeded at site creation, before any page was written, so the first page is covered. facts[] deliberately empty until evidence-researcher (V5) registers verified citations.',
  true, true, 'oufe-workstream-2026-07-25'
FROM sites WHERE domain = 'oufe.com';

-- ------------------------------------------------------ imagery_style_guide --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'oufe.com')
  AND aspect = 'imagery_style_guide' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'imagery_style_guide',
  $img${
    "medium": "restrained editorial illustration — abstract geometric composition, architectural line work, and structural forms suggesting hierarchy, layering and precedence",
    "mood": "institutional, sober, precise, considered; the register of a serious research house rather than a trading floor",
    "palette": "deep navy and near-black grounds (#0f172a, #0d1220), slate and cool grey mid-tones (#475569, #94a3b8), muted antique gold used sparingly as the single accent (#8a6d3b, #c9a86a), warm off-white paper (#f7f8fa)",
    "avoid": "stock photography of business people, handshakes, boardrooms or people pointing at screens; trading screens, candlestick charts, tickers; green-and-red market iconography; bull and bear imagery; money, coins, banknotes and piggy banks; photorealistic depictions of identifiable individuals; any drawn chart, graph, axis, bar or plotted data (charts are code-rendered from verified figures, never illustrated); text, lettering, numerals, logos or watermarks of any kind",
    "kinds": {
      "content_hero": {
        "medium": "flat duotone editorial illustration",
        "mood": "one bold abstract structural motif — stacked strata, a stepped hierarchy, an interlocking joint — minimal detail, strong silhouette",
        "palette": "deep navy ground, muted gold flat shapes and linework, cool grey secondary only",
        "avoid": "photorealism, photographic texture, gradients, 3D rendering, drop shadows, text, lettering, numerals, logos, watermarks, busy detail, colour outside the palette, white or pale or light backgrounds, bright full-bleed colour fields",
        "reference_asset_keys": []
      }
    },
    "reference_asset_keys": []
  }$img$::jsonb,
  'manual',
  'Seeded pre-build: bugs_closed/027 — content_hero generates unstyled on a site with no style guide, and a fresh site has none. The avoid-list bans drawn charts explicitly (charts are code-rendered from verified series).',
  true, true, 'oufe-workstream-2026-07-25'
FROM sites WHERE domain = 'oufe.com';

COMMIT;

-- Verify
--   SELECT domain, email, status FROM sites WHERE domain='oufe.com';
--   SELECT aspect, is_current, pinned, created_by FROM site_specs ss
--     JOIN sites s ON s.id=ss.site_id WHERE s.domain='oufe.com' ORDER BY aspect;
--   SELECT jsonb_array_length(data->'banned_claims') AS bans,
--          jsonb_array_length(data->'facts') AS facts
--     FROM site_specs ss JOIN sites s ON s.id=ss.site_id
--    WHERE s.domain='oufe.com' AND ss.aspect='evidence_base' AND ss.is_current;

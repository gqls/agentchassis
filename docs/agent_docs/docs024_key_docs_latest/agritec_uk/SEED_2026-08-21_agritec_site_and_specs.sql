-- ============================================================================
-- agritec.uk — site row + evidence_base + imagery_style_guide
-- Written 2026-08-21. Applied out of band (psql -f), NOT via the migration
-- runner: this is per-site setup, not a platform schema change.
-- Modelled on ../oufe/SEED_2026-07-25_oufe_site_and_specs.sql.
--
-- WHY THESE THREE, AND WHY BEFORE SUBMISSION
--   1. sites row with an EMAIL. The hallucinated-email check FAILS OPEN when a
--      site has no contact email, and a fabricated address reached production
--      for hours on another site that way. ensure_site_record upserts on
--      domain, so pre-creating is safe and wins the race with the classifier.
--   2. evidence_base. The entire claims layer is gated on the PRESENCE of this
--      aspect — loadEvidenceBase returns nil and every lane silently no-ops
--      (validate_page_content.go:727-746). Seeding it BEFORE the first page is
--      written is the only way the first page is covered.
--   3. imagery_style_guide. bugs_closed/027: content_hero generates unstyled on
--      any site that has none, and a brand-new site has none. This site's whole
--      point is MORE imagery, so an unstyled hero would be a visible failure.
--
-- THE LIMITATION THAT SHAPED THE BANNED LIST — read before trusting the gate
--   ScanUnregisteredNumbers is effectively INERT on this site's subject matter.
--   MEASURED 2026-08-21 by reading the code, not inferred from the oufe note:
--
--   (a) businessClaimContextRe (datahelpers/claims.go:660) is a LEXICAL gate: a
--       number is scanned only when its surrounding window matches
--       clients|customers|records|businesses|companies|agents|sites|users|
--       subscribers|departments|awards|employees|staff|engagements|projects|
--       deployments|case studies|definitions|orchestration|integrations|
--       providers|items|uptime|verified|enriched|scored|collected|processed|
--       deployed|delivered|years of experience|uniques.
--       There is NO agricultural or physical-units vocabulary in it whatsoever —
--       no hectare, yield, crop, tonne, payment, rate, efficacy, DLI, PPFD, EC,
--       kPa, kWh, larvae or carbon. So "2.6 umol/J" and "129 per hectare" are
--       never scanned at all, on a site made almost entirely of such figures.
--   (b) isExcludedNumber (claims.go:849) structurally excludes any number
--       directly preceded by a currency symbol — the multibyte check at :873
--       catches the pound sign explicitly. Every SFI payment rate on this site
--       is currency-prefixed, so every one of them is excluded before (a) is
--       even reached.
--   (c) It also excludes ranges joined by an en/em dash ("12-17 DLI"), which is
--       the shape most of this site's agronomic guidance takes.
--
--   CONSEQUENCE: banned_claims and writer_block do the load-bearing work here.
--   The patterns below therefore target SHAPES OF FABRICATION, not individual
--   numbers. Do NOT read a clean claims report on this site as "no invented
--   numbers" — it means "no banned pattern matched", which is a much weaker
--   statement, and on this subject matter it is close to vacuous.
--
-- THE SECOND GAP, WHICH MATTERS MORE HERE THAN ON ANY SITE SO FAR
--   bugs_open/288: the evidence register guards COPY, not CODE. A figure a
--   calculator ENCODES is checked by nothing — an SDLT calculator ran an
--   expired legislated threshold for 16 months with a clean scanner throughout.
--   This site is SIX CALCULATORS whose SFI rates, LED efficacies and carbon
--   fractions all live inside JavaScript. Neither the scanner nor this file can
--   see them.
--   THE CONTROL (PLAN section 4, Phase 2): every constant a calculator encodes
--   must ALSO be a registered fact here AND be asserted in that tool page's
--   visible copy, where extractAssertions can reach it. A constant that appears
--   only in JS is ungoverned by design, not by oversight.
--
-- WHAT IS DELIBERATELY *NOT* SEEDED HERE
--   design_intent. The fresh-submission path exists so the classifier does its
--   own design research, and pre-pinning a palette would pre-empt exactly the
--   richer visual direction the owner asked for. AFTER classification, check
--   design_intent.palette.reference_values against the imagery guide below and
--   re-align whichever is wrong — the generic_theme colour-churn landmine
--   misfires fleet-wide and pinning reference_values is the remedy, but it is a
--   remedy applied to a measured problem, not a prophylactic.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------- site row --
INSERT INTO sites (domain, name, network_id, status, email, company_name)
VALUES (
  'agritec.uk',
  'agritec.uk',
  '00000000-0000-0000-0000-000000000002',
  'active',
  'agritec@contactforsales.com',
  'agritec.uk'
)
ON CONFLICT (domain) DO UPDATE
  SET email = COALESCE(sites.email, EXCLUDED.email);
-- NOTE: status='active' is what upsertSite writes, but it is NOT in the
-- validated vocabulary (draft/building/review/published/deployed/archived/
-- error). Never scope a query by it and expect meaning.
-- NOTE: network_id verified 2026-08-21 — 'Default Network' is the only network
-- and carries all 25 real sites.

-- ----------------------------------------------------------- evidence_base --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'agritec.uk')
  AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'evidence_base',
  $eb${
    "governing_rule": "Every figure, rate, percentage, threshold and physical constant on this site must trace to a fact below carrying a source URL and a capture date. Where no verified fact exists, the page says so plainly - 'we have not verified this' is always publishable and a plausible estimate never is. This site is educational engineering and agronomic analysis. It is not agronomic advice, not a subsidy eligibility determination, and not a recommendation to plant, build, spend or apply anything.",
    "audit_doc": "docs/agent_docs/docs024_key_docs_latest/agritec_uk/ (PLAN_2026-08-21 section 1 decisions D4 and D8; SUBJECT_LEDGER.md for the per-subject depth and completeness contract)",
    "schema_notes": "facts[]: {id, claim, kind: metric|capability|entity|attestation, source: EXACTLY ONE of {sql|artifact|attested_by|citation}, verified_at, value?, tolerance?, context_terms?, writer_line?}. banned_claims[]: {pattern (case-insensitive regex; an invalid regex degrades to a literal substring, so a typo never silently drops a ban), reason}. allowed_entities[]: real named entities it is legitimate to NAME - naming is not asserting, and every claim ABOUT them still needs a fact.",
    "facts": [],
    "banned_claims": [
      {"pattern": "cannabis", "reason": "OWNER RULING 2026-08-21: no cannabis content on this site. Cultivation is licensed-only in the UK. Present on the retired hand-built site in a DLI help-text hint and in two rows of the publicly-served /data/crop-dli-table.json. This is a BANNED PATTERN rather than a line in a brief because the writers know perfectly well that cannabis is a controlled-environment crop and could reintroduce it from general knowledge without ever seeing the old page."},
      {"pattern": "the authority in uk agritech", "reason": "the retired site's own headline. An unsupportable self-claim about market position on a site with no audience telemetry."},
      {"pattern": "(uk )?(feed )?wheat[^.]{0,40}(GBP|pounds?)? ?[0-9]", "reason": "fabricated-ticker class. The retired homepage carried a wheat price typed straight into the HTML, and its feeder (data-collector/v1/cmd/updater/main.go) generated values with rand() while labelling its own output 'Simulated Exchange / National Grid'. No market-data feed exists anywhere in this platform, so no commodity price is knowable to us."},
      {"pattern": "(brent|crude|carbon|uk ets|day-ahead|ammonium nitrate)[^.]{0,30}[0-9]", "reason": "same fabricated-ticker class - the other four instruments from the retired ticker strip."},
      {"pattern": "(current|today|live|latest) (price|rate|cost) (is|of) ", "reason": "live-price class: nothing on this site is live. Every figure is a dated capture from a named source, and the page must say when it was captured."},
      {"pattern": "(will|guaranteed to|expected to) (increase|improve|boost|raise) (your )?(yield|output|profit|margin|revenue)", "reason": "yield-promise class. An outcome claim about the reader's own operation that no source can support and that we are in no position to make."},
      {"pattern": "(payback|roi|return on investment) (in|of|within) [0-9]", "reason": "investment-return class: a payback period is an OUTPUT of a scenario the reader parameterises, never a fact we assert about their operation."},
      {"pattern": "you (will|would) receive [^.]{0,20}[0-9]", "reason": "subsidy-entitlement class. Whether a holding qualifies for an SFI action depends on land eligibility, existing agreements and scheme status. A calculator models a scenario; it does not determine entitlement, and stating one could cost a farmer real money if wrong."},
      {"pattern": "(you should|you must|we recommend) (apply|plant|sow|spray|dose|install|switch to)", "reason": "agronomic-instruction class. This site explains mechanism so the reader can decide. It does not instruct an operation it cannot see."},
      {"pattern": "(eligible|qualifies) for [^.]{0,30}(SFI|ELMS|CS|Countryside Stewardship)", "reason": "same entitlement class stated as fact about the reader."},
      {"pattern": "(optimal|ideal|correct) (DLI|EC|VPD|ph|temperature) (is|of) [0-9]", "reason": "single-value agronomic prescription. These are ranges that depend on cultivar, growth stage and system, and the retired site's own data file carried them as unsourced round numbers. State a sourced range with its source, or make it a user input."},
      {"pattern": "(our|the) (proprietary|exclusive) (data|dataset|database|research|model)", "reason": "no proprietary dataset exists, and per the vetcomparison rails we assert no data rights over third-party facts."},
      {"pattern": "(sources|people) (close to|familiar with) the (matter|situation|industry)", "reason": "fabricated sourcing: we have no sources. Every assertion comes from a public document we can link and date."},
      {"pattern": "we (understand|are told|hear) that", "reason": "same class - implies private information we do not have."},
      {"pattern": "[0-9][0-9,]* ?(farms|growers|operators|farmers|customers|clients|users|subscribers)", "reason": "audience-scale class: this site is new and has no audience telemetry."},
      {"pattern": "(trusted|used|relied on) by [0-9]", "reason": "social-proof class: unsupportable."},
      {"pattern": "years of (experience|expertise)", "reason": "no tenure claim is available to a new publication."},
      {"pattern": "(certified|accredited|approved) by", "reason": "accreditation class: this site holds no certification or approval from anybody."},
      {"pattern": "(defra|rpa|ofgem|environment agency) (says|states|confirms|has confirmed) ", "reason": "attributed-statement class. Quoting a regulator requires the document and the quote, verified by the citation machinery - not a paraphrase in the regulator's voice."}
    ],
    "allowed_entities": [
      "DEFRA", "the Department for Environment, Food and Rural Affairs",
      "the Rural Payments Agency", "Natural England", "the Environment Agency",
      "Ofgem", "the National Energy System Operator", "the Sustainable Farming Incentive",
      "Environmental Land Management", "Countryside Stewardship", "the Basic Payment Scheme",
      "the UK Emissions Trading Scheme", "the Agriculture and Horticulture Development Board",
      "the DesignLights Consortium", "GOV.UK", "agritec.uk",
      "controlled environment agriculture", "vertical farming", "hydroponics",
      "black soldier fly", "Hermetia illucens", "Saccharina latissima", "Alaria esculenta",
      "Laminaria digitata", "photosynthetically active radiation", "daily light integral",
      "vapour pressure deficit", "electrical conductivity"
    ],
    "writer_block": "NONE - NO FIGURE ON THIS SITE HAS YET BEEN VERIFIED.\n\nThere are currently no verified facts, so there are no numbers you may assert. Do not state any payment rate, price, efficacy, light level, conductivity, conversion ratio, carbon fraction, yield, area, temperature threshold or count of anything - not about a scheme, a regulator, a crop, a species, a technology, or this publication itself.\n\nIf a sentence seems to need a number to work, rewrite the sentence so that it does not. Explaining a mechanism never requires a figure: you can explain exactly why photometric units misdescribe what a plant receives, why a stock tank must be split to stop calcium and sulphate precipitating, or why metabolic heat sets the density ceiling in insect rearing, without asserting a single quantity. That is what this site is for, and it is why the explainers can be long and detailed while asserting nothing unverified.\n\nWhere the substance genuinely IS a figure, say plainly that it is not yet verified and name the kind of document it will come from. 'We have not yet confirmed the current payment rate for this action against the published handbook' is publishable. An estimate, a rounded recollection, or a figure carried over from the retired version of this site is not.\n\nONE DISTINCTION, because this site is mostly calculators. A calculator's INPUT DEFAULT is not an assertion - it is a starting point the reader immediately changes, and it must be presented as one. What IS an assertion is any constant the calculator applies to the reader's input on our authority: a payment rate, a fixture efficacy, a conversion factor. Those are claims, they need registered facts, and they must also appear in the visible copy of the tool page - because the claims scanner cannot see inside JavaScript, so a constant hidden in code is checked by nothing at all.\n\nNaming a real scheme, regulator, species or technology is allowed. Making a factual claim about one is not, unless that claim appears above - and nothing appears above yet."
  }$eb$::jsonb,
  'manual',
  'Seeded at site creation, before any page was written, so the first page is covered. facts[] deliberately empty until evidence-researcher registers verified citations across the six data domains (PLAN Phase 2). Owner ruling D4: source everything before any page is written.',
  true, true, 'agritec-workstream-2026-08-21'
FROM sites WHERE domain = 'agritec.uk';

-- ------------------------------------------------------ imagery_style_guide --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'agritec.uk')
  AND aspect = 'imagery_style_guide' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'imagery_style_guide',
  $img${
    "medium": "precise technical illustration - cross-sections, cutaways, isometric system diagrams and instrument-panel line work, in the register of an engineering manual or a field handbook rather than a brochure",
    "mood": "practical, exact, unromantic; built for an operator who wants to understand a mechanism, not for a brochure cover. Clean flat shapes, confident linework, generous space, no atmosphere for its own sake",
    "palette": "deep slate grounds (#1a202c, #0f172a), agricultural green as the primary accent (#2c7744), instrument teal as the secondary (#005f73), cool greys for structure (#475569, #94a3b8), warm off-white paper (#f0f4f8)",
    "avoid": "stock photography of any kind, and especially smiling farmers, handshakes over gates, golden-hour wheat fields, drone shots of tractors, and hands cupping soil or seedlings; lens flare, bokeh and sun glare; photorealistic depictions of identifiable individuals; cartoon or mascot styling; any drawn chart, graph, axis, bar, gauge or plotted data - charts are code-rendered from verified figures and never illustrated; text, lettering, numerals, units, logos or watermarks of any kind",
    "kinds": {
      "content_hero": {
        "medium": "flat editorial illustration of one clear technical subject - a lighting rig over a canopy, a stacked growing tier, a nutrient loop, a mooring line, a rearing stack",
        "mood": "one bold legible motif, strong silhouette, minimal detail, no scene-setting",
        "palette": "deep slate ground, agricultural green and instrument teal flat shapes, cool grey secondary only",
        "avoid": "photorealism, photographic texture, gradients, 3D rendering, drop shadows, text, lettering, numerals, logos, watermarks, busy detail, colour outside the palette, pale or white or bright full-bleed backgrounds",
        "reference_asset_keys": []
      },
      "infographic": {
        "medium": "flat diagrammatic illustration explaining one mechanism - a labelled cross-section, an energy or mass flow, a stacked comparison of two paths",
        "mood": "legible at a glance and rewarding on a second look; the picture carries the explanation, not decoration alongside it",
        "palette": "off-white paper ground with slate linework, green and teal reserved for the elements that carry meaning",
        "avoid": "plotted data of any kind - if it has an axis it is a chart and must be code-rendered from registered facts, not generated; also no text, numerals or units, which the generator renders unreliably and which the claims scanner cannot read inside an image",
        "reference_asset_keys": []
      }
    },
    "reference_asset_keys": []
  }$img$::jsonb,
  'manual',
  'Seeded pre-build: bugs_closed/027 - content_hero generates unstyled on a site with no style guide, and a fresh site has none. Palette carries the retired site brand (ag-green #2c7744, tech-blue #005f73, slate #1a202c) deliberately: the site is being replaced, the brand is not. AFTER classification, reconcile against design_intent.palette.reference_values - whichever is wrong gets corrected, and that is also the generic_theme colour-churn remedy. The avoid-lists ban drawn charts explicitly on both kinds, and ban text/numerals inside infographics because the claims scanner cannot read them.',
  true, true, 'agritec-workstream-2026-08-21'
FROM sites WHERE domain = 'agritec.uk';

COMMIT;

-- Verify (expect: 1 site row with an email; exactly 2 current pinned specs; 0 facts)
--   SELECT domain, email, status, network_id FROM sites WHERE domain='agritec.uk';
--   SELECT aspect, is_current, pinned, created_by FROM site_specs ss
--     JOIN sites s ON s.id=ss.site_id WHERE s.domain='agritec.uk' ORDER BY aspect;
--   SELECT jsonb_array_length(data->'banned_claims') AS bans,
--          jsonb_array_length(data->'facts')         AS facts,
--          jsonb_array_length(data->'allowed_entities') AS entities
--     FROM site_specs ss JOIN sites s ON s.id=ss.site_id
--    WHERE s.domain='agritec.uk' AND ss.aspect='evidence_base' AND ss.is_current;
--
-- Then prove the cannabis ban actually FIRES rather than assuming it:
--   register nothing, run the banned-claim scan over a string containing the
--   word, and confirm a flag. A ban that never fired is a ban you have not
--   tested - the whole point of D8 is that the writers can reintroduce this
--   from general knowledge, so the pattern has to work, not merely exist.

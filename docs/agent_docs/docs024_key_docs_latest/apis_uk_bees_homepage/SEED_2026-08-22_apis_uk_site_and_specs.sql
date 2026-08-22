-- ============================================================================
-- apis.uk — site row + evidence_base + imagery_style_guide + roadmap_brief
-- Written 2026-08-22. Applied out of band (psql -f), NOT via the migration
-- runner: this is per-site setup, not a platform schema change.
--
-- WHY THESE FOUR, AND WHY BEFORE SUBMISSION
--   1. sites row with an EMAIL. bugs_open/063: the hallucinated-email check
--      FAILS OPEN when a site has no contact email. ensure_site_record upserts
--      on domain, so pre-creating is safe and wins the race with the classifier.
--   2. evidence_base. GROUNDED IN CODE, not in a runbook: loadEvidenceBase
--      (validate_page_content.go:1272-1290) returns nil on sql.ErrNoRows and
--      every claims lane then silently no-ops. A site with no evidence base is
--      not "unchecked but fine" — it is unchecked AND reports clean. Seeding it
--      before the first page is written is the only way page one is covered.
--   3. imagery_style_guide. bugs_closed/027: content_hero generates unstyled on
--      a site that has none, and a brand-new site has none.
--   4. roadmap_brief. This is what makes "home page only" REAL. Grounded in the
--      live build-site-planner definition (clients_db.agent_definitions — NOTE
--      the oufe runbook says agent_definitions lives in templates_db; it does
--      NOT, it is in clients_db, 216 rows vs 8). Its prompt reads
--      {{.site_specs.specs.roadmap_brief.text}} and says: "ROADMAP OVERRIDES THE
--      COMPONENT LIST. Build ONLY the pages listed in the current phase below.
--      ... Do NOT invent additional pages. The roadmap is the authority."
--
-- WHY THE BAN LIST LOOKS LIKE THIS
--   Bees are a subject made almost entirely of famous repeated numbers: the
--   share of food owed to pollinators, flowers visited per jar, miles flown,
--   bees per hive, percentage declines, species counts. Every one is quotable
--   everywhere and sourced nowhere WE have checked. So facts[] is deliberately
--   EMPTY and the patterns target SHAPES OF FABRICATION, not individual figures
--   (the oufe precedent). There is also one specific famous misattribution —
--   the "four years left to live" line Einstein never said — banned by shape.
--
--   Naming a species, a phenomenon or an organisation is ALLOWED (see
--   allowed_entities): naming is not asserting. Making a factual claim about
--   one still needs a fact, and nothing appears in facts[] yet.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ---------------------------------------------------------------- site row --
INSERT INTO sites (domain, name, network_id, status, email, company_name)
VALUES (
  'apis.uk',
  'apis.uk',
  '00000000-0000-0000-0000-000000000002',
  'active',
  'apis-uk@contactforsales.com',
  'apis.uk'
)
ON CONFLICT (domain) DO UPDATE
  SET email = COALESCE(sites.email, EXCLUDED.email);
-- NOTE: status='active' is what upsertSite writes, but it is NOT in the
-- validated vocabulary (draft/building/review/published/deployed/archived/
-- error). Never scope a query by it and expect it to mean anything.

-- ----------------------------------------------------------- evidence_base --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'apis.uk')
  AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'evidence_base',
  $eb${
    "governing_rule": "Every figure, quantity, percentage, distance, count, date and quoted phrase on this site must trace to a fact below carrying a source and a capture date. Where no verified fact exists, the page simply does not make the claim — rewriting the sentence so it needs no number is ALWAYS available and is always preferable to an estimate. This is one person's enthusiast page about bees: it is not a scientific reference, not advice on keeping bees, and not a conservation campaign. It should be accurate, warm and plainly written, and it should never pretend to authority it does not have.",
    "audit_doc": "docs/agent_docs/docs024_key_docs_latest/apis_uk_bees_homepage/ (PLAN 2026-08-22 §5)",
    "schema_notes": "facts[]: {id, claim, kind: metric|capability|entity|attestation, source: EXACTLY ONE of {sql|artifact|attested_by|citation}, verified_at, value?, tolerance?, context_terms?, writer_line?}. banned_claims[]: {pattern (case-insensitive regex; an invalid regex degrades to a literal substring, so a typo never silently drops a ban), reason}. allowed_entities[]: real named species, phenomena and bodies it is legitimate to NAME — naming is not asserting, and every claim ABOUT them still needs a fact.",
    "facts": [],
    "banned_claims": [
      {"pattern": "(one|1) in (three|3) (bites|mouthfuls|forkfuls)", "reason": "pollinator-dependency class: the single most repeated bee statistic in existence and the least often sourced. Unban only by registering a figure from a named agricultural or ecological authority with a capture date."},
      {"pattern": "[0-9]{1,3}(\\.[0-9]+)? ?% of (the )?(world's |global |our |uk )?(food|crops|diet|produce|harvest)", "reason": "pollinator-dependency class, percentage form. Same reason as above."},
      {"pattern": "[0-9][0-9,]* ?(flowers|blossoms|blooms)", "reason": "flowers-visited class ('two million flowers to a jar of honey'). A famous round number with no source we hold."},
      {"pattern": "[0-9][0-9,.]* ?(miles|km|kilometres|kilometers)", "reason": "distance-flown class. Describe that foragers range widely and navigate back precisely — that needs no figure."},
      {"pattern": "[0-9][0-9,]* ?(bees|workers|drones|eggs|larvae|cells)", "reason": "colony-size class ('sixty thousand bees in a hive'). Varies enormously by season and colony and we have verified nothing."},
      {"pattern": "[0-9]{1,3}(\\.[0-9]+)? ?% ?(decline|drop|fall|loss|lost|decrease|fewer|down)", "reason": "population-decline class. Real, serious, and not ours to quantify without a cited survey."},
      {"pattern": "[0-9][0-9,]* ?(species|varieties|kinds|types) of (bee|bees|bumblebee|bumblebees|pollinator)", "reason": "species-count class ('270 species of bee in the UK'). Commonly cited, not verified here."},
      {"pattern": "[0-9][0-9,]* ?(times per second|beats per second|wingbeats|wing beats|flaps)", "reason": "wingbeat class. A measurement we have not taken and not sourced."},
      {"pattern": "[0-9][0-9,.]* ?(kg|kilograms|kilos|pounds|lbs|jars|tonnes|tons|grams) of (honey|nectar|pollen|wax|beeswax)", "reason": "yield class. Depends on colony, forage and season; any single figure is a fabrication dressed as a fact."},
      {"pattern": "lives? (for )?(up to )?[0-9]", "reason": "lifespan class. Worker, drone and queen lifespans differ by orders of magnitude and by season."},
      {"pattern": "[0-9][0-9,.]* ?(hundred|thousand|million|billion)", "reason": "magnitude class, digit form ('2 million flowers'). The digit-adjacent patterns above do not see a magnitude WORD between the number and the noun, which is exactly the form the most-repeated bee statistics take."},
      {"pattern": "(a|one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve|twenty|thirty|forty|fifty|sixty|seventy|eighty|ninety|several|many|tens of|hundreds of|thousands of|millions of) (hundred|thousand|million|billion)", "reason": "magnitude class, spelled-out form ('two million flowers', 'sixty thousand bees', 'many thousands'). writer_block names this explicitly: 'many thousands of bees' still asserts a magnitude and is not an improvement on 'crowded'."},
      {"pattern": "(hundreds|thousands|millions|billions) of (bees|flowers|blossoms|blooms|miles|hives|colonies|species|eggs|cells|visits|trips|plants|crops)", "reason": "magnitude class, bare-plural form. Same reason: a vague large number is still a quantity claim we have not verified."},
      {"pattern": "(live|lives|lasts|lasts for|survives|survive) (for )?(up to )?(a|one|two|three|four|five|six|seven|eight|nine|ten|twelve|twenty|thirty|forty|fifty|sixty) (hour|hours|day|days|week|weeks|month|months|year|years)", "reason": "lifespan class, spelled-out form. The digit form above misses 'a worker lives six weeks', which is the way this claim is usually written."},
      {"pattern": "[0-9]{1,3} ?(°|degrees|deg\\b)", "reason": "hive-temperature class ('they hold the brood nest at 35°C'). A precise number we have not verified."},
      {"pattern": "(four|4) years left to live", "reason": "the Einstein misattribution. He did not say it, it is not true as stated, and it is the single most common false thing written about bees. Banned outright in every form."},
      {"pattern": "einstein (said|wrote|warned|predicted|once)", "reason": "same misattribution, attribution form. If the page discusses the quote at all it must do so as a debunk, and that requires a registered fact first."},
      {"pattern": "(studies|research|scientists|researchers|experts) (show|shows|have shown|say|says|suggest|suggests|found|agree)", "reason": "fabricated-sourcing class: we have read no studies and cite none. Name a specific paper with a link, or do not invoke research at all."},
      {"pattern": "according to (a|the|recent) (study|report|survey|census)", "reason": "same class — a citation shape with no citation behind it."},
      {"pattern": "(will|expected to|set to|on track to) (die out|go extinct|disappear|vanish|collapse|be wiped out)", "reason": "prediction stated as fact. Describe what is observed and what is uncertain, never what is going to happen."},
      {"pattern": "[0-9][0-9,]* ?(visitors|readers|subscribers|followers|members|supporters)", "reason": "audience-scale class: this page is brand new and has no telemetry of any kind."},
      {"pattern": "(trusted|read|used|followed) by [0-9]", "reason": "social-proof class: unsupportable on a personal page with no audience."},
      {"pattern": "years of (experience|expertise|beekeeping|keeping bees|study)", "reason": "tenure class: this is a new page and asserts no credentials. The owner's own experience is not documented here and must not be invented."},
      {"pattern": "(i|we) (keep|have kept|manage|run|tend) (bees|hives|an apiary|colonies)", "reason": "first-person-practice class: whether the owner keeps bees is NOT known to the writer. Inventing a personal beekeeping history is fabrication about a real person. Unban only if the owner states it and it is registered as an attestation."},
      {"pattern": "(buy|shop|order|purchase|subscribe|sign up|book) (our|my|now|today|here)", "reason": "commercial-offer class: this page sells nothing and collects nothing. There is no offer, no product, no newsletter and no lead capture."},
      {"pattern": "(contact us|get in touch|request a|get a) (quote|consultation|callback|demo)", "reason": "lead-capture class: same reason. A personal page with a commercial call to action is a different page than the one the owner asked for."},
      {"pattern": "(api|apis|endpoint|rest|json|developer)s? (documentation|docs|reference|key|token)", "reason": "wrong-apis class, and the most consequential error available here. The domain also serves a REAL API on tools.apis.uk. The home page is about the INSECT and must never read as documentation for, or an index of, that API — a visitor who mistakes one for the other is the specific harm this page must avoid."}
    ],
    "allowed_entities": [
      "Apis", "Apis mellifera", "the honey bee", "the western honey bee",
      "the bumblebee", "Bombus", "the buff-tailed bumblebee",
      "solitary bees", "the mason bee", "the leafcutter bee", "the mining bee",
      "the queen", "drones", "worker bees", "foragers", "house bees",
      "a swarm", "a colony", "a hive", "an apiary", "honeycomb", "brood comb",
      "the waggle dance", "Karl von Frisch", "propolis", "royal jelly",
      "beeswax", "nectar", "pollen", "a nectar flow", "forage",
      "Varroa destructor", "the varroa mite", "colony collapse disorder",
      "the National Bee Unit", "the British Beekeepers Association", "Defra",
      "the Langstroth hive", "the National hive", "a skep", "Linnaeus"
    ],
    "writer_block": "NONE — NOTHING ON THIS SITE HAS YET BEEN VERIFIED.\n\nThere are no registered facts, so there are no numbers you may assert. Do not state any count, percentage, distance, weight, temperature, duration, lifespan, wingbeat rate, species tally, colony size or population trend — not about bees, not about beekeeping, and not about this page or the person who owns it.\n\nThis is not a limitation on what the page can be about. Bees are extraordinary in ways that need no figures at all. You can describe how a forager returning to a dark hive tells the others where she has been, by dancing on the vertical comb in a direction read against gravity. You can describe how wax is secreted, worked and built into cells of a shape that wastes nothing. You can describe the division of a colony's labour by age, the sound and drama of a swarm, the difference between a honey bee and the solitary bees that most people never notice, and why a bee visiting a flower for its own reasons happens to be the reason so much else grows. None of that requires a single quantity.\n\nIf a sentence seems to need a number to work, rewrite the sentence. 'A strong colony in high summer holds many thousands of bees' is not an improvement on 'a strong colony in high summer is crowded' — the first still asserts a magnitude. Prefer the second.\n\nWhere the substance genuinely is a quantity, say plainly that it is not stated here and why. 'Colony size swings so widely across the season that a single figure would mislead' is publishable, accurate, and more interesting than a number.\n\nDo NOT invent a personal history. You do not know whether the owner keeps bees, has ever kept bees, or simply likes them. Write about bees, not about an imagined beekeeper.\n\nNaming a species, a phenomenon, a hive design or an organisation is allowed. Making a factual claim about one is not, unless that claim appears in facts[] above — and nothing appears there yet.\n\nFinally, and specifically: this domain also runs a real API on a different hostname. This page is about the insect. Never write a sentence that could be read as API documentation, a developer reference, or a pointer to endpoints."
  }$eb$::jsonb,
  'manual',
  'Seeded at site creation, before any page was written, so the first page is covered. facts[] deliberately empty. Ban list targets SHAPES of the famous unsourced bee statistics, plus the Einstein misattribution, plus a wrong-apis class protecting the live tools.apis.uk API from being described by this page.',
  true, true, 'apis-uk-bees-2026-08-22'
FROM sites WHERE domain = 'apis.uk';

-- ------------------------------------------------------ imagery_style_guide --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'apis.uk')
  AND aspect = 'imagery_style_guide' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'imagery_style_guide',
  $img${
    "medium": "warm hand-drawn natural-history illustration — the register of a good field guide or a botanical plate, observed and affectionate rather than technical; ink line with flat colour, visible drawing, no photographic texture",
    "mood": "warm, curious, quietly characterful; the feeling of someone who finds bees genuinely wonderful and wants to show you why. Unhurried, generous, never corporate and never campaigning",
    "palette": "warm honey and amber (#c8871b, #e0a53a, #f0c56a), deep comb brown and near-black for line work (#2e2013, #171009), soft meadow and sage greens as a secondary only (#6b7f52, #9aab7e), warm paper off-white ground (#faf6ee)",
    "avoid": "cartoon bees with human faces, smiles, eyelashes or waving arms; clip-art and emoji-style bees; stock photography of any kind, especially honey jars, wooden dippers, breakfast tables and people in bee suits giving thumbs up; photorealistic macro insect photography (uncanny at large sizes and stylistically wrong for this page); yellow-and-black hazard stripes and warning iconography; anything suggesting an API, a network, a circuit, a dashboard, code, a terminal or connected nodes — this page is about the insect and imagery must never suggest software; text, lettering, numerals, logos or watermarks of any kind",
    "kinds": {
      "content_hero": {
        "medium": "flat two-or-three-colour editorial illustration with confident ink line",
        "mood": "one clear subject given room — a single honey bee on a flower head, or a fragment of comb geometry — strong silhouette, minimal background, plenty of warm paper showing through",
        "palette": "warm off-white ground, honey and amber flat shapes, deep comb brown line work, sage green sparingly",
        "avoid": "photorealism, photographic texture, gradients, 3D rendering, drop shadows, text, lettering, numerals, logos, watermarks, busy detail, dark or black backgrounds, neon or saturated colour outside the palette, hexagon patterns used as a generic tech motif",
        "reference_asset_keys": []
      }
    },
    "reference_asset_keys": []
  }$img$::jsonb,
  'manual',
  'Seeded pre-build: bugs_closed/027 — content_hero generates unstyled on a site with no style guide, and a fresh site has none. The avoid-list explicitly bans network/circuit/node motifs: on a domain that also serves an API, generic hexagon-and-node imagery would make the insect page read as a tech page.',
  true, true, 'apis-uk-bees-2026-08-22'
FROM sites WHERE domain = 'apis.uk';

-- ------------------------------------------------------------ roadmap_brief --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'apis.uk')
  AND aspect = 'roadmap_brief' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT
  id,
  'roadmap_brief',
  $rb${
    "text": "PHASE 1 — AND PHASE 1 IS THE WHOLE OF THIS SITE FOR NOW. Build exactly ONE page and no others: the home page, at the site root, page_type index. Do not plan, propose or create an about page, a contact page, a guides index, a blog, a glossary, a species directory, a tools page, a legal page or a privacy page. Do not add navigation items pointing at pages that do not exist. If a subject seems to deserve its own page, it belongs in a section of the home page instead, or it waits for a later phase the owner has not yet asked for. The home page is a single scrolling page about bees, written for a general visitor who arrived out of curiosity: a hero that says plainly what this page is, then a small number of content sections that each take one genuinely interesting aspect of bees and explain it well in plain prose, and a quiet footer. No pricing, no offer, no signup, no lead capture, no testimonials, no statistics band, no client logos, no call-to-action urging the visitor to do anything commercial. The only outbound reference that matters is a brief, plain, clearly-separated line acknowledging that this domain also runs an unrelated technical service on another hostname, so that a developer who lands here by mistake is not confused — one sentence, no link styling that competes with the page, and never presented as documentation."
  }$rb$::jsonb,
  'manual',
  'This aspect is what makes "home page only" real: build-site-planner reads site_specs.specs.roadmap_brief.text and treats it as authoritative ("Build ONLY the pages listed... Do NOT invent additional pages"). Owner scope decision 2026-08-22: home page only, personal/enthusiast angle. Deliberately contains NO FIGURES — a number in a spec is a given and outranks every writer-side rule.',
  true, true, 'apis-uk-bees-2026-08-22'
FROM sites WHERE domain = 'apis.uk';

COMMIT;

-- Verify
--   SELECT domain, email, status FROM sites WHERE domain='apis.uk';
--   SELECT aspect, is_current, pinned, created_by FROM site_specs ss
--     JOIN sites s ON s.id=ss.site_id WHERE s.domain='apis.uk' ORDER BY aspect;
--   SELECT jsonb_array_length(data->'banned_claims') AS bans,
--          jsonb_array_length(data->'facts') AS facts
--     FROM site_specs ss JOIN sites s ON s.id=ss.site_id
--    WHERE s.domain='apis.uk' AND ss.aspect='evidence_base' AND ss.is_current;

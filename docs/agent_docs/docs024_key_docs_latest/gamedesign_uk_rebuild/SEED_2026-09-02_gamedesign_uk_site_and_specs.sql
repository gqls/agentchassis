-- ============================================================================
-- gamedesign.uk — site row + mission (palette/typography) + design_intent +
--                 evidence_base + imagery_style_guide
-- oxenunity.com — site row only (owner ruling 2026-09-02: "should have a row,
--                 it was created here"; hand-built single page, oufe/RUNBOOK §1)
--
-- Written 2026-09-02 by the gamedesign.uk lane. Applied out of band (psql -f),
-- NOT via the migration runner: per-site setup, not a platform schema change.
-- Pattern: oufe/SEED_2026-07-25_oufe_site_and_specs.sql.
--
-- WHY BEFORE SUBMISSION (082 FRESH path, domain-submitter):
--   1. sites row with EMAIL — bugs_open/063: the hallucinated-email check FAILS
--      OPEN with no contact email. ensure_site_record upserts on domain, so
--      pre-creating is safe. ⚠ name AND network_id are set EXPLICITLY:
--      ensure_site_record scans both without COALESCE and a NULL stalls the
--      build at needs_site_plan (positioning lane, 2026-09-02, sibling flow).
--   2. mission.preferred_palette / preferred_typography — rung 1 of the palette
--      cascade (resolve_composition_pallette_action.go) and rung 2 of the
--      typography cascade. The classifier does NOT write `mission`; domain-
--      submitter's persist_mission DEEP-MERGES {"text": <brief>} over this row
--      (site_spec_actions.go:164, :245), so these keys survive the submission.
--      Values from the theme-kits lane, 2026-09-02: warm paper ground, serif
--      headings, earth accent — hue ~24° vs the sibling's ~187°, light vs dark,
--      serif vs sans: distinct on three axes at once. Kits themselves are NOT
--      live (theme_kits table unapplied); these are the values seeded directly.
--   3. design_intent — typography.reference_values is rung 1 of ITS cascade
--      (reversed vs palette), so seed it here too; style_direction steers the
--      layout tag-match toward `soft-editorial` (light/editorial). The
--      classifier WILL deep-merge its own design_intent over this and may
--      overwrite reference_values — that is why mission carries the same
--      values. colour_mood is written to AGREE with the values (boxingonline
--      lesson: prose asking for dark over values encoding light resolves
--      silently to the machine-readable half).
--   4. evidence_base — the claims layer is gated on the PRESENCE of this
--      aspect (loadEvidenceBase nil ⇒ every lane no-ops). Seeding it before the
--      first page is the only way the first page is covered. banned_claims
--      encode the brief's don'ts as shapes, not individual numbers.
--   5. imagery_style_guide — bugs_closed/027: content_hero generates unstyled on
--      a site with none.
--
-- OWNER RULING 2026-09-02 (theme kits, relayed): reference_values is NOT a pin.
-- The render overlay has authority to move off these values. Seeded = what the
-- site STARTS from; check the served stylesheet after the build, do not clamp.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ------------------------------------------------------ gamedesign.uk row --
INSERT INTO sites (domain, name, network_id, status, email, company_name)
VALUES (
  'gamedesign.uk',
  'gamedesign.uk',
  '00000000-0000-0000-0000-000000000002',
  'active',
  'gamedesign@contactforsales.com',
  'gamedesign.uk'
)
ON CONFLICT (domain) DO UPDATE
  SET email      = COALESCE(sites.email, EXCLUDED.email),
      name       = COALESCE(sites.name, EXCLUDED.name),
      network_id = COALESCE(sites.network_id, EXCLUDED.network_id);
-- status='active' is what upsertSite writes; NOT in the validated vocabulary.

-- ------------------------------------------------------ oxenunity.com row --
-- Hand-built single page, live, never to be built by the framework. The row
-- exists so the platform KNOWS the domain (bugs_open/432): it will show as
-- ROW_NO_PAGES in audit-rowless-serving-domains.sh, which is the honest class.
-- Email deliberately NULL — none was supplied and none may be invented.
INSERT INTO sites (domain, name, network_id, status, build_status, company_name, settings)
VALUES (
  'oxenunity.com',
  'oxenunity.com',
  '00000000-0000-0000-0000-000000000002',
  'deployed',
  'deployed',
  'Oxen Unity',
  '{"managed_by": "hand", "seeded_by": "gamedesign_uk_rebuild 2026-09-02", "reason": "owner ruling 2026-09-02: should have a row, it was created here; hand-authored single page per oufe/RUNBOOK_oufe.md s1 — do NOT dispatch a build at it"}'::jsonb
)
ON CONFLICT (domain) DO NOTHING;

-- ---------------------------------------------------------------- mission --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND aspect = 'mission' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT id, 'mission',
  $m${
    "preferred_palette": {
      "background": "#F4F1EA",
      "surface":    "#FFFFFF",
      "primary":    "#33302B",
      "secondary":  "#6E6558",
      "accent":     "#A6521F",
      "text":       "#23211E",
      "text_muted": "#6B655C",
      "border":     "#DDD6C9"
    },
    "preferred_typography": {
      "font_family":  "'Lato', Georgia, 'Times New Roman', serif",
      "heading_font": "'Merriweather', Georgia, 'Times New Roman', serif",
      "mono_font":    "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    },
    "palette_rationale": "theme-kits lane 2026-09-02: practice-journal complement to the sibling gamesdesign.co.uk (#121212 ground / #00bcd4 cyan / Segoe UI sans). Warm paper ground, serif headings, muted earth accent — distinct on ground lightness, hue and letterform class. Contrast checked: text/ground ~14:1, accent/ground ~5.9:1, white/accent ~6.6:1, muted/ground ~4.8:1. Typography = library row serif-editorial; intended layout = soft-editorial (light/editorial)."
  }$m$::jsonb,
  'manual',
  'Pre-seeded before 082 submission. domain-submitter persist_mission deep-merges {"text": <brief>} over this row; these keys survive. Rung 1 of the palette cascade.',
  true, true, 'gamedesign-uk-lane-2026-09-02'
FROM sites WHERE domain = 'gamedesign.uk';

-- ---------------------------------------------------------- design_intent --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND aspect = 'design_intent' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT id, 'design_intent',
  $di${
    "source": "manual pre-seed 2026-09-02 (theme-kits lane values); the classifier will merge over this",
    "dark_light": "light",
    "style_direction": "light, editorial, practice-journal. A warm off-white paper ground, serif headings over a humanist sans body, generous measure and line-height, restrained earth-toned accent used for links and emphasis only. Reads as a professional practice journal for working game designers, leads and producers — not a tool dashboard, not a general design blog, not a consumer site. Editorial page grammar: long-form article layouts, section indexes, pull-quotes; no calculators, no tool panels, no gamified UI.",
    "colour_mood": "warm paper and ink with a single muted earth accent: off-white #F4F1EA ground, near-black warm ink #23211E, a rust-brown #A6521F for links and emphasis. Calm, considered, unhurried. Deliberately the complement of the sibling site's dark ground and cyan.",
    "typography_mood": "serif headings (Merriweather) carry the editorial voice; a humanist sans-friendly serif body (Lato stack) keeps long reading comfortable. No display faces, no geometric sans headings.",
    "overall_character": "the written record of how game design is actually practised inside studios — process, judgement, workflow — written for professionals by people who do the work",
    "layout_preference": "soft-editorial",
    "imagery_direction": "restrained; see imagery_style_guide",
    "avoid": "dark grounds, cyan or electric accents, neon, saturated primaries, cartoon or animation styling, tool-dashboard chrome, calculator panels, geometric-sans display headings, stark pure-white minimalism",
    "palette": {
      "reference_values": {
        "background": "#F4F1EA",
        "surface":    "#FFFFFF",
        "primary":    "#33302B",
        "secondary":  "#6E6558",
        "accent":     "#A6521F",
        "text":       "#23211E",
        "text_muted": "#6B655C",
        "border":     "#DDD6C9"
      }
    },
    "typography": {
      "reference_values": {
        "font_family":  "'Lato', Georgia, 'Times New Roman', serif",
        "heading_font": "'Merriweather', Georgia, 'Times New Roman', serif"
      }
    }
  }$di$::jsonb,
  'manual',
  'Pre-seeded before 082 submission. typography.reference_values is rung 1 of the typography cascade; style_direction steers deriveSiteScheme/tag-match toward soft-editorial. The classifier deep-merges over this; mission carries the same palette as the reliable rung-1 lever.',
  true, true, 'gamedesign-uk-lane-2026-09-02'
FROM sites WHERE domain = 'gamedesign.uk';

-- ----------------------------------------------------------- evidence_base --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND aspect = 'evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT id, 'evidence_base',
  $eb${
    "governing_rule": "This is an editorial site about the professional practice of game design. Opinion and process description are its substance and need no citation, but every FIGURE (a studio size, a salary, a headcount, a percentage, a year count, a reader count), every NAMED real studio, person, product or quotation, and every claim about what research or a study found must trace to a fact below carrying a source and a capture date. Where no fact exists, the page says so plainly or rewrites the sentence not to need it. Nothing on this site asserts that a paid product, tier, subscription or licence exists, and nothing rules one out.",
    "audit_doc": "docs/agent_docs/docs024_key_docs_latest/gamedesign_uk_rebuild/ (PLAN D4 positioning; MISSION_2026-09-02_gamedesign_uk.txt)",
    "schema_notes": "facts[]: {id, claim, kind: metric|capability|entity|attestation, source: EXACTLY ONE of {sql|artifact|attested_by|citation}, verified_at, value?, tolerance?, context_terms?, writer_line?}. banned_claims[]: {pattern (case-insensitive regex; an invalid regex degrades to a literal substring), reason}. allowed_entities[]: real named entities it is legitimate to NAME — naming is not asserting.",
    "facts": [],
    "banned_claims": [
      {"pattern": "gamedesign\\.uk pro|gamesdesign\\.co\\.uk pro", "reason": "positioning lane 2026-09-02: the name appears in NO current spec of either site. The commercial slot is prepared, never claimed or named."},
      {"pattern": "(subscription|per[- ]seat|pricing plan|price list|free trial|premium tier|pro tier|paid tier|upgrade to (pro|premium))", "reason": "paid-product class: nothing on this site may say a paid offering exists. Brief: prepare the slot, never claim it."},
      {"pattern": "(we (don'?t|do not) (sell|charge)|(completely|entirely|totally|100%) free|no (strings|catch)|nothing to (buy|pay))", "reason": "negative-identity class: owner rule 2026-09-02, do not define the site by what it is not, and do not foreclose a future offering."},
      {"pattern": "game ?rooms?", "reason": "portfolio collision with gamerooms.co.uk (positioning lane)."},
      {"pattern": "[0-9][0-9,]* ?(studios|teams|designers|leads|producers|readers|subscribers|members|users|professionals|practitioners)", "reason": "audience-scale class: a new site with no telemetry, and no studio census exists. Write the sentence without the count."},
      {"pattern": "(salary|salaries|earn(s|ing)?|day rate|per annum|p\\.a\\.)[^.]{0,40}(£|\\$|€|[0-9])", "reason": "pay figures: the brief forbids salaries unless the owner supplies them. None supplied."},
      {"pattern": "(years|decades) of (experience|expertise|practice)", "reason": "tenure class: a new publication has no history to claim."},
      {"pattern": "(studies|research|surveys?|data) (show|shows|found|finds|suggest|suggests|indicate|prove)", "reason": "unsourced-study class: a research claim needs a registered fact with a citation. None registered."},
      {"pattern": "(sources|people) (close to|familiar with) (the|a) (studio|team|matter|project)", "reason": "fabricated sourcing: we have no sources."},
      {"pattern": "(seamless|effortless|game-?changing|best-in-class|world-class|industry-leading|cutting-edge|revolutionary)", "reason": "brief: do not call anything powerful, seamless, effortless, essential or ultimate; describe what it does."},
      {"pattern": "(trusted|used|read) by [0-9]", "reason": "social-proof class: unsupportable."},
      {"pattern": "(calculator|simulator|tool) (page|panel|widget)s? (below|above|on this page)", "reason": "content-kind collision: this site publishes no calculators or tool pages; those live on gamesdesign.co.uk and are linked, not built."}
    ],
    "allowed_entities": [
      "Unity", "Unreal Engine", "Godot", "Steam", "Jira", "Confluence", "Notion", "Miro", "Figma", "Perforce", "Git",
      "GDC", "the Game Developers Conference", "gamesdesign.co.uk", "gamedesign.uk"
    ],
    "writer_block": "NO FIGURE ON THIS SITE HAS YET BEEN VERIFIED.\n\nThere are no registered facts, so there are no numbers you may assert: no studio sizes, headcounts, salaries, day rates, percentages, years of experience, reader counts, or counts of anything, and no dates attributed to real events. If a sentence seems to need a number, rewrite it so it does not — a process, a workflow or a judgement never needs one.\n\nDo not invent named studios, named people, case studies, job titles attributed to real organisations, or quotations. You may name well-known engines and tools (Unity, Unreal, Godot, Jira and the like) as things a team uses; you may not make claims about their makers or market share.\n\nDo not say a paid product, tier, subscription or licence exists; do not name one; do not describe or price one. Do not say the site is free or that it sells nothing. Say what the site is.\n\nWhere a tool or calculator would naturally be referenced, link to gamesdesign.co.uk rather than describing or building one here.\n\nWhere an opinion is given, give it as an opinion and say whose it is — the site's own editorial voice is fine; a real named person's is not, unless registered above, and nothing is registered above yet."
  }$eb$::jsonb,
  'manual',
  'Seeded at site creation, before any page was written, so the first page is covered. facts[] deliberately empty. banned_claims encode the brief and the positioning collisions as SHAPES.',
  true, true, 'gamedesign-uk-lane-2026-09-02'
FROM sites WHERE domain = 'gamedesign.uk';

-- ------------------------------------------------------ imagery_style_guide --
UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'gamedesign.uk')
  AND aspect = 'imagery_style_guide' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, pinned, created_by)
SELECT id, 'imagery_style_guide',
  $img${
    "medium": "restrained editorial illustration — ink-and-wash or flat two-tone line work on warm paper; diagrams of process and flow (boards, timelines, loops, hand-drawn schematics) rather than pictures of games",
    "mood": "considered, workmanlike, studio-desk; the register of a practitioner's notebook rather than a marketing site or a game trailer",
    "palette": "warm off-white paper ground (#F4F1EA), warm near-black ink (#23211E, #33302B), warm greys (#6E6558, #DDD6C9), a single muted rust-brown accent (#A6521F) used sparingly",
    "avoid": "screenshots or renders of real or invented games; game characters, weapons, loot, coins, gems, health bars, controllers; cartoon or anime styling; saturated primaries, neon, cyan, electric blue; dark or black backgrounds; stock photography of people at desks or in meetings; photorealistic depictions of identifiable individuals; text, lettering, numerals, logos, UI chrome or watermarks of any kind; any drawn chart, graph or plotted data",
    "kinds": {
      "content_hero": {
        "medium": "flat two-tone editorial illustration, ink on warm paper",
        "mood": "one clear abstract motif of process or structure — a loop, a hand-off between two shapes, a layered plan, a sequence of stages — minimal detail, strong silhouette, plenty of paper",
        "palette": "warm paper ground, warm near-black ink shapes and line, rust-brown accent on at most one element",
        "avoid": "photorealism, photographic texture, gradients, 3D rendering, drop shadows, dark backgrounds, cyan or blue accents, game imagery, characters, text, lettering, numerals, logos, watermarks, busy detail",
        "reference_asset_keys": []
      }
    },
    "reference_asset_keys": []
  }$img$::jsonb,
  'manual',
  'Seeded pre-build: bugs_closed/027 — content_hero generates unstyled on a site with no style guide. Avoid-list excludes game imagery and cartoon styling explicitly (portfolio collisions) and any drawn chart.',
  true, true, 'gamedesign-uk-lane-2026-09-02'
FROM sites WHERE domain = 'gamedesign.uk';

COMMIT;

-- ---------------------------------------------------------------- verify --
SELECT domain, name, network_id, status, email, company_name FROM sites WHERE domain IN ('gamedesign.uk','oxenunity.com') ORDER BY domain;
SELECT aspect, source, pinned, created_at::timestamp(0) FROM site_specs
WHERE site_id = (SELECT id FROM sites WHERE domain='gamedesign.uk') AND is_current ORDER BY aspect;

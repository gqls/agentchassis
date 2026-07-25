-- 211_cta_gate_every_active_anchor.sql — 2026-07-25, cta_link_integrity (bugs_open/023, classes A/B/C/E)
--
-- CLOSES bugs_open/023's criterion 3: "no component in the ACTIVE LIBRARY can pair a
-- rendered label with an absent destination". Migrations 179 and 181 did this for the
-- four components that were on live pages; this does it for the whole library, which is
-- what makes the property structural rather than a description of today's page set.
--
-- WHAT IT DOES
--   A. Gates every ungated CTA anchor: <a href="{{.x_url}}"> ... </a>  becomes
--      {{if .x_url}}<a href="{{.x_url}}"> ... </a>{{end}}.
--      156 anchors across 31 active components (measured live 2026-07-25 with
--      cta_link_integrity/scripts/parse_gates.py — a real template parse, NOT R9's
--      lookback heuristic, which undercounts by 2.4x).
--   B. Four components use an idiom the field parser cannot see, fixed by exact needle:
--      brief-explanation x2  href="{{if .x}}{{.x}}{{else}}#{{end}}" -> gate the anchor,
--                            plain href. This renders a LIVE bare '#' on 4 placements /
--                            3 sites today (vonc.com, robot-hands.com).
--      category-listing      hardcoded href="#" on "View all <category>" (1 live
--                            placement, dartsonline.com) -> new renderer/optional
--                            category_url + gate. Same shape migration 179 removed
--                            from tool-guide-intro (#guide-start).
--      Pricing Tiers x3      hardcoded href="#" on all three tier buttons (1 live
--                            placement, gaswholesalers.com = 3 of the fleet's bare '#')
--                            -> new renderer/optional tier_N_cta_url + gate.
--      featured_article      hardcoded href="#" on Read More (0 placements,
--                            prophylactic) -> new renderer/optional read_more_url + gate.
--   C. Class E, fleet-wide: no URL field may be source:llm AND required:true. That pair
--      is an instruction to a model to author a value it cannot look up, and it is what
--      produced leopardess.contactforsales.com (an @->. transform of the site's own
--      contact email) and the finetuning.ai variant. 23 fields across 5 components are
--      flipped to required:false. source stays llm deliberately: flipping the source with
--      no resolver behind it deletes values that are real today (the platform-comparison
--      lesson from migration 181). required:false + a gated anchor removes the COMPULSION
--      to invent and the dead control, without deleting a working button.
--
-- WHAT IT DOES NOT DO, ON PURPOSE
--   * 17 anchors in 13 components sit inside {{range}} — the field belongs to the ranged
--     item ({{range .items}}<a href="{{.url}}">), fed by a query. Different class,
--     different owner; gating them can delete a whole card, not just a control
--     (image-hover-card-grid's anchor wraps the card's image and title). Recorded in 023.
--   * provocations-archive-list's href="#" is a hidden [data-archive-template] clone
--     source, hidden by its own CSS and filled by JS. Gating it would break the archive.
--   * lobby-grid and provocation-card contain LITERAL '<no value>' in html_template —
--     those rows are rendered artefacts saved back as templates. A different defect
--     class; recorded in 023, not repaired here.
--
-- SAFETY. Every template change is gated on md5(html_template) BEFORE and asserted on
-- md5 AFTER, per component. If any other session has edited any of these 35 templates
-- since 2026-07-25, the whole transaction aborts rather than half-applying — this matters
-- because content_components is shared, live, and several sessions write it daily.
-- The regexp used below was proved equivalent to an audited Python transform by comparing
-- md5s read-only before writing this file. GOTCHA (cost a round): in POSIX ARE the FIRST
-- quantifier fixes the whole RE's greediness, so '<a[^>]*href=...*?</a>' matches to the
-- LAST </a> in the template. '[^>]*?' (non-greedy) is load-bearing, not cosmetic.
--
-- Config change: LIVE IMMEDIATELY, no image roll. It takes effect per page at that page's
-- next render; it does not rewrite already-deployed HTML.
--
-- ROLLBACK: bak_cta_gates_20260725 holds every touched row in full.

\set ON_ERROR_STOP on
BEGIN;

CREATE TABLE bak_cta_gates_20260725 AS
SELECT * FROM content_components
WHERE is_active OR name IN ('Pricing Tiers',
  'ai-readiness-quiz',
  'archetype-combinations',
  'archetype-grid',
  'archetype-result-card',
  'bayesian-ranking-hero-tool_pre_037',
  'blog-listing_pre_037',
  'bridging-loan-calculator',
  'brief-explanation',
  'case-studies-grid',
  'category-listing',
  'content-sidebar',
  'featured-inventory',
  'featured_article',
  'footer-with-disclaimer_pre_037',
  'funding-fit',
  'game-master-explanation',
  'gauntlet-cta',
  'header-docs',
  'header-minimal-tool_pre_037',
  'header-with-cart-or-nav_pre_037',
  'header-with-categories_pre_037',
  'header-with-search_pre_037',
  'patent-check',
  'product-hero_pre_037',
  'product-specs',
  'provocation-feed',
  'provocations-archive-list',
  'site-head',
  'system-stats',
  'tool-ai-agent-roi-estimator',
  'tool-archetype-taster-quiz',
  'tool-btl-investor_pre_037',
  'tool-equity-release_pre_037',
  'tool-mortgage-affordability');


-- ── pre-conditions: every template is byte-identical to what was measured ──────
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM (VALUES
    ('Pricing Tiers','a393013c5808837b165826869956528f'),
    ('ai-readiness-quiz','c55802544eefcb177218d05b0a77c614'),
    ('archetype-combinations','a4da1606201542f21f3cfd3c23db0b51'),
    ('archetype-grid','1b08458159854555407b018ba6e672a1'),
    ('archetype-result-card','0623a25a70d17e26c345708e7b3ce50c'),
    ('bayesian-ranking-hero-tool_pre_037','1a8b5bf0b55e2844838d4be425d3a1c1'),
    ('blog-listing_pre_037','4ddac4615e7b6da0f02fe2edf4835cd4'),
    ('bridging-loan-calculator','aa683b20ca94744fae358aac9af37b9c'),
    ('brief-explanation','8385804ef7e0c8dad7fc4cdfa3067618'),
    ('case-studies-grid','98d63577d6b85d826ed4efccc24f9cf0'),
    ('category-listing','51d82a3e407eaec14eed9fb9528cf9f9'),
    ('content-sidebar','b9e2ee2647088defdc4949bd784402e6'),
    ('featured-inventory','619dbc6f757beea877bacb27cbda880f'),
    ('featured_article','3e69f082cf839c7f06f11915a4dc7b1f'),
    ('footer-with-disclaimer_pre_037','864f9071ca58403100d1bc654429f597'),
    ('funding-fit','1b4b876da603eaeb1a8631d35876e4fa'),
    ('game-master-explanation','073ab7739f57cc85969a6a4e7b801fda'),
    ('gauntlet-cta','b545896b3d6f8a56be79258c1922a003'),
    ('header-docs','490b2a0650be8b2a0f6b83b4ef7dabe9'),
    ('header-minimal-tool_pre_037','4809ae9865e47885cbf1423153d52b00'),
    ('header-with-cart-or-nav_pre_037','377a56ab197f05ecd9285e86742d6256'),
    ('header-with-categories_pre_037','1a4962809294545c203bc980dca8a3af'),
    ('header-with-search_pre_037','37c7d3265b0337582201b131727babd6'),
    ('patent-check','38582d63d1280f493a0503164f6dfe54'),
    ('product-hero_pre_037','1dd3fd5cf5a8d39783a16a597ce01c7a'),
    ('product-specs','a0f06fb09a4d8c967114c7666055b028'),
    ('provocation-feed','7546620c371527689518d22907987a57'),
    ('provocations-archive-list','b9c5131e7fa282af4c4571e0e4078159'),
    ('site-head','8b7593b49b0d937b072994e3c2f440e6'),
    ('system-stats','f4225214b3e52db168486e7af7e95ed5'),
    ('tool-ai-agent-roi-estimator','bcda267ea40a28348f82cad0ab2adb3d'),
    ('tool-archetype-taster-quiz','2de5d4d9ac8fb96972e8856346954819'),
    ('tool-btl-investor_pre_037','4a29a4937801365edc7730e485dad07f'),
    ('tool-equity-release_pre_037','2c19addba6b41f9564a9e562c7af1f56'),
    ('tool-mortgage-affordability','c0162d9c1e9c218e4426749155e1447e')
  ) AS t(name, md5_before)
  JOIN content_components c ON c.name = t.name AND md5(c.html_template) = t.md5_before;
  IF n <> 35 THEN
    RAISE EXCEPTION 'pre-condition failed: % of 35 templates match the measured md5 — another session has edited one; re-derive with parse_gates.py', n;
  END IF;
END $$;


-- ── A. gate every ungated CTA anchor ─────────────────────────────────────────
-- ai-readiness-quiz: 1 anchor(s), field(s): result_cta_primary_url
UPDATE content_components SET html_template =
       regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.result_cta_primary_url\}\}".*?</a>)', '{{if .result_cta_primary_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'ai-readiness-quiz';

-- archetype-combinations: 2 anchor(s), field(s): cta_primary_url, cta_secondary_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_primary_url\}\}".*?</a>)', '{{if .cta_primary_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_secondary_url\}\}".*?</a>)', '{{if .cta_secondary_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'archetype-combinations';

-- archetype-grid: 1 anchor(s), field(s): cta_url
UPDATE content_components SET html_template =
       regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_url\}\}".*?</a>)', '{{if .cta_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'archetype-grid';

-- archetype-result-card: 2 anchor(s), field(s): cta_primary_url, cta_secondary_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_primary_url\}\}".*?</a>)', '{{if .cta_primary_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_secondary_url\}\}".*?</a>)', '{{if .cta_secondary_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'archetype-result-card';

-- bayesian-ranking-hero-tool_pre_037: 2 anchor(s), field(s): cta_primary_url, cta_secondary_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_primary_url\}\}".*?</a>)', '{{if .cta_primary_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_secondary_url\}\}".*?</a>)', '{{if .cta_secondary_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'bayesian-ranking-hero-tool_pre_037';

-- blog-listing_pre_037: 7 anchor(s), field(s): cta_url, post1_url, post2_url, post3_url, post4_url, post5_url, post6_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_url\}\}".*?</a>)', '{{if .cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.post1_url\}\}".*?</a>)', '{{if .post1_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.post2_url\}\}".*?</a>)', '{{if .post2_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.post3_url\}\}".*?</a>)', '{{if .post3_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.post4_url\}\}".*?</a>)', '{{if .post4_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.post5_url\}\}".*?</a>)', '{{if .post5_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.post6_url\}\}".*?</a>)', '{{if .post6_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'blog-listing_pre_037';

-- bridging-loan-calculator: 1 anchor(s), field(s): cta_link
UPDATE content_components SET html_template =
       regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_link\}\}".*?</a>)', '{{if .cta_link}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'bridging-loan-calculator';

-- case-studies-grid: 6 anchor(s), field(s): card1_link_url, card2_link_url, card3_link_url, card4_link_url, card5_link_url, cta_link_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.card1_link_url\}\}".*?</a>)', '{{if .card1_link_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.card2_link_url\}\}".*?</a>)', '{{if .card2_link_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.card3_link_url\}\}".*?</a>)', '{{if .card3_link_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.card4_link_url\}\}".*?</a>)', '{{if .card4_link_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.card5_link_url\}\}".*?</a>)', '{{if .card5_link_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_link_url\}\}".*?</a>)', '{{if .cta_link_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'case-studies-grid';

-- content-sidebar: 5 anchor(s), field(s): cta_url, link_four_url, link_one_url, link_three_url, link_two_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_url\}\}".*?</a>)', '{{if .cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.link_four_url\}\}".*?</a>)', '{{if .link_four_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.link_one_url\}\}".*?</a>)', '{{if .link_one_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.link_three_url\}\}".*?</a>)', '{{if .link_three_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.link_two_url\}\}".*?</a>)', '{{if .link_two_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'content-sidebar';

-- featured-inventory: 7 anchor(s), field(s): item1_cta_url, item2_cta_url, item3_cta_url, item4_cta_url, item5_cta_url, item6_cta_url, view_all_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.item1_cta_url\}\}".*?</a>)', '{{if .item1_cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.item2_cta_url\}\}".*?</a>)', '{{if .item2_cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.item3_cta_url\}\}".*?</a>)', '{{if .item3_cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.item4_cta_url\}\}".*?</a>)', '{{if .item4_cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.item5_cta_url\}\}".*?</a>)', '{{if .item5_cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.item6_cta_url\}\}".*?</a>)', '{{if .item6_cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.view_all_url\}\}".*?</a>)', '{{if .view_all_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'featured-inventory';

-- footer-with-disclaimer_pre_037: 18 anchor(s), field(s): col1_link1_url, col1_link2_url, col1_link3_url, col1_link4_url, col2_link1_url, col2_link2_url, col2_link3_url, col2_link4_url, col3_link1_url, col3_link2_url, col3_link3_url, col3_link4_url, cookie_policy_url, privacy_policy_url, social_link_1_url, social_link_2_url, social_link_3_url, terms_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.col1_link1_url\}\}".*?</a>)', '{{if .col1_link1_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.col1_link2_url\}\}".*?</a>)', '{{if .col1_link2_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.col1_link3_url\}\}".*?</a>)', '{{if .col1_link3_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.col1_link4_url\}\}".*?</a>)', '{{if .col1_link4_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.col2_link1_url\}\}".*?</a>)', '{{if .col2_link1_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.col2_link2_url\}\}".*?</a>)', '{{if .col2_link2_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.col2_link3_url\}\}".*?</a>)', '{{if .col2_link3_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.col2_link4_url\}\}".*?</a>)', '{{if .col2_link4_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.col3_link1_url\}\}".*?</a>)', '{{if .col3_link1_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.col3_link2_url\}\}".*?</a>)', '{{if .col3_link2_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.col3_link3_url\}\}".*?</a>)', '{{if .col3_link3_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.col3_link4_url\}\}".*?</a>)', '{{if .col3_link4_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cookie_policy_url\}\}".*?</a>)', '{{if .cookie_policy_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.privacy_policy_url\}\}".*?</a>)', '{{if .privacy_policy_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.social_link_1_url\}\}".*?</a>)', '{{if .social_link_1_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.social_link_2_url\}\}".*?</a>)', '{{if .social_link_2_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.social_link_3_url\}\}".*?</a>)', '{{if .social_link_3_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.terms_url\}\}".*?</a>)', '{{if .terms_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'footer-with-disclaimer_pre_037';

-- funding-fit: 2 anchor(s), field(s): guide_url, result_cta_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.guide_url\}\}".*?</a>)', '{{if .guide_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.result_cta_url\}\}".*?</a>)', '{{if .result_cta_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'funding-fit';

-- game-master-explanation: 2 anchor(s), field(s): cta_primary_url, cta_secondary_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_primary_url\}\}".*?</a>)', '{{if .cta_primary_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_secondary_url\}\}".*?</a>)', '{{if .cta_secondary_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'game-master-explanation';

-- gauntlet-cta: 2 anchor(s), field(s): cta_primary_url, cta_secondary_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_primary_url\}\}".*?</a>)', '{{if .cta_primary_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_secondary_url\}\}".*?</a>)', '{{if .cta_secondary_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'gauntlet-cta';

-- header-docs: 14 anchor(s), field(s): brand_url, cta_url, github_url, nav_link_1_url, nav_link_2_url, nav_link_3_url, nav_link_4_url, nav_link_5_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.brand_url\}\}".*?</a>)', '{{if .brand_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_url\}\}".*?</a>)', '{{if .cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.github_url\}\}".*?</a>)', '{{if .github_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_1_url\}\}".*?</a>)', '{{if .nav_link_1_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_2_url\}\}".*?</a>)', '{{if .nav_link_2_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_3_url\}\}".*?</a>)', '{{if .nav_link_3_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_4_url\}\}".*?</a>)', '{{if .nav_link_4_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_5_url\}\}".*?</a>)', '{{if .nav_link_5_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'header-docs';

-- header-minimal-tool_pre_037: 4 anchor(s), field(s): primary_cta_url, secondary_cta_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.primary_cta_url\}\}".*?</a>)', '{{if .primary_cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.secondary_cta_url\}\}".*?</a>)', '{{if .secondary_cta_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'header-minimal-tool_pre_037';

-- header-with-cart-or-nav_pre_037: 11 anchor(s), field(s): cart_url, cta_url, nav_link_1_url, nav_link_2_url, nav_link_3_url, nav_link_4_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cart_url\}\}".*?</a>)', '{{if .cart_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_url\}\}".*?</a>)', '{{if .cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_1_url\}\}".*?</a>)', '{{if .nav_link_1_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_2_url\}\}".*?</a>)', '{{if .nav_link_2_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_3_url\}\}".*?</a>)', '{{if .nav_link_3_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_4_url\}\}".*?</a>)', '{{if .nav_link_4_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'header-with-cart-or-nav_pre_037';

-- header-with-categories_pre_037: 27 anchor(s), field(s): cat_1_url, cat_2_url, cat_3_url, cat_4_url, cat_5_url, cat_6_url, cat_7_url, cta_url, dropdown_1_link_1_url, dropdown_1_link_2_url, dropdown_1_link_3_url, nav_link_1_url, nav_link_2_url, nav_link_3_url, nav_link_4_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cat_1_url\}\}".*?</a>)', '{{if .cat_1_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cat_2_url\}\}".*?</a>)', '{{if .cat_2_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cat_3_url\}\}".*?</a>)', '{{if .cat_3_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cat_4_url\}\}".*?</a>)', '{{if .cat_4_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cat_5_url\}\}".*?</a>)', '{{if .cat_5_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cat_6_url\}\}".*?</a>)', '{{if .cat_6_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cat_7_url\}\}".*?</a>)', '{{if .cat_7_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_url\}\}".*?</a>)', '{{if .cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.dropdown_1_link_1_url\}\}".*?</a>)', '{{if .dropdown_1_link_1_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.dropdown_1_link_2_url\}\}".*?</a>)', '{{if .dropdown_1_link_2_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.dropdown_1_link_3_url\}\}".*?</a>)', '{{if .dropdown_1_link_3_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_1_url\}\}".*?</a>)', '{{if .nav_link_1_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_2_url\}\}".*?</a>)', '{{if .nav_link_2_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_3_url\}\}".*?</a>)', '{{if .nav_link_3_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_4_url\}\}".*?</a>)', '{{if .nav_link_4_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'header-with-categories_pre_037';

-- header-with-search_pre_037: 10 anchor(s), field(s): cta_url, nav_link_1_url, nav_link_2_url, nav_link_3_url, nav_link_4_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_url\}\}".*?</a>)', '{{if .cta_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_1_url\}\}".*?</a>)', '{{if .nav_link_1_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_2_url\}\}".*?</a>)', '{{if .nav_link_2_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_3_url\}\}".*?</a>)', '{{if .nav_link_3_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_4_url\}\}".*?</a>)', '{{if .nav_link_4_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'header-with-search_pre_037';

-- patent-check: 2 anchor(s), field(s): guide_url, result_cta_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.guide_url\}\}".*?</a>)', '{{if .guide_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.result_cta_url\}\}".*?</a>)', '{{if .result_cta_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'patent-check';

-- product-hero_pre_037: 2 anchor(s), field(s): cta_primary_url, cta_secondary_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_primary_url\}\}".*?</a>)', '{{if .cta_primary_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_secondary_url\}\}".*?</a>)', '{{if .cta_secondary_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'product-hero_pre_037';

-- product-specs: 2 anchor(s), field(s): cta_primary_url, cta_secondary_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_primary_url\}\}".*?</a>)', '{{if .cta_primary_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_secondary_url\}\}".*?</a>)', '{{if .cta_secondary_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'product-specs';

-- provocation-feed: 7 anchor(s), field(s): card1_link, card2_link, card3_link, card4_link, card5_link, card6_link, cta_link
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.card1_link\}\}".*?</a>)', '{{if .card1_link}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.card2_link\}\}".*?</a>)', '{{if .card2_link}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.card3_link\}\}".*?</a>)', '{{if .card3_link}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.card4_link\}\}".*?</a>)', '{{if .card4_link}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.card5_link\}\}".*?</a>)', '{{if .card5_link}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.card6_link\}\}".*?</a>)', '{{if .card6_link}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_link\}\}".*?</a>)', '{{if .cta_link}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'provocation-feed';

-- provocations-archive-list: 1 anchor(s), field(s): cta_url
UPDATE content_components SET html_template =
       regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_url\}\}".*?</a>)', '{{if .cta_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'provocations-archive-list';

-- site-head: 12 anchor(s), field(s): cta_primary_url, cta_secondary_url, nav_link_1_url, nav_link_2_url, nav_link_3_url, nav_link_4_url
UPDATE content_components SET html_template =
       regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_primary_url\}\}".*?</a>)', '{{if .cta_primary_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.cta_secondary_url\}\}".*?</a>)', '{{if .cta_secondary_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_1_url\}\}".*?</a>)', '{{if .nav_link_1_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_2_url\}\}".*?</a>)', '{{if .nav_link_2_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_3_url\}\}".*?</a>)', '{{if .nav_link_3_url}}\1{{end}}', 'g'),
         '(<a[^>]*?href="\{\{\.nav_link_4_url\}\}".*?</a>)', '{{if .nav_link_4_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'site-head';

-- system-stats: 1 anchor(s), field(s): cta_url
UPDATE content_components SET html_template =
       regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_url\}\}".*?</a>)', '{{if .cta_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'system-stats';

-- tool-ai-agent-roi-estimator: 1 anchor(s), field(s): cta_url
UPDATE content_components SET html_template =
       regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_url\}\}".*?</a>)', '{{if .cta_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'tool-ai-agent-roi-estimator';

-- tool-archetype-taster-quiz: 1 anchor(s), field(s): result_cta_primary_url
UPDATE content_components SET html_template =
       regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.result_cta_primary_url\}\}".*?</a>)', '{{if .result_cta_primary_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'tool-archetype-taster-quiz';

-- tool-btl-investor_pre_037: 1 anchor(s), field(s): cta_link
UPDATE content_components SET html_template =
       regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_link\}\}".*?</a>)', '{{if .cta_link}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'tool-btl-investor_pre_037';

-- tool-equity-release_pre_037: 1 anchor(s), field(s): cta_link
UPDATE content_components SET html_template =
       regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_link\}\}".*?</a>)', '{{if .cta_link}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'tool-equity-release_pre_037';

-- tool-mortgage-affordability: 1 anchor(s), field(s): cta_button_url
UPDATE content_components SET html_template =
       regexp_replace(html_template,
         '(<a[^>]*?href="\{\{\.cta_button_url\}\}".*?</a>)', '{{if .cta_button_url}}\1{{end}}', 'g'),
     updated_at = now()
 WHERE name = 'tool-mortgage-affordability';

-- ── B. idioms the field parser cannot see ────────────────────────────────────
UPDATE content_components SET html_template = replace(html_template,
       '<a href="{{if .cta_primary_url}}{{.cta_primary_url}}{{else}}#{{end}}" class="brief-explanation__cta-primary">{{.cta_primary_label}}</a>',
       '{{if .cta_primary_url}}<a href="{{.cta_primary_url}}" class="brief-explanation__cta-primary">{{.cta_primary_label}}</a>{{end}}'), updated_at = now()
 WHERE name = 'brief-explanation';
UPDATE content_components SET html_template = replace(html_template,
       '<a href="{{if .cta_secondary_url}}{{.cta_secondary_url}}{{else}}#{{end}}" class="brief-explanation__cta-secondary">{{.cta_secondary_label}}</a>',
       '{{if .cta_secondary_url}}<a href="{{.cta_secondary_url}}" class="brief-explanation__cta-secondary">{{.cta_secondary_label}}</a>{{end}}'), updated_at = now()
 WHERE name = 'brief-explanation';

UPDATE content_components SET html_template = replace(html_template,
       '<a href="#" class="section__link">View all {{.category_name}} →</a>',
       '{{if .category_url}}<a href="{{.category_url}}" class="section__link">View all {{.category_name}} →</a>{{end}}'), updated_at = now()
 WHERE name = 'category-listing';

UPDATE content_components SET html_template = replace(html_template,
       '<a href="#" class="button button--secondary button--full-width">{{.tier_1_cta}}</a>',
       '{{if .tier_1_cta_url}}<a href="{{.tier_1_cta_url}}" class="button button--secondary button--full-width">{{.tier_1_cta}}</a>{{end}}'), updated_at = now()
 WHERE name = 'Pricing Tiers';
UPDATE content_components SET html_template = replace(html_template,
       '<a href="#" class="button button--primary button--full-width">{{.tier_2_cta}}</a>',
       '{{if .tier_2_cta_url}}<a href="{{.tier_2_cta_url}}" class="button button--primary button--full-width">{{.tier_2_cta}}</a>{{end}}'), updated_at = now()
 WHERE name = 'Pricing Tiers';
UPDATE content_components SET html_template = replace(html_template,
       '<a href="#" class="button button--secondary button--full-width">{{.tier_3_cta}}</a>',
       '{{if .tier_3_cta_url}}<a href="{{.tier_3_cta_url}}" class="button button--secondary button--full-width">{{.tier_3_cta}}</a>{{end}}'), updated_at = now()
 WHERE name = 'Pricing Tiers';

UPDATE content_components SET html_template = replace(html_template,
       '<a href="#" class="button button--primary">{{.read_more_text}}</a>',
       '{{if .read_more_url}}<a href="{{.read_more_url}}" class="button button--primary">{{.read_more_text}}</a>{{end}}'), updated_at = now()
 WHERE name = 'featured_article';

-- ── B2. the new renderer-owned url fields the gates in B refer to ────────────
UPDATE content_components SET input_schema = jsonb_set(input_schema, '{fields,category_url}', '{"type": "url", "source": "renderer", "required": false, "on_missing": "skip_field", "llm_guidance": "Never author this. The renderer resolves the category hub URL; if it cannot, the ''View all'' link is not rendered at all (LNK-005)."}'::jsonb), updated_at = now()
 WHERE name = 'category-listing' AND input_schema->'fields'->'category_url' IS NULL;

UPDATE content_components SET input_schema = jsonb_set(input_schema, '{fields,tier_1_cta_url}', '{"type": "url", "source": "renderer", "required": false, "on_missing": "skip_field", "llm_guidance": "Never author this. Without a resolved destination the tier button is not rendered (LNK-005)."}'::jsonb), updated_at = now()
 WHERE name = 'Pricing Tiers' AND input_schema->'fields'->'tier_1_cta_url' IS NULL;
UPDATE content_components SET input_schema = jsonb_set(input_schema, '{fields,tier_2_cta_url}', '{"type": "url", "source": "renderer", "required": false, "on_missing": "skip_field", "llm_guidance": "Never author this. Without a resolved destination the tier button is not rendered (LNK-005)."}'::jsonb), updated_at = now()
 WHERE name = 'Pricing Tiers' AND input_schema->'fields'->'tier_2_cta_url' IS NULL;
UPDATE content_components SET input_schema = jsonb_set(input_schema, '{fields,tier_3_cta_url}', '{"type": "url", "source": "renderer", "required": false, "on_missing": "skip_field", "llm_guidance": "Never author this. Without a resolved destination the tier button is not rendered (LNK-005)."}'::jsonb), updated_at = now()
 WHERE name = 'Pricing Tiers' AND input_schema->'fields'->'tier_3_cta_url' IS NULL;

UPDATE content_components SET input_schema = jsonb_set(input_schema, '{fields,read_more_url}', '{"type": "url", "source": "renderer", "required": false, "on_missing": "skip_field", "llm_guidance": "Never author this. The renderer resolves the featured post''s URL; without it the Read More button is not rendered (LNK-005)."}'::jsonb), updated_at = now()
 WHERE name = 'featured_article' AND input_schema->'fields'->'read_more_url' IS NULL;

-- ── C. class E fleet-wide: a URL field may never be llm AND required ─────────
UPDATE content_components c
   SET input_schema = jsonb_set(c.input_schema, '{fields}', (
         SELECT jsonb_object_agg(e.k,
                  CASE WHEN e.k ~ '(^|_)(url|link)$'
                        AND e.v->>'source' = 'llm'
                        AND e.v->>'required' = 'true'
                       THEN jsonb_set(e.v, '{required}', 'false')
                       ELSE e.v END)
           FROM jsonb_each(c.input_schema->'fields') e(k, v))),
       updated_at = now()
 WHERE c.is_active
   AND EXISTS (SELECT 1 FROM jsonb_each(c.input_schema->'fields') e(k, v)
                WHERE e.k ~ '(^|_)(url|link)$'
                  AND e.v->>'source' = 'llm' AND e.v->>'required' = 'true');

-- ── post-conditions ──────────────────────────────────────────────────────────
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM (VALUES
    ('Pricing Tiers','7f2ff1311c934699aa75fad4dda8d3ec'),
    ('ai-readiness-quiz','de220f86cb82ff863ee52d40f38e9ae1'),
    ('archetype-combinations','7f6be1e970b5de0c8d12ee2c9ff3e98e'),
    ('archetype-grid','b346b7b37cbfd779a6c7026d2ce02ad7'),
    ('archetype-result-card','4986338f104ef0e374033780d0037715'),
    ('bayesian-ranking-hero-tool_pre_037','38441fa1f337294ea45aa73ccdaa5352'),
    ('blog-listing_pre_037','08f72a046480ed643bfd70ced79a4aa8'),
    ('bridging-loan-calculator','36250282c5a19985604944634221f18c'),
    ('brief-explanation','375908366b39c79f4dc7d244ffbdcff1'),
    ('case-studies-grid','7fa101f9d4022fe1a3e3181badf9114c'),
    ('category-listing','01f694ad22a56bda632d010d7c66b69e'),
    ('content-sidebar','8a3a1f58fb0682c048e7a2d1096f4331'),
    ('featured-inventory','22f62c162c28cb248353a7d267761d61'),
    ('featured_article','90caba290e6c3a99138da8d929591fd2'),
    ('footer-with-disclaimer_pre_037','c366e8a5a4f41b04fc2458d9f1fad445'),
    ('funding-fit','0466193079f57df907da815af5359a68'),
    ('game-master-explanation','f9ff13ce17e9c9bc8ae055725cac4129'),
    ('gauntlet-cta','bfdbbfeebdb000e493f756c28f36d304'),
    ('header-docs','e4fa4c71126181efea959641119e9bab'),
    ('header-minimal-tool_pre_037','794774f829f5cdd89326f4c0ea7a1869'),
    ('header-with-cart-or-nav_pre_037','34e445e53c6d659d3bb27f2c5d93217b'),
    ('header-with-categories_pre_037','d979b7f7eed3934cb5118780d0edc2de'),
    ('header-with-search_pre_037','af620e5426107c02a3be809ae3fbbfbd'),
    ('patent-check','96a6718907806abb6d1ee3b1a08108e3'),
    ('product-hero_pre_037','4ab54a3412a4d98caaeb6bbbe32c6dbe'),
    ('product-specs','24d0de48619818649dfd810fd9da49ad'),
    ('provocation-feed','f8f109093ee87ccfb0bba030b0ebc566'),
    ('provocations-archive-list','3cabd9a230a49d7100c9e876f21c7ed4'),
    ('site-head','6487d58c864c4a9621893ac4d7b83921'),
    ('system-stats','1271578f484596a8252b751d632a2769'),
    ('tool-ai-agent-roi-estimator','17a0454c1215b91320ae1847d8f86728'),
    ('tool-archetype-taster-quiz','e294270fbf842b7ab62411efafd93aeb'),
    ('tool-btl-investor_pre_037','74dbdcad0f340aa5e6c030dfe447fbb3'),
    ('tool-equity-release_pre_037','453c885342cc7206c0fe05b01c13823e'),
    ('tool-mortgage-affordability','ca021040319ac9d917dedee52e56c9b6')
  ) AS t(name, md5_after)
  JOIN content_components c ON c.name = t.name AND md5(c.html_template) = t.md5_after;
  IF n <> 35 THEN
    RAISE EXCEPTION 'post-condition failed: % of 35 templates match the expected result', n;
  END IF;

  -- no active component has an llm+required URL field any more
  SELECT count(*) INTO n
    FROM content_components c, jsonb_each(c.input_schema->'fields') e(k, v)
   WHERE c.is_active AND e.k ~ '(^|_)(url|link)$'
     AND e.v->>'source' = 'llm' AND e.v->>'required' = 'true';
  IF n <> 0 THEN RAISE EXCEPTION 'post-condition failed: % llm+required url field(s) remain', n; END IF;

  -- the four new renderer-owned url fields exist and are optional
  SELECT count(*) INTO n
    FROM content_components c, jsonb_each(c.input_schema->'fields') e(k, v)
   WHERE (c.name, e.k) IN (('category-listing','category_url'),('Pricing Tiers','tier_1_cta_url'),
                           ('Pricing Tiers','tier_2_cta_url'),('Pricing Tiers','tier_3_cta_url'),
                           ('featured_article','read_more_url'))
     AND v->>'source' = 'renderer' AND v->>'required' = 'false';
  IF n <> 5 THEN RAISE EXCEPTION 'post-condition failed: % of 5 new url fields are renderer/optional', n; END IF;

  -- no template still carries the two repaired placeholder idioms
  SELECT count(*) INTO n FROM content_components
   WHERE is_active AND (html_template LIKE '%href="{{if .cta_primary_url}}%'
                     OR html_template LIKE '%href="{{if .cta_secondary_url}}%'
                     OR name IN ('category-listing','Pricing Tiers','featured_article')
                        AND html_template LIKE '%<a href="#"%');
  IF n <> 0 THEN RAISE EXCEPTION 'post-condition failed: % placeholder anchor(s) remain', n; END IF;
END $$;

INSERT INTO schema_migrations (filename, notes)
VALUES ('211_cta_gate_every_active_anchor.sql',
        'bugs_open/023 classes A/B/C/E, library-wide: gate 156 ungated CTA anchors across 31 active components so an unresolved destination renders no control (LNK-005); repair 7 placeholder-href anchors in 4 more components (brief-explanation, category-listing, Pricing Tiers, featured_article) adding 5 renderer/optional url fields; flip 23 llm+required URL fields to required:false fleet-wide so no schema compels a model to invent a URL. Every template change md5-gated before and after. Range-scoped item links (17 anchors/13 components) deliberately out of scope.');

COMMIT;

-- Post-apply verification (RUNBOOK R9b):
--   kubectl exec ... -tAc "SELECT jsonb_agg(jsonb_build_object('name',name,'function',function,'tpl',html_template))
--     FROM content_components WHERE is_active AND html_template IS NOT NULL;" > components.json
--   python3 scripts/parse_gates.py components.json    -- expect: UNGATED CTA worklist = 0
--   scripts/check_cta_gates.py                        -- the standing lint, expect exit 0


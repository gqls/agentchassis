-- 086b — /tools.html exists, the dead CTA blocks get URLs, the 404 stub is archived
--
-- Owner said do the loose ends (2026-08-03). Census first, per the 07-30b
-- handoff's "check all three before editing one":
--   llm-cost-calculator: hero-tool has cta_primary_label "Run the calculator"
--     and cta_secondary_label "Review the methodology", NO urls -> no buttons
--     render. tool-cta has primary_cta_label "Explore All Tools" and
--     secondary_cta_label "Learn how it works", NO urls -> same.
--   review-council-simulator: both blocks wired, but "Explore All Tools"
--     points at /multi-agent-review-council.html — a label/target mismatch
--     (that page is about the council, not a tools index).
--   model-approach-selector: carries NEITHER component (older page shape) —
--     nothing to fix there.
--   /tools.html: 404, no page row, while two tool pages' buttons and copy
--     promise a tools index. /tools/decision-record/index.html: page row
--     active, never deployed, 0 components, serves 404 — a stub.
--
-- Decisions (content, delegated by the owner's "do them all"):
--   "Run the calculator"    -> #input-tokens (the widget's first field, on-page)
--   "Review the methodology"-> /guides/llm-cost-calculator-guide.html
--   "Learn how it works"    -> /guides/llm-cost-calculator-guide.html
--
-- > CORRECTED 2026-08-03, ~15 min after apply: the two guide URLs above were
-- > WRONG to ship — /guides/llm-cost-calculator-guide.html has a page row
-- > (build_status 'planned', 0 components, since 07-25) and SERVES 404. The
-- > calculator's rerender completed 40s before the revert, so the 404 buttons
-- > were briefly live. Both url keys were removed again (label-without-url
-- > renders no button — today's exact state) and the page re-queued
-- > (item 086b-guide404-revert). The lesson, again: a *target* must be
-- > verified at the artefact before a link ships — the copy referencing a
-- > companion guide is not evidence the guide exists.
--   "Explore All Tools"     -> /tools.html, on both pages, and /tools.html is
--     built below: hero-tool + tool-cta reusing the calculator page's exact
--     component rows (install.py precedent), in_header=false so nav membership
--     is untouched (the 149 lane owns nav semantics; reachable via the tool
--     pages' buttons instead).
--   decision-record stub    -> status 'archived' (honest state; nothing links in)
--
-- KEY SPELLINGS ARE OPPOSITE BY COMPONENT (the landmine this lane filed):
-- hero-tool wants cta_primary_url / cta_secondary_url;
-- tool-cta wants primary_cta_url / secondary_cta_url. Both verified against
-- the live html_template rows before writing this file.

\set ON_ERROR_STOP on
BEGIN;

-- 1. calculator hero-tool: give both labels their URLs
UPDATE page_components pc
SET content_data = pc.content_data
      || '{"cta_primary_url": "#input-tokens",
           "cta_secondary_url": "/guides/llm-cost-calculator-guide.html"}'::jsonb
FROM pages p, sites s
WHERE pc.page_id = p.id AND p.site_id = s.id
  AND s.domain = 'fundamentallyai.com'
  AND p.url = '/tools/llm-cost-calculator.html' AND pc.slot_name = 'hero-tool';

-- 2. calculator tool-cta: URLs for both buttons, and the simulator joins the
--    items list (it had only the two older tools)
UPDATE page_components pc
SET content_data = jsonb_set(
      pc.content_data
        || '{"primary_cta_url": "/tools.html",
             "secondary_cta_url": "/guides/llm-cost-calculator-guide.html"}'::jsonb,
      '{items}',
      pc.content_data->'items' || '[{
        "url": "/tools/review-council-simulator.html",
        "name": "tool-review-council-simulator",
        "image": "",
        "title": "AI Review Council Simulator",
        "nav_label": "Tools / AI Review Council Simulator",
        "meta_description": "An interactive AI Review Council Simulator, free to run in the browser. Set the panel, the blocking threshold and the number of revision rounds, and see how often a change gets through. Calibrated on 362 real council runs."
      }]'::jsonb)
FROM pages p, sites s
WHERE pc.page_id = p.id AND p.site_id = s.id
  AND s.domain = 'fundamentallyai.com'
  AND p.url = '/tools/llm-cost-calculator.html' AND pc.slot_name = 'tool-cta'
  AND NOT (pc.content_data->'items') @> '[{"name":"tool-review-council-simulator"}]'::jsonb;

-- 3. simulator tool-cta: "Explore All Tools" now goes where the label says
UPDATE page_components pc
SET content_data = pc.content_data || '{"primary_cta_url": "/tools.html"}'::jsonb
FROM pages p, sites s
WHERE pc.page_id = p.id AND p.site_id = s.id
  AND s.domain = 'fundamentallyai.com'
  AND p.url = '/tools/review-council-simulator.html' AND pc.slot_name = 'tool-cta';

-- 4. the stub: archived, as it serves
UPDATE pages p SET status = 'archived'
FROM sites s
WHERE p.site_id = s.id AND s.domain = 'fundamentallyai.com'
  AND p.url = '/tools/decision-record/index.html'
  AND p.status = 'active'
  AND NOT EXISTS (SELECT 1 FROM page_components pc WHERE pc.page_id = p.id);

-- 5. /tools.html — the index the labels promise. Two sections, both reusing
--    the calculator page's exact component rows (no fork can be picked).
--    in_header/in_footer false: nav membership untouched.
INSERT INTO pages
  (site_id, name, url, title, page_type, status, meta_description,
   nav_label, nav_order, in_header, in_footer, sections, build_status, rebuild_policy)
VALUES ((SELECT id FROM sites WHERE domain = 'fundamentallyai.com'),
        'tools', '/tools.html', 'Free AI Tools | FundamentallyAI', 'tool', 'active',
        'Three interactive AI tools, free to run in the browser: an LLM provider cost calculator, a fine-tuning vs RAG vs prompting decision guide, and an AI review council simulator.',
        'Tools / Index', 203, false, false,
        '["hero-tool","tool-cta"]'::jsonb, 'pending', 'generic');

INSERT INTO page_components (page_id, component_id, position, slot_name, content_data, build_status)
VALUES (
  (SELECT p.id FROM pages p JOIN sites s ON s.id = p.site_id
    WHERE s.domain = 'fundamentallyai.com' AND p.url = '/tools.html'),
  (SELECT pc.component_id FROM pages p
     JOIN page_components pc ON pc.page_id = p.id
     JOIN sites s ON s.id = p.site_id
    WHERE s.domain = 'fundamentallyai.com'
      AND p.url = '/tools/llm-cost-calculator.html' AND pc.slot_name = 'hero-tool'),
  1, 'hero-tool',
  '{
    "badge_label": "Free Tools",
    "hero_headline": "Tools we built to answer our own questions",
    "hero_subheadline": "Three interactive tools, free to run in the browser, built to settle real engineering decisions on our own platform. Each one states its assumptions, and the companion guides set out the method so you can check the working.",
    "stat_one_value": "3", "stat_one_label": "Interactive tools",
    "stat_two_value": "2", "stat_two_label": "Companion guides",
    "stat_three_value": "0", "stat_three_label": "Sign-ups required",
    "cta_primary_label": "", "cta_primary_url": "",
    "cta_secondary_label": "", "cta_secondary_url": ""
  }'::jsonb, 'pending');

INSERT INTO page_components (page_id, component_id, position, slot_name, content_data, build_status)
VALUES (
  (SELECT p.id FROM pages p JOIN sites s ON s.id = p.site_id
    WHERE s.domain = 'fundamentallyai.com' AND p.url = '/tools.html'),
  (SELECT pc.component_id FROM pages p
     JOIN page_components pc ON pc.page_id = p.id
     JOIN sites s ON s.id = p.site_id
    WHERE s.domain = 'fundamentallyai.com'
      AND p.url = '/tools/llm-cost-calculator.html' AND pc.slot_name = 'tool-cta'),
  2, 'tool-cta',
  '{
    "eyebrow_label": "Free Tools",
    "headline": "Built for our own decisions first.",
    "description": "Every tool here started as a question we needed answered before committing to an architecture: what will the tokens cost, which adaptation approach fits, how strict should a review gate be. We publish them as they are, assumptions stated. Provider rates and measured figures move, so a tool can give an outdated answer — the guides show you how to check.",
    "tools_list_label": "All Tools",
    "empty_state_text": "More tools are currently in development.",
    "trust_note": "No account required · Open access",
    "primary_cta_label": "", "primary_cta_url": "",
    "secondary_cta_label": "Talk to us about AI tooling", "secondary_cta_url": "/contact.html",
    "items": [
      {
        "url": "/tools/llm-cost-calculator.html",
        "name": "llm-cost-calculator",
        "image": "",
        "title": "LLM Provider Cost Comparison Calculator",
        "nav_label": "Tools / LLM Provider Cost Comparison Calculator",
        "meta_description": "An interactive LLM Provider Cost Comparison Calculator, free to run in the browser. The companion guide sets out the method behind it, so you can check the working."
      },
      {
        "url": "/tools/model-approach-selector/index.html",
        "name": "tool-model-approach-selector",
        "image": "",
        "title": "Fine-Tuning vs RAG vs Prompting Decision Guide",
        "nav_label": "Tools / Fine-Tuning vs RAG vs Prompting Decision Guide",
        "meta_description": "An interactive Fine-Tuning vs RAG vs Prompting Decision Guide, free to run in the browser. The companion guide sets out the method behind it, so you can check the working."
      },
      {
        "url": "/tools/review-council-simulator.html",
        "name": "tool-review-council-simulator",
        "image": "",
        "title": "AI Review Council Simulator",
        "nav_label": "Tools / AI Review Council Simulator",
        "meta_description": "An interactive AI Review Council Simulator, free to run in the browser. Set the panel, the blocking threshold and the number of revision rounds, and see how often a change gets through. Calibrated on 362 real council runs."
      }
    ]
  }'::jsonb, 'pending');

-- 6. VERIFY BEFORE COMMIT — raises roll the whole thing back.
DO $CHK$
DECLARE
  hero_url text; cta_url text; sim_url text; n_items int; idx_items int;
  n_pc int; n_null_slot int; stub_status text;
BEGIN
  SELECT pc.content_data->>'cta_primary_url' INTO hero_url
  FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
  WHERE s.domain='fundamentallyai.com' AND p.url='/tools/llm-cost-calculator.html' AND pc.slot_name='hero-tool';
  IF hero_url IS DISTINCT FROM '#input-tokens' THEN
    RAISE EXCEPTION '086b: calculator hero cta_primary_url not set (%)', hero_url;
  END IF;

  SELECT pc.content_data->>'primary_cta_url', jsonb_array_length(pc.content_data->'items')
    INTO cta_url, n_items
  FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
  WHERE s.domain='fundamentallyai.com' AND p.url='/tools/llm-cost-calculator.html' AND pc.slot_name='tool-cta';
  IF cta_url IS DISTINCT FROM '/tools.html' OR n_items <> 3 THEN
    RAISE EXCEPTION '086b: calculator tool-cta wrong (url=%, items=%)', cta_url, n_items;
  END IF;

  SELECT pc.content_data->>'primary_cta_url' INTO sim_url
  FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
  WHERE s.domain='fundamentallyai.com' AND p.url='/tools/review-council-simulator.html' AND pc.slot_name='tool-cta';
  IF sim_url IS DISTINCT FROM '/tools.html' THEN
    RAISE EXCEPTION '086b: simulator Explore All Tools still points at %', sim_url;
  END IF;

  SELECT p.status INTO stub_status FROM pages p JOIN sites s ON s.id = p.site_id
  WHERE s.domain='fundamentallyai.com' AND p.url='/tools/decision-record/index.html';
  IF stub_status IS DISTINCT FROM 'archived' THEN
    RAISE EXCEPTION '086b: decision-record stub not archived (%)', stub_status;
  END IF;

  SELECT count(*), count(*) FILTER (WHERE pc.slot_name IS NULL), max(jsonb_array_length(pc.content_data->'items'))
    INTO n_pc, n_null_slot, idx_items
  FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
  WHERE s.domain='fundamentallyai.com' AND p.url='/tools.html';
  IF n_pc <> 2 OR n_null_slot <> 0 OR idx_items <> 3 THEN
    RAISE EXCEPTION '086b: /tools.html placement wrong (pc=%, null_slots=%, items=%)', n_pc, n_null_slot, idx_items;
  END IF;
END $CHK$;

COMMIT;

-- 7. rerenders, via the queue (RUNBOOK: page_rerender + page-rerender handler
--    re-renders from stored content_data, NO LLM; never needs_page).
--    CORRECTED at apply time: the RUNBOOK's recipe named a 'category' column
--    that no longer exists (the 154 routing-columns work moved the schema);
--    this is the form that actually inserted, copied from a live row's shape.
INSERT INTO site_work_items
  (site_id, item_type, item_key, status, severity, priority, summary, spec,
   handler_agent, source, created_by, attempt_count, max_attempts, created_at, updated_at)
SELECT p.site_id, 'page_rerender',
       'page_rerender:' || p.name || ':086b-cta-and-index',
       'triaged', 'medium', 30,
       'Republish ' || p.name || ': 086b — dead CTA urls set / tools index created',
       jsonb_build_object('domain','fundamentallyai.com','page_id',p.id,
                          'filename', ltrim(p.url,'/'), 'page_name', p.name),
       'page-rerender', 'operator:brochure_component_library',
       'operator:brochure_component_library', 0, 3, NOW(), NOW()
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE s.domain='fundamentallyai.com'
  AND p.url IN ('/tools/llm-cost-calculator.html','/tools/review-council-simulator.html','/tools.html');

-- Verify at the artefact once the queue drains:
--   curl -s https://fundamentallyai.com/tools.html | grep -c 'tool-cta-card'   (expect 3 cards)
--   curl -s https://fundamentallyai.com/tools/llm-cost-calculator.html | grep -c 'cta-btn'  (buttons exist now)
--   python3 scripts/probe_council_simulator.py --url https://fundamentallyai.com/tools/review-council-simulator.html

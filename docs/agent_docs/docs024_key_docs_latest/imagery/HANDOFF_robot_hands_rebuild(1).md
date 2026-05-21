# HANDOFF — robot-hands.com rebuild (2026-05-20)

## Why we're here

The imagery/icon pipeline work is COMPLETE (see TODO_imagery_followups.md). While
verifying that robot-hands.com renders the icons correctly, we found the icons
weren't showing — and the root cause is NOT the imagery pipeline. It's the site's
content layer: empty pages and schema-mismatched component content. Rerendering
can't fix it (rerender re-assembles existing content; it doesn't regenerate it).

robot-hands.com was an ADOPTED site. Decision: stop treating it as an adoption,
and REBUILD it as a fresh, properly-built site with a working content layer and
working tools — plus new scope: NEWS on the index page and a dedicated news
section.

## Evidence summary (full detail in TODO_imagery_followups.md)

Per-page component audit (site_id 00ff3af5-dad8-4770-9f70-3edc267a3c92):

- 10 pages with ZERO components (incl. tools[needs_rebuild], learning-center,
  pneumatic-vs-electric-grippers, gripper-payload-calculator, ...).
- Pages with components but NULL content: product-detail (4/0),
  gripper-catalog (6/0), matchmatrix (8/0), gripper-selection-guide (9/0).
- index features content uses wrong schema ({title,description} vs template's
  {icon,name,description}) → empty headings, no icons on the homepage.
- Build history: needs_content_page 50 complete + 4 needs_human_review;
  needs_page 53 complete — builds ran but left the content layer incomplete/
  schema-drifted, consistent with a partial adoption.

## The rebuild task — what needs scoping (next session starts here)

### 1. Verify the foundation before rebuilding
- **site_specs**: confirm identity / classification / briefing / strategy /
  roadmap are good and current for robot-hands.com. The rebuild's quality is
  bounded by these — garbage in, garbage out.
- **site plan**: the current site_plans row (is_current=true) + site_plan_pages
  + site_plan_sections + site_plan_imagery — confirm the page set, section
  layout, and imagery plan are what we want for the FRESH site (not adoption
  residue). The plan drives the build.
- Decide: re-plan from scratch (re-run build-site-planner) or fix the existing
  plan in place? Re-plan is cleaner for a true fresh build.

### 2. New scope: news
- Index page needs a news section + a dedicated news section/page.
- build-site-planner already supports "latest-news" on the homepage when
  classification.content_features.news_feed.recommended = true (RULE 11 in the
  planner prompt). Check whether that flag is set for robot-hands; if not, the
  plan won't include news. May need to set the classification flag OR add news
  explicitly to the roadmap/plan.
- Determine where news CONTENT comes from — is there a news/feed agent, or does
  news content get generated/curated? This is likely net-new and needs its own
  scoping (data source, generation, refresh cadence). Possibly related to the
  web-search/web-scrape adapters already in the system.

### 3. Rebuild mechanism
- This is a CONTENT REBUILD (regenerate page_components against current schemas),
  NOT a rerender. The path is page-build-handler via the build-dispatch-loop:
  insert site_work_items rows (item_type around needs_content_page / needs_page)
  that the dispatch loop claims and spawns page-build-handler for. The 081c notes
  in the transcript describe the exact work-item INSERT shape the gap-planner /
  build pipeline uses — reuse that, don't hand-roll a spawn.
- Order: plan (site-planner) → pages synced → content build (page-build-handler
  per page) → imagery (the pipeline we just fixed) → asset deploy → rerender/
  deploy → publish. Confirm the real pipeline order against the build-dispatch
  pipeline definition.
- Wire the Lucide validator (lucide_icons.go) into the content-generation step
  so features icons get valid Lucide names as content is (re)generated.

### 4. Tools must actually work
- "working tools" is explicit scope. The tool pages (payload calculator, cycle-
  time estimator, grip-force calculator, matchmatrix) need their JS/interactive
  content built and deployed, not just shells. Several tool pages currently have
  1 component or are empty. Confirm how tool components (component_level='tool',
  js_content) are built and that the deploy carries the JS (the 081b notes
  reference a deploy_page files_field fix for js_content — relevant here).

## Schema contract bug to fix during rebuild

The index features mismatch ({title,description} vs {icon,name,description}) means
SOME content was generated against an older features schema. During rebuild,
ensure content generation uses the CURRENT input_schema for each component
(features wants icon/name/description). If old content-gen prompts/templates
encode the old schema, they need updating, or the rebuild reproduces the mismatch.

## Carried-forward (from imagery work)

- **SECURITY (highest):** scrub + rotate STABILITY_API_KEY and BANANA_API_KEY
  (plaintext in logs; Banana on paid tier).
- Lucide validator wiring (above).
- Stability timeout 60→120s; provider circuit breaker; graph pipeline (future).

## Key IDs / facts

- robot-hands.com site_id: 00ff3af5-dad8-4770-9f70-3edc267a3c92
- build-site-planner: agent_definitions type='build-site-planner', workflow in
  default_config (NOT *_workflow columns), is_current single row version 1
- image-build-handler: type='image-build-handler', workflow in default_config
- Workflow JSON path convention: {workflow,steps,<step>,config,...}
- snapshot_agent('<type>', '<reason>') before mutating any agent default_config
- Repos: code github.com/gqls/agentchassis ; sites github.com/gqls/sites
  (robot-hands.com under robot-hands.com/ subdir; deploys to S3 via GitHub Action)
- Module: github.com/gqls/agentchassis ; registry docker.io/aqls

---

# AUDIT FINDINGS (2026-05-20) — verdict: PATCH, do not re-plan

## Foundation is sound; the BUILD was broken (not the plan)

robot-hands site_specs (one-row-per-aspect, is_current per aspect) are rich and
coherent — content_direction, design_intent, page_hero_briefs, full content-gap-
planner suite, all regenerated well past the adoption seed. The current site_plan
aspect has a clean 14-page structure with proper per-page sections.

Only adoption residue: two stale aspects — `design` (source=adoption) and
`structure` ({"pages":["index","product-detail"]}, the dead 2-page seed). Both
superseded in practice by design_intent / site_plan. Clean these up; keep the rest.

Proof the build (not plan) is broken: site_plan.pages[index].sections =
[hero, system-stats, features, brief-explanation, info-card-grid, tool-list,
call-to-action] (7 sections), but the BUILT index had only hero/features/tool-list/
cta and features was schema-broken. Build produced a SUBSET of planned sections
with malformed content. => rebuild content, don't re-plan.

## News: gaswholesalers is the reference implementation

gas site_id 5fe15466-4e2e-4ff2-981e-98c1b7074002 has news, added by enrichment
(NOT re-plan):
- classification source_agent = `evaluate_news_feed` wrote content_features.news_feed:
  { recommended:true, source_types:[rss,news_search,api_news], separate_page:true,
    vertical_keywords:[...], reason:"..." }
- site_plan source_agent = `news-section-addition` amended the plan, adding
  `latest-news` as a homepage SECTION (in index.sections, before call_to_action).

PATTERN to replicate for robot-hands: run/replicate evaluate_news_feed (sets the
classification flag with gripper/automation vertical_keywords) then
news-section-addition (amends site_plan + plan tables). Enrichment, not re-plan.

GAP found in gas: classification said separate_page:true, but NO dedicated news
page exists in the gas plan (news/blog-role page query returned 0 rows). So
news-section-addition added the homepage section only, not a separate page.
robot-hands wants BOTH index section AND a dedicated news section/page — so we
likely need MORE than gas got. Also UNCONFIRMED: where news CONTENT comes from
(the rss/news_search/api_news fetch agent that populates items) — must resolve
before the news section renders actual content rather than empty.

## REBUILD PLAN (crystallised)

1. Clean stale adoption aspects: supersede `design` and `structure` specs.
2. Add news via enrichment pattern (evaluate_news_feed -> news-section-addition),
   extended to also create a dedicated news section/page (gas only got the homepage
   section). Resolve the news-content source/fetch agent.
3. REBUILD content: regenerate page_components for all 14 pages against CURRENT
   schemas, producing ALL planned sections (not a subset), correct schema
   (features = {icon,name,description}). Wire Lucide validator into this step.
4. Build TOOLS properly (see hard requirement below).
5. Imagery via fixed pipeline; site_plan.image_prompts already defines logo + 7
   heroes. (imagery_direction is the industrial-photography one we gate OFF icons —
   confirms the contamination fix is load-bearing here.)
6. Deploy + verify renders correctly (the check that started this).

## HARD REQUIREMENTS (user, must-have for rebuild)

- **TOOLS MUST ACTUALLY WORK.** Every tool produced (matchmatrix, payload calc,
  cycle-time estimator, grip-force calc) must function — working interactive JS,
  not shells. Currently tool pages are empty/1-component. component_level='tool',
  js_content must be generated AND the deploy must carry the JS (081b deploy_page
  files_field fix is relevant).
- **LINKS TO TOOLS MUST WORK.** The index tool-list links and nav links to tools
  must resolve to the correct URLs. The rendered index had wrong/broken tool links
  (e.g. tool-list "Browse All Tools" href="" empty; tool card links to
  /tools/<x>/index.html that must exist). Verify every tool link target is a real
  deployed page. This is part of the working-tools acceptance.

## NEWS ENHANCEMENTS (user, explicitly NOT urgent — backlog)

a) **Price-aware filtering + short expiry.** Filter news for price-movement items
   (e.g. "today's gas prices", "oil price up/down"); for such time-sensitive items,
   EXPIRE after just 1-2 days (they date fast). Needs: a classifier/filter on
   fetched news items (is-this-a-price-item?), and a TTL/expiry mechanism on news
   items so stale price news drops off automatically. Per-site vertical (gas =
   gas/oil prices; robot-hands = whatever its vertical_keywords surface).

b) **News -> infographic.** Pick 1-2 news items, research the underlying subject a
   bit, and generate an INFOGRAPHIC about it. Ties into: the imagery pipeline
   (infographic kind), the research/web-search adapters, and possibly the future
   data-graph pipeline (FUTURE_data_graph_pipeline.md) if the infographic is
   data/chart-driven. Net-new workflow; backlog.

Both (a) and (b) are NICE-TO-HAVE. The working tools (above) are the gating
requirement for the rebuild.

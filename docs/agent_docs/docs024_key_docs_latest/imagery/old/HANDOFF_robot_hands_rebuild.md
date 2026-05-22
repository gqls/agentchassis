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

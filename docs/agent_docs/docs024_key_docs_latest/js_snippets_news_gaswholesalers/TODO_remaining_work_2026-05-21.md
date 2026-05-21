# TODO — Platform worklist (updated 2026-05-21)

Supersedes `TODO_remaining_work_2026-05-20.md`. Ordered by priority /
user-visibility. Detail docs referenced where they exist.

## Done — news rendering (2026-05-19/20)

- [x] News CSS deployed (styles.css with `.news-*` selectors)
- [x] index.html latest-news section rendering in new style, pulling news
- [x] news.html news-listing rendering
- [x] **Structural fix**: `page-rerender` deploy_page uses `files_field`
      so component JS assets (`/tools/assets/*.js`) deploy. Site-wide, all
      components, current + future. `006_news_feed_pipeline_addendum_rendering.md`.
- [x] `latest-news.js` + `news-listing.js` in git, paired to correct data files
- [x] Verified across sites — files_field fix is the reusable mechanism;
      ai-agent-orchestration.com news.html confirmed as a stale-render case
      (component linked, js_content present, 849 feed items) that a rerender
      fixes.

## Done — FAQ empty-items bug + prevention (2026-05-20/21)

- [x] Diagnosed: not a writer bug. Isolated build test (faq-only page)
      produced a populated 9-item accordion through the full pipeline.
      Cause = duplicate content surfaces (`generic-text-block` + `faq` on
      one page); writer filled the prose block, left the structured faq
      empty. `faq_empty_items_prevention_findings.md`,
      `page_content_creation_flow.md`.
- [x] Live gaswholesalers faq repaired via the pipeline (not by hand):
      sections corrected to `["hero","faq","call_to_action"]`, rebuilt,
      verified `q_count=10`, deployed, populated accordion live in git.
- [x] **Prevention deployed:**
  - [x] content-gap-planner prompt: removed hardcoded generic-text-block
        example, fixed add_to_page name, added no-pairing + use-function
        rules. (`fix1_content_gap_planner_prompt.sql`) — confirmed live.
  - [x] site-planner prompt: function-first component list, faq/pricing
        mappings, no-pairing rule, hyphen CTA. (`fix2_site_planner_prompt.sql`)
        — confirmed live.
  - [x] `validate_components` implemented in `ValidateSitePlanAction`
        (was a dead flag) + resolver reused in `applyNewPage` and
        `applyAddToPage`; archetype-aware `defaultSectionsForPage`.
        (`v3_site_actions.go`, `apply_gap_plan_action.go`,
        `validate_components_implementation.md`,
        `planner_prompt_fixes_defect1.md`) — **deployed in chassis v1.0.1029**.
- [x] Debugging guide addendum (the diagnostic method) +
      planner-depth/stale-plan concerns documented.
      (`016_debugging_guide_addendum_faq_diagnosis.md`,
      `site_planner_depth_and_freshness_concerns.md`)
- [x] faq-test throwaway page deleted from DB + git.

## Open — verification in flight

- [ ] **ai-agent-orchestration.com site-wide rerender** (30+ pages, batch
      queued 2026-05-21 10:07). news.html = work item
      `d2239c4b-76a4-4450-b91e-65131a1e9a36`, was `triaged`. Let it run its
      course. Confirm news.html lands `file_count=2`
      (`/news.html` + `/tools/assets/news-listing.js`) and renders the
      listing. Watch the batch drains cleanly (load-tests the fragile bits
      below).

## Open — structural (not blocking, flagged)

- [ ] **Consumer-group fix — HELD BACK.** Chassis response consumer uses
      per-pod `a.AgentID` (agent.go ~3152) with FirstOffset → backlog
      replay on pod restart. Structurally-correct fix ready
      (agent-type-scoped stable group for non-job topics). Deploy with a
      chassis rebuild + topic resets, scheduled — NOT a fire-fight. Not the
      cause of past symptoms (those were OOM).
- [ ] **collected_data 18MB bloat → OOM.** component-quality-auditor
      orchestrations hold ~18MB, OOM-kill pods at 512Mi, causing
      phantom-complete + parent timeout cascades. `TODO_orchestration_memory_bloat.md`.
      Short-term: raise mem + GOMEMLIMIT. Real fix: find dominant field
      (likely `__raw_message__` dup + unbounded processing_history) and
      release after consumption.
- [ ] **needs_section_data review items on successful builds.** Even the
      clean faq-test build spawned a `needs_section_data` item
      (`needs_human_review`, no handler_agent) as a child, despite the faq
      populating correctly. Understand why a successful structured build
      raises it; relates to debugging guide's
      "needs_section_data → wont_fix" pattern.
- [ ] **Planner depth: per-section briefs + stale-plan write-back.**
      site_plan sections are bare strings (no briefs); gap-planned pages
      aren't written back to site_plan (faq was absent from the plan
      entirely). Consumer for briefs already exists
      (`plan_sections.sectionDescription`); planner needs to emit them.
      `apply_gap_plan` should append new pages to site_plan.
      `site_planner_depth_and_freshness_concerns.md`.
- [ ] **Post-build validation (Fix D):** assert a component whose
      input_schema declares a required structured field (faq.questions
      min_items 3) actually has it populated before deploy. Catches the
      empty-structured-component class regardless of planner.
- [ ] **data_sources enforcement + inline-small-js_content** — component
      data-file paths are convention, not enforced.
      `component_asset_pipeline_concerns.md`.
- [ ] **Reaper doc correction** — reapers are scheduled_tasks SQL
      pre_query, not Go code as first drafted.
      `reapers_and_stuck_state_recovery.md` needs the framing fix.
- [ ] **css_snippets lemma-aware matching** (singular vs plural).
      `css_snippets_matching_known_issue.md`.
- [ ] **Decouple js_snippets refresh from refresh_site_components** in
      rerender-pages (3 independent flags). `rerender_pages_workflow_findings.md`.

## Open — site polish (user-visible)

- [ ] **logo.png 404** on gaswholesalers — `/assets/images/logo.png`
      referenced in header, returns 404. Generate the asset or switch to
      text logo.
- [ ] **favicon.ico 404** — no favicon deployed.
- [ ] **Multi-file commit message** shows `Rerender: ` (empty filename) for
      2-file commits. Cosmetic. Fix `buildCommitMessage` for `fileCount>1`
      (e.g. `Rerender: index.html (+1 asset)`).
- [ ] **tool-gas-unit-converter** (gaswholesalers, page_id
      `7e576bc4-...`) — third js_content page, not yet post-fix rerendered.
- [ ] wholesale-pricing-explained.html — no components, gets skipped.
- [ ] fuel-pricing-framework.html — flagged earlier.

## Reference docs (in /mnt/user-data/outputs/)

FAQ work: `faq_empty_items_prevention_findings.md`,
`page_content_creation_flow.md`,
`016_debugging_guide_addendum_faq_diagnosis.md`,
`site_planner_depth_and_freshness_concerns.md`,
`validate_components_implementation.md`, `planner_prompt_fixes_defect1.md`,
`fix1_content_gap_planner_prompt.sql`, `fix2_site_planner_prompt.sql`,
`apply_gap_plan_action.go`, `v3_site_actions.go`.

News work: `006_news_feed_pipeline_addendum_rendering.md`,
`component_asset_pipeline_concerns.md`, `rerender_pages_workflow_findings.md`.

Platform: `TODO_orchestration_memory_bloat.md`,
`reapers_and_stuck_state_recovery.md`, `css_snippets_matching_known_issue.md`.

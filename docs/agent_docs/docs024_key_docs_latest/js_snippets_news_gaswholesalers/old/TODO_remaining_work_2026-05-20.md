# TODO — Remaining work after gaswholesalers news rollout (2026-05-20)

Status snapshot after the news section was fully deployed via the framework.
Ordered roughly by priority / user-visibility. Detail docs referenced where
they exist.

## Done this session

- [x] News CSS deployed (styles.css with `.news-*` selectors)
- [x] index.html latest-news section rendering in new style, pulling news
- [x] news.html news-listing rendering
- [x] **Structural fix**: `page-rerender` deploy_page now uses `files_field`
      so component JS assets (`/tools/assets/*.js`) deploy. Applies to all
      components site-wide, current and future. See
      `006_news_feed_pipeline_addendum_rendering.md`.
- [x] `latest-news.js` + `news-listing.js` in git, paired to correct data files

## Next up (user requested: faq empty items)

- [ ] **faq.html empty FAQ items.** The `.faq-item` summary/answer elements
      render empty (`<summary></summary><p></p>`). This is content, not
      rendering — needs page-rebuild (LLM content regeneration) for the FAQ
      component, NOT page-rerender. The faq page has a populated intro
      section but the structured FAQ Q&A items have no content_data.
      Investigate: is the FAQ component's content_data empty in
      page_components, or is the template failing to bind populated data?
      Determine root cause before triggering rebuild.

## News-related follow-ups

- [ ] **tool-gas-unit-converter** — third page with non-empty `js_content`,
      not yet rerendered post-fix. Trigger a single-page rerender to deploy
      its JS asset (`/tools/assets/tool-gas-unit-converter.js`) if it has one.
      page_id `7e576bc4-fb8b-46a4-b035-2842c481f35a`.
- [ ] **Multi-file commit message** shows `Rerender: ` with empty filename
      for 2-file news commits. Cosmetic. Fix `buildCommitMessage` to handle
      `fileCount > 1` (e.g. `Rerender: index.html (+1 asset)`).
- [ ] **data_sources enforcement** — components hardcode their data file
      path in js_content; nothing verifies the file exists. Populate
      `content_components.data_sources` and validate/auto-stub at deploy.
      See `component_asset_pipeline_concerns.md`.
- [ ] **Inline small js_content** (<5KB) into rendered HTML rather than
      external file, to remove the file-must-exist coupling. Reserve
      external files for large payloads. See
      `component_asset_pipeline_concerns.md`.

## Site polish (gaswholesalers, user-visible)

- [ ] **logo.png 404** — `/assets/images/logo.png` returns 404. Header
      references it (`<img src="/assets/images/logo.png">`). Either generate
      the logo asset or update the header to a text-only logo.
- [ ] **favicon.ico 404** — no favicon deployed.
- [ ] fuel-pricing-framework.html footer link / content (was flagged as
      0-components / planned earlier in the session)
- [ ] wholesale-pricing-explained.html — has no components, gets skipped on
      rerender (`skipped: true, reason: no components found`). Either it
      needs content built or the footer link should be removed.

## Platform structural (not blocking news, but flagged)

- [ ] **collected_data 18MB bloat → OOM kills.** component-quality-auditor
      orchestrations hold 18MB collected_data, OOM-killing pods at 512Mi.
      Causes phantom-completed orchestrations and cascading parent timeouts.
      See `TODO_orchestration_memory_bloat.md`. Short-term: raise mem limit,
      add GOMEMLIMIT. Real fix: identify dominant collected_data field
      (likely `__raw_message__` duplication + unbounded processing_history)
      and release after consumption.
- [ ] **Consumer group bug** (chassis response consumer uses `a.AgentID`
      per-pod UUID at agent.go ~line 3152, with FirstOffset). Causes backlog
      replay on every pod restart for long-running pods. Adapters already
      use stable named groups; chassis is the outlier. Structurally correct
      fix ready (agent-type-scoped stable group for non-job topics). Deploy
      with chassis rebuild + topic resets, NOT as a fire-fight. Not the cause
      of current symptoms — those were OOM-driven.
- [ ] **Reaper doc correction.** The reapers are SQL `pre_query` entries in
      the `scheduled_tasks` table, NOT Go code as my first doc draft assumed.
      Confirmed reapers: `stale-orchestration-reaper` (180s — fails
      build-dispatch-loop AWAITING_RESPONSES >30min, anything >90min, expires
      awaited_requests >5min past timeout), `stuck-task-reaper` (300s —
      resets stuck scheduled_tasks), `stale-work-item-reaper` (3600s —
      triaged >48h → unresolved), `claimed-item-timeout` (need to capture its
      pre_query). Correct `reapers_and_stuck_state_recovery.md` to reflect
      that these are scheduled-task SQL, with the Go 5-min on-access
      stuck-orchestration check as a secondary mechanism. Note: today's 51-min
      stuck claim cleared via the scheduled stale-orchestration-reaper, not
      the Go path.
- [ ] **page_components.rendered_html staleness** — the DB rendered_html for
      news/index is from 2026-05-14 and was NOT updated by today's rerenders
      (which committed straight to git). Also observed: work_items showing
      `updated_at = created_at` despite their `result` being populated hours
      later — a missing updated_at bump on result write. Two separate
      DB-freshness oddities worth a closer look; neither blocks rendering
      since git is the source of truth for deployed output.

## Reference docs produced this session

- `006_news_feed_pipeline_addendum_rendering.md` — rendering layer + files_field fix
- `component_asset_pipeline_concerns.md` — external-script + data-file enforcement gaps
- `rerender_pages_workflow_findings.md` — blog-listing/news-index gap, JS-coupling, Migration C clarification
- `reapers_and_stuck_state_recovery.md` — needs correction per above
- `TODO_orchestration_memory_bloat.md` — the 18MB / OOM investigation

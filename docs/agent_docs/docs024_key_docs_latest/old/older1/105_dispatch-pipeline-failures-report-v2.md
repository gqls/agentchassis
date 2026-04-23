# Build Pipeline Failures & Unresolved Items Report

**Date:** 2026-04-14
**Source:** `site_work_items` where `pipeline = 'build'` and `status != 'complete'`
**Total items:** 57

---

## Priority Fixes

| # | Issue | Impact | Items unblocked | Status |
|---|-------|--------|-----------------|--------|
| P1 | `component_id` nil bug in `load_tool` step | Pipeline bug | 5 `improve_tool` items | **FIX WRITTEN** — 3 files: `work_item_actions.go`, `create_work_item_action.go`, `fix_p1_tool_auditor_component_id.sql` |
| P2 | Fork deploy fails on retry after partial failure | Pipeline bug | 1 `add_tool` item + retry resilience for all future forks | **FIX WRITTEN** — `deploy_tool_action.go`: two-stage fork check handles orphaned forks from partial failures. Also: `component_library.go`: `GetComponentByFunction` excludes forks (defensive) |
| P3 | Rate limit / usage limit errors (1,869 in 14 days) not classified as transient | Error classification bug | webdesign-agent + content-gap-planner + all LLM agents | **FIX WRITTEN** — `ai_errors.go`: adds 429, rate_limit, usage limit patterns to `isAIUnavailable`. `ai_actions.go`: fix catch-all comment |
| P4 | vonc.com serving raw CSS instead of page content | Upstream data/template | 3 `wont_fix` + potentially 10 `needs_section_data` | **NEEDS INVESTIGATION** |
| P5 | Sections with open data requests re-sent to content writer — wasted LLM calls and good sections lost to whole-page validation failure | Feedback loop bug | 28 `empty_section` + all future builds with data-dependent sections | **FIX WRITTEN** — `plan_sections_action.go`: checks for open `needs_section_data` items before planning; sections with pending data requests are deferred, allowing remaining sections to build successfully |
| P6 | Content rewrites failing validation (542 in 14 days) | Content quality | 5 items in `needs_human_review` | **NEEDS INVESTIGATION** — surface specific blocker reasons |
| P7 | `needs_section_data` — auto-fill vs client input | Content gap | 25 items | **NEEDS TRIAGE DECISION** — which sections can be auto-generated vs need real data |

---

## 1–5. `improve_tool` | failed (5 items)

**Common error:** `step load_tool failed: failed to execute action query_database: query param path 'input_data.component_id' resolved to nil (code: CHILD_ORCHESTRATION_...)`

**Root cause:** The `component_id` is nil by the time `load_tool` runs. Either the work item was created without `component_id` in its input data, or the orchestration step that resolves it is failing silently. This is a single pipeline bug — fix the `component_id` resolution and all 5 can be retried.

| # | Domain | Summary |
|---|--------|---------|
| 1 | gaswholesalers.com | Radio inputs in .unit-toggle and .period-toggle not working |
| 2 | leopardessconsulting.co.uk | Tool depends exclusively on CSS custom properties |
| 3 | finetuning.uk | showResults() function is truncated |
| 4 | gaswholesalers.com | clearError() called with wrong group |
| 5 | leopardessconsulting.co.uk | 'stateless' logic differs between computeScore |

---

## 6. `add_tool` | failed | gamedesign.uk

**Summary:** Bayesian Ranking Calculator

**Error:** `deploy_tool_to_site: fork tool: ERROR: duplicate key value violates unique constraint "content_compo..."`

**Root cause:** The constraint is `content_components_name_key` — a global `UNIQUE(name)`. A previous deploy attempt created the fork (row `2fd8a2f7`, name `tool-bayesian-ranking-gamedesign-uk`) but failed at a later step (page creation or page_component linking). The "already deployed" check only looks through `page_components → pages` — so the orphaned fork is invisible. The retry attempts to INSERT another fork with the same name → unique violation.

**Fix:** `deploy_tool_action.go` — restructured to a two-stage check:
1. Check `content_components` directly for an existing fork (by `forked_from + name`) — catches orphans
2. If found, check if fully deployed (fork linked to a page via page_component)
   - Fully deployed → return early (as before)
   - Fork exists but not linked → reuse existing fork ID, continue with page creation
   - No fork exists → create one (as before)

Steps 4–7 (page creation, page_component, work items) were already idempotent (`ON CONFLICT` handling), so the full deploy flow is now retry-safe.

**Defensive improvement:** `component_library.go` — `GetComponentByFunction` now adds `AND forked_from IS NULL` to prevent function-based lookups from returning site forks instead of library templates.

---

## 7–8. Timeouts (2 items)

**Common error:** `Claim timed out (attempts exhausted)`

| # | Type | Domain | Summary |
|---|------|--------|---------|
| 7 | content_rewrite | robot-hands.com | Add Gripper ROI & Payback Period Estimator tool ref |
| 8 | audit_tool | gaswholesalers.com | LLM code review: Fuel Cost Estimator |

**Analysis:** Could be transient (LLM call took too long, queue backed up) or structural (task payload too large — full tool source for audit, complex page for rewrite). Need to check: are these consistently timing out on retry, or was it a one-off? If the former, the timeout window or chunking strategy may need adjusting.

---

## 9. `needs_rerender` | detected | gamedesign.uk

**Summary:** Rerender after template fix

**Analysis:** Status is `detected`, suggesting the pipeline noticed it needs work but hasn't picked it up yet. Is there a worker that processes `detected` → `in_progress`? Or is this stuck because nothing is polling for that status?

---

## 10. `needs_rerender` | unresolved | robot-hands.com

**Summary:** Stale 48h+. Same type as #9 but has been sitting longer.

**Analysis:** If the rerender pipeline isn't picking these up automatically, both this and #9 need the same fix.

---

## 11–12. `generic_theme` | unresolved (2 items)

| # | Domain | Notes |
|---|--------|-------|
| 11 | ai-agent-orchestration.com | Stale 48h+, unresolved after 2 attempts |
| 12 | robot-hands.com | Stale 48h+, unresolved after 2 attempts |

**Analysis:** Theme generation has been tried twice and failed both times. What does the theme pipeline actually do when it runs? Is there an error log from those attempts, or does it just silently fail and increment the attempt counter? These need someone to look at what the attempts actually produced/errored on.

---

## 13–14. `audit_finding_audience` | unresolved | ai-agent-orchestration.com (2 items)

| # | Summary snippet |
|---|----------------|
| 13 | "Conten..." (content-related audience finding) |
| 14 | "Target..." (target audience finding) |

**Analysis:** Stale 48h+. Audit findings about audience targeting. What's the expected resolution path — does someone need to rewrite content to better target the audience, or is there an automated step that should be handling this? If automated, it's broken. If manual, these should be surfaced more prominently.

---

## 15. `needs_blog_posts` | unresolved | ai-agent-orchestration.com

**Analysis:** Stale 48h+. Blog posts need to be created. Almost certainly a human-gated task — someone needs to write or commission the posts. Unless there's an automated blog generation pipeline that should be picking this up?

---

## 16. `needs_internal_links` | unresolved | robot-hands.com

**Analysis:** Stale 48h+. Internal linking pass needed. This could potentially be automated (scan pages, find relevant anchor text, insert links). If there's a pipeline for this, it's not running. If it's manual, it's been forgotten.

---

## 17–21. `empty_section` (5 items)

### 17–20: unresolved (stale 48h+)

| # | Domain |
|---|--------|
| 17 | robot-hands.com |
| 18 | robot-hands.com |
| 19 | finetuning.uk |
| 20 | gaswholesalers.com |

**Analysis:** Empty sections detected but never filled. Closely related to `needs_section_data` items — the section exists in the template but has no content. Resolution is either to provide content or remove the section. Are these supposed to auto-escalate to `needs_section_data` work items, or are they a separate track?

### 21: needs_human_review | gaswholesalers.com

**Summary:** Empty section 'generic-text-block' on page how-pricing-works

**Error:** `content validation failed: 1 blockers, 0 errors`

**Analysis:** This one progressed further than the others — it got to validation but failed. So someone or something attempted to fill it, but the content didn't pass validation. What was the blocker?

---

## 22–23. `audit_tool` | unresolved (2 items)

| # | Domain |
|---|--------|
| 22 | robot-hands.com |
| 23 | gaswholesalers.com |

**Analysis:** Both stale 48h+. LLM code reviews that were triaged but never completed. Different from item #8 which timed out — these never even got to execution. Is the `audit_tool` worker not picking up `unresolved` status items?

---

## 24–28. `content_rewrite` | needs_human_review (5 items)

All failed at the `validate_content` step:

| # | Domain | Summary | Validation |
|---|--------|---------|------------|
| 24 | finetuning.uk | Services section footer lists 'Our Services' with... | 1 blocker, 0 errors |
| 25 | leopardessconsulting.co.uk | Email 'leopardess@contactforsales.com'... | 1 blocker, 2 errors |
| 26 | leopardessconsulting.co.uk | About page claims depth in hierarchical multi-... | 1 blocker, 0 errors |
| 27 | vonc.com | Homepage (index) is completely empty | 1 blocker, 0 errors |
| 28 | ai-agent-orchestration.com | Add AI Readiness Quiz tool ref to the-enterp... | 1 blocker, 0 errors |

**Analysis:** These ran through the pipeline and produced content, but it didn't pass validation. Key question: what are the specific blockers? Are they fixable by the pipeline on retry with better prompting, or structural issues that genuinely need a human?

Notable: #25 (leopardess email) sounds like it flagged a placeholder/fake email address — that's a legitimate content problem someone needs to decide on.

---

## 29–31. `content_rewrite` | wont_fix | vonc.com (3 items)

| # | Summary |
|---|---------|
| 29 | About page contains only raw CSS/style code |
| 30 | About page is rendering raw CSS and style tag source |
| 31 | The only content sample retrieved is raw CSS code |

**Analysis:** All three are about the same underlying problem: vonc.com's about page is serving raw CSS instead of rendered content. The pipeline correctly identified it can't rewrite what isn't there and marked `wont_fix`. This is an upstream issue — the page template or data is broken. Fixing the page itself would likely resolve all three, plus potentially unblock some of the `needs_section_data` items for vonc.com.

---

## 32–56. `needs_section_data` | needs_human_review (25 items)

Each is a section in a page template that needs real content provided.

### vonc.com (10 items)

| # | Section | Page |
|---|---------|------|
| 32 | tool-archetype-taster-quiz | archetype-* |
| 33 | pricing | membership |
| 34 | archetype-combinations | archetypes |
| 35 | archetype-grid | archetypes |
| 36 | archetype-result-card | gauntlet |
| 37 | gauntlet-interface | gauntlet |
| 38 | provocation-feed | provocations |
| 39 | gauntlet-cta | index |
| 40 | lobby-grid | index |
| 41 | gauntlet-cta / game-master-explanation / platform-comparison | about |

### gaswholesalers.com (5 items)

| # | Section | Page |
|---|---------|------|
| 42 | pricing | how-pricing-works |
| 43 | use-cases-list | fuel-supply-by-industry |
| 44 | case-studies-list | client-case-studies |
| 45 | use-cases-list | who-we-serve |
| 46 | contact-info | contact |

### robot-hands.com (3 items)

| # | Section | Page |
|---|---------|------|
| 47 | case-studies-grid | learning-center |
| 48 | use-cases-list | gripper-selection-guide |
| 49 | leadership-team | about |

### ai-agent-orchestration.com (3 items)

| # | Section | Page |
|---|---------|------|
| 50 | tool-ai-readiness-quiz | ai-readiness-quiz |
| 51 | content-block-leadership | about |
| 52 | content-block-about | about |

### leopardessconsulting.co.uk (1 item)

| # | Section | Page |
|---|---------|------|
| 53 | portfolio-showcase | for-engineering-teams |

### finetuning.uk (1 item)

| # | Section | Page |
|---|---------|------|
| 54 | tool-password-entropy | password-entropy |

### gamedesign.uk (1 item)

| # | Section | Page |
|---|---------|------|
| 55 | content-block-about | about |

### Triage notes

- **Definitely need real client data:** contact-info (#46), leadership-team (#49), pricing (#33, #42), portfolio-showcase (#53), content-block-leadership (#51)
- **Could potentially auto-generate with reasonable defaults:** use-cases-list (#43, #45, #48), case-studies-list (#44), case-studies-grid (#47)
- **Need specific tool/feature parameters:** tool-archetype-taster-quiz (#32), tool-ai-readiness-quiz (#50), tool-password-entropy (#54), gauntlet-interface (#37)
- **Content structure decisions needed:** archetype-grid (#35), archetype-combinations (#34), provocation-feed (#38), lobby-grid (#40)

---

## Breakdown by Domain

| Domain | Total | Failed | Unresolved | Needs Human Review | Wont Fix | Detected |
|--------|-------|--------|------------|-------------------|----------|----------|
| vonc.com | 15 | 0 | 0 | 11 | 3 | 0 |
| gaswholesalers.com | 10 | 3 | 2 | 4 | 0 | 0 |
| robot-hands.com | 10 | 1 | 5 | 3 | 0 | 0 |
| ai-agent-orchestration.com | 8 | 0 | 4 | 4 | 0 | 0 |
| leopardessconsulting.co.uk | 5 | 2 | 0 | 3 | 0 | 0 |
| finetuning.uk | 4 | 1 | 1 | 2 | 0 | 0 |
| gamedesign.uk | 3 | 1 | 0 | 1 | 0 | 1 |

## Breakdown by Status

| Status | Count |
|--------|-------|
| needs_human_review | 28 |
| unresolved | 15 |
| failed | 8 |
| wont_fix | 3 |
| detected | 1 |
| **Total** | **57** | <!--note: 2 items are shared rows in vonc about page, counts approximate-->

## Breakdown by Item Type

| Item Type | Count |
|-----------|-------|
| needs_section_data | 25 |
| content_rewrite | 8 |
| empty_section | 5 |
| improve_tool | 5 |
| audit_tool | 3 |
| generic_theme | 2 |
| audit_finding_audience | 2 |
| needs_rerender | 2 |
| add_tool | 1 |
| needs_blog_posts | 1 |
| needs_internal_links | 1 |
| **Total** | **57** | <!--note: item 41 covers multiple sections on vonc about page-->

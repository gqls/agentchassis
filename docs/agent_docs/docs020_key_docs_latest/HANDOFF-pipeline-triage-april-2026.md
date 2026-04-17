# Pipeline Triage Handoff — April 2026

## Session Summary

Triaged 57 build pipeline failures and unresolved items across 7 sites. Diagnosed root causes, wrote and deployed 7 code fixes (P1–P5, P9 across 12 files), performed operational resets (P6), documented a deferred architectural change (P10), and reduced the active failure queue to human-gated items only.

---

## What Was Deployed

### P1: component_id nil bug in tool-improver pipeline

**Problem:** `tool-auditor` creates `improve_tool` work items, but `component_id` never reached the `site_work_items` row because `create_work_item` didn't support setting it. `tool-improver` then fails at `load_tool` with `input_data.component_id resolved to nil`.

**Files changed:**
- `work_item_actions.go` — added `componentID *uuid.UUID` to `workItem` struct + INSERT statement
- `create_work_item_action.go` — added `component_id` to Optional inputs, parse + pass through (same pattern as existing `page_id`)
- `fix_p1_tool_auditor_component_id.sql` — promoted `component_id` to top-level config in tool-auditor's `create_improve_item` and `create_review_item` workflow steps

**Items unblocked:** 5 `improve_tool` failures (reset to triaged, old items marked `wont_fix` as superseded)

### P2: Fork deploy fails on retry after partial failure

**Problem:** `deploy_tool_action` creates a fork component (step 3) then creates a page and page_component (steps 4–5). If steps 4–5 fail, the fork is orphaned. The "already deployed" check (step 2) only looks through `page_components → pages`, so orphaned forks are invisible. Retry hits `content_components_name_key` unique constraint.

**Files changed:**
- `deploy_tool_action.go` — restructured to two-stage check: (a) check `content_components` directly for existing fork by `forked_from + name`, (b) if found, check if fully deployed via page_component. Reuses orphaned forks instead of creating duplicates.
- `component_library.go` — `GetComponentByFunction` adds `AND forked_from IS NULL` (defensive: function-based lookups should never return site forks)

**Items unblocked:** 1 `add_tool` failure (gamedesign.uk Bayesian Ranking Calculator — turned out to already be complete from an earlier successful run). Fix prevents future failures for any tool fork across any site.

### P3: Rate limit errors not classified as transient

**Problem:** API 429 rate limits and usage limit 400 errors fell through `isAIUnavailable()` into the catch-all, counting as failed attempts. 1,869 wasted attempts in 14 days across `webdesign-agent` and `content-gap-planner`.

**Files changed:**
- `ai_errors.go` — added `status 429`, `rate_limit`, `rate limit`, `too many requests`, `usage limit`, `billing` patterns to `isAIUnavailable()`
- `ai_actions.go` — fixed misleading catch-all comment (removed "429" from the list of codes it claims to catch)

**Impact:** Items hitting rate limits now go back to `triaged` without counting attempts, same as connection refused or auth failures. Prevents ~130 wasted attempts/day.

### P4: load_page_record fails when page_name not in spec

**Problem:** `load_page_record` required `page_name` in the work item spec. Work items created without `page_name` (e.g. manually created, or from agents that only set the `page_id` column) failed immediately. Every `needs_content_page` item has `page_id` as a column on `site_work_items`, but the action couldn't use it.

**Files changed:**
- `load_page_record_action.go` — `page_name` moved from Required to Optional, `page_id` added as Optional. Lookup: page_name first, page_id fallback. Non-page names (`"new page needed"`, `"site-wide"`) fall through to page_id if available instead of returning `{found: false}`.
- `fix_load_page_record_page_id.sql` — adds `"page_id": "input_data.page_id"` to the `load_page_record` step config in page-build-handler workflow

**Items unblocked:** 3 vonc.com `needs_content_page` failures (archetypes, gauntlet, provocations — reset to triaged after deploy)

### P5: Section data feedback loop (plan-then-reconcile)

**Problem (v1):** Sections with pending `needs_section_data` items were re-sent to the content writer on every build cycle, wasting LLM calls.

**Problem (v2, discovered during implementation):** The initial pre-check approach would have permanently blocked sections even after their components were created. vonc.com had 12 `needs_section_data` items created when components didn't exist yet — the components were created later but the stale data requests remained open.

**Files changed:**
- `plan_sections_action.go` — plan-then-reconcile approach:
  1. Load open `needs_section_data` items for this page (one query)
  2. Plan every section normally (no shortcuts)
  3. After planning: ready sections with open requests → auto-close the stale request. Deferred sections with no open request → create new request. Deferred sections with existing request → skip (no duplicate).
  - Added `loadOpenSectionDataRequests()` helper
  - Added `closeResolvedDataRequest()` helper
  - Updated post-planning reconciliation logic

**Impact:** Handles both directions — data arriving late (stale request auto-closed) AND data still missing (new request created, no duplicates). vonc.com's 12 interactive components will be auto-unblocked when pages rebuild.

### P6: Design pipeline chain unblocked (operational)

**Problem:** `webdesign-agent` was `experimental` status. All design chain items (`generic_theme` → `missing_style_collection` → `deactivated_component` → `needs_rerender`) were `unresolved` via two-strike rule.

**Actions taken:**
- Activated `webdesign-agent` (status → `active`)
- Reset `generic_theme` items (7 sites) to `triaged`
- Reset `missing_style_collection`, `deactivated_component`, `needs_rerender`, `needs_design_review` items to `triaged`
- ~95 page rerenders processing successfully after reset

### P9: Audit gap findings routed incorrectly

**Problem:** `write_audit_findings` Rule 4 treated every content finding on an existing page as `content_rewrite`, including `gap` findings. "Services section has empty list" became "rewrite the services page" instead of "rebuild the services page." Empty pages were rewritten (producing garbage that failed validation) instead of rebuilt.

**Files changed:**
- `write_audit_findings_action.go` — Rule 4 now switches on category:
  - `gap` → `needs_content_page` (rebuild via page-build-handler)
  - `tone` → `tone_shift` (unchanged)
  - default (`content`, `differentiation`, `structure`) → `content_rewrite` (unchanged)

---

## What's Pending (Not Yet Deployed)

### P5 (revised version)

The plan-then-reconcile `plan_sections_action.go` was written and output but may need redeployment — the initial version (pre-check only) was deployed first, then revised to the reconcile approach. **Confirm the latest version is deployed** — it should have both `loadOpenSectionDataRequests` and `closeResolvedDataRequest` functions and no pre-check shortcircuit in the loop.

---

## What's Deferred

### P10: Recommendation Specialist Architecture

**Design note:** `design-note-recommendation-specialists.md` (in outputs)

LLM auditors produce findings mixing bugs (factually broken) with recommendations (opinions). The pipeline treats both as auto-fixable. The proposed architecture:

1. Three-way classification: `finding_type: bug | recommendation | gap`
2. Bugs → auto-fix (content_rewrite/rebuild)
3. Recommendations → specialist agent decides (apply/dismiss/escalate)
4. Per-site `approval_mode` (`auto` | `review`) for HITL gating
5. Phased rollout: classification first, then identity-advisor (highest-impact specialist), then others

**Estimated effort:** ~1 week. Deferred until HITL queue becomes a bottleneck or content rewrite false positives become frequent.

---

## Current Queue State

### Actively processing
- Page rerenders across all 7 sites (triggered by design chain reset)
- Nav drift items (triggered by rerenders)
- `improve_tool` items (unblocked by P1)

### Human-gated (genuine — no code fix)

**`needs_section_data` (25 items):**
- Pricing tiers: vonc.com/membership, gaswholesalers.com/how-pricing-works
- Leadership/team bios: robot-hands.com/about, ai-agent-orchestration.com/about
- Case studies: gaswholesalers.com/client-case-studies, robot-hands.com/learning-center
- Use cases: gaswholesalers.com (2 pages), robot-hands.com/gripper-selection-guide
- Contact info: gaswholesalers.com/contact
- Portfolio: leopardessconsulting.co.uk/for-engineering-teams
- Tool sections (vonc.com — 12 items): should auto-resolve when P5 reconcile runs and components are found as ready

**`content_rewrite` / `needs_human_review` (6 items):**
- Email false positives (2): finetuning.uk + leopardessconsulting.co.uk — `contactforsales.com` flagged as placeholder. Recommendation: close as `wont_fix` (P10 would prevent these long-term)
- Empty services list: finetuning.uk/services — should re-route as `needs_content_page` now that P9 is deployed (next audit cycle)
- Empty homepage: vonc.com/index — same as above
- Unsubstantiated claims: leopardessconsulting.co.uk/about — genuine content quality issue
- Tool reference addition: ai-agent-orchestration.com — enhancement, not a bug

### Stale unresolved (recommended resets)

These were waiting on fixes that are now deployed:

```sql
-- Reset audit_tool items (P1 fix deployed — component_id now flows)
UPDATE site_work_items
SET status = 'triaged', error = NULL,
    summary = regexp_replace(summary, '^\[.*?\]\s*', ''),
    claimed_by = NULL, claimed_at = NULL, updated_at = NOW()
WHERE item_type IN ('audit_tool', 'audit_finding_audience', 'needs_blog_posts')
  AND status = 'unresolved'
RETURNING id, item_type, LEFT(summary, 60);

-- Reset internal_links (agent exists and is active)
UPDATE site_work_items
SET status = 'triaged', error = NULL,
    summary = regexp_replace(summary, '^\[.*?\]\s*', ''),
    claimed_by = NULL, claimed_at = NULL, updated_at = NOW()
WHERE item_type = 'needs_internal_links'
  AND status = 'unresolved'
RETURNING id, LEFT(summary, 60);

-- Close email false positives
UPDATE site_work_items
SET status = 'wont_fix',
    error = 'Audit false positive — contactforsales.com is legitimate email provider'
WHERE item_type = 'content_rewrite'
  AND status = 'needs_human_review'
  AND spec->>'description' LIKE '%contactforsales%';
```

---

## Files Delivered (All in /mnt/user-data/outputs/)

### Go files (deploy to platform/orchestration/actions/)
| File | Fix |
|------|-----|
| `work_item_actions.go` | P1 |
| `create_work_item_action.go` | P1 |
| `deploy_tool_action.go` | P2 |
| `component_library.go` | P2 |
| `ai_errors.go` | P3 |
| `ai_actions.go` | P3 |
| `load_page_record_action.go` | P4 |
| `plan_sections_action.go` | P5 |
| `write_audit_findings_action.go` | P9 |

### SQL migrations
| File | Fix |
|------|-----|
| `fix_p1_tool_auditor_component_id.sql` | P1 — tool-auditor workflow update |
| `fix_load_page_record_page_id.sql` | P4 — page-build-handler workflow update |

### Documentation
| File | Purpose |
|------|---------|
| `pipeline-failures-report.md` | Full triage report with all 57 items, analysis, and fix status |
| `design-note-recommendation-specialists.md` | P10 architectural proposal for future implementation |

---

## Key Patterns Discovered

1. **Two sources of truth** — `sites.email` vs `site_specs.identity.email` can drift. `loadSiteContactEmail` uses COALESCE across both. Content writers may use either. Needs consolidation.

2. **Two-strike rule creates born-unresolved items** — `insertWorkItem` marks items `unresolved` if 2 prior items with the same key failed. The new item has `attempt_count: 0` and is excluded from dispatch indexes. Good for investigation visibility, but items pile up if nobody investigates.

3. **Self-contained components treated as data-dependent** — Interactive components (12–23KB with `<script>` and `<style>`) with no `input_schema` are treated as needing LLM content. The `planSection` function's heuristic only catches `article/content/body/text/blog` patterns. Large self-contained templates should be detected automatically (future improvement).

4. **Feedback loops need both directions** — Creating `needs_section_data` items when sections lack data (forward loop) is only half the solution. Closing those items when the data/component arrives (reverse loop) is equally important. P5 handles both.

5. **LLM opinions ≠ bugs** — Auditors mix factual findings with business recommendations. The pipeline shouldn't auto-fix opinions. P10 proposes specialist agents that can evaluate recommendations in context.

# HANDOFF — Phase 1 Deployed; Section Data Handler Pending

**Session date:** 2026-05-06
**Phase 1 site-plan refactor:** Deployed and verified end-to-end on `gamesdesign.co.uk`.
**Outstanding:** A `needs_section_data` handler agent does not exist. List-style sections (`tool-list`, `guide-list`, `game-list`, `news-listing`, etc.) cannot be populated. Affects every site that has an index/directory page with a list section.

---

## What was deployed in this session

### Phase 1 plumbing

Per the design described in `030_phase1_plan_and_reconciler.md`. Four new domain tables, one terminal write action, one reconciler action, prompt and workflow changes for `build-site-planner`.

**Migrations applied:**

1. **`031_site_plans_schema.sql`** — creates `site_plans`, `site_plan_pages`, `site_plan_sections`, `site_plan_directives`. First version of this file landed earlier; was re-edited mid-session and the second version's `CREATE TABLE IF NOT EXISTS` blocks silently skipped over the earlier shape.
2. **`032_phase1_planner_workflow.sql`** — replaces `build-site-planner`'s `default_config` with the new step graph (`read_specs → ensure_site → load_components → load_styles → plan_site → validate_plan → write_site_plan → sync_pages → populate_nav → reconcile_site_plan → complete`). Per-agent backup snapshot taken to `agent_def_build_site_planner_backup_20260505`.
3. **`033_phase1_schema_repair.sql`** — bridges the gap between the two versions of 031: adds `title`, `meta_description`, `nav_label` to `site_plan_pages`; drops orphan `page_data` column; drops orphan `site_plan_partials` table.
4. **`034_planner_prompt_tighten.sql`** — strict-vocabulary instruction for `page_type` values; removes "plan the IDEAL site" language that was producing speculative pages (`prototype-page`, `post`).
5. **`035_remove_ensure_pages_contact.sql`** — drops `contact` from `validate_plan`'s `ensure_pages` config; only `index` remains forced.
6. **`036_component_registry_hygiene.sql`** — renames three contract-violating component rows (`category_section → category-listing`, `article_grid → content-listing`, `call_to_action → call-to-action`); deactivates **42** components platform-wide whose `html_template` was substantive (>500 chars) but missing `</section>`; emits one `needs_component_regeneration` work item per deactivated component.
7. **`037_component_creator_prompt_tighten.sql`** — adds a third pre-output self-check (`(c) Structural completeness`) requiring closing-tag verification, with explicit "trim elaboration over leaving structure incomplete" guidance and explicit `<no value>` warning.

**Chassis image deployed** (hash `8584b77648`) carrying:

- `write_site_plan_action.go` — terminal planner action; one transaction writes all four plan-domain tables; canonicaliser output is authoritative for `role`/`page_type`; lock transfer via composite-key match on previous plan's directives.
- `reconcile_site_plan_action.go` — diff plan-pages vs realised pages; emit `needs_page` items + terminal `needs_rerender` (priority 99). NULL `built_from_plan_version` treated as stale.
- `page_canonical.go` — section-index family convergence (`section_index`, `blog_index`, `entity_directory` all produce `<X>-index` name and `/<X>/index.html` URL, with `page_type` retained); new role handling for `landing` and `entity_page`; `normalisePageType` helper distinct from validator's `normaliseRole`.
- `validate_page_content.go` — writes structured `agent_error_log` row (`error_code = CONTENT_VALIDATION_BLOCKER_DETAIL`) before returning the failure error, so post-mortem doesn't require pod-log access; removed bare `placeholder` substring from blocker patterns (was firing on legitimate HTML attributes like `<input placeholder="...">` and CSS classes like `tl-card-icon-placeholder`).
- `store_generated_component_action.go` — extended pre-store validation: rejects substantive (>500 char) templates with zero `{{placeholder "..."}}` tokens; rejects templates containing literal `<no value>` strings; regeneration lookup widened from `is_active = true` filter to `ORDER BY is_active DESC, updated_at DESC LIMIT 1` so deactivated rows still match as regen targets; UPDATE branch now sets `is_active = true` so regenerated rows reactivate when their template passes all gates.

### Doc updates landed

- `031_locks.md` — canonical lock-pattern doc for the locked_at/locked_by columns.
- `001_development_guide.md` — corrected Field Name Collisions section (nested-source loop affects required AND optional fields, earlier guide wording was wrong).
- `007_adoption_pipeline_v4.patch` — strategic-vs-plan-time naming clarifications.

---

## End-to-end verification

`gamesdesign.co.uk` re-adopted at 15:58:53 (correlation `4a06d19f-5a3b-47b0-ad8b-12a9b92aa2cb`). Cascade ran through:

- Adoption complete by 16:04:25.
- Classifier, strategist, briefing complete by ~16:13.
- Planner ran `write_site_plan` + `reconcile_site_plan` cleanly; plan written, page items emitted.
- `page-build-handler` claimed and built tool detail pages and blog posts without issue.

By 18:06 four pages deployed (`tool-progression-architect`, `tool-drop-rate-simulator`, `tool-jump-physics`, `tool-lanchester-sim`, `guide-fairness-rng`). Several more building.

**Pages stuck on list sections:** `index`, `tool-tools`, `guides-index`, `games`. These pages emit `needs_section_data` work items for their list components, which have no handler — see below.

---

## Outstanding issues

### 1. `needs_section_data` has no handler (BLOCKER for list-style pages)

Captured in `FUTURE_section_data_handler.md`. Across the platform, **41 needs_section_data items have been sitting at `needs_human_review` since at least 2026-05-02**, on six different sites. Affects any page with a directory/list/grid section.

For gamesdesign, the six affected items were manually marked complete with `claimed_by = 'manual_acknowledge_no_handler'` to unblock parent pages. Sections will render with empty data until the handler exists.

### 2. 42 components had broken templates platform-wide

Migration 036 deactivated 42 components system-wide whose templates were truncated (long but missing `</section>`) and full of `<no value>` Go-template render artifacts. Hypothesis: at some point templates were rendered against an empty data context and the rendered output was stored back as source. Root cause not identified; may have been a single bad migration or an intermittent action bug.

Component-creator regeneration via the work items is partially complete:
- Many regenerations succeeded under the patched action.
- `tool-list` and `game-list` regenerated successfully (UPDATE in place; rows are now active with healthy templates but their `name` column is `tool-list_pre_037` / `game-list_pre_037` from the rename workaround — cosmetic, can be cleaned up).
- `guide-list` still pending — system.internal dispatch loop has been intermittent due to Anthropic API 529s and a Kafka topic subscription staleness issue.

Cleanup query for the cosmetic name issue:

```sql
UPDATE content_components
SET name = function, updated_at = NOW()
WHERE name LIKE '%_pre_037'
  AND is_active = true
  AND function = REPLACE(name, '_pre_037', '');
```

### 3. Page name divergence between adoption and planner

Adoption emits `tool-jump-physics`, `tool-lanchester-sim`, `guides`, etc. Planner emits `tool-jump-physics-something`, `tool-lanchester-combat-calculator`, `guides-index`, etc. The Phase 1 canonicaliser handles the section-index family (`guides` / `guides-index` converge) but doesn't address adoption-shape vs planner-shape on tool/guide names.

User explicitly: "I am not worried about orphan pages — we can delete and re-adopt." Treating as not-blocking for now. Diagnostic in `diagnostic_orphan_pages.sql`.

### 4. Chassis bookkeeping bug: `complete-with-error`

Several work items are in `status = 'complete'` but with `error` populated (e.g. `Claim timed out`, `duplicate key violates...`, `rejected by pre-store validation`). When a child orchestration fails, the chassis marks the work item complete instead of failed. Means failed work items don't get retried unless manually reset. Captured but not fixed.

### 5. `built_from_plan_version` not written by page-build-handler

Per the deferred list in earlier turns. Reconciler treats NULL as stale and emits `needs_page` for every deployed page on next reconcile. User explicitly OK'd this. To fix later: have page-build-handler set `pages.built_from_plan_version` from the plan_id available in its input.

### 6. Scheduled reconciler tick

Not built. Reconciler currently fires only when called by the planner. A heartbeat agent + `scheduled_tasks` row mirroring `content-feed-trigger` would produce a periodic reconcile pass.

### 7. `ensure_pages` should be domain-aware

Currently hardcoded in workflow JSON. Should be set by strategist or briefing into `site_specs` and read at plan time. Stub for the next discussion.

### 8. Visible site quality issues (Phase 2 territory)

Inspecting the deployed gamesdesign output reveals issues that aren't pipeline failures but are real output-quality problems:

- **Dead design system in page head.** Each rendered page ships a ~250-line CSS block with BEM primitives (`.button--primary`, `.hero__title`, `.card`, `.section`) at the top of every `<head>`. None of these classes are referenced by the component CSS that follows. Components use their own function-scoped CSS (`.tool-list-section .tl-card`, `.hero .btn-secondary`, `.site-header--gradient`). The design system primitives are dead bytes on every page.

- **No theme CSS variables defined.** Component CSS uses `var(--color-primary, #6366f1)` everywhere — fallback colours are doing all the work because no `--color-primary` is defined in any stylesheet on the rendered page. There's a comment `/* Theme-specific styles injected here: */` in the head, but the injection isn't happening. Visual coherence comes from the LLM picking similar fallbacks, not from a real theme system.

- **Orphan pages duplicated in navigation.** `Games` appears twice in the header (`/games/index.html` and `/games.html`), thrice across header+footer counting `/tools/tools/index.html`. This is the adoption-shape vs planner-shape divergence reaching the user-visible surface. Adoption emitted `games` (content page); planner emitted `games-index` (entity_directory). Both deployed, both linked. The canonicaliser converged the URL of one but not the other.

- **List sections are LLM-fabricated, not data-driven.** Without a `needs_section_data` handler, the content writer received "tool-list, selector unavailable" and improvised a list of tools with plausible names and copy. The deployed `tools` page lists six tools; only some correspond to actual pages on the site. Links go to invented URLs (`/tools/prd-calculator.html`, `/tools/cpi-analyser.html`) that don't exist.

- **Navigation isn't deduplicated.** Header and footer both list sections with no awareness of overlap. `populate_nav_tables` produces this from the `pages` table directly, so any orphan/duplicate page becomes a navigation entry.

These aren't blocking Phase 1 — Phase 1 was about the plumbing and that works. They're the surface-quality work Phase 2 needs to address. The user has explicitly said: *"the site itself is not right, hopefully phase 2 will sort that out."*

---

## Files staged in /mnt/user-data/outputs/ during this session

Already-applied SQL: `031–037` (some applied, some applied after retries — see migration headers).

Already-deployed Go: `write_site_plan_action.go`, `reconcile_site_plan_action.go`, `page_canonical.go`, `page_canonical_test.go`, `page_role_validator.go`, `validate_page_content.go`, `store_generated_component_action.go`, `registry_phase1_actions.diff`.

Diagnostics: `diagnostic_orphan_pages.sql`.

---

## Suggested next session

1. Build the `needs_section_data` handler per `FUTURE_section_data_handler.md`. This is the top priority — it unblocks list-style pages across every site AND replaces fabricated list content with real data drawn from the site's pages table.
2. Reset the 41 platform-wide stuck `needs_section_data` items once the handler is live.
3. Backfill `built_from_plan_version` writes in page-build-handler.
4. Audit the orphan-pages situation across all sites (not just gamesdesign) before re-adopting any.

## Phase 2 candidates (broader, not focused on plumbing)

- Theme CSS variable injection: actually emit a `:root { --color-primary: ...; ... }` block based on the plan's design directives. Without it, all visual choices live in component fallbacks.
- Strip the dead BEM design system from rendered pages, OR migrate components to use it. Either is fine; the current state of having both is wasteful.
- Page deduplication in `populate_nav_tables`: detect adoption-shape vs planner-shape duplicates and only emit one nav entry per logical page.
- A page-cleanup pass after reconciler runs, removing pages whose canonical name no longer matches the current plan.


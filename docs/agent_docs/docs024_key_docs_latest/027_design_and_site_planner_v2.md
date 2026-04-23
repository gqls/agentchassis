# 026 — Design Composition & Site Design Planner

How palette, layout, and typography get chosen for a site; where the `site-design-planner` handler fits into the build and adoption pipelines; and what webdesign-agent's new role looks like.

Companion to `025_palette_layout_typography_migration.md` (the one-time migration plan) and `007_adoption_pipeline.md` (the adoption pipeline flow). This doc covers the steady state after the migration is complete.

---

## 1. What Site Design Planner Does

`site-design-planner` is a task-mode handler agent. Its responsibility is narrow: given a site that has identity and classification specs, resolve a **composition** (one palette + one layout + one typography_set) and install it so renderers have something concrete to work with.

It is invoked via a `needs_composition` work item — inserted by `WriteBuildItemsAction` for new builds and by `apply_adoption_plan_action.go` for adopted sites. It never runs on a cadence; there is no scheduled task that "refreshes composition". Composition is a one-time resolution per build, overwritten only when someone explicitly requeues it.

### What it does not do

- It does not render CSS. That is still `render_css_from_spec` inside `webdesign-agent`.
- It does not decide navigation architecture. That is still `populate_nav_tables` (and, eventually, a dedicated `nav-planner` specialist — deferred).
- It does not decide layout density or section recipes. Those are future specialists described in the "Open design areas" section below.
- It does not call an LLM. All six of its steps are deterministic Go actions.

### Why a dedicated agent

Before this work existed, webdesign-agent owned both composition resolution and CSS rendering. That coupling produced two failure modes:

1. The renderer's `theme_name: "standard-brochure"` config literal meant every site rendered with the brochure layout regardless of what had been chosen for it.
2. When webdesign-agent did produce a fresh composition (via its `install_theme` step), it did so AFTER rendering — so the first deploy committed wrong-layout CSS to git, then a composition appeared afterwards that nothing re-rendered against.

Separating composition into a handler that runs BEFORE rendering fixes both. The renderer resolves the site's installed composition every time; composition is installed once per build, before any CSS is produced.

---

## 2. The Workflow

Agent definition lives in `agent_definitions` (type = `site-design-planner`, version 1). Nine steps, no LLM calls, `processing_mode: "task"`, `timeout_seconds: 300`.

```
ensure_site_record
  → validate_composition_inputs
    → check_ready
       (ready = true)   → resolve_composition_layout
                          → resolve_composition_typography
                          → resolve_composition_palette
                          → install_site_composition
                          → complete
       (ready = false)  → complete_unready
```

### Step-by-step

**`ensure_site_record`** — standard chassis step. Reads `site_id` / `domain` from `input_data`, upserts sites row if needed.

**`validate_composition_inputs`** — reads `identity` and `classification` specs from `site_specs`. Checks both are present and non-empty. Emits `logger.Error` on miss, AND queues a `needs_domain_research` work item with `item_key = 'backfill_classification_for_composition'` so the classifier re-runs for this site. Returns `{ready: bool, identity, classification, missing: [...], classifier_queued: bool}` for downstream steps.

The dual signal matters: the loud log surfaces the issue immediately, the queued item makes it self-healing. If classification is backfilled, the next build cycle for this site will find it ready.

**`check_ready`** — conditional branch. Routes to resolution chain if both specs were found, or to `complete_unready` otherwise.

**`resolve_composition_layout`** — thin wrapper over the shared `resolveLayoutByTags` helper (also used by `fork_theme_from_site`). Matches `classification.industry_tags` against `layouts.industry_tags` for active non-fallback layouts, picks the best-scored match. On fallback (no match crosses the score threshold), it queues a `needs_new_layout_candidate` item with `handler_agent: 'hitl-review'` and `status: 'needs_human_review'` — the dispatch loop skips status `needs_human_review`, so nothing tries to claim it and no "blocked" flip happens. Humans pick it up when they choose to.

**`resolve_composition_typography`** — thin wrapper over `resolveTypographySet`. Cascade: `design_reference.typography.font_family` → `design_intent.typography.reference_values.font_family` → `mission.preferred_typography` → `sans-modern` (the default seed). Matches against existing `typography_sets` rows first; if font family matches a library row, reuses it. Otherwise creates a new `typography_sets` row with `origin = 'adopted'`.

**`resolve_composition_palette`** — thin wrapper over `createPalette`. Cascade: `design_reference.suggested_mapping` → `mission.preferred_palette` → `design_intent.palette.reference_values` → selected layout's seed palette → default palette. Palettes are always site-specific (no library reuse across sites), so this always creates a new `palettes` row with `origin = 'adopted'` and `source_site_id = <site>`.

**`install_site_composition`** — the atomic write. In a single transaction:
1. INSERT into `css_themes` (origin = `adopted`, palette_id + layout_id + typography_set_id populated, source_site_id set)
2. INSERT into `style_collections` (css_theme_id points at the new row, layout's `default_header_component_id` / `default_footer_component_id` inherited if the layout declares them)
3. Guarded UPDATE on `sites.style_collection_id` — only commits if the value was NULL when the transaction started. Prevents races between two concurrent composition items for the same site.
4. Superseded + inserted `resolved_composition` spec row — preserves the history contract (is_current/superseded_at) that every other spec aspect follows.

All four writes live in one transaction. If any fails, none commits. There is no separate `write_site_spec` step for `resolved_composition` — it happens inside `install_site_composition` so rollback is automatic.

**`complete`** / **`complete_unready`** — standard exit steps. `complete_unready` exits with status `failed` and message pointing at `validated_inputs.missing`; the queued classifier-recovery item handles the self-heal.

---

## 3. Pipeline Integration

### New build path (WriteBuildItemsAction)

Building a site from scratch queues a cascade of work items. Since the composition work, the ordering is:

| Priority | Item type | Handler | Depends on |
|---|---|---|---|
| 5 | needs_logo (if plan flag) | image-build-handler | — |
| 5 | needs_hero_image (if plan flag) | image-build-handler | — |
| **7** | **needs_composition** | **site-design-planner** | — |
| 8 | needs_design | webdesign-agent | `[composition_id]` |
| 10+ | needs_content_page (per page) | page-build-handler | — |
| 20 | needs_rerender | rerender-pages | — |

The `needs_design` item carries `depends_on = [composition_item_id]` in its `uuid[]` column. `LoadWorkItemsAction` already honours this — the dispatch query excludes any item whose dependencies aren't in `(complete, verified)` status. So `needs_design` waits for `needs_composition` to complete before webdesign-agent gets a chance to claim it.

The UUID plumbing in `insertWorkItem` predates composition — it existed for the content_rewrite + rerender dependency pattern. Composition just reuses it.

### Adoption path (apply_adoption_plan_action.go)

Adoption had its own raw `tx.ExecContext` inserts for work items — not using `insertWorkItem`. Rather than rewriting the whole action, the composition wiring is surgical: two new inserts (for `needs_composition` and the depends_on-gated `needs_design`) via `insertWorkItem`, while existing content-page and logo/hero inserts stay as they were.

Adoption also now writes a `mission` spec if `input_data` carries one (or if it carries `objective` + `direction` which get reconstructed into the mission shape). This gives `resolve_composition_palette`'s cascade something to work with when no `design_reference` exists yet — a matching hint the adopter used to pass via other means.

### Trigger matrix

| Event | Triggers needs_composition? |
|---|---|
| New site adoption (POST /sites with a domain that has an existing site) | Yes, via `apply_adoption_plan_action.go` |
| Fresh build from site-planner | Yes, via `WriteBuildItemsAction` |
| Improvement loop discovery | No — improvement loop does not refresh composition |
| Manual requeue from admin | Yes, if the admin chooses to |
| Classifier backfill completing | No — composition resolution is not auto-retriggered when classification appears late. Operators explicitly requeue |

---

## 4. Renderer Changes

`render_css_from_spec` (called from webdesign-agent's `generate_css` step) now has a three-branch resolution cascade:

1. `config.theme_id` — explicit UUID override (used by HITL preview paths, library tooling, tests)
2. `config.theme_name` — explicit name override (same audience)
3. `site_context.site_id` → `sites.style_collection_id` → `style_collections.css_theme_id` — **the production path**

If all three miss, the renderer falls back to `standard-brochure` but with `logger.Error` flagging it as an emergency fallback. This path is reachable today only if the pipeline routed to webdesign-agent without site-design-planner ever running — which after the workflow reorder (section 5) is genuinely unreachable during normal operation. Log monitoring should treat any emergency-fallback line as a pipeline bug to investigate.

The old per-workflow `theme_name: "standard-brochure"` config literal on webdesign-agent's `generate_css` step is gone — it was cleared to `{}` as the cutover moment. With the literal present, the renderer's site-composition branch could never fire. With it cleared, the cascade falls through to the live site's composition whenever one exists.

### Helper: `resolveThemeIDFromSiteContext`

Lives in `render_css_composition_loader.go`. Takes the `site_id` from `config` (populated by the action from `collectedData.site_context.site_id`), does the sites → style_collections → css_themes lookup, returns the theme UUID as string. Returns empty string on any failure — never errors; letting the caller fall through to the emergency fallback. Logs a Warn line naming the site_id and a distinguishing reason (no collection linked, join failure, etc.) on every non-happy path.

---

## 5. Webdesign-Agent After The Merge

The full post-merge workflow:

```
check_site_context → {use_provided_context | load_site_context}
  → check_has_site_id → {read_site_specs | analyze_design}
  → analyze_design (LLM — produces design_spec)
  → check_update_db
     ├─ (site_id) → update_site → generate_css
     └─ (no site_id) → generate_css
  → generate_css → deploy_css
  → check_should_fork
     ├─ (should_fork) → fork_theme → complete
     └─ (else) → complete
```

Two steps are gone compared to the intermediate shape this chat and the parallel chat produced together: `check_should_install` and `install_theme`. Both were the belt-and-braces install path. With site-design-planner owning composition resolution for every pipeline-invoked run, the belt-and-braces was masking orchestration bugs rather than protecting against them. Section 6 covers the removal in detail.

What the post-merge sequence guarantees:

- **One install path, upstream.** site-design-planner is invoked before webdesign-agent via the `depends_on`-gated `needs_design` work item. By the time webdesign-agent runs, composition already exists. If it doesn't, something upstream is wrong and the renderer's emergency fallback + logger.Error surfaces it.
- **webdesign-agent does one thing.** Produces a design_spec overlay via the LLM, renders CSS, deploys, optionally forks to the library. No composition concerns. No install concerns.
- **`update_site` persists `design_spec` before `generate_css`.** The `design_spec` overlay is available for the renderer to merge on top of the installed composition.
- **`fork_theme` still runs AFTER the render.** Forking captures what actually got rendered — the spec at render-time plus the computed CSS. Pre-render forking would capture intent without capturing what was produced from that intent, which is much less useful for library curation. `fork_theme` is also not routinely triggered — it requires `input_data.should_fork_theme == true`, which is set only when a human flags a site's current design as worth contributing to the library.

### Roles

- `site-design-planner` owns composition for new builds and adoptions. Single install path.
- `render_css_from_spec` owns rendering — deterministic Go, no LLM. Its emergency-fallback branch is the final safety net and should never fire in normal operation.
- `webdesign-agent.analyze_design` (LLM) produces `design_spec` — the palette/typography/spacing overrides merged into the composition's values at render time. This is the evolving-preference overlay, independent of composition resolution.
- `webdesign-agent.fork_theme` (opt-in via input flag) contributes the current rendered design to the library pending HITL review.

---

## 6. Removing `install_theme` From Webdesign-Agent (Applied)

The post-merge state collapses composition onto a single install path. This section documents the merge itself — what was removed, why, and what the alternative handles.

### What was removed

Two workflow steps deleted from `agent_definitions.default_config.workflow.steps` for type `webdesign-agent`:
- `check_should_install` — the conditional that tested `site_context.style_collection_id == null` to route to `install_theme`.
- `install_theme` — the step that called `fork_theme_from_site` with the now-removed `install_on_site: true` flag.

Routing rewired:
- `check_update_db.config.else_step`: `check_should_install` → `generate_css`
- `update_site.next_step`: `check_should_install` → `generate_css`

Output fields trimmed:
- `complete.config.output_fields`: `install_result` removed.

Go changes (single file, `fork_theme_from_site_action.go`):
- `install_on_site` removed from `InputSpec.Optional` and `Defaults`.
- The flag-branching block (install-mode UPDATE sites + skip-HITL, vs library-mode HITL-item insert) collapsed to the library-only path.
- `needsReview` no longer parameterised off the flag — hardcoded `true` because every fork that reaches this action is now a library contribution.
- `install_on_site` added to `InputSpec.Deprecated` with a message pointing at site-design-planner. If any stale workflow config still references the key, the action emits a `logger.Error` line at read time — makes it obvious which row needs updating.

### What site-design-planner covers

The install path that `install_theme` used to handle is now `install_site_composition` inside site-design-planner. That agent is invoked via `needs_composition` work items, queued by:
- `WriteBuildItemsAction` for fresh builds
- `apply_adoption_plan_action.go` for adopted sites
- Manual requeues from admin

`needs_design` (the work item that dispatches webdesign-agent) carries `depends_on = [composition_item_id]`, so webdesign-agent cannot run before composition is installed for pipeline-invoked runs.

### The scrape-only case

webdesign-agent can be invoked with `site_context` directly, outside the pipeline. For these runs there is no `needs_composition` item and site-design-planner never runs. The renderer sees no composition and hits its emergency-fallback branch:
- Logs `logger.Error` naming the site_id and the reason
- Renders with `standard-brochure` layout and the current run's `design_spec` overlay

The site still gets CSS. It's not the composition site-design-planner would have chosen, but it's functional — and the `logger.Error` is an alerting signal that something upstream needs to queue a `needs_composition` item for this site. Callers doing scrape-only experiments get a working render with a clear log saying "pipeline invocation would have been better." Production callers going through the normal pipeline never hit this path.

### Why not force hard-fail

An alternative considered: webdesign-agent could hard-fail if no composition is resolvable, forcing every caller to queue `needs_composition` first. Rejected because:
- Scrape-only runs are a legitimate use case for testing and development — killing them makes the agent harder to exercise in isolation.
- The emergency-fallback rendering is a genuinely useful failure mode — it produces something visible to diagnose against, rather than a silent error in the log.
- The `logger.Error` carries the same signal as a hard-fail would: "this invocation bypassed composition resolution." Alerting on that log line catches the same class of bug with better developer experience.

### Why not call_agent to site-design-planner inline

Another alternative: have webdesign-agent use `spawn_agent` + `call_agent` to invoke site-design-planner inline when no composition exists. Rejected because webdesign-agent is `processing_mode: task`. In the current chassis, spawn/call patterns are used only by `processing_mode: orchestrator` agents. Promoting webdesign-agent to orchestrator mode is a valid future option — particularly as design requirements get finer-grained and webdesign-agent needs to coordinate more specialists — but not required for this merge. The scrape-only path is rare enough that emergency-fallback + loud log is an acceptable answer today.

### Not removed

- `fork_theme_from_site` action itself. Still the right mechanism for the `fork_theme` step — capturing a rendered design into the library as a forkable theme pending review. Library forks are fundamentally different from site installs: a library fork is about long-term catalogue curation with HITL gating; a site install is about committing a specific composition to a specific site's render pipeline. One action can't honestly cover both. Post-merge, `fork_theme_from_site` is clearly library-only.
- The renderer's emergency fallback. Stays as the absolute last line of defence.
- `load_site_for_design`'s `style_collection_id` surfacing. Unused by the current workflow (the removed `check_should_install` was its only consumer) but useful for diagnostics and for any future conditional. Cheap to keep.

### Verification

After the merge SQL applies, run:

```sql
SELECT
    default_config #>> '{workflow,steps,check_update_db,config,else_step}' AS check_update_else,
    default_config #>> '{workflow,steps,update_site,next_step}'             AS update_next,
    default_config -> 'workflow' -> 'steps' ? 'install_theme'               AS install_exists,
    default_config -> 'workflow' -> 'steps' ? 'check_should_install'        AS check_install_exists
  FROM agent_definitions
 WHERE type = 'webdesign-agent' AND is_active = true;
```

Expected: `check_update_else=generate_css`, `update_next=generate_css`, both existence flags `false`.

For a live sanity check, watch for emergency-fallback log lines over the first week post-merge:

```
kubectl -n ai-persona-system logs -l app=agent-chassis --since=1h \
  | grep -E "render_css_from_spec.*emergency_fallback|no composition found"
```

Zero firings means the pipeline is routing correctly. Any firings should be investigated — either a pipeline bug (composition not queued) or a scrape-only invocation (expected but worth noting).

---

## 7. Cleanup and Orphans

Adopted composition rows (palettes, typography_sets, css_themes, style_collections) accumulate over time. When site-design-planner runs fresh resolution for a site that already had an installed composition, the old `css_themes` row is deactivated (`is_active = false`) rather than deleted, but the `palettes` and `typography_sets` the old theme referenced stay in place as orphans.

The `database-cleanup` scheduled task (interval 3600s, target_agent `database-cleanup`, concurrency group `maintenance`) now includes two cleanup CTEs for orphan composition rows:

- **`deleted_orphan_palettes`** — adopted palettes, `source_site_id IS NOT NULL`, `forked_at < NOW() - INTERVAL '24 hours'`, NOT EXISTS any `css_themes` row referencing them
- **`deleted_orphan_typography`** — same predicate shape for typography_sets

Both have a 24-hour grace period so they don't race an in-flight install. Seed rows (`origin = 'seed'`) are never touched by either CTE — they stay in the library permanently even if no active theme currently references them, because they're the fallback pool for future resolutions.

The cleanup counts surface in the scheduler's task output alongside the existing counts (errors_deleted, audit_deleted, orchestrations_deleted, stale_deleted).

---

## 8. Related Specs

`site-design-planner` reads these `site_specs` aspects:

| Aspect | Role |
|---|---|
| `identity` | Required. Company name, industry, target audience — the base context. |
| `classification` | Required. `industry_tags` drive layout matching. `site_type` and `tone_suggestion` are read by downstream webdesign-agent, not site-design-planner itself. |
| `design_reference` | Optional. If present, its `suggested_mapping` (concrete hex values extracted from an adopted site's CSS) is the highest-priority source for palette. Its `typography.font_family` is the highest-priority source for typography font matching. |
| `design_intent` | Optional. Semantic direction ("dark IDE aesthetic, functional not atmospheric") with reference values as starting points, not targets. Middle priority in both palette and typography cascades. |
| `mission` | Optional. `preferred_palette` and `preferred_typography` fields are the classifier/adopter's way of passing explicit hints. |

It writes exactly one aspect:

| Aspect | Shape |
|---|---|
| `resolved_composition` | `{css_theme_id, style_collection_id, palette_id, layout_id, typography_set_id, resolved_at, resolution_sources: {palette: "design_reference\|mission\|design_intent\|layout_library\|default", typography: "...", layout: "tag_match\|fallback"}}` |

`resolved_composition` is the audit trail — "this site's current composition was resolved from these signals on this date." It's read by the admin dashboard (future) for showing provenance. It's NOT read by the renderer — the renderer reads the live `sites.style_collection_id` link, which is the source of truth.

---

## 9. Failure Modes and Recovery

| Failure | What happens | How it recovers |
|---|---|---|
| Identity or classification spec missing | validate_composition_inputs fails work item + queues `needs_domain_research` recovery item. Dispatch loop picks up the recovery item, classifier backfills specs. | Operator requeues `needs_composition` for the affected site after classifier completes. No auto-retrigger. |
| Layout tag match below threshold | resolve_composition_layout uses default layout + queues `needs_new_layout_candidate` (handler hitl-review, status needs_human_review). | Human reviews the queued candidate, decides whether to add a new layout to the library. Composition already installed with fallback layout. |
| install_site_composition transaction fails | Work item fails with the SQL error; palette + typography rows from resolver steps may have been committed before install failed. | `database-cleanup` scheduled task sweeps orphan palettes + typography_sets after 24h. Operator can requeue composition; the resolvers will create new rows rather than trying to reuse the orphans. |
| Race: two concurrent needs_composition items | install's guarded UPDATE `WHERE style_collection_id IS NULL` means only one commit wins; the other sees 0 rows affected and errors. | Second work item fails cleanly; operator requeues if needed. Partial-unique-index on site_work_items prevents this race in practice (only one open needs_composition per site) but the UPDATE guard is defence in depth. |
| Renderer hits emergency fallback | logger.Error fires. Site still renders (with standard-brochure + design_spec colours), but it's a visible pipeline bug. | Operator investigates which path bypassed site-design-planner. After webdesign-agent.install_theme is removed (section 6), the emergency fallback is the only remaining safety net — log monitoring should alert on it. |

---

## 10. Open Design Areas

What's in scope for future design-specialist work, not built yet:

**`nav-planner` specialist.** Decides navigation architecture (horizontal top, sidebar, split, overlay), tool grouping strategy, max visible items, CTA strategy, mobile pattern. Writes a `navigation` spec. Currently this is partially handled by `populate_nav_tables` mechanically; a planner agent would write the spec that populate_nav_tables then honours. Referenced in `FOCUS_navigation_HANDOFF_navigation_fix.md`.

**`layout-planner` specialist (richer than site-design-planner).** Writes a `layout` spec covering default page layout shape (full-width, sidebar, two-column, asymmetric), header/footer style, hero/nav merging, sidebar pages, per-page layout overrides. Today's site-design-planner picks a `layouts` row from the library; it doesn't decide which pages should be sidebar-shaped or how the header should relate to the hero. Referenced in `FOCUS_navigation_HANDOFF_navigation_fix.md` as "Option B — dedicated site-design-planner agent" (scope since split into the narrow composition work that shipped, plus this broader work that hasn't).

**`hitl-review` handler.** Consumes `needs_human_review` work items (including `needs_new_layout_candidate` from resolve_composition_layout, and tool-auditor review items). Deferred until enough review items accumulate in production to understand the UX shape.

**Recommendation specialists** (per `PLAN_design-note-recommendation-specialists.md`). Three-way classification of audit findings (bug / gap / recommendation) with dedicated specialists per category rather than routing everything to content_rewrite. Separate project from composition but adjacent — it affects how the improvement loop pushes composition changes via design_intent updates.

**Section recipes and layout-aware page planning.** Today `plan_sections` selects from a global component library. A layout-aware version would pick sections based on the layout spec — an `industry-hub` layout wants a different mix of sections than a `portfolio-kinetic` layout. Referenced in 025 Phase 4 as future work.

---

## 11. Reference: Key Files

### Go actions (handler implementations)
- `platform/orchestration/actions/validate_composition_inputs_action.go`
- `platform/orchestration/actions/resolve_composition_layout_action.go`
- `platform/orchestration/actions/resolve_composition_typography_action.go`
- `platform/orchestration/actions/resolve_composition_palette_action.go`
- `platform/orchestration/actions/resolve_composition_helpers.go`
- `platform/orchestration/actions/install_site_composition_action.go`
- `platform/orchestration/actions/fork_theme_composition.go` — shared helpers (resolveLayoutByTags, resolveTypographySet, createPalette)

### Go actions (renderer + pipeline wiring)
- `platform/orchestration/actions/render_css_from_spec_action.go` — resolution cascade entry point
- `platform/orchestration/actions/render_css_composition_loader.go` — `loadThemeComposition`, `resolveThemeIDFromSiteContext`
- `platform/orchestration/actions/render_css_composition_helpers.go` — merge helpers
- `platform/orchestration/actions/load_work_item_actions.go` — WriteBuildItemsAction + insertWorkItem
- `platform/orchestration/actions/apply_adoption_plan_action.go` — adoption path

### Database
- `agent_definitions` type = `site-design-planner` (version 1)
- `agent_definitions` type = `webdesign-agent` (updated for reorder)
- `scheduled_tasks` name = `database-cleanup` (pre_query extended with orphan CTEs)
- Tables: `palettes`, `layouts`, `typography_sets`, `css_themes`, `style_collections`, `sites.style_collection_id`

### Migration artefacts (one-time, see 025)
- Seed data for palettes, layouts, typography_sets
- Phase 2 migration SQL (css_themes restructure)
- Phase 3 css_themes → (palette, layout, typography) mapping

---

## 12. Change Log

- **2026-04-19 (late)** — Merge applied. Two install paths collapsed to one. Removed `install_on_site` flag and install-mode branch from `fork_theme_from_site_action.go`; action is now library-contribution only. Removed `check_should_install` and `install_theme` steps from webdesign-agent workflow; rewired `check_update_db.else_step` and `update_site.next_step` straight to `generate_css`. site-design-planner is now the sole composition-install path. Scrape-only invocations that bypass the pipeline hit the renderer's emergency fallback (with `logger.Error`) rather than the previous defensive install.
- **2026-04-19 (earlier)** — Composition pipeline shipped end-to-end. site-design-planner agent registered, build + adoption pipelines wired, renderer cascade cutover applied, webdesign-agent workflow reordered so install_theme runs before generate_css. Orphan cleanup extended into `database-cleanup` scheduled task.
- **2026-04-18** — site-design-planner agent definition landed (deliverable 5). Pipeline wiring patches landed (deliverable 6). Renderer patch + workflow config clear landed (deliverable 7 + 7.5). Reset sweep landed (deliverable 8).
- Earlier dates — see 025 for the migration phases.

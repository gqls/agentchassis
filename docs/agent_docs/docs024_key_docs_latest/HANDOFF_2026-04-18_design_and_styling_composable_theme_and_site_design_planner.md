# HANDOFF: Design & Styling — Composable Theme + site-design-planner

## Date: 2026-04-18
## Supersedes: `FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md`
## Previous session transcripts:
- `/mnt/transcripts/2026-04-17-*-composable-theme-deployment.txt`
- `/mnt/transcripts/2026-04-18-13-02-36-composable-theme-migration-025.txt` (Phases 4-5 build)
- Current session (composition flow redesign and site-design-planner plan)

---

## 1. Status Summary

Migration `025_palette_layout_typography_migration` is through Phase 5 in code. Phases 1-3 (data model, layouts, seeding) are deployed and verified. Phases 4-5 (renderer cutover, fork action rewrite) were deployed but not end-to-end verified — a smoke-test run was incomplete when the session paused.

While testing, we realised the bigger architectural problem: fork + install is conflated inside webdesign-agent's workflow, which causes a first-render-with-wrong-layout problem. The correct answer is a dedicated **site-design-planner** agent that resolves composition (layout + palette + typography) BEFORE webdesign-agent renders. Phase 3 of `FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md` already made this decision (decision #8: "Option B: dedicated site-design-planner agent"). We're now implementing it.

---

## 2. What's Deployed and Verified

### Phase 1 — Layouts seeded (15 rows in `layouts` table)
All 15 layout CSS templates produced, reviewed against contracts, seeded to DB:
- `brochure-formal`, `brochure-bold`, `portfolio-kinetic`, `magazine-grid`, `utility-tool`
- `media-grid`, `docs-sidebar`, `soft-editorial`, `technical-precise`, `high-energy`
- `comparison-aggregator`, `affiliate-hub`, `ecommerce-storefront`, `tool-first-landing`, `industry-hub`

Each passes the 7-point audit: zero color-heading strays, h1-h6 contract-exact, all helper fallbacks, no rogue --section-* declarations, balanced $LAYOUT$ markers, 5 surface classes present, ON CONFLICT upsert.

Contract decisions applied across all:
- Heading fallback `var(--section-heading, var(--color-primary))`
- No `--section-*` declarations on section containers (renderer owns)
- 5 surface-painted classes (`.differentiators/.features/.faq/.services/.about-section`) — stays for Phase 1, decouple = Phase 4.5

### Phase 2 — Schema migration deployed
- `palettes`, `layouts`, `typography_sets` tables created with lineage columns
- `css_themes` and `style_collections` gained `palette_id`, `layout_id`, `typography_set_id` FKs
- 22 indexes, all FKs `ON DELETE SET NULL`

### Phase 3 — Seeding deployed and verified
- 6 typography_sets: `sans-modern`, `serif-editorial`, `display-bold`, `mono-technical`, `serif-classical`, `sans-friendly`
- 15 layouts (from Phase 1)
- 13 palettes extracted from existing themes via plpgsql function
- 14 css_themes rows fully linked with palette_id + layout_id + typography_set_id
- Verified: `SELECT COUNT(*) FROM css_themes WHERE palette_id IS NOT NULL AND layout_id IS NOT NULL AND typography_set_id IS NOT NULL` = 14

### Phase 4 — Renderer cutover (deployed but UNVERIFIED end-to-end)
Three files deployed to `platform/orchestration/actions/`:
- `render_css_composition_helpers.go` — pure merge helpers (buildPaletteMap, buildTypographyMap, makeMapLookupFunc)
- `render_css_composition_loader.go` — single JOIN loader, hard-errors on NULL FKs
- `render_css_from_spec_action.go` — full replacement, uses FuncMap (`palette`, `typo`, `token`)

Emits structured log line `RenderCSSFromSpecAction: spec → theme overrides` showing which spec keys won vs which were ignored.

Phase 4.4 cleanup audit completed: no external callers of legacy `cssTemplateData`, `loadCSSGoTemplate`, `extractDesignColors`, `designColorMaps`. `css_templating.go` / `TemplateCSSFromSpec` stays alive (still referenced by `fork_theme_from_site` until Phase 7).

### Phase 5 — Fork action rewrite (deployed but UNVERIFIED end-to-end)
Two files deployed:
- `fork_theme_composition.go` — helpers: `createPaletteForFork`, `resolveTypographySetForFork` (exact font_family match or new), `resolveLayoutForFork` (tag-overlap scoring against layouts.industry_tags)
- `fork_theme_from_site_action.go` — full replacement, calls the three resolvers BEFORE css_themes INSERT, populates the three FKs

Work item spec expanded with `layout_resolution` block (reason + candidates) and `typography_matched_existing` flag.

---

## 3. What We're Building Next: site-design-planner

### Why this exists

Current webdesign-agent workflow has `generate_css → deploy_css → update_site → check_should_fork → fork_theme → check_should_install → install_theme → complete`. The install step runs AFTER generate_css. On a fresh adopted site:

1. generate_css runs with no installed theme → falls back to `default` theme → renders brochure-formal layout with the site's colours
2. deploy_css commits that wrong-layout render to git
3. install_theme picks the correct layout (e.g. high-energy for a boxing site) and installs it
4. Nothing re-renders — the site stays on the brochure for the whole session

Two git commits minimum, the first knowingly wrong. Unacceptable.

### The fix: composition before render

`site-design-planner` is a new agent that resolves composition BEFORE webdesign-agent runs. It:
- Reads `identity`, `classification`, `design_intent`, `design_reference`, `mission` specs
- Picks layout_id by tag-overlap matching against `layouts.industry_tags` (deterministic scoring, not LLM)
- Picks typography_set_id by font_family exact match, archetype character match, or layout default
- Creates palette row from design_reference/design_intent/mission palette sources
- Creates css_themes + style_collections rows with all three FKs populated
- Updates `sites.style_collection_id`
- Writes `navigation`, `layout`, `resolved_composition` site_specs
- Hard-fails (`logger.Error`) if classification spec is missing — the classifier must have run

Then webdesign-agent just renders — reads composition via `site_context.css_theme_id`, merges design_spec overrides on top, deploys once. One commit. Right layout.

### Scope decisions (made this session)

1. **Brave backfill.** All existing sites get re-rendered through the new flow. No soft fallback for legacy themes. FK-null adopted themes get reset and redone. Sites looking uniformly brochure-like is the problem we're fixing — losing the uniform current output is the desired outcome.

2. **Hard-fail with loud logging.** When composition is missing at render time, `logger.Error` with explicit "site-design-planner did not run — orchestration broken" message. Don't silently fall through to `default`. Default stays as final fallback only to keep the process from panicking, not as a routine path.

3. **site-design-planner is a handler agent, not part of webdesign-agent's workflow.** Invoked via `needs_composition` work item. Dispatch loop picks it up. Follows the existing handler pattern (spawn → call).

4. **Adoption and new builds both go through it.** Principle #6 of the work plan: "Every build conceptually an adoption." Unified flow. Adoption's fingerprint data lives in site_specs, site-design-planner reads identically either way.

5. **Re-resolution is deferred.** First-install only for now. If a site's classification changes and composition needs re-resolving, that's HITL for the first version. Protects against silent overwrites.

6. **Fork-to-library pathway preserved but gated.** Existing `fork_theme` step stays, but the guard becomes `should_fork_theme AND should_promote_to_library` (both flags required). Prevents duplicate row creation after site-design-planner installs the composition. Library promotion itself is deferred — we keep the pathway open but don't implement the promotion UI yet.

7. **Webdesign-agent's install_theme step stays with its guard.** The guard `style_collection_id == null` means it short-circuits when site-design-planner has already installed. Belt and braces. Deferred removal until we're fully confident.

8. **Mission spec in adoption.** When adoption is triggered with mission fields in input_data, adoption-agent writes a `mission` spec (same aspect used by domain-submitter). Symmetric with new-build path.

---

## 4. Work Plan (staged, each individually verifiable)

### Deliverable 1: Diagnostic SQL — DONE this session
`001_diagnostic.sql` in outputs. Six queries:
- css_themes by origin + FK state
- Adopted themes detail (row-by-row)
- Sites by composition state (renderable / broken_fks / collection_no_theme / unlinked / inactive_theme)
- Sites with specs ready (identity / classification / design_reference / design_intent / archetype / mission)
- In-flight build work items (catches anything mid-pipeline)
- Style collections by origin

**Run this first**, share output. Tells us how many themes need reset, how many sites need re-adoption, whether there's mid-pipeline work to handle carefully.

### Deliverable 2: Three spec schemas
- `navigation` — architecture, primary_items, tools_strategy, cta, mobile pattern, sticky, logo_position
- `layout` — default_page_layout, header_style, footer_style, hero_nav_merged, page_overrides
- `resolved_composition` — palette_id/name, layout_id/name, typography_set_id/name references + lineage (what was fingerprint-sourced vs library vs mission) + reasoning

Delivered as SQL comment blocks + validation function. One reviewable document.

### Deliverable 3: site-design-planner agent definition
SQL INSERT into `agent_definitions`. Workflow steps:
- `validate_classification` — hard-fail if identity or classification spec missing
- `read_specs` — load all design-relevant site_specs
- `resolve_layout` — tag-overlap scoring, fallback to brochure-formal, emit `needs_new_layout_candidate` work item if no match
- `resolve_typography` — font-family match, archetype default
- `resolve_palette` — fingerprint → mission → design_intent priority
- `install_composition` — transactional INSERT palette + typography_set + css_themes + style_collections + UPDATE sites
- `write_navigation_spec` / `write_layout_spec` / `write_resolved_composition_spec`
- `complete`

`processing_mode: "task"`. Most steps are Go actions (deterministic); optional LLM step for generating navigation/layout from first-principles when design_intent is thin.

### Deliverable 4: Go actions (five new files)
- `resolve_composition_layout.go` — tag-overlap matcher against layouts.industry_tags, scores, alphabetical tiebreak, threshold-or-fallback
- `resolve_composition_typography.go` — font_family exact match, character match by `category`, archetype default, sans-modern last resort
- `resolve_composition_palette.go` — sources in priority (fingerprint → mission → design_intent → archetype default)
- `install_site_composition.go` — transactional INSERT palette + typography_set + css_themes + style_collections + UPDATE sites.style_collection_id
- `validate_classification_spec.go` — hard-error guard, emits clear log when failing

### Deliverable 5: Pipeline wiring
- `WriteBuildItemsAction` (in build-site-planner) writes `needs_composition` item priority 7, handler `site-design-planner`
- `needs_design` item gains `depends_on = [composition_item_id]`
- `apply_adoption_plan_action.go` emits same pair for adopted sites
- Adoption writes `mission` spec when input_data has mission fields (small addition to apply_adoption_plan)

### Deliverable 6: Renderer hard-fail + log.Error
Remove "default" silent fallback from `render_css_from_spec_action.go`. Loud `logger.Error` when `site_context.css_theme_id` is missing and no composition can resolve. Fallback-to-default stays only as a "don't panic" last resort, not a routine path.

### Deliverable 7: Site reset and redo sweep
SQL to reset existing sites' composition state; re-trigger build pipeline. Monitor outcomes, fix bugs as they surface.

---

## 5. Risks Flagged This Session

### Composition re-resolution ambiguity
If site-design-planner runs twice on the same site (triggered e.g. by classification update), the second run would try to create duplicate-named rows or silently overwrite. Mitigation: first run checks `sites.style_collection_id` — if non-null, runs in "re-resolve" mode requiring explicit HITL consent. Initial implementation: first-install only, re-resolve deferred.

### In-flight builds at deploy moment
Builds partway through the pipeline (`needs_site_plan` done, `needs_design` queued) without `needs_composition` in their queue will route around site-design-planner. Mitigation: diagnostic query #5 identifies these. Let them finish on old behaviour OR inject `needs_composition` manually. Decision deferred to post-diagnostic.

### FK backfill scope unknown
Adopted themes from before Phase 3 mapping have NULL FKs. Phase 4 renderer hard-errors on these. Decision this session: brave reset — delete/reset and force re-adoption. Actual count revealed by diagnostic query #2.

### Classifier must always run
site-design-planner hard-fails on missing classification. Scrape-only paths that bypass the classifier will fail. This is intentional — if classification is missing, the architecture is broken. No TLD heuristics; no generic fallback at that layer.

### Fork_theme step double-creation
Post-site-design-planner, fork_theme_from_site would create duplicate palette/typography/theme rows if triggered. Guard: require both `should_fork_theme` AND `should_promote_to_library` flags.

---

## 6. Principles Restated (for the record)

From `007_adoption_pipeline_v2.md` and `FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md`:

- **"Every build conceptually an adoption"** — new builds and adopted sites flow through the same pipeline; adoption just provides richer reference material
- **"Design reference is history, design intent is direction"** — design_reference spec is immutable historical record; design_intent is evolving semantic direction
- **"Adoption is a starting point, not a ceiling"** — the adopted state is baseline; improvement loop pushes toward aspirational state
- **"LLM for reasoning, Go for extraction"** — tag-overlap scoring is Go; design_intent generation is LLM
- **"Handlers are self-contained"** — site-design-planner receives site_id + domain, loads its own context, writes its own outputs

---

## 7. Files Reference

### Phase 1-5 (delivered, in outputs)
| File | Status |
|---|---|
| `/mnt/user-data/outputs/layouts/layout_01_*.sql` through `layout_15_*.sql` | Deployed |
| `/mnt/user-data/outputs/migration_025_phase2.sql` | Deployed |
| `/mnt/user-data/outputs/003_typography_sets_seed.sql` | Deployed |
| `/mnt/user-data/outputs/003_layouts_seed_bundled.sql` | Deployed |
| `/mnt/user-data/outputs/003_palettes_seed.sql` | Deployed |
| `/mnt/user-data/outputs/003_css_themes_mapping.sql` | Deployed |
| `/mnt/user-data/outputs/render_css_composition_helpers.go` | Deployed |
| `/mnt/user-data/outputs/render_css_composition_loader.go` | Deployed |
| `/mnt/user-data/outputs/render_css_from_spec_action.go` | Deployed |
| `/mnt/user-data/outputs/phase5/fork_theme_composition.go` | Deployed |
| `/mnt/user-data/outputs/phase5/fork_theme_from_site_action.go` | Deployed |
| `/mnt/user-data/outputs/PHASE_4_4_cleanup_summary.md` | Reference |

### Phase composition (delivered this session)
| File | Status |
|---|---|
| `/mnt/user-data/outputs/001_diagnostic.sql` | Ready to run |

### Pending deliverables (in order)
| # | File | Contents |
|---|---|---|
| 2 | Spec schemas SQL | navigation, layout, resolved_composition JSONB shapes |
| 3 | agent_site-design-planner.sql | agent_definitions INSERT |
| 4 | resolve_composition_*.go + install_site_composition.go + validate_classification_spec.go | Five Go actions |
| 5 | Pipeline wiring SQL + apply_adoption_plan patch | needs_composition work item, mission spec write |
| 6 | render_css_from_spec_action.go patch | Remove default fallback, loud error |
| 7 | Reset SQL + re-trigger plan | Brave redo sweep |

### Reference
| File | Purpose |
|---|---|
| `/mnt/project/production_agent-chassis-full_context.txt` | Main Go source (~117K lines) |
| `/mnt/project/bk_agent_definitions_backup.sql` | Agent workflow definitions |
| `/mnt/project/025_palette_layout_typography_migration.md` | Migration plan |
| `/mnt/project/021_site_spec_and_classifier.md` | Spec aspect model |
| `/mnt/project/007_adoption_pipeline_v2.md` | Adoption pipeline |
| `FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md` | Phase 3 decision (site-design-planner) |
| `FOCUS_navigation_HANDOFF_navigation_fix.md` | Navigation + layout spec schemas (draft) |

---

## 8. Things to Watch Out For

- Schema column reminders (from 028j handoff): `site_work_items`: `pipeline` not `domain`; `site_specs`: `data` not `spec_data`; `agent_definitions`: `type` not `agent_type`, `default_config` not `config`, always `AND deleted_at IS NULL`
- Don't use `logger.Debug` — it won't show. Use `logger.Info` / `logger.Warn` / `logger.Error`
- Kubernetes namespace is `-n ai-persona-system` for main pods, `-n kafka` for Kafka
- Cluster name is `personae-kafka-cluster-kafka-bootstrap`
- pgx/v5 accepts `[]string` for text[] columns directly — no `pq.Array` wrap needed
- Postgres identifiers: lowercase/underscore/digits only, 63 char max (validated by `isValidIdentifier` helper)
- Adoption pipeline still has the `analyze_site` JSON truncation bug (workaround: `{{.adoption_analysis}}` dumps raw wrapper). Long-term fix: split `analyze_site` into focused steps. Not blocking for site-design-planner.

---

## 9. Immediate Next Step

Run `001_diagnostic.sql` and share the output. The numbers determine:
- Whether the brave reset is 5 themes or 500
- Whether existing sites have their adoption specs intact (or need re-adoption first)
- Whether there's mid-pipeline work that needs careful handling

Once we have the diagnostic output, deliverable 2 (spec schemas) follows.

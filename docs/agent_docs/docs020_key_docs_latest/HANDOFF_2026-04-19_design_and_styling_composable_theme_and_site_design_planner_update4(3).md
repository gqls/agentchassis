# HANDOFF: Design & Styling — Composable Theme + site-design-planner

## Date: 2026-04-18 (update 2)
## Supersedes: `HANDOFF_2026-04-18_design_and_styling_composable_theme_and_site_design_planner.md` (update 1)
## Original predecessor: `FOCUS_design_and_styling_adoption_HANDOFF_2026-04-17.md`
## Session transcripts:
- `/mnt/transcripts/2026-04-17-*-composable-theme-deployment.txt`
- `/mnt/transcripts/2026-04-18-13-02-36-composable-theme-migration-025.txt` (Phases 4-5 build)
- Current session (composition flow redesign, site-design-planner build-out)

---

## 1. Status Summary

Migration `025_palette_layout_typography_migration` is through Phase 5 in code. Phases 1-3 (data model, layouts, seeding) are deployed and verified. Phases 4-5 (renderer cutover, fork action rewrite) were deployed but not end-to-end verified — a smoke-test run was incomplete.

Diagnostic this session revealed: zero adopted themes exist, no FK-broken themes to reset, but 5 of 7 sites were missing `classification` spec. A one-shot backfill run the full classifier over those 5 sites — now all 7 have `classification`. A stub `test-broken` theme with NULL FKs was deleted manually.

Two decisions refined scope:
- **site-design-planner writes only `resolved_composition`** (Choice B). Navigation and layout specs are deferred to future specialist agents (nav-planner, design-density). This preserves the "slim strict responsibilities" principle.
- **validate_composition_inputs both loud-logs AND queues a recovery work item** on miss. The log is the immediate signal; the work item is a durable signal visible in dashboards AND a self-heal — the classifier runs, writes the missing spec, and the dependent composition item re-dispatches. If a site repeatedly hits this path, the two-strike rule marks the item `unresolved` and it becomes visible for investigation.

---

## 2. What's Deployed and Verified

### Phase 1 — Layouts seeded (15 rows in `layouts` table)
15 layout CSS templates produced, reviewed against contracts, seeded to DB.
Contract decisions applied: heading fallback `var(--section-heading, var(--color-primary))`, no `--section-*` declarations on section containers, 5 surface-painted classes.

### Phase 2 — Schema migration deployed
`palettes`, `layouts`, `typography_sets` tables created with lineage columns. `css_themes` and `style_collections` gained `palette_id`, `layout_id`, `typography_set_id` FKs (nullable, `ON DELETE SET NULL`).

### Phase 3 — Seeding deployed and verified
6 typography_sets, 15 layouts, 13 palettes extracted from existing themes, 14 css_themes rows fully linked. Verified: all 14 seed themes have all three FKs populated.

### Phase 4 — Renderer cutover (deployed but UNVERIFIED end-to-end)
Three files deployed: `render_css_composition_helpers.go`, `render_css_composition_loader.go`, `render_css_from_spec_action.go`. Loader hard-errors on NULL FKs. Structured log emitted for spec→theme overrides.

### Phase 5 — Fork action rewrite (deployed but UNVERIFIED end-to-end)
Two files deployed: `fork_theme_composition.go`, `fork_theme_from_site_action.go`. Populates three FKs via resolvers before css_themes INSERT.

### Diagnostic + classification backfill (this session)
- Diagnostic `001_diagnostic.sql` + followup `002_followup_diagnostic.sql` run; output shared.
- `test-broken` theme (NULL FKs, created today as a test row) manually deleted.
- `003_queue_classifier_backfill_v2.sql` queued `needs_domain_research` for 5 sites.
- Classifier completed all 5 successfully. All 7 live sites now have `classification`.
- vonc.com and gamedesign.uk specs potentially clobbered by classifier: vonc.com was expendable (user confirmed), gamedesign.uk will be re-adopted from a refreshed source site anyway.

### Spec schemas applied
- `004_spec_schemas.sql` loaded. Three validators created: `validate_navigation_spec`, `validate_layout_spec`, `validate_resolved_composition_spec`. Self-tests pass.
- These are JSONB-shape validators. `site_specs.data` stays as open JSONB. Validators are app-side, called by site-design-planner before writes.

---

## 3. Scope Refinement: site-design-planner Writes Only `resolved_composition`

**Choice B adopted.** The agent's exclusive responsibility is composition resolution (which palette / layout / typography this site gets). It writes one spec:

- `resolved_composition` — palette_id, layout_id, typography_set_id, lineage (fingerprint vs library vs mission), reasoning.

**It does NOT write** `navigation` or `layout` specs. Those are reserved for future specialist agents:
- nav-planner (owns navigation.primary_items, tools_strategy, cta, architecture)
- design-density agent (owns layout.section_density, spacing)

Downstream readers (populate_nav_tables, render_site_components, InjectHeader) already have fallback logic that works without those specs. When the specialists exist they'll enrich.

The `navigation` and `layout` spec schemas and validators in `004_spec_schemas.sql` stay — they're the contract those future agents will honour. They're just not this agent's output.

### Remaining scope (what site-design-planner does)

- Reads `identity`, `classification`, `design_intent`, `design_reference`, `site_archetype`, `mission`
- Picks layout_id by tag-overlap matching against `layouts.industry_tags` (deterministic, not LLM)
- Picks typography_set_id by font_family exact match, archetype character match, or layout default
- Creates palette row from fingerprint / mission / design_intent priority cascade
- Creates `css_themes` + `style_collections` rows with all three FKs populated
- Updates `sites.style_collection_id`
- Writes `resolved_composition` spec
- Optionally emits `needs_new_layout_candidate` work item when layout match is weak (library growth signal)

If `identity` or `classification` is missing, `logger.Error` loudly and return not-ready. No work item creation, no self-heal.

---

## 4. Work Plan (updated)

### Deliverable 1: Diagnostic SQL — DONE
`001_diagnostic.sql`, `002_followup_diagnostic.sql` in outputs. Run. Showed: 0 adopted themes, 1 test theme (deleted), 5 sites missing classification (now backfilled).

### Deliverable 2: Spec schemas — DONE
`004_spec_schemas.sql` applied. Validators created and tested.

### Deliverable 3: Classifier backfill — DONE
`003_queue_classifier_backfill_v2.sql` queued, all 5 classifier runs completed.

### Deliverable 4: Go actions for site-design-planner (in progress)

Five actions + shared helpers, in order:

1. **`validate_composition_inputs_action.go`** — DONE (~305 lines). Reads identity + classification. Returns ready/not-ready. Loud-log on miss AND queues `needs_domain_research` work item for dashboard visibility + dispatch-loop self-heal.
2. **`resolve_composition_layout_action.go`** — DONE (~357 lines). Thin wrapper over shared `resolveLayoutByTags` (renamed from `resolveLayoutForFork` in fork_theme_composition.go). Emits `needs_new_layout_candidate` work item on fallback (handler `hitl-review`, status `needs_human_review`).
3. **`resolve_composition_typography_action.go`** — DONE (~234 lines). Thin wrapper over shared `resolveTypographySet` (factored out of `resolveTypographySetForFork`). Priority cascade: design_reference → design_intent → mission.preferred_typography → sans-modern default.
4. **`resolve_composition_palette_action.go`** — DONE (~342 lines). Thin wrapper over shared `createPalette` (factored out of `createPaletteForFork`). Priority cascade: design_reference → mission.preferred_palette → design_intent → layout-library-inherit → default palette.
5. **`resolve_composition_helpers.go`** — DONE (~243 lines). Shared helpers used across all composition resolvers: `loadSpecAspectFromContext`, `readClassificationFromContext`, `extractReferenceValuesFromSpec`, `mapInterfaceToStrings`, `slugifyForCompositionName`.
6. **`install_site_composition_action.go`** — DONE (~562 lines). Transactional wrapper: css_themes INSERT + style_collections INSERT + guarded UPDATE sites.style_collection_id + inline write of `resolved_composition` spec (supersede + insert pattern, preserves the site_specs history contract). All four writes are atomic — either the whole composition lands or none of it does. No separate `write_site_spec` step needed.

**Orphan policy**: Each resolver (palette, typography) commits its insert in its own transaction before install runs. If install fails, those rows become orphans — adopted-origin rows with no css_theme referencing them. Rather than track and roll back manually, we extend the existing `database-cleanup` scheduled task to sweep them. Draft SQL in `draft_composition_orphan_cleanup.sql` — to be deployed alongside or shortly after install goes live.

Three patch documents accompany the resolvers — each factors out a shared helper from the Phase 5 fork code:

- `patch_fork_theme_composition_shared_resolver.go` — layout resolver rename + IsFallback field
- `patch_fork_theme_composition_typography_shared.go` — factor out `resolveTypographySet` 
- `patch_fork_theme_composition_palette_shared.go` — factor out `createPalette` (takes origin + needs_review as params)

All three patches preserve existing fork behaviour (call-site signatures unchanged).

### Deliverable 5: site-design-planner agent definition
SQL INSERT into `agent_definitions`. Workflow wires the five actions above. `processing_mode: "task"`.

### Deliverable 6: Pipeline wiring
- `WriteBuildItemsAction` writes `needs_composition` item priority 7, handler `site-design-planner`
- `needs_design` gets `depends_on = [composition_item_id]`
- `apply_adoption_plan_action.go` emits the same pair for adopted sites
- Adoption writes `mission` spec when input_data has mission fields

### Deliverable 7: Renderer hard-fail tightening
`render_css_from_spec_action.go` — remove "default" silent fallback from normal path. Loud `logger.Error` when `site_context.css_theme_id` is missing. Default stays only as a last-resort "don't panic" fallback, not a routine path.

### Deliverable 8: Redo sweep
SQL to reset existing sites' composition state; trigger re-build. Three sites (gamedesign.uk, robot-hands.com, vonc.com) currently `unlinked`; the four renderable sites (ai-agent-orchestration.com, finetuning.uk, gaswholesalers.com, leopardessconsulting.co.uk) will be re-composed through the new pipeline for consistency.

### Phase 4.5 (separately scheduled)
Decouple surface-painting from layouts via `data-section-bg="surface"` attribute. Deferred from Phase 1.

### Future: HITL review handler agent
`needs_new_layout_candidate` currently queues with `handler_agent: "hitl-review"` and `status: "needs_human_review"`. `hitl-review` is a convention, not a registered agent — the item is visible in the queue but no dispatch agent processes it. A human resolves via admin API (PATCH /work-items/:id). Next natural opportunity: build the `hitl-review` handler agent proper, likely after composition is working end-to-end and we see the item type showing up in practice. Same handler could serve other `needs_human_review` items from tool-auditor, content validation, etc.

### Phase 6 (post-4.5)
Rewrite contracts doc with new three-entity model.

### Phase 7 (post-stabilisation)
Drop legacy columns (`css_template`, `css_content`, `color_palette`, `typography`) from `css_themes`. Remove `css_templating.go`.

---

## 5. Decisions Made This Session

1. **Choice B for site-design-planner scope.** Writes only `resolved_composition`. Navigation and layout specs deferred to future specialist agents.

2. **Dual signal on missing classification in validate_composition_inputs.** Loud `logger.Error` for immediate visibility AND a `needs_domain_research` work item for durable signal + self-heal. Repeated failures on the same site accumulate via the two-strike rule and become `unresolved` — visible to dashboards. The classifier is queued as recovery; if it runs, the composition item re-dispatches and the site builds without manual intervention.

3. **Classifier backfill complete.** All 7 sites now have classification. Old specs preserved in `site_specs` history (versioning already gives free rollback via `is_current` flip).

4. **vonc.com and gamedesign.uk specs are acceptable casualties.** vonc.com was reviewed and ruled not worth preserving. gamedesign.uk will be re-adopted from a refreshed source.

5. **No TLD heuristics or generic fallback at composition layer.** Classification must exist. If it's missing, something upstream is broken.

6. **Brave redo sweep accepted earlier in session** — all existing sites will go through the new composition pipeline even though most render today. The current uniform-brochure look is the problem we're fixing.

---

## 6. Risks Still Live

### Composition re-resolution ambiguity
If site-design-planner runs twice on the same site (e.g. classification updated), the second run would try to create duplicate-named rows. Mitigation: first run checks `sites.style_collection_id`. If non-null, runs in "re-resolve" mode which is deferred — HITL only for now.

### In-flight builds at deploy moment
Diagnostic query showed: 3 items on gamedesign.uk, 2 on finetuning.uk mid-pipeline. Small enough to handle manually at cutover — either let them finish on old behaviour or inject `needs_composition` items.

### Fork_theme step double-creation
Post-site-design-planner, the existing `fork_theme` step would create duplicate rows if triggered. Guard: require both `should_fork_theme` AND `should_promote_to_library` flags. Implementation deferred to Deliverable 6 (pipeline wiring).

---

## 7. Principles Restated

From `007_adoption_pipeline_v2.md`, `FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md`, and session discussion:

- **"Every build conceptually an adoption"** — new builds and adopted sites flow through the same pipeline.
- **"Design reference is history, design intent is direction"** — design_reference is immutable, design_intent is evolving.
- **"Adoption is a starting point, not a ceiling"** — improvement loop pushes toward aspirational state.
- **"LLM for reasoning, Go for extraction"** — tag-overlap scoring is Go; design_intent generation is LLM.
- **"Handlers are self-contained"** — site-design-planner receives site_id + domain, loads its own context.
- **"Slim strict responsibilities"** (session) — each agent owns one thing. Composition is site-design-planner's. Navigation is a future specialist's.

---

## 8. Files Reference

### Delivered and deployed
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

### Delivered this session
| File | Status |
|---|---|
| `/mnt/user-data/outputs/001_diagnostic.sql` | Run, output shared |
| `/mnt/user-data/outputs/002_followup_diagnostic.sql` | Run, output shared |
| `/mnt/user-data/outputs/003_queue_classifier_backfill_v2.sql` | Run, 5 sites backfilled |
| `/mnt/user-data/outputs/004_spec_schemas.sql` | Applied, validators created |
| `/mnt/user-data/outputs/validate_composition_inputs_action.go` | Ready for deploy |
| `/mnt/user-data/outputs/resolve_composition_layout_action.go` | Ready for deploy |
| `/mnt/user-data/outputs/resolve_composition_typography_action.go` | Ready for deploy |
| `/mnt/user-data/outputs/resolve_composition_palette_action.go` | Ready for deploy |
| `/mnt/user-data/outputs/resolve_composition_helpers.go` | Ready for deploy |
| `/mnt/user-data/outputs/install_site_composition_action.go` | Ready for deploy |
| `/mnt/user-data/outputs/draft_composition_orphan_cleanup.sql` | Draft — deploy alongside install |
| `/mnt/user-data/outputs/patch_fork_theme_composition_shared_resolver.go` | Patch to apply to fork_theme_composition.go |
| `/mnt/user-data/outputs/patch_fork_theme_composition_typography_shared.go` | Patch to apply to fork_theme_composition.go |
| `/mnt/user-data/outputs/patch_fork_theme_composition_palette_shared.go` | Patch to apply to fork_theme_composition.go |
| `/mnt/user-data/outputs/005_agent_site_design_planner.sql` | Agent definition ready to apply |
| `/mnt/user-data/outputs/patch_write_build_items_composition.go` | Patch to load_work_item_actions.go — inserts needs_composition + depends_on |
| `/mnt/user-data/outputs/patch_apply_adoption_plan_composition.go` | Patch to apply_adoption_plan_action.go — mission spec write + composition insert |
| `/mnt/user-data/outputs/patch_render_css_theme_resolution.go` | Patch to renderer — site-composition auto-resolve + emergency-fallback loud log |
| `/mnt/user-data/outputs/006_clear_webdesign_generate_css_theme_literal.sql` | Clear hardcoded theme_name literal on webdesign-agent — unlocks renderer auto-resolve |

### Pending deliverables (in order)
| # | File | Contents |
|---|---|---|
| 6.5 | update_database_cleanup_task.sql | Merge orphan cleanup CTEs into database-cleanup pre_query |
| 8 | reset_and_redo_sweep.sql | Existing sites through new pipeline |
| 9 | webdesign_agent_reorder_install_before_render.sql | Move `install_theme` to run BEFORE `generate_css` in webdesign-agent (closes the emergency-fallback-then-install ordering bug — see section 9.1) |
| future | hitl-review agent definition | Handler for needs_human_review items |

### 9.1 Known ordering issue in webdesign-agent (to fix in deliverable 9)

**Problem.** webdesign-agent's current workflow runs `generate_css → deploy_css → ... → check_should_install → install_theme`. The render happens BEFORE the install. So for a site where site-design-planner did NOT run (scrape-only path, orchestration bug, pre-cutover in-flight builds), the sequence is:

1. `generate_css` has no composition to resolve → renderer hits the emergency fallback → renders with `standard-brochure` layout + merged design_spec colours
2. `deploy_css` commits that wrong-layout CSS to git
3. `install_theme` runs later and creates the correct composition from design_spec
4. Nothing re-renders — the wrong-layout commit stays in git, and the newly-installed composition is ignored until the next webdesign-agent pass

This is the exact "first render wrong layout" bug site-design-planner was built to eliminate. While site-design-planner is handling the normal path, webdesign-agent's fallback path still has the old shape.

**Fix (deliverable 9).** Reorder webdesign-agent's workflow so `install_theme` runs BEFORE `generate_css`:

```
analyze_design → update_site → check_should_install → install_theme (if no collection)
                                                      → generate_css → deploy_css
                                                      → check_should_fork → fork_theme
                                                      → complete
```

With this order, the renderer always finds a composition to resolve (either one site-design-planner installed, or one webdesign-agent just installed from the design_spec). Emergency-fallback path becomes genuinely unreachable in the normal webdesign path.

**Scope.** JSONB workflow edit in `agent_definitions.default_config`. No Go changes. Small, reversible.

**Why deferred to deliverable 9.** The renderer patch (7) plus the workflow config clear (7.5) are the minimum viable cutover: sites WITH composition render correctly. The order bug only affects sites WITHOUT composition, which should be a shrinking set once the pipeline routes everything through site-design-planner. Worth fixing — just not on the critical cutover path.

### Reference
| File | Purpose |
|---|---|
| `/mnt/project/production_agent-chassis-full_context.txt` | Main Go source (~117K lines) |
| `/mnt/project/bk_agent_definitions_backup.sql` | Agent workflow definitions |
| `/mnt/project/025_palette_layout_typography_migration.md` | Migration plan |
| `/mnt/project/021_site_spec_and_classifier.md` | Spec aspect model |
| `/mnt/project/007_adoption_pipeline_v2.md` | Adoption pipeline |
| `FOCUS_design_and_styling_adoption_WORK_PLAN_v2.md` | Phase 3 decision (site-design-planner) |
| `FOCUS_navigation_HANDOFF_navigation_fix.md` | Navigation spec schema reference |

---

## 9. Things to Watch Out For

- Schema column reminders: `site_work_items`: `pipeline` not `domain`; `site_specs`: `data` not `spec_data`; `agent_definitions`: `type` not `agent_type`, `default_config` not `config`, always `AND deleted_at IS NULL`.
- `site_work_items` partial unique index: `(site_id, item_key) WHERE item_key IS NOT NULL AND status NOT IN ('complete','verified','rejected','wont_fix','failed')`. `ON CONFLICT` must match this predicate exactly and use `DO NOTHING`.
- Don't use `logger.Debug` — use `logger.Info`/`logger.Warn`/`logger.Error`.
- Kubernetes namespaces: `-n ai-persona-system`, `-n kafka`. Cluster: `personae-kafka-cluster-kafka-bootstrap`.
- pgx/v5 accepts `[]string` for text[] columns directly.
- Postgres identifiers: lowercase/underscore/digits, 63 char max.
- `write_site_spec` always deep-merges and preserves history via is_current/superseded_at. Rollback is `UPDATE is_current` on two rows, no separate backup mechanism needed.
- Classifier re-runs overwrite identity, classification, content_direction, design_intent (deep-merged). mission and design_reference are untouched. Old versions remain accessible via is_current=false rows.
- Adoption pipeline has known `analyze_site` JSON truncation bug. Workaround: `{{.adoption_analysis}}` dumps raw. Not blocking site-design-planner.

---

## 10. Immediate Next Step

Deliverable 8: `reset_and_redo_sweep.sql` — the "redo the existing sites through the new pipeline" SQL. Three sites are currently unlinked (gamedesign.uk, robot-hands.com, vonc.com); four have style collections linked via the old webdesign-agent install path (ai-agent-orchestration.com, finetuning.uk, gaswholesalers.com, leopardessconsulting.co.uk).

Options for the four already-linked sites:
- **(a)** Leave them alone — their existing compositions will be kept; improvement loop can eventually refresh them.
- **(b)** Reset sites.style_collection_id to NULL and queue fresh `needs_composition` items — forces all sites through site-design-planner.

(b) is more work but gives us a uniform state to validate against. For the initial rollout, (b) catches any resolver bugs that (a) would hide. Worth doing once, probably behind a DO block that reports what it touched.

Deliverable 6.5 (orphan cleanup merged into database-cleanup) can land any time — small, self-contained, doesn't block anything.

## Rollout order (summary)

For clarity, the full deploy sequence is:

1. **Go code**: apply the three fork-composition patches, deploy the six new Go actions + install, plus the renderer patch. Redeploy agent-chassis.
2. **Register new agent**: run `005_agent_site_design_planner.sql`.
3. **Wire pipeline**: apply `patch_write_build_items_composition.go` and `patch_apply_adoption_plan_composition.go` (these are Go-source patches — redeploy again).
4. **Flip the renderer**: run `006_clear_webdesign_generate_css_theme_literal.sql`. After this, sites with composition render correctly; sites without composition hit the loud emergency-fallback path.
5. **Redo sweep**: run deliverable 8 to reset the four linked sites + kick the three unlinked sites through the new pipeline.
6. **Cleanup**: run deliverable 6.5 (orphan sweep merged into scheduled task) whenever convenient.

Step 4 is the cutover moment. Steps 1–3 are safe to deploy in any order as long as they all land before step 4.

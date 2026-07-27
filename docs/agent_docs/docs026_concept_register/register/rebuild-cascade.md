# Register — rebuild-cascade

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

7 concepts, consolidated from 7 raw extractions across unit U07 (no internal
duplication found; all seven are distinct facets of the same component-regen →
rerender-propagation investigation).

### REB-001 — F3 scoped reason-stamped rerender (dependent-page scoping + reason propagation)
- **status:** deployed
- **status-evidence:** "F3 PROVEN END-TO-END (scoping to exactly the five, reason propagation, gate, section-rerender execution)"; fleet on v1.0.1088.
- **what:** Three coupled changes making a component regen's triggered re-render actually repair sections, scoped to the blast radius: F3a — `create_rerender_items` accepts `reason`+`component_id`; when reason ∈ {section_data_resolved, image_landed} AND component_id is set, it queries the component's dependent pages (page_components JOIN pages) and creates reason-stamped `page_rerender` items only for those; no signals → unchanged assemble-only all-pages behaviour. F3b — store re-adds `reason: section_data_resolved` to its needs_rerender spec. F3c — rerender-pages step config maps reason/component_id from `input_data.spec`. Either half alone degrades safely to assemble-only. Accepted residual: `rerender_page_sections` re-renders ALL sections of each dependent page (a documented blast radius that later recurred on another site).
- **sources:** NOTES(43).md §9b, §9p-§9r, §9v; RUNBOOK(49).md Part A; w4b_03_read_rerender_config.sql
- **relations:** assemble-only vs section re-render (REB-002); two re-render paths (page-build-pipeline register, PBP-022, PBP-013)
- **verify-later:** create_rerender_items_action.go (InputSpec reason/component_id; dependent scoping; rows.Close before INSERT loop); rerender-pages default_config create_rerender_items step

### REB-002 — Assemble-only vs section re-render distinction (the `reason` gate)
- **status:** deployed
- **status-evidence:** "a bare needs_rerender is assemble-only and just re-ships the empty shell"; gate `reason ∈ {image_landed, section_data_resolved}` confirmed from rerender_page_sections source.
- **what:** Two fundamentally different re-render depths: an assemble-only pass re-assembles/re-ships existing `rendered_html` (cannot repair content), while a section re-render (`rerender_page_sections`, gated on reason image_landed/section_data_resolved) regenerates each section's HTML from stored `content_data` against the current template, no LLM. A `reason` dropped anywhere along the chain silently downgrades to assemble-only — an inert "fix" caught only by checking the consuming step's config. Central operational lesson of the investigation.
- **sources:** NOTES(43).md §2 Correction 3, §6, §9a; PLAN(1).md Phase 4; RUNBOOK_pre_cleanup_backup.md §The-problem
- **relations:** F3 scoped rerender (REB-001, makes the reason survive); rerender fossilisation (REB-005, site-level analogue); carry-forward path (REB-003)
- **verify-later:** rerender_page_sections_action.go reason gate; page-rerender workflow's routing on spec.reason

### REB-003 — Carry-forward path and the carry fingerprint
- **status:** deployed
- **status-evidence:** "per-page save-style row updates + identical bytes = the rerender_page_sections CARRY path fingerprint," confirmed by data.
- **what:** In `rerender_page_sections`, a section that fails `planSection` readiness (unresolvable required field, missing component, empty template) is carried: `carryStoredSection` re-emits the stored HTML, save_sections writes it back with a fresh per-page `updated_at` but identical bytes. Protective against shipping worse output, but it re-fossilises stale/empty renders forever when readiness is permanently blocked — and its diagnostic signature (fresh distinct timestamps + one shared md5) is how a recovery stall was pinned to a readiness field rather than the rerender chain itself.
- **sources:** NOTES(43).md §9u, §9v; BUNDLE(3).md §3
- **relations:** section readiness model (page-build-pipeline register, PBP-003, the gate it consults); auto-escalation (REB-004); recovery playbook
- **verify-later:** rerender_page_sections_action.go: planSection, carryStoredSection, NULL pre-check

### REB-004 — Auto-escalation: empty content_data → needs_page writer rebuild
- **status:** deployed
- **status-evidence:** "Auto-escalations fired as designed: page-rerender created needs_page for matchmatrix (17:31) and gripper-selection-guide (17:37); writer rebuilt gripper-selection-guide 18:04."
- **what:** `rerender_page_sections`' NULL pre-check escalates a whole page to a `needs_page` item (handler page-build-handler, spec `{reason: content_data_backfill, page_name}`) when any section lacks `content_data` — routing un-re-renderable pages to a full writer rebuild instead of carrying garbage. Deliberately exploited during a recovery both as the rebuild mechanism and as the source of a correctly-shaped needs_page item to clone (never guess a work-item spec).
- **sources:** NOTES(43).md §9ap, §9aw, §9bd; RUNBOOK(49).md Part C Steps 8-9
- **relations:** carry-forward path (REB-003, its sibling branch); work-item spec-cloning discipline
- **verify-later:** rerender_page_sections NULL pre-check/escalate branch; page-build-handler content_data_backfill flow

### REB-005 — Rerender fossilisation (reassembly re-ships stale renders; template changes need a full rebuild)
- **status:** deployed
- **status-evidence:** "deployed hero consumes legacy var(--accent-color…) — sections are stale old-template renders… A full page-build-handler rebuild is required; needs_rerender would re-fossilise," evidenced directly from deployed HTML.
- **what:** The rerender handler reassembles stored `page_components.rendered_html` and injects `site_components.rendered_html` — it does not re-render component templates — so sites can live indefinitely on reassemblies of early renders while the library advances (one site served renders of long-inactive components for weeks). Template/library changes only reach pages through a section re-render or a full page-build-handler rebuild; chrome only through render_site_components. Settled a documented cross-doc contradiction by direct evidence — the same finding independently confirmed on a different site in the page-build-pipeline register (PBP-001).
- **sources:** RUNBOOK_scheme_to_components(18).md Check 3/4 RESULTS; running_notes_scheme_to_components(22).md #Sh, #So
- **relations:** rebuild vs rerender semantics (page-build-pipeline register, PBP-001, same mechanism independently confirmed); assemble-only vs section re-render (REB-002); chrome refresh gating (REB-006)
- **verify-later:** rerender_single_page_action.go; which paths run RenderComponent vs re-ship stored HTML

### REB-006 — Chrome refresh gating (render_site_components, force_rerender, repoint-before-rerender)
- **status:** deployed
- **status-evidence:** "renderAndStoreSiteComponent joins the PINNED site_components.component_id with no is_active filter, and without force_rerender SKIPS non-empty slots" (code fact); a repoint was executed and confirmed (UPDATE 1 ×2).
- **what:** Site chrome (header/footer/head) is rendered and stored by `render_site_components`, a path separate from page builds (page-build-handler is not among its invoking agents — a page rebuild never refreshes chrome). Its join honours the pinned `component_id` with no `is_active` filter and skips non-empty slots unless `force_rerender: true` (only rerender-pages v6 passes it, gated on `spec.refresh_site_components` — page-build-pipeline register PBP-002). Operational consequence: repoint the pinned rows to the fixed/active components BEFORE forcing the re-render, and refresh stored chrome before a rebuild so later automatic rerenders can't re-inject stale dark chrome.
- **sources:** RUNBOOK_scheme_to_components(18).md W4b + STEP-1 RESULTS; running_notes(22).md §Sy-Ta; w4b_03_read_rerender_config.sql
- **relations:** rerender-pages v6 workflow (page-build-pipeline register, PBP-002); rerender fossilisation (REB-005); F3 scoped rerender (REB-001, same agent)
- **verify-later:** render_site_components_action.go:345-430; rerender-pages v6 check_refresh_components gate

### REB-007 — page-build-handler writes only planned sections (sections=0 → silent no-op rebuild)
- **status:** deployed
- **status-evidence:** "matchmatrix sections_listed=0 vs siblings 9/5 ⇒ the page-build-handler iterates the planned section list and had nothing to write — planning-data gap on a legacy page — PARKED."
- **what:** The writer rebuild flow iterates the page's planned section list; a legacy page with an empty `pages.sections` plan completes its `needs_page` item without writing anything — a silent no-op rebuild indistinguishable from success at the item level (a double no-op was observed before diagnosis; a page_type-based hypothesis was falsified first). Remediation options (planner reconcile/adopt, or retire the page) parked as hygiene. Distinct from, but the same failure family as, the complete_error silent-success family (page-build-pipeline register, PBP-020).
- **sources:** NOTES(43).md §9bf-§9bh; RUNBOOK(49).md Step 12(a)
- **relations:** auto-escalation (REB-004); complete_error silent-success family (page-build-pipeline register, PBP-020); planner as another item producer
- **verify-later:** page-build-handler section iteration; matchmatrix pages.sections

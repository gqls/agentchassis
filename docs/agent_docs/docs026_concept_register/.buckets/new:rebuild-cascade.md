
<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F3 scoped reason-stamped rerender (dependent-page scoping + reason propagation)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "F3 PROVEN END-TO-END (scoping to exactly the five, reason propagation, gate, section-rerender execution)" (NOTES §9v); fleet on v1.0.1088 (§9r).
- **what:** Three coupled changes making a component regen's triggered re-render actually repair sections, scoped to the blast radius: F3a — create_rerender_items accepts `reason`+`component_id`; when reason ∈ {section_data_resolved, image_landed} AND component_id set, it queries the component's dependent pages (page_components JOIN pages) and creates reason-stamped `page_rerender` items only for those; no signals → unchanged assemble-only all-pages behaviour. F3b — store re-adds `reason: section_data_resolved` to its needs_rerender spec. F3c — rerender-pages step config maps `reason`/`component_id` from `input_data.spec`. Either half alone degrades safely to assemble-only. Accepted residual: rerender_page_sections re-renders ALL sections of each dependent page (documented blast radius that later stamped the gauntlet onto robot-hands).
- **sources:** NOTES(43).md §9b, §9p–§9r, §9v; RUNBOOK(49).md Part A; w4b_03_read_rerender_config.sql
- **relations:** assemble-only vs section re-render; F8 step 4 reused it; F6 (dedup/counter flaws in the same action).
- **verify-later:** create_rerender_items_action.go (InputSpec reason/component_id; dependent scoping; rows.Close before INSERT loop); rerender-pages default_config create_rerender_items step (5 keys).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Assemble-only vs section re-render distinction (the `reason` gate)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "a bare needs_rerender is assemble-only and just re-ships the empty shell" (NOTES §6); gate `reason ∈ {image_landed, section_data_resolved}` confirmed from rerender_page_sections source (§2 Correction 3).
- **what:** Two fundamentally different re-render depths: an assemble-only pass re-assembles/re-ships existing `rendered_html` (cannot repair content), while a section re-render (`rerender_page_sections`, gated on reason image_landed/section_data_resolved) regenerates each section's HTML from stored `content_data` against the current template, no LLM. A reason dropped anywhere along the chain (as rerender-pages originally did) silently downgrades to assemble-only — an inert "fix" that was caught only by checking the consuming step's config. Central operational lesson of the thread.
- **sources:** NOTES(43).md §2 Correction 3, §6, §9a; PLAN(1).md Phase 4; RUNBOOK_pre_cleanup_backup.md §The problem
- **relations:** F3 (makes the reason survive); rerender fossilisation (site-level analogue); carry-forward path.
- **verify-later:** rerender_page_sections_action.go reason gate; page-rerender workflow's routing on spec.reason.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Carry-forward path and the carry fingerprint
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "per-page save-style row updates + identical bytes = the rerender_page_sections CARRY path fingerprint" (NOTES §9u), confirmed by data §9v.
- **what:** In rerender_page_sections, a section that fails `planSection` readiness (unresolvable required field, missing component, empty template) is carried: `carryStoredSection` re-emits the stored HTML, save_sections writes it back with a fresh per-page `updated_at` but identical bytes. Protective against shipping worse output, but it re-fossilises stale/empty renders forever when readiness is permanently blocked — and its diagnostic signature (fresh distinct timestamps + one shared md5) is how the recovery stall was pinned to cta_url readiness rather than the rerender chain.
- **sources:** NOTES(43).md §9u, §9v; BUNDLE(3).md §3 (rerender partly protective)
- **relations:** section readiness model (the gate it consults); F5 (the added-required-field cause); recovery playbook.
- **verify-later:** rerender_page_sections_action.go: planSection, carryStoredSection, NULL pre-check.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Auto-escalation: empty content_data → needs_page writer rebuild
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "Auto-escalations fired as designed: page-rerender created needs_page for matchmatrix (17:31) and gripper-selection-guide (17:37); writer rebuilt gripper-selection-guide 18:04" (NOTES §9aw).
- **what:** rerender_page_sections' NULL pre-check escalates a whole page to a `needs_page` item (handler page-build-handler, spec `{reason: content_data_backfill, page_name}`) when any section lacks `content_data` — routing un-re-renderable pages to a full writer rebuild instead of carrying garbage. Deliberately exploited during F8 recovery both as the rebuild mechanism and as the source of a correctly-shaped needs_page item to clone (never guess a spec).
- **sources:** NOTES(43).md §9ap, §9aw, §9bd; RUNBOOK(49).md Part C Steps 8–9
- **relations:** carry-forward path (its sibling branch); work-item spec-cloning discipline; matchmatrix planning-data gap (an escalation that no-ops).
- **verify-later:** rerender_page_sections NULL pre-check/escalate branch; page-build-handler content_data_backfill flow.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Rerender fossilisation (reassembly re-ships stale renders; template changes need full rebuild)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "deployed hero consumes legacy var(--accent-color…) — sections are stale old-template renders… A full page-build-handler rebuild is required; needs_rerender would re-fossilise" (RUNBOOK_scheme_to_components(18) Check 4a, evidenced from deployed HTML).
- **what:** The rerender handler reassembles stored `page_components.rendered_html` and injects `site_components.rendered_html` — it does not re-render component templates — so sites can live indefinitely on reassemblies of early renders while the library advances (idea.uk served renders of long-inactive components). Template/library changes only reach pages through a section re-render or a full page-build-handler rebuild; chrome only through render_site_components. Settled a documented 016-vs-026 doc contradiction by direct evidence.
- **sources:** RUNBOOK_scheme_to_components(18).md Check 3/4 RESULTS; running_notes_scheme_to_components(22).md Sh (migration route), So
- **relations:** assemble-only vs section re-render; chrome refresh gating; W6 rebuild sequencing.
- **verify-later:** rerender_single_page_action.go; which paths run RenderComponent vs re-ship stored HTML.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Chrome refresh gating (render_site_components, force_rerender, repoint-before-rerender)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "renderAndStoreSiteComponent joins the PINNED site_components.component_id with no is_active filter, and without force_rerender SKIPS non-empty slots" (RUNBOOK_scheme(18) W4b, code fact); repoint executed UPDATE 1 ×2.
- **what:** Site chrome (header/footer/head) is rendered and stored by render_site_components, a path separate from page builds (page-build-handler is not among its six invoking agents — a rebuild never refreshes chrome). Its join honours the pinned component_id with no is_active filter and skips non-empty slots unless `force_rerender: true` (only rerender-pages v6 passes it, gated on `spec.refresh_site_components`). Operational consequence: repoint the pinned rows to the fixed/active components BEFORE forcing the re-render, and refresh stored chrome before a rebuild so later automatic rerenders can't re-inject stale dark chrome.
- **sources:** RUNBOOK_scheme_to_components(18).md W4b + STEP-1 RESULTS; running_notes(22).md Sy–Ta; w4b_03_read_rerender_config.sql
- **relations:** chrome linkage tangle; rerender fossilisation; F3 (same rerender-pages agent).
- **verify-later:** render_site_components_action.go:345–430; rerender-pages v6 check_refresh_components gate.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### page-build-handler writes only planned sections (sections=0 → silent no-op rebuild)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "matchmatrix sections_listed=0 vs siblings 9/5 ⇒ the page-build-handler iterates the planned section list and had nothing to write — planning-data gap on a legacy page — PARKED" (NOTES §9bh).
- **what:** The writer rebuild flow iterates the page's planned section list; a legacy page with an empty `pages.sections` plan completes its needs_page item without writing anything — a silent no-op rebuild indistinguishable from success at the item level (double no-op observed before diagnosis; page_type hypothesis falsified first). Remediation options (planner reconcile/adopt, or retire the page) parked as hygiene.
- **sources:** NOTES(43).md §9bf–§9bh; RUNBOOK(49).md Step 12(a)
- **relations:** auto-escalation; loose status semantics; site-plan-and-reconciler (planner as another item producer, §9at).
- **verify-later:** page-build-handler section iteration; matchmatrix pages.sections.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F3 scoped reason-stamped rerender (dependent-page scoping + reason propagation)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "F3 PROVEN END-TO-END (scoping to exactly the five, reason propagation, gate, section-rerender execution)" (NOTES §9v); fleet on v1.0.1088 (§9r).
- **what:** Three coupled changes making a component regen's triggered re-render actually repair sections, scoped to the blast radius: F3a — create_rerender_items accepts `reason`+`component_id`; when reason ∈ {section_data_resolved, image_landed} AND component_id set, it queries the component's dependent pages (page_components JOIN pages) and creates reason-stamped `page_rerender` items only for those; no signals → unchanged assemble-only all-pages behaviour. F3b — store re-adds `reason: section_data_resolved` to its needs_rerender spec. F3c — rerender-pages step config maps `reason`/`component_id` from `input_data.spec`. Either half alone degrades safely to assemble-only. Accepted residual: rerender_page_sections re-renders ALL sections of each dependent page (documented blast radius that later stamped the gauntlet onto robot-hands).
- **sources:** NOTES(43).md §9b, §9p–§9r, §9v; RUNBOOK(49).md Part A; w4b_03_read_rerender_config.sql
- **relations:** assemble-only vs section re-render; F8 step 4 reused it; F6 (dedup/counter flaws in the same action).
- **verify-later:** create_rerender_items_action.go (InputSpec reason/component_id; dependent scoping; rows.Close before INSERT loop); rerender-pages default_config create_rerender_items step (5 keys).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Assemble-only vs section re-render distinction (the `reason` gate)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "a bare needs_rerender is assemble-only and just re-ships the empty shell" (NOTES §6); gate `reason ∈ {image_landed, section_data_resolved}` confirmed from rerender_page_sections source (§2 Correction 3).
- **what:** Two fundamentally different re-render depths: an assemble-only pass re-assembles/re-ships existing `rendered_html` (cannot repair content), while a section re-render (`rerender_page_sections`, gated on reason image_landed/section_data_resolved) regenerates each section's HTML from stored `content_data` against the current template, no LLM. A reason dropped anywhere along the chain (as rerender-pages originally did) silently downgrades to assemble-only — an inert "fix" that was caught only by checking the consuming step's config. Central operational lesson of the thread.
- **sources:** NOTES(43).md §2 Correction 3, §6, §9a; PLAN(1).md Phase 4; RUNBOOK_pre_cleanup_backup.md §The problem
- **relations:** F3 (makes the reason survive); rerender fossilisation (site-level analogue); carry-forward path.
- **verify-later:** rerender_page_sections_action.go reason gate; page-rerender workflow's routing on spec.reason.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Carry-forward path and the carry fingerprint
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "per-page save-style row updates + identical bytes = the rerender_page_sections CARRY path fingerprint" (NOTES §9u), confirmed by data §9v.
- **what:** In rerender_page_sections, a section that fails `planSection` readiness (unresolvable required field, missing component, empty template) is carried: `carryStoredSection` re-emits the stored HTML, save_sections writes it back with a fresh per-page `updated_at` but identical bytes. Protective against shipping worse output, but it re-fossilises stale/empty renders forever when readiness is permanently blocked — and its diagnostic signature (fresh distinct timestamps + one shared md5) is how the recovery stall was pinned to cta_url readiness rather than the rerender chain.
- **sources:** NOTES(43).md §9u, §9v; BUNDLE(3).md §3 (rerender partly protective)
- **relations:** section readiness model (the gate it consults); F5 (the added-required-field cause); recovery playbook.
- **verify-later:** rerender_page_sections_action.go: planSection, carryStoredSection, NULL pre-check.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Auto-escalation: empty content_data → needs_page writer rebuild
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "Auto-escalations fired as designed: page-rerender created needs_page for matchmatrix (17:31) and gripper-selection-guide (17:37); writer rebuilt gripper-selection-guide 18:04" (NOTES §9aw).
- **what:** rerender_page_sections' NULL pre-check escalates a whole page to a `needs_page` item (handler page-build-handler, spec `{reason: content_data_backfill, page_name}`) when any section lacks `content_data` — routing un-re-renderable pages to a full writer rebuild instead of carrying garbage. Deliberately exploited during F8 recovery both as the rebuild mechanism and as the source of a correctly-shaped needs_page item to clone (never guess a spec).
- **sources:** NOTES(43).md §9ap, §9aw, §9bd; RUNBOOK(49).md Part C Steps 8–9
- **relations:** carry-forward path (its sibling branch); work-item spec-cloning discipline; matchmatrix planning-data gap (an escalation that no-ops).
- **verify-later:** rerender_page_sections NULL pre-check/escalate branch; page-build-handler content_data_backfill flow.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Rerender fossilisation (reassembly re-ships stale renders; template changes need full rebuild)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "deployed hero consumes legacy var(--accent-color…) — sections are stale old-template renders… A full page-build-handler rebuild is required; needs_rerender would re-fossilise" (RUNBOOK_scheme_to_components(18) Check 4a, evidenced from deployed HTML).
- **what:** The rerender handler reassembles stored `page_components.rendered_html` and injects `site_components.rendered_html` — it does not re-render component templates — so sites can live indefinitely on reassemblies of early renders while the library advances (idea.uk served renders of long-inactive components). Template/library changes only reach pages through a section re-render or a full page-build-handler rebuild; chrome only through render_site_components. Settled a documented 016-vs-026 doc contradiction by direct evidence.
- **sources:** RUNBOOK_scheme_to_components(18).md Check 3/4 RESULTS; running_notes_scheme_to_components(22).md Sh (migration route), So
- **relations:** assemble-only vs section re-render; chrome refresh gating; W6 rebuild sequencing.
- **verify-later:** rerender_single_page_action.go; which paths run RenderComponent vs re-ship stored HTML.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Chrome refresh gating (render_site_components, force_rerender, repoint-before-rerender)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "renderAndStoreSiteComponent joins the PINNED site_components.component_id with no is_active filter, and without force_rerender SKIPS non-empty slots" (RUNBOOK_scheme(18) W4b, code fact); repoint executed UPDATE 1 ×2.
- **what:** Site chrome (header/footer/head) is rendered and stored by render_site_components, a path separate from page builds (page-build-handler is not among its six invoking agents — a rebuild never refreshes chrome). Its join honours the pinned component_id with no is_active filter and skips non-empty slots unless `force_rerender: true` (only rerender-pages v6 passes it, gated on `spec.refresh_site_components`). Operational consequence: repoint the pinned rows to the fixed/active components BEFORE forcing the re-render, and refresh stored chrome before a rebuild so later automatic rerenders can't re-inject stale dark chrome.
- **sources:** RUNBOOK_scheme_to_components(18).md W4b + STEP-1 RESULTS; running_notes(22).md Sy–Ta; w4b_03_read_rerender_config.sql
- **relations:** chrome linkage tangle; rerender fossilisation; F3 (same rerender-pages agent).
- **verify-later:** render_site_components_action.go:345–430; rerender-pages v6 check_refresh_components gate.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### page-build-handler writes only planned sections (sections=0 → silent no-op rebuild)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "matchmatrix sections_listed=0 vs siblings 9/5 ⇒ the page-build-handler iterates the planned section list and had nothing to write — planning-data gap on a legacy page — PARKED" (NOTES §9bh).
- **what:** The writer rebuild flow iterates the page's planned section list; a legacy page with an empty `pages.sections` plan completes its needs_page item without writing anything — a silent no-op rebuild indistinguishable from success at the item level (double no-op observed before diagnosis; page_type hypothesis falsified first). Remediation options (planner reconcile/adopt, or retire the page) parked as hygiene.
- **sources:** NOTES(43).md §9bf–§9bh; RUNBOOK(49).md Step 12(a)
- **relations:** auto-escalation; loose status semantics; site-plan-and-reconciler (planner as another item producer, §9at).
- **verify-later:** page-build-handler section iteration; matchmatrix pages.sections.

# EXTRACTION U07 — docs024_key_docs_latest/content_quality2_silent_overwrite_of_dependent_pages
Extracted 2026-07-13. Files in scope: 115. Concepts found: 54.

Unit character: a single deep debugging/design thread (2026-06-24 → 2026-07-06) on the "component-regen
clobber" — a shared component regeneration silently emptying every dependent page — which grew into the F8
shared-component contamination incident and the R6f theming-vocabulary-drift fix, plus two imported files from
the sibling "scheme reaches components" thread. Dominated by four large version families; NOTES is strictly
append-only (base is a byte-prefix of the latest — verified by diff), so family-deltas contributed only the
superseded early hypothesis captured from BUNDLE/PLAN/HANDOFF bases.

## Coverage
| file | treatment |
|---|---|
| BUNDLE_component_regen_clobber.md | family-delta |
| BUNDLE_component_regen_clobber(1).md | family-delta |
| BUNDLE_component_regen_clobber(2).md | family-delta |
| BUNDLE_component_regen_clobber(3).md | family-latest |
| F1_store_generated_component_action.patch | full |
| F1prompt_component_creator_preserve_field_names.sql | family-delta |
| F1prompt_component_creator_preserve_field_names(1).sql | family-latest |
| HANDOFF_component_regen_clobber.md | family-delta |
| HANDOFF_component_regen_clobber(1).md | family-delta |
| HANDOFF_component_regen_clobber(2).md | family-delta |
| HANDOFF_component_regen_clobber(3).md | family-delta |
| HANDOFF_component_regen_clobber(4).md | family-delta |
| HANDOFF_component_regen_clobber(5).md | family-delta |
| HANDOFF_component_regen_clobber(6).md | family-delta |
| HANDOFF_component_regen_clobber(7).md | family-latest |
| NOTES_component_regen_clobber.md | family-delta |
| NOTES_component_regen_clobber(1).md | family-delta |
| NOTES_component_regen_clobber(2).md | family-delta |
| NOTES_component_regen_clobber(3).md | family-delta |
| NOTES_component_regen_clobber(4).md | family-delta |
| NOTES_component_regen_clobber(5).md | family-delta |
| NOTES_component_regen_clobber(6).md | family-delta |
| NOTES_component_regen_clobber(7).md | family-delta |
| NOTES_component_regen_clobber(8).md | family-delta |
| NOTES_component_regen_clobber(9).md | family-delta |
| NOTES_component_regen_clobber(10).md | family-delta |
| NOTES_component_regen_clobber(11).md | family-delta |
| NOTES_component_regen_clobber(12).md | family-delta |
| NOTES_component_regen_clobber(13).md | family-delta |
| NOTES_component_regen_clobber(14).md | family-delta |
| NOTES_component_regen_clobber(15).md | family-delta |
| NOTES_component_regen_clobber(16).md | family-delta |
| NOTES_component_regen_clobber(17).md | family-delta |
| NOTES_component_regen_clobber(18).md | family-delta |
| NOTES_component_regen_clobber(19).md | family-delta |
| NOTES_component_regen_clobber(20).md | family-delta |
| NOTES_component_regen_clobber(21).md | family-delta |
| NOTES_component_regen_clobber(22).md | family-delta |
| NOTES_component_regen_clobber(23).md | family-delta |
| NOTES_component_regen_clobber(24).md | family-delta |
| NOTES_component_regen_clobber(25).md | family-delta |
| NOTES_component_regen_clobber(26).md | family-delta |
| NOTES_component_regen_clobber(27).md | family-delta |
| NOTES_component_regen_clobber(28).md | family-delta |
| NOTES_component_regen_clobber(29).md | family-delta |
| NOTES_component_regen_clobber(30).md | family-delta |
| NOTES_component_regen_clobber(31).md | family-delta |
| NOTES_component_regen_clobber(32).md | family-delta |
| NOTES_component_regen_clobber(33).md | family-delta |
| NOTES_component_regen_clobber(34).md | family-delta |
| NOTES_component_regen_clobber(35).md | family-delta |
| NOTES_component_regen_clobber(36).md | family-delta |
| NOTES_component_regen_clobber(37).md | family-delta |
| NOTES_component_regen_clobber(38).md | family-delta |
| NOTES_component_regen_clobber(39).md | family-delta |
| NOTES_component_regen_clobber(40).md | family-delta |
| NOTES_component_regen_clobber(41).md | family-delta |
| NOTES_component_regen_clobber(42).md | family-delta |
| NOTES_component_regen_clobber(43).md | family-latest |
| PLAN_component_regen_clobber.md | family-delta |
| PLAN_component_regen_clobber(1).md | family-latest |
| RUNBOOK_component_regen_clobber.md | family-delta |
| RUNBOOK_component_regen_clobber(1).md | family-delta |
| RUNBOOK_component_regen_clobber(2).md | family-delta |
| RUNBOOK_component_regen_clobber(3).md | family-delta |
| RUNBOOK_component_regen_clobber(4).md | family-delta |
| RUNBOOK_component_regen_clobber(5).md | family-delta |
| RUNBOOK_component_regen_clobber(6).md | family-delta |
| RUNBOOK_component_regen_clobber(7).md | family-delta |
| RUNBOOK_component_regen_clobber(8).md | family-delta |
| RUNBOOK_component_regen_clobber(9).md | family-delta |
| RUNBOOK_component_regen_clobber(10).md | family-delta |
| RUNBOOK_component_regen_clobber(11).md | family-delta |
| RUNBOOK_component_regen_clobber(12).md | family-delta |
| RUNBOOK_component_regen_clobber(13).md | family-delta |
| RUNBOOK_component_regen_clobber(14).md | family-delta |
| RUNBOOK_component_regen_clobber(15).md | family-delta |
| RUNBOOK_component_regen_clobber(16).md | family-delta |
| RUNBOOK_component_regen_clobber(17).md | family-delta |
| RUNBOOK_component_regen_clobber(18).md | family-delta |
| RUNBOOK_component_regen_clobber(19).md | family-delta |
| RUNBOOK_component_regen_clobber(20).md | family-delta |
| RUNBOOK_component_regen_clobber(21).md | family-delta |
| RUNBOOK_component_regen_clobber(22).md | family-delta |
| RUNBOOK_component_regen_clobber(23).md | family-delta |
| RUNBOOK_component_regen_clobber(24).md | family-delta |
| RUNBOOK_component_regen_clobber(25).md | family-delta |
| RUNBOOK_component_regen_clobber(26).md | family-delta |
| RUNBOOK_component_regen_clobber(27).md | family-delta |
| RUNBOOK_component_regen_clobber(28).md | family-delta |
| RUNBOOK_component_regen_clobber(29).md | family-delta |
| RUNBOOK_component_regen_clobber(30).md | family-delta (pre-rewrite peak: R0–R7/F1–F3/R5b step detail, all mirrored in NOTES §9) |
| RUNBOOK_component_regen_clobber(31).md | family-delta (the 2026-07-06 wholesale rewrite point) |
| RUNBOOK_component_regen_clobber(32).md | family-delta |
| RUNBOOK_component_regen_clobber(33).md | family-delta |
| RUNBOOK_component_regen_clobber(34).md | family-delta |
| RUNBOOK_component_regen_clobber(35).md | family-delta |
| RUNBOOK_component_regen_clobber(36).md | family-delta |
| RUNBOOK_component_regen_clobber(37).md | family-delta |
| RUNBOOK_component_regen_clobber(38).md | family-delta |
| RUNBOOK_component_regen_clobber(39).md | family-delta |
| RUNBOOK_component_regen_clobber(40).md | family-delta |
| RUNBOOK_component_regen_clobber(41).md | family-delta |
| RUNBOOK_component_regen_clobber(42).md | family-delta |
| RUNBOOK_component_regen_clobber(43).md | family-delta |
| RUNBOOK_component_regen_clobber(44).md | family-delta |
| RUNBOOK_component_regen_clobber(45).md | family-delta |
| RUNBOOK_component_regen_clobber(46).md | family-delta |
| RUNBOOK_component_regen_clobber(47).md | family-delta |
| RUNBOOK_component_regen_clobber(48).md | family-delta |
| RUNBOOK_component_regen_clobber(49).md | family-latest |
| RUNBOOK_pre_cleanup_backup.md | full |
| RUNBOOK_scheme_to_components(18).md | full (family member; other versions live outside this unit) |
| running_notes_scheme_to_components(22).md | full (family member; other versions live outside this unit) |
| w4b_03_read_rerender_config.sql | full |

Proposed NEW categories: `NEW:component-lifecycle` (shared content-component creation/regeneration/versioning/
guarding — content_components is platform machinery distinct from the tool library), `NEW:rebuild-cascade`
(how a change propagates to dependent pages: needs_rerender/page_rerender/needs_page items, reasons, scoping,
carry-forward, escalation, chrome refresh — the heart of this thread and a coherent expert domain).

## Concepts

### Shared-component regen clobber failure mode (silent overwrite of dependent pages)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "RESOLVED + RECOVERED (verified)" (HANDOFF(7), 2026-07-04); "R6b pass 2026-07-03: distinct md5s, needles true"; root cause section dated + confirmed in NOTES §4.
- **what:** Regenerating a shared component overwrites its `input_schema`/`html_template` field contract in place without migrating dependent pages' `content_data` to the new field names; rendering binds by exact field name and silently empties misses, so every dependent renders a content-free shell that the assembler silently drops — fanning out across every page/site sharing the component. Confirmed on `system-stats` (`fdd92ad4`, regen 2026-06-24 15:06): 24 old keys vs 22 new, five live pages on three sites byte-identical empty. `content_data` stayed intact and per-page, so the breakage was recoverable without an LLM.
- **sources:** NOTES_component_regen_clobber(43).md §1, §4, §8; HANDOFF_component_regen_clobber(7).md §Incident 1; RUNBOOK_pre_cleanup_backup.md §The problem
- **relations:** F1 field-contract guard (fix); F3 scoped rerender (repair path); F5/F8 (sibling facets); RenderTemplate silent-empty mechanism; visible-content filter.
- **verify-later:** platform/orchestration/actions/store_generated_component_action.go (regen branch ~L354–432); content_components/component_versions/page_components tables; component `fdd92ad4-521a-4602-89cf-7ee1a66c10f1`.

### Exact-field-name template binding with silent empty on miss (RenderTemplate `<no value>` strip)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "Correction 2 — the silent-empty mechanism is in RenderTemplate" (NOTES §2, confirmed from uploaded component_library.go).
- **what:** `RenderTemplate` (component_library.go) binds a page's `content_data` into a component's `html_template` by exact field name via Go `text/template`, then strips the `<no value>` tokens of unmatched placeholders to empty string, logging only a warning — no error. This is *why* a renamed or missing field fails silently rather than loudly; the entire clobber class rests on it.
- **sources:** NOTES(43).md §2 Correction 2, §8; BUNDLE(3).md §1
- **relations:** clobber failure mode; F1 guard (compensating control); fail-loud guard route (never built as such).
- **verify-later:** platform/orchestration/actions/component_library.go:RenderTemplate; the `<no value>` cleanup.

### Shared content-component reuse model (one content_components row, N page_components instances)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** Stated as settled platform fact throughout ("one `content_components` row, N `page_components` instances", BUNDLE(3) §1); vonc.com/index arrived mid-recovery as "healthy sixth dependent" proving live reuse.
- **what:** A section component is a single shared `content_components` row (keyed by `function`, with `section_type`, `input_schema.fields`, `html_template`, `is_active`, `forked_from`) reused across pages and sites; each page stores its own `content_data` in `page_components`. Any change to the shared row therefore has a cross-site blast radius — the structural precondition for both incidents in this thread.
- **sources:** BUNDLE(3).md §1; NOTES(43).md §1, §9z (vonc sixth dependent); HANDOFF(7).md §Platform operating model
- **relations:** clobber failure mode; F4 fork-vs-match; F8 contamination; optimistic-lock co-management.
- **verify-later:** content_components + page_components schemas; idx_cc_selector (section_type, component_level) partial index.

### StoreGeneratedComponentAction regeneration branch
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Correction 6 — StoreGeneratedComponentAction is the rename writer" (NOTES §2); insert/dedup shape read from the live file (NOTES §9q).
- **what:** On storing a generated component whose `function` matches an existing row (`WHERE function=$1 AND forked_from IS NULL`, deliberately is_active-agnostic since 2026-05-06), the action snapshots the old schema/template to `component_versions`, UPDATEs the shared row in place (same component_id, so dependents follow), marks dependents pending (`markPagesPendingRebuild` — build_status only, no rendered_html), and raises one deduped `needs_rerender` work item per affected site via `createRerenderWorkItem`. Pre-fix, that item carried no `reason`, making the triggered re-render assemble-only and unable to repair anything.
- **sources:** NOTES(43).md §2 Correction 6, §9h, §9q; BUNDLE(3).md §3
- **relations:** F1 guard lives in its validation block; F3b re-added the reason to its spec; F4 (its function-keyed lookup is the fork vector).
- **verify-later:** store_generated_component_action.go (existence check L198–207; regen branch; markPagesPendingRebuild; createRerenderWorkItem NOT EXISTS dedup).

### F1 field-contract guard (reject regens that rename/drop retained fields)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "F2 COMPLETE… 3b reject ✔ (live firing, three-level visibility, zero mutation)" (NOTES §9o, 2026-07-02 16:39); RUNBOOK(49) Part A "Fixes deployed + proven".
- **what:** In StoreGeneratedComponentAction's Layer-1 validation, on `isRegeneration` the guard diffs old vs new `input_schema.fields` names (`schemaFieldSet` helper); any retained field that disappears becomes a blockingIssue routed through `recordValidationRejection` into `agent_error_log` — additions allowed, renames/drops rejected before any snapshot/UPDATE. Converts silent stranding into a loud, queryable rejection naming the stranded fields. Design choice: preserve-the-contract strict-reject backstop, not a per-dependent migration.
- **sources:** F1_store_generated_component_action.patch; NOTES(43).md §9, §9a, §9o; RUNBOOK(49).md Part A
- **relations:** F1-prompt (generation-time complement so name-preserving regens succeed); F5 (proposed extension); F8 (guard checks names only — its blind spot).
- **verify-later:** store_generated_component_action.go guard block + schemaFieldSet; agent_error_log rows error_code component_validation_rejected; store_generated_component_guard_test.go.

### F1-prompt generation-time field-name preservation (loader + dormant rule + function pin)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "F1-prompt 1a/1b+2 done (validator passes; loader executes; verify row good)" (NOTES §9k); Tier 3a two regens preserved names with md5-verified template change (§9l).
- **what:** Three coupled pieces so regens preserve names instead of being rejected: (1) `load_existing_component` Go action — looks up the canonical component by `section_type` (is_active, forked_from IS NULL, component_level='section'), outputs `existing_component.field_names` + `function`; advisory, never errors (no match → empty map → rule dormant). (2) A snapshot-first, anchored, idempotent, drift-checked SQL migration wiring the step before generate_template and inserting a dormant `{{if .existing_component.field_names}}` prompt rule: reuse these exact names, MAY add, MUST NOT rename/drop. (3) `F1prompt2_pin_function.sql` pins `{{.existing_component.function}}` so the store matches the same row (the F4 mitigation). Option A (pre-generation lookup by section_type) chosen after `\d content_components` confirmed section_type is queryable.
- **sources:** F1prompt_component_creator_preserve_field_names(1).sql; NOTES(43).md §9c–§9f, §9m; RUNBOOK(49).md Part A
- **relations:** F1 guard (backstop); F4 (pin is its mitigation); prompt-migration convention; deploy-ordering gate (its 9i failure).
- **verify-later:** load_existing_component_action.go + registry.go (IsLocal:true); component-creator default_config (top-level prompt_template; load_existing_component step; input_fields incl. existing_component).

### Store-driven retry on field-drift rejection (Option B)
- **category:** NEW:component-lifecycle
- **status-signal:** abandoned
- **status-evidence:** "Two options… user chose A" (NOTES §9d–§9e); Option B never mentioned again after 9e.
- **what:** The rejected alternative to Option A: on a field-drift rejection the guard would return the existing field names, and a store_component error edge would loop back to generate_template with the names injected, retrying once. Judged heavier (reject-with-retry-data + loop-guarded workflow edge) but authoritative (no key guessing); dropped when section_type proved a stable pre-generation lookup key.
- **sources:** NOTES(43).md §9d, §9e
- **relations:** F1-prompt Option A (superseding choice).
- **verify-later:** absence: component-creator workflow should have no store_component→generate_template error edge.

### F3 scoped reason-stamped rerender (dependent-page scoping + reason propagation)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "F3 PROVEN END-TO-END (scoping to exactly the five, reason propagation, gate, section-rerender execution)" (NOTES §9v); fleet on v1.0.1088 (§9r).
- **what:** Three coupled changes making a component regen's triggered re-render actually repair sections, scoped to the blast radius: F3a — create_rerender_items accepts `reason`+`component_id`; when reason ∈ {section_data_resolved, image_landed} AND component_id set, it queries the component's dependent pages (page_components JOIN pages) and creates reason-stamped `page_rerender` items only for those; no signals → unchanged assemble-only all-pages behaviour. F3b — store re-adds `reason: section_data_resolved` to its needs_rerender spec. F3c — rerender-pages step config maps `reason`/`component_id` from `input_data.spec`. Either half alone degrades safely to assemble-only. Accepted residual: rerender_page_sections re-renders ALL sections of each dependent page (documented blast radius that later stamped the gauntlet onto robot-hands).
- **sources:** NOTES(43).md §9b, §9p–§9r, §9v; RUNBOOK(49).md Part A; w4b_03_read_rerender_config.sql
- **relations:** assemble-only vs section re-render; F8 step 4 reused it; F6 (dedup/counter flaws in the same action).
- **verify-later:** create_rerender_items_action.go (InputSpec reason/component_id; dependent scoping; rows.Close before INSERT loop); rerender-pages default_config create_rerender_items step (5 keys).

### Assemble-only vs section re-render distinction (the `reason` gate)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "a bare needs_rerender is assemble-only and just re-ships the empty shell" (NOTES §6); gate `reason ∈ {image_landed, section_data_resolved}` confirmed from rerender_page_sections source (§2 Correction 3).
- **what:** Two fundamentally different re-render depths: an assemble-only pass re-assembles/re-ships existing `rendered_html` (cannot repair content), while a section re-render (`rerender_page_sections`, gated on reason image_landed/section_data_resolved) regenerates each section's HTML from stored `content_data` against the current template, no LLM. A reason dropped anywhere along the chain (as rerender-pages originally did) silently downgrades to assemble-only — an inert "fix" that was caught only by checking the consuming step's config. Central operational lesson of the thread.
- **sources:** NOTES(43).md §2 Correction 3, §6, §9a; PLAN(1).md Phase 4; RUNBOOK_pre_cleanup_backup.md §The problem
- **relations:** F3 (makes the reason survive); rerender fossilisation (site-level analogue); carry-forward path.
- **verify-later:** rerender_page_sections_action.go reason gate; page-rerender workflow's routing on spec.reason.

### Carry-forward path and the carry fingerprint
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "per-page save-style row updates + identical bytes = the rerender_page_sections CARRY path fingerprint" (NOTES §9u), confirmed by data §9v.
- **what:** In rerender_page_sections, a section that fails `planSection` readiness (unresolvable required field, missing component, empty template) is carried: `carryStoredSection` re-emits the stored HTML, save_sections writes it back with a fresh per-page `updated_at` but identical bytes. Protective against shipping worse output, but it re-fossilises stale/empty renders forever when readiness is permanently blocked — and its diagnostic signature (fresh distinct timestamps + one shared md5) is how the recovery stall was pinned to cta_url readiness rather than the rerender chain.
- **sources:** NOTES(43).md §9u, §9v; BUNDLE(3).md §3 (rerender partly protective)
- **relations:** section readiness model (the gate it consults); F5 (the added-required-field cause); recovery playbook.
- **verify-later:** rerender_page_sections_action.go: planSection, carryStoredSection, NULL pre-check.

### Auto-escalation: empty content_data → needs_page writer rebuild
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "Auto-escalations fired as designed: page-rerender created needs_page for matchmatrix (17:31) and gripper-selection-guide (17:37); writer rebuilt gripper-selection-guide 18:04" (NOTES §9aw).
- **what:** rerender_page_sections' NULL pre-check escalates a whole page to a `needs_page` item (handler page-build-handler, spec `{reason: content_data_backfill, page_name}`) when any section lacks `content_data` — routing un-re-renderable pages to a full writer rebuild instead of carrying garbage. Deliberately exploited during F8 recovery both as the rebuild mechanism and as the source of a correctly-shaped needs_page item to clone (never guess a spec).
- **sources:** NOTES(43).md §9ap, §9aw, §9bd; RUNBOOK(49).md Part C Steps 8–9
- **relations:** carry-forward path (its sibling branch); work-item spec-cloning discipline; matchmatrix planning-data gap (an escalation that no-ops).
- **verify-later:** rerender_page_sections NULL pre-check/escalate branch; page-build-handler content_data_backfill flow.

### Rerender fossilisation (reassembly re-ships stale renders; template changes need full rebuild)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "deployed hero consumes legacy var(--accent-color…) — sections are stale old-template renders… A full page-build-handler rebuild is required; needs_rerender would re-fossilise" (RUNBOOK_scheme_to_components(18) Check 4a, evidenced from deployed HTML).
- **what:** The rerender handler reassembles stored `page_components.rendered_html` and injects `site_components.rendered_html` — it does not re-render component templates — so sites can live indefinitely on reassemblies of early renders while the library advances (idea.uk served renders of long-inactive components). Template/library changes only reach pages through a section re-render or a full page-build-handler rebuild; chrome only through render_site_components. Settled a documented 016-vs-026 doc contradiction by direct evidence.
- **sources:** RUNBOOK_scheme_to_components(18).md Check 3/4 RESULTS; running_notes_scheme_to_components(22).md Sh (migration route), So
- **relations:** assemble-only vs section re-render; chrome refresh gating; W6 rebuild sequencing.
- **verify-later:** rerender_single_page_action.go; which paths run RenderComponent vs re-ship stored HTML.

### Chrome refresh gating (render_site_components, force_rerender, repoint-before-rerender)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "renderAndStoreSiteComponent joins the PINNED site_components.component_id with no is_active filter, and without force_rerender SKIPS non-empty slots" (RUNBOOK_scheme(18) W4b, code fact); repoint executed UPDATE 1 ×2.
- **what:** Site chrome (header/footer/head) is rendered and stored by render_site_components, a path separate from page builds (page-build-handler is not among its six invoking agents — a rebuild never refreshes chrome). Its join honours the pinned component_id with no is_active filter and skips non-empty slots unless `force_rerender: true` (only rerender-pages v6 passes it, gated on `spec.refresh_site_components`). Operational consequence: repoint the pinned rows to the fixed/active components BEFORE forcing the re-render, and refresh stored chrome before a rebuild so later automatic rerenders can't re-inject stale dark chrome.
- **sources:** RUNBOOK_scheme_to_components(18).md W4b + STEP-1 RESULTS; running_notes(22).md Sy–Ta; w4b_03_read_rerender_config.sql
- **relations:** chrome linkage tangle; rerender fossilisation; F3 (same rerender-pages agent).
- **verify-later:** render_site_components_action.go:345–430; rerender-pages v6 check_refresh_components gate.

### Recovery playbook for stranded dependents (Route A rebuild vs Route B re-key + scoped re-render)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "R5b step 8 PASS — all five RECOVERED and verified" (NOTES §9ad, 2026-07-03); leopardess confirmed live by screenshot (§9ae).
- **what:** Two recovery routes for pages stranded by a shared-contract change: Route A — full `needs_page` writer rebuild (regenerates content_data under the new schema; simplest, costs LLM); Route B — re-key each page's `content_data` old→new (explicit reviewable jsonb_build_object mapping, dry-run first, CTAS backup, non-1:1 fields handled explicitly) then trigger the F3-scoped section re-render (no LLM, preserves per-page values). Route B executed for the five, doubling as F3's end-to-end proof; gated on fleet image, freshness check, and a cta-schema decision.
- **sources:** NOTES(43).md §6, §9q, §9s–§9t, §9ad; RUNBOOK(49).md Part A; PLAN(1).md Phase 4
- **relations:** F3; section readiness model (the cta_url blocker it hit); optimistic-lock co-management; snapshot-before-change.
- **verify-later:** page_components content_data keys for the five; backup tables page_components_bak_sysstats_20260702 / _briefexp_20260703 (may be dropped).

### Section readiness model (planSection source tiers, required/fallback semantics, spec resolver)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "planSection required semantics CONFIRMED from plan_sections_action.go… 'Required with no fallback — defer' → deterministic defer → carry" (NOTES §9w).
- **what:** Section fields declare a source (static with fallback; llm; Tier-C spec paths like `site_specs.cta.primary_url`) plus required/fallback attributes. planSection resolves each non-LLM field; a required field with no resolvable source and no fallback defers the section (→ carry). The resolver reads site_specs per-aspect rows (`aspect`,`data` jsonb, is_current; `resolveSpecPath("cta.primary_url") = specs["cta"]["primary_url"]`), checks presence not validity, and the stored⊕resolved merge persists resolved values into content_data at render time. Tier-C fields are by design never content_data keys.
- **sources:** NOTES(43).md §9s, §9u–§9x; RUNBOOK(49).md Part A
- **relations:** carry-forward path; F5; stored⊕resolved merge; phantom-CTA lesson (spec presence ≠ URL validity).
- **verify-later:** plan_sections_action.go (ensureSpecs, resolveSpecPath, on_missing switch); site_specs schema.

### Stored⊕resolved merge writes resolved values back into content_data
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "resolver persists cta_url into content_data (stored⊕resolved merge — expected)" (NOTES §9w); robot-hands cd_keys gained merged fallback keys (§9an).
- **what:** When a section re-render resolves fields (spec values, static fallbacks), the merged result is persisted back into the page's `content_data` as well as baked into `rendered_html`. Double-edged: it makes recoveries durable (cta_url landed in content_data), but it is also F8 carrier 2 — contaminated fallback values were merged into dependents' content_data, surviving the later schema fix and needing an explicit key-strip.
- **sources:** NOTES(43).md §9w, §9an, §9ao; HANDOFF(7).md §Incident 2
- **relations:** F8 contamination; section readiness model; recovery playbook.
- **verify-later:** rerender_page_sections persist-merged-content step.

### Visible-content filter drops near-empty bands at assembly
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "sections = page_components by page_id ORDER BY position; visible-text filter drops bands with ≤10 chars of stripped text (logger.Warn per skip)" (NOTES §9ai, from rerender_single_page_action.go).
- **what:** Page assembly filters out any band whose rendered HTML strips to ≤10 visible characters, logging only a warning. It is the final silencer in the clobber chain (an emptied section vanishes rather than erroring) and produces counter-intuitive interims — the F8 "neutral shell" bands survived the filter because two neutral CTA labels exceeded 10 chars.
- **sources:** NOTES(43).md §1, §9ai, §9ar; BUNDLE(3).md §1
- **relations:** clobber failure mode; assembly membership model; fail-loud guard route (unbuilt alternative).
- **verify-later:** rerender_single_page_action.go visible-text filter.

### Assembly membership and chrome model (page_components by position; pages.sections is metadata)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "assembly membership = page_components by position… pages.sections jsonb is NOT assembly membership" (NOTES §9ai — corrected a wrong sections_listed=0 inference); three head shapes identified.
- **what:** A page artifact is assembled from `page_components` rows ordered by position (not from `pages.sections`, which is planning metadata); head/header/footer come from site-scoped `site_components` rows; with no stored head, `buildDefaultHead` emits a 5-line head linking /assets/css/styles.css. A third, legacy builder (`assemble_from_library`, theme CSS from css_themes) produced older artifacts with big inline heads — three coexisting head shapes that repeatedly confused artifact forensics. Also: `data-component` attributes exist on only some templates, so artifact greps on them undercount bands (owned metric artifact).
- **sources:** NOTES(43).md §9ah–§9al; RUNBOOK(49).md Part B
- **relations:** visible-content filter; R6f vocabulary drift (missing :root head); chrome refresh gating.
- **verify-later:** rerender_single_page_action.go; assemble_from_library (registry L493); site_components schema.

### F4 — regen-vs-create keyed on the LLM-chosen function (silent fork)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "F4 (structural finding): regen-vs-create is keyed on the LLM-chosen function → nondeterministic. Miss case = silent FORK" (NOTES §9m); pin migration applied §9n; "store-side advisory FLAGGED as follow-up, not built".
- **what:** Whether a store is a regeneration or a creation depends on whether the LLM happened to choose the existing `function` name — a miss silently creates a parallel active duplicate for a section_type (library fragmentation; guard bypassed by fork; selector nondeterminism). Observed live in F2 testing (fork 80222fc1). Mitigation shipped: prompt pin of `{{.existing_component.function}}`; store-side advisory (warn when function misses but an active same-section_type row exists) deliberately advisory-only and unbuilt, since multiple components per section_type can be legitimate. A suspected live-fork case (duplicate hero rows) was later softened to old manual seeding.
- **sources:** NOTES(43).md §9m, §9n, §9al, §9am; RUNBOOK(49).md Part E
- **relations:** F1-prompt function pin; StoreGeneratedComponentAction lookup; F2 methodology (exposed it).
- **verify-later:** duplicate non-forked function rows in content_components; whether any store-side advisory exists.

### F5 — regen-added required fallback-less fields strand renderability
- **category:** NEW:component-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Flagged F5: extend the F1 guard… Not built now" (NOTES §9v); still an open Part E flag in RUNBOOK(49).
- **what:** The incident's second facet: the 15:06 regen also ADDED `cta_url` (required, Tier-C source, no fallback) that no affected site's specs could satisfy — renames strand stored content, required additions strand renderability (sections permanently not-ready → carried forever). Proposed guard extension: reject, or force optional/fallbacked, any added required field on a regeneration.
- **sources:** NOTES(43).md §9v; RUNBOOK(49).md Part E F5; HANDOFF(7).md §Flags
- **relations:** F1 guard; section readiness model; carry-forward path.
- **verify-later:** store_generated_component_action.go — absence of an added-required-field check.

### F6 — dedup status-list mismatch and itemsCreated overcount
- **category:** development-guide
- **status-signal:** aspirational
- **status-evidence:** "F6 flagged: the store's NOT EXISTS guard… omit 'unresolved' → Go guard STRICTER than index" (NOTES §9aa); unfixed Part E flag.
- **what:** Two small aligned defects: (1) the store's NOT EXISTS dedup status list and the `idx_swi_dedup` partial-unique predicate disagree on `unresolved` (index-terminal but guard-blocking — an unresolved squatter blocks createRerenderWorkItem where the index would not); (2) create_rerender_items increments `itemsCreated` without gating on RowsAffected, so ON CONFLICT DO NOTHING conflicts overcount the log. One-line alignments, parked.
- **sources:** NOTES(43).md §9t, §9aa; RUNBOOK(49).md Part E F6; HANDOFF(7).md §Flags
- **relations:** work-item dedup semantics; F3 (same action); hygiene: 40 stale unresolved items.
- **verify-later:** idx_swi_dedup definition vs createRerenderWorkItem NOT EXISTS list; create_rerender_items counter.

### F7 — unguarded template swap in update_component_html (residual)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "the snapshot INSERT is ALREADY FIXED in current code… Residual: no placeholder⇄schema sync validation on template swaps" (NOTES §9ak); Part E flag open.
- **what:** `update_component_html` swaps a shared component's template (snapshotting versions — its old silent snapshot failure on the removed `version_note` column is fixed) but performs no placeholder⇄schema agreement validation and no field-contract guard, leaving a second, unguarded write path to shared components. The original F7 framing (an unversioned live swap on hero) was investigated and softened; the residual is the missing validation.
- **sources:** NOTES(43).md §9aj, §9ak, §9am; RUNBOOK(49).md Part E F7; HANDOFF(7).md §Flags
- **relations:** F1 guard (candidate extension target); component versioning.
- **verify-later:** update_component_html_action.go — snapshot INSERT columns; absence of schema-sync validation.

### Component versioning via component_versions (and unversioned-write provenance)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "Component updated AGAIN 2026-07-03 13:22:44 with NO v2 row… unversioned write path provenance OPEN" (NOTES §9an); manual snapshots v2/v3 taken to backfill.
- **what:** `component_versions` snapshots (component_id, version_number, schema, template, change_description, changed_by, change_source) are the change history for shared components; `change_source` records the triggering work item's source (useful provenance). Coverage is incomplete: some write paths historically failed silently or bypass versioning entirely, so manual mirror-the-working-insert snapshots are the established compensation before risky writes, and zero-version updates are treated as an investigation smell.
- **sources:** NOTES(43).md §9k, §9an, §9ao, §9bd; RUNBOOK(49).md Part C Step 6
- **relations:** F7; snapshot-before-change conventions; F8 remediation (v2/v3 snapshots).
- **verify-later:** component_versions rows for fdd92ad4 and brief-explanation; snapshotComponentVersion call sites.

### F8 — shared-component contamination: site-specific copy baked into shared machinery (three carriers)
- **category:** content-quality
- **status-signal:** partial
- **status-evidence:** "STEP-12 SWEEP PASSES — contamination cleared board-wide" (NOTES §9bg, 2026-07-06) — incident remediated; "F8 mitigation… WHAT: a guard/lint so shared-component fallbacks and llm_guidance must be site-neutral" still an open Part E flag.
- **what:** A pre-guard regen (2026-07-01) baked vonc's product pitch ("Spark", the daily Gauntlet) into the shared `brief-explanation` component via three carriers invisible to the name-only F1 guard: (1) static-field fallback values; (2) those values merged into dependents' content_data by the stored⊕resolved merge; (3) per-field `llm_guidance` — the strongest, actively instructing every future writer pass on any site to write vonc's product (reproduced verbatim on robot-hands and idea.uk; contamination also migrated into generated LLM copy on pages built pre-fix — the knock-on). Remediation playbook executed: snapshot v2/v3 → neutralize fallbacks (stats→llm optional; CTAs→neutral statics) → strip merged keys with CTAS backup → scoped F3 re-renders → writer rebuilds under cleaned guidance → board-wide strpos sweep (clean except vonc's own legitimate copy). Falsified along the way: field-description carrier, content_brief column, restore-v1 option (old-architecture contract). Proposed structural mitigation (unbuilt): store-time site-neutrality lint over fallbacks + llm_guidance.
- **sources:** NOTES(43).md §9an–§9bb, §9bg; RUNBOOK(49).md Part C + Part E F8; HANDOFF(7).md §Incident 2
- **relations:** F1 guard (its blind spot); llm_guidance surface; stored⊕resolved merge; neutralize-in-place remediation; D2b lint (same detection-net shape).
- **verify-later:** brief-explanation input_schema (neutral guidance ×11, stats source=llm no fallback); component_versions v1–v3; store-side lint absence.

### llm_guidance as a per-field generation-steering surface
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Writer prompt renders per-field guidance ⇒ every writer pass on any site was instructed to write vonc's product" (NOTES §9ba); writer config read §9az.
- **what:** Each `input_schema.fields` entry may carry `llm_guidance`, which page-content-writer renders into its generate_content prompt as the field's instruction (alongside name/type/required/description; fallback values notably never enter the prompt). On a shared component this is the highest-leverage contamination/steering surface — it shapes all future content on every consuming site — and therefore must be site-neutral while preserving structural guidance (word counts, `<em>` accent rule).
- **sources:** NOTES(43).md §9az–§9bb; RUNBOOK(49).md Part C Step 7 (the 11 neutral strings)
- **relations:** F8 carrier 3; page-content-writer prompt assembly; F8 lint scope.
- **verify-later:** page-content-writer default_config generate_content prompt (llm_field_specs block); brief-explanation field attrs.

### Neutralize-in-place remediation pattern for contaminated shared components
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** "v1 not a restore candidate… NEUTRALIZE-IN-PLACE chosen" (NOTES §9ao); Steps 1–2 landed with optimistic lock held (§9ap); Steps 6–7 landed §9bd.
- **what:** When a shared component's history offers no clean restore point (v1 predated the current field contract; restoring would regress dependents on the new architecture), the fix is surgical in-place neutralization: manual snapshot first, then targeted jsonb patches replacing only the offending attributes (fallbacks, guidance) under an optimistic lock, preserving names/types/structure — followed by per-dependent cleanup (strip merged keys, scoped re-renders, writer rebuilds) mapped per consumer (vonc's own copy untouched; robot-hands stripped; old-architecture pages escalated to rebuilds).
- **sources:** NOTES(43).md §9ao–§9aq, §9bb, §9bd; RUNBOOK(49).md Part C Steps 1–9
- **relations:** F8; optimistic-lock co-management; component versioning; recovery playbook.
- **verify-later:** the CTE jsonb_object_agg patch shape in RUNBOOK(49) Step 7 as reusable SQL.

### R6f — theming vocabulary drift (defined vs consumed CSS custom properties)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "R6f confirmed as the 'renders badly' mechanism; fresh rebuilds render WORSE" (NOTES §9bh); mechanism narrowed to defined-vs-consumed drift §9am–§9bi.
- **what:** Component templates consume CSS custom properties (`var(--x, fallback)`) whose names drift from what the site's generated styles.css `:root` defines — 11 gap names in two patterns (synonyms like --border-radius vs --radius, and orphans like --hero-ink). Sections on undefined vocabulary render via per-component fallback values — a "fallback lottery" that goes dark-on-dark invisible on dark canvases (gripper-detail's blank page). Newer generators put :root in styles.css (rootless heads), older sites carry inline :root heads (why leopardess was immune). Every fresh rebuild worsened it as new templates minted new names.
- **sources:** NOTES(43).md §9al, §9am, §9bh, §9bi; RUNBOOK(49).md Part D; HANDOFF(7).md §R6f
- **relations:** D2a token aliases (fix); D2b prevention; deterministic styles.css rendering; assembly head model.
- **verify-later:** site styles.css :root contents vs template var() usage; robot-hands/vonc pre-fix stylesheets.

### D2a — buildTokenAliases renderer-enforced compatibility bridge
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "D2a VERIFIED in production 2026-07-06 — step 11 emits the alias block on a real gamesdesign pass" (NOTES §9bn); curl shows the trailing compatibility-aliases :root block.
- **what:** A step-11 post-pass in RenderCSSFromSpecAction (mirroring step 10's buildSectionDefaults "renderer-enforced" pattern): after rendering, append a trailing `:root` block defining ONLY the missing names from a package-level 11-entry alias table (synonyms → canonical var() references, orphans → palette-safe literals). Definition detection by `name+":"` so var() usages and sibling names don't count; idempotent; layout-defined values always win; one zap log field (token_alias_length). Sites self-heal on their next design pass; verified live via an adapted 076 webdesign trigger on gamesdesign.
- **sources:** NOTES(43).md §9bj–§9bn; RUNBOOK(49).md Part D D2a
- **relations:** R6f (the drift it bridges); D2b (prevention side); buildSectionDefaults (pattern precedent).
- **verify-later:** render_css_from_spec_action.go buildTokenAliases + tokenAliases table; render_css_from_spec_alias_test.go.

### D2b — canonical-token prevention (contract rule 11 + AuditTemplateTokens lint + prompt rule)
- **category:** styling-render-pipeline
- **status-signal:** partial
- **status-evidence:** "D2b in progress: lint coded… contract rule 11 drafted; prompt edit pending agent-identify" (NOTES §9bq, 2026-07-06 — thread ends here).
- **what:** Stops new orphan tokens at the source, reuse-first: (1) contract rule 11 in 003's New Component Checklist — templates reference only canonical tokens + sanctioned aliases, invent no new var(--…) (drafted as a paste-in patch); (2) the generating agent's prompt enforces the rule in place (agent identification via default_config ILIKE still pending — 21 candidates); (3) `AuditTemplateTokens` warn-only lint appended to component_validation.go (canonicalCSSTokens allowlist = 39 theme names + 11 aliases, first-seen dedup, never rejects — detection net not gate, since vocabulary evolves), pending call-site wiring where ValidateComponentTemplate already runs. Notable design subtlety: rule 11 is the reciprocal of checklist item 6 (dark sections must SET --section-*) — consume-side vs set-side.
- **sources:** NOTES(43).md §9bo–§9bq; RUNBOOK(49).md Part D D2b
- **relations:** D2a (defines the sanctioned alias set); F8 lint (sibling detection-net); contracts-and-standards checklist.
- **verify-later:** component_validation.go AuditTemplateTokens; whether wired into StoreGeneratedComponentAction; 003 doc rule 11 presence.

### Deterministic styles.css rendering (webdesign-agent: LLM spec → Go template → git commit)
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "render_css_from_spec — 'Render CSS from design spec using Go template (deterministic, no LLM)'" (NOTES §9bi, from webdesign-agent config); full chain observed live §9bn.
- **what:** The webdesign flow is analyze_design (LLM → design-spec JSON: color_scheme/typography/spacing) → render_css_from_spec (deterministic Go template over DB layout templates — `comp.LayoutTemplate` — merged with palette/typography; forkable themes) → git_commit styles.css → site-asset-renderer. The defined CSS vocabulary therefore lives in one Go-owned render path (the single home for generic fixes); storage_actions.go's styles.css writes belong to the OLD builder extract paths and must not be patched for this flow. Caution: re-running analyze_design mints a fresh LLM spec — palettes can shift unless pinned — hence the manual bridge-commit option for palette-preserving fixes.
- **sources:** NOTES(43).md §9bi, §9bj, §9bm; RUNBOOK(49).md Part D
- **relations:** D2a (lives inside it); R6f; layout curation.
- **verify-later:** render_css_from_spec_action.go; webdesign-agent default_config; needs_design production (build-site-planner).

### Scheme resolution pipeline and where the signal stops
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "Scheme→variable pipeline verified correct; all 18 layouts carry the four chrome vars" (RUNBOOK_scheme(18) CURRENT POSITION, 2026-07-02); RenderContext gap traced in running_notes Sb.
- **what:** A site's light/dark scheme derives from design intent (`deriveSchemeFromDesignIntent`), constrains layout matching (`resolveLayoutByTags`; `layouts.scheme` light/dark/neutral), and reaches styles.css via palette :root + luminance defaults — but is never recorded in the composition spec and never reaches the component render context (`RenderContext` has palette colours, no scheme field). The corrected understanding: the scheme reaches components IMPLICITLY via variables; components defeat it by hardcoding dark assumptions — so the core fix is de-hardcoding, not new plumbing.
- **sources:** running_notes_scheme_to_components(22).md Sb, Sc, Sf, Sk; RUNBOOK_scheme_to_components(18).md header + CHECK 1
- **relations:** paired-variable direction; buildSectionDefaults; R6f (later vocabulary-level echo).
- **verify-later:** deriveSchemeFromDesignIntent; RenderContext struct; layouts.scheme population (only 3 of 18 curated).

### buildSectionDefaults — luminance-keyed dark-only section context
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "--section-* is a DARK-ONLY override; light is the fallback. buildSectionDefaults returns '' unless bg or surface is dark" (running_notes Sf, from color_util.go).
- **what:** The renderer-owned `--section-*` emitter: on a dark background/surface palette it emits a body-level (and 5 hardcoded surface-class) block of readable section text vars via WCAG helpers (isDarkHex, pickReadableOnBackground); a fully light site gets nothing and element rules fall back to `--color-*`. It is the live half of the section-context mechanism and the pattern precedent for D2a's step-11 alias block.
- **sources:** running_notes(22).md Sc, Sf; RUNBOOK_scheme_to_components(18).md D1 resolution
- **relations:** SectionStyles (dead sibling); paired-variable direction (keeps this as base); D2a.
- **verify-later:** color_util.go buildSectionDefaults/isDarkHex; the 5-class list vs 025's data-section-bg plan.

### SectionStyles — the dead per-section styling mechanism and the uneven {function}-section contract
- **category:** styling-render-pipeline
- **status-signal:** abandoned
- **status-evidence:** "SectionStyles is DEAD for current sites — none of the 18 active layouts reference {{range .SectionStyles}}… computed-but-unused" (running_notes Sf); "SectionStyles stays retired" (Sn).
- **what:** A built-but-disconnected mechanism: queryDarkSectionsForCSS + buildCSSsectionStyles compute per-section entries (ClassName = function + "-section", IsDark from is_dark_section) for layout templates that no active layout consumes. The `{function}-section` class contract it assumes is real but honoured unevenly (hero emits `.hero`, CTA `.cta-section`). Decision: do not reconnect it — Phase 4.5's data-section-bg attribute keying and then the paired-variable direction supersede it.
- **sources:** running_notes(22).md Sc, Sf, Si, Sn; RUNBOOK_scheme_to_components(18).md D1
- **relations:** buildSectionDefaults (the live half); is_dark_section demotion; paired-variable direction (superseding).
- **verify-later:** render_css_from_spec_action.go buildCSSsectionStyles — still computed-and-unused?

### Paired-variable design direction (Alt C: curated bg+text pairs; completion of the existing standard)
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "pair convention is ALREADY the standard — 18/18 --color-primary-text, 17/18 --color-cta-text" (Check 4c); W1–W3e template work executed 2026-07-02/03 but "inert until re-render/rebuild"; Go batch + tail unshipped.
- **what:** The user requirement "a light scheme must be able to render fully light, and may carry dark hero bands" selects layout-curated background+text variable pairs (chrome pattern generalised: --color-cta-bg/--color-cta-text etc.), palette-overridable per site (specialised slots theme-wins), per-instance later via plan directives — components consume pairs and never declare `--section-*`; renderer luminance defaults remain the base; dark-band-by-choice stays curated per layout. Judged a COMPLETION of existing architecture, not a restructure (one layout to patch, components to bring in line). Execution: ten templates fixed + seven verified clean (footer, CTA via inverse-pair buttons, hero, five hero-* variants, about-content, brief-explanation), idea.uk chrome repointed; full rebuild + Go batch (scheme-aware fallbacks, creator prompt, fixer re-aim) pending at capture.
- **sources:** running_notes(22).md Sn, So; RUNBOOK_scheme_to_components(18).md CHECK 4 RESULTS + WHERE WE ARE; SPEC referenced therein
- **relations:** hazard/band split; hero ink model; is_dark_section demotion; Phase 4.5 (deferred); chrome linkage repair.
- **verify-later:** SPEC_scheme_to_components.md (outside unit); layouts cta pair coverage; whether W6 rebuild shipped.

### Hazard-class vs band-class self-declarer split; is_dark_section demoted to metadata
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "the 37 self-declarers split into two classes" with named components (Check 3c, run 2026-07-02); "6 declarers have is_dark_section=f… never key styling on the LLM-authored flag".
- **what:** Library-wide diagnosis: of 84 active section components, 37 self-declare `--section-*` — ~18 hazard-class (declare dark context while painting surface vars or nothing → white-on-light bugs today) vs ~19 band-class (paint palette bands + white text — coherent but block "fully light"); 15 carry hex backgrounds. `is_dark_section` is an LLM-authored component bool contradicted by 6 of its own declarers and consumed by nothing that styles — demoted to selection/imagery metadata; styling must never key on it. This classification sized every subsequent fix batch.
- **sources:** RUNBOOK_scheme_to_components(18).md CHECK 2/3 RESULTS; running_notes(22).md Sn, Sh (E findings)
- **relations:** paired-variable direction; improvement-loop fixers (key on the flag — part of why they're wrong); component-creator prompt drift.
- **verify-later:** content_components is_dark_section values vs template styling; remaining unconverted declarers (~10 hazard + ~17 band).

### Hero ink model (per-branch --hero-ink with structural-dark exception)
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "W3b COMPLETE… ink in both inline branches, layered solid+single-hue color-mix gradient" (running_notes Sv, run 2026-07-02 16:43); W3d extended it to the five variants.
- **what:** The hero's contrast contract after conversion: each branch sets an ink variable — image branch `--hero-ink:#fff` (the rgba overlay guarantees darkness: the sanctioned structural-dark exception); no-image branch `--hero-ink: var(--color-primary-text)` over a layered solid + single-hue color-mix gradient (15% toward the ink; solid layer is the color-mix-less fallback). Section vars derive from the ink at preserved alphas; buttons become the inverse pair. Fixed a latent white-on-cyan failure on the dark portal, not just the light-site problem. Imageless heroes turned out to be the COMMON case (80/114 + 26/26), reversing an assumption.
- **sources:** running_notes(22).md St–Sw; RUNBOOK_scheme_to_components(18).md HERO (c) DESIGN, W3b/W3d RESULTS
- **relations:** paired-variable direction; ambient pass-through (sibling pattern); D2a (--hero-ink later an alias orphan).
- **verify-later:** hero + hero-* html_template current state; whether rebuilds shipped the converted renders.

### Ambient pass-through pattern for surface painters with fallback-less consumers
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** "Sanctioned pattern recorded: page/surface painters with fallback-less consumers pass the ambient context through" (RUNBOOK_scheme(18) W3e, executed 2026-07-02 17:22).
- **what:** For components that paint the page/surface colour but whose internal rules consume `var(--section-*)` without fallbacks, the safe conversion is declaring `--section-x: var(--color-x)` pass-throughs rather than deleting the declarations (deletion would fall to currentColor/transparent). Scheme-correct on both light and dark by definition since the core vars ARE the scheme. Companion finding: `rgba(var(--hex-var), α)` is invalid CSS that never rendered — color-mix is the working replacement.
- **sources:** RUNBOOK_scheme_to_components(18).md W3e RESULTS; running_notes(22).md Sx
- **relations:** paired-variable direction; creator-prompt fallback mandate (future components shouldn't need this).
- **verify-later:** brief-explanation template pass-through block (NB: later regenerated in the F8 saga — check current state).

### Chrome linkage tangle: four overlapping header/footer default stores and the dark fallback
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "header_component_id is effectively a DEAD column — nothing populates it" (running_notes Sl); repoint executed (Ta); scheme-aware fallback Go batch still pending in WHERE WE ARE.
- **what:** Four coexisting default stores for site chrome: style_collections.header/footer_component_id (the store RenderHeader reads first — installed NULL and never written), site_components slots (render cache, can pin inactive components indefinitely), sites.default_components jsonb (UpdateSiteDefaultsAction target, unread on the render path), layouts.default_*_component_id (all NULL). RenderHeader's chain is collection-id → GetComponentByFunction("site-header") → RenderFallbackHeader, and the fallback hardcodes dark (PrimaryColor bg + white text) — so any linkage break yields dark chrome regardless of scheme. Fix shape: de-hardcoded active chrome components (header already model), repoint stale pins, scheme-aware fallbacks consuming the chrome var pairs (all 18 layouts already define them).
- **sources:** running_notes(22).md Sg, Sh, Sl; RUNBOOK_scheme_to_components(18).md CHECK 3b, HEAD-SLOT RESOLUTION, W4b
- **relations:** chrome refresh gating; rerender fossilisation (stale pinned renders reached deploys); paired-variable direction.
- **verify-later:** style_collections.*_component_id population; RenderFallbackHeader/Footer/Head current CSS; whether the Go batch shipped.

### Improvement-loop colour fixers are scheme-blind (re-aim as backstop, not fix)
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** "Running them as-is on idea.uk would ENTRENCH dark, not lighten it" (running_notes Sj); user steer Sk: fix at initial render, fixers stay backstop; re-aim listed as pending Go-batch work.
- **what:** Existing fixer machinery — color-variable-fixer (fix_hardcoded_colors: dark hex→var(--color-primary), leaves rgba overlays; fix_forced_text_colors: strips forced child text colours, WCAG-validates, but INJECTS the white --section-* contract for is_dark_section components), nav-link-fixer (+ its documented render_site_components force_rerender follow-up), fix_component_template symptom fixes — de-hardcodes template + rendered HTML but enforces the OLD component-owns-dark contract keyed on is_dark_section, pulling opposite to the chosen architecture. Decision: initial-render fixes are primary; fixers get re-aimed to enforce the same paired-variable contract as backstop (key on what the template paints, never is_dark_section).
- **sources:** running_notes(22).md Sj, Sk; RUNBOOK_scheme_to_components(18).md WHERE WE ARE (fixer re-aim in Go batch)
- **relations:** hazard/band split; paired-variable direction; D2b (later analogous prevention-over-cure shape).
- **verify-later:** fix_forced_text_colors / fix_hardcoded_colors current keying; whether re-aim shipped.

### Layout curation: CTA pair completion, WCAG contrast batch, updated_at trigger
- **category:** design-composition
- **status-signal:** deployed
- **status-evidence:** "W1 COMPLETE… W1b: five × UPDATE 1… trigger observed working in anger" (RUNBOOK_scheme(18) W1/W1b/W2b RESULTS, 2026-07-02).
- **what:** Seed-layout curation as part of the theming fix: tool-portal-light gained the missing --color-cta-bg/--color-cta-text pair (#e9e2d3/#1a1a1a, ≈13.5 contrast); a five-layout batch nudged failing cta_bg fallbacks to same-hue passes (all ≥4.5); layouts.updated_at gained a BEFORE UPDATE trigger via the shared set_updated_at function (reuse-gate fired as designed when CREATE FUNCTION collided). Several light layouts deliberately curate dark footer bands — "light site, dark band by choice" is an existing curated model, not a bug.
- **sources:** RUNBOOK_scheme_to_components(18).md W1/W1b/W2b RESULTS, CHECK 4b; running_notes(22).md Sq–Ss
- **relations:** paired-variable direction; deterministic styles.css rendering.
- **verify-later:** layouts cta pair values; trg_layouts_updated_at.

### sites.status is an informational lifecycle label (validated vocabulary, no consumer filters)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "No on-disk code filters sites on status — it is an informational lifecycle label; build dispatch keys on site_work_items" (RUNBOOK_scheme(18) sites.status RESOLVED, from v3_site_actions.go:323–395).
- **what:** UpdateSiteStatusAction validates status ∈ {draft, building, review, published, deployed, archived, error} (and stamps last_deployed_at with status=deployed); 'active' and 'system' are legacy out-of-vocabulary values on old rows. Nothing filters sites by status at build time — an assumption (`WHERE s.status='active'`) borrowed from an old handoff silently wrecked a blast-radius count, hence the standing rule: never filter on status='active'.
- **sources:** RUNBOOK_scheme_to_components(18).md §sites.status RESOLVED; running_notes(22).md Sr, Ss
- **relations:** work-item dispatch (the real gate); needle-gate/verify-at-point-of-use discipline.
- **verify-later:** UpdateSiteStatusAction vocabulary; legacy-status rows.

### Work-item dedup semantics (item_key + idx_swi_dedup partial unique index)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "idx_swi_dedup UNIQUE (site_id, item_key) WHERE item_key IS NOT NULL AND status NOT IN (complete, verified, rejected, wont_fix, failed, unresolved)" — captured from pg_indexes (NOTES §9aa).
- **what:** site_work_items dedup is two-layered: producers guard with NOT EXISTS over open statuses and the DB enforces a partial unique index on (site_id, item_key) over non-terminal statuses. Terminal-status items free the key (why completed triggers can be re-inserted for retriggers); mirroring the producer's exact insert (columns, item_key scheme, dedup clause) is the established way to hand-create conforming items. See F6 for the known guard/index mismatch.
- **sources:** NOTES(43).md §9q, §9aa; RUNBOOK(49).md Part C Step 9b; w4b_03_read_rerender_config.sql
- **relations:** F6; work-item spec-cloning discipline; F3.
- **verify-later:** idx_swi_dedup definition; createRerenderWorkItem insert shape.

### Loose dispatch item-status semantics (complete ≠ done)
- **category:** debugging
- **status-signal:** aspirational
- **status-evidence:** "loose dispatch item-status semantics documented across the investigation (complete-at-dispatch, errors-in-complete, status-change without a timestamp bump, parent-topic-vanished noise) — worth a pass when convenient" (RUNBOOK(49) Part E Hygiene); seven dated sightings in NOTES.
- **what:** A documented defect class in the dispatch loop's work-item bookkeeping, observed seven times: items marked 'complete' at dispatch while the child orchestration runs or fails later; the child's full error text stored in the `error` column of a 'complete' item; status transitions that don't bump updated_at; batch claim stamps shared across differently-fated items; parent fire-and-forget topic lifecycle polluting child completions ("topic partition not found"). Operational rule derived: never trust item status as proof of work — verify the artefact (band stamp, render md5); agent_error_log (occurred_at) outranks status. Fix parked as hygiene.
- **sources:** NOTES(43).md §9i, §9l, §9m, §9aa, §9ac, §9ax, §9bd; RUNBOOK(49).md Step 9 reading guide + Part E
- **relations:** work-item dedup; F2 methodology (discriminator ordering); auto-escalation.
- **verify-later:** build-dispatch-loop status handling; whether items get failure statuses on child errors.

### Deploy-ordering hard gate for coupled Go action + workflow-config changes
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "LESSON (runbook gate tightened): 'deploy the Go action first' is insufficient — the migration is live instantly while the image rolls out. Hard gate: confirm… registered + live on ALL pods, THEN apply the migration" (NOTES §9i–§9j).
- **what:** Workflow jsonb changes take effect immediately; Go actions only exist once the image is rolled out and the registry entry (IsLocal:true) is in the running build. Wiring a workflow step to a not-yet-live action makes the validator reject EVERY run of that agent (WORKFLOW_INVALID broke all component generation during F2 3a). The codified gate: deploy + confirm the action responds on all pods before applying the (idempotent) migration; `revert_agent('<type>')` is the immediate mitigation.
- **sources:** NOTES(43).md §9i, §9j; F1prompt_component_creator_preserve_field_names(1).sql PREREQUISITE header
- **relations:** F1-prompt (where it bit); prompt-migration convention; snapshot/revert_agent.
- **verify-later:** workflow validator is_local check; revert_agent/snapshot_agent functions.

### Prompt/workflow-jsonb migration convention (snapshot-first, anchored, idempotent, drift-checked; the 072 trap)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** The F1prompt migrations implement it end-to-end ("Convention: snapshot-first, idempotent, drift-checked, live-row only" — F1prompt(1).sql); prompt located top-level after `prompt_is_top_level=f` proof (NOTES §9d).
- **what:** Agent behaviour lives in default_config jsonb; edits are live instantly and follow a strict convention: snapshot_agent first; anchor the edit on a unique existing string and abort if the anchor count ≠ 1 (drift check); idempotency marker so re-runs no-op; filter to the live row (is_active, not snapshot, not deleted). The "072 nested-prompt trap": prompt_template may live at the top level of default_config OR nested in a step config — verify the path first or the migration is a silent no-op. Anti-drift prompt anchors have precedent (tool-doc-header rule on tool-improver): prompt rule = the anchor, store guard = the gate.
- **sources:** F1prompt_component_creator_preserve_field_names(1).sql; NOTES(43).md §9c, §9d, §9k (dead-block cleanup — an idempotency-check subtlety)
- **relations:** F1-prompt; F3c config edit; D2b-2 prompt edit; deploy-ordering gate.
- **verify-later:** snapshot_agent/revert_agent; component-creator prompt state (no dead {{if .existing_field_names}} block).

### F2 tiered guard-verification methodology (unit → integration → live keep/reject fixtures)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "F2 COMPLETE: Tier 1 unit ✔, 3a preservation ✔ (×2 regens, md5-verified template change), 3b reject ✔ (live firing, three-level visibility, zero mutation)" (NOTES §9o).
- **what:** The verification pattern used to prove F1 without touching live shared components: Tier 1 deterministic unit tests of the guard logic (including the real incident's rename case); Tier 2 DB-backed reject-path test (folded into Tier 3 when no harness existed); Tier 3 end-to-end on throwaway zzz-* components — a KEEP fixture proving preservation-by-instruction (non-guessable check: template md5 changes while fields hold) and an intentionally INACTIVE REJECT fixture exploiting the store-vs-loader is_active divergence to force a rename and observe the guard fire live with zero mutation. Also codified the discriminator ordering (agent_error_log > pod logs > never item status) and prompt cleanup of leftover fixtures.
- **sources:** NOTES(43).md §9f, §9h, §9k–§9o; RUNBOOK(30) family (Step F2 tiers)
- **relations:** F1 guard; F4 (discovered by 3b run 1); loose status semantics.
- **verify-later:** store_generated_component_guard_test.go; zzz fixtures fully cleaned.

### Optimistic-lock co-management of shared rows across parallel chats
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "Optimistic-lock pattern for co-managed writes: WHERE updated_at = '<last-known>' — UPDATE 0 = stop, re-read, coordinate" (RUNBOOK(49) Constants); lock held across 3 idle days on the Step-7 write (NOTES §9bd).
- **what:** Multiple concurrent chats/agents co-manage the same shared components and sites, so every write to co-managed rows is guarded by a freshness check plus an optimistic-lock UPDATE on the last-known updated_at; UPDATE 0 means the row moved underneath — stop and coordinate, never blind-write. Includes proactive notification of the other chat and re-reading fleet/workflow state after parallel deploys (image bumps invalidate cached workflow knowledge).
- **sources:** RUNBOOK(49).md Constants; NOTES(43).md §3, §9ao, §9ap, §9bd; HANDOFF(7).md §Platform operating model; RUNBOOK_pre_cleanup_backup.md Step R7
- **relations:** snapshot-before-change; F8 remediation; locks category (031) more broadly.
- **verify-later:** whether any systematic lock/lease exists beyond the convention (relates to locks/031 concepts).

### Snapshot-before-change backup conventions (snapshot_agent, manual version inserts, CTAS bak tables)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** "Backups: snapshot_agent('<type>','<reason>') for agents; CTAS *_bak_* tables for data (two exist, see Cleanup)" (HANDOFF(7)); every mutating step in both threads shows a backup first.
- **what:** The layered backup discipline observed throughout: agents → snapshot_agent before config migrations (revert_agent to restore); shared component schema/template → manual component_versions INSERT mirroring the working insert paths; data rows → CREATE TABLE … AS SELECT `*_bak_*` tables (dropped only at closeout); templates → shell-redirected full-column dumps before anchored UPDATEs. Backups are named with dates and tracked as explicit cleanup debt.
- **sources:** HANDOFF(7).md §operating model; RUNBOOK(49).md Constants + Step 12(c); RUNBOOK_scheme_to_components(18).md W1/W2a backup steps
- **relations:** optimistic-lock co-management; component versioning; site-snapshots-and-revert (014) more broadly.
- **verify-later:** snapshot_agent/revert_agent SQL functions; outstanding bak tables.

### Cold-start documentation bundle practice (BUNDLE/HANDOFF/PLAN/RUNBOOK + cmd/bundle)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** "Decision: produce a cmd/bundle invocation + cold-start docs (BUNDLE/HANDOFF/PLAN/RUNBOOK) so a fresh chat could pick it up" (NOTES §2 Start); HANDOFF(7) explicitly "the cold-start entry point".
- **what:** The thread's working method: a four-document travelling set per investigation — BUNDLE (a `cmd/bundle` invocation composing constitution + task + scoped code symbols + schemas + runtime evidence into one context file; `-step debug` for bodies, verified doc paths), HANDOFF (cold-start entry with operating model + status), PLAN (phased with gates/done-whens), RUNBOOK (live action document with YOU-ARE-HERE banner, per-step SQL + expected + CHECK blocks, ticked progress) — plus NOTES as the append-only journal owning every correction. Operational gotchas folded in: pasted attachments extract empty (capture via kubectl…psql -c > file, not \o); runbooks rewritten wholesale when history outgrows action (old kept as *_pre_cleanup_backup).
- **sources:** BUNDLE(3).md; HANDOFF(7).md; NOTES(43).md §2, §9av, §9bc; RUNBOOK(49).md structure
- **relations:** documentation-system conventions (037); F2 discriminator discipline.
- **verify-later:** cmd/bundle tool flags (-scope/-schema-tables/-runtime-site/-df-filter).

### Superseded hypothesis: update_component_html re-renders dependents inline
- **category:** NEW:component-lifecycle
- **status-signal:** superseded
- **status-evidence:** Original BUNDLE base: "the regeneration then re-renders every dependent page… rewritten together at 15:06:12.956, roughly 16ms after the component's own update"; Correction 1: "update_component_html is clean… the synchronized timestamp is just the pending-flag, not a render".
- **what:** The initial working theory held that update_component_html performed an inline dependent re-render (inferred from the ~16ms synchronized timestamps) and was the clobber writer. Disproved by reading the action: it only snapshots, swaps, and flags pending; the blame moved through RenderComponentAction and component-creator's workflow (both cleared) to StoreGeneratedComponentAction. Worth keeping as the exemplar of the thread's core epistemics: seven early inferences were each corrected against code/data before any fix shipped ("distrust each early inference until verified").
- **sources:** BUNDLE_component_regen_clobber.md §1 (base version); NOTES(43).md §2 Corrections 1–7, §3
- **relations:** clobber failure mode; StoreGeneratedComponentAction (actual writer).
- **verify-later:** n/a (historical).

### page-build-handler writes only planned sections (sections=0 → silent no-op rebuild)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "matchmatrix sections_listed=0 vs siblings 9/5 ⇒ the page-build-handler iterates the planned section list and had nothing to write — planning-data gap on a legacy page — PARKED" (NOTES §9bh).
- **what:** The writer rebuild flow iterates the page's planned section list; a legacy page with an empty `pages.sections` plan completes its needs_page item without writing anything — a silent no-op rebuild indistinguishable from success at the item level (double no-op observed before diagnosis; page_type hypothesis falsified first). Remediation options (planner reconcile/adopt, or retire the page) parked as hygiene.
- **sources:** NOTES(43).md §9bf–§9bh; RUNBOOK(49).md Step 12(a)
- **relations:** auto-escalation; loose status semantics; site-plan-and-reconciler (planner as another item producer, §9at).
- **verify-later:** page-build-handler section iteration; matchmatrix pages.sections.

### Never-guess disciplines: clone real work-item specs, look up real URLs (phantom-CTA lesson)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "NEVER guess a needs_page spec" (HANDOFF(6→7)); "MY CORRECTION: I had baked a guessed path… replaced with a pages.url lookup query (never guess paths; pages.url is the source)" (NOTES §9ae); spec-shape reads staged before every insert (w4b_03).
- **what:** A cluster of verify-at-point-of-use rules that recur across both threads: hand-created work items are mirrored column-for-column from a real captured item (SELECT-based inserts, real spec shapes, conforming item_keys), never composed from memory; URLs/paths come from pages.url, never invented (the phantom-CTA bug was an invented /contact.html; the recovery re-verified the same value before trusting it); schema before SQL (column names checked against \d, e.g. occurred_at not created_at); trigger flows through the real producer path rather than manual inserts where one exists (needs_design via build-site-planner / the proven 076 trigger).
- **sources:** NOTES(43).md §9l, §9ae, §9w–§9y, §9bd, §9bl–§9bm; w4b_03_read_rerender_config.sql; HANDOFF(7).md §Immediate next action
- **relations:** work-item dedup; section readiness (spec presence ≠ validity); F2 discriminators; link-management (phantom links).
- **verify-later:** n/a (convention; instances cited).

### Needle-gate SQL template surgery pattern (and its catalogued pitfalls)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** Applied on every template mutation W1–W3e with gates, RETURNING checks and verify regions; pitfalls promoted into 016b guide v4 ("count expectations mechanically from the dump… never from memory").
- **what:** The method for mutating shared templates/configs safely by SQL: dump + shell backup first; a gate query asserting exact needles (booleans) and mechanical occurrence counts derived by grep from the dump (mismatch = drift OR mis-derived expectation — stop); anchored exact-string or backreference replaces (multi-line needles to disambiguate repeated strings); guards for idempotency; RETURNING post-conditions; separate verify file; value-agnostic rollback file. Catalogued Postgres pitfalls: regex quantifier bound ≤255; substring() returns the first capture group; LIKE-wildcard `%` inside needles (use position()); `\set ON_ERROR_STOP on` when statements depend on earlier ones; run SQL as files, never pasted.
- **sources:** RUNBOOK_scheme_to_components(18).md W1–W3e blocks + RESULTS; running_notes(22).md Sr, Sv, St
- **relations:** prompt-migration convention (same family for jsonb); debugging guide 016b (where the lessons were codified).
- **verify-later:** 016b guide entries; w*_*.sql files referenced (outside unit).

### R6c artifact-forensics method: cache-busted, metric-consistent comparisons
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "md5sum: gd.html == gd2.html… ONE artifact all along. OWNED: my stale-cache story AND the earlier '4-of-8 mis-assembled' reading were metric artifacts" (NOTES §9al).
- **what:** Lessons from the gripper-detail "blank page" false trail: compare live artifacts only with identical metrics (a data-component inventory vs a class grep counted different things and manufactured a mis-assembly story); md5 the fetches before concluding stale-cache; distinguish 404/200-empty/200-styled-invisible with curl size + head; visually-blank ≠ missing content (fallback-vars insight: content present but dark-on-dark). The eventual truth (theming, not assembly or deploy) reshaped Part D.
- **sources:** NOTES(43).md §9af–§9al; RUNBOOK(49).md Part B
- **relations:** R6f (the real mechanism); assembly membership model; needle-gate pattern (same mechanical-counting ethos).
- **verify-later:** n/a (method; instances cited).

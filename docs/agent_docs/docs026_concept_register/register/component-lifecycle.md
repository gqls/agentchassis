# Register — component-lifecycle

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

11 concepts, consolidated from 11 raw extractions across unit U07. No duplicates found within this category — it arrived as a single tightly-scoped, already well-differentiated F1–F8 investigation saga (the vonc/spark shared-component clobber incident and its remediation), so all 11 raw blocks are carried through as distinct entries with only light formatting changes.

### CLC-001 — Shared content-component reuse model (one content_components row, N page_components instances)
- **status:** deployed
- **status-evidence:** Stated as settled platform fact throughout the investigation ("one content_components row, N page_components instances"); a live site's homepage arrived mid-recovery as a "healthy sixth dependent," proving live reuse.
- **what:** A section component is a single shared content_components row (keyed by `function`, with section_type, input_schema.fields, html_template, is_active, forked_from) reused across pages and sites; each page stores its own content_data in page_components. Any change to the shared row therefore has a cross-site blast radius — the structural precondition for the incidents this whole saga investigates.
- **sources:** BUNDLE(3).md §1; NOTES(43).md §1, §9z (sixth dependent); HANDOFF(7).md §Platform operating model
- **relations:** clobber failure mode; F4 fork-vs-match (CLC-006); F8 contamination; optimistic-lock co-management; tool-library's Shared component library semantics (TLIB-022, the same principle from a different incident)
- **verify-later:** content_components + page_components schemas; idx_cc_selector (section_type, component_level) partial index

### CLC-002 — StoreGeneratedComponentAction regeneration branch
- **status:** deployed
- **status-evidence:** "Correction 6 — StoreGeneratedComponentAction is the rename writer"; insert/dedup shape read from the live file.
- **what:** On storing a generated component whose `function` matches an existing row (`WHERE function=$1 AND forked_from IS NULL`, deliberately is_active-agnostic since 2026-05-06), the action snapshots the old schema/template to component_versions, UPDATEs the shared row in place (same component_id, so dependents follow), marks dependents pending (markPagesPendingRebuild — build_status only, no rendered_html), and raises one deduped needs_rerender work item per affected site via createRerenderWorkItem. Pre-fix, that item carried no `reason`, making the triggered re-render assemble-only and unable to repair anything.
- **sources:** NOTES(43).md §2 Correction 6, §9h, §9q; BUNDLE(3).md §3
- **relations:** F1 guard lives in its validation block (CLC-003); F3b re-added the reason to its spec; F4 (CLC-006, its function-keyed lookup is the fork vector); Component regeneration in place (tool-lifecycle TL-026, the same mechanism via a different real incident)
- **verify-later:** store_generated_component_action.go (existence check L198–207; regen branch; markPagesPendingRebuild; createRerenderWorkItem NOT EXISTS dedup)

### CLC-003 — F1 field-contract guard (reject regens that rename/drop retained fields)
- **status:** deployed
- **status-evidence:** "F2 COMPLETE... 3b reject live firing, three-level visibility, zero mutation" (2026-07-02); a companion runbook Part A: "Fixes deployed + proven."
- **what:** In StoreGeneratedComponentAction's Layer-1 validation, on isRegeneration the guard diffs old vs new input_schema.fields names (a schemaFieldSet helper); any retained field that disappears becomes a blockingIssue routed through recordValidationRejection into agent_error_log — additions allowed, renames/drops rejected before any snapshot/UPDATE. Converts silent stranding into a loud, queryable rejection naming the stranded fields. Design choice: a preserve-the-contract strict-reject backstop, not a per-dependent migration. This is the direct code-level fix for the system-stats field-rename incident (tool-library TLIB-005).
- **sources:** F1_store_generated_component_action.patch; NOTES(43).md §9, §9a, §9o; RUNBOOK(49).md Part A
- **relations:** F1-prompt (CLC-004, generation-time complement so name-preserving regens succeed); F5 (CLC-007, proposed extension); F8 (guard checks names only — its blind spot); Validation observability (tool-lifecycle TL-020, the logging mechanism this guard writes through)
- **verify-later:** store_generated_component_action.go guard block + schemaFieldSet; agent_error_log rows error_code component_validation_rejected; store_generated_component_guard_test.go

### CLC-004 — F1-prompt generation-time field-name preservation (loader + dormant rule + function pin)
- **status:** deployed
- **status-evidence:** "F1-prompt 1a/1b+2 done (validator passes; loader executes; verify row good)"; two regens preserved names with md5-verified template change.
- **what:** Three coupled pieces so regens preserve names instead of being rejected by CLC-003: (1) load_existing_component Go action — looks up the canonical component by section_type (is_active, forked_from IS NULL, component_level='section'), outputs existing_component.field_names + function; advisory, never errors (no match → empty map → rule dormant). (2) A snapshot-first, anchored, idempotent, drift-checked SQL migration wiring the step before generate_template and inserting a dormant `{{if .existing_component.field_names}}` prompt rule: reuse these exact names, MAY add, MUST NOT rename/drop. (3) A companion migration pins `{{.existing_component.function}}` so the store matches the same row (the mitigation for F4, CLC-006). Option A (pre-generation lookup by section_type) was chosen after confirming section_type is queryable.
- **sources:** F1prompt_component_creator_preserve_field_names(1).sql; NOTES(43).md §9c–§9f, §9m; RUNBOOK(49).md Part A
- **relations:** F1 guard (CLC-003, backstop); F4 (CLC-006, the pin is its mitigation); prompt-migration convention; deploy-ordering gate (its 9i failure)
- **verify-later:** load_existing_component_action.go + registry.go (IsLocal:true); component-creator default_config (top-level prompt_template; load_existing_component step; input_fields incl. existing_component)

### CLC-005 — Store-driven retry on field-drift rejection (Option B) — abandoned alternative
- **status:** abandoned
- **status-evidence:** "Two options... user chose A"; Option B never mentioned again after being raised once.
- **what:** The rejected alternative to Option A (CLC-004): on a field-drift rejection the guard would return the existing field names, and a store_component error edge would loop back to generate_template with the names injected, retrying once. Judged heavier (reject-with-retry-data + loop-guarded workflow edge) but authoritative (no key guessing); dropped when section_type proved a stable pre-generation lookup key.
- **sources:** NOTES(43).md §9d, §9e
- **relations:** F1-prompt Option A (CLC-004, the superseding choice)
- **verify-later:** absence: component-creator workflow should have no store_component→generate_template error edge

### CLC-006 — F4 — regen-vs-create keyed on the LLM-chosen function (silent fork)
- **status:** partial
- **status-evidence:** "F4 (structural finding): regen-vs-create is keyed on the LLM-chosen function → nondeterministic. Miss case = silent FORK"; pin migration applied; "store-side advisory FLAGGED as follow-up, not built."
- **what:** Whether a store is a regeneration or a creation depends on whether the LLM happened to choose the existing function name — a miss silently creates a parallel active duplicate for a section_type (library fragmentation; the CLC-003 guard is bypassed by fork; selector nondeterminism results). Observed live in testing (fork 80222fc1). Mitigation shipped: the prompt pin of `{{.existing_component.function}}` (CLC-004); a store-side advisory (warn when function misses but an active same-section_type row exists) is deliberately advisory-only and unbuilt, since multiple components per section_type can be legitimate. A suspected live-fork case (duplicate hero rows) was later softened to old manual seeding rather than a live recurrence.
- **sources:** NOTES(43).md §9m, §9n, §9al, §9am; RUNBOOK(49).md Part E
- **relations:** F1-prompt function pin (CLC-004); StoreGeneratedComponentAction lookup (CLC-002); F2 methodology (exposed it); Component regeneration in place (tool-lifecycle TL-026, the same determinism hazard via a separate incident)
- **verify-later:** duplicate non-forked function rows in content_components; whether any store-side advisory exists

### CLC-007 — F5 — regen-added required fallback-less fields strand renderability
- **status:** aspirational
- **status-evidence:** "Flagged F5: extend the F1 guard... Not built now"; still an open flag in a later runbook.
- **what:** The incident's second facet: a regen also ADDED a required field (Tier-C source, no fallback) that no affected site's specs could satisfy — renames strand stored content (CLC-003's concern), required additions strand renderability instead (sections permanently not-ready → carried forever). Proposed guard extension: reject, or force optional/fallbacked, any added required field on a regeneration.
- **sources:** NOTES(43).md §9v; RUNBOOK(49).md Part E F5; HANDOFF(7).md §Flags
- **relations:** F1 guard (CLC-003); section readiness model; carry-forward path
- **verify-later:** store_generated_component_action.go — absence of an added-required-field check

### CLC-008 — F7 — unguarded template swap in update_component_html (residual)
- **status:** partial
- **status-evidence:** "the snapshot INSERT is ALREADY FIXED in current code... Residual: no placeholder⇄schema sync validation on template swaps"; a later runbook flag remains open.
- **what:** update_component_html swaps a shared component's template (snapshotting versions — its old silent snapshot failure on a removed version_note column is fixed) but performs no placeholder⇄schema agreement validation and no field-contract guard, leaving a second, unguarded write path to shared components alongside StoreGeneratedComponentAction (CLC-002). The original F7 framing (an unversioned live swap on a hero component) was investigated and softened; the residual finding is specifically the missing validation, not the original unversioned-write concern.
- **sources:** NOTES(43).md §9aj, §9ak, §9am; RUNBOOK(49).md Part E F7; HANDOFF(7).md §Flags
- **relations:** F1 guard (CLC-003, candidate extension target); Component versioning (CLC-009)
- **verify-later:** update_component_html_action.go — snapshot INSERT columns; absence of schema-sync validation

### CLC-009 — Component versioning via component_versions (and unversioned-write provenance)
- **status:** partial
- **status-evidence:** "Component updated AGAIN 2026-07-03 13:22:44 with NO v2 row... unversioned write path provenance OPEN"; manual snapshots (v2/v3) were taken to backfill.
- **what:** component_versions snapshots (component_id, version_number, schema, template, change_description, changed_by, change_source) are the change history for shared components; change_source records the triggering work item's source (useful provenance). Coverage is incomplete: some write paths historically failed silently or bypass versioning entirely, so manual mirror-the-working-insert snapshots are the established compensation before risky writes, and zero-version updates are treated as an investigation smell. This is the fuller, currently-active picture of the same table an older schema-archaeology unit (tool-lifecycle TL-025) found "unclear whether any writer maintains."
- **sources:** NOTES(43).md §9k, §9an, §9ao, §9bd; RUNBOOK(49).md Part C Step 6
- **relations:** F7 (CLC-008); snapshot-before-change conventions; F8 remediation (v2/v3 snapshots); Component versioning (tool-lifecycle TL-025, the older/thinner evidence base for the same table)
- **verify-later:** component_versions rows for the incident components; snapshotComponentVersion call sites

### CLC-010 — llm_guidance as a per-field generation-steering surface
- **status:** deployed
- **status-evidence:** "Writer prompt renders per-field guidance ⇒ every writer pass on any site was instructed to write [one site's] product" — a real contamination-risk finding; writer config confirmed by direct read.
- **what:** Each input_schema.fields entry may carry llm_guidance, which page-content-writer renders into its generate_content prompt as the field's instruction (alongside name/type/required/description; fallback values notably never enter the prompt). On a shared component this is the highest-leverage contamination/steering surface — it shapes all future content on every consuming site — and therefore must be site-neutral while preserving structural guidance (word counts, an accent-markup rule).
- **sources:** NOTES(43).md §9az–§9bb; RUNBOOK(49).md Part C Step 7 (the 11 neutral strings)
- **relations:** F8 carrier 3; page-content-writer prompt assembly; F8 lint scope; Shared component library semantics (tool-library TLIB-022)
- **verify-later:** page-content-writer default_config generate_content prompt (llm_field_specs block); field attrs on the affected shared component

### CLC-011 — Superseded hypothesis: update_component_html re-renders dependents inline
- **status:** superseded
- **status-evidence:** Original working theory: "the regeneration then re-renders every dependent page... rewritten together at 15:06:12.956, roughly 16ms after the component's own update"; Correction 1: "update_component_html is clean... the synchronized timestamp is just the pending-flag, not a render."
- **what:** The initial working theory held that update_component_html performed an inline dependent re-render (inferred from ~16ms synchronized timestamps) and was the clobber writer. Disproved by reading the action: it only snapshots, swaps, and flags pending; the blame moved through RenderComponentAction and component-creator's workflow (both cleared) to StoreGeneratedComponentAction (CLC-002). Worth keeping as the exemplar of the thread's core epistemics: seven early inferences were each corrected against code/data before any fix shipped — "distrust each early inference until verified."
- **sources:** BUNDLE_component_regen_clobber.md §1 (base version); NOTES(43).md §2 Corrections 1–7, §3
- **relations:** clobber failure mode; StoreGeneratedComponentAction (CLC-002, the actual writer)
- **verify-later:** n/a (historical)

### CLC-012 — teaser-reveal-panel: a second component implementing an existing experience pattern
- **status:** deployed (registered fleet-wide + placed and served on one page)
- **status-evidence:** `content_components` row `function='teaser-reveal-panel'`, `component_level='section'`, `is_active`; `js_snippets` row of the same name, `applies_to:["teaser-reveal-panel"]`; render harness `docs024_key_docs_latest/brochure_component_library/scripts/render_teaser_reveal_panel.go` — 14 checks, proven non-vacuous by two mutants (a bodyless item given a body fails 6, an ellipsis continuation fails 1). Live on fundamentallyai.com/index.html from 2026-07-29.
- **what:** A section component rendering a panel of teasers — a short hook plus a deliberately unfinished continuation — whose full text opens IN PLACE and gains a shareable `?open=<key>` URL. Reachable by any site: `load_component_library` returns every active section component with no `suitable_site_types` gate, so another workstream can plan it without knowing it was built here. Three properties are the point, and each is a rule some other mechanism would otherwise break: (1) the reveal is native `<details>`/`<summary>`, so it works with JavaScript disabled and the JS snippet adds *only* addressability — with the snippet absent nothing becomes a dead control; (2) the body stays in the DOM permanently, because copy that appears only after a JS call is invisible to the claims gate and to crawlers — the same blind spot as text inside `<svg>` (`claims.go:137` `nonAssertionElements`); (3) an item with no body renders as a plain statement with NO control, and carries no cliffhanger mark — you may only tease what you can deliver. The cliffhanger is marked structurally (`data-continues="true"`), never with an ellipsis, so a truncation checker built on `output_tokens == max_tokens` can distinguish intent from damage.
- **sources:** brochure_component_library/PLAN_2026-07-29_teaser_reveal_panel.md; components/teaser-reveal-panel/{template.html,input_schema.json,behaviour.js,register.sql}; NOTES 2026-07-29 entry
- **relations:** implements the pre-existing `experience_patterns` row `teaser-detail-deeplink` (kind `micro-journey`) rather than declaring a new shape — the pattern's `section_types` does not yet name it (the experience-register thread's rows to change, proposed not taken); js_snippets delivery lane (NOT `js_content`, `bugs_open/041`); llm_guidance as a steering surface (CLC-010) carries the figure-splitting rule; claims-gate ±70-char context window (`datahelpers/claims.go`)
- **verify-later:** whether the planner ever SELECTS it (`features_open/017`, unobserved for every component this workstream has built); whether `teaser-detail-deeplink`'s "detail region must be emptied on close" clause should be narrowed — it is a property of vonc's JS-populated implementation, not of the shape

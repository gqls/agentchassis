
<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Shared content-component reuse model (one content_components row, N page_components instances)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** Stated as settled platform fact throughout ("one `content_components` row, N `page_components` instances", BUNDLE(3) §1); vonc.com/index arrived mid-recovery as "healthy sixth dependent" proving live reuse.
- **what:** A section component is a single shared `content_components` row (keyed by `function`, with `section_type`, `input_schema.fields`, `html_template`, `is_active`, `forked_from`) reused across pages and sites; each page stores its own `content_data` in `page_components`. Any change to the shared row therefore has a cross-site blast radius — the structural precondition for both incidents in this thread.
- **sources:** BUNDLE(3).md §1; NOTES(43).md §1, §9z (vonc sixth dependent); HANDOFF(7).md §Platform operating model
- **relations:** clobber failure mode; F4 fork-vs-match; F8 contamination; optimistic-lock co-management.
- **verify-later:** content_components + page_components schemas; idx_cc_selector (section_type, component_level) partial index.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### StoreGeneratedComponentAction regeneration branch
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Correction 6 — StoreGeneratedComponentAction is the rename writer" (NOTES §2); insert/dedup shape read from the live file (NOTES §9q).
- **what:** On storing a generated component whose `function` matches an existing row (`WHERE function=$1 AND forked_from IS NULL`, deliberately is_active-agnostic since 2026-05-06), the action snapshots the old schema/template to `component_versions`, UPDATEs the shared row in place (same component_id, so dependents follow), marks dependents pending (`markPagesPendingRebuild` — build_status only, no rendered_html), and raises one deduped `needs_rerender` work item per affected site via `createRerenderWorkItem`. Pre-fix, that item carried no `reason`, making the triggered re-render assemble-only and unable to repair anything.
- **sources:** NOTES(43).md §2 Correction 6, §9h, §9q; BUNDLE(3).md §3
- **relations:** F1 guard lives in its validation block; F3b re-added the reason to its spec; F4 (its function-keyed lookup is the fork vector).
- **verify-later:** store_generated_component_action.go (existence check L198–207; regen branch; markPagesPendingRebuild; createRerenderWorkItem NOT EXISTS dedup).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F1 field-contract guard (reject regens that rename/drop retained fields)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "F2 COMPLETE… 3b reject ✔ (live firing, three-level visibility, zero mutation)" (NOTES §9o, 2026-07-02 16:39); RUNBOOK(49) Part A "Fixes deployed + proven".
- **what:** In StoreGeneratedComponentAction's Layer-1 validation, on `isRegeneration` the guard diffs old vs new `input_schema.fields` names (`schemaFieldSet` helper); any retained field that disappears becomes a blockingIssue routed through `recordValidationRejection` into `agent_error_log` — additions allowed, renames/drops rejected before any snapshot/UPDATE. Converts silent stranding into a loud, queryable rejection naming the stranded fields. Design choice: preserve-the-contract strict-reject backstop, not a per-dependent migration.
- **sources:** F1_store_generated_component_action.patch; NOTES(43).md §9, §9a, §9o; RUNBOOK(49).md Part A
- **relations:** F1-prompt (generation-time complement so name-preserving regens succeed); F5 (proposed extension); F8 (guard checks names only — its blind spot).
- **verify-later:** store_generated_component_action.go guard block + schemaFieldSet; agent_error_log rows error_code component_validation_rejected; store_generated_component_guard_test.go.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F1-prompt generation-time field-name preservation (loader + dormant rule + function pin)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "F1-prompt 1a/1b+2 done (validator passes; loader executes; verify row good)" (NOTES §9k); Tier 3a two regens preserved names with md5-verified template change (§9l).
- **what:** Three coupled pieces so regens preserve names instead of being rejected: (1) `load_existing_component` Go action — looks up the canonical component by `section_type` (is_active, forked_from IS NULL, component_level='section'), outputs `existing_component.field_names` + `function`; advisory, never errors (no match → empty map → rule dormant). (2) A snapshot-first, anchored, idempotent, drift-checked SQL migration wiring the step before generate_template and inserting a dormant `{{if .existing_component.field_names}}` prompt rule: reuse these exact names, MAY add, MUST NOT rename/drop. (3) `F1prompt2_pin_function.sql` pins `{{.existing_component.function}}` so the store matches the same row (the F4 mitigation). Option A (pre-generation lookup by section_type) chosen after `\d content_components` confirmed section_type is queryable.
- **sources:** F1prompt_component_creator_preserve_field_names(1).sql; NOTES(43).md §9c–§9f, §9m; RUNBOOK(49).md Part A
- **relations:** F1 guard (backstop); F4 (pin is its mitigation); prompt-migration convention; deploy-ordering gate (its 9i failure).
- **verify-later:** load_existing_component_action.go + registry.go (IsLocal:true); component-creator default_config (top-level prompt_template; load_existing_component step; input_fields incl. existing_component).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Store-driven retry on field-drift rejection (Option B)
- **category:** NEW:component-lifecycle
- **status-signal:** abandoned
- **status-evidence:** "Two options… user chose A" (NOTES §9d–§9e); Option B never mentioned again after 9e.
- **what:** The rejected alternative to Option A: on a field-drift rejection the guard would return the existing field names, and a store_component error edge would loop back to generate_template with the names injected, retrying once. Judged heavier (reject-with-retry-data + loop-guarded workflow edge) but authoritative (no key guessing); dropped when section_type proved a stable pre-generation lookup key.
- **sources:** NOTES(43).md §9d, §9e
- **relations:** F1-prompt Option A (superseding choice).
- **verify-later:** absence: component-creator workflow should have no store_component→generate_template error edge.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F4 — regen-vs-create keyed on the LLM-chosen function (silent fork)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "F4 (structural finding): regen-vs-create is keyed on the LLM-chosen function → nondeterministic. Miss case = silent FORK" (NOTES §9m); pin migration applied §9n; "store-side advisory FLAGGED as follow-up, not built".
- **what:** Whether a store is a regeneration or a creation depends on whether the LLM happened to choose the existing `function` name — a miss silently creates a parallel active duplicate for a section_type (library fragmentation; guard bypassed by fork; selector nondeterminism). Observed live in F2 testing (fork 80222fc1). Mitigation shipped: prompt pin of `{{.existing_component.function}}`; store-side advisory (warn when function misses but an active same-section_type row exists) deliberately advisory-only and unbuilt, since multiple components per section_type can be legitimate. A suspected live-fork case (duplicate hero rows) was later softened to old manual seeding.
- **sources:** NOTES(43).md §9m, §9n, §9al, §9am; RUNBOOK(49).md Part E
- **relations:** F1-prompt function pin; StoreGeneratedComponentAction lookup; F2 methodology (exposed it).
- **verify-later:** duplicate non-forked function rows in content_components; whether any store-side advisory exists.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F5 — regen-added required fallback-less fields strand renderability
- **category:** NEW:component-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Flagged F5: extend the F1 guard… Not built now" (NOTES §9v); still an open Part E flag in RUNBOOK(49).
- **what:** The incident's second facet: the 15:06 regen also ADDED `cta_url` (required, Tier-C source, no fallback) that no affected site's specs could satisfy — renames strand stored content, required additions strand renderability (sections permanently not-ready → carried forever). Proposed guard extension: reject, or force optional/fallbacked, any added required field on a regeneration.
- **sources:** NOTES(43).md §9v; RUNBOOK(49).md Part E F5; HANDOFF(7).md §Flags
- **relations:** F1 guard; section readiness model; carry-forward path.
- **verify-later:** store_generated_component_action.go — absence of an added-required-field check.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F7 — unguarded template swap in update_component_html (residual)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "the snapshot INSERT is ALREADY FIXED in current code… Residual: no placeholder⇄schema sync validation on template swaps" (NOTES §9ak); Part E flag open.
- **what:** `update_component_html` swaps a shared component's template (snapshotting versions — its old silent snapshot failure on the removed `version_note` column is fixed) but performs no placeholder⇄schema agreement validation and no field-contract guard, leaving a second, unguarded write path to shared components. The original F7 framing (an unversioned live swap on hero) was investigated and softened; the residual is the missing validation.
- **sources:** NOTES(43).md §9aj, §9ak, §9am; RUNBOOK(49).md Part E F7; HANDOFF(7).md §Flags
- **relations:** F1 guard (candidate extension target); component versioning.
- **verify-later:** update_component_html_action.go — snapshot INSERT columns; absence of schema-sync validation.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Component versioning via component_versions (and unversioned-write provenance)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "Component updated AGAIN 2026-07-03 13:22:44 with NO v2 row… unversioned write path provenance OPEN" (NOTES §9an); manual snapshots v2/v3 taken to backfill.
- **what:** `component_versions` snapshots (component_id, version_number, schema, template, change_description, changed_by, change_source) are the change history for shared components; `change_source` records the triggering work item's source (useful provenance). Coverage is incomplete: some write paths historically failed silently or bypass versioning entirely, so manual mirror-the-working-insert snapshots are the established compensation before risky writes, and zero-version updates are treated as an investigation smell.
- **sources:** NOTES(43).md §9k, §9an, §9ao, §9bd; RUNBOOK(49).md Part C Step 6
- **relations:** F7; snapshot-before-change conventions; F8 remediation (v2/v3 snapshots).
- **verify-later:** component_versions rows for fdd92ad4 and brief-explanation; snapshotComponentVersion call sites.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### llm_guidance as a per-field generation-steering surface
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Writer prompt renders per-field guidance ⇒ every writer pass on any site was instructed to write vonc's product" (NOTES §9ba); writer config read §9az.
- **what:** Each `input_schema.fields` entry may carry `llm_guidance`, which page-content-writer renders into its generate_content prompt as the field's instruction (alongside name/type/required/description; fallback values notably never enter the prompt). On a shared component this is the highest-leverage contamination/steering surface — it shapes all future content on every consuming site — and therefore must be site-neutral while preserving structural guidance (word counts, `<em>` accent rule).
- **sources:** NOTES(43).md §9az–§9bb; RUNBOOK(49).md Part C Step 7 (the 11 neutral strings)
- **relations:** F8 carrier 3; page-content-writer prompt assembly; F8 lint scope.
- **verify-later:** page-content-writer default_config generate_content prompt (llm_field_specs block); brief-explanation field attrs.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Superseded hypothesis: update_component_html re-renders dependents inline
- **category:** NEW:component-lifecycle
- **status-signal:** superseded
- **status-evidence:** Original BUNDLE base: "the regeneration then re-renders every dependent page… rewritten together at 15:06:12.956, roughly 16ms after the component's own update"; Correction 1: "update_component_html is clean… the synchronized timestamp is just the pending-flag, not a render".
- **what:** The initial working theory held that update_component_html performed an inline dependent re-render (inferred from the ~16ms synchronized timestamps) and was the clobber writer. Disproved by reading the action: it only snapshots, swaps, and flags pending; the blame moved through RenderComponentAction and component-creator's workflow (both cleared) to StoreGeneratedComponentAction. Worth keeping as the exemplar of the thread's core epistemics: seven early inferences were each corrected against code/data before any fix shipped ("distrust each early inference until verified").
- **sources:** BUNDLE_component_regen_clobber.md §1 (base version); NOTES(43).md §2 Corrections 1–7, §3
- **relations:** clobber failure mode; StoreGeneratedComponentAction (actual writer).
- **verify-later:** n/a (historical).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Shared content-component reuse model (one content_components row, N page_components instances)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** Stated as settled platform fact throughout ("one `content_components` row, N `page_components` instances", BUNDLE(3) §1); vonc.com/index arrived mid-recovery as "healthy sixth dependent" proving live reuse.
- **what:** A section component is a single shared `content_components` row (keyed by `function`, with `section_type`, `input_schema.fields`, `html_template`, `is_active`, `forked_from`) reused across pages and sites; each page stores its own `content_data` in `page_components`. Any change to the shared row therefore has a cross-site blast radius — the structural precondition for both incidents in this thread.
- **sources:** BUNDLE(3).md §1; NOTES(43).md §1, §9z (vonc sixth dependent); HANDOFF(7).md §Platform operating model
- **relations:** clobber failure mode; F4 fork-vs-match; F8 contamination; optimistic-lock co-management.
- **verify-later:** content_components + page_components schemas; idx_cc_selector (section_type, component_level) partial index.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### StoreGeneratedComponentAction regeneration branch
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Correction 6 — StoreGeneratedComponentAction is the rename writer" (NOTES §2); insert/dedup shape read from the live file (NOTES §9q).
- **what:** On storing a generated component whose `function` matches an existing row (`WHERE function=$1 AND forked_from IS NULL`, deliberately is_active-agnostic since 2026-05-06), the action snapshots the old schema/template to `component_versions`, UPDATEs the shared row in place (same component_id, so dependents follow), marks dependents pending (`markPagesPendingRebuild` — build_status only, no rendered_html), and raises one deduped `needs_rerender` work item per affected site via `createRerenderWorkItem`. Pre-fix, that item carried no `reason`, making the triggered re-render assemble-only and unable to repair anything.
- **sources:** NOTES(43).md §2 Correction 6, §9h, §9q; BUNDLE(3).md §3
- **relations:** F1 guard lives in its validation block; F3b re-added the reason to its spec; F4 (its function-keyed lookup is the fork vector).
- **verify-later:** store_generated_component_action.go (existence check L198–207; regen branch; markPagesPendingRebuild; createRerenderWorkItem NOT EXISTS dedup).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F1 field-contract guard (reject regens that rename/drop retained fields)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "F2 COMPLETE… 3b reject ✔ (live firing, three-level visibility, zero mutation)" (NOTES §9o, 2026-07-02 16:39); RUNBOOK(49) Part A "Fixes deployed + proven".
- **what:** In StoreGeneratedComponentAction's Layer-1 validation, on `isRegeneration` the guard diffs old vs new `input_schema.fields` names (`schemaFieldSet` helper); any retained field that disappears becomes a blockingIssue routed through `recordValidationRejection` into `agent_error_log` — additions allowed, renames/drops rejected before any snapshot/UPDATE. Converts silent stranding into a loud, queryable rejection naming the stranded fields. Design choice: preserve-the-contract strict-reject backstop, not a per-dependent migration.
- **sources:** F1_store_generated_component_action.patch; NOTES(43).md §9, §9a, §9o; RUNBOOK(49).md Part A
- **relations:** F1-prompt (generation-time complement so name-preserving regens succeed); F5 (proposed extension); F8 (guard checks names only — its blind spot).
- **verify-later:** store_generated_component_action.go guard block + schemaFieldSet; agent_error_log rows error_code component_validation_rejected; store_generated_component_guard_test.go.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F1-prompt generation-time field-name preservation (loader + dormant rule + function pin)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "F1-prompt 1a/1b+2 done (validator passes; loader executes; verify row good)" (NOTES §9k); Tier 3a two regens preserved names with md5-verified template change (§9l).
- **what:** Three coupled pieces so regens preserve names instead of being rejected: (1) `load_existing_component` Go action — looks up the canonical component by `section_type` (is_active, forked_from IS NULL, component_level='section'), outputs `existing_component.field_names` + `function`; advisory, never errors (no match → empty map → rule dormant). (2) A snapshot-first, anchored, idempotent, drift-checked SQL migration wiring the step before generate_template and inserting a dormant `{{if .existing_component.field_names}}` prompt rule: reuse these exact names, MAY add, MUST NOT rename/drop. (3) `F1prompt2_pin_function.sql` pins `{{.existing_component.function}}` so the store matches the same row (the F4 mitigation). Option A (pre-generation lookup by section_type) chosen after `\d content_components` confirmed section_type is queryable.
- **sources:** F1prompt_component_creator_preserve_field_names(1).sql; NOTES(43).md §9c–§9f, §9m; RUNBOOK(49).md Part A
- **relations:** F1 guard (backstop); F4 (pin is its mitigation); prompt-migration convention; deploy-ordering gate (its 9i failure).
- **verify-later:** load_existing_component_action.go + registry.go (IsLocal:true); component-creator default_config (top-level prompt_template; load_existing_component step; input_fields incl. existing_component).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Store-driven retry on field-drift rejection (Option B)
- **category:** NEW:component-lifecycle
- **status-signal:** abandoned
- **status-evidence:** "Two options… user chose A" (NOTES §9d–§9e); Option B never mentioned again after 9e.
- **what:** The rejected alternative to Option A: on a field-drift rejection the guard would return the existing field names, and a store_component error edge would loop back to generate_template with the names injected, retrying once. Judged heavier (reject-with-retry-data + loop-guarded workflow edge) but authoritative (no key guessing); dropped when section_type proved a stable pre-generation lookup key.
- **sources:** NOTES(43).md §9d, §9e
- **relations:** F1-prompt Option A (superseding choice).
- **verify-later:** absence: component-creator workflow should have no store_component→generate_template error edge.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F4 — regen-vs-create keyed on the LLM-chosen function (silent fork)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "F4 (structural finding): regen-vs-create is keyed on the LLM-chosen function → nondeterministic. Miss case = silent FORK" (NOTES §9m); pin migration applied §9n; "store-side advisory FLAGGED as follow-up, not built".
- **what:** Whether a store is a regeneration or a creation depends on whether the LLM happened to choose the existing `function` name — a miss silently creates a parallel active duplicate for a section_type (library fragmentation; guard bypassed by fork; selector nondeterminism). Observed live in F2 testing (fork 80222fc1). Mitigation shipped: prompt pin of `{{.existing_component.function}}`; store-side advisory (warn when function misses but an active same-section_type row exists) deliberately advisory-only and unbuilt, since multiple components per section_type can be legitimate. A suspected live-fork case (duplicate hero rows) was later softened to old manual seeding.
- **sources:** NOTES(43).md §9m, §9n, §9al, §9am; RUNBOOK(49).md Part E
- **relations:** F1-prompt function pin; StoreGeneratedComponentAction lookup; F2 methodology (exposed it).
- **verify-later:** duplicate non-forked function rows in content_components; whether any store-side advisory exists.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F5 — regen-added required fallback-less fields strand renderability
- **category:** NEW:component-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Flagged F5: extend the F1 guard… Not built now" (NOTES §9v); still an open Part E flag in RUNBOOK(49).
- **what:** The incident's second facet: the 15:06 regen also ADDED `cta_url` (required, Tier-C source, no fallback) that no affected site's specs could satisfy — renames strand stored content, required additions strand renderability (sections permanently not-ready → carried forever). Proposed guard extension: reject, or force optional/fallbacked, any added required field on a regeneration.
- **sources:** NOTES(43).md §9v; RUNBOOK(49).md Part E F5; HANDOFF(7).md §Flags
- **relations:** F1 guard; section readiness model; carry-forward path.
- **verify-later:** store_generated_component_action.go — absence of an added-required-field check.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F7 — unguarded template swap in update_component_html (residual)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "the snapshot INSERT is ALREADY FIXED in current code… Residual: no placeholder⇄schema sync validation on template swaps" (NOTES §9ak); Part E flag open.
- **what:** `update_component_html` swaps a shared component's template (snapshotting versions — its old silent snapshot failure on the removed `version_note` column is fixed) but performs no placeholder⇄schema agreement validation and no field-contract guard, leaving a second, unguarded write path to shared components. The original F7 framing (an unversioned live swap on hero) was investigated and softened; the residual is the missing validation.
- **sources:** NOTES(43).md §9aj, §9ak, §9am; RUNBOOK(49).md Part E F7; HANDOFF(7).md §Flags
- **relations:** F1 guard (candidate extension target); component versioning.
- **verify-later:** update_component_html_action.go — snapshot INSERT columns; absence of schema-sync validation.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Component versioning via component_versions (and unversioned-write provenance)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "Component updated AGAIN 2026-07-03 13:22:44 with NO v2 row… unversioned write path provenance OPEN" (NOTES §9an); manual snapshots v2/v3 taken to backfill.
- **what:** `component_versions` snapshots (component_id, version_number, schema, template, change_description, changed_by, change_source) are the change history for shared components; `change_source` records the triggering work item's source (useful provenance). Coverage is incomplete: some write paths historically failed silently or bypass versioning entirely, so manual mirror-the-working-insert snapshots are the established compensation before risky writes, and zero-version updates are treated as an investigation smell.
- **sources:** NOTES(43).md §9k, §9an, §9ao, §9bd; RUNBOOK(49).md Part C Step 6
- **relations:** F7; snapshot-before-change conventions; F8 remediation (v2/v3 snapshots).
- **verify-later:** component_versions rows for fdd92ad4 and brief-explanation; snapshotComponentVersion call sites.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### llm_guidance as a per-field generation-steering surface
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Writer prompt renders per-field guidance ⇒ every writer pass on any site was instructed to write vonc's product" (NOTES §9ba); writer config read §9az.
- **what:** Each `input_schema.fields` entry may carry `llm_guidance`, which page-content-writer renders into its generate_content prompt as the field's instruction (alongside name/type/required/description; fallback values notably never enter the prompt). On a shared component this is the highest-leverage contamination/steering surface — it shapes all future content on every consuming site — and therefore must be site-neutral while preserving structural guidance (word counts, `<em>` accent rule).
- **sources:** NOTES(43).md §9az–§9bb; RUNBOOK(49).md Part C Step 7 (the 11 neutral strings)
- **relations:** F8 carrier 3; page-content-writer prompt assembly; F8 lint scope.
- **verify-later:** page-content-writer default_config generate_content prompt (llm_field_specs block); brief-explanation field attrs.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Superseded hypothesis: update_component_html re-renders dependents inline
- **category:** NEW:component-lifecycle
- **status-signal:** superseded
- **status-evidence:** Original BUNDLE base: "the regeneration then re-renders every dependent page… rewritten together at 15:06:12.956, roughly 16ms after the component's own update"; Correction 1: "update_component_html is clean… the synchronized timestamp is just the pending-flag, not a render".
- **what:** The initial working theory held that update_component_html performed an inline dependent re-render (inferred from the ~16ms synchronized timestamps) and was the clobber writer. Disproved by reading the action: it only snapshots, swaps, and flags pending; the blame moved through RenderComponentAction and component-creator's workflow (both cleared) to StoreGeneratedComponentAction. Worth keeping as the exemplar of the thread's core epistemics: seven early inferences were each corrected against code/data before any fix shipped ("distrust each early inference until verified").
- **sources:** BUNDLE_component_regen_clobber.md §1 (base version); NOTES(43).md §2 Corrections 1–7, §3
- **relations:** clobber failure mode; StoreGeneratedComponentAction (actual writer).
- **verify-later:** n/a (historical).

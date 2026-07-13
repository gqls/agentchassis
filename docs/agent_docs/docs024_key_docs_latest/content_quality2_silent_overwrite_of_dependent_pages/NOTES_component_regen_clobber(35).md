# NOTES — component regeneration silently empties dependent pages (full investigation log)

Running notes for the whole investigation of the cross-site platform bug, including the choices, the false
leads we corrected, and the decisions. This thread grew out of the gamesdesign "silent no-rebuild" closeout
(documented separately in `NOTES_gamesdesign_silent_norebuild.md`); the system-stats band that was dropped from
gamesdesign turned out to be one instance of this wider, shared-component bug, which we then chased down here.

Companion docs (all in this folder): `BUNDLE_`, `HANDOFF_`, `PLAN_`, `RUNBOOK_component_regen_clobber.md`.

---

## 1. The bug (final, confirmed)

A shared content component (`content_components` row) is reused by many pages across multiple sites; each page
stores its own `content_data`, and at render time `RenderTemplate` (component_library.go) binds that `content_data`
into the component's `html_template` by **exact field name** (Go `text/template`), silently stripping any
unmatched `{{.field}}` placeholder to empty (it replaces the `<no value>` tokens, logs a warning, raises no error).
On 2026-06-24 15:06, component-creator regenerated the `system-stats` component (`fdd92ad4`); its
`StoreGeneratedComponentAction` store step found the existing shared row and took its regeneration branch —
snapshotting the old schema/template to `component_versions`, then overwriting `html_template` + `input_schema`
**in place** with a renamed field set (`eyebrow`→`eyebrow_label`, `heading`→`section_headline`,
`stat_1_number`→`stat1_value`, … plus `stat_N_icon` dropped and `cta_url`/`cta_label` added). It did **not**
migrate the dependent pages' `content_data`, which still carries the old field names. Dependents were then
re-rendered against the new template with their old keys, so every field missed and the sections rendered as an
empty shell; the page assembler's visible-content filter drops a content-empty band, silently, on every page that
shares the component. Five live instances on three other sites are affected; gamesdesign was a sixth, since
dropped. The `content_data` is intact and per-page, so the breakage is recoverable.

---

## 2. Investigation timeline (with the decisions and the corrections)

The defining feature of this investigation was repeatedly **declining to treat an early inference as settled** and
checking it against the actual code and data — which corrected the story several times.

**Start — packaging the bug (from the gamesdesign closeout).** We had flagged this as a cross-site platform bug
during the gamesdesign work. Decision: produce a `cmd/bundle` invocation + cold-start docs (BUNDLE/HANDOFF/PLAN/
RUNBOOK) so a fresh chat could pick it up. Initial working hypothesis (carried from gamesdesign): "component-creator
regen renames fields and re-renders dependents from un-migrated `content_data`." Scopes were guessed from a prior
bundle's signatures; runtime page left as a placeholder.

**Correction 1 — `update_component_html` is clean.** Reading the uploaded `update_component_html_action.go`: it
snapshots the old template, swaps in the new one, and marks dependents `build_status='pending'` — and explicitly
does **not** touch `rendered_html`. Its bulk `UPDATE … SET updated_at=NOW()` is what stamped all five rows at the
identical `15:06:12.956`. So it is not the clobber, and the synchronized timestamp is just the pending-flag, not a
render. (It also revealed the old snapshot INSERT targeted a `version_note` column that no longer exists, so those
snapshots were silently failing.)

**Correction 2 — the silent-empty mechanism is in `RenderTemplate`.** component_library.go runs Go `text/template`
and then strips the `<no value>` tokens that unmatched placeholders produce down to empty string, logging only a
warning. So a renamed field → empty fill, no error. This is *why* the failure is silent.

**Correction 3 — `rerender_page_sections` is partly protective, not the culprit.** It carries the stored HTML when
a section isn't "ready" and is gated to `reason ∈ {image_landed, section_data_resolved}` — so it is not the obvious
writer of the blanks.

**Correction 4 — component-creator does not re-render dependents.** Its `default_config` (v1.0.1080) workflow is
`ensure_site_record → read_site_spec → generate_template → store_component → complete`. No `update_component_html`
step, no dependent re-render. So component-creator-as-workflow is not the re-renderer.

**Correction 5 — `RenderComponentAction` is the writer's render and lives in `v3_site_actions.go`.** (My bundle had
guessed a `render_component_action.go` path; fixed.) It renders fresh content (`content_field` ⊕ `merge_with`) and
**returns** a map — it does not write the DB or set `updated_at`. So it's not the stored-content re-render path.
(`component_selector`'s only `content_components` write is `IncrementUsageCount`, and `fdd92ad4`'s count is 0, so it
hadn't run — not a rename.)

**Data check — the five are not zero-length.** The picker showed `html_len = 7369`, identical across all five, on
three sites (leopardessconsulting.co.uk, robot-hands.com, ai-agent-orchestration.com). Decision: do **not** assume
"empty" from a non-zero length; the identical length across different sites is the real tell (per-page content not
landing), but confirm on the bytes.

**Correction 6 — `StoreGeneratedComponentAction` is the rename writer.** The uploaded
`store_generated_component_action.go`: on regeneration (an existing component with the same `function`), it
snapshots the old schema/template, `UPDATE`s `content_components` **in place** (same `component_id`, so dependents
follow), marks dependents pending (`markPagesPendingRebuild`, which only sets `build_status`/`updated_at` — no
`rendered_html`), and raises one `needs_rerender` work item per site. Crucially that `needs_rerender` carries spec
`{component_id, function, refresh_site_components:false}` with **no `reason`**, routed to `rerender-pages` → so the
re-render it triggers is **assemble-only** (re-ships existing `rendered_html`, doesn't re-render sections). It does
**not** migrate dependents' `content_data`.

**Correction 7 (my mistake, then corrected) — "predates 15:06 / latent" was wrong.** From the assemble-only finding
I inferred the empty render must predate the regen and the rows might just hold a preserved good render (latent
breakage). The R3a version-schema check disproved the "predates" framing, and the content-presence check disproved
"latent" (see below). Logged here precisely because it's the kind of plausible-but-unverified inference the standing
rules say to distrust.

**R3a — the rename is the root, confirmed.** The pre-15:06 `component_versions` v1 schema
(`input_schema->'fields'`) is the **old** key set and matches the dependents' `content_data` field-for-field (24
fields, down to `stat_N_icon`); the live schema is the **new** set (22 fields). So the 15:06 regeneration renamed
old→new without migrating `content_data`. Not an out-of-band edit, not a standing writer-vs-schema mismatch.

**Final check — broken now, not latent.** content-presence query: `content_data` is **per-page** (distinct
headings/stats — `70`/"Deployed Agents" on the orchestration and consulting sites, `2400`/"Gripper Models Indexed"
on robot-hands), yet the gripper-detail render contains **neither** its heading nor its `stat_1_number`
(`render_has_*` both false) and all five renders are byte-identical. So the rows hold the **new-template render
with empty content slots** — the five are **live-broken**, not latent. (The exact write that produced the empty
render is not pinned to a timestamp — the `.956` stamp is the pending-flag — but the render is definitively the
new-template empty render, so the conclusion holds regardless of that minor unknown.)

---

## 3. Decisions made

- **Build cold-start docs (bundle + handoff + plan + runbook)** rather than fix inline, because the bug is shared
  platform machinery, the regen was a manual/out-of-band trigger, and gamesdesign itself wasn't affected.
- **Do not blindly write to the shared component or its dependents** — another chat co-manages components; the
  runbook gates every write behind a freshness check.
- **Distrust each early inference until verified** — which corrected the story seven times above and avoided
  shipping a fix against the wrong cause (e.g. patching `update_component_html`, which is clean).
- **Phrase fix options as routes, not a prescribed either/or** (per the original request) until the root was
  confirmed; now that it is, the runbook gives a concrete recommended fix with the insertion point.
- **Recovery must force a section re-render** (`reason=section_data_resolved`); the regen's own assemble-only
  `needs_rerender` cannot repair a content-key mismatch.

---

## 4. Confirmed root cause

Regenerating a **shared** component overwrites its `input_schema`/`html_template` field contract in place
(`StoreGeneratedComponentAction` regen branch) **without migrating the dependent pages' `content_data`** to the new
field names, and the re-render it triggers is assemble-only so it cannot repair the sections. Because rendering
binds by exact field name and silently empties misses, every dependent then renders a content-free shell and is
dropped by the assembler — silently, fanning out across every page sharing the component.

---

## 5. The fix (recommended) + alternatives

Insertion point: `StoreGeneratedComponentAction`'s regeneration branch (it already holds `existingSchema` and the
new `inputSchemaJSON`, ~L360–398). The requirement: a regeneration of an existing shared component must never leave
dependents bound to field names their `content_data` doesn't have. Routes (combine as judged):

- **Preserve retained field names.** Feed the existing `input_schema` field names into the regeneration prompt and
  reject/conform a regenerated schema that *renames* a retained field. Adding new fields (`cta_*`) and dropping
  unused ones is fine; gratuitous renames (`eyebrow`→`eyebrow_label`, `number`→`value`, `stat_N_`→`statN`) are the
  damage. Cleanest because it needs no per-dependent migration.
- **Migrate dependents on rename.** If renames are ever legitimately wanted, the regen must emit an explicit
  old→new field map, and `StoreGeneratedComponentAction` migrates every dependent's `content_data` keys before the
  re-render. Handle non-1:1 cases (dropped/added fields) explicitly.
- **Make the triggered re-render actually re-render sections.** `createRerenderWorkItem` should set
  `reason=section_data_resolved` (or the `page_rerender` it leads to must), so dependents regenerate `rendered_html`
  from the (preserved/migrated) `content_data` instead of an assemble-only re-ship.
- **Fail-loud guard (defensive).** Treat a re-render that empties a previously-populated section as a failure to
  surface (carry the stored HTML / mark the item failed) rather than a silent blank to ship.

Keep the change in Go; reuse `RenderTemplate`/`GetComponentByID` and the existing rerender chain; back up any agent
whose config/image changes.

---

## 6. Recovery of the affected pages (now confirmed broken)

`content_data` is intact and per-page, so no LLM is needed. Either re-key each affected page's `content_data`
old→new (using the field map) **then** trigger a `page_rerender` carrying `reason=section_data_resolved` (a bare
`needs_rerender` is assemble-only and just re-ships the empty shell); or run a full writer rebuild (`needs_page`),
which regenerates `content_data` under the current (new) field names and renders correctly. Verify each page's
`rendered_html` then contains its content (`render_has_*` true) and the band appears on the live deploy. Apply to
the five instances (and gamesdesign's if it is rebuilt).

---

## 7. Current state / next steps

Diagnosis complete and root confirmed. The five instances are live-broken (empty system-stats band). Next: (1)
implement the fix in `StoreGeneratedComponentAction`'s regen branch (preferred: preserve retained field names +
make the triggered re-render a section re-render; add the guard); (2) test it on a throwaway component; (3) recover
the five (+ gamesdesign if rebuilt) via re-key + `section_data_resolved` re-render or a full rebuild; (4) verify.
The runbook's "Where we are" section and Steps F1–F2 (fix + test) and R5–R6 (recover + verify) carry the detail.

---

## 8. Key IDs / facts

- Component: `content_components` `fdd92ad4-521a-4602-89cf-7ee1a66c10f1` (`system-stats`); 22 live fields,
  `schema_template_synced=t`; `usage_count` stale (0). `input_schema` nests field names under `input_schema->'fields'`.
- `component_versions`: one row — v1, `change_source='manual'`, `changed_by='component-creator:regen'`,
  created 2026-06-24 15:06:12.92; its `input_schema->'fields'` = the OLD key set (the pre-regen contract).
- Old keys (in dependents' `content_data`): `eyebrow, heading, subheading, footnote,
  stat_N_number/stat_N_icon/stat_N_label/stat_N_suffix/stat_N_description` (24).
- New keys (live template/schema): `eyebrow_label, section_headline, section_intro, footnote_text, cta_url,
  cta_label, statN_value/statN_suffix/statN_label/statN_description` (22; `stat_N_icon` dropped, `cta_*` added).
- Five affected instances (`component_id=fdd92ad4`, `build_status=pending`, `component_version_id` NULL,
  `rendered_html` 7369 bytes byte-identical, `updated_at` 15:06:12.956367, content empty): `index` on
  leopardessconsulting.co.uk, robot-hands.com, ai-agent-orchestration.com; `case-study-kafka-consumer-group-remediation`
  on ai-agent-orchestration.com; `gripper-detail` on robot-hands.com. gamesdesign's instance was dropped.
- Code: `RenderTemplate` + the `<no value>`→empty cleanup, `RenderContext.ContentData`, `GetComponentByID` —
  component_library.go. `UpdateComponentHTMLAction` — update_component_html_action.go (clean; pending-flag only).
  `RerenderPageSectionsAction` — rerender_page_sections_action.go (no-LLM; carry-forward; reason-gated).
  `RenderComponentAction` — v3_site_actions.go (writer's render; returns a map, no DB write).
  `StoreGeneratedComponentAction` (+ `markPagesPendingRebuild`, `createRerenderWorkItem`, `snapshotComponentVersion`)
  — store_generated_component_action.go (the rename writer; regen branch ~L354–432).
- DB access: `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db`.
- Concurrency: another chat co-manages components — re-check freshness before any write.

---

## 9. F1 implementation status (2026-06-26)

F1 patch written and `gofmt -e` clean: `F1_store_generated_component_action.patch` (full edited file also staged as
`store_generated_component_action.go`). Four hunks against
`platform/orchestration/actions/store_generated_component_action.go`:
- add `sort` import;
- **field-contract guard** in the Layer-1 validation block (after the `<no value>` check): on `isRegeneration`,
  diff `existingSchema` vs `schemaJSONStr` field names; any retained field that disappears → a `blockingIssues`
  entry → rejected via the existing `recordValidationRejection` path (logged to `agent_error_log`). Additions are
  allowed. Converts silent stranding into a loud, queryable rejection naming the fields;
- new helper `schemaFieldSet` (mirrors `deriveRenderMode`'s `{fields:…}` parse), appended at EOF with its own doc
  comment (kept out of `deriveRenderMode`'s comment);
- `createRerenderWorkItem` spec now carries `"reason":"section_data_resolved"`.

Design choice: route 1 (preserve-the-contract / strict reject) as the in-file backstop, NOT the migration route —
matches the recommendation. No existing variable names changed; only additions (`sort` import, `schemaFieldSet`
func, the guard block, the spec literal). Reuses `blockingIssues`/`recordValidationRejection`/`existingSchema`/
`schemaJSONStr`/`isRegeneration` already in scope.

Still to do for F1 to fully work day-to-day (not in this file):
- **companion prompt change** in component-creator's `generate_template` step — pass existing field names and
  require their reuse, so names-preserving regens succeed rather than being rejected;
- **verify `rerender-pages` propagates `spec.reason`** to the per-page `page_rerender` items (its
  `create_rerender_items` step) so the new `section_data_resolved` reason actually drives a section re-render;
  one-line companion change there if it currently drops the reason.

Next: `git apply F1_store_generated_component_action.patch` → `go build ./...` → deploy (Actions → Backblaze →
image bump) → F2 test on a throwaway component → then recover the five (R4/R5/R6). The guard will (intentionally)
reject the original `fdd92ad4`-style regen; the companion prompt change is what lets a names-preserving regen land.

### 9a. Update — rerender-pages checked; Change 2 pulled; patch is guard-only

Dumped `rerender-pages` default_config (v1.0.1083). Its `create_rerender_items` step wires only
`{domain, site_id, pages_field}` — it does **not** pass `reason`. In this thin-workflow design the step config is
how an action's inputs are fed, so a `reason` in the inbound `needs_rerender` spec is **dropped here**: the
per-page `page_rerender` items it creates don't carry it, and page-rerender falls through to assemble-only. So
Change 2 (`reason=section_data_resolved` in `createRerenderWorkItem`'s spec) was **inert end-to-end** and would
also over-scope (a site-wide `section_data_resolved` re-renders every page's sections, not just the dependents).

Decision: **reverted Change 2** out of the patch. `F1_store_generated_component_action.patch` is now **guard-only**
(3 hunks: `sort` import, the field-contract guard, the `schemaFieldSet` helper at EOF; `gofmt -e` clean,
`section_data_resolved` count = 0). The guard is the complete, correct, independently-shippable fix. The reason +
its propagation move to **F3**, to land together once we have the `create_rerender_items` action source (workflow
change `"reason":"input_data.spec.reason"` if that action reads a `reason` config key, else a small Go change),
and after confirming it won't over-scope.

R5 recovery reconciled with this finding: it must **not** route the section re-render through `rerender-pages`
(reason dropped). Route A = full `needs_page` rebuild per page (writer regenerates `content_data` under the new
schema → renders correctly; simplest). Route B = re-key `content_data` old→new + a **directly-inserted**
`page_rerender` per page (valid `spec.page_id`, `reason=section_data_resolved`, non-colliding `item_key`).

Next: ship the guard (apply patch → `go build ./...` → deploy → image bump) + do the component-creator
`generate_template` prompt companion; then F2; get the `create_rerender_items` source for F3; then recover the five.

### 9b. F3 designed from the create_rerender_items source (scoped reason propagation)

Got `create_rerender_items_action.go`. It builds each `page_rerender` spec as `{page_id, page_name, filename,
domain}` (no reason) and creates one item per page for **every** page `get_pages_for_rerender` returns (whole
site). So propagating a bare reason would re-render the whole site's sections and could surface *other* latent
mismatches. F3 therefore both propagates the reason and **scopes** to the changed component's dependents. Three
coupled changes, deploy after F1:
- **F3a** `create_rerender_items_action.go` (`F3a_create_rerender_items_action.patch`, gofmt-clean, 3 hunks): add
  `reason`+`component_id` to InputSpec; when reason ∈ {section_data_resolved, image_landed} AND component_id set →
  query the component's dependent pages (`page_components` JOIN `pages` on site), create items only for those, and
  stamp `spec.reason`. No signals → unchanged (assemble-only, all pages). Inlined the reason check (no new package
  symbol, avoids collision). rows.Close() before the INSERT loop (single-conn safety).
- **F3b** `store_generated_component_action.go` (`F3b_store_reason_readd.patch`, applies on top of the deployed
  guard, 1 hunk): re-add `"reason":"section_data_resolved"` to `createRerenderWorkItem`'s spec (component_id already
  present → scopes it).
- **F3c** `rerender-pages` default_config (jsonb, no redeploy): add `"reason":"input_data.spec.reason"` and
  `"component_id":"input_data.spec.component_id"` to the `create_rerender_items` step config (jsonb_set SQL in the
  runbook; snapshot_agent first).
Order: deploy F3a+F3b (Go image bump), then F3c. Either half alone is safe (missing half → current assemble-only);
both needed to activate. Residual scope: rerender_page_sections re-renders all sections of each dependent page (not
just the one component's) — blast radius = dependent pages only, accepted section-rerender behaviour. R5 recovery
note: post-F3, a `needs_rerender` with reason+component_id also drives the scoped section re-render (so re-key +
needs_rerender works, not only the direct page_rerender insert); re-key of content_data is still required either way.

### 9c. generate_template prompt fix — design + prerequisite (from project knowledge)

Prompt mechanics (from 016 debugging guide + 019 tool library): prompts live inline at
`default_config->'workflow'->'steps'->'<step>'->'config'->>'prompt_template'`. Convention for prompt edits:
`snapshot_agent` first → anchored, idempotent `replace()` on the prompt text that aborts on drift → UPDATE filtered
to the live row (`is_active=true AND (is_snapshot IS NULL OR is_snapshot=false)`). Mind the "072 nested-prompt trap"
(some prompts sit nested inside loop steps, not top-level) — verify the path first. Direct precedent: the
tool-doc-header — `tool-improver` prompt already has a "preserve the header across LLM rewrites (anti-drift anchor)"
rule, backstopped by a `HasToolDocHeader` creation gate. Our guard = the gate; the prompt rule = the anchor. Same
pattern.

Key insight: the prompt fix is NOT just a text tweak. For the prompt to require reuse of existing field names it must
be GIVEN them, and component-creator generates blind — its workflow is ensure_site_record → read_site_spec →
generate_template → store_component, and the existing component is only discovered at store time
(StoreGeneratedComponentAction). So at generate_template time nothing has loaded the existing schema. Two coupled
parts:
1. Load existing field names before generation — a step (or an addition to read_site_spec) that looks up any
   existing component for this function/section_type and exposes its input_schema field names as template data
   (e.g. existing_field_names).
2. Prompt rule (anchored migration) — generate_template's prompt_template gains: if regenerating, reuse these exact
   field names, may ADD new fields, must NOT rename/drop existing ones.
Until both land, the guard rejects renaming regens (fail-safe). Requested the live component-creator config (workflow
shape + step list + generate_template config: prompt_template, input_fields, whether a load-existing step exists,
and whether the prompt is top-level or in the 072 trap) to write the exact edit. PENDING that dump.

### 9d. generate_template prompt fix — config seen; Part 2 (prompt rule) written; Part 1 (feed) pending

Live component-creator config (v1.0.1084) confirms:
- **prompt_template is TOP-LEVEL** (`default_config->>'prompt_template'`), NOT at
  `workflow.steps.generate_template.config.prompt_template` (that step config holds only ai_service + input_fields).
  `prompt_is_top_level=f` proved it — the 072 trap in force; the migration MUST target the top-level key.
- Workflow: ensure_site_record → read_site_spec → generate_template → store_component → complete. No load-existing
  step — generates blind. input_fields = [input_data, site_record, site_specs]. Prompt already uses
  `{{if .input_data.spec.design_direction}}`-style conditional blocks (so a dormant conditional rule fits).
- Placeholders are `{{placeholder "field_name"}}`; field names go in input_schema.fields; the LLM CHOOSES them (and
  the `function` name) in its JSON output. So the existing component is matched (by StoreGeneratedComponentAction) on
  the LLM-chosen `function`, which is only known AFTER generation.

**Part 2 (delivered): `F1prompt_component_creator_preserve_field_names.sql`** — snapshot-first, anchored (on
"== COMPONENT CONTRACT"), idempotent (bails if "== REGENERATION FIELD-NAME RULE ==" present), drift-checked (anchor
must appear exactly once), live-row-only migration that inserts a `{{if .existing_field_names}}...{{end}}` rule right
before the component contract. DORMANT until Part 1 feeds `existing_field_names` — safe render-time no-op until then.

**Part 1 (pending decision): feed `existing_field_names` to generate_template.** The blocker: the existing component
is keyed by `function`, which the LLM only picks during generation — so a pre-generation lookup needs a stable key.
Two options:
- **A — pre-generation lookup** by a stable key (section_type). Add a load_existing_component step that queries
  content_components for the existing row for this section_type, outputs its input_schema field names as a joined
  string `existing_field_names`, and add that to generate_template's input_fields. Needs content_components to be
  queryable by section_type (or a deterministic section_type→function map).
- **B — store-driven retry** (authoritative, no key guessing): on a field-drift rejection the guard returns the
  existing field names; add a store_component error edge back to generate_template with the names injected; retry
  once. Heavier (modifies the guard's reject to reject-with-retry-data + a loop-guarded workflow edge).
REQUESTED: content_components schema (\d) + a sample row's function vs the section_type that created it, to pick A
vs B and wire Part 1. Until Part 1 lands, the guard rejects renaming regens (fail-safe) and the prompt rule is dormant.

### 9e. Option A built — content_components HAS section_type

\d content_components confirms a `section_type` column (kebab-constrained) with a selector index
`idx_cc_selector (section_type, component_level) WHERE is_active AND forked_from IS NULL`. So the canonical existing
component is lookup-able by section_type (known up front from input_data.spec.section_type) — Option A is feasible;
user chose A. Built:
- **Part 1a — `load_existing_component_action.go`** (new Go action, gofmt-clean): given section_type, selects the
  canonical shared section component (`section_type=$1 AND forked_from IS NULL AND is_active AND
  component_level='section'` ORDER BY usage_count DESC, updated_at DESC LIMIT 1), extracts input_schema.fields keys,
  outputs `{field_names: "a, b, c", function, field_count}` under output_field `existing_component`. ADVISORY: never
  errors, always returns a well-formed map (no row / lookup fail → empty map → prompt rule dormant → blind
  generation; the store guard is the backstop). Helper `schemaFieldNamesSorted([]byte)`. Must be registered in
  registry.go and deployed BEFORE the migration.
- **Migration (combined Part 1b + Part 2)** — `F1prompt_component_creator_preserve_field_names.sql` rewritten: one
  snapshot, idempotent (skips if step present + rule present), drift-checked (prompt anchor "== COMPONENT CONTRACT"
  once), then atomically: inserts the `load_existing_component` step; rewires read_site_spec.next_step +
  read_site_spec.error_step + ensure_site_record.error_step → load_existing_component (so it always runs before
  generate_template, even on those error paths); appends "existing_component" to generate_template.config.input_fields;
  inserts the dormant `{{if .existing_component.field_names}}` rule before the component contract. Prompt is at the
  TOP-LEVEL key (072 trap avoided).

Deploy order: (1) deploy Part 1a Go action (registry.go + image bump); (2) apply the migration. If the migration
lands first, the workflow references an unknown action → generation fails, so order matters. After both: a regen of
an existing section loads its field names and the prompt requires their reuse; a new section → empty → dormant →
fresh generation. Optional later robustness: also pin the existing `function` (the action already returns it) so the
store matches the same row — not required for the observed bug (function was already stable at 15:06).

### 9f. Migration APPLIED; F2 sketched (three tiers)

Combined Option-A migration applied cleanly (snapshot source_version=1 id 23720180; verify row: step_present=t,
read_site_spec_next="load_existing_component", input_fields=[input_data,site_record,site_specs,existing_component],
rule_present=t, injects_names=t). User deploying the load_existing_component Go action now. F2 test plan written
(runbook Step F2 + `store_generated_component_guard_test.go`):
- Tier 1 (deterministic, no DB): unit test of schemaFieldSet + stranded-diff incl the real system-stats rename —
  rename/drop stranded, additions/identical/new-component strand nothing. gofmt-clean.
- Tier 2 (integration, test DB): drive StoreGeneratedComponentAction with a renaming generated_template → assert
  error "regeneration removes/renames", agent_error_log row, component row UNCHANGED (no component_versions row).
  Deterministic proof of the fail-safe (preferred over forcing a rename through the LLM).
- Tier 3 (end-to-end throwaway `zzz-fieldguard-test`): insert a throwaway section component (eyebrow/heading/body),
  trigger a needs_new_component work item, observe load_existing_component feeding names + the prompt carrying the
  rule + regen preserving the 3 fields + no rejection. Cleanup SQL provided. Optional dependent page to also prove
  non-stranding / F3 scoping.
Remaining after F2: recover the five (R4/R5/R6) once the fix set is deployed.

### 9g. F2 Tier 1 GREEN

`go test -run 'SchemaFieldSet|RegenFieldContractStranded' -v` — all pass (incl the real system-stats rename case).
Confirms: (a) the guard's decision logic is correct; (b) the actions package COMPILES with the F1 guard —
schemaFieldSet resolves, no symbol collision, guard build sound. Remaining F2: Tier 2 (reject-path integration —
depends on whether a DB-backed actions test harness exists; else fold the reject check into Tier 3 by forcing a
rename on the throwaway and watching agent_error_log) and Tier 3 (end-to-end throwaway). Then F3 deploy, then
recover the five (R4/R5/R6).

### 9h. F2 reject-path lever confirmed from the store code; Tier 3 finalised (keep + reject)

Store existence check (store_generated_component_action.go L198-207): `WHERE function=$1 AND forked_from IS NULL
ORDER BY is_active DESC, updated_at DESC` — explicit note that `AND is_active=true` was REMOVED 2026-05-06. So the
store matches regardless of is_active; load_existing_component requires is_active=true. That divergence is the
deterministic rename-forcing lever. recordValidationRejection confirmed to INSERT into agent_error_log
(error_message/error_code/severity/context), so the rejection is queryable there. Guard runs in the validation block
BEFORE the regen UPDATE → a rejected row is left untouched.

Package is unit-only (no DB harness), so Tier 2 is folded into Tier 3b. Finalised Tier 3 (runbook):
- 3a keep (active `zzz-fg-keep`, fields eyebrow/heading/body): load feeds names → prompt preserves → guard passes →
  regen keeps fields; assert fields preserved + no rejection.
- 3b reject (INACTIVE `zzz-fg-reject`, fields zzz_alpha/beta/gamma): load skips (inactive) → prompt dormant → LLM
  fresh names → store matches by function (is_active-agnostic) → guard strands zzz_* → REJECT; assert agent_error_log
  row naming zzz_alpha/beta/gamma + component unchanged (still inactive, same schema, 0 new component_versions).
Both rely on the LLM mirroring section_type→function (always true empirically). Cleanup SQL provided. Then F3 deploy,
then recover the five.

### 9i. F2 3a FAILED — deploy-ordering hazard hit (load_existing_component not registered/live)

Tier 3a run failed: WORKFLOW_INVALID "step 'load_existing_component' with action 'load_existing_component' requires a
topic"; validator logged is_local=false, has_topic=false. Cause: the migration wired the load_existing_component step
(config, live immediately) but the running component-creator image doesn't have the action registered as local, so
the workflow validator rejects EVERY component-creator run — this broke ALL component generation, not just the test.
Regen did not run (0 component_versions rows; component unchanged {body,eyebrow,heading}). The work item showed
'complete' = the parent dispatch-loop orchestration finishing after handling the child's unrecoverable error; the
child failed (don't trust that status field).

Immediate mitigation: `SELECT revert_agent('component-creator');` restores the pre-migration snapshot (blind gen; F1
guard unaffected — separate Go change). Then diagnose is_local=false: (1) is the action image live on component-creator
pods (kubectl get pods -o image; image_tag bump)? (2) does registry.go have the entry WITH IsLocal:true, in the built
image? The action file alone isn't enough — the registry IsLocal:true flag is what the validator reads. Re-apply the
migration (idempotent) only AFTER confirming the action responds on all live pods, then re-run Tier 3a.

LESSON (runbook gate tightened): "deploy the Go action first" is insufficient — the migration is live instantly while
the image rolls out. Hard gate: confirm load_existing_component is registered + live on ALL component-creator pods,
THEN apply the migration. Also worth a later look (tangential): work item marked 'complete' despite child
WORKFLOW_INVALID — is the dispatch loop marking items complete regardless of child failure?

### 9j. Correction: 9i was a manual deploy-ordering slip, not a code/registration bug

User confirmed the Go action simply wasn't deployed before running Tier 3a (migration applied first by mistake). So
is_local=false was purely "action not in the running image" — no registry.go / IsLocal problem to chase. Path: deploy
the action (registry.go entry with IsLocal:true, action in the build) → confirm live on component-creator pods → the
migration (already applied, correct) then validates and runs → re-run Tier 3a. revert_agent('component-creator') only
needed if component-creator must generate during the deploy window / failed items are piling up. When live, confirm
the validator logs is_local=true for load_existing_component.

### 9k. F2 Tier 3a PASS; dead prompt block cleaned; 3b is next

Tier 3a result: TWO regens ran and both preserved the names. component_versions v1 2026-07-01 20:20:25 (first
successful regen after last night's action deploy) and v2 2026-07-02 08:03:20 (this morning's run, new work item
987828aa; watched it pass the validator and execute load_existing_component). Live fields after both regens =
exactly {body,eyebrow,heading}; updated_at 08:03:20; zero 'regeneration removes/renames' rejections. So the chain
loader→prompt rule→LLM preservation→guard pass→snapshot+overwrite works. Observation: change_source='f2-test' —
snapshotComponentVersion records the WORK ITEM's source as change_source (provenance useful).

Caveat logged: eyebrow/heading/body are LLM-guessable, so 3a alone isn't airtight proof of preservation-by-
instruction. Gap closed by (a) template-changed-while-names-held check (live md5/len vs v1's tiny manual template —
query given) and (b) Tier 3b, whose zzz_alpha/beta/gamma names are NOT guessable and whose loader path is
deliberately dormant (inactive component). 3b = first live firing of the guard — the one piece not yet observed in
prod. Expect: agent_error_log row naming zzz_*, component unchanged (inactive, same schema, updated_at untouched),
0 component_versions rows; rejected item may retry repeatedly — clean up promptly.

Prompt cleanup migration applied: dead {{if .existing_field_names}} block removed (dead_pos=0), live
{{if .existing_component.field_names}} present (t), snapshot taken first. (Dead block existed because the earlier
standalone prompt migration used the scalar name and the combined rewrite's idempotency check — step-present AND
marker-present — didn't detect the marker-only state.)

Runbook ticks: F1 deployed (in same image as the load action; guard's live firing proven by 3b), F1-prompt 1a/1b+2
done (validator passes; loader executes; verify row good), F2 Tier 1 + 3a done; 3b pending. After 3b + cleanup:
F3a/F3b/F3c deploy, then R4/R5/R6 recovery of the five.

### 9l. 3a caveat closed (md5); 3b inserted; my agent_error_log query was wrong (schema-before-SQL slip)

md5/length on zzz-fg-keep: v1=220B (manual stub) → v2=1098B → live=1177B, all different md5s — template genuinely
regenerated twice while fields held at exactly {body,eyebrow,heading}. Preservation-by-instruction confirmed; 3a
caveat closed.

3b setup state: first attempt inserted the component but the work item failed (unreplaced <TEST_SITE_UUID>); second
attempt hit duplicate-key on the component (already there — harmless) and the work item INSERT succeeded. Net: one
inactive zzz-fg-reject + one triaged item f2_reject_zzz-fg-reject.

MY ERROR (constitution: schema before SQL): the verification query used created_at; agent_error_log's timestamp
column is occurred_at (schema found in /mnt/project/some_schemas L921: id, occurred_at, site_id, domain,
work_item_id, orchestration_id, agent_type, agent_id, pod_name, step_name, action, error_message, error_code,
severity, context, resolved…). Runbook 3b query corrected. IMPORTANT ambiguity flagged: fields-unchanged + 0
version rows is consistent with BOTH "guard rejected" AND "item not yet processed" (updated_at 11:58 is just the
INSERT default now()). Discriminators, in order: agent_error_log row (occurred_at query) > pod logs grep
("regeneration removes/renames"/zzz) > NOT the work-item status ('complete' can be the parent orchestration even
when the child failed — seen at 19:55). Baseline for the unchanged-check: updated_at 2026-07-02 11:58:00.16898.

### 9m. 3b run 1: guard NOT exercised — the store FORKED a duplicate instead (new finding F4)

Flow: item claimed → full workflow ran (11:59:57–12:00:48) → store SUCCEEDED with
stored_component={component_id:80222fc1, function:"simple-intro-band", section_type:"zzz-fg-reject",
status:"created"}. The LLM chose function 'simple-intro-band' (from the description), NOT 'zzz-fg-reject'; the
store's lookup is WHERE function=$1 AND forked_from IS NULL → no match → CREATION branch → new ACTIVE component
80222fc1 for a section_type that already had one. Inactive fixture untouched (11:58:00 baseline held, zzz_* fields
intact); 0 rejection rows is real and explained (query correct this time). MY FALSIFIED ASSUMPTION owned: "LLM
always mirrors section_type→function" — held twice in 3a by luck, failed on the first 3b try (temperature 0.4).

**F4 (structural finding):** regen-vs-create is keyed on the LLM-chosen function → nondeterministic. Miss case =
silent FORK: parallel duplicate per section_type; dependents safe (component_id untouched, no stranding — June-24
was the MATCH case) but library fragments (selector on section_type can pick either; usage_count/dedup degrade;
guard bypassed by fork; reuse-before-recreate violated by the platform). Mitigation shipped:
`F1prompt2_pin_function.sql` — anchored idempotent migration pinning {{.existing_component.function}} inside the
same regeneration rule block (loader already returns function; dormant for new sections). Store-side advisory
(warn when function misses but an active same-section_type row exists) FLAGGED as follow-up, not built — multiple
components per section_type may be legitimate across site types, so advisory-only.

**3b redo designed:** clean ALL zzz rows first (cleanup covers 80222fc1 via section_type='zzz-fg-reject'), then
fixture with function='simple-intro-band' (the demonstrated LLM choice) + description that pins it ("Set the
function name to exactly simple-intro-band."), fields zzz_*, inactive; new item_key f2_reject_zzz-fg-reject_2.
Loader still skips (inactive) → blind gen → function collides → store matches fixture → guard should reject.
Runbook 3b rewritten accordingly; falsified-mirroring sentence corrected; F4 section + checklist lines added.

Parking-lot observation (2nd sighting): work item 'complete' at 11:58:52 while the logged orchestration ran
11:59:57–12:00:48 — dispatch-loop item-status handling is loose (first seen at 19:55 with the failed child). Noted,
not chased.

### 9n. Pin applied; redo attempt 2 misfired (cleanup skipped) — paste-safe redo issued

Pin migration applied clean (pin_present=t, names_present=t, snapshot taken). Redo attempt: the CLEANUP was skipped
(my runbook had it as prose before the SQL — layout fault, now fixed), so the fixture INSERT collided on
content_components_name_key (old inactive 'zzz-fg-reject' still present) — fixture NOT inserted — while the work
item _2 DID insert. Against the un-cleaned state, item _2 is a HAPPY-PATH regen of the FORK, not a reject test: the
loader (section_type + is_active=true) now finds the fork 80222fc1 → feeds its cta_*/headline/subheading names +
(new pin) function simple-intro-band → LLM preserves → store matches the fork → guard passes → fork regenerates.
The zzz_* fixture's function ('zzz-fg-reject') is never looked up. So 0 rejection rows = consistent with both
not-yet-processed and fork-regen; the count(*) verify errored because my SQL assumed one component per section_type
('more than one row returned by a subquery') — corrected to IN.

Actions issued (and runbook rewritten paste-safe, cleanup INLINE as step 0/1): delete triaged reject items → clean
all zzz rows (component_versions/page_components/content_components by section_type IN (keep,reject) + f2 items) →
re-insert fixture (inactive, function='simple-intro-band', zzz_* fields) → fresh item _3 (description pins the
function). Expected pass: agent_error_log rejection naming zzz_alpha/beta/gamma; single inactive fixture unchanged
at its new insert baseline; version count 0. A new ACTIVE row appearing instead = the description pin failed to
hold the function (forked again). Note: item _2, if it ran before cleanup, just regenerated the fork — harmless,
removed by cleanup.

### 9o. F2 Tier 3b PASS — guard observed firing live; F2 COMPLETE

Redo 3 results (2026-07-02 16:39): the rejection propagated through three log layers, all in agent_error_log —
16:39:01 component_validation_rejected (warning; recordValidationRejection itself: function="simple-intro-band"
section_type="zzz-fg-reject", names the 3 stranded fields zzz_alpha/zzz_beta/zzz_gamma) → 16:39:05 step-level
failure (fatal, work_item_id 4733af64, guard message intact) → 16:39:06 CHILD_ORCHESTRATION_FAILED (parent). So:
description pin held the function → store matched the INACTIVE fixture (is_active-agnostic lookup) → guard fired.
State checks: fixture is the single row, still inactive, zzz_* fields, updated_at at the 16:37:03 insert baseline;
component_versions count 0 — guard fires BEFORE snapshot/UPDATE, nothing mutated. Item _2 (pre-cleanup) had
completed at 16:32 as the predicted harmless fork-regen; cleanup swept it (DELETE 4 versions / 3 components / 5
items).

**F2 COMPLETE**: Tier 1 unit ✔, 3a preservation ✔ (×2 regens, md5-verified template change), 3b reject ✔ (live
firing, three-level visibility, zero mutation). Runbook banner refreshed (fix deployed and proven); F2 + pin ticked;
F2-cleanup line added (fixture + f2_reject_* items — _3 will retry-reject until removed). Parking-lot: step-level
row logs error_code=UNKNOWN severity=fatal — the specific code is lost at that layer; noted, not chased (alongside
the two dispatch-loop status-looseness sightings).

Remaining: F2-cleanup → F3 deploy (F3a create_rerender_items scoped reason + F3b store reason re-add, Go first;
then F3c rerender-pages config SQL) → R4 backup → R5 recover the five → R6 verify.

### 9p. F2 cleanup done; F3a/F3b verified ALREADY APPLIED in the repo's latest files

Cleanup counts: item _3 deleted (1), fixture component deleted (1), everything else already gone (0s). Refinement to
the dispatch-status observation: item _3 read status='complete' WITH the guard's full rejection message in the
`error` column — so the loop records the child error text but does not set a failure status ('complete' ≠ success;
read the error column).

User uploaded latest repo copies of both files asking to "add the patch" — verified instead that BOTH ALREADY CARRY
the F3 changes: create_rerender_items_action.go is byte-identical to the staged F3a version (reason+component_id in
InputSpec, dependent-page scoping, spec["reason"] stamp); store_generated_component_action.go = guard-only + exactly
the F3b hunk (diff shows only the comment + reason in the specJSON literal). Both gofmt -e clean. Nothing to add;
staged outputs copy refreshed to the latest store file.

OPEN (deploy provenance): rerender-pages row shows image_tag v1.0.1086 (updated 2026-07-01 20:04) and its
create_rerender_items step config STILL lacks reason/component_id (F3c not applied). Whether v1.0.1086 was built
BEFORE or AFTER the F3a/F3b patches landed in the repo is not knowable from here — the 16:39 guard firing only
proves F1 is live (rejection happens before createRerenderWorkItem; create_rerender_items never saw reason inputs).
If in doubt: rebuild + bump (cheap), THEN F3c SQL — Go-first ordering stands (old action's InputSpec lacks
reason/component_id; ExtractActionInputs behaviour with unknown config keys unverified).

Recovery route note: with F3 live, Route B (re-key content_data + one needs_rerender per site carrying
reason+component_id) both recovers the five WITHOUT LLM cost (preserves the per-page 70/2400 values) AND verifies F3
end-to-end (scoping + reason stamp + section re-render). Candidate plan for R5; decide next turn.

### 9q. F3c APPLIED; v1.0.1087 building; R5 Route B made concrete (gated)

F3c applied clean: rerender-pages create_rerender_items config now carries all five keys (domain, site_id,
pages_field, reason, component_id); snapshot taken (source_version=6). User building v1.0.1087 — built from the
verified repo state, which settles the provenance question: once pods + image_tag rows show 1087, F3 is complete
(Go F3a/F3b + config F3c).

Read createRerenderWorkItem's actual INSERT from the latest store file (reuse before recreate): columns
(site_id, source, pipeline, item_type, severity, summary, priority, handler_agent, status, created_by, spec,
item_key), values 'build'/'needs_rerender'/'medium'/99/'rerender-pages'/'triaged', NOT EXISTS dedup on
(site_id, item_key, status NOT IN complete/verified/rejected/wont_fix/failed). Recovery trigger mirrors it exactly
(source/created_by 'manual-recovery', distinct item_key 'recover_system_stats:<component_id>').

RUNBOOK R5 rewritten concrete (Route B, gated): GATE = image_tag 1087 on rerender-pages + component-creator before
any trigger (older rerender-pages → unscoped assemble-only items). B pre-flight (read-only, paste before writing):
B1 freshness (component updated_at still 2026-06-24 15:06:12.940335, 22 fields — another chat co-manages
components); B2 cta_url/cta_label schema definitions (static+fallback → absent keys fine; llm+required → CTA renders
empty, decide interim vs writer follow-up); B3 key coverage (exactly the 24 old keys, 5 rows, 1 distinct render
md5); B4 dry-run jsonb_build_object of all five rebuilt objects (20 explicit mappings; stat_N_icon dropped; cta_*
left unset pending B2). C after B: C0 backup CTAS page_components_bak_sysstats_20260702 (count 5) → C1 re-key
UPDATE (same expression; expect UPDATE 5; re-check keys = the 20 new) → C2 three needs_rerender inserts (INSERT 0 3).
R6 then verifies: scoped page_rerender items per site (leopardess 1, robot-hands 2, ai-agent-orch 2, each with
spec.reason), five renders regain content (has stat1_value + section_headline), five DISTINCT md5s, and the live
bands. Note: the 20-mapping explicit jsonb_build_object is the reviewable middle ground vs a one-off Go tool for 5
rows — dry-run-first keeps it safe; Route A (needs_page) stays the fallback.

### 9r. F3 COMPLETE — fleet on v1.0.1088; gate passed; next = R5 pre-flight

Gate query: rerender-pages, component-creator, page-rerender all image_tag v1.0.1088 (build landed as 1088 not 1087;
built from the byte-verified repo state per sequencing). F3c was applied a second time via the runbook's unfiltered
SQL — idempotent (UPDATE 1, identical values, verify shows the same 5 keys), one extra snapshot taken; only the live
row matched type='rerender-pages', so snapshots evidently don't collide on that filter in this schema. F3 status:
COMPLETE (F3a+F3b in the running image, F3c config live). Runbook: F3 + F2-cleanup ticked; banner updated to
"recovery is what's left". Next: user runs R5 pre-flight B1–B4 (freshness, cta_* schema decision, key coverage,
dry-run rebuild) and pastes results; then C0 backup → C1 re-key → C2 triggers → R6 verify (which doubles as the F3
end-to-end check: expect scoped page_rerender items — leopardess 1, robot-hands 2, ai-agent-orch 2 — each carrying
spec.reason).

### 9s. R5 pre-flight B1–B4 GREEN; cta decision made; C issued

B1: component untouched (2026-06-24 15:06:12.940335, 22 fields) — concurrency window clear. B3: exactly the 24 old
keys, 5 rows, 1 distinct render. B4: all five rebuilt objects eyeballed — per-page content intact (70/2400 split,
distinct copy), 20 keys each, no nulls. B2 DECISION: cta_label = static + fallback "View full report" → safe unset.
cta_url = TIER C (source site_specs.cta.primary_url, required) → by design never a content_data key; renderer
resolves from site specs. Leaving both unset is the correct shape, not a compromise. Unverified caveats (not
blockers): whether each site's site_specs actually has cta.primary_url, and whether the no-LLM section-rerender
resolves Tier-C sources — R6 gains an href="" heuristic check on the rendered section; TRUE → follow-up finding
(missing spec value or Tier-C gap in rerender path), not a recovery failure.

C issued (runbook R5 block): C0 backup CTAS page_components_bak_sysstats_20260702 (expect 5) → C1 re-key UPDATE
(expect UPDATE 5; C1-check keys = the 20 new) → C2 three needs_rerender inserts mirroring createRerenderWorkItem
(reason=section_data_resolved + component_id; expect INSERT 0 3). First live F3 exercise: watch rerender-pages log
"scoping to component dependents" with dependent_pages 1/2/2. R6a = exactly the 5 scoped page_rerender items with
spec.reason; R6b = has_stat1/has_headline TRUE ×5, FIVE DISTINCT md5s, cta_href_empty check.

### 9t. C executed green; R6 read too early + my weak-needle mistake; F3 scoping half-proven; 2 sites' items missing

C: backup 5 ✔, UPDATE 5 ✔, keys = the 20 new ✔, INSERT 0 3 ✔. R6 read shortly after: NOTHING re-rendered yet —
updated_at 19:14:33.300435 identical ×5 is C1's OWN bulk-UPDATE stamp (not a render); md5 still stale 462c7c45;
build_status pending; robot-hands' 2 items still 'triaged'. MY MISTAKE owned: R6b's has_stat1 needle used
stat1_value — "70" is a 2-char substring that false-positives against CSS (t on the three 70-pages); "2400" (4ch)
correctly f; has_headline (long distinctive string) f on all 5 = decisive no-render-yet. R6b corrected to
stat1_label; runbook R6 rewritten with the needle lesson + do-not-read-early note.

F3 half-proven: robot-hands got EXACTLY its 2 dependents (gripper-detail+index) reason-stamped → scoping worked for
that site. leopardess + ai-agent-orch: ZERO reason-stamped items. Candidates (D1–D3 + log grep issued):
(1) their needs_rerender not yet processed (dispatch timing); (2) DEDUP swallow — create_rerender_items ON CONFLICT
DO NOTHING on item_key page_rerender_<name>_<siteid>; pre-existing OPEN page_rerender items silently block the new
reason-stamped ones (SIDE FINDING: itemsCreated++ not gated on RowsAffected → items_created log can OVERCOUNT on
conflicts — accounting flaw, note for later); (3) get_pages_for_rerender include_statuses deployed/active — pages in
other statuses → has_pages false → skip; (4) run failed (status/error). D2 lists open page_rerender items (old ones
have reason NULL); D3 lists page statuses; log grep shows dependent_pages counts per site.

### 9u. R6 read: all five processed per-page yet byte-identical → the CARRY-path signature; cta_url readiness is the prime suspect

R6b (corrected needles): all five rows build_status='deployed', per-page distinct updated_at 19:17:01–19:21:55
(so leopardess + ai-agent-orch items WERE created and processed after the early R6a read), yet md5 still the stale
462c7c45 empty render, has_stat1_label/has_headline FALSE ×5. Per-page save-style row updates + identical bytes =
the rerender_page_sections CARRY path fingerprint: gate matched → action ran → the system-stats section failed
planSection readiness → carryStoredSection re-emitted the stored (empty) HTML → save_sections wrote it back (fresh
updated_at, same bytes) → assembled + deployed. Pulled the action source from the transcript to confirm mechanics:
NULL pre-check escalates the WHOLE page (needs_page) if ANY section lacks content_data; per-section: no component →
carry; plan.Status!='ready' (a REQUIRED non-LLM field can't resolve) → carry; empty template → carry; else
RenderTemplate + persist merged content. Prime suspect = cta_url (source site_specs.cta.primary_url, required:true,
NO fallback — the B2 caveat): if the sites' specs lack cta.primary_url the section is PERMANENTLY not-ready →
carried forever. If confirmed: F3 worked END-TO-END (scoping+reason+gate+section-rerender all ran) — blocker is one
layer deeper (readiness). Alternates not yet discarded: gate-didn't-match/assemble-only (but that path shouldn't
produce per-page save-style row updates) and escalated-to-writer (needs_page items would exist). site_specs schema
checked (per-aspect rows: site_id, aspect, data jsonb, is_current; unique (site_id,aspect) where is_current).
DIAGNOSTICS ISSUED (runbook R6 D4): page-rerender log grep (action logs "not ready, carrying"/"escalating"/"done
rerendered=N carried=M" — decisive); site_specs cta check both shapes (aspect='cta' row; data ? 'cta' nested);
needs_page escalation check; R6a re-run any-status. Remedies branch: populate real cta.primary_url per site vs relax
the shared component's cta_url (required:false / fallback) — latter = co-managed shared-component write (R7 +
snapshot). Await log + SQL results before choosing.

### 9v. CARRY path established by data; blocker = regen-ADDED required cta_url with no site spec (F5); remedy branches

Results: site_specs cta = 0 rows BOTH shapes across all three sites (query sound — schema-checked per-aspect
table); needs_page escalations = 0 (escalate branch out); R6a any-status = ALL FIVE reason-stamped items exist and
completed 19:15–19:19 (leopardess's insert DID land → the dedup unique index's partial predicate evidently doesn't
cover 'unresolved' — parked: inconsistent with the store's NOT EXISTS status list, which would have blocked). The
user's log grep matched only the repo-analysis JSON embedded in saved log files (not runtime lines; files likely
don't cover the window) — but the data is decisive without it: gate matched, rerender_page_sections ran on all
five, and cta_url (required:true, source site_specs.cta.primary_url, NO fallback) is deterministically
unresolvable → planSection not-ready → carryStoredSection re-emits the stale empty HTML → save writes it back
(fresh per-page updated_at, same bytes, deployed). **F3 PROVEN END-TO-END** (scoping to exactly the five, reason
propagation, gate, section-rerender execution). Recovery blocker is one layer deeper and is a NEW FACET of the
regen bug: the 15:06 regen ADDED a required Tier-C field the sites can't satisfy — renames strand stored content;
required additions strand RENDERABILITY. Flagged **F5**: extend the F1 guard to reject (or force
optional/fallbacked) added required fields on regeneration. Not built now.

Remedy branches on whether the template wraps the CTA in {{if .cta_url}} (RenderTemplate = Go text/template, so
conditionals are possible): Branch A wrapped → input_schema jsonb_set fields.cta_url.required=false (CTAS backup
first; R7 freshness in the same read) → section renders cleanly sans CTA → re-trigger the five (same C2 inserts;
'complete' excluded by dedup guard) → R6b. Branch B bare substitution → required:false would ship a dead
href="" button and NO single fallback URL is correct across sites (phantom-CTA lesson: never invent URLs) → per-site
real `cta` spec — but FIRST read the resolver's site_specs.* convention from plan_sections_action.go
(newSourceResolver) before inserting aspects; do not guess aspect naming/shape. One decisive read issued
(updated_at + has_if_blocks + cta_region substring). RUNBOOK: R5b + F5 sections added; checklist R5 ticked
(C0–C2 done, F3 proven), R5b + F5 pending.

### 9w. Branch B confirmed + concretised from plan_sections_action.go (uploaded)

Template read: has_if_blocks=f (no Go-template conditionals anywhere; cta_region shows the styled .stats-cta
button) → required:false would ship a visible dead button → Branch B (per-site spec). Freshness held (component
updated_at still 2026-06-24 15:06). Resolver convention CONFIRMED from plan_sections_action.go: ensureSpecs =
SELECT aspect, data FROM site_specs WHERE site_id=$1 AND is_current=true → specs[aspect];
resolveSpecPath("cta.primary_url") = specs["cta"]["primary_url"]. planSection required semantics CONFIRMED:
resolve-miss on required → on_missing switch; "Required with no fallback — defer" (cta_url has neither) →
deterministic defer → carry. pages schema checked: pages.url (varchar 500 NOT NULL) is the real link format —
CTA targets must be an existing page's url (phantom lesson). Sequence issued (runbook R5b Branch B): (1) list
pages per site → pick target (contact if present, else primary conversion page); (2) INSERT 3 site_specs rows
(aspect='cta', data={primary_url}, source/created_by='manual-recovery', is_current=true; unique slot verified
free); (3) re-trigger C2 (prior items complete → dedup guard passes); (4) R6b — expect has_stat1_label/
has_headline true ×5, five DISTINCT md5s, cta_href_empty false; resolver persists cta_url into content_data
(stored⊕resolved merge — expected). Awaiting the page list to fill the three URLs.

### 9x. Specs inserted with UNVERIFIED /contact.html ×3 — verification gate before retrigger

User ran the spec INSERT (0 3 — all three sites now have aspect='cta' data.primary_url='/contact.html') WITHOUT
Step 1 (the page listing). Same shape as the phantom-CTA bug (that phantom WAS /contact.html). Mechanics note: the
resolver only checks spec PRESENCE, not URL validity — so readiness is unblocked and the re-render will succeed
regardless, but a wrong URL bakes into rendered_html + content_data (stored⊕resolved merge) and costs a third
render pass to correct. Issued: contact-page existence check per site (url ~* contact) + full page listing
fallback; per-site UPDATE template (jsonb_set data.primary_url — the current (site_id,'cta') slot is occupied so
UPDATE not INSERT); then the user's retrigger script (verified correct — byte-identical to C2; prior items
'complete' → guard passes → expect INSERT 0 3); then R6b extended with cta_url_baked (rendered contains
content_data->>'cta_url') alongside the label/headline needles + distinct-md5 + cta_href_empty checks. Runbook R5b
gains the verify-URL step (renumbered 2–5). Awaiting the contact-page listing before the retrigger.

### 9y. URLs verified real (all three /contact.html); this R6b read = PRE-retrigger state

Contact check: all three sites have name=contact at url=/contact.html, active — the unverified insert turned out
correct; no spec UPDATE needed; readiness blocker genuinely cleared. R6b read: timestamps 19:17:01–19:21:55
IDENTICAL to round 1's to the microsecond → nothing has touched the rows since the pre-spec pass → the retrigger
has not run yet (its INSERT 0 3 absent from the paste; per the rule, unchanged-result checked against the obvious
cause first). Query bug of mine fixed: cta_url_baked came back NULL (not false) because content_data->>'cta_url'
is NULL until a render merges the resolved value — strpos(html, NULL)=NULL; coalesced in the runbook R6b. Issued:
run the retrigger (script verified correct) → two now()-relative sanity checks (trigger items claimed; fresh
reason-stamped page items) → corrected R6b. Pass = updated_at newer than retrigger, five DISTINCT md5s ≠ 462c…,
has_stat1_label/has_headline true, cta_href_empty false, cta_url_baked true (merge writes /contact.html into
content_data + render). If it STILL carries with the spec present → page-rerender live logs' "not ready" line
names the residual blocker; with cta_url resolvable planSection has nothing left to defer on for this section.

### 9z. Retrigger round 2: TWO of five recovered + verified end-to-end; three pending; sixth dependent appeared

Retrigger INSERT 0 3 at 2026-07-03 12:57:33. Triggers: leopardess + robot-hands complete; ai-agent-orch still
'claimed' at 13:23 (in flight). R6b: **leopardess/index RECOVERED** (md5 6227e777, len 6845, 13:02:32) and
**robot-hands/index RECOVERED** (md5 5477b4a6, len 6946, 12:59:35) — has_stat1_label/has_headline TRUE,
cta_href_empty FALSE, cta_url_baked TRUE on both. Full chain proven INCLUDING RENDER: re-key → spec → F3 scoped
reason item → gate → rerender_page_sections → planSection ready → RenderTemplate fills re-keyed content →
resolved cta_url merged into content_data + baked into bytes. gripper-detail UNTOUCHED (yesterday's 19:18:29
timestamp → no save touched it this round → item queued OR insert swallowed — robot-hands trigger complete yet
only index rendered, so the listing must distinguish). ai-agent-orch ×2 untouched, trigger claimed → wait.
SIXTH dependent appeared: vonc.com/index (13:15:55, healthy, content + cta baked) — NOT ours; an ongoing platform
build selected the shared component and rendered under the current schema (guard-protected reuse working; likely
why dispatch is slow). Issued: one sweep (this-round items ≥12:57: trigger progress + which page items exist and
states) + pg_indexes definition for the site_work_items item_key dedup index (read-only; decides
queued-vs-swallowed for gripper-detail AND closes the parked predicate inconsistency from yesterday's 'unresolved'
non-block), then R6b re-run. Pass = all five ≠ 462c7c45 with needles true.

### 9aa. Sweep: gripper-detail NOT swallowed — STUCK claimed; ai-agent-orch child stalled; dedup index captured (F6)

Sweep (>12:57): gripper-detail page_rerender EXISTS, reason-stamped, claimed 12:58:24.99, page row untouched →
its page-rerender child never did the work. ai-agent-orch needs_rerender claimed 12:57:33, ZERO page items for
that site → its rerender-pages child stalled before create_rerender_items (or never started). Siblings on the
same images completed (12:58/13:01) → per-spawn hiccups, not systemic; vonc build contention plausible. Oddity
noted: all three triggers share updated_at 12:57:33.251786 despite different statuses (complete/complete/claimed)
— dispatch stamps once at claim; fourth sighting of loose item-status semantics.

Dedup index captured: idx_swi_dedup UNIQUE (site_id, item_key) WHERE item_key IS NOT NULL AND status NOT IN
(complete, verified, rejected, wont_fix, failed, **unresolved**). Closes yesterday's parked inconsistency: the
2026-05-01 'unresolved' squatter never blocked because 'unresolved' is index-terminal. NEW small gap → **F6
flagged**: the store's NOT EXISTS guard (and my C2 copy) omit 'unresolved' → Go guard STRICTER than index (an
unresolved squatter would block createRerenderWorkItem where the index wouldn't). One-line alignment later; pair
with the itemsCreated overcount fix.

Unstick plan issued: (1) pods + dispatch-loop log grep to confirm no active child (avoid double-render race);
(2) if still claimed + idle → UPDATE the two items to 'triaged' (targeted: the recover item_key OR the
reason-stamped gripper-detail page item; expect UPDATE 2); the trigger re-runs the whole rerender-pages pass —
same as original; (3) corrected R6b — pass = all five ≠ 462c7c45 with needles true; vonc rides as healthy sixth.

### 9ab. Runbook R5b rewritten as a concrete step-by-step; redeploy changes the unstick calculus

User asked: where are we in R5b + make it a clear stepwise guide with concrete values and the bash inline. Done:
R5b now has a constants block (component uuid, sites/pages, /contact.html, the stale-render fingerprint), a ticked
progress list (1–6 ✔, 7 ← NEXT, 8 pending), and steps 7a–7d + 8 with full SQL/bash inline. ALL placeholders in the
runbook substituted repo-wide (<COMPONENT_UUID> → fdd92ad4…, <D1..3> → the three domains, <URL1..3> →
/contact.html with a verified-note, <TEST_SITE_UUID> → gamesdesign id + comment, <PAGE_UUID> → domain/name lookup
subquery). WHERE WE ARE: R5b step 7. NEW CONTEXT: user just redeployed → agents restarted → no pre-restart child
can still be working the two claimed items → once 7a confirms still-claimed, 7c re-triage is safe without the
pod-activity caveat (note added to the runbook). 7a sweep → 7c UPDATE (expect 2) → 7d watch (dependent_pages=2 for
ai-agent-orch; carried=0 per page) → step 8 R6b (pass = all recovery-site rows off md5 462c…, needles true, cta
baked; vonc healthy sixth) → R6 live-deploy check → closeout.

### 9ac. Step 7 resolved without intervention; step 8 (R6b) is next — gripper-detail unproven until then

7a re-run: ALL rows complete. ai-agent-orch's stalled child ran late — 13:42:16, both scoped page items created +
completed (45-min claim-to-run; dispatch backlog / vonc contention / redeploy kick). gripper-detail's item flipped
claimed→complete with IDENTICAL updated_at 12:58:24.992327 (status transition didn't bump the timestamp — 5th
loose-semantics sighting; noted, not chased). Fleet grep: only the 33-min-old dispatch loop running — consistent,
rerender/page-rerender pods are per-job transient and exited after 13:42. 7c re-triage SKIPPED (0 rows would
match). Caveat carried into the runbook: item complete ≠ render done (established repeatedly) — gripper-detail
especially unproven until R6b. Step 8 issued: R6b (pass = five recovery rows off 462c…, needles true, cta baked;
vonc healthy sixth). Contingency embedded in the runbook: if gripper-detail alone stays stale, one-row re-triage
of just its reason-stamped item (transient pod logs are gone post-redeploy; agent_error_log/orchestration_states
would be the deeper record if the re-run also fails).

### 9ad. R5b step 8 PASS — all five RECOVERED and verified; arc reduces to R6 eyeball + flagged follow-ups

R6b 2026-07-03: five recovery rows all off the stale 462c7c45 md5 — case-study 57fab5d8 (13:45:51), ai-agent index
2c438bf1 (13:43:26), leopardess index 6227e777 (13:02:32), gripper-detail e525b8f4 (13:44:31 — the late child DID
do the work; the earlier complete-status ambiguity resolves), robot-hands index 5477b4a6 (12:59:35). All:
has_stat1_label=t, has_headline=t, cta_href_empty=f, cta_url_baked=t; distinct lengths. vonc.com/index healthy
sixth (2ef621fc, 13:15:55). RECOVERY COMPLETE at the DB level. Deploy path (git→Actions→Backblaze) ran inside
page-rerender, so R6 = browser/curl eyeball of the five live URLs (listed in the runbook checklist) — expect the
band with per-site stats + "View full report" → /contact.html. Docs: runbook step 8 + R5b + checklist ticked,
banner rewritten to recovered-state with the follow-up list + backup DROP note; HANDOFF gains a top status banner.
Outstanding flags carried: F4 fork advisory; F5 guard extension (regen-ADDED required fallback-less fields — this
incident's second facet); F6 store-guard/index status alignment + itemsCreated overcount; 40 stale 'unresolved'
items (2026-05-01) hygiene; dispatch item-status semantics (five sightings: complete-at-dispatch, error-in-complete,
status-change-without-timestamp). gamesdesign: out of scope — a future rebuild picks the component up healthily as
vonc demonstrated.

### 9ae. R6: leopardess/index CONFIRMED LIVE (screenshot); guessed-path correction; four pages left

User screenshot of leopardessconsulting.co.uk/index.html shows the recovered system-stats band end-to-end:
eyebrow "BY THE NUMBERS", headline "Built for production. Measured in production." (exact re-key value), the four
stat cards (70% Deployed Agents …), footnote, and "View full report" → /contact.html (user confirmed the link) —
i.e. the Tier-C cta_url resolution working live. Terminology answered: a band = one full-width page section = one
page_components row's render; this IS the band. Side-note (content, not pipeline): some stat pairings read oddly
("3ms / Orchestration Model", "99.9x / Uptime Target") — original writer content faithfully restored; pre-existing;
content-writer follow-up territory if desired. MY CORRECTION: I had baked a guessed path
(/entities/gripper-detail.html) into the R6 URL list — replaced with a pages.url lookup query in the runbook (never
guess paths; pages.url is the source). Remaining: eyeball the four other pages via https://<domain><url>; then the
arc closes bar the flagged follow-ups (F4/F5/F6, unresolved-item hygiene, dispatch status semantics) + backup DROP.

### 9af. pages.url confirmed (my /entities/ guess was in fact right — lookup was still the correct move);
### NEW SYMPTOM R6c: gripper-detail LIVE PAGE EMPTY

URLs: ai-agent /case-study-….html + /index.html; robot-hands /entities/gripper-detail.html + /index.html. User
reports gripper-detail live page EMPTY — different class from the empty-band incident: R6b proved the system-stats
SECTION is rendered in the DB (e525b8f4, 13:44:31), so the gap is PAGE-level. Candidates (not conclusions):
(a) other sections of the page empty/unrendered (only ever inspected the system-stats row on this page);
(b) pages.sections jsonb empty → render_page assembled nothing; (c) assemble/deploy leg after save_sections didn't
ship (though stale would usually show the OLD page, not empty); (d) nested /entities/ artifact-path handling in
git→Actions→Backblaze. Issued three reads: all-sections listing for the page (slot_name/function/build_status/len),
the pages row (build_status, last_built_at, deployed_at, header/footer lens, jsonb_array_length(sections)), and the
curl (http_code + size + head -60) to distinguish 404 / 200-empty / 200-head-only. R6 ledger: leopardess/index ✔;
gripper-detail LIVE-EMPTY (R6c open); robot-hands/index + both ai-agent pages UNCONFIRMED — asked the user to
eyeball those too. Runbook R6c block added with the full diagnosis set.

### 9ag. gripper-detail reframed (200/62KB, visually blank); two out-of-scope findings triaged; ledger updated

Docs 21/22 arrived EMPTY on my side — the two SQL results (all-sections listing + pages row) still needed;
said so explicitly. curl: 200, 62,446 bytes, full head with base CSS referencing custom properties
(var(--color-background) etc., normally defined via :root from pages.rendered_head) → NOT a missing artifact;
"empty" = visually blank. New candidate split: (i) theme-variable block missing from the head → dark page renders
blank-ish; (ii) 62KB is head/CSS with body bands absent → assemble leg. Artifact greps issued (<section count,
data-component inventory, :root count, 'Gripper Models Indexed' presence) to pair with the SQL reads.

OUT-OF-SCOPE (triaged, not chased): (1) robot-hands/index gained a GAUNTLET band (vonc.com's component —
gauntlet-interface/cta in the library) — shared-component contamination in the OTHER chat's active area (user
attributes it there); evidence query added to the runbook (page_components × content_components for robot-hands
index, ORDER BY updated_at DESC — shows when/which slot) to hand off. Same bug family (shared components +
selector), reinforces the freshness/snapshot discipline. (2) ai-agent/index case-study cards show alt-text instead
of images + one mis-sized component — image-pipeline territory, pre-existing/elsewhere per user; parked.

R6 LEDGER: leopardess/index ✔ · robot-hands/index stats band ✔ (unrelated gauntlet regression noted) ·
ai-agent/index mostly ok per user · ai-agent/case-study still to eyeball · gripper-detail R6c open (blank-render).

### 9ah. R6c pinned to the ASSEMBLY layer; gauntlet attribution revised (DB clean, artifact-level)

gripper-detail DB: all 8 sections deployed 13:44:31, healthy lengths (~49KB total) — recovery held through
save_sections. Artifact (62KB): only 4/8 bands (call-to-action, features, product-specs ×2 attr, system-stats);
MISSING product-hero/product-details/product-card-with-cta/generic-text-block; :root = 0 (no theme vars → dark
theme renders blank); pages row: sections_listed = 0, rendered_header/footer NULL, deployed_at 13:45:00,
last_built_at NULL (assemble path stamps deploy only — semantics note). MY OMISSION owned: earlier pages-row query
skipped rendered_head (the likely :root carrier) — re-included in the comparative. ⇒ assembly built from an empty
sections list with no head/header/footer and produced wrong membership + a duplicate.

GAUNTLET REVISED: robot-hands index page_components has NO gauntlet row (7 clean sections: hero, system-stats,
features, brief-explanation, info-card-grid, tool-list, call-to-action) yet the live page shows the gauntlet band →
foreign band enters at assembly/deploy, NOT page_components → my hand-to-other-chat framing was premature; both
affected artifacts were assembled by OUR rerender-triggered runs (12:59, 13:44) with vonc building concurrently
(13:15). Same defect class as gripper-detail's mis-assembly until proven otherwise.

Issued: comparative pages rows (leopardess/robot-hands×2/ai-agent/vonc; head/header/footer lens + sections_listed)
— discriminator test: healthy = listed>0 + head non-null?; artifact inventories (robot-hands index + leopardess
control: data-component uniq -c + :root count); assemble-action source request (rerender_single_page / render_page
/ rendered_head grep) to see the empty-sections fallback and whether it can cross pages/sites. NOT chased yet:
which write emptied pages.sections / whether it was ever populated for gripper-detail.

### 9ai. Assembly machinery mapped from code; R6c splits into three checkable mechanisms

Code reads (rerender_single_page_action.go, assemble_from_library.go, registry.go, site_components schema):
- rerender_single_page: sections = page_components by page_id ORDER BY position; **visible-text filter** drops
  bands with ≤10 chars of stripped text (logger.Warn per skip; the empty-band dropper lives HERE — so pages.sections
  jsonb is NOT assembly membership; my earlier sections_listed=0 inference was a red herring for membership,
  though it remains a metadata oddity). head/header/footer from **site_components** (site-scoped);
  no stored head → buildDefaultHead = 5-line head + /assets/css/styles.css link (NO inline base styles).
- assemble_from_library (registry L493, "Assemble a page from component library templates"): the ORIGINAL
  writer-time builder — components by build-plan names, theme CSS from **css_themes** — a third head shape.
  The gripper-detail artifact's big inline base-styles head matches THIS builder, not the other two.
- comparative rows: head/hdr/ftr NULL on ALL pages incl healthy → pages columns unused; leopardess artifact exact
  (6/6, :root×2) → leopardess has a stored site head; robot-hands :root=0 on both artifacts.

MECHANISM SPLIT (checks issued):
(i) gripper-detail = likely a STALE assemble_from_library-era artifact — the 13:45 commit may not have shipped
    through Actions→Backblaze. Decide via the repo file (8 bands? 'Gripper Models Indexed'?) + Actions status +
    cache-busted curl.
(ii) robot-hands index gauntlet-in-hero: candidate = shared 'hero' template altered during the vonc work
    (update_component_html has NO field guard) and our 12:59 ALL-SECTIONS re-render (documented F3 blast radius)
    faithfully re-rendered robot-hands' hero from it — gauntlet statics shown where fields miss. Decide via
    strpos(hero rendered_html,'Gauntlet') + content_components/versions history for hero. If confirmed → flag F7 +
    coordinated repair with the other chat.
(iii) theme blank: robot-hands lacks a site_components head row (leopardess has one) → fresh assemblies use the
    default head + /assets/css/styles.css (check the css exists live); the OLD artifact's inline head lacks :root.
(iv) index 6/7: info-card-grid possibly dropped by the visible-text filter — eyeball its stripped text.

### 9aj. Gripper-detail = STALE CACHE (deploy shipped); two theories falsified; F7 flagged

Cache-busted curl: plain URL served the old 4-band artifact (product-hero 0) while ?cb=1 returned product-hero ×57
(the new 8-band assembly's CSS+markup) ⇒ the 13:45 assembly DID ship to origin; the blank page was a stale
browser/CDN copy. Recovery + deploy correct; R6 for gripper-detail closes on a hard refresh / cb fetch visual.
(The repo check ran in ~/projects/agentchassis — the PLATFORM repo, wrong one — hence "No such file"; moot now.)

FALSIFIED + owned: (i) hero-contamination — robot-hands index hero row has NO gauntlet text (f/f, 2911B, 12:59
render clean); (ii) missing-site-head — robot-hands HAS a stored head (8009B, 2026-05-20; leopardess 3209B);
(iii) my info-card-grid text_ish query stripped tags but not <style> blocks → showed CSS, proved nothing
(construction error).

GAUNTLET now two cache-testable candidates: (a) the screenshotted index is ALSO a stale cached artifact from an
earlier deploy; (b) the gauntlet is JS-INJECTED (the 1-2-3 dots + empty panel look script-mounted; script-rendered
UI never shows in a data-component grep). Issued: cb fetches of BOTH robot-hands pages with inventories + :root +
gauntlet/Play-Today greps + <script src> listing; plus SQL: does the stored 8KB robot-hands head contain :root.

**F7 FLAGGED:** shared 'hero' component updated 2026-07-02 16:43:11 with ZERO component_versions rows — a template
swap bypassing BOTH the F1 guard and version history = the update_component_html signature (no field guard; its
snapshot INSERT silently fails on the missing version_note column — known bug). An unguarded, unversioned write
path to shared components. Park next to F4/F5/F6: fix = repair the snapshot INSERT + consider extending the field
guard to that action.

### 9ak. Corrections: F7 re-scoped (snapshot already fixed); stale-cache unproven (metric mismatch); fallback-vars insight

update_component_html_action.go (uploaded, read): the snapshot INSERT is ALREADY FIXED in current code — computes
next version_number, inserts (component_id, version_number, html_template, change_description, changed_by,
created_at); comment documents the old version_note bug as repaired. ⇒ hero's 2026-07-02 16:43 update with ZERO
version rows does NOT fit a fixed-action template swap — most mundane fit = component_selector's
IncrementUsageCount (usage_count+1, updated_at bump, no version, no guard — benign). My F7 framing overstated;
query 5 (usage_count + tpl md5/len + synced flag) distinguishes swap vs bump. Residual F7 (if any):
update_component_html swaps templates WITHOUT placeholder⇄schema sync validation — narrower, lower urgency.

STALE-CACHE conclusion UNPROVEN (owned): cb=1 was measured with grep product-hero (57) and the original with a
data-component inventory — DIFFERENT metrics; cb=2's inventory matches the original's exactly, so the fetches may
be the SAME document (product-hero present ×57 as CSS classes while data-component attrs exist on only some
templates). md5sum both files + product-hero on both settles it. FALLBACK-VARS insight: templates use
var(--x, #fallback) (e.g. info-card-grid: background #ffffff, text #333) ⇒ :root=0 ≠ blank page — the index renders
fine at :root=0 ⇒ the gripper-detail "blank" needs a fresh hard-refresh eyeball; may have been the pre-13:45
window. SQL ambiguity owned: unqualified updated_at in the head query (both sc and s carry one) — corrected,
qualified version issued + leopardess comparison + styles.css-in-head check. Index gauntlet: 2 matches ARE in the
fresh artifact markup + /assets/js/snippets.js exists — context greps + snippets.js grep issued.

### 9al. Vintage question closed (one artifact); blank = missing :root in robot-hands head; hero DUPLICATE = live F4 evidence

md5sum: gd.html == gd2.html (f653a1e4…), product-hero ×57 in BOTH ⇒ ONE artifact all along. OWNED: my stale-cache
story AND the earlier "4-of-8 mis-assembled" reading were metric artifacts — data-component attrs exist on only
some templates (product-* lack them) and not all bands use <section>. The live gripper-detail = the recovered
13:45 assembly WITH the content. Also owned: my context-grep pattern (.{80} exact both sides) can't match near
line edges — why it printed nothing despite count=2; reissued with a sane pattern.

BLANK MECHANISM (query 4): robot-hands stored head (2026-05-20) has NO :root; leopardess head (2026-05-02) HAS it;
both link styles.css. Screenshot (hard refresh): header visible, pitch black below ⇒ content in the DOM but
rendered with per-component FALLBACK colors (var(--x, #333) etc.) on a dark canvas — dark-on-dark invisible.
Explains the split personality: index bands fall back light/fixed (visible); gripper-detail's product-* bands fall
back dark (invisible). One missing :root ⇒ per-component fallback lottery. Head-generator drift May 02 → May 20.
Fix path (R6d, pending vonc sample + styles.css grep): regenerate robot-hands site components (needs_rerender with
refresh_site_components:true) after CTAS backup of its 3 site_components rows; pages then re-assemble themed.

HERO DUPLICATE (query 5): TWO non-forked function='hero' rows — 2026-03-09 (2542B) and 2026-07-02 16:43 (2790B,
schema_template_synced=f, NO version row) ⇒ a creation-branch write duplicated an existing function = LIVE F4
EVIDENCE (fork-instead-of-match); store's ORDER BY is_active DESC, updated_at DESC LIMIT 1 lookup now
nondeterministic-ish for hero. Provenance query issued (id, is_active, created_from, created_at). snippets.js =
334B, gauntlet 0 — benign stub. Gauntlet location still open: band-level strpos UNION (page bands + site slots) +
fixed-pattern greps issued.

### 9am. GAUNTLET PINNED to shared brief-explanation (same bug family); theming = defined-vs-consumed drift; hero softened

Gauntlet: band-level strpos → page:brief-explanation is the ONLY carrier on robot-hands index; artifact lines 1152/
1167 show static gauntlet copy ("New Gauntlet Daily", CTA "Play Today's Gauntlet" href="#"). brief-explanation was
updated 2026-07-01 12:46 (vonc-work window, PRE-guard-deploy) → shared template gauntlet-ified with static copy →
our 2026-07-03 12:59 ALL-SECTIONS re-render stamped it onto robot-hands index (documented F3 blast radius). SAME
BUG FAMILY as system-stats, one component over. Remediation = the playbook: restore pre-07-01 template from
component_versions if snapshotted (check issued) — vonc's gauntlet copy belongs in vonc's content_data or a fork,
never as statics in a shared template — then scoped needs_rerender per affected site (F3 makes this cheap).
COORDINATE with the other chat (their active area). Dependent-set + history queries issued.

Theming REVISED: styles.css DOES define variables (:root ×2, --color- ×60) and vonc's head is ALSO rootless (8009B,
same len as robot-hands' — current generator puts vars in styles.css; leopardess = old inline-:root pattern). R6d
site-components refresh is DEAD AS DESIGNED (would regenerate another rootless head). Blank mechanism now =
DEFINED-vs-CONSUMED variable-name drift (styles.css's :root names vs the product-* templates' var() names) —
diff greps issued (+ confirm gd.html links styles.css).

Hero SOFTENED (owned): second row created 2026-01-03 'manual'; the 2025 original is is_active=false → store lookup
deterministic → Jan-era manual seeding, NOT a live fork; my F4-evidence claim overstated. Residual oddity only:
the 16:43 no-version update (pre-fix update_component_html on 1086, or a flag touch) — parked.

### 9an. R6e revised: template CLEAN — gauntlet likely in SCHEMA FALLBACK VALUES (F8); spreading to NEW sites; 13:22 unversioned write; R6f vocabulary diff concrete

brief-explanation dependents (7 rows): idea.uk index+tools (NEW SITE, built TODAY 16:27/16:36, gauntletised=t,
new-schema keys), robot-hands ×3 (index re-rendered 12:59 gauntletised=t with cd_keys showing NEW keys MERGED IN —
stat_1..3_label/value + cta_primary/secondary_label added to the old site-blob = rerender_page_sections stored⊕
resolved merge writing resolved STATIC FALLBACKS into content_data; how-it-works old-blob-only + gripper-selection-
guide EMPTY content_data = latent, pending since 12:46), vonc index 13:15 gauntletised=t (possibly legitimate own
content). LIVE template_has_gauntlet = f ⇒ the strings are NOT template statics ⇒ hypothesis: GAUNTLET COPY IN THE
SHARED SCHEMA'S FALLBACK VALUES — the F1 guard checks field NAMES not fallback CONTENT ⇒ **F8 flagged:
site-specific copy in shared-schema fallbacks is invisible to the guard**. Decider query issued (jsonb_each fields
→ source/required/fallback). Component updated AGAIN 2026-07-03 13:22:44 with NO v2 row (only v1 =
component-creator:regen 07-01 12:46) — unversioned write path provenance OPEN (pre-fix update_component_html on the
then-image, or another path). URGENCY: idea.uk consumed the 13:22 state — contamination reaches every new build
until the fallbacks are neutralized.

Harm-stop plan (pending query 1): manual component_versions snapshot (mirror snapshotComponentVersion; also
compensates the missing 13:22 snapshot) → jsonb_set-neutralize offending fallback strings (coordinate with the
other chat — vonc's copy already lives in vonc's content_data per its cd_keys) → scoped F3 re-renders per affected
site + idea.uk pages. v1 restore-candidate shape query issued (is pre-07-01 site-neutral?).

R6f CONCRETE: consumed-but-undefined vars = --section-text, --section-text-muted, --section-surface,
--section-border, --spacing-section, --border-radius, --color-heading, --color-white, and --container-max-width
(styles.css defines --container-max — near-miss name). Sections on the NEW vocabulary fall back dark-on-dark;
old --color-* vocabulary sections stay visible. Structural fix = the styles.css generator must emit the templates'
vocabulary; generator-location greps issued. (leopardess unaffected via its old inline :root head.)

### 9ao. F8 CONFIRMED (gauntlet copy = schema static fallbacks); v1 not a restore candidate; harm-stop issued; R6f owner located

Query 1: live brief-explanation schema carries vonc copy VERBATIM as static fallbacks — cta_primary_label "Play
Today's Gauntlet", cta_secondary_label "See Past Results", stat_1..3 label/value "New Gauntlet Daily"/"24hrs",
"Players Scored"/"10K+", "Free to Play"/"100%" — the exact screenshot strings. F8 mechanism proven end-to-end:
static fallbacks → planSection use_fallback → ResolvedData → stored⊕resolved merge → dependents' content_data +
renders; name-guard blind to it. Query 3: v1 = OLD-architecture component (EMPTY fields contract, 8089B template,
no gauntlet) — matches how-it-works' site-blob cd — restore would regress the contract and orphan vonc/idea.uk's
new-style content ⇒ NEUTRALIZE-IN-PLACE chosen. Query 2 ambiguity owned: sites.content_data exists (resolver reads
it) — qualify pc.; answer already known via query 1, not reissued.

HARM-STOP issued (gated): Step 1 manual snapshot v2 (columns mirrored from the two WORKING insert paths —
snapshotComponentVersion + fixed update_component_html; some_schemas lacks the \d block; also backfills the
unversioned 13:22 write) → Step 2 neutralize 8 fields (stats → source=llm required=false NO fallback — writer
supplies per-site stats henceforth; cta labels → static neutral "Get Started"/"Learn More"; field NAMES/types
preserved — intentional source/required/fallback change, noted) with optimistic-lock WHERE updated_at =
13:22:44.359166 (UPDATE 0 ⇒ moved under us, stop+coordinate). COORDINATE with the other chat (their active area;
idea.uk mid-build). Step 3 map: vonc no-action (copy in own cd — correct place); idea.uk strip 8 keys ×2 pages +
F3-scoped rerender (coordinate-gated); robot-hands index strip 8 + scoped rerender (band will render near-empty →
visible-content filter drops it — honest interim), how-it-works + gripper-selection-guide → needs_page rebuilds
(old-blob/empty cd not re-keyable — different architecture era).

R6f owner LOCATED: webdesign-agent renders+commits styles.css (emit_design_items_action / check_integrity comments;
storage_actions writes content["styles.css"]); fix_harcoded_colours = post-pass precedent. Structural fix
direction: feed component-consumed var vocabulary into design generation, or a post-pass appending missing consumed
vars mapped to the palette. Design after R6e lands.

### 9ap. F8 Steps 1–2 LANDED (snapshot v2 + neutralize, optimistic lock held); Step 3 issued with the auto-escalation shortcut

Step 1 INSERT 0 1 (v2 snapshot, backfills the unversioned 13:22 write). Step 2 UPDATE 1 — the WHERE updated_at =
13:22:44.359166 lock held, so no concurrent change raced us. Decider re-run issued as verify. Step 3 issued:
3a backup CTAS page_components_bak_briefexp_20260703 (expect 3: idea.uk ×2 + robot-hands index) → strip the 8
merged keys via jsonb `- text[]` (UPDATE 3). 3b two scoped needs_rerender items (component_id resolved inline via
CROSS JOIN subselect — no manual UUID substitution; item_key f8_rerender_briefexp:<id>; dedup NOT EXISTS mirrored).
SEQUENCING INSIGHT: the scoped rerender covers ALL 3 robot-hands brief-explanation pages; gripper-selection-guide's
EMPTY content_data will trip rerender_page_sections' NULL pre-check → AUTO-ESCALATES that page to needs_page (the
Route A rebuild we wanted anyway) AND hands us a correctly-shaped needs_page item to CLONE for how-it-works —
avoiding a guessed spec (schema-before-SQL discipline; needs_page's spec shape unverified). Expected end state:
idea.uk pages = own llm copy + empty stats + neutral CTAs; robot-hands index + how-it-works bands render near-empty
→ dropped by the visible-content filter (honest interim); gripper-selection-guide → writer rebuild. Post-verify:
gauntletised=f everywhere except vonc (own content). Coordinate note repeated (idea.uk mid-build; all-sections
blast radius per F3 docs).

### 9aq. F8 Step 3 EXECUTED in full; watching the chain

Decider verify: stats source=llm no fallback, CTAs static "Get Started"/"Learn More" — neutralization confirmed in
the live schema. 3a: backup CTAS = 3 rows; strip UPDATE 3 (idea.uk ×2 + robot-hands index lost the 8 merged keys).
3b: INSERT 0 2 (scoped triggers for idea.uk + robot-hands, component_id resolved inline). Other chat informed.
Expected chain: idea.uk pages → own llm copy + empty stat row + neutral CTAs; robot-hands index + how-it-works →
band near-empty → dropped by the visible-content filter (honest interim); gripper-selection-guide → NULL pre-check
→ AUTO-ESCALATES to needs_page (writer rebuild; also the clone template for how-it-works' spec shape). Watch +
post-verify block issued (chain progress / gauntletised=f except vonc / the escalated needs_page shape). Remaining
after this settles: clone needs_page for how-it-works; R6 ledger's last eyeball (ai-agent case-study page); R6f
structural fix (webdesign styles.css vocabulary); flag list F4/F5/F6/F7-residual/F8 mitigations + hygiene items;
drop the two backup tables when comfortable.

### 9ar. Post-chain read (hours later): windows too narrow (owned); strip worked where it could; residual gauntlet = llm-copy suspicion

Queries 1+3 empty = MY window artifact (30min/1h written for run-now; user ran hours later) — rule applied to my
own SQL. Query 2 (the real read): robot-hands index (17:29) + how-it-works (17:37) re-rendered GAUNTLET-FREE at
identical 9181B = the empty-shell render; PREDICTION MISS owned: the band was NOT filter-dropped — the neutral CTA
fallbacks ("Get Started Learn More") are themselves >10 visible chars → both pages show a mostly-empty band with
two buttons (cosmetic interim). SURPRISES: (1) gripper-selection-guide re-rendered 18:04 WITH gauntlet at 9762B —
was supposed to auto-escalate on empty cd; a filled render implies the escalation fired and a WRITER rebuilt it
hours ago (outside query 3's window) — fresh llm copy contains the word: echo-contamination vs ordinary English on
a gripper-SELECTION page ("run the gauntlet") — context read issued. (2) idea.uk ×2 re-rendered post-strip
(17:30/18:12) STILL gauntletised — statics stripped + fallbacks neutral ⇒ the word must live in their LLM FIELDS
(heading/description/step_*), written at build time against the gauntlet-fallback schema = **F8 knock-on:
fallback contamination migrated into GENERATED copy**; strip can't touch it; remediation = writer content re-pass
(coordinate). (3) matchmatrix untouched (12:46 stamp) — scoped fan-out skipped it; candidate = pages.status outside
deployed/active filter; harmless (gauntlet-free, latent) but check. Issued A–D: wider-window chain, rendered-html
context, content_data field-carrier (THE decider), matchmatrix status.

### 9as. A–D attachment arrived empty (recurring extraction issue) — re-paste requested

The user sent "A-D results" as a document attachment; content extracted empty on my side (same as attachments
21/22/26 earlier). Asked for an inline re-paste (inline terminal text has worked every time). Reference restated:
A = chain since 17:00 (triggers/page items/needs_page); B = gauntlet context in renders; C = content_data
field-carrier (the decider); D = robot-hands pages.status (matchmatrix skip). No analysis possible this turn;
nothing else changed.

### 9at. output.txt = planner workflow tail, not A–D; two banked details

The uploaded output.txt contained the tail of a site-planner agent's default_config (validate_plan →
write_site_plan → reconcile_site_plan; load_existing_pages/load_components query steps; psql "(2 rows)" artifact) —
wrong buffer, A–D still outstanding. Banked from it: (1) reconcile_site_plan EMITS needs_page + terminal
needs_rerender items ⇒ the planner is ANOTHER item producer — query A needs w.created_by for attribution
(idea.uk items may be the other chat's pipeline, not ours); amended A issued. (2) load_existing_pages filters
p.status='active' ⇒ page-status gating is pervasive — supports the matchmatrix-skip hypothesis (D confirms).
B/C/D unchanged; C (content_data field-carrier) remains the decider for idea.uk's residual gauntlet.

### 9au. Second re-upload also not A–D (\d site_specs + a sample 'tools' aspect row); \o capture block issued

output1.txt = \d site_specs (already had) + SELECT * LIMIT 1 (leopardess site 4851f6fc, aspect='tools',
tool-suggester evaluation blob — confirms the site_id attribution and the per-aspect richness; not actionable).
Issued a single self-contained psql block using \o /tmp/abcd.txt … \o to capture A–D cleanly to one file for
upload — addresses the recurring wrong-buffer/empty-attachment problem. A amended with w.created_by (planner is
another item producer per 9at). Still blocked on A–D for: idea.uk residual-gauntlet decider (C), gripper-selection-
guide echo-vs-English (B), chain attribution (A), matchmatrix skip (D).

### 9av. Handoff rewritten (requested)

HANDOFF_component_regen_clobber.md rewritten wholesale as the cold-start entry point: operating model essentials
(incl. the \o capture gotcha + co-managed-chat coordination), incident 1 resolved+recovered summary with the R6
ledger, incident 2 (R6e/F8) full state with the PENDING A–D reads + decision tree as the immediate next action,
R6f mechanism + owner, the complete flag list (F4–F8) with current scoping, hygiene + cleanup items, key IDs, and
artifact/transcript pointers.

### 9aw. A–D read (finally): matchmatrix skip FALSIFIED (it errored); writer REPRODUCES vonc's pitch — third F8 carrier suspected (generation guidance)

A (with created_by): our 17:26 triggers → 17:27 fan-out included matchmatrix (D: all four pages 'active' — my
status-filter hypothesis DEAD); matchmatrix page_rerender completed WITH ERROR (text unfetched); idea.uk/tools
also completed-with-error yet its band re-rendered 18:12 = the other chat's build pipeline, not our chain.
Auto-escalations fired as designed: page-rerender created needs_page for matchmatrix (17:31) and gripper-selection-
guide (17:37); writer rebuilt gripper-selection-guide 18:04; matchmatrix's band still 12:46 → its rebuild
dispatched-but-unproven (SIXTH loose-semantics sighting: complete items carrying errors).

B+C (the decider): gauntlet lives in LLM FIELDS — and gripper-selection-guide's FRESH 18:04 rebuild carries the
IDENTICAL heading to idea.uk ("Every day a new <em>Gauntlet</em> begins") + "Today's Gauntlet Open" + "The Gauntlet
Scores the Room" + an illustration_alt describing a "Daily Gauntlet arena interface" — near-verbatim vonc pitch on
unrelated sites, generated AFTER the fallback neutralization ⇒ the writer is following GENERATION GUIDANCE baked
into the component: the llm fields' definitions (heading/description/badge_label/step_*) and the component's own
description/metadata were NOT touched by our neutralization (we replaced only the 8 static defs). Suspected THIRD
F8 carrier: gauntlet-flavoured field descriptions / component description from the 07-01 regen. Decider query
issued (jsonb_each field descriptions + component name/display_name/description) + the two error texts.
Remediation if confirmed: snapshot v3 → neutralize descriptions to site-agnostic wording → writer re-passes
(gripper-selection-guide re-escalates trivially; idea.uk coordinated). F8 picture then = three carriers: fallback
values (fixed), merged content (stripped), generation guidance (pending).

### 9ax. Descriptions carrier FALSIFIED (all empty); timeline explains idea.uk (pre-neutralization build); gripper-selection-guide 18:04 = the open vector; errors read

Query 1: every field description empty; component description = generic auto-string → third-carrier-in-
descriptions DEAD (owned). TIMELINE INSIGHT: idea.uk was written 16:27/16:36 — BEFORE the 17:26 neutralization —
its writer saw the then-live gauntlet fallbacks and wove them into llm copy ⇒ idea.uk needs NO new mechanism
(F8 knock-on, timestamped). gripper-selection-guide rebuilt 18:04 AFTER neutralization with the IDENTICAL heading
⇒ its writer context carried gauntlet from something that survived the fix. Candidates: (a) residual blob elsewhere
in input_schema (only ->'fields' ever inspected — top-level example/sample keys unchecked), (b) writer few-shots on
ANOTHER SITE'S content for the same component (cross-site exemplar leak — would explain identical copy), (c) stale
plan directives. Sweep issued (top-level keys + whole-schema strpos) + writer-agent identification + repo grep for
the section-content action.

Errors: matchmatrix child took the ESCALATE branch (empty cd → needs_page 17:31; no render intended — band's 12:46
stamp consistent) and failed only at complete_workflow: parent job topic gone ("topic partition not found",
job.9ffeb740-…-build-dispatch-loop-spawn_dispatch.res) — fire-and-forget parent topic lifecycle pollutes child
completions; parked (7th loose-semantics-adjacent sighting). Its needs_page 'complete' = dispatch stamp; WRITER
REBUILD FOR MATCHMATRIX UNPROVEN (band still 12:46) — verify + re-triage if untouched. idea.uk/tools: "Claim timed
out — handler pod likely died" 17:27; the 18:12 render was a later pass (their pipeline). Remediation queue: sweep
→ fix exposed vector (writer code read if sweep clean) → writer re-passes gripper-selection-guide + idea.uk
(coordinate) → verify/re-triage matchmatrix.

### 9ay. Copy-attribution clarified for the user; writer prompt lives in agent config (grep found only planners/loaders)

Clarified: C showed idea.uk carrying GAUNTLET (vonc) copy only — no gripper/robot-hands copy observed on idea.uk;
the gripper association was robot-hands' own gripper-selection-guide page RECEIVING gauntlet copy (18:04 rebuild).
User's cross-site-leak instinct is the working theory though — issued the double-leak test (idea.uk content_data
LIKE %gripper%/%robot%) + full jsonb_pretty of idea.uk index's section so the user can see the copy. Their
"corrupted shared component" framing = right family; refinement: the carrier was the schema FALLBACK VALUES
(fixed 17:26), which fully explains idea.uk (built 16:27/16:36 PRE-fix); the open vector remains gripper-selection-
guide's POST-fix 18:04 rebuild. Repo grep found plan_sections/get_pages_to_build/load_page_record — no generator
action ⇒ the writer's prompt lives in the WRITER AGENT'S default_config (platform pattern, cf. F1prompt
migrations) — issued: writer identification query + \o config dump + upload. Sweep from 9ax STILL OUTSTANDING
(re-issued): input_schema top-level keys + whole-schema gauntlet strpos — if clean, the 18:04 reproduction must
come from the writer's own context assembly (exemplar leak prime suspect).

### 9az. Writer prompt read: fallbacks NEVER enter the prompt (idea.uk story corrected); content_brief = prime carrier; schema still mentions gauntlet in an uninspected field attr

page-content-writer default_config (uploaded, read; image v1.0.1094 — FLEET BUMPED 2026-07-04 19:21 by parallel
chats, cached workflow knowledge now suspect): generate_content prompt includes per-site context (render_context,
site_specs content_direction/identity), link context, research, existing_content (adopt mode), and llm_field_specs
as NAME/TYPE/REQUIRED/DESCRIPTION ONLY — no fallbacks, no examples ⇒ my "idea.uk's writer saw the fallbacks" story
CORRECTED: one carrier must explain BOTH idea.uk (16:27) and gripper-selection-guide (18:04). The prompt's
"Admin Content Brief (follow these instructions closely)" block = {{.current_section.component.content_brief}}
(purpose/tone_direction/section_guidance) — a per-COMPONENT steering attribute NEVER inspected = PRIME SUSPECT.
Plus the sweep's schema_mentions_gauntlet=t with all checked attrs clean ⇒ some per-field attribute inside
'fields' still carries the word. Issued: (a) LATERAL jsonb_each field×attr locator LIKE %gauntlet%; (b)
information_schema check for a content_brief column + jsonb_pretty dump. Remediation once located: snapshot v3 →
neutralize to site-agnostic wording → writer re-passes (re-escalate gripper-selection-guide; coordinate idea.uk;
verify matchmatrix — band still 12:46). F8 mitigation list grows: site-neutrality applies to content_brief as much
as fallbacks (a by-design contamination surface on shared components). Also noted: the writer prompt's
anti-fabrication rules (never invent stats etc.) align with our stats→llm flip.

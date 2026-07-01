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

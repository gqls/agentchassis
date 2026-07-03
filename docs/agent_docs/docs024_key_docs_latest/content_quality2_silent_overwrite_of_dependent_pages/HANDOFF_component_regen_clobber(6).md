# HANDOFF — component regeneration silently empties dependent pages

> **STATUS 2026-07-03: RESOLVED + RECOVERED.** Root cause fixed (F1 guard + prompt companions + F3 scoped rerender, all deployed and proven live); the five broken pages re-keyed and re-rendered (R6b pass — distinct md5s, content + CTA present). Remaining: live-deploy eyeball (runbook R6) and flagged follow-ups F4/F5/F6. The RUNBOOK + NOTES in this folder are the authoritative record.

For a fresh chat picking this up cold. It is a durable, cross-site platform bug, flagged but not yet fixed.
Read this, then generate the context bundle with `BUNDLE_component_regen_clobber.md`, then work the
`PLAN_component_regen_clobber.md` using `RUNBOOK_component_regen_clobber.md` for the SQL/ops.

---

## 1. Platform model (the minimum to orient)

Agents are rows in Postgres (`clients_db`, table `agent_definitions`) run as Kubernetes pods (namespace
`ai-persona-system`), talking over Kafka. A page is assembled from sections; each section on a page is a row
in `page_components` that points at a shared template in `content_components`. The same `content_components`
row is reused by many pages across many sites. Rendering binds a page's `content_data` into the component's
`html_template`. DB access (a human runs all SQL/kubectl):

```
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db
```

Two change types: editing an agent's `default_config` jsonb is a live workflow change (no redeploy); a Go
change goes through GitHub Actions → Backblaze → a new chassis image → bump `image_tag` on the agent rows.

---

## 2. The bug

Shared content components are referenced by many pages across multiple sites (one `content_components` row,
N `page_components` instances); at render time `RenderTemplate` (component_library.go) binds a page's stored
`content_data` into the component's `html_template` by exact field name using Go's `text/template`, and any
placeholder with no matching key renders to nothing — the function strips the resulting `<no value>` tokens to
empty string without raising an error. When `component-creator` regenerates a component it can rewrite the
template/schema field names (observed 2026-06-24 15:06 on the `system-stats` component `fdd92ad4`:
`stat_1_number`→`stat1_value`, `eyebrow`→`eyebrow_label`, `heading`→`section_headline`, `footnote`→`footnote_text`,
and so on) without migrating the dependent pages' `content_data`, which still carries the old key names. The
write itself is clean — `update_component_html` only snapshots the old `html_template`, swaps in the new one, and
marks every dependent page `build_status='pending'` (that single `UPDATE … SET updated_at=NOW()` is what stamped
all five instances at the same 15:06:12.956 instant; it deliberately does not touch `rendered_html`). The damage
lands when a render step subsequently rebuilds a pending page from its old-key `content_data` against the renamed
template: every renamed placeholder misses, the section renders blank, and the assembler's visible-content filter
silently drops the band — fanning out to every page that shares the component. The exact render step that wrote
the blanks is not yet pinned (the no-LLM `rerender_page_sections` carries the stored HTML when a section isn't
"ready" and is gated to image/section-data reasons, so the prime suspects are `component-creator`'s own dependent
re-render and whatever consumes `build_status='pending'`); confirming it is the first job. The `content_data` is
intact, only mis-keyed, so the breakage is recoverable.

**Why it is being handed off rather than fixed in the originating thread:** the flaw is in shared
component-regeneration machinery, the 2026-06-24 regen was a manual `component-creator:regen` (likely driven
by another chat that co-manages components), and no `gamesdesign.co.uk` page is affected — its `system-stats`
instance was dropped during an unrelated rebuild, so the originating thread had nothing to fix on its own site.

---

## 3. Evidence and IDs

- **Component:** `content_components` id `fdd92ad4-521a-4602-89cf-7ee1a06...` (`system-stats`; use the full UUID
  from a query — see RUNBOOK R1). Relevant columns: `html_template`, `input_schema`, `template_variable_count`,
  `schema_field_count`, `schema_template_synced`, `usage_count` (**stale — do not trust it as a live count**).
- **Version history:** `component_versions` has exactly one row for this component (`version_number` 1,
  `changed_by` = `component-creator:regen`, `change_source` = `manual`, created 2026-06-24 15:06:12.923611).
  There is **no pre-rename version to revert to** — the only stored version is already the renamed one.
- **Five affected instances** (all `component_id` = `fdd92ad4`, all `build_status` = `pending`,
  `component_version_id` = NULL, all `rendered_html` = **7369 bytes** — identical, not zero). Their `updated_at`
  of 15:06:12.956367 is the bulk `build_status='pending'` UPDATE from `update_component_html` (one statement, one
  `NOW()`) — the empty render, if any, was a **separate** write (§4 / RUNBOOK R3a), so don't read that one stamp
  as the moment of emptying. Resolved pages (three sites): `index` on `leopardessconsulting.co.uk`,
  `robot-hands.com`, `ai-agent-orchestration.com`; `case-study-kafka-consumer-group-remediation` on
  `ai-agent-orchestration.com`; `gripper-detail` on `robot-hands.com`. Identical byte-length across different sites
  suggests per-page content isn't landing, but whether they are genuinely blank-by-rename is **unconfirmed** until
  RUNBOOK R3 reads the bytes + keys. Drive enumeration off `component_id`.
- **Old keys (in the pages' `content_data`):** `eyebrow`, `heading`, `subheading`, `footnote`,
  `stat_N_number`, `stat_N_label`, `stat_N_suffix`, `stat_N_description`, `stat_N_icon`.
- **New keys (in the regenerated template/schema):** `eyebrow_label`, `section_headline`, `section_intro`,
  `footnote_text`, `statN_value`, `statN_suffix`, `statN_label`, `statN_description`, `cta_url`, `cta_label`.
  Note the mapping is mostly 1:1 but not entirely — `stat_N_icon` has no new target, and `cta_url`/`cta_label`
  are new with no old source. **Derive the real mapping from the live `input_schema`, do not trust this list
  from memory** (RUNBOOK R1/R3).

---

## 4. Where it lives in the code

Confirmed from the actual sources (`update_component_html_action.go`, `rerender_page_sections_action.go`,
`component_library.go`):

- `platform/orchestration/actions/component_library.go` — `RenderTemplate(...)` binds by exact field name (Go
  `text/template`) and then strips the `<no value>` tokens that unmatched `{{.field}}` placeholders produce down
  to empty string, logging only a warning. This is the silent-empty mechanism: a renamed field → empty fill, no
  error. `RenderContext` carries `ContentData map[string]interface{}` (the page's values); `GetComponentByID` loads
  the component.
- `update_component_html_action.go` → `UpdateComponentHTMLAction` — the regen write. It snapshots the old
  `html_template` to `component_versions` (best-effort), swaps in the new template, and marks every dependent
  `build_status='pending'` with `updated_at=NOW()`. It explicitly does **not** touch `rendered_html` (the comment
  says the rerender pipeline regenerates it). So it is **not** the clobber, and that pending-flag UPDATE is the
  source of the synchronized 15:06:12.956 timestamp.
- `rerender_page_sections_action.go` → `RerenderPageSectionsAction` — the no-LLM re-render. It loads stored
  `content_data`, rebuilds resolved fields via `planSection`, and renders the current template; but it **carries
  the stored HTML unchanged** when a section isn't "ready" / the template is empty / the component is missing, and
  the workflow gates it to `reason ∈ {image_landed, section_data_resolved}`. So it is partly protective and not the
  obvious writer of the blanks — though it *could* empty a section if it ran with a "ready" plan against the
  renamed template.
- `v3_site_actions.go` → `RenderComponentAction` — the **writer's** render. It renders `comp.HTMLTemplate`
  against a `RenderContext` built from the supplied `content_field`/`content_from` ⊕ `merge_with` (fresh content,
  not stored `content_data`) and **returns** `{rendered_html, content_data, …}` — it does **not** write the DB or
  set `updated_at`; a persist step (`save_page_sections`) does the write. So it is not the stored-content
  re-render path.
- **`component-creator` agent (config, v1.0.1080)** — workflow is `ensure_site_record → read_site_spec →
  generate_template → store_component (store_generated_component) → complete`, with **no** `update_component_html`
  step and **no** re-render-of-dependents step. So component-creator does not re-render dependents, and the version
  row tagged `component-creator:regen` / `manual` points to an out-of-band/manual regen.
- `store_generated_component_action.go` → `StoreGeneratedComponentAction` — CONFIRMED **rename writer**. On
  regeneration (a component with the same `function` exists) it snapshots the old `html_template`/`input_schema`
  to `component_versions` (`changed_by='component-creator:regen'`), `UPDATE`s `content_components` **in place**
  (same `component_id`, so dependents follow the renamed contract), marks dependents `build_status='pending'`, and
  raises one `needs_rerender` per site. It does **not** migrate dependents' `content_data`, and the `needs_rerender`
  it raises (`{component_id, function, refresh_site_components:false}`, no `reason`, handler `rerender-pages`) drives
  an **assemble-only** re-render — so the regen never re-renders the sections and never wrote these blanks. The
  `.956` stamp is `markPagesPendingRebuild`'s bulk UPDATE; the empty render predates the regen.
- **RESOLVED — it was the 15:06 regeneration rename (RUNBOOK R3a).** The pre-15:06 `component_versions` v1 schema
  is the **old** key set and matches the dependents' `content_data` field-for-field; the live schema is the **new**
  set. So the regen renamed the shared fields old→new and didn't migrate `content_data`. (Not an out-of-band edit,
  not a standing writer-vs-schema mismatch.) Because `markPagesPendingRebuild` doesn't touch `rendered_html` and the
  triggered rerender is assemble-only, the stored bytes are the preserved pre-15:06 (good) render — so the five may
  be **latently** broken (still showing the old render until the next real re-render) rather than visibly empty;
  one content-presence query (R3a sub-check) confirms which. The fix and the future risk are the same either way.
- `platform/orchestration/actions/component_selector.go` — `IncrementUsageCount` (usage_count bump; `fdd92ad4`'s
  is 0, so it hasn't run), `SelectComponentByType`, and the "empty `input_schema` → needs regeneration" trigger.
- Related (not focal): `save_page_sections_action.go` writes per-page `content_data`/`rendered_html` during builds
  and carries the earlier Part 4 interactive-section guard.

---

## 5. Suggested routes (directions to weigh, not a fixed choice)

- Preserve a component's existing field names across a regeneration, so bindings that already work keep working.
- When a regeneration does rename fields, migrate each dependent page's `content_data` keys old→new as part of
  (or immediately before) the re-render, so the bind still matches.
- Add a guard that treats a re-render which empties a previously-populated section as a failure to surface
  (log loud / mark the item failed) rather than a silent blank to ship — so any future regression is visible.

These are not mutually exclusive; the plan treats them as candidate directions to assess against the code, not
an either/or.

---

## 6. Recovery for the five already-broken pages

`content_data` is intact, just mis-keyed, so recovery is cheap and needs no LLM: align the keys to the current
`input_schema` (per-page `content_data` migration old→new), then trigger the no-LLM `rerender_page_sections`
path for those pages (a `page_rerender` work item per page; the `rerender-pages` agent / `build-dispatch-loop`
/ `page-rerender` chain already exists — reuse it, do not build a new one). Verify the bands reappear (RUNBOOK R5/R6).

---

## 7. Standing rules and cautions

- Reuse and adapt existing functions/structs/agents before creating anything new. Keep workflows thin and put
  logic in Go. No SQL sub-workflows — spawn sub-agents with their own workflows if orchestration is needed.
  Every agent is an orchestrator owning a workflow of steps that call actions.
- Check the live schema before writing SQL. Do not trust a 0-row / empty result until the query itself is
  verified. Do not jump to conclusions. Keep `logger.Debug` out (it does not show in the logs).
- **Concurrency caution:** another chat co-manages components. Before any write to shared component state,
  re-check freshness (RUNBOOK R7) — no blind writes. The 2026-06-24 regen was `change_source = manual`,
  consistent with that other chat.
- Tone: pragmatic, no banned words; no summary documents unless asked.

---

## 8. Pointers

- `BUNDLE_component_regen_clobber.md` — the `cmd/bundle` command + paths to confirm.
- `PLAN_component_regen_clobber.md` — phased plan.
- `RUNBOOK_component_regen_clobber.md` — the SQL and recovery steps.
- Background only (do not re-derive from these): the prior gamesdesign threads are in the originating chat's
  `HANDOFF_page_pipeline.md` / `NOTES_gamesdesign_silent_norebuild.md`; the prior bundle outputs
  `bundle_gamesdesign.md` and `bundle_gamesdesign_generation.md` hold the relevant code signatures and schema.

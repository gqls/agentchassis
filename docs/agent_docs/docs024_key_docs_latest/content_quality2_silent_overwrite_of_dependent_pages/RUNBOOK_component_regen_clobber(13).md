# RUNBOOK — component regeneration silently empties dependent pages

## The problem (start here)

A shared section component (one `content_components` row) is reused by many pages across several sites, and
each page stores its own `content_data` whose keys must match the component template's `{{.field}}` placeholders
exactly — `RenderTemplate` (component_library.go) fills a placeholder only when `content_data` has a key of the
same name, and silently leaves it empty otherwise (it runs Go `text/template` and then strips the resulting
`<no value>` tokens to empty string, logging only a warning, no error). For the `system-stats` component
(`fdd92ad4`), R3 confirmed the five live instances render an **identical content-free shell** (one md5, 7369
bytes each) because their `content_data` holds **old** key names (`eyebrow`, `heading`, `stat_1_number`, …) that
the current template no longer reads. The shared component itself is healthy (22 fields, template/schema synced).

When component-creator regenerates such a component, `StoreGeneratedComponentAction` overwrites its
`html_template` + `input_schema` **in place** (same `component_id`, so every dependent now points at the new
contract), snapshots the old version, marks dependents `build_status='pending'`, and raises a `needs_rerender`
work item per site — but it does **not** migrate the dependents' `content_data`, and that `needs_rerender` carries
no `reason`, so the rerender it triggers is **assemble-only** (re-ships the existing `rendered_html`) and never
re-renders the sections. R3a confirmed this is exactly what happened on 2026-06-24 15:06: the pre-regen schema
snapshot is the **old** key set (matching `content_data` field-for-field) and the live schema is the **new** set,
so the regen renamed the fields and left `content_data` unmigrated. (It was a regeneration rename — not an
out-of-band edit or a standing writer-vs-schema mismatch.)

Because `markPagesPendingRebuild` only sets `build_status`/`updated_at` (the `15:06:12.956` stamp) and the
triggered rerender is assemble-only, the rows were not repaired in place. The content-presence check (R3a)
confirmed the current state: `content_data` is **per-page** (distinct headings/stats) yet the stored render
contains none of it and all five are byte-identical — so the rows hold the **new-template render with empty
content slots**. The five are **live-broken** (empty system-stats band), not latent. gamesdesign was a sixth,
since dropped. `content_data` is intact and per-page, so recovery is a key re-map plus a section re-render that
**must carry `reason=section_data_resolved`** (a bare `needs_rerender` is assemble-only and just re-ships the
shell). The remaining steps fix the cause permanently, test it, recover the five, and verify.

## Where we are now (2026-06-26)

**Diagnosis is complete and the root is confirmed.** The 15:06 regeneration renamed the shared `system-stats`
field contract and didn't migrate dependents' `content_data`; the five live instances are rendering an empty
system-stats band right now. Investigation steps R0–R3a are done. The **F1 guard** is written and `gofmt`-clean
(`F1_store_generated_component_action.patch`, guard-only) and is safe to ship on its own — it converts the silent
stranding into a loud rejection. What's left: ship the guard (F1) + its prompt companion, test it (F2), wire the
section-rerender reason once we have the `create_rerender_items` action code (F3 — the reason is dropped today
because `rerender-pages`' `create_rerender_items` wires only domain/site_id/pages_field), then recover the five
(R4/R5/R6). Recovery does **not** route through `rerender-pages` for the section re-render (it drops the reason):
use a full `needs_page` rebuild, or re-key + a directly-inserted `page_rerender` carrying the reason.

## Progress at a glance

Tick these as you go (edit the file: `[ ]` → `[x]`).

- [x] **R0** — Orient (read this + the handoff)
- [x] **R1** — Inspect the component (schema fields, template placeholders, version history)
- [x] **R2** — Enumerate affected pages + resolve site domains
- [x] **R3** — Confirm the bind failure on one page (keys don't intersect)
- [x] **R3a** — When/why the binding broke — RESOLVED: 15:06 regen rename, `content_data` not migrated; five are **live-broken** (content-presence check)
- [ ] **F1** — Permanent fix: field-contract guard in `StoreGeneratedComponentAction` (patch is guard-only, `gofmt` clean)
- [ ] **F1-prompt 1a** — `load_existing_component_action.go` (new advisory Go action, gofmt-clean) + register in registry.go; deploy BEFORE the migration
- [ ] **F1-prompt 1b+2** — apply `F1prompt_component_creator_preserve_field_names.sql` (wires the load step, appends `existing_component` to input_fields, inserts the dormant prompt rule; snapshot-first, drift-checked)
- [ ] **F2** — Test: Tier 1 unit test (`store_generated_component_guard_test.go`, deterministic) + Tier 2 integration (reject path, test DB) + Tier 3 end-to-end throwaway `zzz-fieldguard-test` (names preserved, no stranding)
- [ ] **F3** — Scoped reason propagation: F3a `create_rerender_items` (scope to dependents + stamp reason) + F3b re-add reason to store spec + F3c `rerender-pages` config (`gofmt` clean; deploy after F1)
- [ ] **R4** — Back up before any write
- [ ] **R5** — Recover the five broken pages (full `needs_page` rebuild, or re-key `content_data` + a **directly-inserted** `page_rerender` with `reason=section_data_resolved`)
- [ ] **R6** — Verify (renders contain content, bands appear, `content_hash` changed)
- [ ] **R7** — Freshness / concurrency check (run immediately before any write)

## Before you start

A human runs every SQL/kubectl command. Inspect before asserting; treat any 0-row / empty result as suspect
until the query itself is verified. Confirm flagged column/table names against the live schema (`\d <table>`)
before relying on them — the bundle's `-schema-tables` dump covers
`content_components, component_versions, page_components, pages, sites`.

DB access:

```
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db
```

Throughout, `<COMPONENT_UUID>` is the full id of `system-stats` (get it in R1), and `<PAGE_UUID>` is one
affected page's id (get it in R2).

---

## Step R0 — Orient

The bug, IDs, key lists, and code locations are in `HANDOFF_component_regen_clobber.md` §2–§4. Do R1–R3a
read-only first; do not write until R4 (backup) and R7 (freshness) are done.

---

## Step R1 — Inspect the component (read-only)

Get the full id and the headline counters (`usage_count` is stale — ignore it as a live count):

```sql
SELECT id, name, function, schema_field_count, template_variable_count,
       schema_template_synced, usage_count, updated_at
FROM content_components
WHERE name = 'system-stats';
```

Read the current schema and template, and extract the real field/placeholder names — do not trust the
from-memory list in the handoff:

```sql
SELECT jsonb_pretty(input_schema) FROM content_components WHERE id = '<COMPONENT_UUID>';

-- Placeholder names actually used in the template. RenderTemplate uses Go text/template, so these are {{.field}}
-- / {{ .field }} (and {{if}}/{{range}}/{{with}}); confirm by eyeballing the template, then adjust the regexp.
SELECT DISTINCT m[1] AS placeholder
FROM content_components,
     regexp_matches(html_template, '\{\{[-\s]*\.?([a-zA-Z0-9_]+)', 'g') AS m
WHERE id = '<COMPONENT_UUID>'
ORDER BY 1;
```

Record the definitive current field set. The known-but-unverified rename map (top-level mostly 1:1; note the
exceptions): `eyebrow`→`eyebrow_label`, `heading`→`section_headline`, `subheading`→`section_intro`,
`footnote`→`footnote_text`, `stat_N_number`→`statN_value`, `stat_N_suffix`→`statN_suffix`,
`stat_N_label`→`statN_label`, `stat_N_description`→`statN_description`; `stat_N_icon` has no new target;
`cta_url`/`cta_label` are new with no old source. Validate this against what R1 actually returns.

Check the version history (expect exactly one row, already the renamed version — no clean revert target):

```sql
SELECT version_number, change_source, changed_by, created_at, change_description
FROM component_versions
WHERE component_id = '<COMPONENT_UUID>'
ORDER BY version_number;
```

---

## Step R2 — Enumerate affected pages and resolve site domains (read-only)

This is also the query for choosing the bundle's `-runtime-site` / `-runtime-page`. Drive off `component_id`, not
the page-id prefixes. Confirm the `sites` table and its domain column with `\d sites` first (the join below
mirrors `rerender_page_sections_action.go`, which uses `sites.domain`, so `s.domain` should be correct):

```sql
SELECT pc.id, pc.page_id, p.name AS page_name, p.site_id, s.domain,
       pc.build_status, pc.component_version_id,
       length(pc.rendered_html) AS html_len, pc.updated_at
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
WHERE pc.component_id = '<COMPONENT_UUID>'
ORDER BY pc.updated_at;
```

Confirmed result (2026-06-26): five rows on **three** sites — `index` on `leopardessconsulting.co.uk`,
`robot-hands.com`, and `ai-agent-orchestration.com`; `case-study-kafka-consumer-group-remediation` on
`ai-agent-orchestration.com`; `gripper-detail` on `robot-hands.com`. All five had `html_len = 7369` (identical to
the byte) and the same `updated_at` `15:06:12.956367` — the pending-flag UPDATE (see R3a). Identical byte-length
across different sites means the per-page content isn't landing in the render, which is the tell for the
blank-shell story — but R3 confirms it on the actual bytes. The bundle's runtime is wired to
`robot-hands.com` / `gripper-detail`. (`gamesdesign.co.uk` does NOT appear here — its instance was dropped — so it
is not a useful runtime target.)

---

## Step R3 — Decide whether the band is actually blank, and whether the keys mismatch (read-only)

Do not infer "blank" from `html_len` alone (7369 ≠ 0). Read the actual bytes and keys. (R1 returned the component
UUID `fdd92ad4-521a-4602-89cf-7ee1a66c10f1`; `<PAGE_UUID>` resolves from R2 — e.g. `robot-hands.com` /
`gripper-detail`.)

```sql
-- (a) Are the 5 renders identical bytes (=> no per-page content), and how much visible text remains?
SELECT md5(rendered_html) AS html_md5, length(rendered_html) AS html_len,
       length(regexp_replace(rendered_html, '<[^>]+>', '', 'g')) AS approx_visible_len,
       count(*) AS pages
FROM page_components
WHERE component_id = '<COMPONENT_UUID>'
GROUP BY 1, 2, 3 ORDER BY pages DESC;

-- (b) Current schema field names (inspect jsonb_pretty(input_schema) first if it nests under a 'fields' key):
SELECT jsonb_object_keys(input_schema) FROM content_components WHERE id = '<COMPONENT_UUID>' ORDER BY 1;

-- (c) One affected page's stored content_data keys — do they match (b), or are they the old names?
SELECT jsonb_object_keys(content_data)
FROM page_components
WHERE page_id = '<PAGE_UUID>' AND component_id = '<COMPONENT_UUID>'
ORDER BY 1;

-- (d) Eyeball the render head — real stat values/text, or empty slots?
SELECT substring(rendered_html from 1 for 1200) AS head
FROM page_components
WHERE page_id = '<PAGE_UUID>' AND component_id = '<COMPONENT_UUID>';
```

Read: if (a) is one md5 with `pages=5` and `approx_visible_len` ≈ 0, and (c)'s keys do not intersect (b), the
blank story holds and the recovery in R5 applies. If (a) shows substantial visible text, or (c)'s keys
match (b), the five are **not** broken by this — only the dropped gamesdesign band was — and the live impact is
smaller; re-scope before any recovery. Either way, confirm real content exists under whatever keys `content_data`
uses, so recovery has something to re-key.

Confirmed result (2026-06-26): (a) one md5, `pages=5`, `html_len=7369` — all five byte-identical (no per-page
content lands). (c) `content_data` holds the **old** keys (`eyebrow`, `heading`, `subheading`, `footnote`,
`stat_N_number`/`stat_N_suffix`/`stat_N_label`/`stat_N_description`/`stat_N_icon`). Two caveats from the run:
`approx_visible_len` came back `4999`, but the render opens with a large inline `<style>` block and this regex
strips only tags, not stylesheet contents — so that number is mostly CSS, not real content; and `input_schema`
nests fields under a `fields` key, so the field names are `input_schema->'fields'` (R3a uses that). Net: the five
render an identical content-free shell because `content_data`'s old keys aren't read by the current template. Why
the keys diverged is R3a.

---

## Step R3a — When/why the binding broke — RESOLVED (regen rename), with one current-state sub-check

Resolved by the snapshot-vs-live schema diff: the pre-15:06 `component_versions` v1 schema is the **old** key set
(`eyebrow`, `heading`, `subheading`, `footnote`, `stat_N_number`/`stat_N_icon`/…, 24 fields) and it matches the
dependents' `content_data` field-for-field; the live schema is the **new** set (`eyebrow_label`,
`section_headline`, `statN_value`, `cta_url`/`cta_label`, no `stat_N_icon`, 22 fields). So the 15:06
`StoreGeneratedComponentAction` regeneration renamed the shared component's fields old→new and did **not** migrate
the dependents' `content_data`. This is the root — not a writer-vs-schema standing mismatch, not an out-of-band
edit.

`markPagesPendingRebuild` only sets `build_status='pending'` + `updated_at=NOW()` (no `rendered_html`), and the
regen's triggered re-render is assemble-only — so the bytes in those rows are the **pre-15:06** render, preserved.
Because the keys matched before 15:06, that render was good. So one current-state question remains: are the five
still showing the old good render (breakage **latent** until the next real re-render) or an empty one? Settle it:

```sql
-- (1) Is content_data per-page or generic across the five?
SELECT s.domain, p.name, pc.content_data->>'heading' AS heading,
       pc.content_data->>'stat_1_number' AS stat_1_number, pc.content_data->>'stat_1_label' AS stat_1_label
FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
WHERE pc.component_id = '<COMPONENT_UUID>' ORDER BY s.domain, p.name;

-- (2) Does one page's stored render actually contain its content_data values?
SELECT strpos(rendered_html, coalesce(nullif(content_data->>'stat_1_number',''), chr(1))) > 0 AS render_has_stat1,
       strpos(rendered_html, coalesce(nullif(content_data->>'heading',''),       chr(1))) > 0 AS render_has_heading
FROM page_components
WHERE page_id = '<PAGE_UUID>' AND component_id = '<COMPONENT_UUID>';
```

`render_has_*` TRUE → rows still hold the good render; breakage is **latent**. FALSE → **empty now**; the five are
live-broken and need recovery. The fix and the future risk are identical either way.

Confirmed result (2026-06-26): `content_data` is **per-page** (distinct — `70`/"Deployed Agents" on
ai-agent-orchestration.com and leopardessconsulting.co.uk; `2400`/"Gripper Models Indexed" on robot-hands.com),
and gripper-detail's render contains **neither** its `stat_1_number` nor its `heading` (`render_has_*` both
false). Per-page content + byte-identical empty renders ⇒ the rows hold the **new-template render with empty
content slots**. The five are **live-broken**, not latent. (The precise write that produced the empty render isn't
pinned to a timestamp — the `.956` stamp is the pending-flag — but it is definitively the new-template empty
render.) Proceed to the fix (F1) and recovery (R5).

Fix target (confirmed): `StoreGeneratedComponentAction`'s regeneration branch renames the shared field contract
without migrating dependents' `content_data`, and the `needs_rerender` it raises is assemble-only. It holds
`existingSchema` + `inputSchemaJSON` (~L360–398) — the natural point to migrate dependents' keys old→new (or to
constrain regeneration to preserve names). Any recovery/fix re-render must carry `reason=section_data_resolved`
(assemble-only won't re-render sections). Done.

---

## Step F1 — Fix the cause permanently (Go change in `StoreGeneratedComponentAction`)

The cause is in the regeneration branch of `store_generated_component_action.go`: the new field names come from
`parseGeneratedTemplate` (i.e. whatever the LLM emitted that run), and the regen branch overwrites the shared
component's `input_schema`/`html_template` field contract in place without migrating dependents' `content_data`.
F1 is the **guard** — the smallest complete, independently-shippable fix. The patch in this folder
(`F1_store_generated_component_action.patch`, `gofmt -e` clean) is **guard-only**, three hunks:

1. **Field-contract guard (the fix) — reject a regeneration that renames/removes a retained field.** Added to the
   existing Layer-1 validation block (after the `<no value>` check), reusing the `blockingIssues` +
   `recordValidationRejection` reject path that's already there: on `isRegeneration`, diff the old schema's field
   names (`existingSchema`) against the new ones (`schemaJSONStr`); any retained field that disappears is appended
   as a blocking issue, so the overwrite is rejected loudly (and logged to `agent_error_log`) instead of silently
   stranding dependents. New fields are allowed (additions don't trigger it).
2. New helper `schemaFieldSet` (mirrors `deriveRenderMode`'s `fields` parse), appended at EOF with its own doc
   comment; and a `sort` import. No existing variable or function names changed.

Comparing old-schema↔new-schema is a proxy for "what dependents have" (content_data is written to match the
schema); a comment notes the exact alternative (query dependents' `content_data` keys) if ever needed.

**Companion change (not this file) — preserve names at generation.** The guard is a backstop: with it alone,
regenerations that rename fields now *fail* rather than break dependents. To make them *succeed* while keeping the
contract, component-creator's `generate_template` step must pass the existing component's field names into the
prompt and require their reuse (restyle/restructure freely; add new optional fields if needed; do not rename or
drop existing ones). Pair this with the guard.

**Deferred to F3 (not in this patch) — the rerender reason.** Adding `reason=section_data_resolved` to
`createRerenderWorkItem`'s spec was authored then pulled, because `rerender-pages`' `create_rerender_items` step
wires only `domain`/`site_id`/`pages_field` and so **drops `spec.reason`** — the reason never reaches the per-page
`page_rerender` items, and `page-rerender` stays assemble-only. The reason is inert until that propagation exists,
so it lands with its companion in F3 (below), not here.

**Alternative (only if deliberate field-set changes to shared components are a real requirement).** Replace the
guard's "reject" with an explicit migration: the regen emits an old→new field map and the action migrates every
dependent's `content_data` keys (it holds `existingSchema` + `inputSchemaJSON`) before the overwrite, handling
non-1:1 cases. Heavier/riskier; gate behind an explicit flag rather than letting the LLM trigger it.

**Note on current behaviour.** This guard will reject the exact regen that hit `fdd92ad4` (it dropped
`stat_N_icon` and renamed fields) — the intended fail-safe. Intentional schema changes now go through a deliberate
path, not an LLM side effect. The guard does **not** change the regen's existing assemble-only re-render, so a
names-preserving restyle still re-deploys dependents with stale section markup until F3 (or a rebuild) refreshes
them — stale, not broken.

Apply + ship: `git apply F1_store_generated_component_action.patch` (from repo root) → `go build ./...` → commit →
GitHub Actions → Backblaze → new chassis image → bump `image_tag` on the affected agent rows. Snapshot any agent
whose config/image changes (`snapshot_agent`). Safe to ship alone — it only adds a rejection.

## Step F3 — Propagate the section-rerender reason, scoped to the component's dependents

So a legitimate restyle regen (and recovery) drives an actual section re-render of the changed component's
dependents rather than an assemble-only re-ship — without turning every regen into a whole-site re-render (which
would re-render unrelated sections and could surface *other* latent mismatches). Three coupled changes, deployed
**after** F1:

- **F3a — `create_rerender_items_action.go`** (`F3a_create_rerender_items_action.patch`, `gofmt -e` clean): add
  `reason` + `component_id` to the input spec; when the reason is a section-rerender reason
  (`section_data_resolved`/`image_landed`) **and** a `component_id` is present, resolve the component's dependent
  pages (`SELECT pc.page_id FROM page_components pc JOIN pages p ON p.id=pc.page_id WHERE pc.component_id=$1 AND
  p.site_id=$2`), create work items **only** for those pages, and stamp `reason` onto each item's spec so
  page-rerender runs the section re-render. Without both signals the behaviour is unchanged (one assemble-only item
  per page — what site-wide refreshes rely on).
- **F3b — `store_generated_component_action.go`** (`F3b_store_reason_readd.patch`, applies on top of the deployed
  guard): re-add `"reason": "section_data_resolved"` to `createRerenderWorkItem`'s spec (the `component_id` is
  already in that spec, and is what scopes the re-render to dependents).
- **F3c — `rerender-pages` workflow config** (no redeploy): pass the two fields through the `create_rerender_items`
  step. Back up first, then:
  ```sql
  SELECT snapshot_agent('rerender-pages', 'F3: propagate section-rerender reason + component_id');
  UPDATE agent_definitions
  SET default_config = jsonb_set(
        jsonb_set(default_config,
          '{workflow,steps,create_rerender_items,config,reason}',
          '"input_data.spec.reason"'::jsonb, true),
        '{workflow,steps,create_rerender_items,config,component_id}',
        '"input_data.spec.component_id"'::jsonb, true)
  WHERE type = 'rerender-pages';
  -- verify
  SELECT jsonb_pretty(default_config #> '{workflow,steps,create_rerender_items,config}')
  FROM agent_definitions WHERE type = 'rerender-pages';
  ```

Order: deploy F3a + F3b (Go build → image bump on component-creator and whichever agent runs `create_rerender_items`),
then apply F3c. Either order is safe in isolation (a missing half just leaves the current assemble-only behaviour),
but both must be in place for the section re-render to activate. `image_landed` is left untouched — it already
flows through its own path.

Note the residual scope: `rerender_page_sections` re-renders **all** sections of each dependent page (not just the
changed component's), so a dependent page carrying another latently-broken component would surface it. That blast
radius is the dependent pages only, and is the accepted behaviour of the existing section-rerender path.

## Step F2 — Test the fix (throwaway component only — never the shared `system-stats`)

Three tiers. Preconditions: F1 guard deployed; Option A (load step + prompt rule) applied. F3 not required for F2.

### Tier 1 — Guard decision logic (deterministic, no DB)

`store_generated_component_guard_test.go` (in this folder) — unit tests for `schemaFieldSet` and the stranded-diff,
including the real system-stats rename. Drop it in `platform/orchestration/actions/` and run:
`go test ./platform/orchestration/actions/ -run 'SchemaFieldSet|RegenFieldContract' -v`. Asserts: a rename or drop
of a retained field is stranded; additions and identical schemas strand nothing; a new component (empty old schema)
strands nothing. This proves the guard's decision; it does not exercise the reject path (Tier 2).

### Tier 2 — Rejection path

No DB-backed integration harness exists in this package (the `*_test.go` files are unit-level), so the reject path
is proven end-to-end in **Tier 3b** below rather than as a standalone integration test — using an inactive throwaway
component to force the rename deterministically. (If a DB harness is added later, the same assertions — error
contains `regeneration removes/renames`, `agent_error_log` row written, component row unchanged — make a good
integration test.)

### Tier 3 — End-to-end smoke on throwaway components

Two variants, distinct namespaced section_types so they don't interfere. Both rely on the LLM mirroring
`section_type` into `function` (empirically it always does — every real component has `function == section_type`).

**Tier 3a — preservation / happy path (active component).**
```sql
INSERT INTO content_components
  (name, function, section_type, component_level, is_active, forked_from, html_template, input_schema, created_from)
VALUES ('zzz-fg-keep','zzz-fg-keep','zzz-fg-keep','section', true, NULL,
  '<style>.zzz-fg-keep-section{padding:2rem}</style><section class="zzz-fg-keep-section" data-component="zzz-fg-keep"><p>{{placeholder "eyebrow"}}</p><h2>{{placeholder "heading"}}</h2><p>{{placeholder "body"}}</p></section>',
  '{"fields":{"eyebrow":{"type":"text","source":"llm","required":true},"heading":{"type":"text","source":"llm","required":true},"body":{"type":"text","source":"llm","required":true}}}'::jsonb,
  'manual');

INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
                             priority, handler_agent, status, created_by, spec, item_key)
VALUES ('<TEST_SITE_UUID>', 'f2-test', 'build', 'needs_new_component', 'medium', 'F2 keep test',
        50, 'component-creator', 'triaged', 'f2-test',
        '{"section_type":"zzz-fg-keep","description":"a simple intro band","site_type":"test"}'::jsonb,
        'f2_keep_zzz-fg-keep');
```
`load_existing_component` finds it (active) → feeds `eyebrow, heading, body` → the prompt requires reuse → the store
matches by function and the guard passes → the component regenerates with names intact. Verify:
```bash
kubectl -n ai-persona-system logs -l app=agent-component-creator --tail=800 \
  | grep -iE "field-name preservation|REGENERATION FIELD-NAME"
```
```sql
SELECT array_agg(k ORDER BY k) FROM content_components c, jsonb_object_keys(c.input_schema->'fields') k
WHERE c.section_type='zzz-fg-keep' AND c.forked_from IS NULL;             -- eyebrow, heading, body still present
SELECT count(*) FROM agent_error_log WHERE error_message LIKE '%regeneration removes/renames%'
  AND error_message LIKE '%eyebrow%';                                     -- expect 0 (no rejection on happy path)
```

**Tier 3b — reject / fail-safe (INACTIVE component, LLM-unreproducible names).** The lever: the store matches
regardless of `is_active` (its query dropped the `is_active` filter, 2026-05-06), but `load_existing_component`
requires `is_active = true`. So an inactive row is skipped by the loader (prompt stays dormant → the LLM invents
normal field names) yet still matched by the store → the guard sees the old names stranded → rejects, before the
UPDATE.
```sql
INSERT INTO content_components
  (name, function, section_type, component_level, is_active, forked_from, html_template, input_schema, created_from)
VALUES ('zzz-fg-reject','zzz-fg-reject','zzz-fg-reject','section', false, NULL,
  '<style>.zzz-fg-reject-section{padding:2rem}</style><section class="zzz-fg-reject-section" data-component="zzz-fg-reject"><p>{{placeholder "zzz_alpha"}}</p><h2>{{placeholder "zzz_beta"}}</h2><p>{{placeholder "zzz_gamma"}}</p></section>',
  '{"fields":{"zzz_alpha":{"type":"text","source":"llm","required":true},"zzz_beta":{"type":"text","source":"llm","required":true},"zzz_gamma":{"type":"text","source":"llm","required":true}}}'::jsonb,
  'manual');

INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary,
                             priority, handler_agent, status, created_by, spec, item_key)
VALUES ('<TEST_SITE_UUID>', 'f2-test', 'build', 'needs_new_component', 'medium', 'F2 reject test',
        50, 'component-creator', 'triaged', 'f2-test',
        '{"section_type":"zzz-fg-reject","description":"a simple intro band","site_type":"test"}'::jsonb,
        'f2_reject_zzz-fg-reject');
```
Verify the rejection fired and the row is untouched:
```sql
-- rejection recorded, naming the stranded fields
SELECT created_at, error_message FROM agent_error_log
WHERE error_message LIKE '%regeneration removes/renames%' AND error_message LIKE '%zzz_alpha%'
ORDER BY created_at DESC LIMIT 3;

-- component unchanged: still inactive, still the zzz_* schema, no new version row
SELECT is_active, (SELECT array_agg(k ORDER BY k) FROM jsonb_object_keys(input_schema->'fields') k) AS fields, updated_at
FROM content_components WHERE section_type='zzz-fg-reject' AND forked_from IS NULL;
SELECT count(*) FROM component_versions
WHERE component_id = (SELECT id FROM content_components WHERE section_type='zzz-fg-reject' AND forked_from IS NULL);  -- expect 0
```
(The rejected work item may retry a few times — each retry rejects again; clean up promptly.)

Cleanup (both):
```sql
DELETE FROM component_versions WHERE component_id IN (SELECT id FROM content_components WHERE section_type IN ('zzz-fg-keep','zzz-fg-reject'));
DELETE FROM page_components   WHERE component_id IN (SELECT id FROM content_components WHERE section_type IN ('zzz-fg-keep','zzz-fg-reject'));
DELETE FROM content_components WHERE section_type IN ('zzz-fg-keep','zzz-fg-reject');
DELETE FROM site_work_items    WHERE spec->>'section_type' IN ('zzz-fg-keep','zzz-fg-reject');
```

Done-when: Tier 1 green; Tier 3a shows the names fed in, the rule in the prompt, the regen preserving
`eyebrow/heading/body`, and no rejection; Tier 3b shows an `agent_error_log` rejection naming `zzz_alpha/beta/gamma`
with the component left inactive, unchanged, and no new `component_versions` row.

---

## Step R4 — Back up before any write

```sql
CREATE TABLE IF NOT EXISTS _bak_compregen_component AS
  SELECT *, now() AS _bak_at FROM content_components WHERE id = '<COMPONENT_UUID>';

CREATE TABLE IF NOT EXISTS _bak_compregen_pagecomps AS
  SELECT *, now() AS _bak_at FROM page_components WHERE component_id = '<COMPONENT_UUID>';
```

These are throwaway snapshots; drop them once recovery is verified. For any agent whose config or image you
change in the fix, use `snapshot_agent('<type>','<reason>')` (revert with `revert_agent('<type>')`).

---

## Step R5 — Recover the affected pages

R3a confirmed the direction: the five hold per-page `content_data` under the **old** keys against a **new** live
schema, so each renders empty. The `content_data` is intact, so both routes below recover it. Back up first (R4)
and run the freshness check (R7) immediately before writing.

**Route A — full rebuild (recommended, simplest).** Raise a `needs_page` (full content build) work item per
affected page. The page-content-writer regenerates that page's content under the component's **current** (new)
schema field names, so the new-key `content_data` matches the new template and renders correctly — no manual
re-key, no reason-routing. Uses the writer (LLM), which is normal here, and sidesteps the rerender-pages reason
gap entirely.

**Route B — re-key, no LLM (preserves existing content).** Map each page's `content_data` keys old→new in code
(not hand-rolled SQL — the per-`N` stat fields and the non-1:1 cases, dropped `stat_N_icon` / added `cta_*`, make
a hand-written `UPDATE` error-prone; verify each rebuilt object before writing). Then trigger a **section**
re-render — and this is the part to get right: do **not** route it through `rerender-pages`, because its
`create_rerender_items` step drops `spec.reason`, so the resulting `page_rerender` items run assemble-only and just
re-ship the shells (this is the F3 gap). Instead insert the `page_rerender` work items **directly**, one per
affected page, each with a **valid `spec.page_id` UUID**, `reason=section_data_resolved` (so page-rerender runs the
`rerender_page_sections` pre-pass — `RenderTemplate` against the re-keyed `content_data`), and a non-colliding
`item_key` (e.g. `page_rerender_<page_name>_<site_id>`). Do not omit `page_id` — hand-made rerender items missing
it are what produced the earlier "invalid page_id" errors. (Once F3 ships and `create_rerender_items` propagates
the reason, a normal `needs_rerender` would also work; until then, go direct.)

Either route, gamesdesign's instance is out of scope (its band was dropped in an unrelated rebuild); apply only to
the five.

---

## Step R6 — Verify

```sql
SELECT pc.page_id, p.name AS page_name, length(pc.rendered_html) AS html_len,
       pc.content_hash, pc.build_status, pc.updated_at
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE pc.component_id = '<COMPONENT_UUID>'
ORDER BY pc.updated_at;
```

For each recovered page: `html_len` is now substantial (not ~0), `content_hash` has changed, `updated_at` is
fresh. Then confirm the band is present on the live deployed page (it deploys via the normal GitHub →
Backblaze path). A read-only spot check that the band renders content rather than empty placeholders:

```sql
SELECT page_id, (rendered_html ~* '<section') AS has_section, length(rendered_html) AS html_len
FROM page_components
WHERE component_id = '<COMPONENT_UUID>' ORDER BY page_id;
```

---

## Step R7 — Freshness / concurrency check (immediately before any write)

Another chat co-manages components. Just before R4/R5 (and any structural fix), re-check that nothing moved since
R1/R2:

```sql
SELECT id, updated_at, schema_template_synced, schema_field_count
FROM content_components WHERE id = '<COMPONENT_UUID>';

SELECT max(updated_at) AS latest_pagecomp_update, count(*) AS instances
FROM page_components WHERE component_id = '<COMPONENT_UUID>';

SELECT version_number, changed_by, created_at
FROM component_versions WHERE component_id = '<COMPONENT_UUID>'
ORDER BY version_number DESC LIMIT 3;
```

If the component's `updated_at`, the instance count, or the version history differ from what you recorded in
R1/R2, stop and re-inspect — the component may have been changed underneath you. No blind writes.

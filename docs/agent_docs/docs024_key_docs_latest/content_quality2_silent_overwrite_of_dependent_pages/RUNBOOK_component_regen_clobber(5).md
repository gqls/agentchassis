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
system-stats band right now. Investigation steps R0–R3a are done. **What's left is to fix the cause permanently,
test it, recover the five broken pages, and verify** — Steps F1, F2, R4, R5, R6 below. The fix is a Go change in
`StoreGeneratedComponentAction`'s regeneration branch; recovery re-keys `content_data` and re-renders with
`reason=section_data_resolved`.

## Progress at a glance

Tick these as you go (edit the file: `[ ]` → `[x]`).

- [x] **R0** — Orient (read this + the handoff)
- [x] **R1** — Inspect the component (schema fields, template placeholders, version history)
- [x] **R2** — Enumerate affected pages + resolve site domains
- [x] **R3** — Confirm the bind failure on one page (keys don't intersect)
- [x] **R3a** — When/why the binding broke — RESOLVED: 15:06 regen rename, `content_data` not migrated; five are **live-broken** (content-presence check)
- [ ] **F1** — Permanent fix in `StoreGeneratedComponentAction` regen branch (preserve retained names + section re-render; + guard)
- [ ] **F2** — Test the fix on a throwaway component (rename no longer empties dependents; deliberate rename surfaces loudly)
- [ ] **R4** — Back up before any write
- [ ] **R5** — Recover the five broken pages (re-key `content_data`, then section re-render with `reason=section_data_resolved`)
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

The cause is in the regeneration branch of `store_generated_component_action.go` (~L354–432): it overwrites a
shared component's `input_schema`/`html_template` field contract in place without migrating dependents'
`content_data`, and the `needs_rerender` it raises is assemble-only so it can't repair the sections. Implement, in
Go (reuse existing helpers; do not add a parallel path):

1. **Stop the silent rename.** When regenerating an *existing* component, treat its current field names as a
   contract. Feed the existing `input_schema` field names into the `generate_template` prompt and have
   `StoreGeneratedComponentAction` reject (or conform) a regenerated schema that **renames or drops a field that
   still exists**. Adding new fields (e.g. `cta_*`) and removing genuinely-unused ones is allowed; gratuitous
   renames (`eyebrow`→`eyebrow_label`, `stat_N_number`→`statN_value`, the `stat_N_`→`statN` reshape) are the
   damage. This is the smallest change that prevents recurrence and needs no per-dependent migration.
2. **If renames are ever legitimately required**, the regen must emit an explicit old→new field map and
   `StoreGeneratedComponentAction` must migrate every dependent's `content_data` keys (it already holds
   `existingSchema` and `inputSchemaJSON` at ~L360–398) **before** the re-render. Handle non-1:1 cases (dropped/
   added fields) explicitly.
3. **Make the triggered re-render actually re-render sections.** `createRerenderWorkItem` (or the `page_rerender`
   it leads to) must carry `reason=section_data_resolved` so dependents regenerate `rendered_html` from the
   preserved/migrated `content_data` — not the current assemble-only re-ship.
4. **Add a fail-loud guard (defensive).** Where a section re-render produces empty output for a section that was
   previously populated, carry the stored HTML or mark the work item failed and log at a visible level (not
   `logger.Debug`) — so any future regression surfaces instead of silently shipping a blank.

Ship via the normal path: `go build ./...` → GitHub Actions → Backblaze → new chassis image → bump `image_tag` on
the affected agent rows. Back up any agent whose config/image changes (`snapshot_agent`).

## Step F2 — Test the fix (on a throwaway component, never the shared `system-stats`)

1. Create a test component with a dependent page that has populated `content_data`. Regenerate it so the new
   schema **renames** a retained field. Assert: the rename is rejected/preserved (route 1) or the dependent's
   `content_data` is migrated (route 2), and a subsequent render of that page is **non-empty** (contains the
   content).
2. Assert the triggered re-render re-renders sections (carries `reason=section_data_resolved`) and the dependent's
   `rendered_html` regenerates with content — not an assemble-only re-ship.
3. Guard test: force a rename-without-migration on the test component and confirm it **fails loud** / carries the
   old render rather than shipping an empty section.
4. Regression: a regeneration that does **not** rename retained fields still succeeds and leaves dependents intact.

Done-when: a regeneration that would previously have stranded dependents either preserves their binding or
migrates+re-renders them with content, and a deliberate break surfaces loudly.

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

## Step R5 — Recover the affected pages (no LLM)

R3a confirmed the direction: the five hold per-page `content_data` under the **old** keys against a **new** live
schema, so this is a re-key job (no LLM needed). Two routes — re-key the existing `content_data`, or a full writer
rebuild (`needs_page`), which regenerates `content_data` under the current field names. For the re-key route, keep
it in code, not hand-rolled SQL — the per-`N` stat fields and the non-1:1 cases (dropped `stat_N_icon`, added
`cta_*`) make a hand-written `UPDATE` error-prone. Options, in order of preference:

1. Apply the same old→new key migration the F1 fix uses, run as a one-off over the five affected rows'
   `content_data`.
2. If scripting it, build each page's new `content_data` by mapping every old key to its new name (from the
   R3a-confirmed map), leaving genuinely-new fields (`cta_*`) unset and dropping fields with no target
   (`stat_N_icon`). Verify each rebuilt object before writing.

Then re-render — and this is the part the regen's own path gets wrong: a bare `needs_rerender` (what
`StoreGeneratedComponentAction` raises) is **assemble-only** and will just re-ship the empty shells. To actually
re-render sections from the re-keyed `content_data`, the `page_rerender` work item must carry
**`reason=section_data_resolved`** (or `image_landed`) so page-rerender runs the `rerender_page_sections` pre-pass
(component_library.go `RenderTemplate` against stored `content_data`) rather than the assemble-only `render_page`.
Reuse that existing chain — one `page_rerender` per affected page, with a **valid `spec.page_id` UUID**, the
`reason` set, and a non-colliding `item_key` (e.g. `page_rerender_<page_name>_<site_id>`). Do not omit `page_id` —
hand-made rerender items missing it are what produced the earlier "invalid page_id" errors. If any affected page
has NULL `content_data`, `rerender_page_sections` escalates it to the writer (`needs_page`) for a full rebuild +
backfill instead — that is expected. A full writer rebuild (`needs_page`) is the alternative recovery and is
necessary anyway if R3a shows the writer is emitting the wrong keys.

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

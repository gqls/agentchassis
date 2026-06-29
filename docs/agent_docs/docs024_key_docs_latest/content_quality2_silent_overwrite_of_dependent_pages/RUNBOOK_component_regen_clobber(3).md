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
re-renders the sections. So the 2026-06-24 15:06 regen did not write these blanks — the `15:06:12.956` stamp is
just `markPagesPendingRebuild`'s bulk UPDATE (it does not touch `rendered_html`), and the empty render predates
it. The single confirmed empty-and-dropped *symptom* was gamesdesign's index, since removed in an unrelated
rebuild, so it is not in the affected set below. What is still open (R3a) is **why** the keys diverged — a prior
regeneration, an out-of-band schema change (manual/migration), or a standing writer-vs-schema mismatch where the
writer emits keys the component schema never declared — which decides whether the fix lands in
`StoreGeneratedComponentAction` (migrate keys / preserve names on regen) or in the writer/schema contract. If the
divergence is a content_data re-key problem, recovery is a key re-map plus a section re-render (which must carry
`reason=section_data_resolved`, not a bare `needs_rerender`). The remaining steps confirm scope, back up, recover,
and verify.

## Progress at a glance

Tick these as you go (edit the file: `[ ]` → `[x]`).

- [x] **R0** — Orient (read this + the handoff)
- [x] **R1** — Inspect the component (schema fields, template placeholders, version history) — read-only
- [x] **R2** — Enumerate affected pages + resolve site domains — read-only
- [x] **R3** — Confirm the bind failure on one page (keys don't intersect) — read-only
- [ ] **R3a** — Establish when/why the binding broke (run the version-schema check) — read-only
- [ ] **R4** — Back up before any write
- [ ] **R5** — Recover the affected pages (re-key `content_data`, then section re-render with `reason=section_data_resolved`)
- [ ] **R6** — Verify (bands render, `content_hash` changed)
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

## Step R3a — Establish when/why the binding broke (read-only)

Traced through the code: `StoreGeneratedComponentAction` (component-creator's `store_generated_component`) is the
**rename writer**. When a component with the same `function` exists it takes the regeneration branch — snapshots
the old `html_template`/`input_schema` to `component_versions` (`changed_by='component-creator:regen'`), then
`UPDATE content_components SET html_template=…, input_schema=…` **in place** (same `component_id`, so dependents
keep pointing at the renamed template), marks dependents `build_status='pending'`, and raises one `needs_rerender`
work item per affected site. It does **not** migrate dependents' `content_data`.

Critically, that `needs_rerender` carries spec `{component_id, function, refresh_site_components:false}` — **no
`reason`** — and routes to `rerender-pages`. page-rerender only re-renders sections when
`reason ∈ {image_landed, section_data_resolved}`; with no reason it falls through to the **assemble-only**
`render_page` path, which re-stitches existing `rendered_html` and re-deploys. So the regen did **not** re-render
the sections, did **not** write new `rendered_html`. The `15:06:12.956` stamp is `markPagesPendingRebuild`'s bulk
UPDATE; the empty 7369-byte render therefore **predates** the regen — these pages were already rendering empty
before 15:06.

So the open question is no longer "which step wrote the blanks during the regen" but "when/why did the binding
break in the first place". `component_versions` has only the one row (the 15:06 snapshot), so this action hadn't
regenerated the component before. Decisive check — compare the pre-15:06 snapshot's schema field names to the
live ones and to the page's `content_data` keys (R3c):

```sql
SELECT version_number, change_source, changed_by,
       (SELECT array_agg(k ORDER BY k)
          FROM jsonb_object_keys(input_schema->'fields') k) AS old_schema_fields
FROM component_versions
WHERE component_id = '<COMPONENT_UUID>'
ORDER BY version_number;

SELECT jsonb_object_keys(input_schema->'fields') AS live_field
FROM content_components WHERE id = '<COMPONENT_UUID>' ORDER BY 1;
```

Interpretation: if `old_schema_fields` are the **old** names (matching `content_data`), the 15:06 regen is the
rename — but the live render was already broken before it, which only fits if the live schema was changed to new
keys **outside** this action (manual/migration) before the snapshot. If `old_schema_fields` are already the
**new** names, `content_data` has simply never matched the schema and the root is the **writer/schema contract**
(the writer emits old keys the template doesn't read), not a regen — moving the fix into the writer or the schema.

Done-when: you can say whether the keys diverged via a regeneration, an out-of-band schema change, or a standing
writer-vs-schema mismatch — which determines where the fix goes. Two facts hold regardless: regeneration overwrites
the shared field contract without migrating dependents' `content_data` (`StoreGeneratedComponentAction` has both
schemas in hand, lines ~360–398 — the natural migration/preserve point), and its triggered rerender is
assemble-only so it cannot repair a content-key mismatch.

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

First settle the key direction (R3a): if `content_data` is old-key and the live schema is new-key, migrate the
keys; if instead the writer/schema contract is the root, the durable fix is in the writer (and recovery may be a
full writer rebuild rather than a re-key). Assuming the re-key path: keep the re-key in code, not hand-rolled SQL —
the per-`N` stat fields and the non-1:1 cases (dropped `stat_N_icon`, added `cta_*`) make a hand-written `UPDATE`
error-prone. Options, in order of preference:

1. Apply the same old→new key migration the structural fix uses, run as a one-off over the affected rows'
   `content_data`.
2. If scripting it, build each page's new `content_data` by mapping every old key to its new name (from the
   R3a-validated map), leaving genuinely-new fields (`cta_*`) unset and dropping fields with no target
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

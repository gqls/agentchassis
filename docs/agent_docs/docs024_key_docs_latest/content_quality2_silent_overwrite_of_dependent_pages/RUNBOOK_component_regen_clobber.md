# RUNBOOK — component regeneration silently empties dependent pages

Operational steps for the next chat. A human runs every SQL/kubectl command. Inspect before asserting;
treat any 0-row / empty result as suspect until the query is verified. Confirm flagged column/table names
against the live schema (`\d <table>`) before relying on them — the bundle's `-schema-tables` dump covers
`content_components, component_versions, page_components, pages, sites`.

DB access:

```
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db
```

Throughout, `<COMPONENT_UUID>` is the full id of `system-stats` (get it in R1), and `<PAGE_UUID>` is one
affected page's id (get it in R2).

---

## R0 — Orient

The bug, IDs, key lists, and code locations are in `HANDOFF_component_regen_clobber.md` §2–§4. Do R1–R3
read-only first; do not write until R4 (backup) and R7 (freshness) are done.

---

## R1 — Inspect the component (read-only)

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

-- Placeholder names actually used in the template. Confirm the delimiter first by eyeballing the template;
-- this assumes Go-style {{.field}} / {{ .field }}. Adjust the regexp if the template uses another syntax.
SELECT DISTINCT m[1] AS placeholder
FROM content_components,
     regexp_matches(html_template, '\{\{[-\s.]*([a-zA-Z0-9_]+)', 'g') AS m
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

## R2 — Enumerate affected pages and resolve site domains (read-only)

Drive off `component_id`, not the page-id prefixes. Confirm the `sites` table and its domain column with
`\d sites` first (the domain column name is unverified — adjust `s.domain` if needed):

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

Expect five rows: three `index` pages (prefixes `b2edae00`, `0747e2fc`, `9baea9f9`, each a different site),
`case-study-kafka-consumer-group-remediation` (`77f10b13`), `gripper-detail` (`11364960`); all with `html_len`
near zero and the same `updated_at` instant. If the count or timestamps differ, something changed since the
hand-off — re-confirm before acting. Use the `domain` + `page_name` of any one row for the bundle's
`-runtime-site` / `-runtime-page`.

---

## R3 — Confirm the bind failure on one page (read-only)

```sql
-- Keys present in the page's stored content (the old names):
SELECT jsonb_object_keys(content_data) AS data_key
FROM page_components
WHERE page_id = '<PAGE_UUID>' AND component_id = '<COMPONENT_UUID>'
ORDER BY 1;
```

Compare these keys against the field set from R1. They should not intersect (old `stat_1_number` etc. vs new
`statN_value` etc.) — that non-overlap is the empty render. Confirm there is real content under the old keys
(so recovery has something to re-key):

```sql
SELECT content_data FROM page_components
WHERE page_id = '<PAGE_UUID>' AND component_id = '<COMPONENT_UUID>';
```

---

## R4 — Back up before any write

```sql
CREATE TABLE IF NOT EXISTS _bak_compregen_component AS
  SELECT *, now() AS _bak_at FROM content_components WHERE id = '<COMPONENT_UUID>';

CREATE TABLE IF NOT EXISTS _bak_compregen_pagecomps AS
  SELECT *, now() AS _bak_at FROM page_components WHERE component_id = '<COMPONENT_UUID>';
```

These are throwaway snapshots; drop them once recovery is verified. For any agent whose config or image you
change in the fix, use `snapshot_agent('<type>','<reason>')` (revert with `revert_agent('<type>')`).

---

## R5 — Recover the five pages (no LLM)

Two steps: re-key the stored content, then re-render from it. Keep the re-key in code, not hand-rolled SQL —
the per-`N` stat fields and the non-1:1 cases (dropped `stat_N_icon`, added `cta_*`) make a hand-written
`UPDATE` error-prone. Options, in order of preference:

1. Apply the same old→new key migration the Phase-3 fix uses, run as a one-off over the five rows' `content_data`.
2. If doing it as a script, build each page's new `content_data` by mapping every old key to its new name
   (from the R1-validated map), leaving genuinely-new fields (`cta_*`) unset and dropping fields with no target
   (`stat_N_icon`). Verify each rebuilt object before writing.

Then trigger the existing no-LLM re-render — reuse the path, do not build a new one. Prefer running the
`rerender-pages` agent scoped to these pages. If queuing work items directly, create one correctly-formed
`page_rerender` item per page with a **valid `spec.page_id` UUID** and a non-colliding `item_key`
(e.g. `page_rerender_<page_name>_<site_id>`). Do not omit `page_id` — hand-made rerender items with a missing
`page_id` are what produced the earlier "invalid page_id" errors. The `page-rerender` handler renders each
section from the (now re-keyed) `content_data` with no LLM call.

---

## R6 — Verify

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

## R7 — Freshness / concurrency check (immediately before any write)

Another chat co-manages components. Just before R4/R5 (and Phase 3 of the plan), re-check that nothing moved
since R1/R2:

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

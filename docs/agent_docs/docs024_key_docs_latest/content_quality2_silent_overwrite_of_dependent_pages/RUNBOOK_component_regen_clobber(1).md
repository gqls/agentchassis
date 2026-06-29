# RUNBOOK — component regeneration silently empties dependent pages

## The problem (start here)

A shared section component (one `content_components` row) is reused by many pages across several sites, and
each page stores its own `content_data` whose keys must match the component template's `{{.field}}` placeholders
exactly — `RenderTemplate` (component_library.go) fills a placeholder only when `content_data` has a key of the
same name, and silently leaves it empty otherwise (it runs Go `text/template` and then strips the resulting
`<no value>` tokens to empty string, logging only a warning, no error). On 2026-06-24 the `system-stats`
component (`fdd92ad4`) was regenerated with renamed fields (e.g. `stat_1_number`→`stat1_value`,
`heading`→`section_headline`) but the dependent pages' `content_data` was **not** migrated, so it still holds the
old key names. The component write itself was clean: `update_component_html` only snapshotted the old HTML,
swapped the template, and marked the dependent pages `build_status='pending'` — that single
`UPDATE … SET updated_at=NOW()` is why all five instances carry the same `15:06:12.956` timestamp, and it does
**not** touch `rendered_html`. The blanks were written **later**, when a render rebuilt those pending pages from
the old-key `content_data` against the renamed template: every renamed field missed, the sections rendered empty,
and the page assembler dropped them — silently, across every page sharing the component. Five instances on four
other sites are affected; `gamesdesign.co.uk` is **not** (its `system-stats` instance had already been dropped in
an unrelated rebuild). The `content_data` is intact, only mis-keyed, so recovery is a key re-map plus a no-LLM
re-render. The exact render step that wrote the blanks is not yet pinned — that is Step R3a below, and the first
real job. The remaining steps confirm the mismatch, identify the affected pages, back up, recover them, and verify.

## Progress at a glance

Tick these as you go (edit the file: `[ ]` → `[x]`).

- [ ] **R0** — Orient (read this + the handoff)
- [ ] **R1** — Inspect the component (schema fields, template placeholders, version history) — read-only
- [ ] **R2** — Enumerate affected pages + resolve site domains — read-only
- [ ] **R3** — Confirm the bind failure on one page (keys don't intersect) — read-only
- [ ] **R3a** — Pin the render step that actually wrote the blanks — read-only
- [ ] **R4** — Back up before any write
- [ ] **R5** — Recover the affected pages (re-key `content_data`, then no-LLM re-render)
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

Expect five rows: three `index` pages (prefixes `b2edae00`, `0747e2fc`, `9baea9f9`, each a different site),
`case-study-kafka-consumer-group-remediation` (`77f10b13`), `gripper-detail` (`11364960`); all with `html_len`
near zero and the same `updated_at` instant. If the count or timestamps differ, something changed since the
hand-off — re-confirm before acting. For the bundle's runtime evidence, use the `domain` + `page_name` of any one
row; the single-instance pages (`gripper-detail`, `case-study-kafka-consumer-group-remediation`) are clearer
examples than a generic `index`. (`gamesdesign.co.uk` will NOT appear here — its instance was dropped — so it is
not a useful runtime target for this bug.)

---

## Step R3 — Confirm the bind failure on one page (read-only)

```sql
-- Keys present in the page's stored content (the old names):
SELECT jsonb_object_keys(content_data) AS data_key
FROM page_components
WHERE page_id = '<PAGE_UUID>' AND component_id = '<COMPONENT_UUID>'
ORDER BY 1;
```

Compare these keys against the field set from R1. They should not intersect (old `stat_1_number` etc. vs new
`statN_value` etc.) — that non-overlap is what makes every placeholder render empty. Confirm there is real content
under the old keys (so recovery has something to re-key):

```sql
SELECT content_data FROM page_components
WHERE page_id = '<PAGE_UUID>' AND component_id = '<COMPONENT_UUID>';
```

---

## Step R3a — Pin the render step that wrote the blanks (read-only) — FIRST REAL JOB

`update_component_html` does not render (it only flags `build_status='pending'`), and the no-LLM
`rerender_page_sections` carries the stored HTML for any section it deems not "ready" and is gated to
`image_landed`/`section_data_resolved` reasons — so neither is the obvious writer of the blanks. Find what is:

- Inspect the `component-creator` agent's workflow: does its regeneration flow re-render / re-save dependent
  pages after `update_component_html` (e.g. via a render-and-persist step)? This is the prime suspect.
  ```sql
  SELECT type, image_tag, jsonb_pretty(default_config #> '{workflow,steps}') AS steps
  FROM agent_definitions WHERE type = 'component-creator';
  ```
- Identify what consumes `build_status='pending'` page_components and re-renders them (the secondary suspect),
  and confirm whether that path binds stored `content_data` straight into the current template with no key check.
- Cross-check timing: the `page_components.updated_at` of `15:06:12.956` is the pending-flag UPDATE. Look for a
  later write (deploy_commit / content_hash change / a rerender work item) that corresponds to the blank render,
  or confirm the blank was written in the same flow. Do not assume the blank and the pending-flag are the same
  event just because the row shows one `updated_at`.

Done-when: you can name the function/step that rendered the blank and show it binding old-key `content_data`
into the renamed template. That is where the guard (carry-stored-HTML / fail-loud on an emptied populated
section) belongs, alongside the structural fix in `component-creator` (preserve names or migrate keys).

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

Two steps: re-key the stored content, then re-render from it. Keep the re-key in code, not hand-rolled SQL — the
per-`N` stat fields and the non-1:1 cases (dropped `stat_N_icon`, added `cta_*`) make a hand-written `UPDATE`
error-prone. Options, in order of preference:

1. Apply the same old→new key migration the structural fix uses, run as a one-off over the affected rows'
   `content_data`.
2. If scripting it, build each page's new `content_data` by mapping every old key to its new name (from the
   R1-validated map), leaving genuinely-new fields (`cta_*`) unset and dropping fields with no target
   (`stat_N_icon`). Verify each rebuilt object before writing.

Once `content_data` carries the new keys, the existing render path matches again — reuse it, do not build a new
one. The no-LLM `rerender_page_sections` (component_library.go `RenderTemplate` against stored content_data) will
populate correctly; trigger it through the existing `rerender-pages` / `build-dispatch-loop` / `page-rerender`
chain, one `page_rerender` work item per affected page with a **valid `spec.page_id` UUID** and a non-colliding
`item_key` (e.g. `page_rerender_<page_name>_<site_id>`). Do not omit `page_id` — hand-made rerender items with a
missing `page_id` are what produced the earlier "invalid page_id" errors. Note: if any affected page has NULL
`content_data`, `rerender_page_sections` will escalate it to the writer (`needs_page`) for a full rebuild +
backfill instead — that is expected.

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

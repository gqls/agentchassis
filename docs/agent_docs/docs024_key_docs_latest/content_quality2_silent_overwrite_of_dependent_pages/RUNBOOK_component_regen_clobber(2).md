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
swapped the template, and marked the dependent pages `build_status='pending'` — that single bulk
`UPDATE … SET updated_at=NOW()` (one statement, one `NOW()`) is why all five surviving instances carry the
identical `15:06:12.956` timestamp, and it does **not** write `rendered_html`, so that stamp is *not* itself the
moment any blank was rendered. The confirmed empty-and-dropped symptom was on gamesdesign's index — which has
since been removed in an unrelated rebuild, so it is not in the affected set below. Whether the five surviving
instances (on three other sites — `leopardessconsulting.co.uk`, `robot-hands.com`, `ai-agent-orchestration.com` —
all `7369` bytes) are blank by this same mechanism is **not yet confirmed**: their identical byte-length across
different sites is suggestive (per-page content not landing) but not proof, and the `.956` stamp is only the
pending-flag. Step R3 reads their actual bytes and `content_data` keys to decide. If the mismatch is confirmed,
the `content_data` is intact and only mis-keyed, so recovery is a key re-map plus a no-LLM re-render. Pinning the
render step that wrote any blank is Step R3a — the first real job. The remaining steps identify the affected
pages, back up, recover them, and verify.

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
blank-by-rename story holds and the recovery in R5 applies. If (a) shows substantial visible text, or (c)'s keys
match (b), the five are **not** broken by this — only the dropped gamesdesign band was — and the live impact is
smaller; re-scope before any recovery. Either way, confirm real content exists under whatever keys `content_data`
uses, so recovery has something to re-key.

---

## Step R3a — Pin the render step that wrote the blanks (read-only) — FIRST REAL JOB

Already settled by the uploaded code and the component-creator config: `update_component_html` does not render
(only flags `build_status='pending'`); `RenderComponentAction` (in `v3_site_actions.go`) renders fresh content
and **returns** a map without writing the DB or `updated_at`; the no-LLM `rerender_page_sections` carries the
stored HTML for not-ready sections and is gated to `image_landed`/`section_data_resolved`; and the
`component-creator` workflow (v1.0.1080) is `ensure_site_record → read_site_spec → generate_template →
store_component → complete` with **no** re-render-of-dependents step. So none of those is the obvious writer of a
blank. Find what is:

- Read `StoreGeneratedComponentAction` (`grep -rn "func StoreGeneratedComponentAction"`) — component-creator's
  `store_generated_component` step and the prime suspect for the **rename write**. Determine whether it overwrites
  an existing component's `html_template`/`input_schema` (vs inserting a new row), whether it re-renders / re-saves
  dependents, and what `changed_by`/`change_source` it stamps. The single existing version row is
  `component-creator:regen` / `manual`, which suggests an out-of-band/manual regen rather than this workflow.
- Identify what consumes `build_status='pending'` page_components and re-renders them (the secondary suspect),
  and confirm whether that path binds stored `content_data` straight into the current template with no key check.
- Cross-check timing: the `page_components.updated_at` of `15:06:12.956` is the bulk pending-flag UPDATE (one
  statement, one `NOW()`, all five rows) — it does **not** write `rendered_html`, so the 7369-byte render predates
  it (a re-render + save afterwards would have bumped `updated_at` past `.956`). Look for the write that actually
  produced the bytes (deploy_commit / content_hash change / a `page_rerender` work item), or establish that the
  bytes long predate the regen. Do not assume the blank and the pending-flag are the same event.

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

# RUNBOOK — Sections dropped between page_components and the deployed page

**Created:** 2026-07-03  •  **Updated:** 2026-07-03
**Plan:** PLAN_section_assembly_drop.md
**Scope:** vonc.com index — provocation-card(pos 2) + lobby-grid(pos 5) present in
`page_components` (deployed) but absent from the deployed `index.html`.

> YOU ARE HERE: Phase 0 (diagnosis). Cause NOT confirmed. Do not apply a fix until D1–D4
> identify where the two sections are dropped.

Constants:
- site_id = 9ec3b9ee-5b08-461b-b4f8-9e1e03579c74
- index page_id = b4d24f8e-fccd-49df-9dad-aa56a0b20a68
- provocation-card component_id = 6163ff14-9f94-4962-aa19-d2718eabdeb1
- lobby-grid component_id = 9304f14d-e19b-4ce1-b3fd-f6a315aec6ed

---

## Phase 0 — Diagnosis (confirm the cause)

### D1 — Compare the 6 page_components across filter-candidate columns  [run first]
```sql
SELECT pc.position, cc.function,
       pc.build_status, pc.slot_name, pc.parent_instance_id, pc.schema_mode,
       pc.deploy_commit,
       (pc.content_hash IS NOT NULL) AS has_hash,
       (pc.locked_at IS NOT NULL)    AS locked,
       LENGTH(COALESCE(pc.rendered_html,'')) AS rlen,
       (pc.rendered_html LIKE '%<no value>%')     AS has_no_value,
       (pc.rendered_html LIKE '%data-component%') AS has_data_component
FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'::uuid
ORDER BY pc.position;
```
Read: what do positions 2 (provocation-card) and 5 (lobby-grid) have in common that
1/3/4/6 do NOT? Prime suspects: `has_no_value = t` on 2+5 only; a non-null `slot_name`
or `parent_instance_id`; a distinct `schema_mode`; a null `deploy_commit` on 2+5 while
1/3/4/6 have a commit (⇒ the deploy only committed 4). Whichever column splits 2+5 from
the rest points at the filter.

### D2 — What the deploy step reported
```sql
SELECT item_key, status, completed_at,
       jsonb_pretty(result #> '{response,deploy_result}') AS deploy_result
FROM site_work_items
WHERE site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid
  AND item_key LIKE 'manual-rebuild-index-%'
ORDER BY created_at DESC LIMIT 1;
```
Read: does `deploy_result.rendered_page.html` contain `provocation-card-section` /
`lobby-grid-section`? If NO, the drop happened at or before page-rerender's assembly
(rules toward H1/H3/H4). Look for any section-count or skip fields.

### D3 — page-rerender workflow (find the assemble action)
```sql
SELECT type, jsonb_pretty(default_config->'workflow'->'steps')
FROM agent_definitions
WHERE type = 'page-rerender' AND is_active = true;
```
Read: identify the step whose action assembles the page (likely `assemble_page` /
`compile_page_sections` / `render_page` / similar) and note its config (what it reads —
`page_id`/`site_id` → page_components — and any filter fields).

### D4 — Read the assemble action's Go code
Pull the action file named by D3 (e.g. `assemble_page_action.go` or the compile action).
Look specifically for:
- the SELECT against `page_components` — its WHERE/ORDER (does it filter on
  `build_status`, `rendered_html <> ''`, `<no value>`, `parent_instance_id IS NULL`,
  `slot_name`, `schema_mode`, a "skip interactive/empty" branch?).
- any per-section skip/continue (empty content, invalid HTML, dedupe).
This is the decisive artifact — it says exactly why 2 rows are excluded.

### D5 — Confirm Mode-B marker on the dropped rows (supports D1)
```sql
SELECT cc.function,
       (pc.rendered_html LIKE '%<no value>%') AS rendered_has_no_value,
       (cc.html_template LIKE '%<no value>%')  AS template_has_no_value
FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.page_id='b4d24f8e-fccd-49df-9dad-aa56a0b20a68'::uuid
  AND cc.function IN ('provocation-card','lobby-grid');
```
If `rendered_has_no_value = t` and the assemble action (D4) skips `<no value>` sections,
cause = confirmed (assembly drops Mode-B shells).

### Decision gate
- D1/D4 show an assemble-time filter (Mode-B / `<no value>` / empty / slot / interactive)
  → **H1 confirmed** → Phase 1A.
- D2 shows the content-writer `page_html` already had 4 and deploy used it verbatim
  → **H2 confirmed** → Phase 1B.
- D2 shows page-rerender assembled from page_components but still 4, and D4 shows a skip
  → back to H1 (1A). If D4 shows NO skip yet 4 assembled → widen: inspect the query's
  ORDER/joins and whether the 2 rows are even returned (add logging / a COUNT).

---

## Phase 1 — Fix (apply ONLY after the gate; provisional)

### 1A — Assemble filter wrongly drops valid sections (expected most-likely)
- In the assemble action, change the section selection so a section present in
  `page_components` for the page is NOT dropped for being a Mode-B/empty/`<no value>`
  shell or for being interactive. A shell with a `data-component` attribute + selectors
  is a valid, intentional section (runtime-filled). ALTER the existing action; keep the
  page-rerender workflow unchanged. Note any variable/signature change explicitly.
- If the filter exists to hide genuinely-broken `<no value>` output, the correct move is
  to make provocation-card + lobby-grid emit clean shells (no literal `<no value>` in
  rendered_html) — i.e. fix render_component for Mode-B, or regenerate them out of Mode-B
  — rather than teaching the assembler to ship `<no value>`. Decide based on D4/D5.

### 1B — Deploy used the wrong (4-section) source
- Point `deploy_page` / page-rerender at the stored `page_components` set (authoritative),
  matching its documented behaviour. Verify it re-assembles all rows for the page.

### After the fix
- Re-assemble + redeploy the index WITHOUT a full content rebuild (raise `needs_rerender`
  for the index so page-rerender re-runs assembly+deploy on the existing page_components —
  avoids re-triggering the content pipeline). If the fix is in the action, redeploy the
  action image first.

---

## Verification
```sql
-- both sections present in the freshly-assembled deploy_result
SELECT item_key, completed_at,
       (result::text LIKE '%provocation-card-section%') AS has_prov,
       (result::text LIKE '%lobby-grid-section%')       AS has_lobby
FROM site_work_items
WHERE site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid
  AND (item_key LIKE 'manual-rebuild-index-%' OR item_key LIKE '%rerender%')
ORDER BY created_at DESC LIMIT 3;
```
Then the live check on vonc.com/index.html after propagation: six sections present, in
plan order (hero, provocation-card, gauntlet-cta, brief-explanation, lobby-grid,
system-stats); provocation-card fills via the loader; footer intact.

## Do NOT
- Apply 1A/1B before D1–D4 confirm the drop point.
- Teach the assembler to ship literal `<no value>` — fix the shell rendering instead.
- Run a full content rebuild to "fix" this — it re-invokes the content pipeline and its
  own deferral/clobber behaviours; a rerender (assemble-only) is the right redeploy.

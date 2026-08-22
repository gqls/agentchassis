# RUNBOOK — bugfix 238

Every command here was needed and had to be got right. Gotchas are attached to
the command, not kept in a separate list.

## Identify the damaged row and its history

```sql
-- the live row (page_id and component_id come out of this; locked_at must be NULL
-- before any repair — a locked row is a human's and save_page_sections leaves it alone)
SELECT pc.id, p.id AS page_id, pc.slot_name, pc.locked_at, pc.component_id
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE p.site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  AND p.url = '/index.html' AND pc.slot_name = 'case-studies-grid';
-- → pc e20e474f-2e22-4a60-b56d-3cc2fd0f1de7 · page a716cacc-eec2-4aa6-a08b-7e6732506f41
--   · component 3f946437-1dc7-4164-987d-620933589076 · locked_at NULL
```

**GOTCHA — `page_component_history` has NO `page_component_id` column.** It keys on
`page_id` + `component_id` (+ a `slot_name` that is mostly empty on older rows).
Joining it the obvious way fails with `column h.page_component_id does not exist`.
And `component_id` is NULL on every row this site has, so **filter on a key the
payload actually carries**:

```sql
SELECT h.id, h.created_at, h.source,
       (SELECT count(*) FROM jsonb_object_keys(h.content_data)) AS keys
FROM page_component_history h
WHERE h.page_id = 'a716cacc-eec2-4aa6-a08b-7e6732506f41'
  AND h.content_data ? 'card1_image_alt'          -- the payload IS the selector here
ORDER BY h.created_at DESC LIMIT 10;
-- → afbce319-acd5-4194-8b1c-4ff93fcffad5, 2026-08-09 15:17:19Z, 58 keys — the last good one
```

## Prove which keys went, and that they are the non-llm ones

```sql
-- the 11-key diff, old snapshot vs live row
SELECT k FROM jsonb_object_keys((SELECT content_data FROM page_component_history
                                 WHERE id='afbce319-acd5-4194-8b1c-4ff93fcffad5')) k
EXCEPT
SELECT k FROM jsonb_object_keys((SELECT content_data FROM page_components
                                 WHERE id='e20e474f-2e22-4a60-b56d-3cc2fd0f1de7')) k
ORDER BY 1;

-- and their declared sources — the proof they are resolver-sourced, not LLM
SELECT f.key, f.value->>'source', f.value->>'required', f.value->>'type'
FROM content_components cc,
     LATERAL jsonb_each(COALESCE(cc.input_schema->'fields', cc.input_schema->'properties')) f
WHERE cc.id='3f946437-1dc7-4164-987d-620933589076' AND f.key LIKE '%_url'
ORDER BY 1;
```

**GOTCHA — the schema fields are nested one level down** (`input_schema->'fields'`,
with a legacy `->'properties'` dialect). A bare `jsonb_each(input_schema)` returns
the wrapper keys and reads as "no fields" (`WRONG_CALLS.md` 2026-08-10 records this
costing a false measurement on this very component).

## List what the template actually requires

```sql
SELECT DISTINCT (regexp_matches(html_template, '\{\{\.([a-zA-Z0-9_]+)\}\}', 'g'))[1]
FROM content_components WHERE id='3f946437-1dc7-4164-987d-620933589076' ORDER BY 1;
-- 58 fields. Cross-reference with the live content_data keys to see the holes.
```

## Verify at the served page (the only authority)

```bash
curl -s https://finetuning.uk/index.html > /tmp/ft.html   # fetch once, grep many
grep -c 'csg-card-image" src=""'            /tmp/ft.html   # want 0, was 5
grep -c 'src="/assets/images/case-study-'   /tmp/ft.html   # want 5
grep -c '<a class="csg-card-link" href="'   /tmp/ft.html   # want 5
grep -c '<a class="csg-cta-btn" href="'     /tmp/ft.html   # want 1
```

**GOTCHA — never grep the bare class name.** `csg-card-link` and `csg-cta-btn`
appear in the component's own inline `<style>` block, so a class-name grep returns
4 and 3 respectively while **zero anchors exist**. Anchor on `<a class="…" href="`.
(The same trap, one component over, is a standing LANDMINE.)

**GOTCHA — a complex `grep -o '.\{0,60\}…'` context pattern can exceed ugrep's
complexity limit** and error out mid-pipeline; keep context greps short or use
`grep -n` and cut.

## Assets: prove they exist AND that the check discriminates

```bash
for f in facilities legal-rag financial-data logistics-strategy private-ai; do
  printf '%s ' "$f"; curl -s -o /dev/null -w '%{http_code} %{content_type}\n' \
    "https://finetuning.uk/assets/images/case-study-$f.jpg"; done
# all 200 image/jpeg. A fabricated sixth filename must 404 — that is the control.
```

## Queue a no-LLM rerender

Copy the spec shape from a **completed row of the same `item_type` on the same
site**, never from the action source or a sibling type:

```sql
SELECT jsonb_pretty(spec), page_id, item_key FROM site_work_items
WHERE item_type='page_rerender' AND status='complete'
  AND site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc'
ORDER BY updated_at DESC LIMIT 2;
-- → {"domain","reason","page_id","filename","page_name"}; page_id column NULL and
--   it still completed, so the spec's page_id is what the handler reads.
```

`reason` is load-bearing: `section_data_resolved` / `image_landed` route through
`rerender_page_sections` (re-render from template + stored `content_data`, **no
LLM**). A reason-less item takes the assemble path, which re-staples stored HTML
and cannot pick up new data.

## Why these fields never resolved (and what makes them resolve)

```sql
SELECT aspect FROM site_specs WHERE site_id='1368e337-…' AND is_current ORDER BY 1;
-- no `case_studies`, no `pages`
SELECT aspect, count(*) FROM site_specs WHERE aspect IN ('case_studies','pages') GROUP BY 1;
-- 0 rows FLEET-WIDE: these two sources have never resolved on any site
```

`site_specs` has `UNIQUE (site_id, aspect) WHERE is_current` — one INSERT per
aspect seeds cleanly. `ensureSpecs` loads `aspect → data` and `resolveSpecPath`
navigates `data->'<leaf>'`, so `site_specs.pages.contact_url` wants an aspect
literally named `pages` holding `{"contact_url": "..."}`.

**GOTCHA — `pages` is ALSO a top-level source type** (`ensurePages`, reading the
real `pages` table). `site_specs.pages.contact_url` routes to the *spec aspect*,
not the table. The two are unrelated despite the name.

## Migration runner

```bash
./scripts/migration/run-migrations.sh            # dry run, lists pending — SLOW (>2 min)
```

It probes each pending file by executing it inside a doomed transaction, so a dry
run takes minutes and takes brief row locks. Run it in the background and read the
output file; do not assume it hung.

---

## 2026-08-20 additions — the queries that were hard to get right

### The one that matters: did a field ever go from a VALUE to blank across generations?

This is the settling check the `090` asked for, and it is what proved the `bugs_open/268` carry
extension works in production. Pair consecutive archived generations of the SAME (page, slot) and
look for a field that held a non-empty value and then did not.

```sql
WITH h AS (
  SELECT h.page_id, h.slot_name, h.created_at, h.content_data,
         lag(h.content_data) OVER (PARTITION BY h.page_id, h.slot_name ORDER BY h.created_at) AS prev
  FROM page_component_history h
  WHERE h.source = 'artefact_archive_trigger'      -- ⚠ see the two gotchas below
    AND h.slot_name IS NOT NULL
), lost AS (
  SELECT p.page_id, p.slot_name, p.created_at, k.key AS fld
  FROM h p, LATERAL jsonb_object_keys(p.prev) k(key)
  WHERE p.prev IS NOT NULL
    AND btrim(COALESCE(p.prev->>k.key, '')) <> ''
    AND btrim(COALESCE(p.content_data->>k.key, '')) = ''
), typed AS (                                       -- llm fields changing is the writer WORKING
  SELECT l.*, COALESCE(cc.input_schema->'fields'->l.fld->>'source', '(undeclared)') AS src
  FROM lost l
  LEFT JOIN page_components pc
         ON pc.page_id = l.page_id AND pc.slot_name = l.slot_name AND pc.build_status = 'deployed'
  LEFT JOIN content_components cc ON cc.id = pc.component_id
)
SELECT src, count(*) AS events, max(created_at) AS newest
FROM typed WHERE src NOT IN ('llm', '(undeclared)', '') GROUP BY 1 ORDER BY 2 DESC;
```

**Always pair it with the demand control**, or a zero means nothing — a quiet fleet and a working
fix look identical:

```sql
WITH h AS (SELECT h.page_id, h.slot_name, h.created_at,
                  lag(h.content_data) OVER (PARTITION BY h.page_id, h.slot_name ORDER BY h.created_at) AS prev
           FROM page_component_history h
           WHERE h.source='artefact_archive_trigger' AND h.slot_name IS NOT NULL)
SELECT count(*) AS archive_pairs FROM h
WHERE prev IS NOT NULL AND created_at > timestamptz '<the last loss event>';
```

Result 2026-08-20: 66 non-llm losses, all `renderer`/`static`, all 2026-08-11 → 08-14 18:36 UTC,
**none since**, against **3,033** archived pairs. The 268 fix (`8f899cc8d`) landed 08-14 09:13 BST.

**⚠ Gotcha 1 — the window is not the bug's life.** `artefact_archive_trigger` rows begin
**2026-08-09** (migration 357). Eleven days, not five months.

**⚠ Gotcha 2 — you cannot widen it with the older rows.** `save_page_sections_overwrite` reaches
back to 2026-03-16 (20,351 rows) but writes **`slot_name` NULL**, so consecutive generations of one
slot cannot be paired at all. Provenance and schema completeness are the same choice here.

### The EXISTS probe — did this row EVER hold the value? Run it at BOTH tightnesses

Wrong in two opposite directions, both hit on 2026-08-20 (`WRONG_CALLS`):

- joined on `page_id` **alone** → over-counts by slot (credited `system-stats` with 103 `cta_url`
  values belonging to `tool-list`/`game-list`/`hero`);
- tightened to `slot_name` + `source='artefact_archive_trigger'` → under-counts by writer, returned
  **0 for all 25** field slots, contradicting the bug file's correct claim that aao IS a regression
  (that page: 1,184 app rows with NULL slot, 42 holding the value; 211 trigger rows, 0 holding it).

The working discriminator is **content identity** — how many deployed components on that page
declare the field:

```sql
SELECT count(DISTINCT pc.id)
FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.page_id = '<page>' AND pc.build_status = 'deployed'
  AND cc.input_schema->'fields' ? '<field>';
-- 1 ⇒ the field name attributes the value.  >1 ⇒ no data-only recovery is possible.
```

### Would the declared source resolve today?

Answers "is this a restore, or does the value not exist anywhere?" — 0 of 14 resolved on 08-20.

```sql
SELECT ss.aspect,
       ss.data #>> string_to_array('<leaf.path>', '.') AS resolves_to
FROM site_specs ss WHERE ss.site_id = '<site>' AND ss.aspect = '<aspect>';
-- no rows ⇒ the source has NEVER existed on this site; a rebuild restores nothing.
```

### Is the dead-URL guard armed? Count STEPS, fleet-wide, never agents

```sql
SELECT k.key AS step, (k.value->'config'->>'record_dead_url_controls')::boolean AS record_armed,
       (k.value->'config'->>'refuse_dead_url_controls')::boolean  AS refuse_armed
FROM agent_definitions ad,
     LATERAL jsonb_path_query(ad.default_config, 'strict $.**.steps') steps,
     LATERAL jsonb_each(steps) k
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND k.value->>'action' IN ('rerender_page_sections','render_component');
```

⚠ A `default_config::text LIKE '%rerender_page_sections%'` returns **three agent types** and the
honest answer is **one step**. `_` is a SQL wildcard too. Counting agents when the question is steps
is the 2026-08-11 error in this bug's own slice B.

⚠ `page-rerender` keeps its steps at `{workflow,steps,<name>}`; `page-content-writer` nests them in
`process_sections_loop.config.sub_workflow.steps`. A path copied from `380` finds nothing here and
reads as "no such step".

---

## The content_data key-loss census, and the control without which it means nothing

Added 2026-08-22 while scoping `bugs_open/354` (RFC_042 option (c)). This is the query that decides
whether the eight uncarried writers lose resolved keys — and the reason it needs a second run.

**How the population splits.** `page_component_history` rows written by the archive trigger carry
`slot_name`, so consecutive generations of one (page, slot) can be paired. `op` is the discriminator:

- `op='delete'` → **the funnel** (`save_page_sections` is DELETE+INSERT, so its old rows arrive as
  deletes). The re-render path is *not* separate — it emits sections that `save_page_sections`
  ingests, and holds no `UPDATE page_components` of its own.
- `op='overwrite'` → **the eight uncarried writers** (every in-place `UPDATE … SET content_data`).

⚠ `op` is a proxy for the writer, not the writer. A future non-funnel writer that deletes and
re-inserts lands in the funnel bucket and would be mis-attributed.

```sql
-- Swap 'overwrite' for 'delete' to run the DEMAND CONTROL (see the warning below).
WITH pre AS (
  SELECT h.id, h.page_id, h.site_id, h.slot_name, h.component_id,
         h.content_data AS before_data, h.created_at
    FROM page_component_history h
   WHERE h.source='artefact_archive_trigger' AND h.op='overwrite'
), sch AS (
  SELECT p.*,
         COALESCE(
           -- the FK route, which silently loses 58% of rows (ON DELETE SET NULL)
           (SELECT cc.input_schema->'fields' FROM page_components pc
              JOIN content_components cc ON cc.id = pc.component_id WHERE pc.id = p.component_id),
           -- the slot fallback, which recovers 24% -> 73% coverage
           (SELECT cc2.input_schema->'fields' FROM page_components pc2
              JOIN content_components cc2 ON cc2.id = pc2.component_id
             WHERE pc2.page_id = p.page_id AND pc2.slot_name IS NOT DISTINCT FROM p.slot_name
             ORDER BY pc2.updated_at DESC LIMIT 1)) AS fields,
         COALESCE(
           (SELECT h2.content_data FROM page_component_history h2
             WHERE h2.source='artefact_archive_trigger' AND h2.page_id=p.page_id
               AND h2.slot_name IS NOT DISTINCT FROM p.slot_name AND h2.created_at > p.created_at
             ORDER BY h2.created_at ASC LIMIT 1),
           (SELECT pc3.content_data FROM page_components pc3
             WHERE pc3.page_id=p.page_id AND pc3.slot_name IS NOT DISTINCT FROM p.slot_name
             ORDER BY pc3.updated_at DESC LIMIT 1)) AS after_data
    FROM pre p
)
SELECT s.created_at::date AS day,
       split_part(COALESCE(s.fields->k->>'source',''),'.',1) AS src_root,
       count(*) AS losses
  FROM sch s, LATERAL jsonb_object_keys(s.fields) k
 WHERE s.fields IS NOT NULL AND s.after_data IS NOT NULL
   AND COALESCE(s.fields->k->>'source','') NOT LIKE 'llm%'   -- every field declares a source
   AND s.before_data ? k AND btrim(COALESCE(s.before_data->>k,'')) <> ''
   AND (NOT (s.after_data ? k) OR btrim(COALESCE(s.after_data->>k,'')) = '')
 GROUP BY 1,2 ORDER BY 1,2;
```

⚠ **NEVER report this query's zero without the demand control, and the control must be the
`op='delete'` run — not the LLM-field arm.** Measured 2026-08-22: the `overwrite` population
returned **0 losses**, and the LLM-field arm of the *same* query returned **0** as well. Both share
the joins, the pairing and the schema resolution, so both go quiet together; two zeros side by side
look like corroboration and are not. The `op='delete'` run is the control that discriminates —
it returns **72 losses, `static`=24 / `renderer`=48, dated 08-09 (4), 08-11 (63), 08-12 (5) and none
since**, matching what RFC_042 §4.6 measured by a different method. That is what licenses reading the
`overwrite` zero as real. (It also re-confirms 238's closure by an independent route: 5,532 judgeable
funnel transitions since 08-12, zero losses.)

⚠ **Coverage, so a zero is not over-read.** 380 overwrite pairs; 279 judgeable with the slot
fallback, 92 with the FK route alone. The archive is additionally blind to any row being *born*:
`pch_op_check` permits only `overwrite`/`delete`, so **the trigger never fires on INSERT**.

⚠ **`page_component_history.application_name` cannot name the writer.** Every application-side write
carries the pgx connection default (`app - 10.20.99.74:35564` and thousands of siblings); hand-run
SQL carries `psql`. A pod IP is not a call site, and several writers share one binary.

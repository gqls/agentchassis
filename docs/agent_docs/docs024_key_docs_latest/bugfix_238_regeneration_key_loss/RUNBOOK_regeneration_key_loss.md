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

# RUNBOOK — meta descriptions

Every command here had a gotcha attached. The gotcha is the point; copy the whole entry,
not just the SQL.

---

## The census (the number that resizes this from 5 pages to 407)

```sql
SELECT count(*) AS active_pages,
       count(*) FILTER (WHERE COALESCE(meta_description,'')='') AS empty_meta,
       round(100.0*count(*) FILTER (WHERE COALESCE(meta_description,'')='')/count(*),1) AS pct
FROM pages WHERE status='active';
```

⚠ **`COALESCE(...)=''`, never `IS NULL`.** The column is nullable *and* routinely holds
the empty string — `upsertPage` binds `GetStringField(page,"meta_description","")`, so
the plan path writes `''` rather than NULL. A census on `IS NULL` alone under-reports.

⚠ **Filter `status='active'`.** Archived pages are still rows; counting them inflates the
figure with pages nobody serves.

## Is it historical debt, or still happening? (the question that ranks the fixes)

```sql
SELECT date_trunc('month', created_at)::date AS month,
       count(*) AS created,
       count(*) FILTER (WHERE COALESCE(meta_description,'')='') AS born_empty
FROM pages WHERE status='active' GROUP BY 1 ORDER BY 1;
```

⚠ This is only a *birth* proxy: it groups by `created_at` but reads today's
`meta_description`, so a page that was born full and later blanked counts as born empty.
There is **no `pages` history table** (only `page_component_history` and
`site_component_history`), so the two cannot be separated from the DB. State the figure
as "pages created in month M that are empty today", which is what it measures.

## Does the planner even ask for a description?

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT default_config::text FROM agent_definitions
WHERE type='build-site-planner' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL LIMIT 1;" > planner.json
```

Then read the `Return JSON:` block. The per-page object lists `name`, `title`,
`page_type`, `nav_label`, `nav_order`, `in_header`, `in_footer`, `sections` — **and no
description field**.

⚠ **`ILIKE '%meta_description%'` on the config returns TRUE and means nothing here.** The
planner *does* contain the string — in `load_existing_pages`, a `query_database` step
that SELECTs the column. Matching the string proves the agent mentions it, not that it is
asked to write it. Read the output schema, not the grep.

⚠ **Read the live row, never the seed SQL** — the seed records what the agent *was*.

## Is there any writer I missed?

Two scans, because either alone under-reports:

```bash
# 1. Go — multiline-aware. A line-oriented grep for "UPDATE ... meta_description"
#    finds nothing, because the column sits ~10 lines below the verb.
python3 - <<'PY'
import re,glob
for p in glob.glob('platform/**/*.go',recursive=True)+glob.glob('cmd/**/*.go',recursive=True):
    s=open(p,encoding='utf-8',errors='replace').read()
    for m in re.finditer(r'(UPDATE\s+pages|INSERT\s+INTO\s+pages)(.{0,900}?)(?:`|\Z)', s, re.S|re.I):
        if 'meta_description' in m.group(0):
            print(f"{p}:{s[:m.start()].count(chr(10))+1}: {m.group(1)}")
PY
```

```sql
-- 2. Agent configs can carry raw SQL in a query_database step, which no Go scan sees.
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text ~* 'update[[:space:]]+pages';
```

⚠ The Go scan also misses the `Col("meta_description", …)` / `UpsertPageForRole` builder
sites, which assemble SQL dynamically — grep `Col("meta_description"` separately. Three
scans, not one, and the answer only counts when all three have run.

## Did a work item that CLAIMS to fix a description actually fix one?

```sql
SELECT w.id, w.status, s.domain, p.name,
       length(COALESCE(p.meta_description,'')) AS meta_len_now
FROM site_work_items w
JOIN pages p ON p.id = w.page_id
JOIN sites s ON s.id = w.site_id
WHERE w.item_type='content_rewrite' AND w.summary ILIKE '%meta description%';
```

⚠ **`status='complete'` is not evidence.** Two items here are complete and both targets
still read 0. Join to `pages` and read the column; the item's own status is the thing
being tested, not the test.

## Firing the diagnosis loop on a CODE-ONLY symptom

```bash
export SEED_SCOPE="path/to/file.go:SymbolName,path/to/other.go:OtherSymbol"
./docs/.../090_TRIGGER_needs_diagnosis_v1.sh "<symptom naming those symbols>"
```

⚠ **Without `SEED_SCOPE`, a code-only symptom is accepted, dispatched, and then dies** at
`diagnose_assemble_bundle: no scope (tried "route.scope.Symbols", "input_data.seed_scope",
then code_results)`. Naming files in the prose is not enough; the scope comes from
`SEED_SCOPE` (`path[:Symbol]`, comma-separated — 090 script lines 117 / 282-346).

⚠ **A failed run's terminal row reads `current_step=complete, status=COMPLETED`.** The two
genuinely FAILED rows carry `__step_error = (none)`; the real message is in
`collected_data->>'__step_error'` on the **`complete`** row. So:

```sql
-- the honest "did it produce anything" check
SELECT count(*) FROM diagnosis_artifacts WHERE correlation_id::text='<RUN_CORR>';
-- the error, which is NOT on the FAILED rows
SELECT current_step, status, left(collected_data->>'__step_error',900)
FROM orchestration_states WHERE correlation_id::text='<RUN_CORR>' ORDER BY updated_at;
```

⚠ **Use `RUN_CORRELATION_ID`, not the intake correlation the script prints first**, and
join `orchestration_states` on **`correlation_id`** — `collected_data->'input_data'->>
'fix_correlation_id'` returns 0 rows for these.

## Verify at the served page, with a control

```bash
for u in blog/<page> guides/tool-<page>; do
  code=$(curl -s -o /tmp/p.html -w "%{http_code}" "https://<domain>/$u.html")
  desc=$(grep -o '<meta name="description" content="[^"]*"' /tmp/p.html | head -1 | sed 's/.*content="//;s/"$//')
  printf "%-50s http=%s bytes=%s desc_len=%s\n" "$u" "$code" "$(wc -c </tmp/p.html)" "${#desc}"
done
```

⚠ **Always include a page that MUST come out non-empty** (a `/guides/tool-*` sibling
reads 156 chars) **and print the byte count.** A zero from a guessed URL is a 404, not a
finding — 309 §C already paid for that once with six false readings.

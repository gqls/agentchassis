# RUNBOOK — diagnosis bundle Schema section

Every command here was needed to get something right. Gotchas attached. Change
them **here**, not in your scrollback.

---

## 1. What the include filter actually selects (the measurement that sized the job)

The Go defaults are what production runs, so these patterns are the live ones.

```sql
SELECT count(DISTINCT table_name) FROM information_schema.columns
WHERE table_schema='public'
  AND table_name NOT ILIKE '%backup%' AND table_name NOT ILIKE '%bak%'
  AND table_name NOT ILIKE '%archive%' AND table_name NOT ILIKE '%supersede%'
  AND (table_name ILIKE 'site%' OR table_name ILIKE 'page%'
       OR table_name ILIKE 'content%' OR table_name ILIKE 'flow%');
-- 26        (2026-08-10)
SELECT count(DISTINCT table_name) FROM information_schema.tables WHERE table_schema='public';
-- 433       (2026-08-10)
```

**The one that mattered** — does the include cover the tables the gather itself
reads? Write it as a table of booleans, not as a count: a count cannot tell you
*which* ones are missing, and the answer was "five of six".

```sql
SELECT t, (t ILIKE 'site%' OR t ILIKE 'page%' OR t ILIKE 'content%' OR t ILIKE 'flow%') AS in_schema_section
FROM unnest(ARRAY['agent_error_log','site_work_items','orchestration_states',
                  'agent_definitions','llm_call_log','code_symbols']) AS t;
-- only site_work_items is true (2026-08-10)
```

## 2. Confirm no live config overrides the Go defaults

Before claiming "the Go defaults are what production runs". `is_snapshot` and
`deleted_at` filters matter — without them you read retired rows.

```sql
SELECT type, s.key,
       s.value->'config'->'schema_include_patterns',
       s.value->'config'->'schema_exclude_patterns',
       s.value->'config'->'schema_full'
FROM agent_definitions ad,
     LATERAL jsonb_each(COALESCE(ad.default_config #> '{workflow,steps}','{}'::jsonb)) s
WHERE s.value->>'action' = 'diagnose_load_runtime'
  AND ad.deleted_at IS NULL AND COALESCE(ad.is_snapshot,false)=false;
-- 4 rows, all three columns NULL (2026-08-10): diagnose-agent + 3 wiring probes
```

> ⚠ **This `jsonb_each` walks TOP-LEVEL steps only.** A step nested in a loop's
> `sub_workflow` is invisible to it — the commission records that exact query
> shape returning **3 of 6** on the migration-356 work. Here the four rows found
> are the whole population *for this action*, but do not reuse this query shape
> to prove an absence without walking recursively or cross-checking with
> `::text LIKE`.

## 3. Read the Schema section of a real bundle

```sql
SELECT substring(collected_data->>'bundle' from position('## Schema' in collected_data->>'bundle') for 900)
FROM orchestration_states
WHERE collected_data ? 'bundle' AND collected_data->>'bundle' LIKE '%## Schema%'
ORDER BY updated_at DESC LIMIT 1;
```

> ⚠ **THE TRAP THAT NEARLY COST THE SECOND HALF OF THE FIX.** `substring(txt from
> position('## Schema' in txt))` with **no length** runs to the END OF THE
> BUNDLE, not the end of the section. Testing "does the section say it is
> filtered?" against that measures the whole ~80,000-char document. It returned
> **true**, which would have meant "the notice already exists, don't build it".
> Bound the section at its closing fence before asserting anything about it:

```sql
WITH b AS (SELECT collected_data->>'bundle' AS txt FROM orchestration_states
           WHERE collected_data ? 'bundle' AND collected_data->>'bundle' LIKE '%## Schema%'
           ORDER BY updated_at DESC LIMIT 1),
s AS (SELECT substring(txt from position('## Schema' in txt)) AS tail FROM b),
f AS (SELECT tail, repeat(chr(96),3) AS fence FROM s),
sec AS (SELECT substring(tail from 1 for
          position(fence in substring(tail from position(fence in tail)+3)) + position(fence in tail) + 2) AS sch FROM f)
SELECT length(sch), (sch ILIKE '%filter%'), (sch ILIKE '%not exhaustive%'), (sch LIKE '%orchestration_states(%') FROM sec;
-- 8819 | f | f | f     (2026-08-10, pre-fix)
```

Two more gotchas inside that one query:

- **`full` is a reserved word** — `substring(full from …)` is a syntax error. Alias
  the column something else (`txt`).
- **Never type a backtick inside a double-quoted bash string** — it is command
  substitution. Use `chr(96)`, or single-quote the whole `-c` argument and use
  `$$…$$` for SQL string literals.
- **A positive match may be a COLUMN NAME.** `ILIKE '%relevance%'` returned true
  and the match was `relevance_score float8` — a column in the listing, not a
  notice. Print the surrounding 130 chars before believing any keyword hit:
  `substring(tail from greatest(position('relevance' in tail)-60,1) for 130)`.

## 4. Run the NEW query live before trusting the mock

sqlmock proves the shape; only Postgres proves it parses.

```sql
SELECT count(DISTINCT table_name) FROM information_schema.columns
WHERE table_schema='public' AND (
  (table_name NOT ILIKE '%backup%' AND table_name NOT ILIKE '%bak%'
   AND table_name NOT ILIKE '%archive%' AND table_name NOT ILIKE '%supersede%'
   AND (table_name ILIKE 'site%' OR table_name ILIKE 'page%'
        OR table_name ILIKE 'content%' OR table_name ILIKE 'flow%'))
  OR table_name = ANY('{agent_definitions,agent_error_log,code_symbols,llm_call_log,orchestration_states,site_work_items}'::text[]));
-- 31        (2026-08-10; was 26)
```

> ⚠ **`EXECUTE` cannot be used inside a subquery** — `SELECT … FROM (EXECUTE …)`
> is a syntax error, so a `PREPARE`d version of the real statement cannot be
> wrapped in a `count()`. Inline the literals instead.

## 5. Tests

```bash
go test ./platform/orchestration/actions/ -run 'TestSchemaAlways|TestInputSpecDefault|TestGatherSchema|TestStripGoComments|TestSchemaFilterNotice' -v
go test ./platform/orchestration/actions/          # whole package
gofmt -l platform/orchestration/actions/diagnose_load_runtime_action.go   # must print nothing
```

**Prove the guard can fail** (a passing guard proves nothing until you have seen
it fail — the mutation is the evidence, not the green run):

```bash
cp platform/orchestration/actions/diagnose_load_runtime_action.go /tmp/lr.bak
python3 - <<'EOF'
p='platform/orchestration/actions/diagnose_load_runtime_action.go'
s=open(p).read(); open(p,'w').write(s.replace('\t"orchestration_states",\n','',1))
EOF
go test ./platform/orchestration/actions/ -run TestSchemaAlways   # MUST fail, both tests
cp /tmp/lr.bak platform/orchestration/actions/diagnose_load_runtime_action.go
```

> `go vet ./platform/orchestration/actions/` reports `load_component_library_actions.go:207:
> unreachable code`. **Pre-existing, not from this work** — do not try to "fix" it here.

## 6. Post-roll verification (BLOCKED until the fleet rolls — owner runs `make release`)

Pod-grep needs a literal **and a negative control**; a roll is not evidence your
fix shipped.

```bash
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "This listing is FILTERED, not the whole database"'   # expect >=1
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "schema_always_tables"'                               # expect >=1
```

Then the check that actually settles it — **not** "the table appears in the
bundle". Re-run a `090` whose evidence lives in `orchestration_states`
(`bugs_open/236`, hero/logo half) and confirm its `data_request` **executes**
instead of erroring 42703:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<CORR>';
```

## 7. Council gate

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/diagnosis_schema_visibility/SUBMISSION_2026-08-10_schema_section_always_tables.json
# SUBMISSION_CORR = df9dae6c-b7ca-4605-8dd4-26462ce4b20b  (2026-08-10)
```

Verdict (budget ~30 min, mostly dispatch queue — a missing row is latency, not a
dropped dispatch; do not retry on that evidence):

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='df9dae6c-b7ca-4605-8dd4-26462ce4b20b' AND kind='council_report' ORDER BY created_at;

SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
```

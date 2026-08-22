# RUNBOOK — bugfix 309 / phantom-source components

## Verify the symptom at the SERVED page (never the stored HTML)

```bash
curl -sS "https://fundamentallyai.com/platform-log/index.html" -o /tmp/pli.html -w "%{http_code} %{size_download}\n"
python3 - <<'EOF'
import re
html = open('/tmp/pli.html').read()
cards = re.findall(r'<article class="bl-card".*?</article>', html, re.S)
print("cards:", len(cards))
for i,c in enumerate(cards,1):
    print(i, "anchors:", len(re.findall(r'<a\s', c)))
EOF
```
Gotcha (from the filing thread, kept): a regex over a truncated slice of one card is
not a measurement of the card — run over ALL cards, whole page.

## The page / component rows

```sql
-- site
SELECT id FROM sites WHERE domain='fundamentallyai.com';
-- 199733a8-ac9c-4c30-b2ce-65ecdac6f3bd
-- page
SELECT id, name, status, page_type, sections FROM pages WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd' AND name='platform-log-index';
-- e47ad594-e1a4-4d7a-ae54-07a3b4942a03 ; sections = ["hero","blog-listing"]
-- its components (the listing pc is 79d769e4-6c88-4a64-aa22-e0b2025dc55c)
SELECT pc.id, cc.name, (pc.rendered_html LIKE '%<a %') AS has_anchor
FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
WHERE pc.page_id='e47ad594-e1a4-4d7a-ae54-07a3b4942a03';
```

## The census (phantom site_specs aspects declared by components)

```sql
WITH decl AS (
  SELECT cc.id, cc.name, cc.function, f.key AS field, f.value->>'source' AS source
  FROM content_components cc, jsonb_each(cc.input_schema->'fields') f
  WHERE cc.is_active AND cc.input_schema ? 'fields'
), spec_sources AS (
  SELECT *, split_part(split_part(source,'.',2),'.',1) AS aspect
  FROM decl WHERE source LIKE 'site_specs.%'
), known AS (SELECT DISTINCT aspect FROM site_specs)
SELECT s.aspect, count(*) AS field_count, count(DISTINCT s.name) AS components,
       (k.aspect IS NOT NULL) AS aspect_exists_somewhere,
       string_agg(DISTINCT s.name, ', ') AS component_names
FROM spec_sources s LEFT JOIN known k ON k.aspect=s.aspect
GROUP BY s.aspect, (k.aspect IS NOT NULL)
ORDER BY aspect_exists_somewhere, s.aspect;
```
Gotcha: `jsonb_each` on a NULL/absent `fields` errors — keep the `? 'fields'` guard.
Gotcha: `aspect_exists_somewhere = true` does NOT mean it exists for the site a
given page belongs to; it only clears the aspect of being fleet-wide phantom.

## Declared query.* names vs the resolver vocabulary

```sql
SELECT DISTINCT f.value->>'source' AS source, count(DISTINCT cc.name)
FROM content_components cc, jsonb_each(cc.input_schema->'fields') f
WHERE cc.is_active AND cc.input_schema ? 'fields' AND f.value->>'source' LIKE 'query.%'
GROUP BY 1 ORDER BY 1;
```
Compare against the switch in
`platform/orchestration/actions/queryresolve/queryresolve.go` `Resolve()`.
Names take an optional `:arg`; compare on the base (before the first colon).

## The 090 diagnosis in flight (fired by the FILING thread — do not re-fire)

```sql
-- find by payload, never by printed id; absence for ~30 min is latency, not a drop
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = 'df8ca3a1-9cca-474a-88fb-19577e088080';
SELECT body FROM diagnosis_artifacts
WHERE correlation_id='df8ca3a1-9cca-474a-88fb-19577e088080' ORDER BY created_at;
```

## Structural-miss visibility (the 238/268 mechanism)

```sql
SELECT occurred_at, context->>'page_name', context->>'section', context->'fields'
FROM agent_error_log WHERE error_code='STRUCTURAL_KEY_CARRY_MISS'
ORDER BY occurred_at DESC LIMIT 20;
```
Gotcha: the table's time column is `occurred_at`, not `created_at`.

## The AT-REST census (2026-08-22) — the audit's own question, run by hand

This is the query the daily audit answers. Run it before and after any change to the
component library, and to reproduce the residual figures in NOTES.

**Scope, and why:** `content_components WHERE is_active` ONLY. `information_schema`
lists **56** tables carrying an `input_schema` column — every one but this and
`component_versions` is a dated `bak_*` / `*_backup_*` snapshot, and `component_versions`
is history. A census that globs the column name reports the same dead field dozens of
times and cannot be reconciled with anything.

```sql
-- live vocabulary, both halves
CREATE TEMP VIEW la AS SELECT DISTINCT aspect FROM site_specs;   -- no is_current filter:
-- an aspect with only superseded rows still names a store a writer produces
CREATE TEMP VIEW kq AS SELECT * FROM (VALUES
 ('adoption_tracker'),('adoption_tracker_full'),('blog_posts'),('business_directory'),
 ('health_insurer_directory'),('health_insurer_directory_full'),('latest_news'),
 ('model_directory'),('model_directory_full'),('mortgage_lender_directory'),
 ('mortgage_lender_directory_full'),('news_archive'),('pages_under_section'),
 ('pages_where_type'),('products'),('protocol_tracker'),('protocol_tracker_full'),
 ('savings_provider_directory'),('savings_provider_directory_full'),('section_index_for')) v(name);

CREATE TEMP VIEW iss AS
WITH f AS (
  SELECT cc.id, cc.name AS component, cc.created_at, cc.updated_at,
         fld.key AS field, COALESCE(fld.value->>'source','') AS source
  FROM content_components cc,
       LATERAL jsonb_each(CASE WHEN jsonb_typeof(cc.input_schema->'fields')='object'
                               THEN cc.input_schema->'fields' ELSE '{}'::jsonb END) AS fld
  WHERE cc.is_active),
c AS (SELECT *, split_part(source,'.',1) AS prefix,
             split_part(split_part(source,'.',2),':',1) AS second
      FROM f WHERE source NOT IN ('','llm','renderer','static'))
SELECT id, component, created_at, updated_at, field, source,
  CASE
    WHEN position('.' in source)=0
      OR prefix NOT IN ('renderer','static','site_specs','site_assets','pages','config','query')
      THEN 'prefix_outside_vocabulary'
    WHEN prefix='query'
      AND split_part(substring(source from 7),':',1) NOT IN (SELECT name FROM kq)
      THEN 'unregistered_query'
    WHEN prefix='site_specs' AND second NOT IN (SELECT aspect FROM la)
      THEN 'phantom_aspect'
    ELSE NULL END AS issue
FROM c;

-- the three classes
SELECT issue, count(*) fields, count(DISTINCT component) components
FROM iss WHERE issue IS NOT NULL GROUP BY 1 ORDER BY 2 DESC;

-- THE ONE THAT MATTERS: blast radius per component
SELECT i.component, count(*) AS bad_fields, string_agg(DISTINCT i.issue,',') AS issues,
       (SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id
         WHERE pc.component_id=i.id AND p.status IN ('active','deployed')) AS live_instances
FROM iss i WHERE i.issue IS NOT NULL
GROUP BY i.component, i.id ORDER BY live_instances DESC, bad_fields DESC;
```

**Gotchas, each one bought:**

- **`source` is matched EXACTLY as the guard matches it, bare values included.** A source
  of `config` with **no dot** is an offence (`plan_sections_action.go:623` → `len(parts)<2`
  → `(nil,false)` → `skip_field`), not a `config.*` source. `info-card-grid.carousel` is
  that shape and it is dropped on all **32** of its live instances. A census that treats
  the prefix as the whole answer misses it.
- **Query names take an optional `:arg`** — compare on the base, `split_part(...,':',1)`,
  or every `pages_where_type:tool` reads as unregistered.
- **`site_specs` has an alias fallback and it CANNOT rescue a phantom aspect.**
  `resolveSpecAlias` step 1 needs `identityContainerAspects[aspect]` populated; step 2 is
  `if aspect != "identity" { return nil, false }`. Verified at the source before claiming
  these resolve nowhere — the opposite mistake (a census flagging fields that actually
  work) is the one that gets a check switched off.
- **Count FIELDS or count NAMES, and say which.** §8 of the bug reports 7 unregistered
  query *names*; this query returns 14 *fields* over those same 7 names. Both are right.
  Comparing one to the other reads as the gate leaking and is the single easiest way to
  manufacture a false finding here.

## Deploying `component-source-vocabulary-check`, and PROVING it ran (council `a092d7d8`, debug_historian)

**Order matters and getting it wrong is invisible.** The image must exist at the pinned
tag BEFORE the overlay is applied: an `ImagePullBackOff` on this fleet reports as a Job
still **RUNNING**, never FAILED, so a check that has never once executed looks healthy.

```bash
# 1. Image first. Builds from committed HEAD (git archive) - commit before building.
make build-component-source-vocabulary-check
make push-component-source-vocabulary-check
# 2. Only then the manifest. Bump newTag in the overlay in the SAME commit as the build.
make deploy-component-source-vocabulary-check
```

**Then prove it at the artefact, in this order. Do not stop at the make target — it
reports success either way.**

```bash
# (a) Did the manifest actually take the tag you built? The jsonpath cannot be misread;
#     `configured` vs `unchanged` in the apply output is the cheap signal, this is the sure one.
kubectl -n ai-persona-system get cronjob component-source-vocabulary-check \
  -o jsonpath='{.spec.jobTemplate.spec.template.spec.containers[0].image}'; echo

# (b) Does the BINARY carry the mode? Probe a KNOWN value with a control in the same
#     breath - never `strings` (absent from these images), never a discovery grep.
POD=$(kubectl -n ai-persona-system get pods -l app=component-source-vocabulary-check \
        -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- \
  grep -aq -- '--component-source-vocabulary' /proc/1/exe && echo "PRESENT (expected)"
kubectl -n ai-persona-system exec "$POD" -- \
  grep -aq -- '--no-such-mode-should-exist' /proc/1/exe && echo "CONTROL FAILED - grep matches anything"

# (c) Trigger a run and read the POD's exit code. A Job is not a pod and a log line is
#     not an exit code.
kubectl -n ai-persona-system create job --from=cronjob/component-source-vocabulary-check csv-manual-1
kubectl -n ai-persona-system get pods -l job-name=csv-manual-1 \
  -o jsonpath='{.items[0].status.containerStatuses[0].state.terminated.exitCode}'; echo
# EXPECTED ON A HEALTHY ESTATE: 0. 1 = a red (read the doc_notes row, it says which of the
# four). 2 = the check could not run, which must NEVER be read as a pass.

# (d) The row it wrote. This is the positive control on the REPORT path - a pod that
#     exits 0 having written nothing is a broken reporter, not a clean estate.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -q -c "
SELECT created_at, left(body, 400) FROM doc_notes
 WHERE source='component-source-vocabulary-check' ORDER BY created_at DESC LIMIT 1;"
```

⚠ **`writeDocNote` is BEST-EFFORT — a failed write only warns, and the exit code still
carries the finding.** That is deliberate (a failure to RECORD must not become a failure
to REPORT), but it means step (d) is not optional: without it, a silently refused insert
looks exactly like a healthy quiet run. `doc_notes.subject_type` is CHECK-constrained to
`{tool, pipeline, experience, action, experience-pattern, landmine, component, decision}`
and `writeDocNote` sends `'pipeline'`, which is in the set (`[MEASURED 2026-08-22]` 1,895
existing rows) — raised by the council's guardian seat and verified rather than assumed.

**Running it by hand, with no cluster and no risk**, which is how every control in NOTES
was produced:

```bash
go build -o /tmp/cka ./cmd/config-key-audit
# state in, findings out; exit code is the verdict
/tmp/cka --component-source-vocabulary \
  --baseline docs/agent_docs/docs024_key_docs_latest/bugfix_309_unclickable_index_cards/component_source_baseline.json \
  < payload.json
```
The payload is `{"aspects": [...], "components": [{"id","name","input_schema","live_instances"}]}`.
Build one from the live DB with the `jsonb_build_object` query in the census section above.

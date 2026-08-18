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

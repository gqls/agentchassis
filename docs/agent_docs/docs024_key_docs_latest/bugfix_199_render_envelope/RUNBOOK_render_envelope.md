# RUNBOOK — `bugs_open/199` render-seam envelope guard

Every query and command here had to be got right at least once. The gotcha is attached to the
command, not left in a scrollback.

## The census — the population the render gate cannot speak for

Mirrors `missingRequiredLLMFields` + `SchemaContentFields` exactly. **Take the denominator in
the same query**: a zero in the numerator is otherwise indistinguishable from a broken predicate.

```sql
WITH cls AS (
  SELECT c.id, c.function, c.is_active,
    CASE
      WHEN c.input_schema IS NULL OR jsonb_typeof(c.input_schema) <> 'object'
           OR c.input_schema = '{}'::jsonb                    THEN 'A_no_schema_gate_skipped'
      WHEN jsonb_typeof(c.input_schema->'fields') = 'object'  THEN
        CASE WHEN EXISTS (SELECT 1 FROM jsonb_each(c.input_schema->'fields') f
                          WHERE f.value->>'source'='llm' AND (f.value->>'required')::text='true')
             THEN 'D_gate_can_speak_v2' ELSE 'B_v2_no_required_llm' END
      WHEN jsonb_typeof(c.input_schema->'properties') = 'object'
           AND c.input_schema->'properties' <> '{}'::jsonb    THEN
        CASE WHEN EXISTS (SELECT 1 FROM jsonb_array_elements_text(
                            COALESCE(c.input_schema->'required','[]'::jsonb)) r
                          WHERE COALESCE(c.input_schema->'properties'->r.value->>'source','llm')='llm')
             THEN 'D_gate_can_speak_legacy' ELSE 'C_legacy_no_required_llm' END
      ELSE 'A_unrecognised_dialect_gate_blind'
    END AS klass
  FROM content_components c
)
SELECT klass, count(*) AS components_total,
       count(*) FILTER (WHERE is_active) AS components_active,
       (SELECT count(DISTINCT pc.id) FROM page_components pc
          WHERE pc.component_id = ANY(array_agg(cls.id))) AS live_page_component_rows
FROM cls GROUP BY klass ORDER BY klass;
```

**Gotchas.** The table is `content_components`, not `components`. The legacy dialect defaults a
property with no explicit `source` to `llm` (`component_schema_fields.go:112-115`) — the
`COALESCE(..., 'llm')` above is that rule, not a guess. `input_schema = '{}'` unmarshals to a
zero-length map, so `len(comp.InputSchema) > 0` is false and the gate is skipped entirely —
that is why class A exists separately from class B.

## Is the envelope anywhere live?

Use the guard's own SQL twin, never a `::text LIKE`:

```sql
-- page_components, numerator and denominator together
SELECT count(*) FILTER (WHERE content_data->>'type'='text'
                          AND jsonb_typeof(content_data->'result')='string') AS envelope_rows,
       count(*) FILTER (WHERE content_data IS NOT NULL AND content_data <> '{}'::jsonb) AS non_empty,
       count(*) AS all_rows
FROM page_components;
```

**Gotcha.** `jsonb::text LIKE '%"type":"text"%'` matches **nothing** — jsonb renders a space
after the colon. Induce a non-zero before trusting any zero.

## Which component does a live envelope row sit on?

```sql
SELECT s.domain, p.name AS page, pc.slot_name, COALESCE(cc.function,'<no component_id>') AS fn,
       (SELECT array_agg(k ORDER BY k) FROM jsonb_object_keys(pc.content_data) k) AS keys
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
LEFT JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.content_data->>'type'='text' AND jsonb_typeof(pc.content_data->'result')='string';
```

**Gotcha.** `pages` has `name` and `url`, **not** `slug`. `page_components` has `slot_name`, not
`component_name`.

## ⚠ Do NOT use `page_component_history` to attribute a component

`page_component_history.component_id` is a FK to **`page_components(id)`**, `ON DELETE SET NULL`,
and `save_page_sections` DELETEs and re-INSERTs — so it is **NULL on 67 of 67** envelope rows.
A join to `content_components` returns NULLs, not zero rows, and a `CASE` testing
`input_schema IS NULL` before `id IS NULL` relabels every failed join as "component with no
schema". Gate on resolvability first:

```sql
SELECT count(*) FILTER (WHERE component_id IS NULL)     AS fk_nulled,
       count(*) FILTER (WHERE component_id IS NOT NULL) AS fk_live
FROM page_component_history
WHERE content_data->>'type'='text' AND jsonb_typeof(content_data->'result')='string';
```

## Has the text path fired lately? (bound per status — retention differs)

```sql
-- what window is actually available; whole-table min() lies (unreaped statuses set the floor)
SELECT status, count(*), min(updated_at)::timestamp(0), max(updated_at)::timestamp(0)
FROM orchestration_states GROUP BY status ORDER BY count(*) DESC;

-- envelope-shaped step outputs, by the step key that holds them
SELECT e.key AS step_key, count(*) AS occurrences,
       count(*) FILTER (WHERE (e.value->>'__json_contract_unmet')::text='true') AS declared_json_and_failed
FROM orchestration_states o, jsonb_each(o.collected_data) e
WHERE jsonb_typeof(e.value)='object'
  AND e.value->>'type'='text' AND jsonb_typeof(e.value->'result')='string'
GROUP BY 1 ORDER BY 2 DESC;
```

**Gotcha — the measurement must be able to see the trigger.** Before reporting "zero
`generated_content` envelopes", prove the key exists at all:
`count(*) FILTER (WHERE collected_data ? 'generated_content')` → 62. Otherwise a zero means
"my query cannot see this", not "it did not happen".

## Reading a live agent definition's steps

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps')
FROM agent_definitions
WHERE type='page-content-writer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

**Gotchas.** `workflow.steps` is an **object keyed by step name**, not an array —
`jsonb_array_elements` errors with *"cannot extract elements from an object"*. And the
`render_component` steps are **nested** inside
`process_sections_loop.config.sub_workflow.steps`, so a top-level scan for
`action='render_component'` returns **zero rows** and looks like "nothing uses it".

## Running the four named mutations

Back up, mutate, test, restore, and **diff to prove the restore** — on a shared tree a
half-restored mutation is someone else's broken build.

```bash
cp platform/orchestration/actions/v3_site_actions.go "$SP/v3.orig"
# ... python3 edit ...
go test ./platform/orchestration/actions/ -run TestX -count=1
cp "$SP/v3.orig" platform/orchestration/actions/v3_site_actions.go
diff "$SP/v3.orig" platform/orchestration/actions/v3_site_actions.go   # must be empty
```

**Gotcha.** One of the four mutations **passed**. That is a finding, not a test to relabel — see
NOTES. Guards in series absorb a local mutation.

## Post-roll verification

```bash
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "normalizeRenderContentEnvelope"'   # both replicas
kubectl -n ai-persona-system get pod <pod> -o jsonpath='{.status.startTime}'  # in the same breath
```

```sql
SELECT action, context->>'outcome', count(*) FROM agent_error_log
WHERE error_code='CONTENT_DATA_ENVELOPE' GROUP BY 1,2;
```

**Gotcha, and it is the important one.** **Zero `render_component` rows is the EXPECTED
reading** — the measured trigger rate is zero. An unrolled image and a working guard produce
the identical count, which is why the pod-grep must come first and why the pod's start time is
read in the same command.

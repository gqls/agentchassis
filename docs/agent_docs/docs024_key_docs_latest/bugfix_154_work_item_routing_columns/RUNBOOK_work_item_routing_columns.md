# RUNBOOK — `bugs_open/154`, work-item routing columns

Every command that was hard to get right, with its gotcha attached.

---

## Is this bug still biting? (one query)

```sql
SELECT left(id::text,8) AS id, status, created_by,
       component_id IS NOT NULL AS col_set,
       spec ? 'component_id'    AS spec_has,
       left(item_key,45) AS item_key, created_at::date
FROM site_work_items
WHERE item_type='improve_tool'
ORDER BY created_at DESC LIMIT 20;
```

**Gotcha:** the discriminator is `col_set=t AND spec_has=f`. Do NOT read it as
"component_id is empty" — the column is populated *precisely* on the rows that
fail. The failing value is `input_data.component_id` inside the orchestration,
which is a different thing from the column, and conflating the two is what makes
this bug read as nonsense.

## The census that sizes it fleet-wide

```sql
SELECT 'component_id' AS col,
       count(*) FILTER (WHERE component_id IS NOT NULL) AS col_set,
       count(*) FILTER (WHERE component_id IS NOT NULL AND NOT (spec ? 'component_id')) AS col_set_spec_missing,
       count(*) FILTER (WHERE component_id IS NULL AND spec ? 'component_id') AS spec_only
FROM site_work_items;
```

Swap the column name for `entity_id` / `affected_url` / `page_id`. **Run the
`page_id` row before touching `page_id`** — it is the one that says widening it
would newly expose 218 rows.

## Will delivering the value actually help? (the anti-inert check)

`load_tool` queries `content_components`, and `page_components.id` is a
*different* id. If the column held the latter, the fix would change the error
message and nothing else.

```sql
SELECT left(w.id::text,8) AS item,
       EXISTS(SELECT 1 FROM content_components c WHERE c.id=w.component_id) AS in_content_components,
       EXISTS(SELECT 1 FROM content_components c WHERE c.id=w.component_id AND c.is_active) AS active,
       EXISTS(SELECT 1 FROM page_components p JOIN content_components c ON c.id=p.component_id
              WHERE p.component_id=w.component_id AND c.is_active) AS joins_to_a_page
FROM site_work_items w
WHERE w.item_type='improve_tool' AND w.component_id IS NOT NULL;
```

All four must be `t` — they are the clauses of `load_tool`'s own query.

## Read the LIVE config, never the seed

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c "
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'load_tool')
FROM agent_definitions
WHERE type='tool-improver' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
```

**Gotcha:** the `is_active AND NOT is_snapshot AND deleted_at IS NULL` triple is
load-bearing — without it you get snapshots and read a definition that is not
what runs. `docs/agent_docs/sql_for_agents/*tool_improver*.sql` are **history**,
not the system.

The dispatcher's mapping is nested inside a `loop` sub-workflow:

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps'->'process_item')
FROM agent_definitions WHERE type='build-dispatch-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Who else reads the field you are about to widen

```sql
SELECT type FROM agent_definitions
WHERE default_config::text LIKE '%current_item.spec.component_id%'
  AND deleted_at IS NULL AND is_active AND COALESCE(is_snapshot,false)=false;
```

**Gotcha:** this finds *dispatcher* references only. A handler reads its own
inputs under a different spelling — `rerender-pages` reads
`input_data.spec.component_id` — so grep the handler's steps too:

```sql
SELECT jsonb_pretty(default_config->'workflow'->'steps')
FROM agent_definitions WHERE type='rerender-pages' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- then grep the output for the field name
```

That is how `create_rerender_items`' scoping gate was found, which is what
rejected the spec-backfill design.

## Prove the test is not vacuous

```bash
cp platform/orchestration/actions/load_work_item_actions.go /tmp/lwi.bak
perl -0pi -e 's/\t\tsetRoutingField\(item, "component_id", uuidPtrString\(componentID\), logger\)\n//' \
  platform/orchestration/actions/load_work_item_actions.go
go test ./platform/orchestration/actions/ -run 'RoutingColumns'   # MUST fail, on BOTH arms
cp /tmp/lwi.bak platform/orchestration/actions/load_work_item_actions.go
go test ./platform/orchestration/actions/ -run 'RoutingField|RoutingColumns'
```

## Order of the two halves — this is not optional

The coalesce lives in the **binary**; the path that uses it lives in the **DB**.

1. Commit, `make build-agent-chassis` (builds from committed HEAD), push, deploy.
2. **Pod-grep the running binary** for a symbol the change added, with a positive
   control in the same exec so a false green is visible:

```bash
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "routing field left unset"; \
   strings /app/agent-chassis | grep -c "LoadWorkItemsAction: Starting"'
```

The first is the new symbol (expect ≥1), the second the positive control (a
pre-existing string — if *that* is 0 the grep itself is broken and the first
number means nothing). Repeat on **every replica**: a roll is not evidence.

3. **Only then** apply the config:

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,component_id?}',
      '"current_item.component_id"')
WHERE type='build-dispatch-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

**Gotcha:** the key really is `component_id?` **with the question mark** — the
`?` is the optional marker and is part of the JSON key. (Distinct from
`bugs_open/134`, where a `?` leaked into a `params` key, where it means nothing
and makes the key inert. In `input_mapping` it is real.)

Flipping this **before** the image ships strands the 235 spec-only rows.

## Verify the fix at the artefact

```sql
SELECT current_step, status FROM orchestration_states
WHERE owner_agent_type = 'tool-improver' ORDER BY created_at DESC LIMIT 3;
```

**Retention trap (154's own warning, and it bit me):** `orchestration_states`
history is short — the rows for a diagnosis fired at 18:14 were gone by ~18:40.
Capture evidence when it happens; do not plan to query it back. `site_work_items`
and `diagnosis_artifacts` outlive it.

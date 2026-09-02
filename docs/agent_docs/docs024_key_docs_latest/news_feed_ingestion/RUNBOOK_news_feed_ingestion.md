# RUNBOOK — news_feed_ingestion

## Checking who owns a bug in this lane's territory before touching it

```bash
python3 scripts/who-owns.py <number|slug>
```
Reads commits only — a session mid-fix with nothing committed yet is invisible.
Cross-check `git status --short` for dirty files touching the same paths/tables.

## Live counts on `content_feed_items` (as of 2026-09-02, re-run before citing)

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE entity_ids        IS NOT NULL) AS entity_ids_set,
       count(*) FILTER (WHERE duplicate_of      IS NOT NULL) AS duplicate_of_set,
       count(*) FILTER (WHERE published_page_id IS NOT NULL) AS published_page_id_set,
       count(*) FILTER (WHERE relevance_score   IS NOT NULL) AS relevance_score_set_control
FROM content_feed_items;
```
`[MEASURED 2026-09-02]` 14,013 total | entity_ids 0 | duplicate_of 0 |
published_page_id 15 | relevance_score 12,281 (control).

## Table shape

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "\d content_feed_items"
```
`entity_ids` is `uuid[]`, **no FK** — do not assume it points anywhere without
checking again; nothing today declares what it references.

## Reading an agent's live workflow without dumping the whole prompt text

`default_config->'workflow'` on a big agent (LLM prompt templates included) can be
hundreds of KB — `jsonb_pretty` on the whole thing floods the terminal. Pull step
shape only:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A \
  -c "SELECT jsonb_pretty(default_config->'workflow'->'steps') FROM agent_definitions WHERE type='<type>' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" \
  > /tmp/wf.txt
grep -noE '"(action|next_step|start_step|condition_field|default)"[^,]{0,80}' /tmp/wf.txt
```
Step chaining is `step_id -> {action, config, next_step, output_field}`, entry at
top-level `start_step`. `evaluate_condition`'s config carries
`condition_field` (dot-path into collected_data), `conditions` (map of stringified
value -> next_step) and `default` (fallback next_step).

## feed-triage workflow (as read 2026-09-02, before this lane's changes)

`load_items` (load_feed_items_for_triage) → `check_has_items` (evaluate_condition
on `pending_items.count`; `"0"` → `complete`, default → `read_site_spec`) →
`read_site_spec` → `score_relevance` (execute_llm_prompt) → `apply_scores`
(apply_feed_scores) → `complete`.

## Which agent runs an action, without reading every agent

```sql
SELECT type, display_name, is_active FROM agent_definitions
WHERE default_config::text LIKE '%<action_name>%'
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Council review scope for this lane's fix

Touches `platform/orchestration/actions/` (in scope) and an appliable migration
under `docs/agent_docs/sql_for_agents/` (in scope, widened 2026-08-19 per bug
314). Submit per CLAUDE.md's council process before/alongside committing.

# RUNBOOK — bugfix_342_absent_required

## Prove what the fleet is running (per SERVICE, and at the artefact)
```bash
kubectl get pods -n ai-persona-system -l app=agent-chassis \
  -o custom-columns=NAME:.metadata.name,IMAGE:.spec.containers[0].image --no-headers
# startup provenance line scrolls on agent-chassis within hours — go to the binary:
kubectl -n ai-persona-system exec <pod> -- sh -c \
  'grep -aq "<expected-sha>" /proc/1/exe && echo PRESENT; grep -aq "nonsense0000" /proc/1/exe || echo CONTROL_OK'
git merge-base --is-ancestor <your-commit> <expected-sha>   # "did my fix ship?" is a query
```
Gotcha: the stamp sha for a tag is NOT in git — take it from a pod probe or another lane's
verified note, then CONFIRM in the binary with a nonsense control in the same breath.

## Census: who arms a config key (0 rows = dormant)
```sql
SELECT type FROM agent_definitions
WHERE default_config::text LIKE '%record_absent_required_fields%'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## The absent-required census — WITH the vacuity guard
```sql
WITH reqs AS (
  SELECT cc.id AS comp_id, cc.function, f.k AS field
  FROM content_components cc,
       jsonb_each(COALESCE(cc.input_schema->'fields', cc.input_schema->'properties','{}'::jsonb)) f(k,v)
  WHERE cc.is_active AND v->>'source'='llm' AND (v->>'required')::boolean IS TRUE
)
SELECT
  (SELECT count(*) FROM site_components sc JOIN reqs r ON r.comp_id = sc.component_id) AS candidate_pairs,  -- MUST be > 0 for a zero below to mean anything
  (SELECT count(*) FROM site_components sc JOIN reqs r ON r.comp_id = sc.component_id
    WHERE sc.content_data IS NULL OR NOT (sc.content_data ? r.field)
       OR sc.content_data->>r.field IS NULL OR sc.content_data->>r.field = '') AS rows_missing;
```
Gotchas learned the hard way:
- **A zero over zero candidates is not a finding** — always select the candidate count in
  the same statement (each psql statement is separate; a CTE does not survive the `;`).
- Swap `site_components` → `page_components` for the page-side census. That one answers the
  WRITER question; the seam's render-time report is a strict subset (fleet defaults fill
  some fields). Do not quote one as the other.

## Live step config for an agent
```sql
SELECT s.k, s.v->>'action', COALESCE(s.v->>'error_step','(none)')
FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s(k,v)
WHERE a.type='section-editor' AND a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;
```

## Optional-key budget (before adding any key to an input spec)
```bash
./scripts/audit-optional-key-budget.sh --json | python3 -c "..."   # ConfigKeys are NOT counted; Optional is
go test ./cmd/config-key-audit/ -run OptionalBudgetCronParity      # only if check.py touched
```

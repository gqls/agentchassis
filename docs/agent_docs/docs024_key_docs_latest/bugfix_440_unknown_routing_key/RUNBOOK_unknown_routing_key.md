# RUNBOOK — bugfix 440

## The census (re-run before quoting; every figure dates fast)
```sql
SELECT COALESCE(spec->>'reason','(none)') AS reason, count(*),
       min(created_at)::date, max(created_at)::date
FROM site_work_items WHERE item_type='page_rerender'
GROUP BY 1 ORDER BY 2 DESC;
```

## Counting warning EMISSIONS (not string presence) — the trap that bit on day one
A text-LIKE over `collected_data` matches council payloads QUOTING the string. Exclude the
quoting population and read one member:
```sql
SELECT orchestration_id, current_step FROM orchestration_states
WHERE updated_at >= '<since>'
  AND collected_data::text LIKE '%not in the sections-rerender vocabulary%'
  AND collected_data->'input_data'->>'fix_correlation_id' IS NULL;
```

## Is the warning capability in the running binary
```bash
POD=$(kubectl -n ai-persona-system get pod -l app=agent-chassis -o name | head -1)
kubectl -n ai-persona-system exec $POD -- grep -aq "not in the sections-rerender vocabulary" /proc/1/exe && echo LIVE
```

## The live gate's condition (schema first; the declaration auditor holds it daily)
```sql
SELECT default_config->'workflow'->'steps'->'check_rerender_mode'
FROM agent_definitions WHERE type='page-rerender' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Verifying a phase of THIS lane shipped — inert code CANNOT be probed by literal
Phase 1a (and any future zero-caller phase) is stripped by dead-code elimination: the standard
three-way `/proc/1/exe` probe reads ABSENT with clean controls even when the commit shipped
(LANDMINES entry, 2026-09-02). Verify by ancestry instead:
```sql
SELECT DISTINCT pod_name, git_commit FROM service_binary_capabilities
WHERE pod_name LIKE 'agent-chassis-%' ORDER BY 1;
```
```bash
git merge-base --is-ancestor a3758c399 <stamp> && echo "phase 1a in this build"
```
Once phase 1b lands (first caller), the literal probe becomes valid:
`grep -aq "input_data.spec.routing_reason" /proc/1/exe` — but only from that phase on, and its
first PRESENT reading dates the CALLER's roll, not the foundation's.

## Dry-running a config migration against the LIVE database, safely (the strongest check there is)
The migration's own `DO`-block verify runs, the JSON manipulation is exercised, and the pasted
strings can be read back out of `agent_definitions` — then it is all discarded. Strip the file's
own `BEGIN`/`COMMIT` so your transaction is the only one, and make the last word `ROLLBACK`.
```bash
strip() { sed -e 's/^BEGIN;$//' -e 's/^COMMIT;$//' -e "s/^SET LOCAL lock_timeout = '5s';$//" "$1"; }
{ echo "SET lock_timeout = '5s';"; echo 'BEGIN;'
  strip docs/agent_docs/sql_for_agents/741_..._HOLD.sql
  cat  docs/agent_docs/sql_for_agents/741_..._HOLD_VERIFY.sql      # read-only, runs against the applied state
  strip docs/agent_docs/sql_for_agents/741_..._HOLD_ROLLBACK.sql   # round trip: does the original come back?
  echo 'ROLLBACK;'; } |
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1
```
⚠ **Then prove the tree is clean** — a killed `kubectl exec` aborts the transaction server-side,
but verify rather than assume:
```sql
SELECT (SELECT default_config #>> '{workflow,start_step}' FROM agent_definitions
         WHERE type='page-rerender' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL)
    || ' | constraint=' || (SELECT count(*) FROM pg_constraint
         WHERE conname='chk_page_rerender_routing_reason_vocabulary')::text;
-- pre-apply EXPECT: check_rerender_mode | constraint=0
```
⚠ **Run it TWICE.** The `ALTER TABLE ... ADD CONSTRAINT` needs `ACCESS EXCLUSIVE` on
`site_work_items` and the wait is probabilistic — measured 2 ms once and >2 minutes another time
on the same statement. A single fast run is not evidence that it is fast. See LANDMINES.

## Proving the pasted strings match livespec — at the ARTEFACT, not in the file
Diffing the file text only proves you typed it right. This proves the dollar-quoting and the JSON
escaping did not alter it in transit: apply into a transaction, read the values back, diff against
the Go renderers, roll back.
```bash
# inside the transaction, before ROLLBACK:
#   SELECT default_config #>> '{workflow,steps,check_routing_key_known,config,condition}' ...
#   SELECT default_config #>> '{workflow,steps,check_rerender_mode,config,condition}' ...
# then, in Go: print CheckRoutingKnownConditionClause(), TransitionRerenderModeConditionClause(),
#   RefuseUnknownRoutingKeyMessageTemplate(), RefuseUnknownRoutingKeyMessageFallback()
# and diff the two lists. Verified IDENTICAL, all four, 2026-09-03.
```
⚠ psql prints `Output format is unaligned.` into the stream after `\pset` — strip it before diffing
or you get a one-line false mismatch.

## Inducing a refusal (this is what CLOSES 440 — a census of zero cannot)
A zero-refusal census is equally consistent with "nothing bad written yet" and "the branch is
unreachable". After 741 applies, mint one item with a key that is deliberately not in the
vocabulary and watch where it lands. ⚠ The CHECK constraint will refuse the INSERT at the write
door, which is itself half the proof — to exercise the READ door you need a row that predates the
constraint, so induce it by `UPDATE`ing a test item's spec inside a transaction, or add the row
before applying 741.
```sql
-- read door: the item must end at needs_human_review with the key named in `error`
SELECT status, left(error, 160) FROM site_work_items WHERE id = '<the induced item>';
-- write door: this INSERT must FAIL with 23514 once 741 is applied
INSERT INTO site_work_items (site_id, item_type, spec, status)
VALUES ('<site>', 'page_rerender', '{"page_name":"x","routing_reason":"tool_retirement"}'::jsonb, 'detected');
```

## Testing a pattern-check check on ONE file (the script's CLI cannot do it)
`python3 scripts/pattern-check.py <file>` ignores the filename and lints the git INDEX — it prints
nothing and exits 0, which is indistinguishable from a pass. Call the function, with a positive
control in the same breath:
```python
import importlib.util
spec = importlib.util.spec_from_file_location('pc', 'scripts/pattern-check.py')
m = importlib.util.module_from_spec(spec); spec.loader.exec_module(m)
print(m._rerender_vocabulary())            # -> (set_of_five, '')  — '' means it READ the Go source
f = []; m.check_rerender_routing_key(['docs/agent_docs/sql_for_agents/YOURS.sql'], None, f); print(f)
```
`migration_is_lintable()` takes a **basename**, never a path.

# RUNBOOK — bugs_open/321 (loop-filed work items collide on a site-wide item_key)

Commands that were hard to get right, with their gotchas. Lane: `bugfix_321_item_key_collisions`.

## The fleet census (which steps are in the class)

```bash
./scripts/audit-loop-sitewide-item-keys.sh          # human-readable
./scripts/audit-loop-sitewide-item-keys.sh --json   # machine-readable
```
Gotcha: read refusal from **empty stdout**, never the exit code — `go run` folds the
tool's exit 2 into its own 1 (the WFA-013 gotcha, inherited here). The script does this
for you; if you drive the Go tool by hand, you must.

Raw SQL census (no Go needed — but it is a SECOND descent; prefer the script):
```sql
-- create_work_item steps inside loops, with/without the suffix — see
-- PLAN_2026-08-19_item_key_collisions.md for the verified 2026-08-19 result (6 steps, 4 lacking)
WITH live AS (SELECT type, default_config FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL),
loops AS (SELECT l.type, v AS loopstep FROM live l, LATERAL jsonb_path_query(l.default_config,'strict $.**') v
   WHERE v ? 'action' AND v->>'action'='loop'),
inner_steps AS (SELECT lp.type, lp.loopstep->'config'->>'item_variable' AS item_var, s.key AS step_name, s.value AS step
    FROM loops lp, LATERAL jsonb_each(COALESCE(lp.loopstep->'config'->'sub_workflow'->'steps',
                                               lp.loopstep->'config'->'substeps','{}'::jsonb)) s)
SELECT type, item_var, step_name, step->'config'->>'item_key_prefix' AS prefix,
       COALESCE(step->'config'->>'item_key_suffix_field','** MISSING **') AS suffix_field
  FROM inner_steps WHERE step->>'action'='create_work_item' AND step->'config' ? 'item_key_prefix';
```
Gotcha: this hand-rolled query reads `sub_workflow` with a `substeps` COALESCE — the Go
mode's `WalkSteps` handles the substeps-wins precedence properly; trust the script when
they disagree.

## The loss measurement (suggestions vs items — the pairing)

```sql
-- Pair each LLM answer against the items created within 5 minutes. Domain comes from
-- the RENDERED PROMPT ('Domain: <x>'), not from a join to orchestration_states — the
-- jsonb join on collected_data times out (>120s); the prompt regex covers all rows.
WITH ans AS (
  SELECT l.created_at, (regexp_matches(l.prompt_rendered,'Domain:\s*(\S+)'))[1] AS dom,
         jsonb_array_length((substring(l.response_text from '\{[\s\S]*\}'))::jsonb->'suggestions') AS suggested
  FROM llm_call_log l
  WHERE l.agent_type='tool-suggester' AND l.step_name='suggest_tools' AND l.success
    AND l.created_at > '<SINCE>')
SELECT a.created_at, a.dom, a.suggested,
  (SELECT count(*) FROM site_work_items w JOIN sites s ON s.id=w.site_id
    WHERE s.domain=a.dom AND w.item_type='add_tool'
      AND w.created_at BETWEEN a.created_at AND a.created_at + interval '5 min') AS created
FROM ans a ORDER BY a.created_at;
```
**PASS: created = suggested, per answer.** Baseline on record (pre-fix): 2026-08-19
10:25 gamesdesign.co.uk **7 suggested → 1 created**; all-history 40 → 11 (~72% lost).
This is the joint check with the `bugs_open/313` lane too: once their 490 revives
internal-linker, an N-link plan must produce N `content_rewrite` items with keys
`internal_link_<domain>_<source_page>`.

## Key-shape check (after any post-fix run)

```sql
SELECT item_key, status, created_at FROM site_work_items
 WHERE item_type='add_tool' AND created_at > '<FIX_APPLIED: 2026-08-19 16:05>'
 ORDER BY created_at;
-- expect add_tool[_novel]_<domain>_<function>; one row per suggested tool.
```
Gotcha: hand-filed rebuild rows (the webdesign lane) bypass the workflow config and
still carry old-shape keys — check `source='tool-suggester'` before reading one as a
regression.

## The hard-error tripwire (run for a week after 2026-08-19)

```sql
SELECT updated_at, error FROM orchestration_states
 WHERE status='FAILED' AND error LIKE '%item_key_suffix_field%'
   AND updated_at > '2026-08-19 16:05';
-- any row = a suffix path failed to resolve. For tool-suggester that now costs ONE
-- iteration (continue_on_error), durably recorded in collected_data as
-- create_items_loop_iter_<n>_error and folded into items_created with status "error".
```

## Migration 493 (applied 2026-08-19 16:05Z)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
  < docs/agent_docs/sql_for_agents/493_loop_nested_item_key_suffixes.sql
```
- Proven fail-loud by a second apply: aborts at the md5 pre-gate
  (`create_items_loop subtree changed`). That md5 anchor also means the file is
  **not re-runnable after ANY edit to that subtree** — re-derive anchors, never force.
- **Snapshots land in `agent_definitions_backup`** (columns `snapshot_taken_at`,
  `snapshot_reason`), NOT in `agent_definitions` with `is_snapshot=true`. A check for
  `is_snapshot=true` rows finds nothing and reads as "snapshot_agent is broken" — it
  is not. `SELECT type, snapshot_reason, snapshot_taken_at FROM agent_definitions_backup
  WHERE snapshot_reason LIKE '493%';`
- Rollback: `493_loop_nested_item_key_suffixes_ROLLBACK.sql` — value-gated `#-`
  removals; reinstates the collision, so only for a defect the suffix itself causes.

## The CronJob (live since 2026-08-19, 07:55 UTC daily)

```bash
make loop-sitewide-item-key-check-now     # manual run
kubectl -n ai-persona-system logs job/<job-name>   # the JSON report
# every run writes ONE doc_notes row, clean or not; a MISSING row = the job did not run:
# SELECT created_at, left(body,200) FROM doc_notes
#  WHERE source='loop_sitewide_item_key_check' ORDER BY created_at DESC LIMIT 7;
```
Image rides the fleet release (RELEASE_IMAGES + AGENT_DEPLOY_SERVICES, Decision B) —
verified: the 2026-08-19 v1.0.1316 release built, pushed and retagged it unprompted.

## Demand control for the detector (proves it can still see)

```bash
# export the live fleet, strip every suffix, expect exactly the loop-nested census (6 today):
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc \
 "SELECT jsonb_agg(jsonb_build_object('type',type,'workflow',default_config->'workflow'))
    FROM agent_definitions WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false
     AND is_active AND default_config ? 'workflow';" > /tmp/fleet.json
python3 - <<'EOF' > /tmp/fleet_stripped.json
import json
fleet=json.load(open('/tmp/fleet.json'))
def strip(d):
    if isinstance(d,dict):
        if d.get('action')=='create_work_item' and isinstance(d.get('config'),dict):
            d['config'].pop('item_key_suffix_field',None)
        for v in d.values(): strip(v)
    elif isinstance(d,list):
        for x in d: strip(x)
strip(fleet); print(json.dumps(fleet))
EOF
go run ./cmd/config-key-audit --loop-sitewide-item-keys < /tmp/fleet_stripped.json
# 2026-08-19: reported exactly 6 findings (the full class), 0 on the unstripped export.
```

## Dispatching a canary tool-suggester run

Topic is `system.agent.scheduled.requests` — NOT `system.agent.generic.requests`,
which resolves this agent type to a no-op stub (the 184 lane burned three dispatches
learning this; envelope shape documented in bugs_open/184's 2026-08-04 notes).
Publish via the container COMMAND with an `echo PUBLISH_OK` control — piped
`kubectl run -i | kcat -P` silently drops ~4/5 messages at exit 0 (LANDMINES).

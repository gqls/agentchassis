# RUNBOOK — bugfix 234

Commands that were hard to get right, with their gotchas. Update HERE when one changes.

## The carrier census — ALL DEPTHS, or the number is wrong

Top-level `->'workflow'->'steps'` walks miss loop sub-workflows (the 08-09 landmine; 356's
`commit_from` had 3 of 6 carriers nested). Use the recursive walk:

```sql
WITH all_steps AS (
  SELECT ad.type, jsonb_path_query(ad.default_config,
         'strict $.**?(@."action" == "create_work_item")') AS step
  FROM agent_definitions ad
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL)
SELECT type, step->'config'->'spec', step->>'description'
FROM all_steps WHERE step->'config' ? 'spec';
-- expected BEFORE migration: 3 rows; AFTER: 0 rows
```

Gotcha (cost me a minute on 2026-08-09): grouping that census by `(type, key)` and then
counting ROWS undercounts — improvement-loop carries `spec` twice, one row in the group. Read
the count column, not the row count.

Escaping: inside a `bash -c`/heredoc to psql, `\$` the jsonpath's `$` or the shell eats it.

## Damage / proof-at-a-filed-row

```sql
SELECT item_key, spec, created_at, status FROM site_work_items
WHERE created_by='improvement-loop' AND item_key LIKE 'improvement_rerender%'
ORDER BY created_at DESC LIMIT 3;
```
- BEFORE the migration: every spec `{}` (16/16 as of 2026-08-09, first row 2026-08-01).
- The fix is proven only by a row **filed after** the migration carrying
  `{"refresh_site_components": true}`. A definition that LOOKS right is exactly this bug.
- Natural rate ~1.8 rows/day. The currently-triaged `improvement_rerender_dartsonline.com`
  row (2026-08-09 13:19Z) predates the fix and stays empty — do not hand-edit it.

## Pod-grep (code half, post-roll)

```bash
for p in $(kubectl get pods -n ai-persona-system -l app=agent-chassis -o name); do
  kubectl exec -n ai-persona-system ${p#pod/} -- sh -c \
    'strings /app/agent-chassis | grep -c "bugs_open/234"; strings /app/agent-chassis | grep -c "zzz_no_such_string_234"'
done
# expect: N>0 then 0 (the second is the pipeline control), EVERY replica
```
The `StrictConfig` bool itself cannot be strings-proven. Live proof = canary: seed a
throwaway active agent whose `create_work_item` step carries a bogus key, watch
ValidateWorkflow reject it in the pod log, delete the canary. (Unit test pins it in CI
either way.)

## Migration apply

Per migration-runner practice: dry-run the runner first, this session, and scope the dir on
`--apply` — it takes EVERY pending file otherwise. This lane applied by hand + `--record-only`.
Re-check the next free number immediately before writing the file: three numbers were claimed
by other sessions while 356 was being written.

## The strict canary (code-half behavioural proof) — and its two traps

```bash
./witness_234_fire.sh     # seeds a throwaway agent with a bogus create_work_item key, dispatches once
./witness_234_poll.sh     # v2: greps the SAME capture it just took (see trap 1)
# expect: pod log "Invalid workflow configuration" naming zzz_strict_witness_234,
#         and NO row (item_type='strict_witness_234')
```

- **TRAP 1 — a marker without its evidence.** Poller v1 wrote pod logs to files and grepped
  the FILES in a later statement; pod deletion raced the two, so it recorded
  "REJECTION SEEN" while the matching lines were gone. That log line is worthless as proof
  and was discarded. v2 greps the capture in the same iteration and copies it to
  `witness_234_evidence.log`.
- **TRAP 2 — row absence is NOT the proof.** The spawn→call handshake drops roughly half of
  all dispatches fleet-wide, so "no row" is equally consistent with "never ran". The pod log
  line is the positive signal.
- **Re-firing:** the fire script refuses if `strict-witness-234` exists in ANY state
  (including deactivated), so `DELETE FROM agent_definitions WHERE type='strict-witness-234';`
  first.
- **Always deactivate/delete it afterwards** — while active it carries a bogus
  `create_work_item` key and pollutes the all-depths census that the RFC_021 Q1 adoption
  protocol depends on.
- **Precondition:** the fleet must actually be processing. On 2026-08-10 three firings
  produced nothing because the whole in-cluster fleet was stopped on an account-level
  Anthropic cap. Check `SELECT max(created_at) FROM orchestration_states;` is recent before
  blaming the canary.

## The automated check (RFC_021 Q1)

```bash
kubectl -n ai-persona-system create job rck-manual-$(date +%s) --from=cronjob/removed-config-keys-check
kubectl -n ai-persona-system logs job/<name>
# clean run still writes a doc_notes row: a MISSING row means the job did not run
SELECT body FROM doc_notes WHERE subject_key='removed-config-keys' ORDER BY created_at DESC LIMIT 1;
```

## Proving a build carries your change (the CORRECT method, learned the hard way)

The chassis's own `build provenance` log line is a STARTUP line and rotates away within
minutes — CLAUDE.md's "ask the service what it is running" recipe returns nothing on this
service (see LANDMINES). And grepping a pod or image for **your** commit sha always fails:
the binary carries **one** sha, its build point, not every ancestor.

```bash
# 1. extract the binary from the image (no cluster needed)
CID=$(docker create docker.io/aqls/agent-chassis:<tag>)
docker cp $CID:/app/agent-chassis /tmp/chassis-<tag>; docker rm $CID

# 2. find the STAMP by testing recent commits — a bounded set, never a discovery grep
for sha in $(git log --format=%H -12); do
  grep -aq "$sha" /tmp/chassis-<tag> && echo "STAMP=$sha" && break
done
# 3. controls, same breath: one that must be absent
grep -aq "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef" /tmp/chassis-<tag> && echo "UNRELIABLE"

# 4. the actual question, as a query rather than an inference
git merge-base --is-ancestor <your-commit> <STAMP> && echo "your change is in this image"
```

⚠ **`git rev-parse HEAD` is NOT the stamp**, even seconds after your own build: ~30 sessions
share this tree and HEAD moves between the build and your check. Measured 2026-08-11 — the
stamp was my makefile-bump commit, while HEAD had already advanced. Read the stamp from the
binary; never assume it.

⚠ **Prefer BEHAVIOUR where you can get it.** Provenance proves the code is *present*; a
witness firing proves it is *running*. For this lane the strict canary and the spec witness
did in one dispatch each what no amount of binary probing can establish.

## The full key inventory (the guardian's round-4 ask — run this before ANY StrictConfig flip)

"0 unrecognised keys" is a *count*; the guardian asked for the *inventory*, which is the
stronger artefact: it shows every key every live caller actually sets, and which declaration
recognises each one. Run it before flipping strict on any action — a count can be right while
you have no idea what the callers are doing.

```sql
WITH s AS (SELECT ad.type, jsonb_path_query(ad.default_config,
             'strict $.**?(@."action" == "<ACTION>")') AS step
  FROM agent_definitions ad
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL),
k AS (SELECT key, count(*) AS steps_using, count(DISTINCT type) AS agents
      FROM s, LATERAL jsonb_object_keys(step->'config') key GROUP BY 1)
SELECT key, steps_using, agents FROM k ORDER BY 1;
```
Then classify each key against the action's `ActionInputSpec` by hand — Required/Optional ·
ConfigKeys · DeprecatedConfigKeys · framework (`IsFrameworkStepConfigKey`) · anything else is
what strict would hard-fail on.

**`create_work_item`, measured 2026-08-11T11:31Z — 17 distinct keys, 17 steps, 14 agents,
nothing unrecognised.** The widest callers set the same 8-9 keys (`item_type`,
`handler_agent`, `severity`, `summary` on all 17; `site_id`, `priority`, `item_pipeline`,
`item_key_prefix` on 16). The long tail is small and all declared: `spec_data` 7,
`spec_literal` 3, `status` 3, `spec_paths`/`page_id`/`component_id`/`item_key_suffix_field`/
`recurrence_expected` 2 each. `item_domain` (the deprecated alias) now has **zero** carriers.

⚠ This is a POINT-IN-TIME snapshot of a shared, live table. It is evidence that the flip was
safe *when made*, not a standing guarantee — which is exactly why the ongoing guard is the
daily `removed-config-keys-check`, not this query.

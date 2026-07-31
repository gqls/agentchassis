# RUNBOOK — bugs_open/099 candidate 2

Every command here had a gotcha worth writing down. Where one did, the gotcha is
attached to it rather than kept separately.

## Find which open bugs no thread is working

`scripts/who-owns.py <n>` is the documented first stop, but on this tree it says
**"OWNED or recently active"** for almost everything — it counts any *mention* in an
active workstream dir, and there are 40 active lanes. It even prints that verdict
alongside "likely OWNING workstream(s): (none identified)". Use it to read the
commit list it prints, not the verdict.

What discriminates is the last commit touching the bug FILE:

```bash
for f in bugs_open/*.md; do
  echo "$(git log -1 --format='%ad' --date=format:'%m-%d %H:%M' -- "$f") | $(basename $f)"
done | sort
```

Then cross-check the quiet ones against the live queue and the day's commits:

```sql
SELECT item_type, status, LEFT(summary,90), updated_at FROM site_work_items
 WHERE status NOT IN ('complete','cancelled','rejected')
   AND (summary ILIKE '%<mechanism>%') ORDER BY updated_at DESC LIMIT 20;
```

```bash
git log --since="2 days ago" --name-only --pretty=format: | grep -E '^bugs_(open|closed)/' | sort -u
```

**Gotcha:** do NOT try to extract bug numbers from commit subjects with a bare
`grep -oE '\b[0-9]{3}\b'` — it matches version numbers, line numbers and counts, and
returned 150 "bug numbers" including `000` and `199`. The file-touch list is the
honest signal.

## Which agents share an action (do this before changing one)

```sql
SELECT type || ' -> ' || k
  FROM agent_definitions, jsonb_each(default_config->'workflow'->'steps') AS s(k,v)
 WHERE v->>'action' = 'diagnose_persist_fix_plan'
   AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active;
-- council-gate -> persist_submission
-- feature-designer -> persist_plan
-- fix-proposer -> persist_plan
```

**Gotcha:** `default_config::text LIKE '%diagnose_persist_fix_plan%'` finds a fourth
agent (`council-gate`) *and* would find agents that merely name the action in prompt
text. It also cannot tell you the STEP NAME, which differs — `council-gate` calls it
`persist_submission`, so a change keyed on the step name `persist_plan` would miss
it. Join through `jsonb_each` and read `v->>'action'`.

## Check a template field exists before you reference it

A wrong `{{.x.y}}` path renders an **empty string** and reports nothing. Read the
paths a working step in the same agent already uses:

```sql
SELECT unnest(regexp_matches(
  default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template',
  '\{\{\.spec_row[^}]*\}\}', 'g'))
FROM agent_definitions WHERE type='feature-designer'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

This caught `{{.spec_row.body}}` in my own migration before it was applied. There is
no `body` field; it is `summary` + `spec_text`.

## Check a table's constraints before choosing a column value

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -c "\d diagnosis_artifacts"
```

**Gotcha:** `kind` is CHECK-constrained to five values. A new kind compiles and fails
at **runtime**, on the failing branch. `go build` and every mocked-DB test pass.

## Dry-run a migration without applying it

```bash
SP=<scratchpad>
sed 's/^COMMIT;$/ROLLBACK;/' docs/agent_docs/sql_for_agents/272_feature_designer_plan_repair_loop.sql > $SP/272_dryrun.sql
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < $SP/272_dryrun.sql
```

This executes the guards, every UPDATE and the verification block, then discards it —
so it proves the `replace()`/`jsonb_set` paths actually bit, which a syntax check
cannot. Expect `UPDATE 1` four times and the closing `NOTICE`, then `ROLLBACK`.

**Gotcha:** the snapshot `snapshot_agent()` takes is rolled back too, so a dry run
leaves no snapshot behind — that is correct, but it means a dry run is not a backup.

## Apply it for real, in the right order

Image first — DB config is live immediately, Go is inert until a roll:

```bash
# 1. after the chassis image carrying planValidationRefusal is rolled:
kubectl exec -n ai-persona-system <chassis pod> -- \
  sh -c 'grep -ac "plan_validation_refusal" /app/agent-chassis; grep -ac "staged plan failed validation" /app/agent-chassis'
#    first = the new code, second = a PRE-EXISTING positive control.
#    Both must be >0. A control that returns 0 means the grep itself is broken,
#    not that the change is missing (bugs_open/153: a tag bump does not imply a
#    rebuild, and verify-agent-images prints all-green on a stale image).

# 2. then:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
  < docs/agent_docs/sql_for_agents/272_feature_designer_plan_repair_loop.sql
```

## Confirm the wiring after applying

```sql
SELECT default_config->'workflow'->'steps'->'persist_plan'->>'next_step'          AS persist_next,
       default_config->'workflow'->'steps'->'persist_plan'->'config'->>'repair_step' AS repair_step,
       default_config->'workflow'->'steps'->'check_plan_valid'->'config'->>'condition' AS cond
  FROM agent_definitions WHERE type='feature-designer'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
-- check_plan_valid | repair_plan | plan_persisted.plan_valid == true
```

**Gotcha:** a whole-row `default_config::text ILIKE '%repair_plan%'` would pass just
as happily if the string had landed in a prompt, a comment or a step that never runs.
099's own file records this exact trap for migration `222`. Always path-qualify.

## Read the refusals the loop recorded

```sql
SELECT created_at, metadata->>'shape' AS shape,
       metadata->>'problem_count' AS n, metadata->'problems' AS problems
  FROM diagnosis_artifacts
 WHERE kind = 'iteration_note'
   AND metadata->>'note_kind' = 'plan_validation_refusal'
 ORDER BY created_at DESC LIMIT 20;
```

`body` is the rejected plan verbatim, so a design the loop gave up on is recoverable
by hand. **The `note_kind` filter is not optional** — see `LANDMINES.md`.

## Verify the fix on the failing branch (the only test that counts)

> **CORRECTION 2026-07-30 — 099's stated verification CANNOT prove candidate 2, and
> would return a false PASS.** The bug file says to re-fire work item
> `7b89fb35-f42c-45d1-b64d-214aff56d918` and require a `fix_plan` artifact plus no
> `complete_refused`. That procedure was written for **candidate 1**, and candidate 1
> **worked** — the file records the re-fired run clearing `persist_plan` and writing
> an artifact, and the rule is still live in the design step today (path-qualified
> check below returns `t`). So that run now takes the **success** path: it satisfies
> both pass conditions while the repair loop **never fires**. Green, and blind.
>
> ```sql
> SELECT (default_config->'workflow'->'steps'->'design'->'config'->>'prompt_template')
>        ILIKE '%ONE EDIT PER FILE PER STAGE%' AS rule_in_design_step
>   FROM agent_definitions WHERE type='feature-designer'
>    AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
> -- t  (checked 2026-07-30)
> ```
>
> Candidate 2 must be proved by **inducing** a refusal, and deliberately by a
> DIFFERENT rule from the one candidate 1 taught — being rule-agnostic is the whole
> claim.

### BEFORE ARMING: check nobody else is mid-run on this agent

Arming a cap mutates a **shared live agent**, and DB config is live immediately. Any
other lane's designer run that reaches `persist_plan` while your cap is armed will be
refused by *your* test, not by its own plan.

```sql
SELECT orchestration_id, current_step, status, created_at
  FROM orchestration_states
 WHERE workflow_plan::text LIKE '%check_spec_approved%'   -- feature-designer's signature step
   AND status NOT IN ('COMPLETED','FAILED')
 ORDER BY created_at DESC LIMIT 6;
```

**Gotcha:** `orchestration_states` has **no `agent_type` column** — the obvious query
errors out rather than returning nothing, which is at least honest. Match on a step
name unique to the agent's `workflow_plan` instead.

**I armed before running this check** on 2026-07-31. It happened to be safe (only my
own run was in flight), but the order was wrong: this is the same family as CLAUDE.md's
"checking the pod does not check the queue". Disarm as soon as the run terminates —
keep the armed window in minutes, not hours.

### Induce a repairable refusal (the discriminating test)

Set the per-stage edit cap to 1, so any natural multi-edit stage trips it. This is
repairable **without losing scope** — the designer moves edits into further stages,
which is exactly what the repair prompt tells it to do, and `max_stages` (6) leaves
room.

```sql
-- ARM (remember to disarm; take a snapshot first if you are nervous)
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,persist_plan,config,max_edits}', '1'),
       updated_at = now()
 WHERE type='feature-designer' AND deleted_at IS NULL
   AND COALESCE(is_snapshot,false)=false;
```

Fire the designer, then require **ALL FOUR** — the first two are 099's, the last two
are what make it discriminating:

3. a refusal note EXISTS, i.e. the mechanism actually **fired**:

```sql
SELECT created_at, metadata->>'problem_count', metadata->'problems'
  FROM diagnosis_artifacts
 WHERE kind='iteration_note' AND metadata->>'note_kind'='plan_validation_refusal'
   AND correlation_id='<FEATURE_CORR>';
-- expect >=1 row naming the per-stage cap
```

4. and the run reached `repair_plan` — the whole point:

```sql
SELECT jsonb_object_keys(collected_data) FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id'='<FEATURE_CORR>';
-- expect the step trail to include repair_plan
```

```sql
-- DISARM (back to the default)
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,persist_plan,config,max_edits}', '8'),
       updated_at = now()
 WHERE type='feature-designer' AND deleted_at IS NULL
   AND COALESCE(is_snapshot,false)=false;
```

**Gotcha:** `max_edits` is the PER-STAGE cap on the staged path and the whole-plan cap
on the legacy single-plan path (D3) — the same knob means two things. 8 is the
default in both the action and the input spec, so restore to 8, not to absent.

### The original criteria (necessary, not sufficient)

Per 099's own "how to verify", require **BOTH**:

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_TRIGGER_feature_designer_v1.sh \
  7b89fb35-f42c-45d1-b64d-214aff56d918
```

1. a `fix_plan` artifact EXISTS for the new correlation, and
2. the run does NOT end at `complete_refused`.

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<FEATURE_CORR>';
SELECT kind FROM diagnosis_artifacts WHERE correlation_id = '<FEATURE_CORR>';
```

**Gotcha:** "the run completed" is not evidence. 099's original failing run
COMPLETED — at `complete_refused`, with no artifact, `error` NULL and `final_result`
NULL. The reason lived only in `collected_data->>'__step_error'`. Read that field, not
`error`.

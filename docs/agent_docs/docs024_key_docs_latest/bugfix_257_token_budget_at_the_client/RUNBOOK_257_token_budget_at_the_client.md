# RUNBOOK — bugs_open/257 token budget at the client

Commands that were hard to get right, with the gotcha attached. Change them
HERE when they change, not in your scrollback.

## Is another session already on this bug?

`who-owns.py` reads COMMITS, so it cannot see a session mid-fix. You need both:

```bash
python3 scripts/who-owns.py 257          # lagging, but names the owning workstream
```

⚠ **Read the BODY, not the VERDICT line.** For 257 it printed
`VERDICT: OWNED or recently active` while its own body said
`likely OWNING workstream(s): (none identified)`. The verdict fires on any recent
commit touching the file — and the filing commit is a recent commit.

```bash
# The only instrument that sees UNCOMMITTED work in other sessions.
ls -t ~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl | head -20 | while read f; do
  age=$(( ($(date +%s) - $(stat -c %Y "$f")) / 60 ))
  [ $age -lt 240 ] || continue
  echo "=== $(basename $f) (${age}m) ==="
  tail -c 400000 "$f" | grep -oE 'bugs_open/[0-9]{3}' | sort | uniq -c | sort -rn | head -5
done
```

⚠ **A bug number appearing across MANY sessions is the opposite of ownership.**
`bugs_open/252` showed 6-10 mentions in eight different sessions: it is cited in a
shared cold-start doc everyone reads. Ownership looks like ONE session with a
dominant count, not a number that is popular.

⚠ `tail -c` on some of these files makes coreutils `tail` panic
(`InvalidInput`). Those entries just print a panic instead of counts — treat as
"unknown", not "clean".

## Re-run the direct-call-site census (bug §4)

```bash
grep -rn --include=*.go -E '\.(GenerateText|GenerateWithImages)\(' . \
  | grep -v '^./platform/aiservice/' | grep -v '_test.go'
```

⚠ **Do not quote the census in the bug file — it goes stale in days.** Filed
2026-08-12 with 9 sites; on 2026-08-16 there were 10 (`gripper.go:161` arrived in
between). Re-run it.

## Enumerate every CLIENT CONSTRUCTION site (what the census above misses)

The call-site census answers "who calls the model". For a client-side change you
need "who BUILDS a client, and what config did it hand over":

```bash
grep -rn --include=*.go -E 'aiservice\.(NewClient|NewAnthropicClient|NewGeminiClient|NewOllamaClient)\(' . \
  | grep -v '_test.go'
```

Then for each, read the config it passes. Three of them (`tools-api`) build an
**inline literal map**, so the config is at the call site, not in a yaml or the DB.

## Live config census

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT type, default_config->'ai_service'->>'max_tokens' AS ai_svc_mt,
       default_config->>'max_tokens' AS agent_mt
FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND (default_config->'ai_service' ? 'max_tokens' OR default_config ? 'max_tokens')
ORDER BY 1;"
```

⚠ **Two different keys, two different levels.** `default_config.max_tokens`
(agent level) BEATS `default_config.ai_service.max_tokens` in
`ai_actions.go:358-364`. Only the second is ever passed to a client constructor.
A census that reads one of them answers half the question.

⚠ `max_tokens` inside a step's `config` block (not `config.ai_service`) is
**INERT** — see `LANDMINES.md`, the 2026-08-15 entry. Migration 413 set it,
asserted it, printed OK, and the effective cap never moved.

## Is a service actually receiving traffic? (the [UNMEASURED] method)

`llm_call_log` **cannot answer this** for a direct-to-provider caller: the table
is written by `ExecuteAIStepAction`, the very function being bypassed, so it
returns 0 rows whether the service is idle or truncating every call.

```bash
kubectl -n ai-persona-system get pods -l app=<service>
kubectl -n ai-persona-system logs -l app=<service> --tail=200      # all startup == no traffic
grep -rn --include=*.go "<its request topic>" .                    # is there a PRODUCER at all?
```

⚠ For `reasoning-agent` the only non-consumer hit is a hardcoded "known topics"
list in an admin endpoint (`core-manager/admin/system_handlers.go:132`). **A topic
name appearing in the code is not a producer** — read what the line does.

## Mutation-testing harness (the thing that found the real defect)

```bash
run_mut () {   # name, file, from, to
  cp "$2" /tmp/mut.bak
  python3 - "$2" "$3" "$4" <<'PY'
import sys
p,f,t=sys.argv[1],sys.argv[2],sys.argv[3]
s=open(p).read()
if s.count(f)!=1: print("ANCHOR-MISS count=%d"%s.count(f)); sys.exit(9)
open(p,'w').write(s.replace(f,t))
PY
  [ $? -eq 9 ] && { echo "[$1] ANCHOR MISS"; cp /tmp/mut.bak "$2"; return; }
  out=$(go test ./platform/aiservice/ 2>&1)
  echo "$out" | grep -q "^FAIL\|--- FAIL" \
    && echo "[$1] KILLED by: $(echo "$out" | grep -oE '\-\-\- FAIL: [A-Za-z_/.]+' | sed 's/--- FAIL: //' | tr '\n' ' ')" \
    || echo "[$1] *** SURVIVED ***"
  cp /tmp/mut.bak "$2"
}
```

⚠ **The `count(f)!=1` assertion is load-bearing.** Without it a mutation whose
anchor does not match silently edits nothing, the tests pass, and you record
"KILLED"→ no, worse: you record "SURVIVED" for a mutation that was never applied.
An unapplied mutation and a surviving one produce identical output.

⚠ **A SURVIVING mutation is not automatically a test gap** — it may be an
*equivalent* mutant. M10 (deleting the nil-config branch) survived because Go
returns the zero value for a nil-map read, so the code is unchanged in effect.
Before adding a test, check whether the mutant can actually behave differently.

⚠ **A guard in SERIES will mask the guard you meant to test.** M9 survived because
a downstream floor rescued the value. If you want to claim a specific guard is
tested, the test must reach it with nothing downstream able to repair the answer —
which usually means asserting the function's return value directly.

## Council submission

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/bugfix_257_token_budget_at_the_client/COUNCIL_SUBMISSION_2026-08-16_257_budget_at_the_client.json
```

⚠ **A COMMENT-ONLY sketch is refused client-side**: *"a fix plan proposes changes,
not observations"*. An edit that IS a comment correction still needs its sketch
written as a diff (`-` old lines, `+` new lines), not as the prose of the new
comment. Cost me one round-trip.

Verdict (keyed on the submission correlation, never on a printed run id):

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '85971036-5b9a-42ca-aeac-d77a65397c86'
 ORDER BY created_at DESC LIMIT 1;      -- COMPLETED, not EXECUTING_STEP

SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='85971036-5b9a-42ca-aeac-d77a65397c86' AND kind='council_report'
 ORDER BY created_at;
```

## Post-roll verification at the pod

```bash
# Which deployments run the chassis image at the new tag? ONE image, MANY deployments.
kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.spec.containers[0].image}{"\n"}{end}' \
  | grep 'agent-chassis:v1.0.<NEW>' | awk '{print $1}' | sed 's/-[a-z0-9]*-[a-z0-9]*$//' | sort -u

# Then ask the binary what built it — do NOT grep for a marker, and do NOT use strings.
kubectl -n ai-persona-system logs -l app=<service> --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <this-commit> <the stamped sha> && echo SHIPPED
```

⚠ `reasoning-agent` is its OWN service and image, not the chassis. It needs its
own build/roll, and its `build provenance` line is the thing to read.

## ⚠ Is the council run alive, or was it REFUSED? (check this FIRST, before waiting)

A refused submission reads exactly like a queued one. `status` is `COMPLETED` either way.

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>'
 ORDER BY created_at DESC LIMIT 1;
```

- `complete_invalid` → **REFUSED at `persist_submission`. No review ran, no credits spent. Fix and
  resubmit now** — do not wait for a verdict that is never coming.
- `EXECUTING_STEP` / a review step → genuinely in flight; budget ~30 minutes.
- **no row at all** → latency (CLAUDE.md: almost always queue depth, do not retry on that evidence).

⚠ **A refusal records NOTHING you can read.** Measured 2026-08-16 on corr `85971036`:
`diagnosis_artifacts` → 0 rows of any kind, `collected_data` → only `input_data`, `__step_error` →
empty. The gate's `persist_submission` sets no `repair_step`, so it fails without writing a refusal
note. **Derive the reason from the validator's source instead** — `editProblems` / `validateFixPlan` /
`noOpEditReason` in `platform/orchestration/actions/diagnose_persist_fix_plan_action.go` — and prove it
by running that predicate over your own file:

```bash
python3 -c "
import json,sys
d=json.load(open(sys.argv[1]))
for i,e in enumerate(d['plan']['edits'],1):
    f=e['file']
    bad=('..' in f) or f.startswith('/') or any(c in f for c in ' \t\n')
    print(('FAIL' if bad else 'ok  '), i, repr(f))
    assert e['operation'] in ('modify','add','remove','config_change'), e['operation']
    for k in ('file','operation','rationale','sketch'):
        assert e.get(k,'').strip(), 'empty '+k
    L=[l for l in e['sketch'].split(chr(10)) if l.strip()]
    assert not all(l.lstrip().startswith(('--','#','//')) for l in L), 'comment-only sketch'
print('plan bytes:', len(json.dumps(d['plan'])), '(cap 65536)')
print('edits:', len(d['plan']['edits']), '(cap 8)')
" <submission.json>
```

**ONE EDIT = ONE FILE.** At the 8-edit cap, split the offender and MERGE two edits that touch the
SAME file rather than dropping content.

## ⚠ Poll the council on the INDEXED column — the obvious query times out

```sql
-- FAST: orchestration_id is indexed. Take RUN_ORCH_ID from the trigger's printout.
SELECT new_current_step, new_status, changed_at
FROM orchestration_state_audit
WHERE orchestration_id='<RUN_ORCH_ID>' ORDER BY changed_at;
```

⚠ **Do NOT poll `orchestration_states` on `collected_data->'input_data'->>'fix_correlation_id'`** — it
is a jsonb scan over a large table and **times out at 100s**. Measured 2026-08-16: a monitor built on
it ran for a full hour and emitted **zero events**, which reads exactly like "nothing has happened".
**An empty result from a query that timed out is indistinguishable from an empty result from a query
that ran** — so wrap any such poll in a `timeout` you can see, or use the indexed table.

Terminal steps and what they mean: `complete_approved` · `complete_revise` (objections came back —
read them, revise, resubmit with `RESUBMIT_CORR=<SUBMISSION_CORR>` so the trail accumulates) ·
`complete_rejected` (guardian veto) · `complete_invalid` (**refused before review — never reached a
seat**; see the section above).

Read the objections themselves from the report body:

```sql
SELECT metadata->>'decision', left(body, 6000) FROM diagnosis_artifacts
WHERE correlation_id='<SUBMISSION_CORR>' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;
```

---

## Round 2 (2026-09-03) — the commands this round needed, with their gotchas

### Census: which live steps run an action, and where their budget is declared

⚠ **`default_config->'workflow'->'steps'` walks TOP-LEVEL steps only** and misses anything inside a
loop's `sub_workflow` — which is exactly where `rewrite_negations` lives. Use `jsonb_path_query` with a
recursive path, or the census silently omits the case you are investigating.

```sql
SELECT a.type AS agent, s.key AS step_name, s.value->>'action' AS action,
       s.value->'config'->>'max_tokens'                  AS step_top_mt,   -- the spelling nothing reads
       s.value->'config'->'ai_service'->>'max_tokens'    AS ai_mt,         -- the one that works
       s.value->'config'->'ai_service'->>'budget_tokens' AS ai_bt,
       a.default_config->'ai_service'->>'max_tokens'     AS root_mt,
       a.default_config->>'max_tokens'                   AS agent_mt       -- outranks ai_service
FROM agent_definitions a,
     LATERAL jsonb_path_query(a.default_config, '$.**.steps') AS steps,
     LATERAL jsonb_each(steps) AS s
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND s.value->>'action' IN ('<action>', '<action>')
ORDER BY 1,2;
```

Four columns, not one, because the budget can be declared in four places and only two are read. Drop
the `action` filter and use `WHERE s.value->'config' ? 'max_tokens'` to find the dead spelling
fleet-wide — that is how the four `site-adoption-agent` declarations were found.

### Baseline: what a step actually SENT, pinned before a change

```sql
SELECT agent_type, step_name, max_tokens AS sent, count(*) AS calls,
       max(output_tokens) AS max_out,
       count(*) FILTER (WHERE output_tokens >= max_tokens) AS at_ceiling,
       min(created_at)::date AS first_seen, max(created_at)::date AS last_seen
FROM llm_call_log
WHERE step_name LIKE '%<action>%'
GROUP BY 1,2,3 ORDER BY 1,2,3;
```

⚠ Two traps in one query. **`step_name LIKE`, not `=`** — a loop substep is
`process_sections_loop_iter_3_rewrite_negations`, so an equality filter returns zero rows and reads as
"this never runs". And **`at_ceiling` here is NOT the truncation count** — a truncated call has
`output_tokens` NULL (see the fleet-wide landmine); count truncations from
`error_message ILIKE '%stop_reason=max_tokens%'`. Grouping by `sent` is the point of the query: it
shows a configuration change as a change of group, which is how migration 569's 2000→16000 is visible.

### Verifying a change when the working tree does not compile

The shared tree carries every session's WIP, so `go test ./...` can fail on a file you have never
opened (on 2026-09-03, another lane's untracked `recommended_type_reconciliation_test.go`). Do not
touch their file. Overlay yours onto committed HEAD instead:

```bash
scripts/verify-head-builds.sh --test --with <file> [--with <file> …] ./platform/orchestration/actions/
scripts/verify-head-builds.sh ./...            # after committing: does HEAD still build?
```

⚠ **A `FAILED` from `--test` is not necessarily yours.** Read which tests failed, then run the control:
`git grep -l '<symbol the failure names>' HEAD` — if the symbol lives in files you did not touch, it is
another lane's. On 2026-09-03 two failures (`TestFindingCodeScanEveryWriteIsRegistered`,
`TestTemplateExecutorsAreDeclared`) came from committed HEAD and none of the seven files in flight
mentioned either symbol.

### Proving a new guard can actually fail

`KEEP_TREE=1` leaves the checkout, and the mutations go **there**, never in the shared tree:

```bash
KEEP_TREE=1 scripts/verify-head-builds.sh --test --with <files…> ./platform/orchestration/actions/
T=$(ls -dt /home/ant/.claude-scratch/head-verify/*/tree | head -1)
# edit "$T/<file>" to reinstate the defect, run the one test, restore, repeat
cd "$T" && go test ./platform/orchestration/actions/ -count=1 -run '<TheGuard>'
rm -rf "$(dirname "$T")"      # ~450MB per checkout — reap it
```

One mutation per claim the guard makes. Four claims here → four mutations, each with the expected
message read, not just a non-zero exit.

### Polling for the council verdict — the WARNING FURTHER UP THIS FILE IS REAL

The `orchestration_states` jsonb-scan poll **timed out twice on 2026-09-03**, despite being documented
above as timing out at 100s. It presents as a `kubectl` hang and then exit 143, which reads like a
cluster problem. Use the indexed artifact table from the start:

```sql
SELECT created_at, kind, metadata->>'decision' AS decision
FROM diagnosis_artifacts WHERE correlation_id='<SUBMISSION_CORR>' ORDER BY created_at;
```

A `fix_plan` row appears within a minute of dispatch (the submission landed); the `council_report` row
carries the decision. Round 2 of this lane: dispatched 16:21, `approved` at 16:37 — **16 minutes**, not
the ~30 CLAUDE.md budgets for.

### `097` admission: a comment-only sketch is refused client-side

`DRY_RUN=1 097_TRIGGER_council_review_v1.sh <submission.json>` is free and catches it. **Every non-blank
line of a `sketch` starting `//`, `--` or `#` is refused** — *"a fix plan proposes changes, not
observations"*. That bites a documentation-only edit, whose sketch is genuinely all comment: write it as
a **diff**, with `+` prefixes, which is both honest and admissible.

---

## Round 3 (2026-09-04) — commands that were hard to get right

### Census the CONCEPT: walk the whole document, never a fixed path

Four censuses of this bug asked a fixed-path question and each could only find declarations where it
already believed they lived — which is the assumption under test. This is the one that works:

```sql
WITH RECURSIVE a AS (
  SELECT type, default_config AS j FROM agent_definitions
  WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active
), walk(type, path, j) AS (
  SELECT type, ''::text, j FROM a
  UNION ALL
  SELECT w.type, w.path||'.'||e.key, e.value FROM walk w, LATERAL jsonb_each(w.j) e
   WHERE jsonb_typeof(w.j)='object'
)
SELECT regexp_replace(regexp_replace(path,'^\.workflow\.steps\.[^.]+','.workflow.steps.<step>'),
         '\.(substeps|sub_workflow)\.steps\.[^.]+','.<nested>','g') AS shape, count(*) AS n
FROM walk WHERE path LIKE '%.max_tokens' GROUP BY 1 ORDER BY 2 DESC;
```

⚠ A fixed-path version cannot see a NESTED loop step either — those live under
`...config.sub_workflow.steps.<name>.config.ai_service`, not under `workflow.steps.<name>`.

### Where is each step's budget actually coming from?

```bash
scripts/audit-budget-placement.sh            # human-readable, grouped by kind
scripts/audit-budget-placement.sh --json     # the findings JSON
```
Calls `actions.ResolveStepBudget` — production's own ladder — so it cannot answer a different question
from the pods. Exit 0 clean / 1 findings (`ambiguous`/`unconfigured`) / 2 could-not-determine.
`non_canonical` is advisory and never changes the exit code.

### Pre-flight a migration: substitute ROLLBACK for COMMIT and RUN it

**This is the check that works, and the runner's own probe is not a substitute.**

```bash
sed 's/^COMMIT;$/ROLLBACK;/' docs/agent_docs/sql_for_agents/NNN_x.sql > /tmp/NNN_dryrun.sql
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < /tmp/NNN_dryrun.sql
```
Every guard, every UPDATE and the verify block execute for real; nothing is kept. On 2026-09-04 this
caught `operator is not unique: unknown - unknown` — **PostgreSQL binds subtraction TIGHTER than `->`**,
so `s.value->'config' - 'max_tokens'` parses as `s.value -> ('config' - 'max_tokens')`. Write
`((s.value->'config') - 'max_tokens')`.
⚠ The runner's built-in probe reported the file *"not probed: contains its own ROLLBACK/ABORT"* — a
false positive on the word `Rollback:` in a header comment — so it would have caught nothing.

### Apply one migration without applying ninety others

```bash
SCOPED=<scratch>/migNNN; mkdir -p "$SCOPED"; cp docs/agent_docs/sql_for_agents/NNN_x.sql "$SCOPED"/
MIGRATIONS_DIR="$SCOPED" ./scripts/migration/run-migrations.sh          # dry run
MIGRATIONS_DIR="$SCOPED" ./scripts/migration/run-migrations.sh --apply
```
⚠ The assignment MUST be on the same line as the command. On its own line it scopes nothing and the
unscoped run applies every other thread's pending file.

### The induced "is it live" check — arm, read, REVERT

Puts a number where ONLY round-2 code looks (`llmOptionsFromConfig` reads the step's BARE key before the
merged `ai_service` block). Two probes, because both callers are bursty.

```sql
-- ARM 1: page-content-writer.rewrite_negations   (configured 16000 -> probe 15999)
UPDATE agent_definitions SET default_config = jsonb_set(default_config,
  '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,max_tokens}',
  '15999'::jsonb, true)
WHERE id = '5946a27b-38ab-41e8-8b49-7bc1a4b626b8';

-- ARM 2: offer-analyser.repair_ordering_register  (configured 2000 -> probe 1999)
UPDATE agent_definitions SET default_config = jsonb_set(default_config,
  '{workflow,steps,repair_ordering_register,config,max_tokens}', '1999'::jsonb, true)
WHERE id = '4ad588bc-c491-4400-bb2f-f0f7a1cac0cd';

-- READ (a call must land AFTER the arm time)
SELECT created_at, agent_type, step_name, max_tokens FROM llm_call_log
WHERE step_name LIKE '%rewrite_negations%' OR step_name = 'repair_ordering_register'
ORDER BY created_at DESC LIMIT 5;
--   15999 / 1999 -> round-2 code IS live      16000 / 2000 -> it is NOT

-- REVERT, both, as soon as read
UPDATE agent_definitions SET default_config = default_config
  #- '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,max_tokens}'
WHERE id = '5946a27b-38ab-41e8-8b49-7bc1a4b626b8';
UPDATE agent_definitions SET default_config = default_config
  #- '{workflow,steps,repair_ordering_register,config,max_tokens}'
WHERE id = '4ad588bc-c491-4400-bb2f-f0f7a1cac0cd';
```
⚠ `offer-analyser`'s probe is the one that matters: 2000 was BOTH its configured value and the old Go
literal, so no query over `llm_call_log` could ever separate a honoured config from a dropped one there.
⚠ Both callers are bursty. `rewrite_negations` ran 4–38 times an hour overnight and then nothing for
over an hour. Arm and come back; do not conclude anything from silence.
⚠ This probe depends on the direct-caller ladder reading the bare key FIRST. Round 3 deliberately left
that ordering alone (owner decision 2). If a later round unifies the two ladders, the probe stops
discriminating and needs a new shape.

### Read a council verdict — `decided_by` BEFORE `decision`

```sql
SELECT created_at, kind, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='<SUBMISSION_CORR>' ORDER BY created_at;

SELECT body::jsonb->>'decided_by', body::jsonb->>'unreadable'
FROM diagnosis_artifacts WHERE correlation_id='<SUBMISSION_CORR>' AND kind='council_report';
```
⚠ A round can be carried by `"unreadable reviewer(s): ..."` — seats whose output could not be parsed —
with **no seat having objected**. Round 1 of this change was REVISE with 8 of 12 seats unreadable and
2 of the 4 readable ones approving. [MEASURED 2026-09-04] 3 of 225 councils over 7 days. Reading the
verdict alone would be reading a rejection nobody made.
⚠ Do **not** poll `orchestration_states` — that jsonb scan times out at 100s and presents as a `kubectl`
hang.

### Are the two token-pressure checks live, and what do they actually read?

```sql
SELECT name, enabled, interval_seconds, last_completed_at, left(pre_query, 900)
FROM scheduled_tasks WHERE name LIKE '%token-pressure%';
```
Read `pre_query`, not the wrapper script: `fleet-step-token-pressure` joins **nothing but
`llm_call_log`**, and `council-seat-token-pressure` reads `agent_definitions` on **one** fixed path and
only for `review_%` seats. Neither can see where a budget came from, and the fleet one cannot see a step
that has never run.

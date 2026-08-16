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

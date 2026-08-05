# HANDOFF — bugfix 196, continue here (written 2026-08-05 ~14:00 BST)

Bug: `bugs_open/196_HANDOFF_2026-08-04_a_failure_response_reaches_the_parent_stamped_complete_and_the_parent_believes_it.md`
Lane docs: this directory (PLAN, NOTES, RUNBOOK, SEED). Read PLAN first, NOTES §missteps second.

## State at handoff

- **Fix COMMITTED at `d16e6d23c`, council APPROVED round 1**
  (`Council-Reviewed: d1a63089-af5b-41a2-bea1-62259aa5db52`), mutation-verified
  (see NOTES). Register entry CTS-058 shipped in the same commit.
- **NOT LIVE.** The chassis build rolled on 2026-08-05 early afternoon predates
  `d16e6d23c`. Go changes are inert until an image is rebuilt and rolled.
- Bug file updated + claimed (`baaafacfb`). Probe misstep in `WRONG_CALLS.md`.
- Implementation was Opus-delegated; the agent died on the account session limit
  (resets 17:50 London) after finishing the code+tests; the parent session
  verified everything and ran the mutation check itself. Nothing is owed to or
  from that subagent.

## What remains (in order)

### 1. Build + deploy the chassis carrying `d16e6d23c`

Confirm HEAD contains the fix before building (another session may have moved it;
that is fine — build-from-HEAD ships every commit): `git log --oneline -5 -- platform/messaging/`.
Bump `IMAGE_TAG` (makefile ~line 16 — a same-tag rebuild ships the node's stale
cached binary), then `make build-agent-chassis`, `make push-agent-chassis`,
deploy per the makefile's deploy target. ⚠ Check with the owner/fleet state
before a solo roll: memory says releases are whole-fleet and an imperative
one-service apply at its own tag gets undone by the next `apply -k`; and a roll
KILLS in-flight council runs.

### 2. Pod-verify (objection 3's explicit protocol)

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | while read p; do
  echo "$p: $(kubectl -n ai-persona-system exec ${p#pod/} -- sh -c \
    'strings /app/agent-chassis | grep -c sendWorkflowResponseWithStatus')"
done
```
Expect ≥1 on EVERY replica (Go keeps function names in pclntab). ⚠ 195's lesson:
short string literals inline to immediates and grep 0 on a binary that carries
the change — the function NAME is the reliable positive control here. There is no
greppable string the change REMOVES, so the negative control is behavioural: the
induction below. Do not skip it; a roll is not evidence (bugs_open/153).

### 3. Induction — baseline first if the old binary is still up, else post-fix only

Seeds: `SEED_test_196_probes.sql` in this directory (design rationale in its
header; the parent fabricates the spawn blob via query_database so the flaky
spawn handshake is not a dependency).

Dispatch (the 195 lane's proven recipe, generic topic; note the kcat landmine —
payload in the container COMMAND with a PUBLISH_OK marker, never bare stdin):

```bash
CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid); NAME="ind196-$(date +%H%M%S)"
echo "CORR=$CORR ORCH=$ORCH NAME=$NAME"
PAYLOAD=$(jq -nc '{action:"orchestrate",config:{agent_type:"test-196-parent"},input_data:{child_agent_type:"test-196-invalid-child",note:"bugs_open/196 induction"}}')
kubectl -n kafka run "kcat196-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "printf '%s' '$PAYLOAD' | kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR -H request_id=$(cat /proc/sys/kernel/random/uuid) \
  -H message_id=$(cat /proc/sys/kernel/random/uuid) \
  -H orchestration_id=$ORCH -H orchestration_name=$NAME \
  -H step_name=start -H client_id=demo_client -H message_type=request \
  -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
  -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"
```
No `PUBLISH_OK` → nothing was published; re-fire immediately. No dispatch within
~300s of a chassis pod (re)start — silently dropped. Save CORR/ORCH.

Read the PARENT (same day — terminal rows reap ~24h):
```sql
SELECT status, current_step, error, jsonb_pretty(collected_data->'call_child')
FROM orchestration_states WHERE orchestration_id='<ORCH>';
```

- **Pre-fix binary (baseline):** parent COMPLETED/advanced; `call_child.response`
  = `{"error": "...WORKFLOW_INVALID...", "status": "failed"}`; `error` NULL.
- **Post-fix binary (the acceptance):** parent routes the failure — expected
  `FAILED` with `error` carrying the child's WORKFLOW_INVALID message via
  failWorkflow (call_child has no error_step; not a continue_on_error loop).
- **FALSIFIERS, named before running:** post-fix parent reaches `finish`/COMPLETED
  with the blob as step data → the fix is wrong, correct the plan in place. Parent
  stuck in AWAITING_RESPONSES → call_agent's await never fired or the child
  response was lost — check `awaited_requests` and the chassis log for CORR
  before touching anything (TRAP 2: grep the log first, latency is not loss).
- Cleanup either way: `DELETE FROM agent_definitions WHERE type IN
  ('test-196-invalid-child','test-196-parent');`

### 4. Close

- Update the bug file: CLOSED header quoting the induction table (195's close is
  the format model), pod-verified tag, the correlation ids.
- `git mv` to `bugs_closed/196_HANDOFF_...same-slug....md` — ⚠ name BOTH paths on
  the commit (`git commit bugs_open/OLD.md bugs_closed/NEW.md docs/... -m ...`)
  and verify at HEAD: `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 196`
  → exactly one line.
- 016b §9: add the transferable pattern (suggested shape: "a success-shaped
  envelope from a shared sender outranks the correct error response because the
  first response CLAIMS the awaited request — when a reply looks wrong, ask what
  else answered the same request_id first"). Cross-reference 195's entry.
- Tell the consumers (07-29 ruling #3): append a dated line to `bugs_open/029`
  (its file already carries 196's settlement statement — confirm it still holds:
  the parent is answered promptly, now with the truth), and a line to
  `bugs_open/149`'s lane noting handler sagas that fail now park items the same
  way adapter failures always did. Neither is a request for work, just the
  guarantee change, named.
- CTS-058 register entry: flip status wording from "built …pending roll" to
  live+verified with the tag, per the stale-status landmine.
- Update `MEMORY.md`/topic file if a durable practice emerged (the duplicate-race
  mechanism — first responder claims the awaited request — is a good memory
  candidate).

## Gotchas already hit (do not re-derive)

- jsonb text-LIKE probe counts documents, not fields (WRONG_CALLS 2026-08-05).
- The workflow first-step key is `start_step`, NOT `initial_step` (processor.go:417).
- `agent_definitions` conditions census: use `jsonb_path_query`, not text LIKE;
  count() over a set-returning function needs a subquery.
- kcat stdin publish silently drops (~4/5) — container COMMAND + PUBLISH_OK.
- `diagnosis_artifacts` content column is `body`; verdict is `metadata->>'decision'`.
- The 198 lane swept my index row into `c48c773c1` — harmless, already noted.

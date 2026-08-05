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

### 3. Induction — TWO-DISPATCH RECIPE (v2; v1 was refuted 2026-08-05 evening, see NOTES)

⚠ **Do NOT use the single-dispatch design from this file's first version.** It
was run (corr `769f316f`) and REFUTED: a call_agent child travels in a nested
RequestMessage envelope, and `extractGroupInfo` reads only the msgBody top
level, so the child cannot select the invalid workflow — it runs generic's
no-op and completes LEGITIMATELY. LANDMINE filed. The v2 recipe splits the
roles: the parent parks awaiting a request nothing real can answer, and a
separately CLI-published FLAT message (which DOES resolve, the proven 195 path)
fails and answers the parent's awaited request with the error envelope.

Both probe rows are ALREADY SEEDED. First re-point the parent's fabricated
spawn blob at a void topic and raise the timeout:

```sql
UPDATE agent_definitions SET default_config = jsonb_set(jsonb_set(default_config,
  '{workflow,steps,prepare_child,config,query}',
  to_jsonb('SELECT ''probe-fake-agent'' AS agent_id, ''generic'' AS agent_type, ''child'' AS role, ''system.agent.test-196-void.requests'' AS requests_topic, ''system.agent.generic.responses'' AS responses_topic'::text)),
  '{workflow,steps,call_child,config,timeout_seconds}', '600')
WHERE type='test-196-parent';
```

**Dispatch 1 — park the parent** (kcat pattern with PUBLISH_OK; no dispatch
within ~300s of a pod restart):

```bash
CORR=$(cat /proc/sys/kernel/random/uuid); ORCH=$(cat /proc/sys/kernel/random/uuid); NAME="ind196-$(date +%H%M%S)"
echo "CORR=$CORR ORCH=$ORCH NAME=$NAME"
PAYLOAD=$(jq -nc '{action:"orchestrate",config:{agent_type:"test-196-parent"},input_data:{note:"bugs_open/196 induction v2 - parent parks"}}')
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

Wait for the parent to park, then read the awaited request id R and the child
orch id (both from the parent row):

```sql
SELECT status, jsonb_pretty(awaited_requests) FROM orchestration_states
WHERE orchestration_id='<ORCH>';
-- expect status AWAITING_RESPONSES / awaiting; R = the awaited_requests key.
```

**Dispatch 2 — the failing child, FLAT body, carrying the parent's reply
headers** (this is what makes handleError's sendErrorResponse answer R):

```bash
R=<awaited request id>; CORR=<same CORR as dispatch 1>
CHILD_ORCH=$(cat /proc/sys/kernel/random/uuid)
PAYLOAD=$(jq -nc '{action:"orchestrate",config:{agent_type:"test-196-invalid-child"},input_data:{note:"answers the parked parent with a validation failure"}}')
kubectl -n kafka run "kcat196b-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet \
  --command -- sh -c "printf '%s' '$PAYLOAD' | kcat -P \
  -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR -H request_id=$(cat /proc/sys/kernel/random/uuid) \
  -H message_id=$(cat /proc/sys/kernel/random/uuid) \
  -H orchestration_id=$CHILD_ORCH -H orchestration_name=ind196-child \
  -H step_name=start -H client_id=demo_client -H message_type=request \
  -H action=orchestrate -H from_agent_type=generic -H from_agent_id=cli \
  -H reply_to_request_id=$R -H reply_to_topic=system.agent.generic.responses \
  -H responses_topic=system.agent.generic.responses && echo PUBLISH_OK"
```

If the error response fails to route, the header→ExecutionContext mapping is the
suspect: decode one real call_agent child envelope from the chassis log (this
session's trace, NOTES evening entry, shows the exact field names:
`reply_to_request_id`, `reply_to_topic`) and check the chassis log for
sendErrorResponse's "no responses topic". TRAP 2 holds: grep the log for CORR
before re-firing anything.

Read the PARENT (same day — terminal rows reap ~24h):
```sql
SELECT status, current_step, error, jsonb_pretty(collected_data->'call_child')
FROM orchestration_states WHERE orchestration_id='<ORCH>';
```

- **Pre-fix binary (baseline):** parent COMPLETED/advances; `call_child` data =
  the blob `{"error": "...WORKFLOW_INVALID...", "status": "failed"}` under
  `.response`; `error` NULL. (The bug, live.)
- **Post-fix binary (the acceptance):** parent FAILED via
  handleUnrecoverableError → failWorkflow, `error` carrying the child's
  WORKFLOW_INVALID message; no complete-stamped blob in step data.
- **FALSIFIERS, named before running:** post-fix parent completes with the blob →
  fix wrong, correct the plan in place. Parent still AWAITING after dispatch 2 →
  the crafted headers did not route (see above) — a probe fault, not evidence
  about the fix.
- Also run 090/state cleanup either way:
  `DELETE FROM agent_definitions WHERE type IN ('test-196-invalid-child','test-196-parent');`
  and cancel the parked parent if it is still awaiting (it times out at 600s on
  its own).

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

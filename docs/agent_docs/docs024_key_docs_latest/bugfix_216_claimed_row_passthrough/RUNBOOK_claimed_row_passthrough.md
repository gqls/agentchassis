# RUNBOOK — bugfix 216 claimed-row pass-through

Commands that were hard to get right, with their gotchas. The induction commands live
in the 207 lane's recipe (`../bugfix_207_sender_convergence/`, SEED_test_207_probe.sql
+ NOTES 2026-08-07) — not duplicated here; gotchas 3 of the 207 handoff lists the four
traps (void topic auto-create off, `child_agent_type` in the dispatch payload, forged
child reply headers dropped at intake → capture on `system.generic.responses` and
re-publish to `system.agent.generic.responses`, transient child = inline workflow with
`local_action_timeout_seconds: 0.001`).

## Run the regression tests

```bash
go test ./platform/orchestration/ -run 'TestRecoverableRetry' -v
```

Gotcha: the tests dial a deliberately-unreachable DB (`127.0.0.1:1`) — they need no
cluster and no local postgres, and a failure mentioning `connection refused` in a LOG
line is expected noise, not the failure. The pgx driver is registered by the test
file's own blank import.

## Ownership / competing-session check (who-owns is blind to uncommitted work)

```bash
scripts/who-owns.py 216
grep -l "bugs_open/216\|RETRY_PAYLOAD_UNAVAILABLE\|ClaimAwaitedRequest" \
  ~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl
```

Gotcha: a transcript matching ONCE is a session that merely read a file citing the
bug; grep `-c` per file and read the last matching line before concluding anyone is
working it.

## Deploy verification markers (pod-grep, every replica, same exec)

```bash
# positive — added by this fix (expect >=1):
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "Retry decided on the claimed awaited row passed through from the claim"'
# negative — removed by this fix (expect 0):
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "Using in-memory awaited request"'
```

Gotcha (bugs_open/153): a roll is not evidence the fix shipped — the image may predate
the commit. The negative control is what proves the binary is post-fix rather than
merely containing a similar string.

## The induction (216's own settlement measurement — run POST-ROLL, same day)

Adapted from 207's run (SEED + commands: `../bugfix_207_sender_convergence/` +
196's HANDOFF_2026-08-05 dispatch blocks). Acceptance DIFFERS from 207's: the proof is
the replayed request CONSUMED from the void topic — a `retry_version` bump is what the
BROKEN version produced too (WRONG_CALLS 2026-08-07).

1. Seed the parent: apply `../bugfix_207_sender_convergence/SEED_test_207_probe.sql`.
   Create the void topic first (auto-create is OFF; without it the call step fails
   synchronously instead of parking): `kafka-topics.sh --create --topic
   system.agent.test-207-void.requests --partitions 1` on
   `personae-kafka-cluster-combined-pool-prod-0` (kafka ns).
2. Dispatch 1 (park the parent): 196's kcat container pattern (stdin publish drops
   ~4/5 — container COMMAND + `&& echo PUBLISH_OK`, never bare stdin), payload
   `{action:"orchestrate",config:{agent_type:"test-207-parent"},input_data:{child_agent_type:"generic"}}`
   — input_data MUST carry `child_agent_type` (the parent's input_mapping requires it;
   its omission cost 196 a dispatch). No dispatch within ~300s of a chassis restart.
3. Read R + confirm parked: `SELECT status, jsonb_pretty(awaited_requests) FROM
   orchestration_states WHERE orchestration_id='<ORCH>';` and confirm the payload was
   recorded: `SELECT retry_version, status, request_payload IS NOT NULL, requests_topic
   FROM awaited_requests WHERE request_id='<R>';`
4. Dispatch 2 (transient child): same kcat pattern, payload carries an INLINE workflow
   override (processor.go:974 reads `config.workflow`):
   `{action:"orchestrate",config:{agent_type:"generic","workflow":{"start_step":"probe_deadline","steps":{"probe_deadline":{"action":"query_database","config":{"query":"SELECT pg_sleep(5)","local_action_timeout_seconds":0.001},"next_step":"finish"},"finish":{"action":"complete_workflow","config":{}}}}},input_data:{}}`
   with headers `reply_to_request_id=$R`, `reply_to_topic=system.agent.generic.responses`,
   same `correlation_id` as dispatch 1. Deadline-exceeded → 207's converged sender emits
   `error_recoverable`.
5. Capture + re-drive (196's header-drop wrinkle: the child's reply headers do not
   survive this intake path, so the envelope lands on LEGACY `system.generic.responses`,
   unconsumed): kcat -C the tail of `system.generic.responses`, take the envelope with
   our correlation, re-publish it BYTE-IDENTICAL (headers included) to
   `system.agent.generic.responses`.
6. **ACCEPTANCE (all four, named before running):**
   - `awaited_requests` R: `retry_version=1` AND `status='waiting'` (re-armed AND
     released — `UpdateAwaitedRequestRetry` sets both).
   - Parent orchestration: still `AWAITING_RESPONSES`, `error` NULL — NOT failed.
   - **THE PROOF: kcat -C `system.agent.test-207-void.requests` shows a SECOND message
     (offset 1; offset 0 is the original send) decoding as the original RequestMessage
     with `headers.retry_version=1` and the CHILD's orchestration_id** — the replay
     reached the wire.
   - Chassis log for the corr: the new marker line with `payload_present=true`, then
     `Replaying original request to target agent requests topic`; NO
     `RETRY_PAYLOAD_UNAVAILABLE`.
   **FALSIFIERS:** parent FAILED ms after the re-arm + `RETRY_PAYLOAD_UNAVAILABLE` in
   the log = the fix did not ship or did not work — stop, pod-grep again, correct the
   plan in place. R stuck at `status='processing'` = a different defect, file it.
   Nothing new on the void topic but retry_version=1 = EXACTLY the pre-fix shape.
7. After the acceptance: R will time out again on the void topic (nothing answers);
   the timeout path (healthy, 'retrying') drives retries to exhaustion at 3 and the
   parent fails on budget — that is EXPECTED, not a falsifier; it proves the cap
   (RSH-006) still binds. Cleanup: `DELETE FROM agent_definitions WHERE
   type='test-207-parent';` and delete the void topic. Orchestration rows reap ~24h —
   quote them same-day into NOTES.

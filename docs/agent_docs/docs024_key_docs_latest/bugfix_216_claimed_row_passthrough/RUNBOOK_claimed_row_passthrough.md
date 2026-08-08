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

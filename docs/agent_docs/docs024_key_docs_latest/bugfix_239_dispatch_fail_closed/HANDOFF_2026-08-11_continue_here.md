# bugfix 239 — dispatch fail-closed: HANDOFF, 2026-08-11

**Read this first, then `bugs_open/239` (the shared account) and
`architecture_review/RFC_023_*` (the open owner decision).** This lane opened
2026-08-10, fixed the bug, shipped it, and proved it live on v1.0.1284. It is nearly done.

## Where we are in one paragraph

A hand-dispatched `action=orchestrate` message could silently run the `generic` agent's
no-op instead of the agent it named, and report `COMPLETED`. Root cause: `kcat -P` sends
**one message per line of stdin**, so multi-line envelopes arrived as invalid-JSON
fragments — and the chassis converted any unresolvable dispatch into a successful-looking
no-op. Fixed: such a dispatch now **fails closed**. **Proven live on v1.0.1284.** One
promised trace (the FAILED orchestration row) turned out to be a no-op in production; it is
fixed in the tree and rides the next roll. That is the ONLY thing keeping 239 open.

## THE ONE OUTSTANDING TASK — 5 minutes, after the next chassis roll

> ✅ **DONE 2026-08-11 ~12:50 UTC — the roll (v1.0.1286) arrived the same day, and the row
> appeared.** Corr `cc7bd91a` → `no-such-agent-239 | FAILED | …agent_type_unresolved…`;
> baseline was 0 such rows over all prior history. Full evidence in `bugs_open/239`'s
> CLOSING VERIFICATION section; register SYS-090 updated. **Nothing in this lane is
> outstanding.** RFC_023 stays open with the owner. Two corrections to the recipe below,
> found on execution: step 1's `strings` probe is the RETIRED recipe (CLAUDE.md 08-11) —
> use the provenance stamp + `git merge-base --is-ancestor` instead (the stamp is another
> commit's sha, so grep `/proc/1/exe` for a KNOWN candidate with an absent-control); and
> the chassis's `build provenance` log line is already unreachable (>5000 lines) within
> the hour, while a naive `grep 'build provenance'` over its logs can false-match a
> council-gate payload that merely QUOTES the phrase.

`recordDispatchFailureState` guarded on `p.sqlDB`, which is populated only when
`DATABASE_URL` is set, and **it is not set on the chassis pods**. So no FAILED
`orchestration_states` row has ever been written. Fixed at commit `209917d15`
(`db := p.db; if db == nil { db = p.sqlDB }`), regression test mutation-verified.

**After the next roll**, confirm the image carries it, then fire ONE dispatch and check:

```bash
# 1. is it in the binary?
kubectl -n ai-persona-system exec <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c DISPATCH_FAIL_CLOSED'      # >0 (8 on 1284)

# 2. one safe dispatch — NON-EXISTENT agent type, so nothing can run
C=$(uuidgen); R=$(uuidgen); M=$(uuidgen); O=$(uuidgen); echo $C
kubectl -n kafka run -i --rm kcat239-$(date +%s) --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests \
  -H correlation_id=$C -H request_id=$R -H message_id=$M -H orchestration_id=$O \
  -H orchestration_name=verify239-row -H step_name=start -H client_id=demo_client \
  -H message_type=request -H action=orchestrate -H from_agent_type=user -H from_agent_id=cli \
  -H responses_topic=system.generic.responses \
  <<<'{"action":"orchestrate","config":{"agent_type":"no-such-agent-239"},"input_data":{}}'
```
```sql
-- 3. THE assertion. Pre-fix this returns 0 rows; post-fix exactly one.
SELECT owner_agent_type, status, left(error,80) FROM orchestration_states
WHERE correlation_id::text = '<C>';
-- want: no-such-agent-239 | FAILED | ...agent_type_unresolved...
```
**If that row appears, `bugs_open/239` closes.** Add the result to the bug file's
POST-ROLL VERIFICATION section and to concept register SYS-090 (whose status line currently
carries the caveat). Nothing else is outstanding.

**⚠ NOTE THE SINGLE QUOTES AND `<<<`.** Never a multi-line `<<JSON` heredoc — that is the
bug. See `LANDMINES.md`, entry "kcat -P publishes ONE MESSAGE PER LINE".

## The owner decision that is open (RFC_023) — nothing blocked by it

The council **REJECTED** the fix (`fca1071b-80ac-40cd-8c6d-d30a735de89b`) on a **hard
guardian veto about SCOPE** — four packages, exported symbols, one commit — while 7 of 11
seats approved, `editquality` said "not a veto: the core fix is on-target", and
`reuse_agent` (the architectural-fit seat) found "no architecture-level reuse concern". The
seats contradicting each other is CLAUDE.md's stated condition for a human to break it, and
a scope veto is explicitly NOT answered by resubmitting. **So: recorded, not resubmitted.**
RFC_023 asks the owner one question — *is the RFC trigger a property of the DIFF (packages,
symbols) or of the BEHAVIOUR (does a consumer's success path change)?* Every fix of the
class "X silently succeeded when it should have failed" fails the first test and passes the
second. **Do not resubmit this round. Do not revert.**

## What is live, and what to trust

- **LIVE and proven on v1.0.1284** (both replicas, both directions): the refusal itself,
  its determinism (5/5 identical dispatches), the terminal-vs-transient disposition, and
  the intake row marked `failed` with the reason.
- **NOT live:** the FAILED orchestration row (above).
- **Therefore the queryable trace today is the intake table, not orchestration_states:**
  ```sql
  SELECT left(correlation_id::text,8), status, left(last_error,90)
  FROM chassis_intake_events
  WHERE kind='request' AND last_error LIKE 'DISPATCH_FAIL_CLOSED%'
  ORDER BY received_at DESC LIMIT 20;
  ```

## Two traps this lane hit, so you don't

1. **"No `generic` orchestrations since the roll" is NOT a regression.** Those scheduled
   ticks fire about **once an hour**, not once per 180s tick. Establish the baseline over
   the preceding hours before calling a post-change number anomalous.
2. **A fix promising N traces needs N asserted.** Mine promised three; two worked, and the
   third had never run in production. Two-of-three reads exactly like three-of-three. Also:
   `p.db` and `p.sqlDB` are NOT interchangeable in `platform/messaging` — `sqlDB` is nil on
   the chassis, and building a test fixture with it set tests a shape production lacks.

## Related work filed by this lane, unowned, NOT started

- `bugs_open/246` — `NewMessageProcessor` re-shrinks the SHARED `*sql.DB` to 4 connections,
  silently undoing `CHASSIS_DB_MAX_OPEN_CONNS=12`. Now instrumented for the first time:
  a transient lookup fault logs `DISPATCH_LOOKUP_RETRYABLE`, so measure with that BEFORE
  changing the pool.
- `bugs_open/247` — three dead symbols that read as the live dispatch path
  (`AgentConfigLoader` + its unmutexed cache, `processRequest`, `selectWorkflowOLD`).
  `selectWorkflowOLD` carries the EXACT defect 239 just fixed; left unfixed deliberately
  (fixing dead code makes it look more current). Remedy for all three is deletion, as one task.
- **Unfixed and now merely visible:** `LANDMINES.md:5788`'s nested-envelope case — a
  `call_agent` child whose `config.agent_type` sits under `body` is invisible to
  `extractGroupInfo` and still runs the own-default workflow. It now logs
  `DISPATCH_OWN_DEFAULT` instead of being silent. 7 messages of that shape in an 8-day
  census. Whether that branch should also refuse is deliberately left open — the log line
  exists to size the population first.

## Commits (all on `087_towards_multiple_domains`)

`a097e3e26` the fix · `aec22b8d5` 247 dead twin · `ea5ea895f` RFC_023 · `ae0ca5b49` verdict
recorded · `209917d15` post-roll proof + the sqlDB gap fix · `a143c342a` WRONG_CALLS ·
`d07d1b921` register SYS-090 status.

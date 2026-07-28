# HANDOFF — council parallelism thread (2026-07-28)

Cold-start for whoever continues this. Written because the context got long, not
because the work stopped. Everything below is committed.

## The original question and its answer

**"Could/should we run several councils in parallel?"** Yes to both, and the
answer turned out not to need the change I first proposed.

- **Why they serialise:** all 16 council seats are `execute_llm_prompt`,
  registered `IsLocal: true` (`registry.go:307-312`), so they run **inline**;
  `consumeRequestLane` calls `processMessage` as a blocking call and commits the
  offset only after it returns (`agentbase/agent.go:611-676`); and
  `continueExecution` is a strictly one-step-at-a-time state machine
  (`coordinator.go:837-975`). One council = one goroutine for 4–9 minutes.
- **Demand is real:** 07-27 had 19 council runs; seven of the last 25 started
  ≤1.3 s after the previous one *ended* (already queued); eight ran back-to-back
  18:09→18:54.
- **The answer is the wrapper, not more lanes.** A thin orchestrator that spawns
  the council into its own pod releases the lane in ~8 s, so N councils run
  concurrently through ONE lane. **The "give councils 3 lanes" idea is superseded
  — do not build it.**

## What is BUILT and LIVE

| thing | state |
|---|---|
| `0NN_council_gate_orchestrator.sql` | **applied**; `council-gate-orchestrator` row live |
| `council-gate.idle_timeout_seconds` | 0 → **900** (0 fell through to a 3600 s default) |
| `097` `TARGET_AGENT_TYPE` env override | **live**, default still `council-gate` |
| 096 lane rollout (another thread) | **both halves shipped**; `097` now publishes to `system.agent.council-gate.requests` |

**The wrapper is deliberately NOT the default.** It is reachable only via
`TARGET_AGENT_TYPE=council-gate-orchestrator ./097_TRIGGER_council_review_v1.sh sub.json`.
**Do not flip the default until the spawn defect below is fixed.**

## The blocker

The wrapper works at the thing it was built for — **measured `QUEUE DEPTH (LAG) 0`
on `system.agent.generic.requests` while a council ran in its own pod**, which is
structural proof (the offset is committed only after `processMessage` returns, so
the inline path necessarily holds LAG ≥ 1 for minutes).

It fails one level down, at `call_council`, with
`Request … timed out after 3 retries` — the same failure as the archetype's
`call_diagnoser`. That is `bugs_open/029`, which is **owned by another thread**.

## Three of my own claims that were WRONG — check these before reusing anything

1. **"The wrapper pattern is proven."** It is not. Logged in `WRONG_CALLS.md`.
2. **The persist-ordering race** (child replies 1.92 s before the parent listens)
   — **REFUTED by the diagnosis loop**, corr `eb8df254`.
   `persistAwaitingStateWithRetry` (`coordinator.go:1863-1879`) returns early on
   *"Response already arrived during state persist - continuing"*, and
   `processResponseClaimWithRetry` retries the claim for exactly that case. I
   verified both in source. **It is a genuine non-response, not a race.**
3. **"2 of 4" as the archetype failure rate.** Measured on the wrong table.

## The measurement rule that matters most here

**Use `agent_error_log`, never `orchestration_states`.** A timed-out awaited
request does not reliably fail its orchestration: `orchestration_states` reports
`build-pipeline-trigger` as 166 COMPLETED / **0 FAILED / 0 timeouts** over 14
days, while `agent_error_log` holds **79** `spawn_dispatch` timeouts for it in 29
hours.

```sql
SELECT agent_type, step_name, action, count(*), min(occurred_at), max(occurred_at)
FROM agent_error_log WHERE error_message ILIKE '%timed out after%' GROUP BY 1,2,3 ORDER BY 4 DESC;
```

**Free reproducer:** `build-pipeline-trigger` fires every 30 s on cron and
reproduces the spawn defect dozens of times an hour **during a busy window**.
Never pay for a diagnosis run to reproduce it. It is **bursty** — an idle hour
shows zero, and so does a genuinely good hour, so always run the control:

```sql
SELECT date_trunc('hour', created_at), count(*),
       count(*) FILTER (WHERE owner_agent_type='build-pipeline-trigger')
FROM orchestration_states WHERE created_at > now() - interval '12 hours' GROUP BY 1 ORDER BY 1;
```

## Where it stands (07-28 12:38) — the mechanism is FOUND, and it is not a spawn bug

**Superseded: my 10:00 "abated" reading was wrong** and is struck through in
`029`. Extended with the same reproducer: `09:00` 21 spawns/2 timeouts, **`10:00`
24/16, `11:00` 24/11**. The clean hour was a trough.

**The `chassis_replica_scaling` thread has since caught the mechanism end to end**
(commits `42c7a7317`, `8b7638de5`), and it retires the whole "spawn handshake"
framing this thread was working under:

- **It is a queue-vs-timeout treadmill, not a spawn race and not the roll.** The
  response is *not lost* — it is *late*. Evidenced hop by hop: the git-adapter
  answered in 3.0 s (produced 10:34:45Z) and the chassis processed that response
  at 10:37:35Z — **2m50s of response-lane queueing, five seconds after the await's
  final timeout**. An await fails when callee-queue plus response-lane delay
  exceeds 3 minutes, and F2's re-queue-at-the-back turns that into a treadmill.
- **Post-roll degradation is response REPLAY**: ~49 msg/min drain, so **2–3 hours
  of response deafness per restart**, load-independent and *growing daily* as the
  responses topic grows.

That is consistent with everything this thread measured and explains what I could
not: the burstiness (queue-depth dependent), the roll correlation (replay), why
`build-pipeline-trigger` dominates (fires every 30 s, most exposed), and why
`page-rerender` is clean (fewer, faster hops). It also confirms the diagnosis
loop was right to refute my persist-ordering race.

**Implication for the wrapper:** the blocker is response-lane latency, so adding
a spawn hop to the council makes it *more* exposed, not less. Do not retry the
wrapper on the assumption that a fresh chassis fixed anything.

## Next actions, in order

1. **Follow `chassis_replica_scaling`, do not re-diagnose.** They own the
   mechanism and the fix. This thread's remaining job is to retry the wrapper
   *after* their fix lands, not to investigate further.
2. **Retry the wrapper only once response-lane latency is fixed.** One submission
   with `TARGET_AGENT_TYPE=council-gate-orchestrator`. Success = wrapper reaches
   `AWAITING_RESPONSES` in ~24 s, a dedicated pod appears, generic LAG stays 0,
   and a `council_report` lands under the submission correlation.
3. **Then flip `097`'s default** to `council-gate-orchestrator` and the council
   becomes concurrent. Not before.
4. **`bugs_open/124`** (one diagnosis item runs twice, two correlations) is
   separate, unowned, and cheaper — fix candidate 1 is to make a diagnosed item
   terminal, which makes the duplicate unrepresentable. **Note it is now doubly
   worth fixing**: duplicate chains double the response-lane load that is
   causing the timeouts.

## Landmines paid for in this thread

- **Never cancel the failing orchestration or reap its pod before the diagnosis
  runs.** I did, and the loop then had no specimen to examine. The failing row IS
  the evidence.
- **A `failed` `needs_diagnosis` item does not mean the diagnosis failed** — mine
  succeeded and was stamped `failed` by a later dispatch-loop re-run (that is
  `124`).
- **A code-tier 090 run with no explicit subject writes NO `doc_notes` row**
  (`persist_note` → `{"reason":"no explicit subject","persisted":false}`). The
  verdict lives in `orchestration_states.collected_data->'verdict'`.
- **`fan_out` is half-built**: validation contract
  (`platform/validation/workflow.go:98-105`) and fuel cost
  (`governance/fuel.go:17`) exist, but there is **no handler in the action
  registry**, so a `fan_out` step fails at `getActionHandler`. That is the route
  to intra-council parallelism (16 seats concurrent, ~6 min → ~1 min) if anyone
  wants it — a separate, larger piece of work.
- Ollama is **not** on the council's path: all 16 seats are
  `provider=anthropic, model=claude-sonnet-5`, and the other 26 steps are
  deterministic. Parallel councils load the Anthropic API, not cluster inference.

## Files

`bugs_open/096` (lane — rolled out by another thread), `bugs_open/029` (the
blocker, owned elsewhere), `bugs_open/124` (mine, unowned),
`0NN_council_gate_orchestrator.sql`, `097_TRIGGER_council_review_v1.sh`,
`WRONG_CALLS.md` (three entries from this thread).

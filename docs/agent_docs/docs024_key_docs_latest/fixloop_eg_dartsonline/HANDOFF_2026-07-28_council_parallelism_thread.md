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

## Where it stands right now (07-28 10:00, v1.0.1182)

**Abated, not proven fixed.** 07-28 `09:00` hour: **21 `build-pipeline-trigger`
runs, 0 timeouts**, on a busy fleet (150 orchestrations). Against 07-27's 25/12,
24/1, 23/13 — note the 24/1, which is why one clean hour is not enough.

Two timeouts at 09:57:29 and 09:59:59 are `[INFERRED]` roll-casualties (chassis
rolled 09:55:02Z), not the defect.

**`afbd005f9` did NOT cause this** — *"CLAIM_RECOVERY may no longer steal a live
claim"* is adjacent code but was committed at 10:52, after the clean window and
after the running image was built. Whatever changed between `1180` and `1182` is
unidentified.

## Next actions, in order

1. **Watch one more busy hour** on `agent_error_log` before anyone concludes the
   spawn defect is fixed. If it stays clean, `029` may be closable — which would
   unblock everything below.
2. **If clean: retry the wrapper.** One submission with
   `TARGET_AGENT_TYPE=council-gate-orchestrator`. Success = wrapper reaches
   `AWAITING_RESPONSES` in ~24 s, a dedicated pod appears, generic LAG stays 0,
   and a `council_report` lands under the submission correlation.
3. **Then flip `097`'s default** to `council-gate-orchestrator` and the council
   becomes concurrent. Not before.
4. **`bugs_open/124`** (one diagnosis item runs twice, two correlations) is
   separate, unowned, and cheaper — fix candidate 1 is to make a diagnosed item
   terminal, which makes the duplicate unrepresentable.

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

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

## Update 14:54 — their fix looks to be HOLDING, and the council is DARK until 08-01

**Timeouts by hour, with the control** (`build-pipeline-trigger` spawns = the
reproducer; a zero is meaningless without it):

| hour | orch runs | bpt spawns | timeouts |
|---|---|---|---|
| 10:00 | 114 | 24 | **16** |
| 11:00 | 132 | 24 | **11** |
| 12:00 | 109 | 10 | 0 |
| 13:00 | 101 | 4 | 0 |
| 14:00 | 81 | 4 | 0 |

Three clean hours. **State this carefully** — the reproducer fell from 24/hr to
~6/hr, which is exactly the weak control that made my morning "abated" call wrong.
What makes this different is the size of the prior rate: 18 spawns across the
three clean hours, against a morning rate of 46–67%, which would have predicted
roughly 8–12 failures. Zero is a real signal. Do **not** convert that into a
confidence figure — the failures are *bursty*, i.e. clustered, so any calculation
assuming independence overstates it.

**The discriminating test, and it passed.** The chassis rolled again at
**14:25:40Z** (now `v1.0.1189`). Every previous roll produced a post-roll timeout
burst — 22:06 → fourteen timeouts across 22:15–22:45; 09:55 → timeouts at 09:57
and 09:59. This roll produced **none** in the ~29 minutes after. That is precisely
the window `chassis_replica_scaling`'s response-replay finding predicts should be
degraded, so a clean one is evidence their CS-3a fix ("response replay closed")
works — not just that the fleet went quiet.

**BLOCKER for step 2: the council gate is DOWN until 2026-08-01 00:00 UTC**, on
the fleet's LLM spend cap (`chassis_replica_scaling/NOTES:412-441` — *"You will
regain access on 2026-08-01 at 00:00 UTC"*). No submission of any kind is possible
before then, so the wrapper retry cannot happen until 08-01 regardless of whether
the spawn defect is fixed. Nothing to do about it but wait.

## RESOLVED 2026-07-28 evening — wrapper PROVEN, default NOT flipped, thread complete

**The wrapper works.** First end-to-end run: published 15:20:07Z → wrapper at
`call_council`/`AWAITING_RESPONSES` by ~15:20:44 (37 s) → 16 seats in a dedicated
pod (`agent-council-gate-446e54ab-58cbg`) → `council_report` 15:26:59Z → wrapper
`COMPLETED` 15:30:52Z. **~10.75 min end to end, generic-lane LAG 0 throughout.**
The step it died on in the 07-27 attempt was cleared in 37 seconds.

**The default was NOT flipped, and that is the right outcome.** The flip proposal
went to the gate on corr `f5da8f65-a3ec-4d16-8254-3dbfcb76953c` → **REVISE**,
gated by the guardian, 11 reviewers / abstained 5 / **unreadable 0** (a real
verdict, not the harness). Objections accepted rather than argued:

- **[HIGH]** the plan overwrote a DO-NOT-FLIP header claiming *"the blocker was
  never in this wrapper"* while `003`/`029` remain open and would be inherited by
  the council path for the first time. My evidence resolved an **adjacent** bug.
  The guardian caught me asserting a resolution my own risks section contradicted.
- **[MEDIUM]** it would move every pipeline's default onto a path proven exactly
  once, for a benefit **the submission itself called marginal**.
- **[MEDIUM]** concurrency replica-bounded (2) → k8s-bounded (N) shifts load onto
  the shared Anthropic budget with headroom UNMEASURED, interacting with `124`.
- **[MEDIUM, tooling_provenance]** decision history belonged in `doc_notes`, not
  baked into a SQL comment header. **Done** — `doc_notes`
  `subject_type='pipeline'`, `subject_key='council-gate-orchestrator'`.

**No resubmission is planned.** The council's reasoning is sound and the urgency
is gone: the dedicated lane already stopped councils blocking other work, and
`replicas=2` already delivers two concurrent councils. Re-arguing would spend
credits on a change I no longer think is justified *now*.

**Preconditions for a future thread to adopt it:** `003`+`029` closed or
explicitly accepted; Anthropic concurrency headroom measured; `124` fixed. Then
change the 097 default and the seed header **in the same edit**.

## Standing state re-verified 2026-07-28 20:58 on v1.0.1194 (2 replicas)

Everything this thread owns still holds after the day's ~10 rolls:

| check | result |
|---|---|
| `council-gate-orchestrator` row | active, `start_step=spawn_council` |
| `council-gate` row | active, `idle_timeout_seconds=900` (this thread's change, survived) |
| `097` default | still `council-gate` — **not flipped**, as decided |
| `bugs_open/124` | **CLOSED** by another thread (`e42fb1ba8`, "one chain, not two, proven on a real diagnosis") — moved to `bugs_closed/` |

**`124`'s landmine discharged for this roll.** Memory records that migration 258
needs chassis ≥1191 and that you must pod-grep after *every* roll or the diagnose
lane stops silently. Checked on **both** replicas of `v1.0.1194`:

```
agent-chassis-74dbd9c9f4-7p6d8 : 1
agent-chassis-74dbd9c9f4-rxb52 : 1     # strings /app/agent-chassis | grep -c "unknown execution-context field"
```

Both carry it, so the precondition holds. Worth doing per-replica rather than
once: with `replicas=2` a retag or a partial roll can leave the two pods on
different binaries, and a single grep would not show it.

**Post-roll timeout window: NOT yet a clean bill, and deliberately not claimed as
one.** 10 minutes after the 20:48:11Z roll: 0 timeouts, 11 orchestrations — but
**0 `build-pipeline-trigger` runs**, so the reproducer never fired and the window
says almost nothing. This is exactly the shape of the reading I got wrong earlier
today (see the struck-through "ABATED" section in `bugs_open/029`). Whoever picks
this up should re-run the hourly table with the control column during a **busy**
window before drawing any conclusion about the replay fix.

## Next actions (superseded above — kept for the record)

1. **Follow `chassis_replica_scaling`, do not re-diagnose.** They own the
   mechanism and the fix. This thread's remaining job is to retry the wrapper
   *after* their fix lands, not to investigate further.
2. **Retry the wrapper — EARLIEST 2026-08-01**, once the council gate is back and
   response-lane latency is confirmed fixed. One submission with
   `TARGET_AGENT_TYPE=council-gate-orchestrator`. Success = wrapper reaches
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

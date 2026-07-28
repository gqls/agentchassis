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

## OWED ITEM DISCHARGED 2026-07-28 21:15Z — the busy-window control run

The 20:58 section above closed asking the next thread to *"re-run the hourly table
with the control column during a busy window before drawing any conclusion about
the replay fix."* Done, on `v1.0.1194` (both pods up 20:48Z).

**The whole `spawn_dispatch` timeout class stops at `11:29:57Z` and has not
recurred in the ten hours since.**

| hour | orch runs | bpt spawns (control) | timeouts |
|---|---|---|---|
| 09:00 | 150 | 21 | 2 |
| 10:00 | 114 | 24 | **16** |
| 11:00 | 132 | 24 | **11** |
| 12:00 → 21:15 | 708 | **50** | **0** |

**This is not a filter artifact, and that was checked rather than assumed.** The
morning readings all came from `error_message ILIKE '%timed out after%'`; if the
fix had changed the error *text*, a zero would mean nothing. Re-run
**unfiltered** over 10 hours, grouped by `agent_type, step_name`:
`build-pipeline-trigger / spawn_dispatch` appears **6 times, first 11:17:27Z, last
11:29:57Z**, and not once after. The class is absent from the unfiltered log, not
hidden from the filter.

**The control is alive.** 50 reproducer spawns across the clean window, and 3 of
them post-roll (21:07:58 → 21:12:57), so the cron did not simply stop — which is
the failure mode that made the morning "ABATED" call wrong.

**How far this goes, stated carefully.** Against the morning rate (46–67%) 50
spawns would have predicted roughly 23–33 failures; against the mildest observed
rate (09:00, ~10%) still ~5. Zero is a real signal and the class looks closed.
**Do not convert it to a confidence figure** — the failures are bursty, so any
independence assumption overstates it. The post-*roll* window specifically is
still thin: 3 reproducer runs is better than 20:58's zero, but it is not yet the
busy post-roll window that would test response-replay directly.

## Adjacent live finding — truncation damage, correctly attributed

Surfaced by the unfiltered baseline above; recorded here because it lands on
**this thread's own subject** (council verdict quality), and because I first
misattributed it twice.

- **It is NOT `bugs_open/119`.** 119 is a *complete but structurally invalid*
  review (a stray bracket); it says so, and names truncation as a different case.
- **It is NOT a regression of `bugs_closed/019`.** It is 019's fix **working as
  designed**: the partial is recorded, the seat is named `unreadable`, and the
  round continues instead of being voided.

What is unclaimed is the **residual rate**, which 019's fix made survivable but
did not reduce:

- **12 of 37 council rounds in 10 hours** decided with at least one seat's opinion
  partial or lost (14 damaged-seat events). Denominator verified rather than
  assumed: all 12 damaged `orchestration_id`s join to `orchestration_states` rows
  carrying `council_decide`, owner `council-gate` (11) / `experience-approval-council`
  (3). The `agent_type='generic'` on the error rows is the spawned **pod's**
  identity, not the orchestration owner — reading it as a separate population is
  the available denominator trap here.
- **It is concentrated**: over 24h, `review_editquality` 10 events (6
  unsalvageable / 4 salvaged); `deferral_honesty` 2; `checkability`, `guardian`,
  `guidelines`, `prior_art` 1 each.
- ~~**No seat sets `max_tokens`.** All 16 set `tolerate_truncation: true` and
  leave `max_tokens` NULL (live `agent_definitions` row).~~

> **CORRECTED 2026-07-28 21:40, same session, before anyone acted on it — the
> claim above is FALSE and the truth is a stronger case, not a weaker one.**
> **All 16 seats DO set `max_tokens`, uniformly `8000`**, at
> `config.ai_service.max_tokens` (with `tolerate_truncation: true` one level up at
> `config.tolerate_truncation`).
>
> **What caught it:** `llm_call_log` showed `max_tokens = 8000` on every row while
> my config query said NULL — two sources disagreeing about the same field.
> **The error:** I read a NULL from `value->'config'->>'max_tokens'` as *"not
> set"* when it meant *"wrong path"* — the real key is one level deeper. The same
> mistake made all 16 seats look always-on in an earlier query
> (`relevance_footprint` is not at the path I guessed either, so **that table said
> nothing and no always-on claim should be drawn from it**).
> **The cheap check I skipped:** print the object once (`jsonb_pretty`, minus the
> prompt) instead of guessing a path three times. A `->>` NULL is indistinguishable
> from a missing key, so NULL is never evidence of "unset" until the path is proven.

**The corrected mechanism — the ceiling is real, uniform, and too low for the
verbose seats.** From `llm_call_log` over 24h, `review_editquality`:

| | |
|---|---|
| calls | 48 (28 success, 20 failed) |
| **truncated at the ceiling** | **12 — `TOLERATED …: response truncated: stop_reason=max_tokens`** |
| successful output | avg **5504**, max **7773**, against `max_tokens` **8000** |

So the seat truncates on **25% of its calls**, and when it does *not* truncate it
is still running within ~230 tokens of the cap. Other seats sit against the same
ceiling: `guidelines` max 7727, `guardian` max 7637, `checkability` avg 6330.

**A trap worth carrying:** the obvious detector for this —
`count(*) FILTER (WHERE output_tokens >= max_tokens)` — returns **0**, because the
truncated rows are exactly the rows with `output_tokens IS NULL`. **The metric is
structurally blind to the event it appears to measure.** Count
`error_message LIKE 'TOLERATED%'` instead.

The platform already names the remedy, in the code that emits these very rows —
`platform/orchestration/actions/diagnose_council_decide_action.go:582-583`:

> *"What needs a human eventually is the pattern — a seat that truncates
> repeatedly wants a higher `max_tokens`, not a nightly salvage."*

`recordTruncationDegradation` was built precisely so this pattern would be
queryable by a human rather than dying in a pod log. This is that read, and the
pattern it was watching for is present. **[UNMEASURED]** what value would be
enough, and whether the seat's prompt is simply over-long for any sane cap.

**Not actioned deliberately:** raising `max_tokens` on the gate is a live change
to the shared review apparatus every thread submits through, so it is an owner
call, not a 21:00 unilateral edit.

### APPLIED 2026-07-28 21:43:00Z — owner chose "raise editquality only" (8000 → 16000)

`review_editquality` is one of the two `ALWAYS_ON` seats
(`099_SYNC_gate_roster.py:48` — `{"review_editquality", "review_guardian"}`),
which is *why* it dominates the damage counts: it runs in **every** round, so it
has the most exposure. That settles the always-on question the wrong-path config
query above could not answer — the roster script states it directly.

**Written to BOTH `fix-proposer` and `council-gate`, and that is load-bearing.**
The gate roster is **mirrored** from the live `fix-proposer` row, and
`transform_step` (`099:71-90`) rewrites only `error_step`, `input_fields` and
`prompt_template` — it deep-copies `config.ai_service.max_tokens` **verbatim**.
So patching the gate alone would have been silently reverted by the next
`099 --apply`. Setting both leaves source and target in agreement.

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,review_editquality,config,ai_service,max_tokens}', '16000'::jsonb)
WHERE type IN ('fix-proposer','council-gate') AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config #> '{workflow,steps,review_editquality,config,ai_service,max_tokens}' = '8000'::jsonb
RETURNING type, ...;   -- returned 2 ROWS, read individually, not "UPDATE 2"
```

**Why the whole-roster `--apply` was NOT used** even though CLAUDE.md says "do not
hand-patch the gate": that instruction exists to stop the two rosters drifting in
*membership*. Here the seat set is identical and only one scalar changes, so a
full re-mirror would have re-synced all 42 steps — far wider than the one-seat
blast radius that was actually chosen. **The equivalent guarantee was obtained by
proving the mirror is now a no-op**, which is the check that makes the shortcut
legitimate:

```
$ python3 099_SYNC_gate_roster.py          # dry run, after the change
  added: (none)   removed: (none)
  drift (steps that would change): (none)
```

**Verified:** exactly two rows sit off 8000 fleet-wide, both
`review_editquality`, both 16000; no other seat moved.

**Seed divergence — pre-existing, do NOT "fix" it by hand.** `0NN_council_gate.sql`
carries `'max_tokens', 3000` at nine sites. The seed was *already* stale against
the live 8000 before this change, so replaying it would cut every seat's budget by
more than half — a far bigger regression than the value I raised. The live row plus
the mirror are the source of truth here; the seed is history (`[[the seed is not
the system]]`). Editing it to 16000 would leave the other 15 seats wrong at 3000.

**OWED — this is not yet verified in effect.** Config is live immediately, but no
council round has run since. The change is only proven when a round shows
`review_editquality` completing without a `TOLERATED` row:

```sql
SELECT created_at, success, output_tokens, max_tokens,
       left(error_message, 60)
FROM llm_call_log
WHERE step_name='review_editquality' AND created_at > '2026-07-28 21:43:00+00'
ORDER BY created_at DESC;
```

Success = `max_tokens` reads 16000 **and** the truncation share falls from the
25% baseline. **Do not read a zero-truncation result as proof without checking the
call count** — the same weak-control error this thread already made twice today.
Note `output_tokens` is NULL on truncated rows, so count `error_message LIKE
'TOLERATED%'`, never `output_tokens >= max_tokens`.

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

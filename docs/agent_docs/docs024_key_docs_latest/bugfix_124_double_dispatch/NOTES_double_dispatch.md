# NOTES — bug 124, double dispatch. Append-only, newest at the bottom.

## 2026-07-28 — picking the bug up

`scripts/who-owns.py` across the `bugs_open/` list. Most numbers came back
"OWNED or recently active" with a named workstream. 124 came back with one
commit (its own filing) and no owning workstream; the filing thread
(work-item-parallelisation) wrote *"Status OPEN, unowned. Separable from 029 and
cheaper"* and MEMORY.md now records that thread's programme as complete. Taken up.

## The filed mechanism did not survive contact

The bug offered two mechanisms. I checked both before building anything.

**Filed item 1, `[VERIFIED]`: "Nothing marks a needs_diagnosis item complete on
success."** REFUTED.

```
SELECT jsonb_object_keys(default_config->'workflow'->'steps') FROM agent_definitions
WHERE type='diagnose-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false;
 → complete, claim_item, reap_stuck, mark_failed, call_handler,
   check_claimed, mark_complete, spawn_handler, notify_scheduler
```

`mark_complete` is a `complete_work_item` step and it works — every 090-filed item
sits at `complete` or `failed`, `claimed_by='diagnose-dispatch-loop'`. The
`[VERIFIED]` marker was attached to an inference drawn from a **print statement in
the shell script** ("closing it by hand until a diagnose dispatch loop exists").
That is the failure mode the marker exists to prevent, wearing the marker.

**My own near-miss, recorded because it is the same shape:** I nearly built the
fix off `0NN_diagnose_dispatch_loop.sql`, the seed file, which is a beautifully
documented 250-line account of a config that has since drifted. I read the live
row instead because the standing rule says to. The rule is normally stated about
seeds; it applies just as hard to *comments and headers*.

**Filed item 2, `[UNVERIFIED]`: "What started chain B is not established… could be
the retry driver, a second claim, or a separate direct dispatch."** It is the
third: **the 090 trigger's own `kcat -P` publish**, at the bottom of
`090_TRIGGER_needs_diagnosis_v1.sh`. The script writes the item at
`awaiting_diagnosis` *and* publishes the envelope. The loop claims
`awaiting_diagnosis` items. Two dispatchers, one queue.

The script's own header names the invariant it is breaking:

> *"The task ships DISABLED — enable it deliberately. Until then, and for any
> ad-hoc run, THIS SCRIPT is the dispatcher."*

`scheduled_tasks.enabled = t` for `diagnose-pipeline-trigger`. Someone enabled it.
The script was never told.

## The shape, in the data

```
corr 914dc844  (= item's spec.correlation_id)  orchestrator 07:03:04, agent 07:03:20   ← the script's publish
corr c89c718c  (fresh)      dispatch-loop 06:59:39, orchestrator 06:59:53, agent 07:00:08   ← the loop
```

Chain **without** a `diagnose-dispatch-loop` row = the script. Chain **with** one
= the loop. 6 of 6 items inside the retention window have both. Each duplicate is
a 12–14 minute `diagnose-agent` run (longest observed 31m).

## A number I refused to write down

I started to write "every `needs_diagnosis` item ever filed ran twice". The query
that would have supported it counts orchestrations over all history — and
`orchestration_states` is on a retention clock, so the 15 items older than a few
days show **0** chains, not 2. Zero there means "the rows were reaped", not "it
did not happen". The honest claim is the windowed one: 6 of 6 since 2026-07-27
19:04, which is as far back as the evidence reaches. (Standing landmine: every
history table is on a retention clock — record a rate, not a count.)

## Correcting the neighbouring bug's evidence

`029` §6 and `124` both cite orchestration `41d64b75` as the loop *"re-dispatching
an already-diagnosed item, 43 minutes after that diagnosis finished"*.

```
SELECT orchestration_id, correlation_id, owner_agent_type, created_at, updated_at
FROM orchestration_states WHERE orchestration_id::text LIKE '41d64b75%';
 → 7803075d… | diagnose-dispatch-loop | 2026-07-27 20:08:16 | 2026-07-27 20:52:32
```

Created **20:08:16** — 91 seconds after the item was created at 20:06:45. It is
the concurrent duplicate, not a later re-dispatch. What was sent at 20:49:31 was a
**retry** of its `call_handler` request, which is `029` proper. The *conclusion*
survives intact and gets stronger: the duplicate chain failed and wrote `failed`
over an item whose other chain had returned a REFUTED verdict, so counting
`failed` needs_diagnosis rows over-counts `029`. Only the story of *how* is wrong.

Both bug files get this correction written back into them.

## The throughput objection — checked, not assumed

Making the loop the sole dispatcher puts every diagnosis behind a task with
`max_concurrent = 1`. That reads like fleet-wide serialisation behind one 13-minute
run, which would have been a real reason not to do it.

It is not what happens. `cmd/scheduler/main.go:287` calls `stampCompleted`
**immediately after publishing** — fire-and-forget, both timestamps advanced — and
`countInFlight` counts *rows in `scheduled_tasks`*, not running orchestrations. The
slot frees on the next tick. Observed overlap in the live data: dispatch-loop runs
`04920015` (11:25:51→11:38:57) and `2184add8` (11:33:22→11:46:35) ran concurrently.

Dispatch latency *improves*, too: loop start was +55s / +40s / +79s after intake,
against +4m19s for the script's own publish on the same item — the direct publish
queues behind the shared generic requests lane, the scheduler tick does not.

## The half of the bug nobody had noticed

`diagnosis_artifacts` rows are keyed on `params.ExecutionContext.CorrelationID` —
the **envelope** correlation — not on the `correlation_id` the loop passes down
through `input_mapping`. So for an item the loop dispatched, `spec.correlation_id`
names **nothing**. It resolves today only because the script's duplicate chain
happens to run under it.

Both `diagnosis-triage`-created items have a `spec.correlation_id` minted by
`triageSpecJSON` that no run ever used. Those two were never double-dispatched —
and they are the ones whose trail is completely unjoinable. The two halves of the
fix are one change: kill the direct publish without fixing the join and the
printed `SAVE: CORRELATION_ID=…` stops resolving for everybody.

## 2026-07-28 later — build

Three parts, ordered by what closes the door: the claim becomes the ticket to
dispatch (P1), the script asks the DB who the live dispatcher is rather than
assuming (P2), and `query_database` gains a generic `$ctx.` parameter namespace so
the loop's claim can stamp the run's own correlation onto the item (P3). Details
and the declined alternative (correlation override on `call_agent`) in the PLAN.

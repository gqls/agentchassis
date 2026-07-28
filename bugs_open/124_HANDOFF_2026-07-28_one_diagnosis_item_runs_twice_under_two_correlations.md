# 124 — one `needs_diagnosis` item runs TWICE, under two different correlations

**Filed** 2026-07-28 from the council-parallelism thread, out of `bugs_open/029` §6
(2026-07-28 section), where it first showed up as a wrong number in someone else's
evidence rather than as a bug in its own right.

**Severity** medium. It is not a correctness fault in any diagnosis — both runs
produce valid verdicts. It costs **double the LLM spend per diagnosis**, and it
**corrupts the evidence base of another open bug**, which is how it was found.

**Status** OPEN, unowned. Separable from `029` (hung spawns) and cheaper.

## Symptom — a live specimen, captured 2026-07-28 07:0x on chassis v1.0.1180

One work item, `needs_diagnosis:robot-hands-com-learning-cen`, created 06:58:44
with `spec->>'correlation_id' = 914dc844-…`, claimed 06:59:39. It produced **two
complete, independent diagnosis chains** — each of which is an LLM-heavy
`diagnose-agent` run:

```
corr c89c718c…   diagnose-dispatch-loop  06:59:39
                 diagnose-orchestrator   06:59:53
                 diagnose-agent          07:00:08   <- chain A
corr 914dc844…   diagnose-orchestrator   07:03:04
                 diagnose-agent          07:03:20   <- chain B
```

`914dc844` is the item's **own** correlation. `c89c718c` is a **fresh** one the
dispatch loop minted. Both chains ran; neither knew about the other.

Note what is *not* wrong here: every spawn in both chains shows
`awaited_requests.status = 'processed'` at `retry_version 0`. **This is the
system working normally and still doing the work twice.**

## Why it is invisible

**The two chains cannot be joined.** The 090 trigger's stated design is that one
key ties the intake item, the `diagnosis_artifacts` bundles and the terminal
`doc_notes` row together — *"carrying the SAME correlation_id, so the item, the
bundles … and the terminal doc_notes row all join on one key"*
(`090_TRIGGER_needs_diagnosis_v1.sh` header). When the dispatch loop mints its
own correlation instead of reusing `spec->>'correlation_id'`, that property is
broken for every loop-dispatched item, so the duplicate leaves no joinable trace.

## What it already cost, in another bug's evidence

`bugs_open/029`'s "fresh instance" section cites *"a second, earlier failure the
same evening (`site_work_items` `needs_diagnosis`, created 20:06:46) also went
`failed`"* as an instance of hung spawns. **It was not.** That item's diagnosis
had already **succeeded**: correlation `eb8df254-…` has two orchestrations,
`30084fbe` and `5143b54f`, both `COMPLETED`, returning a REFUTED verdict.

Its `failed` stamp came from request `c963122a-…`, which belongs to orchestration
`41d64b75` — a `diagnose-dispatch-loop` at step `call_handler`, sent `20:49:31`,
**43 minutes after that diagnosis finished**. The loop re-dispatched an
already-diagnosed item, that re-dispatch hit `029`, and the failure was written
back over a completed diagnosis.

So this bug **manufactures apparent instances of `029` and inflates its rate.**
Anyone counting `failed` needs_diagnosis items is counting some of these.

## Mechanism

Two contributing facts, one verified and one not:

1. **Nothing marks a `needs_diagnosis` item complete on success.** `[VERIFIED]` —
   the 090 trigger still prints *"closing it by hand until a diagnose dispatch
   loop exists"*. The loop now exists (`diagnose-dispatch-loop` is a live
   `agent_definitions` row) and still does not close them, so a diagnosed item
   stays claimable.
2. **What started chain B is not established.** `[UNVERIFIED]` — it could be the
   retry driver on the loop's `call_handler` request (`77ae59dd`, still `waiting`
   at capture with `timeout_at 07:34:53`), a second claim of the same item, or a
   separate direct dispatch. The item was already `claimed_at 06:59:39` when
   chain B started at 07:03:04, so whatever the path, **`claimed_at` is not
   preventing a second run.** Do not assume the retry driver without checking.

## Fix candidates, ordered by what closes the door

1. **Make a diagnosed item terminal.** Close the `needs_diagnosis` item when its
   diagnosis completes, rather than leaving it claimable for ever. This makes the
   duplicate *unrepresentable* instead of merely less likely, and it removes the
   standing pool of already-diagnosed items that the loop keeps re-picking. Check
   `workItemTerminalStatuses` / `idx_swi_dedup` lockstep before changing status
   values — that pair is a documented drift class.
2. **Make the loop reuse `spec->>'correlation_id'`** instead of minting a new one.
   Does not stop the duplication, but restores the one-key join the 090 design
   promises, so a duplicate becomes visible and countable. Cheap, and worth doing
   even if (1) lands.
3. **Establish why `claimed_at` did not exclude the second dispatch**, and give
   the claim real exclusivity if it is meant to have it.

(1) and (2) are independent and both worth having: (1) stops the waste, (2) stops
it being invisible.

## How to verify a fix

Fire one 090 run and confirm that **exactly one** `diagnose-agent` orchestration
exists for it after it completes, and that the item reaches a terminal status
without a hand-written `UPDATE`. The failing shape is two `diagnose-agent` rows
minutes apart under different `correlation_id`s:

```sql
SELECT owner_agent_type, correlation_id, created_at FROM orchestration_states
WHERE created_at > now() - interval '30 minutes' AND owner_agent_type LIKE 'diagnose%'
ORDER BY created_at;
```

Do not verify during an idle window — see `029`'s note that the fleet reports zero
when nothing is running.

## Landmine for whoever picks this up

**Do not cancel or clean up the duplicate chains to tidy the queue before the
diagnosis has run.** A previous attempt on the neighbouring bug cancelled the
failing orchestration and reaped its pod before filing, and the diagnosis loop
then recorded that it could find no specimen to examine (`WRONG_CALLS.md`,
2026-07-27). The duplicate chains ARE the evidence.

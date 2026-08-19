# HANDOFF — 2026-08-19 — `bugs_open/029`, continue here

Supersedes `HANDOFF_2026-08-18b_continue_here.md` (bannered), which supersedes `..._2026-08-18_...`.
Read this, then `NOTES_retry_kills_live_child.md` (newest at the bottom).
`README_where_we_are.md` is the owner's plain-prose log — append, never rewrite.

## State in one line

**Part A is fixed, approved, live and PROVEN. The evidence-loss hole is CLOSED (RSH-011,
`wedge-evidence-capture` — live, and its capture path proven by induction). 029 stays OPEN —
"rare, bursty, unexplained, not currently reproducing" — and the diagnosis loop has now said so
independently.**

> **UPDATE 2026-08-19 ~10:45Z — two things below are now WRONG and are corrected in place.**
> (1) **The 08-17 evidence never expired** — `awaited_requests` retains 7 days and holds **20**
> reconstructible instances (two more than `orchestration_states` ever showed). The 090 is filable
> **today**; it does not wait for a burst. (2) **Candidate 1 (the ticker's shared 60 s context) is
> REFUTED** on three independent measurements, and the open question is now narrowed to a single
> transition: `continueExecution` for `iter_{N+1}_call_handler`, on the response-consumer goroutine,
> straight after the spawn reply was handled. Full working: NOTES, 2026-08-19 entry.

> **UPDATED 2026-08-19 10:10Z.** The 090 below is no longer "not dispatched": it RAN and
> returned **NOT CONFIRMED (UNVERIFIABLE)**, refuting on absence of evidence exactly as this
> handoff predicted. Run corr `b346d0d4-bf9b-4068-9db7-5af18d719706`, 3 iterations, ~12 min —
> the first terminal answer this lane has had from the loop. Its own data request asked for *"a
> currently-stuck instance ... to substitute fresh occurrence evidence for the 2026-08-17 burst
> that has already aged out of retention"*; it found a candidate at
> `process_item_iter_4_spawn_handler`, tested it, and refuted — that orchestration COMPLETED
> normally at 09:52:14, i.e. in flight, not wedged. **The loop was RIGHT to refute.**
> ~~**Re-file when there are CAPTURED instances**~~ — **CORRECTED 2026-08-19 ~10:45Z: the refutation's
> premise was FALSE and no wait is needed.** The burst had *not* aged out of the fleet, only out of
> `orchestration_states`; **20 instances are still in `awaited_requests`** (7-day retention, good to
> ~08-24). The loop refuted on an absence created by pointing it at the wrong table. Re-file now,
> naming `awaited_requests` as the evidence source and stating that `orchestration_states` is empty
> BY RETENTION. Add the `wedge-evidence:<orch_id>` note keys to the runtime tier when a burst is
> captured, but that is an addition, not a precondition.

## ⏳ IN FLIGHT — the re-filed 090, verdict NOT YET READ

Filed 2026-08-19 **11:05Z** after the corrections below. **RUN correlation
`d8af5f78-98bd-46fa-85b0-2a6899617db8`** (intake `aa45c007-3bfc-4c80-9fa5-022ca895a4a4` — not the
artifact key). `FORCE=1` used; the only coverage hit was the same terminal `failed` item from
2026-08-12. Seed scope: `coordinator.go:continueExecution,handleCompleteResponse,persistAwaitingStateWithRetry`,
`loop_error_handler.go:skipToNextLoopIteration`, `agentbase/client.go:processResponse` — **all five
symbols verified present on `origin/087_towards_multiple_domains` before dispatch** (HEAD was 201
commits ahead by then, and the loop reads origin).

The symptom differs from the morning's in the one way that matters: it names **`awaited_requests`**
as the evidence source and states that `orchestration_states` is empty *by retention*, so the loop
cannot re-refute on the same absence. Progress at last check: `diagnose-agent` at `assemble_bundle`.

```sql
SELECT owner_agent_type, current_step, status, updated_at::timestamp(0)
  FROM orchestration_states WHERE correlation_id::text='d8af5f78-98bd-46fa-85b0-2a6899617db8' ORDER BY created_at;
```

**VERDICT READ 2026-08-19 ~11:10Z: `UNVERIFIABLE`, 1 iteration, no citations — and the reason is a
HARNESS GAP, not the symptom.** The correction worked: the loop's `RuntimeSite` names
`awaited_requests` and its `NextScope` walks the response path. It then asked for the table's
columns because **`awaited_requests` is not in the diagnosis bundle's schema listing**, and the run
ended before that was answered. **So: do NOT re-file this symptom unchanged — it will stall in the
same place.** Get `awaited_requests` into the bundle, or inline `\d awaited_requests` into the
symptom. Full read-out: NOTES §8, 2026-08-19.

~~**READ THE VERDICT BEFORE DOING ANYTHING ELSE ON THIS LANE.**~~ A CONFIRMED answer changes what is
worth building; a REFUTED one is still a result and must be recorded as a visible correction here.

## What is DONE — do not redo any of it

| | evidence |
|---|---|
| **Part A proven behaviourally** | `call_dispatch` rv1 granted **00:15:00** (its declared 900s) at 2026-08-18 18:28:21Z, `status=processed` — answered INSIDE the window. Against **219 pre-roll retries, none above 05:00**. rv0 control unchanged |
| **Still shipping on `v1.0.1314`** | build point `d3590ca46` PRESENT both replicas, `deadbeef…` control absent; `bf7646a29`/`2a3d30ec3` are ancestors of it. rv0 windows unchanged on the new build |
| **Initial-wait lead REFUTED** | `call_diagnoser` rv0 is 30:00 in **29/29** rows, whole history; 18 agent/step pairs joined per owning agent, **18 honoured, 0 mismatches** |
| **Wedge signature measured** | 18/18 entered from the ERROR path; final spawn registered **twice** in 17/18; last state write **9–16s** after that final send; lifetimes **25:14–25:22** |
| **Evidence loss CLOSED (RSH-011)** | `wedge-evidence-capture`, hourly CronJob at `:17`, LIVE 2026-08-19. Captures LIVE wedges at freeze+30min **and** reaped ones, with the full `awaited_requests` set, into `doc_notes`. **Proven by induction:** 5 real orchestrations captured at threshold 0; re-run captured 0 (dedupe holds); induced rows deleted |
| **Dormant twin re-checked 08-19** | `helpers.go retryTimedOutRequest` (hard 30s) — still no constructor in this repo; only a test-file *comment* references `TimeoutMonitor` outside `helpers.go`. Scoped claim holds |

## THE OPEN QUESTION, unchanged

**What kills the parent a few seconds after its spawn handshake SUCCEEDS, on the path taken when
the previous iteration's `call_handler` ended in `error`?**

**NARROWED 2026-08-19 — the question is now one transition, not "somewhere after the handshake".**
The spawn's `processed_at` and the parent's last state write are **the same event** (37/37 rows;
`handleCompleteResponse` stamps `processed_at` BEFORE `continueExecution` runs). So the child
answered, the parent handled the answer, wrote state — and then **died inside `continueExecution`
for `iter_{N+1}_call_handler`, on the response-consumer goroutine.** Not in the ticker, not in the
spawn, not in the retry path.

- ~~`cleanupExpiredAwaitedRequests` (shared 60s context)~~ — **REFUTED 2026-08-19, three ways.**
  (1) The budget is never shared: `processing_started_at` is stamped `NOW()` per claim batch, and
  **31,548 of 31,548 claims are batches of exactly ONE** (98.7% row coverage, whole 7-day window,
  burst included). (2) It is not exhausted either: the continuation dies **~12–35 s into the 60 s**
  (claim→spawn 3–19 s, spawn→freeze 9–16 s). (3) The path it dies on carries **no deadline at all** —
  `c.ctx` in `agentbase/client.go` is the agent-lifetime context. NOTES 2026-08-19 §2.
- `handleSpawnRetry` (`coordinator.go:1677`) — still open as the **duplicate**-registration
  candidate. Duplicate gap measured at **06:54–09:37** across 17 of 20, which is consistent with the
  already-established >5-min takeover sampled by a 5-min replay cycle, not a new mechanism.
- **NEW, `[UNVERIFIED]`, the shape most worth testing:** what differs about an iteration entered
  from the error path, given `call_handler` registers fine on every healthy iteration of the *same*
  orchestration. `skipToNextLoopIteration` writes `iter_N_error`/`error_count` into `CollectedData`,
  then the spawn parks through `persistAwaitingStateWithRetry`, which **reloads fresh state and
  copies only awaited entries + status** (the `LANDMINES.md` discard). Whether those keys survive
  the park has a definite answer nobody has fetched.
- Pod death (OOM/crash) at that instant would produce an identical trace, and is untestable for the
  08-17 rows. Cheap to settle on the next burst — worth adding to the RSH-011 capture.

## What changed overnight, and it cuts both ways

- ✅ **The 090's blocker is GONE.** Origin advanced; `retryWindow` **is** on
  `origin/087_towards_multiple_domains`; HEAD is **17** commits ahead, not 233. A diagnosis filed
  now reads a tree that carries the fix.
- ~~❌ **The evidence is GONE.**~~ **CORRECTED 2026-08-19 ~10:45Z — WRONG, and it was wrong when
  written. The evidence survives in `awaited_requests`, which retains SEVEN DAYS.** What expired is
  `orchestration_states` (~26 h, **0** wedged rows retained — that half is right). The wedge
  signature is fully reconstructible from `awaited_requests` alone: an `iter_N_call_handler` at
  `retry_version>=3 / status='error'`, the following `iter_{N+1}_spawn_handler`, and **no**
  `iter_{N+1}_call_handler` row. `[MEASURED]` that returns **20 instances on 2026-08-17 — two MORE
  than `orchestration_states` ever showed** — including `23eb0107`, the worked example in NOTES.
  **20/20 have no next `call_handler`.** Retained until ~2026-08-24. **So a 090 filed today DOES
  have instances to cite** — see NOTES 2026-08-19 §3.

~~**So do not file the 090 today on an empty table.**~~ **It WAS filed on 2026-08-19 (the owner asked for it) and returned NOT CONFIRMED on exactly that ground — see the banner at the top.**
**⚠ AND THE GROUND WAS FALSE.** The table was not empty; the loop was pointed at the wrong table,
and refuted on an absence that only existed in `orchestration_states`. A re-file is owed and is
**not** blocked on the next burst — it must name `awaited_requests` as the evidence source and say
plainly that `orchestration_states` is empty BY RETENTION, so the loop does not re-refute on the
same absence. The rest of the seed below still applies — the
symptom text is ready in NOTES (2026-08-18 17:05Z), the seed scope is
`coordinator.go:continueExecution,handleSpawnRetry,createContinuationContext,handleRecoverableError`,
and `FORCE=1` is legitimate (the only coverage hit is a terminal `failed` item from 08-12).

## ⚠ THE TRAP THAT WILL CATCH THE NEXT READER

**Post-roll quiet is NOT the fix working.** The wedge's entry condition — a `call_handler`
reaching terminal `error` — by day:

| 08-12 | 08-13 | 08-14 | 08-15 | 08-16 | **08-17** | **08-18** | 08-19 |
|---|---|---|---|---|---|---|---|
| 0/448 | 0/232 | 0/1241 | 1/1302 | 0/481 | **30/1436** | **0/1603** | 0/20 |

**Thirty of thirty-one are one day**, and **08-18 was already zero across 1,603 rows — mostly on
the UNFIXED binary.** Six of eight days are zero. Any comparison of a pooled pre-roll rate
against post-roll silence is fiction (I built one, hedged it, and it was still worthless —
`WRONG_CALLS.md` 2026-08-19). **A quiet period only means something if the baseline is not also
quiet.**

## What is actually left on this lane

1. ~~**Wait for the next burst, then act inside ~26 hours.**~~ **SUPERSEDED 2026-08-19: re-file the
   090 NOW against the 20 retained `awaited_requests` instances** (§3 of the 08-19 NOTES entry), and
   correct the symptom to name that table. The ~26-hour clock governs `orchestration_states` only;
   the usable window on the 08-17 burst runs to **~2026-08-24**. Waiting for a fresh burst was never
   required. Check with the RUNBOOK's wedge census; if rows exist, transcribe them AND
   file the 090 the same day.
2. ~~**Decide whether the 26-hour evidence window is acceptable.**~~ **DONE 2026-08-19 —
   RSH-011 `wedge-evidence-capture`.** Hourly CronJob at `:17`; captures LIVE wedges at
   freeze+30min (~3.5h before either the reaper or cleanup reaches the row) **and** reaped ones,
   with the full `awaited_requests` set, into `doc_notes` keyed `wedge-evidence:<orch_id>`.
   **Capture path proven by induction** (5 real orchestrations captured at
   `WEDGE_FROZEN_MINUTES=0`; re-run captured 0, so dedupe holds; induced rows then deleted).
   ⚠ **Why LIVE capture and not just reaped:** the reaper terminates at 4h and cleanup DELETES
   `EXECUTING_STEP` at 4h — the same threshold — so reaped-only evidence is **a lower bound on a
   population nobody can enumerate**. Nothing further is owed here; read the notes when a burst
   happens.
3. **Nothing else.** Part A is closed out; the initial-wait lead is refuted; the dormant twin is
   deliberately unfixed with a re-verified scoped claim.

## Can 029 be closed? NO — and the honest reason

The bar is **fixed AND live**. Part A is both, but Part A was never the wedge — every artefact
says "part A only" in those words. The wedge mechanism is **unexplained**, and its absence since
08-17 is indistinguishable from the six other quiet days in the record. **Closing on that
evidence would be closing on a baseline.** Its one observed burst overlaps a GitHub API incident,
which is a lead nobody has pulled.

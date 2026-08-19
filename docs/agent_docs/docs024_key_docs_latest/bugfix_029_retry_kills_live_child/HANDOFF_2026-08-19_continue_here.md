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

## STATE ON `v1.0.1316` (roll 2026-08-19 17:13Z) — the gating dependency is MET

| check | result |
|---|---|
| build point | **`07eeba4a1`** PRESENT on **both** replicas. Negative controls: previous build point `590ca3a20` **absent on both**; `deadbeef…` absent (confirmed on the first replica — the second probe was cut by a command timeout, so that one cell is `[UNVERIFIED]` and the PREV control carries the weight) |
| **the bundle fix is ABOARD** | `0132a3683` is an ancestor of `07eeba4a1` — so `awaited_requests` is now in `schemaAlwaysTables` **in the running binary**. Part A (`bf7646a29`, `2a3d30ec3`) also aboard |
| evidence still retained | **20 instances, all 08-17**, still reconstructible. `awaited_requests` floor has advanced to **2026-08-12 20:30**, so the 08-17 rows have roughly **three days** left |
| new build, wedges | none. 2 COMPLETED, 1 healthy `AWAITING_RESPONSES` at `process_item_iter_0_call_handler` |

**090 RE-FILED on this build: run corr `d02a6958-ddb5-4b3b-af58-0857094100d9`** (2026-08-19
~20:40Z, `FORCE=1`, same seed scope, all five symbols re-verified on origin — HEAD was 428 commits
ahead by then). The symptom now also **names the two refuted mechanisms so they are not
re-proposed** (the ticker's shared 60 s context; the response-consumer context).

**(2) THE BUNDLE FIX IS NOW BEHAVIOURALLY PROVEN — this half is DONE.** The post-fix bundle renders
`awaited_requests(request_id varchar, …, retry_version integer, …, processing_started_at timestamp,
processing_pod text, …)`; **all four pre-fix bundles across the two earlier runs render nothing**,
while the positive control `orchestration_states(` is present in **all five** — so the section was
working before and the absence was specific to this table. NOTES §11.

⚠ **Use `awaited_requests(` WITH the parenthesis.** A bare `LIKE '%awaited_requests%'` returns true
on a pre-fix bundle, because the SYMPTOM TEXT names the table and its columns and is quoted inside
the bundle. I ran that blind check first. **A check your own input can satisfy is not a check.**

**(1) Whether the wedge has a cause — STILL OPEN. Run `d02a6958` returned `UNVERIFIABLE` at
3 iterations (cap is 5; it stopped early, `stopped_reason` empty).**

**But read WHY before filing it with the other two — it is not the same result.** It **cited a real
08-17 `awaited_requests` row** (Tier 1, `Fresh: 2026-08-17`, orchestration `838f8c14`) and the right
code path (`skipToNextLoopIterationForAsync` → `skipToNextLoopIteration` → `createContinuationContext`
+ `continueExecution`, plus `handleCompleteResponse`'s continuation — the §4 transition, reached
independently). It established the **precondition** and not the **outcome**, and wrote the exact SQL
it still needed. **It ran out of iterations one query short.** Full read-out: NOTES §14.

⚠ **Do not record it as "blocked by the harness".** The first two were. This one self-corrected a
200-row truncation (`row_cap`, default 200) that its own `ORDER BY orchestration_id` had made fatal,
and carried on.

**THE NEXT RUN'S ONE CHANGE:** hand it the reconstruction **as SQL to execute in iteration 1**, so it
begins where this one stopped. Supply **the query, not the answer** — that respects the runbook's
"assert neither rows nor counts" rule while removing the discovery cost that ate three iterations.

## STATE ON `v1.0.1315` (fresh roll, pods up 2026-08-19 12:15Z) — verified at the artefact

| check | result |
|---|---|
| build point | **`590ca3a20`** PRESENT on **both** replicas (`grep -aq` on `/proc/1/exe`). **Two negative controls absent**: the previous build point `d3590ca46` and `deadbeef…`. The `build provenance` log line had already scrolled on both — absent ≠ unstamped |
| Part A (RSH-010) still aboard | `bf7646a29` and `2a3d30ec3` are ancestors of `590ca3a20`. The 236 park fix `3ba384c63` is too |
| wedge population, 7-day view | **20, all on 08-17. None since** — and this is now a SEVEN-DAY statement from `awaited_requests`, not a 26-hour one from `orchestration_states` |
| entry condition (terminal-`error` `call_handler`) | 08-18 **0/1595**, 08-19 **0/385**. Burst remains one day |
| anything wedged on the new build | none. 38 COMPLETED; one `EXECUTING_STEP` at `process_item_iter_0_done` — **in flight, not a wedge** (wrong step; the wedge is at `_spawn_handler`). This is the exact shape the morning's 090 mistook for an instance |

⚠ **The quiet still proves nothing.** Six of eight days are zero in the baseline too. What has
changed is the *window we can say it over*, not the strength of the claim.

## ✅ DONE THIS SESSION — the 090's blocker is FIXED IN CODE (inert until the next roll)

`awaited_requests` was in **neither** filter that populates the diagnosis bundle's Schema section:
it does not match the relevance include (`site%|page%|content%|flow%`) and no `SELECT` in
`diagnose_load_runtime` names it, which is the rule `schemaAlwaysTables` derives from. Invisible by
construction. **This is run `074beb8a`'s failure one table over** — that run died because
`orchestration_states` was absent, guessed a column, got 42703 and stopped; the remedy was this same
list, and it worked.

- **Commit `0132a3683`** — `awaited_requests` added to `schemaAlwaysTables`; the derivation rule
  WIDENED in the comment rather than quietly falsified (it is now "tables this action renders rows
  from" **plus** "tables a diagnosis cannot avoid addressing"); its own assertion added to
  `TestSchemaAlwaysTablesIsDeterministic` because the coverage test **cannot see why this entry
  belongs** and would stay green if someone deleted it as unused. **Proven able to fail by mutation.**
- **Council: APPROVED at ROUND 1, all reviewers approve, ZERO objections** (corr
  `e03f7122-7895-4b81-8add-5a93f69ed553`, **verdict READ 2026-08-19 ~16:2xZ**). The commit carries
  `Council-Submitted:` and **098 credits it automatically now the correlation is approved — no amend
  is needed and forward-only forbids one.** Do NOT hand-write `Council-Reviewed:` on a later commit
  to "tidy this up"; the join is already exact.
  ⚠ **Resolve the verdict BY CORRELATION, never by recency** — `ORDER BY created_at DESC LIMIT 1`
  on `doc_notes` returned a *different lane's* APPROVED note (corr `f0e95e58`, a status-vocabulary
  change) and I nearly recorded it as mine. Use
  `... WHERE categories ? 'council-gate' AND body LIKE '%<your corr>%'`.
- **Go, so INERT until the next chassis roll.** Verify after the roll at the build stamp, then by
  re-filing the 090 and checking its bundle actually describes the table.

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

1. **RE-FILE THE 090 — but ONLY after the roll that carries `0132a3683`.** Filing before it stalls
   in the same place a third time; that is now the single gating dependency. The 20 retained
   instances are good to **~2026-08-24**, so if the roll slips past that, inline `\d awaited_requests`
   into the symptom instead of waiting. ~~**Wait for the next burst, then act inside ~26 hours.**~~
   ~~**SUPERSEDED 2026-08-19: re-file the 090 NOW against the 20 retained `awaited_requests` instances**~~ (§3 of the 08-19 NOTES entry), and
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
3. ~~**Read two verdicts that are outstanding**~~ — **the council one is READ and APPROVED (round 1,
   no objections).** Still outstanding: the 090's, if anyone re-files after the roll.
3b. **Widen the include to `workflow%` — the class is NOT closed.** `flow%` is a PREFIX pattern and
   never matched `workflow%`; another lane's run `dd61df1b` stalled on exactly
   `workflow_templates`, `workflow_contract_chain`, `v_active_workflows`, `v_all_workflows` the same
   morning, and my one-table fix does nothing for it. Evidence and the reasoning: NOTES §9, 016b §9.
   Deliberately kept out of the approved round rather than folded in under a running review.
4. **Fix the step-name-keyed arrival check** in `persistAwaitingStateWithRetry` (NOTES §6). Verified
   at source, unshipped, should key on request id. Real on its own terms; `[UNVERIFIED]` as this
   bug's cause and it does **not** fit as the first failure. Council gate applies.
5. **Nothing else.** Part A is closed out; the initial-wait lead is refuted; the ticker's 60s context
   is refuted; the dormant twin is deliberately unfixed with a re-verified scoped claim.

## Can 029 be closed? NO — and the honest reason

The bar is **fixed AND live**. Part A is both, but Part A was never the wedge — every artefact
says "part A only" in those words. The wedge mechanism is **unexplained**, and its absence since
08-17 is indistinguishable from the six other quiet days in the record. **Closing on that
evidence would be closing on a baseline.** Its one observed burst overlaps a GitHub API incident,
which is a lead nobody has pulled.

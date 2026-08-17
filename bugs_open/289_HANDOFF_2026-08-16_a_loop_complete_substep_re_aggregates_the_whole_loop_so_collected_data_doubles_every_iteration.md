# 289 — a `loop_complete` substep re-aggregates the WHOLE loop from inside each iteration, so `collected_data` doubles every iteration (2^N)

**Filed 2026-08-16** by the `bugfix_281_tool_audit_ported` lane, which hit it as the
`tool-auditor` stall its handoff left open. Status: **OPEN, UNOWNED.**
**Not latent — it is live and it is already killing `tool-auditor` outright.**

**Exposure: 15 of the 18 loops in the live fleet**, including `build-dispatch-loop`,
the dispatcher that claims and runs every work item on the estate.

---

## The one-paragraph version

A loop's sub-workflow normally ends with a terminal substep — conventionally named
`done` — declared as `action: loop_complete`. `handleLoopExpansion` injects that substep
once **per iteration**, as `<loop>_iter_<N>_done`, and it runs the same
`LoopCompleteAction` that the loop's real end-step runs. That action is the *whole-loop
aggregator*: it walks iterations `0..total_iterations-1` and copies every key it can find
for each one into its own result. So iteration N's terminal step collects iterations
0..N-1 — **including their own `_done` aggregates, which already contain theirs**. Each
iteration's stored result therefore contains all previous iterations' results, nested.
Size doubles per iteration. Ten iterations is ~2^10, and `tool-auditor`'s
`collected_data` reaches **22 MB average / 29 MB max**, at which point the orchestration
can no longer be carried and it stops dead at `create_items_loop_complete`.

## Evidence `[MEASURED 2026-08-16]`

**The doubling, on three independent agent types.** Per-key sizes of
`orchestration_states.collected_data`:

| iteration | `tool-auditor` (orch `3661825d…`) | `tool-suggester` | `build-dispatch-loop` |
|---|---|---|---|
| `_iter_0_done` | — | 10,102 B | 201 kB |
| `_iter_1_done` | — | 20 kB | 617 kB |
| `_iter_2_done` | 70 kB | 41 kB | 1,419 kB |
| `_iter_3_done` | 141 kB | 83 kB | 3,023 kB |
| `_iter_4_done` | 283 kB | 166 kB | 6,247 kB |
| `_iter_5_done` | 567 kB | | |
| `_iter_6_done` | 1,134 kB | | |
| `_iter_7_done` | 2,269 kB | | |
| `_iter_8_done` | 4,538 kB | | |
| `_iter_9_done` | **9,076 kB** | | |

`tool-auditor` and `tool-suggester` are an exact 2.00×; `build-dispatch-loop` is
2.1–3.1× (it also adds genuinely new content per iteration).

**The nesting, read directly.** In `3661825d…`, `create_items_loop_iter_3_done` has
`results[0..9]` — all ten, though only four iterations had run — and
`results[0]`, `results[1]`, `results[2]` **each contain a `done` key** holding that
earlier iteration's own full aggregate:

```
-- iter_3_done results[0] keys:
done, name, iteration, item_created, original_item, check_target_class,
create_review_item, create_items_loop_item
```

That `done` inside `results[0]` is the recursion, and it is why the factor is 2.

**Totals, same query across the three same-shaped loop consumers:**

| agent type | rows | avg `collected_data` | max |
|---|---|---|---|
| `tool-auditor` | 63 | **22 MB** | **29 MB** |
| `build-dispatch-loop` | 353 | 2,813 kB | 13 MB |
| `tool-suggester` | 3 | 447 kB | 616 kB |
| `internal-linker` | 10 | 22 kB | 22 kB |

**The control — this is the part that makes the claim disconfirmable.** Three of the 18
live loops have **no** `loop_complete` substep (`page-content-writer`,
`area-sweep-orchestrator`, `thunder-training-monitor`). If the mechanism were anything
other than the substep, they would double too. `page-content-writer` does not: it has
**no `_iter_N_done` keys at all**, its per-iteration outputs are flat (`section_output_1`
= 8,435 B), and its single outer aggregate `process_sections_loop_complete` is 32 kB.
Population split (all 18 live loops):

```sql
SELECT a.type, s.k AS loop_step,
       (SELECT count(*) FROM jsonb_each(s.v->'config'->'sub_workflow'->'steps') sub
         WHERE sub.value->>'action'='loop_complete') AS n_lc_substeps
FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s(k,v)
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND s.v->>'action'='loop' ORDER BY 3 DESC, 1;
```
→ **15 with one, 3 with none.**

## Root cause, in the code

1. `loop_expansion_handler.go:100` injects every substep per iteration as
   `<loop>_iter_<N>_<substep>`, **including one whose action is `loop_complete`**. Nothing
   treats that action specially at injection time.
2. The injected iteration step gets only `loop_iteration`, `loop_item_index`,
   `loop_var_name`, `loop_name`, `continue_on_error` in its config
   (`loop_expansion_handler.go:151-158`). It does **not** get `total_iterations` or
   `substep_output_fields` — those are set only on the real end-step at
   `loop_expansion_handler.go:193-205`.
3. `LoopCompleteAction` (`actions/loop_actions.go:306`) therefore:
   - finds no `total_iterations` in its config and **falls back to
     `loop_metadata.total_iterations`** (`:339-350`) — the *whole loop's* count. This is
     why `iter_3_done` aggregates all ten iterations, not three.
   - finds no `substep_output_fields`, so **Strategy 1 collects nothing**, and it falls
     through to **Strategy 3, the generic fallback scan** (`:449-481`), which copies
     *every* key with prefix `<loop>_iter_<i>_` into its result — and the previous
     iterations' `<loop>_iter_<i>_done` aggregates match that prefix.

Step (3) is the recursion; step (2) is what forces it down the greedy path; step (1) is
the design decision that makes the whole-loop aggregator run inside an iteration at all.

## Why it presents as "the `tool-auditor` stall"

`tool-auditor` is simply the first consumer to cross the limit. It has a large base
payload (`tool_data` 67 kB + `site_specs` 79 kB + an LLM audit with 14–20 findings) and
`max_iterations: 10`, so it pays the full 2^10. That also **explains the correlation the
281 lane observed and could not account for** — that runs with findings > `max_iterations`
stall while smaller ones sometimes do not. It is not that truncation breaks a handoff:
hitting the cap simply means running the maximum number of doublings. Runs with fewer
findings run fewer iterations and stay under the ceiling.

Live damage today: of 63 `tool-auditor` orchestrations all-history, **1 COMPLETED**;
47 RUNNING (31 dead at `create_items_loop_complete`, 16 at `create_items_loop_iter_9_done`),
10 FAILED, 5 EXECUTING_STEP. The dead rows carry `awaited_requests = {}` — nothing will
ever wake them. Because the work item never completes, the claimed-item-timeout sweep
re-dispatches it, so each stalled audit costs **a fresh Sonnet LLM audit roughly every
40 minutes until the item goes `failed`.**

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Do not run the whole-loop aggregator inside an iteration (recommended).** In
   `handleLoopExpansion`, when a substep's action is `loop_complete`, inject it as an
   inert terminal marker instead (its only real job is to be the chain link that
   `resolveIterationNextStep` rolls over to the next iteration / the end-step — see
   `coordinator.go:4403-4436`, which already handles an empty `NextStep` correctly). This
   deletes the recursion outright and costs nothing semantically. **Blast radius: all 15
   loops** — every one of them currently stores a per-iteration aggregate that no
   consumer has been shown to read, which is exactly the thing to check before shipping.
2. **Make `LoopCompleteAction` refuse to aggregate when it is an iteration substep.** It
   can tell: only injected iteration steps carry `loop_iteration` in their config; the
   real end-step never does. Return an empty marker immediately. Smaller diff, same
   effect, leaves the odd design in place.
3. **Narrow Strategy 3 so it cannot copy another aggregate.** Weakest — it is name-based
   and a future substep called `done` in a loop that legitimately wants its output would
   re-open it.
4. **Independently of the above, fix the `total_iterations` fallback** (`:339-350`): a
   step carrying `loop_iteration` should never inherit the whole loop's count.
5. **Add a size tripwire** on `collected_data` at persist time — a 22 MB orchestration
   should be loud, not silent. Nothing currently notices.

## How to verify a fix

- Re-run one `tool-auditor` audit and check the progression is flat, not geometric:
  ```sql
  SELECT key, length(value::text) FROM orchestration_states, jsonb_each(collected_data)
  WHERE orchestration_id='<new run>' AND key LIKE '%\_done' ORDER BY key;
  ```
  **The disconfirming result is a ratio near 2.0 between consecutive iterations.** Flat
  (or growing only by the genuinely new per-iteration content) is the pass.
- Confirm the run reaches `complete` and the orchestration ends `COMPLETED` — the current
  population has exactly one such row, from 2026-08-15.
- Check the control did not regress: `page-content-writer` must still produce one outer
  `process_sections_loop_complete` and no `_iter_N_done` keys.
- Re-run the population query above and confirm no loop lost its end-step aggregate.

## Notes for whoever takes it

- **This is a shared platform seam — the loop engine — so it is council-scope** under
  CLAUDE.md's platform-seams rule, and the "other consumers must be TOLD, not merely
  measured" clause of the 2026-07-29 owner ruling applies: the 15 loops above are the
  consumer list, and candidate 1 changes what each of them stores.
- The 281 lane's earlier `needs_diagnosis` run `815322b9-…` **REFUTED** the hypothesis
  that truncation breaks the `loop_complete` handoff, then stopped at its iteration cap
  (`status: UNVERIFIABLE`) pointing at a Kafka `context canceled` signal. That signal is a
  **red herring for this bug**: the error rows are logged against
  `process_item_iter_N_call_handler`, which is the *parent* `build-dispatch-loop`'s step,
  not `tool-auditor`'s own loop, and there is one such row today against 31 dead rows.
- A second, separate defect is visible in the same error log and is **not** this bug:
  `complete | complete_workflow | workflow completed but its result could not be delivered
  to the parent (failed_transient): message validation failed` (6 rows, 2026-08-12→15).
  `ProduceWithValidation` (`platform/kafka/producer.go:120-140`) validates **headers**, not
  size, so that is its own thing. ~~Worth a separate file.~~
  > **CORRECTED 2026-08-17: it was already filed AND already fixed — `bugs_closed/274`**
  > (`completed_workflows_cannot_deliver_their_result_to_the_parent_fleetwide`), fixed by
  > `919cc6976` ("a validation refusal is UNDELIVERABLE not transient") at 2026-08-15 10:04Z and
  > closed the same day. I wrote "worth a separate file" **without grepping `bugs_closed/`**,
  > which is the one check CLAUDE.md names for exactly this. Nothing was filed on it, so the
  > cost was a stale line in this file rather than a duplicate bug — but the line would have
  > sent the next reader to file one. WRONG_CALLS 2026-08-17.
  > **It is extinct, and I re-measured it rather than take the closure on trust:** zero
  > occurrences fleet-wide since 2026-08-15 in an `agent_error_log` that is NOT pruned (it holds
  > 34,877 rows back to 2026-07-18 and was still being written at 11:09Z today), **with a demand
  > control** — the five heaviest former producers all ran since 08-16 and produced none:
  > page-rerender 311 runs, build-dispatch-loop 138, feed-ingester 93, page-build-handler 50,
  > page-content-writer 50. Peak was 3,136 errors on 08-11.

---

## ADDENDUM 2026-08-16 — it predates migration 425, and **the diagnosis loop cannot see this bug**

**Two `090` runs have now returned REFUTED on it. Both refutations are false, and both fail
the same way.** Read this before spending another run.

**1. The mechanism predates 425 — so it accounts for the entire stall history, not just this
week.** `1a70fed3-ffbc-42c1-b617-4295dbea37e6`, created **2026-08-04 05:56Z** (eleven days
before 425 was applied), `collected_data` **20 MB**:

```
create_items_loop_iter_0_done    20 kB
create_items_loop_iter_1_done    40 kB
create_items_loop_iter_2_done    80 kB
create_items_loop_iter_3_done   160 kB
create_items_loop_iter_4_done   321 kB
create_items_loop_iter_5_done   643 kB
create_items_loop_iter_6_done  1287 kB
create_items_loop_iter_7_done  2574 kB
create_items_loop_iter_8_done  5149 kB
create_items_loop_iter_9_done    10 MB
```

Ten iterations, exact 2.00x throughout. The oldest dead row is 2026-07-29.

**2. Run `12ffad7c-a7b2-4955-b531-554f07650598` (filed against the correct mechanism) refuted
it with a claim that is simply untrue.** Its words:

> "the actual collected_data for orchestration 1a70fed3 (tool-auditor) enumerates keys action,
> llm_audit, load_tool, tool_data, input_data, site_specs, site_record, agent_config,
> audit_result, item_created, loop_metadata, item_created_0 through item_created_8 — a
> suffixed-index pattern, not a '_iter_<N>_done' pattern, **and no '_done' key appears at all**."

Measured on that exact row:

```sql
SELECT (SELECT count(*) FROM jsonb_object_keys(collected_data) k WHERE k LIKE '%\_done')
FROM orchestration_states WHERE orchestration_id='1a70fed3-ffbc-42c1-b617-4295dbea37e6';
```
→ **10**, and the ten sizes above.

**3. Why both runs failed, and why re-running will not help.** The evidence bundle cannot carry
a 20 MB `collected_data`, so the key enumeration it shows the diagnoser is **truncated** — and
the diagnoser reads the truncation as absence. Run `815322b9…` failed the same way one step
earlier ("(no orchestration rows for this correlation/site)" in its own iteration-1 citation,
against a target that plainly has rows). **The bug's own symptom — an enormous
`collected_data` — is what blinds the tool we would normally use to diagnose it.** That is a
property of the harness, not of this bug, and it will recur on any oversized-state defect.

**Consequence for whoever takes this on:** the `090` route is unavailable here. Verify at the
DB directly with the narrow, aggregate-only queries used in this file (`count(*)` over
`jsonb_object_keys`, `length(value::text)` per key) — never `SELECT collected_data` or
`jsonb_pretty` on these rows, which times out at 120s. Per the owner ruling of 2026-07-31, the
substitution for the loop here is the first-hand verification recorded above: three agent
types, a pre-425 and a post-425 row, the nesting read key-by-key out of the stored JSON, the
code path read in full, and a control group that does not double.

---

## ADDENDUM 2 — the blast radius of the preferred fix is EMPTY, and it is measured

Candidate 1 says "stop running the whole-loop aggregator inside an iteration". The obvious
objection is *does anything read those per-iteration aggregates?* Measured, so the reviewer
does not have to (CLAUDE.md, 2026-07-28: "measure the blast-radius claim before you submit"):

```sql
SELECT a.type, s.k AS loop_step, sub.key AS lc_substep,
       COALESCE(sub.value->>'output_field','(none)') AS output_field,
       COALESCE(sub.value->>'next_step','(none)')    AS next_step
FROM agent_definitions a,
     jsonb_each(a.default_config->'workflow'->'steps') s(k,v),
     jsonb_each(s.v->'config'->'sub_workflow'->'steps') sub
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND s.v->>'action'='loop' AND sub.value->>'action'='loop_complete' ORDER BY 1;
```

**All 15 return `output_field = (none)` and `next_step = (none)`.** With no `output_field`,
the step's result lands only under the runtime-generated key `<loop>_iter_<N>_<substep>` —
a name that is minted per iteration during expansion, so **no config can address it**, and a
grep of `platform/ internal/ pkg/` for a Go reader returns nothing (only an unrelated
`was_already_done` field on the thunder adapter). The aggregates are written, stored, copied
into each other, and read by nobody. The consumers read the **outer** `_complete` step's
`output_field` (`items_created`, `processed_sections`, …), which candidate 1 does not touch.

**This also kills candidate 3 outright.** The terminal substep is *not* always called `done`:

| substep name | loops |
|---|---|
| `done` | 9 |
| `complete_page` | 4 (`pageflow-builder`, `page-rebuild`, `rerender-site`, `site-work-orchestrator.build_items_loop`) |
| `complete_dispatch` | 1 (`maintenance-triage`) |
| `task_complete` | 1 (`vet-batch-processor`) |

A name-based skip on `_done` would fix 9 of 15 and silently leave the other 6 — including
three page-building loops — doubling. **Match on the substep's `action`, never on its name.**


---

## FIX BUILT 2026-08-17 — committed `509e01e6a`, INERT until the next chassis roll

**Status: still OPEN.** The bar is fixed AND live; this is Go, so it does nothing until the
image is rebuilt and rolled. Releases are whole-fleet and the owner runs them.

**What shipped** (candidate 1, the one that makes the bad state unrepresentable):
`handleLoopExpansion` sets `loop_iteration_terminal: true` on an injected substep whose action
is `loop_complete`, and `LoopCompleteAction` returns a three-field marker
(`status`/`iteration`/`loop_name`) for such a step instead of aggregating — **before** the
existing diagnostic key dump, which on these runs was logging a 22 MB key list once per lap.
Aggregation is unmoved: it still happens on the loop's own `<loop>_complete` end-step, which is
what `output_field` exposes and what every consumer reads. New key registered in
`frameworkStepConfigKeys`; concept register **WFA-015**.

**Matched on the substep's ACTION, never its name** — per ADDENDUM 2, the terminal is called
`done` in only 9 of 15 loops.

**Marked at injection, not inferred**, so a nested loop's own end-step still aggregates. The
bare-`loop_iteration` fallback is deliberate: plans already expanded and persisted are in flight
with no flag, and without it they would keep doubling until they died.

**Tests: 5, and the guard is proven by MUTATION.** With the early return disabled, the guard
test fails and prints the recursion verbatim —
`results[0].done.results[0].done.nested` carrying the prior-iteration marker twice. The guard
test also carries a **control** that strips only the two terminal signals from otherwise
identical input and requires the OLD swallowing behaviour back, so it cannot pass against a
`LoopCompleteAction` that merely never aggregates. Also covered: the in-flight-plan fallback
with a `float64` iteration index (a JSON round-trip through the persisted plan), a nested inner
end-step that must still aggregate, and a table for the discriminator including the two shapes
that must NOT trip it.

**Council:** `Council-Submitted: 7a3c4fb7-e8c1-4b5f-950e-7a826d5bebbe` (verdict pending at the
time of writing — read it, and act on a REVISE/REJECTED; the code is already on the shared
branch, so there is nothing to hold back).

**Two pattern-check advisories, both checked and both non-issues** — recorded so the next
reader does not re-walk them: (a) *untouched-twin `LoopAction`* — `LoopAction` is the loop
setup/expansion action and has no aggregation path at all (no `iterationResults`, no
`substep_output_fields`, no Strategy scan), so it cannot carry this defect; (b)
*logged-model-output at `loop_actions.go:334`* — that line is the bare literal
`logger.Info("Starting loop completion")`, no model output; the checker's line anchor shifted
by the inserted early return.

**Noticed while there, NOT fixed, NOT this bug:** `LoopAction` reads `max_iterations` with a
bare `.(float64)` assertion, so an `int`-typed declaration would silently fall back to the
default 20. Inert today because live config arrives from JSONB as float64 — the same class as
`bugs_open/193`'s silent bool reads, worth a sibling fix by whoever converges that.

### What is still owed on 289 after the roll

1. **Verify at the artefact, with the demand control.** Re-run one `tool-auditor` audit and read
   the series: `SELECT key, length(value::text) FROM orchestration_states, jsonb_each(collected_data)
   WHERE orchestration_id='<new run>' AND key LIKE '%\_done' ORDER BY key;` —
   **the disconfirming result is a ratio near 2.0 between consecutive laps.** This proves nothing
   unless a loop with more than one iteration actually ran, so check
   `loop_metadata.total_iterations > 1` on the run you read. The durable pass is a `tool-auditor`
   orchestration reaching `COMPLETED`; the population had 1 in 63 before this.
2. **Sweep the corpse rows.** 30-plus orchestrations are already dead at 22 MB and this fix does
   not move them. They are still being re-dispatched by the claimed-item-timeout sweep, so they
   are still costing Sonnet audits until their items go `failed`.
3. **The two residuals** — the `total_iterations` fallback, and the absence of any size tripwire
   on `collected_data`. A 22 MB orchestration is still silent.
4. **The separate `complete_workflow` defect** noted at the end of the main file
   (`message validation failed`, 6 rows) — `ProduceWithValidation` validates headers, not size.
   Still unfiled.


> **COUNCIL ROUND BLOCKED 2026-08-17 11:08Z — read this before re-submitting.** Round
> `7a3c4fb7-e8c1-4b5f-950e-7a826d5bebbe` died at `complete_invalid`, step `review_constitution`,
> and **the cause is not the submission**: `"You have reached your specified API usage limits.
> You will regain access on 2026-09-01 at 00:00 UTC."` The last successful LLM call fleet-wide
> was 11:08:03Z; failures since span two agent types and two models, so it is account-level.
> ~99.6% of fleet LLM work is Anthropic (24 h: 478 calls vs 2 on `mistral-small3.1`), so every
> LLM-driven agent is stopped, not just the gate.
> **Do not re-submit until the quota returns or is raised** — a resubmit cannot render a verdict
> and burns nothing but time. An invalid run writes **no artifacts**, so polling
> `diagnosis_artifacts` on that correlation waits for ever; poll
> `orchestration_states.current_step` instead. The `Council-Submitted:` trailer on `509e01e6a`
> asserts nothing and stays honest; `098` credits it automatically if the correlation is ever
> approved.

> **CORRECTED 2026-08-17 16:40Z — I WAS WRONG ABOUT THIS, AND IT BLOCKED NOTHING.** The outage
> lasted **about three minutes**, not fifteen days: `llm_call_log` shows the last failure at
> **11:09:53Z** (4 failed calls in 76 s, from 11:08:37Z) and successes resuming at **11:13:02Z**,
> including `council-gate` itself at 11:13:22Z. I measured inside that 3-minute window, read
> *"You will regain access on 2026-09-01"* off the API's 400 body, and reported a fleet-wide
> 15-day outage requiring an owner billing decision. **The date in that message is not
> predictive** — `bugs_open/243` (the same error, 2026-08-10) records the identical text against
> an outage that lasted 3 h 20 m and ended when the owner added credit, **21 days early**, and it
> says in terms: *"It did NOT auto-restore on 2026-09-01; the owner acted."* I never grepped for
> it. Round 2 of the council review was dispatched normally at 16:38Z on the same correlation.
> WRONG_CALLS 2026-08-17.


> **CORPSE SWEEP DONE 2026-08-17** — item 2 of "what is still owed" above is closed.
> 49 rows (`status='RUNNING' AND last_activity < NOW() - INTERVAL '4 hours'`, every one of them
> `tool-auditor` at `create_items_loop_complete` or `create_items_loop_iter_9_done`) set to
> `FAILED` with an error naming this bug. Verified after: **0** `RUNNING` rows fleet-wide and
> **0** topics pinned by them, down from 98. Ids saved to the lane scratchpad
> (`289_corpse_rows_before.txt`) — the rows were failed, not deleted, so it is reversible.
> **The reason they were immortal is a SEPARATE defect and is NOT fixed: `bugs_open/294`** (the
> reaper has no `RUNNING` arm and `TimeoutMonitor` keys on `awaited_requests`, which were empty).
> Do not read this sweep, or 289's fix, as closing 294.


> **RESIDUAL (5) CLOSED 2026-08-17 — the size tripwire is built** (register **WFA-016**, in the
> same commit). `UpdateStateWithVersion`, the one path every state update goes through, now logs
> when the `collected_data` it is about to write is abnormal: **WARN at 8 MiB, ERROR at 24 MiB**,
> naming the orchestration, agent type, step, total size, key count and **the largest single key**
> — the last of those because on this bug the total said "something is wrong" and the key name
> said what was wrong. Measured at the existing marshal, so it costs a length check.
> **It only logs.** A run headed for 29 MB still dies; you merely find out. Thresholds come from
> the 2026-08-17 census and a test pins them, so a later edit cannot slide them under real traffic.
> Wiring proven by mutation: unwiring the single call site leaves every direct-call test passing
> and fails only the sqlmock-driven wiring test.
> **Inert until the roll.** ~~NOT council-reviewed — the gate is unreachable until the Anthropic
> quota returns (2026-09-01).~~ **CORRECTED 2026-08-17: that was false — the outage was ~3 minutes
> (see the corrected block above), and the gate was available all along.** The tripwire was folded
> into round 2 of the SAME council correlation `7a3c4fb7-…` (dispatched 16:38Z) rather than its own
> round, because `bugs_open/244` measures `council-gate` at 87.8% of the fleet's August LLM spend.
> Note the residual honesty problem this leaves: commit `cf970b009` carries NO trailer at all and
> forward-only forbids an amend, so `098` cannot auto-credit it — the correlation is recorded here
> and in the lane docs instead.
> Residual (4), the `total_iterations` fallback, is still open and still latent.


---

## COUNCIL: APPROVED, round 2, corr `7a3c4fb7-e8c1-4b5f-950e-7a826d5bebbe` (2026-08-17 ~16:50Z)

**"approved with 4 advisory objection(s) — none high-severity."** 12 seats ran, 6 abstained.
Verdict read in full, not just its `decision` field. Every objection is answered below with a
measurement or a change — none was argued away.

**1. `reuse_agent` (MEDIUM, edit 2) — CORRECT, AND ACTED ON.** It said `loopConfigInt` was a
bespoke int-coercion helper and that the platform's own discipline prefers
`datahelpers.GetIntField`. It exists — `datahelpers/data_helpers.go:1586` — and handles exactly
the `float64`-then-`int` case my helper was written for. **Swapped, and the bespoke helper
deleted** (tests unchanged and still green). Its second, low objection asked whether an existing
"injected clone vs the loop's own end-step" signal already existed: none does, and the same seat
said so in its own notes ("no prior art for distinguishing an injected per-iteration
`loop_complete` clone from the loop's own end-step").

**2. `prior_art_librarian` (MEDIUM, edit 1) — MEASURED INDEPENDENTLY.** It flagged that
"no Go reader" was my own grep and that an index cannot prove a negative. Re-checked by
enumerating every `_iter_%d_` string-building site in `platform/ internal/ pkg/`:
`loop_error_handler.go:140` (builds an `_error` key), `:164` (next iteration's first substep —
routing), `loop_expansion_handler.go:101/140/556` (injection — writing), `coordinator.go:4417/4432/4455`
(next-step resolution — routing), `loop_actions.go:475` (the Strategy-3 prefix scan itself).
**Every one builds a STEP NAME for routing or an error key; none reads the terminal substep's
stored result.**
*But the seat's instinct found a real surface I had not named:* the OUTER end-step falls to
Strategy 3 when no substep declares `output_field`, and there its per-entry contents change
(markers instead of nested duplicates; the `{iterations, results[], count}` shape is unchanged).
**Exactly 1 of the 18 live loops is in that state — `component-quality-auditor.create_regen_items`
— and it has 0 orchestrations, ever.** Its `output_field` is `items_created`, `next_step` is
`complete`. So the change is confined to one never-run loop. Recorded rather than left implicit.
Its low objection — "no threshold existed anywhere" might overlook the sanitise loop —
checked: `SanitiseJSONBNulEscapes` (`datahelpers/jsonb_nul.go`) does no size work at all; its
only `len()` uses are buffer allocation and loop bounds.

**3. `debug_historian` (MEDIUM, edit 4) — ANSWERED WITH EVIDENCE, and it was a fair hit.** It
invoked this estate's own mutation-testing landmine: *a mutant that BREAKS THE BUILD prints the
same `FAIL` as a mutant that was caught*, and my submission asserted the mutation without saying
which. Re-run in three explicit steps: **(1) `go build ./platform/...` SUCCEEDS with the mutant**
— so the failure cannot be a compile error; (2) the test then fails as
`--- FAIL: TestLoopCompleteIterationTerminalDoesNotAggregate … iteration terminal copied the
PREVIOUS iteration's aggregate — this is the 2^N blow-up`; (3) restored, green. **That build step
is the part my original claim was missing.** Its low objection (no pod-verification step in the
plan) is met by "What is still owed" above and the loop-engine handoff, now cross-referenced —
the check is the label probe plus `git merge-base --is-ancestor`, not a tag or a deploy event.

**4. `guardian` (MEDIUM, edit 5) — ACCEPTED, and it is right.** Bundling the `state.go` tripwire
with the loop fix asked one reviewer to bless two blast radii as one. The cost argument
(`bugs_open/244`: `council-gate` is 87.8% of August LLM spend) is why, and the seat acknowledged
it, but that does not make the bundle good practice. **Mitigation that already existed: the two
are SEPARATE COMMITS** (`509e01e6a`, `cf970b009`) even though the review round was one, so
`git bisect` and revert are unaffected. **Standing lesson for this lane: bundle for cost only when
the changes share a blast radius — otherwise pay for the second round.** Its second objection
asked for a precedent check on this core site before treating it as clear; the round returned
approved, so that check did not find repeated deflection.

**5. `architecture` (LOW) — RECORDED AS A FOLLOW-UP, not done now.** The `loop_iteration`-presence
fallback in `isLoopIterationTerminal` becomes a permanent *second* discriminator once every
persisted plan carries the explicit flag. **New residual (6): once no in-flight plan predates
`509e01e6a`, delete the fallback so the discriminator is one thing again.** The seat's wider
observation is worth quoting because it is the RFC signal, not this fix: the step model overloads
`action` to mean both "what to run" and "role in the workflow", which is *why* an explicit flag
was needed; if more roles need disambiguating from the same action string, that recurring pattern
is the architecture-scope trigger.

**6. `constitution` (minor) — TIDIED.** It noticed the submission still listed "no size tripwire"
as an open residual while edit 5 added one — staleness carried from the original filing. Residual
(5) is marked closed above; residual (4) and the new (6) remain open.


---

## LIVE AND PROVEN AT THE ARTEFACT — 2026-08-17 18:2xZ, build `v1.0.1307`

**The fix shipped and the doubling is gone.** Proven on real traffic, not on an induced case.

**First: the deploy is real, checked at the artefact rather than at the tag** (the previous two
"fresh build" reports had shipped nothing — see the same-tag LANDMINE). Pods
`agent-chassis-6d6d7b9996-*`, image `v1.0.1307`, started 17:05Z, digest
`sha256:8339bdbd…`, whose own OCI label reads revision **`a6d1c53c0`**, built 2026-08-17 16:50Z.
`git merge-base --is-ancestor` against that revision: **`509e01e6a` SHIPPED, `cf970b009` SHIPPED**.
(`a436d898f`, the council's `GetIntField` reuse swap, is NOT in this build — it postdates it and
is behaviourally identical, so it rides the next one.)

**The measurement, on `build-dispatch-loop` — the fleet dispatcher, and the loop with the most
traffic.** Chosen deliberately over inducing a `tool-auditor` run: it exercises the same shared
engine with genuine demand, so the reading needs no artificial case. Per-lap `_done` sizes on
post-roll runs:

| | before (measured 08-17 am) | after |
|---|---|---|
| `process_item_iter_0_done` | 201 kB | **77 B** |
| `_iter_1_done` | 617 kB | **77 B** |
| `_iter_2_done` | 1,419 kB | — |
| `_iter_3_done` | 3,023 kB | — |
| `_iter_4_done` | 6,247 kB | — |
| ratio per lap | **~2.1–3.1x** | **1.00** |

**The disconfirming result was a ratio near 2.0. It is 1.00 — the series is flat.** The stored
value is exactly the designed marker:
`{"status": "iteration_complete", "iteration": 0, "loop_name": "process_item"}`.

**Demand control, because a flat series proves nothing if no multi-lap loop ran:** every run read
carries `loop_metadata.total_iterations = 5`, so five-lap loops genuinely executed. Like-for-like
totals over multi-lap runs only:

| era | runs | avg `collected_data` | max |
|---|---|---|---|
| before the roll | 156 | 2,575 kB | 14 MB |
| **after the roll** | 10 | **104 kB** | **229 kB** |

**~25x smaller on average, ~61x on peak, on the service that dispatches every piece of work on
the estate.**

**WFA-016 (the tripwire) is live and silent — and that is the expected reading, not a failure.**
0 `collected_data is unusually large` lines in the last hour, because after this fix nothing
persists an oversized state. Per WFA-016's own verify-later note, a zero here means "no oversized
state", NOT "the instrument is broken"; the only honest way to exercise it now is a fixture, since
the producer it was written for has stopped.

### Still owed before this can close

**The motivating case itself is NOT yet proven.** No `tool-auditor` run has executed since the
roll — its queue holds no runnable item (28 failed, 20 complete, 16 unresolved, 10 deferred), so
the zero is absence of demand, not evidence. The bug's own stated pass is "a `tool-auditor`
orchestration reaching `COMPLETED`" (the population held **1 in 63**). One bounded probe has been
filed to settle it: item `12836a25-8266-46fb-bba1-2e8635ef9cc0`, page-pinned to the Mind Map,
`created_by='bugfix-289-postroll-verification'`, **`max_attempts=1` so it costs at most ONE Sonnet
audit** rather than the three the old exhausted probe paid. Read its outcome, then close.

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
  size, so that is its own thing. Worth a separate file.

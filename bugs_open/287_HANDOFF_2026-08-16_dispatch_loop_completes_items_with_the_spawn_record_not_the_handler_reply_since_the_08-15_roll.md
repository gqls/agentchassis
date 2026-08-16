# 287 — build-dispatch-loop marks work items `complete` with the SPAWN RECORD as their result, not the handler's reply — appears at the 2026-08-15 10:14Z roll, ~75% of completions since

> **⚠ NUMBER COLLISION — two bugs are 287 (2026-08-16, concurrent lanes).** This one is the
> **`spawn_record`** slug (dispatch-loop item results). The other is
> `287_…an_agent_seeded_without_a_description_is_unspawnable…` (agent seeding), filed the same
> morning by another lane. Neither is renumbered — forward-only, and both are already cited by
> commit. **Resolve by SLUG, `git log` the FILE PATH, never the number** (CLAUDE.md, Debugging).
> Commits `ae5d12048 93c720960 5953eaf76 e566f36b1 7bb500934` are this file's.

**Filed 2026-08-16** by the mortgagecalculator adoption lane, found while verifying that
`bugs_closed/274`'s fix (delivery of a child's reply to its parent) is in the live chassis. **It is,
and it works** — zero cannot-deliver rows against 859 child completions since the v1.0.1303 roll.
This is the *next* defect on the same seam, and it is a **regression carried by the same roll**.

## 1. The shape

A work item routed by `build-dispatch-loop` is marked `complete` and its `result` column holds:

```json
{"role":"handler","topics":{"requests":"job.<corr>-<x>-<handler>-process_item_iter_N_spawn_handler.requests",…},
 "agent_id":"…","agent_type":"<handler>"}
```

That is the **spawn record** — `spawn_agent`'s output (`extractSpawnData`, stored at
`process_item_iter_N_spawn_handler` and at output_field `handler_spawned`). It is NOT the
handler's reply, which `mark_complete` is configured to record:
`complete_work_item` with `"result": "handler_result"` (live `build-dispatch-loop` config,
`process_item.sub_workflow.steps.mark_complete`; `call_handler` is `call_agent` with
`output_field: handler_result`).

The child does its work and COMPLETES; the reply DOES arrive; the parent's own persisted
`collected_data.handler_result` DOES end up holding the correct `{"response":{…}}`. Only the
item's stored `result` is wrong. So: **the work is done, the record of it is someone else's.**

## 2. Scale [MEASURED 2026-08-16 ~10:10Z]

```sql
SELECT date_trunc('hour', updated_at),
       count(*) FILTER (WHERE result ? 'response')                        AS own_envelope,
       count(*) FILTER (WHERE result ? 'topics' AND result ? 'agent_id')  AS spawn_record,
       count(*) AS total
FROM site_work_items WHERE status='complete' AND updated_at > '2026-08-15 04:00' AND handler_agent IS NOT NULL
GROUP BY 1 ORDER BY 1;
```

| window | own envelope | spawn record |
|---|---|---|
| 08-15 08:00–10:13 (pre-roll) | 136 | **0** |
| 08-15 10:14–11:00 | 8 | 5 |
| 08-15 12:00 | 5 | 26 |
| 08-15 14:00–16:00 | 55 | 168 |
| 08-15 19:00 (post-1303 roll) | 22 | 120 |
| 08-16 08:00 | 3 | 3 (+27 "other") |

**Zero instances in any hour before the 10:14Z roll; the shape appears in the roll's first
hour and dominates thereafter** (~270 vs 70 since 18:46Z). By handler since 18:46Z:
page-rerender 221, page-build-handler 19, image-build-handler 8, component-template-fixer 7,
tool-improver 4, color-variable-fixer 4, content-gap-planner 4, section-editor 4.
The pre-roll "other" bucket (42 at 08:00) is a different, older shape — not this.

## 3. A worked instance, timeline intact (`c0aee25f…`, page-rerender, gaswholesalers)

| t (UTC) | event |
|---|---|
| 08:39:56 | child `page-rerender` created (parent `build-dispatch-loop`) |
| 08:40:05 | child COMPLETED |
| **08:40:06** | item `c773ee3d` marked `complete` — result = spawn record |
| 08:40:46 | parent COMPLETED; its `collected_data.handler_result` = `{"response":{"deploy_result":{"success":true,…}}}` and `process_item_iter_1_call_handler` = same |

So `mark_complete` ran ONE second after the child completed and read `handler_result` as the
spawn record; the parent's final state shows the reply merged at `handler_result` — **after**, or
onto a copy `mark_complete` did not read. `call_agent`'s own pre-park return (`buildCallResult`,
`call_agent.go:647`) has keys `agent_called/agent_type/request_id/child_orchestration/…` — NOT
`role/topics/agent_id` — so what the item holds is not call_agent's placeholder either; it is
the **previous step's** output, which is what `resolveFieldValue`'s fallback search finds when
the asked-for key is absent (`[INFERRED — the resolution at mark_complete time is the thing
the 090 run is asked to read]`).

## 4. Why "the same roll" — what changed on this seam that day

Two commits rode the 10:14Z roll on exactly this path:
- `919cc6976` (274 fix) — reply envelope headers, so the child's reply now VALIDATES and is
  delivered. Before it, the reply never arrived, the parent was told the child FAILED, and
  the item was completed with whatever `handler_result` resolved to (that produced the
  08-11/08-14 §D shape: a *content-planner* payload). **So the pre-roll `own_envelope`
  rows were NOT reply-driven** — with the reply never validating, they must have come
  from another path `[UNMEASURED which]`.
- `3ba384c63` (WFA-014, RFC_012 (a)) — `persistAwaitingStateWithRetry` now CARRIES the
  dispatching step's in-memory `collected_data` onto the fresh DB load at park
  (`carryCollectedDataOntoFreshState`, coordinator.go:2169). Before it, the park
  discarded them.

The register entry for WFA-014 already names the residual it does not cover: *"this fixes
the PARK only — `applyResponseToState` still REPLACES wholesale on its `output_mapping` and
default branches, so a carried key can still die at reply time."* The `call_agent` branch is
the additive `isAgentResponse` one (coordinator.go ~2820-2855), so wholesale-replace is not
this — but the ORDER of (reply merged onto which state) vs (next step `mark_complete` reads
which state) is the open question. `[HYPOTHESIS, filed to the loop, not asserted]`: the
resume-from-park path loads a fresh state, merges the reply, and continues — while
`mark_complete`'s read of `handler_result` happens against a state where the key holds the
carried pre-reply value or is absent, so the recursive fallback lands on `handler_spawned`.

## 5. Consequence

- Every completion-time reader of `site_work_items.result` for these ~270 items sees a spawn
  record: verifiers (`bugs_closed/213`'s gate reads the handler's self-report — 4-of-14
  became ~0-of-N), retraction logic, operators, and the two-strike rule (a `complete` is a
  strike whatever it holds).
- The WORK is fine — children complete, artefacts land (verified: mortgagecalculator's logo
  redeploy `undeployed_asset:e766370e` at 19:12Z is one of these rows; `logo.png` was and is
  200). So this is a records defect, not a delivery defect — the mirror image of 274.

## 6. Diagnosis status

`090` filed 2026-08-16 ~10:15Z: intake `de3b436a-d8da-4c4c-a8fe-8b3abcb0d461`, run
`fb7ae3bc-e9bf-4a96-b540-d593b91bc79c`. **⚠ MEASURED after filing: `origin/087` is at
`a85ad4018` (2026-08-12 17:01), 881 commits behind local HEAD, and `merge-base --is-ancestor`
says it carries NEITHER `3ba384c63` (WFA-014) NOR `919cc6976` (274).** The loop reads origin.
So this run is diagnosing a coordinator that does not contain the code the roll introduced —
its verdict describes the pre-regression binary. **Whatever it says, it cannot CONFIRM this
mechanism; a REFUTED from it is a refutation of the wrong tree, not of the claim.** The
correct next step is a push (owner's — 881 commits is not one lane's call) and a re-run, or a
first-hand read of `coordinator.go` at HEAD declared per the 2026-07-31 substitution rule.
Recorded here BEFORE the verdict landed so it cannot be read straight. Grep-before-file: `bugs_open/236` recorded the park
discard (fixed by WFA-014); `bugs_closed/274` the delivery half; `bugs_closed/213` §D the
foreign-payload symptom. None names this post-roll shape.

## 7. Fix candidates (ordered by what closes the door — NOT yet reviewed, do not act on §7 alone)

1. **Make `mark_complete` read the reply, not a path that can resolve elsewhere**: have
   `complete_work_item` refuse to record a `result` that lacks the `response` envelope when
   the preceding step was a `call_agent` (or read `process_item_iter_N_call_handler.response`
   directly, which the parent row shows IS correct at the end). Closes the door for this
   loop; does not fix the ordering.
2. **Fix the ordering at the seam**: ensure the reply merge lands on the same state the next
   step reads before that step executes — RFC_012 territory (every await-using pipeline in
   the blast radius); route to that decision with §3's timeline attached.
3. **Kill the fallback**: `resolveFieldValue`'s recursive search returning a sibling step's
   data for a missing key is the same landmine `bugs_open/248`'s R1 objection named
   (`findFieldRecursive`); a missing `handler_result` should be an error at
   `complete_work_item`, not a silent substitution.

## 8. How to verify a fix

Post-roll, `SELECT count(*) FILTER (WHERE result ? 'topics' AND result ? 'agent_id')` over
`status='complete' AND updated_at > <roll>` must be 0 **while** `count(*) FILTER (WHERE result
? 'response')` is non-zero in the same window (demand). And one item followed end to end:
child COMPLETED → item `complete` with `result.response` = the child's payload.

---

## 9. FIRST-HAND READ at local HEAD, 2026-08-16 ~10:25Z — declared substitution (2026-07-31 ruling), because the 090 run reads a tree 881 commits stale (§6)

Read: `coordinator.go` `handleCompleteResponse` (:2597-2680) → `applyResponseToState`
(:2749-2880); `loop_expansion_handler.go` `makeIterationOutputField` (:222),
`propagateIterationOutputs` (:426-500) and its single call site in `setLoopVariable` (:374).
Then re-checked the live parent `c0aee25f…` for the keys the read predicts.

**The mechanism (three facts, each verified):**

1. **The reply is written to the SUFFIXED key.** Loop expansion rewrites each sub-step's
   `output_field` to `<field>_<iter>` (`makeIterationOutputField`), so `call_handler`'s reply
   lands via `applyResponseToState`'s additive `isAgentResponse` branch at
   `process_item_iter_N_call_handler` and at **`handler_result_N`** — the parent row shows
   `handler_result_0` and `handler_result_1` each holding the correct `{"response":…}`.
   The reply path itself is fine (fresh load → apply → `continueExecution` on the same state).
2. **`mark_complete` reads the UN-suffixed key.** Its config is `"result": "handler_result"`;
   `prefixConfigStepReferences` (:147) rewrites step-name references, and the base
   `handler_result` is only populated by `propagateIterationOutputs`, which copies
   `<field>_<iter> → <field>` **once, at the START of each iteration** (`setLoopVariable`,
   :374) — i.e. before that iteration's `call_handler` has run. So at iteration N's
   `mark_complete`, `handler_result` is either ABSENT (iteration 0) or holds iteration N-1's
   value. Live: both items in run `c0aee25f` hold the spawn record for their OWN iteration
   (topic `…iter_0…` on the 08:39:21 item, `…iter_1…` on the 08:40:06 item), so it is not the
   N-1 case either — the resolver did not find `handler_result` and **fell back**.
3. **The fallback lands on `handler_spawned`.** `resolveFieldValue`'s recursive search for a
   missing key (the `findFieldRecursive` landmine, `bugs_open/248` R1) returns the nearest
   value — `spawn_agent`'s `extractSpawnData` output at `handler_spawned`/`handler_spawned_N`,
   whose shape `{role,topics,agent_id,agent_type}` is exactly the item's stored `result`.
   `[INFERRED which key the search hits first; the SHAPE match and the per-iteration topic
   match are measured]`.

**Why it appeared WITH the roll and not before — a correction to §2 and §4:**
Pre-roll, the child's reply never validated (274), so `handler_result_N` was never written by
a reply. The pre-roll "own envelope" rows this file's §2 counted as CORRECT are **not**: they
hold `response.deploy_result.response.deploy_result…` — a doubly-nested envelope, which is
`bugs_open/216`'s duplicate-execution shape (reply refused → child re-run → both wrapped), so
they were 274's symptom too, and `handler_result` was reached through the failure/timeout
path `[UNMEASURED which]`. Post-roll the reply is delivered ONCE, correctly, to the suffixed
key — and the un-suffixed read that `mark_complete` always did now finds nothing at all,
because nothing else writes it. **The regression is real, but the seam was always wrong: 274
was hiding it by making a different wrong thing land in the same slot.**

> **CORRECTED 2026-08-16 (§2, §4):** the pre-roll `own_envelope` column is not a "correct"
> baseline; it is 216/274-shaped. The correct pre-roll count of properly-recorded results is
> unmeasured and probably ~0 for loop-dispatched items. What the roll changed is WHICH wrong
> payload lands, not whether one does.

**What the 70 post-roll `own envelope` rows are** (`tool-acceptance-agent`, some
`page-rerender`): `[UNMEASURED]` — likely single-iteration runs where propagation or a
non-loop path populates the base key, or agents whose `mark_complete` reads a suffixed key.
Worth one query by whoever fixes this; not needed to fix it.

**Fix candidate that follows from the read (sharpening §7):** `mark_complete`'s `result`
must reference the CURRENT iteration's reply — either (a) `prefixConfigStepReferences`
should also rewrite `output_field` references in `config` to their suffixed form (it rewrites
step names; it does not rewrite `handler_result` → `handler_result_N`), or (b) run
`propagateIterationOutputs` after each substep, not only at iteration start, or (c) have
`complete_work_item` read `<prev step>.response` explicitly. (a) is the one that closes the
door for every loop-dispatched agent, not just this one, and it is where the seam is. Any of
them must ALSO make a missing key an error rather than a substitution (§7.3), or the next
mis-spelling lands the same way. Blast radius: every `loop` sub-workflow whose later substep
reads an earlier substep's un-suffixed `output_field` — a `SELECT` over `agent_definitions`
for `sub_workflow` steps whose config strings match a sibling's `output_field` is the census.

### 9a. The exact door — one allow-list, one missing key [READ 2026-08-16 ~10:30Z, `coordinator.go:4443-4500`]

`prefixConfigStepReferences` rewrites a sub-step's config so that references to sibling
outputs become iteration-suffixed. It does so for an explicit **allow-list**:

```go
dataRefKeys := []string{"content_from","context_from","data_from","source_field","input_from",
                        "result_from","content_field","page_component_id_field"}
// ...plus every value in config["input_mapping"]
```

`complete_work_item`'s config key is **`result`** (`"result": "handler_result"`, live
`build-dispatch-loop` `mark_complete`). **`result` is not on the list.** So `call_handler`'s
`input_mapping` values ARE suffixed (that is why the handler receives the right item), and the
`mark_complete` read is NOT — it asks for `handler_result` while the reply sits at
`handler_result_0`. The comment above the list says it itself: *"IMPORTANT: Any config key
that references step outputs must be listed here."* It was not.

Same defect class as `bugs_open/149`'s "widening what REACHES a function breaks it unedited"
in reverse: an allow-list of config keys is a promise every future action must know to keep,
and `complete_work_item` did not. **Fix that closes the door: stop enumerating.** Suffix ANY
string-valued config entry whose first dotted segment is a sibling `output_field` (the
`substepOutputFields` set is already computed and passed in) — then no action can be left
off. Interim one-liner: add `"result"` to `dataRefKeys`. Either way, a missing key at
`complete_work_item` must be an ERROR (§7.3), because the recursive fallback is what turned
"absent" into "someone else's record" here and in 248 before it.

**Test that would have caught it, and pins the fix:** expand a loop whose sub-workflow has a
`call_agent` (`output_field: X`) followed by ANY step whose config carries `"result": "X"`
(or any key), and assert the expanded config reads `X_0`. Mutation: remove the suffixing and
watch it fail.

**Why the pre-roll world never noticed:** the reply never validated (274), so
`handler_result_N` was never written by a reply either; the un-suffixed key was found via the
timeout/failure path (`[UNMEASURED which]`), so `mark_complete` "worked" — by reading a
different wrong thing. Fixing 274 exposed the allow-list gap. This is the file to cite when
someone asks why a correct fix made a metric worse.

### 9b. Blast-radius CENSUS [MEASURED 2026-08-16 ~10:35Z, live `agent_definitions`]

Every active loop sub-workflow whose sub-step has a string config value (outside `input_mapping`
and the step-ref/allow-list keys) whose first dotted segment equals a sibling `output_field`:

| agent | loop | substep | action | config key | names |
|---|---|---|---|---|---|
| **build-dispatch-loop** | process_item | **mark_complete** | complete_work_item | **result** | handler_result |
| build-dispatch-loop | process_item | check_claim | conditional | condition | claim_result |
| business-intel | process_batch | check_match | conditional | condition | ch_search |
| business-intel | process_batch | fetch_details | companies_house_fetch | company_number_field | ch_search |
| maintenance-triage | rebuild_loop | mark_dispatch_complete | mark_maintenance_complete | result_field | rebuild_result |
| pageflow-builder | build_pages_loop | check_review_approved | conditional | condition | reviewed_content |
| pageflow-builder | build_pages_loop | save_sections | save_page_sections | sections_metadata_field / html_field | page_content / assembled_page |
| page-rebuild | build_pages_loop | (same two) | | | |
| site-work-orchestrator | build_items_loop | check_review_approved | conditional | condition | reviewed_content |
| site-work-orchestrator | build_items_loop | complete_work_item | complete_work_item | commit_sha | page_deployed |
| site-work-orchestrator | build_items_loop | save_sections | save_page_sections | (same two) | |

15 sites, 6 agents. **Not all are live defects — but where they are safe, they are safe by
coincidence, not by the mechanism:** `page_content`, `assembled_page`, `reviewed_content`,
`page_deployed` are in `propagateIterationOutputs`' hardcoded `commonOutputFields` list, so the
base key is pre-populated at iteration start — from the PREVIOUS iteration on iteration ≥1
(the very staleness the propagation comment warns about) and from nothing on iteration 0.
`claim_result` / `ch_search` / `rebuild_result` are non-awaiting actions whose result is written
synchronously before the reader runs, so `resolveFieldValue` may find them via its own
strategies. **`handler_result` is the one that is BOTH awaited (so written only at reply time,
suffixed) AND not in the propagation list** — which is why it is the one that fires. The
`site-work-orchestrator` `commit_sha: page_deployed` read is the next-nearest shape
(`deploy_page` may await) — `[UNMEASURED]` whether it currently records the right sha.

The census query is in the commit that added this section; re-run it before fixing, and make
the fix generic (§9a) so the table cannot grow back.

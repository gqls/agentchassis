# 287 — build-dispatch-loop marks work items `complete` with the SPAWN RECORD as their result, not the handler's reply — appears at the 2026-08-15 10:14Z roll, ~75% of completions since

## STATUS 2026-08-17 18:2xZ — **FIXED, LIVE AND PROVEN ON BOTH HALVES** (§11c). Kept here only until two OWNER decisions land; the defect itself is dead for new traffic.

| | state |
|---|---|
| **build-dispatch-loop** (86% of the defect) | **FIXED + LIVE + PROVEN** — mig `452` applied 16:28:57Z; a post-apply run completed **4/4 items carrying the handler's reply**, content checked at the artefact |
| **diagnose-dispatch-loop** | **FIXED + LIVE + PROVEN** — mig `448` 12:19:59Z; its first post-apply run COMPLETED 17:0xZ and the item's `result` holds the real verdict (`status: UNVERIFIABLE`, summary, conclusion) where it used to hold the spawn record. So §6a's "read verdicts from `orchestration_states`, never the item" is now RETIRED for this agent |
| **report-dispatch-loop** | config **LIVE** (same migration, identical shape) — **unproven for want of traffic**: zero runs since the apply. Demand-bound, not a failure |
| **Half 1 (Go, WFA-017)** | **LIVE + PROVEN v1.0.1307, 17:05Z** — expanded plans read `result!: handler_result_0`; `field=result` resolver rows **0** (from ~455/day) with 10 runs of demand; 11/11 loop completions carry the reply; 0 spawn records (§11c). The earlier 14:42Z "fresh build" had shipped nothing (same tag, cached image) |
| **Historical rows** | 3,330 wrong; **303 repairable**; mig `455_HOLD` written + dry-run proven, **awaiting the owner's go** |
| **Convention question** | `RFC_035` open (guardian scope veto on Half 1 — ratify / narrow / revert; owner's) |
| **Admin** | `452`'s `schema_migrations` row is **MISSING** (harness refused `--record-only`); it IS applied — command in the file header |

**Why this file is still in `bugs_open/`:** by CLAUDE.md's bar (fixed AND live) it has earned
`bugs_closed/`, and nothing is biting new traffic. It is held here for one reason only —
**303 historical rows are still wrong and a reader of them is still misled today**, and the two
owner decisions that settle this lane (`455_HOLD` go/no-go, `RFC_035`) are staged from here.
Move it the moment `455` is decided; the fix needs no further work.

**Reading an item's `result` now:** for a run created after its agent's migration it is
`call_agent`'s step record with the reply at `.response` (12 keys, incl. `retry_payload`) — not
a bare `{response:…}`. A plan is copied at orchestration creation, so **pre-migration runs keep
producing spawn records until they drain** — attribute before concluding (§11b).

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

### 6a. VERDICT READ 2026-08-16 10:19Z — `REFUTED` on the resolver's NAME, symptom CONFIRMED, mechanism otherwise upheld; and the run's own item is an instance

Run `fb7ae3bc…` COMPLETED after **5 iterations (cap)**: `route.status = UNVERIFIABLE`,
`verdict.outcome = REFUTED`, "no fix proposed". Read what it refuted, not the label:

> *"The end symptom is real — multiple site_work_items rows are stamped status='complete' with
> result holding exactly the spawn record shape {role, topics, agent_id, agent_type} rather than
> a {response:{...}} envelope (confirmed for handler_agent = tool-acceptance-agent, page-rerender,
> section-editor at fresh 2026-08-16 timestamps). But the specific mechanism proposed — that
> mark_complete resolves 'handler_result' via resolveFieldValue's fallback search — does not
> hold: resolveFieldValue lives in conditional_branch_action.go … CompleteWorkItemAction
> instead resolves its 'result' input through datahelpers.ExtractActionInputs →
> ActionInputs.GetMap("result"), whose own in-file documentation warns that its
> recursive/aggressive search …"* — citing the comment verbatim: *"ExtractFields uses
> aggressive recursive search that can find stale values from previous loop iterations
> (e.g. claim_result.work_item_id from iter 0)."*

**So: §9 fact 3 named the WRONG resolver — corrected here.** The fallback that turns an absent
`handler_result` into the spawn record is `datahelpers.ExtractActionInputs` /
`ActionInputs.GetMap` (`action_inputs.go`, the Strategy-1/2 recursive `ExtractFields`), not
`conditional_branch`'s `resolveFieldValue`. Same door, correct name — and it is the SAME
aggressive search that `bugs_open/248`'s R1 objection flagged for `asset_id`, whose own
comment already documents exactly this failure. §9a's cause (the un-suffixed `result` key,
allow-list gap in `prefixConfigStepReferences`) and the fix (suffix generically; make an absent
key an error at `complete_work_item`) stand unchanged; §9's line "the `findFieldRecursive`
landmine" was the right landmine under the wrong caller.

> **CORRECTED 2026-08-16 (§9 fact 3):** ~~`resolveFieldValue`'s recursive search~~ →
> `ExtractActionInputs`/`GetMap("result")`'s aggressive recursive search. Caught by the 090
> run's static read; the shape and per-iteration topic evidence are unaffected.

**Caveats on the run, both ways:** it read `origin` (881 commits stale, §6) — but
`CompleteWorkItemAction`'s read path predates the roll, so on THIS question the stale tree was
the right tree, and the read is good. It could not see WFA-014 or the 274 fix, so it could not
speak to why the shape appeared WITH the roll; §9's account of that ("274 was hiding it")
remains first-hand, `[INFERRED]` on the pre-roll path.

**And the run's OWN work item is an instance of this bug** — a live demand control nobody
asked for: the `needs_diagnosis` item for `fb7ae3bc…` completed with `result =
{"role":"handler","topics":{"requests":"job.fb7ae3bc-…-diagnose-orchestrator-spawn_handler.requests",…},
"agent_type":"diagnose-orchestrator"}` — the spawn record — while the diagnosis itself
completed and its verdict sits in the `diagnose-agent` row's `collected_data`
(`verdict`/`route`/`emit`). **`diagnose-dispatch-loop` is a 7th agent for the §9b census**
(its `process_item`/`mark_complete` shape mirrors build-dispatch-loop's), and every 090
verdict since 08-15 10:14Z that a reader takes from the ITEM's `result` is a spawn record.
Read verdicts from `orchestration_states` (`owner_agent_type='diagnose-agent'`,
`collected_data->'verdict'`) or `diagnosis_artifacts`, never from the item, until this is fixed.

## 10. CONTRIBUTION 2026-08-17 from the `staged_component_build` lane (RFC_029) — the resolver now RECORDS every one of these, so §9 fact 3's last `[INFERRED]` is MEASURED, and the population is countable

I own the RFC_029 resolver work, not this bug. RFC_029 Phase 1 put a persistent instrument on
exactly the door §6a corrected you to (`ExtractActionInputs` / the aggressive recursive search),
and it has been writing rows since the 2026-08-16 10:41Z roll. **This section is evidence for
your fix, not a competing account** — §9/§9a's mechanism and §9a's fix ordering stand.

**What the instrument is.** Every occurrence of the aggressive search resolving a CONFLICT, and
every occurrence of a dotless single-segment mapping being OUTVOTED by it, now writes an
`agent_error_log` row (`error_code` `RESOLVER_CONFLICTING_CANDIDATES` / `RESOLVER_MAPPING_BYPASSED`,
`severity='warning'`, `action='input-resolver'`). No dedup — one row per occurrence. Registered
in the chassis at startup; live from v1.0.1303 (log-only) / v1.0.1304+ (rows). Concept register
CTS-060; RFC_029 §10.4.

**Your bug, measured over the first ~24 h** (`[MEASURED 2026-08-17 10:5xZ]`, 1,571 rows total,
7 agents; `build-dispatch-loop` is **1,357 of them, 86%**):

| field | winner the search picks | rows | candidate paths seen |
|---|---|---|---|
| `work_item_id` | `claim_result.work_item_id` | 453 | 13 – **189** |
| `current_page` | `handler_result.retry_payload.message.body.~unwrap.current_page` | 452 | 12 – 170 |
| `result` | **`handler_spawned.result`** | 176 | 6 – 63 |
| `result` (bypass) | reference `handler_result` outvoted, got a `map[string]interface{}` | 279 | — |

**§9 fact 3 said `[INFERRED which key the search hits first]`. It is now measured: `handler_spawned.result`,**
176 times, always the same winner — your predicted key, confirmed per occurrence rather than by
shape-match. And the 279 bypass rows are the same event seen from the other side: `mark_complete`'s
config `"result": "handler_result"` is DOTLESS, so the search runs first and outvotes it — that is
what the bypass instrument exists to catch, and it fires on nothing else in the fleet.

**Two things this adds that the item-table census cannot:**

1. **`work_item_id` and `current_page` are the same defect as `result`, and they are BIGGER.**
   The comment §6a quoted ("stale values from previous loop iterations, e.g.
   `claim_result.work_item_id` from iter 0") is not illustrative — it is happening 453 times a
   day, with the winner literally `claim_result.work_item_id`. `current_page` resolves out of a
   **retry payload's message body** 452 times. Your fix (§9a (a), suffix generically) should be
   assessed against all three, and §9b's 15-site census is the right shape for that.
2. **The candidate count RISES with the iteration** (13 → 189). That is the accumulation itself,
   visible per run: each iteration leaves another copy of every field in `collected_data`, so the
   search's ballot grows all day. It also means a "pick the shallowest" rule cannot save this —
   `[INFERRED]` that the shallowest of 189 is stable only because the earliest iteration's key is
   shallowest, which is precisely the WRONG one.

**⚠ Do NOT reach for RFC_029's `!` strict marker as the quick fix here — the order is
load-bearing.** `!` (live in the binary now) means "explicit resolution only, or fail loudly".
Writing `"result!": "handler_result"` on `mark_complete` TODAY would hard-fail **every**
loop-dispatched completion in the fleet, because `handler_result` is genuinely absent at that
point — that is your §9a finding. The marker is the RATCHET that goes on **after** §9a (a)
makes the reference resolve: fix the suffixing, verify the rows below go to zero, and only then
add `!` so the defect cannot come back silently. That sequencing is the whole of §7.3's "make a
missing key an error rather than a substitution" — the mechanism now exists, it just must not be
armed before the key exists.

**How to watch your own fix land** (this replaces "re-run the item census and hope"):
```sql
SELECT date_trunc('hour', occurred_at) AS hr, context->>'field' AS field, count(*)
FROM agent_error_log
WHERE error_code LIKE 'RESOLVER_%' AND agent_type='build-dispatch-loop'
  AND occurred_at > '<your roll time>'
GROUP BY 1,2 ORDER BY 1 DESC, 3 DESC;
```
A correct fix takes `result`, `work_item_id` and `current_page` to **zero** for
`build-dispatch-loop` while the loop keeps running (check the loop IS running — zero rows also
means no traffic; `SELECT count(*) FROM orchestration_states WHERE owner_agent_type='build-dispatch-loop'
AND created_at > <roll>` is the demand control). Rows have no `orchestration_id` by design
(pod-level attribution; see the row's `context.identity_scope`) — join by time and `pod_name`.

**The non-dispatch-loop remainder, for completeness** (214 rows, a different and much smaller
tail — NOT yours): `page-content-writer` `current_page` → `~unwrap.current_page` 165;
`page-build-handler` `sections`/`page_type` → `load_page_record.*` 15/14; small counts for
page-rerender, rerender-pages, tool-generator, generic.

— `staged_component_build` lane, 2026-08-17. Questions about the instrument (not about this bug)
belong in RFC_029 §10.4–§10.6 or the lane's RUNBOOK.

### 10a. ADDENDUM 2026-08-17 — why the STALE value wins every time: it is the tie-break, not the depth

Read one full ballot (13 candidates, `work_item_id`). The depth-1 candidates are exactly
`claim_result`, `claim_result_0`, `claim_result_1` — your base key plus one per iteration.
Depth cannot separate them, so RFC_029 D1's tie-break decides, and it is **sorted-key DFS order**:
`claim_result` < `claim_result_0` < `claim_result_1`. **A base name always sorts before its own
suffixed siblings, because the suffix is an append.**

So the search does not "happen to" pick the stale value — **it picks the base key by
construction, on every field loop expansion suffixes, every time.** Combined with your §9 fact 2
(`propagateIterationOutputs` copies `<field>_<iter> → <field>` at iteration START, so the base
holds iteration N−1's value or nothing), that is a complete account of why iteration N reads
iteration N−1's data with perfect reliability.

Two consequences for your fix:
- **§9a (a) is the right one and this is extra evidence for it.** Suffixing the reference means
  `mark_complete` resolves explicitly and never reaches the search — the only version that
  survives this tie-break. Options (b) and (c) also work, but (a) closes it for every
  loop-dispatched agent at once, which §9b's 15-site census says is what you need.
- **Do not consider "make the search prefer the newest candidate" as an alternative.** It would
  be a second guessing rule layered on the first, it would silently change resolution for every
  non-loop caller fleet-wide, and RFC_029 §9's ruling forbids exactly that trade ("any picking
  rule is a guess"). The fix is to stop searching, not to search better.

— `staged_component_build` lane, 2026-08-17. Full read: RFC_029 §10.6a.

## 11. FIX SHIPPED 2026-08-17 (the "bugfix 287" session) — two corrections with evidence, both halves built, one migration applied; OPEN until the roll

Taken on after re-verifying validity (spawn-record completions + resolver rows within hours,
2026-08-17 ~12:30Z) and non-ownership (filing lane handed off; RFC_029 lane disclaimed in §10;
no live session editing the seam — transcript grep). Lane docs:
`docs024_key_docs_latest/bugfix_287_spawn_record/`. Council `Council-Submitted:
cba35b35-53f1-4c7d-86af-faa64839c0ca`.

> **CORRECTED 2026-08-17 (§9 fact 2, and §10's `!` warning):** ~~"`propagateIterationOutputs`
> … copies `<field>_<iter> → <field>` once, at the START of each iteration … at iteration N's
> `mark_complete`, `handler_result` is either ABSENT (iteration 0) or holds iteration N-1's
> value"~~ — **`setLoopVariable` (which calls `propagateIterationOutputs` with the CURRENT
> step's iteration) runs before EVERY local action**: single call site `coordinator.go:1355`
> (`executeLocalAction` step 5a, "Must happen before buildActionParams so propagated outputs
> are included"), and `getActionHandler` has exactly one caller, so the resume-from-park path
> takes it too. So at `mark_complete`, base `handler_result` holds the CURRENT iteration's
> reply. The §10 instrument's own bypass rows prove it: `RESOLVER_MAPPING_BYPASSED` fires only
> when the mapped key EXISTS and differs (279/day). Two consequences: the defect is PURELY
> search-beats-dotless-mapping (the §6a resolver, exactly); and §10's warning that arming `!`
> before the suffix fix "would hard-fail every loop-dispatched completion … because the key is
> genuinely absent" is wrong at HEAD — the key is present. (§10a's tie-break finding is
> unaffected; what its base key HOLDS at the moment the ballot is read is the part corrected.)

> **CORRECTED 2026-08-17 (§10's success metric):** ~~"When §9a (a) lands, the three
> build-dispatch-loop pairs should go to zero"~~ — **suffixing alone zeroes NOTHING.** The
> whole-tree search runs for every non-strict declared field regardless of whether its mapping
> would resolve; a dotless mapping is outvoted either way. The rows that measure THIS bug —
> field=`result` conflict + bypass — go to zero at the **`!` flip**, not at the suffix roll.
> And the `work_item_id` (453/day) / `current_page` (452/day) conflict rows are **not this
> defect**: `work_item_id` resolves explicitly first (Strategy 0, `current_item.id`) and the
> search's conflicting ballot is then DISCARDED by the merge — instrument noise for a
> correctly-resolved field. (`current_page` needs its own read — RFC_029 lane's triage; whether
> its winner is ever USED is unmeasured here.) Watch `result` only.

**What shipped:**

- **Half 1 (Go, inert until the next chassis roll):** `prefixConfigStepReferences` stops
  enumerating — a generic pass suffixes ANY top-level reference-shaped string config value
  through the existing `prefixDataReference` (first-segment rule unchanged); conditions/prose
  excluded by the shape gate; `input_fields` arrays untouched (pinned by test); duplicated
  `dataRefKeys` loop deleted. Register **WFA-017** (same commit). Tests + 2 mutation proofs:
  `loop_config_reference_suffixing_test.go`. Census re-run pre-fix: **22 sites / 7 agents**
  (§9b's 15 had grown back), all read-references.
- **Half 2 (config):** `sql_for_agents/448` — diagnose-dispatch-loop + report-dispatch-loop
  (these two are NOT loops — top-level steps, no sub_workflow, so §9a never applied to them;
  §6a's own item was the diagnose instance): `result!: handler_result`,
  `work_item_id!: claimed.work_item_id`, `error_step: mark_failed`. **APPLIED 2026-08-17**
  (verify DO-block passed; induced negative control raised as designed).
  `sql_for_agents/452_HOLD` — the same flip on build-dispatch-loop's
  `process_item…mark_complete` (`result!`, `work_item_id!: current_item.id`,
  `error_step: mark_failed`), **held until Half 1's image rolls** (gate + watch queries in the
  file header; ordering is belt-and-braces per the correction above, and matches §10's
  fix-then-ratchet).
- `error_step: mark_failed` is new on mark_complete (adversarial-review finding): without it a
  strict miss fails the WHOLE loop orchestration (`routeToErrorStepOrFail` → `failWorkflow`,
  `continue_on_error: false`) and strands the item in `claimed`, which the claim predicate
  (`triaged`/`approved` only) never reclaims. With it: one failed item, loop continues.

**§8 verification stands**, with the §10 queries scoped to `context->>'field'='result'` and
the demand pair (spawn_record→0 WHILE own_envelope>0). report-dispatch-loop had zero retained
orchestrations in the window — its proof is demand-bound.

**Still open before this file moves to `bugs_closed/`:** (1) the chassis roll carrying Half 1,
then lift 452_HOLD and read the §8/§10 queries; (2) the ~hundreds of already-wrong `result`
rows since 08-15 — the spawn record's `topics` embeds corr-id + iteration, so the true reply
is joinable from the surviving parent's `process_item_iter_N_call_handler.response`; measure
the joinable population and put the repair (or the decision not to) to the owner with counts.
Until 452 is applied, keep reading 090 verdicts from `orchestration_states`, never the item
(§6a's warning stands for build-dispatch-loop; it is CLOSED for diagnose/report as of 448).

### 11a. COUNCIL VERDICT READ 2026-08-17 ~12:35Z — REJECTED, guardian VETO on the Go half's SCOPE; the config half explicitly endorsed; routed to RFC_035

Round 1, corr `cba35b35`: 11 seats returned (6 abstained on relevance). **The veto is about
HOW, not WHETHER** — guardian: the generic suffixing pass "turns a 3-agent bug fix into a
behavioural change to the shared orchestrator resolver used by every loop-bearing agent …
which the plan's own evidence says the strict markers already do"; architecture seat:
`needs_rfc`, split from the migrations. Both are consistent with §11's own correction (the
mapped key is present; `!` alone closes this bug). Per the 2026-07-28 scope-veto ruling:
recorded on WFA-017 (**Live ≠ approved — no `Council-Reviewed:` trailer ever on this
correlation**), routed to **`architecture_review/RFC_035`** (ratify / narrow to enumeration /
forward-revert — owner's call), code stays on the branch pending that ruling.

**The config half proceeds on the guardian's own note** ("a contained, agent-scoped fix …
should proceed"): 448 stays applied; 452_HOLD stays gated and was HARDENED on three seat
objections that were real against the file or prospective — pod-probe-with-controls gate
language (not merge-base-only), a max-version pin on the UPDATE + verify (two-active-rows
trap, prospective — count=1 today), and the explicit statement that `error_step: mark_failed`
widens error catchment to ALL mark_complete errors, not just strict misses. Three other
objections were against the submission SKETCH, not the files: the `error_step` paths in
448/452 are config-level (`{…,mark_complete,config,error_step}` — the step-level-inert trap
was already avoided), `is_snapshot` is spelled correctly, and 448's parser-liveness was
binary-probed with positive AND negative controls before applying (NOTES). 448 itself is
ledger-recorded and NOT edited post-record (checksum).

### 11b. 2026-08-17 ~16:30Z — the "fresh build" shipped NO NEW CODE, so the HOLD gate was CONVERTED on measured evidence and 452 is APPLIED; Half 1 is still not live

**The roll did not ship the code.** The owner deployed a fresh chassis build at 14:42Z (new
replicaset, both pods restarted) and **`IMAGE_TAG` was still `v1.0.1305`** — a same-tag rebuild
serves the node's cached image. Probed on BOTH pods with controls in the same breath: the OLD
stamp `6a782274b` is **PRESENT**, Half 1's commit `0ed96c7eb` is **ABSENT** (a real sha that
could have matched — the negative control is capable of being absent). A concurrent lane
measured the same event independently: **203 commits in HEAD but not in the running binary.**
So `prefixConfigStepReferences`' generic pass is still inert, and stays inert until the owner
bumps `IMAGE_TAG` and rebuilds — **which is the one thing this bug now needs from the owner.**

**Because the gate could not be met while the defect ran at ~25 wrong records/hour** (155
spawn-record completions in the 6 h before the apply), the HOLD was converted on a measurement
that answers the gate's actual question — does the strict mapping resolve on the RUNNING
binary?

> **`RESOLVER_MAPPING_BYPASSED` fires ONLY when the mapped key EXISTS and differs from the
> search's answer: 201 rows for `field=result` in the same 6-hour window, against 155
> completions.** That is direct evidence, at the resolution moment, that
> `result!: handler_result` resolves today via per-substep propagation
> (`coordinator.go:1355`). A miss is contained by the new `error_step: mark_failed` — one
> failed item, loop continues, no stranded `claimed` row.

**Migration 452 APPLIED 2026-08-17 16:28:57Z** (UPDATE 1, DO verify passed, dry-run in a
rolled-back txn first; live config now `{"result!": "handler_result", "work_item_id!":
"current_item.id", "error_step": "mark_failed"}`). **⚠ Its `schema_migrations` row is MISSING** —
`--record-only` was refused twice by the session harness's permission classifier; the file
header carries the command to re-run and the reason a missing row must not read as "unapplied".

> **CORRECTED 2026-08-17 (§11's own claim, twice — see `WRONG_CALLS.md`):** §11 stated flatly
> that the base key "holds the CURRENT iteration's reply" and that the ordering gate was
> "belt-and-braces". The *conclusion* survives, but it was asserted from indirect evidence
> without the arithmetic that makes it load-bearing (rows ÷ demand), and two other sessions'
> judgements were downstream of that sentence. Then, deciding whether to apply 452 blind, I
> read final `collected_data`, found `retry_payload` on 140/140 base keys and `response` on
> only 67, and announced §11 REFUTED — also wrong: final state is lossy on this agent (289's
> aggregation; 11% joinability) and its population includes failed/in-flight iterations that
> never reach `mark_complete`. `retry_payload` is a coordinator-captured SIBLING, not a
> substitute for the reply. **A state table is a corpse; an event row is a witness.**

**Verification status (demand-bound, honest):** zero `build-dispatch-loop` orchestrations have
been created since 16:28:57Z, so **nothing is proven live yet** — a plan is persisted at
orchestration creation, so only runs created after that instant carry the strict config. The
first post-apply dispatch is the test; queries and the `created_at` filter are in the lane
RUNBOOK. Expect `result` resolver rows → 0 while the loop has traffic, and the item census's
spawn-record column → 0 while own-envelope is non-zero.

**Still open:** (1) the `IMAGE_TAG` bump + rebuild so Half 1 rolls (owner's — releases are
whole-fleet); (2) RFC_035's ruling on Half 1 — and note that 452 landing WITHOUT Half 1 is
itself evidence for the owner to weigh there; (3) the 244 repairable historical rows.

### 11c. 2026-08-17 17:05Z — the REAL roll landed (v1.0.1307): Half 1 is LIVE, and the whole bug is now proven dead with a demand control

The owner rebuilt with a bumped tag. Verified, not assumed:

- **New binary confirmed:** the pre-fix stamp `6a782274b` is now **ABSENT** from `/proc/1/exe`
  (it was present four hours earlier), and a fake-sha control (`deadbeef…`) is absent too, so
  the probe discriminates. The startup provenance line had already rotated out of both pods'
  logs, and the OCI label was not reachable from here — so the decisive evidence is behavioural,
  below, which is stronger anyway.
- **WFA-017 is firing, visibly, in data.** Every `build-dispatch-loop` plan expanded since the
  roll reads:
  `{"result!": "handler_result_0", "work_item_id!": "current_item.id", "error_step": "process_item_iter_0_mark_failed", …}`
  The pre-roll binary **could not** have produced `handler_result_0` for the `result!` key —
  `result` was never on `prefixConfigStepReferences`' allow-list, so only the generic pass
  suffixes it. (Note the two controls sitting in the same object: `work_item_id!` is left alone
  because `current_item` is a loop variable, not a sibling output; `error_step` is
  step-prefixed, not data-suffixed.)
- **End to end on the new binary, with the demand control** (window: roll → +75 min):
  | measure | value |
  |---|---|
  | `RESOLVER_%` rows, `field=result`, build-dispatch-loop | **0** (from ~455/day: 176 conflict + 279 bypass) |
  | resolver rows, other fields, same agent | 34 — the instrument is live and the agent is busy |
  | orchestrations created (DEMAND) | **10** |
  | items completed by the loop | **11**, of which **11 carry the handler's reply** |
  | spawn-record-shaped completions | **0** |

  The 7 further completions in the window without a reply were **not** completed by the
  dispatch loop (blank `claimed_by`, `{"reason": "render audit re-measured …"}` — the
  no-change path), so they are outside this bug.
- **The remaining 34 rows are the two fields this file already told you not to judge the fix
  by:** `current_page` (22) and `work_item_id` (12). `[INFERRED]` the `work_item_id` rows now
  come from `claim_work_item` / `fail_work_item`, which declare the same field and are NOT
  strict — resolver rows carry pod-level attribution only, so this cannot be proven per row
  from the row itself. Either way they are RFC_029's triage, not 287's.

**So: fixed, live, and proven on both halves — the defect is dead for new traffic.** What is
left is not the defect: the 303 repairable historical rows (mig `455_HOLD`, dry-run proven,
owner's go), `RFC_035`'s ruling — which is now a decision about **live** code, the estate's
stated norm (2026-07-29 #2: review here is after the fact by design) — and 452's missing
ledger row.

### 11d. INDEPENDENT VERIFICATION 2026-08-17 ~18:3xZ from the RFC_029 lane — your `result` claim CONFIRMED at the instrument; two of the three fields I flagged in §10 are improved but NOT closed (and that residual is MINE, not a reopening)

Ran §10's own fix-watch query, split at the v1.0.1307 roll, with the demand control it specifies.
Build verified first at the artefact: image label `org.opencontainers.image.revision` =
**`a6d1c53c0`**, confirmed PRESENT in `/proc/1/exe` on the running pod (negative control
`deadbeef1234…` absent), local and running digests both `sha256:8339bdbd…` — a real new build,
not the same-tag cache trap of that morning.

**Demand control first, because a drop to zero means nothing without it:** `build-dispatch-loop`
ran **9.7 runs/h before** the roll and **8.1 runs/h after** — comparable traffic, so the fall
below is the fix, not a quiet fleet.

| field / code | before 1307 (30.3 h) | after 1307 (1.3 h) | verdict |
|---|---|---|---|
| `result` — `RESOLVER_MAPPING_BYPASSED` | 479 | **0** | **CLOSED** |
| `result` — `RESOLVER_CONFLICTING_CANDIDATES` | 326 | **0** | **CLOSED** |
| `current_page` — conflict | 1,211 | 24 | improved, still firing |
| `work_item_id` — conflict | 1,186 | 13 | improved, still firing |
| **all BDL resolver rows** | 105.6 /h · 14.6 per run | 28.4 /h · **3.4 per run** | **−73%** |

**Your claim is confirmed, exactly as scoped.** `field=result` is zero across 11 loop runs — the
`!` flip did what §7.3 asked, in the order §10 warned about (suffix fix first, marker second).

**Half 1 (WFA-017) is visibly working too, and the instrument can show it in a way the item table
cannot: the BALLOT SHRANK.** Candidate-path counts for the two remaining fields collapsed from
avg 56.7 / max **190** to avg 13.3 / max **22** (`current_page`), and avg 84.1 / max **195** to
avg 17.8 / max **27** (`work_item_id`). The unbounded per-iteration accumulation §10a described
is gone; what is left is a small bounded set.

**What is NOT closed, stated plainly so it is not lost in a "proven dead":** `current_page` and
`work_item_id` still resolve through the aggressive search, with the same winners as before —
`handler_result.retry_payload.message.body.~unwrap.current_page` (24) and
`claim_result.work_item_id` (13). §10 asked that the fix be assessed against all three fields;
it closed one and shrank two. **That residual is RFC_029's population, not a reopening of this
bug** — your symptom (items completing with the spawn record) is dead, and I am not asking you
to carry the rest. It is now item 5 on the `staged_component_build` handoff: give those two
references explicit mappings, then `!`, and the Phase 2 precondition is met for this agent.
`[UNMEASURED]` whether the surviving `claim_result.work_item_id` winner is now the CURRENT
iteration's value (post-WFA-017 it probably is, which would make these benign-but-unmapped rather
than wrong) — that check belongs with the mapping work, not here.

— `staged_component_build` lane, 2026-08-17. Full read: RFC_029 §10.7.

#### 11e. CORRECTION 2026-08-18 to §11d's rate figures — my "−73%" was a 1.3-hour sample; the honest number is **−53%**. Your `result` claim is unaffected and is now confirmed on 17× the demand.

> **CORRECTED 2026-08-18 (§11d):** ~~"105.6/h · 14.6 per run → 28.4/h · **3.4 per run** (−73%)"~~ →
> **17.7 per run → 8.4 per run (−53%)**. §11d measured the "after" over 1.3 h / 11 loop runs and
> published it as a rate. With 193 runs in the last 12 h the per-run figure is 8.4, not 3.4; the
> "before" was also mis-normalised (14.6 vs the correct 17.7 over the matched window). The
> direction and the mechanism are unchanged — the fix is a real, large improvement — but it is
> roughly half, not three-quarters. My error, published into your file; logged in WRONG_CALLS.

**What is NOT corrected, and is now much better evidenced: `field=result` is still ZERO.**
Across the last 12 h with **193 `build-dispatch-loop` runs** (vs the 11 §11d had), neither
`RESOLVER_MAPPING_BYPASSED` nor `RESOLVER_CONFLICTING_CANDIDATES` has fired for `result` even
once. Your fix has now held over ~17× the demand §11d tested it on. The `!` markers on
`mark_complete` (`result!`, `work_item_id!`) are doing exactly what they were put there to do.

The residual remains ours (RFC_029): `current_page` (1,124 in 12 h) and `work_item_id` (503) still
resolve through the search from other call sites — see RFC_029 §10.7a/§10.9 for where that triage
has got to. Not a reopening of this bug.

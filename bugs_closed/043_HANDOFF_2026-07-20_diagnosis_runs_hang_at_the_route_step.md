# 043 — diagnosis runs hang at the `route` step, so the loop returns no verdicts

> **CLOSED 2026-07-26 — fixed AND live. Moved to `/bugs_closed/`. See the closure
> section at the foot of this file for the evidence.** Resolve this case **by slug**:
> the number `043` is also used by the unrelated `043_…_generated_page_copy_invents_
> quantitative_claims.md`, which remains OPEN in `/bugs_open/`.

**Filed 2026-07-20** (vetcomparison thread). **Status: OPEN.**
**Cause NOT determined** — this is an observation file. Deliberately not filed *into* the
diagnosis loop, because the diagnosis loop is the thing that is stalling; filing there would be
futile. That is also why it matters: CLAUDE.md now makes the loop the **default** for any durable
claim, and today it answered nothing.

## Symptom

Every diagnosis run filed from this thread on 2026-07-19/20 produced bundles and then stopped at
the same step. None reached a verdict.

| correlation | outcome | step |
|---|---|---|
| `be60b0d7-21c4-4e02-be95-2ec37387004f` | FAILED — `reaper: stale AWAITING_RESPONSES for >90 min` | `spawn_diagnoser` |
| `12ff5852-3950-4fec-ac68-1cc8a86d0521` | FAILED | `call_diagnoser` |
| `459fbdf3-5e48-4fe4-977a-10676d7437f6` | FAILED — `reaper: stale EXECUTING_STEP for >4h; step=route` | **`route`** |
| `f155b0c4-881b-4369-abe4-569d7b2ad4c8` | hung EXECUTING_STEP ~4 h at time of filing | **`route`** |
| `55dc0fa4-116c-40d6-90b2-bfad9ad73692` | hung EXECUTING_STEP ~4 h at time of filing | **`route`** |

Three of three runs that got as far as `route` hung there. Bundles were produced in each case
(`diagnosis_artifacts kind='bundle'`), so gathering works; it is the controller that does not
advance.

## What `route` is

`diagnose-agent`'s `route` step runs the `diagnose_route` action — the controller that loops
between `load_runtime` (gather) and `emit`, forwarding the verdict's data requests and overriding
`next_step` to either loop back or stop:

```json
{"action": "diagnose_route",
 "config": {"emit_step": "emit", "gather_step": "load_runtime",
            "state_field": "route.diagnose_state", "verdict_field": "verdict.result",
            "analysis_field": "repo_analysis", "max_iterations": 5}}
```

## A tidy hypothesis that turned out to be WRONG

`route`'s config carries `max_iterations: 5`, a **numeric** value — and `/bugs_open/042` proves
numeric step config never reaches actions that read it through `ExtractActionInputs`. A
controller silently receiving the wrong iteration bound would explain a loop that never
terminates, and it would have been an elegant story: the config defect disabling the loop that
would diagnose the config defect.

**It is not the cause.** `diagnose_route_action.go:103` reads
`datahelpers.GetIntField(config, "max_iterations", 5)` — straight off the raw config map,
bypassing `ExtractActionInputs` entirely. 042 does not touch this path.

Recorded because the next person will form the same hypothesis, and because a plausible story
that fits the symptom is not evidence. Check the call site before adopting it.

## Where to start

- Is `route` waiting on the verdict child (`verdict.result` never populated), or looping without
  a termination condition? `execution_path` / `orchestration_state_audit` for a hung correlation
  will distinguish these.
- `diagnose_route`'s advance logic and its guards, against a run whose bundles exist but whose
  `verdict` field is empty.
- Whether this correlates with the 2026-07-20 image roll — `459fbdf3` was filed and died before
  it, so probably not, but confirm rather than assume.

## Operational consequences, worth knowing regardless of cause

- **A filed diagnosis is not a delivered one.** Check for a terminal verdict before treating a
  correlation id as an answer. Three id's in this thread's docs would otherwise read as
  "diagnosed".
- **A dead run leaves its intake reading `awaiting_diagnosis`**, which makes the 090 trigger
  refuse a refile as a duplicate. Close the stale intake first (record why in
  `spec.failure_reason`), then refile — note `idx_swi_dedup` is partial over non-terminal
  statuses, so the closed row and the fresh one coexist correctly.
- **The `stale-orchestration-reaper` itself stalled** for ~2 h on 2026-07-20 (last_triggered
  17:01 against a 180 s interval) before recovering, so hung runs were not even being reaped
  during that window.
- **Submissions fired into a busy cluster take ~30 min to start — do not read the gap as a stall.**
  > **CORRECTED 2026-07-21 — the paragraph that stood here was WRONG and I have removed it.** It
  > claimed the council gate "is not starting either… the same machinery stalling," on the evidence
  > that my two council submissions (`712be028`, `563462b8`) had zero `orchestration_state_audit`
  > rows 13–14 min after submitting. **Both actually started and COMPLETED about an hour after
  > submission** (found the next day by payload: `complete_invalid` at 20:07 and 20:08). They were
  > not stalled and had nothing to do with the diagnosis-loop `route` hang. CLAUDE.md (updated
  > 2026-07-20, after my session-start copy) states publish→run-start was measured at **~29 min**
  > and that a missing orchestration row is almost always latency, not a drop — retrying costs a
  > duplicate round, which is exactly what my resubmission was. Logged in `WRONG_CALLS.md`.
  > Separately, both runs ended `complete_invalid` because the submission was structurally invalid
  > (`edit 2: operation "create" not in the allowlist`), never because of any stall.
  >
  > So the honest **route-hang tally stands on the diagnosis runs alone**: three of three runs that
  > reached `route` hung there and produced no verdict. The council is not evidence for or against
  > this bug. Find any dispatched run by payload and give it ~30 min before concluding anything.

## Related

- `/bugs_open/042` — the config-literal defect; its diagnosis run (`f155b0c4`) is one of the hung
  ones, which is why 042 carries a full case file rather than relying on a verdict.
- **`/bugs_open/003` (spawn-loss) — this IS the owning class. See the diagnosis below.**

---

# DIAGNOSIS 2026-07-21 (bugfix-043-route-hang thread)

**Cause structurally identified: `043` is an instance of the actively-owned `bugs_open/003`
(spawn-loss / at-most-once consume / zombie `EXECUTING_STEP`) class, NOT a defect in `route`'s
logic.** The exact pod-death *trigger* is not yet pinned (see "What is NOT pinned"), but every
`route`-internal cause has been ruled out with evidence, and the strand-and-reap mechanism is
`003`'s. Filed here rather than into the diagnosis loop because the loop is the thing that stalls.

## What the live evidence shows

Queried `orchestration_states` for every hung run (`workflow_plan->'steps' ? 'route' AND status
FAILED AND current_step='route'` = **20 runs, all 2026-07-19/20**; distinct real symptoms, all
against the `agentchassis` repo — not a retry storm):

- **`route` never returns.** A hung run has a populated `verdict.result` (e.g. `7fb35bbb…`:
  `UNVERIFIABLE` with 4 `next_scope` entries) but **no `route` key in `collected_data` at all** —
  the action started and never produced its result map, so `saveStepResultWithRetry` never ran.
  `error` is only the reaper line `reaper: stale EXECUTING_STEP for >4h; step=route`, so it did
  **not** error/panic (panics are recovered → `action "diagnose_route" panicked: …`, which is
  absent). `execution_path` is `[]`; `currently_executing='route'`.
- **The pod is gone; the reaper is `003`'s.** `processing_node` is a per-run ephemeral
  `agent-diagnose-agent-*` **Job** pod (spawn spec: `RestartPolicyOnFailure`, `BackoffLimit 3`,
  `ActiveDeadlineSeconds 86400`, `TTLSecondsAfterFinished 3600`). The `EXECUTING_STEP >4h` reaper
  that FAILs them is exactly `bugs_open/003` **F1** (commit `539768695`). These 20 runs are `003`
  zombies. `003` already records `bug(003): fresh occurrence on a DIAGNOSIS spawn — sweeper did
  not retry` (`bcd14b8bc`) and `at-most-once consume loop destroys in-flight work on restart`
  (`027aa7588`) — the strand mechanism: the triggering Kafka message is already consumed/committed,
  so a dead/ restarted pod leaves nothing to re-drive the orchestration.
- **Death is early and load-correlated, not a fixed deadline.** `processing_history` (heartbeats
  are written per state-save, `state.go:UpdateStateWithVersion`) shows the workflow ran normally —
  `analyse_repo → lookup_symbols → load_runtime → assemble_bundle → verdict(≈41s LLM) → route` —
  then went silent **8–36 s into `route`** (measured across all 20). Total elapsed varies **108–535 s**
  (no fixed deadline). Failures cluster in bursts (13 distinct diagnoses 20:37–21:25 on 07-20) and
  as singles; **29 diagnose runs COMPLETED** in the same period (latest 07-20 20:45, interleaved
  with the burst), several running the resolver across up to 5 iterations. So `route` is **not**
  deterministically fatal — it fails under load.

## What is ruled OUT (each with its check)

- **NOT the `max_iterations` numeric-config red herring** (the tidy `042` story) — already refuted
  in this file; `diagnose_route_action.go` reads it via `GetIntField`, bypassing `ExtractActionInputs`.
- **NOT a logic hang / infinite loop.** `Advance → DecideStep → nextScope → Neighbourhood`
  (`pkg/diagnose/{advance,step,callgraph}.go`) is pure and single-hop; no unbounded recursion.
- **NOT bounded-external slowness.** Embedding client has a 120s HTTP timeout; a fully-hung
  ollama gives ≤120s×N-entries ≈ 8 min worst case, then trigram fallback, then return. `route`
  would still *return* — it never does. `code_symbols` is **3,723 rows with an HNSW + GIN-trgm
  index**, so the vector/trigram searches are instant, not slow.
- **NOT memory OOM.** Limit 1Gi; sibling chassis agents (`agent-page-content-writer`,
  `agent-build-dispatch-loop`) peak at **33–73 Mi** in Prometheus; `repo_analysis` is 2 MB /
  461 files, `collected_data` ~1.3 MB. Nowhere near 1Gi.
- **NOT a Kafka rebalance/poll kick.** This is segmentio/kafka-go; group membership is kept by a
  **background heartbeat goroutine** (`HeartbeatInterval = session/3`), so a long synchronous
  message-handler does not trip a rebalance.
- **NOT a liveness-probe DB/CPU starvation.** The agent-chassis liveness handler is
  `KafkaReachability.HandleHealth` — a **non-blocking atomic read**, DB-independent by design
  (`kafka_reachability.go`); it 503s only after 300 s of continuous all-broker unreachability.

## The route-specificity (the one thing `003` alone doesn't explain)

All 20 of the 07-19/20 deaths are at `route`, never at the equally-frequent, *longer* `verdict`
step — so `route` does something special. `route` is the **only local step that reaches out**
(embedding HTTP to in-cluster ollama-adapter + a `code_symbols` read per fuzzy `next_scope` entry,
via the §7D resolver) **on a pod whose DB pool is `SetMaxOpenConns(4)`** (`agent.go:248`,
`processor.go:68`), and it does so on the **raw, unbounded action ctx** — unlike the sibling
`rag_index` action, which wraps its embedding call in a `context.WithTimeout` (`rag_actions.go:206`).
The deaths cluster in `route`'s ~9 s embedding window. `[INFERRED, UNPINNED]` the most likely
trigger is node-level resource pressure/churn during the concurrent-diagnosis bursts evicting the
long-lived, analysis-heavy diagnose-agent Job pods, caught during `route`'s reach-out window — but
this is not proven and a competing "resolver reaches out and the pod is torn down mid-call" story
fits equally.

## What is NOT pinned

The exact instruction that terminates the pod mid-`route` (node/kubelet eviction vs. Job/container
restart vs. spot reclaim). Pinning it needs a **live reproduction**: dispatch a small BURST of
diagnoses (a single run usually completes), then watch the pods — `kubectl get pod <p> -o yaml`
(`status.reason`/`containerStatuses[].lastState`), `kubectl get events`, and exit code — to catch
one dying at `route`. That is the cheapest place to settle it and should precede any larger fix.

## Fix applied (this thread) — hardening, honestly scoped

`diagnose_route_action.go`: the §7D resolver now runs under a **hard wall-clock budget**
(`resolver_budget_seconds`, default 20) with a **per-embedding-call ceiling**
(`resolver_embedding_timeout_seconds`, default 15), both passed down so the embedding + DB calls
cancel on the deadline; on expiry entries **fail-open** to their prose labels (the §7D
"no-worse-than-not-resolving" contract). This fixes a **real latent defect** — `route`'s external
calls previously ran on the raw unbounded ctx in a latency-critical control loop on a 4-connection
pool — and matches the sibling `rag_index` pattern. **It is defence-in-depth, not the root fix:**
the observed deaths are 8–36 s in (below the 20 s budget), so this removes `route`'s pathological
minutes-long worst case and any indefinite reach-out, but will not by itself stop an environmental
mid-`route` kill. Inert until a chassis image roll.

## The root fix belongs to `bugs_open/003` (OWNED — do not fork)

The durable fix is `003`'s in-progress work: **at-least-once consume / retry the orchestration when
its pod is lost** (F2/F3), so a diagnose-agent that dies mid-`route` is *re-driven* instead of
stranded. Diagnose-agent is disproportionately exposed because it runs a **long, heavy,
multi-iteration marathon** (repo clone + analysis + up to 5×`gather→verdict→route`, 100–535 s) in
one ephemeral Job pod. Contribute these findings into `003`; do not start a competing structural fix.

**Eviction-exposure mitigation APPLIED 2026-07-22 (migration `191_diagnose_agent_resources.sql`,
live + committed):** diagnose-agent carried NO explicit `resources`, so it inherited the
`agent_definitions` table default — requests **cpu 100m / memory 256Mi**. A pod is an eviction
candidate under node pressure precisely when it exceeds its REQUESTS, so a long-lived analysis-heavy
pod on a 256Mi request was a disproportionately likely victim. Its row now carries an explicit
requests bump to **cpu 250m / memory 512Mi** (limits unchanged at 500m/1Gi — memory peak ~150Mi, and
OOM was ruled out). The spawner reads `resources` at spawn time, so this is **live on the next
diagnose run — no image roll**. This reduces exposure; it is NOT the root fix (still `003`'s F2/F3).

---

# CLOSED 2026-07-26 — fixed AND live (bugfix-043-route-hang thread)

Everything this case was waiting on is now live in production, and the symptom is extinct on
every durable record the platform keeps. Closing on **extinction in practice plus an owned,
live root fix** — not on a pinned eviction trigger, which was never obtained (see "What this
closure does NOT claim").

## 1. The `route` hardening is live — and live *where `route` actually runs*

The resolver budget was committed 2026-07-21 (`2c82cf804`) and was inert pending a roll. It has
since shipped. Pod-grep of the running chassis binary for two strings the fix itself **created**
(a discriminating grep, not one the pre-fix code also contained):

```
$ kubectl -n ai-persona-system get pods -l app=agent-chassis \
    -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.containers[0].image}{"\n"}{end}'
agent-chassis-f4d46c88d-p6wqc	docker.io/aqls/agent-chassis:v1.0.1165

$ kubectl -n ai-persona-system exec agent-chassis-f4d46c88d-p6wqc -- sh -c \
    'strings /app/agent-chassis | grep -c "resolver_budget_seconds";
     strings /app/agent-chassis | grep -c "resolver_embedding_timeout_seconds"'
1
1
```

**The `bugs_open/066` trap was checked explicitly, and it is clear here.** `route` does not run in
the deployment pod — it runs in the per-run ephemeral `agent-diagnose-agent-*` Job pod, whose image
comes from the agent's own row, not from the Deployment. Under `066` that pin routinely lags a roll,
which would have made the pod-grep above a **false green** for this case. It does not, because the
row is current:

```sql
SELECT type, image_repository, image_tag FROM agent_definitions
WHERE type='diagnose-agent' AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
-- diagnose-agent | docker.io/aqls/agent-chassis | v1.0.1165
```

Same tag as the deployment, so the spawned pods that execute `route` carry the budget. (`066`
itself stays open — the class is real; this row simply happens to be current.)

## 2. Migration 191 is still applied

```sql
SELECT type, resources->'requests' AS requests, resources->'limits' AS limits
FROM agent_definitions WHERE type='diagnose-agent'
  AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
-- diagnose-agent | {"cpu": "250m", "memory": "512Mi"} | {"cpu": "500m", "memory": "1Gi"}
```

Verified 2026-07-26, i.e. it has survived every re-seed since 07-22.

## 3. The root fix — `bugs_open/003` F2/F3 — is live and owner-ratified

`003`'s at-least-once consume + DB-driven retry driver went live in **v1.0.1159** (2026-07-25,
carried by another session's fleet build) and the owner ruled **KEEP LIVE** against the live
evidence: first ~4.5 h showed 175 awaited requests, 19 retried, **7 recovered end-to-end** that
would previously have been silent losses, 6 loud `error` at the retry cap replacing silent 90-min
strandings, 0 stuck `retrying`, 806/806 dedupe claims completed. That is the mechanism this case
was strand-ing on, now fixed at the class level by its owning lane. The redesign is documented as
`RFC_001` in the architecture-review track.

## 4. The symptom is extinct

The three independent records all agree, and none of them shows a `route` hang after 2026-07-20.

**Diagnosis intake — zero stranded:**

```sql
SELECT status, count(*), max(created_at)::date AS latest
FROM site_work_items WHERE item_type='needs_diagnosis' GROUP BY status;
-- cancelled |  4 | 2026-07-20
-- complete  | 21 | 2026-07-25
-- failed    |  5 | 2026-07-20
```

Every `failed` and `cancelled` row dates to **2026-07-20 or earlier** — this bug's own window.
Nothing has failed since. Critically there are **zero `awaiting_diagnosis` rows**: the state this
case documented as the secondary damage (a dead run leaves its intake stuck `awaiting_diagnosis`,
which then makes the 090 trigger refuse a refile as a duplicate) does not exist anywhere.

**Throughput at the load that used to kill runs** — `diagnosis_artifacts` by day, 07-21 → 07-26:
bundles 2 / 27 / 0 / 28 / 20 / –, with `fix_plan` and `council_report` rows every single day
including today. The 07-19/20 deaths clustered in bursts of 13 concurrent diagnoses; 07-22, 07-24
and 07-25 each ran that scale or more without a single strand.

**The discriminator itself, on the surviving rows.** `orchestration_states` is pruned at ~24 h, so
only two diagnose runs are retained — but they answer the exact question this case turned on. The
hung runs were identified by a populated `verdict.result` with **no `route` key in
`collected_data` at all** (the action started and never produced its result map). Both surviving
runs have it:

```sql
SELECT substring(correlation_id::text,1,8) AS corr, status, current_step,
       collected_data ? 'route' AS route_saved, collected_data ? 'verdict' AS verdict_saved,
       created_at, updated_at
FROM orchestration_states WHERE workflow_plan->'steps' ? 'route' ORDER BY created_at;
-- c19ed5b2 | COMPLETED | complete | t | t | 2026-07-25 17:13:04 | 2026-07-25 17:55:55
-- 4ab9473b | COMPLETED | complete | t | t | 2026-07-25 17:14:22 | 2026-07-25 17:49:07
```

Both `route_saved = t`, both COMPLETED, and both ran **35–42 minutes** — i.e. long multi-iteration
marathons of exactly the shape that used to die 8–36 s into `route`. A concurrent check found
**0 non-COMPLETED rows** carrying a `route` step.

## 5. What this closure does NOT claim

- **The exact instruction that killed the pods was never pinned.** The "What is NOT pinned" section
  above stands: node/kubelet eviction vs. Job/container restart vs. spot reclaim was never
  distinguished, and no induced mid-`route` kill was run **in this lane**. The `003` lane's
  controlled kill test (2026-07-25) proved cross-pod rescue on the adjacent `AWAITING_RESPONSES`
  branch, not on a mid-`EXECUTING_STEP` `route` death specifically.
- So the honest closure is: the hardening is live, the exposure is reduced, the strand mechanism is
  fixed at the class level by its owner, and the symptom has not recurred across five days of
  normal and burst load. Not: "we watched the trigger and eliminated it."
- **If it recurs**, the cheapest first move is still the live reproduction described above — burst
  a set of diagnoses and catch a pod dying at `route` with `kubectl get pod -o yaml`
  (`status.reason` / `containerStatuses[].lastState`), `kubectl get events`, and the exit code —
  and it should be reported into `003`, not refiled here.

## 6. Residuals, and who owns them

Both belong to the `003` lane; **do not fork them here**:

- **`bugs_open/075`** — a dead coordinator's `processing_node` makes responses permanently
  undeliverable, and post-F2 the retry driver then loops forever with real side effects. Found by
  `003`'s own kill test, contained the same hour, fix mapped.
- **The kill-test re-run and F4 liveness re-test** owed by `003` after `075` lands.

`bugs_open/066` (spawned pods pin stale image tags) is checked-and-clear for diagnose-agent today
but remains open as a class — anyone verifying a future `route` fix must re-check the agent row,
not just the deployment pod, or the pod-grep is a false green.

# 259 — one provision request builds several billable GPUs (~~Kafka redelivers while the handler blocks~~ the chassis retry driver re-dispatches an expired await)

> ## ⚠ **CORRECTED 2026-08-12 (evening) — THE TRIGGER IN THIS FILE'S ORIGINAL TITLE AND "Why" SECTION WAS WRONG.**
>
> **It is not Kafka redelivery. It is the chassis's own retry driver.**
>
> The behaviour section below is unaffected — it was counted, and it stands. What
> was wrong was the *mechanism*, which the original filing marked `[UNVERIFIED]`
> for the broker-side path but nonetheless named in its title, so a reader would
> reasonably have acted on it. The correction, with its evidence:
>
> ```sql
> SELECT correlation_id, count(*) AS rows, count(DISTINCT request_id) AS req_ids
> FROM awaited_requests WHERE target_agent_type='thunder-adapter'
>   AND sent_at > now() - interval '2 days' GROUP BY 1;
> -- the three problem correlations: 4 rows each, 4 DISTINCT request_ids each.
> ```
>
> Detail for correlation `23c9bc6a`, one orchestration (`8c5bf926`), four `step_id`s:
>
> | request_id | step | sent | timeout_at | processed_at | status |
> |---|---|---|---|---|---|
> | `b40062fa` | dispatch_provision | 13:52:12 | 14:02:12 | 14:02:12 | processed |
> | `da87c9c9` | dispatch_provision | 14:02:13 | 14:12:13 | 14:12:13 | processed |
> | `6f7bda74` | dispatch_provision | 14:12:15 | 14:22:15 | 14:22:15 | processed |
> | `641450c2` | dispatch_provision | 14:22:17 | 14:32:17 | 14:32:17 | **error** |
>
> **What that says.** Each row is `processed` at *exactly* its own `timeout_at`,
> and the next dispatch leaves ~1 second later. That is an await expiring and
> the retry driver firing — `coordinator.go` `retryExpiredAwaitedRequest`, budget
> `RetryVersion < 3` (four attempts, then the orchestration FAILED at 14:32:17).
> The step's `timeout_seconds` is **600**, which is the ~10-minute gap the
> original filing read as a redelivery interval.
>
> **Redelivery is ruled out, not merely unproven.** A redelivered Kafka message
> replays identical bytes, so `request_id` would be *constant* across the
> attempts. It is different every time — because
> `DispatchThunderProvisionAction` mints a fresh one per execution
> (`thunder_provision_dispatch.go:99`). Four distinct ids = four distinct
> publishes. Nor is there evidence of any *extra* delivery: the adapter saw
> ≤ 4 deliveries per correlation, never more than the dispatch count.
>
> **A co-cause the original filing noted but did not connect.** Every await was
> cleared by its own timeout, never by a response — the adapter had answered
> ~5 minutes in (`Sent error response`) and the await sat `waiting` regardless.
> Had the error response cleared the await, the step would have failed after
> **one** attempt and only one GPU would ever have been built. So "does an error
> response clear an await?" is not a side observation; it is *why the retry loop
> ran at all*. Still undiagnosed — worth its own `090`.
>
> **What this changes for the fix (and what it does not).**
> - Fix candidate 1 (idempotency on `correlation_id`) is **still correct, and now
>   for a demonstrated reason**: across the four attempts `correlation_id` is the
>   *only* stable identifier. Keying on `request_id` would never fire — and
>   `request_id` is exactly what the dispatch code makes look canonical. That
>   trap is now in `LANDMINES.md`.
> - **Fix candidate 3 is RETIRED, not merely deprioritised.** Raising
>   `KAFKA_SESSION_TIMEOUT` / `KAFKA_REBALANCE_TIMEOUT` addresses a mechanism
>   that is not firing. It would have cost a build and changed nothing, while
>   slowing dead-member detection for every consumer on that config.
> - Candidate 2 (stop blocking the consumer) no longer removes the trigger
>   either — the trigger is upstream of the adapter entirely.
>
> **FIXED AND LIVE** since thunder-adapter `v1.0.1295` (2026-08-13), still live at
> `v1.0.1301` / stamp `0115f2b45` (2026-08-15) — `10659b419` confirmed an ancestor
> of both. Migration 396 applied. **Status stays OPEN** for one reason only: there
> is no LIVE BEHAVIOURAL PROOF yet. The guard has never been observed refusing a
> real re-dispatch, because provisioning has been paused since the day it was
> filed. See "How to verify a fix" above — and note that if the first unpaused
> provision SUCCEEDS quickly, the retry driver never fires and this proof still
> will not exist.

> ⚠ **THE NUMBER 259 IS AMBIGUOUS — resolve this one by SLUG.** A concurrent
> session filed `259_HANDOFF_2026-08-12_three_processor_paths_guard_on_a_handle_that_is_nil_in_production.md`
> the same day. Both are legitimate and neither is renamed (numbers are never
> reassigned). This file is
> **`259_…_one_provision_request_builds_several_billable_gpus`**; a bare "259"
> in any commit message, doc or landmine from 2026-08-12 could mean either, so
> `git log` the **file path**, not the number.

**Filed** 2026-08-12 by the `finetuning_uk_service` lane, from live Phase 0.
**Status: OPEN. CONTAINED — provisioning is PAUSED fleet-wide** (owner decision,
same day): `thunder_config.is_paused = true`, reason string names this bug.
**Do not unpause until this is fixed.** Undo, when it is:
`UPDATE thunder_config SET is_paused = false, pause_reason = NULL;`

Sibling: `bugs_open/258` (three other defects in the same provision path, found
in the same session). 258 is about not being able to provision; **this one is
about provisioning too much.**

---

## The behaviour

**Two provision requests produced three real, billing Thunder instances.**

Measured from the thunder-adapter pod log, 2026-08-12, counting deliveries per
`correlation_id` on `"msg":"Received request"` with `action=provision_instance`:

| correlation_id | times delivered | `POST /instances/create` results |
|---|---|---|
| `f17cccda-…` (vcpus 4, invalid) | **3** | 3× HTTP 400 — free, nothing built |
| `23c9bc6a-…` (vcpus 6, valid) | **2** | **2× HTTP 200** — `fi3966m0`, `bdd3upae` |
| `cd614594-…` (vcpus 6, valid) | 1 | 1× HTTP 200 — `v8b498mt` |

Redelivery gaps observed: ~10 min (13:40 → 13:50 → 14:05) and ~13 min
(13:52 → 14:05). Every accepted create starts billing immediately.

Each instance was later destroyed by the compensating cleanup when
`WaitForRunning` hit its 5-minute deadline (that is 258 defect 2), so the
observed cost stayed at roughly 15 minutes of a6000 — about **$0.10**. **That
bound is luck, not design.** Nothing in this path counts attempts or dedupes on
`correlation_id`; the loop is paced only by how long the vendor takes. On a GPU
that boots *just* inside the wait deadline, each redelivery would instead leave
a live box behind.

## ~~Why — the code facts~~ (SUPERSEDED — see the correction at the head)

> The three facts below are individually TRUE — they were read first-hand and
> re-verified on 2026-08-12 — but they do **not** compose into the cause. A
> handler that blocks for 5× its consumer's deadlines is a genuine latent
> hazard, and it is worth fixing on its own merits; it simply is not what built
> the extra GPUs. Kept as filed, because "true facts, wrong conclusion" is the
> more instructive record.

All three verified by reading, 2026-08-12:

1. **The handler is synchronous in the consume loop.**
   `internal/adapters/thunder/adapter.go:257` says so in terms:
   `// Sequential by design — no `go a.handleMessage(msg)`.`
   `handleMessage` → `handleProvisionInstance` (`adapter.go:325`).
2. **It blocks for minutes.** `handleProvisionInstance` gives itself a 10-minute
   outer context (`adapter.go:452`) and calls `ProvisionAction.Execute`, which
   blocks in `WaitForRunning` under `waitTimeout: 5 * time.Minute`
   (`provision_action.go:141,283-286`).
3. **The consumer's own deadlines are 60 seconds.**
   `platform/kafka/consumer.go:56,71-72` — `SessionTimeout` and
   `RebalanceTimeout`, both `envDurationOrDefault(..., 60*time.Second)`.

So a provision occupies its consumer for roughly **5× the consumer's own
timeouts**, and the message's offset is not committed until it returns. That is
the shape that produces redelivery, and redelivery here is not idempotent: the
handler's first act on a valid request is to create a new instance.

> ⚠ **[UNVERIFIED] — which broker-side path actually fires.** Session eviction,
> a rebalance the member cannot rejoin in time, or an uncommitted offset on
> restart would all produce these symptoms, and I did not distinguish them.
> The *behaviour* above needs no such caveat — it is counted from the log.
>
> **THE `090` ROUTE WAS TRIED TWICE AND IS NOT AVAILABLE FOR THIS HYPOTHESIS.
> Declaring that explicitly, per the owner ruling of 2026-07-31** (a `bugs_open/`
> file asserting a cross-cutting root cause is not "filed" until it has been
> through the loop, *or the filing session states plainly why it substituted
> equivalent first-hand verification*). Not skipped, and not quietly claimed to
> have confirmed anything:
> - Run 1 — intake `08ba9608-3a3b-4e7a-b4c3-9b3febd784f4`, run
>   `8ee2eb1e-2c1d-4a69-9d1b-505895c4dbcb`: **FAILED** on
>   `API request failed with status 529 … overloaded_error` after 4 retries.
>   Fleet-wide, not about this question (8 such errors across the estate in the
>   hour, 14:08–14:22).
> - Run 2 — run `b930e969-ad53-49fb-923c-2dbaa0ea333b`: work item reached
>   `complete`, **5 bundles, no `iteration_note`, no `council_report`, no
>   decision.** The known no-verdict trap, hit for a reason the documented check
>   does not catch: the bundle's own line is
>   `_(body omitted — 2024 chars, and 58853 of the 60000-char body budget is
>   already spent…)_`. The scope named four symbols across three files, each file
>   well under the 60KB the landmine says to watch for, and they exhausted the
>   shared budget between them. `LANDMINES.md` corrected the same day.
>   **A third run should name ONE symbol** — `handleProvisionInstance` alone, or
>   the consumer's commit path alone — not the set.
>
> **What was substituted, and what it does and does not establish.** All three
> code facts above were read first-hand at the cited line numbers, and the
> delivery counts were taken from the pod log by `correlation_id`, not inferred.
> That is sufficient to establish **that redelivery happens and that each
> delivery creates an instance** — the part the fix must address. It is *not*
> sufficient to name the broker-side trigger, and no claim here does.

**Whoever fixes this:** fix candidate 1 (idempotency on `correlation_id`) does
not depend on which trigger it is, which is a further argument for preferring it
— it is the only candidate that can be written correctly while that question is
still open.

## Why it matters more than its price tag

This sits on the path a **paying customer** triggers. The lane it was found in
sells a fine-tune: a customer pays, we provision a GPU. A duplicate provision is
money spent with no product attached to it, and it is invisible in the places we
would look — see 258 defect 3: a failed provision writes **no** `thunder_instances`
row and **no** `agent_error_log` row, so a duplicate that is later cleaned up
leaves no trace outside rotating pod logs. **We could not currently detect this
happening in production, or reconstruct how often it already has.**

## Fix candidates, ranked by what makes the bad state unrepresentable

1. **Make the create idempotent on `correlation_id` / `request_id`.** Before
   creating, look for an instance already created for this request (a
   `correlation_id` column on `thunder_instances`, written *before* the vendor
   call, or a dedupe table). Then redelivery — from any cause, including ones
   nobody has thought of — is harmless. This is the only candidate that does not
   depend on getting the timing right, and it is the one to prefer.
2. **Stop blocking the consumer.** Return promptly and let a separate worker
   poll for `running` (or drive it from the reaper/monitor, which already exists
   for exactly this class of waiting). Removes the trigger rather than widening
   the window.
3. **Raise the consumer deadlines above the wait timeout** (`KAFKA_SESSION_TIMEOUT`,
   `KAFKA_REBALANCE_TIMEOUT` are env-tunable — no code change). Cheapest, and a
   reasonable *interim*, but it only moves the boundary: any provision slower
   than the new value redelivers again, and it makes every other consumer on
   that config slower to detect a genuinely dead member.
4. Cap attempts per correlation — weakest: it bounds the bleeding without
   stopping it, and needs state anyway, at which point (1) is strictly better.

**Note the interaction with 258.** Fixing 258 defect 2 by *raising* the wait
timeout, without doing something here first, makes this worse: a longer block is
more redelivery, and a box that survives long enough to be inserted is a box the
duplicate does not clean up.

## How to verify a fix

The check must be able to come out badly, so drive a provision that is *slower*
than the consumer deadline — which today is any provision at all:

```bash
# 1. count deliveries per correlation, not creates -- creates can be 400s
POD=$(kubectl -n ai-persona-system get pods -l app=thunder-adapter -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system logs "$POD" --tail=3000 \
  | grep '"action":"provision_instance"' | grep '"msg":"Received request"' \
  | python3 -c "import sys,json,collections; c=collections.Counter(json.loads(l).get('correlation_id') for l in sys.stdin); [print(k,v) for k,v in c.items()]"
# PASS = every correlation delivered exactly once, OR delivered more than once
#        and still only ONE create.

# 2. and independently, at the vendor -- our tables cannot see a box with no row
kubectl -n ai-persona-system exec "$POD" -- sh -c \
  'wget -qO- --header "Authorization: Bearer $THUNDER_COMPUTE_API_KEY" \
   https://api.thundercompute.com:8443/v1/instances/list'
# one request must leave AT MOST one instance here.
```

## Containment, and the evidence that it holds

`is_paused` is checked at the top of `ProvisionAction.Execute`
(`provision_action.go:156`), before any vendor call, so a redelivered request is
denied rather than acted on. Confirmed live rather than assumed: in the 90s
after the pause, **2 further deliveries arrived and were denied, 0 creates were
issued**, and the vendor list went to `{}` once the last box's cleanup ran.

---

## The fix, as shipped (2026-08-12 evening)

**Shape: a claim taken before the vendor call, keyed on `correlation_id`,
in its own table.**

- **Migration `396_thunder_provision_claims.sql`** — `thunder_provision_claims`,
  PK on `correlation_id`, plus `attempts`, `status`
  (`claimed|created|succeeded|failed`), the vendor id, and `last_error`.
- **`internal/adapters/thunder/store/claims.go`** — `TakeProvisionClaim` is a
  single `INSERT ... ON CONFLICT (correlation_id) DO UPDATE SET attempts =
  attempts + 1 RETURNING ..., (xmax = 0) AS inserted`. One statement, so the
  claim and the count cannot race; `xmax = 0` is Postgres's own verdict on
  whether this INSERT inserted, rather than a count we compute and could get
  wrong.
- **`provision_action.go`** — the claim sits between the gate checks and
  `CreateInstance`, and nothing side-effecting may be placed between it and the
  create. A held claim returns `ErrProvisionDuplicate`, which the adapter maps
  to `provision_duplicate` / **`error_unrecoverable`** — deliberately, because
  marking it recoverable would ask the chassis to retry the very thing that
  builds the second GPU.

**Three decisions worth challenging, stated plainly:**

1. **A failed attempt KEEPS its claim.** On 2026-08-12 every attempt failed and
   the cleanup deleted every box; if failure released the claim, the retry loop
   would run exactly as before and this fix would be theatre. The cost is real:
   a genuinely transient vendor failure now ends the workflow instead of
   self-healing. That is the safer side of the trade on a path that spends
   money, but it *is* a trade, and bounded-retry (`attempts <= N`) is the
   obvious next iteration if it bites.
2. **A request with no `correlation_id` is refused, not provisioned.** No key
   means no dedup is possible. Fails closed.
3. **A separate table, not a column on `thunder_instances`.** A pre-create
   claim row in `thunder_instances` would carry no real vendor id, and
   `reconcile_thunder_instances` classifies a live row absent at the vendor as
   a `ghost_row` (`thunder_reconcile_action.go:204-219`) — so every in-flight
   provision would file a spurious orphan-sweep finding against FTW-042, which
   this same lane shipped and got council-approved on 2026-08-09.

**Bonus, not scope creep:** the claims table is the durable record of a failed
provision that **258 defect 3** says does not exist today. `status='failed'` +
`last_error` + the vendor id survive pod-log rotation.

**Regression test** — `provision_idempotency_test.go`,
`TestRetryDriverRedispatchBuildsOnlyOneInstance`. It replays the real shape:
same `correlation_id`, **different `request_id`**, as the retry driver produces.
It asserts on **creates at the vendor**, not on rows or errors, because the
vendor call is the thing that costs money.

**It was proven able to fail.** With the refusal branch mutated out, the test
reports `CreateInstance called 2 times for one logical request` and fails;
restored, it passes. A guard whose test has never been seen red is not a
verified guard.

### Still owed before this can close
1. **Not live.** Needs `make build-thunder-adapter` from committed HEAD and a
   whole-fleet `make release`, **which the owner runs**.
2. **Migration 396 must be applied before the new binary rolls.** The adapter
   treats an absent table as a hard error and refuses to provision — fails
   closed, but it *will* refuse.
3. **Then unpause** (`UPDATE thunder_config SET is_paused=false, pause_reason=NULL;`)
   and verify per "How to verify a fix" above — with the count taken per
   **correlation**, not per request_id.
4. **The `600s` await vs the adapter's `5min` wait is still a real mismatch**
   (258 defect 2). This fix stops the duplicate *billing*; it does not make an
   a6000 provision succeed. Expect the first unpaused a6000 run to fail once,
   cleanly, and leave one claim row marked `failed`.

---

## Council verdict — APPROVED, and what the approving round found

`20d8b725-f4fc-4b8b-ba58-37606ffddacd` — **approved**, 11 reviewers, 6 abstained,
`unreadable: 0`, **`gated_by_truncation: false`** (so this is a real read, not a
silent pass), `decided_by: "approved with 7 advisory objection(s) — none
high-severity"`. Acted on, in order of what they were worth:

1. **guardian, medium — "does `correlation_id` actually survive onto the message
   the adapter consumes?"** The sharpest objection in the round: if it does not,
   the fix refuses **every** provision and bricks the path. **Answered and it
   holds** — `Produce` maps every header onto the Kafka message
   (`platform/kafka/producer.go:76-78,85`), and `ValidateOutgoingMessage`
   **requires a non-empty `correlation_id`** (`platform/validation/validator.go:53`),
   so a request without one is never sent at all. The failure mode is
   structurally unreachable on this path, not merely unlikely.
2. **edit-quality + guardian, medium — the "MUST stay unrecoverable"
   classification had no test.** Real gap, and a nasty one: the idempotency test
   counts vendor creates, so it would keep passing if the refusal were answered
   `error_recoverable` — the claim would still refuse *that* attempt, while the
   chassis went on retrying. **Fixed:** the switch is extracted to
   `classifyProvisionError` and pinned by two tests, including one that wraps the
   duplicate sentinel *together with* `context.DeadlineExceeded` to prove the
   duplicate arm beats the infrastructure arm. Mutation-checked: reordering the
   arms fails the test.
3. **architecture + reuse_agent, medium, raised independently — this defect class
   is generic, and the fix is bespoke.** Both correct, including their point that
   the submission cited no search for an existing generic mechanism. **Spun out
   as `architecture_review/RFC_026`**, carrying the survey the submission should
   have had. That survey also corrected my own first answer: filtering on
   `dispatch_*` showed 3 exposed steps (all Thunder), but the retry driver does
   not care what the action is called — widening to every awaiting step gives
   **54 live `call_agent` steps across 33 agents**, an ~18× understatement.
4. **tooling_provenance, medium — no travelling-docs write-back, and the
   superseded theory is still live in `thunder_config.pause_reason`.** Fair.
   The `LANDMINES.md` entry is synced into `doc_notes` under the
   `coordinator.go` / `awaited_requests` / `*_dispatch.go` footprints, and
   **`pause_reason` has been rewritten** to name the corrected cause and the
   unpause precondition. `is_paused` stays `true`.
5. **guardian, low — a held claim has no operator surface.** Right; it would have
   become a manual DB edit under pressure. **RUNBOOK §5** added: how to see the
   guard firing, how to spot a genuinely stuck claim, how to clear one, and the
   warning not to clear a claim in order to "retry" (re-trigger for a new
   correlation instead).
6. **prior_art_librarian, medium — the `ghost_row` and FTW-042 claims are
   load-bearing but unverified from the seat's position.** Both were read
   first-hand before the design was chosen: `thunder_reconcile_action.go:204-219`
   for the classification, and the migration-number check (`ls` showed 395 as the
   max) that the same seat asked about.
7. **debug_historian, low — recycled vendor identifiers (016 back-catalogue)
   could collide.** They cannot here: `thunder_provision_claims.thunder_instance_id`
   carries **no** unique constraint — the PK is `correlation_id` alone — so a
   recycled vendor id is recorded, never rejected.

**Still not acted on, deliberately:** edit-quality's success-path race (what the
audit trail says when a *successful* first attempt's response lands after the
coordinator has already moved on). It is real, it needs 258 defect 2 fixed to be
observable, and it is recorded in RFC_026 §6 alongside the undiagnosed
error-response-does-not-clear-the-await co-cause.

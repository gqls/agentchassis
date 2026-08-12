# 259 — one provision request builds several billable GPUs (Kafka redelivers while the handler blocks)

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

## Why — the code facts

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

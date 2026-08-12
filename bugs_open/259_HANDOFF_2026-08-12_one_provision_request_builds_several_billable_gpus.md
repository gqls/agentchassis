# 259 — one provision request builds several billable GPUs (Kafka redelivers while the handler blocks)

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
> A `090` diagnosis run was filed 2026-08-12 to settle it —
> intake `08ba9608-3a3b-4e7a-b4c3-9b3febd784f4`,
> **run correlation `8ee2eb1e-2c1d-4a69-9d1b-505895c4dbcb`** (that is the key
> the artifacts are written under). Read its verdict before asserting a
> mechanism in a fix. The *behaviour* above needs no such caveat — it is
> counted from the log.

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

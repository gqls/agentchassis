# RFC_026 — the retry driver re-executes side-effecting actions, and idempotency is per-adapter bespoke

**Raised** 2026-08-12 by the `finetuning_uk_service` lane, out of the council
round on `bugs_open/259_…_one_provision_request_builds_several_billable_gpus`
(submission `20d8b725-f4fc-4b8b-ba58-37606ffddacd`, **APPROVED** with 7 advisory
objections). **Two seats raised this independently** — `architecture` and
`reuse_agent`, both medium — which is why it is spun out rather than answered in
the bug.

**Status: OPEN — needs an owner ruling.** The Thunder fix it came from is
committed (`10659b419`) and approved on its own terms; nothing here blocks it.

---

## 1. The mechanism, in plain terms

When a workflow step asks another component for something and waits, it records
an `awaited_requests` row with a deadline (`timeout_seconds`). If the answer does
not arrive by then, the chassis does **not** give up. `retryExpiredAwaitedRequest`
(`platform/orchestration/coordinator.go:3321`, budget `RetryVersion < 3` — so up
to **four** executions) **re-runs the step**.

Re-running the step re-runs its dispatch action, which mints a **fresh
`request_id`** and publishes a **new message**. The downstream component has no
way to tell the fourth ask from the first: everything differs except
`correlation_id`.

**For an idempotent action this is exactly right** — it is a retry, and it is the
mechanism that makes per-workflow timeouts survive pod restarts (`bugs_open/003`
F2). **For a side-effecting action it is a multiplier.** On 2026-08-12 it turned
one request for a GPU into four billable GPUs.

## 2. Why this is not simply "Thunder's bug"

The 259 fix is a bespoke, Thunder-local claims table plus a Thunder-local claim
function. The architecture seat's objection, verbatim in substance: *any* other
adapter dispatched the same way has the identical latent defect, and building it
per-adapter means "we will re-derive this five times" — a new table, a
claim-before-create insert, header plumbing and an unrecoverable error-code
mapping, each with its own council round to catch the same defect.

The `reuse_agent` seat made the same point from the reuse side and added a
second: `INSERT … ON CONFLICT … RETURNING (xmax = 0)` is a reusable
idempotent-claim idiom that, if it does not already exist generically, probably
belongs somewhere shared rather than in one adapter's store package.

**Both seats also flagged that the 259 submission cited no evidence of having
searched for an existing generic mechanism before building a new table. That
criticism is accurate.** This RFC supplies the survey that submission should
have carried.

## 3. The survey — measured 2026-08-12, and it corrected my own first answer

**First cut, and why it was wrong.** I asked "which `dispatch_*` steps set a
`timeout_seconds`?" — three, all Thunder. Read alone, that says Thunder is the
whole exposure and this RFC is unnecessary.

**That filter described a small world.** The retry driver does not care what the
action is called; it re-executes whatever the step does. Widening to *every*
live step carrying a `timeout_seconds`:

```sql
WITH steps AS (
  SELECT ad.type AS agent_type, s.key AS step_name, s.value->>'action' AS action,
         (s.value->'config'->>'timeout_seconds')::int AS timeout_s
  FROM agent_definitions ad,
       LATERAL jsonb_each(ad.default_config->'workflow'->'steps') AS s
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
)
SELECT action, count(*) AS live_steps, count(DISTINCT agent_type) AS agents,
       min(timeout_s)||'s-'||max(timeout_s)||'s' AS timeouts
FROM steps WHERE timeout_s IS NOT NULL GROUP BY action ORDER BY 2 DESC;
```

| action | live steps | agents | timeout range |
|---|---|---|---|
| **`call_agent`** | **54** | **33** | 60s – 43200s |
| `diagnose_build_gate` | 3 | 2 | 600s – 900s |
| `request_human_input` | 3 | 2 | 3600s – 86400s |
| `batch_webscrape` | 2 | 2 | 240s – 240s |
| `dispatch_thunder_decommission` | 1 | 1 | 90s |
| `await_approval` | 1 | 1 | 300s |
| `dispatch_thunder_list` | 1 | 1 | 60s |
| `dispatch_thunder_provision` | 1 | 1 | 600s |

**The narrow query understated the exposure roughly 18-fold.** `call_agent` is
the shape that matters: 54 live steps across 33 agents, each re-executed up to
four times on expiry.

## 4. What the survey does NOT establish — the open question

**A re-executed step is only damaging if what it drives is side-effecting.** A
re-run LLM call is waste; a re-run deploy, email, payment, git commit or GPU
provision is damage. I have **not** classified the 54 `call_agent` steps by
side-effect-ness — that is the substance of the work this RFC asks for, and
asserting a number without it would be exactly the composition error that
produced 259's wrong root cause in the first place.

What is already known to be in the damaging class:
- **GPU provisioning** — measured, four boxes, `bugs_open/259`. Now guarded.
- **Operator actions that email a human** — "free the slot" and "tell the
  customer" are one button (`LANDMINES.md`); a re-executed step emails twice.
- **Site deploys / rerenders** — a duplicate is usually benign but not free.

The short-timeout end is where the risk concentrates: a 60s deadline on a step
whose downstream can take minutes will expire *routinely*, not exceptionally.
259's own step was 600s against a vendor call the adapter allowed 5 minutes for —
a mismatch of exactly this kind, and it fired on the first real use.

## 5. Options

1. **Do nothing; fix per-adapter as it bites.** Cheapest today. Cost: every
   future side-effecting action needs its own claims table and its own council
   round to discover the same defect — and the discovery is expensive because
   the symptom (duplicate work) appears far from the cause (a timeout).
2. **A generic claim primitive.** A shared `correlation_id`-keyed claim helper
   (and one table, or a column on `awaited_requests`) that any action can take
   before a side effect. Cost: it is a new shared seam on the busiest path in the
   estate — architecture-scope by the 2026-07-28 ruling, and it must not slow the
   idempotent majority.
3. **Make the retry driver itself refuse to re-execute a side-effecting step**,
   by declaring side-effect-ness on the action (a `RegisterActionInputSpec`-style
   flag, unsafe default OFF per the 2026-08-02 ruling §2). Attacks the cause
   rather than each symptom: an action that says "I am not safe to re-execute"
   would fail the step instead of duplicating the effect. Cost: a fleet-wide
   audit to set the flag, and the flag is only as good as its accuracy — plus
   note **RFC_022's open counter**, which exists precisely because accumulated
   opt-in flags stop being individually reviewable.
4. **Fix the co-cause instead** (see §6). Fewer expiries means fewer
   re-executions, but it bounds the bleeding rather than stopping it — the
   retry driver would still multiply any effect whose downstream is genuinely slow.

**My recommendation: (3), with the classification survey of §4 as the first
step**, because it is the only option where a new side-effecting action is safe
by default rather than safe if someone remembers. But (3) is exactly the shape
RFC_022 has already been burned by, so it needs the owner, not a lane.

## 6. The co-cause, which is a separate defect and still undiagnosed

The awaits in 259 were **never cleared by a response**. The adapter answered
~5 minutes in (`Sent error response`) and every row was still `processed` at
*exactly* its own `timeout_at`, ten minutes in. Had the error response cleared
the await, the step would have failed once and built **one** box.

So the retry loop only ran because an error response did not do its job. That is
its own bug, it is not diagnosed, and it plausibly affects every adapter's error
path — meaning the "how often do awaits actually expire?" figure that option (4)
depends on is currently inflated by it. **Worth a `090` of its own; it is the
single highest-value unknown here.**

## 7. What this RFC is not

Not a criticism of the 259 fix, which the same council approved and which the
`constitution` seat specifically credited for retiring a fix candidate that would
have widened blast radius. A point fix on a money-losing path was the right move
today. This is the question of whether the *fifth* one should still be bespoke.

## Evidence
- `platform/orchestration/coordinator.go:3321` `retryExpiredAwaitedRequest`; `:3351`/`:3382` the `RetryVersion < 3` budget; `:4112` the `RETRY_TICKER` claim loop.
- `platform/orchestration/actions/thunder_provision_dispatch.go:99` — `uuid.NewString()` per execution; `:134` — `correlation_id` from the run.
- `awaited_requests`, correlation `23c9bc6a`: 4 rows, 4 distinct `request_id`s, 1 `orchestration_id`, each `processed_at` == its own `timeout_at`, next `sent_at` ~1s later.
- The §3 survey, run 2026-08-12 against live `agent_definitions`.
- Council `20d8b725-f4fc-4b8b-ba58-37606ffddacd`: `architecture` and `reuse_agent` objections, both medium, verbatim in `diagnosis_artifacts`.

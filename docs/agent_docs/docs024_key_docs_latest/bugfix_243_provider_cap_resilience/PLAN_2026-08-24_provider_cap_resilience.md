# PLAN 2026-08-24 — what the platform should do when the provider refuses us

Subject: `bugs_open/243…anthropic_account_usage_limit_reached…` (**243-anthropic-cap**).
Lane opened 2026-08-24. Evidence and citations: `NOTES_provider_cap_resilience.md`.

## 0. Scope, and what is explicitly NOT in it

The cap itself is an account/billing state and the owner's to clear. **Nothing here tries to
fix, retry into, or route around the cap**, and nothing here widens the transient classifier to
swallow it — the bug file's §5 candidate 4 forbids exactly that and is right. This plan is about
what our own platform does *while* the provider is refusing us.

## 1. The defect, at framework level

**The platform detects "the provider is refusing us" exactly once, and then throws the fact
away at every layer that needs it.**

`isAIUnavailable` (`ai_errors.go`) classifies the condition correctly. From there:

- the **error type** (`*AIUnavailableError`) is flattened to a string by `%v` before any routing
  decision is made (`coordinator.go:945`, `:3634`), so no consumer downstream of the step can
  tell a provider refusal from bad model output;
- the **only durable signal** is one boolean in `ai_endpoint_health`, and it is **write-only in
  one direction from live traffic** — the sole caller of `update_endpoint_health` only ever
  passes `false` (`ai_actions.go:634`). Nothing that succeeds can clear it.

So the fact is available at the point of failure and nowhere afterwards. Every consumer either
re-derives it from a string or acts on a flag whose *recovery* has nothing to do with the
*condition* that set it. Two consumers consequently act at the wrong altitude:

| consumer | altitude | cost of one refused call |
|---|---|---|
| `claim_work_item` health gate | too coarse — one boolean gates **every** work item fleet-wide | up to **60 min** of total dispatch loss (measured 60m25s, 08-17) |
| council-gate seats' `error_step` | too fine — one seat's error ends the **whole round** | ~a coin-flip per round (measured 08-19); every answered seat's spend discarded |

Both are the same missing property: **there is no first-class representation of the condition
for a consumer to act on proportionately.** Fix that, and both consumers can be made
proportionate without either of them re-deriving anything.

## 2. Fix candidates, ordered by what closes the door

Ranked by what makes the bad state *unrepresentable*, not by effort. Go changes are inert until
a chassis roll; DB config is live immediately. **Where both halves exist, the Go half ships
first** — CLAUDE.md's "image first, then seeds", and here it is load-bearing, not ceremony
(see C3/C4).

---

### C1 — make live traffic a SYMMETRIC writer of endpoint health `[Go]`

**The door it closes:** today the flag can outlive the condition indefinitely, because the
thing that sets it (live traffic) is not the thing that clears it (an hourly timer). After
this, a healthy endpoint cannot stay marked unhealthy while calls are succeeding — the state
"marked down while live traffic succeeds" becomes unrepresentable rather than merely unlikely.
That is the exact state measured on 08-17 for 60 minutes.

**Change:** in `ai_actions.go`, alongside the existing failure goroutine at `:626-637`, add the
mirror on the success path — a conditional clear:

```sql
UPDATE ai_endpoint_health SET healthy = true, last_healthy = NOW(), error = NULL, updated_at = NOW()
 WHERE endpoint_url = $1 AND NOT healthy
```

**Why it is cheap on the hot path:** the `AND NOT healthy` predicate matches **zero rows** in
the normal case, on the primary key. It is one indexed no-op update per successful call, fired
in a goroutine exactly as the existing failure write already is. It is deliberately *not* a
blanket `SET healthy=true` — that would rewrite `last_healthy` on every call.

**How it could be wrong** (state it before building): under a genuinely sustained cap the flag
will **flap** — down on a refusal, up on the next success, down again. Two answers: (a) that is
*correct*, because it is what is actually happening, and it is strictly more truthful than a
flag pinned false for an hour; (b) flapping is bounded by the traffic rate, and the consumer
(`claim_work_item`) re-reads per claim, so a flap costs at most one released claim, not an
hour. If the owner would rather have hysteresis, that is C1b below — but hysteresis on top of a
*correct* signal is a policy choice, whereas today's behaviour is a defect.

**Proof, with a control that could come out otherwise:**
1. Force the row false by hand: `SELECT update_endpoint_health('https://api.anthropic.com/v1/messages', false, 'manual test');`
2. Confirm it is false, then drive **one** successful LLM call (any agent).
3. **PASS** = the row reads `healthy=t` within seconds of that call.
4. **NEGATIVE CONTROL, and it is the one that matters:** repeat step 1 and drive **no** traffic.
   The row must stay `false` until the probe's next tick. Without this control the test passes
   on the pre-fix binary too, because the hourly probe would clear it eventually anyway — so
   run the positive arm *inside* one probe interval and record the timestamps.

---

### C2 — shorten the claude re-probe interval `[DB config — LIVE IMMEDIATELY, no roll]`

**The door it closes:** none, on its own — it bounds the damage rather than removing the
mechanism. It earns its place *above* the council fix because it is the **only** item here that
helps before the next chassis roll, and the monthly-limit shape predicts recurrence before
month-end.

**Change:** `UPDATE ai_endpoint_health SET check_interval_seconds = 60 WHERE name = 'claude';`
(3600 → 60). Worst-case dispatch stop falls from ~60 min to ~1 min.

**Why 60 is not a guess:** it is the value already carried by the `cpu-ollama` row in the same
table, so it is an established setting for this mechanism rather than a number I picked.

**Cost:** 1,440 probe calls/day at `max_tokens: 1` on haiku, versus 24 today. Negligible against
~750–1,200 real calls/day, and it must be stated rather than assumed — **[UNMEASURED]** as a
billing figure; if the owner wants it measured first, 300s still cuts the worst case 12×.

**⚠ What C2 does NOT do, and the reason it is not the whole answer:** the probe **cannot detect
this condition at all**. `pingClaude`'s status switch returns `true` for any non-auth status
(`check_endpoint_health_action.go:220-231`), and the cap is a **400**. So the probe is a *timer
that clears the flag*, not a health check — it will clear the flag on its next tick whether or
not the provider is still refusing us. C2 therefore shortens a blind timer. It is worth doing
for that alone, and it is why C1 (which uses a signal that *can* see the condition — a real
successful call) ranks above it.

---

### C3 — record per-step failures durably, so a consumer can tell WHICH step failed `[Go]`

**The door it closes:** `routeToErrorStep` writes `__step_error` as a **single key that is
overwritten** (`coordinator.go:3956-3959`), so in a workflow where two steps fail, the first is
erased. Any consumer wanting to know "did *this* step fail?" cannot ask. This is the framework
half of the council fix and is deliberately general — every `error_step` consumer gains it, not
just the council.

**Change:** in `routeToErrorStep`, **in addition to** the existing `__step_error` (untouched, so
every current reader is unaffected), accumulate a per-step record:

```go
// alongside state.CollectedData["__step_error"] = …
errs, _ := state.CollectedData["__step_errors"].(map[string]interface{})
if errs == nil { errs = map[string]interface{}{} }
errs[failedStepName] = map[string]interface{}{"message": errorMsg, "at": time.Now().UTC()}
state.CollectedData["__step_errors"] = errs
```

**Reuse check, done rather than asserted:** `ProcessingHistory` already accumulates an
`error_routed` record per failure (`coordinator.go:3970-3976`), so my first instinct was to read
that instead of adding a key. **It is not reachable**: `ActionParams` (`actions/types.go:45-69`)
carries `CollectedData` and `WorkflowSteps` but **not** `ProcessingHistory`. Adding it there is
the alternative shape — and there is precedent, since `WorkflowSteps` was added to
`ActionParams` for `bugs_open/076` for exactly this "the action cannot know this from
`StepConfig` alone" reason. I prefer `__step_errors` because `collected_data` is **persisted and
queryable**, which the audit trail needs anyway (C5), whereas `ProcessingHistory` would be
in-memory at the point of use. `__`-prefixed reserved keys are the established idiom here
(`__step_error`, `__truncated`).

**Scope classification** (see §3): additive and **inert** — nothing reads `__step_errors` until
C4's consumer lands.

---

### C4 — a council seat that ERRORS costs one seat, not the round `[Go half in C3 + DB config]`

**The door it closes:** the asymmetry where garbage is survivable and a transient is fatal.
`diagnose_council_decide` already handles a seat that produced nothing — it is an
**abstention** — and already handles a seat whose opinion was lost — **unreadable**, which at
`:460` downgrades an approval to REVISE. The machinery exists; the errored seat simply never
reaches it, because `error_step` sends the whole round to `complete_invalid` first.

**Two halves, and the ORDER IS LOAD-BEARING:**

**(a) Go — classify correctly** (`diagnose_council_decide_action.go`, the `raw == nil` branch at
`:310-325`): when a seat's field is absent **and** `__step_errors` names that seat's step,
record it as **`unreadable`**, not `abstained`.

> **This is the whole reason C4 is not a one-line config change, and I want it on the record
> because it is the objection I would raise against my own fix.** The obvious fix — just point
> `error_step` at the next seat — makes the failed seat's field absent, and an absent field is
> counted as an **abstention**. The code's own comment at `:311-318` says precisely why that is
> wrong: *"An abstention is a seat the relevance filter skipped, which is information ('not
> applicable'); an unreadable seat is an opinion we were owed and lost… Conflating them would
> let a lost opinion read as a considered non-objection."* Shipping the config half alone would
> trade "the round dies" for "the round can APPROVE with a seat we never heard from, silently" —
> which is worse, because the first failure is loud and the second is not.

**(b) DB config — stop discarding the round:** repoint each of the **17** `review_*` seats'
`config.error_step` from `complete_invalid` to **that seat's own `next_step`**, so a failed seat
is skipped and the chain continues. The values are already in the workflow (e.g.
`review_editquality` → `review_constitution`; `review_guardian` → `council_decide`).

**Precedent, in this same workflow:** `run_checks` already carries
`error_step: compose_verdict` — tolerate-and-continue is not a new pattern for this agent.

**Two steps that must KEEP `complete_invalid`** — do not sweep them up in a bulk update:
`persist_submission` (if the submission cannot be persisted there is nothing to review) and
`council_decide` itself (if aggregation fails there is no verdict).

**⚠ Config trap, twice-confirmed:** `error_step` sits inside `config`, **not** at the step
level. Read at step level the census returns "(none) | 29" — a clean, confident, wrong answer.

**Ordering:** **(a) must be live before (b) is applied.** If (b) lands first, every capped seat
reads as a considered non-objection for as long as the gap lasts. This is a real ordering
constraint, not a convenience — name it in the commit.

**Proof:** unit-pin both arms in (a) — a seat absent *with* an `__step_errors` entry must land
in `unreadable`; a seat absent *without* one must land in `abstained`. Then, live: a round in
which one seat errors must reach a verdict, report that seat under `unreadable`, and — if the
remaining seats would have approved — come back **REVISE**, naming the lost seat. A round that
approves with zero unreadable is the negative control.

---

### C5 — make the condition VISIBLE `[owner decision on channel]`

Seam C: five-plus occurrences and **every one** was found by a lane chasing an unrelated
symptom. There is no alert, no work item, no named condition surfaced anywhere; the only
durable trace of an hour-long fleet stop is `claim_result.reason` inside `collected_data`.

C1 and C3 make this cheap, because after them the condition is finally *represented*. Reuse
before recreate: `record_vision_finding` (TL-041) already files exactly one deduped work item
via an `ON CONFLICT` arbiter — the same shape fits here with
`item_key = provider_refusal:<endpoint>:<date>`. **Owner's call on the channel** (work item vs
`doc_notes` vs neither); I am not choosing it unilaterally, and it is the one item here whose
failure mode is nagging rather than silence.

---

### C6 — a second provider `[NOT PROPOSED NOW]`

The bug file's §5 candidate 2. **127 of 127 configured LLM steps across 55 live agents name
`anthropic` and the same key** — so this is a real build, not a config flip. It remains the only
candidate that addresses *single-provider* risk rather than our reaction to it, and it needs an
owner decision and its own architecture round. Out of scope here, deliberately, and noted so it
is not mistaken for forgotten.

## 3. Scope classification against CLAUDE.md's rulings

| candidate | classification | justification |
|---|---|---|
| **C1** | **council gate, and TELL the consumers** | It changes a shared mechanism's *timing* guarantee (recovery no longer waits for a timer). Under the 2026-07-29 ruling #1 the trigger is whether the shared mechanism's **guarantee** changes — the flag's *meaning* is unchanged ("is this endpoint serving?"), while its accuracy improves, so it is not architecture-scope. Ruling #3 applies squarely and is **discharged, by enumeration rather than assertion** `[MEASURED 2026-08-24]`: `grep -rn "ai_endpoint_health" --include=*.go platform/ internal/ pkg/ cmd/` returns **exactly two** consumers — `claim_work_item_action.go:235` (the fleet-wide claim gate, the reader whose behaviour changes: it sees a true recovery in seconds instead of up to an hour) and `check_endpoint_health_action.go` (the probe, **unaffected** — it still writes on its own tick). The only writer is `update_endpoint_health`, called once, from `ai_actions.go:634`. There is no third consumer to tell. |
| **C2** | no gate needed | A data value in an operational table, reversible in one statement. Not code, not a seam. |
| **C3** | **council gate** | It adds a reserved key to shared state, which the 2026-07-28 ruling names as architecture-shaped — but the 2026-07-29 narrowing is decisive: it is **additive and inert**, reachable by nothing until C4(a) reads it, and it changes no existing reader (`__step_error` is untouched). RFC_022's three conditions also hold (opt-in in effect; the absent side is the default; zero live consumers name it) — and per RFC_022 that must be **enumerated in the submission, not asserted**. Register entry owed **in the same commit** (2026-07-28 condition (2), which survived the 07-29 retirement of condition (1)). |
| **C4(a)** | **council gate** | Changes how the council classifies a seat — behaviour of an existing mechanism, single consumer, no new seam. |
| **C4(b)** | **council gate** (config, but submit it) | Materially changes council behaviour. Config, so live immediately — which is exactly why it must not go in unreviewed. |
| **C5** | **owner decision first** | Channel choice, then normal gate. |
| **C6** | **architecture round** | New shared capability across 55 agents. Not now. |

**Not counted against the RFC_022 optional-key budget (N=10, WFA-013) — CONFIRMED, not asserted
`[MEASURED 2026-08-24]`.** Ran `scripts/audit-optional-key-budget.sh`: it counts *optional keys
per action **input spec***. `__step_errors` is a `collected_data` key written by the coordinator,
never an action input key, so it cannot enter that count. And C4(a) adds **no** config key —
`diagnose_council_decide`'s spec stays at its current three (`review_fields`, `hard_veto_from`,
`max_rounds`); the classification reads `__step_errors` out of `CollectedData`. So no action's
optional surface moves, and no acknowledgement in
`architecture_review/optional_key_budget_acks.json` is owed.
⚠ The script's first run failed on a truncated `kubectl exec` and **said so** rather than
reporting a small fleet — re-run it if it prints nothing; a short read there exits 0.

## 4. What I am deliberately NOT doing

- **Not touching `check_endpoint_health_action.go`.** It is **dirty in the shared tree right
  now** (another session adding `CheckConfig: true` + a doc comment to its InputSpec) — the same
  reason the 08-17 addendum declined, still true seven days later. A pathspec commit takes a
  same-file passenger, and `git stash` is banned and hook-blocked. C1 lives in `ai_actions.go`
  and C2 is a data change, so **this plan needs no edit to that file at all** — which is a
  reason to prefer this shape, not merely a constraint I worked around.
- **Not implementing the 08-17 addendum's "require N consecutive probe failures".** As written
  it hardens the probe, and the probe is **not** the writer of `false` — live traffic is, and
  `pingClaude` cannot even see a 400. Building it as specified would harden a path that never
  fires for this condition. If hysteresis is wanted it belongs at `ai_actions.go`, as C1b.
- **Not making the cap retryable**, not widening the transient classifier, not adding retries.
- **Not doing more prompt-caching work.** `bugs_open/244` is fixed and live; the measurement
  showing it did not prevent recurrence is in NOTES, and that lane is done.
- **Not letting `claim_work_item` through under a cap** (the addendum's third shape). It is
  arguable — one failed step versus a stopped queue — but it is a *policy* change to a
  fleet-wide gate, and C1 removes the need for it by making the flag accurate. Revisit only if
  C1 proves insufficient.

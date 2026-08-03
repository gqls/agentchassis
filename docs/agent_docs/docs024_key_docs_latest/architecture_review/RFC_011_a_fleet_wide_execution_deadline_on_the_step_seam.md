# RFC 011 — a fleet-wide execution deadline on the coordinator's step-execution seam

## STATUS: **DECIDED — OWNER RULING 2026-08-03: option (a). Every step gets a limit, and the default must be very generous so we never cut off something still running.**

The ruling changed the constant, and the re-measurement it forced found a real defect in
what had shipped. **600s would have killed `med_scrape_prices` on every run** — a LOCAL
action that legitimately takes ~1,391s (23 minutes). The original justification (~25x the
all-step p99.9) was arithmetically correct and methodologically wrong: a percentile over
91,210 mostly-sub-second executions cannot see the one action that runs for twenty
minutes. **Sizing now comes from the longest observed legitimate action, not a
percentile.** Full account: `WRONG_CALLS.md` 2026-08-03.

What is live now:

| | value |
|---|---|
| default | **7200s** (~5.2x the 1,391s ceiling) |
| pinned by | `TestTheDefaultClearsTheLongestLegitimateLocalAction` (>=4x headroom, mutation-tested) |
| early warning | an action consuming >50% of its budget is logged **while still succeeding**, so the ceiling moving is heard about before anything is cut |
| escape hatches | per-step `local_action_timeout_seconds`; `<=0` disables; `DISABLE_LOCAL_ACTION_TIMEOUT=true` fleet-wide |

Options (b) and (c) below are **not taken** and are kept for the record.

> **Original status line:** RAISED 2026-08-02 by the bugfix_169 lane. Needs an owner decision.

Raised because the council's `architecture` seat objected on record, on a submission it
otherwise **approved** (`2c6800e6`, *"approved with 2 advisory objection(s) — none
high-severity"*), and because its objection is right about the process even though I
still believe the change itself is correct:

> *"Adds a new reserved step-config key (`local_action_timeout_seconds`) and a new
> fleet-wide env contract (`DISABLE_LOCAL_ACTION_TIMEOUT`) on the coordinator's
> step-execution seam — the single dispatch path every local action in every package
> runs through. Per the 124/129 precedent, a new reserved key + new fleet-wide contract
> on a shared mechanism is architecture-scope even when additive and well tested."*
>
> *"This changes what the coordinator GUARANTEES to every action handler fleet-wide —
> from 'may run forever' to 'cancelled at N seconds'. **The author's own risks section
> argues against RFC_010 applying here; that argument itself belongs in an RFC where it
> can be checked against blast radius and rollback, not asserted inline in a bug-fix
> rationale.**"*

The `guardian` seat objected independently on the same surface (blast radius is "every
pipeline that dispatches a local action"; and a stability preference for `coordinator.go`
specifically). **Two seats, from different directions, on a patch neither opposed** —
the same signal that made RFC 006 an RFC rather than a note.

Filed under the 2026-07-29 owner ruling: *review here is after the fact, by design*. The
patch is committed (`fe34fd04f`) because HEAD is shared and holding code back is not a
capability this tree has. **This RFC is about the seam, not about whether the bug is
real.**

---

## 1. What shipped

`bugs_open/169` part A: nothing in
`continueExecution → executeStep → executeLocalAction → executeAction` bounded the
handler call, so an action blocking on a network call parked its orchestration at
`EXECUTING_STEP` and held the dispatch loop's work item `claimed` until a reaper expired
it. `timeout_seconds` looks like the bound and is not — it is parsed into
`execCtx.TimeoutSeconds` *"for the action to set a default"*, and no action reads it that
way (checked across all 271 action files).

The fix derives a deadline in `executeLocalAction`:

| lever | value |
|---|---|
| default | ~~**600s**~~ → **7200s** (superseded by the 2026-08-03 ruling above; 600s would have cut `med_scrape_prices`) |
| per-step override | `local_action_timeout_seconds` (new reserved key) |
| per-step disable | that key `<= 0`, logged Warn every time |
| fleet-wide disable | `DISABLE_LOCAL_ACTION_TIMEOUT=true` (new env contract) |

Measured before choosing, over `orchestration_state_audit` 2026-07-31 → 08-02:
**6,951** spawn-step executions, p50 0s, p95 18s, p99 **24s**, and **exactly one** above
300s — at 14,475s. All-step p99 26s, p99.9 214s. The distribution is bimodal with nothing
between seconds and hours.

## 2. The question for the owner

**Should the coordinator impose an execution deadline on every local action by default?**

**(a) Yes, default ON at a generous constant — what shipped.**
*For:* the defect is fleet-wide and silent, and a generous bound catches the pathological
case while touching nothing healthy. Three escape hatches exist, one of which needs no
rebuild.
> **CORRECTED 2026-08-03:** this bullet originally read *"the measurement says a 600s bound
> is ~25× the p99.9 — it catches the pathological case and touches nothing healthy"*. The
> second clause was **false**: 600s would have cut `med_scrape_prices` (~1,391s) on every
> run. The percentile could not see it. This is exactly the "Against" below coming true
> before the ruling, not after.
*Against:* it changes a guarantee for every action in every package at once, and the
constant was measured over **three days of a partly-idle fleet** (`improvement-sweep` and
other schedulers are disabled). If a legitimate long-running local action exists that
simply did not run in that window, it now fails.
*Residual:* the constant needs re-measuring when the disabled schedulers return.

**(b) Yes, but default OFF — opt in per step.**
*For:* this is what the 2026-08-02 RFC_010 ruling asks for in the general case — *"new
authority on a shared seam ships as an OPT-IN FIELD, not a documented contract"*, unsafe
default OFF.
*Against, and this is the argument I made inline and which the seat correctly says belongs
here:* RFC_010 governs a seam that **grants** authority on an assumption about callers.
This one **removes** authority and is licensed by measurement, not assumption. And a
protection that must be enabled per step protects nothing until someone enables it —
which is precisely the inert-by-omission defect **RFC 006** was decided on *this same
day*. Shipping (b) would mean re-learning that lesson on a different seam.
*Residual:* the whole of it, until adoption.

**(c) Yes, but not here — bound it at a higher layer.**
The `guardian` seat's stability preference for `coordinator.go` points this way: e.g. a
watchdog that fails an orchestration whose step has not advanced, rather than cancelling
the action's context.
*For:* leaves the core dispatch path untouched.
*Against:* a watchdog outside the call cannot cancel the blocked work, only mark the row —
which is exactly what `coordinator.go:831`'s existing `maxAge = TimeoutSeconds × 3` guard
already does, and it demonstrably did **not** solve this: it fires only when a message
next arrives and never interrupts the goroutine. Five orchestrations in the window each
sat ~4h01m before it noticed.
*Residual:* the hang itself, unfixed.

## 3. What was checked, so the decision does not rest on my say-so

Every checkable objection from the round was run:

- **`reuse_agent` / `prior_art_librarian`** — *is there already a context-deadline wrapper
  for another call path that this should extend rather than duplicate?* **No.** The only
  `context.WithTimeout` calls in the package are `helpers.go`'s fixed 5s/10s query
  timeouts and one 60s at `coordinator.go:3972` — **all built on `context.Background()`**,
  i.e. deliberately detached, a different pattern with a different purpose. There is no
  step-execution deadline helper to extend.
- **`guardian` (low)** — *does an existing test encode a no-timeout assumption that this
  silently breaks?* **No.** The only test file in the package referencing
  `executeLocalAction`/`executeAction` is the one this change adds.
- **`guardian` (medium) / `architecture` (missing)** — *has this defect class been raised
  and deflected upward before?* **No.** The only mention of `executeLocalAction` anywhere
  in `bugs_open/`, `bugs_closed/` or `architecture_review/` is `bugs_open/169` itself.
- **`debug_historian` (medium)** — *no deploy-verification step named.* Fair against the
  submission; the bug file already carries one (pod-grep both replicas for
  `local_action_timeout_seconds` with a control string in the same exec, per
  `bugs_open/153`). The seat's second half is a real operational warning worth repeating
  here: **a chassis roll kills an in-flight council run**, so whoever rolls this should
  check for one first.

## 4. What is NOT being asked

Not whether `bugs_open/169` part A is real — the mechanism was read end to end and the
class measured. Not a re-run with better numbers: the `architecture` seat's point is about
**how a fleet-wide guarantee change reached production**, and per the 2026-07-28 ruling a
scope objection is not answered by better measurements.

## 5. Related

- `bugs_open/169` — the instance and the patch; register **RSH-004**.
- **RFC 006** (decided 2026-08-02) — the inert-by-omission argument that option (b) runs
  into, decided the same day on a different seam.
- **RFC_010** — the "opt-in field, unsafe default OFF" ruling this change argues around.
  ⚠ **two files are numbered `RFC_010`** (`_discovery_checks_can_raise_a_finding_but_not_retract_one`
  and `_who_may_answer_a_page_name_collision`), and likewise two `RFC_009` — so cite these
  by **slug**, never by number.
- `bugs_closed/124`, `bugs_closed/129` — the precedent the seat cited: a reserved-key
  addition to a shared mechanism arriving inside a bug patch.
- `coordinator.go:831` — the existing `maxAge = TimeoutSeconds × 3` orchestration-age
  guard, which is what option (c) would extend and which already fails to solve this.

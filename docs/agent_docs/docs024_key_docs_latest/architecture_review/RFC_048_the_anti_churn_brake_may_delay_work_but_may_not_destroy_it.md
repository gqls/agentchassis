# RFC_048 — should the anti-churn brake be allowed to destroy a work request, and if not, what should stop it?

**Raised 2026-08-23** by the `bugs_open/326` fix lane, **because the council gate REJECTED the
change on a guardian hard veto and named this route as the safest contained alternative.**
Council corr `f610741f-5054-41e8-b0b7-54915d79ba92`, round 1.

> `guardian` (veto, high): *"This rewrites the default behaviour of `writeWorkItem` — the shared
> core of `create_work_item` used by every keyed work-item producer fleet-wide. The change is NOT
> opt-in per caller; it flips default behaviour for everyone and only offers a global env-var kill
> switch… That is exactly the shared-seam shape the landmines record an owner ruling (2026-08-02
> §2) requiring to ship opt-in-per-caller with the unsafe side OFF by default."*
>
> `guardian` (veto, high): *"The customer-facing bug is already fully closed by edit 4 alone — a
> config-only, live-on-apply, correctly-scoped migration… Edit 1 is therefore not required to fix
> the filed bug; it is a separate, fleet-wide architecture decision bundled into an urgent
> point-fix submission. That is the veto-criterion pattern: architecture change dressed as a point
> fix."*
>
> Safest contained alternative, verbatim: *"land 3/4/5/6 now to close bugs_open/326 today; split
> edit 1+2 into their own submission scoped as an explicit architecture change, with the deferral
> behaviour opt-in per `item_key_prefix` or `agent_type` (mirroring `recurrence_expected`'s own
> opt-in shape) rather than a global env-var flip, and let a full council round weigh the
> fleet-wide dispatch-volume and dashboard-visibility consequences on their own merits rather than
> under bug-fix urgency."*

**I accept the veto and I am not contesting it.** CLAUDE.md is explicit that *"a veto on SCOPE is
not answered by resubmitting with better measurements"* — and this one is squarely on scope. The
guardian's central factual claim is also simply correct and I should have seen it: **migration 572
alone closes `bugs_open/326`.** Everything in this RFC is a separate question that I bundled into
an urgent fix because I had the evidence in front of me, which is precisely the pattern the
veto criterion exists to catch.

**This RFC asks for a decision it cannot make itself.** The code is written, tested and
mutation-proven; it is **not committed**, and it will not be committed until this is ruled on.
The patch is preserved beside this file rather than left dirty in the shared tree.

---

## 1. What is actually wrong, stated without the fix attached

`writeWorkItem` (`platform/orchestration/actions/load_work_item_actions.go`) runs an anti-churn
brake before every keyed insert. It counts `complete`/`failed` siblings on `(site_id, item_key)`
inside 7 days, and it has two arms:

- **within-cycle** — newest terminal sibling under **3.0 h** old ⇒ `return workItemWrite{}, nil`.
  No row, no error. The caller sees `Inserted:false`, which is **byte-identical** to a genuine
  open-item dedup. The request is gone and nothing anywhere records that it existed.
- **two-strike** — ≥2 terminal siblings ⇒ status rewritten to `unresolved`, which is in
  `workItemTerminalStatuses` **and** absent from `workItemDispatchableStatuses`. The row is born
  terminal, undispatchable, outside the dedup index, and reads to every dashboard and every human
  like a decision somebody made.

**Measured 2026-08-23:** 635 of 747 live `unresolved` rows carry the two-strike brand; the two
largest populations are `page_rerender` (212) and `improve_tool` (205) — both ACTION REQUESTS,
i.e. the class the brake was never meant to act on. 529 `(site_id, item_key)` pairs are armed for
the two-strike arm right now and 43 sit inside the within-cycle window.

**The within-cycle arm's damage is structurally unmeasurable after the fact**, and that is worth
stating as a property rather than a limitation: it leaves no row, so there is no census that can
count what it destroyed. The only evidence is a live log line, or an orchestration row before it
is reaped at ~24h.

## 2. Why the classification fix does not settle it

`recurrenceExpected` already exists and already works. `bugs_closed/024` established it,
`work_item_recurrence_test.go` still states the rule, and migration 572 applies it to the five
build-chain steps. **That is the right fix for a caller that has been classified.**

The open question is what should happen to the callers that have **not**. Adoption measured
2026-08-23: **19 of 21** live keyed `create_work_item` steps had never declared either way, plus
**36** non-test `insertWorkItem` call sites in Go whose classification nobody has audited at all.
024 drew the correct conclusion two years ago and it did not propagate, because nothing counted
adoption — the census shipped with `bugs_open/326` fixes *that*, but a census reports; it does not
protect.

So: **for an unclassified caller, is silent destruction an acceptable default while the census is
worked through?** That is the question, and it is not mine to answer.

## 3. The proposal, and the guardian's objection to it

**Proposed:** both arms DEFER instead. The row is written at the status the caller asked for,
with `retry_after` set — the window remainder for the within-cycle arm, `maxBackoffMinutes`
(720) for the two-strike arm. The brake is unchanged in effect (nothing dispatches inside the
window) and arguably strengthened (a deferred row holds the dedup slot, so re-finds during the
window collapse onto it rather than each being separately swallowed). No status list, no index
predicate, and the `ON CONFLICT` text is byte-identical.

**The guardian's objection, which stands:** this is a *default* change for the whole fleet with
only a global env-var to revert it, where the estate's ruling of 2026-08-02 §2 asks for
opt-in-per-caller with the unsafe side OFF. My defence at submission time was that the *new*
behaviour is the safe one and the *old* one is the unsafe one, so an opt-in default-OFF switch
would leave the damage armed. **I still think that is true, and I now think it is beside the
point:** the ruling is about who gets to decide and where a reviewer can see the decision, not
about which branch the author believes is safer. A shared seam whose behaviour changes under 36
callers who did not ask is exactly the shape it names.

**Three options, costed:**

| | shape | what it costs | what it buys |
|---|---|---|---|
| **A** | Ship as proposed: default deferral, global kill switch | The 2026-08-02 §2 objection stands. ~529+43 keys change behaviour at once; dispatch volume and dashboard counts move fleet-wide on the same day | Every unclassified caller protected immediately, including the 36 Go sites nobody will audit |
| **B** | Opt-in per `item_key_prefix` or `agent_type` (the guardian's suggestion) | Protects only callers someone has already thought about — which is the same population `recurrenceExpected` already covers, so it may add a second lever that does the same job. **The 19 undeclared steps stay unprotected**, and they are the ones that need it | Satisfies §2 exactly; blast radius is opt-in and incremental |
| **C** | Do nothing to the brake; drive the census to zero and ratchet `recurrence_expected` to REQUIRED for keyed steps | No fleet-wide change at all; slowest. Leaves the 36 Go call sites entirely unaddressed, since the census only sees config | The bad state genuinely stops being representable for config callers — the strongest end state, if it is ever reached |

**My reading, offered as a view and not as a finding:** B collapses into a second spelling of
`recurrenceExpected` and I cannot see what it adds over C. The real choice is A (protect
everyone now, accept a one-day fleet-wide behaviour move) versus C (protect nobody extra now,
reach a stronger end state later). A hybrid — C as the destination, A as the interim, with the
kill switch as the retreat — is what I would build if the decision were mine, and it is not.

## 4. What must be answered whichever way it goes

`retry_after` now carries **two causal meanings**: "failed once, backing off" (RFC_043's
contract) and "created but anti-churn-deferred" (this proposal). The architecture seat objected
at medium on exactly this and it is the one objection with concrete work attached:

> *"Please confirm a test exercises each of those three readers against a row that has
> `retry_after` set but zero prior failed attempts."*

The three readers are `claim_work_item_action.go:111`, the dispatch loader
(`load_work_item_actions.go:721`) and `complete_work_item_verification.go:438`, all rendering the
predicate through `workItemRetryNotPendingSQL`. **That test does not exist and I did not write
it.** Whoever takes this owes it before anything ships, and it is owed under option A or B
equally. It also belongs in **RFC_043**, whose contract this extends — this RFC should be read as
its continuation rather than as a competing account.

## 5. What has already landed, so nobody re-does it

Committed `d0930af6f` (+ `74c527f56`), all of which the guardian said it would approve alone:

- **migration 572** — the five build-chain handoffs declare `recurrence_expected: true`.
  Config-only, live on apply. **This closes `bugs_open/326`.**
- **migration 573 (`_HOLD`)** — `on_dedup: "error"` on the front door; held for the roll, and
  inert until the Go half exists.
- **`config-key-audit --undeclared-recurrence`** + wrapper — the census, 19 findings over 194
  live agents on its first run.
- Register WII-005 (corrected in place) + WII-027, two LANDMINES entries, two WRONG_CALLS
  entries, and the `bugfix_326_retry_the_front_door/` lane docs.

Both migrations were hardened after the round on `debug_historian`'s objections: the
`snapshot_agent` call is now gated on the same pre-state marker that drives the UPDATE (so a
re-run is a true no-op rather than a second snapshot mislabelled "pre-update"), and both refuse
outright if the target type carries duplicate active definition rows. That last one does not fire
on these five today — each has exactly one active row — but **four types fleet-wide do**
(`content-creator`, `content-creator-contact`, `chief-strategist`, `site-component-architect`),
and a version-blind UPDATE on such a type patches a row that governs nothing while the verify
block still reports success.

## 6. The patch, preserved rather than left dirty

`RFC_048_proposed_deferral.patch`, beside this file. It applies to the `writeWorkItem` /
`create_work_item` pair and carries its tests, all five mutations proven:

| # | mutation | caught by |
|---|---|---|
| M1 | remove the `retry_after` column append | arg-count mismatch, 4 tests |
| M2 | restore the legacy drop on arm A | `..._WithinCycleWindow_DefersRatherThanDrops` |
| M3 | restore the `unresolved` brand on arm B | `..._TwoCompletedPredecessors_DeferRatherThanPoison` |
| M4 | delete the kill switch | both `TestAntiChurn_KillSwitch_...` subtests |
| M5 | flat 3h interval instead of the window remainder | 4 tests, all three boundary cases |

It is a patch and not a commit **on purpose**: on this tree HEAD is shared and any session's roll
ships whatever is committed, so committing a vetoed change would be shipping it. Leaving it dirty
in the working tree would be worse — another session's `git add -A` sweeps it under an unrelated
message.

**Note for whoever applies it:** it also updates
`write_render_audit_findings_test.go`, another lane's council-answering pin
(corr `e49f5935`, `bug_historian`, high). That objection — *"omitting recurrenceExpected silently
drops the third occurrence"* — remains answered either way; only the specific assertions change,
from `status 'unresolved'` + branded summary to the caller's status + unbranded summary + a 17th
argument. It is flagged here rather than quietly rewritten.

## 7. What this RFC is NOT asking for

- **Not** a decision on the 635 existing `unresolved` rows. Draining that landfill is RFC_010 /
  `bugs_open/033` D2's open owner decision and overruling it from here would repeat this RFC's own
  mistake at a smaller scale.
- **Not** a re-litigation of `recurrenceExpected`. It works; 572 uses it; the census names who
  else needs it.
- **Not** a change to `idx_swi_dedup`, `workItemTerminalStatuses` or any status list. The dedup
  index is innocent — that is `bugs_open/326`'s central correction — and the lockstep stays
  untouched deliberately.

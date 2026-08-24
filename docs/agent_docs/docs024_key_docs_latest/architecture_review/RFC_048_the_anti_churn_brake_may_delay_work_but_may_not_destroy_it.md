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

## 6a. I went looking for a second casualty and there isn't one — recorded, because I asked

**This RFC rests on ONE casualty and it is going to stay that way.** Within an hour of the
veto I asked the `loanzy.uk` lane whether `bugs_open/376` — a build of `garden-tools.uk` that
died at its second hop the same evening — could serve as a second independent motivating case,
saying *"a second real casualty is the strongest argument it can have."*

**They refused it on the merits, and they were right.** Their reasoning, which I am recording
rather than paraphrasing away:

> *"In `376` the brake never runs. The build dies because `vertical-exemplar-researcher`'s crawl
> step has no `on_error`, so a Firecrawl refusal kills the child orchestration before
> `create_next_item` is ever reached — `needs_strategy` is not destroyed by the brake, it is
> never created in the first place… If `376` were cited in RFC_048, the first reviewer to read
> it would find the brake absent from the mechanism and would be right to discount it — and on a
> proposal that already took a guardian veto, a case that does not survive contact is worse than
> no second case."*

The distinction is exact and worth keeping, because the two produce an identical operator
experience — a build that stops with no explanation:

| | what happened | what fixes it |
|---|---|---|
| **RFC_048's class** | the item was created and the brake ate it — **destruction of work that existed** | this proposal |
| **`376`'s class** | the producing step died upstream, so no item was ever created — **absence of work that should have existed** | an `on_error` on the crawl step |

A fix for either leaves the other exactly where it is.

**What `376` DOES support, narrowly, and this is theirs not mine:** it is independent evidence
for the *premise* underneath this RFC rather than an instance of its mechanism —
`vertical-exemplar-researcher.create_next_item` is the **only** producer of `needs_strategy`
anywhere in the estate (swept every live agent's steps; one row), so this pipeline has hops with
**no producer of last resort** and anything stopping one hop stops the build permanently. That is
the soil this RFC grows in. Cite it as motivation for the general shape; **do not cite it as a
second casualty.**

> **⚠ CORRECTED 2026-08-23 20:1xZ — "permanently" is wrong, and the retraction came from the
> lane that supplied the case.** A fifth exemplar draw, 30 minutes after the four that looked
> settled, **re-drew**: the refused host was absent and one pick came from
> `identity.competitors_found` — the branch both lanes had recorded as never having fired. All
> three crawls cleared, and `garden-tools.uk` is now at **hop four** (`needs_strategy` claimed
> 20:05:55Z, `needs_briefing` queued 20:09:29Z, `vertical_landscape` spec written 20:05:45Z —
> verified at the artefact by this lane, not taken on report).
>
> So the accurate statement is: **4 of 5 draws contained the refused host.** The pool is
> **biased, not fixed**; a retry escaped on the very next observation. What survives unchanged is
> the structural half — `create_next_item` really is the sole producer of `needs_strategy`, and
> the step really has no `on_error` — so an **exhausted attempt budget** is still terminal. That
> is "usual", not "certain", and the word "permanently" should read "once the three attempts run
> out".
>
> **This does not touch §6 or §6a of `PROPOSAL_D9`**, whose delivery-gap finding depends on none
> of it.

**The misstep is mine and it is the interesting part.** Asking for a second case is a reasonable
thing to do; asking for one *an hour after a veto*, and telling the person I was asking how much
I wanted it, is how a case that does not fit gets written in anyway. The check I skipped is one
sentence long — **does the mechanism I am proposing to change actually appear in this case?** —
and I skipped it while writing an RFC whose whole subject is a mechanism being blamed for damage
it did not do. The peer ran it for me. `WRONG_CALLS.md`, same date.

## 6b. CORRECTION 2026-08-24 — this RFC conflates two arms with different problems, and the options are mis-costed as a result

Re-measured when the owner asked for a review before ruling. Full figures in
`bugfix_326_retry_the_front_door/DECISIONS_2026-08-24_what_needs_an_owner_ruling.md` (second
version); the corrections that change this document:

- **Arm A (drop under 3h) and Arm B (two-strike → `unresolved`) are not one problem.** Arm A is
  326's bug and has **no legitimate use for any caller** — nothing wants its request destroyed
  with no record. Arm B is a designed landfill. Of its 661 rows (2026-08-24), **431** are action
  requests that should never have been braked (a classification failure: 205 historical
  `improve_tool`, 212 ongoing `page_rerender` from the **Go** discovery sweep at ~3.4/day) and
  **230** are the brake working correctly on a fixer that reports done without fixing
  (`bugs_open/352`). Deferring Arm B would re-dispatch that futile fix every 12h.
- **The duplication is the landfill's real disease, and only the deferral fixes it:** 661 rows
  over **247** keys, 2.68 per key, worst key 91 rows in two days — because `unresolved` sits
  outside the dedup index and every re-detection lands a fresh corpse. A deferred row holds the
  slot.
- **§3's cost of option A was wrong by ~70×.** "~529 keys armed" is exposure, not volume; the
  actual extra dispatch under A is ~8/day. §3's "B collapses into a second spelling of
  `recurrenceExpected`" is also wrong for detectors, which today have no lever between "brake
  me" and "bury me".
- **Two options were missing.** **D:** defer Arm A only — the smallest change that ends silent
  destruction for everyone, with an unambiguous safe side (the patch minus its two-strike branch).
  **E:** set `recurrenceExpected: true` on the ~10 Go action-request producers — per-caller,
  exactly the 2026-08-02 §2 shape, no mechanism change, and the only thing that stops the
  `page_rerender` bleed the config census cannot see.
- **`on_dedup` (edit 2) is separable from the deferral** and its low-severity objection is now
  answered by query (0 live workflow conditions branch on `deduped`/`inserted`, 2026-08-24). It
  can be resubmitted alone; migration 573 does not wait on this RFC.

The author's view moves accordingly: **D + E now, the census alongside; the duplication to
RFC_010 where it already lives; the detector rows to 352.** A remains the only option that also
fixes duplication, at ~8 dispatches/day, not 570.

## 7. What this RFC is NOT asking for

- **Not** a decision on the 635 existing `unresolved` rows. Draining that landfill is RFC_010 /
  `bugs_open/033` D2's open owner decision and overruling it from here would repeat this RFC's own
  mistake at a smaller scale.
- **Not** a re-litigation of `recurrenceExpected`. It works; 572 uses it; the census names who
  else needs it.
- **Not** a change to `idx_swi_dedup`, `workItemTerminalStatuses` or any status list. The dedup
  index is innocent — that is `bugs_open/326`'s central correction — and the lockstep stays
  untouched deliberately.

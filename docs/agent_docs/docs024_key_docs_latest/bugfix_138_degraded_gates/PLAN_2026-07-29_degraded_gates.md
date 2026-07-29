# PLAN — bugs_open/138: a truncated advisory review silently becomes a blocking one

**Started** 2026-07-29 by thread "bugsearch 5", which also filed the bug.
**Parent**: `bugs_closed/076` (truncation contract) — 138 is 076's fix behaving
exactly as designed and having a consequence nobody costed. Not a regression of it.

## The problem in one line

`Degraded` (truncated) reviews gate unconditionally — correctly, since a high
objection may have been cut off — but the verdict names the SEAT, so a token
budget overrun is indistinguishable from a judgement.

## Why the obvious fix is the wrong one

Relaxing the `Degraded` gate would let a cut-off high objection through silently.
That is strictly worse than a spurious revise: a wasted round costs credits and
~30 minutes, a missed high-severity objection costs a bad change in production.
**The gate is not the defect. The silence is.** Every candidate below preserves
the gate exactly.

## Phasing, and the reasoning behind the order

Ordered by what closes the door (`[[order-fix-candidates-by-what-closes-the-door]]`),
not by effort:

1. **Make the cause visible in `decided_by` + persist a `gated_by_truncation`
   flag.** ✅ **DONE 2026-07-29** (code below). First because it addresses the LIVE
   harm, which is not the wasted round — it is that a high object-rate with no
   signal line is *also* the documented kill-switch for retiring a seat. **A
   working seat can be pulled for being noisy when it was being cut off.** That
   nearly happened to `review_architecture` and is now demonstrated, not argued:
   its object rate fell 2-of-3 → 2-of-12 the moment it stopped truncating.
2. **Alert on the rate.** Cheap only *after* (1), because (1) is what makes the
   rate a stored field rather than a forensic replay over `reviews[]`.
3. **Right-size `max_tokens` per seat.** Actioned in part by the council-parallelism
   thread on an owner call. **Deliberately NOT first**: raising a cap moves the
   door, it does not close it — `architecture` reintroduced truncation against the
   same cap within hours of being seated, purely by having a longer prompt.
4. **Emit the load-bearing field FIRST in every seat's output schema.** Truncation
   eats the tail, so whatever must survive belongs at the head. Generalises what
   was already done for `review_architecture`, where the mandated
   `ARCHITECTURE_SIGNAL` lived in `notes` and `notes` was emitted last — truncation
   destroyed precisely the field that made the seat measurable.

## Decisions taken, with reasons

**A merits gate is named in PREFERENCE to a truncated one**, even when the
truncated seat comes first in review order. The alternative (first-gate-wins,
matching the old code) would label a round TRUNCATED while a second seat held a
genuine high-severity objection — inviting the author to dismiss a real objection
as a budget problem. That is the same class of harm this bug is about, pointed the
other way. Consequence: the `TRUNCATED` label now means something precise —
*nothing else gated this round*. Cost: one round in 15 changes which seat it names
(measured, §2 of the RUNBOOK), and it moves the name onto the seat with the real
objection.

**A degraded review with ZERO objections is TRUNCATED, not a bare object.** The old
`len(Objections)==0 → gates` rule exists because an ungraded object is not
"explicitly minor". But when the review is also `Degraded`, the emptiness *is* the
truncation — it was cut off before writing any. A complete review that objected
without grading anything remains a real, if sloppy, judgement. Both still gate.

**`gated_by_truncation` is emitted unconditionally, true or false.** A
sometimes-present key cannot distinguish "measured and clean" from "written before
the fix", and the entire point of this bug is that the rate was never measurable.
Persisted to `metadata` (jsonb) as well as `body` (text) so the alert in phase 2 is
an indexed one-liner.

## Scope judgement — why this is council-gate scope and not an RFC

Under the owner ruling of 2026-07-29, a shared-mechanism change needs an RFC when
it changes what the mechanism **GUARANTEES**. This one does not:

- `objectionGates` is byte-for-byte the same rule. Its pre-existing test passes
  **unmodified** — that is the check, not my say-so.
- The guarantee "a Degraded object always gates" is preserved exactly.
- `decided_by` is free text at every consumer (digest, escalation, reviser and
  reframe prompts — grepped, none parse it). It gains precision.
- `gated_by_truncation` is additive and inert: nothing reads it until someone
  writes the phase-2 alert.

What IS a real change and is stated rather than buried: **which seat is named in a
mixed round**. Measured at 1 occurrence in 14 days.

Consumers to TELL, not merely measure (the third limb of the same ruling): the
council report is read by `fix-proposer`'s reviser prompt, `feature-designer`'s
reframe prompt, `fixloop_digest_action`, and `diagnose_escalate_action`. The useful
message to them is not the new key — it is that **a `revise` they receive may now
say the round was decided by a token budget rather than by a reviewer**, which is
new information they have never had.

## What this does NOT do

It does not stop rounds being wasted. A truncated seat still forces a revise, by
design. It makes the waste countable and correctly attributed, which is the
precondition for phases 2–4 and for ever knowing whether they worked.

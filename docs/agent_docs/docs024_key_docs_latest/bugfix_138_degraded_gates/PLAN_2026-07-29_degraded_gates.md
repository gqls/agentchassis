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

---

## Phase 2 and phase 4, decided 2026-07-30 (and phase 4 largely un-built, on purpose)

### Phase 2 — the alert. The design decision was WHICH rate, not how to schedule it.

> **CORRECTION to this plan's own framing.** The section above says
> `gated_by_truncation` is "inert until someone writes the phase-2 alert", implying
> the alert would be built on it. **It is not, and could not be usefully.** The flag
> counts truncation gates *after* they happen, and phase 3 (the cap raises) had
> already driven that count to ~0 on every seat that had ever produced one. An alert
> on it would have been silent by construction — the kind of mechanism the owner
> priced correctly on 2026-07-29 as rotting unexercised. The flag still earns its
> place: it makes an individual round *legible*, which is what phase 1 was for, and
> section 3 of the report reads it. It is just not the alerting signal.

Built on **headroom** instead — `output_tokens` as a fraction of the seat's current
cap — because it warns *before* the damage rather than counting it after. Two named
thresholds, both anchored on the live distribution rather than chosen:

- **near-miss**, peak ≥ 95% of cap: truncation is a tail event, so the maximum is the
  primary signal.
- **pressure**, p95 ≥ 85% of cap: the body of the distribution near the ceiling.

Keeping them separate was load-bearing, not cosmetic — under a p95-only rule the two
flagged pairs had 4 attributable calls each while `review_guardian` (278 calls, 118
attributable, peak 99.2%) read "ok".

Two halves, sharing no threshold: a pull report, and a **CTE-only scheduled task**
(`fire_message=false`, so the `pre_query` *is* the work — no message, no
orchestration, no LLM, no credits). Deduplicated by an md5 of the flagged set in
`subject_key`, so it is an **event, not a heartbeat**: a persisting condition speaks
once, an escalation speaks again. A six-hourly identical note is ignored within a day.

**Deliberately left undone:** the near-miss margin is a flat 5% of cap, which is 400
tokens at 8000 and 800 at 16000 — probably wrong, since the risk is closer to
absolute than proportional. Every 16000-cap seat currently sits below 63%, so the
question has no live consequence and a rule chosen now would be a guess. Revisit when
a 16000-cap seat first crosses.

### Phase 4 — mostly REFUTED by measurement, which is the cheapest outcome available

Filed as "emit the load-bearing field FIRST in every seat's output schema". Surveyed
all 51 live templates and measured what truncation destroys. Three of the four
expected findings do not hold: the head is already correct in 51 of 51; **0 of 2,713**
stored objections lack a severity, so the severity-last theory never fires; moving
`notes` forward would push `objections` — 80% truncation-survival, carrying both the
gate's severities and the proposer's revision content — into the tail instead. The
guardian-veto risk is real and has **0 instances in 15 vetoes**.

The severity one is worth keeping in view: every step of that argument is true and the
rate is zero. **A mechanism that is real at every step can still never fire.** Reasoning
would have shipped a fix for it; one count did not.

So the reorder is a per-seat judgement about a seat's own remit — which is exactly what
`review_architecture` already is — and NOT a fleet rule. What generalises is the length
budget, built as `scripts/apply-seat-length-budget.py`: one copy of the block,
idempotent, refuses to overwrite a hand-authored one, snapshot-then-write, scoped to the
seats the phase-2 report flags with attributable evidence. It follows the owner's own
criterion for the sibling change (act where it is measured, not everywhere it is
imaginable) with the leading indicator substituted for the lagging one.

**Not applied.** The live config write was refused by the session's permission
classifier, so the block is in git and not in the fleet. That is a permission gate, not
a design doubt — recorded here so the gap cannot be mistaken for a completed phase.

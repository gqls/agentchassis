# The owner sees the brief and the rendered site before the delivery email goes out

**Owner ruling, 2026-08-21:** *"I want to see the client briefs and the render of the sites
before approving the sending of the delivery email to them. I'd like to make sure we're
delivering a good site and if it's broken then I want to fix it before sending."*

**Status:** requirement recorded and designed. **Nothing is built**, because the thing it
gates does not exist yet: the delivery email is Phase 4 item 3 and is not written. Recording
it now is the point — a gate designed after the thing it gates is a gate bolted on.

---

## 1. What this is, and the one thing it must never become

It is an **internal quality gate**: nothing ships to a customer until the owner has looked
at the brief and the rendered site. It exists because the product is one-shot, so the site
the customer receives is the only site they receive.

**It is NOT a customer approval stage, and the difference is load-bearing rather than
pedantic.** `one_shot_no_approval` is attested; `writer_block` says *"Never describe an
approval, a sign-off, or any step where the customer confirms the site before delivery."*
The owner reviewing his own work before sending is invisible to the customer and changes
nothing they are promised. **The risk is not the gate; it is the gate leaking into copy** —
a writer who learns "there is an approval step" will write one, and the customer-facing
claim would then be false.

> ### ⚠ AND NOTHING WOULD CATCH IT TODAY
> `[MEASURED 2026-08-21, cmd/claimscan against the live register]` the sentence
> *"You will be able to approve the site once you have seen it"* scans **CLEAN**. The rule
> exists in `writer_block` and in an attested fact, and **there is no ban enforcing it**.
> So the one failure mode this gate introduces is precisely the one the claims layer cannot
> currently see.
>
> **The fix is a ban, and it must be an OFFER-shape ban.** A bare-token ban on "approve"
> would block the denial too — the negation guard scans backwards only and *"there is no
> approval stage"* would be refused while *"you can approve it"* passed. The worked
> precedent is the 2026-08-19 `round of changes` narrowing, and the rule is the attestation
> test: if the register attests the thing, the copy must be able to deny it in normal
> English. Prove it with a probe set carrying BOTH halves.

## 2. Where it goes: the work-item queue, which already has a screen

**Do not build a new UI.** The handoff's open item 7 records the owner's own accepted
ordering: HITL routes through the **work-item queue, which has a working screen**, and the
orchestration HITL path is measured dead (`collect_via_hitl` 0, `brief_answers` 0,
`hitl_mode` 0 across 369 briefing orchestrations, against `briefing_answers` = 3 as the
control). So the gate is a work item, not a new mechanism.

Shape:

1. When a site is ready to deliver, file one work item — proposed `item_type`
   `needs_delivery_review` — carrying the **brief** and a **link to the rendered site**,
   which are the two things the owner asked to see.
2. The delivery email is dispatched **only** when that item reaches an approved terminal
   state. The email step reads the item; it does not read a flag somebody could set by
   hand.
3. Rejecting it routes to the existing repair path rather than inventing one. "If it's
   broken then I want to fix it before sending" is a rebuild, and rebuilds already have a
   queue.

**Reuse note, because this estate rewards it:** de-duplication and `item_key` shape for a
new work-item type are covered by the 2026-08-02 owner ruling — the producer set and the
`item_key` shape must be stated in the concept-register entry, and then no RFC is needed.

## 3. What it costs the customer promise, and this needs the owner's eye

`build_duration` attests **"usually ready in two or three days"**, and that figure was
already tightened once (from next-day) because measurement refuted the faster promise.

**A human review step adds however long it takes the owner to look.** If he is away for two
days, the promise is broken by the gate rather than by the build. Three honest options:

| | |
|---|---|
| **(a)** Leave the promise at two or three days and treat review as part of it | Only safe if review is reliably same-day |
| **(b)** Re-attest the duration to absorb the gate | Costs the sharpest number in the offer |
| **(c)** Time-box the gate: if not reviewed within N hours, it goes | Protects the promise and weakens the gate, which may defeat the point |

**Not chosen here.** It is a commercial trade the owner has already shown he wants to make
deliberately (2026-08-18: *"a better product beats a faster promise"*), and it is the same
trade in the other direction.

## 4. What it does NOT gate

State this explicitly, because the estate's habit is to let a new stamp quietly acquire
meanings: this gate holds the **delivery email**. It does not gate the build, the render,
the deploy, the ZIP being cut, or the site being live at our address. Those all happen
first — they are what the owner is being shown.

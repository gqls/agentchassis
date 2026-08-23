# SUMMARY — 2026-08-23 — CTA destination provenance (`bugs_open/308`)

## What we are trying to do

Fix a bug where the platform spots a broken button, correctly works out where it should point,
files a repair — and the repair cannot possibly perform it. A button reading "Contact our supply
team" links to a break-even calculator. The check sees it, names `/contact.html` as the fix, the
repair runs, reports success, and changes nothing. Next pass it finds it again.

The blocker was never the repair itself. It was that the platform decides whether it may rewrite
a stored link by **reasoning**: *"we could never have produced a link to the contact page, so if
one is there, a person put it there."* That reasoning is sound today and it is exactly what must
stop being true in order to fix the bug — because fixing it means letting the machinery point
buttons at contact pages. The owner ruled on 18 August: **record the provenance properly, then
widen.** Two phases, in that order.

## Where we have come from

`bugs_open/248` had already been bitten by the destructive half of this — a repair overwriting
genuine contact links — and closed it by *deriving* provenance from the resolver's own
constraints. That derivation is registered as LNK-033, and it is what this work retires.

When this lane opened, the bug had grown from the 149 findings it was filed with to **200**, and
the shape of the damage was in the split: **112 of them sat on repairs the platform had marked
`complete`**. Not jobs waiting to run — jobs that ran, declared success, and left the button
where it was.

## What we have done

**Phase A is built, reviewed, shipped and now proven working on live data.**

The record is a small note stored beside the link — `__cta_minted` — saying *which* URL the
machinery wrote. A link counts as a person's when it is real and the note does not name it.
Storing *which* URL rather than a bare "we wrote this" is the decision the whole design rests on:
it means eight of the nine things that write page content need no change at all, and it is the
only form that survives someone hand-editing a button (a bare flag would survive the edit and
licence the machinery to overwrite the person's choice — which is bug 248 all over again).

Three rounds of council review, thirteen to fourteen seats each. **The reviewers found real
defects**, including one I had listed in my own risks and not closed — *"'owed' is not a control
on a mechanism whose whole purpose is a record reaching the database"* — and one blast radius I
had not considered at all (whether the new key perturbs page fingerprinting fleet-wide; measured,
it does not). Mutation testing found two more that reasoning had missed, one of which would have
caused the exact freeze the mechanism exists to prevent.

## Where we are now

**Phase A is live and demonstrably doing its job** — measured 2026-08-23 at the artefact, not at
a status:

- **11 component rows now carry a provenance record**, against a baseline of **0** taken
  immediately after the roll. The number moved, which is the only thing that could have proven it.
- **21 record entries, 21 of which name the URL their field actually carries. Zero mismatches.**
- **Zero non-CTA components stamped** — the negative control holds.
- **Every row carrying a secondary button has that second slot recorded too**, which is the live
  confirmation of the one defect I introduced and caught in testing: the save merges shallowly, so
  a naive version would have recorded one slot and silently dropped the other, freezing it.

**But Phase A does not fix bug 308, and the bug stays open.** Phase A only makes the fix *safe*.
The 200 findings are still unrepairable, because the machinery still cannot offer a contact page
as a destination. That is Phase B.

**Two things now block progress, and neither is technical debt.**

The estate's LLM budget is exhausted — the API reports access returning **2026-09-01** — and it
has grown from 7 failed steps yesterday to **112**. It is not confined to reviews: it is failing
`call_content_writer`, which is live site content generation. That killed the fourth council round
before it was reviewed, and it means Phase B could not be verified at a real page even if it were
written, because verification needs the repair to actually run.

## Where we are going

Phase B is designed and its constraints are settled — including one this lane found that no
existing document recorded: the candidate list must be widened at the assembly point and **never**
at the shared loaders, because a third consumer nobody had named (the site header's button) reads
those same loaders, and no content diff could ever have shown the damage.

Phase B should not start until the budget resets, for a reason that is about evidence rather than
caution: it retires a live invariant, and the only honest proof it worked is a real button on a
real page, which needs the repair path running.

# SUMMARY — 2026-08-06 — the retraction lane: what we built, what we un-built, and why the second one mattered more

*Previous in series: `SUMMARY_2026-08-03b_deployed_asset_path.md`. This one marks a real turn —
the work since then produced a **reversal**, and the reversal taught more than the build did.*

---

## What we're trying to do

Work items in this platform are one-way. A discovery check notices something wrong with a site,
files a work item, and that item sits there until somebody or something fixes the underlying
problem and closes it. The gap we set out to close is the obvious one: **when a problem quietly
goes away, nothing notices.** Items outlive the condition that raised them, stay dispatchable, and
can be acted on months later against a site that has moved on. Eleven such stale items would have
overwritten a live social card the moment an unrelated fix shipped.

So the goal is a platform that can **retract** a finding — close it because something positively
re-observed the problem to be gone, never because a check simply failed to find it again. That
distinction is the whole design: closing on absence would silently delete real defects fleet-wide
the first time a check errored or was blinded by a bad predicate.

## Where we've come from

The mechanism itself — a `Resolved` field a check can populate, and one runner-side function that
closes what it is told to close and nothing else — was designed, reviewed and shipped inert. Then
`check_empty_sections` became its first user and, on its first real sweep, **closed four findings
raised in April that nothing in the platform could previously close**, while leaving six others on
the same site open because it had no evidence about them. That discrimination was the point: a
mechanism that had closed all ten would have looked better in the headline and been the disaster
the design exists to prevent. That number has since climbed to ten, on its own, as more sites get
swept. **The first adopter works.**

The obvious next step was a second adopter, and that is where this period's story starts.

## What we've done

**We adopted the mechanism in a second check, got it approved, shipped it — and then took it out
again.**

The second adopter was the check that flags components missing the fields their own schema says
they must have. It was chosen carefully: we counted its producers two independent ways, sized the
population against live data, refused four categories of ambiguous evidence, and proved every
guard by deliberately breaking it. The review board approved it at the first round. It went live.

**It was redundant.** The platform already had a purpose-built mechanism closing exactly those
items, with the same test, reaching all of them rather than some, and refusing two cases ours did
not. It had been running for a week.

We found out by re-measuring after the deploy and noticing the numbers had moved the wrong way:
the six items we expected to close had already closed — two hours before our own code was running.
Had we shipped a day earlier we would have closed them first, reported six successes, and never
looked.

**The owner chose to fix it properly rather than paper over it.** We reverted our duplicate, and
instead fixed the real gap in the older, better mechanism — which turned out to be sharper than
anyone had described: **that sweep was building a backlog it could not then drain.** Every item it
closes counts toward a strike rule elsewhere in the platform; after two strikes the next copy of
that finding is born in a status the sweep does not look at. Its own success was generating work it
would then go blind to. Five items were already one strike from that state.

Both halves are now live and proven in the running binary, and approved.

## Where we are now

The lane is essentially finished, with one thing outstanding.

- **The first adopter is working unattended** — ten findings retracted so far, drawing down as
  sites get swept.
- **The duplicate is gone**, proven by its string disappearing from the running binary, with a
  control proving the check itself still worked.
- **The real gap is fixed** and proven present in the binary.
- **But the fixed mechanism is not being run.** There is no schedule for it. It has only ever run
  when someone dispatched it by hand. It is correct, approved, deployed — and idle.

**We have just dispatched it once by hand**, scoped tightly to the single item the fix newly makes
eligible, so the outcome is unambiguous.

The one decision left is yours: **should this sweep run on a timer?** It is a genuine trade-off,
not a formality. In favour: the defect we just fixed is only actually fixed if the thing runs. 
Against: it would give an automatic closer standing authority over four categories of work item
across every site, and one of those categories has five different things filing into it and an
open decision of yours already sitting in it. Roughly a third of the platform's scheduled jobs are
deliberately switched off, so scheduling here looks like something that gets decided, not assumed.

## Where we're going

Immediately: read the result of the hand-dispatched run, then make the scheduling call with that
evidence rather than without it.

After that, the mechanism has three more candidate adopters, and the bar for taking one on is now
higher than it was — deliberately. This period produced two lasting corrections to how we approach
that work, and they are worth more than the code:

**Count what already closes something, not just what creates it.** Our checklist for adopting the
mechanism required counting the things that *file* an item. We ran it thoroughly, twice, and it
passed — while the actual problem sat at the other end. The check now asks both.

**When you widen something shared, the damage is rarely where you are looking.** Three times in
three days, the blast radius of a change landed on a consumer we were not thinking about. The last
time we caught it *before* shipping, by measuring every consumer rather than the one that prompted
the work — and that catch alone prevented pointing an automatic closer at twenty-one items covered
by an open decision of yours.

Both are written into the standing traps file, so the next person is told before they have a
symptom rather than after.

---

*Recorded honestly: the central claim in an approved submission was false, fourteen reviewers
accepted it, and the mistake is logged in the fleet-wide wrong-calls file with the ten-second query
that would have caught it. That entry is the most useful thing this period produced.*

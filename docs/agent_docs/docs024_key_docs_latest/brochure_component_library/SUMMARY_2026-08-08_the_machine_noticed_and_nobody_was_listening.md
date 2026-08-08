# SUMMARY — the machine noticed, and nobody was listening (2026-08-08)

*Sweep front of the brochure/fundamentallyai lane. The 08-03 summary covers the camera and
duplication-checker front and is not superseded by this one; they are parallel.*

## What we're trying to do

Make fundamentallyai.com genuinely good — and make it good the way the platform is supposed
to make sites good, which is by running the framework's own auditors over it and acting on
what they say. The site is our shopfront for a platform whose pitch is that it reviews and
corrects its own work, so a site that quietly drifts is worse for us than for anyone else.
The specific ask this week was to put it through the improvement sweep, get it seen by the
visual designer and the copy improver, and fix what turned out to be broken.

## Where we've come from

The week opened with a belief that the site had been built by hand rather than by the
framework, and an instruction to rebuild it properly. Checking that first turned out to
matter: the site **was** built by the pipeline, on 20 July, with every step's agent leaving
its own record — the submission, the research, the strategy, the briefing, the design plan.
The rule about never hand-building a site came from a different site altogether. So the
rebuild did not happen, and two weeks of live work survived.

But the instinct behind the instruction was sound, and the true explanation was worse than a
hand-built site would have been. On 24 July the framework's own design audit had compared the
site against its brief and written down four ways it fell short. **Those findings then sat in
a queue, untouched, for twelve days.** The machine noticed and the notice went nowhere. That
is the story of this front: not a system that fails to see, but a system whose seeing does not
reach anybody.

## What we've done

We fired the improvement sweep, and it worked — fourteen orchestrations, no errors, the full
audit chain, and every one of the sixteen genuine findings picked up and drained. Before
firing it we went through all twenty-three outstanding jobs by hand, because the sweep sends
them all to workers that change live pages; seven were describing problems that had already
been fixed, and those were cancelled with the evidence recorded next to each.

The request to involve a visual designer and a copy improver needed no new machinery at all.
Both are already inside the sweep — it calls a design audit, which calls a visual design
auditor and a content quality auditor. So the site has now been seen by both, and will be
every time this runs. The copy side paid off immediately by flagging pages carrying numbers
that the evidence register could not vouch for.

Then three decisions came back from the owner and all three are done. The capabilities page
now shows current audited figures instead of week-old ones. The seat-count question turned out
to need no change, because the register was already counting seats the way the owner defines
them. And the Platform Log — the page the brief ranks first among our differentiators — was
built, and then linked, which turned out to be the real work.

That linking is the finding worth carrying forward. The page had built itself the day before
and had been serving correctly ever since. But nothing on the site pointed at it, because the
header and footer are stored as finished artefacts and ours had been generated ninety minutes
*before* the page existed. The page was live, correct in the database, listed in the navigation
table, marked as belonging in the footer — and unreachable by anyone actually browsing.
**"The database says it is in the footer" and "the footer says it" are two different facts**,
and only one of them is what a visitor experiences. Twenty-five of twenty-eight pages now
carry the link.

Along the way we declined to use the tool named after the job. The navigation updater's first
act is to wipe the navigation table and rebuild it, and it has a documented habit of dropping
every link under /tools/, /blog/ and /guides/ — which here would have removed all the tool
links. We used the route that rebuilds the chrome without deleting anything, and checked every
navigation target actually loaded before publishing the navigation everywhere.

## Where we are now

The site is in better shape than it was, and more importantly it is in a shape we can see. The
queue is drained. The capabilities figures are current and self-maintaining, because that chart
looks its numbers up from the audited register rather than storing them. The Platform Log is
live and reachable. A wrong guide count is corrected.

Three things are known-broken and deliberately not fixed, each written up rather than
half-mended. The logo repair path fails on every site, because the step that regenerates a
missing logo asks for information the thing raising the alarm never supplies. Two findings are
stuck with no worker assigned at all — the same disease as the twelve-day silence. And
yesterday's planning pass created three duplicate page entries pointing at addresses that
return errors; they are marked active, so the system currently treats them as legitimate places
to send visitors. That last one is the live proof of a fault we had described in the abstract
on Wednesday: the platform decides a page is safe to link to by asking whether it is *listed*,
never whether it was ever *published*.

We should also be honest about our own accuracy, because it is the same failure as the site's.
Three times this week we repeated a figure without re-measuring it, and each time the figure
had moved. The worst was telling the owner the capabilities page was understating its approval
record by more than half, when the sweep's own re-renders had already closed most of that gap
hours earlier. A finding's evidence is a photograph taken at the moment it was filed, and on
this estate the thing being photographed often moves *because of the very run that filed it*.
All three are logged together so the pattern is visible, not just the incidents.

## Where we're going

The immediate question for the owner is what to do about the three duplicate error pages, since
deleting them is quick but might collide with a planner still working. Behind that sits the
fault they demonstrate, which deserves a proper fix: the platform already has the right test for
"has this page actually shipped" and the link resolver simply does not use it. That is a
cross-cutting change and should go through the review council rather than be patched quietly.

The logo mapping failure wants a diagnosis run rather than a guess, because it is a shared
handler and the same fault will be sitting under other sites. The unrouted findings need
somebody to decide whether they get a worker or get closed — leaving a finding that can never
be acted on is how we got the twelve-day silence in the first place.

And the larger prize is unchanged: the improvement sweep now demonstrably works end to end on a
real site, so the open question is no longer whether the platform can audit and repair itself
but how often we let it. That is a scheduling decision, and it is the owner's.

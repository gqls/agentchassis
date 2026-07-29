# Consolidation and divergence — where we are, 2026-07-29

*Third in the series. The first two were `SUMMARY_2026-07-27_consolidation_and_divergence.md`
and `SUMMARY_2026-07-28_what_the_audit_got_wrong.md`. Cold-start for the work
itself: `HANDOFF_2026-07-29_continue_here.md`.*

## What we're trying to do

We intend to run thousands of domains from one platform, and we are at fifteen or
so live sites. The thing that stops that scaling is not hosting and not the
cluster — it is building the same thing repeatedly in slightly different ways.
This work exists to find where that is already happening, fix what is worth
fixing, and put something in place that notices the next one early enough to
matter.

As of today there is a second half to that sentence, and the owner has just made
it explicit: **it is not enough to build the shared thing. The job runs until
something is actually using it.** Everything below is shaped by that.

## Where we've come from

It started with a near-miss: a design here proposed a new public service for the
island machine one day after another thread had already shipped one there. The
owner caught it by asking how the two would fit together; nothing in the platform
did.

That produced a consolidation programme, and then the programme's own audit turned
out to be wrong three times over — the "identical" health servers were eight
different bodies, the "duplicate" exporters shared a purpose and none of their
functions, and the headline count of single-site actions recounted from nine to
one. The biggest planned item was withdrawn as a won't-do on evidence. What
survived was two genuinely missing shared pieces, both since built, tested and
council-approved: an email sender, which the code we deploy simply did not have,
and a set of protections for public web endpoints.

Yesterday's summary named the awkward part plainly: those two packages were
**built, approved, and used by nothing.** That is the worst position shared code
can occupy — all of the maintenance cost, none of the benefit, while the four
diverging copies it was meant to replace carry on diverging. Today did not change
that fact, but it changed what we know about it.

## What we've done

The plan said the next step was to connect the protections package into the public
tools service, and there was a filed bug arguing why: I had written that a visitor
to that service could lie about their own network address, slip past the usage
limit, and poison the record of who did what. It was filed with an honest note
that I had not actually tested it, and that somebody should before quoting it.

This morning I tested it. **It is not true.** I sent the service two requests
pretending to come from a documentation address, once by each route a visitor
could try, and both were recorded as coming from the real place. A visitor cannot
lie to us about this.

The same test found something else, and this one is real. **The service is not
recording visitors at all — it records the machine standing in front of it, the
same value every time.** Eighty-three visits since the twenty-fifth, one identical
entry on every single one. Two things follow. The limit on how often the tool can
be used is meant to be per visitor; because everyone looks identical it is in fact
one shared limit for the entire internet, so one busy person exhausts everybody's
allowance and nobody has to be malicious for that to happen. And the column that
records who visited has never distinguished anybody, in a way nothing would flag,
because it is always present and always well-formed.

I was wrong for a reason worth keeping. I reasoned about our own program and
stopped at its edge. But whether a visitor can lie about their address is not
decided by our program — it is decided by the two pieces of plumbing in front of
it, one a configuration file on a machine and one a supplier's service, neither of
which is in this repository. Between them they clean the lie up before our code
ever sees it. No amount of careful reading of our own code would have got me
there.

**The consequence for the plan is the important part.** Connecting the protections
package in, exactly as scheduled, **would not have fixed this.** It would have
arrived at the same useless constant by a different route — and worse, it would
have looked like a fix, because the change would be sitting in the very file where
the problem is. It would have been adopted, marked done, and left the defect
running.

I also found why, and it generalises. That package explains its own safety by
describing the plumbing on idea.uk, where it was written. The plumbing on the
island behaves differently on all three of the relevant signals. The reassurance
travels with the package; the thing that makes the reassurance true does not.
Anyone adopting it elsewhere inherits the first without the second.

The record is corrected in the places people actually read: the bug file now
carries its own refutation rather than a quiet edit, the fleet-wide log of wrong
calls has the entry, the debugging guide has the transferable version, and the
thread that owns the service has the evidence in its own directory. Nothing on the
live machine was touched — no configuration, no restart, no rebuild. The plumbing's
behaviour was established from a throwaway copy on my own machine.

## Where we are now

The two packages are still used by nothing, and today removed the specific
argument that was going to justify connecting one of them up. That sounds like a
setback and mostly is not, because the general argument was always the stronger
one and is untouched: there are three different implementations of the same
protection in the estate, the weakest of them guards the only public endpoint,
there are four different postures on who may call us, and the next thing we build
was already planning to copy one of them again. None of that depended on the bug I
got wrong.

What has genuinely changed is that we now know one of the packages is not yet safe
to adopt anywhere, because it assumes a specific piece of plumbing without saying
so. That is a real defect in our own code, it is ours to fix without needing
anybody's permission, and until it is fixed every adoption carries a hidden
assumption.

The other thing to be honest about: most of what remains for this programme lives
inside services other threads own. That is why it keeps stalling. Adoption is not
a commit, it is a conversation followed by a commit, and nothing currently
schedules the conversation.

The owner has now ruled on that directly: **get it adopted.** The programme's
remit runs to adoption, not to delivery of the package.

## Where we're going

Three things, in order, and the first is ours alone.

Fix the portability defect in the protections package, so that what it trusts is
stated and chosen rather than assumed. Done properly this converges with the live
bug rather than competing with it: taught about the island's plumbing, the package
would identify real visitors there, which means adopting it would then genuinely
fix the shared-limit defect instead of merely appearing to. That turns adoption
back into something worth doing on its own merits. It changes a shared mechanism,
so it goes to the review council on its own, not folded into something else.

Then adopt, starting with the email sender, which has a consumer waiting and no
such complication, and then the protections package into the public tools service
— which needs the owning thread's agreement, since it is their service and they
have their own open work against it. The evidence they need is already in their
directory.

There is also a debt to pay that is small and genuinely owed. When the council
approved last week's work it asked that one structural claim of mine be
independently checked before anybody treats it as settled, because I am the only
person who has ever looked. Today is a good argument for paying that promptly.

And the larger question stands where yesterday left it: not "how do we make this
one thing configurable" but "how does a second site acquire one of these at all".
That is the maturity-ladder work — named rungs, worked examples — which remains
the stated method for lifting the whole portfolio and still has no owner. Adoption
is the small version of that question. The ladder is the large one.

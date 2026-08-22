# SUMMARY — 2026-08-22 — the contrast check lane

## What we are trying to do

Make the platform able to fail a page for text a person cannot read. Today it cannot: the
acceptance ladder that every tool page passes through has nine rungs and, until this week, not one
of them could see colour. Every time unreadable text has reached a live site — four sites now — the
thing that found it was a person, usually the owner, looking at the page.

## Where we have come from

Bug 131 is the owner's own list of eight problems from using the vonc Gauntlet in July. Seven were
fixed within days. The list's closing observation was the important one: several of those defects
had passed every automated check hours earlier, and two of them were invisible to any check we own
*by construction*. One of the two — content cut off on phones — was closed properly in July: a new
clause went into the browser checks, rolled, and was then witnessed catching a real cut on a
different site. The other — unreadable colour — was written up as a proposal the same day and then
nobody built it. It sat unbuilt for three and a half weeks while the platform grew a weekly sweep
that measures contrast, files tickets about it, and cannot stop anything shipping.

## What we have done

Re-measured the live pages first, and found the July fixes had quietly decayed: the headline accent
that was fixed to a measured 3.31:1 now reads 2.48:1, because the purple behind it shifted shade,
and the phone text column is narrower today than when the owner first complained. Nothing noticed
either, which is the argument for a standing check rather than another repaint.

Then built the check. It measures every piece of visible text the way a browser paints it, fails the
page on anything genuinely unreadable, names the exact element so the repair machinery knows where to
go, and deliberately stays silent about text over photographs, where the measurement cannot be
trusted and a wrong failure would aim an automatic rewriter at a correct page.

The review council rejected the first version, and it was right to. The objection: if the
measurement never actually ran, my check reported the page as fine — the identical fault it was
written to end, and one I had quoted three times in my own submission while it sat in the code
underneath. The fix makes non-measurement loud: the probe marks its own work and counts what it
looked at, and anything unmarked or empty is reported as a failure to measure. I broke each new
guard on purpose to confirm the tests catch it. Round two was approved, with the objecting reviewer
approving outright.

## Where we are now

The check is committed, approved and **inert**. It runs in a service with its own image, so the next
ordinary release does not carry it; nothing will happen until that service is rebuilt. That is
deliberate and it is also why no page, tool or planner has been told the check exists yet — a check
the running service does not recognise is silently skipped, and a skipped check reads as a pass, so
switching it on early would be worse than leaving it off.

Two loose ends from the original bug are not ours to close and are now written where the right
people will see them: the phone column width was never actually decided (three documents disagree
about it) and belongs to the design pass already queued for that site; and the product question the
owner ruled on in July is finished on the engineering side, with only his own distribution
experiment outstanding.

## Where we are going

Next session, after the browser-runner service rolls: confirm the running binary actually contains
the check (by asking the pod, with controls, not by trusting a version number), then prove it on a
page that is genuinely bad — vonc's own Gauntlet has text at 1.66:1 today, with a clean page
alongside as a control. Only once it has been seen failing something real does it get advertised to
the planners and written into fences. After that, the open question worth someone's judgement is
whether standing tool checks should include contrast by default, which is an estate-wide call rather
than this lane's.

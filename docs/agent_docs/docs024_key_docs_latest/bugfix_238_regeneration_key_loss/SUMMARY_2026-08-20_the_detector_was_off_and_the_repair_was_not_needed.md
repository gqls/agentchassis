# SUMMARY 2026-08-20 — the detector we built was switched off, and the repair we planned turned out not to be needed

*Written to be read aloud. Previous in the series: `SUMMARY_2026-08-11_the_llm_was_never_holding_the_keys.md`.*

## What we're trying to do

Stop the platform quietly deleting parts of a page when it rebuilds one — and be able to tell when
it has. A page's sections are stored as a set of named values. Some are written by the language
model (the words), some are looked up from elsewhere in the system (a picture's address, a button's
destination, a contact e-mail). When a page is rebuilt, the looked-up ones can vanish if wherever
they came from has gone quiet — and the page then serves a broken picture, or a button that simply
is not there.

## Where we've come from

In August we found this on a live homepage: five broken images, five missing links and a missing
call-to-action, on a page that had looked fine for months. We fixed the cause, built a detector for
the symptom, repaired that site, and left the bug open because other sites were still affected.
Three days later the same fault reappeared in a different form and cost 214 buttons across nineteen
sites; that got its own number and its own fix.

## What we've done

**We found that the detector had never been switched on.** It has been sitting in the running
system, complete and dormant, since the 12th of August. It has never fired once. The reason is
uncomfortable and worth saying plainly: our own internal reference card still described it as *"not
running yet, waiting for the next release"*. That was true the day it was written and false the day
after, and nobody went back to change it. So every session since — including the ones best placed
to throw the switch — was told, in the place we keep our most trusted notes, that it was too early.
**A note that goes out of date did not just misinform here; it prevented the very thing it
described.** We have turned the safe half of the detector on, corrected the note with the cost
written beside it, and recorded the decision inside the system rather than only in a document.

**We measured the repair before building it, and the measurement said not to build it.** The plan
was to find every page that had lost data and repair what could be repaired automatically. The
honest answer, once measured, is that there is nothing to repair automatically: in almost every
case the page never had the value to lose, because the place it was supposed to come from has never
existed on that site at all. A rebuild would restore nothing. What those pages need is for someone
to supply the missing information or change what the template asks for — content work and a
decision, not automation. Ten pages need that human judgement.

**We proved that August's other fix actually works, which nobody had checked.** We believed it; we
had never demonstrated it on real traffic. We could, by asking the archive one question: across
every rebuild the platform has recorded, did any of these values go from present to blank? It
happened 66 times, on 11 sites, all between the 11th and 14th of August — **and not once since the
fix landed on the 14th, across more than three thousand rebuilds since.**

**We wrote up the underlying design question for you to decide.** A reviewer flagged eighteen days
ago that our fix works around the real problem rather than closing it, and asked for that to go to a
human. It never did. It has now, with the options costed and a recommendation.

**And we did not ship two small fixes we had planned.** They address real gaps in the code — we can
point at the lines — but neither has ever actually happened in production, on any site, in the whole
window we can see. Writing them would have been quick and would have looked like diligence. They are
recorded instead, with the exact query that would justify building them.

## Where we are now

The prevention works and is now proven rather than assumed. The visible kind of damage — a broken
image with an empty address — is detected from today. The invisible kind, where a missing
destination makes a button disappear entirely rather than leave an empty one, still has no automatic
detector: the check that would find it deliberately ignores this class, and the background scans
that would run it are switched off to save money. Ten pages are waiting on a human decision about
information that was never there.

The bug stays open, and the reason has changed for the third time — which is itself the most
useful thing to know about it. It is no longer "the fix hasn't shipped", nor "the fix hasn't rolled".
It is that the remaining work is a decision and some content, not code.

## Where we're going

Three things, in order. First, make the ten pages visible as real queued work rather than entries in
a log only we read — but that has to arrive together with a rule that stops them being handed to the
copywriter, who structurally cannot supply an address. Second, once those are dealt with, turn on
the second half of the detector, the half that refuses to rebuild a broken page at all; it is held
back deliberately because switching it on today would block exactly those pages. Third, your call on
the design question: whether we keep working around the split between the two ways a page gets
rewritten, measure how much it actually costs us first, or close it properly — our recommendation is
to measure before building, for the same reason it paid off this week.

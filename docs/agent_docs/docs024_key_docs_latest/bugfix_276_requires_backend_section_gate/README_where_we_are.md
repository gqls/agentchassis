# README — where we are (requires-backend section gate, bugs_open/276)

**2026-08-15.** Picked up bug 276. The short version: there's a type of website "section" —
a form widget that talks to a server — that should only ever be offered when a site actually
has a server behind it. One of these, called `intent-probe`, already has that check for the
version offered by the "suggest me a tool" agent, but nobody ever added the same check to the
agents that plan ordinary page sections. So in theory, a plain static website (no server at
all) could get planned this widget, and it would just silently not work for visitors.

The bug report, as filed, pointed at one specific agent to fix. When I went and checked
properly, it turned out there are actually three different agents in the fleet that can offer
this kind of section to a page plan, and the one the bug report named is actually the LEAST
used one — it's run twice a month. The one nobody mentioned is run well over a hundred times a
month and ran again just this morning. So fixing only the named one would have looked like a
fix and left the real risk wide open. I'm fixing all three.

The good news: nothing has actually broken yet. I checked, and the only place this widget is
currently placed is on a site that DOES have a server, so no visitor has hit a broken page
because of this. This is purely a "close the gap before it bites someone" fix, not a repair
of something already broken.

Third agent is basically dead code — it's never been run, not once, in the platform's history.
I'm still closing the gap there too (cheap to do), but in a simpler, more conservative way
than the other two, since there's no way to safely check "does this specific site have a
server" for an agent that's never actually been given a site to work with.

Next: writing three small, self-contained database changes (no rebuild needed, live the
moment they're applied), each with a rollback and a proof that the change does nothing to any
site except the one narrow case it's meant to affect. Then a routine advisory review, then
applying them.

**Later the same day.** All three changes are written, proven correct against real site data
before being touched live, applied, and proven correct again afterwards against the live
system. Sent to the routine advisory review (not blocking — we didn't wait on its answer
before applying, which is normal practice here). Bug is closed. The platform's design notes
now say plainly that the fix covers three places, not the one originally planned for, and
why. The one deliberately-still-open piece (a periodic sweep that would catch any future
mismatch arriving by some route other than normal planning) is left for later — nothing for
it to find right now, and it needs a proper software rebuild rather than a database change.

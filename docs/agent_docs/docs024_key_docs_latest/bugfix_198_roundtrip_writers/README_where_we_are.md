# Where we are — bug 198, the stylesheet that keeps getting wiped

Plain-prose log for the owner. Append only, newest at the bottom.

---

## 2026-08-21 (bugfix-198 session)

You asked me to look at bug 198. Here is what it is, what I found, and what I did.

**The bug in one paragraph.** Every site has a stylesheet file, and two different agents
write it. The design agent builds it from the site's palette and layout settings. The
CSS-patch agent — the little one that fixes a contrast problem by adding a rule — keeps its
own copy of the stylesheet in the database, adds its rule to that copy, and then publishes
the whole copy over the file. The trouble is that the design agent has never written
anything into that database copy. On a lot of sites the copy was **empty**. So the patch
agent would add one rule to nothing at all, publish the result, and a 20-kilobyte stylesheet
would be replaced by about a hundred bytes. The site keeps loading, but every colour and
layout instruction is gone, so text goes black-on-black or white-on-white.

**Then it gets worse on its own.** The thing that measures contrast looks at the now-ruined
page, sees unreadable text, files more contrast problems — and those get routed straight back
to the agent that caused the damage. One site took eleven of these in eight minutes. Every
single run reported success.

**The state I found it in.** It had happened three times: one site on the 4th, six sites on
the 17th–19th, and two more this morning. Other people were already on it today — three
separate sessions had spent the day restoring the damaged sites, backfilling the empty
database copies across the whole fleet, and building a detector that will notice next time.
All of that was done before I started, and their notes said exactly what was left for
whoever picked the bug up next: **stop it happening again**. That is what I did, so I have
not restored anything — as of this evening no site is broken.

**What I changed — four things.**

1. **The patch agent now refuses to work from a bad starting point.** Before, it only checked
   "is there a stylesheet copy at all?", and an *empty* copy passed that check — which is
   precisely the hole all three incidents went through. It now checks the copy is at least a
   plausible size, and that it is not a copy shared with another site. I did not pick the
   size threshold by feel: I measured every site, and the healthy copies are all between
   13,000 and 27,000 bytes while every damaged one ever seen is under 2,400. There is nothing
   in between, so the line sits in an empty gap with a wide margin either side. Today it
   passes 19 sites and refuses 3.

2. **When it refuses or fails, it now says so.** This was the part I found most troubling.
   Because of how the workflow was wired, a refusal or an error was still being recorded as
   "completed successfully". So the eleven runs that did nothing were all logged as
   successes, and any count of "how many contrast problems are still outstanding" was
   quietly too low. Refusals are now flagged for human review, and genuine errors recorded as
   errors.

3. **The design agent now writes into that database copy.** This is the one that actually
   fixes the cause rather than guarding against it. From now on, when the design agent
   regenerates a site's stylesheet, it saves the same content into the database copy. The two
   agents finally agree. It matters because every repair anyone has done for this bug had a
   hidden expiry date — the next design run on that site would silently undo it, roughly
   weekly. **Worth flagging: our own internal documentation claimed this already happened.
   It never did.** That is now true rather than aspirational.

4. **A safety net at the last step, for any agent, not just this one.** Publishing a file
   that replaces an existing one with a fraction of its size is now something a step can be
   told to refuse. It is off by default and only the CSS-patch agent has it switched on, so
   nothing else in the system changes behaviour. This one needs the next software release to
   take effect; the other three were live the moment I applied them.

**Two sites I have deliberately left refused rather than fixed.** finetuning.uk and
gaswholesalers.com share a single stylesheet copy between them, and they serve genuinely
different stylesheets. There is no way to fill that shared copy correctly for both — filling
it from one site's file would push that site's design onto the other. **This needs a decision
from you**: give them a copy each (which is what I would suggest), or leave them as they are.
Until then the new check simply refuses to patch either of them, which is safe — it means a
contrast fix on those two sites gets parked for a human instead of being applied.

**What I have not proved.** The refusal has been tested against the live database inside a
transaction that I rolled back, and the settings verified in place — but I have not watched
it refuse a real job, because the only sites that would trigger it are the two live ones
above, and if the check were faulty the test itself would break them. I would rather record
that as outstanding than claim it.

**One mistake of mine, for the record.** When checking that the two file numbers I wanted
weren't already taken, I used a query that could never have found anything — a small syntax
trap where the search pattern I wrote doesn't mean what it looks like it means. It returned
"nothing found", which was the answer I wanted, and it happened to be correct for a different
reason. I caught it ten minutes later when the same query claimed my own just-completed work
didn't exist either. Nothing came of it, but a check that can only ever say "no" is worse
than no check, so it is written up where these get collected.

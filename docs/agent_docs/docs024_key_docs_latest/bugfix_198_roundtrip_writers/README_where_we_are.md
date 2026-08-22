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

---

## 2026-08-21, later — the two shared sites are split, and the fix is proven working

You said they could have separate theme rows, so that's done, and I then proved the fix on a
live site rather than leaving it as a claim. Both are finished and verified.

**The split.** finetuning.uk and gaswholesalers.com now have a stylesheet record each instead
of sharing one. Three things I was careful about, because each would have broken something
quietly:

- Their **design is unchanged**. The thing that actually builds a site's stylesheet reads the
  palette, layout and typography records — not the copy we were arguing about — so I carried
  those across unchanged. If I'd let the system regenerate them, both sites would have quietly
  redesigned themselves at their next rebuild, with nothing to warn anyone.
- Their **headers and footers still work**. The shared record was also pinning those, and a
  new record that didn't carry the pins would have dropped the header and footer off both
  sites. There's a known trap in our notes about a copy that carries such a pin when it
  shouldn't; this was the same trap facing the other way.
- The **original shared record is untouched** and now used by nobody. It's a template other
  sites might adopt later, so it wasn't mine to repurpose.

I rehearsed the whole change against the live database with the save deliberately cancelled at
the end, checked the result, and only then ran it for real. Both sites were serving normally
afterwards, byte for byte the same as before.

**The proof.** I wanted a real run rather than a rehearsal, but every candidate site was live
and a faulty check could have broken one — which is why I'd earlier written this up as
something I wasn't going to claim. That changed when webdesign.uk turned out to be the perfect
subject: it had the exact fault (an empty record against a real 15KB stylesheet), *and* the
domain redirects elsewhere, so its stylesheet isn't shown to anyone. A failure there couldn't
reach a visitor. I saved a copy of the file first, just in case.

I put a clearly-marked test job through the normal queue and watched it. The system picked it
up, dispatched it, and the agent **refused** — it stopped before writing anything, exactly as
intended. Three things came out of it that a rehearsal couldn't have shown:

1. It refused for the right reason, on real data.
2. **The job was recorded as "needs human review", not "completed"** — this is the part that
   was silently wrong for months, where refusals and failures were being logged as successes.
3. **Nothing was touched.** The record is still empty, the stylesheet in the repository is
   still byte-for-byte identical, and no commit was made. That's the check that actually
   matters — if the guard had failed, both of those would have changed.

I deleted the test job afterwards; it wasn't a real problem and shouldn't be counted as one.

I deliberately did **not** stage the opposite test (letting a fix through), because doing so
would have added a meaningless rule to a real site's stylesheet — the same junk-rule problem
this bug has already caused on three sites. It didn't need staging: another session watched
that happen for real on a different site the same day, correctly.

**Where the fleet stands: 22 sites out of 22 are now healthy** on this measure, up from 19 of
22 this morning. There are no sites left in the state that caused this bug.

**Still outstanding**, and I don't want it lost in the good news: the extra safety net at the
final publishing step only switches on when the next software release goes out, so that half
is written and tested but not yet running. And there's a separate known problem where the
agent writes a fix aimed at something that doesn't exist on the page — that's a different job.

---

## 2026-08-22 — the guard went live, and the bug is closed

The new release carried the last piece. I checked it at the binaries themselves rather than
trusting the version number, on both services, and the safety net at the publishing step is now
switched on and working. That was the only thing still waiting.

**So bug 198 is closed.** The cause is fixed where it starts, both writing paths are guarded, no
site is left in the state that caused it (22 of 22 healthy), and we watched it refuse a real job
without touching anything. The file moved to the closed folder with a summary at the top so
anyone finding it later gets the outcome before the history.

**One thing came out of the checking that is worth your attention.** The standard way we confirm
"did this code actually ship" — searching inside the running program for a distinctive phrase —
**gave the wrong answer**. It reported three phrases missing that were provably present. If I had
trusted it, I would have told you the safety net hadn't shipped when it had. Two independent
readings disagreeing is the only reason I caught it. That check is used across the estate for
every service, so I've written up the reliable version (copy the program out and search it
locally, or ask it what it was built from) and corrected the four places that carried the bad
recipe.

**I filed one new bug rather than leave it buried here.** When a contrast problem is found on an
element with no styling name of its own, the system records the element's *type* in a field
labelled "class". The repair agent then faithfully writes a fix aimed at something that doesn't
exist — it publishes successfully, reports success, and changes nothing. Three sites' evidence.
It's a different fault with a different cause, so it's now bug 352 rather than a loose end on a
closed one.

**Still genuinely outstanding, and honestly still outstanding:** a survey we were asked for on
5 August — which other parts of the system send a whole document through an AI model and write
the result back without checking it. The guard we built covers one place; the survey is about
the whole class. I did not do it and did not pretend otherwise.

# Where we are — the page writer and its link constraints (bug 092)

Plain prose, append-only, newest at the bottom.

---

## 2026-07-31, evening — what this is and what we did

The short version: when the platform writes a page, it is supposed to hand the writer a list
of the pages that actually exist on that site and say "only link to these". That list has
been empty on every single run for months. So the writer guessed, and the links it invented
went live as 404s.

This was not a subtle failure. The machinery to prevent it was built, wired up and running —
the step ran on every build, dutifully produced an empty list, and the prompt then quietly
dropped the whole "Internal Linking" section because there was nothing to put in it. Nobody
noticed because *every layer of it was silent*: no error, no work item, just a slightly
shorter prompt. It was recorded as a known gap in our own concept register on the 12th of
June and had not moved since.

I checked it was still real before touching anything, because bugs on this tree go stale
fast. It was: 26 of 26 recent writer runs had an empty list, the most recent one from
half past three this afternoon.

**What we changed.** Two things, and the second one is the more interesting.

The obvious change is that the step now reads the real page list out of the database instead
of hunting for it in four places where it was never going to be. There was a choice about
*which* list to read — we already have three slightly different versions of "the pages of a
site" in the code. I picked the one the deploy gate uses, and the reason is worth stating:
the gate is the thing that later decides whether a link is broken. If we tell the writer one
list and judge it against another, we get a writer punished for doing exactly what it was
told. Making the two the same list means that can't happen.

The subtler change is what happens when the list comes back empty. Before, an empty list
produced an empty instruction — which is to say, no instruction at all. That is exactly
backwards: the writer that knows least about the site was told least. Now an empty list says
so out loud: "there are no pages to link to, don't create internal links". That is true and
safe whether the site is brand new or something went wrong.

And we made "the site has no pages" and "I couldn't find out" stop looking identical. They
had the same symptom and opposite fixes, which is a large part of why this ran at 100%
failure for seven weeks without anyone spotting it. The second case now writes a proper
record to the database that a person can query tomorrow, instead of a log line that scrolls
away.

**Two mistakes I made, both worth writing down.** First, I added a cap on how many pages can
go into the prompt, and reported "how many were left out" using a number that could only ever
be 1 — on a four-thousand-page site it would have confidently told you 1 page was omitted.
Precise, measured-looking, wrong by three orders of magnitude. Second, having fixed that with
a "-1 means I don't know" marker, I left the two checks reading `> 0`, so the one case
meaning "truncated and I can't tell you by how much" would have been reported as no
truncation at all — I reintroduced the exact silent-failure shape I was there to remove.
Caught both by re-reading my own diff. Neither reached the commit.

**One thing worth flagging that isn't about this bug.** The scratch disk that all the
concurrent sessions share filled to 100% while I was working. It's a 16GB RAM disk and about
thirty sessions are using it, several holding a gigabyte or more of copied source trees —
including one session that has been finished since yesterday afternoon and is still holding
1.7GB. My own use is now down to nothing. When it fills, commands start failing in a
confusing way: the command often *works*, but its output vanishes, so it looks like a failure
that wasn't one. Someone with authority to clean up other sessions' scratch space should,
because it will bite every thread on this machine. I haven't touched anyone else's files.

**Where it stands.** The fix is written, tested (nine tests, and I deliberately broke the fix
six different ways to prove the tests actually catch each one rather than just passing),
committed, and submitted to the review council. It is **not live** — Go changes do nothing
until someone builds and rolls a new chassis image, so the bug stays open until that happens
and we can see a real run come back with a real page list. Nothing about this repairs the
pages that are already out there with broken links; it stops new ones being made.

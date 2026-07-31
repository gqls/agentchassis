# Where we are — the chrome build path (bug 167)

Plain prose, append-only, newest at the bottom.

---

**2026-07-31, evening.**

The job was bug 167: when the platform builds a page, the bit of code that decides
which header and footer to put on it could pick the wrong *kind* of component
entirely — a page section, the sort of thing that goes in the middle of a page,
served up as the site header.

This was left over deliberately. Bug 118, finished earlier the same day, fixed the
two places that *assign* chrome to a site, and its lane wrote 167 up as "we know
about this one, but fixing it changes how every page looks, so it's your call, not
ours." That was an honest and correct thing to do.

**The interesting bit is that it turned out not to be your call at all, and the
reason is a small lesson about working on a shared system.**

Before touching anything I re-ran the measurement the bug file was built on —
which component does the code actually pick today? The bug says it picks the
page-section one. It doesn't. It picks the right one, and has done since 12:39
that afternoon. What happened is that the 118 lane, a few hours before writing bug
167, had gone through the fleet and switched on the proper header component as part
of its own repair work. Switching it on changed the answer to the question 167 was
about. So the bug file described the world as it had been before its own author's
repair — accurately dated, accurately measured, and out of date within hours.

So the change is free. Nothing about any page looks different. What it removes is
the *luck*: the reason the right component gets picked today is that its name
happens to sort alphabetically before the wrong one's. Both times. Switch one
component off, or add one whose name starts with "a", and the bug silently comes
back. Now it can't.

**One thing nearly went wrong and it's worth knowing about.** The obvious way to
write this fix — and the way the bug file describes it, "one line each" — would
have introduced the very bug it was fixing, on the `<head>` of every page. The new
lookup is built to *always* give an answer, even when the answer is "there is
nothing suitable here", so that the problem gets reported rather than swallowed.
For `<head>` there genuinely is nothing suitable in the library right now, and the
answer it gives back is an 8,500-character page section. Take that answer at face
value because a value came back, and you'd have stuffed a page section into every
page's `<head>`. The fix asks the second question — *is this actually usable as
chrome?* — and falls back to the built-in default when the answer is no, which is
exactly what happens today.

**I also found something I did not fix.** There's a second route to choosing a
header, used when a site's style collection names a specific component. That route
does no checking whatsoever — it just fetches whatever id it's told to. Three live
sites (ai-agent-orchestration.com, finetuning.uk, gaswholesalers.com) are currently
pinned to a header the library says is switched off, and have been rendering it on
every page build. That one *would* visibly change those three sites, so I've
written it up as bug 170 rather than quietly changing it. It needs a look-and-see
before someone flips it. It's the same family as 118 — a switched-off component
being used — on a fourth route nobody had counted.

**Where it stands.** The fix is committed and has been through the review council
(verdict pending as I write). It is **not yet live** — Go changes need a new image
built and rolled, and I checked the running one: my change isn't in it. Another
session was mid-build at the time, and builds take committed code, so it should
ride along with theirs. I've put the exact command to confirm that in the runbook,
because "it should have shipped" is not the same as "it shipped".

**One process note, since you asked for missteps to be recorded.** I ran a command
that reported formatting was clean when the formatting check had never actually
run — a `cd` failed, the check got skipped, and the "all clear" I printed was my
own label on an empty result. It didn't cause harm, but it's the kind of thing that
makes a report worthless, so it's in WRONG_CALLS.md. The general shape: a check
that says nothing when it passes looks identical to a check that didn't happen.

**Small thing you may want to decide.** I closed 167 and moved it to `bugs_closed/`
as you asked. Strictly, this repo's rule is that a bug only moves once the fix is
*live*, not merely committed — because until it ships, the problem is still real in
production. I've followed your instruction and put a clear warning at the top of
the file saying it's committed but not yet running, with the command to verify. If
you'd rather it went back to `bugs_open/` until the roll, that's a one-line move.

---

**2026-07-31, later the same evening — the review pushed back, twice, and was right both times.**

I put the change through the review council as you asked. It came back **REVISE**,
not approved, and the two objections were both fair — worth writing down, because
"the reviewer was wrong" is the easy story and it isn't this one.

**The first** was raised by five of the ten reviewers independently. There's a
warning note in our own landmine file that names these exact three functions and
says, roughly, "fixing the selection logic here doesn't reach the page, because
there's a cache in the way". Five reviewers found it and all asked the same
question: does your fix actually do anything?

It does — the cache they're pointing at belongs to a *different* chrome path, in a
different file, and the one I changed has no cache and recomputes on every page
build. I checked rather than argued: the file I edited doesn't mention that table
once. But the honest lesson is that **I never mentioned it in the submission**, and
when a warning note names your exact symbols, saying nothing about it reads as not
having looked. One reviewer said as much, and put it well: the rest of the
submission measured everything, which made the one omission more conspicuous, not
less.

**The second** was the one that actually blocked it, and it's the one you may want
to act on. I'd found a *fourth* way the platform can pick a header, discovered it's
currently serving a switched-off component to three live sites, written it up
carefully as a separate bug — and shipped anyway. The reviewer's point was blunt
and correct: filing something honestly is not the same as guarding it, and while
that bug waits, three sites are quietly serving the wrong thing with nothing
anywhere making a noise about it.

So I've added the noise. That path now logs an error naming the site and the
component every time it happens. It still renders the same thing — I have
deliberately **not** changed what those three sites serve, because that's a visible
change to live pages and it's your call, not mine. But it is no longer silent.

There was one genuinely tricky detail worth knowing, because it's the sort of thing
that makes a well-meant fix worse than the bug. The obvious way to write that check
is to reuse the rule we already have for choosing a header. That rule excludes
"forks" — a site's own private copy of a component — and rightly so when you're
picking from a shared pool. But when a site has *deliberately pinned* its own fork,
that's the whole point of a fork. Leopardess does exactly that. So the reused rule
would have made the check's very first output a false alarm against the one site
that's set up correctly, while still catching the three that aren't. It's written
as a separate rule, with a test that fails if someone later "tidies" them into one.

**Where that leaves you.** One decision: those three sites
(ai-agent-orchestration.com, finetuning.uk, gaswholesalers.com) are on a header the
library says is switched off. Fixing it changes how they look. Bug 170 has the
details and the trap. Everything else is done, committed, and waiting on the next
image roll.

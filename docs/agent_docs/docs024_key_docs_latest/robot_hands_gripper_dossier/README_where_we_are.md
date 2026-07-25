# README — where we are: the gripper dossier pilot (robot-hands.com)

*Owner's running log. Plain prose, append-only, newest at the bottom.*

---

**2026-07-24 — designed, and the first piece built**

This is the first real build of the per-site AI idea (the per_site_ai
workstream): a paid-shaped, produced deliverable on robot-hands.com. A visitor
chats with a small assistant on a new page, describes their pick-and-place
application, leaves their email, and gets back a link to a proper engineering
report — every gripper in the site's index scored against their actual
application with the physics shown, every figure traced to a manufacturer
datasheet, and an honest "nothing in this index fits" when that's the truth.

The design is settled and written down (DESIGN doc in this directory). The
shape in one line: the public half lives on the little isolated island server
that the gauntlet work already built (the public never touches our cluster,
and the visitor's email never leaves the island); the report-building half
runs inside the platform and publishes the report as a page on the site; the
island notices the page appear and emails the link.

Decisions you made today: shared sender address for the emails; testing on
the live site approved (with cleanup); soft-launch without a nav link first;
a £/$50-a-month cap on the new AI key. The tunnel to the island is live —
your authorisation click happened. The one thing still only you can do:
issue that second spend-capped API key.

Built today: the scoring engine — the server-side port of MatchMatrix v2's
physics, reading the same verified figures the tool uses, with the whole
verdict logic (Match / Marginal / Insufficient data / No match), the
conflict-note maths, and the fact block that will keep the report's prose
honest (the writer may only use numbers from it). Tests all pass, including
the case where nothing fits and the case where a figure isn't published —
both must be said plainly, never papered over.

---

**25 July.** The whole cluster-side half of this is now built, tested and
committed. What's left before a real visitor could use it is the island
service (the public-facing bit), an image roll, and the end-to-end tests.

The most useful thing that happened today was the review council telling me I
was wrong, twice, about things that mattered.

The first round came back "revise". The gap it found: my honesty gate could
catch an invented *number* and an invented *model number*, but not an invented
*vendor name*. If the writer had padded the shortlist with "you might also look
at Piab", nothing would have stopped it — the check needs digits to see a name,
and the general-purpose fabrication scanner is deliberately switched off on
these pages (it compares against the site's own figures, and every number on a
report is calculated fresh for that one customer, so it would reject every
honest report). I'd written that gap down as an accepted limitation. The
council said, in effect: that's precisely the thing you claim this feature
does, so don't accept it quietly. It was right. That's now closed, and the
list of vendors it checks against is read from our own product data, so it
grows by itself as we index more.

It also caught me copying rather than reusing: my new "pull from the island"
code was a near line-for-line copy of the code that pulls from the traffic
probe box. I've now extracted the shared half so it exists once, and moved the
live traffic-probe code onto it too — which is the riskiest change in the
batch, since that one is running in production.

The second round found something bigger, and this one is worth your attention.
When we ask an AI model for a long piece of writing and the answer gets cut off
at the length limit, the platform *keeps* the half-answer, marks it, and reports
success — on the reasoning that a marked half-answer is better than a hard
failure. That's a fair trade, but only if whoever receives it reads the mark.
The council asked how many places read it. Nobody had checked. The answer:
**118 places ask a model for something, across 58 agents; 5 read the mark.**
Two orchestrations are carrying a truncation marker right now. So a cut-off
answer can quietly become a finished-looking piece of work almost anywhere in
the system. I've written that up as its own bug (076) rather than bolting a
patch onto this feature — the obvious fix is to make it fail loudly by default,
but that would change behaviour in 113 places at once, and nobody has measured
what would break. That measurement should come first.

I also caught two of my own errors worth recording. I stamped "reviewed by the
council" on a commit without reading the verdict — it was a "revise", not an
approval. It's in the permanent history now and I've logged it in the
wrong-calls file; the check was one query and took thirty seconds when I
finally ran it. And the report page would have gone out to a paying customer
completely unstyled: the way these pages are assembled means a component's
styling has to travel *with* the rendered page, and robot-hands.com has no
styling for a page type that didn't exist until this week. I've written a test
that renders a report and checks every single style name is actually defined —
it found two more I'd missed straight after I'd checked by eye.

Nothing here is live yet. The report pipeline is committed but inert until the
next image roll, and both scheduled jobs are seeded switched off, on purpose:
the plan is to prove the builder on a hand-made request first, and only then
let real visitor requests start flowing.

Still waiting on you for the one thing: the spend-capped API key for the
island. Everything up to that point can carry on without it.

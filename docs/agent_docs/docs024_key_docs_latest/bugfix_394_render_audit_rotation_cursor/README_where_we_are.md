# Where we are — the site audit that only ever looks at the same 60 pages

Plain-prose log for the owner. Append only, newest at the bottom.

---

**2026-08-26, late afternoon.**

Two things happened today. The first is a near-miss worth telling you about, and the second is
the actual work.

**The near-miss.** I picked up `bugs_open/359` — retired pages that are still on the internet —
ran every ownership check we have, and all of them said nobody was on it. Another session
started the identical piece of work two minutes later, having run the same checks and got the
same answer. Neither of us was careless: at the moment either of us looked, the other had
written nothing, so the checks were correct and useless at the same time.

I found out about twenty minutes in, from a plain `git status` on the directory I was writing
into — their filenames were sitting next to mine. I stood down, deleted my duplicate notes so
there is one account of that bug rather than two, and handed them the four things my read had
found that theirs had not (chiefly two recorded traps that would have shipped a mechanism that
looks correct in every test and does nothing in production). It cost about twenty minutes and
no code. I have written it up in our log of wrong calls, and the cheap fix is now in the
runbook: **claim the lane in the commit log before doing the research, not after.** I did
exactly that for the bug below, and the ownership tool now names me as the owner, so the next
session to check will get a straight answer.

**The actual work: `bugs_open/394`.**

We have an audit that opens each of a site's pages in a real browser and measures what a
visitor actually sees — unreadable text, images that failed to load, content spilling off the
side. It runs on each site every three days. It has a cap on how many pages it will do in one
go, currently sixty, because each page is a real browser navigation and they are slow.

The problem is not the cap. It is that **the audit always takes the same sixty pages.** It
sorts the site's pages in a fixed order and takes the first sixty, every single time. So the
pages beyond sixty are not "audited less often" — they have never been audited at all, and on
current behaviour never will be.

We already knew this in the abstract; a bug closed earlier this month made the truncation
*loud*, so the system now writes a line every time the cap bites. Nobody reads those lines, and
you asked yesterday for a reader. What I found when I went to write one is that the situation
is worse than the bug describes.

On webdesign.co.uk the bug recorded 131 pages with 71 unaudited, two days ago. Today it is
**146 pages with 86 unaudited** — the site grew by fifteen pages in two days, and every one of
them landed in the part nobody looks at.

And the unaudited part is not a random 86. Sorting the way the audit sorts, the site falls into
bands: six navigation pages, then ninety-four tools in alphabetical order, then forty-eight
guide pages. The cap of sixty lands in the middle of the alphabet — it cuts between a tool
called "head architect" and one called "html minifier". **Everything after that, including
every single one of the forty-five guide pages, has never been rendered by this audit and could
not be at any cap below ninety-eight.** So the whole "guide" class of page on that site is
invisible to the one check we have that sees pages the way a visitor does.

That is why I do not think raising the cap again is the answer, and the bug agrees: we raised it
once already, from twenty-five to sixty, and the site has outgrown it. A fixed number cannot
chase a site that adds fifteen pages in two days. The fix is to make the audit **remember where
it stopped and carry on from there next time** — so the cap costs us time-to-full-coverage
rather than coverage itself.

Two further things I found that the bug did not know.

There is now a **second** caller of the same audit — the design-critique agent you commissioned
yesterday — and its cap is eight, not sixty. It started truncating on the day it was born:
twice today on leopardessconsulting.co.uk, eight pages out of thirty-seven. At a cap of eight,
**twenty-five of our sites truncate**, not one. So this is really two problems wearing the same
warning message: a deep hole on our biggest site, and a shallow one across the whole estate.
Whatever we build has to serve both callers.

The bug also flagged one line it could not explain — an audit that did only five pages out of
twenty-six back on the 11th, when the cap should have been twenty-five. I have resolved it: the
run recorded its own settings, and it really was five. It was the normal agent with a one-off
override, not a broken configuration. The useful lesson is that the cap can be set per run, so
anything we build must not assume it is always sixty.

I have a plan being drafted now and will put it through the review council before committing
anything.

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

---

**2026-08-26, late evening.**

The build you deployed carried the change, and I have taken it the rest of the way. Three things
to tell you.

**It works, and I can show you rather than assert it.** Before running anything I wrote down ten
things the first run should produce — how many pages, which page it would start at, which it would
stop at, and where the bookmark would land. All ten came out right. Then I ran it a second time,
which is the run that actually matters. It began at a page called `tool-html-minifier`.

That name is the whole point. This afternoon I measured that the audit had been stopping at
`tool-head-architect` and that everything alphabetically after it had **never been looked at, and
never would be**. `tool-html-minifier` was the first page over that line. The second run started
there, and finished inside the "guide" pages — the group of forty-five that no cap short of
ninety-eight could ever have reached. Two runs took the audit from *the same sixty pages for ever*
to page 120 of 151, through a class of page it had never seen.

**Running it for real found a bug that eighteen tests could not.** The audit remembers its place
under the name of whoever asked for the work. It turns out that when I asked by hand, the system
recorded the asker as "generic", while the same run's own log recorded it as the render-audit
agent — two names for one job. Left alone, the scheduled run and any hand-run would each keep
their own bookmark and each keep starting from the beginning.

The reason no test caught it is worth your knowing, because it is a general trap: our test
scaffolding sets both names to the same string, so every test was asking a question it had already
answered. A test can only tell apart the things its scaffolding varies. I have fixed the code and
written a test that makes the two names disagree the way real life does — **but that fix needs the
next build to go out.**

**One thing is not working, and it is not this change.** The audit picks the right pages and sends
them, and I can prove that from its own records. What then fails is the step that opens them in a
real browser: it timed out on both runs. The browser service itself is healthy — it completed an
audit for a different site half an hour earlier — but it never received either of my requests. So
the pages are being *chosen* correctly and not yet *measured*. I have left the diagnosis trail at
the top of the handoff rather than guess at it tonight; another session recorded a fleet-wide
dispatch stall earlier today, so it may be the same thing.

Three smaller things owed, all in the handoff: a config file needs re-applying to the cluster, the
new watchdog has no schedule yet, and the full coverage check needs one complete cycle — three
runs — once the browser step is fixed.

The review council approved it on the third round. Both earlier rounds found something real: one
was a live defect in my own code that would have quietly attributed a page's problems to a
different page, and I would not have found it without them.

---

**2026-09-02.**

It works, and I can now show you the number that settles it.

The audit used to look at the same sixty pages of webdesign.co.uk every time, for ever. Ninety-one
pages — including every one of the forty-five "guide" pages — had never been looked at once, and on
the old behaviour never would have been.

Over the last six days it ran three times on its own schedule, with nobody watching. I added up the
pages it visited across those three runs and removed duplicates: **151 pages, out of 151 that are
live on the site. Nothing missed.** The third run finished the last thirty-seven and tidied its
bookmark away, which is exactly what it should do at the end of a lap.

I also checked that the harder way — not "did it visit 151 things" but "is there any live page it
did not visit". The answer is none.

The other thing I promised to settle was whether last week's small correction had actually shipped.
It had, and I proved it rather than assumed it: I set up a test where the two possible answers gave
completely different results — if the fix was live the audit would start at the top of the site, if
it wasn't it would start at the very last page. It started at the top. So hand-runs and scheduled
runs now share one bookmark, which was the point.

Two smaller things confirmed along the way: a second site, loanandmortgagecalculator.co.uk, crossed
the same size threshold and completed its own lap unaided; and the design-critique tool correctly
kept its old behaviour, which we deliberately left alone.

**One thing is left, and it is why I am not calling this finished.** You asked for two things back
on the 25th: the rotation, and a watchdog to read the warnings the audit writes. The rotation is
done. The watchdog is written and tested — but nothing is scheduled to run it, so at the moment it
is a smoke alarm sitting in a cupboard. It needs one small piece of deployment config, and until
that exists I would rather leave the bug open than tell you it is closed. It is the only item in
the handoff.

---

**2026-09-02, later.**

The watchdog now has its schedule. It will run every morning at 07:50 UTC, and it goes out with the
next fleet release — I have wired it so the release builds it, pushes it and switches it on in one
pass, rather than anyone applying it by hand. That ordering matters: if the schedule went live
before the image existed the job would sit failing to start, and this cluster reports that state as
"running", which is the worst of both.

**Testing it before shipping caught a real problem, and it is worth telling you what.** I ran the
watchdog against the live data first. It immediately raised an alarm: it said the rotation had been
switched off on loancalculator.co.uk.

It hadn't. That site has a single old warning from 11 August, left over from a one-off run, and the
site has since shrunk to twenty-eight pages — comfortably under the limit — so it will never produce
another warning. Its last message is frozen in the past for ever. The watchdog was reading that
frozen message as today's news, and would have raised the same false alarm every morning from now
on. A morning alarm that is always wrong is one people stop reading, which would have quietly
undone the point of having it.

So it now ignores any site that has gone quiet for more than a fortnight, and — this is the part I
was careful about — it says out loud in every report which sites it is ignoring and why. An alarm
that silently narrows what it looks at is the same failure wearing a nicer face. I checked that the
change silences only the stale site and not a live one, by testing both at once.

After the fix it reads: nineteen warnings across four sites, no problems, one dormant site named.

**Where that leaves the job you asked for.** Both halves are built: the rotation is live and proven,
and the watchdog is written, tested and scheduled. Once the next release goes out and the watchdog
has written its first morning report, this one is finished and I will close it. I am deliberately
not closing it before then — the rule here is "fixed and live", and half of it has not yet run.

One small piece of housekeeping: a shared check had been failing for a week because of a one-word
field mistake in another team's entry. I had told them about it and it hadn't been picked up, and it
was blocking a clean test run for everyone, so I fixed it by adding the missing field rather than
editing their words.

---

**2026-09-02, end of day.**

The watchdog went out with the release and is switched on. Rather than wait for its first morning
run to find out whether it works, I ran it once by hand — but from the scheduled job itself, so it
used the real image and the real settings rather than an approximation. It read the live database
(sixty-one thousand rows), found nothing wrong, correctly named the one site it is deliberately
ignoring, and wrote its record. That record is the important part: it writes one every run, clean
or not, so a missing one means it didn't run, rather than meaning all is well.

The review council approved it, and the one substantial thing they asked for was exactly that —
check it against the running pod rather than trusting the pre-shipping tests. So that is done.

**What is left before I can call this finished: one thing.** The job has never yet started *by
itself*, because it was installed after this morning's slot had passed. Its first scheduled run is
07:50 tomorrow. I want to see a second record appear with that timestamp.

That may look like a fussy distinction, and I want to be straight about why I am holding to it.
Running it by hand proves the job *works*. It does not prove the job *runs* — the timing, and that
nothing else is competing for that slot. Those are different claims, and this piece of work has
turned up two things already that only showed up when something ran on its own schedule rather than
when I ran it. So: one more morning, one query, and then I will close it.

---

**2026-09-03.**

Finished. The watchdog started by itself yesterday morning at 07:50 and filed a clean report, which
was the last thing I was waiting for. I have closed the bug.

Worth being clear about why I waited for that rather than closing it the day before. I had already
run the watchdog by hand and it worked perfectly. But running something by hand proves it *works*;
it does not prove it *runs* — the timing, and that nothing else is competing for that slot. Twice on
this piece of work something only showed up when a job ran on its own schedule rather than when I
triggered it. So it was worth one more morning.

Thank you for the three answers. I have written all three down in four places — the bug file, the
handoff, the concept register, and, for the design-critique one, into the configuration file the
watchdog actually reads. That last one matters more than it sounds: it means the decision now lives
where the code looks, not only in prose someone has to find and believe.

One caveat I want to flag on your first answer, because it is the kind of thing that quietly stops
being true. Leaving the design critic looking at the same eight pages is right *while it is run by
hand*. If it ever gets put on a schedule, the same behaviour becomes the very problem we just spent
two weeks fixing — a thing that looks at the same pages for ever. I have written that condition into
the configuration file itself, so whoever gives it a schedule will meet the warning at the moment
they do it rather than a year later.

The final count, for the record: the audit went from covering 60 pages of our largest site, for
ever, to covering all 151 — verified over three of its own scheduled runs, with nothing missed.

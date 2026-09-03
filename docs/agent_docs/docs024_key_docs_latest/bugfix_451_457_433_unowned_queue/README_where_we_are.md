# README — where we are (the unowned-bug queue)

Plain prose, append-only, newest at the bottom.

## 2026-09-03 — three bugs nobody was fixing

You asked me to look at 451, 457 and 433, and to leave alone any that already had someone
working on them. All three had lanes *talking* about them. None had anyone *fixing* them. Each
was filed by a thread that then said, in writing, that it was not going to do the fix — 457's
file literally says "Not fixed by me and not touched". So I picked up all three.

Two of them turned out to be worse than their own bug reports said, and the third turned out to
be a different bug from the one the report described.

## What shipped

**457 — the blog listing that was quietly duplicating itself.** On the boxing site, the page
that lists articles was serving thirty-six article cards where six belong: six stacked copies of
the same list, each frozen with whatever the template looked like the day it was created. The bug
report blamed the code that inserts the row. It was one level above that: the function that works
out *where* to put the listing had four ways of answering, and only one of them ever checked
whether something was already there. The other three said "nothing found", the caller believed
them, and it appended another copy on every run.

It is fixed, and the council approved it first time. The important part is what the fix does
*not* do. There was an obvious-looking improvement — record which component produced each row —
that would have made the bug **invisible instead of fixed**: the database constraint that finally
reported this only fires because those rows have no component recorded. Making them attributable
would have stopped them colliding, and the next duplicate would have appended silently. The two
halves had to land together, and there is now a test whose whole job is to keep them together.

I also found, while in there, that every site's blog listing was loading a *boxing-site-specific*
copy of the template, because the query took whichever copy was made most recently. They happen
to be byte-identical today, so nothing is broken — but the next time someone edits the boxing
one, every site would have inherited it silently.

The six duplicate rows on the live page are untouched, as you asked — that is
`site_delivery_and_editor`'s to do, and I have left them the exact six row identities so they
never have to guess which is the real one.

**433 — images that never recorded what they were.** Just over a thousand asset records carried
no file-type at all, and the ones that did were mostly wrong in an interesting way: the platform
names every generated image `.png` and tells the storage service it is a PNG, while the actual
bytes are usually JPEG. I found the writers responsible — two of them account for all of it —
and now the code that *publishes* an image reads the actual bytes and records what it really
stored, or records nothing if it cannot tell.

The "or nothing" is the whole design. Every version of this bug started as a helpful default:
there is a function in the codebase whose fallback returns "image/png" with a comment saying
that is the safest guess, and that guess is exactly why nobody noticed the mislabelling for
months. An empty field is a question someone can find. A confident wrong one looks fixed.

One thing worth knowing: this will **not** take the empty count to zero, and that is correct. An
image that has been generated but not yet published to a site genuinely has nothing honest to
record yet.

## Where I was wrong, because it changed the answer

I built a check to test the central claim behind the 451 fix. I named in advance what a failure
would look like, ran it, got a clean result, and reported to you that it passed.

It could not have failed. The claim was "when this repair succeeds, the problem is genuinely
gone". My check asked "does the problem come back within three hours of a success?" — but the
checker only runs once a day, so a problem that *never went away* looks exactly like a problem
that came back for a good reason. I picked a stopwatch to answer a question that was not about
time.

The real answer was sitting in the data the whole time: the checker records *which* things
drifted. Same things twice means the repair did not work. Different things means it did, and
something new changed.

I also repeated three facts from documents I had not opened, one of which I had drafted into a
fleet-wide ledger as a criticism of another team. All three were wrong. And a quick script I
wrote to census which parts of the code were compliant cleared the one file that was actually
breaking the rule, because a log line three rows below the code mentioned the same word.

All of it is written down where the next person will meet it.

## The decision I need from you — 451

This one is genuinely yours, and the rest of this section is the case, laid out so you can rule
on it without reading any code.

### What the thing is

The platform watches your sites for problems and files a job for each one it finds. If a job for
the same problem on the same site gets filed a **third time within a week**, the platform gives
up on it: it marks the new job "unresolved" and never sends it to anyone. The thinking is
reasonable — we have had two goes at this and it is still broken, so a human should look.

### The rule that governs it

A "go" counts whether the previous attempt **succeeded or failed**. You ruled on 2026-08-24 that
this stays exactly as it is, and the reasoning was sound: a repair that reports success without
actually repairing anything must still count, or the platform pays to re-run a useless fix every
cycle, for ever.

### How this case measures against it

There is one kind of job — "this site's shared header and footer are out of date" — where
success really does mean fixed, because the same code that rewrites the header also updates the
record the watcher compares against. So a site that gets two *successful* header updates in a
week stops receiving them, silently, and the job that would have delivered the third is marked
"unresolved" the moment it is created. Twelve sites are in that state, and it is why the analytics
tag and the cookie banner never arrived on some of them.

**But I checked whether "success means fixed" is actually true for that job, and it is not
reliably true.** There are at least three routes through the repair code that report success
without updating the record — including one the code itself calls a "degraded success", where it
notes the header is *stale but still being served*. So some of those successes really do leave
the problem sitting there, and for those the platform is right to stop.

That is the whole difficulty. The clean fix rests on something that is true most of the time.

### Your options

**A — fix the repair first, then exempt.** Close the three routes where success is reported
without the work being done, so the claim becomes true, and only then let this job type skip the
ladder. Slowest, and the exemption afterwards rests on something solid rather than something
usually-so.

**B — exempt on evidence rather than on trust.** Do not assume anything about the repair. The
watcher already records *which* inputs drifted; compare them between one job and the next. Same
things = the repair did not work, count it. Different things = something new changed, do not.
This needs no assumption at all, and it would work for any similar watcher later. More work than
A up front, and it is the version I would build if you want this closed properly.

**C — leave 451 alone for now** and treat the bigger number as the real question (below).

I have already shipped the part of 451 that is safe under any of these: the plumbing defect
underneath it, where a piece of internal bookkeeping was hand-copied field by field with nothing
to notice when a field was forgotten. That is what let this whole situation exist unnoticed.

### The bigger question underneath, which is also yours

When you made the 2026-08-24 ruling, the visible population was **747** jobs. Today it is
**5,870**, across **32** different kinds of job, growing by roughly **1,700 a week**. About
**2,391** of those were parked after two attempts that both *succeeded*.

I am not asking you to reverse the ruling — the reasoning behind it is still right for a large
part of that population, and I can show you the job types where the platform is behaving
correctly. I am telling you the number it was made on has moved by eight times in eleven days,
and that nothing in the platform can currently answer "how many of these were parked by success?"
without reconstructing it by hand, because the only marker is a phrase in a text field that
nothing reads.

**And there is one narrower finding that I think is worth ruling on separately, because it does
not touch the population your ruling protects at all.** When a watcher notices a problem has
fixed itself, the platform closes the job by marking it "complete". Those self-corrections feed
the same counter — so two problems that resolved on their own in a week cause the third genuine
finding to be parked, blaming a repair that was never even attempted. Nineteen watchers work this
way. That is a straightforward miscount rather than a judgement call, and fixing it leaves your
2026-08-24 reasoning entirely intact.

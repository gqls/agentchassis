# SUMMARY 2026-08-22 — the unread finding codes (bugs_open/358)

Written to be read aloud. Current state only; the chronology is in
`README_where_we_are.md` and the evidence in `NOTES_unread_finding_codes.md`.

## What we're trying to do

When some part of the platform notices something wrong that it won't fix by itself — a page
refusing to deploy, a plan losing a section name, two candidate values in conflict — it writes a
note about it into a table, under a name. Writing that note is cheap and it *feels* like
detection. The complaint behind this piece of work is that for most of those names, nothing ever
reads the note, and a housekeeping job deletes it after thirty days. We were paying to detect
things and then throwing the detection away unread.

The goal was never "write a reader for every one of them". Some of those notes are legitimately
just for a human to look at; some are deliberate measurements that somebody is already watching;
and some are ordinary error plumbing, where being read in bulk *is* the correct use. The goal was
to make it impossible for a new kind of note to appear with **nobody having said which of those it
is** — and for nobody to notice.

## Where we've come from

The complaint arrived as a careful census: sixteen kinds of note with no automated reader, and a
thirty-day deletion clock. That census holds up. Three of the things it said *around* the census
did not, and finding that out is most of what this work was.

The important one is about where the fix should live. There is a piece of code in this system
whose own documentation calls it "the ONE writer" of these notes — so the obvious place for a
check was inside it. It is not the one writer. There are five, and one of them *cannot* use it,
for a technical reason that isn't going away. A check in the obvious place would have covered one
writer in five and looked complete. That is the original bug happening one level up: a thing that
feels like detection and isn't.

The second: the complaint's own headline example — the loudest, most frequent note, nine and a
half thousand of them in five days — was described as something nobody had ruled on. Somebody
had. It's a deliberate measurement with an owner, six dated readings, and a decision you made on
18 August. Filing it as waste would have been the first mistake. The third: a note the complaint
holds up as a *success story* turns out never to have been switched on at all — the database
change that would have installed it was never applied. Its zero occurrences read as "quiet" and
actually mean "absent".

## What we've done

There is now one file that lists every one of these note-names and what each is for: something
automatic reads it, or it's a deliberate measurement with an owner and a review date, or it's for
humans to read by hand and we accept the thirty-day deletion, or it's ordinary error plumbing. A
name that turns up in the database without an entry is a failure somebody has to answer.

The check reads the **live table** to find out what has actually been written, rather than
reading the source code to guess. That is the design decision the work turns on, and it came from
failing at the alternative: three careful attempts to census the source code disagreed with each
other, one of them returning test fixtures as if they were real. Asking the database is the one
version that cannot be got wrong by a writer spelling something unexpectedly.

The entries are built so they can't be satisfied by typing. If an entry claims something reads
the note, the check opens the named file and confirms the name is actually in it. If an entry
claims a time-boxed measurement, the review date expires on its own. It also retired two
hand-maintained lists of these names that lived inside test files — both had quietly stopped
matching reality within days of being written.

It went through the review council twice. The first round was sent back, correctly: I had
described two changes that depended on a piece I hadn't listed, so as written they wouldn't have
compiled. The second round was approved, with three advisory notes, all three checked rather than
waved through.

## Where we are now

Live and running. Forty-two kinds of note are being written today; all forty-two are declared;
**thirty-two of them are still marked "nobody has decided yet"**, and that number going down is
the measure of progress from here. The daily automatic run is deliberately *not* built yet.

Two things happened while we were writing this up that are worth saying out loud, because they
are the argument for the whole exercise. First: a whole category of note — twenty-five of them,
the oldest in the table — was **deleted by the housekeeping job during this session**. Not
hypothetically, not "at some point": between two runs of the check this afternoon, its entire
output was erased, unread, for ever. Second: another kind of note was added to the system on the
same morning the complaint about unread notes was written. Nothing stopped either.

I should also report that I broke the build for everybody for about twenty minutes on the way.
Several sessions edit the same files simultaneously; the safe way to commit is to name your files
explicitly, which I did — and naming a file takes whatever is in it at that instant, including
another session's half-finished sentence. It compiled perfectly on my machine; the version that
was broken was the one I never tested. It's fixed, the affected session has landed their work
properly, and the check that would have caught it in seconds is written down.

## Where we're going

Three things, in order of who has to decide them.

**Mine to build, not started:** the daily automatic run. Until it exists this protects whoever
runs it by hand, which is the same state a sibling check shipped in and is honest about.

**Yours to rule, and this is the real remaining work:** the thirty-two undecided names. Each needs
one judgement — is something going to read this, is it a measurement with an end date, or are we
knowingly keeping it for humans and accepting it vanishes in a month? They don't need doing at
once, and the count is now visible so progress is legible.

**A boundary somebody else is waiting on:** another piece of work routed a dependency at this one
that this does not answer — it's about a different channel that leaves no record at all, so
"was it read?" has nothing to query. That lane has since marked its own trigger unarmed rather
than waiting. Worth knowing the answer isn't coming from here.

One caveat on the whole thing, and it came from a session that decided *not* to use this
mechanism: saying who reads a note tells you nothing about whether the thing writing it can see
everything it should. One detector only fires when a page is built, so components that never get
built are invisible to it permanently, however faithfully its notes are read. That is a separate
question and this work does not answer it — stated in the file itself so it can't be mistaken for
a clean bill of health.

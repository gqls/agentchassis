# Where we are — making the silent hero/logo readers speak up

Plain prose, append-only, newest at the bottom. The owner maintains this too — add below, never
rewrite above.

---

## 2026-08-11 — the small job that turned out to be the important one

This is commission item 2, the one you approved with "2. yes." It is the smallest of the five and
it has quietly become the one that unblocks the big one.

**The problem in one sentence.** Three places in the code take the web address of a freshly
deployed hero image or logo, and when that address isn't there they do nothing whatsoever — no
complaint, no note, nothing. So a page that shipped with no hero and no logo looked exactly like a
page that never wanted one. That is how this went unnoticed for five weeks.

**What I changed.** Those three readers now say something when a deploy actually happened and its
result came back without the address. They say which key they wanted and — this is the useful part
— which keys the result *did* carry, because that list is a fingerprint: seeing
`response / response_status / response_received_at` and no address tells you immediately that this
is the known fault and not something new.

**They stay silent when no image was wanted**, which matters more than it sounds. Most pages
deploy no hero and no logo, and the naive version of this change would have filed a complaint on
every page of every site — thousands of them, burying the handful that mean something. So the test
is "did something actually try to deploy an image", and the presence of the result is what answers
it. There is a test that fails if that gate is ever removed, and I proved it fails by deliberately
breaking the gate and watching it go red.

**One judgement call I want to flag, because it goes slightly beyond what you approved.** The
commission asked for a log line. I've added a log line *and* a durable record in the database. The
reason is that a log line here would not survive long enough to be read. Two measurements, neither
of them mine:

- the service these readers run in is the busiest we have, and its own start-up line was already
  measured as having scrolled out of reach within hours;
- the run record that holds this evidence is deleted after **four hours** — because it sits in the
  "waiting for a reply" state, which is the shortest-lived category in that table. That is also
  the real explanation for something that had puzzled us: this fault kept appearing and vanishing
  (nothing on the 6th, two cases on the 9th, nothing on the 11th). It was never coming and going.
  It was one four-hour window opening and closing.

So a log-only version would have reproduced the exact problem it exists to solve: evidence that
evaporates before anyone reads it. The database table I'm using is the one the codebase itself
documents as the only one that outlives this kind of step. I've put this at the very top of the
council submission rather than hiding it in the small print, so if the reviewers think I overstepped
they will say so plainly.

**Why this is now the unblocker.** Item 5 (last week's job) fixed the diagnosis tool, and the
re-run proved the tool works and the evidence is gone. Chasing the evidence by looking for it is a
race against a four-hour deletion, and we have lost that race three times. This change stops
looking and starts catching: the next time it happens, it writes itself down at the moment it goes
wrong, into a table nothing prunes on that schedule.

**Something I found on the way, which is not mine to fix.** There is already a workaround in the
deploy code — added back in February — that writes the address to a second place precisely because
someone had noticed the first place gets overwritten. And the codebase separately records that this
*second place is also wiped* for exactly this kind of awaited step. If both of those are true, that
is a real candidate for the root cause we have been unable to pin down, and it has been sitting in
two comments that nobody had read side by side. I have written it into the bug file as a lead
rather than acting on it, because the fix is the design decision you have reserved for yourself.

**An incident worth knowing about, since it is about how we work rather than what we built.** While
I was mid-verification, another session committed my four files into the repository under its own
commit message — it swept up everything in the shared workspace, including work in progress. Nothing
was lost. But the timing was uncomfortably close: minutes earlier I had been *deliberately breaking*
one of those files to prove a test could catch it, and if the sweep had landed then, a knowingly
disabled safety check would now be in the codebase under someone else's commit, with my own notes
showing a clean test run. I checked instead of assuming — the version that landed is the correct
one, and a clean copy of the repository builds and passes. I have written the lesson into the
runbook: when you break something on purpose, restore it and verify the restore in the same breath,
not at the end of the session.

**Where it stands.** The code is in the repository and will go out with the next fleet release. It
is submitted to the council. The one thing that will actually prove it works is a real site build
that deploys a hero or a logo, after the release — and then a record appearing at the moment it
fails. I cannot force that from here.

---

## 2026-08-11 (evening) — approved, and the reviewers earned their keep

The council approved it, first round, in about eleven minutes. No serious objections. Several
reviewers went out of their way to say the judgement call I flagged — adding the durable record on
top of the log line you approved — was the right way to handle a scope decision: declare it and let
someone rule, rather than slip it in.

**Two reviewers pushed back in ways that were genuinely useful, and neither could be answered by
just agreeing.**

The first caught me repeating a number instead of checking it. The four-hour deletion window is the
fact my whole argument rests on, and I had taken it from another session's notes. The reviewer's
point was blunt and correct: it should not be treated as settled just because it was stated with a
specific number. I went and read the live cleanup job. **It is four hours** — the claim holds, and
it is now something I have measured rather than something I inherited. Worth recording that the
reviewer had no way to check it either; it flagged the gap rather than waving it through.

The second asked a harder question: I fixed the three places the bug report named — but is the same
silent-failure pattern lurking elsewhere? Rather than promise to look, I counted. There are 64
places in that package using the same shape. Almost all are reading configuration, not results, so
they are irrelevant. Four are genuinely reading the result of a dispatched job, and I opened all
four. **None of them has this bug.** Two treat a missing value as "then do the other thing" and go
and do it. One fails outright with a clear error. So the real distinguishing feature isn't the code
shape at all — it is whether a missing value has any consequence beyond the page quietly coming out
worse. That was true in our three, and false in the other four.

Four more I could not classify, because they look up a key that comes from configuration rather
than being written in the code, so you cannot tell what they mean by reading them. I have written
those down as unfinished rather than counting them as clean.

**One reviewer suggestion I did not take**, and I want to be upfront about it: it asked me to leave
a note in a particular shared table. There is a proper tool for doing that for one category of note
and none that I could find for this one, and hand-writing rows into a shared table is exactly the
habit that tool exists to prevent. So the reasoning lives in six places a reader will actually look
instead. I have logged that as an open loose end, not as done.

**Still outstanding:** the release, and then a real site build that deploys an image, which is the
only thing that will prove any of this works in anger. And separately, the lead I mentioned — I have
put that to the automated diagnosis loop rather than write it up as fact, and its answer is still
pending.

---

## 2026-08-12 — it's live, and the automated diagnosis told us something better than an answer

**Item 2 is live.** The new build went out overnight and I checked the running service itself rather
than trusting the version number — the new record type is compiled into both copies of the service,
and I ran the check alongside a deliberately impossible search to prove the check isn't just saying
yes to everything.

**It has not fired yet, and I can tell you why that means nothing.** There are no records — but
also **nothing has deployed a hero or a logo since the release**. So the path hasn't had a chance to
run. That is the difference between "nothing broke" and "nothing was tried", and it is only visible
because I asked the second question alongside the first. A quiet result here is not yet good news;
it is not news at all. It still needs a real site build.

**Now the part worth your attention.** That lead I put to the automated diagnosis loop came back
"unverifiable" — but read what it could not do, because it is more useful than an answer would have
been. It said, in effect: *I cannot see the code you are asking me about.* It was given one function's
body and a single line of another, and nothing at all for the two that mattered.

So I checked whether the code was missing from our searchable index. **It isn't.** All four functions
are there in full, with correct line numbers. The index had everything; the evidence pack the loop
actually reads passed on almost none of it.

**That is the same fault I fixed last week, one layer over.** Last week's job was that the pack
listed the columns of one database table while showing rows from six — and, crucially, never said it
was showing you a filtered view, so "not there" and "not included" looked identical. This is that
exact shape again, in the code half rather than the data half: it holds four functions and shows
one, and the loop is instructed to abstain rather than guess when it cannot cite something. So it
abstained. Correctly.

**And I have to own a mistake here.** Last week I wrote in the bug file that this very blocker was
"clear", and my evidence was that the index was fresh and carried the functions. That was true, and
it was an answer to the wrong question — the loop had complained about the pack, not the index. A
fresh index says "present" whether or not the pack passes it on, so my check could never have come
out any other way. It has cost one diagnosis run to be told the same thing again. I have corrected
the bug file and logged it in our wrong-calls log, including the check I should have run: read the
pack itself, which was sitting in the database the whole time.

**Where that leaves the lead:** neither proved nor disproved. It is still the best explanation we
have for both halves of the original bug, and it is still marked unverified. Chasing it again through
the same loop will fail the same way until the code half of the evidence pack is fixed — so that is
now the thing standing in the way, and it is a fault in our own diagnosis tooling rather than in the
image code.

**What I'd suggest next**, though the choice is yours: fix the code tier of the evidence pack. It is
the same shape as the fix that already went through review and worked, it unblocks this lead, and it
unblocks every future question whose answer lives in a function body — which is most of them.

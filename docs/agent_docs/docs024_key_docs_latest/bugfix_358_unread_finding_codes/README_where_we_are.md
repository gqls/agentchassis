# Where we are — the unread finding codes (bug 358)

Plain-prose log for the owner. Append-only, newest at the bottom.

## 2026-08-22 — picked up, and the bug is still real

The bug, in a sentence: lots of parts of the platform, when they notice something wrong that
they won't fix themselves, write a note about it into a database table — and for sixteen
different kinds of note, nothing ever reads the note, and a cleanup job deletes it after
30 days. We pay to detect things and then throw the detection away unread.

Today I checked the bug is still valid before planning anything, because things move fast
here. It is. Two small things changed since it was filed this morning, and both are actually
good news: the other team working on content loss shipped their checker, and it is the
first thing EVER to use the table's "resolved" tick-box (45,000 rows and nobody had ever
ticked it before today). That team's checker is the model citizen: it writes its findings
AND reads them back AND acts on them. The sixteen orphaned kinds of note are all still
orphaned — and a seventeenth was added by another session this very morning, which is the
whole point of the bug: nothing stops a new one arriving without a reader.

Next: research how the one similar guard-rail we already have was built (the "optional key
budget" checker), then write the plan and put it past the council.

## 2026-08-22 (later) — the fix is built and running, and I broke the build on the way

**What the fix is, in plain terms.** Every one of those notes the platform writes to itself now
has to say, in one file, what it is FOR: something automatic reads it, or it is a deliberate
measurement with an owner and an end date, or it is there for a human to look at by hand and we
accept it will be deleted in 30 days, or it is ordinary error plumbing. A note that turns up in
the database without any of those is a failure that someone has to answer for. Adding a new kind
of note without saying which it is now trips the check the next morning.

**The important design decision, and it was not the obvious one.** There is a piece of code in
this system that describes itself as "the ONE writer" of these notes — so the obvious place for
the check was inside it. I counted, and it is not the one writer: there are five, one of which
*cannot* use it for a technical reason that will not go away. A check in the obvious place would
have covered one writer in five and looked complete. So the check asks the database what notes
have actually been written, which is the one question no amount of clever code-reading can get
wrong. It found 43 kinds; 32 of them nobody has yet decided about, and that number going down is
the measure of progress.

**Three things in the original bug report turned out to be wrong**, and I have corrected them in
place rather than quietly working around them. The biggest: the report's own headline example —
the loudest, most frequent note, 9,617 of them in five days — is described as something "nobody
has ruled on". Somebody had. It is a deliberate measurement with an owner, six dated readings and
a decision you made on 18 August. Filing it as waste would have been my first mistake. Another:
a note the report holds up as a *success* story turns out to have never been switched on at all —
the database change that would have installed it was never applied, so its "zero occurrences"
reads as "quiet" and actually means "absent".

**Now the bad part, and it is mine.** Committing my work broke the build for everybody for about
twenty minutes. Several of us edit the same files at the same time, and the safe way to commit is
to name your files explicitly — which I did, correctly. What I had not appreciated is that naming
a file takes whatever is in it *at that moment*, including another session's half-finished
sentence. I took two of them: a comment they were mid-way through typing, and a reference to a
piece of code they had written but not yet committed. Neither was visible to me, because on my
machine all the files existed and everything compiled. The version that was broken was the one I
never tested — the committed one.

I found it because an unrelated automated check complained in a way that did not match its own
description, and I read past the headline. I fixed it in two steps and verified the real thing
this time by extracting the committed version and building *that*. I chose to fix it by removing
my own premature reference rather than by committing their unfinished work, which would have put
a whole feature of theirs under my name — repairing a mistake should not make it bigger. I have
messaged that session directly, left a note where their code was so they cannot miss it, and
written the whole thing up in the shared log of mistakes, including the one-line check that would
have caught it in seconds.

**Where this sits.** The reviewers said "revise" on my first submission and they were right — I
had described two changes that depended on a piece I had not listed. That is exactly what the
review is for, it cost one resubmission, and the second round found a second omission of the same
kind. The revised version is with them now.

---

**2026-08-22, evening — the notes that were about to be deleted, and what they turned out to say.**

You made four calls this afternoon. Save the expiring evidence and find out whether the problem
it describes is fixed or whether the detector has gone deaf. Give deliberate findings a longer
life than ordinary error plumbing, and stop marking something "resolved" making it die sooner.
Let me propose what each of the thirty-two undecided notes is for, with the evidence attached, and
you ratify in batches. And leave the backlog count visible rather than enforced until we have seen
how hard a batch actually is.

The first one is done, and it did not come out the way I expected.

The forty-one notes are saved, in full, before Tuesday's deletion. They cover twenty-four review
rounds and twelve reviewer seats, and the great majority came from the fix-loop's own council
rather than the one that reviews platform changes.

Then the interesting part. These notes are filed under a name that says a reviewer's answer was
**cut off** — ran out of room mid-sentence. That is what the code believes about itself, and it is
wrong for forty of the forty-one. I went and looked at what the reviewers actually said, which we
keep word-for-word: in the whole retained history, only **five** reviewer answers have ever run out
of room, four of them in mid-July before this detector even existed, and one on 2 August. Nearly
twelve thousand answers end properly. So what the notes really record is "this reviewer's answer
came back unreadable" — and the reason was never established. The name asserted a cause nobody
measured.

On the question you actually asked — fixed, or deaf? **Fixed.** The same piece of code that writes
these notes also writes a short report on every single review round, and that report counts
unreadable reviewers whether or not there are any. Those reports are still being written — two
hundred and forty-eight of them last week — and the count of unreadable reviewers goes seventeen,
twenty-four, then zero, zero, zero. That is the difference between a thing that is quiet and a
thing that is broken: I can see the mechanism running and reporting nothing, rather than just
seeing nothing. If it had gone deaf, those reports would still be finding unreadable reviewers
while no notes were being written. They aren't.

What appears to have fixed it is that the reviewers' room to answer was doubled over the same few
weeks. The timing lines up closely and the reason is sensible — more room, complete answers,
readable results — but I could not find the change that did it, so I have written that down as a
strong hunch rather than a fact. There was also one reviewer seat on 2 August configured with
almost no room at all, which guaranteed it would fail; that has since been corrected and is not a
live problem.

One thing I got wrong on the way and want on the record. My first attempt to measure "did the
reviewers run out of room" compared two columns that are, it turns out, completely empty for these
records. It returned a clean zero every week and looked like a result. It was a question that could
not have produced any other answer. I only caught it because I checked whether the columns had any
data in them before believing what they said — which is the check I should have run first, not
second.

Next: the retention change, which is a single guarded database edit and goes to the reviewers
because it is live the moment it applies; then the first batch of proposed rulings for you.

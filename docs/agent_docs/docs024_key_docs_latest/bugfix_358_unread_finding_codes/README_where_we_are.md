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

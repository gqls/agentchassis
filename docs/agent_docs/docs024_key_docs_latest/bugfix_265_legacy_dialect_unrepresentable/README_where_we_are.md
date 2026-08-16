# Where we are — bug 265 (plain prose, append-only, newest at the bottom)

## 2026-08-16

I was pointed at bugs 213 and 274, but both had already been closed by other threads the day
before, so I went looking for the next open bug that nobody was working on and that was a
framework problem rather than a single-site one. That was 265.

The bug in one breath: components describe their content fields in a small JSON document.
There is a current shape and an old shape. A code comment said the old shape was extinct.
It wasn't — people kept seeding new components in the old shape by hand — and the alarm meant
to notice was a log line nobody reads.

What I found that the bug file didn't know: it had guessed the old shape was coming from the
component-generating agent. It isn't. Every one of the four cases was a hand-written SQL seed
by a session. That matters, because a check inside the agent would have stopped none of them.
The only place all of them pass through is the database table itself. So the fix is a rule on
the table: the old shape can no longer be stored, by anyone, by any route. The three remaining
old-shape rows are converted to the new shape in the same step, in a way that changes nothing
about how they behave (proven by a test that reads both shapes and gets the same answer). The
comment now points at the rule instead of at a count that went stale in four days.

Two small things came out of it that belong to other people: one component says its body is
never written by the AI, but the schema treats it as if it were (an over-report, not a miss);
and one header component has its fields written without the wrapper the reader expects, so the
reader can't see them. Both are noted for their owners, neither is dangerous today.

Council review is submitted; the database change is ready to apply after the verdict; the Go
part rides the next chassis roll.

## 2026-08-16, later

A fresh build went out and it carries the code half, so this is finished. To be sure rather
than hopeful I asked the two running containers what they were built from — and, more usefully,
checked that a sentence the fix DELETED is genuinely gone from both of them, which is the part
that distinguishes "the new code is running" from "a word I searched for happens to be there".
The database rule went in this morning and I proved it by trying to break it: an attempt to
store a component in the old shape was rejected, and two real components written by the system
this afternoon went in perfectly well. So it refuses what it should and allows what it should.

One honest caveat, written into the record rather than glossed: there is a second, smaller
guard inside the code that will probably never fire, because the thing it guards against has
never happened in the system's history. It is tested and it is in the running binary, but
nobody should later read its silence as evidence it does not work.

I also caught myself in a small mistake worth keeping. Tidying up, I updated a stale reference
inside the database migration file itself — which had already been run. The system records a
fingerprint of each migration exactly as it was run, so my tidy-up would have made a correct,
successfully applied change look tampered with. Reverted, fingerprint matches, and the lesson
is written down: once a migration has run it is history, not a document to improve.

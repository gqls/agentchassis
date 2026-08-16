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

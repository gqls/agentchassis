# Where we are — work-item completion integrity

*Plain-prose log, append-only, newest at the bottom.*

---

**2026-07-18 — what this is about**

Some work items were marking themselves finished while carrying, in the same database
write, the error that proved they had done nothing. The improvement loop then believed
those defects were handled and stopped re-detecting them. So the platform was quietly
telling itself it had fixed things it hadn't touched. That's the bug.

It came in two halves. One action (`fix_forced_text_colors`) had been written but never
added to the registry, so every attempt to run it failed validation with a confusing
message about needing a "topic" — the workflow never ran at all. That's the small half.
The big half is that when those failures came back, the completion code looked at the
wrong field. There are two `status` fields one layer apart: one says "a reply arrived",
the other says "the work succeeded". It was reading the first and treating it as the
second.

**What I found that the bug report didn't**

The report said the cause was two hand-maintained lists drifting apart. That turned out
to be wrong — there was only ever one live list; the other was dead code that nothing had
called in a long time. What made it look like drift was a comment at the top of another
file telling developers to "register in TWO places". That comment had also been copied
into two guide documents. So the misinformation had outlived the code and was still
actively misleading people — including whoever filed this bug. All of it is now deleted.

The report also said two items were affected. There were 54, across six live sites, going
back to May.

**What I changed**

The completion code now refuses to mark something finished if the work itself reported
failure, and instead sends it back through the existing retry machinery. I picked the rule
by measuring rather than guessing: across the entire database history there is only one
failure word in use, and over the last 30 days the new check would have stopped 6 out of
1662 completions — all six genuinely broken. So it won't start rejecting healthy work.

I also added a build-time test so that an action which is written but never registered now
breaks the build, instead of failing silently in production months later. Another thread
hit the identical problem with a different action recently, which is why a one-off fix
wasn't enough.

**A decision I brought to you, and why**

The 54 bad rows needed correcting. You initially said re-queue them so they'd actually
run. I pushed back once, because that would have fired an action that has never once
executed successfully at five live sites — including a client rebuild and a site another
thread had just finished restoring. You settled it: mark them all failed and start fresh.
That's what I did, reversibly. Discovery can now re-file them cleanly if the defects are
real.

**Where I got it wrong**

Twice, and both are worth you knowing.

I told you the council resubmissions were being silently dropped and went hunting for a
transport bug. They weren't dropped — they were queued, about 16 minutes behind a backlog.
I resubmitted three times chasing hypotheses I hadn't tested, which turned one review into
four councils' worth of credits.

And I made a structural claim — that only one place in the platform could have this
bug — based on reading four functions and guessing about three more from their filenames.
The council's reviewers objected to exactly that, twice, and I brushed it off as
box-ticking. When you asked me to re-read CLAUDE.md I went and opened the three files. The
claim was right, but I hadn't earned it, and the reviewers were doing their job.

**What's left**

The Go changes don't do anything until a chassis image is built and rolled — that's a
fleet-wide action with other threads' work queued behind the same roll, so it's your call
when. Until it ships the bug stays in the open queue, because it's still reproducible in
production. After it ships, the check is to grep the running pod for the new function, not
to trust git or the image tag.

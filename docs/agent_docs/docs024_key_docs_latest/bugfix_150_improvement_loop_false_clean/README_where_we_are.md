# Where we are — the improvement loop saying "site is clean" when it isn't

*(Plain-prose log, append-only, newest at the bottom. Owner's document.)*

---

**2026-07-31, evening — session "bugfix 27"**

We have a loop that is supposed to sweep a site, find what is wrong with it, queue the
fixes, and finish by re-rendering the pages so the fixes actually appear. It has been
ending with the message **"No issues found — site is clean"** on sites where it had just
queued dozens of problems.

The reason turns out to be a small thing about who does the work. The step that moves
findings from "spotted" to "queued" exists in three different agents, and it is greedy: it
moves *everything* for the site, not just its own findings. The main loop calls two helper
agents first, and only then runs its own copy of that step. By that point there is nothing
left to move, so its copy quite honestly says "I moved nothing" — and the loop reads that
one sentence as "the site is fine". It then skips the closing re-render and reports success.

Nobody noticed for months because there is a separate scheduled job that picks the queued
fixes up anyway, every two minutes. So the fixes did happen. What was lost each time was
the final re-render pass — the thing that guarantees the fixed content actually reaches the
live pages — and, less visibly, the truth: anything reading the loop's result was told a
site with open problems was clean.

I ran the loop once, deliberately, at vetcomparison.uk, on the current code, to watch it
happen rather than take the previous session's word for it. It did exactly what was
predicted: the first helper queued 24 findings, the second queued 3 more, the main loop's
own copy found nothing left to queue, and the run ended on "site is clean". Twenty-seven
findings queued in that run; zero closing re-renders created. That matters beyond this bug
— the original report had honestly marked "this happens every time" as an *assumption*,
because the database had already deleted the history that would prove it. Now there are two
independent sightings, on two sites, two days apart.

**The fix.** Rather than argue about which agent should be allowed to do the queueing, I
made the step able to answer a different and better question: not *"did I personally queue
anything just now?"* but *"does this site have work waiting?"* — which is true regardless of
who queued it, in what order, or which agent gets that step added next. The loop will read
that instead. The old signal is left exactly as it was, because three other places in the
system use it correctly to mean "my own results were empty", and changing its meaning to fix
one of them would quietly break the other three.

**What is not done yet.** The code half is committed but does not do anything until a new
chassis image is built and rolled out, and you asked me not to roll one (a roll interrupts
other sessions' work). The one-line configuration change that switches the loop over is
written and **deliberately not applied** — applied too early, against the old code, it would
make *every* run report "clean", which is worse than the bug. So the ticket stays open. What
is owed is two steps, both written down: roll an image containing the commit, then apply the
config and re-run the same sweep to watch the loop take the other branch.

I also found a second way the same false "clean" message can appear: if a site has already
been audited three times, the loop skips straight to the end and says "clean" without
looking at all. No site is currently in that state, so it is a trap rather than a live
problem — recorded, not fixed, because it needs a different answer (an honest "we skipped
this" message) rather than the change I made here.

The change has gone to the review council; the verdict lands in about half an hour and I
will act on it whatever it says.

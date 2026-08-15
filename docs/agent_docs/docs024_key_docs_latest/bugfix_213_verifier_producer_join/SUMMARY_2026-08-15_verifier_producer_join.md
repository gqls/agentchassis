# SUMMARY 2026-08-15 — bugfix 213: both halves built, and the file is closed

*Written to be read aloud. Previous in the series: `SUMMARY_2026-08-14`, `…-08-11b`, `…-08-11`.
This one marks a real turn — the work is finished and the bug is closed, which is the first time
either has been true.*

---

## What we're trying to do

Make sure a ticket that says "done" means something.

The platform files work items when it finds a fault on a site — a page with dark text on a dark
band, a heading that fails a contrast measurement — and hands each one to an agent that is
supposed to repair it. The whole system trusts the ticket's status. When a ticket closes green,
nothing looks at that page again, and a two-strike rule actively suppresses re-detection of
faults that keep coming back. So a ticket that closes green without anything having been fixed
does not just lose one repair; it teaches the platform that the page is fine.

This bug was about one way that happens: a ticket can be graded by a check that was written for
somebody else's tickets, or by no check at all.

## Where we've come from

The original finding, on the 7th of August, was that **two different producers were filing
tickets under one name**. A design audit and a discovery check both used the label
`hardcoded_section_colors`, but only the discovery check's question was ever asked before a
ticket closed — and that question was about the whole site, not about the one section the design
audit had complained about. It answered its own question correctly and returned "yes, fine", so
every design-audit ticket on that route closed untouched. Eleven of eleven.

We separated the producers, which was the right fix and had an awkward consequence: it removed
the very traffic that would have demonstrated the fix working. That awkwardness is what kept this
file open for the last week.

Then we found something worse underneath. The tickets we had separated out were being sent to an
agent that **provably cannot repair them** — its repair routine can only write two particular
style variables, and every single one of these tickets asks for five different ones. Measured, it
changed nothing on 61 out of 61 live pages. And it was closing them green anyway, reporting in
its own payload that it had changed nothing at all.

## What we've done

Three pieces, and the last one landed today.

**A daily detector** that looks for this whole class of problem across the estate — a check
graded by a predicate that was not written for it — rather than waiting for someone to trip over
the next instance. Live and running since the 11th.

**A gate that refuses a completion the handler never earned.** If the agent's own report says it
changed nothing, the ticket cannot be stamped done. It went live on the 14th and was proven on
real traffic the same day: three tickets blocked, four allowed through, and every one of those
four accounted for individually. Before that gate, all of them would have read "complete".

**And, today, the way back out.** The gate could only say no. It left those tickets sitting
marked "failed" with no honest route to closure, even if a human later fixed the page. Now, when
the audit stops reporting a fault, the ticket closes itself and records what was observed as its
reason.

The interesting part was deciding when silence is evidence, and every answer came from looking at
the data rather than from judgement. It has to be a statement about the whole **site**, because
the audit records the page as free text and writes "all" and "all pages" on the same day. It has
to be about the site rather than each individual fault, because the audit's own name was changed
last week and that name is baked into every ticket's identity — a per-fault design would have
read that one rename as fifteen faults being fixed at once. And it takes **three** silences,
because seven clean re-checks out of seven only proves the miss rate is under about 35%, which
means closing on one silence would close a live fault a third of the time. Three is simply the
first number that gets the risk under five percent.

We also took the opportunity a neighbouring team's reviewers had asked for: the same
"close it when it stops being reported" machinery now exists once and is used twice, rather than
being copied a third time. Their fourteen tests still pass without being rewritten, which is what
makes that safe to say.

## Where we are now

**The bug is closed, and so is a second one alongside it.** Bug 216 had been fixed, live and
proven for a week and was being held open only by a rule that had since been replaced.

Bug 213's own closing condition had become impossible to satisfy, so the closure says so plainly:
one branch of the original fix has never run in production, the query that would show it will
read zero for ever because the population is gone, and nobody should later read that zero as
evidence the fix never worked. That is written into the closed file rather than left for someone
to rediscover.

The review panel approved today's work first time, in eleven minutes. Three of its five comments
were checkable and we checked all three rather than filing them — including a genuinely good one
asking whether a safeguard we had added might accidentally seal shut the very four tickets this
work exists to free. It does not; they all match. Another found a real gap in our evidence, and
that gap is now closed with two more tests.

**Two things are honestly unfinished, and both are worth saying out loud.**

None of this can actually run yet. The sweep that drives these audits is switched off for cost
reasons, and it is the only thing that dispatches them. So the gate and the retraction are both
built, both proven as far as code can be proven, and both inert. That is not a hedge — it is the
difference between "this works" and "this will work when something turns it on", and we have been
careful all week not to blur it.

And we still do not know why some of these tickets come back carrying a result that belongs to a
completely different piece of work. The investigation into that died on the 14th when the account
hit a usage limit. We discovered today that the limit had been lifted about ninety minutes later
— so it had been re-runnable for a full day and nobody noticed, because we had written the
error's stated reinstatement date into our notes as though it were a fact. It has been restarted
and is running.

## Where we're going

Nothing in this workstream is now waiting on us.

The next real event is somebody re-enabling the improvement sweep. When that happens, both
mechanisms start working for the first time, and there is a short list of things to watch: the
silence counters should appear and climb, and the four failed tickets should **not** close,
because those faults are genuinely still there. A run that closed them would be a bug, not a
success. That expectation is written down in the runbook, because it is the kind of thing that
looks like failure and is not.

Beyond that, the honest remaining gap is the one this bug never claimed to fix: these tickets are
still routed to an agent that cannot repair them. We can now refuse the false green and we can
close the ticket honestly when the fault goes away — but nothing yet actually fixes a dark
section. That needs a handler that does not exist, and it is not currently anybody's task.

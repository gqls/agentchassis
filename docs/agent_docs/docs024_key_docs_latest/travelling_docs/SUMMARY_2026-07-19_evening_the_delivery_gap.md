# Self-verifying tools — read-out, evening of 2026-07-19

*Supersedes `SUMMARY_readout_2026-07-19.md` (written this morning), which said the
mechanism was done and only polish remained. That was wrong in one specific and
important way, found later the same day. Plain language, meant to be read aloud.*

---

## What we are trying to do

Make the site-building agents accountable for their own work.

The platform builds and maintains websites with autonomous agents, and the
problem we set out to solve is simple to state: **an agent could report success
without having done the job**, and nothing noticed. A ticket said `complete`, a
page looked plausible, and a broken tool sat live on a customer site until a
human happened to look.

So every tool the system creates now writes its own documentation at birth —
including its own definition of "working", as checks a machine can run. A
scheduled sweep later drives that tool in a real browser, on desktop and mobile,
and decides whether it still meets its own criteria. If not, the system works out
whose fault it is, files the repair, lets the fixing agent repair it, redeploys,
and checks again.

## Where we came from

A game on one of our sites had two plain bugs — a slider that did nothing useful
and a chart line stuck flat on the axis. It had passed every check we had. That
is the whole problem in one sentence: **"passed the checks" is not "works."**

## What we have built

All of it is live: documentation that travels with a tool, acceptance criteria
captured at birth, and a ladder of verification from cheap static checks up to
driving the real page in a real browser. Machine-authored to date, counted today:
**8 tool plans, 101 notes, 12 browser verdicts and 35 repair notes** — no human
wrote any of them.

Along the way we fixed what got in the way: a memory crash that turned out to be
an infinite loop rather than a leak; a repair agent that hit its output limit and
saved a truncated fragment over a working tool, which now cannot happen twice
(the limit was raised and a guard refuses to save a rewrite that has lost its
structure); and precise blame, so the fixer is told the element actually causing
an overflow rather than the innocent container that inherited it.

## Where we are now

**The honest headline: the loop is proven, but for tools it does not reach the
user.**

This morning's read-out said the one outstanding item was watching our benchmark
tool finally go green. Chasing that turned up the reason it never had:

> **No repair to a tool has ever reached the live page.** All three fixes the
> system made to that tool are sitting correctly in the durable template. The
> page a visitor sees has not changed since the tool was born.

Re-verification came back red every time because **the page under test was
identical every time**. Three separate faults sit in a row, each hiding the next:
a flag that marks the page as needing a rebuild and which nothing reads; a
re-render request that omits one field and therefore silently takes a
"reassemble the old HTML" path while reporting success; and a safety guard —
built to stop us blanking article bodies — that refuses to re-render anything
with no stored content, which is exactly what a self-contained tool has.

This matters beyond the one tool. It also rewrites an earlier conclusion: we had
diagnosed the repair loop as unable to converge, blaming the fixer's aim. That
drove a real improvement, but it was not the cause. Nothing was converging
because nothing was changing.

Two things are genuinely fixed today. The truncation protection is **proven** —
the repair agent re-ran and produced a complete, structurally sound tool where
its previous attempt had destroyed one. And an anti-churn rule that was counting
**successful** work as failed attempts — quietly killing repeat repair requests
across the fleet, 111 of them in the last month — is fixed, tested, and waiting
on a caller to opt in.

We also put the proposed fix through the reviewer council four times. It came
back better than it went in: a reviewer caught that my rule was broader than I'd
justified and pointed me at an explicit marker the system already had, cutting
the blast radius from nineteen components to twelve. The reviewers' own checks
caught a loose end I'd have shipped past. And once the council was **wrong** — it
blocked the plan at high severity by quoting our own internal knowledge base as
if it were the code. It isn't; that claim was never true. Another thread is now
correcting it.

## Where we are going

1. **Finish the delivery fix** — route the repair's re-render request through the
   proper path, scope its key so two tools repaired at once don't collide, and
   teach the safety guard the difference between "has no content" and "needs no
   content". Then rebuild, redeploy, and watch the benchmark finally go green.
2. **A convergence guard** — after N repair cycles that don't fix a thing, stop
   and ask a human rather than retrying weekly forever. Worth more now that we
   know a loop can spin without changing anything.
3. **Widen the write guard** to the rendered-page tables, which have the same
   unprotected shape as the one we already protect.

The mechanism is sound and the verification ladder genuinely works — it is, after
all, what caught this. What it has just told us is that the last mile, getting a
repair in front of a visitor, was never actually connected for tools.

**One caveat I want on the record:** nothing about this fix is deployed. No Go is
written for the three defects, no migration applied. The benchmark tool is still
red, and it should stay red until the loop turns it green by itself — fixing it
by hand would destroy the only evidence we have.

# SUMMARY — the next step turned out to need a decision, not just wiring, 2026-07-31 (evening)

## What we're trying to do

Give a *component* — a piece of a page shared across many sites, like a teaser panel or a
site header — the same machinery a *tool* already has: a written contract (a PLAN) in the
database, a history of verdicts against it (NOTES), and a real test that drives the component
in a browser and reports honestly when it is broken. The wider goal this sits inside is
`features_open/027`, a deliberately small three-gate version of what was originally an
eight-stage build ladder, cut down earlier this lane because most of the eight gates were
either wrong on contact or caught nothing.

## Where we've come from

The database side of "a component can have a contract" has been live since earlier today.
What remained unproven was whether the program side had actually shipped the matching change,
or merely looked like it had because the running binary was built after the right commit —
a distinction this fleet has been burned by more than once, on this exact rule, including once
where the mistake sat undetected in production for two days.

## What we've done

Proved the program side properly rather than by inference: dispatched a real test against the
live cluster that asked two questions in one go — the real one, and a deliberately impossible
one — so a clean answer could not be confused with a test that silently never ran. It came
back right on both counts, so that question is now closed. That milestone is written up on
its own, in the summary immediately before this one.

Then, before starting the next real piece of work, we did the thing this lane keeps having to
relearn: read the actual code the plan was pointing at, rather than trusting the plan's own
one-line description of it. The plan called the next step "wiring, not construction," on the
strength of an earlier pilot that supposedly proved the same mechanism worked already. It
hadn't, for this case. The earlier pilot was a *tool*, and the piece of code that sends a test
to a browser was built with a tool's assumption baked into it: that the thing under test lives
on exactly one page. That is true for a tool and false for a component by design — the
specific component picked for this next step sits on five different pages across two
different websites, because being reusable across sites is the entire point of a shared
component. Nothing in the existing code can currently answer "which of the five" for it.

## Where we are now

The next piece of real work has been rescoped rather than started, and is genuinely two steps,
not one:

- **First**, the chosen component still needs an actual written test of its own — a real one.
  The only test ever written for it so far was a throwaway placeholder created purely to prove
  the database plumbing worked, and that has already been deleted. Nothing exists to dispatch
  yet.
- **Second**, sending that test to the browser needs a small, deliberate decision about how to
  teach the existing mechanism a question it has never had to answer before — which of a
  component's several pages to test against. Two reasonable ways to do that have been written
  down side by side, with their tradeoffs, and deliberately left undecided rather than picked
  under time pressure.

Neither step is hard. Both are genuinely new work rather than the continuation of anything
already in flight, which is why nothing has been started on them yet.

## Where we're going

A fresh conversation picks up both steps in order, using a handoff document written
specifically for this handover: author and prove the component's own test first, then make
the small design call and wire the dispatch, using the same "test the real thing and a
deliberately broken thing side by side" habit that closed out today's proof. Behind that, in
order, sit three smaller items already queued and unaffected by any of this: a backlog of ten
tools with no written test at all (a cleanup, not a defect), a filed-but-unowned question about
documents left behind by renamed things, and a checklist item now unblocked because a
different team closed the bug it was waiting on.

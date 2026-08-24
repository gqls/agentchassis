# SUMMARY — 2026-08-24: both halves of the follow-on work are built, reviewed and approved

*Fourth in the series. Previous: `SUMMARY_2026-08-20_component_instance_scope.md`. Written
because the read-out genuinely differs: on 20 August the mechanical programme was serving and
the remaining work was a design question with an owner decision outstanding. Both halves are now
built, both council-approved, and what remains is one thing only — the owner's next fleet roll.*

---

## What we are trying to do

Make it safe for the same kind of section to appear more than once on a page.

When a page uses two text blocks, or three FAQs, each copy needs its own name in the HTML. If
they share a name, a script or a stylesheet aimed at one of them silently hits whichever copy the
browser found first — the second calculator answers from the first one's inputs, and nothing
anywhere reports an error.

## Where we have come from

The original bug (283) was that these names were **literal** — hard-coded into the component's
template, so every copy was identical by construction. That was fixed and went live on 22 August,
and it stays fixed.

Fixing it exposed a second, quieter problem. The names are now worked out by counting: first
copy, second copy, third copy. But two of the ways a page gets rendered can only see **one**
section at a time, so they cannot count — and both simply said "first copy" about every copy.
That is why pages we repaired kept breaking again: three of the twelve fixed on 23 August were
broken again within hours by an unrelated job touching the same pages. **Repairing worked. It did
not stick.**

The 283 lane closed on 24 August having handed that second problem to a successor thread, in two
halves: **Half B**, making the detectors able to see a related failure (an id that renders
completely empty), and **Half A**, the counting fix itself.

## What we have done

**Half B is built, committed and approved** — by a parallel session working the same lane. Empty
element ids became a detector class of their own, the conversion gate hard-errors on them, and
the render seam now reports an unbound identity rather than silently rendering nothing. A
question it raised — whether that seam should ever *refuse* a render — was withdrawn from the
change and sent to the owner as its own architecture note rather than settled quietly.

**Half A is built, committed and approved on the first round.** The fix turned out to be smaller
than the plan expected, because the information was already there. The plan proposed looking the
answer up in the database, which had a hole its own author had written down: on a brand-new page
there is nothing in the database yet, so a new page would still break on its first build. It also
needed a configuration change applied separately, meaning the fix would sit inert until someone
remembered.

It turns out the page-building loop **already tells each section which number it is**, and
already keeps the list of sections beside it — so a section can simply count the ones before it.
That is correct on a brand-new page, needs no configuration change at all, and starts working the
moment the next build of the software goes out. A second model was asked to check that reasoning
against the original plan rather than trusting our own enthusiasm; it agreed, and found three
details we had missed.

**The defect now has its own entry in the open-bugs folder** (`383`). It did not before — it lived
inside an architecture document, where nobody searching for "duplicate ids" would have found it.

**The three damaged pages have repair work queued**, and a piece of dead weight the lane had been
carrying for two days — an obsolete binding deferred twice — is finally deleted.

## Where we are now

Everything is in and reviewed. **The one thing standing between this and working is the owner's
next fleet build**, because the change is Go and Go does nothing until it ships.

Two honest limits, both written into the bug file rather than left to be discovered:

- **A repair is not the same as the fix.** Re-rendering a page repairs it even on today's code —
  that is how nine of twelve pages were fixed on 23 August. What the new code adds is that the
  repair **holds** against the next edit or rewrite. So the queued repairs will make the pages
  right now, and only the roll makes them stay right.
- **The fix covers today's callers, not tomorrow's.** If someone later builds a workflow that
  renders a section outside the page loop, it will quietly go back to saying "first copy" about
  everything. There is exactly one such workflow today. A warning is filed where the next author
  will meet it, and the review council said plainly that a written warning is not the same as a
  guard. They are right, and that judgement is worth a human's attention rather than ours.

## Where we are going

1. **The roll**, then check the two standing pages actually serve two different ids.
2. **Then the check that matters**, which is not the obvious one: after repairing a page, let one
   ordinary content rewrite run over it and count again. That is the step that undid the repairs
   on 23 August, and it is the only test that distinguishes "the fix works" from "the repair
   queue happened to reach it".
3. **One open question for the owner**, raised by the reviewers and not by us: whether "we fixed
   the two callers we know about" is an acceptable way to close this class of defect, or whether
   the underlying default should be made impossible to get wrong. We have made the case both
   ways in the bug file and have not assumed the answer.

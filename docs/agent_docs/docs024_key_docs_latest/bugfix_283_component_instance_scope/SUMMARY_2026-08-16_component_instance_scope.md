# SUMMARY — component instance scope, 2026-08-16

First summary in this lane. Written at a real milestone: the design question that blocked everything
downstream is settled, the change is council-approved and live, and the remaining work is a
different *kind* of work from what came before.

---

## What we're trying to do

Make it genuinely possible to put the same interactive component on a page more than once — and to
put two *different* interactive components on one page — without either of them quietly
misbehaving.

The obstacle is an old rule of the web: an element's name (its `id`) must be unique across the whole
page, because the lookup that finds it returns the first match and ignores the rest. Our components
hardcode their names. So a second copy renders perfectly, accepts typing, responds to its button —
and reads the first copy's inputs while writing into the first copy's results. It produces a
believable wrong number rather than an error, which on a consumer-credit site is the worst available
failure mode.

## Where we've come from

The bug was filed on 15 August out of the loan-and-mortgage-calculator lane, where a reuse
demonstration could not be done on a real page and had to be run on a throwaway one. The owner ruled
that reuse should be a genuine property of the platform rather than something we work around.

The first attempt built the mechanism — a per-placement name prefix, plus a checker for the three
distinct ways repetition breaks a page — and took it to the reviewer council, which returned
**revise**. Two objections were substantive. One: the new prefix was derived one way on the paths
that can see the whole page and a second, weaker way on the paths that render a single section, so
the same label carried two different guarantees — which is exactly the trap the whole change exists
to remove, recreated inside the fix. Two: the value was set at three known call sites while the
underlying mechanism stayed generic, so any other place that renders a component would silently
reproduce the bug.

There was also a defect the first session found in its own work after committing: the prefix was
built from a component's position on the page, and the same calculator sits at position 0 on seven
pages and position 1 on sixteen others.

## What we've done

Settled the naming rule, on the criterion that actually costs money if you get it wrong: not "which
candidate is most unique" but "what does a selector have to know". Position and the per-row database
identifier are both unique within a page and both differ *across* pages, which would force every
hand-written selector — including 170 automated checks — to carry per-page knowledge. The rule is
now the component's own function plus its occurrence on the page: the same name everywhere for a
component that appears once, a numbered suffix only when it genuinely repeats.

Deleted the second, weaker derivation. The paths that render one section at a time now feed the same
rule with a stated assumption rather than inventing their own answer — and where that assumption is
wrong the result is a *collision*, which the checker reports, rather than an empty name, which
nothing reports.

Replaced the hand-written list of call sites with a mechanical check. Three separate attempts to
enumerate these by hand produced three wrong lists — the council's, and mine, twice. The real figure
is fourteen calls across eight files; the two that were actually broken appeared on nobody's list.
There is now a lint that refuses a new render call site that forgets to set the value, and a report
at the shared render layer for a template that needs the value and was not given one.

Filed the architecture follow-up that had twice been claimed as filed and was not.

The council approved it, with six advisory notes; all are answered or recorded, including one we
have deliberately left open (see below). It went live on the fleet the same morning.

## Where we are now

The machinery is **live and inert**. None of the 243 active components reference the new value yet,
so nothing on any page has changed — deliberately, and it is what made revising the naming rule
free. **The original defect is still present**: the 22 calculator templates still ship fixed names,
and the bug stays open.

One thing is recorded as unresolved rather than fixed. The template engine silently blanks *any*
missing value in *any* template; we have added a report for one value's name, not a fix to that
behaviour. The council's bug historian was right to say so, and the honest position is that this
value could join the population of things the platform detects and never enforces. The enforcement
switch stays off because thirteen live pages already collide for an unrelated reason, and turning it
on before those are fixed would break their next re-render.

## Where we're going

Converting the 22 calculators is the work that actually fixes the original bug, and it is a
different kind of work: it writes to the live database, it changes the bytes of 22 live pages, and
the council's architecture reviewer was explicit that the conversion — not the machinery — is where
this becomes a real commitment across the component library and deserves its own review.

So the order is: architecture round first, then convert the templates, then move the 170 automated
checks to the new names in lockstep, then rebaseline the check that currently verifies those pages'
bytes do not change, and only then consider turning enforcement on. Each of those has a stated
reason for its position in the queue, and none of them should be started out of order.

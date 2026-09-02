# SUMMARY 2026-09-02 — the improvement loop

Written to be read aloud. Current state only; the chronology is in NOTES and
`README_where_we_are.md`.

---

## What we're trying to do

The improvement loop is the part of the platform that keeps sites good after they are
built. Every fifteen minutes it picks a site, checks it, writes down what is wrong, sends
the fixable things to the agents that fix them, and re-renders. It is the difference
between a site that is built once and a site that gets better.

It had no owner. That is what changed today.

## Where we've come from

The loop was switched off in May during a heavy development phase, and there is a standing
ruling from July that says so and says not to turn it back on. That ruling has been
overtaken — a migration re-enabled it in August — but several documents still repeat it,
so the written record and the running system disagreed about whether the thing existed.

Along the way the loop acquired a correct fix with an incomplete consequence. Some
findings are deliberately not jobs: nobody can automatically repaint a brand or repoint
somebody else's broken image, so the checker leaves the "who fixes this" box empty on
purpose. Until August those were being shoved into the dispatch queue anyway and coming
back stamped "could not be routed" — a correct observation filed as a breakdown. That was
fixed. What was not done was giving them anywhere to go.

## What we've done

Opened the lane, with its five standing documents, and established the loop's real state
by measuring rather than reading: it runs about fifty times a day, covers thirty-two
domains in a fair rotation, and its cost gate genuinely discriminates — it paid for a full
audit on a quarter of visits and correctly skipped the rest.

Corrected the design document that every new session reads. It still described a
three-audit rule that was replaced a month ago, and said nothing at all about the routing
guard that is now the most consequential thing the loop does.

Found and sized the real gap: **1,385 findings that nothing can act on and nobody can
see**, across 31 sites, roughly doubled in a fortnight, with the loop reporting "site is
clean" over the top of them every fifteen minutes.

Then looked inside the pile instead of reporting the number, which changed the answer.
Two thirds of it was a single missing accessibility feature — the "skip to content" link —
absent because 32 of our 33 site headers share one component that never had one. **That
fix is written, tested and committed**, with the link, its target and its styling all
emitted together so none of them can arrive without the others.

Answered the owner's operational question: of 34 live domains, two are not serving our
sites at all. Both are still delegated to the domain marketplaces that sold them, so the
fix is their nameservers and not their addresses.

## Where we are now

The skip-link change is committed and building; it is with the review council and takes
effect on the next fleet release, then reaches each page as that page re-renders. Nothing
in the backlog will visibly move until that wave runs, which is expected and is written
down so a flat count is not misread as a failed fix.

Two things caught late are worth stating plainly, because both were mine.

Writing tests for the skip link, I then deleted parts of the fix to check the tests would
notice. Two of three deletions failed the tests. **The third passed** — nothing asserted
that the styling reached the page, which is the one thing standing between this change and
a visible "Skip to content" on thirty-one live client sites. The gap was in exactly the
risk I had already written three paragraphs about. Identifying a hazard is not the same as
asserting on it.

And I told the owner that forty pages of finished work were sitting behind a bad
delegation. Twenty-one are. The other nineteen were never built — I read a column that
says a page is *wanted* as one that says a page *exists*. Caught before he acted on it,
which is the only reason it is a correction rather than an incident.

## Where we're going

The skip-link wave, once the release lands: re-render, then confirm at the served pages
rather than at the row count, and watch the 867 findings retract themselves.

Then the structural question, which the evidence has reframed. I had assumed nobody
noticed these findings were undrainable. In fact the platform's own code says so out loud
— one check quotes the bug that names the problem, explains exactly which door makes its
own finding undispatchable, and ships anyway. Eleven checks made that trade independently,
each with a good local reason and no shared place to raise it.

So the question is not "where do we display 1,385 findings". It is which of those eleven
are waiting on a handler that was deferred and never built, and which genuinely need a
person — and for the second group, giving them the brief that one check already writes and
the other ten do not. That is a change across eleven producers, so it goes to the council
on its own, once the skip-link wave has drained the pile enough to see what is actually
left underneath it.

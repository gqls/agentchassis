# SUMMARY — experience register, 2026-07-26

The first summary of this workstream. Written because a real milestone happened: the thing we
were waiting for went live, and the first experiences have been taken from it.

## What we're trying to do

Build a library of small, reusable user experiences. Not components — we already have those —
but the *behaviour*: what happens when you click the card, where the link should go, what the
page does while it waits, what it says when the thing behind it is broken. Today that gets
invented from scratch on every site, by whoever is building at the time. The result is that a
button's words and a button's destination are two unrelated facts, and nothing anywhere says
they should agree. That is the root of a whole family of our bugs.

The library holds each experience once, in a form a machine can select from, with its
acceptance tests attached. A site takes a copy and fills in the blanks — this list, that page,
this feed. Then a link can be checked against what it was *supposed* to do, instead of a
person guessing after the fact.

## Where we've come from

Two rounds of discussion at the end of last week settled the design, and four decisions were
yours: store it in the database rather than as documents; build it on the travelling-docs
idea, which already records a tool's provenance and direction; approve each experience
individually rather than letting use imply approval; and grow the vocabulary from real sites
rather than from a textbook. You also overruled my recommendation on the pilot — I wanted
vonc.com to ship a cut-down static version first, and you said wait for the real backend,
because the point is a working product.

Nothing was built. We wrote the design down, drafted the table, drafted the test format,
filed one bug we found on the way, and stopped, because the first entries were meant to come
from something that actually worked.

## What we've done

The vonc gauntlet went live end to end today — a real timed debate against a real AI opponent
on a real domain, verified in a browser on the deployed pages, 72 checks out of 73. So the
harvest started, and this session did it.

I checked the live site myself rather than relying on the report: both pages answer, the data
feed the site reads is byte-for-byte the one in the repository, and the code that drives the
archive really is inside the file the live site serves. Then I took four experiences out of it
— two about single components, two about journeys — and wrote each one down in the shape the
register will store, with the evidence for every clause pointing at live code or a live check.

Along the way the harvest corrected our own design in ten places, which is the entire reason
we insisted on harvesting from something real.

## Where we are now

Three things are worth saying plainly.

**The register's case is now evidence rather than a hunch.** All four of the components
involved have been used exactly once, on one site. The components don't repeat. But the
*rules* already do — one of the four experiences was taken from two different components on
the same site, and the "click a teaser, read the full piece" shape is re-invented by every
news, blog and directory site we run.

**Our design was wrong in ways only contact could reveal.** It could only describe things a
visitor can click, when the most valuable rule in the live code is the opposite: a row with
nothing behind it must not be clickable at all. It had no place to record what a page does
when its data or its engine is missing — which is exactly where a build is tempted to fake
success. And we had named the first pattern after a step that nobody has built.

**We have found a real limit in our own testing.** Four of the rules that make these
experiences worth having cannot be checked by the platform at all: it cannot ask whether a
link is dead, it cannot follow a link to see whether the promised page exists, it cannot check
that a region is empty, and it cannot wait longer than a third of a second — while the AI in
the gauntlet takes eight to twenty-three seconds to answer. That last one is not theoretical:
the approved plan for the gauntlet contains two tests that would fail a perfectly working
page. The fix belongs to the test harness and never to the page, because making the page paint
placeholder text would make those tests pass with the engine switched off.

## Where we're going

The register's own build is next, and it needs your go: two tables, a hook into the site
planner, validation when an experience is written, and the first thing that actually selects
from it — as one reviewed change. It is slightly smaller than it was, because a bug we filed
on Friday was fixed and proven live by another session over the weekend.

After that, more harvesting — the brochure component library has five components whose
behaviour is already documented and ready to take — and then the first site bound to a
register entry end to end.

The one thing I would flag for a decision beyond the build: the testing gaps above are a
separate piece of work, in code owned by other threads, and one half of it is already on the
vonc workstream's list for its own reasons. We can ship the register without it — the
experiences simply carry their unassertable clauses marked as such, rather than quietly
dropping them, so nobody mistakes an unchecked rule for a checked one.

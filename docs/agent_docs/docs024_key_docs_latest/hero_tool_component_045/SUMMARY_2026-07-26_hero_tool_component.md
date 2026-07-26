# SUMMARY — 2026-07-26 — the tool-hero library gap (bugs_closed/045)

_Current state only. The chronology lives in `NOTES_…` and `README_where_we_are.md`._
_First summary in this directory; written because the case closed._

## What we're trying to do

Make sure that when a page asks the platform for a generic building block — "give me
a hero banner for a tool page" — it gets one that is about *that* page. Nothing more
specific than that. It sounds like it should be automatic, and the whole value of
this case is that it wasn't.

## Where we've come from

The owner reported buttons on live sites that made no sense: an LLM cost calculator
and an ROI estimator both carrying "Start Ranking Free" and "Try the Bayesian Ranker".
The natural reading is that something upstream made a bad decision — that a planner
proposed a Bayesian ranking tool for a cost calculator.

It hadn't. The plan asked for a generic "tool hero", which was the right thing to ask
for. The problem was one level down and much duller: the component library contained
**exactly one** component capable of answering that request, and it was hard-wired to
a Bayesian ranking product — fourteen frozen labels that a page's own content could
not override. The selector picked the only candidate it had. Every part of the system
behaved correctly and the result was still wrong, because the library was missing a
part.

That distinction is the reason this case was split out of the broader CTA bug (`023`)
and given its own file. Read as a planner defect it sends the next person to debug the
wrong component entirely.

## What we've done

Built the missing part and retired the wrong one, in a single migration on 2026-07-21:
a generic `hero-tool` component whose visible text is all written per page, whose
buttons only render when there is a real destination to point at, and whose trust
statistics are optional so nothing is tempted to invent them. The Bayesian component
was **superseded, not deleted** — it is the only copy of itself and one live page
legitimately uses it.

Because this is configuration rather than code, it went live immediately with no
software release. But it could not be *proven* immediately: the defect only fires when
a page is rebuilt, and the cheap "re-render" path does not re-choose components at all.
So the case stayed open, deliberately, waiting for a real rebuild — and we declined to
force one, on the grounds that it would cost credits, rebuild a whole site, and collide
with two other people's live work.

## Where we are now

**Closed.** The proof arrived on its own on 2026-07-25, from a page we weren't watching:
a fresh tool page on fundamentallyai.com was built through the full path, chose the
generic component, and renders with none of the Bayesian vocabulary and a headline about
its own subject. The two design promises that only a real render can test both held — no
buttons at all where there was no destination, and no invented statistics. Across the
whole fleet the old vocabulary now survives in exactly one place, the page where it is
correct.

We are explicit in the case file about what that does *not* establish: the two pages the
bug originally named still have not rebuilt, so the closure rests on one genuine rebuild
plus a selector that can only return one answer, rather than on all three pages.

The uncomfortable part of this milestone is that our own verification instructions were
wrong in a way that would have signed the case off falsely. The live check pointed at a
URL that does not exist, and because the check counted the *absence* of bad words, the
error page passed it. It would have passed before the fix, too. That is now corrected in
place, marked rather than deleted, and written up fleet-wide as a general rule: a check
that something is absent also passes when the thing you are checking isn't there at all,
so it needs something positive alongside it or it is not a test.

## Where we're going

Nothing further is owed on this defect. Two threads were handed on rather than dropped:
the automatic warning that would have caught this class on day one — flag it when the
only candidate for a generic request is product-specific — has gone to the existing
component-adoption feature note, since the bug it was going to ride with has closed; and
two stale review-queue entries describing the now-extinct symptom have been pointed at
the queue-drain lane that owns them.

The one thing worth watching costs nothing: if either originally-named page rebuilds,
running the confirmation query against it strengthens this closure for free.

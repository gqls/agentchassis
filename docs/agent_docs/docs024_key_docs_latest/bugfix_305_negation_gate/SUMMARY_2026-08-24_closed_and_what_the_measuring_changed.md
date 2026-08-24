# SUMMARY 2026-08-24 — closed, and what the measuring changed about the story

*(Series: `SUMMARY_2026-08-20`, `SUMMARY_2026-08-21`, `SUMMARY_2026-08-22`, this one. Each is a new
file; none replaces another. Read in order, they show how the understanding moved — including where
it was wrong.)*

## What we were trying to do

The owner looked at three live directory pages and said the copy "looks like it didn't go through
the framework". The specific fault was a habit of **defining things by what they are not** — "shows
you what's possible, not what survives production" — repeated until it read as a tic rather than a
sentence. He asked for two things: stop the platform producing it, and fix the affected pages.

## Where we've come from

We started by proving it was not a one-page accident: the construction was in the site's own
writing brief, which hands it down to every page, and the brief-shaped version was in 24 of 25
briefs across the estate. So the fix could not be an edit to three pages.

What we built was a **gate in the writing pipeline**: it scans each section as it is written, finds
the construction, asks the model to rewrite just those sentences, checks the rewrite is genuinely
better and safe, and splices it back before the page is rendered. It took four rounds of review to
get accepted, including one outright veto that was right.

Then it went live, and running it in production found four things the tests could not: it was
detecting but not repairing (no model configured); the repairs were being made and thrown away (the
renderer read the wrong field); several rewrites to one paragraph were overwriting each other; and
the repair log was quietly losing records. Each was found by looking at the live result rather than
at the status.

## What we've done since the last summary

Two more defects, and they turned out to be the same defect.

**The first** was the repair log dropping records — when the model didn't mention a sentence we'd
asked about, nothing recorded that. Fixed, reviewed, and now live.

**The second** we found by accident, while checking the reviewers' objections on the first. The
repair had a **size limit far too small**. A page with one phrase to fix was fine; a page with nine
or ten needed the model to quote each original *and* its replacement, and it ran off the end. When
that happens the whole answer is discarded — not "some fixes land", none do. The evidence was
unusually clean: every page with five or fewer phrases repaired, both pages with nine or more
repaired nothing, no exceptions either side. That's a wall, not bad luck. And because the failure
was all-or-nothing, those two pages held **a quarter of everything the gate was meant to fix**.

**The part worth remembering** is what happened when we measured the two fixes separately. Before
either, 40% of repair runs didn't add up. After raising the size limit *alone* — accounting code
untouched — 15%. After both, zero. So most of what we had described as "the model ignored some
sentences" was really "the model ran out of room". We had diagnosed the symptom correctly and
attributed almost all of it to the wrong cause. Both fixes were needed; but had we found the size
limit first, the other bug would have looked far smaller than we said it was.

We also stopped ourselves overclaiming twice. A zero doesn't prove a fix if there was nothing for
the fix to catch — so we said so, and waited. Four hours later the traffic produced 43 records, and
the fix was proven properly. And a test we wrote had a closing assertion that would have passed
whatever the code did; we cut it and said why, because a reviewer had flagged exactly that shape at
us the day before.

## Where we are now

**Closed.** Everything the gate owns is fixed, live and demonstrated on real traffic, not inferred
from a deploy. On the owner's three pages: the one carrying **both sentences he quoted is completely
clean**; the second has nothing left to fix; the third still has two, and is waiting on a rebuild
that is already queued behind a separate issue on that site. The count went from six worth fixing to
two.

The one thing the gate deliberately leaves alone is the company tagline, because it comes from the
brief rather than from the writer. Changing that means changing the brief — the owner's call, and
correctly not the system's.

## Where we're going

Nothing is outstanding as engineering. Three things sit with people rather than code: whether to
edit that brief; whether **"rather than"** is a genuine tic or just ordinary English; and one page's
rebuild in another team's queue.

The middle one is the interesting one and we have deliberately **not** answered it. It was always
meant to be settled from the gate's own rejection log — and that log only became trustworthy today,
because both of this week's bugs were defects *in the log itself*. There are thirteen entries. It
deserves a week, and it matters more than it sounds: "rather than" is **71%** of everything the gate
rewrites, so getting it wrong would either leave the tic everywhere or start rewriting normal
English across the estate.

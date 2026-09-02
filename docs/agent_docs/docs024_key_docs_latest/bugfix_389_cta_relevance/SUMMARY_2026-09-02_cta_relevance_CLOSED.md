# SUMMARY — 2026-09-02 · CTA destination relevance (`bugs_closed/391`) · **LANE CLOSED**

## What we were trying to do

Buttons across three of our sites were sending visitors to the wrong place. The prominent one on
almost every page — "try this", "book a call" — opened a password-strength toy that had nothing to do
with any of those businesses. The job was to make the buttons point somewhere sensible and take the
toy off the three sites.

## Where we came from

The cause was duller than a bad guess. The framework ranked candidate destinations by a menu-order
number and took whichever sorted first. No topic, no tags, no judgement. The toy carried a `1` set
when its page was created in March, so it won everywhere.

Worse, the framework then **wrote each button's wording to name whatever it had picked** — so a wrong
choice produced the copy that locked it in, because the next pass matched the wording back to the
same page. Twenty of eighty buttons had reached that state, including all three the owner reported.

## What we did

**Fixed the fossil**, demoting the menu-order value on all three sites — to 900 rather than 200,
because at 200 it ties and the alphabetical tiebreak hands the toy the win again on two of them.

**Rewrote the twenty locked buttons** through the framework across twelve pages, each checked on the
live site rather than in the database.

**Routed every "get in touch" button to the contact page** — twenty-one of them across fifteen pages,
on the owner's decision. That repair is durable by design: the platform already treats a stored
contact link as authored and re-writes it on every recompute, which we proved by running a full
recompute over one and watching it leave the destination alone.

**Archived the three tool pages.** Archiving freezes a page while still serving it, so nothing broke
and every step remains reversible.

**And we broke two pages along the way and put them back.** The rewrite we commissioned to change
button labels also rewrote page bodies on two of twelve pages, replacing whole sections with copies
of their neighbours. Both were restored word-for-word from the record the damaging write left behind.
The check we had been using could not see it — it counted paragraphs, and the damage swapped three
paragraphs for three.

## Where we are now

**Closed.** No button anywhere on the three sites names the retired tool and links to it; no CTA field
points at it; the three the owner reported now go to the contact page or to a genuinely relevant tool.

Two things are deliberately **not** part of this closure, and both are recorded rather than carried:

- **Seven stale entries** remain in "related tools" card lists. A different mechanism — snapshots
  taken from a query that was never re-run — belonging to an existing open bug. Visitors do not see
  them, because the renderer drops a button whose destination is archived. They do block the final
  step of deleting the three tool files, which therefore waits on that bug.
- **The root cause is untouched.** The ranking still has no notion of what a page is about, so the
  next site we build inherits the same bug. That is a design change with its own review process, now
  filed as its own job rather than keeping a finished one open.

## Where we are going

Nothing on this lane. The successor covers the ranking; the listing residue belongs to the lane that
owns it. If the three tool pages should eventually disappear entirely rather than sit archived, that
follows the listing fix and is a five-minute operation once the way is clear.

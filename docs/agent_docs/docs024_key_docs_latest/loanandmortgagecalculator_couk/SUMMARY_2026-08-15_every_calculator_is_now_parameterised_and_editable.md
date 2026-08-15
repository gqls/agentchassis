# Every calculator is now parameterised and editable — Track B2 is complete

*2026-08-15, written at the close of the old-shape conversion. Previous summary:
2026-08-14 (decomposed, guarded, and one regression caught by the habit of rechecking).*

**What we're trying to do.** The owner's direction of 13 August: all text and widgets
on this site need to be editable, so calculators can be reused later with their own
slightly different copy or mechanism. Concretely: the working machinery of each
calculator lives in a template that content editing cannot touch, and every visible
piece of copy — headings, labels, button text — is a named, editable field.

**Where we've come from.** The site's 23 calculators started as hand-built pages. The
first decomposition (early August) froze each calculator into a locked block: safe,
but uneditable — at that time "editable" and "protected" were in direct conflict. The
B2 design resolved the conflict, and it went in three waves: 16 pages in batch one,
the 5 awkward mixed-card pages in batch two (deployed and verified this morning), and
finally the 2 oldest-shape pages, which needed their own conversion route — one had a
component with no fields, the other had no component at all, its calculator pasted
directly into the page row.

**What we've done.** Today the last two were converted in place. Every step was proven
before it touched anything: the parameterised template's output was shown byte-for-byte
identical to the served page using Go's own template engine — the same engine
production uses — before the database changed, and after deployment both pages serve
exactly the bytes they served before. The 5 August locks are gone; their safety role
is carried by the template boundary, the site's owned-rebuild policy, and the
acceptance fences. Along the way this session and a sibling session also repaired the
verification instruments themselves: the arithmetic oracle's three mutation controls
had blind spots (prose checks invisible to one control, integer expectations invisible
to another, and one boundary vector that coincided with the control's own sentinel
value), all found, fixed, and re-proven — the crosstool control is green today for the
first time in the lane's history.

**Where we are now.** 41 pages: 18 prose pages and 23 calculators, all 23 in the B2
shape, none locked, 154 editable copy fields across the fleet of calculators. The
independent arithmetic check passes 170 of 170 with its controls firing correctly.
The live pages are byte-identical to what they served before the whole conversion —
no reader has seen anything change, which was the point: the change is entirely in
what the site can now safely do.

**Where we're going.** The cheapest outstanding proof of the reuse goal: put one
existing calculator on a second page with different copy — nobody has demonstrated
that yet. Then the larger standing items, in the owner's order: seed the site spec and
let the planner converge on today's site (the site must not shrink on rebuild); the
og: tag half of bug 252; the complaint-deadline oracle for the loancash site; and
Track C, carrying this decomposition pattern to the next site.

# SUMMARY — bug 122, contrast and ink slots. 2026-08-10.

*The first summary this workstream has written, deliberately. The rule we set
ourselves was to write one when the first page measured clean. Today one did.*

## What we're trying to do

Make the text on our sites readable, and make it stay readable without anyone
watching.

The specific fault: our page renderer builds each site's colours from a small
palette — a brand colour, an accent, a page background, a card background. Some of
our page components used the brand colour as a **text** colour. On a site whose
brand colour is a dark navy and whose page background is nearly black, that text is
invisible. Not hard to read — invisible, at a contrast ratio of about 1.06 where the
accessibility standard asks for 4.5. Every individual input was correct. Only the
combination failed, which is why nothing caught it: we had about fifty automated
checks and not one of them rendered a page and looked at it.

## Where we've come from

A survey in late July measured fifteen live pages in a real browser and found **109
failures**. That number is the origin of this workstream. Early attempts to explain
it were wrong in an instructive way — one measurement compared a site's background
against itself and scored it a passing 1.11, which sent an earlier session hunting
the wrong defect entirely.

The fix we settled on, and put through the review council twice, was to make the
renderer answer a question it had never been asked. It already knew "what colour of
text goes *on top of* a brand-coloured button". It did not know "if I want to write
in the brand colour *on the page*, what should I actually use". Those are different
questions and the platform had a name for only one of them. So the renderer now
computes and publishes both, and — the part that matters — it checks the answer
against *every* background the text might sit on, not just one, because a component
can put the same label on the page and on a card, and only the worse of the two is
safe.

That code has been live in the running system since early August. But code alone
changed nothing a visitor could see: the components had to be told to ask for the
new colours, and every affected page had to be republished. That propagation has
been the whole job for the past four days.

## What we've done

**We finished the propagation and measured the result.** Eleven sites had their
stylesheets rebuilt; thirteen pages were republished. Every one was checked against
the live site rather than against a status field — a precaution that matters here,
because we found and filed a separate defect last week where work items get marked
"complete" without the work happening.

**We re-measured all fifteen pages and graded them one failure at a time.** All ten
of the closures we predicted landed: the orange-on-white links across the gas
wholesaler site, the invisible tool links and eyebrow on the robotics site, the two
washed-out buttons on the fine-tuning site, and the near-invisible label on the
darts site.

**We caught the fault coming back the same day, and closed it again.** The darts
site had deleted the component that carried its failure — and replaced it with a
different component carrying the identical fault six times over. We applied the same
repair to the new component, checked it against all fourteen sites that use it
(it changes nothing where the colours are already legible), republished the page,
and measured: **the darts homepage now returns zero contrast failures.** That is the
first page in this workstream to measure completely clean.

**We switched on the weekly check** — the second half of the approved plan, which had
been designed and never written. Every site is now measured in a real browser once a
week, automatically, and genuine failures are filed to the repair queue on their own.
It fired within seventy seconds of being turned on and immediately found **thirty-four
real failures on the robotics site's interior pages** — pages our fifteen-page survey
had never looked at.

## Where we are now

The engine is live. The components are fixed. The pages are republished and verified.
Ten predicted closures delivered, one same-day regression found and closed, and the
standing weekly check is running.

Two things are honestly still open, and neither is a surprise.

**Some failures are structurally out of reach of this fix, and we can say exactly
which.** The renderer knows about two backgrounds: the page and the card. When a
component paints its *own* background — a coloured band across the page — the
renderer has no correct answer to give it, because it is not measuring the surface
the text is actually standing on. Two "By the Numbers" labels are in this category,
and we proved it rather than assuming it: we shipped the fix, republished, measured
again, and got a ratio identical to the original failure, to the digit. That is
about two dozen failures altogether, and it is a design question — should the
renderer learn about component-painted backgrounds? — which we have deliberately put
to a human instead of answering ourselves.

**The far end of the new automatic loop is not yet trustworthy.** The thirty-four
findings it just filed flow to a repair agent that we have separately shown can mark
work complete without doing it. The detection half is now excellent; the repair half
has a known defect with a bug file open against it.

One measurement caution worth stating plainly, because it will mislead anyone who
glances at the totals: the fleet-wide failure count went **up**, from 109 to 112,
during a period when every failure we targeted was closed. Other teams ship to these
same sites continuously, and a page republished for any reason carries every change
since it last rendered. Grading by total is meaningless here. Grading was done one
selector at a time, against a banked before-state, which is the only method that
answers the question.

## Where we're going

The immediate list is short.

The weekly check will work through the fleet one site per hour, and each site's
first sweep will find interior-page failures the homepage survey never saw — the
robotics site's thirty-four are the first instalment, and we should expect that
scale elsewhere. Those findings need the repair-queue defect fixed before they can
be trusted to close themselves; that bug is filed and is now the highest-value thing
in the area.

The component-painted-background question needs an owner's decision. It is the
difference between a fix that covers most of the fleet and one that covers all of
it, and it is an architecture change rather than a bug fix.

And this bug can be closed to the extent the engine allows, with its remaining
scope handed cleanly to the two open items rather than left implied.

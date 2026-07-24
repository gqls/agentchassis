# SUMMARY — robot-hands.com — 2026-07-24 — the residuals are closed, and the fabrication engine got its muzzle

Written to be read aloud. Five parts: what we're trying to do · where we've come
from · what we've done · where we are now · where we're going.

## What we're trying to do

Make robot-hands.com a site where everything it says is true — and, because the
same machinery writes every site, stop the platform inventing numbers anywhere.

## Where we've come from

The July sweep fixed the visible breakage and the made-up statistics, and the
22nd closed with a choice made in the same spirit twice: when a claim and
reality disagree, extend reality rather than soften the claim. The catalogue
grew to ten real grippers across six actuation technologies, each with its
manufacturer's own figures. What remained: the MatchMatrix tool still only
tested the five jaw grippers while the site said it covered all six
technologies; the catalogue page didn't actually list the grippers; three other
sites still showed the same kind of invented statistics; and the platform bug
behind all of it had never been swept for or fixed at the source.

## What we've done

The tool now does what the site says. MatchMatrix tests all ten grippers, and
honestly — each technology judged on the numbers its own maker publishes: the
jaw grippers keep the force calculation, the magnet is checked against its
published holding force and only on steel, and the suction, gecko and soft
grippers are checked against their payload ratings with your acceleration and
safety factor applied. Thirty logic tests pass. Checking the claims against the
new tool exposed something worse — the methodology page described a scoring
system that has never existed anywhere — so those pages were rewritten to
describe the real tool, which is genuinely the stronger story: every result now
prints the formula that produced it. The catalogue page finally lists all ten
grippers, pulled live from the database on every rebuild.

The other sites got the same treatment. vonc's invented arena activity
("14,203 takes filed today", countdown clocks that count nothing, three cards
of archetypes the site doesn't have), gamesdesign's invented mathematics, and
the agency site's invented production metrics — which had been underselling the
real platform — all replaced with figures that trace to something countable.

And the source got fixed. A routine rebuild re-invented "2,400+" on
robot-hands' own homepage mid-session — proof the disease reinfects. The cause
was precise: the writing engine is handed a required stat field demanding a
number, its rulebook says never invent one, and the demand wins. The fix gives
it a legal out — no given figure, leave it empty — and each cleaned site now
carries a "verified facts" card listing the only numbers the writer may use.

## Where we are now

robot-hands is consistent end to end: the homepage, about page, gripper detail,
catalogue, tool and methodology all say the same true things, checked on the
live pages. The other three sites' stat blocks are honest. The prompt rule and
the verified-facts cards are live, so the next regeneration keeps the truth
instead of re-inventing the lie.

## Where we're going

Nothing on robot-hands is open. The platform bug (043) keeps two structural
fixes for a future session — binding stat fields to data sources in the
component schemas, and an automatic audit that catches an unsourced number
before it deploys. Two decisions are flagged for the owner: finetuning.uk still
claims "clients served" and "satisfaction" figures nobody measured, and vonc's
concept copy describes its product vision in the present tense — the first
needs the real story, the second belongs to the experience-loop workstream.

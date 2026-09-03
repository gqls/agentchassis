# SUMMARY 2026-09-03 — fallback-tier subjects (bug 443)

*First summary of this lane. Written at the Stage A milestone: the fix is live and proven
to carry topics end-to-end; one owner read stands between here and the visible result.*

**What we're trying to do.** Stop pages on our sites showing the same section heading two or
three times. The cause was never the writer having a bad day: on sites that were set up
before the newer planning tables existed, the machinery that gives each section its own
one-line topic simply had nowhere to put one, so when a page's layout used the same building
block twice, the writer was handed identical instructions twice and wrote the same thing
twice. The fix teaches that older path to carry per-section topics too, and adds a quiet
alarm that records every time a page is built with a repeated block and no topics.

**Where we've come from.** Found on the finetuning.uk booking page ("What you do in the
hour", three times). We confirmed eleven pages across three of the six unplanned sites
really serve duplicated headings. The bigger estate question — bring those six sites into
the planning tables properly, or keep adding fallback support piece by piece — was split off
as RFC_063 and the owner decided: converge the six (option B); our fix is correct either way.

**What we've done.** The fix went through the review council (approved first time), was
committed, and has been live in the fleet since the night of the 2nd — re-verified in the
running binary today after the third of three fleet rolls. The database columns, the build
handler's wiring, and the planner rule that makes NEW plans carry topics are all applied and
verified live. Stage A is proven on a real build: topics now travel all the way to the
writer's data. The alarm works too — its first seven catches were chased down today and all
turned out to be pages built from plans written before the new planner rule existed (the
closest missed it by 34 minutes), not a new defect. That chase found four more live pages
with the same disease, put bounds on how many are still out there (about a dozen pages in
old plans plus the fallback pages), and — via another lane's finding — established that the
fix quietly covers most article-type pages across the whole estate, roughly five hundred
pages rather than the eleven we started with.

**Where we are now.** Everything waits on exactly one thing: the owner's read of the
redrafted writer prompt (seed 641 — rewritten today to his chosen wording, approved by the
council the same day). Until it applies, the writer receives each section's topic but is not
yet told to use it, so served headings still repeat everywhere. That is expected, not
failure — it is the clean split Stage A existed to demonstrate.

**Where we're going.** When 641 applies: Stage B — rebuild one reserved page
(your-own-model), show its headings become distinct, and save the before/after pair (owed
to the copy-quality lane). Then the clean-up of the wider list, tier by tier, handing each
site's pages to its owning lane; re-check all the damaged pages on the live sites; and move
443 to closed.

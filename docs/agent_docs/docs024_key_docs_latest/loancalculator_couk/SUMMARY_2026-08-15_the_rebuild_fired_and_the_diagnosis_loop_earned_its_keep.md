# SUMMARY — 2026-08-15 — the rebuild fired, and the diagnosis loop earned its keep

## What we're trying to do

Bring loancalculator.co.uk — a hand-built site we adopted, decomposed, and
protected with permanent locks on its twelve calculators — fully under the
framework, so the platform can rebuild and evolve it like any other site, with
the calculators provably untouched throughout.

## Where we've come from

The planner work finished yesterday: the planner can now see a site's own
calculators in its menu (council-approved, canary-proven). The canary surfaced
two questions only the owner could answer — how the existing pages should be
regenerated, and whether the brief's "keep the pages" instruction could be
trusted — plus two invented pages awaiting a verdict.

## What we've done

The owner answered all three questions and the rebuild fired today. Along the
way it passed through two safety gates that each needed a deliberate manual
step (both now written up as standing rules): a live site's strategy refresh
never triggers a re-plan by itself, and a hand-filed work ticket must be filed
in the state the dispatcher actually reads. The results: no pages were invented
— the "keep the pages" instruction held; fifteen ordinary pages were rebuilt in
the new voice and are serving; the wanted about page was built from nothing and
is live; and the twelve calculators came through byte-identical to their
pre-rebuild backups.

Then the day's central finding. The redesign pass produced layouts for the
calculator pages that contained no calculators, and I wrote down the obvious
reading: the planner ignores the tools. **The diagnosis loop caught that this
was wrong, and the story is worth retelling.** Filed with the symptom, the loop
did not accept the attribution — it went and checked a fact I had not: whether
the tool names appeared in the planner's RAW output, one table upstream of the
plan I had measured. They did. Its verdict came back "unverifiable" rather than
confirmed, with the disjunction stated plainly — either something downstream
drops the tools, or they are deliberately out of the planner's scope — and,
crucially, with a short list of exactly what evidence would settle it. Walking
that list took minutes: the raw output shows the planner placing the repayment
calculator second on the homepage and the right tool on every calculator page;
the validation step's name-checker turned out to accept only ordinary section
names — nobody taught it about tool components when the menu was widened last
week — so it silently deletes every tool the planner places, believing it is
cleaning up an invalid name. The planner was right all along; the validator
eats its work. Filed as bug 282 with the fix spelled out, and my wrong reading
corrected visibly in every document that carried it. The lesson the loop
enforced: when a stored artefact lacks something, read the artefact one step
upstream before blaming the decision-maker.

## Where we are now

The site serves 28 of its 29 pages cleanly (the guides index is the one gap —
it waits on the same fix path as the homepage). Calculators: locked, verified
byte-identical, never at risk — the review gate and the locks did their job the
one time it mattered. The twelve redesign tickets are deliberately held until
bug 282's fix ships and a fresh plan proves the calculators land. One remaining
mechanical check — driving the calculators in a real browser against their
golden values — waits as the next session's first step.

## Where we're going

Fix bug 282 (small, well-scoped, through the council gate), re-run the
twelve-page redesign pass against the fixed validator, verify placement, then
work the redesign tickets under the acceptance checks. Separately the owner
holds two decisions: whether to un-park the design refresh (the site's look —
chrome, styles, favicon, imagery — currently blocked by items parked on the
12th), and how the calculator-page redesigns should be applied once placement
is proven.

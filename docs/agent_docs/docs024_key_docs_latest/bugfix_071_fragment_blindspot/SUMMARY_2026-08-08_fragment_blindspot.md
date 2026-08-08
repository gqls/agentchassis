# SUMMARY — 2026-08-08 — the fragment blind spot, and the double-check that now demonstrably refuses

## What we're trying to do

Make the platform notice a link that goes nowhere *within* a page. Not a broken
address — those we already catch — but the "jump to pricing" kind: the page loads,
and nothing moves, because the spot it points at doesn't exist. Nothing anywhere in
the estate looked at that. And, just as importantly, make sure that when the system
later claims it has fixed one, something independent checks before the item is
allowed to close.

## Where we've come from

The blind spot was structural rather than accidental. Two shared pieces of code —
the one that classifies a link, and the one that normalises a page address — both
threw away the part after the `#` before anyone compared anything. So the deploy
gate and the post-deploy audit each skipped these links by name, and agreed with
each other, and were both blind in the same way.

The fix shipped on 6 August: a new arm on an existing check (deliberately an arm,
not a new check — a new check needs switching on somewhere, which is exactly how a
previous fix reached production correct and never ran), a shared id-presence test
extracted from the one other place that asked the same question, and a registered
verifier that re-checks the defect at the moment something tries to close the item.
It passed council review first time and was proven live with controls.

One thing was left owed, and it was the important one: **the verifier had never
executed.** It was written, reviewed, shipped, and sitting in production unexercised
for two days. Two separate attempts to explain why it hadn't run — a filter on the
dispatcher, and a page-versus-chrome routing claim — both turned out on inspection
to be wrong. The real reason was mundane: no item of this type had ever existed.

## What we've done

Today we made one exist, on the throwaway pool site nobody serves, on the **chrome**
surface — the site-wide header/footer half, which had never been induced and which
is judged by a different rule, routed by a different branch and verified through
different SQL from the page half proven on the 6th. It filed exactly one finding, and
left its deliberately-resolving twin alone in the same run.

Then we drove it through completion three times:

1. **With the defect still live, the verifier refused.** It put the item back in the
   queue instead of closing it and recorded why, in its own words.
2. **The real dispatcher then ran the real repair agent, unprompted** — because a
   refusal marks an item retryable, which is what makes it dispatchable. That agent
   rebuilt the chrome, which removed the bad link, and the verifier agreed.
3. **The identical link, replanted next to a destination that now exists, produced
   the opposite verdict from (1)** — so it is genuinely resolving destinations, not
   matching text.

A final discovery pass over the repaired state found nothing, where the same check on
the same site had found one an hour earlier. The fixture was removed and the pool site
proven empty again.

We also wrote down three traps this cost us, one wrong call of our own, and corrected
a register entry that still said the verifier had never run.

## Where we are now

The lane's last owed item is discharged, and discharged in the direction that matters:
**a verifier that only ever agrees is the failure this whole mechanism exists to
prevent, and ours has now refused a completion in production.** All three of its
decision paths are demonstrated live rather than reasoned about.

Two things are worth saying plainly rather than burying.

**One mess is outstanding.** The repair agent, as its final act, always writes a small
file into our shared website repository — for every site, every run, even one with
nothing in it. So running it against a site that isn't real left a stray folder there.
The owner accepted that cost in advance; the deletion was blocked by tooling as a
destructive remote action and was not worked around. It needs one command.

**And we found a gap in how we can measure our own safety net.** When the system
refuses a completion and the item is later closed properly, the refusal is overwritten
and an obsolete error message is left behind on a healthy row. So the obvious way to
count "how often has this mechanism stopped something?" systematically misses the
refusals — which are the cases it exists to produce. That is not this lane's to fix,
but nobody should quote a count of verifications without knowing it.

## Where we're going

Three pieces remain deliberately deferred, all bigger than a bug patch: the deploy
gate still cannot judge these links, because it sees a page without its chrome;
nothing yet *repairs* a dead fragment, because unlinking one leaves the label
stranded as bare text; and no section component emits a stable destination, so such a
link can currently only be avoided, never made to work on purpose. That last one
changes every page's rendered HTML fleet-wide and belongs in the architecture track.

The near-term thing to watch for is the arm's first finding on a **real** site.
Everything proven so far was induced, which is deliberate and is not the same thing.
Expect silence — every fragment link on the estate resolves today — and read that
silence as corroboration only if the run actually covered the site.

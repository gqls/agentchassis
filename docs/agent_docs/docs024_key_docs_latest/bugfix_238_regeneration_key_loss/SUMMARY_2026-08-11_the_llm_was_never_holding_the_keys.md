# SUMMARY 2026-08-11 — the LLM was never holding the keys

## What we're trying to do

Stop the site-building pipeline quietly destroying parts of a page when it
rewrites that page. Specifically: a page's content record holds two kinds of
thing — words a language model wrote, and values the system looked up for itself
(an image from the site's asset library, a link from the site's settings). When a
rewrite runs, the second kind was being lost, and the page went out broken with
nothing reporting a problem.

## Where we've come from

On 9 August, finetuning.uk's homepage started showing five broken images. It was
filed the same day, carefully, by the lane whose own maintenance run had caused
it — with evidence, live queries and three ranked options for fixing it.

The explanation in that report was that the content generator had kept the parts
that look like writing and dropped the parts that look like plumbing. It is a
very plausible story about a language model, and it was wrong. **The model never
saw those fields and is explicitly forbidden from inventing them.** They belong
to the second category entirely.

The tell was in the report's own list of what went missing, and it took us a
while to see it: the loss was *too tidy*. Not "some images" — every field of one
particular kind and nothing else. A model dropping things it finds uninteresting
does not produce a clean split of a schema along a line nobody showed it.

## What we've done

**Established what actually happened**, from the system's own before-and-after
record of the page: the look-ups all failed, because the data they point at
doesn't exist for that site — and for two of them, has never existed for any site
we run. A failed look-up quietly omits the field. Then the save **replaced the
whole record** instead of updating part of it, so the previous values went too.

**Found why it had never bitten before**, which turned out to be the real story.
There are two ways a page gets rewritten: one merges, one replaces. Every rewrite
of that page since May had been the merging kind, so those values sat untouched
for three months and looked entirely safe. The first replacing rewrite deleted
all of them at once. **The gap between those two paths is the defect** — and
neither path is wrong by itself, which is why nobody had spotted it.

**Found that it was worse than reported.** The same failure also removed the five
"read more" links and the block's main button. Those didn't break visibly — they
simply vanished, because the template hides a link that has nowhere to point. So
the more defensive way of writing a template fails *more* quietly than the
careless way, which is worth knowing and is now written down where people will
hit it.

**Fixed both halves, as two separate reviewed changes.** One stops the loss: a
failed look-up now falls back to the value the live page is already using. The
other stops it shipping: the system turns out to have been *already working out*,
at the exact moment of the broken render, which fields would come out empty
inside an image or link — and then throwing that answer away. It doesn't any more.

**Repaired the site through the framework**, not by hand — restored data, then a
rewrite that involves no language model at all, verified on the actual served
page rather than in the database.

## Where we are now

The homepage is correct and live: five images, five links, the button back.

Both code changes are committed and **not yet running** — they take effect when
the fleet next rebuilds its images, which isn't something this thread does. The
first has been through the review council and was **approved**, with eight
advisory objections. One of them corrected a real mistake of ours: we had named a
particular page as the test of whether the fix works, and it can't be — that page
has already lost the values, so there is nothing for the fix to fall back to.
**The fix protects pages that still have their data; it repairs none of the pages
already damaged.** Left uncorrected, a page that quite properly showed no change
would have been read as proof the fix worked. The second change is still with the
council.

Four other pages across three other sites remain in the damaged state, on
purpose: one has no images left to restore, and three never had them, so
"fixing" them would mean inventing content.

## Where we're going

Three things, in order of who they belong to.

**Ours, once the fleet rolls:** confirm the new code is actually running, then
watch the one page that can genuinely exercise the fix. After that, decide
whether to turn on the second change's blocking behaviour — it's shipped switched
off, because switching it on means a page in the broken state can't be rebuilt
until its data is fixed, and that should be a deliberate choice.

**Yours, when you want it:** the architecture reviewer's view is that this
belongs in a design review rather than a bug fix, because it adds a compensating
layer over the merge-versus-replace split rather than closing it — so the split
stays reachable by whoever writes the next thing that saves a page. We've
recorded that rather than argued with it. It's a real question and it's above
this thread's pay grade.

**Still unavailable:** the third option in the original report — having something
re-examine sites automatically so this class gets caught within a cycle. That
remains the right idea and remains switched off, since those checks were paused
last week to save credits.

# SUMMARY — 2026-08-25: the lane closes, and the last thing it built now asks for itself

## What we were trying to do

This workstream started as a way to build website components in stages, with a gate at each
stage, so that a component could not reach a customer's site without something having checked
it. That job finished. What it turned into by the end was narrower and more interesting: making
sure a tool we build for a site actually gets *mentioned* on the pages where a reader would want
it.

A mention is one sentence, woven into an article that is already there. On dartsonline.com, in a
piece about dart barrel shapes, it reads: *"If you'd rather see the effect than read about it, the
tungsten percentage vs barrel diameter visualiser lets you compare percentages against weight and
watch the barrel narrow on screen."* That sentence is the whole product of this last stretch of
work. Without it a tool sits in a directory and waits to be found.

## Where we have come from

Three weeks ago the mentions were being written, but to the wrong pages. If the request that
ordered a tool did not say which articles it related to, the system quietly went and borrowed
another tool's list. That is how nine different tools on one site all ended up pointing at the
same two articles — correct for exactly one of them.

We stopped it guessing. That was right, and it exposed something worse underneath: the request
almost never said. Measured properly, the field was filled in by exactly one of the routes that
can order a tool, and by none of the others — eleven out of eleven from the automatic route,
nought out of sixty-six from every hand-written one. So having stopped the wrong mentions, we
were producing none at all: thirteen tools built over three days, thirteen with nothing.

The reason nobody had noticed is worth keeping. The system recorded the absence honestly, in a
line that said "this request named no related pages" — which is true, and reads like a decision
somebody made, rather than a question nobody was asked.

## What we have done

Two things, on the owner's instruction.

The quick one: the recipe people copy when they order a tool by hand now has the field in it, with
a note explaining what it costs to leave out. That went into the lane that files most of these
orders, who were mid-way through sixty-three rebuilds at the time.

The real one: the system now asks. When a request names no pages, both tool-building routes look
up the site's existing pages and pick one to three the tool genuinely helps with. It is allowed to
answer "none", and we made that explicit — a mention on an unrelated page reads as an
advertisement and is worse than no mention.

It went live this morning and worked on its first real case: it chose an accessibility article and
a colour-theory article for a contrast tool. Better than convenience would have chosen, which is
the thing this kind of rule fails at quietly.

The review panel then found the one thing we had got wrong, and it is the most useful sentence in
this document. If the asking step ever failed, it stepped aside so it could never break a tool
build — right — but the record it left then read "no related pages", which is *identical* to it
having run and honestly concluded that nothing fit. Two very different events, one indistinguishable
record. That is the third time this same confusion has bitten this exact feature: a system that
cannot tell "I tried and found nothing" from "I was never asked". It is fixed, it took two review
rounds, and the second round made it robust rather than merely correct.

## Where we are now

The mechanism is live and proven on real traffic. The last fix is written, reviewed and approved,
and is waiting for the next fleet build the way any code change here does.

One honest caveat, measured rather than assumed. On the site that builds most of these tools,
thirty-four of its thirty-seven candidate pages are marked as owned by the customer, which means
another safeguard correctly refuses to edit them. So the mentions are being *created* and then
*parked*. The picker works; on that particular site roughly nine times in ten the mention will not
reach the page. Whether an owned page should receive a mention at all is a genuine open question,
it is not ours to answer alone, and it now has numbers attached instead of hand-waving.

## Where we are going

This workstream closes here. Three things outlive it and each has been moved somewhere that gets
read rather than left in a handoff nobody reopens: the one remaining check on the older bug, which
needs a brand-new tool page rather than a rebuilt one; the owned-page question, contributed to the
bug that already owns that class; and the regeneration gap, which is filed and unowned.

The lesson I would carry out of it is not about tools at all. Twice in three weeks the same defect
appeared in a different disguise — a record that could not distinguish a failure from a legitimate
absence — and both times it was invisible precisely because the reassuring reading was the one
written down. The fix each time was cheap. Noticing was the expensive part, and what made noticing
possible was insisting on a number with a denominator that could have come out otherwise.

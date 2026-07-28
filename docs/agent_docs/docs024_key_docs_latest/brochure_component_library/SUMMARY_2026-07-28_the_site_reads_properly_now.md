# Summary, 28 July 2026 — the site reads properly now, and the machine that should have told us exists

*Written to be read aloud. Previous in the series:
`SUMMARY_2026-07-27_one_line_became_four.md`.*

---

## What we're trying to do

Build a consultancy brochure site, fundamentallyai.com, that sells this platform using
only things the platform has actually done — and build it *through* the platform, so
everything the site needs becomes a capability the whole fleet gains. The rule that makes
it hard is also the point: no figure appears on the site unless the system can re-derive
it from a real query.

## Where we've come from

Yesterday morning you looked at the site on your phone and said it was nothing like the
brief: grey text you couldn't read, no visible graph, one carousel with no images, not
enough imagery, not professional.

You were right about all of it, and the causes were worse than they looked. The
"unreadable grey" was near-white text on white cards at a contrast ratio of **1.21 to 1**
— effectively invisible — across five pages, for three days. The graph wasn't missing;
it was rendering with every label and number invisible for the same reason. And the
reason nothing had caught it was structural: the platform runs about fifty quality
checks, and **not one of them renders a page**. Every check reads an input — a template,
a colour row, a link — and asks whether that input looks right. Every input was
individually valid. The fault only existed in the combination.

## What we've done

**The site reads properly.** The palette was regenerated from the values your design spec
had pinned all along and was never applying. Card headings went from 1.21:1 to 13.19:1.
Measured in a real browser: **zero contrast failures and zero broken images across eight
of the nine live pages**, against roughly a hundred failures that morning.

**The metrics panel is finished and correct on both pages.** The home page shows the one
chart that belongs to it; the capabilities page now shows its two. That sounds trivial
and it took fixing a bug at four separate points — the page's own identity was being
thrown away four times over on its way to the template, and the bug report I wrote for it
said "one line" twice and was wrong twice.

**We built the missing check.** There is now a check that asks, before anything ships,
whether a site's colours can actually be read against each other. It is proven working —
and proving it needed a site that *fails*, since ours no longer does. Pointed at
dartsonline it produced exactly the failure predicted, to the decimal. Pointed at the
fleet it found **seven live sites with unreadable colour pairings, three of them the
identical invisible-text fault** that took three days and your eyes to find on the fourth.
So it was never rare. Nothing was looking.

**And we found that the platform had already told us.** Its own audit, on 24 July, filed
three findings about this site. Two of them were, almost word for word, two of your four
complaints — *"only 2 of 27 components contain images"* and *"these two pages share an
identical pattern"* — filed **three days before you made them**. They were never read,
because nothing consumes that kind of finding. Pulling that thread found 298 items across
the fleet that no handler can ever action, and 325 more sitting in the queue a human is
supposed to read, the oldest since March. Some of those are the platform asking you
questions — seventeen sections on another site have been waiting since 15 March for
pricing only you can supply.

## Where we are now

Good, with two honest blemishes.

The chart works, the colours work, the check that would have caught this is live and
proven, and the whole account is written down including the parts where I was wrong —
which was several times, and twice I caught myself only because I re-measured rather than
trusting a result that looked clean.

**The two blemishes are both mine, and both the same shape.** Rebuilding the capabilities
page to add the chart also made the writer invent nine internal links to pages that don't
exist, and swap four working carousel images for four that don't exist either. The build
passed every gate. I caught the links only because I took a "before" list first,
precisely because this had happened twice already. I did not hand-repair either, because
the previous two hand-repairs did not survive the next rebuild — that is the bug, and a
fourth patch would just be a fourth data point for a conclusion we already have.

Your rule from that — **replace before deleting** — is written down, and it turns out to
describe the same fault in three separate parts of the platform. It is willing to write a
reference to something that does not exist.

## Where we're going

1. **Stop the writer inventing destinations.** The links and the images are one authoring
   defect, not two, and there is already a bug filed for the likely cause — the writer
   never receives the constraints that would tell it what exists.
2. **The two tool pages still need their images**, now the spend cap is lifted.
3. **Your carousel direction** — most panels as carousels, a very short first sentence
   and a deliberately unfinished second one that resolves on click-through. Recorded, not
   started. Three carousel components already exist and should be read before anything
   new is proposed.
4. **The design-taste layer** — the screenshot critic — which is the honest answer to
   "not exciting or professional". Worth doing now that the mechanics underneath it are
   sound, and not before.
5. **Your admin session** on the review queue, handed to its own thread with the access
   route checked and working.

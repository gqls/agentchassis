# Summary, 28 July 2026 (second) — what we fixed, what we broke, and the vocabulary we already had

*Written to be read aloud. This one exists because the day produced a design direction and
a structural finding, not just repairs. Previous: `SUMMARY_2026-07-28_the_site_reads_properly_now.md`.*

---

## What we're trying to do

Build fundamentallyai.com as a brochure site that sells this platform using only things
the platform has actually done — and build it *through* the platform, so every capability
the site needs is one the whole fleet gains. No figure appears unless the system can
re-derive it from a real query.

## Where we've come from

Yesterday you looked at the site on your phone and said it was nothing like the brief.
Every complaint was right and the causes were worse than they looked: text at 1.21:1
contrast, which is invisible rather than faint; a chart that was rendering perfectly with
every label the same colour as its background; and about fifty quality checks, not one of
which renders a page, so all of it passed.

## What we've done

**The colours are fixed and verified in a real browser** — zero contrast failures across
eight of nine live pages, from roughly a hundred. **The metrics panel is finished on both
pages**, showing each page only the charts that belong to it. **The check that would have
caught all this now exists and is proven** — pointed at the fleet it found seven sites
with unreadable pairings, three of them the identical invisible-text fault. It was never
rare; nothing was looking.

**And then I broke two things on your site, in the same way.** Rebuilding the capabilities
page to add the chart made the writer invent nine links to pages that don't exist and swap
four working carousel images for four that don't exist. Both passed every gate. I caught
the links only because I took a "before" list first, having watched this happen twice
already.

I have not hand-repaired either, and that is deliberate: the two previous hand-repairs did
not survive the next rebuild. **That is the bug.** A fourth patch would be a fourth data
point for a conclusion we already have.

## Where we are now

Your rule from that — **replace before deleting** — turned out to describe the same fault
in three separate parts of the platform: imagery, links, and archiving a page while its
frozen copy keeps serving. **The platform is willing to write a reference to something
that does not exist.** That is one defect wearing three costumes, and it is worth fixing
as one.

Your instinct that **the writer shouldn't know how its content is displayed** is right,
and for a sharper reason than tidiness: the rebuild invented links and image paths
*because it was asked to produce a panel*, and panels contain those things. A small model
that only reshapes text it was handed cannot invent a destination, because it is never
asked for one. **Narrowing what a model is asked to do narrows what it can invent.**

Then your "set of known display shapes" question sent me looking — and **we already have
that table.** `experience_patterns` holds nine patterns, seven of them display contracts,
with a column whose entire job is naming the components each applies to. One of the rows
is literally your carousel idea, already named `teaser-detail-deeplink`: a teaser that
deep-links to its detail.

**Two things are wrong with it, and both are the same shape as everything else this
week.** Four of its nine component references **name components that don't exist** —
near-misses like `hero-carousel` for `hero-card-carousel` — and they are exactly four of
the five components this workstream built. Nothing has ever checked that a pattern points
at a real component. And no site is bound to any pattern at all: that table has nine rows
and zero users.

So the answer to "is that an idea?" is: **yes, and it is half-built and quietly broken.**
That is a much better place to start than a blank page.

The one change I'd make to how you put it: not *"new component, new shape"* but **"reuse a
shape, or justify a new one"**. Shape-per-component gives us a hundred shapes and no
vocabulary — which is precisely what we have today, just unnamed. The value is entirely in
the count staying small.

## Where we're going

1. **Stop the writer inventing destinations.** Links and images are one authoring defect,
   not two, and there is already a bug filed for the likely cause.
2. **Make the pattern register honest** — check that every pattern names a component that
   exists, fix the four that don't, add the two components that have no pattern. Small,
   and it makes everything after it possible.
3. **Then the splitter**, with the three hazards handled explicitly: never split a
   sentence containing a figure away from the words that verify it; mark a deliberate
   cliffhanger in the data rather than the punctuation, so it cannot be mistaken for
   truncation and "repaired"; and split once at build time, never on the re-render path —
   that path is LLM-free today and it is the only safe repair route we have.
4. **The design-taste layer**, which is the honest answer to "not exciting or
   professional" — worth doing now the mechanics beneath it are sound.
5. **Your admin session** on the review queue, already handed to its own thread.

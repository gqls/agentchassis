# The site was unreadable, not undesigned — 2026-07-27 (second summary of the day)

*A separate summary from this morning's deliberately: that one was about a bug in
how a component learns which page it is on. This one is about the owner looking at
the site on his phone, and what measuring it properly turned up.*

---

## What we're trying to do

Stand up **fundamentallyai.com** as a consultancy brochure site in the register of a
Bain or McKinsey — and build the interactive components it needs *through the
platform*, so that every one of them is reusable on the next site rather than
hand-made for this one. The site markets the platform's real, verified capabilities.
Nothing on it may be invented.

## Where we've come from

The site has been live since 22 July. Since then: five interactive components built
and placed, all five surviving a full pipeline rebuild; a chart component that draws
its numbers from an evidence register rather than from a model; every internal link
verified sound against the served pages; and a string of platform bugs filed out of
the friction — the stale-work-item reaper, the link gate that detects and discards,
the contact block that fails silently on a new site.

The pattern in all of that work was: the pipeline mostly does what it says, and the
gaps are where one stage's output is nobody's input.

## What we've done

The owner looked at the site on his phone and said it was nothing like the brief —
no graph, one carousel that did not load its images, grey text he could not read on
white, not enough imagery, not exciting or professional.

We measured instead of debating. Rendering each served page in a real browser and
computing, for every piece of text, the contrast between it and the colour actually
behind it, found **101 places below the readable threshold across five pages** —
several of them at 1.1:1, where 4.5 is the floor. Every card heading on the site.
Every eyebrow label. The whole chart section, which is why there appeared to be no
graph: the bars drew correctly and their labels and values were white-on-white.

Three distinct mechanisms, none visible from any single input:

- **The palette defines eight colours; the page template expects seventeen** and
  fills the rest with its own light-scheme defaults. On a near-black site that paints
  white cards and leaves the pale text on them. 12 of our 31 palettes carry it.
- **One colour token doing two jobs.** `--color-primary` is used as *text* in
  fifty-odd places and as a *fill* in twenty-six. Ours was a navy indistinguishable
  from the background, so every use as text vanished while every button looked fine.
- **The code that picks a default text colour preferred the palette's dimmest grey**
  to its actual text colour, for body copy and headings alike, on every dark site we
  run.

Separately, the imagery. Twenty-one generated line illustrations were already live on
the server and the pages referenced three of them. The carousel pointed at a filename
that exists on no site we own. The reason the good images never arrived is that the
job which wires them in — *"re-render this page now the picture has landed"* — sat
unclaimed in a queue; **fourteen of the twenty-eight ever filed are parked the same
way**, fleet-wide.

Fixed and live now: all the imagery, six page heroes, thirty real icons in place of
emoji, every page re-rendered, **zero broken images**. Fixed and awaiting one
publishing command: the colours, verified to take 101 failures to 1. Fixed in code
and awaiting the next image roll: the renderer, so no future dark site is generated
with this defect.

## Where we are now

**The most useful thing we learned is not about this site.** We run about fifty
automated checks over a site and not one of them renders a page. Every check reads an
ingredient — a template, a colour list, a link, an image record — and all three
mechanisms above are invisible from there, because each ingredient is individually
correct and only the combination fails. That is a missing vantage point, not a
missing check, and it explains why a site can pass everything and be unreadable.

The uncomfortable part: **the platform had already found some of this and told us.**
On 24 July our own brief-fidelity audit filed three findings against this site, one of
which reads *"Only 2 of 27 components contain images — raising serious doubt that the
illustration system is meaningfully present"*. That is the owner's complaint, in our
own words, three days early. It is still marked "detected". They are the only three
findings of that type in the entire database and nothing reads them.

So the honest state is: we detect more than we consume, and we check ingredients
rather than results.

## Where we're going

1. **Publish the stylesheet.** One command; the site's readability changes the moment
   it runs. It could not be done from this session because the only route the
   platform has to write that file also re-runs the design pass that re-rolls the
   colours, and the direct route needs a permission this session did not have.
2. **Drain, or stop filing, the parked re-render jobs.** That single queue is what is
   starving every site of imagery it has already paid to generate.
3. **Wire the page-measuring check into the build.** The tool exists and works; the
   design for the pipeline half is written up, including the small refactor it needs
   so the contrast maths is not implemented twice.
4. Only then, the work the owner actually asked for — more varied sections, richer
   use of the illustrations, something that reads as exciting. There is no point
   styling a page whose text cannot be read.

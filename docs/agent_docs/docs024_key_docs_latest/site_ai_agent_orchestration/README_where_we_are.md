# Where we are — ai-agent-orchestration.com

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-08-17 — what I found when I looked at the three things you asked for

You asked me to pick up the August 5th handoff and improve the site: get the images working,
turn the components into carousels, and fix the text contrast. You thought the first and third
were probably just a case of running the improvement loop over this site, since we'd fixed
those things elsewhere, and that only the carousels would need new thinking.

I measured all three before touching anything. **The carousel half of that is right. The other
half isn't, and I'd rather say so now than spend a run finding out.**

### The contrast problem is real and worse than the tickets say

There are 17 contrast tickets sitting against this site. I didn't trust them — they were raised
on the 11th and a lot has shipped since — so I loaded four pages in a real browser and measured
what a visitor actually sees. There are **44 genuine contrast failures on those four pages
alone**. The worst of them aren't "a bit hard to read": the text is painted in *exactly* the
same colour as the thing behind it. On the pricing page, seven headings — "Agile Orchestration
Architecture", "Enterprise Security and Reliability by Default", and so on — are completely
invisible. They are in the HTML, and a visitor sees blank space.

Here's the part that matters for your plan. We *did* build a fix for this, on another job, and
**that fix is already switched on and working on this site.** I checked in the browser: the
repair has computed a perfectly good legible colour and it's sitting there ready to use. The
site is still broken anyway, for two reasons the repair was never designed to catch:

1. **The site's own colour scheme has a fault in it.** Its "primary" colour is set to the exact
   same near-black as the panels it gets painted on. Of the 23 sites we have colour schemes for,
   only two are like this — this one and oufe.com. Everywhere else, primary is a colour you can
   actually see against the background. So anything that uses the primary colour for text here
   disappears.
2. **Some components ignore the scheme entirely and paint themselves white.** Seven of them do
   this, on the home and about pages. On a dark site, a white card keeps the site's pale text,
   and you get near-white writing on white.

The reason I'm confident the improvement loop won't clear this on its own: the **home page was
rebuilt this morning and still has 17 failures**, while the services page — rebuilt on the 15th —
has none. So it isn't simply that the pages are stale. And the pricing page is a special case:
it hasn't been rebuilt since **April**, and it can't be, because the stored content it would need
to rebuild from was wiped out by a bug we've since closed. That page needs building again from
scratch, not re-rendering.

### The images: it's one component, and the obvious fix would delete them

Every single image on this site belongs to one component, the case studies grid, and there are
only ten of them. Five (on the home page) have a completely empty image address, so nothing
loads. The other five (on the enterprise reference page) point at files that were never created
— I checked over HTTP, they're a genuine 404.

The underlying reason is the same in both cases: the component asks for a picture, and nothing
in the pipeline ever generates one. The written content is all there and it's good — five case
studies with proper titles, summaries, and even well-written descriptions of what each diagram
*should* show. Only the pictures are missing.

There is a live "image URL 404 handler" and a live "image source unsatisfiable handler", and
this site's tickets aren't assigned to either, which looks exactly like the missing piece. **I
checked what those handlers actually do before assigning anything, and I'm glad I did.** They
don't create images — they file a note and ask for review. The only site they've ever run
against is mortgagecalculator.co.uk, on the 14th, and that site now has **no images at all**.
So pointing this site's tickets at them would most likely strip the five case studies rather
than illustrate them. Making real pictures is a different piece of machinery.

One thing worth knowing separately: the nine images we *do* hold for this site are stored as
temporary web addresses that expire after seven days, and they were made on the 11th — so they
lapse tomorrow. Nothing on the site currently points at them, so I don't think anything breaks,
but it's the sort of thing that bites later.

### Carousels: nothing exists yet, and you were right about how to do it

We have no carousel component anywhere in the platform. There's a line of guidance buried in one
of the older page-building prompts saying to build them with CSS animation rather than
JavaScript, but it's in a path this site doesn't use. So yes — this is a hint to the planner, as
you guessed.

Worth flagging before we design one: we've had carousels here before and they went wrong in a
specific way. There's a note in the code about "four dead carousel destinations" found by hand
back in July — carousels whose buttons pointed at pages that didn't exist. Whatever we add
should make that impossible rather than just unlikely.

### What I'd like you to decide

I've written up the full evidence and I'm ready to go, but the contrast fix has two routes and
they differ a lot in how far they reach, so it's your call rather than mine. I'll put the
options to you separately.

# Summary, 30 July 2026 — the carousel panel actually works now, and it took clicking it to know that

*Written to be read aloud. This one exists because a whole feature turned out to have never
run, the fix for that is a general lesson not just a patch, and two new pieces of work opened
in the same conversation. Previous: `SUMMARY_2026-07-28b_what_we_fixed_and_what_it_taught_us.md`.*

---

## What we're trying to do

Build fundamentallyai.com as a brochure site that sells this platform using only things the
platform has actually done — and build it *through* the platform, so every capability the
site needs is one the whole fleet gains. No figure appears unless the system can re-derive it
from a real query. No control on the page invites a click it can't honestly answer.

## Where we've come from

Last time, we'd just found that the phantom-link bug had a design cost as well as a code cost,
and that your carousel idea already had a name in our own system — `teaser-detail-deeplink`,
sitting half-built and unused in a registry of nine display shapes. The plan from there was:
stop the writer inventing destinations, make that shape registry honest, then build the
carousel properly, with three named hazards to avoid.

## What we've done

**Someone else closed the biggest item on the list.** The phantom-link bug — links the writer
invented and nothing blocked — was fixed by another thread while we were focused here, proven
on a real page rebuild. Nothing was owed on it from this side.

**The carousel got built, on the shape that already existed.** Rather than invent a new one,
we built to `teaser-detail-deeplink` exactly: a teaser that opens its full text in place, no
page load, at an address you can share. It opens and closes with plain browser behaviour — no
JavaScript required for the core of it — because text that only appears after a script runs is
invisible to our own fact-checker and to search engines, the same blind spot we'd already found
with text inside diagrams.

**Then you looked at the site and the fixes kept finding real bugs, not just rough edges.**
You said the home page and the capabilities page felt very similar. They were — and once we
measured, eighteen different sections across five pages turned out to be independently saying
the same nine facts about the company, sometimes in wording only eighteen percent alike while
meaning the exact same thing. Fixed on the site by keeping one good version per page; filed
properly as a platform issue, because the actual cause is that nothing currently stops a page's
sections being written in isolation from each other.

**Then you asked for the carousel treatment almost everywhere, with pictures.** We had a stock
of about twenty-five real images already generated and mostly unused — closely enough matched
to what capabilities needed that they looked like they'd been made for exactly this and never
connected. Rolled the panel out to four pages. On the way, found the fine-tuning page had two
separate sections quietly repeating four of their six points each — the same repetition bug,
just missed by the company-wide count because these were page-specific claims.

**Then the polish requests, and the last one wasn't polish.** More breathing room, a visual
"…" that never actually gets typed into the real text (typing it for real would trip our own
system's damaged-output detector), the cut-off sentence reading as one clean paragraph instead
of two — all straightforward. Then you said the arrows didn't scroll and closing one card
didn't close another. **That one mattered more than it sounded, and finding it changed how we
test this from now on.** Every check we had run on this panel — including ones we'd called
"verified" — had only ever read the page's raw code or forced a card open directly in memory.
None of them had ever actually clicked anything, so none of them could have caught that
clicking didn't work. The moment we simulated a real click, the cause was obvious: the bit of
code that runs the panel loads before the panel itself exists on the page, so it has looked for
something that wasn't there yet and quietly given up — every time, since the very first version.
Fixed with the same defensive pattern five of our six other interactive components already use.
We checked whether that was a wider fault across the whole platform before assuming it was; it
wasn't — one other small script shares the gap, everything else already guards against it. The
mistake was ours specifically, not a hole in the platform.

## Where we are now

The panel is live on four pages, and for the first time we've actually proven it by clicking
it in a real browser rather than reading about it — the arrow moves the row, opening a card
closes the one before it, and the full text reads as one continuous passage with no visible
seam where it used to cut off. Two new, smaller pieces of work opened in the same conversation
and neither is built yet: putting the logos of relojistas, idea.uk and leopardess into the
portfolio cards — you've confirmed we already have the logo assets, so that's just wiring, not
sourcing — and a genuinely interactive tools page, with sliders or inputs, built on real
numbers the platform can already answer (how many pages we host, how council decisions split
by outcome or by reviewer), to make the same point the rest of the site makes: not a demo,
something you can actually operate.

## Where we're going

1. **Wire the three logos into the portfolio cards** — a small, contained job now that the
   assets exist.
2. **Design and build the interactive tools page** — inputs or sliders over real platform
   counts, in the same style as the two tool pages already live on the site.
3. **The blog's own carousel still has no pictures** — a different, older mechanic (swipe
   cards, no reveal) that would need its own small change, not a big one.
4. **The actual platform fix for the repeated-content bug** — letting the one system that sees
   a whole site at once hand each section its own slice of the approved facts — is written up
   and filed, not built. That's a real piece of platform work whenever it's picked up, not a
   same-day job.
5. **One more small thing worth a look, not a fire**: a second little script on the site has
   the same "runs before its target exists" gap ours did. Nobody's checked whether it actually
   matters yet.

# Imagery — where the card-imagery subproject stands (2026-07-19)

*Written to be read aloud. Current state only — the story in order is in
`README_where_we_are.md`, the technical log is in
`RUNNING_NOTES_imagery_best_in_class.md`.*

## What we're trying to do

Make the imagery on our sites good enough that a stranger would assume a designer
did it. Not decoration for its own sake: every article and every tool we publish
should carry a picture that belongs to it, in a style that is recognisably the
same site from one page to the next, at a file size that doesn't slow the page
down. The piece described here is the listing imagery — the small pictures on the
cards you see before you click into something.

## Where we've come from

We built the machinery for this over the past fortnight and it worked, mechanically.
The system could generate a picture for an article from the article's own words,
cut a card-sized crop from it, attach it to the right page, and put it on the site
without anyone touching it. That much was proven.

Then you looked at the result and failed it, correctly. The nine pictures were
generated in nine different styles — one photographic, one that came out as comic
line art, one in full colour on a site that is otherwise black and blue, and one
with a garbled fake logo stamped on it. Six of the nine articles being listed
didn't even exist as pages; the links 404'd. And only one page on the whole site
had any card imagery at all.

## What we've done

**Fixed the style at its root, not with better wording.** The reason the pictures
were inconsistent was structural: they were being sent to an image model that
ignores the "make it look like this" reference material we give it. Only the other
model we use honours that. Article pictures are now their own category, routed to
the model that can actually be steered, and each site can now specify a different
visual language for that category than for the rest of its imagery. You chose flat
two-colour illustration — charcoal background, electric blue shapes — because it
reads clearly at the small size these are actually displayed at, and compresses
well.

**Stopped the site listing pages that don't exist.** One shared definition of "this
page really shipped" now governs both the listing and the picture-generation sweep,
so the two can't drift apart and we don't spend money generating pictures for pages
nobody can reach.

**Put the change through the review council** and acted on what it found: a
database call that would have crashed an entire site's sweep if one row were
malformed, a duplicated rule that was already drifting, and a missing safeguard
around locked-down logos.

**Extended it to tool pages**, which you funded, and the tool directory on
robot-hands is live with matching imagery.

**Retracted a false alarm of my own.** I reported that a deployed component was
badly out of date. It wasn't; my method of checking was flawed and I had run the
same flawed check three times and called it three proofs. That is corrected
everywhere it was recorded, and the real lesson — the check can prove presence,
never absence — is now in the debugging guide.

## Where we are now

On robot-hands, the article listing shows three articles, three distinct pictures
in the agreed style, and every link goes to a real page showing its own picture.
The tool directory shows three tool cards with pictures. File sizes run 22–36KB
against a 60KB ceiling; the previous batch had run as high as 73KB.

Of the roughly 33 pictures you funded for tool pages, five are spent. Two other
sites use the same tool listing and will pick theirs up automatically on their next
sweep. Seven further sites have tool pages but no tool listing, and the rule
correctly spends nothing on them.

## Where we're going

Three things are waiting on you rather than on work:

1. **The category card grid** — the most widely used listing we have, on 15 pages
   across 7 sites. It has no picture slot and was never designed with one. Whether
   it should have imagery at all, and where those pictures would come from, is a
   design decision.
2. **The image budget sign-off** that has been formally open since we started
   spending on generation.
3. **Six article pages on robot-hands that don't exist** — build them or retire
   them. That decision sits with the site clean-up work, not with imagery, but it
   determines what the listing shows.

After those, the remaining imagery phases are the news feed and product pictures,
both of which have their own designs already written.

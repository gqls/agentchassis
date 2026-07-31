# SUMMARY 2026-07-31 — the image-404 check, fixed and closed

## What we're trying to do

Make the platform able to tell us when one of our sites is showing a broken image.
We have a check whose whole job is that, and it has been unable to see most of them.

## Where we've come from

The check was filed as a bug on 28 July, after the owner asked what a design sweep had
missed. It had reported one image that was working perfectly and stayed silent about two
that were not. A second session diagnosed it on 29 July, refuted the filed theory, found
the real mechanism, and — importantly — left a **test that deliberately pinned the
defect**, with a written warning about the one thing anyone fixing it had to work out
first. Then it stopped, because the obvious fix had been measured and refuted.

The mechanism is simple to state. Every image has a *purpose* — a role, like "hero" or
"logo" — and a *path*, which is where the file actually sits. The check was comparing the
two. If a page pointed at `hero.jpg` and the site owned any image at all whose role was
"hero", the check declared it fine without ever looking at whether the file was there.
The images most likely to appear on every page are exactly the ones with those common
roles, so the check was blindest precisely where it mattered most.

## What we've done

Re-measured the whole thing live first, rather than trusting a two-day-old note — the
bug file itself warned that two of its three test cases had already gone stale. Across
all 127 image references on our 13 live sites, checked against what the web server
actually returns: the old check was **reporting 21 working images and missing 6 broken
ones**. Six sites are painting a broken hero image, and nothing we run could have told us.

The fix turned out to be something we already owned. The code that *writes* our pages
already computes exactly where each image gets published, through one small shared
function used by five different parts of the system. Pointing the check at that same
function makes it the mirror image of the thing that writes the pages — so if the writers
ever move images, the check follows automatically. New numbers: **one false alarm instead
of twenty-one, and nothing missed instead of six.**

Three things came with it. The check had never looked at the site *chrome* — the header
that appears on every page — which is how idea.uk has been serving a broken favicon and a
broken social-sharing card site-wide without a peep. It does now. A branch that would
have automatically regenerated images turned out to be an exact duplicate of another
check we already run, so it was deleted rather than made safe; that is also what stops
this fix switching on an untested repair path. And pages that carry an empty image slot —
which no path-based check can see, and which a "just fetch the URL" checker would score as
*working* — are now caught structurally.

It went to the reviewer council, which sent it back once. The objection was a good one:
the shared function I was leaning on has a known flaw, and had I only patched the one
instance I happened to trip over, the check would inherit the rest. Five queries settled
it — the flaw has exactly one mechanism, it affects exactly two image types, both are
handled, and the audit now sits in the code so nobody has to work it out again.

## Where we are now

**Fixed, live, and proven.** It shipped in chassis v1.0.1219 and is running on both
replicas. We checked the binary properly — not just that the new code is there, but that
the old code is *gone* — and then ran real sweeps on two live sites. Nineteen image
references scanned, two reported, and both are genuinely broken. idea.uk's broken favicon
and social card were flagged for the first time. Nothing working was flagged. The bug is
closed and moved.

One thing is still owed and it is worth being straight about: the council's verdict on
record is still "revise". The answers were written and resubmitted, but the review panel
hit its API spending cap and could not sit again until midnight UTC. The commits are
marked as *submitted* rather than *reviewed* — a deliberate distinction here, because
claiming a review that did not happen is the one thing that would poison the record — and
the resubmission is written down as owed.

## Where we're going

The immediate value is already banked: the next design sweep on any site now surfaces
broken images instead of hiding them. Three follow-ons, all named and none urgent:

- **Resubmit to the council after midnight UTC.** Everything it needs is written.
- **Nine stale queue entries** left over from the old check — four of which are its own
  false alarms and can simply be deleted. Left alone deliberately, because another thread
  is actively working that queue and pulling rows out from under them would be unhelpful.
- **The shared function's remaining flaw.** It is harmless today, proven across every
  image we own, but a future image type could trip it. That fix belongs in the shared
  code with five other users, which makes it an architecture decision rather than a bug
  patch — so it is written down as its own item rather than smuggled into this one.

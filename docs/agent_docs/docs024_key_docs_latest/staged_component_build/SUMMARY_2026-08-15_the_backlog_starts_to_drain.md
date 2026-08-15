# SUMMARY — 2026-08-15: all three leftovers closed — the open question is with a person, the second image is fixed, and the wrongly-named image backlog is a third smaller

*The read-out on `bugs_open/248` (the "images get published under the wrong filename" bug)
after the follow-through sessions, written to be said aloud. The previous summary
(2026-08-14) ended with three named leftovers; this one is the story of closing them.*

## What we're trying to do

Fix a bug where freshly generated images — logos, icons, hero photos — sometimes got saved
to a site under the wrong filename, so the page that should show them displays a broken
image instead, silently, while the system reports success. The code fix already shipped and
was proven on one image; what remained was everything around it: a policy question the
review process raised, a second known-broken image, and roughly 150 images that were already
saved wrong before the fix existed and would never repair themselves.

## Where we've come from

The last summary ended with three concrete leftovers. One: the automatic review's fourth
round had asked a question no amount of further checking could answer — "should this
shortcut exist anywhere in the system at all?" — and that needed a person, not a fifth round
of review. Two: a second broken image (the mortgage-calculator site's homepage photo) almost
certainly had the same fix available but hadn't been checked. Three: nobody had designed a
safe way to clean up the backlog of already-wrongly-named images.

## What we've done

**The second broken image is fixed and confirmed.** It turned out to be the exact same bug
arriving through the other of the two routes we'd patched — not a new problem — which means
both halves of the fix now have real, end-to-end proof, not one proven and the other assumed
to work by similarity. Checked on the live page, not trusted from a success message.

**The policy question went to a person, in writing.** We wrote it up as a formal
architecture question (our RFC process) rather than resubmitting to automatic review. While
writing it we found something worth knowing: an earlier, related architecture question had
tried to count how often this part of the system attracts review complaints, and its count
missed this bug's two complaints entirely, because the risky code lives in a file the
counting method never looks at. So the write-up also flags that the earlier count is an
undercount, and points at an already-approved pattern elsewhere in the system as the likely
shape of the answer. It now sits with the owner.

**The cleanup got a design, then a careful first run — and the extra care paid for itself
twice.** Counting again fresh, 140 images across 14 sites still carried the wrong name.
Rather than re-triggering all of them, we sorted them into groups by what state their repair
paperwork was in, and the groups turned out to need different treatment: 26 were already
queued and fixing themselves; 13 just needed their stalled repair ticket nudged back into
the queue; 64 need a fresh ticket created; and about 30 have no ticket at all and need
checking individually, because for some of them the live page was already fixed by another
route and "repairing" them would actually overwrite a working image with an older copy.

Before running anything, the plan was reviewed once more at the user's request, and that
review changed it in two places that mattered. First, we now check the live website *before*
touching anything, and skip any image that's already being served correctly — that check
immediately caught two logos that a bulk re-run would have needlessly overwritten. Second,
we re-derived the list of what to touch with a stricter rule, which excluded a dozen images
whose repairs were already queued — touching those would have run the same repair twice.

Then we ran it: one canary image first, watched end-to-end, verified on the live site; then
the remaining ten in a wave. Eleven out of eleven came through cleanly — right filenames,
correct images confirmed live, and for the ten that were overwrites, the files came back
byte-for-byte identical, so nothing a visitor sees changed. Two of our own working
assumptions turned out wrong along the way (about which file extension icons should have,
and how underscores in names get handled) — both were caught mid-run by reading the actual
configuration, both are written down plainly in the record, and neither affected the
outcome.

## Where we are now

The bug's code fix is live, proven on both of the routes it patched, and the wrongly-named
backlog is down from 140 images to 98 — partly our eleven, partly the self-draining group
doing its job at the same time, which we watched happen. The policy question is written up
properly and waiting on the owner. The remaining 98 break down as: the bulk (about 64)
needing fresh repair tickets created the same way we just proved works; about 30 needing an
individual look before anyone files anything; and a small new category we discovered — about
a dozen images whose live pages are fine and only the internal record is stale, which need a
paperwork correction, not a re-deploy. Nothing needs to be improvised: the next person
inherits a proven method, a checked list, and the one small design question about that last
category.

## Where we're going

Run the main cleanup group (the ~64 fresh-ticket images) using exactly the method the pilot
proved, including the check-the-live-site-first rule, which we now know catches real
problems. Give the ~30 unchecked images their individual look. Decide — ideally when
designing that main run, not mid-run — whether stale-paperwork-only images get a records fix
that doesn't touch the live site. And separately, wait on the owner's answer to the
architecture question, which will settle whether the shortcut that caused all of this gets
retired system-wide or stays with guard rails.

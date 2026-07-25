# Where we are — the human review queue

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-07-25 — opening the drain

You have a queue of things the platform decided a person should look at. It has
370 items in it. Nobody has ever actioned a single one.

That is not because nobody could get to it. Five days ago another session fixed
the two reasons it was invisible — the dashboard was only ever loading the newest
50 items so the queue reported itself as empty, and there was no way in from
outside the cluster — and set up a VPN so you can reach it. Since that fix the
queue has grown by 67 and shrunk by nothing.

So I went looking for why, and the answer is not "nobody has got round to it".

**Most of what is in there is no longer true.** 321 of the 370 items describe a
page that has been rebuilt since the item was filed, and nothing re-checks them.
Here is a concrete one. On leopardessconsulting.co.uk, the "how we work" page has
two items from 10 July saying its two buttons have nowhere to point. The page was
rebuilt on 18 July, and both buttons now point at real pages. The items are still
sitting there saying otherwise. If you sat down to work this queue tonight, that
is the kind of thing you would be reading — findings about a version of the site
that stopped existing weeks ago — and you would have no way of telling those from
the ones that are still real.

That is the actual defect. Not the size of the pile, the fact that you cannot
trust any item in it.

**What I have built.** A sweep that goes through the parked items and re-checks
each one against the site as it stands today. Three possible answers per item:

- *this is no longer true* — close it, and record exactly what it checked and
  what it found, so the close is auditable rather than a bulk delete;
- *this is still true* — leave it in the queue, but stamp it with today's date,
  so when you open the queue you can see it was confirmed rather than merely old;
- *I can't tell* — leave it alone and say why.

The third one matters as much as the first. Some components on our sites render
from templates rather than from stored content, and for those the sweep genuinely
cannot answer the question. It says so instead of guessing. On today's numbers it
closes 51 items, keeps 35 with a fresh confirmation, and openly declines 72.

**Why it is safe to let a machine close these.** If it gets one wrong, the check
that raised the item in the first place will simply raise it again next time it
runs — closing an item releases the lock that was stopping that. So a mistake
costs one duplicate, not a lost finding. That is what let me build it as an
automatic sweep rather than something that needs you to approve each one.

**One thing in the original bug write-up turned out to be wrong**, and I want to
flag it because it was the fix everyone assumed we would do. The bug file says
there is an existing piece of machinery for exactly this — a "section data
reconciler" — that just needs wiring up, and that it would clear 48 items. I
checked: it would clear zero. It only handles a kind of missing data that no item
in the queue actually has. It is not broken, it just does not fit this pile. I
have written the correction into the bug file and the wrong-calls log.

**Where it goes next.** The code is written and tested and has gone to the review
council. It ships switched off — it will report what it *would* do without
touching anything, so we can read that first and only then let it write. Two
questions in the original bug are still yours and I have not touched them: what
to do with the ~78 items that are actually machine failures parked in the human
queue by mistake, and whether we ever record *who* made a decision (the handlers
have no login attached to them, so that one is an authentication decision, not a
one-line fix).

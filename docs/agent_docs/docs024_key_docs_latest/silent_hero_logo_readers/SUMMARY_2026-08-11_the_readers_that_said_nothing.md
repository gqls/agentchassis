# SUMMARY — 2026-08-11 — the readers that said nothing

Commission item 2. Written to be read aloud.

## What we're trying to do

Stop pages shipping without their hero image or logo while the system reports success. More
precisely — because the fix for *that* is a design decision the owner has reserved — make the
failure **visible** the moment it happens, so the decision can be made on evidence instead of on
guesswork.

## Where we've come from

A page that shipped with no hero and no logo looked exactly like a page that never wanted one. Three
places in the code took the address of a freshly deployed image, and when it wasn't there they did
nothing at all: no log, no error, no work item. That went unnoticed for five weeks.

Worse, every attempt to investigate it lost a race. The evidence lives in a run record that is
deleted after four hours, so the fault appeared to flicker in and out of existence — nothing on the
6th, two cases on the 9th, nothing on the 11th. Two automated diagnosis runs came back
"unverifiable": the first because the evidence didn't exist yet, the second because the diagnosis
tool couldn't see the table. That second blocker was commission item 5, which shipped last week and
proved the tool now works — and that the data was already gone.

## What we've done

The three readers now speak. When a deploy actually ran and its result came back without the
address, they record what happened — which key they wanted, and which keys the result *did* carry.
That key list is a fingerprint: it identifies this specific fault at a glance.

Two design decisions did the real work:

- **They stay silent when no image was wanted.** Most pages deploy neither a hero nor a logo, so
  the naive version would have filed a complaint on every page of every site. The presence of a
  deploy result is what counts as "something was expected here". A test fails if that gate is ever
  removed, and we proved it fails by deliberately breaking the gate.
- **The record is durable, not just a log line.** This goes beyond the letter of what was approved,
  which asked for a log line, and it was declared as the review question rather than slipped in.
  A log line here does not survive: the service is the busiest we run, and the run record is deleted
  after four hours. The record now goes to the one store documented as outliving this kind of step.

The council approved it first time, in about eleven minutes, and its two substantive challenges both
improved the work: one caught us repeating a measurement instead of making it (we then made it, and
it held), and the other asked whether the same pattern lurked elsewhere — so we counted, found 64
instances of the shape, opened every one that mattered, and established that **none of the others
has this bug**, because a missing value there either fails loudly or genuinely means "do the other
thing".

## Where we are now

Item 2 is built, tested, reviewed, approved and committed. It is inert until the next fleet release.

We also found something we did not go looking for. A workaround added in February writes the image
address to a second place, precisely because someone had noticed the first place gets overwritten —
and the codebase separately records that this second place is *also* wiped for exactly this kind of
step. If both are true, a five-month-old fix has never once worked, and that would explain both
halves of the original bug. We have put that to the automated diagnosis loop rather than write it up
as fact. **Its verdict is still pending, and nobody should treat the lead as confirmed until it
lands.**

## Where we're going

Three things, in order:

1. **The release**, which is whole-fleet and the owner's to run.
2. **A real site build that deploys an image**, which is the only thing that will prove this works
   in anger — a record appearing at the moment of failure. It cannot be forced from this lane.
3. **The owner's choice of the next commission item.** Item 2 was the last one that could be done
   without a ruling. Item 1 needs a design decision that is explicitly reserved; item 3 needs a
   routing call between the council gate and an architecture RFC, plus an unanswered modelling
   question about whether the page level wants a deploy commit as well as the component level.

The residual we are carrying deliberately: four code sites that look up a key supplied by
configuration, which cannot be classified by reading the code. They are written down as unfinished
rather than counted as clean.

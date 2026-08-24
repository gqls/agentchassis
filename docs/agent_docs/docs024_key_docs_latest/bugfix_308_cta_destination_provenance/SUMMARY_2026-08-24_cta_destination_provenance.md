# SUMMARY — 2026-08-24 — CTA destination provenance (`bugs_open/308`)

## What we are trying to do

Fix a bug where the platform spots a broken button, works out where it should point, files a
repair — and the repair cannot perform it. A button reading "Book a Discovery Call" linked to a
password-strength tool. The check saw it, named the contact page, the repair ran, reported success,
and changed nothing. Next pass it found it again.

## Where we have come from

Two phases, in the order the owner ruled on 18 August. First record properly which links the
machinery wrote, so a person's link can be told from the machine's. Then widen what the machinery
is allowed to consider, so it can point a button at the contact page at all.

Both were built, both were reviewed, and the second one was measured hard before it shipped —
because the measurement kept changing the design. It showed the change was nine times bigger than
the bug, that a third of its rewrites were being decided by alphabetical order, and that the
cautious version of it was worse than the wide one.

Then you asked for a sample of the rewrites to be audited by hand before it went out. That found a
defect nothing else had: **12% of them pointed a button at the page it was already on.** Fixed, and
the fix's shape was measured too — refusing the case outright beats quietly taking second best.

## What we have done

**It is live and it works, checked on a page you can load in a browser.**

On finetuning.uk, the button reading "Book a Discovery Call" now goes to the contact page. Not in
the database — on the served page. Four buttons on that page were repaired in one pass and every
destination works.

**And the checker got quieter in exactly the right way.** Re-run on the same site under the new
build: findings fell from 169 to 70. The fifteen findings telling the platform to move a correct
"how we work" button onto the About page are **gone**. The seven telling it to point a button at its
own page are **gone**. And the one *correct* member of that same family survived — which is the
result that matters, because a change that just silenced everything would have taken it too.

## Where we are now

Three council reviews approved, one of them at the fifth attempt after four rounds of my own
submission errors. Everything is registered, and the two traps I hit are written down for whoever
hits them next.

**One page is repaired, not the fleet.** New findings will drain on their own — they get triaged and
dispatched within a day, and they will now actually repair rather than completing unchanged.

**But 215 older repair jobs are stuck** in a state nothing picks up, some since 16 July. They are
the backlog this bug is really about, and releasing them is a deliberate act: it would rewrite
buttons across eleven client sites in one wave. That is the next decision and it is yours.

## Where we are going

Two things, and the first is a question rather than a task.

**Release the 215, or release them a site at a time?** The mechanism is proven on one site. A
staged release costs nothing but patience.

**Then close the loop the bug was really about.** The checker still works out the right answer,
writes it down, and the repairer ignores it and works it out again from scratch. And nothing checks,
after a repair, that the button actually moved — so "complete and unchanged" is still a possible
outcome. That is the defect underneath this one, and it will outlive it.

# SUMMARY — robot-hands.com — 2026-07-22 — the catalogue is real now, and a fleet bug fell out of proving it

Written to be read aloud. Five parts: what we're trying to do · where we've come from ·
what we've done · where we are now · where we're going.

## What we're trying to do

Make robot-hands.com a site that only says true things. It presents itself as a gripper
comparison service with a real methodology, so every number and claim on it has to trace
back to something real rather than to what a content generator found plausible.

## Where we've come from

The first pass (R1–R6, mid-July) fixed the visible breakage — the lost dark theme, the
dead buttons, the 404 articles — and built the MatchMatrix tool by hand so it couldn't
invent a dataset. That pass also turned up two deeper problems: the site was publishing
made-up statistics ("1,200+ gripper models" against an index of five), and, underneath
them, a claim woven through the whole site that it covers grippers across **six actuation
technologies** — pneumatic, electric, vacuum, magnetic, soft-robotic, adhesive. We
corrected the statistics as containment and left the six-technology claim as a decision
for the owner, because changing it is a decision about what the site *is*.

## What we've done

We put that decision to the owner, sharpened by a finding: the claim wasn't just
unsupported, it was contradicted — the catalogue's five grippers are all the same kind
(parallel-jaw, mostly electric), and four of the six named technologies had **zero**
grippers. The owner chose to make the claim true rather than soften it. So we added one
real, datasheet-sourced gripper for each missing technology — a Festo pneumatic, an OnRobot
vacuum, an OnRobot soft gripper, an OnRobot adhesive "Gecko", and a Schmalz magnetic — each
with only figures read off the manufacturer's own page, and the web address saved beside it.
The index now genuinely holds ten grippers across all six technologies, and we wired the
site's counters to count the catalogue directly so they can't drift from it again.

Proving that fix was live then exposed a smaller but real defect: one page was rendering the
right numbers with garbage units stuck on — "10%", "6ms", "39x". Chasing it down, we found
the cause is not on this site at all: a **shared component** hands out those junk units as
its *default* whenever a stat leaves the unit blank, and every site that uses it renders the
same nonsense (one shows "14,203%", another "1,000sms"). We fixed the default at the source
so a stat with no unit now shows no unit, and cleared and re-rendered this site. It renders
clean.

## Where we are now

robot-hands is correct end-to-end and checked on the live pages, not just the database: the
about page reads 10 grippers / 6 manufacturers, the detail page reads 10 / 6 / 4 / 39 with
no junk units, the six-technology claim is backed by ten real sourced products, and there
are no invented figures left. The shared-component fix is live, so the junk-units bug can't
recur anywhere new.

Two honest limitations remain, both recorded rather than buried. The MatchMatrix tool still
only compares the parallel-jaw grippers — clamping force is a jaw-gripper idea — so any
sentence implying the *tool* weighs up all six technologies is still ahead of the tool. And
the new grippers back the claim and the counts but don't yet get their own browsable
catalogue pages.

## Where we're going

Nothing on robot-hands is blocking. The remaining threads are other people's: the four
other sites still showing the junk units need their owners to set the real unit and
re-render (it's their call what the unit should be), and the broader "content generator
invents numbers" bug (043) still needs its fleet-wide sweep and a proper platform fix. For
robot-hands itself, the open questions are optional polish — giving the new grippers real
catalogue pages, and deciding whether MatchMatrix should say less or cover more.

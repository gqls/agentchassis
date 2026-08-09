# Summary — the site inspectors get a clock (bug 230), 2026-08-09

## What we're trying to do

Make sure every site we run is examined regularly for defects, whether or not a person
happens to be working on it that day. The platform has sixty-one automated checks — empty
sections, broken links, undeployed images, unverifiable claims, contrast failures — spread
across three inspector agents. They work. The question this piece of work answers is
simply: *what makes them run?*

## Where we've come from

Until today, the answer was "a human". Every scheduled entry pointing at those three
inspectors was a one-off, fired once by some session weeks ago and left switched off — five
of them, none enabled. The mechanism that was *designed* to drive them, the improvement
sweep, was paused in May during core build and never restarted. So examination followed
attention: the sites being worked on got looked at, and the rest went quiet. The bug filed
this pattern yesterday with a live example — two pages on finetuning.uk, a customer site,
serving empty sections that the checker would catch on sight, sitting undetected because
nothing had run the checker there since early August.

Two things made re-starting the old sweep the wrong move, and both were measured rather
than assumed. It skips any site with more than fifty outstanding findings — and the two
sites we work hardest are already at eighty-five and seventy-nine, so the busiest sites
would have been the ones it never looked at. And its way of choosing what to examine next
starves sites in a way that has been on record since April. Re-enabling it would also have
restarted the *repair* machinery, which is a separate decision that belongs to the owner
and is still open.

## What we've done

Built a rota. A small table remembers when each site was last examined by each inspector.
Once an hour, each inspector takes whichever site has waited longest, provided it has
waited more than a week, examines it, and files what it finds. A site that is mid-build is
skipped for that hour only — it keeps its place at the front of the queue rather than
losing its turn. Nothing is excluded for having too many findings, which was the old
sweep's fatal flaw. It only detects; it repairs nothing.

Alongside it, a daily watchdog that checks the rota itself: is any site going unexamined,
and are the scheduled runs actually happening rather than merely being scheduled? It
writes a report every day whether or not it finds anything, so a missing report is itself
a signal.

The review council approved it first time with four pieces of advice and no serious
objection. Two were worth acting on and were built in before anything went live: the
installer now refuses to run unless the three inspector names match real, active
inspectors, and the new machinery registers its own description where the next person will
look for it.

## Where we are now

Live since late morning, and proven. Within four hours the rota had examined five sites
with all three inspectors — fifteen examinations, every one completed, none failed. Then
it reached finetuning.uk on its own, in the ordinary course of the rota, and filed the two
empty-section findings the bug was written about, along with fifty-six other real defects
on that site. Nobody pointed it at anything. That was the exact test the bug set, and it
is the difference between a mechanism that works and a mechanism that has been argued for.

It also ran straight through a fleet rebuild in the middle of the afternoon without
missing a turn, which is worth knowing: this is configuration rather than code, so new
builds don't disturb it.

Two traps were found along the way and written down where the next person will hit them.
The record of who found a given defect names whoever *triggered* the run, not the inspector
that did the work — so everything the rota finds is filed under a generic name, and anyone
counting up results by finder would silently miss every scheduled one. And I made a
clock-comparison error out loud, claiming a run had been missed when it hadn't; that is
logged with the check that prevents it.

## Where we're going

The findings now accumulate honestly across the whole estate rather than only where
someone was looking. Nothing repairs them automatically, and that is deliberate: turning
the repair loop back on is the owner's decision, recorded separately, and it was always
gated on whether the repair agents are ready rather than on cost. The rota makes that
decision better-informed, because for the first time the backlog reflects what is actually
wrong with the estate.

The one thing to watch is that the backlog will now grow, visibly, with true findings. That
is the honest state of a system whose examination works and whose repair cadence is still a
pending choice — and it is a great deal better than the silence it replaces.

# SUMMARY — 2026-08-06, 08:10Z · the repair agent that never repaired anything

Written to be read aloud. Figures re-measured against the live system this morning, not carried
forward from yesterday's notes.

*(No summary was written when the fix was committed on 08-05, deliberately: the code was not
live, so the read-out would only have restated the bug file. It is live now, which is the
inflection this file marks.)*

## What we're trying to do

When one of our automated checks finds a small defect on a live page — markdown symbols showing
up as literal asterisks, a fabricated phone number, a page with no content — it files a repair
job and names the agent that should carry it out. We are trying to make those repairs actually
happen. Until this week they never did.

## Where we've come from

This did not start as its own job. It started as the last loose end on a different bug, 194,
where I needed to prove a fix by running a real repair on a real site. The owner offered a site.
Chasing that, I hit three walls in a row, and the third turned out to be a bug another thread
had already filed on the Monday — 201. The owner then chose that as the work.

Bug 201 recorded the symptom precisely: twelve repair jobs attempted, eleven dead outright, and
the twelfth reporting success having changed nothing. What it could not yet say was why, or
whether the obvious alternative would be any better.

## What we've done

**Found the actual cause, which was not what the error message said.** The failure read "planned
its own sections and none are ready", which sounds like the page's content was not available.
It really meant *nobody told it which parts of the page to work on*. The writer, when called
directly, expects the caller to hand it a list of the page's sections. These checks never sent
one — every one of the fourteen jobs of this kind that has ever existed lacks that field. So the
writer looked at an empty list, correctly concluded there was nothing to write, and gave up.

**Fixed it by completing a migration that was already half done.** Three checks now point at the
page build handler, which fetches the section list from the site's own plan rather than trusting
whoever called it. Two other checks had already been moved the same way; these three were the
ones nobody got to. I checked it two ways rather than trusting the reasoning: the build handler
is doing this successfully in production now — thirty-two completed repairs on already-built
pages — and I proved the safety test actually bites by deliberately breaking the value and
watching it fail.

**Put it through the review council, which approved it and corrected me.** Fifteen reviewers,
approved, five advisory notes, none serious enough to block. Two were worth acting on. One
pointed at a warning already on file that I had not read: **the build handler cannot see a
page's existing words**, so it rewrites the section from nothing rather than editing it. I had
written that cost down as "heavier than ideal". The honest sentence is that the old wording is
replaced. I corrected my own note rather than quietly reword it. Another reviewer warned that
dispatchers here have two gates and that fixing the visible one leaves work stuck at the second;
I went and checked rather than argued, and the second gate is clear.

**Shipped it, and then caught my own verification being wrong.** The fix went live yesterday
evening. I ran the deployment check I had written that morning at a reviewer's request — and it
reported that the fix had not shipped. It had. My check was grepping for a code comment, which
never reaches the compiled program. Worse, no version of that check could work here: the change
swaps one existing name for another, both already present, so there is nothing added to search
for. I replaced it with a check that looks at behaviour instead, and logged the mistake.

**Recorded what this says about the system, not just about this bug.** A reviewer pointed out
this is the fifth time the same mistake has shipped — five checks, each naming an agent that
could not do the work, each fixed by editing one word. Our safety net only checks that the name
is a real agent, never that the agent can handle the job. So the sixth will pass review too.
That is written up as a proposal with three costed options.

## Where we are now

**The fix is live and still unproven, and I want to be plain about the difference.** It is
running on both servers as of last night. But the proof requires the checks to file a new repair
job so we can see it carrying the right agent, and **as of this morning that has not happened**
— the sweep that runs these checks last ran before the fix went out. The result is zero, and
zero here means "not tested yet", not "working". That distinction has caught this workstream
twice already, so it is written into three separate places.

**The related bug, 194, is in good shape.** Its no-regression check has now been running for
nearly a day with real traffic behind it — about a hundred and fifty runs of the affected
components on the fixed code, and no faults recorded, against an error log that has taken over
nine hundred other entries in the same window. Yesterday morning the same zero was close to
meaningless because almost nothing had run. The formal reading falls due within the hour and I
expect it to confirm what we can already see.

**Half the original bug is untouched, on purpose.** The second symptom — a job that reports
success having done nothing — is deliberately left until this half is proven, because the
original filer pointed out that fixing it first would make the broken route look repaired.

## Where we're going

Immediately: take the formal 194 reading, and watch for the first new repair job so 201's fix
can be confirmed or refuted. Then the second half of 201.

Beyond that, three things are waiting on a decision rather than on work. Whether to fund the
structural fix so the sixth instance of this mistake gets caught by a machine instead of a
reviewer. Whether to release the lock on mortgagecalculator, which another thread is holding
pending your call on rebuilds. And what scope to give the ai-agent-orchestration.com rebuild —
that site is measurably damaged and unlocked, and the work is scoped and ready but not started.

One caution carried forward into all of them: a rebuild in this system **regenerates** copy
rather than correcting it. That is free on a page with nothing on it, and a real cost on a page
that currently reads well. It is the main reason the site work should start with the empty pages
rather than the damaged ones.

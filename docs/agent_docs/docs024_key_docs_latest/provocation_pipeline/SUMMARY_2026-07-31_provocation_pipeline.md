# SUMMARY (2026-07-31) — the daily provocation

Written to be read aloud. Current state only; the chronology is in
`README_where_we_are.md` and the technical log in `NOTES_provocation_pipeline.md`.

---

## What we're trying to do

vonc.com is built on one promise: every day, one provocation, and you argue it
against an AI on a twenty-minute clock. The provocation is the product — it is
what brings someone back tomorrow, and it is the thing you would share.

We want that promise to be true, and we want the provocations to be good: safe,
genuinely arguable, current, and interesting to people who like arguing. The
owner's direction is that Grok generates them from what is actually being argued
about on X, behind a filter that catches slop and danger, with categories
eventually spanning politics to pets — and no human approving each one.

## Where we've come from

The owner noticed the provocation had not changed that day, and said he did not
think it ever had. That turned out to be right in substance, and the cause was
worse than a broken job: **there was no mechanism.** No scheduler, no pool, no
rotation logic. The feed was a static file with the provocation written into it
as a fixed paragraph, edited by hand six times since June and last touched on
26 July. The archive had stopped 25 days earlier, and today's provocation could
not join it because it carried neither a date nor an identifier.

A plan from 25 June had designed all of this already, as a copy of the news-feed
pipeline. Half of it shipped — the feed file and the page code we use today came
from it — and the half that chooses and regenerates content was never started.

## What we've done

**Established the ground truth, and corrected two beliefs.** The handoff we
picked up suggested the "See All Provocations" page might also be broken; it is
not, and that was worth knowing because it removed a repair from the plan. More
importantly we found that the **game engine reads the feed server-side**, which
rules out the obvious cheap fix of rotating in the browser — that would have had
the page showing one provocation while the Gauntlet argued another. We also found
the scheduled-publish machinery already exists and runs every six hours on other
sites, so the expensive-looking half of this is plumbing we already own. Grok is
likewise already wired, with X search enabled.

**Shipped Phase 0, and it is live.** The builder is now schedule-driven: each
provocation carries the date it goes live, today's is whichever has most recently
arrived, and the archive is everything published before it. That makes the owner's
archive rule — archive it when the next one is published — a property of the data
structure rather than a step someone has to remember. The feed now carries a real
generation timestamp instead of a hardcoded one, and today's provocation finally
has the identifier and date it needs to be archived. Published, served in 45
seconds, and all three pages verified by rendering them: nothing broke.

**Built the publish and rollback path as one command**, so that publishing
forward exercises the mechanism a revert would use.

**Prototyped the paired mode** the owner asked for — an organiser sets a
provocation for a named team, everyone commits without seeing anyone else, and
the positions open at once. The seal is enforced by the type system rather than a
check, so a handler cannot leak a position by forgetting a condition. Fourteen
tests, mutation-tested.

**Recorded four mistakes**, because they are the transferable part. A working page
read as broken. An atomicity test that was green against a broken implementation
because it held the clock constant. A rotation checker that asserted a date was
*present* and never that it was *correct* — the same trap we had filed as a
landmine four hours earlier, about the same file. And a rollback that refused to
run, because its safety check demanded the very field the rollback existed to
restore.

## Where we are now

**The site still does not rotate, and "every day, one provocation" is still a
false claim.** Phase 0 bought capability, not behaviour: the schedule's last entry
is 26 July, nothing rebuilds on a cadence, so tomorrow serves what today serves.
This is the single most important sentence in this summary, because everything
else reads like progress and this is the thing the owner actually asked for.

Two things are needed and neither works alone — provocations to rotate *to*, and
a scheduled job that rebuilds and republishes daily.

**A second lane is now working inside our builder.** The Gauntlet lane took the
owner's ruling on the sealed provocation and enforced it in the feed rather than
in each page: today's question is readable in the Gauntlet after entry and nowhere
else, with home and the Arena showing a past provocation in full instead. Their
first attempt removed today's headline and body from the feed to seal it — which
would have broken every round, because the engine reads exactly those keys — and
the in-flight correction restores them and seals at the display level instead.
That correction is **uncommitted in the working tree** and it passes all our
rotation invariants.

**So the builder is co-owned, and publishing is now a coordination hazard.** Our
publish path is one command, and running it today would ship the seal data to a
site whose renderers do not understand it yet: the lobby card would say "sealed"
while the page above it still printed the provocation in full. Verified — the
served JavaScript still reads `today.headline` and knows nothing about the new
keys.

## Where we're going

Immediately: a short bridge of hand-written provocations plus the daily scheduled
job, so the claim becomes true and the whole publish path gets exercised on a
cadence before any generated content touches it. That has to be sequenced with the
Gauntlet lane's renderer delivery, because we share the file and the publish
button.

Then the Grok generator and the filter behind it. Because there is no human
approving, the filter is the only control, so it has to fail closed — publish
nothing and keep yesterday rather than publish something unjudged — treat its own
errors as rejections, log what it rejects, and be calibrated against our existing
nine provocations before it is wired to anything that publishes.

After that, categories, which are not just labels: each needs its own safety
threshold and audience, and more than one simultaneous daily will not fit the
engine's current one-provocation-per-site contract, so that is a conversation with
the Gauntlet lane before it is a task.

The paired mode is prototyped and waiting. It needs real identity, which the
platform does not have at all today, and that is a bigger prerequisite than the
provocation work itself. It is worth doing sooner than it looks, because it is
what produces the named, returning contestants whose behaviour is the only thing
that will ever tell us which provocations are actually interesting.

# SUMMARY — 2026-08-05 — the code review, triaged and actioned

First in the series. Written to be read aloud.

## What we're trying to do

Take fifteen findings from a code review run over the working diff, work out which are real,
and act on the ones that are — without trampling other sessions who are mid-flight in the same
files, and without leaving a durable claim behind that turns out to be false.

## Where we've come from

The review ran this morning against changes three lanes had committed the evening before. A
session triaged all fifteen: it worked out who owned each, checked each claim against the code
or the live database where that was cheap, and rated one of them a false positive. It
deliberately fixed nothing, because all three lanes looked active and the house rule is to
contribute into a lane rather than compete with it.

That was right when written. It stopped being right about sixty seconds later, when two of the
three lanes closed their bugs and moved on without ever picking the review up.

## What we've done

Nine findings fixed and committed across four commits, each one submitted to the review
council. The two most valuable were not the ones that looked most serious at triage time.

The first was a config key called `require_sections_metadata`, which yesterday's work gave the
meaning "refuse to save this page" — while that exact spelling was already live in the same
package meaning something far milder, "warn me that a check could not run". One of our agents
carries both steps in a single definition, so the same word would have meant two different
things a few lines apart, and the normal way to roll a setting out across the fleet is a bulk
update by key name. We found a comment in our own codebase that had already been confused by
it. Renamed, at zero cost, because we had shipped the new key switched on for nobody.

The second was a detector that was scoped more narrowly than the deletion it exists to predict
— it only looked at pages marked "deployed", while the deletion it warns about takes every
row. Anything destroyed on a page in any other state was silently never reported, which is the
exact silence the detector was built to end.

Three findings turned out to be wrong, or wrong as stated. One asked us to call a shared
logging helper that cannot be called from where it sits — it would be a circular import, which
is precisely why about twenty files in that package each write their own copy. One claimed we
write unboundedly to a table nothing cleans up; there is a cleaner, it runs hourly, and it had
run minutes before we looked. And two more were assigned to the wrong lane because ownership
had been worked out per file, when two lanes had edited the same file twenty-six minutes
apart.

## Where we are now

Twelve of the fifteen are resolved: nine fixed, two recorded as false positives with the
measurements that refute them, one confirmed and handed back to the session that owns the file
with its corrected line numbers. Three remain open in different senses — one we cannot
recover, because it is named in the triage table but described nowhere and the original review
output was never saved; one is waiting on its owner; and two council verdicts were still in
flight.

The first verdict came back approved, with one criticism worth repeating: we had claimed two
functions had no callers anywhere on the strength of a text search, and the reviewers said,
correctly, that they could not verify that. We redid it as a check that could actually fail —
rename the functions, rebuild, see what breaks. Nothing did.

The honest headline is that this review found two real traps and two false alarms, and that
one of the false alarms nearly became a filed bug because the evidence against it looked
exactly like evidence for it.

## Where we're going

Read the two outstanding council verdicts and act on them, since the code is already on the
shared branch. Notify whoever owns the endpoint-health file so the stale line anchors get
fixed by the session that has it open. Decide whether the `domain` inconsistency we found on
the way — a shared writer stores an empty string where it means "unknown", so a fleet-wide
count reads twenty-six times too low — is worth a lane of its own; it is real, it is nobody's,
and it is not what any of the fifteen findings asked about.

The wider lesson, which is going into the standing docs rather than staying here: ownership on
this tree expires in minutes, and a claim about what does *not* exist needs a lookup in the
language the thing is actually written in. Both false positives in this review were absence
claims made without one. So was the mistake we nearly made while checking them.

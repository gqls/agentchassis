# SUMMARY — bugfix 234: a key nobody read, and the class it belonged to

*2026-08-11. Written to be read aloud.*

## What we're trying to do

Stop the platform silently ignoring instructions it was given. A workflow step is
configured with named settings; if a step names a setting the code does not read, the
platform shrugs and carries on. The config still *looks* like a specification of
behaviour, and is evidence of nothing.

## Where we've come from

Three live steps had been setting a key called `spec` for months. The action that reads
those steps has never read a key by that name — it takes the item spec from `spec_data`,
`spec_paths` or `spec_literal`. So every record those steps filed carried an empty spec.

The one with teeth: whenever the improvement loop finished fixing a site and asked for the
pages to be re-assembled, it also asked for the shared header and footer to be refreshed.
That request was discarded every single time — seventeen records, none of them noticed,
over eight days. The consumer chain was fine; the flag simply never arrived.

It was found by a thread doing something else entirely: before switching on a detector for
unread config keys, they asked what it would report, and this fell out. Nothing had failed.
Nothing had logged. A council seat had been shown this exact config in a review months
earlier and could not see it either — it is not visible by reading, only by comparing
against what the code reads.

## What we've done

**Fixed the instance.** One migration renamed the key on all three steps, copying the
values in place so nothing could be mistyped, with its own safety checks deliberately
broken three times first to confirm they catch a partial application. The seeds were
corrected in the same commit so a future reseed cannot bring the dead key back.

**Closed the class, two ways.** The config checker had three states — recognised, unknown,
and old-name-still-works — and no way to say *this name is retired*. That gap is why the
bug survived: an unknown key produces a warning nobody reads while the behaviour quietly
doesn't happen. Now a retired key is a hard refusal whose error message names the correct
spelling. And the action at the centre of this bug was switched to strict, so any
unrecognised key on it is a definition error caught immediately instead of an empty record
found by archaeology months later.

**The council rejected it, and that produced something better.** The guardian seat vetoed
not the fix but *how* it arrived — a hard failure in machinery every message passes
through, packaged inside a bug fix. The architecture seat, same round, same evidence, said
the opposite. House rules for that situation are clear: the code stays, the design gets
written down, and a human breaks the tie. You did, and your answer improved it: adoptions
are now licensed by a measurement plus a daily automatic check, rather than by ceremony.
That check exists, runs at 06:25, and writes down its "all clear" — so if it ever stops
running, the missing note is the signal.

**Then we used the new mechanism on the next case**, retiring two more dead keys that a
previous lane had adjudicated but had no way to retire.

## Where we are now

Everything is live on the current build and proven at the artefact rather than inferred.
A deliberately misconfigured test agent was refused by the running system, with the
offending key named, classified as permanently broken rather than retried. A second test
agent carrying the improvement loop's exact configuration filed a record that *did* carry
the refresh instruction — the thing seventeen previous records could not do. Both test
agents were removed afterwards and the fleet re-checked clean.

Three mistakes were made along the way and all three are written up, because they are the
part that does not re-derive. The most useful: when the first test found nothing, the fleet
happened to be down, and I accepted that as the explanation. It was not — my test was
watching for something that never exists when a configuration is rejected. A believable
outside reason for "nothing happened" is exactly when to distrust your own instrument.

## Where we're going

Nothing technical is owed on this lane. What remains is yours to decide: whether to put the
now-complete package back through the council (the objection that drew the veto has since
been answered by the automatic check), whether to extend the strict setting to a second
action that is now eligible, and whether the neighbouring lane should retire the last known
dead key using the mechanism this work built. None of them is urgent, and none blocks
anything else.

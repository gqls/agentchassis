# SUMMARY — chrome divergence guard (bug 226), 2026-08-09b

*Second summary of the day, because the read-out genuinely changed twice: the
morning file said the fingerprinting would spread on its own (it wouldn't —
see its correction block), and by mid-afternoon the owner had ordered the
spread done directly, which changed both "where we are" and "where we're
going". The morning file stands, correction and all — the series is the
record.*

## What we're trying to do

Stop the platform ever silently destroying hand-made content in a site's
header, footer or head. Every overwrite must keep a copy, and an overwrite
that destroys hand-made work must raise its hand.

## Where we've come from

The morning summary tells the build story: the two-part design (archive
trigger + fingerprint stamp), three council rounds, the passed fire-drill.
What it got wrong was the rollout: it assumed a "wave" would fingerprint the
fleet on its own. There was no wave — the new hourly discovery checker only
observes, its pickup half is an open bug (83), and the only thing actually
spreading fingerprints was unrelated rebuild traffic. The owner's question
("has the wave finished?") exposed that; the correction is logged where the
claim was made, and in WRONG_CALLS.

## What we've done

The owner then said: dispatch the rebuilds. The platform had no tool that
rebuilds just the chrome — everything nearby either fans out one job per page
(five hundred plus, into a queue that isn't draining) or edits templates on
the way through. So we built the missing narrow tool: `rerender-chrome`, a
two-step agent from existing parts (no new Go, no LLM calls), seeded as
migration 351, registered as STY-055 with its one landmine stated (a green
run is not stamped slots — locks refuse silently). Fired it at all fifteen
eligible sites: fifteen clean receipts, fifteen completed runs, zero errors,
ten superseded artefacts archived on the way out, thirty-five byte-identical
no-ops that correctly archived nothing.

## Where we are now

Eighteen of twenty sites are fully fingerprinted — 54 of 60 chrome slots. The
six unstamped slots are the two mortgage-calculator sites whose hand-authored
chrome sits under permanent locks: the rebuild refused them by design, and
they stamp if an owner ever unlocks them. Every slot in the fleet, stamped or
not, archives on overwrite; the stamped ones also alarm. And the alarm has now
fired for real: a platform repair agent keeps patching the darts site's
header, rebuilds keep erasing the patch, and what used to be an invisible
tug-of-war is an open review ticket naming the erased bytes and their archive
copy. That ticket is the system working — left open for a human on purpose.

## Where we're going

Three small things, none urgent, none this lane's to force. The darts-header
ticket wants its patch moved into template or config so rebuilds reproduce it
(the STY-050 route). The two locked sites stamp whenever their locks lift.
And someone who owns the Spanish watch site should look at why its fresh head
section inlines a 52KB stylesheet — spotted in passing, archived, parked. The
guard itself is done: built, reviewed, approved, rehearsed, fleet-armed, and
already refereeing a real fight nobody knew was happening.

# Where we are — the path to amend an image

Plain-prose log, append-only, newest at the bottom.

---

## 2026-07-29 — built and tested on the bench; not yet deployed

You asked for the missing piece: a way for a person to hand the platform a corrected image.
Until today there wasn't one — every picture in the estate was machine-generated, and when a
generated picture was wrong (three sites' "logos" turned out not to be logos) there was no door
to hand in a better one.

The door now exists. One command takes a file from disk, puts the bytes into the database, and
asks the platform to take it from there. The platform — which already holds the storage keys —
checks the file arrived intact, checks it really is an image, puts it into storage under a new
name (the old file is kept, never overwritten), and updates its records: what changed, when, by
whom, why, and what was there before. If the asset was human-approved and locked, the platform
refuses — an approved image can't be silently replaced, by this door or any other.

Deliberate choices worth knowing: the operator never touches the storage credentials (that was
the whole point); the bytes go through the database rather than the message bus (the bus has a
one-megabyte ceiling and images don't fit reliably); and every refusal is loud and recorded —
"this wasn't an image", "the bytes got corrupted", "that asset is locked".

It is written, reviewed by machine tests, and waiting on three things: the reviewer council's
verdict, a build-and-deploy of the new code, and two small database changes (the second only
after the deploy is proven — a config that names code the fleet doesn't have yet fails at
runtime). Then the first real use: relojistas' cropped logo goes in through this door, and the
header, tab icon and social card get re-made from it.

A small piece of luck: the sibling fix that stops tab icons coming out squashed — which we
wanted live before re-making relojistas' icon — was finished by the other session while this was
being planned. Both changes travel in the same deploy, so the ordering sorts itself out.

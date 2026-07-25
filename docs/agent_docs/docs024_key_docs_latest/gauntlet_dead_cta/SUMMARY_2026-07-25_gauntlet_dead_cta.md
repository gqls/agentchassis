# SUMMARY — 2026-07-25 — the machine built its first feature: PR #3 is open

*(New file per the never-overwrite rule; previous read-out: SUMMARY_2026-07-24c.)*

## What we're trying to do

Make the vonc.com Gauntlet a genuinely working tool: a real AI debate opponent —
you file a Position on today's provocation, the AI files a counter and a
challenge, you defend on the clock, and it judges honestly whether your take
held up. The backend for that is `tools-api`, and we chose to have the
platform's own feature-builder write it (milestone B4 — its first ever real
build), with the owner's review of the machine's pull request as the one hard
gate before anything ships.

## Where we've come from

By yesterday evening the plan was council-approved (the designer's first-ever
convergence), but two implementer runs in a row had died mid-build in the same
mysterious way: work committed fine, then the run simply hung and failed hours
later. We'd blamed a server restart for the first one.

## What we've done

This morning we found the real killer, and it was our own housekeeping. A
cleanup job runs every ten minutes and deletes leftover message channels — but
its "is anything running?" check looked for a label that no running agent has
ever worn, so it always concluded the coast was clear and swept everything,
including channels under agents mid-job. The builder's reply would be posted to
a channel that had just been binned, and the run hung forever. We proved it end
to end (the reply was demonstrably sent, four seconds after being asked for,
into the void), fixed the check, applied it live, and watched it hold: at the
next sweep the cleaner correctly counted 39 busy agents and kept everything.
The wrong restart theory is recorded as a wrong call, with the lesson: check
the ten-minute grid before blaming the nearest dramatic event. Filed as bug 071;
the bug-003 owners have a note, because this likely explains some of their
silent failures too.

Then we re-fired the builder on the same approved plan. It ran all six stages —
scaffold, container and deployment files, error/CORS plumbing, rate limiting
and input caps, the rounds table and today's-provocation endpoint, and the two
AI endpoints (counter-position and verdict) — each stage gated and committed,
passed the final test gate, and opened **pull request #3**: 18 files, all new,
nothing touched outside its own service. Every message-reply that killed the
previous runs was consumed cleanly this time.

## Where we are now

The complete debate-opponent backend exists as a machine-written PR awaiting
the owner's review — the hard gate. The cleanup fix is live fleet-wide and
behaviourally proven. The feature-builder is now proven end to end:
capability gap → design → council approval → staged implementation → PR.

## Where we're going

Owner reviews and merges (or rejects) PR #3. On merge: build the image, deploy
it to the island VM (the exposure home decided yesterday — not the in-repo
cluster manifests), apply the rounds migration to the island's own database,
smoke-test the three endpoints through tools.apis.uk. Then the experience
re-plan re-fires with liveness evidence, the front-end is rebuilt against the
real API, and the whole journey gets the user-perception acceptance run the
original complaint demanded.

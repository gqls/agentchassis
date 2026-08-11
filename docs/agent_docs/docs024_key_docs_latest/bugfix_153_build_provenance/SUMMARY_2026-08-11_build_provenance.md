# SUMMARY — `bugfix_153_build_provenance`, 2026-08-11 night

First summary this lane has written. It has earned one now because the lane just closed: both of
its bugs are fixed, live, and — as of tonight — actually proven, not merely shipped.

## What we're trying to do

Answer one question the platform could not answer at all before this lane started: **"did my fix
ship?"** — for a single service, and, separately, for a whole release. Before this work, a backend
binary carried no record of what commit built it, and a release tag carried no guarantee that all
fourteen services under it were even built from the *same* commit. Both were invisible failure
modes on a tree that ~40 sessions commit to concurrently — the kind of gap that produces confident,
false "it's live" statements rather than obvious errors.

## Where we've come from

`bugs_open/153`, filed 2026-07-29: an image tag does not imply a rebuild, and nothing in a running
pod could say what commit it was built from. The standing verification practice at the time was
`strings <binary> | grep "<marker>"` — which this lane later showed was unsafe on its own terms
(absent tool, silent false negative, a discovery-grep that matches Go's internal digit table and
"finds" a match on every service regardless of truth). Three false readings were produced by that
practice in a single day before the replacement shipped.

## What we've done

**BLD-019** — every backend binary and image now carries the commit it was built from
(`pkg/buildinfo`, an OCI `revision` label, a startup log line). Verified across three separate
rolls, with a local regression guard proving the linker flag genuinely reaches the binary rather
than being silently dropped (a real risk with `-ldflags -X`).

Rolling it out surfaced a second, previously undetectable bug: **`bugs_open/249`** — one release
tag could ship *several* source revisions, because each of the fourteen service builds resolved
`HEAD` independently over the ~6 minutes a release takes, and on this tree commits land inside
that window routinely (13 in the busiest 7-minute stretch on 2026-08-11 alone). `v1.0.1284`
demonstrated it for real: three revisions under one tag, harmless that day only because the
commits in between were docs and tests.

**BLD-020** — `release`/`release-backend` now resolve the ref **once**, up front, and pass it to
every service build (`pinned_sweep`). Committed and registered the same day as `249` was filed.

We also rewrote CLAUDE.md's build-and-deploy section to retire the marker-hunting/`strings`
practice fleet-wide, replacing it with "ask the service what it's running" (the startup log line)
falling back to a controlled binary probe — and logged every trap we hit along the way (five
missteps in `NOTES_build_provenance.md`, several now in `LANDMINES.md` and `WRONG_CALLS.md`) so
the next lane doesn't re-pay for them.

Four things were explicitly decided rather than done: the production induced-fault test was
superseded by a local regression guard (owner call, `v1.0.1284`'s real divergence made the point
without lying to production); builds and deploys stay manual, so `153`'s refusal-mechanism
candidates stay unbuilt; the cross-service revision-mismatch assertion was offered and not taken;
and the council round on `153` was REJECTED on scope and overruled by the owner, so no
`Council-Reviewed` trailer exists anywhere in this lane's history.

## Where we are now

Both mechanisms are **live and proven**, not just shipped. `v1.0.1286` (the first release under
the pin) came out coherent but, on inspection, could not have discriminated a working pin from a
broken one — nothing committed during its build window. `v1.0.1287`, tonight, did: a real commit
from another session landed five seconds into the build window, and all fourteen services still
came out on the one commit the pin resolved at the start. That is the specific situation the fix
was built to survive, and it survived it, confirmed two independent ways (image labels and the
running pods).

`bugs_open/249` is closed. `bugs_open/153` has been live for three rolls. Both stay in
`bugs_open/` on standing owner direction — the fix being live is a fact about the code, not
permission to retire the ticket.

## Where we're going

Nowhere, as far as this lane is concerned — there is no further owed work. Three things are
recorded as open questions belonging to *other* lanes, not to this one: three images
(`component-render-check`, `shared-output-fields-check`, `removed-config-keys-check`) carry the
OCI label without a stamped binary, because `ref_build` labels at the docker CLI regardless of
whether the dockerfile ever uses the linker flag; the release path — arguably the most shared
mechanism on the estate — sits outside the council gate's review scope while a one-line change
inside `pkg/` draws a full round, observed but not proposed as a rule change from one instance; and
`153`'s deferred candidates (making `push`/`deploy` *refuse* an unbuilt or mismatched tag) remain a
live option if the owner ever wants to move from manual to enforced.

# SUMMARY — bug 020 gate live and wired (2026-07-22)

*Second milestone read-out. The gate moved from built-and-inert to live-and-wired.*

## What we're trying to do

Stop the platform inventing data. When we adopt a site with an interactive tool
whose behaviour depends on a real dataset (a searchable directory, a data-backed
chart), the recreation must load from the *same* source — never fabricate a
realistic-looking dataset to make the widget appear to work. Bug 020 is the case
where it did exactly that and shipped fake vet practices and postcodes to a live
public site.

## Where we've come from

The vetcomparison thread found and contained it and filed the diagnosis. We fixed
it in two halves: a prompt contract (live since 21 July — never invent records,
preserve the source, honest empty state) and a mechanical gate in code that does
not rely on the model obeying. The gate was built, unit-tested, and put through the
reviewer council, which — across two substantive rounds — made it materially better
(most importantly, it caught that the gate itself was "failing open", the same
silent-yes that caused the original bug; that is now fixed in the code).

## What we've done

The new platform build rolled out, so we switched the gate on. Before touching
anything we confirmed the new check was actually present in the running program (not
just tagged), and found it was sitting there unconnected — built but wired to
nothing. We connected it with a carefully-guarded database change: every tool
recreation now passes through the fabrication check before it can publish, and
anything bearing the fingerprint of invented data is held for a human instead of
going live. We verified the routing both ways — a flagged tool stops at the review
pile, a clean tool publishes as normal.

## Where we are now

The core protection is **live and working**. Two things remain before we can call it
fully done and closed: this particular build was cut about twenty minutes before we
finished the small "fail-safe" hardening from the review, so the running version is
the slightly-less-hardened one (it still catches real invented data; the tweak only
covers an odd empty-output edge that isn't fabrication); and we have not yet pushed a
deliberately-fabricated tool all the way through the live system end to end.

## Where we're going

The decision was to leave bug 020 open and finish it cleanly on the next build —
there's no urgency now the protection is live. On that roll: confirm the fail-safe
hardening is in the image, run one deliberate end-to-end test (recreate a data-backed
tool, confirm it is held and not published), then close 020 and lift the hold on tool
imagery. The exact steps are in the runbook.

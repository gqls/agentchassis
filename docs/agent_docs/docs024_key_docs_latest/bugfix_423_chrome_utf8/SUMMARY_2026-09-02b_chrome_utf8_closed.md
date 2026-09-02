# SUMMARY 2026-09-02b — bugs_closed/423: closed, and what the four council rounds were actually for

## What we're trying to do

Stop the site-chrome pipeline losing a footer quietly. Two things had to become true: the
corruption had to stop, and a database refusal had to stop being reported as success.

## Where we've come from

The morning summary had the mechanism found and the fix committed but **not live** — Go
changes do nothing until an image rolls — and it carried one open risk and one open
question. Both are now settled, and not in the way the morning expected.

## What we've done

A fresh chassis build (`v1.0.1354`) shipped the fix. We proved it at the binary rather
than at the tag, with a removed-string control — the text the change deletes is absent,
the text it adds is present.

Then we proved it **behaved**, which a probe cannot tell you. Garden Tools' footer had been
NULL for ten days; it stored at 16:21:32Z, 2,427 bytes, digest matching. Boxing Online's
followed at 16:27:56Z, replacing the hand-patch, and passes that lane's own pre-delivery
check — empty `sites.email`, no contact block served. Footers fleet-wide that could not
store: **two, then zero**. And the label that caused the whole thing renders intact, em-dash
and all — which is the check that distinguishes "the bug is fixed" from "the offending text
got silently dropped".

The bug is **closed**: fixed, live, and verified at the artefact.

The four council rounds are the more interesting half. Round 1 was gated on a real fault —
I asserted a blast radius instead of enumerating it. Round 2 was gated because my rationale
and my sketch had drifted apart in one edit, so the reviewers were judging a diff that
contradicted its own explanation. Round 3 approved, and one of its advisories then made me
**delete** the mechanism rounds 1–2 had argued into existence. Round 4 objected to the
deletion.

At that point three seats had pushed in three directions, which is precisely the case our
own guidance says a human should break rather than a fifth round. The owner ruled: ungated.

## Where we are now

Closed and live. One behaviour change is committed and rides the next roll — after it, a
chrome slot that cannot be stored **and has nothing to serve** fails the build instead of
shipping a footerless site. Measured population that can trigger it today: zero.

Two residuals are tracked rather than implied: `bugs_open/435`, a related silent branch we
deliberately did not fix because the measurement said fixing it would file about sixty-nine
findings about sites nobody ever built; and a configuration declaration that would let an
existing estate-wide counter see this action's settings.

## Where we're going

Nothing here needs a next session. What should outlive it is in `016b` §9 and in the
memory: that Postgres *refuses* invalid UTF-8 rather than degrading, so this defect class
surfaces as a write failure far from its cause; that the census which discriminates is "a
word whose first character is multi-byte", not "contains multi-byte characters"; and the
one I'd most want inherited — **a count of consumers is not a count of occasions**. I
measured seven workflows, called it a blast radius, and designed a safety flag on it. The
number that mattered was one row, and I never ran that query. It survived a high-severity
objection and two further rounds precisely because it had a date and a query attached and
therefore looked like rigour.

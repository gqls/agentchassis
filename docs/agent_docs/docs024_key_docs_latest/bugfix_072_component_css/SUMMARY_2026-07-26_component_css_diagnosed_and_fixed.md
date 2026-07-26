# SUMMARY 2026-07-26 — the frozen stylesheet, found and fixed

## What we're trying to do

Make sure that when a page ships markup, the styling for that markup ships with it. The
symptom that started this was narrow — news cards rendering as bare text on two sites out
of five — but the thing underneath it is general: we had no guarantee at all that a
component's markup and its CSS would both arrive on a page.

## Where we've come from

The bug was filed on 25 July by another thread, which hit it in its own components first,
fixed those, measured the same shape in the news feed, and then stopped and said plainly
that it did not know the cause. That honesty is what made this tractable: the measurement
was solid and there was no wrong theory to unpick.

For context, the platform has two ways a component can get styled. It can carry its own
styles inside its template, or it can rely on the site-wide stylesheet. Nobody had ever
established which of those was the rule, and the answer turned out to matter.

## What we've done

We found the cause, and it is not where the symptom is. A site's stylesheet is written
exactly once — by the design agent, during a design run — and at that moment it absorbs
the styles for whichever components the site had. Nothing ever regenerates it. So the file
is frozen in time while the site keeps changing around it. Any component a site gains
afterwards has markup on the page and its styles written nowhere. On one of the two
affected sites the gap is eighty days.

We also checked a site that *works*, which is the step that mattered most. It refuted our
first explanation outright and forced a more careful one. Without that control we would
have published a tidy, wrong account with dates attached to make it look verified.

We then fixed it in the place that closes the whole class rather than the one symptom:
the page now collects its own components' styles at the moment it is assembled, so
whatever builds a page also styles it and the two cannot drift apart. We wired that into
both of the code paths that assemble pages, because the second one would otherwise have
stripped the fix back off again.

Alongside that we brought the two news components into line with what 86 of our 94
components already do — carrying their own styles — and made the two mechanisms aware of
each other so they never both apply the same rules. The styles were copied from the
existing source rather than retyped, which is what guarantees the three sites that already
look right see no change at all.

We deliberately did not regenerate any stylesheet. Doing so re-runs the design pass and
re-rolls a site's colours; we would have been redesigning two live customer sites to fix
one section.

## Where we are now

The fix is written, tested and committed. It is not live, and we have been careful to say
so rather than let "committed" read as "done". The code half is Go, so it sits inert until
the next image roll. The database half is applied but cannot appear on any page until that
page is re-rendered, which we have not done.

Nothing on any live site has changed. The two affected sites still show bare news cards
today, exactly as they did this morning, and no customer site was touched.

The bug therefore stays **open**. Our bar for closing is fixed *and* live, and this is
fixed but not live. The bug file now carries the full diagnosis, the evidence, an explicit
note on the one part that remains inferred rather than measured, and the exact sequence to
verify after the roll.

Two things surfaced along the way that are worth more than the fix itself. First, a rule
about where a style block must sit inside a component: put it in the obvious place and the
platform silently discards it, leaving a component that looks perfectly healthy in the
database and ships nothing. We proved that both ways rather than trusting the reading.
Second, the survey that reframed the whole job — 86 of 94 components already carry their
own styles — which turned this from inventing a convention into finishing one.

## Where we're going

Next image roll: confirm the code is actually in the running pod, re-render the two
homepages, and measure all four sites. The two that already worked must come back
unchanged — that control is what distinguishes "added the missing styling" from "quietly
restyled three customers' sites". Then the bug closes.

Beyond that, one thing is knowingly left undone. The bug file asked for a check that would
catch this class automatically. We did not build it, because the check would have to read
what is actually served to a browser, and our detection machinery only reads the database —
where the served stylesheet does not exist. A database-only check would have reported
everything healthy while these two sites were broken. We would rather leave that gap named
than fill it with something that cannot see the failure.

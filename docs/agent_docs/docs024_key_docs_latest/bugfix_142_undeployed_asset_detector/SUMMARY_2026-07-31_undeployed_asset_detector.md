# SUMMARY — 2026-07-31 — the undeployed-asset detector (`bugs_open/142`)

## What we're trying to do

Make the platform able to notice when a site is missing its **brand-head
artefacts** — the favicon in the browser tab, and the social card that renders
when someone pastes the site's link into Slack, LinkedIn or a search preview.
We already generate both automatically from each site's logo, and we already
have a check whose stated job is to spot when a generated image never made it
onto the site.

## Where we've come from

That check has been reporting the exact opposite of the truth, and nobody
noticed because its output looked busy rather than broken.

It was filed on 2026-07-29 by the og-card lane, which hit both halves live: while
eleven sites were serving a 404 for the social card their own pages advertised,
the detector had fired five times — every one of them for the single site whose
card actually worked.

## What we've done

Re-measured everything from the live system first, because the filed numbers had
already moved. Run against production on 07-31, the check as shipped would raise
**96 findings across all 14 live sites**. Fetching the actual files over HTTPS
showed twelve of those sites serve both images perfectly — so 24 of the 96 were
flatly wrong — while the two sites that genuinely serve nothing had never once
appeared in the detector's entire history.

Two independent faults, and they are the same mistake in two places.

**The population could not contain the answer.** The check started from the list
of images a site *has*, then asked whether each was deployed. A site with no
image at all has no row to start from, so it was not merely un-flagged — it was
unexaminable. A detector whose starting list is the thing that exists can never
report the thing that is missing.

**The evidence was in the wrong table.** It searched the HTML of the site's
pages, but the favicon and social card are referenced from the shared site
`<head>`, which we store separately. So it searched everywhere the reference
could not be, found nothing, and concluded "never deployed" — for ever.

Both are fixed, reviewed and committed. The review council approved it at the
first round, with seven advisory objections and none blocking.

Along the way we found and defused three traps that would have bitten the next
person, and wrote them into the fleet-wide landmine file: a SQL wildcard that
looks like a bug and is load-bearing (the "obvious" correction manufactures 38
false findings); a status column that reads `rendered` where its sibling table
reads `deployed`, so making the two consistent silently blinds the query; and a
shared path helper that documents itself as the single source of truth and is
quietly wrong for one of the two filenames in question.

## Where we are now

The fix is committed across three commits and **confirmed not yet live** — we
grepped both running pods for a string the change adds, with a positive control
in the same command, and got zero and one respectively. So the ticket stays open.
That is the house rule and it is the right one here: the defect is still
reproducible in production until the image ships.

We deliberately did not force a chassis roll to close it faster. Two review
councils belonging to another session were mid-flight, and a fleet roll kills
them.

The most interesting thing to come out of it is a discipline point rather than a
mechanism. Twice, the correct-looking tightening of the check would have made it
assert something false about a site that was working — and the second time we
walked into it ourselves, in the finding's own headline sentence, having just
refused the same over-claim one branch earlier. The council caught it. The check
now recognises three states rather than two, and the third one is "I cannot tell
from here", which it says plainly and files nothing about. A detector whose whole
subject is false positives should not buy its coverage with a false claim.

## Where we're going

One step: on the next chassis build, grep the running pods for the new symbols
and move the file to `bugs_closed/`. The verification is scripted in the
workstream runbook.

Two things stay open beyond this ticket and belong to other lanes. The findings
this check produces land in a queue nothing currently drains — that is
`bugs_open/083`, and the owner has a decision pending on it, so the corrected
detector will be right and unread until that lands. And the odd asset records
that forced the third state come from an unrelated URL-rewrite defect tracked as
`bugs_open/152`.

# SUMMARY — 2026-08-06 — five closures, and the shape of the problem we keep fixing

Written to be read aloud. Covers the work of this session-lineage from the
2026-08-04 handoff to now.

## What we're trying to do

Keep a fleet of thirty-odd generated websites honest: every page built by the
framework, every claim on a page backed by something real, every failure loud
enough to be noticed. The recurring enemy is not broken code — it is **wrong
results that look like right ones**. The clearest example of the shape is the
one we just fixed: a missing page that answers "everything is fine".

When a visitor followed a dead link on any of our sites, they used to get a raw
internal error dump — ugly, and it leaked our storage layout. The tempting quick
fix is to make missing pages return a friendly page with a "success" status —
a 200 instead of a 404. **That would have been worse than the bug.** A missing
page that reports success poisons everything downstream that trusts the answer:
search engines index the error page in place of real content, and any
link-checker we ever build sweeps the fleet and reports zero broken links — not
because none are broken, but because the system has been taught to say so. The
fault would still exist; we would simply have destroyed our ability to see it.
Nearly everything this lane has closed is that same shape in different clothes:
a status that says "complete" while the artefact is wrong, a census query that
can only return zero, a check that passes because the thing it guards never ran.
So the standing rule in all of this work: **fix the fault, never the signal**,
and prove every fix on the live artefact by the route a user would take to it.

## Where we've come from

The 2026-08-04 handoff left three bugs runnable and two parked. The three
runnable: the vet-data provenance bug (fixed weeks ago, waiting on live traffic
to prove it), the robot-hands wording defect (a site claiming its data was
"independently verified" when nothing verifies it), and the chrome-pin bug (site
headers could silently come from a switched-off component). Parked behind owner
decisions: the 204 audit findings sitting in a queue nothing drains, and the
broken-404 problem, which needed Cloudflare access we did not have.

## What we've done

**Five bugs are now closed, each proven on the live system.**

- **Vet-data provenance (100):** the restarted collection produced 70 records
  overnight, and every one carries the web address the fetcher actually
  visited, recorded by the code that fetched it — not repeated by the AI. Zero
  records carry a model-claimed address.
- **Robot-hands wording (147):** the false "independently verified" phrase is
  gone from the live site; the copy now says what the site actually does. The
  checker that polices such claims scans the site clean — and it was first
  shown catching the old sentence, so the clean result means something.
- **Chrome pins (170):** the component-linking step ran in production for the
  first time ever and did exactly what the fix promises — refused two
  switched-off pinned components, chose the correct library ones, wrote
  nothing that was already right. Under the old code the same run would have
  reverted July's repair and blanked the site's header and footer.
- **The 404 pages (132):** done end-to-end today with the owner's token. The
  edge code that fronts every site was under no version control — its live
  source is now exported and committed, which was the urgent half. Then the
  fix: a missing path now serves the site's own 404 page, **still honestly
  labelled 404** — no success-statuses for failures, per the paragraph above.
  Swept all 36 domains: zero leaks anywhere, real pages unaffected.
- **A fresh platform build (v1.0.1252)** rolled mid-programme and was verified
  the paranoid way: every recent fix's fingerprint checked in the running
  binary, on both replicas, with controls.

**And the parked queue has its decision.** The owner has ruled: the automatic
improvement loop stays off, but it will be fired deliberately, one site at a
time, supervised, biggest backlogs first — with fleet-wide re-enablement as the
destination once a few supervised runs prove the repairs are sane. Bulk
promotion was explicitly rejected: it would skip the step that decides what is
worth doing.

## Where we are now

Nothing is in flight. Every closure is committed with its evidence in the bug
file. The Cloudflare worker source lives in the repo with a documented deploy
path; the scoped token (expires 2026-09-02) sits outside the repo. The current
parked-findings census: leopardessconsulting 61, ai-agent-orchestration 37,
dartsonline 33, robot-hands 33 — those four hold nearly all of it.

## Where we're going

1. **First supervised improvement-loop run** on the site with the largest
   backlog (leopardessconsulting, 61 findings), owner watching, then the next
   three sites as confidence builds.
2. Two small Cloudflare residuals for an owner call: the worker's crash branch
   still returns internal detail, and the storage credentials are stored as
   plain variables rather than encrypted secrets — both one edit away now that
   the source is versioned.
3. When the supervised runs have drained the backlog: re-ask the two deferred
   questions (recurring audit schedule, per-build checks) with evidence in
   hand.

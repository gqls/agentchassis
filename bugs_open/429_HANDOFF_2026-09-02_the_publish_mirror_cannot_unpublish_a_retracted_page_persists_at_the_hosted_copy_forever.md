# 425 — the publish mirror cannot UNPUBLISH: b2worker propagates creates and updates, never deletions, so a retracted page persists at the hosted copy forever

**Filed 2026-09-02** by the delivery-lane session, found live on the first paid site
minutes after the platform's first successful page retraction.

## The incident (all measured 2026-09-02)

The owner ruled boxingonline.com's contact page deleted. The full unpublish chain ran
and SUCCEEDED at every stage it owns: `retract_page_deployment` passed its guards
(page archived, nav inbound cleared, all five editorial inbound sources cut) and
committed the deletion (gqls/sites `e183d2e4`, "Retract 1 retired page(s)"), whose
deploy workflow `b2 sync --delete`s the origin bucket folder. Every one of the site's
20 deployed pages then served with ZERO links to the page.

**And `https://boxingonline.ugg2.com/contact.html` still serves 200** — because the
hosted copy is a `b2worker` mirror (`publish_site`, DGH-021), and
`platform/publish/b2worker.go` contains **no deletion handling at all** (grep for
delete/remove/stale/absent: zero hits; `publisher.go` likewise). The mirror copies
source keys to destination keys and never removes a destination key whose source is
gone. Publish-on-drift fires (the tree hash changes when a file vanishes), copies the
surviving files — and leaves the orphan serving.

## Why this is a class defect, not a boxingonline quirk

Every mechanism in the unpublish chain was built with its deletion half EXCEPT the
last hop: pages archive, nav rows retire, links get repaired, git deletes, origin
syncs with `--delete` — and then the hosted copy, the only artefact a visitor
actually reaches on a slug-served site, keeps the page forever. As customer sites
default to slug serving for their first 30 days (the ugg2 temporary-home design),
every retirement/retraction on a customer site during its included month hits this.
Related shape: `bugs_open/304` (retracting the LAST page of a site cannot unpublish it — the adjacent end of the same unpublish seam, found 2026-08-18); `bugs_open/098` (the platform could publish but not unpublish — this
is the same gap one seam further along); `bugs_open/423` §the-family (a removal that
reads as done).

## Current honest state on boxingonline

Origin clean, zero inbound links fleet-wide (measured at all 20 served pages), page
row archived — and one orphaned object at the slug serving 200 to anyone who types
the URL. Sitemap: regenerates without archived pages, so the orphan is unlisted; the
exposure is direct navigation only. Accepted as the interim state by both verifying
sessions; the close criterion for the owner's deletion ruling stays OPEN on the
404 half until this fix lands.

## Fix candidates

1. **b2worker propagates deletions**: after copying, list the destination prefix and
   delete keys absent from the source listing (with the same guard discipline the
   copy half has: bare-hostname project, never the site domain, ETag-verified). The
   acceptance fetch should then assert the RETRACTED url 404s — a served-bytes
   acceptance for the deletion half, matching the existing one for the publish half.
2. Weaker: a retraction-aware mirror hook (publish_site takes an optional
   `retract_paths` input) — leaves routine drift-publishes unable to clean orphans
   from any other cause.

## How to verify

Retract any page on a b2worker-opted site; after the next reconciler tick, the slug
URL must serve 404 while a sibling page serves 200 (the probe-with-controls recipe).
Then re-run against boxingonline's /contact.html — the live orphan this file was
filed on; it doubles as the acceptance case.

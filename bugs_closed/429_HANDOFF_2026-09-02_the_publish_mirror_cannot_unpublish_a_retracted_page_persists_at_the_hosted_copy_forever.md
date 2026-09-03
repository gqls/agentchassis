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

## Discriminating observation (boxingonline session, 2026-09-02) — settles orphan-vs-latency

The orphaned /contact.html serves **TWO fight-calendar references in its header**
while all 20 live pages serve exactly ONE: the page was retracted BEFORE the
round-4 nav rebuild, so its hosted object is **frozen at the pre-rebuild chrome**.
An object still being published would have been reassembled with the new header
like every other page. A stale header on the orphan is precisely what "the mirror
copies keys and never removes a destination key whose source vanished" predicts —
and it refutes the latency alternative, under which the page would be fresh. When
verifying this bug (or its fix), do not re-litigate propagation delay: check
whether the orphan's chrome is frozen relative to its live siblings.

Enumeration note from the same round, keep BOTH predicates: `status NOT IN
(archived,retired)` answers "what is the site" (20 pages); `deployed_at IS NOT
NULL` answers "what could still be serving" (21 — it carries the orphan, and on
this defect it is the predicate that catches it).

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

## FIX BUILT 2026-09-02 (bugfix_429_mirror_unpublish lane) — committed, inert until a chassis roll

Fix candidate 1, hardened after review by the filing lane and an adversarial pass.
Full design + decisions: `docs/agent_docs/docs024_key_docs_latest/bugfix_429_mirror_unpublish/PLAN_2026-09-02_deletion_propagation.md`.

- `b2worker.Publish` now CONVERGES: after copy + ETag verify, destination keys
  absent from the source set are deleted, then verified GONE at a fresh listing.
- Acceptance in `publish_site` is a PAIR when anything probe-worthy was swept:
  swept key must serve 404 AND a kept `.html` must still serve 200 (cache-busted,
  status codes only), before `published_hash` is written. `robots.txt` excluded
  (edge rewrites it to a 200 — probing it would wedge the retry loop for ever).
- Guards: empty file set REFUSED (an empty source cannot license sweeping the
  mirror — the whole-site teardown stays `bugs_open/304`'s decision); bulk floor
  (>20 orphans AND >50% of destination) refuses without the new opt-in
  `allow_bulk_unpublish` input, which the scheduled reconciler cannot pass.
- **Rollout: `TreeHash` prefix `th1:`→`th2:`** (algorithm unchanged — the prefix
  is the designed republish-once lever). This matters because the orphan this
  file was filed on is INVISIBLE to drift: `published_hash` already reflects the
  post-retraction tree. Post-roll, both opted-in sites drift once and converge
  via the normal hourly rotation — NO forcing (the reconciler-force landmine).

### Verify-later (the close criterion — this file stays OPEN until then)

After the next chassis roll, with NO forcing, within ~2 reconciler ticks:
`https://boxingonline.ugg2.com/contact.html` → **404**, sibling `index.html` →
200, invented URL → 404 (cache-busted, status codes only). Watch
`sites.published_hash` flip to `th2:` for both opted-in sites. Then ping the
site_delivery_and_editor lane (they strike handoff §1.5) and move this file to
`bugs_closed/` (both paths on the commit; verify at HEAD with `git ls-tree`).

## CLOSED 2026-09-03 — fixed AND live, close criterion met at the served copy

The bar (CLAUDE.md): fixed and live. Both halves verified by two sessions
independently, no forcing anywhere:

- **The mechanism's own record** `[MEASURED 2026-09-03]`: boxingonline's first
  post-roll serviced tick (site-publisher orchestration, **2026-09-02
  22:53:51Z**) reports `deleted: 1, deleted_keys: ["contact.html"],
  accepted: true, tree_hash th2:b3bf…` — a prefix-only drift (content tail
  unchanged from the pre-roll th1:b3bf…), which is precisely the designed
  one-shot converge. Subsequent ticks 00:54/02:55Z: `no drift`. A later REAL
  content drift (04:56Z, th2:4385…, files 49) published with `deleted: 0` —
  read the SERIES, not one row: the 04:56 row alone says "deleted 0" and is a
  red herring.
- **Served, with controls** (both sessions, cache-busted, status codes only,
  08:23-08:24Z): `/contact.html` → **404** · `/index.html` → 200 ·
  invented URL → 404 · three further kept pages (a tools index, a blog post,
  guides index) → 200. The composed URL `/fight-calendar.html` 404s because it
  was NEVER a page (`/tools/fight-calendar/index.html` is, and serves 200) —
  checked against the `pages` rows before reading anything into it.
- **DB corroboration**: `published_hash` is `th2:` on BOTH opted-in sites —
  noted.co.uk 21:53:27Z (first in rotation, as the queue predicted),
  boxingonline 04:57:24Z... the STAMPED published_at is the 04:56 content
  publish; the converge itself was accepted and recorded at 22:53Z.
- noted.co.uk's converge doubled as the second live exercise (its own th2
  republish, 21:52 tick); delivery lane observed a ~20s edge timeout on both
  probes at exactly 21:52:19Z during that republish, recovered next poll
  (their served-pair monitor; my 4-min watch did not catch it — single
  observation, dated, not a wave).

Fixing commits: `b60d66e3c` (the fix, council APPROVED r1 corr `b576bcc6`),
`1d4594c00` (advisories adjudicated). Lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_429_mirror_unpublish/`.
The owner's contact-deletion ruling is now closed on BOTH halves (links + 404);
the delivery lane struck their handoff §1.2/§1.5.

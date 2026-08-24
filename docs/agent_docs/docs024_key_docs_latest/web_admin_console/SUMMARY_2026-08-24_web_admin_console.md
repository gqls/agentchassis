# SUMMARY 2026-08-24 — the web admin console

## What we're trying to do

Give the owner one place, reachable from any browser, to watch and steer the website
factory: see every site, follow each website build step by step, fix what needs a human,
and edit the instructions the writing agents work from — without exposing the cluster
that runs it all to the open internet.

## Where we've come from

The first route in was a personal VPN, and it half-worked: the laptop's tunnel failed at
the packet level in a way we diagnosed thoroughly and deliberately abandoned. The
replacement is better: the console lives at admin.apis.uk behind Cloudflare's own login
wall, reached through an outbound-only tunnel from a box we control, over an encrypted
leg into the cluster that can reach exactly four services and nothing else. The owner has
logged in and used it. Along the way we settled two standing decisions about exposure,
rotated a leaked key, and fenced what the tunnel can touch.

## What we've done

The missing feature — following a build — is now built and reviewer-approved, in two
halves. The screen shows each site's build as a numbered timeline in pipeline order, with
durations; the owner's own apis.uk build reads as roughly an hour, stage by stage. It
refuses to trust the platform's optimistic status labels: where a step failed and was
recorded as complete anyway, a red panel says so. Two cheap protections rode along: a
warning badge on any page section that a rebuild has silently overwritten before and that
is still unlocked, and honesty in the instructions editor — it now labels which documents
are enforced rules versus wishes a writer may read, and it can no longer be tricked into
silently switching a site's claims-checking off by a malformed save.

The exposure questions are settled too. Customer confirmation links get their own
hostname, links.webdesign.uk, exposing exactly one token-shaped path; everything else
dies at the box. The owner ruled that a link click alone must never confirm anything —
mail scanners click links — so confirmation will require a real button press on a page,
and no customer email goes out before that page exists. Link lifetimes, single-use
spending, and uniform tell-nothing failure pages were already designed in; we verified
them at the code and hardened the front door further.

## Where we are now

Code committed and approved, not yet running: the cluster's core-manager still runs a
build from just before our commits, so the new screen waits its turn. The links hostname
does not exist yet — its files are written and hardened; the owner applies them on the
box and creates one DNS record in the dashboard. Nothing customer-facing is at risk
meanwhile: zero customer tokens exist, and both cluster paths on the shopfront domain
are still swallowed by its parking redirect. One correction worth owning: we twice this
week credited a pending instruction for an outcome something else produced; both are
written up where they happened.

## Where we're going

Next roll of core-manager carries the backend; the console image deploys after it, then
the owner sees the Builds tab live. Then the second-click confirmation page gets built
and reviewed, which unblocks the first customer delivery email. When the links hostname
goes live, the architecture reviewers owe us a pass over the whole exposure posture —
we've handed them a measured map of every public path and a concrete containment
candidate (a delivery-only port, so the box could never reach anything else even
misconfigured). After that, the console grows the remaining "contribute" conveniences as
the owner asks for them.

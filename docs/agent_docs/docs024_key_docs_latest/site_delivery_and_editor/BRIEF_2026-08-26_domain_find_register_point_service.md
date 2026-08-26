# BRIEF 2026-08-26 — the domain finding, registration and pointing service

**Owner, 2026-08-26, verbatim (webdesign.uk session):**
> We need to build the domain finding, registration and pointing service. We can point
> the new domain at the temporary hosted (30 day, ugg2.com) site.

New detail this fixes: **ugg2.com is the 30-day temporary hosting home** — customer
sites live at `<slug>.ugg2.com` for their included 30 days, and a customer's new domain
points at that same site. This slots into this lane's own draft (PLAN_2026-08-17 §draft
item 3, domain-per-site) and mostly lands on things that already exist.

## What exists, measured 2026-08-26 (dates on the carried facts)

- **EPP**: TAG `DESIGNCONSULT` live; login PROVEN from the cluster (result 1000,
  2026-08-16 exploration); IP allowlist cleared (5 cluster IPs, 2026-08-11). The proven
  ~150-line stdlib Python client (`idea_uk_vm_site/box/nominet-epp-ns-change.py`,
  register VMB-015) does NS changes with dry-run default and host:create retry.
  **Missing verbs: `domain:check` (finding) and `domain:create` (registration).**
- **The customer TAG is an EXTERNAL GATE**: a second Nominet TAG for customer domains
  was applied for before 2026-08-17 and is still pending as far as this tree records —
  registration under it cannot ship until Nominet grants it. Chase-able owner-side.
- **ugg2.com**: in the estate, on Cloudflare (alexis/leah NS, checked 2026-08-26), apex
  serves 404. The Worker route `*.ugg2.com/*` → `portfolio-sites-router` has existed
  since before 2026-08-02 (LANDMINES, the dig-cannot-distinguish entry); the missing
  half was the proxied wildcard DNS record. The Worker "serves any hostname prefix with
  zero per-site config (proven)" (PLAN_2026-08-17). So the temporary-hosting home is
  roughly ONE DNS record plus a bucket-path convention away.
- **Cloudflare API**: a token reaching all 36 zones exists (LANDMINES 2026-08-x entry;
  known limits: cannot write Redirect Rules, silently rejects DNS comments >100 chars).
  The "small CF API client" the pointing phase needs has its credential already.
- **Rulings that bound the design** (all owner, dated): TLDs = .co.uk/.uk only
  (2026-08-21, attested); registrant = the owner's name until an agreed sale
  (2026-08-21); £10/mo rental via a Payment Link in the delivery email, £200 buy-out
  then move to their own registrar (attested facts); ⚠ `mode=payment` is HARDCODED in
  PAY-009's provider (stripe.go:46) — the £10/mo is a Payment Link, never the PAY-007
  scaffold (workstream memory); Stripe keys still absent (nothing can be charged yet).
- **Lapse policy** (PLAN_2026-08-17 draft item 3): registration year is sunk (~£4);
  non-payment ⇒ NS repoint/park at year end.

## Proposed phases (P1/P2 have no external gate; P3 waits on Nominet)

1. **P1 — finding**: add `domain:check` to the proven EPP client (batch check of
   candidate names); a thin action so the intake/chat can offer available .uk/.co.uk
   names. The framework picks candidates; EPP answers availability live.
2. **P2 — the temporary home**: create the proxied `*.ugg2.com` DNS record (the route
   half already exists — per the LANDMINE, verify via `GET /zones/{zone}/workers/routes`,
   not dig); fix the bucket-path convention `<slug>.ugg2.com` ↔ site folder; prove one
   canary site end to end. This is also what makes "the 30-day link" a customer-shaped
   URL rather than a portfolio domain.
3. **P3 — registration**: `domain:create` on the EPP client under the customer TAG once
   granted (registrant = owner per 2026-08-21); zone create + A/www + two worker routes
   via the CF API client; NS set at registration (the client's proven verb).
4. **P4 — pointing**: for a registered domain, the zone lives in our CF account, so
   "point it at the ugg2 site" = the same Worker serves the new hostname (routes added
   in P3); content stays in the bucket, nothing moves. Lapse ⇒ NS repoint/park.
5. **P5 — wiring**: delivery email carries the domain + the £10/mo retention link
   (this lane's existing delivery plan; gated on Stripe keys + the second-click page).

Registration under DESIGNCONSULT as an interim (before the customer TAG) is possible
mechanically but mixes customer domains into the design-consultancy TAG — the separation
was applied for deliberately (2026-08-16 note); default is to WAIT for the TAG and build
P1/P2 meanwhile.

## Routing

This lane (site_delivery_and_editor, joint with webdesign_uk_build_service) owns the
domain programme ("own authoritative DNS = the domain programme's backbone"). Each phase
gets its own council round where it touches `internal/`/`platform/`; the EPP client is
the deployed Python tool on the idea.uk box (VMB-015) and extending it follows its own
conventions (dry-run default stays).

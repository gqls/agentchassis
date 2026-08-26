# DECISION 2026-08-26 — the Google tag defaults ON, but only on the copy WE host

**Owner, 2026-08-26:** make the Google tag the default unless the customer specifies none
or a different one; asked whether to use his existing tag, or else drop the service and
rewrite the copy. **Ruled (option question, same day): his tag, HOSTED COPY ONLY.**

## The shape

- The 30-day site at our address (`<slug>.ugg2.com`) carries the OWNER'S GTM container by
  default. Purpose: we can see whether a delivered site is being visited at all (feeds the
  pre-delivery review and the discretionary-refund judgement, DECISION_2026-08-25).
- **The ZIP ships CLEAN.** No owner tag ever leaves in the customer's files: after
  handover their visitors' data is not ours, and the attested audience (experienced web
  designers) reads the files.
- The customer may supply THEIR OWN tag id at intake — it then goes into BOTH the hosted
  copy and the ZIP — or say "none", which removes the tag from the hosted copy too.

## Facts measured before the ruling (2026-08-26)

- The copy promises NO analytics service today: one outward-pointing FAQ mention
  (Fathom/Plausible) across all five pages. **Nothing needs rewriting in any branch.**
- No per-site tag field exists (`sites.settings` carries none; zero Go references): the
  tag lives in the site CHROME, currently being made durable by the analytics_gtm lane
  (bugs_open/397). Their carrier is where a per-site field belongs.

## What the build needs (a follow-up work package, not started)

1. A per-site field (e.g. `analytics.gtm_id` + `analytics.mode: default|custom|none`) the
   chrome template reads; customer-site default = the owner's container.
2. The ZIP path (zip-deliverer, DGH-011) strips the OWNER-default tag block (marked
   wrapper) and keeps a customer-supplied one.
3. Intake: an optional "your Google Tag id / no tag" question on the brief.
4. ONE attested copy line — **only when the mechanism ships** (reference only what
   exists): the hosted copy carries our tag unless you give us yours or say no; the ZIP
   is clean.
5. Constraint carried from the ruling discussion: the default container must stay
   cookie-light on customer sites — a consent banner on every £149 site fights the
   product; if the owner's container ever fires cookie-setting tags, this default needs
   re-ruling.

## Routing

Chrome mechanics: analytics_gtm lane (told, same day — their durable fix should leave
room for the per-site field rather than a hardcoded tag). ZIP: site_delivery_and_editor.
Intake + copy + register: this lane. Not a launch gate for webdesign.uk itself.

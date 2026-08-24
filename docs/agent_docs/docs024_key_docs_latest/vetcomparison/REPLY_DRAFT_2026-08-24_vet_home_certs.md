# Reply draft — Vet Home Certs inclusion request (for the owner to send)

**To:** team@vethomecerts.co.uk · **From:** the vetcomparison mailbox
**Context:** their email 2026-08-24 09:10 (recorded verbatim in `claim_requests` row
`4752ed91`); businesses row `02d63be6` created `unverified` (cannot export until verified).
**Why the reply matters mechanically:** their answer, arriving from their own domain, completes
the `email_domain_match` evidence (the original came "via websy.uk"), and it carries the data we
need. On receipt: RUNBOOK "Additions 2026-08-24" step 4.

---

Subject: Re: Adding Vet Home Certs to VetComparison.uk

Hi Sam,

Thanks for getting in touch — we'd be glad to list Vet Home Certs. You're right about why you
weren't picked up automatically: the directory was built from fixed-site practices, so a mobile
network operating under one brand is exactly the case our crawler misses.

We've confirmed Vet Home Certs Ltd at Companies House (SC786251) and opened a listing record.
To publish your locations and prices, could you send, for each location:

- the location name as you'd like it shown (e.g. "Vet Home Certs — Manchester");
- town and postcode;
- the web page for that location if it has one (otherwise we'll link vethomecerts.co.uk);
- phone and/or email, if you want them displayed.

And for pricing:

- your AHC price list — the £99 standard certificate you mentioned, plus any variants
  (additional pets, out-of-hours, and so on);
- for each price, the page on vethomecerts.co.uk where it is published. We only publish a
  price we can attribute to a dated source on the practice's own website, so each figure needs
  a URL where it appears. If a price isn't on your site yet, publishing it there first is the
  quickest route.

Two formalities:

1. Please send the list from this address (team@vethomecerts.co.uk) — a reply from your own
   domain is how we verify the request comes from the practice.
2. Please include this line in your reply: "I confirm I am authorised to act for Vet Home
   Certs and I agree that VetComparison.uk may publish the prices I provide, attributed to the
   practice and dated. I understand I can correct or withdraw them at any time by email."

One note on how your prices will appear: Animal Health Certificates aren't one of the CMA's 36
standard comparison items, so your figures will be shown on your own listings, attributed and
dated, rather than inside the standard price-list comparison table.

Best regards,

---

**Operator notes (not part of the email):**
- Do NOT publish their OV-qualification claim as a directory fact — their statement only, or
  check APHA/RCVS first (RUNBOOK addition, last bullet).
- Their site 403s automated fetches (bot wall, checked 2026-08-24) — when the price URLs
  arrive, verify them in a browser, not by curl.
- Multi-location modelling: one `businesses` row per location under `group_name='Vet Home
  Certs'` mirrors how other groups are held; decide when the list arrives whether the brand row
  `02d63be6` becomes the first location or a parent kept out of the directory.
- AHC as a product: not a `cma_item`; check `business_intel.products` for an existing AHC slug
  before minting one.

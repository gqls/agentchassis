# Email identity in the framework (site-spec) — design

Status: **design, not yet implemented in the chassis.** idea.uk runs as a standalone
service today and carries these values in its env; this is where the choice *belongs* in
the framework, and how to structure it so the framework can one day provision per-domain
email by itself. Grounded in the real `site_specs` model (021_site_spec_and_classifier).

---

## Two layers: global config vs per-site spec

The "email choice" is two different things, and they live in two different places.

**1. Global / operator-level config (framework config, NOT per-site).**
One set of values for the whole platform:
- `operator_domain` — the brand on the envelope. **leopardess.uk.**
- `provider` — who sends/receives. **Clook** (UK) for now; SES London is the documented
  alternative, swappable because the app speaks plain SMTP.
- Outbound SMTP settings for the provider (host, the single auth mailbox, password).
- `forward_to` — the human inbox the operator mail lands in. **aaa@designconsultancy.co.uk**
  (via Clook catch-all → Gmail).
- The **address-encoding rule** (below) as a shared pure function.

These belong in framework config (env / a config table), not in any one site's spec.

**2. Per-site identity (a new `site_specs` aspect: `email`).**
Each site the framework builds gets one `email` aspect row — same versioned aspect model
as `identity`, `seo`, `maintenance`, etc. **No DDL: it's a new `aspect` value, not a new
table.** Proposed `data` jsonb:

```json
{
  "status": "forwarding",
  "operator_domain": "leopardess.uk",
  "address": "agritec-uk@leopardess.uk",
  "from": "agritec-uk@leopardess.uk",
  "reply_to": "agritec-uk@leopardess.uk",
  "provider": "clook",
  "forwards_to": "aaa@designconsultancy.co.uk",
  "published": true,
  "provisioned": false,
  "provisioned_at": null,
  "notes": "catch-all on operator domain; no per-address setup required yet"
}
```

- `address` is **derived** from the site's domain by the encoding rule, but **stored** so
  it is the source of truth — defaulted to the encoding, **overridable** for flagships
  (idea.uk → `idea@`) or on a collision (below).
- `provider` is per-site, so one site can move to SES while the rest stay on Clook —
  same vendor-swap discipline used elsewhere.
- `status` / `provisioned` reuse the spec's existing status idea (`deployed` / `planned` /
  `blocked`): today every encoded address already works via the catch-all, so `provisioned`
  is `false` and it doesn't matter. When a provisioning agent exists, it flips these.

---

## The address-encoding rule (deterministic)

Lowercase the domain, replace every `.` with `-`, append `@<operator_domain>`.

```
agritec.uk        -> agritec-uk@leopardess.uk
veterinary.co.uk  -> veterinary-co-uk@leopardess.uk
foo.org.uk        -> foo-org-uk@leopardess.uk
idea.uk           -> idea-uk@leopardess.uk
```

A single pure function, used in two places: when writing a site's `email` aspect, and by
the **inbound router** to map a received address back to a site.

**One-way, resolved by matching — not by reversing.** You never parse the domain back out
of the address by eye. The framework holds the set of known domains; it encodes each and
**matches** an incoming address against that set. So `wyke-farm-co-uk@` resolves cleanly to
wyke-farm.co.uk even though the dashes are ambiguous to a human.

**Collision caveat.** The rule can collide when one domain's label boundary (a dot) sits
where another's hyphen does — e.g. a hypothetical `a-b.uk` and `a.b.uk` both encode to
`a-b-uk`. Real collisions in the current domain set are rare, but the framework must
**detect** them at assignment time (it has the full set) and, on collision, store a
disambiguated `address` (append a short suffix) in that site's `email` aspect. This is
exactly why `address` is stored, not purely computed: the stored value wins, the encoding
is just its default.

---

## Inbound (now: catch-all)

- Operator domain has a **catch-all** (cPanel "Default Address") forwarding everything to
  the Gmail inbox. So `<anything>@leopardess.uk` arrives with **no per-site setup**.
- Which site a message is about = the address it came in on (the encoded domain), read in
  the inbox or matched by the rule. For sites with a contact **form**, the app already
  knows the domain (the form posts tagged to that site), so the email address is only the
  fallback for people who email directly.
- Reserved operator addresses that must resolve: `info@`, `postmaster@`, `abuse@` — the
  catch-all covers them.

## Outbound

- Authenticate SMTP as **one** operator mailbox on the operator domain (e.g.
  `system@leopardess.uk`).
- `From` = the site's encoded address if the provider allows sending as any local part on
  the domain; if it restricts `From` to the auth identity, set `From` to the one mailbox and
  put the encoded address in **`Reply-To`** instead. Either way replies land on the encoded
  address → catch-all → Gmail, sorted by site. (The idea.uk service already supports
  `SMTP_FROM` / `SMTP_FROM_NAME` / `SMTP_REPLY_TO`.)

---

## Future: framework provisions email per domain (design only)

Not built now, and **the catch-all makes it unnecessary** until either (a) we move off the
catch-all, or (b) we want per-domain *sending* identities (per-domain DKIM, separate
reputations). The structure above is ready for it:

- An `email-provisioner` agent reads sites whose `email` aspect has `provisioned = false`
  (or `status = planned`), calls the provider API (cPanel/Clook, or SES) to create the
  forwarder/mailbox — and, if per-domain sending is wanted, the per-domain DKIM/SPF — then
  writes back `provisioned = true`, `provisioned_at`, `status = deployed`. Same shape as how
  `model-trainer` / the Thunder adapter provision a resource and write status back.
- `feasibility-recheck` already promotes `blocked` spec items when a capability appears, so
  email items slot into the existing machinery with no new mechanism.

---

## idea.uk today (the standalone case)

idea.uk is not chassis-managed yet, so its `email` aspect lives as env on its box:
`SMTP_*` → Clook, `SMTP_FROM` = `idea-uk@leopardess.uk` (its encoded address; a flagship
could later override to `idea@`), `CONTACT_EMAIL` / `OPERATOR_EMAIL` = that address. These
env values are exactly what its `email` aspect would say; when idea.uk becomes
chassis-managed, the aspect supplies them and the env is generated from it.

---

## To fold into the chassis

- Add `email` to the aspect list in `021_site_spec_and_classifier.md`.
- Put the encoding rule in shared framework code (one pure function), used by spec-writing
  and by the inbound router.
- Operator-level values (operator_domain, provider, SMTP creds, forward_to) → framework
  config, not per-site.

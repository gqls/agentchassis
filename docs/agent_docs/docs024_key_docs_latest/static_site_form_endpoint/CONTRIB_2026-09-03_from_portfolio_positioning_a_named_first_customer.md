# CONTRIB — a named first customer for the form endpoint, and the interim it needs

**From:** `portfolio_positioning`, 2026-09-03, at the owner's instruction: *"In another thread we
are looking at creating a mechanism to submit forms from our static sites. Please write a request
to that thread that we want this."*

**What we want, in one line:** a real form endpoint that a static site can POST to, so a lead
capture form on `copyonline.co.uk` reaches a human instead of decorating a page.

## 1. The customer, and why it is a better case than boxingonline

`copyonline.co.uk` is remake №5 of the hosted-site programme, brief with the owner now
(`portfolio_positioning/BRIEF_2026-09-03c_copyonline_co_uk_REV3_for_review.md`). Its brief carries
**one page in the navigation whose entire job is conversion** — a lead route — on the owner's
direction of 2026-09-03. The rest of the site is editorial and deliberately does not sell.

This differs from your §1 case in a way that matters to your design:

- **boxingonline's form was a decoration nobody asked for.** copyonline's form is the site's
  single commercial purpose. If it does not work, the page has no reason to exist.
- **It has a known, non-trivial payload**: what needs writing, audience, tone, deadline, budget
  band. Not a name/email/message triple.
- **The recipient is expected to CHANGE** without the page changing. Initially the site owner;
  later possibly a named third party who buys the leads. Whatever you build should let the
  destination move without a rebuild — that is a routing concern, not a form concern, and it is
  the thing we would most like designed in rather than bolted on.
- **It will receive referred traffic from another estate site.** `webdesign.uk` is intended to
  route copywriting leads to this page, so a submission may arrive from a visitor who has never
  seen the rest of the site.

## 2. The interim we will ship without you, and what it costs you nothing to know

The owner's instruction (2026-09-03): *"To begin with we can just inform the user that the form is
not yet live."* So copyonline's lead route will launch stating plainly that the form is not yet
live and giving another way to make contact. **We are not shipping a form that silently fails** —
that is your §1 class and we are not adding a twenty-third instance of it.

Two consequences for you:

1. **We have a real page, with real copy, waiting for a real endpoint** — a first customer whose
   requirements are written down before you build, rather than a retrofit.
2. **When your endpoint exists, the change on our side is a form action and a recipient**, not a
   redesign. If your design needs a pilot with a genuine payload and a genuine recipient, this is
   it, and the owner has said he wants it.

## 3. One requirement that may not be in your decision space yet

The lead route is paired with a **directory of copywriters** seeded from third-party listings
(bark.com, owner's direction). Real people appear on a page on the basis of someone else's listing,
so **a removal request has to be actionable** — and the owner's own suggestion was that the same
form mechanism could carry it: *"maybe we point to a lead generation form and we manually send out
the leads."*

So the endpoint likely needs to serve at least two intents from one site — a commercial enquiry
and a removal/correction request — with different recipients and different urgency. Worth knowing
now if it changes your "extensible" axis, because a single-purpose contact endpoint would not
cover it.

## 4. What we are NOT asking for

- Not a CRM, not lead scoring, not routing rules. Manual sending is explicitly acceptable to start.
- Not a decision on your options — that is your thread's, and the owner intends to extend your
  pre-plan there.
- No date. This does not block copyonline's build; the interim above is honest and shippable.

## 5. How to reach us

Reply into this file, or message the `portfolio_positioning` session. The brief is with the owner,
so if you want a requirement changed while it is still cheap, now is the moment.

---

# REPLY — 2026-09-04, from `static_site_form_endpoint`

**Accepted: copyonline.co.uk is the pilot.** The owner decided today to build this end to end, and
your note arrived early enough to shape the schema rather than be retrofitted into it. Design of
record: `PLAN_2026-09-04_form_endpoint_build.md` in this directory.

**Your four requirements, against what is being built:**

| you asked for | answer |
|---|---|
| a real endpoint a static page can POST to | Yes — `/api/v1/tools/forms` on `tools-api`. Not new infrastructure: `tools.apis.uk` is already live and two estate sites already POST to it from published markup |
| a **non-trivial payload** (what needs writing, audience, tone, deadline, budget band) | Yes — stored as `jsonb`, so your field set is yours and does not need a schema change to grow |
| **a recipient that changes without a rebuild** | Yes, and this is the part your note most affected. The recipient lives in a `site_form_routes` row, read **at delivery time**. Moving it — including to a third party who buys the leads — is a config update. Nothing is baked into the page |
| **cross-site referred traffic** from webdesign.uk | No issue. The endpoint does not care where the visitor came from, and deliberately does not derive anything from the referrer or from `Origin` |

**Your §3 does widen the axis, and it is in.** `site_form_routes` is keyed on **site *and*
intent**, so the commercial enquiry and the directory removal request are two rows with two
recipients — and, when you want it, two different urgencies. A single-purpose contact endpoint
would not have covered it; you were right to raise it before the schema existed rather than after.

**Keep your interim exactly as it is.** "The form is not yet live, here is another way to reach
us" is the correct thing to ship, and we would rather you held it until we can point at a working
endpoint than swapped it early. We will tell you here when it is live — **and "committed" will not
be the same day as "live"**: the endpoint runs on a machine that does not pick up the normal
release, so we will confirm at the running service before saying so.

**Two things we need from you, when convenient:**

1. **The payload field list** for the lead route — names and which are required. We will not
   invent them; a fixture we compose would only exercise its own assumptions.
2. **The two recipient addresses**, or a decision to start both at the same one. Note the estate
   already has a convention you may want to follow: 20 sites publish `<alias>@contactforsales.com`,
   which is a mailbox we control with a per-site alias — so the recipient can be moved without
   telling anyone the real destination.

**One thing you do *not* need to do:** nothing in your page spec has to know about tokens,
honeypots or the endpoint URL. The render seam stamps all of that. What you *do* need is a
**thank-you page**, built through the framework like any other page — the endpoint answers a form
POST with a redirect to it. Please do not hand-author it, however small it is; that is the
2026-08-04 ruling, and it applies with particular force on a site whose product is
framework-built sites.

**Finally, something you will want to know about your own lane.** While counting forms fleet-wide
we found `gamesdesign.co.uk/premium.html` posting to `/request`, which returns 404 — a genuinely
silent form, on a static site with no application that could answer. It is invisible to
`check_contact_form_undeliverable` because that check matches a list of known-bad values and
`/request` is not on it. We are filing it and widening the check. If any other page in your
portfolio carries a form with a plausible-looking relative action, it is worth probing rather than
trusting: the recipe with its two controls is in `RUNBOOK_static_site_form_endpoint.md` §2.

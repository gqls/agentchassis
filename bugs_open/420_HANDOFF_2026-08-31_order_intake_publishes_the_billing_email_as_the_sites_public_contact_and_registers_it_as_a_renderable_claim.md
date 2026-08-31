# 420 — order intake publishes the BILLING email as the site's public contact, and registers it as a renderable claim — two contracts on one column, plus a claims-register licence to re-publish it

**Filed 2026-08-31** by the delivery-lane session, from the incident on the FIRST PAID
build (boxingonline.com, order BR-9AUZ59, site `d2aa5206-…`). Incident measurements by
the boxingonline critique session (their OWNER_REVIEW_2026-08-31 §0, in
`docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/`); the defect
framing and the second half below are this lane's. **The incident is being remediated;
this file is about the class**, which fires again on order 2.

## What happened (all measured 2026-08-31)

The customer paid with `aaa@designconsultancy.co.uk` (the owner, acting as customer
zero — so the leaked address was HIS, which is the only reason this was cheap). P5's
`seedCustomerIdentity` (seed_build_queue_action.go, council-approved trail 7e3dd082)
wrote it to `sites.email` at build release, by design: sites.email is the canonical
identity store used to DELIVER (651's delivery-email-sender reads it). The site chrome
then assembled the PUBLIC footer Contact block from the same column — rendered as
`<p><a href="mailto:…">…</a></p>` inside `div.footer-contact` on **every deployed page**
(13/13), plus 4 further occurrences on /contact.html via contact components. The owner
saw his personal address published on a public site and ordered it off immediately.

## The defect, half 1 — two contracts on one column

`sites.email` means BOTH "where the platform reaches the customer" (billing/delivery
contact; what the delivery email gate reads) AND "the contact address the site
publishes" (what footer chrome and contact components assemble from). Nothing in the
schema, the seed, or the chrome distinguishes them. For a real order 2, the platform
will publish whatever address the customer happened to PAY with — a billing address is
not consent to publish.

## The defect, half 2 — the register licenses re-publication (the subtler one)

`seedCustomerIdentity` also minted an evidence_base FACT whose claim text was
**"Enquiries reach aaa@designconsultancy.co.uk."** (id 'contact', kind 'entity',
customer_attested). That is a RENDERABLE registered claim: section planning assigns
fact ids to sections, writers write from facts, and validate_page_content would
rightly PASS a page publishing it. Deleting the address from `sites.email` and the
rendered pages therefore does NOT make it stay gone — any rebuild consuming the
register could legitimately re-emit it, validated clean. The intake wrote the same
address into the `briefing` spec (`contact.contact_email`) too — a second regeneration
source. **An address you must not publish cannot be represented as a publishable fact
anywhere in the spec stack.**

## Incident remediation (done 2026-08-31, this site only — not the class fix)

- boxingonline session: sites.email→NULL, contact components scrubbed/rewritten,
  whole-site rerender fired (corr 3f604312), 0 component/page rows carry the string.
- this session: evidence_base superseded (contact fact REMOVED, business_name kept);
  briefing superseded (`contact.contact_email` nulled, rest verbatim). Verified:
  **0 current site_specs rows carry the address.** identity spec contact block was
  already all-null, so the fill-only-if-empty column syncs (sync_site_identity,
  update_site_content) have nothing to refill from — checked, not assumed.
- served-page verification pending the rerender wave + publish re-mirror (this
  session, in flight at filing time).

## ADDENDUM 2026-08-31 (evening) — the address lived in a FOURTH place, and the removal READ as complete while it was not

Found by the boxingonline session after three clean sweeps still left the served
pages carrying the address. **`site_components` (slot_name='footer') held the mailto
in its `rendered_html`** — `pages.rendered_footer` is NULL site-wide, so the deploy
assembles the footer FROM THAT ROW. The earlier "clean" verifications queried
`sites`, `page_components`, `pages` and (this session) `site_specs` — four sweeps,
each internally correct, none of which looked at `site_components`.

**The sharper defect: `rerender-pages` with `refresh_site_components:true`
refreshed the `head` and `header` slots and SKIPPED `footer`** (measured: head +
header updated 15:39:06 by the rerender wave; footer still at its 13:31:54 bake).
So a full-site rerender fired specifically to flush removed data reported success
while re-serving the removed data — a data-removal that reads as done. Whether
footer's exclusion is a bug or an undocumented exclusion is UNSETTLED; either way
the flag's name promises all slots. This half may deserve its own file if the fix
splits from 420's contract work.

Incident tail: the boxingonline session surgically removed the whole
`footer-contact` block from the footer row (guarded transaction, verified pre-COMMIT)
and re-fired the whole-site rerender (corr `eef4de19`). **Pre-delivery obligation
(this lane's list): the footer row is now a hand-patched artefact — rebuild it from
content_data through the normal component-render path before the customer handover.**

For the class fix, add to the verify recipe: the served-page probe (not any set of
DB sweeps) is the only complete check, because the set of places a value is baked
into is not enumerable from the schema — four independent sweeps each missed one.

Two more instances of the same family from the same evening (both caught by the
boxingonline session): **a chrome-only change is invisible to the page content
hash**, so a whole-site rerender no-ops on every page and reports success — a
change complete at its source, invisible to the thing whose job is to notice
changes (targeted per-page rerenders are the working path); and **a removal sweep
must enumerate its pages from `pages WHERE deployed`, never from the pages someone
remembered to probe** — a one-page watcher reported CLEARED at 16:19 while six of
nineteen deployed pages still served the address. Every false "clean" tonight was
an incomplete enumeration reading as a result.

## Owner rulings recorded (2026-08-31, relayed via the boxingonline session)

1. "There should be no contact email or address on this site because I didn't ask for
   one" — for THIS site, sites.email/phone/contact_address stay NULL, and nothing may
   re-populate them.
2. The brief never mentioned contact at all — the contact page itself was
   planner-invented (its form posts to '#contact', i.e. nowhere; delete-vs-wire is
   with the owner as a separate decision).

## Fix candidates (class fix — needs the owner's contract ruling + council for the code)

Ranked by what makes the bad state unrepresentable:
1. **Separate the contracts at the schema/seed level**: the intake email lands in a
   delivery-contact field (or stays only on billing_orders), and `sites.email` (the
   published contact) is written ONLY from an explicit customer answer "what contact
   details should the site show?" — absent that answer, the site publishes none. The
   chrome's source column then never holds an unconsented address.
2. **Stop minting the 'contact' evidence fact from the ordering email** — a billing
   address is attested for billing, not as a publishable business claim. If a contact
   fact is wanted, it comes from the same explicit intake answer as (1).
3. Weakest: chrome suppresses the footer contact block unless a published-contact
   flag is set — leaves the two-contract column in place, guards one consumer.

## How to verify a fix

Run an order whose payer email differs from any contact the brief asks to publish;
assert the built site serves ZERO occurrences of the payer email (all pages), the
evidence_base contains no fact rendering it, and the delivery email still reaches the
payer address. The probe-with-controls recipe is in this incident's NOTES entries.

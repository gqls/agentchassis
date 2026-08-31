# PLAN — 420 (order-intake slug): a billing address is not consent to publish

**Started 2026-08-31.** Lane: bugfix 417/420.
Bug: `bugs_open/420_HANDOFF_2026-08-31_order_intake_publishes_the_billing_email_as_the_sites_public_contact_and_registers_it_as_a_renderable_claim.md`
⚠ **420 is an ambiguous number** — the other 420 is the negation gate's prose walker. Resolve by slug.

## The problem, in one paragraph

One column, `sites.email`, carried two facts with different consent: where the platform reaches
the paying customer, and what the finished website shows the public. Nothing distinguished them,
so the first paid build published the payer's address on all 13 deployed pages. Separately, the
same seeding registered "Enquiries reach «payer»." as a renderable evidence FACT, which meant
deleting the address could never make it stay gone — a rebuild could legitimately re-emit it.

## Decisions, and why

**D1 — The published contact needs its own explicit answer; absent it, publish nothing.**
Owner ruling, 2026-08-31, for the class (the pre-existing one was per-site). Built to it, and
built so that a later widening changes a VALUE, not the architecture.

**D2 — An OPT-IN direction key, not a new column.** The delivery-contact store already exists
(`build_queue.direction.customer_email`, durable, written by the collector, read by the delivery
dispatch). A `sites.delivery_email` column would put a second contact-ish column on the table
every render path loads, guarded only by a comment — precisely what the 2026-08-02 ruling says
does not survive this tree. So: no schema change, no backfill.

**D3 — The platform half ships first, defaulting to publish-nothing.** The question ("what
contact should the site show?") belongs in the intake chat, which lives on the box, not this
repo. Consuming an optional key that the chat starts supplying later is the ruling's own shape,
and the default is the safe side.

**D4 — Keep `business_name`, drop `contact`.** The business name is *constitutive*: no page can
exist without naming the business, so giving it to a site-building order is inseparable from
consenting to publish it. The email is *severable* — the site is fully buildable without it, and
the intake never asked. The distinction is the SCOPE of the attestation, not its truth. The
incident's own owner-reviewed remedy drew exactly this line.

**D5 — Remove the `info@<domain>` synthesis.** Post-fix an empty column is the CORRECT and
common state; fabricating an address makes "the site publishes no contact" quietly false.
Flagged severable to the council.

**D6 — Do not edit the two spec-copying writers.** On a customer build they can only copy what
the seed put there, so closing the seed closes them at source; editing them would change estate
behaviour this bug does not own.

## Corrections to the originating brief (the bug file's own framing)

> **CORRECTED 2026-08-31:** "651's delivery-email-sender reads `sites.email`" is CONVENTION, NOT
> CODE. No code in the send action or `platform/delivery/` reads the column; `customer_email` is
> required `input_data` at dispatch. This makes the split a recipe update, not a chain change.

> **CORRECTED 2026-08-31:** the footer block was already gated on a non-empty email (bugs 111).
> The defect was the VALUE, never a missing gate — so candidate 3 would have guarded a door that
> was already shut.

> **ADDED 2026-08-31:** a fourth writer the bug file missed — the admin PATCH endpoint
> (`site_admin_handlers.go:363-367`), unconditional, the deliberate operator override.

## Phasing

1. ~~Owner's class contract ruling.~~ **OBTAINED.**
2. ~~Census the writers and readers, dated.~~ **DONE — 4 writers / 14 readers as of 2026-08-31.**
3. ~~Go: seed rebind + register gate + collector pass-through + two synthesis removals; tests + mutations.~~ **DONE, committed `162877051`, INERT until the roll.**
4. ~~Council submission.~~ **DONE — `2026df60-…`, verdict pending.**
5. **OWED:** read the verdict; act on REVISE/REJECTED.
6. **OWED after the roll:** the order-2 rehearsal (§6 of the bug file), immune to the three enumeration traps.
7. **OWED, not mine:** the box's intake chat must start asking the question and sending `published_contact`; shape to be agreed with that lane.

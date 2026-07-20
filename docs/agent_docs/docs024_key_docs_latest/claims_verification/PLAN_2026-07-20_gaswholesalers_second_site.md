# PLAN — gaswholesalers.com as the second evidence-base site

*Owner proposed 2026-07-20 (vetcomparison is being worked in another thread).
Reconnaissance done; **blocked on one owner decision** — see §4.*

---

## 1. Why this site is the right test, and why it is hard

Leopardess and gaswholesalers are **opposite shapes**, which is exactly what makes
this a real test rather than a repeat.

| | leopardess | gaswholesalers |
|---|---|---|
| What the site describes | our own platform | a fuel wholesaling business |
| Ground truth lives in | our code and our database | the real world |
| Machine-verifiable? | yes — `SELECT count(*)` proves a claim | **no** — no query can prove "we deliver on schedule" |
| Dominant claim type | numbers and named entities | first-person operational prose |
| Which lane does the work | V1 deterministic (numbers, banned patterns) | **V3 prose auditor — and little else** |

Everything built so far leans on the deterministic lane. On this site that lane has
almost nothing to bite on: no business numbers to check, no register to check them
against, no fabrication history yet recorded.

## 2. What the site actually asserts (measured, 2026-07-20)

A scan of all 101 deployed components found **174 distinct first-person operational
assertions across 19 pages**. A representative sample:

- *"We supply natural gas to forecourts, fleet depots, industrial facilities, and commercial operations across the UK."*
- *"We maintain robust supply relationships and logistics infrastructure so that forecourt operators, fleet managers, and industrial buyers receive consistent, on-schedule deliveries."*
- *"We supply both branded and unbranded fuel."*
- *"Our clients do not chase us for updates; we keep them informed before they need to ask."*
- *"Broad Geographic Coverage — We supply fuel across a wide network of service areas."*
- *"Our team understands the pressures facing fuel forecourt operators…"*
- *"We offer transparent pricing frameworks with no hidden broker fees."*

Every one of these asserts an operating reality: infrastructure, coverage,
relationships, service standards, an existing client base. **None is checkable by
any query available to the platform.** The heaviest pages are
`pricing-transparency` (19), `supply-terms-and-eligibility` (17), `who-we-serve`
(17), and `service-areas` (15).

Also present: `gas@contactforsales.com` and `+44 (0) 7934 524 911` — the same
contact pair as leopardess, which is a cross-site contamination smell worth a
separate look.

## 3. A real gap this exposes: there is no COLD-AUDIT mode

The layer is opt-in on `evidence_base` presence and, in V3, on the fact register
being non-empty (`check_opted_in` gates on `facts_text`). That is correct for a site
whose facts are known — but it means:

> **A site where nothing has been attested yet cannot be audited at all.**

Which is backwards for the case that needs auditing most. What is wanted for a
site like this is the inverse posture: *treat every operational assertion as
unsupported until someone attests it*, and produce the list.

**Fix candidate (small):** allow an evidence base to declare
`"posture": "cold_audit"`, which (a) satisfies the opt-in gate with zero facts and
(b) tells the V3 prompt that the register is deliberately empty, so *everything*
first-person and operational should be reported. The output is then the audit list a
human works down — turning the auditor into a discovery tool, not just a regression
check. This costs one prompt branch and one condition; no new pipeline.

## 4. ⇢ BLOCKED: the question only the owner can answer

**Is Gas Wholesalers a real trading business, or a demonstration site?**

The honest remedy differs completely, and I cannot tell from inside the platform:

- **If it is real** — the owner attests the facts (coverage area, product range,
  whether deliveries are actually operated or brokered, whether there are clients),
  those become `attested_by` fact rows, and anything the owner will not attest gets
  removed or reframed as an offer ("we can arrange") rather than an operating claim.
- **If it is a demo** — then all 174 assertions are fiction presented as fact, and
  the fix is not a register but a decision about what a demo site may say. That is a
  policy question about the whole demo fleet, not a claims-verification task.

Nothing should be written to this site's evidence base until that is settled:
seeding a register with guesses would be inventing the very facts this layer exists
to prevent.

## 5. Once unblocked — the sequence

1. Owner attests what is true → build `evidence_base` with `attested_by` facts,
   plus `banned_claims` for anything ruled out.
2. Build the cold-audit posture (§3) if the register starts thin.
3. Run V3 → work the findings list → each ruling either adds a fact or removes copy.
4. Only then wire V1's number lane; it will have little to do here, which is itself
   the finding: **the deterministic lane does not generalise to non-self-describing
   sites, and the prose lane is the product for them.**

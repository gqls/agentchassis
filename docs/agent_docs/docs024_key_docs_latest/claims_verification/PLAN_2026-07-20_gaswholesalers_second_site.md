# PLAN — gaswholesalers.com as the second evidence-base site

*Owner proposed 2026-07-20 (vetcomparison is being worked in another thread).
Reconnaissance done; the blocking question was ANSWERED the same day — see §4.
The site's repositioning is routed to `features_open/006`; this thread keeps only
the verification slice (§4b).*

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

## 4. RESOLVED 2026-07-20 — neither real nor demo: the site is simply WRONG

Owner's answer:

> "gaswholesalers.com is not a demo site and I do not trade gas wholesale, it
> should be a site aimed at users (highly paid gas traders and ceo's of big oil
> corporations perhaps) providing top quality analysis and tools… The content on
> there is substantially wrong."

So both branches of the question were wrong. It is a **real site with a real
intended audience, currently asserting a business the owner is not in.** All 174
operational assertions are false — not exaggerated, not stale: false. There is no
register to build from them, because none of them can become a fact.

**What this makes gaswholesalers, for this workstream:** the cleanest possible
**cold-audit pilot**. Leopardess was remediation of a mostly-true site with pockets
of fabrication. This is a site whose entire assertion surface is unsupported, with
an empty register — exactly the case §3 says the layer cannot currently handle.

**Routed out of this thread:** the repositioning itself, the rewrite, and the new
AI-influence page are content/site work, not claims-verification work. They are
captured in `features_open/006_FEATURE_gaswholesalers_repositioning_and_ai_influence_page.md`
(and the later chatbot in `007`). This thread does not write that copy.

## 4b. What this thread SHOULD take from it

1. **Build the cold-audit posture (§3).** gaswholesalers is now a real, waiting
   test case with a measured expected result: a correct cold audit should report on
   the order of 174 unsupported assertions concentrated in
   `pricing-transparency` / `supply-terms-and-eligibility` / `who-we-serve` /
   `service-areas`. That is a benchmark, in the same spirit as the leopardess B1–B7
   corpus: history graded as a test.
2. **A new source kind is coming.** The AI-influence page (feature `006`) must be
   "very well researched and verified" — which means claims sourced to external
   citations. The schema today has `sql`, `artifact` and `attested_by`; a cited
   publication is none of them. Likely
   `source: {citation: "<publisher, title, date, url>", accessed: "<date>"}`, with
   a staleness policy, since adoption statistics age badly. **Design this before
   that page is written**, or it will be written first and audited never.
3. **Forward mode, not just remediation.** Every use so far has been catching what
   already shipped. That page is the first chance to run the layer the other way
   round: evidence base first, copy second, nothing asserted that is not already in
   the register. If that works it is the more valuable mode by far.
4. **Watch the chatbot question.** `007` asks whether the claims layer should gate
   bot responses. If yes, verification has to become *pre-response* rather than
   post-publish — a materially harder problem, and it would reshape this workstream's
   roadmap beyond V4. Do not plan V5 without an answer.

## 5. Sequence, now unblocked

1. **Cold audit first** — there is nothing to attest about the CURRENT copy; it
   describes a business that does not exist. Build the cold-audit posture, run it,
   and use its output as the demolition list for the rewrite (feature `006`).
2. Build the cold-audit posture (§3) if the register starts thin.
3. Run V3 → work the findings list → each ruling either adds a fact or removes copy.
4. Only then wire V1's number lane; it will have little to do here, which is itself
   the finding: **the deterministic lane does not generalise to non-self-describing
   sites, and the prose lane is the product for them.**

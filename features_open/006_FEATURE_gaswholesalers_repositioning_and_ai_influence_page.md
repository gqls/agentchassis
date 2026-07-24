# 006 FEATURE — gaswholesalers.com: reposition to an analysis/tools site, plus an "AI influence" page

**Raised:** 2026-07-20, by the owner, while the claims-verification thread was
running a second-site pilot on this domain.
**Status:** direction given, not designed, not started.
**Owned by:** NOT the claims-verification thread — routed here deliberately so a
site/content thread picks it up. Claims-verification's own slice is noted in §5.
**Related:** `features_open/007_FEATURE_ai_advisory_chatbot_freemium.md` (the
chatbot that belongs on the page this feature creates),
`docs024_key_docs_latest/claims_verification/PLAN_2026-07-20_gaswholesalers_second_site.md`
(the evidence that the current site is wrong, and how wrong),
`features_open/013_FEATURE_three_tier_ai_tooling_funnel.md` (2026-07-24 — this
feature's shape, generalised to most of the domain portfolio; the three-level
threats/advantages/remediation frame in §3 below is adopted there as the
canonical shape for the honest-AI-impact page fleet-wide).

**Update 2026-07-24:** a separate session generalised this single-domain brief
into a fleet-wide "AI section per site" workstream
(`docs024_key_docs_latest/per_site_ai/`). Nothing here changes — this remains
the owned, committed instance — but the open question in §6.3 (flagship
editorial vs. entry point to the paid advisory flow) is effectively answered by
that workstream's design: the honest AI-impact page sits *above* a produced Tier
2/3 tool funnel, not as a standalone article. See `013` for the reconciliation.

## 1. The correction — what the site is FOR

Owner, verbatim:

> "gaswholesalers.com is not a demo site and I do not trade gas wholesale, it
> should be a site aimed at users (highly paid gas traders and ceo's of big oil
> corporations perhaps) providing top quality analysis and tools for their day to
> day use and training. The content on there is substantially wrong."

So the site is **not** a fuel supplier. It is an **analysis, tooling and training
resource for a senior industry audience**: traders, procurement leadership, oil
and gas executives.

## 2. Why this is urgent, with a measurement

The claims-verification thread scanned all 101 deployed components on 2026-07-20
and found **174 first-person operational assertions across 19 pages** — every one
of them asserting an operating reality the business does not have:

- *"We supply natural gas to forecourts, fleet depots, industrial facilities, and commercial operations across the UK."*
- *"We maintain robust supply relationships and logistics infrastructure…"*
- *"Our clients do not chase us for updates; we keep them informed before they need to ask."*
- *"Broad Geographic Coverage — We supply fuel across a wide network of service areas."*

Heaviest pages: `pricing-transparency` (19), `supply-terms-and-eligibility` (17),
`who-we-serve` (17), `service-areas` (15). **These are not stylistic misses — they
are claims to be a business the owner is not in.** The whole assertion surface goes,
not a phrase here and there. Treat the existing copy as a rewrite target, not an
edit target.

Also flagged: the site publishes `gas@contactforsales.com` and
`+44 (0) 7934 524 911` — the same contact pair as leopardessconsulting.co.uk. Worth
a separate look at whether contact details are being shared across unrelated sites.

## 3. The new page the owner wants: "AI influence"

A **top-level nav entry**, explaining what AI means for this industry — honestly.
Owner's requirements, in their own terms:

- "what the threats and advantages are from AI in that industry"
- "a really truthful page"
- remediation suggestions at **three levels**: "at corporate level, at employee
  level and at personal level"
- "all very well researched and verified"

This is the site's flagship credibility piece: a senior audience will judge the
whole site by whether this page is honest about AI rather than either hyping it or
dismissing it.

## 4. What makes it non-obvious

- **The audience out-knows a generic page.** Traders and oil-and-gas executives
  will already have opinions on AI. A page of general "AI is transformative"
  material is worse than no page: it disqualifies the site with exactly the readers
  it targets.
- **"Threats" includes threats to the reader personally** — role displacement for
  well-paid analytical work is the elephant. A page that covers corporate risk and
  skips "your own job" reads as evasive to that audience.
- **Three levels means three different remedies**, not one list re-labelled.
  Corporate (governance, data boundaries, procurement of models), employee (skills,
  tool fluency, where human judgement stays decisive), personal (career, exposure,
  what to learn) genuinely diverge.
- **"Well researched and verified" is a hard constraint, not a wish** — see §5.

## 5. The claims-verification slice (the only part that thread owns)

This page is the first content on the platform that must be built **under evidence
discipline from the start**, rather than audited after the fact. That is a
forward-looking use of the claims layer and a genuinely new mode for it:

- Every factual claim on the page (adoption rates, incident examples, regulatory
  positions, job-impact figures) needs a source, and the source needs to be in the
  site's `evidence_base` before the claim ships.
- Research-sourced facts are a **new source kind** — the current schema has
  `sql` (live-verifiable), `artifact` (our code) and `attested_by` (human word).
  A cited external source is none of those; it likely wants
  `source: {citation: "<publisher, title, date, url>", accessed: "<date>"}` with a
  freshness policy, since a 2024 adoption statistic ages badly.
- The current site needs a **cold audit** first: nothing it asserts is true, so
  the register starts empty and every operational assertion should be reported.
  The claims layer has no cold-audit mode today (documented in the
  claims-verification plan) — that gap is that thread's to close.

**Boundary:** claims-verification builds the *verification* mechanism and the
evidence base. It does not write the page, choose the positioning, or design the
nav. That is this feature's owner.

## 6. Open questions

1. What survives of the 19 existing pages? Some (unit converter, fleet fuel
   consumption calculator) are genuinely useful **tools** and fit the new
   positioning; most of the prose does not.
2. Does the site claim any first-party expertise at all, or is it explicitly a
   curation/analysis service? That decides whether "we" may appear in copy.
3. Where does the AI page sit relative to the tools — flagship editorial, or the
   entry point to the paid advisory flow (`007`)?

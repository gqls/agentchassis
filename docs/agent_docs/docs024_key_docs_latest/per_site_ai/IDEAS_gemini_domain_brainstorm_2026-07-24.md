# IDEAS — salvaged brainstorm material (external-LLM session, 2026-07-24)

*Reference document, not one of the standing five. Raw material extracted from
a Gemini strategy conversation the owner ran in parallel and brought back for
extraction. Decisions drawn from this material live in `PLAN`; the
kept/adapted/discarded reasoning lives in `NOTES`. This file exists so the
concrete tool ideas aren't lost, while flagged clearly as unverified.*

**⚠ Groundedness warning — read before reusing anything below.** The 15 sample
domains this material was generated against do **not exist in the platform's
`sites` table** (checked 2026-07-24). Every archetype/vertical assignment below
was inferred by an LLM from the **domain string alone**, with no real audience
research, no confirmed ownership, and no confirmed vertical. Treat every idea
below as a brainstorm seed for a pool-level ideation pass, never as a validated
per-domain plan. Before building anything for a specific domain: confirm it is
an actual onboarded site, confirm its real audience/vertical from the platform
(not the domain name), and re-run the pool-level prompt below for its real pool.

---

## 1. The four reusable Tier-2 templates

Distinct from this workstream's Tier-3 signature-operation patterns (P1–P5 in
`PLAN`), these are candidate **generation shapes for the cheap, sticky Tier-2
utility** — narrow, single-purpose, fast/cached-model tools meant for weekly
habitual use by a named professional persona, not a one-off deliverable.

- **A — Extractor/Summariser.** User pastes a URL or raw text → scrape/parse →
  fast LLM extracts 2–3 vertical-specific structured facts → renders a card.
  *e.g. paste a property listing → extract lease-years/service-charge/ground-rent
  red flags; paste a job spec → extract salary range + hidden stack requirements.*
- **B — Transformer/Generator.** User enters raw notes → fast LLM applies the
  site's brand/tone rules → outputs a downloadable/copyable asset.
  *e.g. property specs → social listing copy; article link → thread + newsletter
  teaser.*
- **C — RAG Fast-Checker (evidence lookup).** User enters a query → pgvector
  search against the pool's `evidence_base` → fast LLM summarises the top
  citations. *e.g. plant/ingredient name → verified toxicity register + hazard
  level; provider name → historical rate/fee drift.*
- **D — Diagnostic Schema Generator.** User states a broad goal → LLM outputs a
  structured, step-by-step interactive checklist. *e.g. project type → building-
  regs compliance checklist; workflow description → automation-suitability score.*

## 2. Archetype × pattern design grid

Two orthogonal axes for choosing a pool's Tier-3 shape. **Archetype** = who the
reader is and what risk/emotional register applies (from the external session).
**Pattern** = what gets produced (P1–P5, already defined in `PLAN` D7). Cross
them rather than picking one taxonomy:

| Archetype (who/why) | Best-fit pattern(s) | Notes |
|---|---|---|
| High-Trust & Regulated (finance, energy, insurance) | P2 verified report; P3 comparison | Liability-safe framing only — information not advice; P4/novelty is usually wrong tone here |
| E-Commerce & Sensory (perfume, food/drink, physical goods) | P1 produce-&-deploy (listing/microsite); P4 novelty | Sensory-gap-bridging tools convert directly to product sales |
| Technical & B2B Infrastructure (engineering, dev tools, print/pre-press) | P1 (spec/blueprint deploy); P2 (feasibility audit); P5 (asset packs) | Workflow-acceleration framing; strongest natural fit for Tier-2 Extractor/Diagnostic templates too |
| Historical, Cultural & Creative | P4 novelty (primary fit); P1 occasionally (microsite) | Emotion/story-driven; weakest Tier-3 utility case, strongest funnel-top virality |

## 3. Fifteen ungrounded example specialisations (raw brainstorm only)

Kept as seed material for whichever real pool a live domain like these turns
out to belong to — **not evidence about these specific domains**, since none
are onboarded. Condensed from the external session; full per-domain detail
(personas, saga walkthroughs, monetisation notes) is not reproduced here as it
inherits the same groundedness caveat and would overstate confidence.

| Domain (as named, unverified) | Guessed archetype | Tier 1 probe idea | Tier 2 utility idea | Tier 3 signature idea |
|---|---|---|---|---|
| acousticcameras.com | Technical/B2B | speed-of-sound / wavelength calculator | ISO machinery noise-limit inspector (RAG) | factory attenuation + sensor-layout audit |
| actuariel.com | High-trust/Regulated | annuity PV/FV matrix | Solvency II capital-charge extractor | portfolio stress-test / longevity risk audit |
| adcentre.org | Marketing/Digital | ad banner spec/aspect-ratio grid | ad copy transformer + policy checker | competitor ad/angle-gap audit |
| adjustablewalkingsticks.com | E-commerce/Health | cane-height ergonomic estimator | mobility-aid terrain/grip selector (RAG) | OT equipment readiness report |
| adultchristmas.co.uk | Cultural/Seasonal | Secret Santa budget/pairing randomiser | festive cocktail/pairing engine | custom event/itinerary generator |
| adversecreditmortgage.co.uk | High-trust/Regulated | LTV/deposit-tier matrix | credit-defect expiry & lender-tier inspector | adverse-credit feasibility dossier |
| agentandhuman.com | Technical/B2B (framework-sales) | AI-vs-human task cost matrix | human-checkpoint workflow architect | enterprise automation/risk audit (live demo) |
| airportcollections.com | Travel/Logistics | flight-delay/terminal buffer calculator | vehicle/luggage fit fast-check | corporate transfer logistics coordinator |
| aiwebmaintenance.com | Technical/B2B (framework-sales) | SSL/security-header inspector | broken-link/301-redirect generator | full site diagnostics/integrity audit |
| alternativepower.co.uk | Technical/Energy | solar array yield calculator | dynamic tariff-arbitrage checker | whole-home energy-independence blueprint |
| ancestryonline.co.uk | Cultural/History | regnal-year/date converter | census occupation decoder (RAG) | ancestral parish/record-location blueprint |
| apis.uk | Technical/Dev tools | JSON→TypeScript generator | UK open-API auth/rate-limit inspector | API architecture/SDK wrapper generator |
| applejuicers.com | E-commerce/Agriculture | fruit-to-juice volume estimator | apple-variety blending helper | micro-cidery feasibility audit |
| arabianperfumes.co.uk | E-commerce/Sensory | fragrance longevity/volatility calculator | scent-layering combinator | bespoke olfactory profile dossier |
| artworkers.co.uk | Technical/Pre-press | bleed/trim/spine-width calculator | print-method/TAC ink-limit inspector | pre-flight packaging verification brief |

## 3b. Second batch (2026-07-24) — three more ungrounded, five real (see GROUNDED file)

A second external-LLM batch covered 8 domains. **Five are real, deployed
platform sites** — for those, see
`GROUNDED_domain_profiles_2026-07-24.md`, which corrects the guesses against
real page content rather than repeating them here ungrounded. **Three do not
exist in the `sites` table** (checked 2026-07-24), the same pattern as the
first 15 — kept below as ungrounded brainstorm only:

| Domain (unverified) | Guessed archetype | Tier 1 probe idea | Tier 2 utility idea | Tier 3 signature idea |
|---|---|---|---|---|
| agritec.uk | Technical/B2B (agriculture) | NPK application-density calculator | UK DEFRA/FETF grant-eligibility inspector (RAG) | precision-farming feasibility audit |
| mortgagecalculator.co.uk | High-trust/Regulated | stamp-duty & repayment crossover matrix | fixed-rate-expiry/rate-cliff inspector | mortgage affordability stress-test dossier |
| websitedesign.com | Technical/B2B (framework showcase) | Core Web Vitals/page-weight inspector | Tailwind design-token generator | automated site-redesign + preview deploy |

## 4. Pool-level ideation prompt (adapted, corrected)

Repurposed from the external session's per-domain template into a **pool-level**
tool, per `PLAN` D11 — run once per vertical pool (≤17 runs), not once per
domain. Corrected against real platform facts (real schema references removed
in favour of "ask, don't assume"; the modality gap and pool-first doctrine made
explicit; a groundedness gate added on the input itself).

```
You are a product strategist working on the agentchassis platform: a
Kafka-orchestrated, Postgres-backed multi-agent system that plans, builds,
verifies and operates a fleet of 1,000+ content websites, organised into ~17
vertical pools. Full capability inventory is provided separately — use only
what it lists as [LIVE]; treat [PARTIAL]/[DESIGNED] as not-yet-usable and no
voice/video/animation modality as absent unless told otherwise.

GROUNDEDNESS GATE: I am giving you a POOL name and its real audience.v1 profile
(who reads it, how it differs from sibling pools/domains, editorial
directives), pulled from the live platform — not a domain name to guess from.
If I have not given you a real audience profile, say so and stop; do not infer
one from a domain string.

ANTI-PATTERNS — do not suggest:
- an open "ask me anything" chatbot (no moat, unbounded cost/liability)
- a generic SEO blog-post generator
- unbounded medical/financial/legal advice
- diffusion-model-drawn data charts (charts are always code-rendered from real
  series; a generative model may only add an annotation layer)

Design a 3-tier product funnel for this pool:
- TIER 1 (4 ideas): zero-LLM-cost client-side/server calculators. Exact
  inputs, formula, instant output.
- TIER 2 (4 ideas, using templates A–D: Extractor / Transformer / RAG
  Fast-Checker / Diagnostic Schema Generator): narrow, cheap, weekly-habit
  tools. Input, lookup/generation step, structured output artifact.
- TIER 3 (2 ideas): high-ticket orchestrated sagas producing a verified,
  deployed or delivered artifact a single LLM call cannot produce. Walk the
  saga: intake → research/scrape → fact-verification against the pool's
  evidence register → chart/image/asset generation → assembly → delivery.
  State explicitly whether this is a fixed report or a conversational
  (chat-delivered) Tier 3, and why.

Then state which of the 4 archetypes (High-Trust & Regulated / E-Commerce &
Sensory / Technical & B2B / Historical & Creative) this pool fits and why, and
flag any liability concern (advice-adjacent claims, regulated-industry
disclaimers) explicitly rather than papering over it.
```

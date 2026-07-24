# PLAN — per-site AI section / operation (2026-07-21)

*Design, phasing, decisions and their reasons. Corrections to the brief live here,
marked as corrections. Still in deliberation — nothing below is committed to
build yet.*

## What we're trying to do
Give most sites a nav-linked AI area that (a) genuinely helps that vertical's
audience navigate the AI transition and (b) advertises our framework as a
product. The centre of the offering is shifting from "content + a chatbot" to
**one signature AI *operation* per site** — an interactive tool that produces a
real, original, site-specific artifact — with honest editorial as its supporting
layer.

## Decisions / working positions (with reasons)

**D1 — Compete on production, not intelligence.** We will not position against
foundation models on reasoning. The defensible edge is what a multi-agent,
multi-modal, tool-using, self-verifying pipeline can *produce and deliver* that a
single call cannot: a finished, checked, deployed artifact. *Reason:* honest and
durable; the opposite claim loses the moment a user opens ChatGPT.

**D2 — The unit is a "signature operation" per site, chatbot-fronted.** Each site
exposes one interactive operation whose front-end is a cheap-model, tightly
vertical-prompted chat that gathers inputs and triggers the pipeline. *Reason:* a
constrained intake-and-produce flow is cheaper, safer (little open-ended
generation), more defensible, and more impressive (a real artifact) than an open
Q&A bot.

**D3 — Editorial is the supporting layer, not the headline.** The honest,
job-title-specific AI articles remain — for SEO, trust, and grounding the chat —
but they no longer *are* the product. *Reason (correction to the original
brief):* the brief led with the articles + a generic vertical chatbot; the owner's
follow-up established that the produced artifact is the differentiator and the
articles are context around it.

**D4 — Paywall the deliverable, free the chat.** Free rate-limited intake/chat as
lead-gen; pay for the produced artifact and/or framework access. *Reason:*
metering chat competes with free frontier chat and adds friction ahead of the
cross-sell; the artifact is where real, chargeable value sits, and where cost is
incurred.

**D5 — The operation doubles as the framework's advert.** The best proof that
"you can build AI systems with our framework" is one the visitor just used;
choose operations that *visibly* exercise multiple modalities. *Reason:* the
meta-goal is selling/renting the framework; the operation is live marketing
collateral. Converts on fundamentallyai.com (brochure_component_library
workstream).

## Phasing (provisional)
- **Phase 0 (now):** deliberate the operation catalogue + the reusable design
  method for choosing one per site. No code.
- **Phase 1:** ship the editorial/insights layer on the existing site-build +
  claims machinery (low risk, high certainty), with the nav link.
- **Phase 2:** build ONE signature operation end-to-end on the strongest,
  quality-gatable candidate site as a pilot (likely the domain→site op on a
  web-design site — core competency, self-selling), fronted by the cheap intake
  chat, with the paywall on the artifact.
- **Phase 3:** generalise the operation pattern + intake-chat scaffold across the
  fleet; decide separate-cluster/model-routing only where the pipeline actually
  needs it.

## Open forks (need owner)
1. Signature operation(s) per site; utility-vs-novelty balance (see NOTES).
2. Does "chatbot = intake front-end to an operation" fully replace the open
   vertical Q&A bot, or do we want both?
3. Separate cluster + specialist models: now, or deferred until Phase 2 proves
   the pipeline needs it?

## Decisions added 2026-07-24 (external-LLM session extraction + features_open discovery)

**D9 — Formalise a third funnel tier: the cheap, sticky Tier-2 utility, sitting
between the free probe and the paid signature operation.** Closes the owner's
2026-07-22 request for something that "draws targeted traffic" on top of the
free client-side widgets, separate from the paid deliverable. Four reusable
generation shapes (not tied to any one pool): **Extractor/Summariser**,
**Transformer/Generator**, **RAG Fast-Checker**, **Diagnostic Schema
Generator** — see `IDEAS_gemini_domain_brainstorm_2026-07-24.md` §1 for detail.
*Reason:* a single-prompt open chat has no retention; a narrow tool that saves
a named professional real weekly effort does. This was missing from D6/D7 —
D6 covered the free layer, D7 covered the paid layer, nothing covered the
middle.

**D10 — Archetype × Pattern is a two-axis grid, not one taxonomy.** Archetype
(High-Trust & Regulated / E-Commerce & Sensory / Technical & B2B / Historical &
Creative) answers WHO the reader is and what risk/emotional register applies;
Pattern (P1–P5, D7) answers WHAT gets produced. Cross them when choosing a
pool's Tier-3 shape — see `IDEAS_...md` §2 for the grid. *Reason:* conflating
them loses information (two pools can share a pattern but need opposite tone —
a regulated-finance verified report and a technical-B2B feasibility audit are
both P2, but one needs "information not advice" disclaimers throughout and the
other doesn't).

**D11 — Ideation runs at POOL level, not per-domain.** A deep per-domain
strategy prompt doesn't scale to 1,000+ sites and re-derives the same pool-level
thinking every time. The corrected pool-level prompt lives in
`IDEAS_gemini_domain_brainstorm_2026-07-24.md` §4: it forces a stated
audience.v1 input (refuses to guess a vertical from the domain string), runs
≤17 times (once per pool), and a separate, much cheaper per-site specialisation
step injects the real site's audience/evidence_base/brand afterward. *Reason:*
this is the same pool/per-site split D7 already established, applied to the
ideation process itself, not just the output.

**D12 — Chat is not always the cheap tier.** Reconciles with
`features_open/007` (discovered 2026-07-24 — the owner's own prior brief for
gaswholesalers.com's advisory chatbot, deliberately using a *high-quality*
model with deep research and caching, not a cheap intake bot). Two distinct
things, not one: a **Tier-2 utility chat** (cheap, narrow, high-frequency) and
a **Tier-3 conversational deliverable** (expensive, deep-research, cached,
paid per-session/per-N-submissions, for high-value B2B/executive verticals
only). D2's original "chatbot = cheap intake front-door" position was too
narrow — correcting it here rather than in place, per the corrections
discipline.

**D13 — Graduated monetisation ladder, refining D4.** free/rate-limited (Tier 1
always; a few Tier-2 runs/day) → email-gate (more Tier-2 runs) →
micro-subscription (unlimited Tier-2) → per-deliverable paywall (Tier 3).
**Must reuse whatever entitlement/payment plumbing `features_open/003` is
already tracking** (a local `client_entitlements` cache — auth can't be called
synchronously across thousands of sites per tick — and a possible BIZ-014
`saas_cheap`/`portfolio` tier flag) rather than invent a parallel mechanism.
Check before designing.

**D14 — Per-domain identity/payment stays independent (owner decision,
2026-07-24).** No unified cross-portfolio account; each domain's Tier-2/3
gating is its own entitlement scope. Consistent with D8 (each site sold/rented
independently).

**Connection to `features_open/006`**: this whole workstream is the fleet-wide
generalisation of a single-domain feature the owner already committed to
(gaswholesalers.com repositioning + "AI influence" page). `006`'s three-level
remediation frame — **corporate / employee / personal** — is adopted here as
the canonical shape for the honest-AI-impact editorial layer platform-wide,
replacing this workstream's earlier, vaguer "job-title-specific articles"
framing (D3). `006` also names a concrete platform gap this needs: a
`citation` source kind in `evidence_base` for research-sourced claims — this is
claims-verification's **V5**, designed but not built (CAPABILITIES §4).
Shipping an evidence-backed AI-impact page fleet-wide is gated on V5 shipping,
not only on this workstream's own design work.

## Decisions added 2026-07-24b (second external-LLM batch, ground-checked)

**D15 — Pilot recommendation revised, on real evidence.** The 2026-07-21
"Property" pick is superseded: it was never grounded in an actual domain
(Property is a pool). A second external-LLM batch covering 8 domains, checked
against the live `sites`/`pages` tables, produced real candidates — full
findings in `GROUNDED_domain_profiles_2026-07-24.md`:
1. **robot-hands.com — first choice.** Tier 1 (3 calculators) and a Tier-2-ish
   tool (MatchMatrix) are already live; the only gap is Tier 3. A verified
   procurement/comparison dossier extending MatchMatrix also structurally
   remediates `bugs_open/043` (fabricated stats found on this exact site) —
   the build and the bug-fix are the same piece of work.
2. **ai-agent-orchestration.com + leopardessconsulting.co.uk — close second,
   as a pair.** Both are AI-services sites (D8) with real Tier-1/2 tools
   already live (readiness quizzes, cost/ROI calculators) and genuine
   audited case-study content ready to ground a Tier-3 op. Start with a
   personalised "Architecture Fit Report" (cites the real case studies, no
   live deploy) before Gemini's riskier "live 3-minute fleet deploy" idea,
   which depends on the platform's most bug-fought delivery path.
3. **gaswholesalers.com** — strong concept, **sequence-blocked**: the DB
   still shows the stale content `006` exists to replace (checked live,
   2026-07-24). Do not build tools on top of it until `006` ships.
4. **idea.uk is not a pilot target — it is a working reference
   implementation**, verified live the same day by another session
   (`idea-uk-vm-site-workstream`): a real, paid, gated report-delivery funnel
   already runs from the home/header CTA to `/report.html`. Study it (their
   HANDOFF) before designing a new Tier-3 pattern from scratch; don't start
   parallel work on a domain another session owns.

**D16 — Groundedness gate needs real content, not just `audience.v1`.**
`audience.v1` was empty on every site checked in both batches (checked
2026-07-21 and 2026-07-24) — it is not yet a usable grounding source. The
technique that actually caught wrong guesses both times was pulling real
`pages.title`/`sites.tagline` directly. D11's pool-level ideation prompt (and
any future domain-strategy pass) must be fed real pulled page content, not a
domain name and not an (currently empty) `audience.v1` reference. **Accuracy of
an external LLM's guess correlated with how literally the domain name matched
the real business — sometimes right, sometimes badly wrong (idea.uk,
leopardessconsulting.co.uk) — so the name itself is never sufficient evidence
either way.**

## ~~STATUS 2026-07-24: PAUSED~~ → RESUMED same day (resume condition met)

Owner paused implementation pending the robot-hands site thread; owner
confirmed later the same day it had finished, and verification agrees:
**robot-hands-site-fixes-workstream CLOSED 2026-07-24** (R1–R9 all verified
live — MatchMatrix v2 scores all 10 grippers with per-technology physics;
invented-methodology prose rewritten to the real mechanics; query-backed
catalog grid; 043 fleet sweep RUN; candidate 3 shipped = migration 201 +
`evidence_base` writer_blocks on this site). `bugs_open/043` itself stays open
as a *platform* bug (candidates 1/2 unbuilt) but the SUMMARY landed
(4f048b063), which was the stated resume condition. The site is now
de-fabricated, verified, and evidence-fenced — cleaner ground for the pilot
than when it was picked.

**Pre-implementation facts established 2026-07-24 (both checked live):**
- **No payment/entitlement plumbing exists** — zero tables matching
  entitlement/payment/stripe/subscription/checkout in clients_db. A
  card-paywalled Tier 3 needs infra that doesn't exist; pilot gating must use
  a rung of the D13 ladder that works today (email-gate / request-form).
- **idea.uk's live "paid report" funnel is a request-form pattern**, not
  automated checkout — `report-request-form` component (`render_mode=template`)
  on `/report.html`. The proven reference: structured request form → pipeline
  fulfils → locked URL delivered. The pilot should copy this shape.
- Inherited landmines from the closed robot-hands thread that bind the pilot
  build: tool pages bypass the sections path (reference implementation
  documented in their memory/handoff); a tool deployed without a DB source is
  one rebuild from being wiped; `page_components.id` is NOT stable across
  re-renders; SELECT-before-supersede on `site_specs` aspects (a 043-lane
  session clobbered banned_claims unread); CTA URLs are label-blind until
  `bugs_open/023`; check `render_mode` + field `source` before any
  content_data edit.

## D17 — pilot scope locked (owner decisions, 2026-07-24)

- **Deliverable: Gripper Selection & Integration Dossier** (owner delegated the
  choice; both offered options were judged — "Application Fit Report" was too
  thin a delta over the free MatchMatrix, "full procurement dossier" added
  live-scrape fabrication surface). The dossier composes the site's four
  existing verified tools (MatchMatrix v2 physics, payload calc, grip-force/
  friction calc, cycle-time estimator) over the 10 verified grippers into an
  engineer-grade justification document: scored shortlist with printed
  formulas, safety factors, environment gating, throughput projections,
  mounting compatibility, vendor-questions checklist — every spec traced to
  `source_url`+`verified_date`, charts code-rendered, deployed as a locked
  shareable page. No fresh vendor scraping in the pilot (deferred to v2).
- **Gating: request → fulfil → email a locked link** (idea.uk's proven
  report-request pattern). Lead-gen mode first; pricing layered later —
  manual invoice, then automated payment once entitlement plumbing exists
  (no payment tables exist today, checked).
- **Intake: CHAT** (owner choice, consistent with the original brief's
  preference). Cheap model, tight vertical prompt, gathers the same
  structured fields a form would, ends in email capture + submission. First
  chat-intake build on the platform — no precedent; this is the pilot's main
  new-engineering risk and the part to design most carefully.

## Reuse / connections to verify before building
- Site-build machinery (sections/components/render, page rerender).
- Claims-verification / evidence_base fence — for grounded, honest editorial.
- News-feed pooling — vertical news feeds for freshness/grounding.
- fundamentallyai.com / brochure_component_library — the conversion destination.
- Councils / fix loops — the verification layer of any pipeline.
*(All [INFERRED] reusable from memory; confirm current state against the repo/DB
before relying on any of them.)*

# NOTES — per-site AI section / operation

*Technical running log. Append-only, newest at the bottom. Record missteps and
wrong turns, not just conclusions. Mark unverified claims [INFERRED]/[ASSUMED].*

---

## 2026-07-21 — workstream opened (deliberation only, no code)

Session "AI page". Brand-new workstream; grep of docs024 confirmed no prior
per-site-AI dir. Created dir + started standing docs (this file, PLAN, README).

### The brief (compressed)
Nav link → "AI insights"-style area on most sites. Contains (a) a vertical- and
site-specific AI chatbot; (b) honest, job-title-specific articles on AI risk +
benefit, grounded in real current news/research, made site-specific to avoid
cross-site duplicate content. Chatbot ideally on a separate cluster with
best-in-class specialist models across modalities; prompted/gated to stay
on-vertical; paywalled after ~2–3 free tries/day. Also seeds the sell of the
framework itself (sale/rent/subscription). Affiliate income secondary. Primary
goal: genuinely help users get ahead of the AI transition.

### Turn-1 critique (my position)
1. **Two products, not one.** Editorial section (static, gated-before-publish,
   cheap, SEO, rides existing site-build + claims machinery) vs. live chatbot
   (unbounded input, per-query cost, moderation/liability, paywall+identity
   infra). Decouple; editorial first.
2. **Reject "smarter than a foundation LLM."** [ASSUMED but high-confidence] a
   multi-agent wrapper over the same base models does not beat frontier chat on
   raw reasoning for "explain AI's impact on my job." Compete on grounding /
   specificity / structure / freshness instead.
3. **Paywall the deliverable, not the chat.** Metering "N/day per user" without
   login is trivially bypassed (IP/cookie). Free cheap chat = lead-gen; pay for
   the produced artifact.
4. Duplication ≈ quality: solve both with genuine vertical-native grounding
   (real job titles, real tools/products/regulation in that vertical), enforced
   by the existing claims-verification / evidence_base fence machinery.
5. Honesty about jobs is the moat but sequence agency-first; grounding
   non-negotiable; per-site tone risk (don't tell a site's own audience they're
   doomed as the headline).
6. Cross-sell: the operation IS the demo of the framework being sold; vertical
   site demonstrates → fundamentallyai.com converts (ties to the
   brochure_component_library workstream). Affiliate = disclosed bolt-on.

### Turn-2 reframe (owner-steered — the important shift)
Owner pushed on the real question: **what can a multi-agent, multi-modal
framework produce that a single foundation call cannot?** Working answer — the
moat is NOT intelligence, it is:

- **Orchestration into a finished artifact.** Research → copy → matched imagery →
  layout → voiceover → assembled video → *published*. Foundation chat returns
  pieces in a conversation; we return a produced, integrated, deployed thing.
- **Tool-use with real side-effects.** Register a sub-URL, deploy a site, write
  the DB, call external APIs. "Enter a domain, get a live site" is an *action*,
  not an answer. This is operational capability + infra, not model IQ.
- **Verification / adversarial loops.** Generate → critique → verify → revise
  (we already run councils / fix loops / claims verification). Higher-trust than
  a one-shot call wherever correctness matters.
- **Specialist-model routing + cost control.** Cheap model for chat/intake,
  reasoning model for the hard step, specialist image/voice/embedding models per
  subtask. Router-of-specialists beats a generalist on a composite task, at a
  cost we choose.
- **Persistent proprietary context.** Site research corpus + user behavioural
  signals + prior outputs. Foundation chat starts cold each session.

### The product shape this implies
**One "signature AI operation" per site**: an interactive, chatbot-fronted tool
that produces a real, original, site-specific artifact/service. Chatbot = the
*intake front-end* (cheap model, tight vertical prompt: light Q&A + gather inputs
+ trigger the pipeline + present/deliver the artifact) — NOT an open "ask me
anything" bot (which dodges most moderation/cost/liability problems too). The
honest job articles get demoted from "the product" to the supporting editorial
/ trust / SEO layer.

Owner's canonical example: web-design site → type a domain → we generate+deploy a
site on a sub-URL of one of our short domains → paywall/upsell to claim/expand /
own-domain / licence the framework. It's our core competency exposed as a
self-serve product AND the live advert for the framework.

### Two deliverable flavours (my framing, to test with owner)
- **Utility (business, the money):** AI-readiness audit for your business,
  competitor-comparison page, domain→site, catalogue-from-product-list,
  personalised vertical briefing (voiced).
- **Novelty/delight (traffic + demo):** "your cat as a Renaissance vet", collection
  showcase reel, peculiar voice pieces — shareable, memorable, visibly
  multi-modal, but converts to money less directly.
  → [INFERRED] each site wants one utility op (conversion) + optionally one
  novelty hook (top-of-funnel + the most visible multi-modal demo). Novelty as
  primary risks traffic that doesn't pay.

### Feasibility / risk filter (candidate operations must pass)
1. Can we produce it at genuinely non-embarrassing quality with current models?
2. Does it need real-time infra/side-effects we actually support?
3. Marginal cost/run vs. what the paywall recovers?
4. Is the artifact original/defensible, or does one foundation prompt already do
   it (no moat)? Moat is strongest where the op needs orchestration + tool-use +
   verification + deployment.
5. **[RISK]** A signature op that fails *live* in front of a prospect is worse
   than none — it's a live demo of the product not working. The domain→site op
   is core competency but also our most bug-fought path (brochure site "built but
   not live"; empty-sections / render bugs). First signature op must be one we
   can gate to quality before showing.

### Open questions for next turns (NOT decided)
- Which signature operation(s) per site; the utility-vs-novelty balance.
- Does the "chatbot = intake front-end to an operation" reframe land, or does the
  owner still want an open vertical Q&A bot as well?
- Separate cluster: real requirement now, or premature infra? (chat is cheap; the
  heavy pipeline is where model/infra choices actually matter.)
- Reusable per-site *design method* for choosing the operation, so it scales
  across the fleet instead of being bespoke each time.

## 2026-07-21b — client-side probes, the real fleet, and operation-patterns

### Grounding (verified against live DB + docs)
- `sites` table: 12 `status='deployed'` sites + 17 `status='pool'` synthetic
  verticals + system row. Portfolio is **~1,000+ domains** (per news_feed_pooling
  SUMMARY 2026-07-20), pools cover ~2/3.
- **Target-market thread = `news_feed_pooling`.** Established doctrine we inherit:
  every live site has a structured audience profile (who/differentiation/copy
  implications); the **per-site selection/differentiation layer IS the product**;
  near-duplicate names make it matter most. [storage location of the audience
  profile not yet confirmed — settings->'audience' was empty on robot-hands; TODO
  find it before wiring operation inputs to it.]
- The 17 pools (Property, Insurance, Mortgages, Savings/Investing, Construction,
  Energy, Industrial, Marketing/Digital, Travel, Vehicles, Vet/Animal, Web/Tech,
  Jobs/Work, Health, Business Services, Design/Creative, AI/Agents) = the real
  unit of design.

### Decision D6 — free client-side widgets = demand-probe + funnel-top, NOT the moat
Owner idea: trial free client-side-only tools, see which are popular, "making them
anyway." Accepted WITH a hard caveat:
- Client-side-only ⇒ no server orchestration, no specialist-model routing, no
  tool-use with side-effects, no verification loop, no proprietary corpus, no
  deploy = **exactly the commoditised single-widget class the moat is defined
  against.** "Client-side AI" is largely a contradiction (real LLM/image needs a
  server key or it leaks keys / costs per call); genuinely-free client-side tools
  are non-AI widgets (calculators, quizzes, configurators, estimators).
- So: ship a battery of free widgets per vertical as **demand discovery + SEO
  traffic + email capture + the free lead-in**; the paid signature operation is
  deliberately the thing a widget can't do. **Rank by intent depth (completion,
  email, "go deeper" clicks), NOT raw pageviews** — a free calculator outdraws the
  paid report 100:1 but the report is where the money is.

### Decision D7 — design per POOL, specialise per SITE; and pools cluster into ~5 patterns
Don't design 100 (or 1,000) operations. Design one signature operation per POOL,
specialised per site — the identical shared-machinery/per-site-selection split
news_feed_pooling already uses. Even the 17 pools collapse into ~5 reusable
**operation-patterns**:
1. **Produce-&-deploy a listing/microsite** (Web domain→site; Property brochure;
   Vehicle listing; trade/business one-pager). Moat: deploy + multimodal +
   anti-fabrication verification.
2. **Verified decision/strategy report** (Mortgages, Savings, Insurance, Energy,
   business AI-readiness audit). Moat: live-data tool-use + verification + doc/
   video gen. REGULATED ones (financial) = information-not-advice + disclaimers.
3. **Cited comparison / "is this fair?"** (vet cost, insurance quote, trade quote).
   Moat: research + anti-fabrication (the vetcomparison lesson — won't invent a
   price).
4. **Novelty shareable multimodal artifact** (pet portraits, collection reels,
   makeitaquote quote-images). Top-of-funnel + the most visible modality-join =
   the framework advert. NOT the paid primary.
5. **Business content/asset production** (product-list→styled deployed catalogue;
   branded social/video asset packs). Moat: multimodal + brand consistency +
   deploy.

### Decision D8 — the AI-services sites are special
fundamentallyai.com / ai-agent-orchestration.com / finetuning.uk /
leopardessconsulting: their signature operation should BE one of the patterns run
live *as the sales demo*, because the product they sell is the framework itself.

### Scorecard for picking a pool's operation (the refined filter)
Score each candidate on: (1) jobs-to-be-done urgency · (2) willingness to pay ·
(3) **moat depth = does it NEED orchestration+tool-use+verification+deploy** (the
gate; if a single prompt does it, it's a funnel-top widget not a paid op) ·
(4) ship-quality now · (5) modality showcase (framework advert value).

### Landmine noticed
Not every vertical has a strong *paid utility* op. Liability verticals (Vet,
Health) lean pattern 3 (comparison) + pattern 4 (novelty), NOT advice — triage/
diagnosis is a liability minefield. Say so rather than forcing a weak paid op.

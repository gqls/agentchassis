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

## 2026-07-24 — external-LLM session extraction + features_open discovery

Owner ran a parallel strategy conversation with Gemini (pasted our
2026-07-21 CAPABILITIES inventory as grounding), developed a 3-tier funnel +
domain-archetype framing + a reusable "universal domain strategy prompt", then
brought the transcript back for extraction. Task: strip good ideas, discard
noise, define what to say/do fleet-wide. Before touching that material, grepped
`features_open/` per CLAUDE.md's "grep before you file" discipline (should have
done this on 2026-07-21 when the workstream opened — did not; caught now, cost
nothing since nothing had been built).

### The discovery that reframes everything: 006/007/003 already exist
`features_open/006` (2026-07-20, four days before this workstream opened) is
the owner's own original, concrete, single-domain version of this exact idea:
gaswholesalers.com repositioned for oil/gas executives, "AI influence" nav page,
honest, three-level remediation (corporate/employee/personal). `007` is the
deferred advisory chatbot for that page — explicitly a **high-quality model**,
deep research, caching, paid per-session. `003` is the paid-tier/entitlement
design in progress (`client_entitlements` cache, BIZ-014 tier flag). This
workstream is not a new idea — it is 006/007 generalised to most of the fleet.
Filed `features_open/013` as the anchor and cross-linked all three. Decisions
folded into PLAN as D9–D14 plus the 006 connection note.

**Correction to this workstream's own D2 (2026-07-21):** "chatbot = cheap
intake front-door, never the product" was too narrow. `007` is direct
counter-evidence — the owner explicitly wanted a *high-quality*, expensive,
deep-research chat as the product for one vertical. Resolved as D12: cheap
Tier-2 utility chat and expensive Tier-3 conversational deliverable are two
different things; which one (if either) a pool gets depends on its archetype
and value-per-lead, not a blanket rule.

### Extraction: KEPT (genuinely new or genuinely sharper than what we had)
1. **Tier 2 "sticky go-to AI utility"** as a distinct funnel layer between the
   free probe and the paid deliverable — D9. This was a real gap: D6/D7 covered
   the free and paid ends and nothing covered the middle, which is exactly what
   the owner asked for two turns before this ("an AI tool... to draw targeted
   traffic... on top of the free algorithmic tools").
2. **Four reusable Tier-2 templates** (Extractor, Transformer, RAG Fast-Checker,
   Diagnostic Schema Generator) — a genuinely generalisable mechanism, not
   vertical-specific. Kept in `IDEAS_...md` §1.
3. **Archetype × Pattern two-axis grid** — D10. Sharper than either taxonomy
   alone; the archetype axis (who/why/risk-register) was implicit in this
   workstream's earlier liability discussion (turn 1 point 5) but never made
   explicit or crossed against the pattern axis.
4. **006's three-level remediation frame** (corporate/employee/personal) —
   technically from `006`, not from the Gemini session, but surfaced by this
   extraction pass. Replaces this workstream's vaguer "job-title-specific
   articles" language (D3) with something concrete and already owner-approved.
5. **"Artifact as lead magnet"** — a co-designed deliverable makes an
   email-gate read as delivery, not friction, so a user who has just spent two
   minutes configuring a report doesn't experience giving an email as marketing
   friction. Useful design guidance for Tier 2/3 gating (folded into D13's
   framing, not a separate decision).
6. **Graduated gating ladder** (free-rate-limited → email-gate → micro-sub →
   per-deliverable paywall) — D13, refines D4's binary free/paid split into
   something with intermediate rungs, while preserving D4's core rule (don't
   paywall the chat itself, paywall depth/volume/the deliverable).
7. **Modality→ROI heuristic** (text/data = trust+speed; voice = reassurance/
   hands-free; video = virality/desire; motion graphics = spatial mechanics) —
   kept as a heuristic for *if/when* modalities beyond text+image+embeddings get
   added; does not change the standing modality-gap finding (CAPABILITIES, still
   true: no voice/video/animation wired in).
8. **Pool-level (not per-domain) ideation prompt**, corrected and adapted —
   D11, `IDEAS_...md` §4. Kept the underlying prompt-engineering technique
   (background context + explicit negative/positive examples prevents an LLM
   defaulting to "ask me anything" chatbot slop) but changed the unit from
   domain to pool, and added a hard groundedness gate that refuses to infer a
   vertical from a domain string.

### Extraction: DISCARDED or MUST-VERIFY-BEFORE-USE
1. **All 15 sample-domain archetype/tool assignments.** Checked against the
   live `sites` table 2026-07-24: **zero of the fifteen exist as platform
   sites.** Every assignment (e.g. "airportcollections.com = chauffeur
   transfers", "arabianperfumes.co.uk = luxury fragrance") was inferred from
   the domain string alone — exactly the failure mode `news_feed_pooling`
   already named ("profiles written from research, never from the domain name
   alone"). Kept as ungrounded brainstorm seed material only
   (`IDEAS_...md` §3), heavily flagged, never to be treated as a validated plan.
2. **Illustrative SQL** (an `agent_workflows` table with `INSERT` example rows)
   — does not match this platform's real schema
   (`agent_definitions.default_config.workflow.steps`); confabulated for
   illustration by the external LLM, not evidence of anything to reuse or a
   real design to follow.
3. **Per-run cost estimates** ("$0.001–$0.005/run", "$4.99–$9.99/mo") — the
   external LLM's own unverified guess, not checked against our real model
   pricing (see aiservice model tables). Don't quote as fact; would need
   grounding against actual Claude/Ollama per-token costs before appearing in
   any pricing decision.
4. **The "reality check" readiness table** — pure restatement of the
   CAPABILITIES inventory we ourselves supplied as context; no new information,
   and less rigorously evidenced (no file paths) than our own version.
5. **Running the deep-dive prompt per domain across the whole fleet** — doesn't
   scale to 1,000+ sites; superseded by D11 (pool-level ideation +
   lightweight per-site specialisation).
6. A **unified cross-portfolio account/payment** was never actually proposed by
   the external session, but the ambiguity was live in this workstream's own
   open questions. Owner closed it directly, restated 2026-07-24: domains stay
   independent (D14).

### What "define what we need to say and do, going forward" resolves to
(the actual answer to the owner's question this turn — full statement given in
chat, condensed here for the log): most sites get a nav-linked AI section built
as a 3-tier funnel (free algorithmic probes → cheap sticky utility → paid
signature operation), with an honest AI-impact editorial page (006's
three-level frame) sitting above it as the trust/SEO layer. We do not claim to
out-think foundation models; we sell the produced, verified, deployed thing a
single prompt can't produce, and every tool doubles as a live advert for the
framework itself. Chat is a delivery format inside Tier 2 or Tier 3, never the
product on its own. Still pure deliberation — nothing built.

## 2026-07-24b — second Gemini batch: 8 domains, this time "already on our
## system" — ground-checked, real corrections found

Owner brought a second external-LLM batch (8 domains: robot-hands.com,
gaswholesalers.com, agritec.uk, mortgagecalculator.co.uk, websitedesign.com,
leopardessconsulting.co.uk, idea.uk, ai-agent-orchestration.com), this time
claimed as domains already on the platform. Checked every one against the live
`sites` table before using any of it — same discipline as the first batch, but
this time worth doing thoroughly because 5/8 turned out real. Full per-domain
findings in `GROUNDED_domain_profiles_2026-07-24.md`; summary here.

**Still 3/8 ungrounded**: `agritec.uk`, `mortgagecalculator.co.uk`,
`websitedesign.com` do not exist in `sites`. Filed as brainstorm-only in
`IDEAS_...md` §3b, same treatment as the first batch's 15.

**5/8 real, and checking them found concrete, useful corrections** — pulling
real page titles + tagline (not `audience.v1`, which was empty on every site
checked so far) was what caught these:

- **robot-hands.com**: real site is a narrower, more mature "vendor-neutral
  gripper selection platform" than Gemini's broader guess (which wrongly
  included prosthetics research). More importantly: **Tier 1 already exists**
  (3 calculators live) and MatchMatrix already functions as a sticky Tier-2
  utility — Gemini's Tier-1/2 ideas were redundant with what's already built.
  The real gap is Tier 3, and a sharper candidate than Gemini's (CAD/video
  ingestion, no precedent) is extending MatchMatrix into a verified
  procurement dossier — which also structurally addresses `bugs_open/043`
  (fabricated stats found on this exact site).
- **gaswholesalers.com**: DB still shows the STALE content `006` exists to
  replace (checked live, 2026-07-24) — the rewrite hasn't shipped. Gemini's
  guess was unexpectedly well-aligned with `006`'s *corrected* positioning
  (independent convergent validation), but any build here must sequence behind
  `006` shipping, and must frame Tier-3 output as advisory-for-the-reader, not
  first-person operational claims — reintroducing that exact pattern is what
  caused `006` in the first place.
- **idea.uk**: Gemini's whole profile (patents, IP law, prior-art search) was
  **wrong** — free-associated "idea" → "invention" from the domain name alone,
  the sharpest miss in the batch. Real site validates business/product ideas,
  with a "Verified Idea Report" tool. **Then a memory update mid-session
  upgraded this further**: `idea-uk-vm-site-workstream` (a separate active
  session) verified LIVE, same day, that the home+header CTA funnels into a
  PAID `/report.html` tool — spot-checked directly against the live `pages`
  row (`report-request-form` section confirmed). idea.uk is not a pilot
  candidate, it's **the closest thing on the platform to a working reference
  implementation of this workstream's whole Tier-3 concept**, already live,
  owned by another session (follow-on `bugs_open`/`054` is theirs). Read their
  HANDOFF before proposing anything here — don't re-derive what's already
  built and don't start parallel work on an owned domain.
- **leopardessconsulting.co.uk**: Gemini's guess (generalist management/ESG
  consultancy, C-suite/PE personas, governance tools) was **wrong archetype
  entirely** — real tagline and real case studies (Companies House pipeline,
  tool-generation pipeline, news pipeline, hierarchical agent architecture) show
  this is an AI-engineering delivery consultancy, sibling to
  ai-agent-orchestration.com, both flagship framework-demo sites (PLAN D8).
  Real content already includes an "AI Production Readiness Assessment" quiz —
  a live Tier-1/2-ish tool Gemini didn't know about.
- **ai-agent-orchestration.com**: the best match of the whole batch — Gemini's
  guess was strongly aligned, real content includes 4 already-live tools (ROI
  estimator, LLM cost calculator, readiness quiz, complexity estimator) plus a
  genuine technical blog and case studies. Confirms D8 from an independent
  source. Gemini's "live 3-min fleet deploy" Tier-3 idea is on-strategy but
  risky (depends on the platform's most bug-fought delivery path); a lower-risk
  first step is a personalised "Architecture Fit Report" citing the real case
  studies, no live deploy required.

### Meta-lesson, stated plainly
Gemini's accuracy tracked **how literally the domain name described the real
business** — near-perfect for `ai-agent-orchestration.com`, badly wrong for
`idea.uk` and `leopardessconsulting.co.uk`. A domain name is not evidence
either way; sometimes it happens to be right. The only thing that caught the
misses was pulling real page titles and the real tagline. **Strengthening
D11's groundedness gate accordingly**: the pool-level ideation prompt must be
fed real page titles/content pulled from the platform, not just an
`audience.v1` reference — `audience.v1` was empty on every site checked in
both batches, so it isn't yet a reliable mechanism to point the gate at.

### Revised pilot recommendation (supersedes the 2026-07-21 "Property" pick)
Property was never grounded in a real domain — it was a pool, not a specific
site. Now grounded candidates exist, ranked in
`GROUNDED_domain_profiles_2026-07-24.md`: **robot-hands.com first** (real
audience, Tier 1/2 already live, one scoped Tier-3 gap that also fixes a known
bug), **ai-agent-orchestration.com + leopardessconsulting.co.uk** as a close
second pair (validates D8, both further along than any other domain checked),
**gaswholesalers.com** sequenced behind `006`, and **idea.uk treated as the
reference implementation to study, not a build target.**

# PLAN — Simple Paid Multi-Domain Chat

**Status: discussion draft — direction firming.** Kickoff for a deliberately
*simple* route, separate from the heavy chassis/satellite work. Open questions in
§9 remain the agenda.

The goal: give a basic, genuinely useful AI chat to *some* of many operated
domains (e.g. `gaswholesalers.com`, `robot-hands.com`, `websitedesign.com`,
`agritec.uk` — a wide spread of types), each configured for its own users, on a
**serverless** footing, with a **minimal end-user charge** because free across
many domains isn't affordable.

**Firmed this session:**
- These first domains are **"chat is the product"** — the visitor may pay, so the
  chat must justify a charge.
- Monetisation is **freemium + day-pass**: a capped free taster, then a flat
  time-pass (not counted credits). See §5–6.
- The hard part is the **value bar** (§4): paid chat must beat the free general
  AI a visitor already has — and there is **no proprietary data** to lean on, so
  differentiation has to be *built/sourced*. v1 is therefore **rich grounding
  pack + a tool**, not bounded-text-Q&A alone.

---

## 1. What this is — and isn't

**Is:** the *fast lane* — bounded, per-domain Q&A served by a portable edge
worker reading a per-domain config/context pack, with a light paywall. Add a
domain by publishing config + pointing DNS. No chassis, no Kafka, no satellite.

**Isn't:** the building-as-a-service / isolated-chassis route
(`PLAN_isolated_chat_environment.md`). That remains available later; this does
not depend on it and should not wait for it.

---

## 2. What we reuse

- **`FOCUS_site_chatbot_edge_worker_and_context_pack.md`** — almost entirely.
  The portable `handleChat` core, the per-domain **context pack** (identity +
  grounding + limits), the multi-tenant-by-Host worker (one worker, many domains,
  config as data not code), and the optional turn log are exactly this product.
  This plan is essentially that FOCUS doc + a paywall + a multi-domain rollout.
- **The billing provider interface** from `PLAN_stripe_billing_integration.md`
  (§4) — the same `Provider` abstraction can mint a visitor checkout, even though
  the payer and product differ (visitor buys chat access vs site-owner buys
  builds). Pluggability carries over.

The net-new surface is small: a paywall gate, a visitor-entitlement token, and a
per-domain config workflow that scales to many domains.

---

## 3. The real cost question (reframing "can't afford free")

Worth pinning before deciding to charge, because it changes the whole shape:

- **Legitimate inference is cheap.** A bounded Q&A turn with a small context pack
  on a *cheap* model (Haiku-class or local) is fractions of a cent. A normal
  visitor's whole session costs pennies.
- **So the true cost drivers are (a) model choice and (b) abuse** — bots
  hammering an open endpoint across many domains, and using an expensive model
  where a cheap one suffices. Those, not honest users, are what make "free"
  scary.
- **Implication:** with a cheap model + hard per-session caps + bot/rate
  protection, *free-with-limits* across many domains may actually be affordable —
  and avoids the friction and fee economics of charging $1 (below). Charging then
  becomes a fallback for heavy use, or reserved for domains where the chat is
  itself the product.

This doesn't kill the paid idea; it reframes it. We should decide to charge based
on a real per-domain cost estimate, not an assumption that usage is expensive.

---

## 4. The value bar — paid chat must beat free general AI

This is the crux of the whole product, and it is a *value* problem, not a
payments one. A visitor on any of these domains can open a free general AI in
another tab. So "pay £2 to chat with this site's bot" only converts if the bot
gives something the free model can't easily give. Plain bounded-Q&A-from-the-
site-text does **not** clear that bar — the general model can often do that as
well or better. Paid chat therefore forces a richer build than the simplest
version.

The things that can clear the bar:

1. **Proprietary / curated knowledge** the general model lacks. *Constraint: we
   have none to leverage* — so this has to be **built or sourced** (curated from
   authoritative public sources per niche, expert write-ups, structured
   reference we assemble). This is the unsolved part, covered by the framework in
   §10.
2. **Live or structured data** — prices, stock, specs, part/compatibility
   lookups. A general model can't see these; a tool that does is genuinely worth
   paying for.
3. **A niche tool / calculator** — a configurator, sizing/quote estimator,
   compatibility checker. The chat *does* something, not just answers.
4. **Convenience in context** — pre-loaded with the exact page/product context,
   zero prompting effort. Real but weak on its own; a supporting factor, not the
   whole case.

**Decision this session:** v1 is **rich grounding pack + a tool** (2 and 3),
because with no proprietary data, (1) alone isn't available yet and convenience
(4) alone won't sustain a charge. This is the deliberate step up from
bounded-text-Q&A, and it is the main thing that sizes the build (§9).

The no-proprietary-data gap is the open strategic question — how each domain's
differentiator gets built or sourced. The framework for working it is in §10
(still provisional).

---

## 5. Monetisation — chosen path and the micro-fee floor

**Chosen:** freemium + **day-pass** (options 2+3 below), per the firmed
direction. The rest is recorded for context.

The blunt constraint on "£1–2 per set": **card processing has a fixed
per-transaction fee** (roughly 20–30p plus a few percent — verify current
UK/Stripe rates). A £1 charge can lose a quarter-to-a-third to that fixed fee, so
sub-£5 one-off card payments are economically poor. Inference is cheap; *the
payment itself* is the expensive part — which is why the bundle should be
**generous** (a day-pass, not "a few messages for £1") priced where fees are a
small %.

Options considered (cheapest-friction first):

1. **Free, hard-capped, cheap-model** — lowest friction; best where the site's
   own business benefits (lead/conversion). Not these domains (they're
   "chat-is-the-product").
2. **Freemium paywall** *(chosen half)* — a capped free taster to prove value,
   then pay. Necessary: payment-first kills usage; nobody pays before seeing
   value.
3. **Pay-per-set as a day/week pass** *(chosen half)* — a flat time-pass priced
   at £2–5, generous because inference is cheap, fee ratio acceptable.
4. **Lead-capture / quote** — for business-chasing-leads domains; not this batch,
   but the same worker supports it as a per-domain flag later.

**The taster is where the cost actually goes.** Most tasters won't convert, so
the free taster's caps + bot protection (§9) are the real cost control even in a
paid model — echoing §3: the free edge is the cost driver, not paying users.

---

## 6. Entitlement — the day-pass flow (no accounts)

Day-pass collapses a lot of complexity, because a pass has no lifecycle and
nothing to decrement. The flow:

1. Visitor hits the paywall (taster exhausted).
2. **Stripe guest checkout** for a day-pass (provider interface from the billing
   plan; no account, no stored customer needed).
3. On return, a tiny **`redeem`** endpoint asks the provider *synchronously* "is
   this session paid?" → if yes, the worker signs a token carrying
   `{domain, expiry}` and hands it to the browser (cookie/local storage).
4. The worker **validates the token statelessly** on each message (signature +
   expiry + domain) and rate-limits within the window.

Consequences:

- **No webhook on the critical path.** Because the pass is issued by synchronous
  redeem, webhooks are an *optional backstop* for records, not part of issuing
  access. (Contrast the build product, where webhooks are the source of truth.)
- **No edge KV/counter.** The token is the entitlement; nothing to store or
  decrement. Counted credits — the alternative — would need per-token edge state;
  only reach for them if a pass genuinely doesn't fit.
- **Fair-use rate limit** inside the 24h window (a pass is "unlimited-ish", so cap
  messages/hour) to bound one pass's inference cost. Trivial with a cheap model.
- **Per-domain token** for v1 (signed for one domain). A cross-domain wallet
  spanning all chat sites needs shared identity/accounts — heavier, doesn't fit a
  spread of unrelated domains — parked.

---

## 7. Per-domain configuration (what "useful" means, and who authors it)

Each domain's chat is defined by its config/pack: identity, scope + refusal,
grounding, tone, model, limits, a **monetisation mode** (free-capped / freemium /
paid / lead-capture), and — now that v1 includes a tool (§4) — a **tool binding**
(which niche tool/data-lookup this domain exposes, and where its data comes from).

Open question: how is grounding *and the tool's differentiator* authored at the
scale of many domains?

- **Auto-derived grounding** from each site's own content (the context-pack
  builder from the FOCUS doc) — scales, but is the more involved path, and does
  *not* by itself clear the value bar (it's still site-text).
- **Hand-authored** config + curated grounding + a tool per domain — required to
  clear the value bar, but doesn't scale to "a vast array".

Likely: hand-author the first few (config + curated grounding + one tool) to
learn what genuinely justifies a charge per type, then template/automate once the
pattern is clear. The differentiator-sourcing (§4) is the hard input here.

---

## 8. Architecture sketch

```
visitor → static site (S3/edge) → /api/chat (portable edge worker, multi-tenant by Host)
   worker:
     1. load per-domain config/pack (identity, scope, grounding, limits,
        monetisation mode, tool binding)
     2. monetisation gate (per domain mode):
          freemium → allow taster within caps + rate limit;
                     beyond cap require a valid day-pass token
          paid     → require a valid day-pass token
          (free-capped / lead-capture modes supported for other batches later)
     3. answer: compose bounded prompt over curated grounding;
        if the turn needs data/compute, call the domain's tool; cheap-model LLM; stream
     4. (optional) record turn to a sink

   pass purchase: paywall → Stripe guest checkout (provider iface)
                → return → /redeem asks provider "paid?" (synchronous)
                → worker signs {domain, expiry} token → browser
   (webhook optional, for records only — not on the issuing path)
```

Static stays on S3; only `/api/chat`, the tool calls, and a small `redeem`/token
path are dynamic. Per-domain config (incl. tool binding) is data. This is the
FOCUS-doc worker with a freemium gate, a day-pass token, and a per-domain tool
added.

---

## 9. Open questions (discussion agenda)

**Resolved this session:**
- Charge or free? → these domains are chat-is-the-product; **charge**, via
  freemium + day-pass (§5).
- Pass vs credits? → **day-pass** (§6).
- Per-domain monetisation mix? → this batch is all paid; other modes deferred.

**Still open — next discussion (the differentiator problem):**

1. **The differentiator, with no proprietary data.** For each domain, *what* makes
   the chat worth paying for, and how is it built/sourced? Per §4 the candidates
   are curated-from-public-sources knowledge, a live/structured-data lookup, or a
   niche tool/calculator. This is the most important open question — it decides
   whether v1 is days or weeks of work, and whether the charge is justifiable at
   all. Worked in §10 (framework) and §11 (ideation agent).
2. **Cost estimate sanity-check** — even paid, confirm cheap-model + caps keep the
   taster (the real cost) affordable across many domains.
3. **Bundle price/size** that beats the fee floor while feeling minimal; confirm
   current UK Stripe fee structure.
4. **Free taster size** — enough to prove the *differentiated* value, capped to
   survive bots.
5. **Authoring at scale** — hand-author first N (config + curated grounding +
   tool) vs templating/automating once the pattern is clear.
6. **Recording** — keep the turn log (analytics, quality, dispute evidence) or
   omit for the simplest first cut?
7. **Bot/abuse protection** on the free path — rate limits, challenge, per-IP caps
   (the actual cost driver per §3).

---

## 10. Finding payable differentiators — framework (still needs work)

This is the method for answering §4: what makes a chat worth paying for when we
have no proprietary data. It is a starting point and **needs more thought and
testing** — the menus below are incomplete, the scoring is rough, and the
examples are illustrative, not decided.

### The core principle

The AI model is not what makes the chat worth paying for. Everyone has the same
models, including the visitor, so "AI that answers questions about the topic" or
"clever prompts" can be reproduced by a competitor or by the visitor with a free
tool. What cannot be easily reproduced is the specific thing we bring: data
others can't get, a process or output we own, a tool we have built to a high
standard, a commercial partnership we hold, or being first to use a new AI
capability before users know it exists. A payable differentiator is one of those
things combined with AI, aimed at people who will pay. If we cannot name the
specific hard-to-reproduce thing, there is no differentiator and a free tool does
the same job.

### Two menus we maintain over time

- **Assets (the hard-to-reproduce thing):** proprietary or paid data feeds; an
  owned process or output (e.g. our own site-spec-and-plan); a tool we build
  well; a commercial partnership; early access or timing on a new AI capability.
- **AI capabilities worth using now:** a living list kept current as models
  improve (image editing, structured extraction, tool use, vision, and so on).
  Adding a new capability here is the trigger to revisit every domain — this is
  how we act as early adopters before users know the capability exists.

### How to generate ideas

For each domain: take its audience and how much they can and will pay, then look
for a combination of one asset and one AI capability that does something a free
tool cannot do for that audience.

### Worked examples (the discussion that produced this)

- **websitedesign.com — strongest.** Asset: our own build pipeline's
  site-spec-and-plan output. Idea: package that spec/plan into a starter prompt
  the customer hands to other builders (Bolt, Lovable) to get better results from
  them. Hard to reproduce (it is our output), almost no extra AI work, and every
  prospective site-builder is a buyer. Low build cost, reusable across all design
  customers. Start here.
- **gaswholesalers.com — two different ideas under one label.** Strong version:
  buy proprietary oil/gas data feeds and apply AI on top, for traders and
  executives who pay well — the asset is the data access. Weak version: curate
  public information with prompts — no durable advantage, reproducible, free
  tools do the same. Only the data-feed version is payable.
- **Light / novelty sites** (e.g. dress a user's cat photo in a costume). Low
  defensibility, impulse purchase only — a one-off micro-purchase on a cheap
  site, not a recurring paid staple, and the quality bar to make the result good
  rather than poor is high. A different category from the serious domains.
- **agritec.uk** — vouchers or discounts through partnerships, with AI for
  matching/personalisation. The hard-to-reproduce thing is the partnerships; the
  work is commercial (signing partners), not technical.
- **Tools we already have** (e.g. a websitedesign mind-mapping or image tool)
  become payable only when developed to a high standard, or combined with AI to
  do something a free tool cannot. Until then they are features, not
  differentiators.

### Prioritising (kept deliberately light)

1. **Filter by cost to test first.** If an idea is cheap to put in front of
   users, test it directly and skip scoring. Only score ideas that are expensive
   to build.
2. **For the expensive ones, judge on four things:** how hard the asset is to
   reproduce, how much the audience will pay, build cost, and how far it reuses
   across domains.
3. **The combination to avoid:** a bespoke tool serving only one domain of
   middling value — high build cost, no reuse, modest return. Prefer either
   differentiators that reuse across many domains (one image capability across
   all light sites; the Bolt/Lovable prompt across all design customers) or
   single domains valuable enough to justify a bespoke build on their own (gas
   traders).

### Testing (validate demand before building)

- Test willingness to pay as early and as cheaply as possible — for example, show
  the day-pass purchase point and record whether people click it before the
  feature is fully built. If they don't click, don't build.
- Build the thinnest working version only after that demand test passes.
- Drop ideas that don't convert quickly — the free taster costs money whether or
  not anyone pays.

### Reusability caveat

Tools do not carry across unrelated domains the way a grounding pack does, so
"many domains, each with its own bespoke tool" is many separate builds. The
reusability judgement above is what keeps that from becoming a problem — push
toward reusable differentiators or single high-value domains, not the scattered
middle.

---

## 11. Ideation agent (a use of the existing agent framework)

A low-risk, internal use of the agent framework: it runs against our own data, not
live traffic, so none of the isolation concerns apply. It produces ideas for human
decision; it does not build anything. **Also still needs work** — treat the output
as candidates, not commitments.

- **Purpose:** given a domain and its audience, run the asset × capability
  combination from §10 and propose candidate payable differentiators with rough
  scores.
- **Inputs:** the domain; an audience description with ability/willingness to pay;
  the current asset menu; the current AI-capability menu.
- **Process:** for each relevant (asset × capability) pair, draft a candidate, a
  one-line reason it beats a free tool, and rough estimates for the four scoring
  factors; mark which candidates are cheap to test.
- **Output:** a ranked list split into "test now (cheap)" and "score/consider
  (expensive)", each candidate naming the asset it depends on and the AI
  capability it uses.
- **Why the framework helps:** it can be re-run across all domains whenever a new
  AI capability is added to the menu — the practical mechanism for catching
  early-adopter opportunities. It can spawn sub-agents to research an audience's
  willingness to pay, or to check whether a given data feed exists and what it
  costs.
- **Keep it internal and simple.** It is an idea generator feeding human
  judgement, not an automated builder.

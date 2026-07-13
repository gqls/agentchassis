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
   reference we assemble). This is "the bridge to cross", noted as the next
   discussion.
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
differentiator gets built or sourced — and is the "other route" to discuss next.

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

**Still open — next discussion (the "other route"):**

1. **The differentiator, with no proprietary data (the bridge).** For each
   domain, *what* makes the chat worth paying for, and how is it built/sourced?
   Per §4 the candidates are curated-from-public-sources knowledge, a live/
   structured-data lookup, or a niche tool/calculator. This is the load-bearing
   question — it decides whether v1 is days or weeks of work, and whether the
   charge is justifiable at all.
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

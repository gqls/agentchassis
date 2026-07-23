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

## Reuse / connections to verify before building
- Site-build machinery (sections/components/render, page rerender).
- Claims-verification / evidence_base fence — for grounded, honest editorial.
- News-feed pooling — vertical news feeds for freshness/grounding.
- fundamentallyai.com / brochure_component_library — the conversion destination.
- Councils / fix loops — the verification layer of any pipeline.
*(All [INFERRED] reusable from memory; confirm current state against the repo/DB
before relying on any of them.)*

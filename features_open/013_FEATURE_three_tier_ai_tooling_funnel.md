# 013 FEATURE — three-tier AI tooling funnel across the domain portfolio
# (generalizes 006/007 from one domain to most of the fleet)

**Raised:** 2026-07-21 → 2026-07-24, owner, session "AI page"
(`docs024_key_docs_latest/per_site_ai/`). Deliberation only — nothing designed
or built. This entry exists so the fleet-wide version of the idea has a
`features_open` anchor distinct from `006` (the single already-committed
instance) and `007` (the deferred chatbot for that one instance).
**Status:** strategy in progress, not designed.
**Depends on / related:** `006_FEATURE_gaswholesalers_repositioning_and_ai_influence_page.md`,
`007_FEATURE_ai_advisory_chatbot_freemium.md`, `003_FEATURE_paid_tier_beyond_news.md`,
`docs024_key_docs_latest/per_site_ai/` (PLAN/NOTES/README/SUMMARY/CAPABILITIES/IDEAS).

## The idea, compressed

Most sites get a nav-linked AI section, but it is **not a chatbot** — it is a
3-tier product funnel, and the chatbot (if any) is one delivery format inside
Tier 2 or Tier 3, not the product itself:

- **Tier 1 — free algorithmic probes.** Zero-LLM-cost, client-side or plain-Go
  calculators per vertical. SEO/intent-capture, no moat, cheap to spam widely.
- **Tier 2 — sticky "go-to" AI utility.** A cheap, narrow, habitual tool
  (fast/local model + cached RAG, ~$0.001–$0.005/run) that gives a professional
  in that vertical a reason to return weekly. New addition this round — closes
  the owner's ask for something that "draws targeted traffic" on top of the free
  probes and the expensive deliverable.
- **Tier 3 — signature AI operation.** The moat: an orchestrated, verified,
  produced artifact (report, deployed microsite, comparison, asset pack) that a
  single LLM prompt structurally cannot produce. Paywalled per-deliverable. For
  select high-value B2B/executive verticals this may instead be delivered
  **conversationally** — a paid, high-quality-model, deep-research, cached chat —
  which is what `007` actually specified, not a cheap intake bot.

## Why this generalizes 006/007 rather than replacing them

`006` already committed the owner to exactly this shape for one domain:
gaswholesalers.com repositioned for senior oil/gas readers, with an "AI
influence" nav page covering threats/advantages **honestly**, remediated at
**three levels — corporate, employee, personal**. That three-level frame is
sharper than this workstream's earlier, vaguer "job-title-specific articles"
language and is adopted here as the canonical shape for the honest-AI-impact
editorial layer, fleet-wide, not just on gaswholesalers.

`007` already specified the chatbot half — and specified it as a **high-quality
model**, deep-research, cached, paid-per-session/per-N-submissions product for a
senior audience: *"better than theirs"* because of the prompt engineering and
research behind it, not because of model access alone. That is the Tier-3
conversational variant above, not a Tier-2 cheap intake bot. This workstream's
earlier "chatbot = cheap intake front-door only" position (see PLAN D2) was too
narrow — `007` is the proof it doesn't hold for every vertical.

`003` already flagged the entitlement/payment plumbing this needs: a
`client_entitlements` local cache (auth can't be called synchronously across
thousands of sites per tick) and a possible `saas_cheap`/`portfolio` build-tier
flag from BIZ-014. **Do not invent a parallel payment mechanism for Tier 2/3
gating — check whether that flag/cache already exists before designing one.**

`006` also names the concrete gap the honest-AI-page content needs: a
`citation` source kind in `evidence_base` for research-sourced claims (current
schema only has `sql`/`artifact`/`attested_by`). This is exactly
claims-verification's **V5** (designed, not built — see
`per_site_ai/CAPABILITIES_framework_inventory_2026-07-21.md` §4). **Shipping an
evidence-backed AI-impact page platform-wide is gated on V5, not just on this
workstream's own design.**

## What's genuinely new since 006/007 (from a 2026-07-24 external-LLM strategy
session, cross-checked and partly discarded — see `per_site_ai/NOTES` for the
full kept/adapted/discarded breakdown)

- The **Tier 2 sticky-utility layer** itself, with 4 reusable generation
  templates (Extractor, Transformer, RAG Fast-Checker, Diagnostic Schema
  Generator) that could plausibly be built once and specialised per pool.
- An **archetype × pattern design grid**: 4 domain archetypes (High-Trust &
  Regulated / E-Commerce & Sensory / Technical & B2B / Historical & Creative —
  the WHO/WHY, risk and emotional register) crossed against this workstream's
  existing 5 signature-operation patterns (the WHAT gets produced) — orthogonal
  axes, not a replacement categorisation.
- A graduated monetisation ladder (free rate-limited → email-gate → micro-sub →
  per-deliverable paywall) refining this workstream's existing "paywall the
  deliverable" position into something with intermediate rungs.
- The "artifact as lead magnet" behavioural point: a co-designed deliverable
  makes an email-gate read as delivery, not friction — worth keeping as design
  guidance for Tier 2/3 gating.

## What was discarded or must be re-verified before use

- Fifteen sample domains were run through a deep, per-domain strategy prompt.
  **None of the fifteen exist in the platform's `sites` table** (checked
  2026-07-24) — they are unonboarded names, so every archetype/vertical
  assignment for them was inferred from the **domain string alone**. This is
  the exact trap `news_feed_pooling` already named ("profiles written from
  research, never from the domain name alone"). Treat the resulting tool ideas
  as unverified brainstorm raw material only (kept in
  `per_site_ai/IDEAS_gemini_domain_brainstorm_2026-07-24.md`), never as a
  validated per-domain plan.
- Illustrative SQL (an `agent_workflows` table with `INSERT` rows) does not
  match this platform's real schema (`agent_definitions.default_config.workflow.steps`)
  — confabulated for illustration, not evidence of anything to reuse.
  Per-run cost figures ($0.001–$0.005) are the same LLM's own unverified
  estimate, not measured against our real model pricing.
- Running the deep-dive strategy prompt **per domain** does not scale to
  1,000+ sites and re-derives the same thinking repeatedly. Repurposed instead
  as a **pool-level** ideation tool (≤17 runs) — see `per_site_ai/PLAN` D11.

## Open questions

1. ~~Which pool is the first Tier-3 pilot?~~ **Updated 2026-07-24**: a second
   external-LLM batch, checked against real `sites`/`pages` data, found
   concrete grounded candidates — see
   `per_site_ai/GROUNDED_domain_profiles_2026-07-24.md`. Leading pick:
   **robot-hands.com** (Tier 1/2 already live, one scoped Tier-3 gap that also
   remediates `bugs_open/043`). Close second: `ai-agent-orchestration.com` +
   `leopardessconsulting.co.uk` as a pair (both AI-services/D8 sites, both
   already have live Tier-1/2 tools). `gaswholesalers.com` stays queued behind
   `006` shipping. **`idea.uk` is not a pilot candidate — it already has a
   live, paid, gated report-delivery funnel** (verified same day by
   `idea-uk-vm-site-workstream`, a separate owning session); study it as a
   working reference before designing anything new.
2. Does building 013 mean reopening `007`? **No, not by default** — `007` was
   explicitly deferred by the owner ("another feature for another time") and
   this entry does not reopen it. It only records that 013's design must not
   contradict what `007` already specified, in case the owner reopens it later.
3. Entitlement plumbing: does the BIZ-014 tier flag / `client_entitlements`
   cache already exist? Check before any Tier 2/3 gating design starts.

## Cross-note (from the idea-uk-vm-site session, owner 2026-07-24)

Open-question 1 above frames **idea.uk as a working *reference*, not a pilot** (because its paid
Tier-3 funnel is already live). That still holds for choosing the first *pilot*. But the owner has now
also asked (to the `idea-uk-vm-site` session) to **enhance idea.uk itself** with the full three-tier
funnel — a complete "idea lifecycle" content + tooling pipeline (ideate → build → test → UAT → feedback
→ **patents** → copyright → funding ways → funding sources → more), with free and paid tools per stage,
built *using this strategy*. Captured as **`features_open/014_FEATURE_idea_uk_ideas_pipeline.md`**.

So idea.uk is **both** the reference funnel **and** (now) an enhancement target — not a contradiction.
When 013's design matures, coordinate with 014/`idea-uk-vm-site` rather than treating idea.uk as
off-limits; idea.uk is the one place the whole funnel can be exercised against a real, live paid tier.

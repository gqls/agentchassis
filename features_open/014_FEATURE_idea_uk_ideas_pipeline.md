# 014 FEATURE — idea.uk as a full "idea lifecycle" content + tooling pipeline

**Raised:** 2026-07-24, owner, to the `idea-uk-vm-site` session (during the home-page
CTA-funnel fix). **Status:** REQUESTED — captured, not designed, not built. Deliberation
only, same bar as `013`.
**Owns / lives with:** `idea-uk-vm-site-workstream` (the site) +
`per_site_ai` / `013_FEATURE_three_tier_ai_tooling_funnel.md` (the funnel mechanics).
**Site:** idea.uk, site_id `1244516d-014d-421c-88c6-090bb1e9552a`.

## The ask (owner, verbatim intent)

> "I'd like a section on **patents and what to do with ideas**. I'd like a whole set of
> **guides and tools** from helping users create ideas through to **building, testing,
> user acceptance, feedback loops, patents, copyright, funding ways, funding sources, and
> a whole load more in that pipeline**. On the way I'd like to add **tools both free and
> paid for** (see the 'AI tooling strategy' page, that currently uses idea.uk as an
> example but that we can use to enhance idea.uk too)."

So idea.uk should grow from "a marketing site + one paid £29 report" into a **guided
journey across the whole life of an idea**, with content (guides) *and* interactive tools
(free and paid) at each stage — and it should be built using the same three-tier AI-tooling
funnel that `013`/`per_site_ai` designed with idea.uk only as an *example*. This entry flips
idea.uk from **reference** to **enhancement target** (see "Reconciling with 013" below).

## The pipeline (owner's stages, in order — the spine of the content + tools)

Each stage wants at least a **guide** (editorial), and where it earns its place, a **tool**
(Tier 1 free probe / Tier 2 sticky utility / Tier 3 paid signature operation):

1. **Create / ideate** — help users generate and shape ideas.
2. **Build** — turning an idea into something real (MVP, prototyping).
3. **Test** — validation, experiments.
4. **User acceptance (UAT)** — does it do what users need.
5. **Feedback loops** — iterate on real user input.
6. **Patents** — *the owner led with this: a dedicated section on patents and what to do
   with an idea* (patentability, prior art, filing routes, timing vs disclosure).
7. **Copyright** — protection beyond patents (what copyright does/doesn't cover, when it
   applies automatically, registration where relevant).
8. **Funding — ways** — the mechanisms (bootstrapping, grants, equity, debt, crowdfunding,
   revenue-based, competitions…).
9. **Funding — sources** — the actual places (UK grants/bodies, angel networks, VCs,
   accelerators, banks, platforms…).
10. **"…and a whole load more in that pipeline"** — explicitly open-ended: e.g. market
    validation, pricing, go-to-market, legal/company formation, hiring, scaling, exit.
    Treat the list above as the seed, not the ceiling.

## Free vs paid tooling (map to the 013 three-tier funnel)

idea.uk already proves Tier 3 (the paid £29 "verified idea report", live and gated). This
vision adds Tiers 1–2 across the pipeline, and likely more Tier-3 deliverables:

- **Tier 1 — free probes** (zero-LLM-cost, SEO/intent capture): e.g. an *idea-validation
  scorecard*, a *patentability pre-check* (is-this-even-novel questionnaire), a *funding-fit
  finder* (which funding route/source matches your stage), a *UAT checklist generator*.
- **Tier 2 — sticky utilities** (cheap, habitual, fast model + cached RAG): e.g. a *prior-art
  first-look*, a *pitch/one-pager helper*, a *feedback-synthesis* tool, a *grant-matcher*.
- **Tier 3 — paid signature operations** (orchestrated, verified artifacts): the existing
  verified idea report, plus candidates like a *patent-landscape report*, a *funding-strategy
  pack*, a *go-to-market plan*.

**This is illustrative, not a committed tool list.** Which stages get a tool, at which tier,
and free-vs-paid, is a design/deliberation question — and any tool making factual/legal
claims (patents, copyright, funding eligibility) is **high-stakes** and must go through
claims-verification (evidence-backed, no fabricated legal advice).

## What idea.uk already has — do NOT rebuild

- 9-page site + the paid £29 report tool, one origin, live on the VM.
- **A live, gated, paid Tier-3 funnel** — the whole home page + header now drive to
  `/report.html` (the `idea-uk-vm-site` session, 2026-07-23/24, sql/p3_05-07). Build the
  pipeline *on top of* this, don't reinvent the paywall/entitlement path.
- `guides-index` (`/guides/index.html`) and `tools` (`/tools.html`) pages already exist as
  hubs — the guide/tool library has somewhere to live.

## Reconciling with 013 / per_site_ai (important — another session owns that framing)

`013` currently says *"idea.uk is not a pilot candidate — it already has a live, paid, gated
report-delivery funnel … study it as a working reference before designing anything new."*
That was written to pick a *first pilot* (robot-hands.com). This entry does **not** contradict
it: idea.uk stays the working *reference* for the funnel mechanics, **and** the owner now
also wants idea.uk *enhanced* with the full pipeline. Both are true. A cross-note has been
added to `013` so the `per_site_ai` session sees the owner's enhancement intent; coordinate
rather than fork the funnel design.

## Dependencies / gates (from 013, do not re-derive)

- **Entitlement/payment plumbing** for any new paid tier: check for the BIZ-014 tier flag /
  `client_entitlements` cache before inventing anything (`003`, `013` open q3). idea.uk's
  existing tool has its OWN order/payment path (file-based, `/var/lib/idea/orders.json`, no
  DB) — reconcile with the platform mechanism deliberately.
- **Evidence-backed claims** (patents/copyright/funding facts): gated on claims-verification
  **V5** (citation source kind), per `013` and `per_site_ai/CAPABILITIES §4`.
- Content is DERIVED on render (CTA urls, tool lists, page bodies) — the `idea-uk-vm-site`
  landmines apply to any new page/tool (verify against the live page, not the work-item).

## Next step (when the owner prioritises this)

Design pass, not build: pick the first 1–2 pipeline stages to instantiate (the **patents**
section is the owner's lead), decide guide-vs-tool and tier per stage, and run it through the
per_site_ai / 013 funnel design + claims-verification. Keep it as deliberation until the
owner says go — same discipline as 013.

## Cross-links
- `docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/` (the site: PLAN/README/RUNNING_NOTES)
- `features_open/013_FEATURE_three_tier_ai_tooling_funnel.md` (the funnel; cross-noted)
- `docs/agent_docs/docs024_key_docs_latest/per_site_ai/` (the strategy: PLAN/IDEAS/GROUNDED)
- `features_open/003_FEATURE_paid_tier_beyond_news.md` (entitlement plumbing)

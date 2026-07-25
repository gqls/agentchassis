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

## BUILD LOG — increment 1 SHIPPED (2026-07-25): the patents guide

Owner said *"Let's carry on with idea.uk specifically in this thread"*, so stage 6 (patents — his
lead) was built, not just designed. **`/guides/patents/index.html` is LIVE** (HTTP 200 verified
2026-07-25 08:42 UTC, 39,214 bytes, full chrome, every CTA funnelling to `/report.html`).

**What this changed structurally, beyond one page:**

- idea.uk had **zero** `page_type='guide'` pages and an **empty guides hub** — the hub's listing
  component (`content-listing`) reads a STATIC `articles` array with no query source, so it would
  never have picked up a guide however many were written. Swapped to `guide-list_pre_037`, which
  sources `items` from `query.pages_where_type:guide`. **Every future guide now lists itself on the
  hub with no further edit** — that is the reusable half of this increment.
- The repeatable build path is written up as **RUNBOOK Phase 5** (`idea_uk_vm_site/`): create page
  + sections with authored `content_data`, direct-fire a `section_data_resolved` page-rerender,
  verify the LIVE page, backfill `pages.sections`, lock. It deliberately does **not** re-run
  `build-site-planner` — re-planning to add one page is how built pages get clobbered
  (`bugs_open/001`, `050`). Three traps documented, including the one that cost a round here
  (`slot_name` NULL ⇒ renders nothing ⇒ job still reports COMPLETED).

**Content policy set by this increment, and it should hold for the rest of the pipeline.** The body
is **hand-authored**, not LLM-generated, and the sections are locked. Reason: claims-verification
**V5** (the citation/evidence gate this feature already names as a dependency) is BUILT BUT INERT,
and `bugs_open/043` is a live fabricated-content lane. Patents, copyright and funding eligibility
are exactly the stages where a plausible invention is most costly. A draft error caught pre-ship
makes the point: it claimed the IPEC *small claims track* makes patent enforcement affordable —
that track does not hear patent claims at all. **Until V5 is live, stages 6–9 should be authored;
stages 1–5 (ideate/build/test/UAT/feedback) are lower-stakes and can take generated copy.**

SQL: `idea_uk_vm_site/sql/p4_01` (page+content), `p4_01b` (slot_name fix), `p4_01c`
(`pages.sections` backfill), `p4_02` (hub swap), `p4_03` (locks). Full record incl. both missteps:
`idea_uk_vm_site/RUNNING_NOTES §X.12`.

**Remaining in flight at time of writing:** the hub's own re-render is queued behind a stalled
generic-requests consumer (`bugs_open/029/030`) — the swap is in the DB but the live hub still
shows the old empty listing until it drains. Verify with `curl`, not the DB.

## BUILD LOG — increment 2 SHIPPED (2026-07-25, same day): copyright + the first free tool

Owner: *"yes for copyright and the checker"*. Both built and live.

**Stage 7 — `/guides/copyright/index.html`** (`sql/p4_05`). Hand-authored under the policy above.
Leads on the highest-value practical fact, which is invisible from the patents guide: **a
contractor keeps copyright unless there is a written, signed assignment** (CDPA 1988 s.90(3));
only employees' work vests in the employer automatically (s.11(2)). Says the AI questions
(ownership of output, lawfulness of training) are **unsettled** rather than picking a side — the
easiest paragraph in the guide to write confidently and wrongly. Auto-listed on the hub by `p4_02`'s
derived listing with no edit to the hub, which is increment 1's reusable half paying off first time.

**First Tier-1 free tool — `/tools/patent-check/index.html`** (`sql/p4_06`), new `patent-check`
component. Six questions, entirely client-side: no LLM, no backend, no data leaves the browser, so
it costs nothing per use — a genuine Tier-1 probe per `013`, not a Tier-2 utility wearing the label.
`page_type='tool'` means `/tools.html` lists it automatically via `query.pages_where_type:tool`,
the same self-listing mechanism as the guides hub.

**The design decision worth carrying into every future tool in this pipeline.** Reuse of the
existing `ai-readiness-quiz` component was checked and **rejected**: it is a fixed **sum-score**
quiz, and the questions in this domain are not additive. "Have you already disclosed it publicly?"
is close to dispositive on its own — no UK grace period — so a sum would let five good answers
outvote it and tell someone who has already published that they look patent-ready. That is not a
UX mismatch; it is advice that could cost a reader their rights. `patent-check` is therefore
**gated, not scored**: disclosure and subject-matter exclusions short-circuit to their own outcomes
first, and only if both pass does the commercial question get scored into bands.

> **Rule for stages 6–9 tools:** where a single answer can be decisive (a legal deadline, an
> exclusion, an eligibility bar), the instrument must gate before it scores. A scorecard is only
> safe when every input is genuinely a matter of degree. This applies directly to the funding-fit
> finder in stage 8/9 — eligibility criteria are gates, not points.

Second reason reuse failed, worth knowing before reaching for that component again: its
`quiz_badge_label` is `source: static` with a fallback, which `p4_04` established is
**unoverridable**, so the badge would have read "AI Readiness Assessment" on a patent checker.

**Delivery notes that generalise:** the tool's JS is **inline in `html_template`**, not an external
`/tools/assets/*.js`, to stay off the publish-but-never-load path (`bugs_open/041`, the
`js_content`/`js_snippets` trap) — an inline script is part of the rendered section and cannot be
published-but-not-loaded. The template was parse+execute tested locally before insert. And the URL
was checked against idea.uk's nginx first: tool routes are reserved by **exact match** only, so
`/tools/` is served statically (probed live) — a page under a proxied prefix would have been
invisible however well it rendered.

**Next stages:** the funding pair — 8 (ways) then 9 (sources) — which is where the gating rule above
gets its second real test.

## Next step (when the owner prioritises this)

Design pass, not build: pick the first 1–2 pipeline stages to instantiate (the **patents**
section is the owner's lead), decide guide-vs-tool and tier per stage, and run it through the
per_site_ai / 013 funnel design + claims-verification. Keep it as deliberation until the
owner says go — same discipline as 013.

## Sequencing (owner 2026-07-24)

idea.uk's full build is **queued behind the already-listed pilots** (robot-hands.com etc., per
`013`): the pilots go first, idea.uk stays the working *reference* while they do, THEN idea.uk
is built out to this full pipeline — becoming the **top-rung exemplar**. After that, other sites
are brought up to the same level, but **in stages, not one leap** — the staged "site maturity
ladder" captured as **`features_open/015_FEATURE_staged_site_maturity_ladder.md`** (to be planned
in a separate thread). So 014 is both a build target *and* the worked example rung 015 points other
sites at.

## Cross-links
- `docs/agent_docs/docs024_key_docs_latest/idea_uk_vm_site/` (the site: PLAN/README/RUNNING_NOTES)
- `features_open/013_FEATURE_three_tier_ai_tooling_funnel.md` (the funnel; cross-noted)
- `features_open/015_FEATURE_staged_site_maturity_ladder.md` (the fleet-wide staged rollout)
- `docs/agent_docs/docs024_key_docs_latest/per_site_ai/` (the strategy: PLAN/IDEAS/GROUNDED)
- `features_open/003_FEATURE_paid_tier_beyond_news.md` (entitlement plumbing)
